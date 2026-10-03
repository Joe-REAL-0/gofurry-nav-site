import { afterEach, describe, expect, it, vi } from 'vitest'
import { resetCsrf } from '../../lib/api'
import { buildShowcaseContentPayload, showcaseAPI, statsCSV } from './api'
import { artworkFileError } from './showcase-assets'
import { contentPatch, contentSchema, contentValues } from './showcase-content'
import { scheduleSchema } from './showcase-schedule'
import { chinaDate, chinaWallTimeToRFC3339, rfc3339ToChinaWallTime, statsRange, validDateRange } from './showcase-time'
import { campaignFixture } from './test-fixtures'

afterEach(() => { vi.unstubAllGlobals(); vi.unstubAllEnvs(); resetCsrf() })
describe('Shanghai wall time is independent of the operator timezone', () => {
  it.each(['UTC', 'America/New_York', 'Pacific/Honolulu', 'Asia/Tokyo'])('%s round trips without DST conversion', timezone => {
    vi.stubEnv('TZ', timezone)
    expect(chinaWallTimeToRFC3339('2026-10-03T18:30')).toBe('2026-10-03T18:30:00+08:00')
    expect(rfc3339ToChinaWallTime('2026-10-03T10:30:00Z')).toBe('2026-10-03T18:30')
    expect(rfc3339ToChinaWallTime(chinaWallTimeToRFC3339('2026-03-08T02:30'))).toBe('2026-03-08T02:30')
    expect(chinaDate(new Date('2026-10-03T17:00:00Z'))).toBe('2026-10-04')
  })
  it('rejects invalid/unzoned dates and bounds inclusive ranges to 366 days', () => {
    expect(() => chinaWallTimeToRFC3339('2026-02-30T18:30')).toThrow()
    expect(rfc3339ToChinaWallTime('2026-10-03T10:30')).toBe('')
    expect(validDateRange('2025-10-03', '2026-10-03')).toBe(true)
    expect(validDateRange('2025-10-02', '2026-10-03')).toBe(false)
    expect(validDateRange('2026-02-30', '2026-03-03')).toBe(false)
    expect(statsRange('all', '2020-01-01T00:00:00Z', '2026-10-03')).toEqual({ from: '2025-10-03', to: '2026-10-03' })
    expect(statsRange('all', null, '2026-10-03')).toEqual(statsRange('30', null, '2026-10-03'))
  })
})
describe('full Stage A content ownership', () => {
  it('focal patches preserve every content field and both locale states', () => {
    const w = campaignFixture(); const body = buildShowcaseContentPayload(w, { focal_x: 0, focal_y: 1 })
    expect(body).toEqual({ internal_name: w.internal_name, content_type: w.content_type, sponsored: false, linked_game_id: 1, locales: w.locales, focal_x: 0, focal_y: 1, primary_action_type: 'game', primary_target: null, secondary_action_type: 'steam', secondary_target: w.secondary_target })
    expect(w.focal_x).toBe(0.5)
  })
  it('Sponsored clears both Editorial Notes and internal game actions discard external targets', () => {
    const w = campaignFixture(); const v = contentValues(w); v.sponsored = true; v.primary_target = 'https://example.test/stale'
    const body = buildShowcaseContentPayload(w, contentPatch(v))
    expect(body.locales.every(locale => locale.editorial_note === null)).toBe(true)
    expect(body.primary_target).toBeNull()
  })
  it.each([['title', 160], ['summary', 400], ['editorial_note', 200]] as const)('counts %s by Unicode codepoints', (field, max) => {
    const values = contentValues(campaignFixture()); values.locales[0][field] = '🐺'.repeat(max)
    expect(contentSchema.safeParse(values).success).toBe(true)
    values.locales[0][field] += '🐺'; expect(contentSchema.safeParse(values).success).toBe(false)
  })
  it('rejects fourth tags, unknown actions, missing internal relations and unsafe targets', () => {
    const v = contentValues(campaignFixture())
    expect(contentSchema.safeParse({ ...v, locales: [{ ...v.locales[0], tags: ['1', '2', '3', '4'] }, v.locales[1]] }).success).toBe(false)
    for (const patch of [{ primary_action_type: 'free-label' }, { linked_game_id: '' }, { primary_action_type: 'website', primary_target: 'http://example.test' }, { secondary_target: 'https://user:pass@example.test' }]) expect(contentSchema.safeParse({ ...v, ...patch }).success).toBe(false)
    expect(scheduleSchema.safeParse({ starts_at: '2026-10-03T18:00', ends_at: '2026-10-03T17:00', weight: 10001, pin_position: '5' }).success).toBe(false)
  })
})
it('rejects oversized and non-AVIF UX inputs without processing bytes', () => {
  expect(artworkFileError(new File(['original'], 'image.avif', { type: 'image/avif' }))).toBe('')
  expect(artworkFileError(new File(['png'], 'image.png', { type: 'image/png' }))).toBeTruthy()
  expect(artworkFileError(new File([new Uint8Array(5 * 1024 * 1024 + 1)], 'large.avif'))).toBeTruthy()
})
it.each(['desktop', 'mobile'] as const)('uploads %s through same-origin session/CSRF multipart transport', async variant => {
  const fetchMock = vi.fn().mockResolvedValueOnce(new Response(JSON.stringify({ code: 1, data: { token: 'csrf', header_name: 'X-CSRF' } }))).mockResolvedValueOnce(new Response(JSON.stringify({ code: 1, data: { primary: 'ready', mirror: 'ready', warnings: [] } })))
  vi.stubGlobal('fetch', fetchMock)
  const file = new File(['original'], 'art.avif', { type: 'image/avif' }); await showcaseAPI.upload('119', variant, file)
  const [path, options] = fetchMock.mock.calls[1]
  expect(path).toBe(`/api/v1/game/showcase/campaigns/119/artwork/${variant}`)
  expect(options).toEqual(expect.objectContaining({ method: 'POST', credentials: 'include', headers: { 'X-CSRF': 'csrf' } }))
  expect((options.body as FormData).get('file')).toBe(file)
  expect(statsCSV('119', '2026-10-01', '2026-10-03')).toBe('/api/v1/game/showcase/campaigns/119/stats.csv?from=2026-10-01&to=2026-10-03')
})
