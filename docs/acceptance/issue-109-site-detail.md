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

## Task B — final Site Detail presentation refinement (2026-09-27)

Task B follows the completed P1–P7 and Task A baseline on `dev`. It changes Site
Detail presentation only. Collector, Backend, API/schema, page/composable/service
runtime, route-state helpers, SSR/SEO, Target selection, history/trend cache and
retry/request ownership remain unchanged. P8 has not started.

### Shared presentation and workspace changes

The existing `site-detail.less` owner now supplies compact translucent Primary
and Secondary surfaces, 8–10px corners, fine borders, restrained evidence colors
and one segmented pattern for secondary navigation, samples and ranges. Native
disclosures remain the third level. Phosphor provides all newly introduced system
icons; no dependency, handwritten SVG or cross-domain appearance owner is added.

- Hero: multi-row 80px desktop / 56px mobile icon, full-column three-line
  description, Name → Domain → Meta → Description order, localized formatted
  view count and compact Visit action.
- Health Strip: exactly six primary values, no secondary helper rows, six desktop
  columns / mobile 2×3. Certificate expiry uses P5's evidence normalization and one
  shared presentation helper; unobserved evidence never becomes zero or a dash.
- Target Context: current Target selector, Ping/HTTP/DNS rows and observed time.
  Relation/provider debug dumps are removed from rendering. The visible selector
  keeps the loaded Target during a pending switch, with the existing separate
  pending message; selection, keyboard/focus and race behavior remain intact.
- Overview: light Site health, conditional Attention, grouped Capability surface
  owning ecosystem/retry actions, and four categorized recent changes. Known
  health codes use frontend zh/en copy and repeated Target reasons group by code;
  unknown-code fallback remains. No frontend health classifier or blacklist is
  introduced and backend Chinese reason messages are not rendered in English.
- Observation: compact protocol/KPI summaries, segmented navigation, waterfall
  surface, 232px history chart, visual redirect chain, dense headers/records and
  two-column Web probes with four/five primary facts and disclosure details. DNS
  labels distinguish informational PTR/TTL/other evidence from private-IP and
  NXDOMAIN-with-answer warnings, without hiding raw flags.
- Security: scope note only in Overview; six concise summaries, separate
  verification/validity, combined certificate identity, header matrix and mapped
  security.txt validation copy. Port and canary summaries remain evidence-only;
  open ports, absent headers, OCSP/SCT and matching canaries imply no safety score
  or WAF deployment verdict.
- Insights: four surfaces (context, grouped seven-row matrix, selected capability
  plus trend, full recent-change set). Adoption/coverage/date are neutral; fact
  state alone is colored. Short mobile dates preserve full semantic dates; null
  stays a single dash. Help icons replace persistent explanatory paragraphs.
  Ecosystem trend remains adoption-only with the existing lazy/cache/race/retry
  lifecycle and `connectNulls=false`.

`siteDetailPresentation.ts` owns shared localized reason labels and display tones.
Security owns shared certificate normalization/expiry and validation mappings;
existing workspace presentation owners retain raw evidence interpretation. All
other locale namespaces are unchanged. Contracts, Agent context and the testing
guide describe the current Task B presentation without rewriting P1–P7 results.

### Tests, debt and review boundary

The new Unit owner verifies latency/HTTP boundaries, shared expiry thresholds,
unobserved/default evidence, TLS 1.2/1.3, bilingual reason/validation mappings and
informational/warning DNS flags. Existing presentation tests cover grouped
Attention, seven registry rows, categories, missing/zero/date precision and
evidence-only Security. Existing Nuxt runtime tests are unchanged.

`site-detail-refinement.spec.ts` adds 20 Functional Browser cases using the same
deterministic Nitro/upstream fixture: three workspace families across
390/768/1440 × Light/Dark, plus zh/en grouped Attention. The six existing P1–P6
owners retain request accounting, route/history/keyboard, pending/race, isolated
failure, cache/retry, SSR and SEO assertions. Only assertions for intentionally
replaced presentation are updated; pending selector identity gains an assertion.

An initial focused run caught a duplicated null-date dash caused by two responsive
date spans. The rendering was corrected, preserving the original assertion.
Screenshot review also corrected cramped mobile English primary tabs and spacing
between mobile adoption/coverage labels and values.

No style baseline updater was run. Site Tailwind appearance, arbitrary appearance
and raw visual debt remain **0 / 0 / 0**, deep selector **0**, legacy dark **0**.
Ambient raw **75** and #108 important **5** remain untouched. The debt manifest,
suppression manifest, dependencies, lockfile and accepted Visual inventory are
unchanged; all **118 PNGs** compare byte-for-byte with the baseline (141 tracked
Visual source/config/PNG files). No final Site Detail golden is created.

Frozen installation, lint, stylelint, policy tooling **75**, exact policy, Unit
**230**, Nuxt **20**, combined Vitest **250**, typecheck, Insights semantics, SEO
recovery, production build and Chromium installation have passed locally.
The final `pnpm run test:browser --workers=1` run passed **578/578 in 14.5 minutes**,
with zero failures, skips or retries, including all six P1–P6 Site owners and the
20 new UI cases. Remote CI and pinned Linux Visual comparison have not been
verified for Task B.

The optional `GOFURRY_SITE_DETAIL_REVIEW_DIR` export uses the existing Functional
owner to produce 68 temporary screenshots: eleven workspace states at all six
viewport/theme combinations, plus zh/en Attention. These are deterministic
fixture review artifacts outside the repository, not production screenshots,
accepted Visual PNGs or a second success-screenshot runner.

Maintainer full visual review is still required: Hero identity/description and
actions; six Health values; Target selector and protocol density; Overview normal
and Attention states; all Observation and Security views; grouped Insights and
analysis; long values/disclosures, mobile wrapping and Light/Dark readability.
Automatic checks and agent screenshot inspection do not replace this acceptance.
Task B engineering exit criteria are satisfied. Full acceptance and entry to P8
still require that maintainer visual confirmation; P8 has not been started.

## First-round visual composition completion (2026-09-27)

This round starts from the completed Overview and Observation composition work
on `dev`. It intentionally replaces the earlier card-border/divider appearance
descriptions, including Task B's four Insights surfaces, while preserving the
P1–P6 runtime and evidence contracts. It is neither P8 nor a second design round.

### Material and composition

The existing Site Detail Less owner supplies shared acrylic Plane A/B/C,
hover/selected and 500ms motion tokens. Content surfaces and facts/matrix/list
rows lose their perimeter borders and decorative dividers. Rows use compact
padding, small corners and 4–6px gaps; selection remains immediate. Focus rings,
the primary active indicator, actual relationship/chart lines and necessary
Target control affordances remain.

- Hero and six Health values share one borderless Identity Note. Its inset CSS
  dashed separator has three aria-hidden perforations. Tags are lighter and the
  Health grid has no rules. Certificate thresholds and Not observed are intact.
- Target Context, Overview composites/change cards and all five Observation
  views consume the same material family. Their existing content composition,
  evidence order and ownership remain intact.
- Observation and Security own their H2 and finite secondary nav in one row,
  including mobile. Overflow stays inside the scrollbar-hidden nav; existing
  tab roles, focus/keyboard behavior and route helpers are unchanged.
- Security retains four evidence-only views: one six-summary plane; one
  TLS/verification/validity composite plus identity and crypto disclosures;
  name/state-over-value header evidence plus security.txt; compact Port
  Observation and Request Behavior Check. Metadata and cases remain native
  disclosures. Open ports, missing headers, OCSP/SCT and matching requests imply
  no score, safety conclusion or WAF deployment verdict.
- Insights has an unboxed workspace header, seven-row Capability Explorer,
  one primary Analysis plane and an unboxed full Recent Change Stream. Retry is
  inline. Adoption/coverage/date stay neutral and the adoption-only chart uses
  the info accent with faint grid/area, borderless tooltip and no legend.

Only Security/Insights locale namespaces change, in both zh/en; new icons use
the existing Phosphor package. The pure Security presentation owner exposes
actual TLS version and evidence tones and puts configured ports in metadata
while showing skipped count in the compact summary. Certificate normalization
and all health/Attention interpretation remain unchanged.

### Runtime, tests and protected inventory

All existing runtime hooks are retained. Page, Hero, Target selector and
secondary-nav business scripts are unchanged. No Collector/Backend, API,
route-state, service/composable, SSR/SEO, request accounting, Site/Target identity,
history/trend lazy/cache/race/retry owner is modified.

The existing refinement Browser owner now has 23 cases. It checks computed
translucency and borderless content, Identity Note, title/nav geometry,
responsive keyboard scrolling, immediate selection, spaced rows, neutral
Security evidence, the single Insights analysis plane and inline retry request
isolation. Short Chinese nav labels fit at 390px; an additional 320px viewport
exercises actual keyboard-driven overflow in both locales. Three existing
Browser owners update only superseded appearance/copy expectations. Existing
request, failure, keyboard, race, chart-state and SEO assertions are retained.
Six Unit cases cover actual TLS presentation, compact port evidence and
security.txt state tones. Nuxt runtime tests are unchanged.

No policy updater is run. Site appearance/arbitrary/raw debt stays **0/0/0**;
deep selector and legacy dark stay **0**. Ambient raw **75** and #108 important
**5**, debt/suppression manifests and the lockfile are unchanged. All **118
accepted Visual PNGs** and their source/configuration inventory are unchanged.
No final Site Detail golden is created or updated.

Frozen install, lint, stylelint, policy tooling **75**, exact style policy, Unit
**243**, Nuxt **20**, combined Vitest **263**, typecheck, Insights semantics, SEO
recovery, production build and Chromium installation have passed locally.
The final `pnpm run test:browser --workers=1` run passed **593/593 in 14.6 minutes**,
with zero failures, skips or retries, including all P1–P6 Site owners and the 23
refinement cases. Remote CI and pinned Linux Visual comparison remain unverified
for this change; local results are not remote acceptance.

### Visual review and exit boundary

The existing optional Functional screenshot export produces **68 temporary
fixture screenshots** outside the repository: Overview, all five Observation
views, all four Security views and Insights across 390/768/1440 × Light/Dark,
plus zh/en populated Attention. A local gallery links those images; it does not
generate screenshots or introduce a runner. Agent inspection does not replace
maintainer approval.

Maintainer review remains pending for 1440 Light/Dark, 768 Light and 390
Light/Dark across all workspaces. Check material consistency, Identity Note and
perforation restraint, absence of decorative outlines, row spacing/hover,
same-row mobile nav, Security neutrality, long values/disclosures and
Explorer + Analysis hierarchy. Do not enter P8, begin a second round or update
final goldens as part of this task.

First-round engineering exit criteria are satisfied. The full exit criteria
remain open only for maintainer visual acceptance; no approval is inferred from
automated checks or agent screenshot inspection.

## Task C — second-round consistency and entry refinement (2026-09-27)

Task C is explicitly authorized after first-round completion on `dev`. It keeps
all first-round workspace compositions while replacing opacity falloff with one
fixed content material and intentionally changing the default route entry.

### Presentation and entry contract

- `--site-detail-panel` supplies every content panel's identical base hue and
  transparency within a theme (46% mix of the existing public surface). Ghost
  disclosures and flows are transparent. Plane A/B/C and workspace tint/hover
  aliases are removed. One 48% accent-soft hover and stronger 72% selected mix
  serve all workspace rows/segments; 500ms color-only hover and immediate
  selection remain. No global primitive or #108 owner changes.
- The local Target trigger is borderless/transparent at rest, 38px minimum
  height, with shared hover/open fill and existing focus ring. Its selection,
  listbox, keyboard, outside dismissal and focus-return script is unchanged.
- Hero puts formatted view count below Visit with stronger warm, non-semantic
  emphasis. Its CDN badge consumes only Current Target `edgeProviderHints`:
  explicitly typed CDN, medium/high confidence, high first and stable tie order.
  WAF/hosting/reverse-proxy hints and low-confidence CDN hints never qualify.
  Known labels are localized; raw confidence/evidence stays out of Hero.
- `siteTargetSignals.ts` is the shared pure owner for CDN selection and protocol
  display status. Observation Overview and Context use identical 8px status dots;
  ordinary success keeps screen-reader text, exceptions stay visible. Slow
  success retains latency tone independently. No health conclusion changes.
- Blank/domain-only routes and `tab=observation` now mean Observation/Performance.
  Builders omit both defaults; Site Overview requires `tab=overview`, and
  Observation Overview requires `tab=observation&view=overview`. Invalid UI state
  falls back to the new default; business Target validation stays authoritative.
  Primary Observation enters Performance from another workspace and retains
  the current view when already active.
- Default SSR renders the Performance loading surface using only Detail and
  Site Insights. Hydration adds exactly one Ping history request with
  `protocol=ping`, `limit=100`, `payload_mode=preview`, then slices 20 locally.
  Existing P4 cache/race/retry internals and P6 trend runtime are untouched.
  Non-Performance Target switches still request Detail only; active Performance
  may load a new Target's uncached Ping history. Neither repeats Insights/View.

API/DB/Collector/Backend, Security runtime, Site/Target ownership and Entity-only
canonical/hreflang/sitemap are unchanged. No P8 or final golden is introduced.

### Verification and review boundary

Route Unit tests cover defaults, explicit Overviews, invalid UI state, builders
and round trips. New pure tests cover confidence/type selection, bilingual CDN
copy, shared stale/failure/unknown status and slow-success latency separation.
The existing Browser owners add default-entry SSR/hydration/reload/back/forward
and Home client-entry accounting, reliable/absent CDN, Target badge updates,
accessible success, visible exceptions and material consistency across all
workspaces at 390/768/1440 Light/Dark. Existing Overview scenarios now select
Overview explicitly; request/error assertions are preserved. The independent
Insights Entity fixture adds only the exact history boundary its new entry needs.

Frozen installation and Chromium installation passed. Unit **268**, Nuxt **20**,
combined Vitest **288**, policy tooling **75**, lint/stylelint, exact style policy,
typecheck, Insights semantics, SEO recovery and the final production build passed.
The final `pnpm run test:browser --workers=1` run passed **607/607 in 15.5 minutes**,
with zero failures, skips or retries. This includes all existing Site runtime
owners and **33** refinement cases (10 added in Task C). The separate focused
entry/refinement run also passed **55/55** before the full run.

Site appearance/arbitrary/raw debt, deep selectors and legacy dark remain zero.
Ambient raw 75 and #108 important 5 are preserved without running the baseline
updater. Debt/suppression manifests and the lockfile are unchanged. All **118
accepted Visual PNGs** are byte-identical; their source/configuration inventory
is unchanged. The existing Functional export supplies **86 temporary screenshots**,
including default Performance, protocol rows and selector hover. Its local gallery
only links those images; it is not a runner or a Visual golden.

Maintainer review is pending for 1440 Light/Dark, 768 Light and 390 Light/Dark:
uniform panel material, visible consistent hover, integrated selector, warm view
count, reliable CDN badge, protocol alignment/exception text and the new entry.
Remote CI remains unverified. This task stops before P8.

Task C engineering exit criteria are satisfied. Full exit remains pending only
for maintainer visual acceptance; agent screenshot review and automated checks
do not imply that approval.


## Task D — composites, help and focused defects (2026-09-27)

Task D runs directly on current `dev` after Task C. It preserves the accepted
material and default Observation/Performance entry. It is not P8 and does not
claim final visual or production acceptance.

### Implemented ownership

- Overview has one Summary/conditional Attention/Capabilities panel. Summary
  fields contain only label/value; distribution and Site-wide helper rows are
  removed. Observation Overview similarly combines Current/Endpoint/Attention.
  Conditional evidence sections use quiet inset dashed separators, without
  copying Hero perforations. DNS signals live inside the ledger; their existing
  informational/warning projection and Target-health Attention source are intact.
- HTTP Response/optional Redirects/Common Headers share one panel. Redirects
  use three-column serpentine placement on desktop and vertical arrows on mobile;
  missing redirects remove both the section and its separator. No hop status is
  inferred. Performance loses its Waterfall heading/help, with space between the
  KPI strip and timings. The API still requests 100 Ping rows once and slices
  20/60/100 locally, defaulting to 20.
- Hero name/domain share a row, metadata/views share the next, and description
  keeps the content column. Current Target CDN selection is unchanged. Security
  Summary/Attention share a panel. TLS/verification/validity are label/value only;
  auxiliary evidence moves into closed native details. Certificate Identity and
  closed Crypto disclosure share one panel. Visible Security/Insights scope
  paragraphs are removed, without changing their Target/Site ownership.
- `SiteChangeStream` owns local Compact/List display for Overview (at most four)
  and Site Insights (complete recent set). Only placement/connector functions are
  shared with Game Timeline. DOM chronology, date precision, categories and unknown
  fallback remain; mobile uses a list with hidden mode controls. The mode selection
  uses immediate shared selected fill; hover remains 500ms without motion.
- `SiteDetailHelpTooltip` replaces active Info/native-title help. It supports
  pointer, focus, click/touch, Escape and blur, has `role=tooltip`/`aria-describedby`,
  inherits Site acrylic tokens through a page-local Teleport, constrains viewport
  width and updates placement during scroll. Responsive arrows have one visible
  icon each; Tailwind visibility lives on wrappers to avoid the icon display rule.
- The four audited charts use finite numeric normalization. Site adoption tooltip
  and Ping values render a dash for absent/invalid evidence. Game player tooltip
  does likewise; malformed priced amounts remain gaps rather than zero-price
  points. Existing explicit unknown/free price semantics remain. Game average has
  at most one decimal, while current/peak remain integer. No chart library added.
- The Gallery fixture registers one exact active trailer request before intentional
  unmount and requires its actual `net::ERR_ABORTED`. A decoded real video is
  reloaded behind the existing release gate to make that cancellation deterministic.
  Request identity, exact URL, error and unmount timing are checked; every other
  failed request remains fatal. No generic browser-error suppression was changed.

Collector/Backend/API/schema, route keys, health/reason projection, SSR/SEO owners,
P4/P6 activation/cache/race/retry and Site/Target identities are unchanged.

### Focused verification actually performed

The task explicitly prohibits a full build and full test suites. The existing
fixture therefore gained opt-in `GOFURRY_FIXTURE_DEV=1`: local Nuxt source with
isolated build/cache paths and the same deterministic upstream/error accounting.
The flag is rejected in CI; the default production fixture is unchanged. These
are **development-runtime results**, not validation of a new production bundle.

Two final focused runs passed **13/13** and **3/3** (15 distinct cases; the no-redirect
case was repeated after adding its absent-separator assertion), with zero retries:

- Task D formatter boundaries, Overview/Observation composites, default history
  sample/request budget, HTTP snake/headers, DNS semantics, Security TLS details,
  help interactions, full/limited changes, Site Target-switch request accounting,
  390 Light and 1440 Dark overflow/visibility checks.
- Real Site ECharts canvas text at the missing middle adoption point includes
  the correct date and `—`, with no undefined/null/NaN/Infinity text.
- Existing Task C blank-entry SSR/hydration, no-redirect/no-CNAME, Target health
  Attention versus raw diagnostics, and Overview 1440 Light composition cases.
- Existing Game real chart theme/tooltip/responsive and Gallery-unmount cases;
  three average fixtures (`35386.394`, `42.5`, `42`), plus actual player/price
  tooltip missing-evidence cases.

Commands used `pnpm exec playwright test` with explicit spec paths, exact `-g`
selection, `--workers=1 --timeout=120000` and the local source flag. Final typecheck,
focused ESLint over **37** changed/new Vue/TS/MJS files, Site Detail Less stylelint,
style policy and `git diff --check` passed. No `pnpm test`, full Browser command,
production build or full theme/route/viewport matrix was run for Task D.

Site Tailwind/arbitrary/raw debt remains **0**, deep and legacy dark **0**.
Ambient raw **75** and #108 important **5** remain unchanged. No budget updater,
dependency upgrade or suppression change. All **118 accepted PNGs** match the
baseline; the **142-file Visual inventory** is unchanged.

### Manual review and exit boundary

The existing Functional screenshot export supplies **14 temporary screenshots**
outside the repository, plus a static index linking them. This is neither a new
runner nor a final golden. Desktop Light covers the touched Site routes; 390 Light
and 1440 Dark are representative checks, not a full matrix. Agent inspection
confirmed combined panels, inset separators, snake arrows, mobile stacking and
absence of overflow in those captures.

Maintainer review remains pending: Hero density, summary two-line hierarchy,
Compact/List and date order, HTTP flow, TLS closed/open disclosures, real tooltip
hover/focus/click/Escape on Site, and average/tooltip display in Game Insights.
Remote CI is **unverified**; earlier Task C CI is not evidence for Task D.
Engineering exit criteria are satisfied within the explicitly focused budget;
manual visual acceptance remains open. Stop here, before P8.


## Task E — Similar Sites discovery (2026-09-27)

Task E runs directly on current `dev` after Task D. It adds Site-level discovery,
not a recommendation score/ranking system, and stops before P8.

### API and candidate ownership

- New independent `/api/v2/nav/sites/:siteId/recommendations` accepts `lang=zh|en`
  and `limit` (default/max eight), registered alongside enabled Detail routes.
  Response schema 1 has UTC generated_at, state, site_id and SiteVo items with
  view_count. Invalid parameters are 400; unavailable reads produce an optional
  unavailable/empty slice. No existing Detail payload is expanded.
- The service reuses the existing localized Site/Group reader and its cache/DB
  fallback. Full group membership identifies eligible groups, even when the
  current Site is below preview rank eight. `BuildHomeGroupsForCache` alone owns
  each group's curated ordering and Top-8. Union/dedupe/self exclusion precede
  SHA-256 ordering by Site ID + candidate ID + UTC date, then the final max eight.
  Shortage stays shortage; there is no rank-nine/global/random backfill.
- No table, migration, recommendation Redis key, scheduler, score or reason field.
  Collector and the existing Home builder/cache owners are unchanged.

### Frontend and request contract

- `useSiteRecommendations` is page-owned, parallel SSR data keyed by Site ID and
  language. Hydration reuses it; Site/locale changes select a new identity, with
  old results guarded. Target, route-only query, local tab and display mode do
  not refetch it. Failures have no error card/retry and do not fail the page.
- One `SiteSimilarSites` instance renders below Current Target on Desktop or as
  Mobile's local panel. Main/Aside proportions are unchanged. ManagedAssetImage,
  name, one-line info, first display domain (including stored Home domain JSON)
  and PhEye/formatted view snapshot are the only item fields. Rows share existing
  borderless material, radius and 500ms hover tokens; no movement or ranking.
- SSR markup filters to SFW; mounted `readDisplayMode`/`subscribeModeChange` only
  filter the raw response locally. NSFW raw payload remains available for mode
  switching, but its items are absent from server-rendered recommendation markup.
  Raw nonempty data controls mobile tab existence; all-filtered Mobile has neutral
  empty copy and Desktop hides the list.
- Four route tabs remain exactly unchanged. Mobile Similar is local-only, does
  not write URL/history/SEO, participates in tab keyboard navigation, and leaves
  the underlying workspace mounted. Resize to Desktop restores that route state.
  The extended tab row scrolls when needed and keeps keyboard selection visible.
- Similar uses localized Entity NuxtLinks and has no view mutation. The destination
  Site page alone posts View; the Browser test verifies exactly one, not zero/two.
- Normal SSR now has Detail + Insights + Recommendations. Hydration keeps the
  existing View and active lazy slice. The document's "Target switch = Detail
  only" shorthand is qualified by the existing P4 Performance exception: a new
  Target may load its uncached Ping history while Performance is active. Task E
  preserves that behavior and guarantees zero recommendation refetches.

The primary route vocabulary/parser, canonical/hreflang/sitemap owners, P4/P6
history/trend/retry owners and health/security evidence semantics are unchanged.
Old Site Browser request-budget assertions and the SEO/Entity fixtures explicitly
include the new slice rather than hiding it from request/error accounting.

### Actual focused verification

Backend, after gofmt of changed Go files:

```text
go test ./apps/nav/recommendations/... ./apps/nav/sitegroup/... ./apps/nav/home/service -run 'Recommendations|Group|Sort' -count=1
```

Passed six recommendation service tests, one controller contract with six request
cases, and three existing Home-group ordering/preview tests. Sitegroup packages
compile but have no test files. Covered multigroup union/dedupe, below-preview
membership, Home Top-8 eligibility, shortage, self exclusion, final limit,
UTC-day/locale/order stability, localization, raw mode and view-count preservation,
and optional missing/empty semantics. No live PostgreSQL/Redis was used.

Frontend used the existing deterministic Nitro/upstream fixture with the Task D
opt-in local source mode (`GOFURRY_FIXTURE_DEV=1`), not stale `.output`. The final
focused run passed **11/11**: ten new recommendation cases plus the existing blank
Task C default SSR/hydration/reload contract. The mobile keyboard/resize case was
then rerun successfully after adding immediate horizontal reveal and its visible
bounds assertion. Zero retries; no fixed sleeps/networkidle; strict error capture
is unchanged. This is development-runtime evidence, not production acceptance.

Coverage includes parallel SSR/SFW markup, exactly one recommendation read,
Desktop Light/Dark, managed image readiness, max eight and view snapshots,
mode filtering without fetch, Target/workspace identity, 390px local tabs and
history/SEO invariance, keyboard/resize, raw-adult empty state, empty/unavailable/
503 isolation, language refetch and exactly-once destination View.

Final typecheck, focused ESLint (19 changed/new TS/Vue files), Site Detail stylelint,
style policy and whitespace checks passed. Site appearance/arbitrary/raw debt,
deep selectors and legacy dark remain **0**. Ambient raw **75** and #108 important
**5** remain unchanged; no baseline updater, dependency or suppression changes.
All **118 accepted PNGs** are byte-identical; the **142-file Visual inventory**
is unchanged. No full frontend suite, full Browser run or production build was run.

### Review and exit boundary

The existing Functional screenshot export produced three temporary images and a
static linking gallery outside the repository: 1440 Light, 390 Light, 1440 Dark.
Agent review corrected missing icon import and narrow English tab spacing; this
is not maintainer approval. Review the aside placement/density, small icons and
view hierarchy, hover/focus, local mobile panel/resize and in-site navigation.

Task E engineering criteria are satisfied within the requested focused budget.
Maintainer visual acceptance remains pending; remote CI is unverified. No P8 or
final golden was started. Both Nav Backend and Nav Web require normal deployment
for the feature; no database or Redis migration is required.

## P8 preflight — contract alignment and CI recovery (2026-09-28)

This preflight is not a new visual round or #109 closure. Its only production
presentation change removes the domain row from Similar Sites and makes its icon
span the remaining two rows. The consumer-free domain parser/import and domain
CSS are deleted. Recommendation API/SiteVo, Home Top-8/daily shuffle, Site/language
identity, mode filtering, localized navigation and destination View remain intact.

Functional contracts now follow Task D: Waterfall retains six timing rows, four
KPIs, total and ready history, and explicitly rejects the retired native-title/
Info help. Exact row gaps, typography dimensions and serpentine pixel placement
leave Functional ownership. Counts, fact states, category/date ownership,
Compact/List controls, accessible tooltips and request/error capture remain.
Redirects inspect ordered nodes and the aria-hidden directional connector wrapper,
not Phosphor's internal SVG structure. No UI was restored to satisfy old tests.

Insights metric/range navigation waits for the URL, pressed controls and displayed
trend identity before readiness. The required sequence remains tls13|90d,
csp|90d, csp|all: three reads before reload, no reads for cached Back/Forward,
and a fourth csp|all read after reload. No request budget is reduced and no force
reload is introduced. `useSiteInsightTrend` and all other runtime owners are
unchanged.

The baseline [Task E CI run](https://github.com/gofurry/gofurry-nav-site/actions/runs/36329241205)
confirmed Browser shards 2/3 failures in these stale/synchronization assertions;
build, image, Visual, Nav Backend and PostgreSQL passed there. That baseline is
diagnostic evidence only, not acceptance of this preflight.

Local source-fixture verification passed **39/39** Chromium cases, workers=1 and
zero retries: six Performance/Waterfall, the metric/range history/cache/reload
case, eighteen final Shell/Security/Insights cases, four recommendation cases,
six affected Task C entry cases and four affected Task D composite/help cases.
The synchronized trend case observed exactly tls13|90d, csp|90d, csp|all, then
csp|all after reload; runtime changes were unnecessary. This is the existing
`GOFURRY_FIXTURE_DEV=1` owner, not a production-build acceptance claim.

Typecheck, focused ESLint, Site Detail stylelint, style policy and whitespace
checks passed. Site appearance/arbitrary/raw debt, deep selector and legacy dark
remain zero; ambient raw 75 and #108 important 5 are unchanged. No budget update,
full local frontend/Browser suite or local production build was run. Accepted
Visual files and the baseline manifest are unchanged.

The new [preflight CI run](https://github.com/gofurry/gofurry-nav-site/actions/runs/36332279244)
completed successfully: repository-policy, nav-web-build, all three Browser
shards, nav-web-visual, nav-web-image and the nav-web aggregate are PASS. The known
Site Detail failures in shards 2/3 are cleared. Go/database jobs were skipped by
the unchanged path rules because this push changes no Backend/Collector code;
the earlier Backend result is not presented as a new execution.

This result is recorded by a documentation-only follow-up; the tested application,
tests, contracts and Visual inventory remain identical to the passing run.
Preflight exit criteria are satisfied and formal P8 can be assigned separately.
No final Visual golden is created or updated, and #109 closure is not claimed.

## P8 — final Visual contract and closure review (2026-09-28)

P8 closes the engineering contract without another product-design round. P1–P7,
Task A–E and preflight form the implemented baseline; their historical acceptance
entries above are not rewritten or promoted into final visual approval. Task C's
Observation/Performance default and single shared acrylic panel supersede earlier
Overview/Plane A–B–C designs. Task D's composites/help/flows, Task E's discovery,
preflight's Name + Info + Views rows without domain, and the certificate fallback
fix are the final product represented here. Collector Task A remains unchanged.

### Final runtime and accessibility audit

- Four URL-owned workspaces remain Overview / Observation / Security / Insights.
  Blank/domain-only URLs select Observation/Performance. Mobile Similar is a local
  fifth panel, with no URL/history/SEO mutation; Desktop restores the route view.
- Site ID owns Insights and exactly-once View; Site/language owns Recommendations
  and the initial Site summary. Target owns Detail protocol/security evidence and
  Ping history. SSR reads Detail + Insights + Recommendations; hydration counts
  View once and loads the active lazy slice only.
- Target switching preserves workspace state and never reloads Site Insights,
  recommendations or ecosystem trend, nor repeats View. Active Performance may
  fetch an uncached new Target's Ping history. The P4/P6 cache, race, activation
  and slice-local retry owners are unchanged.
- Finite explicit certificate days take precedence; otherwise the shared P5
  projection derives whole days from cert_not_after minus HTTP observed_at.
  No browser clock participates. Invalid/missing evidence stays not observed;
  expired/warning/attention/normal thresholds and verification/validity separation
  are unchanged across Health, Security summary and validity details.
- Similar preserves full membership, each group's Home Top-8, union/dedupe/self
  exclusion and deterministic UTC-day selection capped at eight, without global
  backfill. SSR markup is SFW; mode/tab/Target changes filter or navigate locally
  without refetch. Destination navigation alone counts one View.
- Authoritative Detail failure remains separate from optional Insights/trend/
  recommendations failure. Canonical/hreflang/sitemap retain Entity URLs only.
- Existing Functional owners cover primary/secondary tab keys and roving focus,
  local Similar ARIA, Target listbox and focus return, help focus/ARIA/Escape,
  native disclosure keyboard use, textual status and preserved focus-visible.
  P8 adds no accessibility dependency or replacement interaction.

### Dedicated Visual owner and inventory

`tests/browser/fixtures/site-detail-visual.ts` reuses the deterministic local
Nitro/upstream runtime. `tests/browser/visual/site-detail.spec.ts` uses the existing
Visual config and pinned Linux Playwright image. No runner/config is added.

Exactly eight PNGs live in
`apps/cn/nav-web/tests/browser/visual/__snapshots__/site-detail.spec.ts/`:

| Viewport | Golden |
| --- | --- |
| 1440×900 | `site-detail-performance-light-desktop.png` |
| 1440×900 | `site-detail-performance-dark-desktop.png` |
| 1440×900 | `site-detail-overview-light-desktop.png` |
| 1440×900 | `site-detail-security-tls-light-desktop.png` |
| 1440×900 | `site-detail-insights-light-desktop.png` |
| 1440×900 | `site-detail-http-light-desktop.png` |
| 390×900 | `site-detail-performance-light-mobile.png` |
| 390×900 | `site-detail-similar-light-mobile.png` |

These capture the Site Detail root, including content below the viewport; PNG
height may exceed 900px. No arbitrary crop coordinates or masks hide Site content.
Desktop includes the right aside; Mobile Similar is a single column of names,
info and view counts with unchanged URL and no domain line. HTTP has seven redirect
nodes, Overview has health Attention and populated changes, Insights has seven
facts plus a null-gap ecosystem trend, and TLS/Health visibly derive 45 days with
cert_days_left absent.

Evidence and browser time are fixed to 2026-09-27T12:00:00Z. Site/Target payloads,
100 Ping samples, 27 trend points, six changes, eight recommendations and managed
local SVG assets are fixed. The runtime fixture's optional diagnostic timestamp
uses this instant, avoiding unrelated CDN cache expiry when the clock is installed;
other fixtures keep their live timestamps. Strict network/error capture is intact.
An initial preparation attempt caught outbound CDN probes before any PNG was
written; correcting that fixture seed enabled the sole successful generation.
No product bug was hidden or production clock logic changed.

Readiness awaits hydration, View completion, active chart-ready markers, loaded
images, fonts, finite animations and two animation frames; focus is blurred and
the pointer moved away. Overflow and runtime diagnostics are checked before/after
capture. No fixed sleep/networkidle. Only unrelated floating navigation tools are
hidden, following existing Visual fixture ownership.

Visual protects composition/material/spacing/density/type/responsive charts.
Functional retains semantics, requests/cache/race, a11y and errors; exact pixel
gaps are not reintroduced there. The existing 118 PNGs are byte-identical; these
eight make the new inventory 126, without unrelated baseline updates.

### Narrow dead-owner and style closure

Consumer audit covered explicit imports, Nuxt component tags, dynamic names and
tests. No active helper/component or runtime hook was removed. Fifteen unused
selectors were retired: site-capability-row, site-detail-technical,
site-intelligence-change/link, site-observation-chain/protocols/record/sample/
timing-bar/timing-track, site-overview-attention__item/health__status/summary,
site-security-link and site-recent-change. The obsolete intelligence-change
selector also leaves the existing Functional selector list; its effective checks
are unchanged. Active style declarations and product composition are untouched.

Both locales drop only three consumer-free keys: siteIntelligence.coverageHint,
siteObservation.waterfall and siteObservation.timingHint. Current tooltip/help
keys and all active labels remain. No broad i18n/type refactor or legacy clone.

Site Tailwind appearance / arbitrary appearance / raw visual / important / deep /
legacy dark are all **0**. The debt manifest is byte-unchanged, with ambient raw
**75** and #108 important **5** preserved. No budget update/transfer/exception,
suppression change, Backend/Collector/API/schema change or feature is introduced.

### P8 verification and remaining gates

Final local static checks passed: typecheck, stylelint, style-policy tooling
**75/75**, style-policy scan **280 sources**, focused ESLint and whitespace checks.
The three existing certificate/security/Target unit files passed **93/93**, covering
explicit precedence, fallback, invalid evidence, thresholds and no Date.now use.
Frozen install and the production build passed inside the repository's pinned
Linux Playwright/Node 24 image; local Browser/Visual consume that same output,
not the optional dev server or a stale workstation build.

The environment sentinel passed **1/1**. Guarded scoped generation passed **8/8**;
the separate compare run, updates disabled, passed **8/8**. All eight images were
inspected for obvious cropping/overflow; agent inspection is not maintainer approval.

The explicit eight-owner Functional run passed **218/218**, workers=1, retries=0:
site-detail-contract, site-detail-shell, site-overview, site-observation,
site-security, site-insights, site-recommendations and site-detail-refinement.
The existing SEO owner additionally passed **3/3** focused cases covering both
locales' canonical/hreflang/query handling and the Entity-only sitemap. No full
repository Browser suite is run locally for P8.

The [P8 implementation CI run](https://github.com/gofurry/gofurry-nav-site/actions/runs/36335038860)
completed successfully: repository-policy, nav-web-build, all three nav-web-browser
shards, nav-web-visual, nav-web-image and the nav-web aggregate are **PASS**. Go and
database jobs are unaffected/skipped under existing change detection. This is the
new P8 run, not the earlier preflight acceptance. No retry, golden repair or product
change was needed to pass remote CI.

This result is recorded in a documentation-only follow-up, which also routes
`apps/cn/nav-web/AGENTS.md` to the final P8 contract and Visual owners. Application,
tests, fixture and PNGs remain identical to the passing run. The follow-up is still
subject to its own current-head remote gates before delivery; documentation does
not grant an exemption from build/Browser/Visual acceptance.

Maintainer approval of these eight review candidates is **pending**. Review final
quality, Light/Dark material, Hero density, default Performance/chart, Overview,
HTTP snake flow, certificate clarity, Insights density, Desktop aside and Mobile
tabs/Similar. #109 is not yet closure-ready until current-code remote gates and
explicit visual approval pass. No P9 or further design round is started.
