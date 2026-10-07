package bootstrap_test

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofurry/gofurry-admin/internal/app/auth/authorization"
	authmw "github.com/gofurry/gofurry-admin/internal/app/auth/middleware"
	gameadmin "github.com/gofurry/gofurry-admin/internal/app/gameadmin/controller"
	"github.com/gofurry/gofurry-admin/internal/app/shared/audit"
	"github.com/gofurry/gofurry-admin/internal/infra/assets"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type showcaseStore struct {
	objects map[string]assets.Object
	fail    bool
}

func (s *showcaseStore) Put(_ context.Context, o assets.Object) error {
	if s.fail {
		return errors.New("test cloud failure")
	}
	s.objects[o.Key] = o
	return nil
}
func (s *showcaseStore) Head(_ context.Context, k string) (assets.Info, error) {
	if s.fail {
		return assets.Info{}, errors.New("test cloud failure")
	}
	o, ok := s.objects[k]
	return assets.Info{Exists: ok, Size: int64(len(o.Data)), ContentType: o.ContentType, SHA256: o.SHA256, Kind: o.Kind, CacheControl: assets.CacheControl}, nil
}
func (s *showcaseStore) Get(_ context.Context, k string) ([]byte, error) {
	return s.objects[k].Data, nil
}
func showcaseTestAVIF() []byte {
	box := func(kind string, data []byte) []byte {
		b := make([]byte, 8+len(data))
		binary.BigEndian.PutUint32(b, uint32(len(b)))
		copy(b[4:], kind)
		copy(b[8:], data)
		return b
	}
	dims := make([]byte, 12)
	binary.BigEndian.PutUint32(dims[4:], 1600)
	binary.BigEndian.PutUint32(dims[8:], 800)
	props := box("iprp", append(box("ipco", box("ispe", dims)), box("ipma", []byte{0, 0, 0, 0, 0, 0, 0, 1, 0, 1, 1, 1})...))
	return append(box("ftyp", []byte("avif\x00\x00\x00\x00avifmif1")), box("meta", append(append([]byte{0, 0, 0, 0}, box("pitm", []byte{0, 0, 0, 0, 0, 1})...), props...))...)
}
func TestAdminShowcaseThreeDatabase(t *testing.T) {
	base := adminIntegrationDSN(t, requireAdminIntegrationConfig(t))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	root := adminIntegrationSQLDB(t, base, "postgres")
	defer root.Close()
	pools := map[string]*pgxpool.Pool{}
	dsns := map[string]string{}
	for _, domain := range []string{"admin", "game", "nav"} {
		name := integrationDatabaseName("showcase_" + domain)
		createIntegrationDatabase(t, ctx, root, name)
		defer dropIntegrationDatabase(t, root, name)
		dsns[domain] = integrationDatabaseDSN(base, name)
		applyIntegrationBaseline(t, domain, dsns[domain])
		pools[domain] = integrationPool(t, ctx, dsns[domain])
		defer pools[domain].Close()
	}
	redisAddr := os.Getenv("GOFURRY_SHOWCASE_REDIS_ADDR")
	var redisClient *redis.Client
	if redisAddr != "" {
		if !strings.HasPrefix(redisAddr, "127.0.0.1:") {
			t.Fatal("test Redis must be local and disposable")
		}
		redisClient = redis.NewClient(&redis.Options{Addr: redisAddr, DB: 13})
		if e := redisClient.FlushDB(ctx).Err(); e != nil {
			t.Fatal(e)
		}
		defer redisClient.Close()
		defer redisClient.FlushDB(ctx)
	}
	primary := &showcaseStore{objects: map[string]assets.Object{}}
	mirror := &showcaseStore{objects: map[string]assets.Object{}, fail: true}
	api := gameadmin.New(pools["game"], audit.New(pools["admin"])).WithShowcaseAssets(&assets.Service{Primary: primary, Mirror: mirror})
	if redisClient != nil {
		api.WithShowcaseRevision(func(ctx context.Context) error { return redisClient.Incr(ctx, "game:v2:showcase:revision").Err() })
	}
	app := fiber.New()
	principal := &authorization.Principal{Capabilities: authorization.CapabilitiesFor(authorization.RoleOperator)}
	// Route authorization executes the compiled capability policy for the GFA actor below.
	app.Use(func(c fiber.Ctx) error { c.Locals(authorization.PrincipalContextKey, principal); return c.Next() })
	read, write := authmw.Require(authorization.ContentRead), authmw.Require(authorization.ContentWrite)
	app.Get("/campaigns", read, api.ListShowcaseCampaigns)
	app.Post("/campaigns", write, api.CreateShowcaseCampaign)
	app.Get("/campaigns/:id", read, api.GetShowcaseCampaign)
	app.Put("/campaigns/:id/content", write, api.SaveShowcaseContent)
	app.Put("/campaigns/:id/schedule", write, api.SaveShowcaseSchedule)
	app.Post("/campaigns/:id/artwork/:variant", write, api.ShowcaseArtwork)
	app.Delete("/campaigns/:id/artwork/:variant", write, api.ShowcaseArtwork)
	for _, action := range []string{"publish", "pause", "resume", "archive"} {
		app.Post("/campaigns/:id/"+action, write, api.ShowcaseLifecycle)
	}
	app.Delete("/campaigns/:id", write, api.DeleteShowcaseDraft)
	app.Get("/campaigns/:id/stats", read, api.ShowcaseStats)
	app.Get("/campaigns/:id/stats.csv", read, api.ShowcaseStats)
	app.Get("/quality", read, api.ShowcaseQuality)
	app.Post("/games", write, api.CreateGame)
	app.Put("/games/:id/classification", write, api.SaveClassification)
	call := func(method, path, body string, status int) json.RawMessage {
		t.Helper()
		response := requestJSON(t, app, method, path, body, nil, status)
		defer response.Body.Close()
		var envelope integrationEnvelope
		if e := json.NewDecoder(response.Body).Decode(&envelope); e != nil {
			t.Fatal(e)
		}
		return envelope.Data
	}
	create := func(name string) string {
		data := call("POST", "/campaigns", fmt.Sprintf(`{"internal_name":%q,"content_type":"game","sponsored":false}`, name), 200)
		var r struct {
			ID string `json:"id"`
		}
		if e := json.Unmarshal(data, &r); e != nil || r.ID == "" {
			t.Fatal("create", string(data), e)
		}
		return r.ID
	}
	// Real GFA actor prevents fictional identity snapshots.
	if _, e := pools["admin"].Exec(ctx, `INSERT INTO gfa_admin_account(id,username,display_name,role,status,password_hash,session_version,created_at,updated_at) VALUES(119,'showcase-tester','Showcase Tester','operator','active','test',1,now(),now())`); e != nil {
		t.Fatal(e)
	}
	principal.AccountID = 119
	principal.Username = "showcase-tester"
	principal.DisplayName = "Showcase Tester"
	principal.Role = authorization.RoleOperator
	id := create("Showcase integration")
	path := "/campaigns/" + id
	call("POST", path+"/publish", `{}`, 400)
	call("POST", "/campaigns", `{"internal_name":"bad","content_type":"advertisement"}`, 400)
	content := `{"content_type":"game","sponsored":false,"focal_x":0.5,"focal_y":0.5,"locales":[{"lang":"zh","enabled":true,"title":"测试推荐","summary":"完整摘要","tags":["测试"]}],"primary_action_type":"website","primary_target":"https://example.test/game"}`
	call("PUT", path+"/content", content, 200)
	call("PUT", path+"/content", strings.Replace(content, "https://", "http://", 1), 400)
	call("PUT", path+"/content", strings.Replace(content, `["测试"]`, `["1","2","3","4"]`, 1), 400)
	call("PUT", path+"/content", strings.Replace(strings.Replace(content, `"sponsored":false`, `"sponsored":true`, 1), `"summary":"完整摘要"`, `"summary":"完整摘要","editorial_note":"note"`, 1), 400)
	schedule := fmt.Sprintf(`{"starts_at":%q,"ends_at":%q,"weight":100,"pin_position":1}`, time.Now().Add(-time.Hour).UTC().Format(time.RFC3339), time.Now().Add(time.Hour).UTC().Format(time.RFC3339))
	call("PUT", path+"/schedule", schedule, 200)
	upload := func(status int) json.RawMessage {
		t.Helper()
		var b bytes.Buffer
		w := multipart.NewWriter(&b)
		file, e := w.CreateFormFile("file", "showcase.avif")
		if e != nil {
			t.Fatal(e)
		}
		file.Write(showcaseTestAVIF())
		w.Close()
		req := httptest.NewRequest("POST", path+"/artwork/desktop", &b)
		req.Header.Set("Content-Type", w.FormDataContentType())
		response, e := app.Test(req)
		if e != nil {
			t.Fatal(e)
		}
		defer response.Body.Close()
		var envelope integrationEnvelope
		json.NewDecoder(response.Body).Decode(&envelope)
		if response.StatusCode != status {
			t.Fatalf("upload status %d: %s", response.StatusCode, envelope.Data)
		}
		return envelope.Data
	}
	primary.fail = true
	upload(502)
	var key *string
	if e := pools["game"].QueryRow(ctx, `SELECT desktop_object_key FROM gfg_showcase_campaign WHERE id=$1`, id).Scan(&key); e != nil || key != nil {
		t.Fatal("Primary failure changed reference", e)
	}
	primary.fail = false
	data := upload(200)
	if !strings.Contains(string(data), "R2 mirror sync failed") {
		t.Fatal("Mirror warning missing", string(data))
	}
	call("POST", path+"/publish", `{}`, 200)
	call("DELETE", path+"/artwork/desktop", ``, 400)
	call("PUT", path+"/content", strings.Replace(content, "完整摘要", "", 1), 400)
	call("DELETE", path, ``, 400)
	if redisAddr != "" {
		if redisClient.Get(ctx, "game:v2:showcase:revision").Val() == "" {
			t.Fatal("committed writes did not bump revision")
		}
		command := exec.CommandContext(ctx, "go", "test", "./apps/game/v2/controller", "-run", "^TestShowcaseCrossServiceHTTPClosure$", "-count=1", "-v")
		command.Dir = filepath.Join(integrationRepositoryRoot(t), "apps", "cn", "game-backend")
		command.Env = append(os.Environ(), "GOFURRY_SHOWCASE_E2E_GFG_URL="+dsns["game"], "GOFURRY_SHOWCASE_E2E_CAMPAIGN_ID="+id)
		if out, e := command.CombinedOutput(); e != nil {
			t.Fatalf("Game HTTP closure: %v\n%s", e, out)
		} else {
			t.Log(string(out))
		}
		stats := call("GET", path+"/stats", ``, 200)
		if !strings.Contains(string(stats), `"qualified_clicks":1`) {
			t.Fatal("Admin cannot read Game aggregates", string(stats))
		}
		response := requestJSON(t, app, "GET", path+"/stats.csv", ``, nil, 200)
		if !strings.Contains(response.Header.Get("Content-Type"), "text/csv") {
			t.Fatal("CSV content type")
		}
		response.Body.Close()
	}
	conflict := create("pin conflict")
	call("PUT", "/campaigns/"+conflict+"/content", content, 200)
	call("PUT", "/campaigns/"+conflict+"/schedule", schedule, 200)
	if _, e := pools["game"].Exec(ctx, `UPDATE gfg_showcase_campaign SET desktop_object_key='game/showcase/'||id||'/desktop/'||repeat('a',32)||'.avif' WHERE id=$1`, conflict); e != nil {
		t.Fatal(e)
	}
	call("POST", "/campaigns/"+conflict+"/publish", `{}`, 400)
	// The same pin is valid in a disjoint locale; a published edit cannot
	// silently turn that into an overlapping contract.
	call("PUT", "/campaigns/"+conflict+"/content", strings.Replace(content, `"lang":"zh"`, `"lang":"en"`, 1), 200)
	call("POST", "/campaigns/"+conflict+"/publish", `{}`, 200)
	call("PUT", "/campaigns/"+conflict+"/content", content, 400)
	call("POST", "/campaigns/"+conflict+"/pause", `{}`, 200)
	call("PUT", "/campaigns/"+conflict+"/content", content, 200)
	later := fmt.Sprintf(`{"starts_at":%q,"ends_at":%q,"weight":100,"pin_position":1}`, time.Now().Add(2*time.Hour).UTC().Format(time.RFC3339), time.Now().Add(3*time.Hour).UTC().Format(time.RFC3339))
	call("PUT", "/campaigns/"+conflict+"/schedule", later, 200)
	call("POST", "/campaigns/"+conflict+"/resume", `{}`, 200)
	call("PUT", "/campaigns/"+conflict+"/schedule", schedule, 400)
	call("POST", "/campaigns/"+conflict+"/archive", `{}`, 200)
	// Sponsored pins conflict even when their physical positions differ.
	for i := 2; i <= 3; i++ {
		sponsor := create("sponsored pin")
		call("PUT", "/campaigns/"+sponsor+"/content", strings.Replace(content, `"sponsored":false`, `"sponsored":true`, 1), 200)
		call("PUT", "/campaigns/"+sponsor+"/schedule", strings.Replace(schedule, `"pin_position":1`, fmt.Sprintf(`"pin_position":%d`, i), 1), 200)
		if _, e := pools["game"].Exec(ctx, `UPDATE gfg_showcase_campaign SET desktop_object_key='game/showcase/'||id||'/desktop/'||repeat('a',32)||'.avif' WHERE id=$1`, sponsor); e != nil {
			t.Fatal(e)
		}
		status := 200
		if i == 3 {
			status = 400
		}
		call("POST", "/campaigns/"+sponsor+"/publish", `{}`, status)
	}
	call("POST", path+"/pause", `{}`, 200)
	call("POST", path+"/resume", `{}`, 200)
	// GFA failure prevents both GFG commit and the post-commit revision side effect.
	if _, e := pools["admin"].Exec(ctx, `CREATE FUNCTION showcase_test_audit_failure() RETURNS trigger LANGUAGE plpgsql AS $$BEGIN RAISE EXCEPTION 'test audit failure'; END$$; CREATE TRIGGER showcase_test_audit_failure BEFORE INSERT ON gfa_admin_audit_log FOR EACH ROW EXECUTE FUNCTION showcase_test_audit_failure()`); e != nil {
		t.Fatal(e)
	}
	revision := ""
	if redisAddr != "" {
		revision = redisClient.Get(ctx, "game:v2:showcase:revision").Val()
	}
	call("POST", path+"/pause", `{}`, 500)
	var state string
	pools["game"].QueryRow(ctx, `SELECT state FROM gfg_showcase_campaign WHERE id=$1`, id).Scan(&state)
	if state != "published" {
		t.Fatal("Audit failure committed business state")
	}
	if redisAddr != "" && redisClient.Get(ctx, "game:v2:showcase:revision").Val() != revision {
		t.Fatal("uncommitted mutation bumped revision")
	}
	pools["admin"].Exec(ctx, `DROP TRIGGER showcase_test_audit_failure ON gfa_admin_audit_log; DROP FUNCTION showcase_test_audit_failure()`)
	// Redis failure must preserve committed GFG/Audit success.
	if redisAddr != "" {
		redisClient.Close()
	}
	call("POST", path+"/archive", `{}`, 200)
	var audits int64
	if e := pools["admin"].QueryRow(ctx, `SELECT count(*) FROM gfa_admin_audit_log WHERE resource='gfg_showcase_campaign' AND target_id=$1`, id).Scan(&audits); e != nil || audits < 7 {
		t.Fatal("missing audits", audits, e)
	}
	draft := create("delete draft")
	call("DELETE", "/campaigns/"+draft, ``, 200)
	saved := principal.Capabilities
	principal.Capabilities = nil
	call("GET", "/campaigns", ``, 403)
	principal.Capabilities = saved
	// Omission preserves stored eligibility through the existing classification API.
	game := call("POST", "/games", `{"name":"Game","name_en":"Game","info":"Info","info_en":"Info","resources":[],"groups":[],"developers":[],"publishers":[],"appid":119119,"header":"","links":[],"weight":0}`, 200)
	var gameDTO struct {
		ID int64 `json:"id"`
	}
	json.Unmarshal(game, &gameDTO)
	gamePath := "/games/" + strconv.FormatInt(gameDTO.ID, 10) + "/classification"
	call("PUT", gamePath, `{"weight":0,"tag_ids":[],"showcase_eligible":true}`, 200)
	call("PUT", gamePath, `{"weight":0,"tag_ids":[]}`, 200)
	var eligible bool
	if e := pools["game"].QueryRow(ctx, `SELECT showcase_eligible FROM gfg_game WHERE id=$1`, gameDTO.ID).Scan(&eligible); e != nil || !eligible {
		t.Fatal("omission reset eligibility", e)
	}
	// Adult semantics follow code, not the historical ID. Removing the
	// classification permits the exact same otherwise-valid publication.
	if _, e := pools["game"].Exec(ctx, `INSERT INTO gfg_tag_category(id,code,name,name_en,info,info_en,sort_order,create_time,update_time) VALUES(119,'test','Test','Test','','',0,now(),now());
INSERT INTO gfg_tag(id,code,category_id,name,name_en,info,info_en,create_time,update_time) VALUES(119901,'adult',119,'成人','Adult','','',now(),now())`); e != nil {
		t.Fatal(e)
	}
	if _, e := pools["game"].Exec(ctx, `INSERT INTO gfg_game_tag(game_id,tag_id,role,create_time,update_time) VALUES($1,119901,'normal',now(),now())`, gameDTO.ID); e != nil {
		t.Fatal(e)
	}
	linked := create("linked Game")
	linkedContent := strings.Replace(content, `"primary_action_type":"website","primary_target":"https://example.test/game"`, fmt.Sprintf(`"linked_game_id":%d,"primary_action_type":"game"`, gameDTO.ID), 1)
	call("PUT", "/campaigns/"+linked+"/content", linkedContent, 200)
	call("PUT", "/campaigns/"+linked+"/schedule", strings.Replace(schedule, `"pin_position":1`, `"pin_position":null`, 1), 200)
	if _, e := pools["game"].Exec(ctx, `UPDATE gfg_showcase_campaign SET desktop_object_key='game/showcase/'||id||'/desktop/'||repeat('a',32)||'.avif' WHERE id=$1`, linked); e != nil {
		t.Fatal(e)
	}
	call("POST", "/campaigns/"+linked+"/publish", `{}`, 400)
	call("PUT", gamePath, `{"weight":0,"tag_ids":[]}`, 200)
	call("POST", "/campaigns/"+linked+"/publish", `{}`, 200)
}
