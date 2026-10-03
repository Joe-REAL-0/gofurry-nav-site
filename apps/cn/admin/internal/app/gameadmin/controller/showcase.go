package controller

import (
	"context"
	"fmt"
	"io"
	"math"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gofiber/fiber/v3"
	"github.com/gofurry/gofurry-admin/internal/app/shared/adminutil"
	"github.com/gofurry/gofurry-admin/internal/app/shared/audit"
	gamesqlc "github.com/gofurry/gofurry-admin/internal/db/game/sqlc"
	"github.com/gofurry/gofurry-admin/internal/infra/assets"
	"github.com/gofurry/gofurry-admin/pkg/common"
	"github.com/jackc/pgx/v5/pgtype"
)

// WithShowcaseRevision supplies the post-commit side effect for isolated
// integrations. Normal runtime construction uses its existing Redis client.
func (api *GameAPI) WithShowcaseRevision(bump func(context.Context) error) *GameAPI {
	api.store.showcaseRevision = bump
	return api
}

type showcaseLocale struct {
	Lang          string   `json:"lang"`
	Enabled       bool     `json:"enabled"`
	Title         string   `json:"title"`
	Summary       string   `json:"summary"`
	Tags          []string `json:"tags"`
	EditorialNote *string  `json:"editorial_note"`
}
type showcaseWorkspace struct {
	gamesqlc.GfgShowcaseCampaign
	ID            string           `json:"id"`
	LinkedGameID  *string          `json:"linked_game_id"`
	Locales       []showcaseLocale `json:"locales"`
	DerivedStatus string           `json:"derived_status"`
	Ready         bool             `json:"ready_to_publish"`
	Diagnostics   []string         `json:"publication_diagnostics"`
}
type showcaseContent struct {
	InternalName        string           `json:"internal_name"`
	ContentType         string           `json:"content_type"`
	Sponsored           bool             `json:"sponsored"`
	LinkedGameID        *int64           `json:"linked_game_id"`
	Locales             []showcaseLocale `json:"locales"`
	FocalX              float64          `json:"focal_x"`
	FocalY              float64          `json:"focal_y"`
	PrimaryActionType   *string          `json:"primary_action_type"`
	PrimaryTarget       *string          `json:"primary_target"`
	SecondaryActionType *string          `json:"secondary_action_type"`
	SecondaryTarget     *string          `json:"secondary_target"`
}
type showcaseSchedule struct {
	StartsAt    *time.Time `json:"starts_at"`
	EndsAt      *time.Time `json:"ends_at"`
	Weight      int32      `json:"weight"`
	PinPosition *int16     `json:"pin_position"`
}

func showcaseState(c gamesqlc.GfgShowcaseCampaign, now time.Time) string {
	if c.State != "published" {
		return c.State
	}
	if now.Before(c.StartsAt.Time) {
		return "scheduled"
	}
	if !now.Before(c.EndsAt.Time) {
		return "ended"
	}
	return "active"
}
func showcaseString(p *string) string {
	if p == nil {
		return ""
	}
	return strings.TrimSpace(*p)
}
func showcaseMember(v string, options ...string) bool {
	for _, o := range options {
		if o == v {
			return true
		}
	}
	return false
}
func showcaseURL(v string) bool {
	u, e := url.Parse(v)
	return e == nil && u.Scheme == "https" && u.Hostname() != "" && u.User == nil && !strings.ContainsAny(v, "\r\n\t")
}
func showcaseTime(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: t.UTC(), Valid: true}
}

func readShowcaseLocales(ctx context.Context, q *gamesqlc.Queries, id int64) ([]showcaseLocale, error) {
	rows, e := q.ListShowcaseLocales(ctx, id)
	if e != nil {
		return nil, e
	}
	out := make([]showcaseLocale, 0, len(rows))
	for _, l := range rows {
		out = append(out, showcaseLocale{l.Lang, l.Enabled, l.Title, l.Summary, l.Tags, l.EditorialNote})
	}
	return out, nil
}

// Incomplete drafts are allowed. Shape and supplied values are always validated.
func validateShowcaseShape(c gamesqlc.GfgShowcaseCampaign, locales []showcaseLocale) error {
	if strings.TrimSpace(c.InternalName) == "" || utf8.RuneCountInString(c.InternalName) > 160 {
		return common.NewValidationError("internal_name is required (max 160)")
	}
	if !showcaseMember(c.ContentType, "game", "crowdfunding", "tabletop", "merchandise", "other") {
		return common.NewValidationError("invalid content_type")
	}
	if c.Weight < 1 || c.Weight > 10000 || c.PinPosition != nil && (*c.PinPosition < 1 || *c.PinPosition > 4) {
		return common.NewValidationError("invalid weight or pin_position")
	}
	for _, v := range []float64{c.FocalX, c.FocalY} {
		if math.IsNaN(v) || math.IsInf(v, 0) || v < 0 || v > 1 {
			return common.NewValidationError("focal point must be within 0..1")
		}
	}
	if c.StartsAt.Valid && c.EndsAt.Valid && !c.StartsAt.Time.Before(c.EndsAt.Time) {
		return common.NewValidationError("starts_at must precede ends_at")
	}
	if c.LinkedGameID != nil && *c.LinkedGameID <= 0 {
		return common.NewValidationError("invalid linked_game_id")
	}
	primary, secondary := showcaseString(c.PrimaryActionType), showcaseString(c.SecondaryActionType)
	if primary != "" && !showcaseMember(primary, "game", "project", "product", "website") || secondary != "" && !showcaseMember(secondary, "steam", "kickstarter", "website", "other") {
		return common.NewValidationError("invalid fixed action type")
	}
	if primary == "game" && showcaseString(c.PrimaryTarget) != "" {
		return common.NewValidationError("game action cannot have external target")
	}
	for _, target := range []*string{c.PrimaryTarget, c.SecondaryTarget} {
		if v := showcaseString(target); v != "" && !showcaseURL(v) {
			return common.NewValidationError("external target must be an HTTPS URL without credentials")
		}
	}
	seen := map[string]bool{}
	for _, l := range locales {
		if !showcaseMember(l.Lang, "zh", "en") || seen[l.Lang] {
			return common.NewValidationError("locales must have unique zh/en entries")
		}
		seen[l.Lang] = true
		if utf8.RuneCountInString(l.Title) > 160 || utf8.RuneCountInString(l.Summary) > 400 || len(l.Tags) > 3 || utf8.RuneCountInString(showcaseString(l.EditorialNote)) > 200 {
			return common.NewValidationError("localized content exceeds limits")
		}
		if c.Sponsored && showcaseString(l.EditorialNote) != "" {
			return common.NewValidationError("Sponsored cannot contain Editorial Note")
		}
	}
	return nil
}

func showcaseReadiness(ctx context.Context, q *gamesqlc.Queries, c gamesqlc.GfgShowcaseCampaign, ls []showcaseLocale) ([]string, error) {
	issues := []string{}
	langs := []string{}
	if e := validateShowcaseShape(c, ls); e != nil {
		issues = append(issues, e.Error())
	}
	for _, l := range ls {
		if l.Enabled {
			langs = append(langs, l.Lang)
			if strings.TrimSpace(l.Title) == "" || strings.TrimSpace(l.Summary) == "" {
				issues = append(issues, l.Lang+": title and summary required")
			}
		}
	}
	if len(langs) == 0 {
		issues = append(issues, "at least one enabled locale required")
	}
	key := showcaseString(c.DesktopObjectKey)
	if !assets.ValidKey(key) || !strings.HasPrefix(key, fmt.Sprintf("game/showcase/%d/desktop/", c.ID)) {
		issues = append(issues, "valid desktop artwork required")
	}
	if key = showcaseString(c.MobileObjectKey); key != "" && (!assets.ValidKey(key) || !strings.HasPrefix(key, fmt.Sprintf("game/showcase/%d/mobile/", c.ID))) {
		issues = append(issues, "invalid mobile artwork")
	}
	if !c.StartsAt.Valid || !c.EndsAt.Valid {
		issues = append(issues, "complete schedule required")
	}
	p, s := showcaseString(c.PrimaryActionType), showcaseString(c.SecondaryActionType)
	if p == "" || p != "game" && !showcaseURL(showcaseString(c.PrimaryTarget)) || p == "game" && c.LinkedGameID == nil {
		issues = append(issues, "complete primary action required")
	}
	if s != "" && !showcaseURL(showcaseString(c.SecondaryTarget)) || s == "" && showcaseString(c.SecondaryTarget) != "" {
		issues = append(issues, "complete secondary action required")
	}
	if c.LinkedGameID != nil {
		g, e := q.ShowcaseGameStatus(ctx, *c.LinkedGameID)
		if e != nil {
			return nil, e
		}
		if !g.Sfw {
			issues = append(issues, "Adult linked Game cannot enter Showcase")
		}
	}
	if c.PinPosition != nil && c.StartsAt.Valid && c.EndsAt.Valid && len(langs) > 0 {
		conflict, e := q.ShowcasePinConflict(ctx, gamesqlc.ShowcasePinConflictParams{ID: c.ID, StartsAt: c.StartsAt, EndsAt: c.EndsAt, Langs: langs, PinPosition: *c.PinPosition, Sponsored: c.Sponsored, LinkedGameID: c.LinkedGameID})
		if e != nil {
			return nil, e
		}
		if conflict {
			issues = append(issues, "overlapping pin, Sponsored pin, or linked Game pin")
		}
	}
	return issues, nil
}

func showcaseDTO(c gamesqlc.GfgShowcaseCampaign, ls []showcaseLocale, issues []string, now time.Time) showcaseWorkspace {
	var linked *string
	if c.LinkedGameID != nil {
		s := strconv.FormatInt(*c.LinkedGameID, 10)
		linked = &s
	}
	return showcaseWorkspace{c, strconv.FormatInt(c.ID, 10), linked, ls, showcaseState(c, now), len(issues) == 0, issues}
}
func (api *GameAPI) GetShowcaseCampaign(c fiber.Ctx) error {
	id, e := adminutil.ParseIDParam(c)
	if e != nil {
		return common.NewResponse(c).Error(e)
	}
	row, err := api.store.q.GetShowcaseCampaign(c.Context(), id)
	if err != nil {
		return common.NewResponse(c).Error(gameDAOError(err))
	}
	ls, err := readShowcaseLocales(c.Context(), api.store.q, id)
	if err != nil {
		return common.NewResponse(c).Error(gameDAOError(err))
	}
	issues, err := showcaseReadiness(c.Context(), api.store.q, row, ls)
	if err != nil {
		return common.NewResponse(c).Error(gameDAOError(err))
	}
	return common.NewResponse(c).SuccessWithData(showcaseDTO(row, ls, issues, time.Now()))
}
func (api *GameAPI) CreateShowcaseCampaign(c fiber.Ctx) error {
	var req struct {
		InternalName string `json:"internal_name"`
		ContentType  string `json:"content_type"`
		Sponsored    bool   `json:"sponsored"`
		LinkedGameID *int64 `json:"linked_game_id"`
	}
	if e := adminutil.DecodeBody(c, &req); e != nil {
		return common.NewResponse(c).Error(e)
	}
	var row gamesqlc.GfgShowcaseCampaign
	shape := gamesqlc.GfgShowcaseCampaign{InternalName: strings.TrimSpace(req.InternalName), ContentType: req.ContentType, Sponsored: req.Sponsored, LinkedGameID: req.LinkedGameID, Weight: 100, FocalX: 0.5, FocalY: 0.5}
	if e := validateShowcaseShape(shape, nil); e != nil {
		return common.NewResponse(c).Error(gameDAOError(e))
	}
	err := api.store.mutate(c.Context(), audit.MetaFromFiber(c), "create", "gfg_showcase_campaign", func(q *gamesqlc.Queries) (int64, any, any, error) {
		var e error
		row, e = q.CreateShowcaseCampaign(c.Context(), gamesqlc.CreateShowcaseCampaignParams{InternalName: shape.InternalName, ContentType: req.ContentType, Sponsored: req.Sponsored, LinkedGameID: req.LinkedGameID})
		return row.ID, nil, row, e
	})
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	return common.NewResponse(c).SuccessWithData(showcaseDTO(row, []showcaseLocale{}, []string{"draft requires content, artwork and schedule"}, time.Now()))
}
func saveShowcase(ctx context.Context, q *gamesqlc.Queries, c gamesqlc.GfgShowcaseCampaign) (gamesqlc.GfgShowcaseCampaign, error) {
	return q.SaveShowcaseCampaign(ctx, gamesqlc.SaveShowcaseCampaignParams{ID: c.ID, InternalName: c.InternalName, ContentType: c.ContentType, Sponsored: c.Sponsored, LinkedGameID: c.LinkedGameID, State: c.State, StartsAt: c.StartsAt, EndsAt: c.EndsAt, Weight: c.Weight, PinPosition: c.PinPosition, DesktopObjectKey: c.DesktopObjectKey, MobileObjectKey: c.MobileObjectKey, FocalX: c.FocalX, FocalY: c.FocalY, PrimaryActionType: c.PrimaryActionType, PrimaryTarget: c.PrimaryTarget, SecondaryActionType: c.SecondaryActionType, SecondaryTarget: c.SecondaryTarget})
}

type showcaseChange func(*gamesqlc.GfgShowcaseCampaign, *[]showcaseLocale) error

func (api *GameAPI) changeShowcase(c fiber.Ctx, id int64, action string, change showcaseChange) common.Error {
	ctx := c.Context()
	return api.store.mutate(ctx, audit.MetaFromFiber(c), action, "gfg_showcase_campaign", func(q *gamesqlc.Queries) (int64, any, any, error) {
		row, e := q.LockShowcaseCampaign(ctx, id)
		if e != nil {
			return id, nil, nil, e
		}
		ls, e := readShowcaseLocales(ctx, q, id)
		if e != nil {
			return id, nil, nil, e
		}
		before := showcaseDTO(row, append([]showcaseLocale{}, ls...), nil, time.Now())
		if row.State == "archived" {
			return id, before, nil, common.NewValidationError("archived Campaign is immutable")
		}
		if e = change(&row, &ls); e != nil {
			return id, before, nil, e
		}
		if e = validateShowcaseShape(row, ls); e != nil {
			return id, before, nil, e
		}
		if row.State == "published" {
			issues, e := showcaseReadiness(ctx, q, row, ls)
			if e != nil {
				return id, before, nil, e
			}
			if len(issues) > 0 {
				return id, before, nil, common.NewValidationError(strings.Join(issues, "; "))
			}
		}
		row, e = saveShowcase(ctx, q, row)
		if e != nil {
			return id, before, nil, e
		}
		for _, l := range ls {
			tags := l.Tags
			if tags == nil {
				tags = []string{}
			}
			if e = q.SaveShowcaseLocale(ctx, gamesqlc.SaveShowcaseLocaleParams{CampaignID: id, Lang: l.Lang, Enabled: l.Enabled, Title: l.Title, Summary: l.Summary, Tags: tags, EditorialNote: l.EditorialNote}); e != nil {
				return id, before, nil, e
			}
		}
		return id, before, showcaseDTO(row, ls, nil, time.Now()), nil
	})
}
func (api *GameAPI) SaveShowcaseContent(c fiber.Ctx) error {
	id, e := adminutil.ParseIDParam(c)
	if e != nil {
		return common.NewResponse(c).Error(e)
	}
	var req showcaseContent
	if e = adminutil.DecodeBody(c, &req); e != nil {
		return common.NewResponse(c).Error(e)
	}
	err := api.changeShowcase(c, id, "content", func(row *gamesqlc.GfgShowcaseCampaign, ls *[]showcaseLocale) error {
		if req.InternalName != "" {
			row.InternalName = strings.TrimSpace(req.InternalName)
		}
		row.ContentType, row.Sponsored, row.LinkedGameID = req.ContentType, req.Sponsored, req.LinkedGameID
		row.FocalX, row.FocalY = req.FocalX, req.FocalY
		row.PrimaryActionType, row.PrimaryTarget = req.PrimaryActionType, req.PrimaryTarget
		row.SecondaryActionType, row.SecondaryTarget = req.SecondaryActionType, req.SecondaryTarget
		*ls = append([]showcaseLocale{}, req.Locales...)
		for _, lang := range []string{"zh", "en"} {
			found := false
			for _, l := range *ls {
				found = found || l.Lang == lang
			}
			if !found {
				*ls = append(*ls, showcaseLocale{Lang: lang, Tags: []string{}})
			}
		}
		return nil
	})
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	return api.GetShowcaseCampaign(c)
}
func (api *GameAPI) SaveShowcaseSchedule(c fiber.Ctx) error {
	id, e := adminutil.ParseIDParam(c)
	if e != nil {
		return common.NewResponse(c).Error(e)
	}
	var req showcaseSchedule
	if e = adminutil.DecodeBody(c, &req); e != nil {
		return common.NewResponse(c).Error(e)
	}
	err := api.changeShowcase(c, id, "schedule", func(row *gamesqlc.GfgShowcaseCampaign, _ *[]showcaseLocale) error {
		row.StartsAt, row.EndsAt = showcaseTime(req.StartsAt), showcaseTime(req.EndsAt)
		row.Weight, row.PinPosition = req.Weight, req.PinPosition
		return nil
	})
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	return api.GetShowcaseCampaign(c)
}
func (api *GameAPI) ShowcaseLifecycle(c fiber.Ctx) error {
	id, e := adminutil.ParseIDParam(c)
	if e != nil {
		return common.NewResponse(c).Error(e)
	}
	action := c.Path()[strings.LastIndex(c.Path(), "/")+1:]
	err := api.changeShowcase(c, id, action, func(row *gamesqlc.GfgShowcaseCampaign, _ *[]showcaseLocale) error {
		switch {
		case action == "publish" && row.State == "draft", action == "resume" && row.State == "paused":
			row.State = "published"
		case action == "pause" && row.State == "published":
			row.State = "paused"
		case action == "archive":
			row.State = "archived"
		default:
			return common.NewValidationError("invalid Campaign lifecycle transition")
		}
		return nil
	})
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	return api.GetShowcaseCampaign(c)
}
func (api *GameAPI) DeleteShowcaseDraft(c fiber.Ctx) error {
	id, e := adminutil.ParseIDParam(c)
	if e != nil {
		return common.NewResponse(c).Error(e)
	}
	err := api.store.mutate(c.Context(), audit.MetaFromFiber(c), "delete", "gfg_showcase_campaign", func(q *gamesqlc.Queries) (int64, any, any, error) {
		before, e := q.LockShowcaseCampaign(c.Context(), id)
		if e != nil {
			return id, nil, nil, e
		}
		if before.State != "draft" {
			return id, before, nil, common.NewValidationError("only never-published Drafts can be deleted")
		}
		_, e = q.DeleteShowcaseDraft(c.Context(), id)
		return id, before, nil, e
	})
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	return common.NewResponse(c).Success()
}
func (api *GameAPI) ShowcaseArtwork(c fiber.Ctx) error {
	id, e := adminutil.ParseIDParam(c)
	if e != nil {
		return common.NewResponse(c).Error(e)
	}
	variant := c.Params("variant")
	if !showcaseMember(variant, "desktop", "mobile") {
		return common.NewResponse(c).Error(common.NewValidationError("invalid artwork variant"))
	}
	if _, err := api.store.q.GetShowcaseCampaign(c.Context(), id); err != nil {
		return common.NewResponse(c).Error(gameDAOError(err))
	}
	var key *string
	var publication *assets.Publication
	if c.Method() == fiber.MethodPost {
		file, err := c.FormFile("file")
		if err != nil || file.Size > assets.MaxSize {
			return common.NewResponse(c).Error(common.NewValidationError("file required; max 5 MiB"))
		}
		reader, err := file.Open()
		if err != nil {
			return common.NewResponse(c).Error(common.NewValidationError("cannot read file"))
		}
		defer reader.Close()
		data, err := io.ReadAll(io.LimitReader(reader, assets.MaxSize+1))
		if err != nil {
			return common.NewResponse(c).Error(common.NewValidationError("cannot read file"))
		}
		object, err := assets.NewObject("showcase-"+variant, id, file.Filename, data)
		if err != nil {
			return common.NewResponse(c).Error(common.NewValidationError(err.Error()))
		}
		result, err := api.showcaseAssets.Publish(c.Context(), object)
		if err != nil {
			return common.NewResponse(c).Error(common.NewError(common.RETURN_FAILED, 502, "COS publication failed; no database reference changed"))
		}
		publication = &result
		key = &result.ObjectKey
	}
	err := api.changeShowcase(c, id, "artwork."+variant, func(row *gamesqlc.GfgShowcaseCampaign, _ *[]showcaseLocale) error {
		if variant == "desktop" {
			row.DesktopObjectKey = key
		} else {
			row.MobileObjectKey = key
		}
		return nil
	})
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	if publication != nil {
		return common.NewResponse(c).SuccessWithData(publication)
	}
	return common.NewResponse(c).Success()
}
