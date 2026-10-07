import { useQuery } from '@tanstack/react-query'
import { useState } from 'react'
import { Link } from 'react-router-dom'
import { Detail, DetailGrid, Section } from '../../components/admin/page'
import { EmptyState, ErrorState, LoadingState } from '../../components/admin/states'
import { StatusBadge, TechnicalLabel } from '../../components/admin/status'
import { Button } from '../../components/ui/button'
import { showcaseAPI } from './api'
import { contentLabels, reasonLabels } from './showcase-shared'
import { chinaDate, dateBefore, displayTime } from './showcase-time'
import type { Action, Locale, QualityDay } from './types'

export function ActionView({ value }: { value?: Action }) {
  if (!value) return <>无</>
  if (value.game_id) return <Link className="text-primary hover:underline" to={`/game/games/${value.game_id}`}>{value.type} · 游戏 #{value.game_id}</Link>
  return <span className="break-all">{value.type} · {value.target || '—'}</span>
}
export const qualityLabels: Record<Exclude<keyof QualityDay, 'stat_date'>, string> = {
  invalid_token_events: 'Invalid Token', invalid_origin_events: 'Invalid Origin', filtered_user_agent_events: 'Filtered User Agent', duplicate_impressions: 'Duplicate Impressions', duplicate_clicks: 'Duplicate Clicks', session_rate_limited: 'Session Rate Limited', ip_rate_limited: 'IP Rate Limited', malformed_events: 'Malformed Events',
}
export function ShowcaseQuality() {
  const [open, setOpen] = useState(false)
  const to = chinaDate(); const from = dateBefore(to, 6)
  const query = useQuery({ queryKey: ['showcase', 'quality', from, to], queryFn: () => showcaseAPI.quality(from, to), enabled: open })
  return <details className="rounded-lg border bg-surface p-4" onToggle={event => setOpen(event.currentTarget.open)}><summary className="cursor-pointer text-sm font-medium">数据质量 · 近 7 天</summary><div className="mt-4">
    {query.isLoading ? <LoadingState /> : query.error ? <ErrorState message={query.error.message} onRetry={() => void query.refetch()} /> : !query.data?.daily.length ? <EmptyState title="暂无数据质量记录" /> : <DetailGrid>{Object.entries(qualityLabels).map(([key, label]) => <Detail key={key} label={label}>{query.data.daily.reduce((sum, day) => sum + day[key as keyof typeof qualityLabels], 0).toLocaleString()}</Detail>)}</DetailGrid>}
  </div></details>
}
export function ShowcaseComposition({ locale }: { locale: Locale }) {
  const query = useQuery({ queryKey: ['showcase', 'composition', locale], queryFn: () => showcaseAPI.composition(locale) })
  return <div className="grid min-w-0 gap-4"><Section title="当前编排" description="Composer 当前返回的运营视图 · Region CN" actions={<Button variant="secondary" disabled={query.isFetching} onClick={() => void query.refetch()}>刷新编排</Button>}>
    {query.isLoading ? <LoadingState /> : query.error ? <ErrorState message={query.error.message} onRetry={() => void query.refetch()} /> : <>
      {query.data && <div className="mb-4 grid gap-1 text-xs text-muted-foreground"><span>生成：{displayTime(query.data.generated_at)} · 有效至：{displayTime(query.data.valid_until)}（上海时间）</span><TechnicalLabel>Snapshot {query.data.snapshot_id}</TechnicalLabel></div>}
      {!query.data?.items.length ? <EmptyState title="当前没有可展示的 Showcase" /> : <div className="grid gap-4">{query.data.items.slice(0, 4).map(item => <article key={item.key} className="grid min-w-0 gap-3 rounded-md border p-4" data-showcase-item>
        <div className="flex flex-wrap items-center gap-2"><span className="font-mono text-sm text-muted-foreground">{String(item.position).padStart(2, '0')}</span><StatusBadge tone={item.sponsored ? 'info' : 'neutral'}>{reasonLabels[item.reason]}</StatusBadge><h3 className="break-words font-semibold">{item.title}</h3>{item.sponsored && <span className="text-sm text-muted-foreground">商业推广</span>}</div>
        <p className="break-words text-sm text-muted-foreground">{item.summary}</p>
        <DetailGrid><Detail label="Source">{item.source === 'managed' ? 'Managed · 人工策展' : 'Automatic · 自动发现'}</Detail><Detail label="内容类型">{contentLabels[item.content_type]}</Detail><Detail label="Position">#{item.position}</Detail>
          <Detail label="Campaign ID">{item.campaign_id ? <Link className="text-primary hover:underline" to={`/game/showcase/${item.campaign_id}`}>#{item.campaign_id}</Link> : '—'}</Detail><Detail label="Game ID">{item.game_id ? <Link className="text-primary hover:underline" to={`/game/games/${item.game_id}`}>#{item.game_id}</Link> : '—'}</Detail>
          <Detail label="Primary"><ActionView value={item.primary_action} /></Detail><Detail label="Secondary"><ActionView value={item.secondary_action} /></Detail>
        </DetailGrid>
        {item.artwork.kind === 'steam' ? <div><p className="mb-2 text-xs text-muted-foreground">Steam Artwork</p>{item.artwork.url && <img className="max-h-28 max-w-full rounded object-contain" src={item.artwork.url} alt={`${item.title} Steam 缩略图`} />}</div> : <div className="grid min-w-0 gap-1 rounded bg-surface-muted p-3 text-xs"><p className="font-medium">Managed Artwork · 已配置</p><code className="break-all">{item.artwork.desktop_object_key}</code>{item.artwork.mobile_object_key && <code className="break-all">{item.artwork.mobile_object_key}</code>}</div>}
      </article>)}</div>}
    </>}
  </Section><ShowcaseQuality /></div>
}
