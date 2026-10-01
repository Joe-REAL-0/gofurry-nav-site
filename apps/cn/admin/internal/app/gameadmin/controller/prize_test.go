package controller

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/gofurry/gofurry-admin/internal/app/gameadmin/models"
)

func TestPrizeTranslationAndKeyNormalization(t *testing.T) {
	var payload models.PrizePayload
	err := json.Unmarshal([]byte(`{"title":"活动","title_en":"Event","desc":"描述","desc_en":"Description","prize":{"title":" 礼品卡 ","title_en":" Gift card ","platform":" 平台 ","platform_en":" Platform ","keys":[" KEY-A\r","","  ","KEY-B ","KEY-A"]}}`), &payload)
	if err != nil {
		t.Fatal(err)
	}
	body := normalizePrizeBody(payload.Prize)
	if body.Title != "礼品卡" || body.TitleEn != "Gift card" || body.Platform != "平台" || body.PlatformEn != "Platform" || !reflect.DeepEqual(body.Keys, []string{"KEY-A", "KEY-B"}) {
		t.Fatalf("normalized: %+v", body)
	}
	encoded, _ := json.Marshal(body)
	dto := prizeDTO(models.Prize{Title: payload.Title, TitleEn: payload.TitleEn, Desc: payload.Desc, DescEn: payload.DescEn, Prize: string(encoded)})
	if dto.TitleEn != "Event" || dto.DescEn != "Description" || !reflect.DeepEqual(dto.Prize, body) {
		t.Fatalf("translation round trip: %+v", dto)
	}
}
