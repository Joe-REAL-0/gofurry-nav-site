package controller

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofurry/gofurry-admin/internal/app/shared/adminutil"
	gamesqlc "github.com/gofurry/gofurry-admin/internal/db/game/sqlc"
	"github.com/gofurry/gofurry-admin/pkg/common"
	"github.com/jackc/pgx/v5/pgtype"
	"strconv"
	"strings"
	"time"
)

func (api *GameAPI) ListShowcaseCampaigns(c fiber.Ctx) error {
	page, e := strconv.Atoi(c.Query("page_num", "1"))
	size, se := strconv.Atoi(c.Query("page_size", "20"))
	state, status := c.Query("state"), c.Query("derived_status")
	if e != nil || se != nil || page < 1 || page > 1000000 || size < 1 || size > 100 || !showcaseMember(state, "", "draft", "published", "paused", "archived") || !showcaseMember(status, "", "draft", "scheduled", "active", "ended", "paused", "archived") {
		return common.NewResponse(c).Error(common.NewValidationError("invalid pagination or Campaign state"))
	}
	var sponsored *bool
	if v := c.Query("sponsored"); v != "" {
		b, e := strconv.ParseBool(v)
		if e != nil {
			return common.NewResponse(c).Error(common.NewValidationError("invalid sponsored filter"))
		}
		sponsored = &b
	}
	now := time.Now()
	at := pgtype.Timestamptz{Time: now, Valid: true}
	keyword := strings.TrimSpace(c.Query("keyword"))
	if len(keyword) > 160 {
		return common.NewResponse(c).Error(common.NewValidationError("keyword too long"))
	}
	total, e := api.store.q.CountShowcaseCampaigns(c.Context(), gamesqlc.CountShowcaseCampaignsParams{State: state, Sponsored: sponsored, Keyword: keyword, DerivedStatus: status, AtTime: at})
	if e != nil {
		return common.NewResponse(c).Error(gameDAOError(e))
	}
	rows, e := api.store.q.ListShowcaseCampaigns(c.Context(), gamesqlc.ListShowcaseCampaignsParams{State: state, Sponsored: sponsored, Keyword: keyword, DerivedStatus: status, AtTime: at, RowLimit: int32(size), RowOffset: int32((page - 1) * size)})
	if e != nil {
		return common.NewResponse(c).Error(gameDAOError(e))
	}
	type item struct {
		gamesqlc.ListShowcaseCampaignsRow
		ID            string  `json:"id"`
		LinkedGameID  *string `json:"linked_game_id"`
		DerivedStatus string  `json:"derived_status"`
	}
	items := make([]item, 0, len(rows))
	for _, r := range rows {
		var id *string
		if r.LinkedGameID != nil {
			s := strconv.FormatInt(*r.LinkedGameID, 10)
			id = &s
		}
		items = append(items, item{r, strconv.FormatInt(r.ID, 10), id, showcaseState(gamesqlc.GfgShowcaseCampaign{State: r.State, StartsAt: r.StartsAt, EndsAt: r.EndsAt}, now)})
	}
	return common.NewResponse(c).SuccessWithData(adminutil.BuildPageResponse(total, items))
}
