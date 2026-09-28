import { useLayoutEffect, useMemo, useRef, useState } from 'react'
import { Button } from '../../components/ui/button'
import { Textarea } from '../../components/ui/input'
import { cn } from '../../lib/utils'
import { renderUpdateMarkdown } from './markdown'
import { transformMarkdown, type MarkdownAction } from './markdown-toolbar'
import './release-notes.css'

const actions: [MarkdownAction, string][] = [['h2', 'H2'], ['h3', 'H3'], ['bold', 'Bold'], ['italic', 'Italic'], ['list', 'List'], ['quote', 'Quote'], ['link', 'Link'], ['image', 'Image'], ['code', 'Code']]

export function MarkdownEditor({ value, onChange, label, readOnly, busy }: { value: string; onChange: (value: string) => void; label: string; readOnly: boolean; busy: boolean }) {
  const [panel, setPanel] = useState<'edit' | 'preview'>(readOnly ? 'preview' : 'edit')
  const area = useRef<HTMLTextAreaElement>(null)
  const selection = useRef<{ start: number; end: number } | null>(null)
  const html = useMemo(() => renderUpdateMarkdown(value), [value])
  useLayoutEffect(() => {
    if (selection.current && area.current) {
      area.current.focus()
      area.current.setSelectionRange(selection.current.start, selection.current.end)
      selection.current = null
    }
  }, [value])
  const apply = (action: MarkdownAction) => {
    if (!area.current) return
    const next = transformMarkdown(value, area.current.selectionStart, area.current.selectionEnd, action)
    selection.current = next
    onChange(next.value)
  }
  return <div className="grid min-w-0 gap-3">
    <div className="flex gap-2 lg:hidden" aria-label="正文展示方式">
      <Button type="button" variant={panel === 'edit' ? 'primary' : 'ghost'} aria-pressed={panel === 'edit'} onClick={() => setPanel('edit')}>编辑</Button>
      <Button type="button" variant={panel === 'preview' ? 'primary' : 'ghost'} aria-pressed={panel === 'preview'} onClick={() => setPanel('preview')}>预览</Button>
    </div>
    <div className="grid min-w-0 items-start gap-4 lg:grid-cols-2">
      <div className={cn('min-w-0 space-y-2 lg:block', panel !== 'edit' && 'hidden')}>
        {!readOnly && <div className="flex flex-wrap gap-1" aria-label="Markdown 工具栏">{actions.map(([action, text]) => <Button key={action} type="button" size="sm" variant="secondary" disabled={busy} onMouseDown={event => event.preventDefault()} onClick={() => apply(action)}>{text}</Button>)}</div>}
        <Textarea aria-label={label} className="min-h-80 font-mono" value={value} readOnly={readOnly} disabled={busy} ref={area} onChange={event => onChange(event.target.value)} />
      </div>
      <section aria-label={`${label}预览`} className={cn('min-w-0 lg:block', panel !== 'preview' && 'hidden')}>
        <p className="mb-2 text-xs text-muted-foreground">Markdown 预览</p>
        {value ? <div className="release-markdown" dangerouslySetInnerHTML={{ __html: html }} /> : <p className="text-sm text-muted-foreground">暂无正文</p>}
      </section>
    </div>
  </div>
}
