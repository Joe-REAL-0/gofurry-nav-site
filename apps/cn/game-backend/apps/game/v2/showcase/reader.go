package showcase

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	gamesqlc "github.com/gofurry/gofurry-game-backend/internal/db/game/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Reader interface {
	Read(context.Context, time.Time, string, string) (Inputs, error)
}
type SQLReader struct{ Pool *pgxpool.Pool }

func (r SQLReader) Read(ctx context.Context, now time.Time, lang, region string) (Inputs, error) {
	input := Inputs{Candidates: []Candidate{}, Diagnostics: map[string][]Diagnostic{"upcoming": {}, "new_release": {}, "trending": {}}}
	tx, e := r.Pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if e != nil {
		return input, e
	}
	defer tx.Rollback(ctx)
	q := gamesqlc.New(tx)
	at := pgtype.Timestamptz{Time: now, Valid: true}
	managed, e := q.ListShowcaseManaged(ctx, gamesqlc.ListShowcaseManagedParams{Lang: lang, AtTime: at})
	if e != nil {
		return input, e
	}
	boundary, e := q.NextShowcaseBoundary(ctx, gamesqlc.NextShowcaseBoundaryParams{Lang: lang, AtTime: at})
	if e != nil {
		return input, e
	}
	if boundary.Valid {
		input.Boundary = boundary.Time
	}
	games, e := q.ListShowcaseGames(ctx, lang)
	if e != nil {
		return input, e
	}
	rows, e := q.ListShowcasePlayerFacts(ctx)
	if e != nil {
		return input, e
	}
	facts := map[int64][]PlayerFact{}
	horizon := time.Time{}
	for _, f := range rows {
		avg := 0.0
		if f.AvgPlayers != nil {
			avg = *f.AvgPlayers
		}
		facts[f.TrackingPeriodID] = append(facts[f.TrackingPeriodID], PlayerFact{f.FactDate.Time, avg, f.SuccessfulSamples, f.ExpectedSamples})
		horizon = f.ProcessedThrough.Time
	}
	releases := map[string]*Release{}
	for _, g := range games {
		gameID := strconv.FormatInt(g.ID, 10)
		release := gameRelease(g)
		releases[gameID] = release
		trend := Trend{}
		if g.TrackingPeriodID != nil {
			trend = Trending(facts[*g.TrackingPeriodID], horizon)
		}
		for _, pool := range []string{"upcoming", "new_release", "trending"} {
			c, d := automaticCandidate(g, pool, now, trend)
			input.Diagnostics[pool] = append(input.Diagnostics[pool], d)
			if d.Candidate {
				input.Candidates = append(input.Candidates, c)
			}
		}
	}
	for _, row := range managed {
		if c, ok := managedCandidate(row); ok {
			c.Item.Release = releases[c.Item.GameID]
			input.Candidates = append(input.Candidates, c)
		}
	}
	return input, tx.Commit(ctx)
}
func value(p *string) string {
	if p == nil {
		return ""
	}
	return strings.TrimSpace(*p)
}
func date(p pgtype.Date) *string {
	if !p.Valid {
		return nil
	}
	s := p.Time.Format(time.DateOnly)
	return &s
}
func gameRelease(g gamesqlc.ListShowcaseGamesRow) *Release {
	if g.Availability == nil || g.Precision == nil {
		return nil
	}
	return &Release{value(g.Availability), value(g.Precision), date(g.ExactDate), g.ReleaseYear, g.ReleaseMonth, g.ReleaseQuarter, date(g.WindowStart), date(g.WindowEnd)}
}
func validHTTPS(raw string) bool {
	u, e := url.Parse(raw)
	return e == nil && u.Scheme == "https" && u.Hostname() != "" && u.User == nil && !strings.ContainsAny(raw, "\r\n\t")
}
func steamArtwork(raw string) bool {
	if !validHTTPS(raw) {
		return false
	}
	u, _ := url.Parse(raw)
	host := strings.ToLower(u.Hostname())
	for _, suffix := range []string{"steamstatic.com", "steamcdn-a.akamaihd.net", "st.dl.eccdnx.com", "steamchina.queniuam.com", "steamusercontent.com"} {
		if host == suffix || strings.HasSuffix(host, "."+suffix) {
			return true
		}
	}
	return false
}
func automaticCandidate(g gamesqlc.ListShowcaseGamesRow, pool string, now time.Time, trend Trend) (Candidate, Diagnostic) {
	id := strconv.FormatInt(g.ID, 10)
	locale := strings.TrimSpace(g.Title) != "" && strings.TrimSpace(g.Summary) != ""
	for _, t := range g.Tags {
		locale = locale && strings.TrimSpace(t) != ""
	}
	d := Diagnostic{GameID: id, Name: g.Title, ShowcaseEligible: g.ShowcaseEligible, SFW: g.Sfw != nil && *g.Sfw, LocaleReady: locale, ArtworkReady: steamArtwork(g.ArtworkUrl), Release: gameRelease(g), ExcludedReasons: []string{}}
	if !d.ShowcaseEligible {
		d.ExcludedReasons = append(d.ExcludedReasons, "not_approved")
	}
	if !d.SFW {
		d.ExcludedReasons = append(d.ExcludedReasons, "adult")
	}
	if !d.LocaleReady {
		d.ExcludedReasons = append(d.ExcludedReasons, "locale_incomplete")
	}
	if !d.ArtworkReady {
		d.ExcludedReasons = append(d.ExcludedReasons, "steam_artwork_unavailable")
	}
	weight := 0
	today := now.UTC().Truncate(24 * time.Hour)
	switch pool {
	case "upcoming":
		if value(g.Availability) == "upcoming" {
			weight = 100
			planned := g.ExactDate
			if !planned.Valid {
				planned = g.WindowStart
			}
			if planned.Valid && planned.Time.Before(today) && g.WindowEnd.Valid && !g.WindowEnd.Time.Before(today) {
				planned.Time = today
			}
			if planned.Valid && !planned.Time.Before(today) {
				weight = 150
				if !planned.Time.After(today.AddDate(0, 0, 180)) {
					weight = 300
				}
			} else if g.WindowEnd.Valid && !g.WindowEnd.Time.Before(today) {
				weight = 150
			}
		}
	case "new_release":
		if g.FirstAvailable.Valid {
			age := int(today.Sub(g.FirstAvailable.Time) / (24 * time.Hour))
			if !g.FirstAvailable.Time.After(today) && age <= 30 {
				weight = 100
				if age <= 7 {
					weight = 300
				} else if age <= 14 {
					weight = 200
				}
			}
		}
	case "trending":
		d.Trending = &trend
		if trend.Eligible {
			weight = trend.Weight
		}
	}
	if weight == 0 {
		d.ExcludedReasons = append(d.ExcludedReasons, pool+"_requirements")
	}
	d.Candidate = len(d.ExcludedReasons) == 0
	d.FirstAvailable = date(g.FirstAvailable)
	d.PoolFailures = diagnosticPoolFailures(d, pool, today)
	tags := append([]string{}, g.Tags...)
	if len(tags) > 3 {
		tags = tags[:3]
	}
	item := Item{Key: "auto:" + pool + ":" + id, Source: "automatic", Reason: pool, ContentType: "game", GameID: id, Title: strings.TrimSpace(g.Title), Summary: strings.TrimSpace(g.Summary), Tags: tags, Artwork: Artwork{Kind: "steam", URL: g.ArtworkUrl}, Release: d.Release, PrimaryAction: Action{Type: "game", GameID: id}}
	if g.Appid > 0 {
		item.SecondaryAction = &Action{Type: "steam", Target: fmt.Sprintf("https://store.steampowered.com/app/%d/", g.Appid)}
	}
	return Candidate{Item: item, Weight: weight}, d
}

var managedKey = regexp.MustCompile(`^game/showcase/([1-9][0-9]*)/(desktop|mobile)/[a-f0-9]{32}\.avif$`)

func keyFor(key, id, variant string) bool {
	m := managedKey.FindStringSubmatch(key)
	return len(m) == 3 && m[1] == id && m[2] == variant
}
func managedCandidate(row gamesqlc.ListShowcaseManagedRow) (Candidate, bool) {
	c, l := row.GfgShowcaseCampaign, row.GfgShowcaseCampaignLocale
	id := strconv.FormatInt(c.ID, 10)
	if row.Sfw == nil || !*row.Sfw || !row.GameExists || !l.Enabled || strings.TrimSpace(l.Title) == "" || strings.TrimSpace(l.Summary) == "" || len(l.Tags) > 3 || c.Weight < 1 || c.Weight > MaxSelectionWeight || c.Sponsored && value(l.EditorialNote) != "" || !keyFor(value(c.DesktopObjectKey), id, "desktop") {
		return Candidate{}, false
	}
	if value(c.MobileObjectKey) != "" && !keyFor(value(c.MobileObjectKey), id, "mobile") {
		return Candidate{}, false
	}
	gameID := ""
	if c.LinkedGameID != nil {
		gameID = strconv.FormatInt(*c.LinkedGameID, 10)
	}
	primary := Action{Type: value(c.PrimaryActionType)}
	switch primary.Type {
	case "game":
		if gameID == "" {
			return Candidate{}, false
		}
		primary.GameID = gameID
	case "project", "product", "website":
		primary.Target = value(c.PrimaryTarget)
		if !validHTTPS(primary.Target) {
			return Candidate{}, false
		}
	default:
		return Candidate{}, false
	}
	var secondary *Action
	if s := value(c.SecondaryActionType); s != "" {
		switch s {
		case "steam", "kickstarter", "website", "other":
		default:
			return Candidate{}, false
		}
		if !validHTTPS(value(c.SecondaryTarget)) {
			return Candidate{}, false
		}
		secondary = &Action{Type: s, Target: value(c.SecondaryTarget)}
	} else if value(c.SecondaryTarget) != "" {
		return Candidate{}, false
	}
	reason := "editorial"
	if c.Sponsored {
		reason = "sponsored"
	}
	item := Item{Key: "campaign:" + id, Source: "managed", Reason: reason, ContentType: c.ContentType, Sponsored: c.Sponsored, CampaignID: id, GameID: gameID, Title: l.Title, Summary: l.Summary, Tags: append([]string{}, l.Tags...), EditorialNote: value(l.EditorialNote), Artwork: Artwork{Kind: "managed", DesktopObjectKey: value(c.DesktopObjectKey), MobileObjectKey: value(c.MobileObjectKey), FocalX: &c.FocalX, FocalY: &c.FocalY}, PrimaryAction: primary, SecondaryAction: secondary}
	pin := 0
	if c.PinPosition != nil {
		pin = int(*c.PinPosition)
	}
	return Candidate{Item: item, Weight: int(c.Weight), Pin: pin}, true
}
