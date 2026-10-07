package controller

import (
	"bytes"
	"encoding/json"
	"io"
	"regexp"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/gofiber/fiber/v3"
	"github.com/gofurry/gofurry-admin/internal/app/gameadmin/models"
	"github.com/gofurry/gofurry-admin/internal/app/shared/adminutil"
	"github.com/gofurry/gofurry-admin/internal/app/shared/audit"
	"github.com/gofurry/gofurry-admin/pkg/common"
)

var collectionCodePattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

func decodeCollectionBody(c fiber.Ctx, out any) common.Error {
	d := json.NewDecoder(bytes.NewReader(c.Body()))
	d.DisallowUnknownFields()
	if err := d.Decode(out); err != nil {
		return common.NewValidationError("请求字段或 JSON 格式不正确；已有 Code 不可更改")
	}
	var extra any
	if err := d.Decode(&extra); err != io.EOF {
		return common.NewValidationError("请求须为单个 JSON 对象")
	}
	return nil
}
func normalizeCollectionContent(c *models.CollectionContent) common.Error {
	c.Name = strings.TrimSpace(c.Name)
	c.NameEn = strings.TrimSpace(c.NameEn)
	c.Info = strings.TrimSpace(c.Info)
	c.InfoEn = strings.TrimSpace(c.InfoEn)
	if c.Name == "" || c.NameEn == "" || utf8.RuneCountInString(c.Name) > 160 || utf8.RuneCountInString(c.NameEn) > 160 {
		return common.NewValidationError("中英文名称均须为 1–160 个字符")
	}
	if utf8.RuneCountInString(c.Info) > 500 || utf8.RuneCountInString(c.InfoEn) > 500 {
		return common.NewValidationError("中英文简介分别最多 500 个字符")
	}
	return nil
}
func canonicalCollectionIDs(ids []int64) ([]int64, common.Error) {
	canonical := append([]int64{}, ids...)
	for _, id := range canonical {
		if id <= 0 {
			return nil, common.NewValidationError("游戏 ID 必须为正整数")
		}
	}
	slices.Sort(canonical)
	return slices.Compact(canonical), nil
}
func validateCollectionHome(in models.ReplaceCollectionHome) common.Error {
	if in.Revision == "" || len(in.Slots) != 5 {
		return common.NewValidationError("须携带 revision 并完整提交第 1–5 位")
	}
	seen := map[int16]bool{}
	ids := map[int64]bool{}
	for _, s := range in.Slots {
		if s.Slot < 1 || s.Slot > 5 || seen[s.Slot] {
			return common.NewValidationError("第 1–5 位须各出现一次")
		}
		seen[s.Slot] = true
		if s.CollectionID != nil {
			if *s.CollectionID <= 0 || ids[*s.CollectionID] {
				return common.NewValidationError("所选分区 ID 须为正整数且不可重复")
			}
			ids[*s.CollectionID] = true
		}
	}
	return nil
}
func (api *GameAPI) ListGameCollections(c fiber.Ctx) error {
	page := adminutil.ParsePageQuery(c)
	if c.Query("page_size") == "" {
		page.PageSize = 50
	}
	status := c.Query("status")
	if status != "" && status != "draft" && status != "published" && status != "archived" {
		return common.NewResponse(c).Error(common.NewValidationError("游戏分区状态无效"))
	}
	eligible := c.Query("home_eligible")
	if eligible != "" && eligible != "true" && eligible != "false" {
		return common.NewResponse(c).Error(common.NewValidationError("home_eligible 须为 true 或 false"))
	}
	total, items, err := api.store.listCollections(c.Context(), page, status, eligible == "true")
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	return common.NewResponse(c).SuccessWithData(adminutil.BuildPageResponse(total, items))
}
func (api *GameAPI) CreateGameCollection(c fiber.Ctx) error {
	var in models.CreateCollection
	if err := decodeCollectionBody(c, &in); err != nil {
		return common.NewResponse(c).Error(err)
	}
	if len(in.Code) < 1 || len(in.Code) > 64 || !collectionCodePattern.MatchString(in.Code) {
		return common.NewResponse(c).Error(common.NewValidationError("Code 须为 1–64 位小写字母、数字及单个连字符"))
	}
	if err := normalizeCollectionContent(&in.CollectionContent); err != nil {
		return common.NewResponse(c).Error(err)
	}
	result, err := api.store.createCollection(c.Context(), audit.MetaFromFiber(c), in)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	return common.NewResponse(c).SuccessWithData(result)
}
func (api *GameAPI) GetGameCollection(c fiber.Ctx) error {
	id, err := adminutil.ParseIDParam(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	result, e := readCollection(c.Context(), api.store.q, id)
	if e != nil {
		return common.NewResponse(c).Error(collectionError(e))
	}
	return common.NewResponse(c).SuccessWithData(result)
}
func (api *GameAPI) UpdateGameCollection(c fiber.Ctx) error {
	id, err := adminutil.ParseIDParam(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	var in models.UpdateCollection
	if err = decodeCollectionBody(c, &in); err != nil {
		return common.NewResponse(c).Error(err)
	}
	if err = normalizeCollectionContent(&in.CollectionContent); err != nil {
		return common.NewResponse(c).Error(err)
	}
	result, err := api.store.updateCollection(c.Context(), audit.MetaFromFiber(c), id, in)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	return common.NewResponse(c).SuccessWithData(result)
}
func (api *GameAPI) GetGameCollectionMembers(c fiber.Ctx) error {
	id, err := adminutil.ParseIDParam(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	result, err := api.store.collectionMembers(c.Context(), id)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	return common.NewResponse(c).SuccessWithData(result)
}
func (api *GameAPI) ReplaceGameCollectionMembers(c fiber.Ctx) error {
	id, err := adminutil.ParseIDParam(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	var in models.ReplaceCollectionMembers
	if err = decodeCollectionBody(c, &in); err != nil {
		return common.NewResponse(c).Error(err)
	}
	if in.GameIDs == nil {
		return common.NewResponse(c).Error(common.NewValidationError("game_ids 须为完整数组"))
	}
	in.GameIDs, err = canonicalCollectionIDs(in.GameIDs)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	result, err := api.store.replaceCollectionMembers(c.Context(), audit.MetaFromFiber(c), id, in)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	return common.NewResponse(c).SuccessWithData(result)
}
func (api *GameAPI) TransitionGameCollection(action string) fiber.Handler {
	return func(c fiber.Ctx) error {
		id, err := adminutil.ParseIDParam(c)
		if err != nil {
			return common.NewResponse(c).Error(err)
		}
		var in models.CollectionVersion
		if err = decodeCollectionBody(c, &in); err != nil {
			return common.NewResponse(c).Error(err)
		}
		result, err := api.store.transitionCollection(c.Context(), audit.MetaFromFiber(c), id, in.Version, action)
		if err != nil {
			return common.NewResponse(c).Error(err)
		}
		return common.NewResponse(c).SuccessWithData(result)
	}
}
func (api *GameAPI) GetGameCollectionHome(c fiber.Ctx) error {
	result, err := api.store.collectionHome(c.Context())
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	return common.NewResponse(c).SuccessWithData(result)
}
func (api *GameAPI) ReplaceGameCollectionHome(c fiber.Ctx) error {
	var in models.ReplaceCollectionHome
	if err := decodeCollectionBody(c, &in); err != nil {
		return common.NewResponse(c).Error(err)
	}
	if err := validateCollectionHome(in); err != nil {
		return common.NewResponse(c).Error(err)
	}
	// An omitted ID is not an explicit empty placement in a full replacement.
	var raw struct {
		Slots []map[string]json.RawMessage `json:"slots"`
	}
	if err := json.Unmarshal(c.Body(), &raw); err != nil {
		return common.NewResponse(c).Error(common.NewValidationError("首页入口格式不正确"))
	}
	for _, slot := range raw.Slots {
		if _, present := slot["collection_id"]; !present {
			return common.NewResponse(c).Error(common.NewValidationError("每个位置须显式提供 collection_id，空位使用 null"))
		}
	}
	result, err := api.store.replaceCollectionHome(c.Context(), audit.MetaFromFiber(c), in)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	return common.NewResponse(c).SuccessWithData(result)
}
