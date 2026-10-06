import { useEffect, useRef, useState, type ChangeEvent, type CompositionEvent } from 'react'

// Search owners receive committed text only. Keep IME drafts out of URL/query state.
export function useCompositionSafeSearch({ value, onCommit }: { value: string; onCommit: (value: string) => void }) {
  const [displayedValue, setDisplayedValue] = useState(value)
  const composing = useRef(false)
  const emitted = useRef(value)

  useEffect(() => {
    emitted.current = value
    if (!composing.current) setDisplayedValue(value)
  }, [value])

  const commit = (next: string) => {
    // Browsers may dispatch a final change after compositionEnd.
    if (emitted.current === next) return
    emitted.current = next
    onCommit(next)
  }
  const reset = (next: string) => {
    composing.current = false
    emitted.current = next
    setDisplayedValue(next)
  }

  return {
    composing,
    reset,
    inputProps: {
      value: displayedValue,
      onCompositionStart: () => { composing.current = true },
      onCompositionEnd: (event: CompositionEvent<HTMLInputElement>) => {
        composing.current = false
        setDisplayedValue(event.currentTarget.value)
        commit(event.currentTarget.value)
      },
      onChange: (event: ChangeEvent<HTMLInputElement>) => {
        setDisplayedValue(event.currentTarget.value)
        if (!composing.current) commit(event.currentTarget.value)
      },
    },
  }
}
