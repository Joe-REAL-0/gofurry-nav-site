package controller

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/gofurry/gofurry-nav-backend/apps/nav/recommendations/models"
	"github.com/gofurry/gofurry-nav-backend/common"
)

type reader interface {
	GetRecommendations(int64, string, int) models.Response
}
type API struct{ reader reader }

func New(service reader) *API { return &API{reader: service} }
func (api *API) GetRecommendations(c fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("siteId"), 10, 64)
	if err != nil || id <= 0 {
		return common.NewResponse(c).ErrorWithCode("siteId 参数非法", fiber.StatusBadRequest)
	}
	lang := c.Query("lang", "zh")
	if lang != "zh" && lang != "en" {
		return common.NewResponse(c).ErrorWithCode("lang 参数非法", fiber.StatusBadRequest)
	}
	limit, err := strconv.Atoi(c.Query("limit", "8"))
	if err != nil {
		return common.NewResponse(c).ErrorWithCode("limit 参数非法", fiber.StatusBadRequest)
	}
	return common.NewResponse(c).SuccessWithData(api.reader.GetRecommendations(id, lang, limit))
}
