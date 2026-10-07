import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { act, cleanup, fireEvent, render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { createMemoryRouter, RouterProvider } from 'react-router-dom'
import { afterEach, beforeEach, expect, it, vi } from 'vitest'
import { ToastProvider } from '../../app/toast'
import { ApiError, getJSON, listJSON, sendJSON } from '../../lib/api'
import { collectionEndpoint, eligibleCollectionEndpoint, homeKey, loadCollectionWorkspace, workspaceKey } from './api'
import { CollectionListPage } from './collection-list-page'
import { CollectionWorkspacePage, CreateCollectionPage } from './collection-workspace-page'
import { CollectionHomeCurationPage } from './home-curation-page'
import type { CollectionHome, CollectionMember, CollectionWorkspace, GameCollection } from './types'

const auth = vi.hoisted(() => ({ capabilities: ['content.read', 'content.write', 'audit.read'] }))
vi.mock('../auth/auth-context', () => ({ useAuth: () => ({ can: (capability: string) => auth.capabilities.includes(capability) }) }))
vi.mock('../../lib/api', async original => ({ ...await original<typeof import('../../lib/api')>(), getJSON: vi.fn(), listJSON: vi.fn(), sendJSON: vi.fn() }))

function collection(id = 1): GameCollection {
  return { id, code: `collection-${id}`, name: `分区 ${id}`, name_en: `Collection ${id}`, info: '中文简介', info_en: 'Description', status: 'draft', version: 5, member_count: 2, sfw_member_count: 1, home_slot: null, published_at: null, archived_at: null, created_at: '2026-10-06T00:00:00Z', updated_at: '2026-10-06T00:00:00Z' }
}
let current: GameCollection
let members: CollectionMember[]
let home: CollectionHome
beforeEach(() => {
  auth.capabilities = ['content.read', 'content.write', 'audit.read']
  current = collection()
  members = [{ game_id: 1, name: '安全游戏', name_en: 'Safe Game', appid: 101, adult: false }, { game_id: 2, name: '成人游戏', name_en: 'Adult Game', appid: 102, adult: true }]
  home = { revision: 'original', slots: Array.from({ length: 5 }, (_, i) => ({ slot: i + 1, collection: i === 0 ? { ...collection(), status: 'published', home_slot: 1 } : null })) }
  vi.mocked(getJSON).mockImplementation(async path => {
    if (path.endsWith('/home-curation')) return home
    if (path.endsWith('/members')) return { collection_id: current.id, version: current.version, members }
    if (path.startsWith('/api/v1/game/games/')) return { game: { id: 3, name: '新增游戏', name_en: 'New Game', appid: 103 }, tags: [{ tag_id: 140, code: 'adult' }] }
    if (path.endsWith('/2')) return { ...collection(2), status: 'published' }
    return current
  })
  vi.mocked(listJSON).mockImplementation(async path => path.startsWith(collectionEndpoint)
    ? { list: [{ ...collection(), status: 'published' }, { ...collection(2), status: 'published' }], total: 51 }
    : { list: [{ id: '3', label: '新增游戏' }], total: 1 })
  vi.mocked(sendJSON).mockImplementation(async (path, _method, payload) => {
    if (path.endsWith('/home-curation')) {
      const data = payload as { slots: { slot: number; collection_id: number | null }[] }
      home = { revision: 'saved', slots: data.slots.map(s => ({ slot: s.slot, collection: s.collection_id ? collection(s.collection_id) : null })) }; return home
    }
    if (path === collectionEndpoint) { current = { ...current, ...payload as GameCollection, id: 42, version: 1 }; members = []; return current }
    if (path.endsWith('/members')) { const ids = (payload as { game_ids: number[] }).game_ids; members = members.filter(m => ids.includes(m.game_id)) }
    if (path.endsWith('/publish')) current = { ...current, status: 'published' }
    else if (path.endsWith('/unpublish') || path.endsWith('/restore')) current = { ...current, status: 'draft' }
    else if (path.endsWith('/archive')) current = { ...current, status: 'archived' }
    else if (!path.endsWith('/members')) current = { ...current, ...payload as Partial<GameCollection> }
    current = { ...current, version: current.version + 1 }; return current
  })
})
afterEach(() => { cleanup(); vi.restoreAllMocks(); vi.resetAllMocks() })

function setup(path = '/game/collections/1') {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false, gcTime: 0 }, mutations: { retry: false } } })
  const router = createMemoryRouter([
    { path: '/game/collections', element: <CollectionListPage /> },
    { path: '/game/collections/new', element: <CreateCollectionPage /> },
    { path: '/game/collections/home-curation', element: <CollectionHomeCurationPage /> },
    { path: '/game/collections/:id', element: <CollectionWorkspacePage /> },
    { path: '/system/audit', element: <p>操作审计页</p> },
  ], { initialEntries: [path] })
  render(<QueryClientProvider client={client}><ToastProvider><RouterProvider router={router} /></ToastProvider></QueryClientProvider>)
  return { client, router }
}
async function ready() { await screen.findByRole('textbox', { name: '中文名称' }) }
const changeName = (name: string) => fireEvent.change(screen.getByRole('textbox', { name: '中文名称' }), { target: { value: name } })

it('keeps URL search composition-safe and resets pagination only at final commit', async () => {
  const { router } = setup('/game/collections?page_num=2&keyword=old&status=published')
  const input = await screen.findByRole('textbox', { name: '搜索游戏分区' })
  fireEvent.compositionStart(input)
  fireEvent.change(input, { target: { value: 'zhong' } })
  fireEvent.change(input, { target: { value: '中' } })
  expect(router.state.location.search).toContain('page_num=2')
  expect(router.state.location.search).toContain('keyword=old')
  fireEvent.compositionEnd(input, { target: { value: '中文' } })
  fireEvent.change(input, { target: { value: '中文' } })
  await waitFor(() => expect(new URLSearchParams(router.state.location.search).get('keyword')).toBe('中文'))
  expect(router.state.location.search).toContain('page_num=1')
  expect(vi.mocked(listJSON).mock.calls.filter(([, , , q]) => q === '中文')).toHaveLength(1)
  expect(vi.mocked(listJSON).mock.calls.some(([, , , q]) => q === 'zhong' || q === '中')).toBe(false)
  const user = userEvent.setup(); await user.click(screen.getByRole('combobox', { name: '分区状态' })); await user.click(await screen.findByRole('option', { name: '草稿' }))
  expect(router.state.location.search).toContain('status=draft')
})

it('creates a Draft and navigates to its returned real ID', async () => {
  const { router } = setup('/game/collections/new')
  fireEvent.change(screen.getByRole('textbox', { name: /^Code/ }), { target: { value: 'new-collection' } })
  changeName('新的分区')
  fireEvent.change(screen.getByRole('textbox', { name: '英文名称' }), { target: { value: 'New Collection' } })
  fireEvent.click(screen.getByRole('button', { name: '创建草稿' }))
  await waitFor(() => expect(router.state.location.pathname).toBe('/game/collections/42'))
  expect(sendJSON).toHaveBeenCalledWith(collectionEndpoint, 'POST', { code: 'new-collection', name: '新的分区', name_en: 'New Collection', info: '', info_en: '' })
})

it('keeps Code immutable and saves content with the displayed version', async () => {
  setup(); await ready()
  expect(screen.getByRole('textbox', { name: /Code/ })).toHaveAttribute('readonly')
  expect(screen.getByRole('link', { name: '操作审计' })).toHaveAttribute('href', '/system/audit?resource=gfg_game_collection')
  changeName('新名称'); fireEvent.click(screen.getByRole('button', { name: '保存内容' }))
  await waitFor(() => expect(sendJSON).toHaveBeenCalledWith(`${collectionEndpoint}/1`, 'PUT', { version: 5, name: '新名称', name_en: 'Collection 1', info: '中文简介', info_en: 'Description' }))
  expect((vi.mocked(sendJSON).mock.calls[0][2] as Record<string, unknown>).code).toBeUndefined()
})

it('uses complete membership and advances shared baseVersion while retaining dirty content', async () => {
  setup(); await ready(); changeName('保留的内容草稿')
  fireEvent.click(screen.getByRole('button', { name: '移除 成人游戏' }))
  expect(screen.getByRole('button', { name: '发布' })).toBeDisabled()
  fireEvent.click(screen.getByRole('button', { name: '保存成员' }))
  await waitFor(() => expect(sendJSON).toHaveBeenCalledWith(`${collectionEndpoint}/1/members`, 'PUT', { version: 5, game_ids: [1] }))
  await waitFor(() => expect(screen.getByRole('button', { name: '保存内容' })).toBeEnabled())
  expect(screen.getByRole('textbox', { name: '中文名称' })).toHaveValue('保留的内容草稿')
  fireEvent.click(screen.getByRole('button', { name: '保存内容' }))
  await waitFor(() => expect(sendJSON).toHaveBeenLastCalledWith(`${collectionEndpoint}/1`, 'PUT', expect.objectContaining({ version: 6, name: '保留的内容草稿' })))
})

it('content save preserves membership draft and rebases its later complete-set save', async () => {
  setup(); await ready(); fireEvent.click(screen.getByRole('button', { name: '移除 成人游戏' })); changeName('保存内容先')
  fireEvent.click(screen.getByRole('button', { name: '保存内容' }))
  await waitFor(() => expect(screen.getByRole('button', { name: '保存成员' })).toBeEnabled())
  expect(screen.queryByText('成人游戏')).not.toBeInTheDocument()
  fireEvent.click(screen.getByRole('button', { name: '保存成员' }))
  await waitFor(() => expect(sendJSON).toHaveBeenLastCalledWith(`${collectionEndpoint}/1/members`, 'PUT', { version: 6, game_ids: [1] }))
})

it('adds games with code-based Adult inspection without drag or order controls', async () => {
  setup(); await ready()
  fireEvent.focus(screen.getByRole('combobox'))
  fireEvent.click(await screen.findByRole('option', { name: '新增游戏' }))
  await screen.findByRole('button', { name: '移除 新增游戏' })
  expect(getJSON).toHaveBeenCalledWith('/api/v1/game/games/3/workspace')
  expect(screen.getAllByText('Adult')).toHaveLength(2)
  expect(screen.queryByRole('button', { name: /上移|下移|排序/ })).not.toBeInTheDocument()
  expect(document.querySelector('[draggable="true"]')).toBeNull()
  fireEvent.click(screen.getByRole('button', { name: '保存成员' }))
  await waitFor(() => expect(sendJSON).toHaveBeenCalledWith(`${collectionEndpoint}/1/members`, 'PUT', { version: 5, game_ids: [1, 2, 3] }))
})

it('background version changes do not overwrite dirty drafts or silently rebase; 409 retains both', async () => {
  const { client } = setup(); await ready(); changeName('本地草稿'); fireEvent.click(screen.getByRole('button', { name: '移除 成人游戏' }))
  await act(async () => { client.setQueryData<CollectionWorkspace>(workspaceKey(1), { collection: { ...current, name: '他人修改', version: 9 }, members }) })
  vi.mocked(sendJSON).mockRejectedValue(new ApiError('此游戏分区已被其他操作修改，请重新加载后重试。', 409))
  fireEvent.click(screen.getByRole('button', { name: '保存内容' }))
  expect(await screen.findByText(/草稿已保留/)).toBeInTheDocument()
  expect(sendJSON).toHaveBeenCalledTimes(1)
  expect(sendJSON).toHaveBeenCalledWith(`${collectionEndpoint}/1`, 'PUT', expect.objectContaining({ version: 5 }))
  expect(screen.getByRole('textbox', { name: '中文名称' })).toHaveValue('本地草稿')
  expect(screen.queryByText('成人游戏')).not.toBeInTheDocument()
  expect(screen.getByRole('button', { name: '保存成员' })).toBeDisabled()
  vi.spyOn(window, 'confirm').mockReturnValue(true)
  current = { ...current, name: '服务器最新', version: 9 }
  fireEvent.click(screen.getByRole('button', { name: '重新加载 / 放弃修改' }))
  await waitFor(() => expect(screen.getByRole('textbox', { name: '中文名称' })).toHaveValue('服务器最新'))
})

it('clean background refresh advances data and version normally', async () => {
  const { client } = setup(); await ready()
  await act(async () => { client.setQueryData(workspaceKey(1), { collection: { ...current, name: '刷新名称', version: 7 }, members }) })
  await waitFor(() => expect(screen.getByRole('textbox', { name: '中文名称' })).toHaveValue('刷新名称'))
  changeName('更新'); fireEvent.click(screen.getByRole('button', { name: '保存内容' }))
  await waitFor(() => expect(sendJSON).toHaveBeenCalledWith(`${collectionEndpoint}/1`, 'PUT', expect.objectContaining({ version: 7 })))
})

it.each([['draft', '发布', 'publish'], ['draft', '归档', 'archive'], ['published', '取消发布', 'unpublish'], ['archived', '恢复为草稿', 'restore']] as const)('confirms %s lifecycle %s before a versioned write', async (status, label, action) => {
  current.status = status; setup(); await ready()
  fireEvent.click(screen.getByRole('button', { name: label }))
  expect(sendJSON).not.toHaveBeenCalled()
  fireEvent.click(within(await screen.findByRole('dialog')).getByRole('button', { name: '确认操作' }))
  await waitFor(() => expect(sendJSON).toHaveBeenCalledWith(`${collectionEndpoint}/1/${action}`, 'POST', { version: 5 }))
})

it('read-only users inspect complete content and members without write or audit actions', async () => {
  auth.capabilities = ['content.read']; setup(); await ready()
  expect(screen.getByRole('textbox', { name: '中文名称' })).toHaveValue('分区 1')
  expect(screen.getByRole('textbox', { name: '中文名称' })).toBeDisabled()
  expect(screen.getByText('成人游戏')).toBeInTheDocument()
  expect(screen.queryByRole('button', { name: /保存|发布|归档|恢复|移除/ })).not.toBeInTheDocument()
  expect(screen.queryByRole('link', { name: '操作审计' })).not.toBeInTheDocument()
})

it('archived workspace is read-only with Restore as the only mutation', async () => {
  current.status = 'archived'; setup(); await ready()
  expect(screen.getByRole('textbox', { name: '中文简介' })).toBeDisabled()
  expect(screen.getByText('安全游戏')).toBeInTheDocument()
  expect(screen.queryByRole('combobox')).not.toBeInTheDocument()
  expect(screen.queryByRole('button', { name: /保存内容|保存成员|发布|归档|移除/ })).not.toBeInTheDocument()
  expect(screen.getByRole('button', { name: '恢复为草稿' })).toBeEnabled()
})

it('protects dirty content/member navigation and unload and disables lifecycle', async () => {
  const confirm = vi.spyOn(window, 'confirm').mockReturnValue(false)
  const { router } = setup(); await ready(); fireEvent.click(screen.getByRole('button', { name: '移除 成人游戏' }))
  expect(screen.getByRole('button', { name: '发布' })).toBeDisabled()
  expect(screen.getByRole('button', { name: '归档' })).toBeDisabled()
  const event = new Event('beforeunload', { cancelable: true }); window.dispatchEvent(event); expect(event.defaultPrevented).toBe(true)
  fireEvent.click(screen.getByRole('link', { name: '返回' })); await waitFor(() => expect(confirm).toHaveBeenCalled())
  expect(router.state.location.pathname).toBe('/game/collections/1')
})

it('home has five editable slots, fixed sixth, eligible picker, and full placement payload', async () => {
  setup('/game/collections/home-curation'); await screen.findByRole('heading', { name: '#6 全部分区' })
  expect(screen.getAllByRole('combobox')).toHaveLength(5)
  expect(screen.getByText('固定入口 · 无需配置')).toBeInTheDocument()
  expect(listJSON).toHaveBeenCalledWith(eligibleCollectionEndpoint, 1, 50, '')
  fireEvent.focus(screen.getAllByRole('combobox')[1])
  expect(screen.queryByRole('option', { name: /分区 1/ })).not.toBeInTheDocument()
  fireEvent.click(await screen.findByRole('option', { name: /分区 2/ }))
  await waitFor(() => expect(screen.getByRole('button', { name: '保存编排' })).toBeEnabled())
  fireEvent.click(screen.getByRole('button', { name: '清空第 1 位' }))
  fireEvent.click(screen.getByRole('button', { name: '保存编排' }))
  await waitFor(() => expect(sendJSON).toHaveBeenCalledWith(`${collectionEndpoint}/home-curation`, 'PUT', { revision: 'original', slots: Array.from({ length: 5 }, (_, i) => ({ slot: i + 1, collection_id: i === 1 ? 2 : null })) }))
})

it('home dirty draft keeps original revision through background changes and 409', async () => {
  const { client } = setup('/game/collections/home-curation'); await screen.findByText('#6 全部分区')
  fireEvent.click(screen.getByRole('button', { name: '清空第 1 位' }))
  await act(async () => { client.setQueryData(homeKey, { ...home, revision: 'external' }) })
  vi.mocked(sendJSON).mockRejectedValue(new ApiError('编排已修改', 409))
  fireEvent.click(screen.getByRole('button', { name: '保存编排' }))
  expect(await screen.findByText(/编排草稿已保留/)).toBeInTheDocument()
  expect(sendJSON).toHaveBeenCalledTimes(1)
  expect(sendJSON).toHaveBeenCalledWith(`${collectionEndpoint}/home-curation`, 'PUT', expect.objectContaining({ revision: 'original' }))
  expect(screen.getByRole('button', { name: '清空第 1 位' })).toBeDisabled()
  const event = new Event('beforeunload', { cancelable: true }); window.dispatchEvent(event); expect(event.defaultPrevented).toBe(true)
})

it('home is inspectable but not editable for content.read', async () => {
  auth.capabilities = ['content.read']; setup('/game/collections/home-curation'); await screen.findByText('#6 全部分区')
  expect(screen.getByRole('link', { name: '分区 1' })).toBeInTheDocument()
  expect(screen.queryByRole('combobox')).not.toBeInTheDocument()
  expect(screen.queryByRole('button', { name: '保存编排' })).not.toBeInTheDocument()
})

it('does not combine mismatched content and membership versions', async () => {
  vi.mocked(getJSON).mockImplementation(async path => path.endsWith('/members') ? { version: 9, members } : current)
  await expect(loadCollectionWorkspace(1)).rejects.toThrow('正在更新')
  expect(getJSON).toHaveBeenCalledTimes(4)
})
