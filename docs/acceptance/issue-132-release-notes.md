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

## P3: Public Release Notes and final closure — 2026-09-28

This final phase supersedes P1's transitional index body and the historical Updates
Timeline contract. P1/P2 stage records above retain their original evidence. There
is no P4, new lifecycle, schema migration, permission, media library or upload.

### Final ownership

- `/updates` and `/en/updates`: editorial Latest emphasis, optional metadata,
  complete returned history grouped by China-site month; no body-derived summary,
  accordion, client pagination or load more.
- `/updates/:id` and `/en/updates/:id`: SSR sanitized Markdown, optional release
  metadata, full-SHA GitHub link, older `previous` / newer `next` links, locale
  identity preservation and authoritative HTTP 404/503.
- Public list contains exactly six fields: `id/title/summary/version/commit_sha/
  published_at`. Detail continues to own body. SQL/sqlc/lifecycle are unchanged.
- `updateMarkdown.ts` and `UpdateMarkdown.vue` own the sole safe HTML boundary;
  all 37 shared P2 Markdown/security cases pass using exactly markdown-it 15.0.2
  and sanitize-html 2.17.7. Admin implementation/dependencies remain unchanged.
- Existing `useLocaleHead` owns canonical/hreflang. Detail uses localized title,
  summary/fallback description, article OG URL/type and publication metadata.
  No JSON-LD or body-derived SEO is added.
- Sitemap adds both locales from the public Chinese Release Notes inventory;
  unavailable/malformed inventory fails closed. The existing 100-item endpoint
  cap remains a documented future inventory boundary.

### Proven P1 defect corrected

A focused regression first reproduced an eight-hour publication display shift:
pgx's zone-less timestamp fields were serialized with UTC `Z`, despite the domain
using China wall time. The shared public projection now attaches UTC+8 without
shifting calendar fields. Index/detail/neighbors return the same correct instant.
The regression passed after the fix. No stored date, SQL predicate, ordering,
Admin publication behavior or migration changed.

### Retirement / style audit

Consumer-audited Timeline Summary/Entry/YearGroup components, both divider SVGs,
year/load-more runtime, obsolete date helpers, unused service re-export, Timeline
CSS and Browser/Visual assumptions are deleted. No hidden Timeline clone remains.
Unrelated Insight timelines and the global NavBar/Footer/background stay intact.
`frontend-style-debt.json` and ESLint suppressions are unchanged: new Updates debt
is zero; only existing ambient raw 75 and #108 important 5 remain.

### Visual inventory

`tests/browser/visual/updates-page.spec.ts`, using the shared deterministic
`fixtures/updates.ts`, owns exactly these eight root captures:

- `updates-light-desktop.png`
- `updates-dark-desktop.png`
- `updates-light-mobile.png`
- `updates-dark-mobile.png`
- `update-detail-light-desktop.png`
- `update-detail-dark-desktop.png`
- `update-detail-light-mobile.png`
- `update-detail-dark-mobile.png`

Viewports are 1440×900 and 390×900. The complete root capture includes the article's
Markdown, local image and fixed neighbors. Local time/data/media, fresh CDN
routing diagnostics, production Nitro SSR/hydration, semantic request accounting,
fonts/images/finite-animation readiness, focus blur and no-overflow checks are
shared across the test owner. No external service, arbitrary sleep or retry is
used. Only unrelated fixed tools are hidden as in the prior Visual owner.

Generation was scoped to this spec in the digest-pinned Linux/Node 24/Playwright
1.60.0 image. Two subsequent independent scoped compare runs passed (8 each).
Only these four replaced Index PNGs and four new Article PNGs change; no unrelated
golden is updated. Automated comparison is not maintainer visual approval.

### Local verification

Passed from Nav Web:

```text
pnpm install --frozen-lockfile
pnpm run lint
pnpm run stylelint
pnpm run style:policy:test
pnpm run style:policy
pnpm run typecheck
pnpm run test:unit
pnpm run test:nuxt
pnpm run insights:semantics
pnpm run seo:recovery:test
pnpm run build
```

- Unit: 16 files / 357 tests, including 37 shared Markdown cases plus metadata,
  calendar grouping, SEO/status and fail-closed inventory coverage.
- Nuxt: 6 files / 21 tests, including the real sanitized component boundary and
  reactive Markdown replacement.
- Policy tooling: 75 tests; measured debt unchanged.
- Focused Updates Browser: 24 cases covering locale/theme/device, single SSR API
  calls without hydration duplication, navigation, legacy optional metadata,
  local error recovery, 404/503 and real locale switching.
- Focused SEO Browser: three sitemap cases cover canonical release inventory and
  failures of existing/Updates inventory; all pass.
- Backend: `go test ./apps/nav/updates/... ./routers -count=1` passes, including
  exact public projection and the reproduced publication-time defect. The Nav
  persistence test owner also compiles (`go test ./apps/nav/navPage/dao -run '^$'`);
  its body assertion now correctly belongs to detail, not the retired list body.
  No claim of a new local PostgreSQL integration run is made. No SQL/sqlc changed.

No complete repository Browser suite or unrelated Visual update was run locally.
Build retains existing third-party chunk/circular dependency advisories.

### Remote and maintainer gates

Fast, Manual Full Regression and Manual Visual are separate #134 gates. Local
results do not establish remote acceptance. Dispatch/current-code status must be
checked in Actions; this committed record does not predeclare remote success.

Maintainer final review is **pending**, intentionally combined across P1–P3:

- Admin: Draft, zh/en, toolbar/preview, Save, Publish Now, Schedule, Unpublish.
- Index: Timeline absent, Latest hierarchy, history density/months, both themes
  and Desktop/Mobile.
- Detail: reading width, Markdown typography/code/quote/image, metadata/commit,
  older/newer and locale navigation.

Implementation is ready for this combined review. #132 final closure still
requires the selected current-code remote gates and explicit maintainer approval;
no automatic closure or P4 is authorized by this record.

## P3.1 — compact writing and reading refinement (2026-09-28)

This final refinement supersedes P2's toolbar and P3's all-history index layout;
the earlier phase records above remain historical evidence.

- Admin combines shared metadata and zh/en content in one Section. The unused
  toolbar implementation and its transform tests are deleted after consumer audit.
  Equal Desktop editor/preview panes become a local Edit/Preview switch on narrow
  screens. Markdown rendering, authorization, dirty guards and every publication
  mutation are unchanged.
- Public Index removes explanatory header copy, tightens Latest and month rows,
  and SSR-renders only Latest plus 20 history items. `useUpdateIndex` owns explicit
  page appends, pending-click protection, error/retry and locale/unmount races.
- The smallest public index extension adds page/page_size and total/has_more.
  SQL keeps publication visibility and descending date/ID ordering. The default
  unpaged 100-item sitemap request remains intact. No new migration, lifecycle,
  Admin API, permission or media integration is introduced. Deploy Nav Backend
  together with this paginated Nav Web client; rebuild Admin for its embedded UI.
- Detail removes the redundant release label, uses a 760px maximum body and
  quieter Previous/Next links. Previous still means older; Next still means newer.
  Markdown fixture/renderer, canonical/hreflang and authoritative 404 are unchanged.

### Focused local evidence

- Admin lint (existing unrelated warnings only), typecheck and embedded build pass.
  Release Notes tests: 3 files / 58 cases, preserving lifecycle/security coverage.
- Nav Web lint, typecheck, stylelint and style policy pass. The debt manifest is
  unchanged: no Updates debt, ambient raw 75 and Insights important 5 retained;
  deep selectors and legacy dark entries remain zero.
- Focused Markdown/presentation/Nuxt tests: 4 files / 62 cases. New pagination
  cases cover bounded first requests, explicit appends, overlap deduplication,
  concurrent clicks, failed-page retry, locale race and unmount cancellation.
- Production build passes. Updates Browser: 27 cases; focused sitemap Browser:
  3 cases. New desktop/mobile cases prove first-page SSR contains exactly Latest
  plus 20 history entries, no hidden all-data load, and one request per action.
- Updates service/controller and router tests pass. SQL is regenerated; sqlc vet
  and `go run ./check-sqlc` pass. `TestPostgresNavBackendPersistenceSemantics`
  passes against disposable local PostgreSQL 18, including public-only totals,
  page boundaries and stable timestamp-tie ordering. Shared/production data is
  untouched; Goose ran only as part of that disposable test database setup.
- Exactly the existing eight Updates goldens were regenerated in the digest-pinned
  Linux runner, followed by two independent scoped compare passes (8 each).
  Unrelated PNGs and the shared shell are unchanged.
- Temporary Admin browser review uses mocked read-only API responses and the
  production build. It verifies equal Desktop panes, narrow preview/language
  switching, no overflow and no browser errors. At 390px it uses the existing
  sidebar-collapse control; the shared Admin shell was not redesigned.

### Final review gate

Current-code Fast CI and one Manual Visual run are required after push; this
record does not claim their completion. No additional Manual Full is dispatched
for this focused refinement. Prior P3 runs are not P3.1 acceptance.

Maintainer approval remains **pending**. Review Admin writing/preview and existing
Save/Publish/Schedule/Unpublish controls; public compact Latest/month density and
More loading; article typography, both themes/devices and Previous/Next wording.
The eight scoped goldens plus temporary Admin review captures support that review.
P3.1 proceeds directly to #132 final maintainer review; no P4 is planned or created.
