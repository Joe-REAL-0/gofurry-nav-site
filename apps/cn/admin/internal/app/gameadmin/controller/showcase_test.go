package controller

import (
	"context"
	gamesqlc "github.com/gofurry/gofurry-admin/internal/db/game/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
	"strings"
	"testing"
	"time"
)

func TestShowcasePublishReadiness(t *testing.T) {
	website, target := "website", "https://example.test/product"
	key := "game/showcase/1/desktop/" + strings.Repeat("a", 32) + ".avif"
	now := time.Now()
	row := gamesqlc.GfgShowcaseCampaign{ID: 1, InternalName: "Campaign", ContentType: "tabletop", Weight: 100, FocalX: 0.5, FocalY: 0.5, PrimaryActionType: &website, PrimaryTarget: &target, DesktopObjectKey: &key, StartsAt: pgtype.Timestamptz{Time: now, Valid: true}, EndsAt: pgtype.Timestamptz{Time: now.Add(time.Hour), Valid: true}}
	locales := []showcaseLocale{{Lang: "zh", Enabled: true, Title: "Title", Summary: "Summary", Tags: []string{"one"}}}
	if issues, e := showcaseReadiness(context.Background(), nil, row, locales); e != nil || len(issues) != 0 {
		t.Fatal(issues, e)
	}
	for _, tc := range []struct {
		name   string
		change func(*gamesqlc.GfgShowcaseCampaign, *[]showcaseLocale)
	}{
		{"locale", func(_ *gamesqlc.GfgShowcaseCampaign, ls *[]showcaseLocale) { (*ls)[0].Enabled = false }},
		{"missing_en", func(_ *gamesqlc.GfgShowcaseCampaign, ls *[]showcaseLocale) {
			*ls = append(*ls, showcaseLocale{Lang: "en", Enabled: true})
		}},
		{"art", func(c *gamesqlc.GfgShowcaseCampaign, _ *[]showcaseLocale) { c.DesktopObjectKey = nil }},
		{"wrong_art_owner", func(c *gamesqlc.GfgShowcaseCampaign, _ *[]showcaseLocale) {
			s := "game/showcase/2/desktop/" + strings.Repeat("a", 32) + ".avif"
			c.DesktopObjectKey = &s
		}},
		{"schedule", func(c *gamesqlc.GfgShowcaseCampaign, _ *[]showcaseLocale) { c.EndsAt = c.StartsAt }},
		{"action", func(c *gamesqlc.GfgShowcaseCampaign, _ *[]showcaseLocale) {
			s := "custom_label"
			c.PrimaryActionType = &s
		}},
		{"internal_game", func(c *gamesqlc.GfgShowcaseCampaign, _ *[]showcaseLocale) {
			s := "game"
			c.PrimaryActionType = &s
			c.PrimaryTarget = nil
		}},
		{"sponsored_note", func(c *gamesqlc.GfgShowcaseCampaign, ls *[]showcaseLocale) {
			c.Sponsored = true
			s := "Note"
			(*ls)[0].EditorialNote = &s
		}},
		{"http", func(c *gamesqlc.GfgShowcaseCampaign, _ *[]showcaseLocale) {
			s := "http://example.test"
			c.PrimaryTarget = &s
		}},
		{"credentials", func(c *gamesqlc.GfgShowcaseCampaign, _ *[]showcaseLocale) {
			s := "https://user:password@example.test"
			c.PrimaryTarget = &s
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			copy := row
			ls := append([]showcaseLocale{}, locales...)
			tc.change(&copy, &ls)
			if issues, e := showcaseReadiness(context.Background(), nil, copy, ls); e == nil && len(issues) == 0 {
				t.Fatal("invalid public state accepted")
			}
		})
	}
	if got := showcaseState(gamesqlc.GfgShowcaseCampaign{State: "published", StartsAt: row.StartsAt, EndsAt: row.EndsAt}, now.Add(-time.Second)); got != "scheduled" {
		t.Fatal(got)
	}
	if got := showcaseState(gamesqlc.GfgShowcaseCampaign{State: "published", StartsAt: row.StartsAt, EndsAt: row.EndsAt}, row.EndsAt.Time); got != "ended" {
		t.Fatal(got)
	}
}
