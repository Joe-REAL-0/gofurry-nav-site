# Issue #132 — Release Notes acceptance

## P1: domain and API foundation — 2026-09-28

Implemented on `dev`. P2/P3 UI is outside this acceptance; the current Timeline
and generic Admin editor are unchanged. The [domain contract](../release-notes.md)
records the transitional index body and draft-only create behavior. P1 must not
be independently deployed as the completed Release Notes product.

### Development migration

- The configured shared development GFN was reached from the workstation using
  `gofurry_migrator`; no credentials/addresses are recorded here.
- Preflight found **zero** non-deleted future-dated legacy notices.
- Pinned Goose `status → up → status` applied only the new migration
  `20260928010000_nav_update_release_notes.sql` successfully.
- All **26** legacy rows were compared before and after migration. IDs, both
  titles/bodies, publication/create/update timestamps and deleted flags were
  unchanged. All rows became published; version/SHA remained null and summaries
  empty. No notice was edited for this acceptance.
- The PostgreSQL session uses UTC. Release publication uses an explicit China-site
  database clock; no global timezone or timestamp-column migration was made.
- Read-only schema comparison on the migrated shared GFN passed against
  `tools/db-baseline/expected-final/gfn.json` using the existing schema inspector.

### Focused executable coverage

- `tools/db-baseline/TestReleaseNotesMigrationPreservesLegacyAndGuardsDown`:
  legacy-row preservation, nullable draft date, default draft, metadata/state
  constraints, null-date Down rejection before schema change, safe Down/reapply.
- Existing fresh/baseline/upgrade acceptance scoped to GFN; regenerated only the
  final GFN snapshot. Historical schema snapshots/migrations remain unchanged.
- Nav updates service/controller: independent localization, transitional index
  body, explicit detail projection, 404/503 distinction, absent neighbors.
- Nav PostgreSQL persistence owner: public order, tied dates, bounded older/newer
  neighbors skipping hidden records, draft/scheduled/deleted/null-date visibility,
  and session-timezone independence, using a rolled-back isolated fixture.
- Admin unit tests: permissive draft, optional arbitrary version, normalized SHA,
  invalid metadata/date, Chinese publication requirements and China wall dates.
- Admin PostgreSQL API lifecycle: forced draft create, content-save state isolation,
  Publish Now/scheduling/unpublish, timestamp preservation, validation, soft delete,
  five distinct audit actions, and rollback when audit storage is unavailable.
- Production route registration: public index/detail reach their handlers; Admin
  publish/unpublish enforce ContentWrite, reject readers and require authentication.

### Local verification — passed

- `task generate:sqlc` and `task check:sqlc` (vet plus generated-code comparison).
- Nav Backend: `go test ./apps/nav/updates/... ./routers`, plus the existing
  `TestPostgresNavBackendPersistenceSemantics` with its new Release Notes fixture.
- Admin: `go test ./internal/app/navadmin/... ./internal/transport/http/router`,
  and `go test ./internal/bootstrap -run TestReleaseNoteAdminLifecycle -count=1`
  with isolated PostgreSQL integration configuration.
- Tools: `TestReleaseNotesMigrationPreservesLegacyAndGuardsDown`,
  `TestPostgresFreshAndBaselineAdoption/gfn`, and read-only development
  `TestManagedAssetsDevSchema` (full final-schema comparison).
- Admin React: `pnpm run typecheck`; no frontend source changes were necessary.
- Admin and Nav Backend package compilation (`go test ./... -run '^$'`), plus
  Nav Collector's regenerated sqlc package compilation; these are compilation
  checks, not claims of a full module test run.
- Production policy tests/checker and `git diff --check`.

Tests use a disposable local PostgreSQL 18 container, not shared development data.
Remote CI has not been run for this change; no remote acceptance is claimed.
P1 establishes the APIs required by P2; it does not claim P2/P3 completion.
