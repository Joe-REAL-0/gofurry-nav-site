# #134 — Frontend CI runtime governance

## Scope and current contract

The implementation splits frontend feedback into explicit cost tiers, preserving
product behavior, dependencies, fixture diagnostics, zero retries and the existing
Playwright image digest. Hosted runners use Ubuntu 24.04. CHANGELOG keeps the prior
#109 notes and records this new CI model under Unreleased.

| Tier | Trigger | Gate |
| --- | --- | --- |
| Fast | Selected dev push / PR | One install + all static/unit/Nuxt/semantic checks + build + seven-owner Smoke; stable `nav-web` depends only on detection and Fast |
| Full | Main push / manual selected ref / nightly dev | One build shared across three explicit regression groups; final `nav-web-full-gate` requires every group |
| Image | Main/manual Full only | Existing Buildx/GHA cache and real deployment context; nightly skip is asserted by the Full gate |
| Visual | Manual selected ref only | One pinned build and all existing comparison specs; no snapshot updates |

Nightly is 20:17 UTC. Its build resolves dev HEAD once; group/image checkouts and
artifact identity consume that exact source SHA. Manual uses the selected workflow
ref without a second ref input. Only the one-day production `.output` archive is
shared. Smoke, each group and Visual retain separate seven-day failure artifacts.

The existing docs-only optimization and conservative CI/tooling/SQL selection
remain. Normal frontend dev/PR edits no longer select full regression, Visual or
image jobs. CI implementation edits can still select Go/database jobs; their
duration is not frontend Fast latency. Repository policy runs alongside Fast.

## Complete inventory

Smoke retains game-detail, hero, nav-home-locales, resource-routing and static-locales.
Only site-detail and games-search are added: seven files, 30 cases. The new cases
reuse the existing production Nitro/local upstream owners, asserting SSR/hydration,
default Site Performance or one basic search with strict diagnostics. No interaction
matrix is copied from regression.

`.github/scripts/nav-web-regression-groups.mjs` is the explicit inventory owner:

| Group | Files | Cases selected by Playwright | Supplied planning test-time estimate |
| --- | ---: | ---: | ---: |
| games | 12 | 110 | ~298s |
| sites-other | 15 | 266 | ~277s |
| insights-nav | 18 | 224 | ~252s |
| Total | 45 | 600 | |

The estimates are not newly measured job times. Inventory tests fail on unowned
new/current files, duplicate ownership (including within one group), deleted or
renamed members, empty groups and unknown CLI selectors. Each job uses the emitted
explicit file list, not a regex or numeric shard. Workers=1/fullyParallel=false/
retries=0 remain. All pre-existing Smoke, regression and Visual owners and PNGs
are unchanged; no appearance debt or application source changes accompany #134.

## Local verification

- `node --test .github/scripts/detect-changes.test.mjs`: 7 passed.
- `node --test .github/scripts/nav-web-regression-groups.test.mjs`: 7 passed.
- `go test ./check-production-policy`: passed, including YAML parsing, trigger/gate
  boundaries, runner pins, exact build-source propagation and seven-owner Smoke.
- `go run ./check-production-policy`: passed for all six production Go modules.
- Fresh Nav Web production build: passed; no development fixture or stale output.
- `pnpm run test:browser:smoke -- --workers=1`: 30/30 passed in 48.4s, zero retries.
- Each group's actual CLI output was passed to `playwright test --list --workers=1`:
  110/266/224 cases in exactly 12/15/18 files. Full regression was not executed locally.
- Focused ESLint on both new Smoke files and whitespace checks: passed.

## Before/after measurements and remote acceptance

The immediately preceding full frontend [P8 run](https://github.com/gofurry/gofurry-nav-site/actions/runs/36335038860)
provides the before measurement, not acceptance of #134:

| Old job / critical path | Measured duration |
| --- | ---: |
| Workflow creation to completion | 12m49s |
| nav-web-build | 2m26s |
| Browser shard 1 / 2 / 3 | 8m28s / 4m38s / 4m04s |
| Visual | 4m23s |
| Deployment image | 2m08s |

Fast's 3–5 minute goal is a target. Record its actual job duration and workflow
creation-to-`nav-web` completion separately; CI-change-only Go/database work must
not be presented as ordinary frontend feedback time.

Remote Fast, manual Full and manual Visual results remain pending until this
implementation is pushed and the respective runs actually finish. Nightly wiring
is locally verified; no real nightly execution is claimed.

GitHub requires new manual/scheduled workflows to exist on the default branch
before they can receive those events. The current default is main; implementation
on dev alone is not proof of enabled dispatch/nightly. See [GitHub's workflow trigger
contract](https://docs.github.com/en/actions/reference/workflows-and-actions/events-that-trigger-workflows).
No default-branch change or main merge is authorized by this acceptance record.
If dispatch is blocked, retain Full/Visual as unverified rather than run a hidden
fallback or restore expensive daily gates. #134 is closure-ready only after Fast,
manual Full (including image) and manual Visual actually pass.
