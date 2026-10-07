package service

import (
	"reflect"
	"testing"
	"time"

	"github.com/gofurry/gofurry-nav-backend/apps/nav/navPage/models"
	commonmodels "github.com/gofurry/gofurry-nav-backend/common/models"
)

func TestConvertGroupRecordsOrdersMappingsByGroupWeight(t *testing.T) {
	svc := &navPageService{}
	groups := []models.GfnSiteGroup{
		{ID: 10, Name: "论坛社区", NameEn: "Forums", Info: "论坛", InfoEn: "Forums", Priority: 1},
	}
	mappings := []models.GfnSiteGroupMap{
		{ID: 9, GroupID: 10, SiteID: 300, Weight: 10},
		{ID: 3, GroupID: 10, SiteID: 100, Weight: 20},
		{ID: 5, GroupID: 10, SiteID: 200, Weight: 20},
	}

	result := svc.convertGroupRecords(groups, mappings, "zh")
	if len(result) != 1 {
		t.Fatalf("group count = %d", len(result))
	}

	want := []string{"200", "100", "300"}
	if !reflect.DeepEqual(result[0].Sites, want) {
		t.Fatalf("sites order = %v, want %v", result[0].Sites, want)
	}
	if result[0].SiteWeights["200"] != 20 || result[0].SiteWeights["300"] != 10 {
		t.Fatalf("site weights were not carried: %#v", result[0].SiteWeights)
	}
}

func TestConvertRecordsIncludesUpdateTime(t *testing.T) {
	svc := &navPageService{}
	records := []models.GfnSite{
		{
			ID:         7,
			Name:       "兽时社区",
			NameEn:     "Shoutime",
			Domain:     `{"domain":["www.shoutime.net"]}`,
			Info:       "中文站点",
			InfoEn:     "Forum site",
			Nsfw:       "0",
			Welfare:    "0",
			UpdateTime: models2LocalTime("2026-06-05 12:34:56"),
		},
	}

	result := svc.convertRecords(records, "zh")
	if len(result) != 1 {
		t.Fatalf("record count = %d", len(result))
	}
	if result[0].UpdateTime != "2026-06-05 12:34:56" {
		t.Fatalf("update_time = %q", result[0].UpdateTime)
	}
}

func models2LocalTime(value string) commonmodels.LocalTime {
	return commonmodels.LocalTime(parseTestTime(value))
}

func parseTestTime(value string) time.Time {
	tm, err := time.ParseInLocation("2006-01-02 15:04:05", value, time.Local)
	if err != nil {
		panic(err)
	}
	return tm
}
