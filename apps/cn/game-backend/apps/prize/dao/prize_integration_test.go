package dao

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/gofurry/gofurry-game-backend/apps/prize/models"
	gamesqlc "github.com/gofurry/gofurry-game-backend/internal/db/game/sqlc"
	"github.com/gofurry/gofurry-game-backend/roof/env"
	"github.com/jackc/pgx/v5"
	"go.yaml.in/yaml/v4"
)

// Uses only a rollback-only fixture row; never starts scheduling or sends email.
func TestPrizeTranslationsPersistThroughClose(t *testing.T) {
	path := os.Getenv("GOFURRY_GAME_BACKEND_INTEGRATION_CONFIG")
	if path == "" {
		t.Skip("set GOFURRY_GAME_BACKEND_INTEGRATION_CONFIG")
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var config struct {
		Database env.DataBaseConfig `yaml:"database"`
	}
	if err = yaml.Unmarshal(content, &config); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	conn, err := pgx.Connect(ctx, config.Database.ConnectionString())
	if err != nil {
		t.Fatal("connect integration database failed")
	}
	defer conn.Close(ctx)
	tx, err := conn.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(context.Background())
	id := -time.Now().UnixNano()
	_, err = tx.Exec(ctx, `INSERT INTO gfg_prize(id,title,title_en,"desc",desc_en,prize,"key",start_time,end_time,create_time,status)
 VALUES ($1,'活动','Event','描述','Description','{"title":"奖品","title_en":"Reward","platform":"平台","platform_en":"Platform","keys":["TEST-ONLY"]}','join','2026-10-01','2026-10-07','2026-10-01',true)`, id)
	if err != nil {
		t.Fatal(err)
	}
	dao := &PrizeDAO{queries: gamesqlc.New(tx)}
	var record models.GfgPrize
	if err := dao.GetById(id, &record); err != nil {
		t.Fatal(err)
	}
	if record.TitleEn != "Event" || record.DescEn != "Description" {
		t.Fatal("read translations missing")
	}
	active, activeErr := dao.GetLotteryActive()
	if activeErr != nil {
		t.Fatal(activeErr)
	}
	found := false
	for _, item := range active {
		if item.ID == id {
			found = item.TitleEn == "Event" && item.DescEn == "Description"
		}
	}
	if !found {
		t.Fatal("active projection lost translations")
	}
	originalPrize := record.Prize
	record.Status = false
	if count, err := dao.Save(id, record); err != nil || count != 1 {
		t.Fatalf("close: %d %v", count, err)
	}
	if err := dao.GetById(id, &record); err != nil {
		t.Fatal(err)
	}
	if record.Status || record.TitleEn != "Event" || record.DescEn != "Description" || record.Prize != originalPrize {
		t.Fatal("close lost translations/keys")
	}
	history, historyErr := dao.GetLotteryHistory()
	if historyErr != nil {
		t.Fatal(historyErr)
	}
	for _, item := range history {
		if item.ID == id {
			if item.TitleEn != "Event" || item.DescEn != "Description" {
				t.Fatal("history translations missing")
			}
			return
		}
	}
	t.Fatal("closed prize missing from history")
}
