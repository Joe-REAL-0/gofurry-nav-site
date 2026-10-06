package controller

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v3"
	v2models "github.com/gofurry/gofurry-game-backend/apps/game/v2/models"
	v2service "github.com/gofurry/gofurry-game-backend/apps/game/v2/service"
	"github.com/gofurry/gofurry-game-backend/common"
)

func (api *GameV2API) WithCollections(service *v2service.CollectionService) *GameV2API {
	api.collections = service
	return api
}

func collectionQuery(c fiber.Ctx) v2models.CollectionQuery {
	parse := func(value string) int64 {
		n, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return 0
		}
		return n
	}
	page := parse(c.Query("page"))
	size := parse(c.Query("page_size"))
	return v2models.CollectionQuery{Lang: c.Query("lang"), Mode: c.Query("mode"), Page: page, PageSize: size}
}

func collectionResponse(c fiber.Ctx, value any, err error) error {
	c.Set("Cache-Control", "no-store")
	if errors.Is(err, v2service.ErrCollectionNotFound) {
		return common.NewResponse(c).ErrorWithCode("Collection not found", fiber.StatusNotFound)
	}
	if err != nil {
		return common.NewResponse(c).ErrorWithCode("Collections unavailable", fiber.StatusServiceUnavailable)
	}
	return common.NewResponse(c).SuccessWithData(value)
}

func (api *GameV2API) GetCollectionHome(c fiber.Ctx) error {
	if api.collections == nil {
		return collectionResponse(c, nil, errors.New("uninitialized"))
	}
	value, err := api.collections.Home(c.Context(), collectionQuery(c))
	return collectionResponse(c, value, err)
}

func (api *GameV2API) GetCollections(c fiber.Ctx) error {
	if api.collections == nil {
		return collectionResponse(c, nil, errors.New("uninitialized"))
	}
	value, err := api.collections.List(c.Context(), collectionQuery(c))
	return collectionResponse(c, value, err)
}

func (api *GameV2API) GetCollection(c fiber.Ctx) error {
	if api.collections == nil {
		return collectionResponse(c, nil, errors.New("uninitialized"))
	}
	value, err := api.collections.Detail(c.Context(), c.Params("code"), collectionQuery(c))
	return collectionResponse(c, value, err)
}
