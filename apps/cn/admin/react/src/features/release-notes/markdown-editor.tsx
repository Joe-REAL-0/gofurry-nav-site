import { useId, useMemo, useState } from 'react'
import { Button } from '../../components/ui/button'
import { Textarea } from '../../components/ui/input'
import { cn } from '../../lib/utils'
import { renderUpdateMarkdown } from './markdown'
import './release-notes.css'

export function MarkdownEditor({ value, onChange, label, readOnly, busy }: { value: string; onChange: (value: string) => void; label: string; readOnly: boolean; busy: boolean }) {
  const [panel, setPanel] = useState<'edit' | 'preview'>(readOnly ? 'preview' : 'edit')
  const id = useId()
  const html = useMemo(() => renderUpdateMarkdown(value), [value])
  return <div className="release-writing-workspace min-w-0 overflow-hidden rounded-md border">
    <div className="flex gap-1 border-b p-2 lg:hidden" aria-label="正文展示方式">
      <Button type="button" size="sm" variant={panel === 'edit' ? 'primary' : 'ghost'} aria-pressed={panel === 'edit'} onClick={() => setPanel('edit')}>编辑</Button>
      <Button type="button" size="sm" variant={panel === 'preview' ? 'primary' : 'ghost'} aria-pressed={panel === 'preview'} onClick={() => setPanel('preview')}>预览</Button>
    </div>
    <div className="release-writing-grid grid min-w-0 lg:grid-cols-2">
      <div className={cn('min-h-0 min-w-0 flex-col lg:flex lg:border-r', panel === 'edit' ? 'flex' : 'hidden')}>
        <label htmlFor={id} className="bg-surface-muted px-4 py-2 text-xs text-muted-foreground">{label}</label>
        <Textarea id={id} aria-label={label} className="min-h-0 flex-1 resize-none rounded-none border-0 bg-transparent p-4 font-mono leading-7 focus:ring-inset" value={value} readOnly={readOnly} disabled={busy} onChange={event => onChange(event.target.value)} />
      </div>
      <section aria-label={`${label}预览`} className={cn('min-h-0 min-w-0 flex-col lg:flex', panel === 'preview' ? 'flex' : 'hidden')}>
        <p className="bg-surface-muted px-4 py-2 text-xs text-muted-foreground">预览</p>
        <div className="min-h-0 flex-1 overflow-auto p-4">
          {value ? <div className="release-markdown" dangerouslySetInnerHTML={{ __html: html }} /> : <p className="text-sm text-muted-foreground">暂无正文</p>}
        </div>
      </section>
    </div>
  </div>
}
