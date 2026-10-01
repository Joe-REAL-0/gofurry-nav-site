package models

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	cm "github.com/gofurry/gofurry-game-backend/common/models"
)

func TestPrizeDisplayTranslationAndSecretBoundary(t *testing.T) {
	prize := PrizeModel{Title: "礼品卡", TitleEn: "Gift card", Platform: "平台", PlatformEn: "Platform", Keys: []string{"SECRET-A", "SECRET-B"}}
	display := prize.Display()
	for _, lang := range []string{"zh", "en", "invalid"} {
		got := display.Localized(lang)
		wantTitle, wantPlatform := "礼品卡", "平台"
		if lang == "en" {
			wantTitle, wantPlatform = "Gift card", "Platform"
		}
		if got.Title != wantTitle || got.Platform != wantPlatform || got.Count != 2 {
			t.Fatalf("%s: %+v", lang, got)
		}
		body, err := json.Marshal(got)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(body), "keys") || strings.Contains(string(body), "SECRET") || strings.Contains(string(body), "_en") {
			t.Fatalf("unsafe projection: %s", body)
		}
	}
	cache, _ := json.Marshal(display)
	if strings.Contains(string(cache), "SECRET") || !strings.Contains(string(cache), "title_en") {
		t.Fatal("history display must retain translations, never keys")
	}
	prize.PlatformEn = "  "
	if got := prize.Display().Localized("en"); got.Title != "Gift card" || got.Platform != "平台" {
		t.Fatalf("field fallback: %+v", got)
	}
}

func TestHistoryLocalizationPreservesCacheAndLegacy(t *testing.T) {
	history := PrizeWinnerCacheSaveModel{PrizeCount: 4, Prize: []PrizeCacheSaveModel{
		{Name: "活动", NameEn: "Event", Desc: "描述", DescEn: "Description", Prize: PrizeDisplay{Title: "奖品", TitleEn: "Reward", Platform: "Steam", Count: 2}, Count: 12},
		{Name: "旧活动", Desc: "旧描述", Prize: PrizeDisplay{Title: "旧奖品", Platform: "旧平台", Count: 1}},
	}}
	for i := range history.Prize {
		history.Prize[i].EndTime = cm.LocalTime(time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC))
	}
	encoded, _ := json.Marshal(history)
	var restored PrizeWinnerCacheSaveModel
	if err := json.Unmarshal(encoded, &restored); err != nil {
		t.Fatal(err)
	}
	en := restored.Localized("en")
	if en.PrizeCount != 4 || en.Prize[0].Name != "Event" || en.Prize[0].Desc != "Description" || en.Prize[0].Prize.Title != "Reward" || en.Prize[0].Count != 12 {
		t.Fatalf("English history: %+v", en)
	}
	if en.Prize[1].Name != "旧活动" || en.Prize[1].Prize.Title != "旧奖品" {
		t.Fatal("legacy fallback failed")
	}
	if restored.Prize[0].Name != "活动" || restored.Prize[0].NameEn != "Event" || restored.Localized("zh").Prize[0].Desc != "描述" {
		t.Fatal("language projection mutated shared cache")
	}
}
