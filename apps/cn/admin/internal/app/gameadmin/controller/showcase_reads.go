package controller

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	_ "time/tzdata"

	"github.com/gofiber/fiber/v3"
	env "github.com/gofurry/gofurry-admin/config"
	gamesqlc "github.com/gofurry/gofurry-admin/internal/db/game/sqlc"
	"github.com/gofurry/gofurry-admin/pkg/common"
	"github.com/jackc/pgx/v5/pgtype"
)

func showcaseDates(c fiber.Ctx) (pgtype.Date, pgtype.Date, common.Error) {
	zone, _ := time.LoadLocation("Asia/Shanghai")
	today := time.Now().In(zone)
	from, e := time.Parse(time.DateOnly, c.Query("from", today.AddDate(0, 0, -29).Format(time.DateOnly)))
	to, te := time.Parse(time.DateOnly, c.Query("to", today.Format(time.DateOnly)))
	if e != nil || te != nil || from.After(to) || to.Sub(from) > 365*24*time.Hour {
		return pgtype.Date{}, pgtype.Date{}, common.NewValidationError("date range must be ordered YYYY-MM-DD, at most 366 days")
	}
	return pgtype.Date{Time: from, Valid: true}, pgtype.Date{Time: to, Valid: true}, nil
}

type showcaseTotals struct {
	ValidImpressions      int64   `json:"valid_impressions"`
	QualifiedClicks       int64   `json:"qualified_clicks"`
	CTR                   float64 `json:"ctr"`
	SessionEstimate       int64   `json:"session_estimate"`
	SessionEstimateMethod string  `json:"session_estimate_method"`
	ClickArtwork          int64   `json:"click_artwork"`
	ClickTitle            int64   `json:"click_title"`
	ClickPrimary          int64   `json:"click_primary"`
	ClickSecondary        int64   `json:"click_secondary"`
}

func (api *GameAPI) ShowcaseStats(c fiber.Ctx) error {
	id, e := strconv.ParseInt(strings.TrimSuffix(c.Params("id"), ".csv"), 10, 64)
	if e != nil || id <= 0 {
		return common.NewResponse(c).Error(common.NewValidationError("invalid campaign id"))
	}
	from, to, err := showcaseDates(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	rows, e := api.store.q.ListShowcaseCampaignStats(c.Context(), gamesqlc.ListShowcaseCampaignStatsParams{CampaignID: &id, FromDate: from, ToDate: to})
	if e != nil {
		return common.NewResponse(c).Error(gameDAOError(e))
	}
	if strings.HasSuffix(c.Path(), ".csv") {
		var data bytes.Buffer
		w := csv.NewWriter(&data)
		_ = w.Write([]string{"stat_date", "valid_impressions", "qualified_clicks", "ctr", "session_estimate_daily_hll", "click_artwork", "click_title", "click_primary", "click_secondary"})
		n := func(v int64) string { return strconv.FormatInt(v, 10) }
		for _, r := range rows {
			ctr := 0.0
			if r.ValidImpressions > 0 {
				ctr = float64(r.QualifiedClicks) / float64(r.ValidImpressions)
			}
			_ = w.Write([]string{r.StatDate.Time.Format(time.DateOnly), n(r.ValidImpressions), n(r.QualifiedClicks), strconv.FormatFloat(ctr, 'f', 6, 64), n(r.SessionEstimate), n(r.ClickArtwork), n(r.ClickTitle), n(r.ClickPrimary), n(r.ClickSecondary)})
		}
		w.Flush()
		c.Set("Content-Type", "text/csv; charset=utf-8")
		c.Set("Content-Disposition", `attachment; filename="showcase-`+n(id)+`.csv"`)
		return c.Send(data.Bytes())
	}
	totals := showcaseTotals{SessionEstimateMethod: "sum_of_daily_hll_estimates; a session may count on multiple days"}
	for _, r := range rows {
		totals.ValidImpressions += r.ValidImpressions
		totals.QualifiedClicks += r.QualifiedClicks
		totals.SessionEstimate += r.SessionEstimate
		totals.ClickArtwork += r.ClickArtwork
		totals.ClickTitle += r.ClickTitle
		totals.ClickPrimary += r.ClickPrimary
		totals.ClickSecondary += r.ClickSecondary
	}
	if totals.ValidImpressions > 0 {
		totals.CTR = float64(totals.QualifiedClicks) / float64(totals.ValidImpressions)
	}
	return common.NewResponse(c).SuccessWithData(struct {
		Totals   showcaseTotals                  `json:"totals"`
		Daily    []gamesqlc.GfgShowcaseDailyStat `json:"daily"`
		Timezone string                          `json:"timezone"`
	}{totals, rows, "Asia/Shanghai"})
}
func (api *GameAPI) ShowcaseQuality(c fiber.Ctx) error {
	from, to, e := showcaseDates(c)
	if e != nil {
		return common.NewResponse(c).Error(e)
	}
	rows, err := api.store.q.ListShowcaseQuality(c.Context(), gamesqlc.ListShowcaseQualityParams{FromDate: from, ToDate: to})
	if err != nil {
		return common.NewResponse(c).Error(gameDAOError(err))
	}
	return common.NewResponse(c).SuccessWithData(struct {
		Daily    []gamesqlc.GfgShowcaseAnalyticsDailyQuality `json:"daily"`
		Timezone string                                      `json:"timezone"`
	}{rows, "Asia/Shanghai"})
}
func (api *GameAPI) ShowcaseProxy(c fiber.Ctx) error {
	cfg := env.GetServerConfig().ExternalServices.GameBackend
	operation := "composition"
	if strings.HasSuffix(c.Path(), "/candidates") {
		operation = "candidates"
	}
	base, e := url.Parse(cfg.BaseURL)
	if e != nil || base.Host == "" || (base.Scheme != "http" && base.Scheme != "https") || cfg.AdminToken == "" {
		return common.NewResponse(c).Error(common.NewError(common.RETURN_FAILED, 503, "Game Backend connection unavailable"))
	}
	base.Path = strings.TrimSuffix(base.Path, "/") + "/api/v2/game/internal/showcase/" + operation
	base.RawQuery = ""
	query := url.Values{}
	for _, key := range []string{"lang", "region", "pool", "page_num", "page_size"} {
		if v := c.Query(key); v != "" {
			query.Set(key, v)
		}
	}
	base.RawQuery = query.Encode()
	timeout := time.Duration(cfg.TimeoutSeconds) * time.Second
	if timeout <= 0 || timeout > 15*time.Second {
		timeout = 10 * time.Second
	}
	ctx, cancel := context.WithTimeout(c.Context(), timeout)
	defer cancel()
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, base.String(), nil)
	if e != nil {
		return common.NewResponse(c).Error(common.NewServiceError("invalid Game Backend request"))
	}
	header := cfg.AdminTokenHeader
	if header == "" {
		header = "X-GoFurry-Admin-Token"
	}
	req.Header.Set(header, cfg.AdminToken)
	client := &http.Client{Timeout: timeout, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	response, e := client.Do(req)
	if e != nil {
		return common.NewResponse(c).Error(common.NewError(common.RETURN_FAILED, 502, "Game Backend unavailable"))
	}
	defer response.Body.Close()
	body, e := io.ReadAll(io.LimitReader(response.Body, 2<<20+1))
	if e != nil || len(body) > 2<<20 {
		return common.NewResponse(c).Error(common.NewError(common.RETURN_FAILED, 502, "invalid Game Backend response"))
	}
	var envelope struct {
		Code int             `json:"code"`
		Data json.RawMessage `json:"data"`
	}
	if response.StatusCode != http.StatusOK || json.Unmarshal(body, &envelope) != nil || envelope.Code != common.RETURN_SUCCESS || len(envelope.Data) == 0 {
		return common.NewResponse(c).Error(common.NewError(common.RETURN_FAILED, 502, "Game Backend Showcase request failed"))
	}
	return common.NewResponse(c).SuccessWithData(envelope.Data)
}
