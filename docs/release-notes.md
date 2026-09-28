# Release Notes — Issue #132

P1 establishes the GFN domain and public/Admin API contracts. P2 provides the
Admin Release Workspace and safe Markdown authoring. P3 completes the editorial
public index and SSR articles, retires the Timeline and transitional list body,
and owns SEO/sitemap plus Functional/Visual acceptance. Final remote and maintainer
review status is recorded separately in the [acceptance ledger](acceptance/issue-132-release-notes.md).

## Storage and publication

`db/nav/migrations/20260928010000_nav_update_release_notes.sql` extends
`gfn_nav_update_notice`. Existing rows, including deleted rows, are backfilled
to `publication_state=published` without changing their previous fields.

- `version`: optional arbitrary label, at most 64 characters; no SemVer rule.
- `commit_sha`: optional, trimmed/lowercased hexadecimal string of 7–64 characters.
- `summary` / `summary_en`: non-null text, default empty.
- `publication_state`: only `draft` / `published`, new records default to draft.
- `published_at`: nullable China-site wall timestamp, retaining the existing type.
- `body` / `body_en`: Markdown source-compatible text; no rendering or rewriting.

Scheduled is derived from published state plus a future timestamp; it is never
stored as a third state. No scheduler, Redis invalidation or background worker is
needed. Public SQL captures one database clock per request using
`CURRENT_TIMESTAMP AT TIME ZONE 'Asia/Shanghai'`. Admin parses unzoned values in
the same zone and converts offset-bearing values to it. This avoids dependence
on the PostgreSQL session timezone without a repository-wide timestamp change.
P3 corrects one P1 serialization defect: pgx returns zone-less calendar fields;
the public projection now attaches their China-site UTC+8 offset instead of
mislabeling them `Z`. Index, detail and neighbors share that projection. Stored
values, SQL visibility and publication ordering are unchanged.

Only non-deleted, published records with a non-null timestamp at or before that
clock are public. Draft/scheduled/deleted/missing detail requests all return 404.

## Public API

Both endpoints retain the existing outer `{code,data}` envelope and `lang=zh|en`.
Title, summary and body independently fall back to the other language; version
and SHA are not localized.

- `GET /api/v2/nav/updates`: existing schema-version-1 ready/empty/error envelope,
  ordered by `published_at DESC, id DESC` (existing limit 100). Items contain only
  `id`, `title`, `summary`, `version`, `commit_sha`, `published_at`. P3 removes the
  P1 transitional `body`, `create_time` and `update_time` fields. Detail owns body.
- `GET /api/v2/nav/updates/:id`: schema version 1, generated time, `state=ready`,
  `item`, `previous`, `next`. Item contains only ID, localized title/summary/body,
  version, SHA and publication time. Neighbors contain ID/title/version/time or
  null. Previous is immediately older and next immediately newer, with ID breaking
  timestamp ties. Each neighbor is a bounded SQL query, applies the same visibility
  predicate and uses the selected release's request clock. Internal lifecycle,
  deleted and audit/creation fields are excluded. Store failure returns 503;
  the existing index error envelope is unchanged.

## Admin API

All mutations retain existing `content.write` authorization and GFA Audit through
the explicit GFN pool. Content update, publish, unpublish and delete lock the
active row before taking the before/after audit snapshots.

| Endpoint under `/api/v1/nav/update-notices` | Behavior |
| --- | --- |
| `POST /` | Always creates draft; partial/empty content and null date accepted |
| `GET /`, `GET /:id` | Include version/SHA/summaries/publication state and nullable date |
| `PUT /:id` | Replaces content/metadata/date, never changes publication state |
| `POST /:id/publish` | Requires Chinese title/body; optional `published_at`; absent/null/empty means database Publish Now |
| `POST /:id/unpublish` | Returns to draft, preserving the publication timestamp |
| `DELETE /:id` | Remains soft deletion |

Publication state is read-only in create/PUT payloads; attempts to set it cannot
publish or unpublish. English, summaries and version are optional. Published
content saves must retain Chinese title/body and a publication timestamp;
unpublish first to return to permissive draft editing. The response exposes the
stored state only; clients must not interpret `published` alone as public visibility.

Audit actions are `create`, `update`, `publish`, `unpublish`, `delete`. Validation
failures do not produce successful mutation audits. Existing cross-database
semantics remain: an audit failure rolls back the open GFN transaction; a committed
GFA audit is not a distributed transaction guarantee.

## Migration operation

Run Goose from the workstation using the configured **development**
`GOFURRY_GFN_MIGRATOR_URL` and `gofurry_migrator`, never an application role.
Before `up`, inspect active legacy notices whose publication timestamp exceeds
the China-site database clock. Such records were immediately public under the old
query; obtain the maintainer's decision before changing their visibility. Do not
rewrite their dates or choose scheduling silently.

Down refuses to proceed while any publication timestamp is null; it never invents
dates. Rollback requires deliberate data resolution by an operator. No production
migration was performed for P1. See [P1 acceptance](acceptance/issue-132-release-notes.md).

## Admin workspace (P2)

`/nav/update-notices`, `/nav/update-notices/new`, `/nav/update-notices/:id` are
explicit routes before the generic Nav resource route. The existing 更新公告
navigation and ContentRead/ContentWrite capabilities are reused. Readers can
browse details, switch languages and preview; all mutation controls require write.

The list retains API search/pagination and shows version, title, display status,
China-site publication time, abbreviated SHA and last update. Status is an Admin
aid only: draft, future published = Scheduled, otherwise Published. Timestamp
parsing is centralized and independent of browser timezone.

The editor uses one form for Chinese/English title, summary and Markdown body,
plus common version/SHA/date. `/new` stays local until Save or confirmed Publish.
First save POSTs a Draft; later saves PUT content without lifecycle fields. Save
success updates only Release Notes caches and clears dirty state; failures retain
input. A background query refresh never resets unsaved form values.

Publish/Schedule first persists current dirty content and stops if saving fails.
New publication POSTs the complete Draft once, then publishes without a redundant
PUT. Publish Now sends no browser timestamp. Schedule requires an explicitly
chosen future China-site time. If creation succeeds but publishing fails, the
saved Draft remains accessible at its ID rather than being created again.

Unpublish is confirmed, disabled while dirty, and preserves the date/editor.
Delete is a separate confirmed soft-delete action and returns to the list.
Dirty forms use the existing React Router blocker and a dirty-only beforeunload
listener; ordinary Save requires no confirmation and never auto-saves.

Markdown uses a selection-based textarea toolbar, side-by-side Desktop preview
and local Edit/Preview switching on narrow layouts. Only sanitized renderer
output enters the preview. Dependencies, headings, links/images, allowlists and
shared P3 fixtures are defined in [the Markdown contract](../contracts/update-markdown.md).
No P2 database, public-page, asset-picker or upload changes are required.

## Public experience (P3)

`/updates` and `/en/updates` render the current locale's index once on SSR;
hydration reuses that payload. Latest has its own reading surface. All remaining
rows are grouped by calendar month in Asia/Shanghai, without pagination, accordion
or load-more state. Missing summary/version/commit is omitted, never inferred from
body. Historical plain-text releases remain valid Markdown on their detail pages.

`/updates/:id` and `/en/updates/:id` fetch only detail, SSR the sanitized article,
and preserve `previous=older`, `next=newer`. Hidden/missing detail produces HTTP
404; upstream failure produces 503. Locale changes preserve the ID. One commit
helper abbreviates the displayed SHA but links the full SHA with safe external
attributes. Publication metadata remains evidence, not a client-generated date.

Public Markdown installs the exact P2 parser/sanitizer versions and consumes every
shared semantic/security fixture. Its single `updateMarkdown.ts` renderer is the
only source for `UpdateMarkdown.vue` HTML. No media service, syntax highlighting,
new lifecycle, schema migration or permission is introduced.

Existing app-level `useLocaleHead({ seo: true })` owns canonical/hreflang. Index
copy is localized. Detail title uses optional version plus title; description is
summary or a localized fixed fallback, never parsed body. Detail uses article OG
metadata, its own OG URL and publication time; there is no JSON-LD.

Sitemap fetches the public Chinese index alongside Sites, Site Groups and Games.
The strict Release Notes inventory parser adds both localized detail URLs and
fails closed with HTTP 503 for unavailable/malformed inventory. The existing
100-release index limit also limits sitemap release coverage; exceeding that
inventory requires future API work, not an additional endpoint in P3.

`tests/browser/regression/updates.spec.ts` owns Index/Article SSR, request accounting,
locale, metadata/Markdown integration, navigation, failure states and responsive
containment. `visual/updates-page.spec.ts` owns exactly eight pinned Linux goldens:
`updates-{light,dark}-{desktop,mobile}.png` and
`update-detail-{light,dark}-{desktop,mobile}.png`. Old Timeline components, SVGs,
runtime state and selectors are retired. Full remote/manual acceptance is a
separate gate from focused local verification; see the ledger.
