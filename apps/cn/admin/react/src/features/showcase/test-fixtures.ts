import type { CampaignWorkspace, Candidate, Composition, DailyStat, Stats } from './types'
export const campaignFixture = (): CampaignWorkspace => ({
  id: '119', internal_name: 'October Campaign', content_type: 'game', sponsored: false, linked_game_id: '1',
  state: 'draft', derived_status: 'draft', starts_at: '2026-10-03T10:00:00Z', ends_at: '2026-10-03T13:00:00Z',
  weight: 100, pin_position: null, desktop_object_key: `game/showcase/119/desktop/${'a'.repeat(32)}.avif`, mobile_object_key: null,
  focal_x: 0.5, focal_y: 0.5, primary_action_type: 'game', primary_target: null, secondary_action_type: 'steam', secondary_target: 'https://store.steampowered.com/app/1/',
  created_at: '2026-10-03T09:00:00Z', updated_at: '2026-10-03T09:00:00Z', ready_to_publish: true, publication_diagnostics: [],
  locales: [{ lang: 'zh', enabled: true, title: '中文活动', summary: '摘要', tags: ['Furry'], editorial_note: '编辑推荐' }, { lang: 'en', enabled: false, title: 'English campaign', summary: 'Summary', tags: ['Game'], editorial_note: 'Editorial note' }],
})
export const compositionFixture = (count = 1): Composition => ({ schema_version: 1, snapshot_id: 'snapshot', generated_at: '2026-10-03T10:00:00Z', valid_until: '2026-10-03T10:05:00Z', items: Array.from({ length: count }, (_, i) => ({
  key: `campaign:${119 + i}`, source: i === 1 ? 'automatic' : 'managed', reason: i === 1 ? 'trending' : 'editorial', content_type: 'game', sponsored: false,
  campaign_id: i === 1 ? undefined : String(119 + i), game_id: '1', title: `Showcase ${i + 1}`, summary: 'Summary', tags: [],
  artwork: i === 1 ? { kind: 'steam', url: 'https://shared.akamai.steamstatic.com/test.jpg' } : { kind: 'managed', desktop_object_key: campaignFixture().desktop_object_key! },
  primary_action: { type: 'game', game_id: '1' }, position: i + 1, tracking_token: 'test-only-token',
})) })
export const candidateFixture = (): Candidate => ({ game_id: '1', name: 'Candidate game', status: 'blocked', first_available: null, pool_failures: [], showcase_eligible: false, sfw: true, locale_ready: true, artwork_ready: true, release: null, candidate: false, excluded_reasons: ['not_approved', 'future_reason'], trending: { recent: { observed_days: 3, average: 700, coverage: 1 }, baseline: { observed_days: 7, average: 100, coverage: null }, momentum: 1800, eligible: true, weight: 1800 } })
export const dailyFixture: DailyStat = { stat_date: '2026-10-03', valid_impressions: 100, qualified_clicks: 20, session_estimate: 80, click_artwork: 5, click_title: 5, click_primary: 8, click_secondary: 2, impression_position_1: 100, impression_position_2: 0, impression_position_3: 0, impression_position_4: 0 }
export const statsFixture = (): Stats => ({ totals: { ...dailyFixture, ctr: 0.125, session_estimate_method: 'sum_of_daily_hll_estimates' }, daily: [dailyFixture], timezone: 'Asia/Shanghai' })
