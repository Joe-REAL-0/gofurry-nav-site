package main

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/gofurry/gofurry-nav-site/tools/internal/schema"
	"github.com/pressly/goose/v3"
)

// This gate reads committed final snapshots only: no environment, DB or writes.
func TestExpectedFinalReadabilityPolicy(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	for _, database := range []string{"gfa", "gfg", "gfn"} {
		t.Run(database, func(t *testing.T) {
			snapshot, err := loadFinalExpected(root, database)
			if err != nil {
				t.Fatal(err)
			}
			assertReadability(t, snapshot)
		})
	}
}

var readabilityBackfills = []struct {
	database, domain string
	previous         int64
}{
	{"gfa", "admin", 20260926010000},
	{"gfg", "game", 20261003010000},
	{"gfn", "nav", 20260928010000},
}

func TestReadabilityBackfillStatements(t *testing.T) {
	// These particular backfills deliberately use one COMMENT per line. Reject
	// any other executable SQL, even if its net schema effect happens to cancel.
	comment := regexp.MustCompile(`^COMMENT ON (TABLE|COLUMN|FUNCTION|TRIGGER) [^;]+ IS '(?:[^']|'')*';$`)
	for _, migration := range readabilityBackfills {
		t.Run(migration.database, func(t *testing.T) {
			path := filepath.Join("../../db", migration.domain, "migrations", "20261006010000_"+migration.domain+"_schema_readability.sql")
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			for i, line := range strings.Split(string(data), "\n") {
				line = strings.TrimSpace(line)
				if line == "" || strings.HasPrefix(line, "--") {
					continue
				}
				if !comment.MatchString(line) {
					t.Errorf("%s:%d: not a metadata COMMENT", path, i+1)
				}
			}
		})
	}
}

func TestReadabilityBackfillsPreserveSchema(t *testing.T) {
	dsn := integrationAdminDSN(t)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	admin := openDatabase(t, dsn, "postgres")
	defer admin.Close()
	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatal(err)
	}
	for _, migration := range readabilityBackfills {
		t.Run(migration.database, func(t *testing.T) {
			name := temporaryDatabaseName(migration.database, "readability")
			createDatabase(t, ctx, admin, name)
			defer dropDatabase(t, admin, name)
			db := openDatabase(t, dsn, name)
			defer db.Close()
			dir, err := filepath.Abs(filepath.Join("../../db", migration.domain, "migrations"))
			if err != nil {
				t.Fatal(err)
			}
			if err = goose.UpToContext(ctx, db, dir, migration.previous); err != nil {
				t.Fatal(err)
			}
			before, err := schema.Inspect(ctx, db)
			if err != nil {
				t.Fatal(err)
			}
			if err = goose.UpToContext(ctx, db, dir, 20261006010000); err != nil {
				t.Fatal(err)
			}
			after, err := schema.Inspect(ctx, db)
			if err != nil {
				t.Fatal(err)
			}
			assertReadability(t, after)
			if diff := schema.Difference(withoutReadabilityComments(before), withoutReadabilityComments(after)); diff != "" {
				t.Fatalf("backfill changed more than comments: %s", diff)
			}
		})
	}
}

func withoutReadabilityComments(snapshot schema.Snapshot) schema.Snapshot {
	for i := range snapshot.Tables {
		table := &snapshot.Tables[i]
		table.Comment = ""
		for j := range table.Columns {
			table.Columns[j].Comment = ""
		}
		for j := range table.Triggers {
			table.Triggers[j].Comment = ""
		}
	}
	for i := range snapshot.Functions {
		snapshot.Functions[i].Comment = ""
	}
	return snapshot
}

func assertReadability(t *testing.T, snapshot schema.Snapshot) {
	t.Helper()
	findings := schema.ValidateReadability(snapshot)
	for _, finding := range findings {
		t.Errorf("%s %s: %s", finding.ObjectKind, finding.ObjectName, finding.Reason)
	}
	if len(findings) > 0 {
		t.FailNow()
	}
}
