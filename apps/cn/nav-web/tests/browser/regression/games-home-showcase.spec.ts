import { test, expect, type GamesHomeScene } from '../fixtures/games-home'
import { showcaseOrigins } from '../fixtures/games-home-showcase-data'
import { mkdir } from 'node:fs/promises'
import { join } from 'node:path'

export async function assertGameHomeShowcaseAppearance(scene: GamesHomeScene) {
  const hero = scene.showcase
  await expect(hero).toHaveCSS('backdrop-filter', 'none')
  await expect(hero).toHaveCSS('border-radius', '16px')
  await expect(hero).toHaveCSS('transform', 'none')
  await expect(hero).toHaveCSS('animation-name', 'none')
  await expect(hero.locator('.game-home-showcase__title')).toHaveCSS('font-weight', '600')
  await expect(hero.locator('.game-home-showcase__title')).toHaveCSS('-webkit-line-clamp', '2')
  await expect(hero.locator('.game-home-showcase__context')).toHaveCSS('font-weight', '600')
  const image = hero.locator('img:visible').first()
  await expect(image).toHaveCSS('object-fit', 'cover')
  const imageBox = await image.boundingBox(), frameBox = await hero.locator('.game-home-showcase__artwork').boundingBox()
  expect(Math.abs(imageBox!.width - frameBox!.width)).toBeLessThan(1)
  expect(Math.abs(imageBox!.height - frameBox!.height)).toBeLessThan(1)
  const box = await hero.boundingBox()
  await image.hover()
  await expect(hero).toHaveCSS('transform', 'none')
  await expect(image).toHaveCSS('transform', 'none')
  expect(await hero.boundingBox()).toEqual(box)
  expect(await hero.locator('.game-home-showcase__context').evaluate(element => getComputedStyle(element).color))
    .not.toBe(await hero.evaluate(element => {
      const probe = document.createElement('span'); probe.style.color = 'var(--gf-danger)'
      element.append(probe); const color = getComputedStyle(probe).color; probe.remove(); return color
    }))
  const hierarchy = await hero.evaluate(element => ['context', 'summary', 'title'].map(part =>
    Number.parseFloat(getComputedStyle(element.querySelector(`.game-home-showcase__${part}`)!).fontSize)))
  expect(hierarchy[0]).toBeLessThan(hierarchy[1]!); expect(hierarchy[1]).toBeLessThan(hierarchy[2]!)
}

test('Empty and failed optional slices keep the original Home content first', async ({ gamesHome }) => {
  const scene = await gamesHome.open()
  await expect(scene.showcase).toHaveCount(0)
  await expect(scene.root.locator('.game-info-shell > :first-child')).toHaveClass('game-info-group')
  scene.assertQuiet()
})

test('Showcase 503 is optional and leaves catalog, statistics, News and sidebar alive', async ({ gamesHome }) => {
  const scene = await gamesHome.open({ showcase: 'showcase-failure' })
  await expect(scene.showcase).toHaveCount(0)
  await expect(scene.groups).toHaveCount(4)
  await expect(scene.stats).toBeVisible()
  await expect(scene.sidebar).toBeVisible()
  await expect(scene.news).toHaveCount(1)
  scene.assertQuiet()
})

for (const locale of ['zh', 'en'] as const) {
  test(`Single editorial SSR and hydrated first item, strict locale and no controls (${locale})`, async ({ gamesHome }) => {
    const scene = await gamesHome.open({ showcase: 'single-editorial', locale })
    const first = scene.snapshot.items[0]!
    expect(scene.rendered).toContain(first.title)
    expect(scene.rendered).toContain(first.summary)
    expect(scene.rendered).not.toContain(locale === 'zh' ? 'Echoes of the Wild' : '荒野回声')
    await expect(scene.showcase.locator('h2')).toHaveText(first.title)
    await expect(scene.showcase.locator('.game-home-showcase__context')).toHaveText(locale === 'zh' ? 'GoFurry 精选' : 'GoFurry Pick')
    await expect(scene.showcase.locator('.game-home-showcase__controls')).toHaveCount(0)
    await expect(scene.showcase.locator('[aria-live]')).toHaveText('')
    await expect(scene.showcase).not.toHaveAttribute('tabindex')
    await expect(scene.showcase).not.toHaveAttribute('role', 'application')
    await expect(scene.showcase.locator('.gf-button')).toHaveAttribute('href', `${locale === 'en' ? '/en' : ''}/games/6101`)
    await expect(scene.showcase.locator('.gf-button')).not.toHaveAttribute('target')
    await assertGameHomeShowcaseAppearance(scene)
    await reviewScreenshot(scene, `editorial-${locale}-desktop`)
    expect(await scene.showcase.boundingBox()).not.toBeNull()
    const hero = await scene.showcase.boundingBox(), catalog = await scene.group(0).boundingBox()
    expect(catalog!.y).toBeGreaterThan(hero!.y + hero!.height)
    scene.assertQuiet()
  })
}

test('Four items preserve supplied order, non-circular boundaries, focus and user-only announcements', async ({ gamesHome }) => {
  const scene = await gamesHome.open({ showcase: 'four-items' })
  const { showcase: hero, page, snapshot } = scene
  const next = hero.getByRole('button', { name: '下一项精选' }), prev = hero.getByRole('button', { name: '上一项精选' })
  await expect(hero.locator('.game-home-showcase__count')).toHaveText('01 / 04')
  await expect(prev).toBeDisabled(); await expect(next).toBeEnabled()
  await expect(hero.locator('[aria-live]')).toHaveText('')
  await next.focus(); await page.keyboard.press('ArrowRight')
  await expect(hero.locator('h2')).toHaveText(snapshot.items[0]!.title)
  for (let i = 1; i < 4; i++) {
    await next.click()
    await expect(hero.locator('h2')).toHaveText(snapshot.items[i]!.title)
    await expect(hero.locator('[aria-live]')).toHaveText(`${i + 1} / 4 · ${snapshot.items[i]!.title}`)
    await expect(hero.locator('.game-home-showcase__count')).toHaveText(`0${i + 1} / 04`)
    if (i < 3) await expect(next).toBeFocused()
  }
  await expect(next).toBeDisabled(); await expect(prev).toBeEnabled()
  await prev.click(); await prev.click(); await prev.click()
  await expect(prev).toBeDisabled()
  await expect(hero.locator('h2')).toHaveText(snapshot.items[0]!.title)
  await expect(hero.locator('[aria-live]')).toHaveText(`1 / 4 · ${snapshot.items[0]!.title}`)
  scene.assertQuiet()
})

for (const width of [1440, 1024, 768, 390]) {
  test(`Showcase responsive composition and bounded long copy (${width})`, async ({ gamesHome }) => {
    const scene = await gamesHome.open({ showcase: 'long-content', width })
    const hero = scene.showcase, media = hero.locator('.game-home-showcase__artwork'), content = hero.locator('.game-home-showcase__content')
    const m = await media.boundingBox(), c = await content.boundingBox(), root = await hero.boundingBox()
    if (width >= 1024) {
      expect(Math.abs(m!.y - c!.y)).toBeLessThan(2)
      expect(m!.width / (m!.width + c!.width)).toBeCloseTo(.64, 2)
      expect(root!.height).toBeLessThan(350)
    } else {
      expect(c!.y).toBeGreaterThanOrEqual(m!.y + m!.height - 1)
      expect(m!.width / m!.height).toBeCloseTo(width < 640 ? 16 / 9 : 2, 1)
      if (width === 390) expect(root!.height).toBeLessThan(450)
    }
    await expect(hero.locator('.game-home-showcase__title')).toHaveCSS('-webkit-line-clamp', '2')
    await expect(hero.locator('.game-home-showcase__summary')).toHaveCSS('-webkit-line-clamp', width < 640 ? '2' : '3')
    await expect(hero.locator('.game-home-showcase__tags li:visible')).toHaveCount(width < 640 ? 2 : 3)
    if (width < 640) await expect(hero.locator('.game-home-showcase__note')).toBeHidden()
    else {
      const note = hero.locator('.game-home-showcase__note')
      await expect(note).toBeVisible()
      const n = await note.boundingBox(), action = await hero.locator('.gf-button').boundingBox()
      expect(n!.y + n!.height).toBeLessThanOrEqual(action!.y)
    }
    expect(await pageOverflow(scene)).toBe(false)
    await expect(hero.locator('.gf-button')).toBeInViewport()
    await reviewScreenshot(scene, `long-content-${width}`)
    scene.assertQuiet()
  })
}

test('Upcoming uses Steam routing and canonical calendar release metadata', async ({ gamesHome }) => {
  const scene = await gamesHome.open({ showcase: 'automatic-upcoming', locale: 'en' })
  await expect(scene.showcase.locator('.game-home-showcase__context')).toHaveText('Upcoming')
  await expect(scene.showcase.locator('.game-home-showcase__meta')).toHaveText('Expected Q2 2027')
  await expect(scene.showcase.locator('img')).toHaveAttribute('src', /shared\.akamai\.steamstatic\.com\/store_item_assets\/steam\/apps\/6101\/showcase\/header\.jpg\?v=2/)
  await expect(scene.showcase.locator('.game-home-showcase__note')).toHaveCount(0)
  scene.assertQuiet()
})

for (const scenario of ['managed-sponsored-tabletop', 'managed-sponsored-merchandise'] as const) {
  test(`Sponsored context and fixed external CTA (${scenario})`, async ({ gamesHome }) => {
    const scene = await gamesHome.open({ showcase: scenario, theme: 'dark', width: scenario.endsWith('merchandise') ? 390 : 1440, locale: 'en' })
    await expect(scene.showcase.locator('.game-home-showcase__context')).toHaveText(scenario.endsWith('merchandise') ? 'Merchandise · Sponsored' : 'Tabletop · Sponsored')
    await expect(scene.showcase.locator('.game-home-showcase__note')).toHaveCount(0)
    await expect(scene.showcase.locator('.gf-button')).toHaveText(scenario.endsWith('merchandise') ? 'View Product' : 'View Project')
    for (const link of await scene.showcase.getByRole('link').all()) {
      await expect(link).toHaveAttribute('target', '_blank')
      await expect(link).toHaveAttribute('rel', 'noopener noreferrer')
    }
    await assertGameHomeShowcaseAppearance(scene)
    await reviewScreenshot(scene, scenario)
    scene.assertQuiet()
  })
}

test('Managed Primary failure tries Mirror with focal coordinates; missing mobile key uses desktop', async ({ gamesHome }) => {
  const scene = await gamesHome.open({ showcase: 'primary-failure', width: 390 })
  const media = scene.showcase.locator('img:visible')
  await expect(media).toHaveAttribute('src', `${showcaseOrigins.mirror}/${scene.snapshot.items[0]!.artwork.mobile_object_key}`)
  await expect(media).toHaveCSS('object-position', '25% 75%')
  expect(scene.assets.some(url => url.startsWith(showcaseOrigins.primary))).toBe(true)
  expect(scene.assets.some(url => url.startsWith(showcaseOrigins.mirror))).toBe(true)
  scene.assertQuiet()
})

test('Absent mobile artwork reuses desktop through ManagedAssetImage', async ({ gamesHome }) => {
  const scene = await gamesHome.open({ showcase: 'mobile-missing', width: 390 })
  await expect(scene.showcase.locator('img')).toHaveAttribute('src', `${showcaseOrigins.primary}/${scene.snapshot.items[0]!.artwork.desktop_object_key}`)
  scene.assertQuiet()
})

for (const scenario of ['media-failure', 'steam-failure'] as const) {
  test(`Exhausted media never removes a slide, CTA or geometry (${scenario})`, async ({ gamesHome }) => {
    const scene = await gamesHome.open({ showcase: scenario })
    await expect(scene.showcase.locator('img')).toHaveCount(0)
    await expect(scene.showcase.locator('h2')).toHaveText(scene.snapshot.items[0]!.title)
    await expect(scene.showcase.locator('.gf-button')).toBeVisible()
    const before = await scene.showcase.boundingBox()
    expect(before!.height).toBeGreaterThan(300)
    if (scenario === 'media-failure') {
      await expect(scene.showcase.locator('.game-home-showcase__count')).toHaveText('01 / 04')
      await scene.showcase.getByRole('button', { name: '下一项精选' }).click()
      await expect(scene.showcase.locator('img')).toHaveCount(0)
      await expect(scene.showcase.locator('.game-home-showcase__count')).toHaveText('02 / 04')
      expect(await scene.showcase.boundingBox()).toEqual(before)
    }
    scene.assertQuiet()
  })
}

test('Real intersection sends one impression per snapshot/item and no duplicate on revisit', async ({ gamesHome }) => {
  const scene = await gamesHome.open({ showcase: 'four-items', height: 200 })
  expect(scene.events).toEqual([])
  await scene.page.setViewportSize({ width: 1440, height: 900 })
  await scene.showcase.scrollIntoViewIfNeeded()
  await expect.poll(() => scene.events.filter(event => event.event === 'impression').length).toBe(1)
  await scene.showcase.getByRole('button', { name: '下一项精选' }).click()
  await expect.poll(() => scene.events.filter(event => event.event === 'impression').length).toBe(2)
  await scene.showcase.getByRole('button', { name: '上一项精选' }).click()
  await scene.showcase.getByRole('button', { name: '下一项精选' }).click()
  await scene.showcase.getByRole('button', { name: '下一项精选' }).click()
  await expect.poll(() => scene.events.filter(event => event.event === 'impression').length).toBe(3)
  expect(scene.events.filter(event => event.tracking_token === scene.snapshot.items[0]!.tracking_token)).toHaveLength(1)
  expect(new Set(scene.events.map(event => event.session_id)).size).toBe(1)
  scene.assertQuiet()
})

for (const source of ['artwork', 'title', 'primary'] as const) {
  test(`Click ${source} precedes impression and navigates to localized game despite analytics 500`, async ({ gamesHome }) => {
    const scene = await gamesHome.open({ showcase: 'tracking-failure', locale: 'en', height: 200 })
    const selector = { artwork: '.game-home-showcase__artwork', title: 'h2 a', primary: '.gf-button' }[source]
    scene.expectGameDestination('6101')
    await scene.showcase.locator(selector).click()
    await expect(scene.page).toHaveURL(/\/en\/games\/6101$/)
    await expect(scene.page.getByRole('heading', { name: 'Showcase destination', exact: true })).toBeVisible()
    await expect.poll(() => scene.events.some(event => event.event === 'click' && event.source === source)).toBe(true)
    expect(scene.events[0]?.event).toBe('click')
    await expect.poll(() => scene.page.locator('.game-detail-page').count()).toBe(1)
    scene.assertQuiet()
  })
}

test('Secondary external popup still opens when analytics returns 500', async ({ gamesHome }) => {
  const scene = await gamesHome.open({ showcase: 'tracking-failure' })
  const url = scene.snapshot.items[0]!.secondary_action!.target!
  scene.expectShowcasePopup(url)
  const popup = scene.page.waitForEvent('popup')
  await scene.showcase.locator('.game-home-showcase__secondary').click()
  const destination = await popup
  await expect(destination).toHaveURL(url)
  await destination.waitForLoadState('load')
  await expect.poll(() => scene.events.some(event => event.event === 'click' && event.source === 'secondary')).toBe(true)
  await expect(scene.showcase).toBeVisible()
  await expect(scene.page.getByRole('alert')).toHaveCount(0)
  await destination.close()
  scene.assertQuiet()
})

async function pageOverflow(scene: GamesHomeScene) {
  return scene.page.evaluate(() => document.documentElement.scrollWidth > innerWidth)
}

// Optional review artifacts only, never accepted Visual snapshots or a new runner.
async function reviewScreenshot(scene: GamesHomeScene, name: string) {
  const directory = process.env.GOFURRY_SHOWCASE_REVIEW_DIR
  if (!directory) return
  await mkdir(directory, { recursive: true })
  await scene.page.mouse.move(1, 1)
  await scene.settle(scene.showcase)
  await scene.page.screenshot({ path: join(directory, `${name}.png`) })
}
