import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { act, cleanup, fireEvent, render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { createMemoryRouter, RouterProvider } from 'react-router-dom'
import { afterEach, beforeEach, expect, it, vi } from 'vitest'
import { getJSON, sendJSON } from '../../lib/api'
import { ShowcaseDiscovery } from './showcase-discovery'
import { candidateFixture } from './test-fixtures'
import type { Candidate, CandidatePage } from './types'

const auth = vi.hoisted(() => ({ write: true }))
vi.mock('../auth/auth-context', () => ({ useAuth: () => ({ can: (capability: string) => capability === 'content.read' || auth.write }) }))
vi.mock('../../lib/api', () => ({ getJSON: vi.fn(), sendJSON: vi.fn() }))
const pending = (): Candidate => ({ ...candidateFixture(), status: 'pending_approval', excluded_reasons: ['not_approved'] })
const result = (items: Candidate[] = []): CandidatePage => ({ items, total: items.length, counts: { eligible: 0, pending_approval: 42, blocked: 90, all: 132 } })
beforeEach(() => { auth.write = true; vi.mocked(getJSON).mockResolvedValue(result()) })
afterEach(() => { cleanup(); vi.resetAllMocks(); vi.useRealTimers() })
function setup(search = '') {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false, gcTime: 0 } } })
  const router = createMemoryRouter([{ path: '/game/showcase', element: <ShowcaseDiscovery /> }], { initialEntries: [`/game/showcase${search}`] })
  render(<QueryClientProvider client={client}><RouterProvider router={router} /></QueryClientProvider>)
  return router
}
async function choose(label: string, option: string) {
  const user = userEvent.setup()
  await user.click(screen.getByRole('combobox', { name: label }))
  await user.click(await screen.findByRole('option', { name: option }))
}
function lastQuery() { return new URL(vi.mocked(getJSON).mock.lastCall![0], 'https://admin.test').searchParams }

it('defaults to eligible, uses server counts, and explains an empty pool with a pending action', async () => {
  const router = setup()
  await screen.findByText('暂无可用候选')
  expect(lastQuery().get('status')).toBe('eligible'); expect(lastQuery().get('sort')).toBe('pool')
  expect(screen.getByRole('button', { name: '待开启42' })).toBeInTheDocument()
  expect(screen.getByRole('button', { name: '可用候选0' })).toHaveAttribute('aria-pressed', 'true')
  vi.mocked(getJSON).mockResolvedValue(result([pending()]))
  fireEvent.click(screen.getByRole('button', { name: '查看待开启' }))
  await screen.findByText('Candidate game')
  expect(lastQuery().get('status')).toBe('pending_approval')
  expect(router.state.location.search).toContain('candidate_status=pending_approval')
  expect(screen.getByRole('link', { name: '去开启' })).toHaveAttribute('href', '/game/games/1?tab=classification')
  expect(sendJSON).not.toHaveBeenCalled()
})

it('restores URL filters, sends global queries, resets paging and preserves browser navigation', async () => {
  const router = setup('?pool=trending&locale=en&candidate_status=all&candidate_keyword=%23105&candidate_sort=game_id&page_num=3&excluded_reason=adult')
  await screen.findByText('当前条件下没有游戏')
  expect(Object.fromEntries(lastQuery())).toEqual({ pool: 'trending', lang: 'en', region: 'CN', page_num: '3', page_size: '20', status: 'all', sort: 'game_id', keyword: '#105', excluded_reason: 'adult' })
  await choose('候选列表排序', '热度选取权重由高到低')
  await waitFor(() => expect(lastQuery().get('sort')).toBe('pool'))
  expect(lastQuery().get('page_num')).toBe('1')
  await choose('自动发现类型', '新作推荐')
  await waitFor(() => expect(lastQuery().get('pool')).toBe('new_release'))
  expect(lastQuery().has('excluded_reason')).toBe(false)
  expect(screen.getByRole('combobox', { name: '候选列表排序' })).toHaveTextContent('首次可玩由新到旧')
  await act(() => router.navigate(-1))
  expect(screen.getByRole('combobox', { name: '自动发现类型' })).toHaveTextContent('热度上升')
  expect(screen.getByRole('combobox', { name: '排除原因' })).toHaveTextContent('成人内容不参与自动展柜')
})

it('debounces name search before requesting all-data filtering and resets the page', async () => {
  setup('?candidate_status=all&page_num=4')
  await screen.findByText('当前条件下没有游戏')
  vi.useFakeTimers()
  const count = vi.mocked(getJSON).mock.calls.length
  fireEvent.change(screen.getByRole('textbox', { name: '搜索列表' }), { target: { value: '狼' } })
  await act(() => vi.advanceTimersByTimeAsync(200))
  fireEvent.change(screen.getByRole('textbox', { name: '搜索列表' }), { target: { value: '狼与猫' } })
  await act(() => vi.advanceTimersByTimeAsync(299))
  expect(getJSON).toHaveBeenCalledTimes(count)
  await act(() => vi.advanceTimersByTimeAsync(1))
  expect(getJSON).toHaveBeenCalledTimes(count + 1)
  expect(lastQuery().get('keyword')).toBe('狼与猫'); expect(lastQuery().get('page_num')).toBe('1')
})

it('switches exclusion search to all statuses and keeps filtering ahead of server pagination', async () => {
  setup('?pool=trending')
  await screen.findByText('暂无可用候选')
  fireEvent.click(screen.getByText('高级筛选'))
  await choose('排除原因', '热度条件未满足')
  await waitFor(() => expect(lastQuery().get('excluded_reason')).toBe('trending_requirements'))
  expect(lastQuery().get('status')).toBe('all')
  vi.mocked(getJSON).mockResolvedValue({ ...result([pending()]), total: 42 })
  fireEvent.click(screen.getByRole('button', { name: '清除筛选' }))
  await screen.findByText('Candidate game')
  fireEvent.click(screen.getByRole('button', { name: '下一页' }))
  await waitFor(() => expect(lastQuery().get('page_num')).toBe('2'))
  await choose('每页条数', '50')
  await waitFor(() => expect(lastQuery().get('page_size')).toBe('50'))
  expect(lastQuery().get('page_num')).toBe('1')
})

it('keeps technical evidence in an accessible dialog and hides mutation wording for readonly users', async () => {
  auth.write = false
  vi.mocked(getJSON).mockResolvedValue(result([pending()]))
  setup('?pool=trending&candidate_status=pending_approval')
  await screen.findByText('Candidate game')
  expect(screen.getByText('热度条件已满足')).toBeInTheDocument()
  expect(screen.queryByText(/增长动量/)).not.toBeInTheDocument()
  expect(screen.queryByRole('link', { name: '去开启' })).not.toBeInTheDocument()
  expect(screen.getByRole('link', { name: '查看游戏' })).toHaveAttribute('href', '/game/games/1?tab=classification')
  const user = userEvent.setup()
  const trigger = screen.getByRole('button', { name: '查看 Candidate game 的诊断' })
  trigger.focus(); await user.keyboard('{Enter}')
  const dialog = await screen.findByRole('dialog', { name: 'Candidate game · 候选诊断' })
  expect(within(dialog).getByText(/增长动量 1,800/)).toBeInTheDocument()
  expect(within(dialog).getByText(/此前 7 天/)).toBeInTheDocument()
  await user.keyboard('{Escape}')
  await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument())
  expect(sendJSON).not.toHaveBeenCalled()
})

it('shows specific gate failures and first availability without reusing canonical release dates', async () => {
  const row = { ...candidateFixture(), first_available: '2026-10-04', release: { availability: 'available', precision: 'day', exact_date: '2024-01-01', year: 2024, month: null, quarter: null, window_start: null, window_end: null }, excluded_reasons: ['new_release_requirements'], pool_failures: ['first_available_expired'] }
  vi.mocked(getJSON).mockResolvedValue(result([row]))
  setup('?pool=new_release&candidate_status=blocked')
  expect(await screen.findByText('首次可玩：2026-10-04')).toBeInTheDocument()
  expect(screen.getByText('首次可玩已超过 30 天')).toBeInTheDocument()
  expect(screen.queryByText(/2024-01-01/)).not.toBeInTheDocument()
  expect(screen.queryByText(/增长动量/)).not.toBeInTheDocument()
})

it('shows query errors with retry rather than an empty candidate claim', async () => {
  vi.mocked(getJSON).mockRejectedValueOnce(new Error('Diagnostic read failed'))
  setup()
  await screen.findByText('Diagnostic read failed')
  expect(screen.queryByText('暂无可用候选')).not.toBeInTheDocument()
  fireEvent.click(screen.getByRole('button', { name: '重试' }))
  await screen.findByText('暂无可用候选')
})

it('distinguishes missing player facts from a zero-player observation', async () => {
  const row = { ...candidateFixture(), excluded_reasons: ['trending_requirements'], pool_failures: ['player_facts_missing'], trending: { recent: { average: 0, observed_days: 0, coverage: null }, baseline: { average: 0, observed_days: 0, coverage: null }, momentum: 0, weight: 0, eligible: false } }
  vi.mocked(getJSON).mockResolvedValue(result([row]))
  setup('?pool=trending&candidate_status=blocked')
  expect(await screen.findByText('暂无已结算玩家数据')).toBeInTheDocument()
  expect(screen.queryByText(/近期 0 人/)).not.toBeInTheDocument()
  fireEvent.click(screen.getByRole('button', { name: '查看 Candidate game 的诊断' }))
  const dialog = await screen.findByRole('dialog')
  expect(within(dialog).getByText('近期 3 天：没有有效观察日，无法计算平均人数。')).toBeInTheDocument()
})
