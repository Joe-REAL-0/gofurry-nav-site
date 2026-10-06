# Database schema readability contract

## Ownership

Goose is the sole application schema owner. `db/admin` owns GFA (Admin
identity, Audit and Collaboration), `db/game` owns GFG (Game), and `db/nav` owns
GFN (Nav). Runtime never runs schema migrations. Persistence and sqlc rules
remain in [database.md](database.md).

Every current Goose-owned application **TABLE, COLUMN, FUNCTION and TRIGGER**
must have a nonempty PostgreSQL `COMMENT` containing Chinese meaning. This
includes historical facts, retained compatibility tables and projection
functions. Function comments identify the full signature; triggers belong to
their owning table. INDEX, CONSTRAINT and SEQUENCE comments are optional.

The executable inventory is `tools/internal/schema.Inspect`: application tables
and their columns/non-internal triggers in `public`, excluding
`goose_db_version`, plus non-extension-owned functions in `public`. PostgreSQL
catalogs, Goose bookkeeping, extension-owned functions and internal constraint
triggers are outside application ownership. If application ownership expands
to another schema or object kind, update the inspector and its tests in the
same change; do not silently omit it from governance.

## Chinese meaning and quality

Comments use Chinese. Necessary terms such as Steam, AppID, UTC, JSON and HTTP
are welcome, but English-only comments fail. The mechanical gate trims
whitespace and requires at least one Unicode Han character. Passing it is
necessary, not proof of semantic quality.

Review comments against migrations, sqlc queries, services/collectors, tests and
accepted contracts. Explain business ownership, units, NULL/unknown versus
false/zero, enum meanings, timestamp/timezone boundaries, historical identity
versus current references, derived states and quality denominators where
relevant. A missing historical acquisition ledger means the scheduled-sample
denominator is unknown, not zero. “状态”, “时间”, “数据”, “ID”, or “是否删除”
alone is insufficient. Stop and report semantics that cannot be established.

Functions explain their operation, domain, inputs/return meaning and material
side effects. Triggers explain when and why their behavior runs. Retained older
metric/detector versions must be described by their own executable semantics.

Semantic changes update the relevant comment in the same new migration. New
objects arrive with comments. Corrections also use a new migration; runtime
tools do not automatically translate or repair comments.

## Immutable history and final schema

Applied historical migrations are immutable. The readability backfill adds one
metadata-only migration per database, containing only
`COMMENT ON TABLE/COLUMN/FUNCTION/TRIGGER`. It cannot change business DDL,
business rows, function bodies or trigger behavior. Corrections roll forward;
there is no artificial Down restoring missing or obsolete meaning.

`tools/db-baseline/expected/*.json` is the immutable audited pre-Goose adoption
boundary and is **not** required to satisfy the new policy.
`expected-final/*.json` represents the current complete Goose chain and **is**
required to satisfy it. Existing optional comments remain part of structural
comparison. Never hand-edit final snapshots to make a gate pass.

Generate final snapshots through the existing harness using an isolated
PostgreSQL 18 admin connection, from `tools/`:

```text
GOFURRY_TEST_POSTGRES_ADMIN_URL=<isolated PostgreSQL admin connection>
GOFURRY_REQUIRE_POSTGRES_MAJOR=18
GOFURRY_UPDATE_SCHEMA_SNAPSHOTS=1 go test ./db-baseline -run TestPostgresFreshAndBaselineAdoption -count=1 -v
```

Unset `GOFURRY_UPDATE_SCHEMA_SNAPSHOTS` and rerun to validate comparison and
adoption. Never enable `GOFURRY_UPDATE_BASELINE_SNAPSHOTS` for readability work.
The harness creates/drops isolated databases; do not point it at production or
shared development. Operational migration is a separate, explicitly authorized
action using the migrator account.

## Executable validation and CI

`schema.ValidateReadability(Snapshot)` returns deterministic findings with
object kind, full object path and reason. Unit tests cover all mandatory kinds,
missing/whitespace/English-only comments, Chinese mixed with technical terms,
optional objects and stable ordering.

Two gates apply:

1. `task check:db-readability` runs `TestExpectedFinalReadabilityPolicy` on the
   three committed final snapshots. It needs no DB, `.env`, network or file
   mutation and joins `task check` and existing repository-policy CI.
2. Existing foundation CI uses fresh PostgreSQL 18, Goose up, actual
   `schema.Inspect`, existing expected-final comparison, then readability
   validation of the actual schema. Update mode validates before writing.
   Historical upgrade and adoption tests remain intact.

A fresh-database backfill test also compares before/after snapshots with only
comments removed, protecting structure and function/trigger definitions. sqlc
is generated normally; propagated model comments are documentation changes,
not runtime behavior changes.

## Non-goals

No ER diagrams, dictionary UI, Admin schema browser, automatic translation,
mandatory index/constraint/sequence comments, schema renaming, normalization,
ORM, startup migrations or unrelated SQL cleanup. This contract does not change
GFA/GFG/GFN business ownership or API contracts.
