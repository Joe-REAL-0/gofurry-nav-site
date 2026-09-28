import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { act, cleanup, fireEvent, render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { createMemoryRouter, RouterProvider } from 'react-router-dom'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { ToastProvider } from '../../app/toast'
import { ApiError, getJSON, listJSON, sendJSON } from '../../lib/api'
import { ReleaseNoteEditorPage, ReleaseNoteListPage } from './release-note-pages'
import { releaseNoteKey, releaseNotesEndpoint as endpoint, type ReleaseNote } from './release-note-model'

const auth = vi.hoisted(() => ({ write: true }))
vi.mock('../auth/auth-context', () => ({ useAuth: () => ({ can: (capability: string) => capability === 'content.read' || capability === 'content.write' && auth.write }) }))
vi.mock('../../lib/api', async original => ({ ...await original<typeof import('../../lib/api')>(), getJSON: vi.fn(), listJSON: vi.fn(), sendJSON: vi.fn() }))

const baseline: ReleaseNote = {
  id: 17, version: 'Preview', commit_sha: 'abcdef0123456789', title: '中文标题', title_en: 'English title',
  summary: '中文摘要', summary_en: 'English summary', body: '中文正文', body_en: 'English body',
  publication_state: 'draft', published_at: null, create_time: '2026-09-28 10:00:00', update_time: '2026-09-28 10:00:00', deleted: false,
}
let stored: ReleaseNote
const clients: QueryClient[] = []
const routers: ReturnType<typeof createMemoryRouter>[] = []

beforeEach(() => {
  vi.resetAllMocks(); auth.write = true; stored = { ...baseline }
  vi.spyOn(Date, 'now').mockReturnValue(Date.parse('2026-09-28T04:00:00Z'))
  vi.spyOn(window, 'confirm').mockReturnValue(false)
  vi.mocked(getJSON).mockImplementation(async () => ({ ...stored }))
  vi.mocked(listJSON).mockImplementation(async () => ({ total: 1, list: [{ ...stored }] }))
  vi.mocked(sendJSON).mockImplementation(async (path, method, body) => {
    if (method === 'DELETE') return undefined
    if (path.endsWith('/unpublish')) stored = { ...stored, publication_state: 'draft' }
    else if (path.endsWith('/publish')) stored = { ...stored, publication_state: 'published', published_at: (body as { published_at?: string } | undefined)?.published_at || '2026-09-28 12:00:00' }
    else stored = { ...stored, ...body as Partial<ReleaseNote> }
    return { ...stored }
  })
})
afterEach(() => { cleanup(); clients.splice(0).forEach(client => client.clear()); routers.splice(0).forEach(router => router.dispose()); vi.restoreAllMocks() })

function workspace(path = '/nav/update-notices/17') {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } })
  clients.push(client)
  const router = createMemoryRouter([
    { path: '/nav/update-notices', element: <ReleaseNoteListPage /> },
    { path: '/nav/update-notices/new', element: <ReleaseNoteEditorPage /> },
    { path: '/nav/update-notices/:id', element: <ReleaseNoteEditorPage /> },
  ], { initialEntries: [path] })
  routers.push(router)
  render(<QueryClientProvider client={client}><ToastProvider><RouterProvider router={router} /></ToastProvider></QueryClientProvider>)
  return { router, client, user: userEvent.setup() }
}

async function editor() { return screen.findByRole('textbox', { name: '中文标题' }) }
async function confirmAction(user: ReturnType<typeof userEvent.setup>, action: string) {
  await user.click(screen.getByRole('button', { name: action }))
  const dialog = await screen.findByRole('dialog')
  await user.click(within(dialog).getByRole('button', { name: `确认${action}` }))
}
function unload() { const event = new Event('beforeunload', { cancelable: true }); window.dispatchEvent(event); return event.defaultPrevented }

describe('Release Notes list', () => {
  it('shows metadata/status, searches/paginates and navigates rows', async () => {
    vi.mocked(listJSON).mockResolvedValue({ total: 101, list: [stored, { ...stored, id: 18, title: 'Future', publication_state: 'published', published_at: '2099-01-01 12:00:00' }, { ...stored, id: 19, title: 'Public', publication_state: 'published', published_at: '2020-01-01 12:00:00' }] })
    const { user, router } = workspace('/nav/update-notices')
    await screen.findByText('Draft')
    expect(screen.getByText('Scheduled')).toBeInTheDocument(); expect(screen.getByText('Published')).toBeInTheDocument()
    expect(screen.getAllByTitle('abcdef0123456789')[0]).toHaveTextContent('abcdef0')
    await user.type(screen.getByRole('textbox', { name: '搜索列表' }), 'Preview')
    await waitFor(() => expect(listJSON).toHaveBeenLastCalledWith(endpoint, 1, 50, 'Preview'))
    await user.click(screen.getByRole('button', { name: '下一页' }))
    await waitFor(() => expect(listJSON).toHaveBeenLastCalledWith(endpoint, 2, 50, 'Preview'))
    await user.click(screen.getByText('Draft'))
    await waitFor(() => expect(router.state.location.pathname).toBe('/nav/update-notices/17'))
  })
  it('hides create for readers but keeps detail links', async () => {
    auth.write = false
    workspace('/nav/update-notices')
    expect(await screen.findByRole('link', { name: '中文标题' })).toHaveAttribute('href', '/nav/update-notices/17')
    expect(screen.queryByRole('button', { name: '新建 Release Note' })).not.toBeInTheDocument()
  })
  it('retries a list failure', async () => {
    vi.mocked(listJSON).mockRejectedValueOnce(new Error('列表暂不可用'))
    const { user } = workspace('/nav/update-notices')
    await screen.findByText('列表暂不可用')
    await user.click(screen.getByRole('button', { name: /重试/ }))
    expect(await screen.findByText('Draft')).toBeInTheDocument()
  })
})

describe('Release editor content and lifecycle', () => {
  it('opens new locally and creates a Draft only on Save', async () => {
    const { user, router, client } = workspace('/nav/update-notices/new')
    const title = await editor()
    expect(getJSON).not.toHaveBeenCalled(); expect(sendJSON).not.toHaveBeenCalled()
    await user.type(title, '新草稿')
    await user.click(screen.getByRole('button', { name: '保存草稿' }))
    await waitFor(() => expect(router.state.location.pathname).toBe('/nav/update-notices/17'))
    expect(sendJSON).toHaveBeenCalledExactlyOnceWith(endpoint, 'POST', expect.objectContaining({ title: '新草稿', title_en: '', body: '', version: null, commit_sha: null, published_at: '' }))
    expect(vi.mocked(sendJSON).mock.calls[0][2]).not.toHaveProperty('publication_state')
    expect(client.getQueryData<ReleaseNote>(releaseNoteKey(17))?.title).toBe('新草稿')
    expect(unload()).toBe(false); expect(window.confirm).not.toHaveBeenCalled()
  })
  it('keeps both languages and shared metadata in one PUT', async () => {
    const { user } = workspace(); const title = await editor()
    await user.clear(title); await user.type(title, '中文修改')
    await user.click(screen.getByRole('button', { name: 'English' }))
    expect(screen.getByRole('textbox', { name: 'English标题' })).toHaveValue('English title')
    await user.clear(screen.getByRole('textbox', { name: 'English标题' })); await user.type(screen.getByRole('textbox', { name: 'English标题' }), 'Edited English')
    await user.type(screen.getByRole('textbox', { name: 'English摘要' }), ' revised')
    await user.click(screen.getByRole('button', { name: '中文' }))
    expect(screen.getByRole('textbox', { name: '中文标题' })).toHaveValue('中文修改')
    expect(screen.getByRole('textbox', { name: 'Version' })).toHaveValue('Preview')
    await user.click(screen.getByRole('button', { name: '保存修改' }))
    await waitFor(() => expect(sendJSON).toHaveBeenCalledWith(`${endpoint}/17`, 'PUT', expect.objectContaining({ title: '中文修改', title_en: 'Edited English', summary_en: 'English summary revised', body: '中文正文', body_en: 'English body' })))
    await waitFor(() => expect(unload()).toBe(false))
  })
  it('saves dirty content before publishing and sends no browser-now payload', async () => {
    const { user } = workspace(); await user.type(await editor(), ' edited')
    await confirmAction(user, '立即发布')
    await waitFor(() => expect(sendJSON).toHaveBeenCalledTimes(2))
    expect(vi.mocked(sendJSON).mock.calls[0]).toEqual([`${endpoint}/17`, 'PUT', expect.objectContaining({ title: '中文标题 edited' })])
    expect(vi.mocked(sendJSON).mock.calls[1]).toEqual([`${endpoint}/17/publish`, 'POST'])
    expect(await screen.findByText('Published')).toBeInTheDocument()
  })
  it('never publishes after a failed save and preserves local input', async () => {
    vi.mocked(sendJSON).mockRejectedValueOnce(new Error('保存被拒绝'))
    const { user } = workspace(); await user.type(await editor(), ' unsaved')
    await confirmAction(user, '立即发布')
    await screen.findAllByText('保存被拒绝')
    expect(sendJSON).toHaveBeenCalledTimes(1)
    expect(screen.getByRole('textbox', { name: '中文标题' })).toHaveValue('中文标题 unsaved')
    expect(unload()).toBe(true); expect(screen.getByText('Draft')).toBeInTheDocument()
  })
  it('creates then publishes a new record without a redundant PUT', async () => {
    const { user, router } = workspace('/nav/update-notices/new')
    await user.type(await editor(), 'New release')
    fireEvent.change(screen.getByRole('textbox', { name: '中文正文' }), { target: { value: '# New content' } })
    await confirmAction(user, '立即发布')
    await waitFor(() => expect(router.state.location.pathname).toBe('/nav/update-notices/17'))
    expect(vi.mocked(sendJSON).mock.calls.map(call => [call[0], call[1]])).toEqual([[endpoint, 'POST'], [`${endpoint}/17/publish`, 'POST']])
    expect(stored.body).toBe('# New content'); expect(unload()).toBe(false)
  })
  it('does not duplicate a created Draft when publish fails', async () => {
    vi.mocked(sendJSON).mockImplementationOnce(async () => {
      stored = { ...stored, title: 'Created', body: 'body' }
      return { ...stored }
    }).mockRejectedValueOnce(new Error('发布失败'))
    const { user, router } = workspace('/nav/update-notices/new')
    await user.type(await editor(), 'Created'); fireEvent.change(screen.getByRole('textbox', { name: '中文正文' }), { target: { value: 'body' } })
    await confirmAction(user, '立即发布')
    await waitFor(() => expect(router.state.location.pathname).toBe('/nav/update-notices/17'))
    await screen.findByRole('heading', { name: '编辑 Release Note' })
    expect(screen.getByText('Draft')).toBeInTheDocument()
    expect(screen.getByRole('textbox', { name: '中文标题' })).toHaveValue('Created')
    expect(screen.getAllByText('发布失败').length).toBeGreaterThan(0)
  })
  it('requires a future date and saves the chosen China time before scheduling', async () => {
    const { user } = workspace(); await editor()
    await confirmAction(user, '定时发布')
    await screen.findAllByText('定时发布需要选择未来时间（Asia/Shanghai）。')
    expect(sendJSON).not.toHaveBeenCalled()
    await user.click(screen.getByRole('button', { name: '发布时间（Asia/Shanghai）' }))
    await user.click(screen.getByRole('button', { name: '29' }))
    await user.click(screen.getByRole('button', { name: '确定' }))
    await confirmAction(user, '定时发布')
    await waitFor(() => expect(sendJSON).toHaveBeenCalledTimes(2))
    expect(vi.mocked(sendJSON).mock.calls[1]).toEqual([`${endpoint}/17/publish`, 'POST', { published_at: '2026-09-29T12:00' }])
    expect(await screen.findByText('Scheduled')).toBeInTheDocument()
  })
  it('blocks dirty Unpublish; clean Unpublish preserves the date and editor', async () => {
    stored = { ...stored, publication_state: 'published', published_at: '2026-09-28 12:00:00' }
    const { user, router } = workspace(); const title = await editor()
    await user.type(title, ' edited')
    expect(screen.getByRole('button', { name: '取消发布' })).toBeDisabled()
    expect(sendJSON).not.toHaveBeenCalled()
    fireEvent.change(title, { target: { value: baseline.title } })
    await waitFor(() => expect(screen.getByRole('button', { name: '取消发布' })).toBeEnabled())
    await confirmAction(user, '取消发布')
    expect(await screen.findByText('Draft')).toBeInTheDocument()
    expect(sendJSON).toHaveBeenCalledExactlyOnceWith(`${endpoint}/17/unpublish`, 'POST')
    expect(router.state.location.pathname).toBe('/nav/update-notices/17')
    expect(stored.published_at).toBe('2026-09-28 12:00:00')
  })
  it('requires confirmation for delete, then leaves without saving dirty content', async () => {
    const { user, router } = workspace(); await user.type(await editor(), ' unsaved')
    await user.click(screen.getByRole('button', { name: '删除公告' }))
    expect(sendJSON).not.toHaveBeenCalled()
    await user.click(within(await screen.findByRole('dialog')).getByRole('button', { name: '确认删除公告' }))
    await waitFor(() => expect(router.state.location.pathname).toBe('/nav/update-notices'))
    expect(sendJSON).toHaveBeenCalledExactlyOnceWith(`${endpoint}/17`, 'DELETE')
    expect(window.confirm).not.toHaveBeenCalled()
  })
  it('keeps state on a failed unpublish or delete', async () => {
    stored = { ...stored, publication_state: 'published', published_at: '2020-01-01 12:00:00' }
    vi.mocked(sendJSON).mockRejectedValue(new Error('操作失败'))
    const { user, router } = workspace(); await editor()
    await confirmAction(user, '取消发布'); await screen.findAllByText('操作失败')
    expect(screen.getByText('Published')).toBeInTheDocument()
    await confirmAction(user, '删除公告'); await waitFor(() => expect(sendJSON).toHaveBeenCalledTimes(2))
    expect(router.state.location.pathname).toBe('/nav/update-notices/17')
  })
  it('readers can switch language/preview but cannot edit or mutate', async () => {
    auth.write = false
    const { user } = workspace(); expect(await editor()).toHaveAttribute('readonly')
    for (const name of ['保存修改', '立即发布', '定时发布', '取消发布', '删除公告', 'Bold']) expect(screen.queryByRole('button', { name })).not.toBeInTheDocument()
    await user.click(screen.getByRole('button', { name: 'English' }))
    expect(screen.getByRole('textbox', { name: 'English标题' })).toHaveValue('English title')
    expect(screen.getByRole('region', { name: 'English正文预览' })).toHaveTextContent('English body')
    expect(sendJSON).not.toHaveBeenCalled(); expect(unload()).toBe(false)
  })
  it('prevents readers from opening new and presents missing detail clearly', async () => {
    auth.write = false
    const { router } = workspace('/nav/update-notices/new')
    expect(screen.queryByRole('textbox')).not.toBeInTheDocument(); expect(sendJSON).not.toHaveBeenCalled()
    vi.mocked(getJSON).mockRejectedValue(new ApiError('missing', 404))
    await act(async () => { await router.navigate('/nav/update-notices/404') })
    expect(await screen.findByText('未找到该公告，可能已被删除。')).toBeInTheDocument()
  })
  it('guards internal navigation and unload only while dirty, clearing after save', async () => {
    const { user, router } = workspace(); const title = await editor()
    expect(unload()).toBe(false)
    await user.type(title, ' changed'); expect(unload()).toBe(true)
    await user.click(screen.getByRole('link', { name: '返回更新公告' }))
    await waitFor(() => expect(window.confirm).toHaveBeenCalledWith('存在未保存的修改，确定离开吗？'))
    expect(router.state.location.pathname).toBe('/nav/update-notices/17')
    await user.click(screen.getByRole('button', { name: '保存修改' }))
    await waitFor(() => expect(unload()).toBe(false))
    vi.mocked(window.confirm).mockClear()
    await user.click(screen.getByRole('link', { name: '返回更新公告' }))
    await waitFor(() => expect(router.state.location.pathname).toBe('/nav/update-notices'))
    expect(window.confirm).not.toHaveBeenCalled()
  })
  it('can explicitly discard edits through the router blocker', async () => {
    const { user, router } = workspace(); await user.type(await editor(), ' changed')
    vi.mocked(window.confirm).mockReturnValue(true)
    await user.click(screen.getByRole('link', { name: '返回更新公告' }))
    await waitFor(() => expect(router.state.location.pathname).toBe('/nav/update-notices'))
    expect(sendJSON).not.toHaveBeenCalled()
  })
  it('toolbar changes the selected textarea source and previews without saving', async () => {
    const { user } = workspace(); await editor()
    const body = screen.getByRole('textbox', { name: '中文正文' }) as HTMLTextAreaElement
    body.focus(); body.setSelectionRange(0, body.value.length)
    await user.click(screen.getByRole('button', { name: 'Bold' }))
    expect(body).toHaveValue('**中文正文**')
    expect(screen.getByRole('region', { name: '中文正文预览', hidden: true }).querySelector('strong')).toHaveTextContent('中文正文')
    expect(sendJSON).not.toHaveBeenCalled()
  })
})
