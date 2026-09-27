# Issue #109 — Site Detail phase ledger

## P1: Runtime and information architecture

Scope: the [Site Detail runtime contract](../../contracts/nav-web-frontend.md#site-detail-runtime-and-route-ownership-109-p1).
This is P1 evidence only; #109 and P2–P8 are not complete. #108 is unchanged.

### Audit and resolved discrepancies

- The route's `hasTargetQuery` / `showInsights` condition hid Site Insights on
  Target URLs, and `insights-entity` asserted that absence. Both are replaced by
  Site-ID-only Insights ownership; the existing panel appearance remains.
- The old three-key array was the only panel catalog. The shared registry now
  contains all seven capabilities, categories, order and bilingual label keys.
  The legacy three-item preview is explicitly a subset, not the P3/P6 catalog.
- `siteRoutes.ts` previously discarded workspace query when selecting a Target.
  It now delegates to the single route-state parser/normalizer/builder; primary
  tab transitions clear previous secondary state, Target transitions preserve it.
- `buildSiteDetailSeo` copy and Entity-only canonical/hreflang/sitemap behavior
  needed no production rewrite. SEO Browser coverage now exercises all UI keys.
- A populated HTTP detail mounted `SitePerformance`, whose mount hook fetched
  Ping observations immediately. This contradicted the required initial budget;
  the old empty-HTTP fixture hid that request. The #109 fixture now supplies HTTP
  evidence and distinct Target status values. Ping history is fetched only by
  existing sample controls, with pending dedupe/cache and stale-result guards;
  no new control, chart, endpoint or history feature is introduced.

### Executable ownership

- Site Detail requests use Site ID, selected Domain and language. Site Insights
  uses only Site ID; View remains a Site-level mounted side effect. A counted
  View value survives Target Detail replacement.
- Seventeen independent `site-detail-contract.spec.ts` cases use the shared
  `insights-runtime` Nitro/upstream transport and exact request/error ledgers.
  They cover zh/en SSR, real Target clicks from all four workspace states,
  invalid UI fallback, authoritative Detail failure, optional Insights failure,
  success-empty, View failure and explicit legacy Ping-history interaction.
- Successful hydrated Target switches retain Site capability semantics and
  workspace query with exactly one additional Detail request, zero additional
  Insights requests and zero additional View POSTs. The HTTP evidence changes
  from fixture status 200 to 201. The normal initial budget is three requests,
  including when the performance component is present.
- Client Detail 503 reaches the page error. Existing transport retry behavior
  remains explicit: SSR GET failures have two upstream attempts; client failures
  have two browser attempts, each with two Nitro proxy upstream attempts.
- The existing `insights-entity` bridge and 1440/390 × Light/Dark runtime cases
  remain. SEO checks require query-free localized Entity canonical/hreflang and
  sitemap inventory. Unit tests cover route vocabularies/cleanup/encoding and
  complete registry metadata. No runner or browser-error allowance was added.

### Local verification (2026-09-26)

Environment: Windows, Node 24.15.0, pnpm 12.6.0, frozen dependency installation.
Chromium installed through the existing Playwright command; retries remain zero.

Passed: lint, stylelint, style-policy tests (74), exact style-policy baseline,
Unit (83), Nuxt (6), combined Vitest (89), typecheck, Insights semantics, SEO
recovery guard, production build and focused Site Detail Browser (17).
The final full `pnpm run test:browser --workers=1` run passed all **423** cases
in 12.3 minutes, with no failures, skips or retries. The earlier interrupted
run is not acceptance evidence. Local P1 exit criteria are satisfied.

Style debt, ESLint suppressions, Visual specs/configuration and the 118 accepted
PNGs are unchanged. No style-budget updater or snapshot generation was run.
Pinned Linux Visual comparison and remote CI are unverified for this change;
no earlier CI run is used as acceptance evidence.

### Maintainer handoff

Confirm on a real Site with multiple Targets: normal and Target URLs retain
Insights; selecting another Target changes its evidence without routing failure;
canonical remains the localized Entity URL. Existing Ping history now requires
one sample-button interaction. No new visual review is requested by P1.

P2 owns Shell/Hero/Health Strip/Target Context; P3 Overview; P4 Observation;
P5 Security; P6 Insights workspace; P7 appearance/debt cleanup; P8 Visual/closure.
After the lightweight maintainer behavior confirmation, P2 can use the P1
owners. Its implementation remains outside this phase; remote CI/Visual status
above must not be represented as a pass.

## P2: Shell and Target context

Scope: [P2 shell contract](../../contracts/nav-web-frontend.md#site-detail-shell-and-target-context-109-p2).
P1 remains the mandatory runtime baseline. This implementation stays on `dev`;
P3–P8, #108 and final Site Detail Visual golden creation remain out of scope.

### Implementation and ownership

- Rewrote the existing Hero in place as identity-only. Removed its hover domain
  popover, edge/keyword/technical content and the old large signal-card path.
- Added six Target-owned health fields, four router primary tabs, a 3:1 desktop
  workspace/context grid, and compact inline context on tablet/mobile. Only the
  tabs and desktop context stick; the public header/Hero/health remain normal flow.
  Mobile health is 2×3; long descriptions are clamped on mobile.
- A single responsive Context and accessible listbox selector use the shared
  `siteTargetPresentation.ts` projection. It preserves missing/zero/false, refuses
  unrelated Target data, never substitutes Site aggregate health, and retains
  infrastructure hint confidence and existing Target relations.
- Checked backend missing-summary constructors: nil Target arrays and Go zero
  timestamps are legitimate responses. Adapter and tests handle them as missing
  evidence rather than throwing or displaying year 1.
- Tab changes use the P1 query owner, restore through reload/back/forward and
  clear foreign workspace state. Target selection preserves valid secondary
  state. Pending retains the shell and explicitly labels last-resolved evidence;
  held-response coverage proves a late A response cannot overwrite newer B.
- Kept one Site Insights slice mounted. Existing preview appears in Overview and
  Ecosystem; legacy observation panels and minimal transport evidence are thin
  transitional workspace content, not P3–P6 implementations. Public labels retain
  the accepted “生态观测 / Ecosystem” naming while the route value remains `insights`.

### Executable verification

- P1's 17 Browser cases retain SSR, authoritative/optional failure boundaries,
  SEO/request budgets and Target-switch assertions. They now click the selector;
  the existing history sample test opens Observation explicitly.
- Twelve additional `site-detail-shell.spec.ts` cases cover six viewport/theme
  combinations, long hostnames, route/history/reload, tab and selector keyboard,
  touch, outside dismissal/focus return, keyboard visibility in a long list,
  pending/race and missing/sticky behavior.
  The same Nitro/upstream fixture, strict error capture and zero retries apply.
- Five pure adapter cases cover Site/Target ownership, null/zero/false, stale
  evidence, deterministic timestamps, relations/hints and actual missing-summary
  shape. A policy test confines the new token declarations to exact roots.

### Style and Visual boundaries

Only Hero/popover/page-root debt was removed: Tailwind appearance **632 → 572**,
raw visual values **463 → 412**, deep selectors **33 → 25**. Arbitrary appearance
(5), `!important` (5), legacy dark entries (0) and every other per-file budget
remain unchanged. The updater ran only after the policy reported these five
stale entries with zero regressions. ESLint pruning removed only the page's two
obsolete `no-explicit-any` suppressions. No budget transfer/increase occurred.

Existing Visual specs/configuration and accepted PNGs are protected independently
from the local manual-review screenshots. No snapshot update or new runner is
part of P2. Remaining legacy panel debt retains its later-phase owners.

### Local and remote acceptance

On 2026-09-26, Windows / Node 24.15.0 / pnpm 12.6.0 passed frozen installation,
lint, stylelint, style-policy tests (75), exact style-policy baseline, Unit (88),
Nuxt (6), combined Vitest (94), typecheck, Insights semantics, SEO recovery guard
and production build. Chromium was installed through the existing Playwright
command. The final focused Site Detail run passed 29 cases. After the last
selector fix, the complete `pnpm run test:browser --workers=1` run passed all
**435 cases in 13.0 minutes**, with no failures, skips or retries. These are the
current local acceptance results, not the earlier exploratory/failing runs.

The protected Visual file hashes and all **118 accepted PNGs** match the P1
baseline; Visual inventory remains 119 tests. No accepted snapshot was generated
or changed. Remote CI and pinned Linux Visual comparison remain **unverified**
for this local change; no push was performed and earlier remote runs are not
acceptance. Technical P2 criteria are locally verified; manual criterion 30 is
still outstanding.

### Required maintainer review before P3

Review 1440 Light/Dark, 390 Light/Dark and preferably 768 Light: identity clarity,
Hero height, six-field density, sticky tabs, 75/25 balance/sidebar width, long
hostnames, selector operation, mobile 2×3/compact context and theme consistency.
Local fixture screenshots support this review; they are not a final golden or
maintainer approval. Also confirm a real multi-Target Site's Visit URL and Target
switch behavior. P2 exit criterion 30 remains pending until the maintainer accepts
the visual direction. Do not enter P3 before that acceptance.

## P3: Site Overview workspace

Scope: [P3 Overview contract](../../contracts/nav-web-frontend.md#site-overview-workspace-109-p3).
The maintainer explicitly requested P3 on current `dev`; the historical P2 review
record above is preserved rather than retrospectively marked approved. P4–P8,
#108, backend/API/schema changes and dependency changes remain outside this work.

### Implementation and ownership

- Replaced Overview's Target protocol checks and old Insights preview with Site
  Health, conditional Attention, seven capability rows and up to four changes.
  The old panel now renders only on the Insights tab; its fetch remains Site-owned.
- Added a pure `siteOverviewPresentation.ts` projection and four small components.
  Site Health reads only the first Site/language summary of the page session.
  Target changes retain that snapshot; reload adopts the next summary. State and
  status remain separate, zero counts are omitted, and Site generated time is UTC.
- Attention prefers human messages, deduplicates described Target problems and
  keeps meaningful uncovered fallback. A healthy fresh Site has no Attention block.
- All capabilities use the registry; HTTP/2's earlier Transport category was
  corrected to Network to match P3. Unsupported stays neutral; backend unknown,
  missing successful facts and unavailable are separate. No ecosystem percentages
  or coverage appear. Empty success still displays seven missing rows.
- Changes reuse shared labels/order/time precision. Exact timestamps use explicit
  UTC for SSR/client agreement; day-only events stay dates. Other callers retain
  the shared formatter's existing default. The full Ecosystem link uses P1 route state.
- The outer P2 shell stays intact. Overview uses a desktop 3:2 capability/change
  layout and stacked tablet/mobile sections, extending `site-detail.less` with
  existing tokens. Current Target evidence remains outside Overview.

### Executable verification

- Twenty-six new pure unit cases cover health states, missing versus unknown,
  reason priority/deduplication/fallback, all capability states, seven-item grouping,
  unavailable/empty, four-item order, exact/day precision and Chinese copy.
- Sixteen `site-overview.spec.ts` cases use the existing deterministic runtime.
  They cover 390/768/1440 Light/Dark geometry/overflow, SSR markup, Attention and
  capability/changes states. A held Target request returns a changed Site summary:
  the Overview remains unchanged, while Current Target changes. Tab remount,
  history and full Ecosystem navigation add no requests; reload adopts the summary.
  A different browser time zone exercises precise timestamp hydration.
- P1 and `insights-entity` assertions were migrated to the appropriate Overview or
  Insights surface without weakening capability/timeline/error/request checks.
  P2 tests, strict diagnostics and zero retries remain unchanged.

### Style and Visual boundaries

The style debt baseline is byte-for-byte unchanged: Tailwind appearance 572,
arbitrary appearance 5, raw visual values 412, important 5, deep selectors 25,
legacy dark entries 0. No updater or suppression pruning ran. The still-active
legacy panel's debt remains for P6; it was not moved, hidden or re-budgeted.

Protected file hashes, all 118 accepted PNGs and the 119-test Visual inventory
remain unchanged. Five local manual-review screenshots cover 1440 Light/Dark,
390 Light/Dark and 768 Light using the existing fixture. Their temporary capture
test was removed; no runner, Visual spec or final P8 golden was introduced.

### Local and remote acceptance

Local Windows / Node 24.15.0 / pnpm 12.6.0 passed frozen installation, lint,
stylelint, style-policy tooling (75), exact style policy, Unit (114), Nuxt (6),
combined Vitest (120), typecheck, Insights semantics, SEO recovery guard and
production build. Chromium installation succeeded. The final focused Overview
run passed 16 contracts plus five temporary manual captures. A precise timestamp
spacing issue found in the first focused run was fixed and verified on the rebuilt
production output; it is not counted as a passing initial run.

The complete current `pnpm run test:browser --workers=1` run passed **451 cases in
14.3 minutes**, with no failures, skips or retries. P1/P2, Entity, SEO and P3
contracts all passed against the final production build. Remote CI and pinned
Linux Visual comparison remain unverified; no push was requested, and earlier
CI runs are not acceptance. Technical exit criteria are locally verified; the
manual criterion below remains outstanding.

### Required maintainer review before P4

Confirm Site-wide versus Current Target clarity, Health hierarchy/height,
conditional Attention prominence, seven-capability density, four-change density,
desktop 60/40 balance, mobile flow and Light/Dark consistency. Review screenshots
support this decision but do not constitute maintainer approval. P3 exit criterion
29 remains pending until the maintainer accepts the Overview; do not enter P4.

## P4: Current Target Observation workspace

Scope: [P4 Observation contract](../../contracts/nav-web-frontend.md#site-observation-workspace-109-p4).
The maintainer explicitly requested P4 against current `dev`; the historical P3
manual-review record above is preserved. P5/P6, #108, backend APIs, migrations,
dependencies and final P8 goldens remain out of scope.

### Runtime and evidence ownership

- Added route-owned Overview/Performance/HTTP/DNS/Web secondary tabs with
  back/forward/reload, roving keyboard focus and bounded mobile horizontal scroll.
  The default view is omitted; the secondary row is not sticky.
- `siteObservationPresentation.ts` owns raw evidence parsing and Target identity.
  Protocol rows separate status/duration/observed/freshness; reported human reasons
  precede raw code fallback. HTTP presents raw headers with native disclosure and
  only actual redirects. DNS groups collected records/children and retains reported
  risks and secondary infrastructure details. Web admits only metadata, robots,
  llms.txt, page assets and RDAP; security probes are excluded.
- Replaced the selected-but-unrequested Ping chart with page-owned history state.
  Performance hydration auto-loads one `limit=100` Ping request; Overview/HTTP/DNS/
  Web and SSR do not. Sample 20/60/100 changes are local. Page-session cache keys
  include Site/Target/protocol; all outcomes are retained until explicit retry or
  a new page session. A late A response can fill A's cache without replacing B.
- Non-Performance Target changes add only Detail. Performance adds Detail plus
  one request for uncached Target history, while Site Insights/View and the P3
  Site snapshot remain stable. Detail failure remains authoritative; history and
  Insights failures remain local and View failure remains harmless.
- Independent timing bars preserve collected stages and Total, explicitly noting
  overlap instead of inventing summed timing. Collector source inspection showed
  `loss_rate` is already a 0–100 percentage: even 0.1 remains 0.1%, not 10%.
  Missing RTT/loss is not fabricated as zero. ECharts is visible/ready only,
  shallow, theme-token-driven, resize-aware and disposed without double-update.
  Loading, empty, unavailable and available-without-RTT all have explicit copy.

### Replacement and measured debt

Consumer audit retired ten Observation components: SitePerformancePanel,
SitePerformance, SiteObservationTabs, SiteMetadataProbePanel,
SiteObservationOverviewPanel, SiteDnsPanel, SiteHttpPanel, SiteMetadataRows,
SiteObservationHistoryPanel and SiteObservationInfoList, plus the unused
useSiteMetadataProbePanel composable. No Legacy/Old clones remain. Security's
metric renderer, light-probe renderer, Site Changes/Insights and P7 leftovers keep
their later-phase ownership; this does not claim total Site debt closure.

Style policy first reported only stale budgets for those real deletions, with no
regressions. Only then did `style:policy:update` lower them, followed by a clean
policy run. Tailwind appearance **572 → 147**, arbitrary appearance **5 → 2**,
raw visual values **412 → 246**, deep selectors **25 → 0**. Important remains 5,
legacy dark entries 0. All other per-file budgets are unchanged; none was raised,
transferred or hidden. ESLint pruning removed exactly 23 obsolete suppressions
from four deleted files and left all other entries intact.

### Executable verification

- Fourteen added Unit cases cover the route helper and evidence projection:
  protocol priority, identity, null/zero/false, percentages, independent timing,
  header normalization, nested DNS, strict Web allowlist and history precision.
- Five real Nuxt cases mock only the API boundary and cover activation, sample
  slicing, empty/failure/ready states, explicit retry, view/Target cache reuse,
  late-A/ready-B isolation and distinct Site identities.
- Thirty new Functional Browser contracts use the shared deterministic fixture.
  They verify five SSR views without history, keyboard/router navigation, gated
  default-sample loading, exact request budgets, cache/race/error behavior, HTTP
  disclosures, DNS groups/chains and Web security exclusion. 390/768/1440 ×
  Light/Dark cover long evidence wrapping, waterfall, Canvas readiness/height and
  history rows. No fixed sleeps, networkidle, retries or diagnostic exemptions.
- P1's history test now activates Performance and checks non-Performance Target
  switching. This intentionally replaces its earlier sample-click trigger with
  P4's auto-lazy contract; P1–P3 runtime, SEO and failure coverage remains required.

### Local and remote acceptance

Windows / Node 24.15.0 / pnpm 12.6.0 passed frozen install, lint, stylelint,
style-policy tooling (75), exact style policy, Unit (128), Nuxt (11), combined
Vitest (139), typecheck, Insights semantics, SEO recovery and production build.
Chromium installation succeeded. After the final P4 UI changes, the focused
P4 run passed 30 contracts plus ten temporary review captures. The capture test
was removed before complete acceptance; it is not a runner or Visual baseline.
The first full Functional run passed 480/481 in 14.0 minutes and exposed an
existing Game Detail locale-transition race: two successful Info requests for
the same identity. The unchanged test passed alone; holding that Info response
then reproduced the duplicate deterministically. A separate minimal repair makes
concurrent consumers defer to the same pending request. Its existing Browser
contract now holds that response and still requires exactly one request; no
assertion, diagnostic or retry policy was weakened. After rebuilding, all nine
Game content cases and 139 Vitest cases passed. The final full
`pnpm run test:browser --workers=1` passed **481 cases in 14.6 minutes**, with no
failures, skips or retries. P1–P4, Entity, SEO and failure boundaries all passed
against this final production build.

All accepted Visual files and 118 PNGs remain unchanged; Visual inventory stays
119 tests. No final P8 golden was generated. Remote CI and pinned Linux Visual
comparison are unverified; earlier CI runs do not count as this change's acceptance.

### Required maintainer review before P5

Review Observation Overview, desktop/mobile Performance, desktop HTTP/DNS/Web
and representative Dark views. Confirm secondary navigation hierarchy, automatic
default loading, timing interpretation, chart height, evidence/disclosure density,
resolution-chain readability, mobile flow and theme consistency. Local screenshots
support review but are not maintainer approval. P4 exit criterion 38 remains
pending until that review; do not enter P5 automatically.

## P5: Current Target Security workspace

The maintainer declared P1–P4 complete and explicitly requested P5 on current
`dev`. Historical phase records above remain intact. The scope is the
[Security contract](../../contracts/nav-web-frontend.md#site-security-workspace-109-p5);
P6, #108, backend/schema changes and final P8 goldens remain outside it.

### Evidence and runtime ownership

Six Security components replace the transitional panel. Overview, TLS & Certificate,
Web Security and Exposure use `selectSiteSecurityView` and the P1 route owner,
including reload/back/forward, roving keyboard navigation and mobile row scrolling.
`siteSecurityPresentation.ts` is a pure Current Target projection; Vue does not
parse raw payload or infer business states. Existing Observation facts are reused
with an optional structural break-all mode for cryptographic values.

P5 adds zero requests. All evidence is SSR Detail data. View switches add none;
Target changes add Detail only, retaining workspace and last-resolved pending
evidence without repeating Insights/View. The raw Header link retains explicit
Target context or its implicit-primary form, avoiding a spurious Detail request.
The existing authoritative Detail and optional Insights/View boundaries remain.

Overview explicitly scopes technical observations and labels certificate
verification. Attention comes only from verification failure, collector remaining
days, reported security.txt validation issues and positive Canary mismatch/error
counts. There is no score, grade, overall verdict or WAF deployment field/claim.

TLS transport, verification and validity are independent. Collector Go defaults
behind `cert_collected=false` cannot become failed certificate evidence; an explicit
verification fact remains usable in older payloads without that flag. The shared
`readSiteCertificateEvidence` also prevents the Health Strip from contradicting
Security with uncollected zero/false defaults. Expiry bands
use collector days, never `Date.now()`. SAN/chain/crypto are disclosures; hashes wrap
in monospace. False OCSP and zero SCT remain neutral observed facts.

HTTP `security_headers` is a boolean map, not raw header text. The projection
prefers `security_header_summary.present`, keeps raw Header values and does not
reimplement policy parsing. Missing, not-observed and unavailable remain distinct.
security.txt retains found/validation-issues/not-found/unavailable/not-observed,
consuming only collector validation errors and no frontend RFC validator.

Port outcomes stay neutral, including open 22/443; metadata is disclosed. WAF
matching is limited to a successful complete sample with known matching counts.
Cases remain folded initially, truncation is explicit outside disclosure, and
mismatch/error counts never become an overall security assessment.

### Legacy and measured appearance debt

Consumer audit found no remaining use of SiteObservationMetricGrid or
SiteLightProbePanel after replacement; both were deleted without clones. Other
legacy owners remain for their later phases. No ESLint suppression change occurred.

Policy first reported only those two stale raw-visual budgets: 58 and 74. After
actual deletion, `style:policy:update` removed exactly those entries and policy
passed. Raw visual debt **246 → 114**; Tailwind appearance **147**, arbitrary **2**,
important **5**, deep selectors **0** and legacy dark **0** are unchanged. No
surviving file budget was raised, transferred or hidden.

### Executable and manual verification

- Added 50 Unit cases (48 Security and two Target regressions) for route transitions, Target identity, booleans/zero/missing,
  collection defaults, expiry boundaries, raw/summary Header precedence, probe
  states, reported errors, truncated/unknown Canary counts and recursive exclusion
  of conclusion fields.
- Added 50 Functional Browser cases on the existing deterministic Site/Nitro/upstream
  fixture. Four SSR views, all navigation/request accounting, pending Target changes,
  optional and authoritative failures, evidence state matrices and native disclosures
  are covered. 390/768/1440 × Light/Dark assert long-value wrapping and neutral ports.
  Browser errors, exact request counts and zero retries remain strict.
- P1–P5 focused acceptance initially passed 125 formal cases plus ten temporary
  screenshot captures. After the final Overview wording, the 50 P5 contracts plus
  ten refreshed captures passed again. Screenshots are outside the repository and
  are review materials, not accepted baselines. The temporary capture test was
  removed before full acceptance; no additional runner was introduced.

The initial full run passed 531 cases in 14.8 minutes. Subsequent integration review
found that the Health Strip still interpreted uncollected certificate defaults as
evidence. Two new Unit tests reproduced it before the shared certificate normalizer
repair; Browser now asserts matching neutral/missing Strip state as well.

Frozen install, lint, stylelint, style-policy tooling (75), exact policy, Unit (178),
Nuxt (11), combined Vitest (189), typecheck, Insights semantics, SEO recovery and
production build have passed locally. Chromium installation succeeded. After the
integration repair and rebuild, P1/P2/P5 focused acceptance passed 79 cases. Final
`pnpm run test:browser --workers=1` passed **531 cases in 14.3 minutes**, with no
failures, skips or retries. P1–P5, Entity, SEO and failure boundaries passed against
the final unchanged production build.

Visual inventory remains 119 tests / 118 accepted PNGs with no accepted file
changes. Remote CI and pinned Linux Visual comparison remain unverified; previous
CI runs are not this change's acceptance.

### Required maintainer review before P6

Review Overview semantics, verification versus validity, certificate/disclosure
density, Header missing tones, security.txt state copy, neutral Port results and
bounded WAF interpretation. Check desktop/mobile Exposure, long values and theme
consistency. P5 exit criterion 40 remains pending until the maintainer accepts
these visuals. Local tests and screenshots are not that approval; do not enter P6.

## P6: Site Insights workspace and Target-independence closure

The maintainer declared P1–P5 complete and explicitly requested P6 on current
`dev`. Earlier manual-review history remains intact. P7, #108 appearance,
backend/schema changes and P8 goldens are outside this phase.

### Ownership, scope and explicit contract reconciliation

Five Site-owned components replace the legacy preview with seven grouped selectable
registry rows, the selected Site fact/date/adoption/coverage, ecosystem adoption
trend and the full recent-change set. There are no secondary tabs. External links
lead to the localized ecosystem and Compare with only this Site preselected.

The P6 brief uses Insights / 洞察 in example public copy, while the existing
executable semantic contract explicitly retires those product names in favor of
Ecosystem / 生态观测. New public copy preserves that accepted naming; internal
owners still use Insights names. The required ecosystem adoption trend and recent
Site changes titles are preserved. No semantic guard was relaxed.

`siteInsightsPresentation.ts` has no Target input. Slice unavailable yields null
fact states/dashes, not seven fabricated backend unavailable facts. Success-empty
retains seven missing facts. P3's older fabricated unavailable row state was also
corrected; real backend unavailable/unknown states remain meaningful. Adoption
and coverage stay distinct and null never becomes zero. `insightChanges.ts` adds
shared presentation categories for known public events, retaining unknown fallback,
the full returned set and explicit UTC versus date-only precision.

Page-owned `useSiteInsights` shares SSR data/state/retry with P3 and P6. Retry only
refreshes Site Insights, validates response identity and cannot let an old Site
response overwrite the new Site. `useSiteInsightTrend` owns hydrated activation,
metric/range cache, classified states, explicit retry and late-response isolation.
Neither owner accepts Target identity. Existing transport retry behavior remains;
the new owners add no automatic retry loop.

SSR remains Detail + Site Insights. Normal hydration adds View; entering Insights
adds exactly one selected uncached metric/range trend. Metric/range are URL-owned,
with helpers retaining Target and the other selection. Target switching adds only
Detail and cannot reset matrix/detail/chart/recent changes, including while pending.
The two optional slices fail and recover independently.

The Site-owned ECharts surface uses lazy canvas, a shallow instance, ResizeObserver,
import revision protection and disposal. Only adoption is plotted and null points
remain gaps. Chart initialization itself is classified loading; load/render failure
is unavailable, and one-point/all-null samples have meaningful presentation.
No #108 rail/workspace or appearance owner is imported.

### Legacy and measured style debt

Consumer audit found SiteInsightsPanel was InsightsEntityTimeline's only remaining
consumer. Both are removed; their isolated Site panel/timeline styles are deleted.
Shared Game rules and all #108 important/domain appearance debt remain unchanged.
The obsolete three-item registry preview metadata is retired, without a new catalog.

Policy initially failed only on the stale six raw-visual entries in insights.less.
After real removal, style:policy:update removed exactly that budget. Raw visual
**114 → 108**; Tailwind appearance **147**, arbitrary **2**, important **5**, deep
**0**, legacy dark **0** are unchanged. No budget increase, transfer or new exception.

### Verification status

Added 20 Unit cases and nine real Nuxt cases for presentation, route selection,
trend activation/cache/race/retry and old Site retry isolation. Added 26 Functional
Browser cases using the existing deterministic Nitro/upstream owner. P1–P3/Entity
assertions now explicitly account for the authorized first trend and new hooks,
retaining original Site request/failure/SEO guarantees and strict error capture.

Frozen install, Unit (198), Nuxt (20), combined Vitest (218), lint, stylelint,
style-policy tooling (75), exact policy, typecheck, Insights semantics, SEO recovery,
production build and Chromium installation have passed. The first combined P1–P6,
Entity and SEO run passed 168/169. The new cache-remount test issued Back before
its Overview navigation settled; it now waits for the semantic panel state before
Back, retaining canvas/cache assertions without sleeps or retries. All 26 P6 tests
then passed with ten temporary review captures (36 cases, 31.5 seconds). The capture
test was deleted before full acceptance; its images are outside the repository,
not a runner or accepted baseline. Final `pnpm run test:browser --workers=1`
passed **557 cases in 13.2 minutes**, with no failures, skips or retries. The final
unchanged production build passed P1–P6, Entity, SEO, failure boundaries and the
remaining frontend suite. Final lint/stylelint/typecheck and combined Vitest (218)
also passed. Protected Visual/PNG/suppression files remain unchanged.

Inventory is 557 Functional Browser / 119 Visual / 118 accepted PNGs. No accepted
Visual files are changed and no P8 golden is generated. Remote CI and the pinned
Linux Visual comparison are unverified; prior CI runs are not current acceptance.

### Required maintainer review before P7

Review the seven-row desktop matrix, supported/stale/unknown/missing facts,
30d/all trends, Site unavailable and trend unavailable independently, full recent
changes, mobile stacking and representative Dark views. Confirm Fact/Adoption/
Coverage separation, Site/Target scope, chart prominence and change-list density.
Local tests and screenshots do not substitute for P6 exit criterion 37: maintainer
visual acceptance. Do not enter P7 automatically.

### Home entry development-runtime correction (2026-09-27)

Maintainer reproduction at `/site/106?domain=srk.games` exposed a client setup
failure: the running Vite transform of SitePrimaryTabs lacked the `useI18n`
import and threw ReferenceError. Both upstream slices and the SSR document were
200. An explicit vue-i18n import repairs this dependency. Actual Home domain
popover navigation in the local development browser now renders all four tabs
without console errors. The earlier production Browser pass did not cover this
development transform failure.

The existing Site Detail Browser owner now also exercises the real Home popover
entry, asserts no document reload and exact Home + Detail/Insights/View accounting.
Its first run exposed a missing saying in the new Home fixture; supplying the
normal Home saying removes that unintended fallback without relaxing diagnostics.
No appearance, style-debt budget or accepted Visual file is changed.
Lint, stylelint, exact style policy, typecheck and production build passed for
this correction; the final focused Site runtime/Shell Browser run passed all
30 cases with zero retries. The full suite and remote CI were not rerun for this
one-import correction. Functional inventory is now 558; Visual remains 119/118 PNGs.

## P7 — Appearance ownership and legacy cleanup

The maintainer's P7 instruction accepts P2–P6 as the visual/behavior baseline.
Work starts on clean `dev`, synchronized with `origin/dev`, including the Home
entry correction above. P7 removes dead ownership only; no P8 work is included.

### Consumer audit and deletion

Repository `rg` searches covered explicit imports, Pascal/kebab Nuxt component
tags, dynamic component/registration/glob paths, application/server source and
tests/scripts. Manifest entries, generated component declarations and historical
documents are not runtime consumers. No dynamic component registry is present.

| Deleted component | Consumer result | Current owner retained |
| --- | --- | --- |
| SiteHealthSummaryPanel | No runtime or test consumer | Overview, Target Context, Observation, Security |
| SiteOverview | No runtime or test consumer; distinct from SiteOverviewWorkspace | P2 Hero and Target Context |
| SiteSignalCards | No runtime or test consumer | Health Strip, Performance, TLS/Certificate |
| SiteChangeEvents | No runtime or test consumer | P6 Recent Site Changes |

No Legacy/Old clones or legacy-only tests remain. The deleted SFCs take their
private formatting/domain/provider helpers, transitions and scoped styles with
them. Shared Site routes, API wrappers and active presentation helpers are kept.

All fourteen `detailTypes.ts` exports were checked separately:

| Type | References before deletion |
| --- | --- |
| DetailInfoItem | Only DetailSection and LightProbeEntry in the same dead file |
| DetailSection | None |
| LightProbeEntry | None |
| SiteHeroBadge | None |
| SiteSignalCard | Only deleted SiteSignalCards |
| ObservationStripItem | None |
| ProtocolTrackEntry | None |
| SecurityHeaderItem | None |
| ChangeEventItem | Only deleted SiteChangeEvents |
| ObservationProtocol | Only ObservationHistoryItem in the same dead file |
| ObservationHistoryItem | None |
| ObservationTone | Only ObservationMetricItem/ObservationInfoItem in the same dead file |
| ObservationMetricItem | None |
| ObservationInfoItem | None |

The whole file is removed, plus the separately confirmed unused
`SiteObservationHistory` return-type alias. That composable's executable code,
cache, requests and public behavior are unchanged. Other exported helpers/types
without external consumers still have internal consumers and are preserved.

### i18n, suppression and documentation

Exact namespace and dynamic-prefix searches found `site.overview.*` and
`site.healthSummary.*` only in deleted components. Retired performance/HTTP/DNS
labels and the old title have no remaining consumers. Both locales remove 84
dead leaves, preserving `site.siteDnsPanel.none` for the active Home popover.
Parsed before/after comparisons preserve every other namespace and message,
including siteDetail/siteOverview/siteObservation/siteSecurity/siteIntelligence.

No deleted component has a suppression entry; `eslint-suppressions.json` is
unchanged. The active useSiteDetailPage, service and shared-code suppressions stay.
The historical style-system table now identifies Site dark/deep exceptions as
retired. Contract, Agent router and frontend guides distinguish historical #124
inventories from current #109 closure without rewriting old acceptance results.

### Appearance and exact debt delta

Every class and token declared by `site-detail.less` still has an active consumer;
no healthy rule is removed. `insights.less` has no residual Site-only selector.
All CSS/Less files, active Site components, APIs, route owners, presentation
helpers, SSR/SEO code, tests and Visual sources remain byte-identical to baseline.
The only active composable edit removes the unused type alias described above.
There is no spacing/color/layout/animation change or active appearance migration.

After deletion, policy failed on exactly seven stale Site rule/file pairs:
Tailwind 42+85+20, arbitrary 1+1 and raw 22+11. Only then was
`style:policy:update` run; its diff removes precisely those pairs and no exception.

| Debt | Before | After |
| --- | ---: | ---: |
| Site Tailwind appearance | 147 | 0 |
| Site arbitrary appearance | 2 | 0 |
| Site raw visual | 33 | 0 |
| Deep selector | 0 | 0 |
| Legacy dark entry | 0 | 0 |
| Ambient raw, out of scope | 75 | 75 |
| #108 important, out of scope | 5 | 5 |

The final manifest contains no `app/components/site/*` entry. No budget was
raised, transferred, renamed or hidden. Visual inventory remains 119 tests and
118 accepted PNGs; hashes of all 141 Visual source/config/PNG files are unchanged.

### Verification and handoff

Frozen installation, lint, stylelint, style-policy tooling (75), exact policy,
Unit (198), Nuxt (20), combined Vitest (218), typecheck, Insights semantics and
SEO recovery, production build and Chromium installation passed. The full
`pnpm run test:browser --workers=1` run passed **558/558 in 13.4 minutes**, with
zero failures, skips or retries. All six P1–P6 Site contract owners remain intact
and pass alongside the rest of the frontend. Remote CI and pinned Linux Visual
comparison remain unverified for P7.

Maintainer lightweight smoke is still required on Desktop and Mobile: Overview,
Observation, Security and Insights; confirm Hero/Health/Target/Tabs render as before
with no layout collapse, unexpected appearance change or console error. No full
appearance review is requested because active appearance is unchanged. Automated
verification does not stand in for that maintainer smoke. P8 remains separate and
no final Site Visual golden has been created or updated.
Engineering exit criteria are satisfied; the maintainer smoke criterion remains
pending. P8 can be requested after that confirmation; it has not been started.
