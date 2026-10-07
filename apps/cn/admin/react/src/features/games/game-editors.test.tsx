import { act, cleanup, createEvent, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { useState } from 'react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { RemoteSelect } from '../../components/admin/operations'
import { listJSON } from '../../lib/api'
import { KeyValueEditor, TagMultiSelect, loadAllTagOptions, summarySchema, SummaryField } from './game-editors'

vi.mock('../../lib/api', () => ({ listJSON: vi.fn() }))
afterEach(() => { cleanup(); vi.useRealTimers(); vi.resetAllMocks() })

describe('game tag searches', () => {
  it.each([0, 300])('keeps remote IME drafts local with debounceMs=%i and queries the final text once', async (debounceMs) => {
    vi.mocked(listJSON).mockResolvedValue({ total: 1, list: [{ id: '1', label: '冒险' }] })
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } })
    const onChange = vi.fn()
    render(<QueryClientProvider client={client}><RemoteSelect endpoint="/api/v1/options/tags" pageSize={10} debounceMs={debounceMs} value={{ id: '99', label: '已选标签' }} onChange={onChange} /></QueryClientProvider>)
    const input = screen.getByRole('combobox')
    act(() => input.focus())
    await screen.findByRole('option', { name: '冒险' })
    vi.useFakeTimers()
    vi.mocked(listJSON).mockClear()
    fireEvent.compositionStart(input)
    const active = input.getAttribute('aria-activedescendant')
    for (const value of ['mao', '冒xian']) {
      fireEvent.change(input, { target: { value } })
      await act(async () => { await vi.advanceTimersByTimeAsync(500) })
      expect(input).toHaveValue(value)
      expect(listJSON).not.toHaveBeenCalled()
      expect(input).toHaveAttribute('aria-activedescendant', active)
    }
    expect(onChange).not.toHaveBeenCalled()
    fireEvent.compositionEnd(input, { target: { value: '冒险' } })
    fireEvent.change(input, { target: { value: '冒险' } })
    if (debounceMs) {
      await act(async () => { await vi.advanceTimersByTimeAsync(299) })
      expect(listJSON).not.toHaveBeenCalled()
    }
    await act(async () => { await vi.advanceTimersByTimeAsync(debounceMs ? 1 : 0) })
    expect(listJSON).toHaveBeenCalledExactlyOnceWith('/api/v1/options/tags', 1, 10, '冒险')
    fireEvent.keyDown(input, { key: 'Escape' })
    expect(input).toHaveValue('已选标签')
    expect(onChange).not.toHaveBeenCalled()
    client.clear()
  })

  it.each(['ref', 'native'])('does not consume IME confirmation/navigation keys (%s), then restores Enter selection', async (guard) => {
    const option = { id: '1', label: '冒险' }
    vi.mocked(listJSON).mockResolvedValue({ total: 2, list: [option, { id: '2', label: '解谜' }] })
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } })
    const onChange = vi.fn()
    render(<QueryClientProvider client={client}><RemoteSelect endpoint="/api/v1/options/tags" value={null} onChange={onChange} /></QueryClientProvider>)
    const input = screen.getByRole('combobox')
    act(() => input.focus())
    await screen.findByRole('option', { name: '冒险' })
    const active = input.getAttribute('aria-activedescendant')
    if (guard === 'ref') {
      fireEvent.compositionStart(input)
      fireEvent.change(input, { target: { value: 'maoxian' } })
    }
    for (const key of ['ArrowDown', 'ArrowUp', 'Enter', 'Escape']) {
      const event = createEvent.keyDown(input, { key, isComposing: guard === 'native', cancelable: true })
      fireEvent(input, event)
      expect(event.defaultPrevented).toBe(false)
      expect(input).toHaveAttribute('aria-activedescendant', active)
      expect(input).toHaveAttribute('aria-expanded', 'true')
    }
    expect(onChange).not.toHaveBeenCalled()
    if (guard === 'ref') fireEvent.compositionEnd(input, { target: { value: '冒险' } })
    await screen.findByRole('option', { name: '冒险' })
    fireEvent.keyDown(input, { key: 'Enter' })
    expect(onChange).toHaveBeenCalledExactlyOnceWith(option)
    client.clear()
  })

  it('preserves an IME draft through blur without committing or selecting it', async () => {
    vi.mocked(listJSON).mockResolvedValue({ total: 1, list: [{ id: '1', label: '冒险' }] })
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } })
    const onChange = vi.fn()
    render(<QueryClientProvider client={client}><RemoteSelect endpoint="/api/v1/options/tags" debounceMs={300} value={{ id: '99', label: '已选标签' }} onChange={onChange} /></QueryClientProvider>)
    const input = screen.getByRole('combobox')
    act(() => input.focus())
    await screen.findByRole('option', { name: '冒险' })
    vi.useFakeTimers()
    vi.mocked(listJSON).mockClear()
    fireEvent.compositionStart(input)
    fireEvent.change(input, { target: { value: 'maoxian' } })
    act(() => input.blur())
    await act(async () => { await vi.advanceTimersByTimeAsync(500) })
    expect(input).toHaveValue('maoxian')
    expect(listJSON).not.toHaveBeenCalled()
    expect(onChange).not.toHaveBeenCalled()
    fireEvent.compositionEnd(input, { target: { value: '冒险' } })
    await act(async () => { await vi.advanceTimersByTimeAsync(300) })
    expect(listJSON).toHaveBeenCalledExactlyOnceWith('/api/v1/options/tags', 1, 50, '冒险')
    expect(input).toHaveValue('已选标签')
    expect(input).toHaveAttribute('aria-expanded', 'false')
    expect(onChange).not.toHaveBeenCalled()
    client.clear()
  })

  it('requests ten results and debounces remote searches while retaining the selection', async () => {
    vi.mocked(listJSON).mockResolvedValue({ total: 1, list: [{ id: '1', label: '冒险' }] })
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } })
    render(<QueryClientProvider client={client}><RemoteSelect endpoint="/api/v1/options/tags" pageSize={10} debounceMs={300} value={{ id: '99', label: '已选标签' }} onChange={vi.fn()} /></QueryClientProvider>)
    await waitFor(() => expect(listJSON).toHaveBeenCalledWith('/api/v1/options/tags', 1, 10, ''))
    const input = screen.getByRole('combobox')
    expect(input).toHaveValue('已选标签')
    vi.useFakeTimers()
    fireEvent.focus(input)
    fireEvent.change(input, { target: { value: '冒' } })
    await act(async () => { vi.advanceTimersByTime(200) })
    fireEvent.change(input, { target: { value: '冒险' } })
    await act(async () => { vi.advanceTimersByTime(299) })
    expect(listJSON).toHaveBeenCalledTimes(1)
    await act(async () => { vi.advanceTimersByTime(1) })
    expect(listJSON).toHaveBeenLastCalledWith('/api/v1/options/tags', 1, 10, '冒险')
    fireEvent.keyDown(input, { key: 'Escape' })
    expect(input).toHaveValue('已选标签')
    client.clear()
  })

  it('filters locally without losing hidden selections or making requests', () => {
    function Harness() {
      const [selected, setSelected] = useState(['1'])
      return <TagMultiSelect options={[{ id: '1', label: '冒险', extra: 'Adventure' }, { id: '2', label: '解谜' }]} selected={selected} onChange={setSelected} />
    }
    render(<Harness />)
    fireEvent.change(screen.getByLabelText('搜索全部标签'), { target: { value: '解谜' } })
    fireEvent.click(screen.getByRole('checkbox', { name: '解谜' }))
    fireEvent.change(screen.getByLabelText('搜索全部标签'), { target: { value: '' } })
    expect(screen.getByRole('checkbox', { name: '冒险' })).toHaveAttribute('aria-checked', 'true')
    expect(screen.getByRole('checkbox', { name: '解谜' })).toHaveAttribute('aria-checked', 'true')
    fireEvent.click(screen.getByRole('checkbox', { name: '冒险' }))
    expect(screen.getByRole('checkbox', { name: '冒险' })).toHaveAttribute('aria-checked', 'false')
    expect(listJSON).not.toHaveBeenCalled()
  })

  it('loads all pages once for local filtering, including tags beyond the first 200', async () => {
    vi.mocked(listJSON).mockResolvedValueOnce({ total: 201, list: Array.from({ length: 200 }, (_, i) => ({ id: String(i), label: String(i) })) }).mockResolvedValueOnce({ total: 201, list: [{ id: '201', label: '最后一项' }] })
    expect(await loadAllTagOptions()).toHaveLength(201)
    expect(listJSON).toHaveBeenLastCalledWith('/api/v1/options/tags', 2, 200)
  })
})

describe('game summaries', () => {
  it.each(['中', 'a', '🐺'])('accepts 400 and rejects 401 Unicode characters (%s)', (char) => {
    expect(summarySchema.safeParse(char.repeat(400)).success).toBe(true)
    expect(summarySchema.safeParse(char.repeat(401)).success).toBe(false)
  })
  it.each(['中文简介', '英文简介'])('shows the count and validation for %s, including prefilled text', (label) => {
    const { rerender } = render(<SummaryField label={label} value={'🐺'.repeat(400)} onChange={vi.fn()} />)
    expect(screen.getByText('已用 400/400 字符 · 剩余 0 字符')).toBeInTheDocument()
    rerender(<SummaryField label={label} value={'a'.repeat(401)} onChange={vi.fn()} />)
    expect(screen.getByText('简介最多 400 个字符')).toBeInTheDocument()
  })
})

it('uses platform choices, preserves legacy spelling and excludes already selected keys', async () => {
  render(<KeyValueEditor platformKeys value={[{ key: ' QQ ', value: 'https://example.test' }, { key: '', value: '' }]} onChange={vi.fn()} />)
  expect(screen.getByRole('combobox', { name: '平台 1' })).toHaveTextContent('QQ')
  fireEvent.click(screen.getByRole('combobox', { name: '平台 2' }))
  expect(await screen.findByRole('option', { name: 'discord' })).toBeInTheDocument()
  expect(screen.queryByRole('option', { name: 'qq' })).not.toBeInTheDocument()
})

it('keeps resources as free text', () => {
  render(<KeyValueEditor value={[{ key: '自定义下载', value: '' }]} onChange={vi.fn()} />)
  expect(screen.getByPlaceholderText('键')).toHaveValue('自定义下载')
  expect(screen.queryByRole('combobox')).not.toBeInTheDocument()
})
