package dao_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	v2dao "github.com/gofurry/gofurry-game-backend/apps/game/v2/dao"
	v2models "github.com/gofurry/gofurry-game-backend/apps/game/v2/models"
	v2service "github.com/gofurry/gofurry-game-backend/apps/game/v2/service"
	gamedb "github.com/gofurry/gofurry-game-backend/internal/db/game/sqlc"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Extends the existing disposable PG18 gate; no separate fixture database.
func assertCollectionTimelineDecorations(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()
	exec := func(sql string) {
		t.Helper()
		if _, err := pool.Exec(ctx, sql); err != nil {
			t.Fatal(err)
		}
	}
	exec(`INSERT INTO gfg_game(id,name,name_en,info,info_en,appid,header,developers,publishers,weight,create_time,update_time)
SELECT id,'作品','Game','简介','Summary',id,'header','[]','[]',0,now(),now() FROM generate_series(145101,145130) id;
INSERT INTO gfg_game_collection(id,code,name,name_en,status,published_at) VALUES
(145001,'decoration-small','小分区','Small','published',now()),(145002,'decoration-large','大分区','Large','published',now());
INSERT INTO gfg_game_collection_item(collection_id,game_id) SELECT 145001,id FROM generate_series(145101,145103) id;
INSERT INTO gfg_game_collection_item(collection_id,game_id) SELECT 145002,id FROM generate_series(145101,145130) id;
INSERT INTO gfg_tag(id,code,name,name_en,info,info_en,category_id,create_time,update_time) VALUES
(145201,'timeline-primary','剧情','Story','','',1,now(),now()),(145202,'timeline-secondary','','Visual novel','','',1,now(),now()),
(145203,'timeline-fallback','回退','','','',1,now(),now());
INSERT INTO gfg_game_tag(game_id,tag_id,role,create_time,update_time) VALUES
(145101,145201,'primary',now(),now()),(145101,145202,'secondary',now(),now()),
(145102,145203,'primary',now(),now()),(145103,145202,'secondary',now(),now()),(145104,145201,'normal',now(),now());
INSERT INTO gfg_game_tag(game_id,tag_id,role,create_time,update_time) SELECT 145130,id,'normal',now(),now() FROM gfg_tag WHERE code='adult';
INSERT INTO gfg_game_comment(id,region,content,score,create_time,game_id,ip,name) VALUES
(145001,'CN','review',4,now(),145101,'127.0.0.1','tester'),(145002,'CN','review',5,now(),145101,'127.0.0.1','tester');
INSERT INTO gfg_game_player_counts(id,run_id,game_id,appid,count,status,collected_at) VALUES
(145001,'dec-old',145101,145101,9999,'success','2026-10-01'),
(145002,'dec-tie-1',145101,145101,10,'success','2026-10-02'),
(145003,'dec-tie-2',145101,145101,1234,'success','2026-10-02'),
(145004,'dec-failure',145101,145101,90000,'failed','2026-10-03'),
(145005,'dec-zero',145102,145102,0,'success','2026-10-02'),
(145006,'dec-only-failure',145103,145103,90000,'failed','2026-10-03');
UPDATE gfg_game SET groups='[{"key":"discord","value":"https://private-community-one"},{"key":"telegram","value":"https://private-community-two"}]' WHERE id=145101;
UPDATE gfg_game SET groups='[{}, {"key":"discord","value":""}, {"key":"","value":"https://x"}]' WHERE id=145102;
UPDATE gfg_game SET groups='[]' WHERE id=145103;
UPDATE gfg_game SET groups='{}' WHERE id=145104;
UPDATE gfg_game SET groups='null' WHERE id=145105;
UPDATE gfg_game SET groups=NULL WHERE id=145106;
UPDATE gfg_game SET groups='[{"key":"   ","value":"https://x"},{"key":"discord","value":"   "}]' WHERE id=145107;
UPDATE gfg_game SET groups='[{"key":" discord ","value":" https://private-community-single "}]' WHERE id=145108;
UPDATE gfg_game SET groups='"non-array"', resources='[{"key":"discord","value":"https://resource"}]', links='[{"key":"telegram","value":"https://link"}]' WHERE id=145109;`)
	t.Run("one decoration row per game enforced by role uniqueness", func(t *testing.T) {
		for _, role := range []string{"primary", "secondary"} {
			_, err := pool.Exec(ctx, `INSERT INTO gfg_game_tag(game_id,tag_id,role,create_time,update_time) VALUES (145101,145203,$1,now(),now())`, role)
			var pgErr *pgconn.PgError
			if !errors.As(err, &pgErr) || pgErr.Code != "23505" || pgErr.ConstraintName != "idx_gfg_game_tag_"+role {
				t.Fatalf("expected unique %s role invariant: %v", role, err)
			}
		}
		rows, err := gamedb.New(pool).BatchCollectionTimelineDecorations(ctx, gamedb.BatchCollectionTimelineDecorationsParams{GameIds: []int64{145101, 145102, 145103}, Lang: "zh"})
		if err != nil || len(rows) != 3 {
			t.Fatalf("one raw row per game: %v %v", rows, err)
		}
		seen := map[int64]bool{}
		for _, row := range rows {
			if seen[row.GameID] {
				t.Fatalf("duplicate game %d", row.GameID)
			}
			seen[row.GameID] = true
		}
	})
	t.Run("timeline decoration semantics and bounded reads", func(t *testing.T) {
		tracer := &collectionQueryTrace{}
		cfg := pool.Config()
		cfg.ConnConfig.Tracer = tracer
		traced, err := pgxpool.NewWithConfig(ctx, cfg)
		if err != nil {
			t.Fatal(err)
		}
		defer traced.Close()
		dao := v2dao.NewReadModelDAO(traced)
		for _, lang := range []string{"zh", "en"} {
			decorations, err := dao.LoadCollectionTimelineDecorations(ctx, []int64{145101, 145102, 145103, 145104, 145105, 145106, 145107, 145108, 145109}, lang)
			if err != nil {
				t.Fatal(err)
			}
			both := decorations[145101]
			primaryName := "剧情"
			if lang == "en" {
				primaryName = "Story"
			}
			if both.PrimaryTag == nil || both.SecondaryTag == nil || both.PrimaryTag.Code != "timeline-primary" || both.PrimaryTag.Name != primaryName || both.SecondaryTag.Name != "Visual novel" {
				t.Fatal("role/locale fallback", both)
			}
			if both.Rating == nil || both.Rating.Average != 4.5 || both.Rating.Count != 2 {
				t.Fatal("rating", both)
			}
			if both.Online == nil || both.Online.Count != 1234 || both.Online.CollectedAt.Format(time.DateOnly) != "2026-10-02" {
				t.Fatal("latest success, id tie or peak drift", both)
			}
			if both.CommunityCount != 2 || decorations[145102].CommunityCount != 0 || decorations[145108].CommunityCount != 1 {
				t.Fatal("community counts", decorations)
			}
			if decorations[145102].SecondaryTag != nil || decorations[145102].PrimaryTag == nil || decorations[145103].PrimaryTag != nil || decorations[145103].SecondaryTag == nil {
				t.Fatal("single roles", decorations)
			}
			if decorations[145102].PrimaryTag.Name != "回退" {
				t.Fatal("primary locale fallback", decorations[145102])
			}
			if decorations[145102].Rating != nil || decorations[145102].Online == nil || decorations[145102].Online.Count != 0 || decorations[145103].Online != nil {
				t.Fatal("null vs real zero", decorations)
			}
			for _, id := range []int64{145104, 145105, 145106, 145107, 145109} {
				d := decorations[id]
				if d.PrimaryTag != nil || d.SecondaryTag != nil || d.Rating != nil || d.Online != nil || d.CommunityCount != 0 {
					t.Fatal("missing/malformed metadata", id, d)
				}
			}
			if tracer.count != 1 {
				t.Fatal("decorations must use one SQL", tracer.count)
			}
			tracer.count = 0
		}
		service := v2service.NewCollectionService(dao, nil)
		for _, tc := range []struct {
			code, mode string
			count      int
		}{{"decoration-small", "sfw", 3}, {"decoration-large", "sfw", 29}, {"decoration-large", "nsfw", 30}} {
			tracer.count = 0
			tracer.decorationCount = 0
			tracer.queries = nil
			start := time.Now()
			detail, err := service.Detail(ctx, tc.code, v2models.CollectionQuery{Mode: tc.mode})
			if err != nil || len(detail.Items) != tc.count || tracer.count != 8 || tracer.decorationCount != 1 || len(tracer.decorationIDs) != tc.count {
				t.Fatalf("%s/%s: queries=%d decorations=%d ids=%v err=%v", tc.code, tc.mode, tracer.count, tracer.decorationCount, tracer.decorationIDs, err)
			}
			if tc.mode == "sfw" {
				for _, id := range tracer.decorationIDs {
					if id == 145130 {
						t.Fatal("hidden adult reached decorations")
					}
				}
			}
			for _, forbidden := range []string{"gfg_game_prices", "gfg_game_requirements", "gfg_game_news", "queryOnlinePeakCounts", "MAX("} {
				if strings.Contains(strings.Join(tracer.queries, "\n"), forbidden) {
					t.Fatal("unexpected full read/peak", forbidden)
				}
			}
			body, _ := json.Marshal(detail)
			if strings.Contains(string(body), "private-community") {
				t.Fatal("community URLs escaped")
			}
			t.Logf("cold %s/%s: %d visible, %d SQL, %s (diagnostic only)", tc.code, tc.mode, tc.count, tracer.count, time.Since(start))
		}
	})
}
