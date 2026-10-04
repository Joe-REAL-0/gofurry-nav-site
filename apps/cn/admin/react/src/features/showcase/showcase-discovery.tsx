import { useQuery } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import { DataTable, type AdminColumn } from '../../components/admin/data-table'
import { StatusBadge } from '../../components/admin/status'
import { Select } from '../../components/ui/select'
import { showcaseAPI } from './api'
import { LocaleSelect, reasonLabels, useShowcaseURL } from './showcase-shared'
import type { Candidate, Pool, Release, TrendWindow } from './types'

export const exclusionLabels: Record<string, string> = { not_approved: '未允许自动 Showcase', adult: 'Adult 内容', locale_incomplete: '当前语言内容不完整', steam_artwork_unavailable: 'Steam Artwork 不可用', upcoming_requirements: '不满足 Upcoming 条件', new_release_requirements: '不满足 New Release 条件', trending_requirements: '不满足 Trending 条件' }
export function releaseText(release: Release | null | undefined) {
  if (!release) return '暂无规范化发售证据'
  return [release.availability, release.precision, release.exact_date || [release.window_start, release.window_end].filter(Boolean).join(' ~ ') || [release.year, release.quarter ? `Q${release.quarter}` : release.month].filter(Boolean).join('-')].filter(Boolean).join(' · ')
}
function WindowEvidence({ title, value }: { title: string; value: TrendWindow }) {
  return <p>{title}：均值 {value.average.toLocaleString()} · {value.observed_days} 观察日 · 覆盖率 {value.coverage === null ? '—（legacy observed）' : `${(value.coverage * 100).toFixed(1)}%`}</p>
}
const columns: AdminColumn<Candidate>[] = [
  { key: 'name', header: 'Game', sortable: false, render: row => <div className="max-w-64 py-2"><p className="break-words font-medium">{row.name || `游戏 #${row.game_id}`}</p><Link className="text-primary hover:underline" to={`/game/games/${row.game_id}?tab=classification`}>打开游戏 · #{row.game_id}</Link></div> },
  { key: 'candidate', header: '候选诊断', sortable: false, render: row => <div className="grid gap-1 py-2 text-xs"><StatusBadge tone={row.candidate ? 'success' : 'neutral'}>{row.candidate ? '符合候选条件' : '未进入候选'}</StatusBadge><span>自动 Showcase：{row.showcase_eligible ? '已允许' : '未允许'}</span><span>SFW：{row.sfw ? '是' : '否'} · 语言：{row.locale_ready ? '完整' : '不完整'} · 素材：{row.artwork_ready ? '可用' : '不可用'}</span></div> },
  { key: 'excluded_reasons', header: '排除原因 / 发售证据', sortable: false, render: row => <div className="max-w-80 whitespace-normal py-2 text-xs">{row.excluded_reasons.map(reason => <p key={reason}>{exclusionLabels[reason] ?? reason}</p>)}<p className="mt-1 text-muted-foreground">{releaseText(row.release)}</p></div> },
  { key: 'trending', header: 'Trending 内部证据', sortable: false, render: row => row.trending ? <div className="min-w-60 py-2 text-xs"><WindowEvidence title="Recent" value={row.trending.recent} /><WindowEvidence title="Baseline" value={row.trending.baseline} /><p>Momentum {row.trending.momentum.toLocaleString()} · Weight {row.trending.weight} · Eligible {row.trending.eligible ? '是' : '否'}</p><p className="text-muted-foreground">内部选取诊断，不是公开评分。</p></div> : '—' },
]
export function ShowcaseDiscovery() {
  const { params, update, locale, page, size } = useShowcaseURL()
  const pool: Pool = params.get('pool') === 'new_release' ? 'new_release' : params.get('pool') === 'trending' ? 'trending' : 'upcoming'
  const query = useQuery({ queryKey: ['showcase', 'candidates', pool, locale, page, size], queryFn: () => showcaseAPI.candidates(pool, locale, page, size) })
  return <DataTable columns={columns} data={query.data?.items ?? []} total={query.data?.total ?? 0} page={page} pageSize={size} search="" searchable={false} onSearchChange={() => undefined} onPageChange={value => update({ page_num: String(value) })} onPageSizeChange={value => update({ page_num: '1', page_size: String(value) })} loading={query.isLoading} error={query.error?.message} onRetry={() => void query.refetch()} toolbar={<>
    <LocaleSelect value={locale} onChange={value => update({ locale: value, page_num: '1' })} />
    <div className="w-64 max-w-full"><Select ariaLabel="自动发现 Pool" value={pool} onValueChange={value => update({ pool: value, page_num: '1' })} options={(['upcoming', 'new_release', 'trending'] as const).map(value => ({ value, label: reasonLabels[value] }))} /></div>
  </>} />
}
