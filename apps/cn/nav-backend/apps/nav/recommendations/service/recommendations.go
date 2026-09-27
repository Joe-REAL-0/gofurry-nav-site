package service

import (
	"bytes"
	"crypto/sha256"
	"sort"
	"strconv"
	"time"

	home "github.com/gofurry/gofurry-nav-backend/apps/nav/home/service"
	navmodels "github.com/gofurry/gofurry-nav-backend/apps/nav/navPage/models"
	"github.com/gofurry/gofurry-nav-backend/apps/nav/recommendations/models"
	"github.com/gofurry/gofurry-nav-backend/common"
)

type siteGroupReader interface {
	GetSiteList(string) ([]navmodels.SiteVo, common.GFError)
	GetGroupList(string) ([]navmodels.GroupVo, common.GFError)
}

type Service struct {
	reader siteGroupReader
	now    func() time.Time
}

func New(reader siteGroupReader) *Service { return &Service{reader: reader, now: time.Now} }

// Discovery uses existing localized read models; no recommendation storage or scoring.
func (svc *Service) GetRecommendations(siteID int64, lang string, limit int) models.Response {
	now := svc.now().UTC()
	response := models.Response{SchemaVersion: 1, GeneratedAt: now, State: "unavailable", SiteID: siteID, Items: []navmodels.SiteVo{}}
	if lang != "en" {
		lang = "zh"
	}
	sites, err := svc.reader.GetSiteList(lang)
	if err != nil {
		return response
	}
	groups, err := svc.reader.GetGroupList(lang)
	if err != nil {
		return response
	}
	id := strconv.FormatInt(siteID, 10)
	// Membership uses all mappings, including Sites below the Home preview boundary.
	matched := make([]navmodels.GroupVo, 0)
	for _, group := range groups {
		for _, member := range group.Sites {
			if member == id {
				matched = append(matched, group)
				break
			}
		}
	}
	candidates := make(map[string]navmodels.SiteVo)
	// Home owns weight/update/id ordering and its Top-8 cutoff. Shuffle only after it.
	for _, group := range home.BuildHomeGroupsForCache(sites, matched) {
		for _, site := range group.Sites {
			if site.ID != id {
				candidates[site.ID] = site
			}
		}
	}
	type candidate struct {
		site navmodels.SiteVo
		hash [sha256.Size]byte
	}
	ordered := make([]candidate, 0, len(candidates))
	for _, site := range candidates {
		ordered = append(ordered, candidate{site, sha256.Sum256([]byte(id + "|" + site.ID + "|" + now.Format(time.DateOnly)))})
	}
	sort.Slice(ordered, func(i, j int) bool {
		if order := bytes.Compare(ordered[i].hash[:], ordered[j].hash[:]); order != 0 {
			return order < 0
		}
		return ordered[i].site.ID < ordered[j].site.ID
	})
	if limit <= 0 || limit > 8 {
		limit = 8
	}
	if limit > len(ordered) {
		limit = len(ordered)
	}
	for _, item := range ordered[:limit] {
		response.Items = append(response.Items, item.site)
	}
	response.State = "ready"
	return response
}
