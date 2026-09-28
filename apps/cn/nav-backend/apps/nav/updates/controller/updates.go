package controller

import (
	"errors"
	"net/http"
	"strconv"
	"sync"

	"github.com/gofiber/fiber/v3"
	"github.com/gofurry/gofurry-nav-backend/apps/nav/updates/models"
	"github.com/gofurry/gofurry-nav-backend/apps/nav/updates/service"
	"github.com/gofurry/gofurry-nav-backend/common"
)

type updatesApi struct{ reader updatesReader }

var UpdatesApi *updatesApi

func init() {
	UpdatesApi = &updatesApi{}
}

func New(reader updatesReader) *updatesApi { return &updatesApi{reader: reader} }

type updatesReader interface {
	GetUpdates(lang string) models.UpdatesResponse
	GetUpdateDetail(id int64, lang string) (models.UpdateDetailResponse, error)
}

var (
	updatesReaderMu      sync.RWMutex
	updatesReaderForTest updatesReader
)

func (api updatesApi) GetUpdates(c fiber.Ctx) error {
	reader := api.reader
	if reader == nil {
		reader = currentUpdatesReader()
	}
	data := reader.GetUpdates(c.Query("lang", "zh"))
	return common.NewResponse(c).SuccessWithData(data)
}

func currentUpdatesReader() updatesReader {
	updatesReaderMu.RLock()
	reader := updatesReaderForTest
	updatesReaderMu.RUnlock()
	if reader != nil {
		return reader
	}
	return service.GetUpdatesService()
}

func setUpdatesReaderForTest(reader updatesReader) func() {
	updatesReaderMu.Lock()
	previous := updatesReaderForTest
	updatesReaderForTest = reader
	updatesReaderMu.Unlock()
	return func() {
		updatesReaderMu.Lock()
		updatesReaderForTest = previous
		updatesReaderMu.Unlock()
	}
}

func (api updatesApi) GetUpdateDetail(c fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		return common.NewResponse(c).ErrorWithCode(service.ErrNotFound.Error(), http.StatusNotFound)
	}
	reader := api.reader
	if reader == nil {
		reader = currentUpdatesReader()
	}
	data, err := reader.GetUpdateDetail(id, c.Query("lang", "zh"))
	if errors.Is(err, service.ErrNotFound) {
		return common.NewResponse(c).ErrorWithCode(service.ErrNotFound.Error(), http.StatusNotFound)
	}
	if err != nil {
		return common.NewResponse(c).ErrorWithCode("release notes unavailable", http.StatusServiceUnavailable)
	}
	return common.NewResponse(c).SuccessWithData(data)
}
