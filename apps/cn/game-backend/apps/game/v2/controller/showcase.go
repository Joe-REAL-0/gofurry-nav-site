package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/gofiber/fiber/v3"
	"github.com/gofurry/gofurry-game-backend/apps/game/v2/showcase"
	"github.com/gofurry/gofurry-game-backend/common"
	"github.com/gofurry/gofurry-game-backend/common/util"
	"io"
	"strconv"
	"time"
)

func (api *GameV2API) WithShowcase(s *showcase.Service, a *showcase.Analytics, signer showcase.Signer) *GameV2API {
	api.showcaseService = s
	api.showcaseAnalytics = a
	api.showcaseSigner = signer
	return api
}
func (api *GameV2API) GetShowcase(c fiber.Ctx) error {
	lang, region := c.Query("lang", "zh"), c.Query("region", "CN")
	if !showcase.ValidateScope(lang, region) {
		return common.NewResponse(c).ErrorWithCode("invalid locale or region", 400)
	}
	if api.showcaseService == nil {
		return common.NewResponse(c).ErrorWithCode("Showcase unavailable", 503)
	}
	snap, e := api.showcaseService.Snapshot(c.Context(), lang, region)
	if e == nil {
		snap, e = api.showcaseSigner.SignSnapshot(snap, lang, region)
	}
	if e != nil {
		return common.NewResponse(c).ErrorWithCode("Showcase unavailable", 503)
	}
	c.Set("Cache-Control", "no-store")
	return common.NewResponse(c).SuccessWithData(snap)
}
func (api *GameV2API) GetShowcaseCandidates(c fiber.Ctx) error {
	lang, region, pool := c.Query("lang", "zh"), c.Query("region", "CN"), c.Query("pool")
	page, e := strconv.Atoi(c.Query("page_num", "1"))
	size, se := strconv.Atoi(c.Query("page_size", "20"))
	if !showcase.ValidateScope(lang, region) || (pool != "upcoming" && pool != "new_release" && pool != "trending") || e != nil || se != nil || page < 1 || page > 1000000 || size < 1 || size > 100 {
		return common.NewResponse(c).ErrorWithCode("invalid candidate query", 400)
	}
	if api.showcaseService == nil {
		return common.NewResponse(c).ErrorWithCode("Showcase unavailable", 503)
	}
	ctx, cancel := context.WithTimeout(c.Context(), 8*time.Second)
	defer cancel()
	input, e := api.showcaseService.Reader.Read(ctx, api.showcaseService.Now(), lang, region)
	if e != nil {
		return common.NewResponse(c).ErrorWithCode("Showcase unavailable", 503)
	}
	items := input.Diagnostics[pool]
	total := len(items)
	start := (page - 1) * size
	if start > total {
		start = total
	}
	end := min(total, start+size)
	return common.NewResponse(c).SuccessWithData(struct {
		Total int                   `json:"total"`
		Items []showcase.Diagnostic `json:"items"`
	}{total, append([]showcase.Diagnostic{}, items[start:end]...)})
}
func (api *GameV2API) ShowcaseEvent(c fiber.Ctx) error {
	var event showcase.Event
	malformed := func() error {
		if api.showcaseAnalytics != nil {
			api.showcaseAnalytics.Malformed(c.Context())
		}
		return common.NewResponse(c).ErrorWithCode("invalid Showcase event", 400)
	}
	if len(c.Body()) > 8192 {
		return malformed()
	}
	decoder := json.NewDecoder(bytes.NewReader(c.Body()))
	decoder.DisallowUnknownFields()
	if e := decoder.Decode(&event); e != nil || decoder.Decode(new(any)) != io.EOF || !showcase.ValidEventShape(event) {
		return malformed()
	}
	if api.showcaseAnalytics != nil {
		api.showcaseAnalytics.Submit(c.Context(), event, c.Get("Origin"), c.Get("User-Agent"), util.GetClientIP(c))
	}
	return c.SendStatus(fiber.StatusNoContent)
}
