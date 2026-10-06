# Baseline adoption utility

Current final schemas also follow the
[database readability contract](../../contracts/database-schema.md).
`task check:db-readability` validates committed `expected-final` snapshots
without a database. Existing foundation tests validate actual fresh PostgreSQL
18 schemas after Goose. `expected` remains the immutable historical adoption
contract; do not update it for comment backfills. Use only the documented fresh
database snapshot-generation path to update `expected-final`.

This one-time utility marks an existing, exact pre-Goose database as having the
audited baseline. It does not execute business DDL or attempt repair.

It rejects the operation unless:

- the selected contract is `gfg`, `gfn`, or `gfa`;
- version `20260823000000` is explicitly selected;
- the connected database name matches the selected contract;
- the standard Goose version table does not exist;
- tables, columns, types, nullability, defaults, sequences, constraints,
  indexes, extensions, functions, triggers, and relevant comments exactly match
  the embedded audited snapshot.

Only after all checks pass does it use Goose's public PostgreSQL version-store
API to record version `0` and the baseline version in one transaction.

Operator usage from `tools/`:

```text
GOFURRY_DATABASE_URL='postgres://...' go run ./db-baseline \
  -database gfg \
  -baseline-version 20260823000000 \
  -confirm-adopt
```

Do not use this command for an empty database. Empty databases use normal
`goose up`. Do not run it until the production schema-only dump has been
structurally compared with a scratch database built from the baseline.
