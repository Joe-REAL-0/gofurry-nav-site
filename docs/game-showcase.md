# Game Showcase backend contract (#119-A)

Stage A owns the backend contracts below. Stage B's React Admin operations consume
them through `/game/showcase` and `/game/showcase/:id`; Stage C adds the optional
Public Nav Web runtime described below. Showcase is a separate optional slice; the existing
`GET /api/v2/game/home` payload and its long-lived cache remain unchanged.

## Durable ownership

Goose migration `20261003010000_game_home_showcase.sql` adds only GFG objects:

- `gfg_game.showcase_eligible boolean NOT NULL DEFAULT false` opts Games into
  Automatic discovery independently of catalog `weight`.
- `gfg_showcase_campaign` owns draft/published/paused/archived Campaigns, scheduling,
  actions, pins and immutable artwork references. `linked_game_id` uses
  `ON DELETE SET NULL`; internal Game actions become ineligible after deletion.
- `gfg_showcase_campaign_locale` owns independent enabled zh/en text and at most
  three tags. Campaign text never falls back to the other language.
- `gfg_showcase_daily_stat` and `gfg_showcase_analytics_daily_quality` contain
  aggregate counts only. Diagnostic entity IDs have no FK; historical statistics
  survive current Game/Campaign deletion. Raw events and visitor identifiers
  never enter these tables.

No data is enrolled or backfilled by migration. Physical Campaign deletion is
restricted to Drafts; lifecycle actions never return a published Campaign to
Draft. Schema rollback deliberately refuses history loss and requires restoring
a verified backup. Initial production Managed Assets cutover must not be repeated.

## Admin API

All paths below start at `/api/v1/game/showcase`; reads use `content.read` and
mutations use `content.write`, existing session authentication and CSRF.

| Method | Path | Responsibility |
| --- | --- | --- |
| GET/POST | `/campaigns` | Bounded list / create Draft with immediate string ID |
| GET | `/campaigns/:id` | Workspace, derived status, publication diagnostics |
| PUT | `/campaigns/:id/content` | Locales, fixed action types, relation, focal point |
| PUT | `/campaigns/:id/schedule` | Absolute timestamps, weight, pin |
| POST/DELETE | `/campaigns/:id/artwork/{desktop|mobile}` | Publish bytes / clear reference |
| POST | `/campaigns/:id/{publish|pause|resume|archive}` | Explicit validated transition |
| DELETE | `/campaigns/:id` | Never-published Draft only |
| GET | `/composition`, `/candidates` | Authenticated Game Backend Composer diagnostics |
| GET | `/campaigns/:id/stats`, `/campaigns/:id/stats.csv` | Campaign daily statistics and totals |
| GET | `/analytics/quality` | Aggregate filter diagnostics |

List pagination uses `page_num` (1..1,000,000) and `page_size` (1..100), with
`state`, `derived_status`, `sponsored`, `keyword` filters. Statistics use inclusive
`from`/`to` ISO dates, at most 366 days, defaulting to the current and prior 29
Asia/Shanghai dates. `session_estimate` totals sum daily HLL estimates: a session
can count on multiple days. They are not exact user counts. CTR is calculated at
read time and never stored as a floating-point source of truth.

Content payloads contain `locales: [{lang,enabled,title,summary,tags,editorial_note}]`,
`content_type`, `sponsored`, `linked_game_id`, `focal_x/y`, and primary/secondary
action type and target. Missing locale entries are disabled. Primary types are
game/project/product/website; secondary types steam/kickstarter/website/other.
External targets require HTTPS without URL credentials. Internal game actions
use the linked Game; no arbitrary button label or HTML is accepted. Sponsored
content has no Editorial Note. Publication requires complete enabled locales,
desktop artwork, actions, ordered start/end timestamps and SFW linked Games.
Published edits repeat publication checks atomically. Pins serialize under a GFG
advisory lock and reject overlapping time/locale/slot conflicts, multiple pinned
Sponsored items and duplicated linked Games. The Game tag-domain lock protects
classification and Campaign publication against concurrent changes.

The existing Game workspace/classification API gains `showcase_eligible`.
Omission preserves the stored value. GFG writes preserve the existing separate
GFA Audit boundary: an Audit failure prevents GFG commit, but this is not a
distributed transaction. After commit, public-affecting Campaign/Game/tag writes
best-effort `INCR game:v2:showcase:revision`; Redis failure preserves success.

## Public and internal reads

`GET /api/v2/game/home/showcase?lang=zh&region=CN` returns the existing response
wrapper with `{schema_version:1,snapshot_id,generated_at,valid_until,items:[]}`.
Empty results are arrays. Managed and Automatic items share one `ShowcaseItem`:
key/source/reason/content_type/sponsored, optional string Campaign/Game IDs,
localized title/summary/tags, editorial-only note, artwork, actions, optional
canonical release, position and signed tracking token. No momentum score or
localized presentation label enters the public response.

Managed artwork contains object keys and focal coordinates; Automatic artwork
uses observed Steam header URLs, preserving the existing V2 asset preference
order and URL bytes. Automatic text uses current locale's site-maintained name,
summary and tag labels with no text fallback. Tag `code='adult'` excludes Games;
numeric legacy Tag IDs have no Showcase meaning.

`GET /api/v2/game/internal/showcase/{composition|candidates}` requires the
existing `RequireAdminToken()` middleware. Candidate diagnostics accept
`pool=upcoming|new_release|trending`, language, region and bounded pagination;
they expose eligibility/exclusion reasons and internal Trending evidence. Admin
proxies these reads using its existing configured Game Backend service address,
token and token header. It does not implement another Composer.

## Frozen Composer V1

`MAX_ITEMS=4`, `MAX_SPONSORED=1`; default Managed<=2 and Automatic>=2 when available.
Pins occupy exact positions and may force the documented balance exception.
After pins, select one Sponsored if possible and at most one default Editorial;
choose daily representatives for upcoming/new_release/trending, rotate pool
allocation hourly, backfill, then assign unpinned items by hourly order. A Game
appears at most once, and Sponsored has no reserved first slot.

Selection uses SHA-256 deterministic exponential races `-ln(u)/weight`, with key
tie-breaking. Managed seeds use UTC hour; Automatic seeds use pool plus UTC date;
position and pool order use UTC hour. Identical input sets, weights and seeds
produce identical output independently of input ordering. Snapshot identity hashes
schema version, revision, locale, region, hour/day buckets and ordered item keys.

- Upcoming: canonical availability `upcoming`; upcoming planned date/window within
  180 days has weight 300, farther known date/window 150, no useful date 100.
- New Release: `gfg_game_first_available.exact_date`, otherwise `window_start`,
  within 30 UTC calendar days. Weights: 0..7=300, 8..14=200, 15..30=100.
- Trending: current tracking identity, finalized player facts through the
  `game.player_facts` processed horizon. Recent=3 dates; baseline=preceding 7.
  At least 2 recent and 4 baseline observed days. Ledger sample coverage in each
  window is at least 50%; legacy-observed rows require those day counts. Averages
  are weighted by successful samples. Baseline>=5, recent>=20, delta>=10,
  ratio>=1.5. Momentum=`delta*min(ratio,3)`, rounded and clamped to 1..10000 for
  internal selection. No raw/current observation substitutes for finalized facts.

Cached Home first-eight updated/latest/popular IDs are preference exclusions for
the corresponding pools. An empty preferred pool falls back to the full pool;
missing Home cache skips the optimization and never triggers a Home rebuild.

Revisioned Showcase keys have a hard maximum TTL of five minutes, additionally
bounded by the UTC hour and nearest enabled Campaign start/end. A bounded local
lock/recheck avoids concurrent rebuilds. No prefix scan/delete invalidation exists.

## Tracking, privacy and operations

Explicit Game Backend YAML adds distinct long random
`game.showcase_tracking_secret` and `game.showcase_analytics_hash_secret`; neither
reuses JWT secrets nor loads implicitly from `.env`. Empty Showcase works without
signing; a populated slice fails closed if signing is misconfigured.

HMAC-SHA256 tokens bind version, snapshot/item/subject, identities, reason, slot,
locale, region, action mask and issued/expiry times. Expiry is snapshot validity
plus two hours; signature comparison is constant-time. Public events use
`POST /api/v2/game/home/showcase/events` with tracking_token, canonical UUIDv4
session_id, impression/click and optional click source. Structurally malformed or
oversized JSON gets 400; filtered valid submissions and Redis failures return 204.

Origin must match configured frontend CORS origins. Plausible nonempty browser UA,
explicit bot/script filtering and submitted-event limits (30/session/minute,
600/IP/minute) precede acceptance. Session identity is HMAC-derived; IP identity
is HMAC-derived with the Asia/Shanghai date. They exist only in ephemeral Redis
keys. Nothing introduces cookies, fingerprints or cross-session identities.
Client IP begins at the socket peer; only configured `server.trusted_proxy_cidrs`
can introduce a forwarded chain. Walk from the nearest trusted hop, falling back
on malformed chains. Fiber universal proxy trust is disabled; the global limiter,
Game view counter and Showcase analytics use this same helper.

Redis Lua atomically deduplicates impressions and the first click source per
session/snapshot/item. Clicks create an implicit impression when needed, including
on their server-clock business date across midnight. Click dedupe still spans the
token lifetime. Every daily row satisfies clicks<=impressions and clicks=sum of
source counters. HLL estimates sessions; subject sets avoid scans. Dedupe lasts
48 hours, rates three minutes, and counters/HLL/quality 72 hours.

The Game Backend maintenance scheduler aggregates every five minutes, recovering
today and the prior two Asia/Shanghai dates. UPSERT writes absolute snapshots,
never increments by Redis totals; older snapshots cannot reduce accepted counts.
Public event/aggregate writes are not operator Audit events. No raw event, raw IP,
IP hash, session or UA is persisted in PostgreSQL by Showcase analytics.

## Verification

React Admin's native Showcase workspace exposes composition, Campaign CRUD and
lifecycle, strict locale content, managed artwork/focal controls, Shanghai-time
schedules, statistics/CSV, aggregate quality and automatic diagnostics. It does
not duplicate Composer or public Hero behavior. Persisted artwork displays object
keys without guessing CDN URLs. Focal edits send the complete existing content
payload; eligibility writes remain in Game Classification. Campaign History is an
`audit.read`-gated entry to system Audit, not a separate history API. See
[Admin frontend contract](../contracts/admin-frontend.md) and
[React Admin acceptance](admin-react.md#showcase-operations-119-b).

Go/sqlc/schema checks follow the root playbook with Go 1.26.7. Pure Composer,
Trending, token, proxy and asset tests run in normal Go suites. Explicit
`GOFURRY_SHOWCASE_REDIS_ADDR=127.0.0.1:<disposable-port>` enables actual Redis tests
on disposable DBs 12/13/14. Never point these tests at shared or production Redis.

`TestAdminShowcaseThreeDatabase` uses the existing isolated Admin integration
configuration, injected COS/R2 stores and three fresh databases. With disposable
Redis it invokes Game Backend's HTTP bridge against that exact published GFG
Campaign, accepts signed events, writes aggregates and reads Admin stats/CSV.
The Game read-model integration additionally checks Adult code, strict locale,
canonical dates, preferred Steam assets and current-period/finalization boundaries.
`TestShowcaseAdditiveMigration` and fresh/baseline adoption verify schema semantics.
Real cloud publication is separate acceptance; these tests perform no cloud writes.

## Public Nav Web runtime (#119-C, before Visual acceptance)

Games Home requests Home and Showcase concurrently during SSR, with exactly one
GET per slice and no hydration GET. A failed Showcase read renders no Hero and
preserves catalog/News/statistics/sidebar. A valid empty slice also preserves the
existing Home layout. The public contract, Composer and backend analytics remain
unchanged; the frontend never reorders, selects or filters returned items.

The native Showcase surface precedes Recently Released. It uses 64/36 Desktop,
stacked 2:1 Tablet and 16:9 Mobile artwork; existing Managed/Steam components own
provider routing and fallback. UI text/action labels are localized, while content
uses the backend's exact locale. Index is page-local, controls are non-circular,
and transitions are instant. No autoplay/swipe or whole-Hero link is introduced.

An impression candidate requires >=50% viewport visibility in a visible tab for
one uninterrupted second. A page-local Set dedupes snapshot/item impressions.
Anonymous UUIDv4 sessions live in sessionStorage (memory only if blocked); clicks
send artwork/title/primary/secondary without awaiting the POST. Event errors are
silent and cannot stop navigation. The backend still owns qualification, implicit
impressions and authoritative dedupe.

Pure/Nuxt tests and the extended Games Home production fixture cover DTOs, asset
keys, session/visibility lifecycle, SSR/network budgets, media failures, navigation,
responsive semantics and event failure isolation. This first implementation does
not accept new Visual snapshots or close #119. Desktop/Tablet/Mobile maintainer
review precedes separately authorized pinned Visual generation and comparison.
