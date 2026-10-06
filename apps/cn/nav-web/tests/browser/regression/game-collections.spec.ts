import { test, expect, openRuntime, settleRuntime, changeCollectionMode, collectionPath } from '../fixtures/game-collections'

for (const locale of ['zh', 'en']) {
  const prefix = locale === 'en' ? '/en' : ''
  test(`Collection Index SSR SFW, preview cardinality, SEO and localized routes (${locale})`, async ({ page, runtime }) => {
    const html = await openRuntime(page, `${prefix}/games/collections`)
    expect(html).not.toContain('Adult preview')
    expect(html).not.toContain('adult-preview.svg')
    expect(runtime.calls).toHaveLength(1)
    expect(runtime.calls[0]!.url.searchParams.get('mode')).toBe('sfw')
    const cards = page.locator('.game-collection-card')
    await expect(cards).toHaveCount(6)
    for (const [i, count] of [3, 2, 1, 0].entries()) await expect(cards.nth(i).locator('img')).toHaveCount(count)
    await expect(cards.first()).toHaveAttribute('href', `${prefix}/games/collections/collection-3`)
    await expect(page).toHaveTitle(locale === 'en' ? 'GoFurry Game Collections - Curated Furry Game Timelines' : 'GoFurry 游戏分区 - 兽人游戏主题合集与发展脉络')
    runtime.assertQuiet()
  })
  test(`Detail Backend chronology, phase boundaries and date precision (${locale})`, async ({ page, runtime }) => {
    const html = await openRuntime(page, `${prefix}/games/collections/collection-3`)
    expect(html).not.toContain('Adult game'); expect(html).not.toContain('adult-game.svg')
    const nodes = page.locator('.game-collection-timeline li')
    expect(await nodes.evaluateAll(elements => elements.map(el => el.getAttribute('data-game-id')))).toEqual(['21', '19', '7', '8', '30', '40', '41', '42', '43', '50', '60'])
    await expect(page.getByTestId('collection-now')).toContainText(locale === 'en' ? 'October 6, 2026' : '2026年10月6日')
    await expect(page.locator('.game-collection-time').first()).toHaveText(locale === 'en' ? 'Approx. 2004' : '约 2004')
    await expect(page.locator('[data-game-id="41"]')).toContainText(locale === 'en' ? 'Q2 2027' : '2027 Q2')
    await expect(page.locator('[data-game-id="40"]')).toContainText(locale === 'en' ? 'still upcoming' : '当前仍未发售')
    for (const id of ['8', '30', '43', '50', '60']) await expect(page.locator(`[data-game-id="${id}"]`)).toHaveAttribute('data-connector', 'none')
    await expect(nodes.first().getByRole('link')).toHaveAttribute('href', `${prefix}/games/21`)
    expect(runtime.calls).toHaveLength(1)
    runtime.assertQuiet()
  })
}

for (const path of ['/games/collections', '/games/collections/collection-3']) {
  test(`NSFW hydration is one client refresh after SFW SSR: ${path}`, async ({ page, context, runtime }) => {
    await context.addInitScript(() => localStorage.setItem('mode', 'nsfw'))
    const html = await openRuntime(page, path)
    expect(html).not.toContain('Adult game'); expect(html).not.toContain('adult-preview.svg')
    await expect(page.locator('.game-collections-page')).toContainText(path.endsWith('collection-3') ? 'Adult game' : 'Adult collection')
    expect(runtime.calls.map(c => c.url.searchParams.get('mode'))).toEqual(['sfw', 'nsfw'])
    expect(page.url()).not.toContain('mode=')
    runtime.assertQuiet()
  })
  test(`Mode race and refresh failure keep the latest ready content: ${path}`, async ({ page, runtime }) => {
    await openRuntime(page, path)
    const old = runtime.hold(url => url.searchParams.get('mode') === 'nsfw')
    await changeCollectionMode(page, 'nsfw'); await old.wait()
    await changeCollectionMode(page, 'sfw')
    await expect.poll(() => runtime.calls.filter(c => c.completed).length).toBe(2)
    old.release(); await old.done(); await settleRuntime(page)
    await expect(page.locator('.game-collections-page')).not.toContainText('Adult game')
    await expect(page.locator('.game-collections-page')).not.toContainText('Adult collection')
    runtime.state.modeFailure = true
    await changeCollectionMode(page, 'nsfw')
    await expect(page.getByRole('alert')).toContainText('仍为你保留')
    await expect(page.locator('.game-collection-card').first()).toBeVisible()
    runtime.state.modeFailure = false
    await page.getByRole('button', { name: '重试', exact: true }).click()
    await expect(page.locator('.game-collections-page')).toContainText(path.endsWith('collection-3') ? 'Adult game' : 'Adult collection')
    expect(runtime.calls.map(c => c.url.searchParams.get('mode'))).toEqual(['sfw', 'nsfw', 'sfw', 'nsfw', 'nsfw'])
    runtime.assertQuiet()
  })
  test(`Initial 503 keeps a real retry surface: ${path}`, async ({ page, runtime }) => {
    runtime.state.failure = true
    runtime.expectedURLs.add(new URL(path, runtime.app.base).href)
    const response = await page.goto(path)
    expect(response!.status()).toBe(503)
    await expect(page.getByRole('alert')).toContainText('暂时无法加载')
    await settleRuntime(page)
    expect(runtime.calls).toHaveLength(1)
    runtime.state.failure = false
    await page.getByRole('button', { name: '重试', exact: true }).click()
    await expect(page.locator('.game-collection-card').first()).toBeVisible()
    expect(runtime.calls).toHaveLength(2)
    runtime.assertQuiet()
  })
}

test('Index load more retains cards on failure, dedupes, and mode success resets page one', async ({ page, runtime }) => {
  await openRuntime(page, '/games/collections')
  runtime.state.pageFailure = true
  await page.getByRole('button', { name: '加载更多' }).click()
  await expect(page.locator('.game-collections-load')).toContainText('加载失败')
  await expect(page.locator('.game-collection-card')).toHaveCount(6)
  runtime.state.pageFailure = false
  await page.getByRole('button', { name: '重试', exact: true }).click()
  await expect(page.locator('.game-collection-card')).toHaveCount(8)
  await changeCollectionMode(page, 'nsfw')
  await expect(page.locator('.game-collection-card')).toHaveCount(6)
  await expect(page.getByRole('button', { name: '加载更多' })).toBeVisible()
  expect(runtime.calls.map(c => c.url.searchParams.get('page'))).toEqual(['1', '2', '2', '1'])
  runtime.assertQuiet()
})

test('Old load-more response cannot append to a new mode snapshot', async ({ page, runtime }) => {
  await openRuntime(page, '/games/collections')
  const more = runtime.hold(url => url.searchParams.get('page') === '2')
  await page.getByRole('button', { name: '加载更多' }).click(); await more.wait()
  await changeCollectionMode(page, 'nsfw')
  await expect(page.locator('.game-collections-page')).toContainText('Adult collection')
  more.release(); await more.done(); await settleRuntime(page)
  await expect(page.locator('.game-collection-card')).toHaveCount(6)
  runtime.assertQuiet()
})

test('Real Index to Detail and locale navigation use independent route identities', async ({ page, runtime }) => {
  await openRuntime(page, '/games/collections')
  await page.locator('.game-collection-card').first().click()
  await expect(page).toHaveURL(/\/games\/collections\/collection-3$/)
  await expect(page.locator('.game-collection-timeline')).toBeVisible()
  await settleRuntime(page)
  await page.getByRole('button', { name: 'EN', exact: true }).click()
  await expect(page).toHaveURL(/\/en\/games\/collections\/collection-3$/)
  await expect(page.locator('h1')).toHaveText('Forest stories 3')
  await expect(page.getByTestId('collection-now')).toContainText('October 6, 2026')
  await settleRuntime(page)
  expect(runtime.calls.map(c => [c.url.pathname, c.url.searchParams.get('lang'), c.url.searchParams.get('mode')])).toEqual([
    [collectionPath, 'zh', 'sfw'], [collectionPath + '/collection-3', 'zh', 'sfw'], [collectionPath + '/collection-3', 'en', 'sfw'],
  ])
  runtime.assertQuiet()
})

test('Empty visible timeline is 200 and unknown-only omits NOW', async ({ page, runtime }) => {
  await openRuntime(page, '/games/collections/adult-only')
  await expect(page.locator('.game-collections-page')).toContainText('当前暂无可展示作品')
  await expect(page.locator('.game-collections-page')).not.toContainText(/成人|隐藏/)
  await expect(page.getByTestId('collection-now')).toHaveCount(0)
  await openRuntime(page, '/games/collections/unknown-only')
  await expect(page.getByTestId('collection-now')).toHaveCount(0)
  await expect(page.locator('.game-collection-timeline li')).toHaveCount(1)
  runtime.assertQuiet()
})

test('Detail missing is authoritative HTTP 404, never an empty collection', async ({ request, runtime }) => {
  runtime.state.missing = true
  const response = await request.get(runtime.app.base + '/games/collections/missing')
  expect(response.status()).toBe(404)
  expect(await response.text()).not.toContain('当前暂无可展示作品')
  expect(runtime.calls).toHaveLength(1)
  runtime.assertQuiet()
})

test('An authoritative 404 during mode refresh enters the real Nuxt error page', async ({ page, runtime }) => {
  await openRuntime(page, '/games/collections/collection-3')
  runtime.state.missing = true
  await changeCollectionMode(page, 'nsfw')
  await expect(page.locator('.error-page__code')).toHaveText('404')
  await expect(page.locator('.game-collection-timeline')).toHaveCount(0)
  expect(runtime.calls).toHaveLength(2)
  runtime.assertQuiet()
})

for (const width of [390, 768, 1440]) {
  test(`Index grid and visible keyboard focus at ${width}px`, async ({ page, runtime }) => {
    await page.setViewportSize({ width, height: 900 })
    await openRuntime(page, '/games/collections')
    const columns = await page.locator('.game-collections-grid').evaluate(el => getComputedStyle(el).gridTemplateColumns.split(' ').length)
    expect(columns).toBe(width === 390 ? 1 : width === 768 ? 2 : 3)
    const card = page.locator('.game-collection-card').first()
    await page.keyboard.press('Tab'); await card.focus()
    await expect(card).toBeFocused()
    expect(await card.evaluate(el => getComputedStyle(el).outlineStyle)).not.toBe('none')
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true)
    runtime.assertQuiet()
  })
}

for (const width of [390, 899, 900, 1440]) {
  test(`Timeline geometry and overflow at ${width}px`, async ({ page, runtime }) => {
    await page.setViewportSize({ width, height: 900 })
    await openRuntime(page, '/games/collections/collection-3')
    const nodes = page.locator('[data-phase="past"] li')
    const boxes = await Promise.all([0, 1, 2, 3].map(i => nodes.nth(i).boundingBox()))
    if (width < 900) {
      expect(boxes[1]!.y).toBeGreaterThan(boxes[0]!.y)
      expect(boxes[1]!.x).toBe(boxes[0]!.x)
    } else {
      expect(boxes[1]!.y).toBe(boxes[0]!.y)
      expect(boxes[3]!.x).toBe(boxes[2]!.x)
      expect(boxes[3]!.y).toBeGreaterThan(boxes[2]!.y)
    }
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true)
    runtime.assertQuiet()
  })
}

test.describe('Date-only presentation in a western timezone', () => {
  test.use({ timezoneId: 'America/Los_Angeles' })
  test('keeps server and browser day labels identical', async ({ page, runtime }) => {
    await openRuntime(page, '/en/games/collections/collection-3')
    await expect(page.locator('[data-game-id="7"] .game-collection-time')).toHaveText('January 2, 2020')
    expect(runtime.count(collectionPath + '/collection-3')).toBe(1)
    runtime.assertQuiet()
  })
})
