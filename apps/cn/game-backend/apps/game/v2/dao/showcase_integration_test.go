package dao_test

import (
	"context"
	"fmt"
	"github.com/gofurry/gofurry-game-backend/apps/game/v2/showcase"
	"github.com/jackc/pgx/v5/pgxpool"
	"strings"
	"testing"
	"time"
)

func assertShowcaseReadModels(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()
	now := time.Now().UTC()
	day := now.Truncate(24 * time.Hour)
	through := day.AddDate(0, 0, -1)
	exec := func(sql string, args ...any) {
		t.Helper()
		if _, e := pool.Exec(ctx, sql, args...); e != nil {
			t.Fatalf("fixture SQL failed: %s: %v", sql, e)
		}
	}
	for id := int64(119101); id <= 119108; id++ {
		exec(`INSERT INTO gfg_game(id,name,name_en,info,info_en,appid,header,developers,publishers,weight,create_time,update_time,showcase_eligible)
   VALUES($1,'标题','Title','摘要',CASE WHEN $1::bigint=119106 THEN '' ELSE 'Summary' END,$1,'https://shared.akamai.steamstatic.com/header.jpg','[]','[]',0,now(),now(),$1::bigint<>119105)`, id)
		availability := "upcoming"
		if id == 119102 || id == 119103 || id == 119107 {
			availability = "available"
		}
		exec(`INSERT INTO gfg_game_release_state(game_id,availability,precision,source,source_region,source_locale,normalizer_version,observed_at) VALUES($1,$2,'tba','steam','CN','zh','test',now())`, id, availability)
	}
	exec(`UPDATE gfg_game SET header='unusable' WHERE id=119108`)
	exec(`INSERT INTO gfg_tag(id,code,name,name_en,info,info_en,category_id,create_time,update_time) VALUES(119901,'adult','成人','Adult','','',1,now(),now()) ON CONFLICT(code) DO NOTHING`)
	exec(`INSERT INTO gfg_game_tag(game_id,tag_id,role,create_time,update_time) SELECT 119104,id,'normal',now(),now() FROM gfg_tag WHERE code='adult'`)
	first := day.AddDate(0, 0, -5)
	exec(`INSERT INTO gfg_game_first_available(game_id,precision,exact_date,release_year,release_month,window_start,window_end,source,inferred,normalizer_version) VALUES(119102,'day',$1,extract(year FROM $1::date)::int,extract(month FROM $1::date)::int,$1,$1,'legacy_manual',false,'test')`, first)
	exec(`INSERT INTO gfg_game_assets(game_id,appid,asset_type,asset_family,source,lang,media_key,url,exists,collected_at,updated_at) VALUES(119101,119101,'header','store','store_browse','zh','header','https://shared.akamai.steamstatic.com/observed-header.jpg',true,now(),now())`)
	exec(`UPDATE gfg_fact_rollup_checkpoints SET source_start_date=$1::date-30,processed_through=$1 WHERE pipeline_key='game.player_facts'`, through)
	exec(`INSERT INTO gfg_game_tracking_periods(id,game_id,appid,tracked_from,tracking_basis,opened_reason) VALUES(119103,119103,119103,$1::timestamptz-interval '30 days','explicit','test'),(119107,119107,119107,$1,'explicit','test')`, through)
	exec(`INSERT INTO gfg_game_tracking_periods(id,game_id,appid,tracked_from,tracked_until,tracking_basis,opened_reason,closed_reason) VALUES(119207,119107,119107,$1::timestamptz-interval '30 days',$1,'explicit','test','replacement')`, through)
	for _, period := range []struct {
		period, game int64
		days         int
	}{{119103, 119103, 10}, {119107, 119107, 1}, {119207, 119107, 10}} {
		for age := 0; age < period.days; age++ {
			average := 100
			if age < 3 {
				average = 700
			}
			date := through.AddDate(0, 0, -age)
			if period.period == 119207 {
				date = date.AddDate(0, 0, -1)
			}
			exec(`INSERT INTO gfg_game_player_daily(tracking_period_id,game_id,appid,fact_date,min_players,max_players,avg_players,median_players,expected_samples,attempted_samples,successful_samples,partial_samples,failed_samples,skipped_samples,missed_samples,canceled_samples,unattempted_samples,quality_basis,projection_version,finalized_at)
    VALUES($1,$2,$2,$3,$4::bigint,$4::bigint,($4::bigint)::double precision,($4::bigint)::double precision,10,10,10,0,0,0,0,0,0,'acquisition_ledger',1,now())`, period.period, period.game, date, average)
		}
	}
	exec(`INSERT INTO gfg_game_player_daily(tracking_period_id,game_id,appid,fact_date,min_players,max_players,avg_players,median_players,expected_samples,attempted_samples,successful_samples,partial_samples,failed_samples,skipped_samples,missed_samples,canceled_samples,unattempted_samples,quality_basis,projection_version,finalized_at)
 VALUES(119103,119103,119103,$1,999999,999999,999999,999999,10,10,10,0,0,0,0,0,0,'acquisition_ledger',1,now())`, day)
	reader := showcase.SQLReader{Pool: pool}
	input, e := reader.Read(ctx, now, "zh", "CN")
	if e != nil {
		t.Fatal(e)
	}
	byKey := map[string]showcase.Candidate{}
	for _, c := range input.Candidates {
		byKey[c.Item.Key] = c
	}
	for _, key := range []string{"auto:upcoming:119101", "auto:new_release:119102", "auto:trending:119103", "auto:upcoming:119106"} {
		if _, ok := byKey[key]; !ok {
			t.Fatal("missing candidate", key)
		}
	}
	if !strings.Contains(byKey["auto:upcoming:119101"].Item.Artwork.URL, "observed-header") {
		t.Fatal("current Steam asset data not used")
	}
	for _, id := range []int{119104, 119105, 119107, 119108} {
		for key := range byKey {
			if strings.HasSuffix(key, fmt.Sprintf(":%d", id)) {
				t.Fatal("ineligible Game", key)
			}
		}
	}
	for _, d := range input.Diagnostics["trending"] {
		if d.GameID == "119103" && (d.Trending.Recent.Average != 700 || d.Trending.Baseline.Average != 100) {
			t.Fatal("horizon/current identity violated", d)
		}
	}
	english, e := reader.Read(ctx, now, "en", "CN")
	if e != nil {
		t.Fatal(e)
	}
	for _, c := range english.Candidates {
		if c.Item.GameID == "119106" {
			t.Fatal("locale fallback")
		}
	}
	// The FK preserves Campaign history, while missing internal targets stop composing.
	exec(`INSERT INTO gfg_showcase_campaign(id,internal_name,content_type,linked_game_id,state,starts_at,ends_at,desktop_object_key,primary_action_type) VALUES(119001,'linked','game',119102,'published',$1::timestamptz-interval '1 hour',$1::timestamptz+interval '1 hour','game/showcase/119001/desktop/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.avif','game')`, now)
	exec(`INSERT INTO gfg_showcase_campaign_locale(campaign_id,lang,enabled,title,summary) VALUES(119001,'zh',true,'Managed','Summary')`)
	input, e = reader.Read(ctx, now, "zh", "CN")
	if e != nil {
		t.Fatal(e)
	}
	found := false
	for _, c := range input.Candidates {
		found = found || c.Item.Key == "campaign:119001"
	}
	if !found {
		t.Fatal("managed linked candidate missing")
	}
	exec(`DELETE FROM gfg_game WHERE id=119102`)
	input, e = reader.Read(ctx, now, "zh", "CN")
	if e != nil {
		t.Fatal(e)
	}
	for _, c := range input.Candidates {
		if c.Item.Key == "campaign:119001" {
			t.Fatal("deleted internal action composed")
		}
	}
}
