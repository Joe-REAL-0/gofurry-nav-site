import { mkdir } from 'node:fs/promises'
import { join } from 'node:path'
import type { Page } from '@playwright/test'
import { test, expect, openRuntime, settleRuntime, assertRuntimeSurface } from '../fixtures/site-detail'

// Optional review artifacts use the Functional Browser owner, never Visual goldens.
async function review(page: Page, name: string) {
  const directory = process.env.GOFURRY_SITE_DETAIL_REVIEW_DIR
  if (!directory) return
  await mkdir(directory, { recursive: true })
  await settleRuntime(page)
  await page.evaluate(() => window.scrollTo(0, 0))
  await page.screenshot({ path: join(directory, name + '.png'), fullPage: true })
}
async function open(page: Page, path = '/en/site/41') {
  const view = page.waitForResponse(response => new URL(response.url()).pathname.endsWith('/sites/41/view'))
  await openRuntime(page, path)
  await (await view).finished()
}

for (const width of [390, 768, 1440]) for (const theme of ['light', 'dark'] as const) {
  const suffix = `${width}-${theme}`
  test(`final shell, Overview and Observation presentation ${suffix}`, async ({ page, context, runtime }) => {
    runtime.state.observationRich = true; runtime.state.fullCapabilities = true; runtime.state.manyChanges = true; runtime.state.historyCount = 100
    await page.setViewportSize({ width, height: 900 }); await context.addInitScript(value => localStorage.setItem('theme', value), theme)
    await open(page)
    await expect(page.locator('[data-site-health]')).toHaveCount(6)
    await expect(page.locator('[data-site-health] .site-detail-note')).toHaveCount(0)
    await expect(page.locator('[data-site-health="certificate"] .site-detail-health__value')).toHaveText('Not observed')
    await expect(page.locator('[data-site-target-trigger]')).toContainText('target.example')
    await expect(page.locator('[data-site-target-context]')).not.toContainText(/shared_canonical|same_host|redirect_to_external|confidence|cloudflare/)
    const hero = await page.locator('[data-site-hero]').evaluate(node => {
      const description = node.querySelector('.site-detail-hero__description')!, meta = node.querySelector('.site-detail-hero__meta')!
      const icon = node.querySelector('.site-detail-hero__icon')!
      return { description: description.getBoundingClientRect().width, meta: meta.getBoundingClientRect().width,
        clamp: getComputedStyle(description).webkitLineClamp, icon: icon.getBoundingClientRect().width }
    })
    expect(hero.description).toBeCloseTo(hero.meta, 0); expect(hero.clamp).toBe('3')
    if (width >= 768) expect(hero.icon).toBeGreaterThanOrEqual(72)
    await expect(page.locator('[data-site-capability-snapshot] [data-site-overview-insights]')).toBeVisible()
    await expect(page.locator('[data-site-change][data-site-change-category]')).toHaveCount(4)
    await assertRuntimeSurface(page, '[data-site-detail]', theme); await review(page, `overview-${suffix}`)
    await page.locator('[data-site-primary-tab="observation"]').click()
    await expect(page.locator('[data-site-observation]')).not.toContainText('Current Target ·')
    await expect(page.locator('[data-site-observation-nav]')).toHaveCSS('border-radius', '8px')
    await expect(page.locator('[data-site-observation-protocol]')).toHaveCount(3)
    await expect(page.locator('[data-site-observation-risks]')).toContainText('DNS observations are missing or stale')
    await expect(page.locator('[data-site-observation-risks]')).not.toContainText('后端旧中文')
    await review(page, `observation-overview-${suffix}`)
    for (const view of ['performance', 'http', 'dns', 'web']) {
      await page.locator(`[data-site-observation-tab="${view}"]`).click()
      if (view === 'performance') {
        await expect(page.locator('[data-site-performance-chart]')).toHaveAttribute('data-site-chart-ready', 'true')
        await expect(page.locator('[data-site-performance] [data-site-performance-sample="20"]')).toHaveAttribute('aria-pressed', 'true')
      }
      if (view === 'http') {
        await expect(page.locator('[data-site-http-redirects] li')).toHaveCount(2)
        await expect(page.locator('[data-site-http-redirects] li').first()).not.toContainText('1.')
        await expect(page.locator('[data-site-http-redirects] svg[aria-hidden="true"]')).toHaveCount(1)
      }
      if (view === 'dns') {
        await expect(page.locator('[data-site-dns-risks] h3')).toHaveText('DNS Observation Signals')
        for (const code of ['ptr_empty', 'low_ttl', 'collector_reported_signal']) await expect(page.locator(`[data-site-dns-signal="${code}"] [data-tone]`)).toHaveAttribute('data-tone', 'neutral')
        for (const code of ['private_ip', 'nxdomain_with_answer']) await expect(page.locator(`[data-site-dns-signal="${code}"] p[data-tone]`)).toHaveAttribute('data-tone', 'warning')
      }
      if (view === 'web') {
        for (const probe of await page.locator('[data-site-web-probe]').all()) {
          expect(await probe.locator(':scope > dl > div').count()).toBeLessThanOrEqual(5)
          await expect(probe.locator('details')).not.toHaveAttribute('open')
        }
      }
      await assertRuntimeSurface(page, '[data-site-detail]', theme); await review(page, `observation-${view}-${suffix}`)
    }
    expect(runtime.count('/sites/41/detail')).toBe(1); expect(runtime.count('/sites/41/insights')).toBe(1); expect(runtime.count('/sites/41/view')).toBe(1)
    expect(runtime.count('/observations')).toBe(1); expect(runtime.count('/trend')).toBe(0); expect(runtime.calls).toHaveLength(4)
    runtime.assertQuiet()
  })

  test(`final Security evidence presentation ${suffix}`, async ({ page, context, runtime }) => {
    Object.assign(runtime.state.security, { enabled: true, days: 12, txt: 'issues', long: true })
    await page.setViewportSize({ width, height: 900 }); await context.addInitScript(value => localStorage.setItem('theme', value), theme)
    await open(page, '/en/site/41?tab=security')
    await expect(page.locator('[data-site-security-scope]')).toHaveCount(1)
    await expect(page.locator('[data-site-security]')).not.toContainText('Current Target ·')
    await expect(page.locator('[data-site-security-summary="headers"]')).toContainText('6 / 6 observed')
    await expect(page.locator('[data-site-security-overview]')).not.toContainText(/X-Frame-Options|Content-Security-Policy|Security Score/)
    await assertRuntimeSurface(page, '[data-site-detail]', theme); await review(page, `security-overview-${suffix}`)
    for (const view of ['tls', 'web', 'exposure']) {
      await page.locator(`[data-site-security-tab="${view}"]`).click()
      await expect(page.locator('[data-site-security-scope]')).toHaveCount(0)
      if (view === 'tls') {
        await expect(page.locator('[data-site-certificate-expiry]')).toHaveAttribute('data-tone', 'warning')
        await expect(page.locator('[data-site-health="certificate"] .site-detail-health__value')).toHaveAttribute('data-tone', 'warning')
        await page.locator('[data-site-certificate-crypto] summary').click()
        await expect(page.locator('[data-site-certificate-crypto] [data-site-evidence="cert_fingerprint_sha256"]')).toBeVisible()
      }
      if (view === 'web') {
        await expect(page.locator('[data-site-security-txt-validation]')).toContainText('Contact information is missing or invalid')
        await expect(page.locator('[data-site-security-txt-validation]')).not.toContainText('contact_missing_or_invalid')
        await expect(page.locator('[data-site-security-headers] > div').first().locator('[data-site-security-raw-headers]')).toBeVisible()
      }
      if (view === 'exposure') {
        await expect(page.locator('[data-site-port-result][data-state="open"] dd[data-tone]')).toHaveAttribute('data-tone', 'neutral')
        await expect(page.locator('[data-site-waf-canary]')).toContainText('1 / 1 expected blocking behaviors matched')
        await expect(page.locator('[data-site-waf-canary]')).not.toContainText(/WAF Enabled|WAF Detected|Protected by WAF/)
      }
      await assertRuntimeSurface(page, '[data-site-detail]', theme); await review(page, `security-${view}-${suffix}`)
    }
    expect(runtime.calls).toHaveLength(3); runtime.assertQuiet()
  })

  test(`final Site intelligence analysis presentation ${suffix}`, async ({ page, context, runtime }) => {
    runtime.state.fullCapabilities = true; runtime.state.intelligence.rich = true
    await page.setViewportSize({ width, height: 900 }); await context.addInitScript(value => localStorage.setItem('theme', value), theme)
    await open(page)
    expect(runtime.count('/trend')).toBe(0)
    await page.locator('[data-site-primary-tab="insights"]').click()
    await expect(page.locator('[data-site-insight-trend]')).toHaveAttribute('data-site-insight-trend-state', 'ready')
    await expect(page.locator('[data-site-insights-workspace] > .site-detail-surface')).toHaveCount(4)
    await expect(page.locator('[data-site-insight-group] > h4')).toHaveText(['Network', 'Transport', 'Web policy'])
    await expect(page.locator('[data-site-capability-matrix] [data-site-capability]')).toHaveCount(7)
    await expect(page.locator('[data-site-capability="tls13"]')).not.toContainText('Transport')
    await expect(page.locator('[data-site-capability-detail] [data-site-insight-trend]')).toHaveCount(1)
    await expect(page.locator('.site-detail-segmented [data-site-insight-range="30d"]')).toHaveAttribute('aria-pressed', 'true')
    const neutral = await page.locator('[data-site-capability-adoption], [data-site-capability-coverage]').evaluateAll(nodes => nodes.map(node => getComputedStyle(node).color))
    expect(new Set(neutral).size).toBe(1)
    await expect(page.locator('[data-site-insight-change]')).toHaveCount(6)
    await expect(page.locator('[data-site-insights-workspace]')).not.toContainText(/score|ranking|percentile/i)
    await assertRuntimeSurface(page, '[data-site-detail]', theme); await review(page, `insights-${suffix}`)
    expect(runtime.count('/trend')).toBe(1); expect(runtime.calls).toHaveLength(4); runtime.assertQuiet()
  })
}

for (const locale of ['en', 'zh']) test('localized multi-target Attention ' + locale, async ({ page, runtime }) => {
  runtime.state.summaryScenario = 'mixed'; runtime.state.extraTargets = ['third.example']
  await open(page, locale === 'en' ? '/en/site/41' : '/site/41')
  await expect(page.locator('[data-site-overview-attention] li')).toHaveText([locale === 'en' ? '2 targets · HTTP is currently unreachable' : '2 个目标 · HTTP 当前无法访问'])
  await expect(page.locator('[data-site-overview-attention]')).not.toContainText('后端旧中文')
  await review(page, `overview-attention-${locale}`)
  expect(runtime.calls).toHaveLength(3); runtime.assertQuiet()
})
