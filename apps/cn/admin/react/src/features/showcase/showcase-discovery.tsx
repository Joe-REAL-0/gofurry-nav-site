import { useQuery } from '@tanstack/react-query'
import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { DataTable, type AdminColumn } from '../../components/admin/data-table'
import { FilterField } from '../../components/admin/operations'
import { StatusBadge } from '../../components/admin/status'
import { Button } from '../../components/ui/button'
import { Dialog } from '../../components/ui/dialog'
import { Select } from '../../components/ui/select'
import { useAuth } from '../auth/auth-context'
import { showcaseAPI } from './api'
import { LocaleSelect, reasonLabels, useShowcaseURL } from './showcase-shared'
import type { Candidate, CandidateStatus, Pool, Release, TrendWindow } from './types'

export const exclusionLabels: Record<string, string> = {
  not_approved: '尚未允许自动展示', adult: '成人内容不参与自动展柜', locale_incomplete: '当前语言内容不完整',
  steam_artwork_unavailable: '缺少可用 Steam 素材', upcoming_requirements: '尚不符合即将发售条件',
  new_release_requirements: '尚不符合新作条件', trending_requirements: '热度条件未满足',
}
const failureLabels: Record<string, string> = {
  not_upcoming: '发售状态不是“即将发售”', first_available_missing: '缺少首次可玩日期', first_available_future: '首次可玩日期尚未到达', first_available_expired: '首次可玩已超过 30 天',
  player_facts_missing: '缺少已结算玩家数据', recent_days_insufficient: '近期有效观察不足 2 天', baseline_days_insufficient: '基线有效观察不足 4 天',
  recent_coverage_insufficient: '近期采样覆盖率不足 50%', baseline_coverage_insufficient: '基线采样覆盖率不足 50%',
  baseline_players_low: '基线平均人数不足 5', recent_players_low: '近期平均人数不足 20', player_delta_low: '平均人数增量不足 10', player_ratio_low: '平均人数增长不足 50%',
}
const statusLabels: Record<CandidateStatus, string> = { eligible: '可用候选', pending_approval: '待开启', blocked: '条件未满足', all: '全部诊断' }
const statusHelp: Record<CandidateStatus, string> = {
  eligible: '已允许自动展示，且当前类型的全部条件已满足；最终展示以“当前编排”为准。',
  pending_approval: '其他条件全部满足，只差在游戏“分类与展示”中允许自动展示。',
  blocked: '仍有内容、素材或类型条件未满足；允许自动展示不会跳过这些条件。',
  all: '包含全部游戏的诊断；默认先列出可用候选，再列出待开启和条件未满足的游戏。',
}
const poolSortLabels: Record<Pool, string> = { upcoming: '发售时间由近到远', new_release: '首次可玩由新到旧', trending: '热度选取权重由高到低' }
const statuses = Object.keys(statusLabels) as CandidateStatus[]

export function releaseText(release: Release | null | undefined) {
  if (!release) return '发售时间未知'
  const availability: Record<string, string> = { available: '已发售', upcoming: '即将发售', unknown: '发售状态未知' }
  const precision: Record<string, string> = { day: '精确日期', month: '月份', quarter: '季度', year: '年份', range: '日期范围', unknown: '日期未定', tba: '日期待定' }
  const date = release.exact_date || [release.window_start, release.window_end].filter(Boolean).join(' 至 ') || [release.year, release.quarter ? `第 ${release.quarter} 季度` : release.month ? `${release.month} 月` : null].filter(Boolean).join(' 年 ')
  return [availability[release.availability] ?? '发售状态未知', date, precision[release.precision] ?? '日期待定'].filter(Boolean).join(' · ')
}
function exclusions(row: Candidate) {
  return row.excluded_reasons.flatMap(reason => reason.endsWith('_requirements') && row.pool_failures?.length
    ? row.pool_failures.map(code => failureLabels[code] ?? code)
    : [exclusionLabels[reason] ?? reason])
}
function WindowEvidence({ title, value }: { title: string; value: TrendWindow }) {
  if (!value.observed_days) return <p>{title}：没有有效观察日，无法计算平均人数。</p>
  return <p>{title}：平均 {value.average.toLocaleString(undefined, { maximumFractionDigits: 1 })} 人 · {value.observed_days} 个有效观察日 · 覆盖率 {value.coverage === null ? '无覆盖率记录（历史数据）' : `${(value.coverage * 100).toFixed(1)}%`}</p>
}
function CandidateEvidence({ row, pool }: { row: Candidate; pool: Pool }) {
  if (pool === 'new_release') return <>首次可玩：{row.first_available ?? '未知'}</>
  if (pool === 'upcoming') return <>{releaseText(row.release)}</>
  if (!row.trending || (!row.trending.recent.observed_days && !row.trending.baseline.observed_days)) return <>暂无已结算玩家数据</>
  return <><p>{row.trending.eligible ? '热度条件已满足' : '热度条件未满足'}</p><p className="text-muted-foreground">近期 {row.trending.recent.average.toLocaleString(undefined, { maximumFractionDigits: 1 })} 人 / 基线 {row.trending.baseline.average.toLocaleString(undefined, { maximumFractionDigits: 1 })} 人</p></>
}
function CandidateDetails({ row, close }: { row: Candidate; close: () => void }) {
  return <Dialog open onOpenChange={open => { if (!open) close() }} title={`${row.name || `游戏 #${row.game_id}`} · 候选诊断`} description="这里解释候选资格，不代表首页展示排名。">
    <div className="grid gap-4 text-sm">
      <dl className="grid grid-cols-2 gap-2"><dt>自动展示许可</dt><dd>{row.showcase_eligible ? '已允许' : '未允许'}</dd><dt>内容分级</dt><dd>{row.sfw ? '非成人内容' : '成人内容'}</dd><dt>当前语言</dt><dd>{row.locale_ready ? '完整' : '不完整'}</dd><dt>Steam 素材</dt><dd>{row.artwork_ready ? '可用' : '不可用'}</dd></dl>
      <div><p className="mb-1 font-medium">阻碍原因</p>{exclusions(row).length ? <ul className="list-inside list-disc text-muted-foreground">{exclusions(row).map(text => <li key={text}>{text}</li>)}</ul> : <p>全部候选条件已满足。</p>}</div>
      <div><p>{releaseText(row.release)}</p><p>首次可玩：{row.first_available ?? '未知'}</p></div>
      {row.trending && <div className="grid gap-2 rounded-md border bg-surface-muted p-3 text-xs"><p className="font-medium">{row.trending.eligible ? '热度条件已满足' : '热度条件未满足'}</p><WindowEvidence title="近期 3 天" value={row.trending.recent} /><WindowEvidence title="此前 7 天" value={row.trending.baseline} /><p>增长动量 {row.trending.momentum.toLocaleString()} · 选取权重 {row.trending.weight}</p><p className="text-muted-foreground">使用已结算玩家数据；选取权重用于自动编排，不是公开评分。</p></div>}
    </div>
  </Dialog>
}

export function ShowcaseDiscovery() {
  const auth = useAuth()
  const { params, update, locale, page, size } = useShowcaseURL()
  const [detail, setDetail] = useState<Candidate | null>(null)
  const pool: Pool = params.get('pool') === 'new_release' ? 'new_release' : params.get('pool') === 'trending' ? 'trending' : 'upcoming'
  const status = statuses.includes(params.get('candidate_status') as CandidateStatus) ? params.get('candidate_status') as CandidateStatus : 'eligible'
  const keyword = params.get('candidate_keyword') ?? ''
  const [debouncedKeyword, setDebouncedKeyword] = useState(keyword)
  useEffect(() => { const timer = setTimeout(() => setDebouncedKeyword(keyword), 300); return () => clearTimeout(timer) }, [keyword])
  const sort = ['name', 'game_id'].includes(params.get('candidate_sort') ?? '') ? params.get('candidate_sort')! : 'pool'
  const reasons = ['not_approved', 'adult', 'locale_incomplete', 'steam_artwork_unavailable', `${pool}_requirements`]
  const excludedReason = reasons.includes(params.get('excluded_reason') ?? '') ? params.get('excluded_reason')! : ''
  const filters = new URLSearchParams({ pool, lang: locale, region: 'CN', page_num: String(page), page_size: String(size), status, sort })
  if (debouncedKeyword) filters.set('keyword', debouncedKeyword)
  if (excludedReason) filters.set('excluded_reason', excludedReason)
  const query = useQuery({ queryKey: ['showcase', 'candidates', filters.toString()], queryFn: () => showcaseAPI.candidates(filters), enabled: keyword === debouncedKeyword })
  const change = (patch: Record<string, string | null>) => update({ ...patch, page_num: '1' })
  const filtered = !!keyword || !!excludedReason
  const columns: AdminColumn<Candidate>[] = [
    { key: 'name', header: '游戏', sortable: false, render: row => <div className="max-w-56 py-3"><Link className="break-words font-medium text-primary hover:underline" to={`/game/games/${row.game_id}?tab=classification`}>{row.name || `游戏 #${row.game_id}`}</Link><p className="mt-1 text-xs text-muted-foreground">#{row.game_id}</p></div> },
    { key: 'status', header: '候选状态', sortable: false, render: row => <StatusBadge tone={row.status === 'eligible' ? 'success' : row.status === 'pending_approval' ? 'info' : 'neutral'}>{statusLabels[row.status]}</StatusBadge> },
    { key: 'evidence', header: '关键依据', sortable: false, render: row => <div className="max-w-64 whitespace-normal py-3 text-xs leading-relaxed"><CandidateEvidence row={row} pool={pool} /></div> },
    { key: 'excluded_reasons', header: '阻碍原因', sortable: false, render: row => { const reasons = exclusions(row); return <div className="max-w-56 whitespace-normal py-3 text-xs leading-relaxed">{reasons.length ? <>{reasons.slice(0, 2).map(text => <p key={text}>{text}</p>)}{reasons.length > 2 && <p className="text-muted-foreground">另有 {reasons.length - 2} 项，查看诊断</p>}</> : <span className="text-muted-foreground">无</span>}</div> } },
    { key: 'actions', header: '操作', sortable: false, render: row => <div className="flex flex-wrap items-center gap-x-3 gap-y-1 py-2"><Link className="whitespace-nowrap text-sm text-primary hover:underline" to={`/game/games/${row.game_id}?tab=classification`}>{!auth.can('content.write') ? '查看游戏' : row.status === 'pending_approval' ? '去开启' : row.status === 'blocked' ? '去检查' : '管理资格'}</Link><Button size="sm" variant="ghost" onClick={() => setDetail(row)} aria-label={`查看 ${row.name || `游戏 #${row.game_id}`} 的诊断`}>查看诊断</Button></div> },
  ]
  return <div className="grid gap-3">
    <div className="flex flex-wrap gap-2" role="group" aria-label="候选状态分组">{statuses.map(value => <Button key={value} variant={status === value ? 'primary' : 'secondary'} aria-pressed={status === value} onClick={() => change({ candidate_status: value })}>{statusLabels[value]}<span className="tabular-nums opacity-75">{query.data?.counts[value] ?? '—'}</span></Button>)}</div>
    <p className="text-xs text-muted-foreground">{statusHelp[status]}{filtered && ' 分组数量已应用搜索与排除原因筛选。'}</p>
    <div className="flex flex-wrap items-start justify-between gap-3 text-xs text-muted-foreground"><details open={excludedReason ? true : undefined}><summary className="cursor-pointer py-1 text-foreground">高级筛选{excludedReason ? ` · ${exclusionLabels[excludedReason]}` : ''}</summary><div className="mt-2"><FilterField label="排除原因"><Select ariaLabel="排除原因" value={excludedReason} onValueChange={value => change({ excluded_reason: value, candidate_status: value ? 'all' : status })} options={[{ value: '', label: '全部原因' }, ...reasons.map(value => ({ value, label: exclusionLabels[value] }))]} /></FilterField></div></details><p className="py-1">列表排序不代表首页展示顺序。</p></div>
    <DataTable columns={columns} data={query.data?.items ?? []} total={query.data?.total ?? 0} page={page} pageSize={size}
      search={keyword} searchClassName="max-w-64" searchPlaceholder="搜索游戏名称或 ID…" onSearchChange={value => change({ candidate_keyword: value })}
      onPageChange={value => update({ page_num: String(value) })} onPageSizeChange={value => change({ page_size: String(value) })}
      loading={query.isLoading || keyword !== debouncedKeyword} error={query.error?.message} onRetry={() => void query.refetch()}
      toolbar={<><LocaleSelect value={locale} onChange={value => change({ locale: value })} /><div className="w-40"><Select ariaLabel="自动发现类型" value={pool} onValueChange={value => change({ pool: value, excluded_reason: null })} options={(['upcoming', 'new_release', 'trending'] as const).map(value => ({ value, label: reasonLabels[value] }))} /></div><div className="w-56"><Select ariaLabel="候选列表排序" value={sort} onValueChange={value => change({ candidate_sort: value })} options={[{ value: 'pool', label: poolSortLabels[pool] }, { value: 'name', label: '游戏名称' }, { value: 'game_id', label: '游戏 ID 由小到大' }]} /></div></>}
      emptyState={<div className="grid min-h-48 place-items-center rounded-md border bg-surface p-6 text-center"><div><p className="font-medium">{status === 'eligible' && !filtered ? '暂无可用候选' : '当前条件下没有游戏'}</p><p className="mt-2 text-sm text-muted-foreground">{status === 'eligible' && !filtered ? '可以查看待开启游戏，或检查其他未满足的条件。' : '尝试调整类型、状态或搜索条件。'}</p><div className="mt-4 flex flex-wrap justify-center gap-2">{status === 'eligible' && <Button variant="secondary" onClick={() => change({ candidate_status: 'pending_approval' })}>查看待开启</Button>}{filtered ? <Button variant="ghost" onClick={() => change({ candidate_keyword: null, excluded_reason: null })}>清除筛选</Button> : <Button variant="ghost" onClick={() => change({ candidate_status: 'all' })}>查看全部诊断</Button>}</div></div></div>} />
    {detail && <CandidateDetails row={detail} close={() => setDetail(null)} />}
  </div>
}
