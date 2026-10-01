package controller

import (
	"context"
	"os"
	"testing"
	"time"

	env "github.com/gofurry/gofurry-admin/config"
	gamesqlc "github.com/gofurry/gofurry-admin/internal/db/game/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"go.yaml.in/yaml/v4"
)

func TestPrizeTranslationSQLRoundTrip(t *testing.T) {
	path := os.Getenv("GOFURRY_ADMIN_INTEGRATION_CONFIG")
	if path == "" {
		t.Skip("set GOFURRY_ADMIN_INTEGRATION_CONFIG")
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var config struct {
		Business struct {
			Game env.DataBaseConfig `yaml:"game"`
		} `yaml:"business_databases"`
	}
	if err = yaml.Unmarshal(content, &config); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	conn, err := pgx.Connect(ctx, config.Business.Game.Postgres.ConnectionString())
	if err != nil {
		t.Fatal("connect integration database failed")
	}
	defer conn.Close(ctx)
	tx, err := conn.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(context.Background())
	q := gamesqlc.New(tx)
	stamp := pgtype.Timestamp{Time: time.Date(2026, 10, 1, 19, 0, 0, 0, time.UTC), Valid: true}
	body := []byte(`{"title":"礼品卡","title_en":"Gift card","platform":"平台","platform_en":"Platform","keys":["TEST-A","TEST-B"]}`)
	created, err := q.InsertPrize(ctx, gamesqlc.InsertPrizeParams{ID: -time.Now().UnixNano(), Title: "活动", TitleEn: "Event", Description: "描述", DescEn: "Description", Prize: body, Key: "join", StartTime: stamp, EndTime: stamp, Status: true})
	if err != nil {
		t.Fatal(err)
	}
	read, err := q.GetPrize(ctx, created.ID)
	if err != nil {
		t.Fatal(err)
	}
	dto := prizeDTO(prizeModel(read))
	if dto.TitleEn != "Event" || dto.DescEn != "Description" || dto.Prize.TitleEn != "Gift card" || dto.Prize.PlatformEn != "Platform" || len(dto.Prize.Keys) != 2 {
		t.Fatal("create/read translations missing")
	}
	updated, err := q.UpdatePrize(ctx, gamesqlc.UpdatePrizeParams{ID: created.ID, Title: "活动二", TitleEn: "Event Two", Description: "描述二", DescEn: "Description Two", Prize: body, Key: "join", StartTime: stamp, EndTime: stamp, Status: false})
	if err != nil {
		t.Fatal(err)
	}
	if updated.TitleEn != "Event Two" || updated.DescEn != "Description Two" || updated.Status {
		t.Fatal("update translations/state missing")
	}
}
