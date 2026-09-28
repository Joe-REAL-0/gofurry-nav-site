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

## P2: Admin Release Workspace and Markdown authoring — 2026-09-28

P2 replaces the generic `update-notices` resource with dedicated list/new/detail
routes before `/nav/:resource`, using the existing AppShell navigation and
ContentRead/ContentWrite capabilities. P1 storage, API and lifecycle semantics
are unchanged; no Backend defect or database migration was required.

The list supports API search/pagination, row/link navigation, release metadata and
centralized Draft/Scheduled/Published presentation in China-site time. One RHF/Zod
form retains both languages plus shared metadata. Create is deferred until Save
or confirmed Publish. Dirty publication awaits persistence; failed saves never
publish. Immediate publication omits the timestamp. Scheduling sends the chosen
future China time. Unpublish is disabled while dirty; confirmed Delete is separate.
Both browser unload and React Router navigation protect unsaved work, including
when write capability is lost after editing. Successful saves clear the guard.

Admin alone adds `markdown-it` 15.0.2, `sanitize-html` 2.17.7 and typings through
pnpm. The [shared Markdown contract](../../contracts/update-markdown.md) and its
37 semantic/security fixture cases govern headings, breaks, formatting, lists,
code, URL policies, image attributes, literal HTML and the final sanitizer.
The toolbar edits textarea selections without rewriting stored source; Desktop
uses an editor/preview grid and narrow layouts use a local switch. Preview styling
uses Admin tokens and has no public visual ownership.

### Local verification — passed

From `apps/cn/admin/react`:

```text
pnpm install --frozen-lockfile
pnpm run lint
pnpm run typecheck
pnpm test
pnpm run build
```

- **29 files / 219 tests passed** in the final full Admin run.
- New feature coverage: 18 workspace/lifecycle/guard tests, 3 model/time tests,
  and 48 Markdown/toolbar tests (including all 37 shared fixture cases).
- Existing router/generic tests now prove dedicated Release Notes ownership;
  all other Admin tests remain in the full run.
- Frozen install preserved the generated pnpm lockfile; Nav Web dependencies and
  lockfile were untouched. No build-script permission changes were needed.
- Lint exits successfully with the existing shared/routing rule warning patterns;
  the new feature files produce no lint warnings. The build retains the existing
  large shared-chunk advisory and emits the normal embedded Admin frontend.
- `git diff --check` passes. No Public Browser, Nav Web Visual, backend regression
  or migration run was performed for this frontend-only phase.

P2 is locally ready for P3. P3 was not started. This record does not claim remote
CI or maintainer visual acceptance for P2; normal Admin page review can accompany
the final #132 acceptance.
