import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { act, cleanup, fireEvent, render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { createMemoryRouter, RouterProvider } from 'react-router-dom'
import { afterEach, beforeEach, expect, it, vi } from 'vitest'
import { ToastProvider } from '../../app/toast'
import { getJSON, listJSON, sendForm, sendJSON } from '../../lib/api'
import { AuditPage } from '../system/audit-page'
import { ShowcaseCampaignPage } from './showcase-campaign-page'
import { ShowcasePage } from './showcase-page'
import { ShowcaseTrend } from './showcase-stats'
import { campaignFixture, candidateFixture, compositionFixture, dailyFixture, statsFixture } from './test-fixtures'
import type { CampaignWorkspace } from './types'

const mocks = vi.hoisted(() => ({ capabilities: ['content.read', 'content.write', 'audit.read'], chart: { setOption: vi.fn(), resize: vi.fn(), dispose: vi.fn() } }))
vi.mock('../auth/auth-context', () => ({ useAuth: () => ({ can: (capability: string) => mocks.capabilities.includes(capability) }) }))
vi.mock('../../lib/api', async original => ({ ...await original<typeof import('../../lib/api')>(), getJSON: vi.fn(), listJSON: vi.fn(), sendJSON: vi.fn(), sendForm: vi.fn() }))
vi.mock('echarts/core', () => ({ use: vi.fn(), init: () => mocks.chart }))
let workspace: CampaignWorkspace
beforeEach(() => {
  workspace = campaignFixture(); mocks.capabilities = ['content.read', 'content.write', 'audit.read']
  vi.mocked(getJSON).mockImplementation(async path => {
    if (path.includes('/composition')) return compositionFixture()
    if (path.includes('/candidates')) return { total: 1, items: [candidateFixture()] }
    if (path.includes('/analytics/quality')) return { daily: [], timezone: 'Asia/Shanghai' }
    if (path.includes('/stats?')) return statsFixture()
    if (path.includes('/campaigns?')) return { list: [workspace], total: 41 }
    if (path.includes('/audit/logs?')) return { list: [], total: 0 }
    return workspace
  })
  vi.mocked(listJSON).mockResolvedValue({ list: [{ id: '1', label: 'Linked Game' }], total: 1 })
  vi.mocked(sendJSON).mockImplementation(async (_path, _method, payload) => {
    if (payload) { const value = payload as Partial<CampaignWorkspace>; workspace = { ...workspace, ...value, linked_game_id: value.linked_game_id == null ? workspace.linked_game_id : String(value.linked_game_id) } }
    return workspace
  })
  vi.mocked(sendForm).mockResolvedValue({ object_key: 'key', primary: 'ready', mirror: 'ready', warnings: [] })
  vi.stubGlobal('ResizeObserver', class { observe() {} disconnect() {} })
  vi.stubGlobal('URL', URL)
  URL.createObjectURL = vi.fn(file => `blob:${(file as File).name}`); URL.revokeObjectURL = vi.fn()
})
afterEach(() => { cleanup(); vi.restoreAllMocks(); vi.resetAllMocks(); vi.unstubAllGlobals() })
function setup(path = '/game/showcase') {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false, gcTime: 0 }, mutations: { retry: false } } })
  const router = createMemoryRouter([
    { path: '/game/showcase', element: <ShowcasePage /> }, { path: '/game/showcase/:id', element: <ShowcaseCampaignPage /> },
    { path: '/system/audit', element: <AuditPage /> }, { path: '/game/games/:id', element: <p>Game Workspace</p> },
  ], { initialEntries: [path] })
  render(<QueryClientProvider client={client}><ToastProvider><RouterProvider router={router} /></ToastProvider></QueryClientProvider>)
  return { router, client }
}
async function choose(label: string, option: string) {
  const user = userEvent.setup()
  await user.click(screen.getByRole('combobox', { name: label }))
  await user.click(await screen.findByRole('option', { name: option }))
}
it.each([0, 1, 4])('displays %i composition items in API order without resolving managed CDN keys', async count => {
  vi.mocked(getJSON).mockResolvedValue(compositionFixture(count)); setup()
  if (!count) expect(await screen.findByText('当前没有可展示的 Showcase')).toBeInTheDocument()
  else {
    await screen.findByText('Showcase 1')
    expect(document.querySelectorAll('[data-showcase-item]')).toHaveLength(count)
    expect(screen.getAllByText('Managed Artwork · 已配置').length).toBe(count === 4 ? 3 : 1)
    const images = screen.queryAllByRole('img')
    expect(images).toHaveLength(count === 4 ? 1 : 0)
    if (count === 4) expect(images[0]).toHaveAttribute('src', compositionFixture(4).items[1].artwork.url)
    expect(document.querySelector('img[src*="game/showcase"]')).toBeNull()
  }
})
it('keeps tabs, locale, pool and pagination in the URL and discovery never writes classification', async () => {
  const { router } = setup(); await screen.findByText('Showcase 1')
  await choose('展示语言', 'English'); await waitFor(() => expect(router.state.location.search).toContain('locale=en'))
  expect(getJSON).toHaveBeenCalledWith('/api/v1/game/showcase/composition?lang=en&region=CN')
  fireEvent.click(screen.getByRole('tab', { name: '自动发现' })); await screen.findByText('Candidate game')
  await choose('自动发现 Pool', '热度上升')
  expect(router.state.location.search).toContain('pool=trending')
  expect(screen.getByText('未允许自动 Showcase')).toBeInTheDocument(); expect(screen.getByText('future_reason')).toBeInTheDocument()
  expect(screen.getByText(/Momentum 1,800/)).toBeInTheDocument()
  expect(screen.getByRole('link', { name: /打开游戏/ })).toHaveAttribute('href', '/game/games/1?tab=classification')
  expect(sendJSON).not.toHaveBeenCalled()
  fireEvent.click(screen.getByRole('tab', { name: '活动' })); await screen.findByRole('link', { name: 'October Campaign' })
  fireEvent.click(screen.getByRole('button', { name: '下一页' }))
  expect(router.state.location.search).toContain('page_num=2')
  expect(vi.mocked(getJSON).mock.calls.some(([path]) => path.includes('/stats'))).toBe(false)
})
it('submits bounded campaign filters from URL and creates a real-ID Draft workspace', async () => {
  const confirm = vi.spyOn(window, 'confirm').mockReturnValue(false)
  const { router } = setup('/game/showcase?tab=campaigns&page_num=2&page_size=20&state=draft&derived_status=draft&sponsored=false&keyword=October')
  await screen.findByRole('link', { name: 'October Campaign' })
  expect(getJSON).toHaveBeenCalledWith('/api/v1/game/showcase/campaigns?page_num=2&page_size=20&keyword=October&state=draft&derived_status=draft&sponsored=false')
  fireEvent.change(screen.getByRole('textbox', { name: '搜索列表' }), { target: { value: 'Winter' } })
  expect(router.state.location.search).toContain('keyword=Winter'); expect(router.state.location.search).toContain('page_num=1')
  await choose('存储状态', '已发布'); await choose('展示状态', '展示中'); await choose('商业属性', '商业推广')
  expect(router.state.location.search).toContain('state=published'); expect(router.state.location.search).toContain('derived_status=active'); expect(router.state.location.search).toContain('sponsored=true')
  fireEvent.click(screen.getByRole('button', { name: '新建活动' }))
  fireEvent.change(screen.getByRole('textbox', { name: /内部名称/ }), { target: { value: 'New Draft' } })
  fireEvent.click(screen.getByRole('button', { name: '创建草稿并进入' }))
  await waitFor(() => expect(router.state.location.pathname).toBe('/game/showcase/119'))
  expect(sendJSON).toHaveBeenCalledWith('/api/v1/game/showcase/campaigns', 'POST', { internal_name: 'New Draft', content_type: 'game', sponsored: false, linked_game_id: null })
  expect(sendForm).not.toHaveBeenCalled()
  expect(confirm).not.toHaveBeenCalled()
})
it('guards dirty locale edits, clears Sponsored notes, and resets after a full content save', async () => {
  const confirm = vi.spyOn(window, 'confirm').mockReturnValue(false)
  const { router } = setup('/game/showcase/119?tab=content')
  await screen.findByRole('textbox', { name: '中文标题' })
  expect(screen.getByRole('checkbox', { name: '启用 中文' })).toBeChecked(); expect(screen.getByRole('checkbox', { name: '启用 English' })).not.toBeChecked()
  fireEvent.click(screen.getByRole('checkbox', { name: '启用 English' }))
  fireEvent.click(screen.getByRole('checkbox', { name: '商业推广' }))
  expect(screen.queryByRole('textbox', { name: '中文编辑推荐语' })).not.toBeInTheDocument()
  fireEvent.click(screen.getByRole('tab', { name: '素材' }))
  await waitFor(() => expect(confirm).toHaveBeenCalled()); expect(router.state.location.search).toBe('?tab=content')
  expect(screen.getByRole('button', { name: '发布' })).toBeDisabled()
  fireEvent.click(screen.getByRole('button', { name: '保存内容' }))
  await waitFor(() => expect(sendJSON).toHaveBeenCalled())
  const body = vi.mocked(sendJSON).mock.calls[0][2] as { locales: Array<{ editorial_note: string | null; enabled: boolean }> }
  expect(body.locales.every(locale => locale.editorial_note === null && locale.enabled)).toBe(true)
  await waitFor(() => expect(screen.getByRole('button', { name: '保存内容' })).toBeDisabled())
  confirm.mockClear(); fireEvent.click(screen.getByRole('tab', { name: '素材' }))
  await screen.findByText('素材焦点'); expect(confirm).not.toHaveBeenCalled()
})
it('shows HTTPS validation and preserves backend save errors', async () => {
  setup('/game/showcase/119?tab=content'); await screen.findByRole('textbox', { name: '中文标题' })
  await choose('Primary 动作', '访问官网')
  fireEvent.change(screen.getByRole('textbox', { name: 'Primary HTTPS 地址' }), { target: { value: 'http://example.test' } })
  fireEvent.click(screen.getByRole('button', { name: '保存内容' }))
  expect(await screen.findByText('请输入不含账号密码的 HTTPS 地址')).toBeInTheDocument(); expect(sendJSON).not.toHaveBeenCalled()
  vi.mocked(sendJSON).mockRejectedValueOnce(new Error('overlapping pin'))
  fireEvent.change(screen.getByRole('textbox', { name: /Primary HTTPS 地址/ }), { target: { value: 'https://example.test' } })
  fireEvent.click(screen.getByRole('button', { name: '保存内容' }))
  expect(await screen.findByText('overlapping pin')).toBeInTheDocument()
  expect(screen.getByRole('textbox', { name: /Primary HTTPS 地址/ })).toHaveValue('https://example.test')
})
it('preserves dirty content on a failed refresh, blocks unloading, then allows confirmed navigation', async () => {
  const confirm = vi.spyOn(window, 'confirm').mockReturnValue(true)
  const { router, client } = setup('/game/showcase/119?tab=content')
  const title = await screen.findByRole('textbox', { name: '中文标题' })
  fireEvent.change(title, { target: { value: 'Unsaved title' } })
  const event = new Event('beforeunload', { cancelable: true })
  fireEvent(window, event); expect(event.defaultPrevented).toBe(true)
  vi.mocked(getJSON).mockRejectedValueOnce(new Error('Workspace refresh failed'))
  await act(() => client.invalidateQueries({ queryKey: ['showcase', 'campaign', '119'] }))
  expect(await screen.findByText('Workspace refresh failed')).toBeInTheDocument()
  expect(title).toHaveValue('Unsaved title')
  fireEvent.click(screen.getByRole('tab', { name: '概览' }))
  await waitFor(() => expect(router.state.location.search).toBe(''))
  expect(confirm).toHaveBeenCalled()
  const clean = new Event('beforeunload', { cancelable: true }); fireEvent(window, clean)
  expect(clean.defaultPrevented).toBe(false)
})
it.each(['Desktop', 'Mobile'] as const)('uploads original %s AVIF with preview cleanup and nonfatal mirror warnings', async label => {
  vi.mocked(sendForm).mockResolvedValueOnce({ primary: 'ready', mirror: 'failed', warnings: ['R2 mirror sync failed'] })
  setup('/game/showcase/119?tab=assets'); await screen.findByText('素材焦点')
  const file = new File(['raw avif bytes'], `${label}.avif`, { type: 'image/avif' })
  fireEvent.change(screen.getByLabelText(`选择 ${label} AVIF`), { target: { files: [file] } })
  expect(await screen.findByRole('img', { name: `${label} 本地预览` })).toHaveAttribute('src', `blob:${label}.avif`)
  fireEvent.click(screen.getByRole('button', { name: `上传 ${label}` }))
  await waitFor(() => expect(sendForm).toHaveBeenCalled())
  const [path, body] = vi.mocked(sendForm).mock.calls[0]
  expect(path).toBe(`/api/v1/game/showcase/campaigns/119/artwork/${label.toLowerCase()}`); expect(body.get('file')).toBe(file)
  expect(await screen.findByText(/已发布到 COS；R2 Mirror 同步失败/)).toBeInTheDocument()
  await waitFor(() => expect(URL.revokeObjectURL).toHaveBeenCalledWith(`blob:${label}.avif`))
})
it('replaces and revokes previews, preserves Primary failure, confirms clear, and sends complete focal content', async () => {
  setup('/game/showcase/119?tab=assets'); await screen.findByText('素材焦点')
  const input = screen.getByLabelText('选择 Desktop AVIF')
  fireEvent.change(input, { target: { files: [new File(['bad'], 'bad.png', { type: 'image/png' })] } })
  expect(await screen.findByText('请选择原始 AVIF 文件')).toBeInTheDocument()
  for (const name of ['one', 'two']) fireEvent.change(input, { target: { files: [new File(['bytes'], `${name}.avif`, { type: 'image/avif' })] } })
  await waitFor(() => expect(URL.revokeObjectURL).toHaveBeenCalledWith('blob:one.avif'))
  vi.mocked(sendForm).mockRejectedValueOnce(new Error('COS publication failed; no database reference changed'))
  fireEvent.click(screen.getByRole('button', { name: '上传 Desktop' }))
  expect(await screen.findByText(/COS publication failed/)).toBeInTheDocument()
  fireEvent.click(screen.getByRole('button', { name: '取消 Desktop 文件' }))
  expect(URL.revokeObjectURL).toHaveBeenCalledWith('blob:two.avif')
  fireEvent.click(screen.getByRole('button', { name: '清除 Desktop' })); expect(sendJSON).not.toHaveBeenCalled()
  vi.mocked(sendJSON).mockRejectedValueOnce(new Error('valid desktop artwork required'))
  fireEvent.click(screen.getByRole('button', { name: '确认清除' }))
  expect(await screen.findByText('valid desktop artwork required')).toBeInTheDocument()
  expect(sendJSON).toHaveBeenCalledWith('/api/v1/game/showcase/campaigns/119/artwork/desktop', 'DELETE')
  fireEvent.change(screen.getByRole('spinbutton', { name: '水平焦点' }), { target: { value: '0' } })
  fireEvent.click(screen.getByRole('button', { name: '保存焦点' }))
  await waitFor(() => expect(sendJSON).toHaveBeenCalledWith('/api/v1/game/showcase/campaigns/119/content', 'PUT', expect.objectContaining({ focal_x: 0, focal_y: 0.5, internal_name: workspace.internal_name, linked_game_id: 1, locales: workspace.locales, primary_action_type: 'game', secondary_action_type: 'steam' })))
})
it.each(['Desktop', 'Mobile'] as const)('clears %s through its endpoint and releases a staged preview on unmount', async label => {
  workspace.mobile_object_key = 'game/showcase/119/mobile/hash.avif'
  setup('/game/showcase/119?tab=assets'); await screen.findByText('素材焦点')
  fireEvent.click(screen.getByRole('button', { name: `清除 ${label}` })); expect(sendJSON).not.toHaveBeenCalled()
  fireEvent.click(screen.getByRole('button', { name: '确认清除' }))
  await waitFor(() => expect(sendJSON).toHaveBeenCalledWith(`/api/v1/game/showcase/campaigns/119/artwork/${label.toLowerCase()}`, 'DELETE'))
  await waitFor(() => expect(screen.queryByRole('button', { name: '确认清除' })).not.toBeInTheDocument())
  fireEvent.change(screen.getByLabelText(`选择 ${label} AVIF`), { target: { files: [new File(['bytes'], 'staged.avif', { type: 'image/avif' })] } })
  await screen.findByRole('img', { name: `${label} 本地预览` })
  cleanup(); expect(URL.revokeObjectURL).toHaveBeenCalledWith('blob:staged.avif')
})
it('converts shared picker wall time explicitly and saves weight/pin', async () => {
  setup('/game/showcase/119?tab=schedule'); await screen.findByRole('button', { name: '开始时间' })
  expect(document.querySelector('input[type="datetime-local"]')).toBeNull()
  fireEvent.click(screen.getByRole('button', { name: '开始时间' })); fireEvent.click(screen.getByRole('button', { name: '小时加一' })); fireEvent.click(screen.getByRole('button', { name: '确定' }))
  fireEvent.change(screen.getByRole('spinbutton', { name: /Weight/ }), { target: { value: '250' } }); await choose('Pin Position', '#3')
  fireEvent.click(screen.getByRole('button', { name: '保存排期与展示' }))
  await waitFor(() => expect(sendJSON).toHaveBeenCalledWith('/api/v1/game/showcase/campaigns/119/schedule', 'PUT', { starts_at: '2026-10-03T19:00:00+08:00', ends_at: '2026-10-03T21:00:00+08:00', weight: 250, pin_position: 3 }))
})
it.each([['draft', '发布', 'publish'], ['published', '暂停', 'pause'], ['paused', '恢复', 'resume'], ['draft', '归档', 'archive'], ['draft', '删除草稿', 'delete']] as const)('confirms %s → %s through its existing lifecycle endpoint', async (state, label, action) => {
  workspace.state = state; workspace.derived_status = state === 'published' ? 'active' : state
  const { router } = setup('/game/showcase/119'); await screen.findByText('活动概览')
  fireEvent.click(screen.getByRole('button', { name: label })); expect(sendJSON).not.toHaveBeenCalled()
  fireEvent.click(screen.getByRole('button', { name: `确认${label}` }))
  await waitFor(() => expect(sendJSON).toHaveBeenCalledWith(`/api/v1/game/showcase/campaigns/119${action === 'delete' ? '' : `/${action}`}`, action === 'delete' ? 'DELETE' : 'POST'))
  if (action === 'delete') await waitFor(() => expect(router.state.location.pathname).toBe('/game/showcase'))
})
it('uses backend publication diagnostics rather than a frontend readiness calculation', async () => {
  workspace.ready_to_publish = false; workspace.publication_diagnostics = ['server-specific publication diagnostic']
  setup('/game/showcase/119'); expect(await screen.findByText('server-specific publication diagnostic')).toBeInTheDocument()
  expect(screen.getByRole('button', { name: '发布' })).toBeDisabled()
})
it.each(['content', 'assets', 'schedule', 'stats'])('keeps %s readable without content.write', async tab => {
  mocks.capabilities = ['content.read']; setup(`/game/showcase/119?tab=${tab}`); await screen.findByRole('heading', { name: 'October Campaign' })
  expect(screen.queryByRole('button', { name: /^(发布|暂停|恢复|归档|删除草稿|保存|上传|清除)/ })).not.toBeInTheDocument()
  if (tab === 'content') expect(screen.getByRole('textbox', { name: '中文标题' })).toHaveValue('中文活动')
  if (tab === 'assets') expect(screen.getByText(campaignFixture().desktop_object_key!)).toBeInTheDocument()
  if (tab === 'schedule') expect(screen.getByRole('button', { name: '开始时间' })).toBeDisabled()
  if (tab === 'stats') expect(await screen.findByRole('link', { name: '导出 CSV' })).toBeInTheDocument()
})
it('hides create without content.write and links real Audit with resource scope only', async () => {
  mocks.capabilities = ['content.read']; const { router } = setup('/game/showcase?tab=campaigns'); await screen.findByRole('link', { name: 'October Campaign' })
  expect(screen.queryByRole('button', { name: '新建活动' })).not.toBeInTheDocument()
  await act(() => router.navigate('/game/showcase/119?tab=history'))
  expect(await screen.findByText(/当前账号没有 audit.read/)).toBeInTheDocument()
  mocks.capabilities.push('audit.read'); await act(() => router.navigate('/game/showcase/119')); fireEvent.click(screen.getByRole('tab', { name: '历史' }))
  fireEvent.click(await screen.findByRole('link', { name: '打开操作审计' }))
  await waitFor(() => expect(getJSON).toHaveBeenCalledWith('/api/v1/audit/logs?page=1&page_size=20&resource=gfg_showcase_campaign'))
})
it('shows backend KPI totals, estimates, date URL state and CSV without fetching CSV as JSON', async () => {
  const { router } = setup('/game/showcase/119?tab=stats&from=2026-10-01&to=2026-10-03')
  expect(await screen.findByText('12.50%')).toBeInTheDocument()
  expect(screen.getByText(/多日总计为每日 HLL/)).toBeInTheDocument()
  expect(screen.getByRole('link', { name: '导出 CSV' })).toHaveAttribute('href', '/api/v1/game/showcase/campaigns/119/stats.csv?from=2026-10-01&to=2026-10-03')
  expect(screen.getByRole('img', { name: '有效曝光与有效点击每日趋势' })).toBeInTheDocument()
  expect(mocks.chart.setOption).toHaveBeenCalledWith(expect.objectContaining({ series: expect.arrayContaining([expect.objectContaining({ name: '有效曝光', data: [100] }), expect.objectContaining({ name: '有效点击', data: [20] })]) }))
  await choose('统计范围', '7日'); expect(router.state.location.search).toContain('from='); expect(router.state.location.search).toContain('range=7')
  await act(() => router.navigate('/game/showcase/119?tab=stats&from=2020-01-01&to=2026-10-03'))
  expect(await screen.findByText('日期范围须有效、起止有序且最多 366 天。')).toBeInTheDocument()
  expect(screen.queryByRole('link', { name: '导出 CSV' })).not.toBeInTheDocument()
  expect(vi.mocked(getJSON).mock.calls.some(([path]) => path.includes('.csv'))).toBe(false)
})
it('renders empty/populated quality without inventing health scores', async () => {
  const { client } = setup(); await screen.findByText('Showcase 1')
  const details = screen.getByText('数据质量 · 近 7 天').closest('details')!
  details.open = true; fireEvent(details, new Event('toggle'))
  expect(await screen.findByText('暂无数据质量记录')).toBeInTheDocument()
  vi.mocked(getJSON).mockImplementation(async path => path.includes('/analytics/quality') ? { daily: [{ invalid_token_events: 9, invalid_origin_events: 0, filtered_user_agent_events: 0, duplicate_impressions: 2, duplicate_clicks: 1, session_rate_limited: 0, ip_rate_limited: 0, malformed_events: 0 }], timezone: 'Asia/Shanghai' } : compositionFixture())
  await act(() => client.invalidateQueries({ queryKey: ['showcase', 'quality'] }))
  expect(await screen.findByText('Invalid Token')).toBeInTheDocument(); expect(within(details).getByText('9')).toBeInTheDocument()
})
it('disposes and resizes the two-series chart, and renders a normal empty trend', () => {
  const { unmount } = render(<ShowcaseTrend daily={[dailyFixture]} />)
  fireEvent(window, new Event('resize')); expect(mocks.chart.resize).toHaveBeenCalled()
  unmount(); expect(mocks.chart.dispose).toHaveBeenCalled()
  render(<ShowcaseTrend daily={[]} />); expect(screen.getByText('暂无统计趋势')).toBeInTheDocument()
})
