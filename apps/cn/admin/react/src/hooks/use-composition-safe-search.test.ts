import { act, cleanup, renderHook } from '@testing-library/react'
import type { ChangeEvent, CompositionEvent } from 'react'
import { afterEach, expect, it, vi } from 'vitest'
import { useCompositionSafeSearch } from './use-composition-safe-search'

afterEach(cleanup)
const change = (value: string) => ({ currentTarget: { value } }) as ChangeEvent<HTMLInputElement>
const end = (value: string) => ({ currentTarget: { value } }) as CompositionEvent<HTMLInputElement>

it('deduplicates a trailing change even before the committed value is echoed by the owner', () => {
  const onCommit = vi.fn()
  const { result } = renderHook(() => useCompositionSafeSearch({ value: '', onCommit }))
  act(() => result.current.inputProps.onCompositionStart())
  act(() => result.current.inputProps.onChange(change('zhongwen')))
  expect(onCommit).not.toHaveBeenCalled()
  act(() => {
    result.current.inputProps.onCompositionEnd(end('中文'))
    result.current.inputProps.onChange(change('中文'))
  })
  expect(onCommit).toHaveBeenCalledExactlyOnceWith('中文')
  expect(result.current.inputProps.value).toBe('中文')
})

it('preserves an active draft across external updates, then synchronizes subsequent values', () => {
  const onCommit = vi.fn()
  const { result, rerender } = renderHook(({ value }) => useCompositionSafeSearch({ value, onCommit }), { initialProps: { value: 'initial' } })
  rerender({ value: 'external' })
  expect(result.current.inputProps.value).toBe('external')
  act(() => result.current.inputProps.onCompositionStart())
  act(() => result.current.inputProps.onChange(change('zhong')))
  rerender({ value: 'route changed' })
  expect(result.current.inputProps.value).toBe('zhong')
  expect(onCommit).not.toHaveBeenCalled()
  act(() => result.current.inputProps.onCompositionEnd(end('中文')))
  expect(onCommit).toHaveBeenCalledExactlyOnceWith('中文')
  rerender({ value: 'back' })
  expect(result.current.inputProps.value).toBe('back')
})
