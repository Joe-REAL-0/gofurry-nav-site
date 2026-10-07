# React Admin development

The sole Admin frontend is `apps/cn/admin/react`. It owns content workspaces and all operational/system workflows and is the production entrypoint embedded by the Go binary.

## IME-safe searches (#141)

`useCompositionSafeSearch` separates displayed IME drafts from committed search
text. Content ideas commit the final keyword to the URL with replace and page 1;
GlobalSearch keeps its two-character minimum and clears both values on every
dialog close. RemoteSelect and Board IdeaPicker debounce committed text only.
RemoteSelect leaves IME Enter/arrow/Escape handling to the input method and retains
the selected option through composing/blur. DataTable keeps its existing IME
implementation. Vitest covers composition, trailing changes, URL synchronization,
0/300 ms remote searches, keyboard selection and dialog reset. Debounce timing uses
fake timers.

The four owners above are covered by #141. Existing technical filters in Collection
(job key and run-result IDs/target/protocol), Changes (event code), and Metrics
(dimension value) still commit on each change; they are identified follow-up gaps,
not silently included in this patch. Ordinary form inputs and TagMultiSelect's
local filtering are unchanged.

## Showcase operations (#119-B)

`src/features/showcase` owns `/game/showcase` (current composition, Campaigns,
automatic discovery) and `/game/showcase/:id` (overview, content, artwork,
schedule/display, statistics). It uses native shared controls and the
existing [Stage A API](game-showcase.md), outside Resource Engine. Navigation,
breadcrumbs and the page title use “首页展柜”. Discovery locale/pool filters share
the table toolbar with column controls; daily trend legends sit above the plot.
Discovery now defaults to “可用候选”, with “待开启” (only missing operator approval),
“条件未满足”, and “全部诊断” groups. Counts and filtered totals come from the internal
Game Backend diagnostic read, before pagination. Search by localized name/Game ID
is debounced by 300 ms; advanced exclusion filters and pool-specific/name/ID sorting
use URL state (`candidate_status`, `candidate_keyword`, `candidate_sort`,
`excluded_reason`) independently of Campaign filters. Changing a filter resets the
page; changing pool clears incompatible exclusion filters. The table shows Game,
status, key evidence, blockers and actions; “查看诊断” opens full release/first-available
and finalized-player evidence in a keyboard-accessible dialog. Classification owns
all eligibility writes. This view requires the matching Admin proxy/Game Backend
read-only diagnostic extension; it does not change Composer or enrollment rules.
`content.read` can inspect every tab; `content.write` permits mutations. New Drafts immediately
open their real ID workspace. Lists use URL filters and do not fetch per-row stats.

Content explicitly edits zh/en; Sponsored clears both Editorial Notes and CTA
types stay fixed. Assets accept original desktop 1600×800/mobile 1200×675 AVIF up
to 5 MiB. The browser checks file hints/size; Go owns authoritative dimensions.
Local previews are revoked on replacement/unmount. Reloaded artwork shows its
object key, since Stage A supplies no persistent CDN URL. Mirror warnings remain
successful Primary publication. Focus coordinates save the full content contract.
A 3×3 grid highlights the nearest preset; staged local preview clicks select
clamped x/y and update object-position immediately. Saving remains explicit and
read-only controls are disabled. Weak technical text replaces the numeric inputs.
No persistent CDN URL is guessed when no local preview exists.

Schedules interpret DateTimePicker wall time as Asia/Shanghai (UTC+08:00), with
explicit RFC3339 conversion. Lifecycle actions and reference clearing confirm;
unsaved edits/staged uploads protect tab navigation, routing and unload. Statistics
display backend totals, daily two-series trends, click sources and same-origin CSV.
Multi-day sessions are summed daily HLL estimates, not distinct people. Discovery
links to Game Classification for eligibility. The Header “操作审计” action is
visible only with `audit.read` and carries only the Campaign resource filter.
History is removed; old `tab=history` falls back to overview without redirect.
“选取权重” and “固定展示位置” replace Weight/Pin labels, with explanatory
help and unchanged 1..10000 / null-or-1..4 payloads.

Vitest/Testing Library cover these contracts without real cloud resources. After
automated checks, human acceptance still covers light/dark and narrow screens,
the full create/edit/upload/schedule/publish/pause/resume/stats/CSV workflow and
Game eligibility reflected in diagnostics. Real Development COS/R2 acceptance
requires explicit authorization. Production data/cloud operations and Stage C
Public Web implementation are outside this stage.

## Local development

Start the existing Go API with the ignored local development config:

~~~text
cd apps/cn/admin
go run . serve --config config/server.yaml
~~~

Then start Vite in another terminal:

~~~text
cd apps/cn/admin/react
pnpm install --frozen-lockfile
pnpm run dev
~~~

Open `http://127.0.0.1:5178`. Vite proxies `/api` and `/csrf` to `http://127.0.0.1:10099`.

Validation:

~~~text
pnpm run typecheck
pnpm test
pnpm run build
~~~

`pnpm run build` clears and writes `apps/cn/admin/internal/transport/http/webui/dist`. The root `task build:admin` target performs this React build before compiling the Go binary and copying the deployment `dist/` companion artifact. No manual asset copy or runtime Node process is used.

The App Shell consumes the current principal from `/api/v1/auth/state`. Missing navigation or actions should first be checked against returned capabilities and backend authorization; never patch around the contract with role comparisons.

Simple resources are defined in `src/features/resources/definitions.tsx`. Site, Game and Release Notes remain dedicated workspaces. `src/features/release-notes` owns list/new/detail routes, bilingual Markdown authoring and the P1 publication APIs; see [Release Notes](release-notes.md) and the [Markdown contract](../contracts/update-markdown.md). New server reads should be small, explicit sqlc-backed read models rather than a generic frontend BFF.

Collection, Metrics, and Changes are under `src/features/operations`; Cloud Resources, DataOps, Audit, and Accounts are under `src/features/system`. `dataops.read` is the only valid Data Operations capability. Operator/Developer/Owner differences must be expressed through `auth.can(...)`, not client-side role matrices.

`/system/cloud` uses shared Admin controls for storage summaries, object inspection, scoped purges, and task history. `cloudops.read` permits reads; `cloudops.manage` permits mirror repair and scoped EdgeOne/Cloudflare purges; `cloudops.purge_all` permits the separate, initially collapsed full-zone EdgeOne action. The backend grants the first two to Owner/Developer, the last only to Owner, and none to Operator.

Authenticated self-service username/password actions use `/api/v1/auth/self/*` with current-password verification and no `account.manage` requirement. Username changes refresh identity without ending the session; password changes clear authentication and require login again.

See [the cutover parity matrix](admin-frontend-parity.md) and [the role operator guide](operations/admin-roles.md) for production acceptance boundaries.

`src/features/collaboration` owns `/collaboration` (ideas/board), visible pipe-delimited line parsing, version conflicts and shared idea context. Creation pages prefill only Steam AppID or Site name; never auto-fetch Steam or create targets. Link failure preserves successful creation and the `?idea=` recovery banner. Use Vitest/Testing Library for these flows; see [Collaboration Center](collaboration-center.md).

The Board tab lazy-loads React Flow for the shared canvas. Node/edge records and Audit remain in GFA; reference cards reuse existing idea/options reads. Drag/resize saves once at gesture end, multiple selected nodes save atomically, and conflicts require explicit reload. The frontend keeps a local movement draft across polling, including remote deletion; it never retries with a newer version automatically.
