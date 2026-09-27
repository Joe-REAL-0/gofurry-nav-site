package service

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	navmodels "github.com/gofurry/gofurry-nav-backend/apps/nav/navPage/models"
	"github.com/gofurry/gofurry-nav-backend/common"
)

type fakeReader struct {
	sites   []navmodels.SiteVo
	groups  []navmodels.GroupVo
	langs   []string
	failure string
}

func (f *fakeReader) GetSiteList(lang string) ([]navmodels.SiteVo, common.GFError) {
	f.langs = append(f.langs, lang)
	if f.failure == "sites" {
		return nil, common.NewServiceError("cache/read model unavailable")
	}
	sites := append([]navmodels.SiteVo(nil), f.sites...)
	for i := range sites {
		sites[i].Name = lang + "-" + sites[i].ID
		sites[i].Info = lang + " info"
	}
	return sites, nil
}
func (f *fakeReader) GetGroupList(lang string) ([]navmodels.GroupVo, common.GFError) {
	f.langs = append(f.langs, lang)
	if f.failure == "groups" {
		return nil, common.NewServiceError("cache/read model unavailable")
	}
	return f.groups, nil
}
func fixture(n int) (*Service, *fakeReader) {
	reader := &fakeReader{}
	for i := 1; i <= n; i++ {
		reader.sites = append(reader.sites, navmodels.SiteVo{ID: fmt.Sprint(i), Domain: fmt.Sprintf("site%d.example", i), Nsfw: "0", Welfare: "0", ViewCount: int64(i * 100), UpdateTime: "2026-09-20"})
	}
	svc := New(reader)
	svc.now = func() time.Time { return time.Date(2026, 9, 27, 12, 0, 0, 0, time.UTC) }
	return svc, reader
}
func ids(sites []navmodels.SiteVo) []string {
	result := []string{}
	for _, site := range sites {
		result = append(result, site.ID)
	}
	return result
}

func TestRecommendationsMultipleGroupsDedupeSelfAndLocalization(t *testing.T) {
	svc, reader := fixture(6)
	reader.groups = []navmodels.GroupVo{{ID: "a", Sites: []string{"1", "2", "3"}}, {ID: "b", Sites: []string{"1", "3", "4"}}, {ID: "unrelated", Sites: []string{"5", "6"}}}
	for _, lang := range []string{"zh", "en"} {
		response := svc.GetRecommendations(1, lang, 8)
		if response.State != "ready" || len(response.Items) != 3 || response.SchemaVersion != 1 || response.SiteID != 1 {
			t.Fatalf("response=%+v", response)
		}
		seen := map[string]bool{}
		for _, site := range response.Items {
			if seen[site.ID] || site.ID == "1" || site.ID == "5" || site.ID == "6" {
				t.Fatalf("invalid candidate %+v", site)
			}
			seen[site.ID] = true
			if site.Name != lang+"-"+site.ID || site.Info != lang+" info" || site.Domain == "" || site.ViewCount == 0 {
				t.Fatalf("lost localized snapshot: %+v", site)
			}
		}
	}
	if !reflect.DeepEqual(reader.langs, []string{"zh", "zh", "en", "en"}) {
		t.Fatal(reader.langs)
	}
}
func TestRecommendationsMembershipBelowPreviewAndHomeOrdering(t *testing.T) {
	svc, reader := fixture(11)
	members := []string{}
	weights := map[string]int64{}
	for _, site := range reader.sites {
		members = append(members, site.ID)
		weights[site.ID] = 1
	}
	weights["1"] = 0  // Current Site is below Top8, but is still a group member.
	weights["2"] = 10 // Old record promoted by the existing Home weight rule.
	reader.sites[1].UpdateTime = "2000-01-01"
	reader.groups = []navmodels.GroupVo{{ID: "a", Sites: members, SiteWeights: weights}}
	before := append([]navmodels.SiteVo(nil), reader.sites...)
	result := svc.GetRecommendations(1, "zh", 99)
	if len(result.Items) != 8 {
		t.Fatal(ids(result.Items))
	}
	for _, site := range result.Items {
		if site.ID == "1" || site.ID == "3" || site.ID == "4" {
			t.Fatalf("outside Home Top8: %s", site.ID)
		}
	}
	if !reflect.DeepEqual(before, reader.sites) {
		t.Fatal("mutated cached sites")
	}
	if len(svc.GetRecommendations(1, "zh", 3).Items) != 3 {
		t.Fatal("limit ignored")
	}
}
func TestRecommendationsShortageNeverUsesRankNineOrUnrelatedSites(t *testing.T) {
	svc, reader := fixture(12)
	reader.groups = []navmodels.GroupVo{{ID: "a", Sites: []string{"1", "2", "3", "4", "5", "6", "7", "8", "9"}}, {ID: "b", Sites: []string{"10", "11", "12"}}}
	for _, limit := range []int{0, 8, 999} {
		result := svc.GetRecommendations(9, "zh", limit)
		if len(result.Items) != 7 {
			t.Fatal(ids(result.Items))
		}
		for _, site := range result.Items {
			if site.ID == "1" || site.ID == "9" || len(site.ID) > 1 {
				t.Fatalf("filled shortage: %s", site.ID)
			}
		}
	}
	if len(svc.GetRecommendations(100, "zh", 8).Items) != 0 {
		t.Fatal("invented membership")
	}
}
func TestRecommendationsStableUTCDailyShuffle(t *testing.T) {
	svc, reader := fixture(5)
	reader.groups = []navmodels.GroupVo{{ID: "a", Sites: []string{"1", "2", "3", "4", "5"}}}
	first := ids(svc.GetRecommendations(1, "zh", 8).Items)
	for i := 0; i < 20; i++ {
		if !reflect.DeepEqual(first, ids(svc.GetRecommendations(1, "en", 8).Items)) {
			t.Fatal("per-request/locale shuffle")
		}
	}
	svc.now = func() time.Time { return time.Date(2026, 9, 28, 7, 59, 59, 0, time.FixedZone("CST", 8*3600)) }
	if !reflect.DeepEqual(first, ids(svc.GetRecommendations(1, "zh", 8).Items)) {
		t.Fatal("used local day instead of UTC")
	}
	svc.now = func() time.Time { return time.Date(2026, 9, 28, 8, 0, 0, 0, time.FixedZone("CST", 8*3600)) }
	second := ids(svc.GetRecommendations(1, "zh", 8).Items)
	if reflect.DeepEqual(first, second) {
		t.Fatalf("date did not rotate selected fixture: %v", first)
	}
	if len(second) != len(first) {
		t.Fatal("date changed eligibility")
	}
}
func TestRecommendationsOptionalUnavailableAndEmpty(t *testing.T) {
	svc, reader := fixture(1)
	for _, failure := range []string{"sites", "groups"} {
		reader.failure = failure
		response := svc.GetRecommendations(1, "zh", 8)
		if response.State != "unavailable" || len(response.Items) != 0 || response.Items == nil {
			t.Fatalf("failure %+v", response)
		}
	}
	reader.failure = ""
	response := svc.GetRecommendations(1, "zh", 8)
	if response.State != "ready" || response.Items == nil || len(response.Items) != 0 {
		t.Fatalf("empty %+v", response)
	}
}

func TestRecommendationsUnionThenDailyLimit(t *testing.T) {
	svc, reader := fixture(18)
	reader.sites[2].Nsfw = "1"
	reader.groups = []navmodels.GroupVo{{ID: "a", Sites: []string{"1", "2", "3", "4", "5", "6", "7", "8", "9"}}, {ID: "b", Sites: []string{"1", "9", "10", "11", "12", "13", "14", "15", "16", "17"}}, {ID: "other", Sites: []string{"18"}}}
	seen := map[string]bool{}
	for day := 1; day <= 30; day++ {
		svc.now = func() time.Time { return time.Date(2026, 9, day, 0, 0, 0, 0, time.UTC) }
		result := svc.GetRecommendations(1, "zh", 8)
		if len(result.Items) != 8 {
			t.Fatal("union limit", ids(result.Items))
		}
		unique := map[string]bool{}
		for _, site := range result.Items {
			if unique[site.ID] || site.ID == "1" || site.ID == "18" {
				t.Fatal("bad union", ids(result.Items))
			}
			unique[site.ID] = true
			seen[site.ID] = true
			number := 0
			fmt.Sscan(site.ID, &number)
			if site.ViewCount != int64(number*100) {
				t.Fatal("view snapshot changed")
			}
			if site.ID == "3" && site.Nsfw != "1" {
				t.Fatal("filtered or changed raw content mode")
			}
		}
		reader.groups[0], reader.groups[1] = reader.groups[1], reader.groups[0]
		if !reflect.DeepEqual(ids(result.Items), ids(svc.GetRecommendations(1, "zh", 8).Items)) {
			t.Fatal("group iteration changed stable shuffle")
		}
	}
	if len(seen) != 16 {
		t.Fatalf("Top8 union candidates never eligible: %v", seen)
	}
}
