# Release Notes — Issue #132 P1

P1 establishes the GFN domain and public/Admin API contracts. It does not ship
the P2 Admin Release Workspace or P3 public index/article UI. **Do not deploy P1
independently as a finished Release Notes feature**: the current generic Admin
editor can create/save drafts but has no publishing controls.

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

Only non-deleted, published records with a non-null timestamp at or before that
clock are public. Draft/scheduled/deleted/missing detail requests all return 404.

## Public API

Both endpoints retain the existing outer `{code,data}` envelope and `lang=zh|en`.
Title, summary and body independently fall back to the other language; version
and SHA are not localized.

- `GET /api/v2/nav/updates`: existing schema-version-1 ready/empty/error envelope,
  ordered by `published_at DESC, id DESC` (existing limit 100). Adds `summary`,
  `version`, `commit_sha`. **Body and the existing index timestamps remain P1
  compatibility fields** because today's Timeline still consumes the index body.
  P3 owns removal of transitional index fields with its UI migration.
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
