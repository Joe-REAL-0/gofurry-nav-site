import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { useState } from 'react'
import { MemoryRouter } from 'react-router-dom'
import { afterEach, expect, it, vi } from 'vitest'
import { getJSON } from '../../lib/api'
import { GlobalSearch } from './global-search'

vi.mock('../../lib/api', () => ({ getJSON: vi.fn() }))
afterEach(() => { cleanup(); vi.resetAllMocks() })

function setup() {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  function Harness() {
    const [open, setOpen] = useState(true)
    return <><button onClick={() => setOpen(true)}>打开搜索</button><GlobalSearch open={open} onOpenChange={setOpen} /></>
  }
  render(<QueryClientProvider client={client}><MemoryRouter><Harness /></MemoryRouter></QueryClientProvider>)
  return client
}

it('does not query intermediate IME text and sends only the final keyword to all four owners', async () => {
  vi.mocked(getJSON).mockResolvedValue({ list: [], total: 0 })
  const client = setup()
  const input = screen.getByPlaceholderText('输入至少两个字符…')
  fireEvent.compositionStart(input)
  for (const value of ['zhong', '中文sou']) fireEvent.change(input, { target: { value } })
  expect(input).toHaveValue('中文sou')
  expect(getJSON).not.toHaveBeenCalled()
  fireEvent.compositionEnd(input, { target: { value: '中文搜索' } })
  fireEvent.change(input, { target: { value: '中文搜索' } })
  await screen.findByText('没有匹配结果')
  expect(getJSON).toHaveBeenCalledTimes(4)
  expect(vi.mocked(getJSON).mock.calls.map(([path]) => {
    const url = new URL(path, 'http://admin.test')
    expect(url.searchParams.get('keyword')).toBe('中文搜索')
    return url.pathname
  }).sort()).toEqual(['/api/v1/options/games', '/api/v1/options/site-groups', '/api/v1/options/sites', '/api/v1/options/tags'])
  client.clear()
})

it('keeps the two-character minimum and clears displayed/committed text when closed', async () => {
  vi.mocked(getJSON).mockResolvedValue({ list: [], total: 0 })
  const client = setup()
  const input = screen.getByPlaceholderText('输入至少两个字符…')
  fireEvent.change(input, { target: { value: 'a' } })
  expect(getJSON).not.toHaveBeenCalled()
  fireEvent.change(input, { target: { value: 'ab' } })
  await screen.findByText('没有匹配结果')
  expect(getJSON).toHaveBeenCalledTimes(4)
  fireEvent.compositionStart(input)
  fireEvent.change(input, { target: { value: 'zhongwen' } })
  fireEvent.click(screen.getByRole('button', { name: '关闭' }))
  await waitFor(() => expect(screen.queryByPlaceholderText('输入至少两个字符…')).not.toBeInTheDocument())
  vi.mocked(getJSON).mockClear()
  fireEvent.click(screen.getByRole('button', { name: '打开搜索' }))
  expect(screen.getByPlaceholderText('输入至少两个字符…')).toHaveValue('')
  expect(screen.getByText('输入名称、域名、Steam AppID 或技术 ID')).toBeInTheDocument()
  expect(getJSON).not.toHaveBeenCalled()
  fireEvent.change(screen.getByPlaceholderText('输入至少两个字符…'), { target: { value: 'cd' } })
  await screen.findByText('没有匹配结果')
  expect(getJSON).toHaveBeenCalledTimes(4)
  expect(client.getQueryCache().find({ queryKey: ['global-search', 'zhongwen'] })).toBeUndefined()
  client.clear()
})

it('resets search on an external dialog close too', async () => {
  vi.mocked(getJSON).mockResolvedValue({ list: [], total: 0 })
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  const view = (open: boolean) => <QueryClientProvider client={client}><MemoryRouter><GlobalSearch open={open} onOpenChange={vi.fn()} /></MemoryRouter></QueryClientProvider>
  const { rerender } = render(view(true))
  fireEvent.change(screen.getByPlaceholderText('输入至少两个字符…'), { target: { value: 'games' } })
  await screen.findByText('没有匹配结果')
  rerender(view(false))
  vi.mocked(getJSON).mockClear()
  rerender(view(true))
  expect(screen.getByPlaceholderText('输入至少两个字符…')).toHaveValue('')
  expect(screen.getByText('输入名称、域名、Steam AppID 或技术 ID')).toBeInTheDocument()
  expect(getJSON).not.toHaveBeenCalled()
  client.clear()
})
