package bootstrap_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	navadmin "github.com/gofurry/gofurry-admin/internal/app/navadmin/controller"
	"github.com/gofurry/gofurry-admin/internal/app/shared/audit"
)

func TestReleaseNoteAdminLifecycle(t *testing.T) {
	config := os.Getenv("GOFURRY_ADMIN_INTEGRATION_CONFIG")
	if config == "" {
		t.Skip("set GOFURRY_ADMIN_INTEGRATION_CONFIG for PostgreSQL integration tests")
	}
	base := adminIntegrationDSN(t, config)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	db := adminIntegrationSQLDB(t, base, "postgres")
	defer db.Close()
	navName, adminName := integrationDatabaseName("gfn"), integrationDatabaseName("gfa")
	for _, name := range []string{navName, adminName} {
		createIntegrationDatabase(t, ctx, db, name)
		defer dropIntegrationDatabase(t, db, name)
	}
	navDSN, adminDSN := integrationDatabaseDSN(base, navName), integrationDatabaseDSN(base, adminName)
	applyIntegrationBaseline(t, "nav", navDSN)
	applyIntegrationBaseline(t, "admin", adminDSN)
	navPool, adminPool := integrationPool(t, ctx, navDSN), integrationPool(t, ctx, adminDSN)
	defer navPool.Close()
	defer adminPool.Close()
	api := navadmin.New(navPool, audit.New(adminPool))
	app := fiber.New()
	app.Post("/notes", api.CreateUpdateNotice)
	app.Get("/notes", api.ListUpdateNotices)
	app.Get("/notes/:id", api.GetUpdateNotice)
	app.Put("/notes/:id", api.UpdateUpdateNotice)
	app.Delete("/notes/:id", api.DeleteUpdateNotice)
	app.Post("/notes/:id/publish", api.PublishUpdateNotice)
	app.Post("/notes/:id/unpublish", api.UnpublishUpdateNotice)
	request := func(method, path, body string, status int) map[string]any {
		t.Helper()
		resp := requestJSON(t, app, method, path, body, nil, status)
		defer resp.Body.Close()
		var envelope integrationEnvelope
		if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
			t.Fatal(err)
		}
		var data map[string]any
		if len(envelope.Data) > 0 {
			if err := json.Unmarshal(envelope.Data, &data); err != nil {
				t.Fatal(err)
			}
		}
		return data
	}
	draft := request("POST", "/notes", `{"publication_state":"published"}`, 200)
	id := int64(draft["id"].(float64))
	path := fmt.Sprintf("/notes/%d", id)
	if draft["publication_state"] != "draft" || draft["published_at"] != nil {
		t.Fatalf("new draft=%+v", draft)
	}
	request("POST", path+"/publish", "", 400)
	request("PUT", path, `{"title":"标题"}`, 200)
	request("POST", path+"/publish", `{}`, 400)
	saved := request("PUT", path, `{"title":"标题","body":"# 正文","summary":"摘要","commit_sha":" ABCDEF0 ","version":"September preview","publication_state":"published"}`, 200)
	if saved["publication_state"] != "draft" || saved["commit_sha"] != "abcdef0" {
		t.Fatalf("content save=%+v", saved)
	}
	published := request("POST", path+"/publish", "", 200)
	if published["publication_state"] != "published" || published["published_at"] == nil {
		t.Fatalf("publish now=%+v", published)
	}
	if queryInt64(t, ctx, navPool, `SELECT count(*) FROM gfn_nav_update_notice WHERE id=$1 AND published_at <= (CURRENT_TIMESTAMP AT TIME ZONE 'Asia/Shanghai') AND published_at > (CURRENT_TIMESTAMP AT TIME ZONE 'Asia/Shanghai') - interval '1 minute'`, id) != 1 {
		t.Fatal("Publish Now does not use authoritative China wall time")
	}
	at := published["published_at"].(string)
	saved = request("PUT", path, fmt.Sprintf(`{"title":"已修改","body":"正文","published_at":%q,"publication_state":"draft"}`, at), 200)
	if saved["publication_state"] != "published" {
		t.Fatal("content PUT changed publication state")
	}
	request("PUT", path, `{"title":"","body":""}`, 400)
	unpub := request("POST", path+"/unpublish", "", 200)
	if unpub["publication_state"] != "draft" || unpub["published_at"] != at {
		t.Fatal("unpublish must preserve publication timestamp")
	}
	scheduled := request("POST", path+"/publish", `{"published_at":"2099-01-01T12:00:00+08:00"}`, 200)
	if scheduled["publication_state"] != "published" || scheduled["published_at"] != "2099-01-01 12:00:00" {
		t.Fatalf("schedule=%+v", scheduled)
	}
	if queryInt64(t, ctx, navPool, `SELECT count(*) FROM gfn_nav_update_notice WHERE id=$1 AND published_at > (CURRENT_TIMESTAMP AT TIME ZONE 'Asia/Shanghai')`, id) != 1 {
		t.Fatal("future date not persisted")
	}
	request("POST", path+"/publish", `{"published_at":"invalid"}`, 400)
	request("POST", "/notes", `{"commit_sha":"not a sha"}`, 400)
	request("DELETE", path, "", 200)
	if queryInt64(t, ctx, navPool, `SELECT count(*) FROM gfn_nav_update_notice WHERE id=$1 AND deleted`, id) != 1 {
		t.Fatal("delete must remain soft")
	}
	for _, action := range []string{"create", "update", "publish", "unpublish", "delete"} {
		if queryInt64(t, ctx, adminPool, `SELECT count(*) FROM gfa_admin_audit_log WHERE resource='gfn_nav_update_notice' AND action=$1 AND target_id=$2`, action, fmt.Sprint(id)) == 0 {
			t.Fatalf("missing audit %s", action)
		}
	}
	if queryInt64(t, ctx, adminPool, `SELECT count(*) FROM gfa_admin_audit_log WHERE action='publish' AND resource='gfn_nav_update_notice'`) != 2 {
		t.Fatal("rejected publications were audited as successful")
	}
	// Audit failure must leave the GFN mutation uncommitted.
	_, err := adminPool.Exec(ctx, `ALTER TABLE gfa_admin_audit_log RENAME TO release_notes_audit_unavailable`)
	if err != nil {
		t.Fatal(err)
	}
	request("POST", "/notes", `{"title":"Must roll back"}`, 500)
	if queryInt64(t, ctx, navPool, `SELECT count(*) FROM gfn_nav_update_notice`) != 1 {
		t.Fatal("audit failure committed GFN draft")
	}
}
