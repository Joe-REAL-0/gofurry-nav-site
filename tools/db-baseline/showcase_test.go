package main

import (
	"context"
	"github.com/pressly/goose/v3"
	"path/filepath"
	"testing"
	"time"
)

func TestShowcaseAdditiveMigration(t *testing.T) {
	dsn := integrationAdminDSN(t)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	admin := openDatabase(t, dsn, "postgres")
	defer admin.Close()
	name := temporaryDatabaseName("gfg", "showcase")
	createDatabase(t, ctx, admin, name)
	defer dropDatabase(t, admin, name)
	db := openDatabase(t, dsn, name)
	defer db.Close()
	dir, _ := filepath.Abs("../../db/game/migrations")
	if e := goose.SetDialect("postgres"); e != nil {
		t.Fatal(e)
	}
	if e := goose.UpToContext(ctx, db, dir, 20261001010000); e != nil {
		t.Fatal(e)
	}
	if _, e := db.ExecContext(ctx, `INSERT INTO gfg_game(id,name,name_en,info,info_en,appid,header,developers,publishers,weight,create_time,update_time) VALUES(11901,'旧资料','Existing','摘要','Summary',11901,'','[]','[]',9,now(),now())`); e != nil {
		t.Fatal(e)
	}
	var before, after string
	if e := db.QueryRowContext(ctx, `SELECT to_jsonb(g)::text FROM gfg_game g WHERE id=11901`).Scan(&before); e != nil {
		t.Fatal(e)
	}
	if e := goose.UpContext(ctx, db, dir); e != nil {
		t.Fatal(e)
	}
	if e := db.QueryRowContext(ctx, `SELECT (to_jsonb(g)-'showcase_eligible')::text FROM gfg_game g WHERE id=11901`).Scan(&after); e != nil || before != after {
		t.Fatal("existing Game changed", e)
	}
	var count int
	var eligible bool
	if e := db.QueryRowContext(ctx, `SELECT showcase_eligible FROM gfg_game WHERE id=11901`).Scan(&eligible); e != nil || eligible {
		t.Fatal("default eligibility", e)
	}
	for _, table := range []string{"gfg_showcase_campaign", "gfg_showcase_campaign_locale", "gfg_showcase_daily_stat", "gfg_showcase_analytics_daily_quality"} {
		if e := db.QueryRowContext(ctx, "SELECT count(*) FROM "+table).Scan(&count); e != nil || count != 0 {
			t.Fatal("non-empty new table", table, e)
		}
	}
	if _, e := db.ExecContext(ctx, `INSERT INTO gfg_showcase_campaign(id,internal_name,content_type,linked_game_id) VALUES(1,'Campaign','game',11901);INSERT INTO gfg_showcase_campaign_locale(campaign_id,lang) VALUES(1,'zh');INSERT INTO gfg_showcase_daily_stat(stat_date,subject_key,subject_kind,campaign_id,game_id,reason) VALUES('2026-10-03','campaign:1','managed',1,11901,'editorial')`); e != nil {
		t.Fatal(e)
	}
	for _, sql := range []string{
		`UPDATE gfg_showcase_campaign SET weight=0 WHERE id=1`, `UPDATE gfg_showcase_campaign SET content_type='ad' WHERE id=1`,
		`UPDATE gfg_showcase_campaign SET pin_position=5 WHERE id=1`, `UPDATE gfg_showcase_campaign SET focal_x='NaN' WHERE id=1`,
		`UPDATE gfg_showcase_campaign_locale SET lang='fr' WHERE campaign_id=1`, `UPDATE gfg_showcase_campaign_locale SET tags=ARRAY['1','2','3','4'] WHERE campaign_id=1`,
		`UPDATE gfg_showcase_daily_stat SET valid_impressions=-1`, `UPDATE gfg_showcase_daily_stat SET qualified_clicks=1,click_primary=1`,
		`UPDATE gfg_showcase_daily_stat SET valid_impressions=1,qualified_clicks=1`,
	} {
		if _, e := db.ExecContext(ctx, sql); e == nil {
			t.Fatal("constraint accepted", sql)
		}
	}
	if _, e := db.ExecContext(ctx, `DELETE FROM gfg_game WHERE id=11901`); e != nil {
		t.Fatal(e)
	}
	var detached bool
	if e := db.QueryRowContext(ctx, `SELECT linked_game_id IS NULL FROM gfg_showcase_campaign WHERE id=1`).Scan(&detached); e != nil || !detached {
		t.Fatal("relation deletion", e)
	}
	if e := db.QueryRowContext(ctx, `SELECT count(*) FROM gfg_showcase_daily_stat WHERE campaign_id=1 AND game_id=11901`).Scan(&count); e != nil || count != 1 {
		t.Fatal("history disappeared", e)
	}
	if _, e := db.ExecContext(ctx, `DELETE FROM gfg_showcase_campaign WHERE id=1`); e != nil {
		t.Fatal(e)
	}
	if e := db.QueryRowContext(ctx, `SELECT count(*) FROM gfg_showcase_daily_stat`).Scan(&count); e != nil || count != 1 {
		t.Fatal("stat identity incorrectly FK-bound")
	}
}
