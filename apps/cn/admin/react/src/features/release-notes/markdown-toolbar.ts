export type MarkdownAction = 'h2' | 'h3' | 'bold' | 'italic' | 'list' | 'quote' | 'link' | 'image' | 'code'

export function transformMarkdown(source: string, start: number, end: number, action: MarkdownAction) {
  const selected = source.slice(start, end)
  let from = start, to = end, inserted: string
  if (['h2', 'h3', 'list', 'quote'].includes(action)) {
    from = start === 0 ? 0 : source.lastIndexOf('\n', start - 1) + 1
    const lineEnd = end > start && source[end - 1] === '\n' ? end - 1 : end
    const nextLine = source.indexOf('\n', lineEnd)
    to = nextLine < 0 ? source.length : nextLine
    const prefix = { h2: '## ', h3: '### ', list: '- ', quote: '> ' }[action as 'h2' | 'h3' | 'list' | 'quote']
    inserted = source.slice(from, to).split('\n').map(line => prefix + line).join('\n')
  } else if (action === 'bold') inserted = `**${selected || '文字'}**`
  else if (action === 'italic') inserted = `*${selected || '文字'}*`
  else if (action === 'link') inserted = `[${selected || 'text'}](https://)`
  else if (action === 'image') inserted = `![${selected || 'alt'}](https://)`
  else if (selected.includes('\n')) {
    const runs = selected.match(/`+/g) ?? []
    const fence = '`'.repeat(Math.max(3, ...runs.map(run => run.length + 1)))
    inserted = `${start > 0 && source[start - 1] !== '\n' ? '\n' : ''}${fence}\n${selected}\n${fence}${end < source.length && source[end] !== '\n' ? '\n' : ''}`
  } else {
    const fence = '`'.repeat(Math.max(1, ...(selected.match(/`+/g) ?? []).map(run => run.length + 1)))
    const text = selected || 'code'
    inserted = `${fence}${text.includes('`') ? ' ' : ''}${text}${text.includes('`') ? ' ' : ''}${fence}`
  }
  return { value: source.slice(0, from) + inserted + source.slice(to), start: from, end: from + inserted.length }
}
