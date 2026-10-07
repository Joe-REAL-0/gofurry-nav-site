import { useQuery } from '@tanstack/react-query'
import { LineChart } from 'echarts/charts'
import { GridComponent, LegendComponent, TooltipComponent } from 'echarts/components'
import * as echarts from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { useEffect, useRef } from 'react'
import { FilterField, Kpi, operationLineSeries } from '../../components/admin/operations'
import { Section } from '../../components/admin/page'
import { EmptyState, ErrorState, LoadingState } from '../../components/admin/states'
import { Alert } from '../../components/ui/alert'
import { DatePicker } from '../../components/ui/date-picker'
import { Select } from '../../components/ui/select'
import { showcaseAPI, statsCSV } from './api'
import { useShowcaseURL } from './showcase-shared'
import { operatingTimezone, statsRange, validDateRange } from './showcase-time'
import type { CampaignWorkspace, DailyStat } from './types'

echarts.use([LineChart, GridComponent, LegendComponent, TooltipComponent, CanvasRenderer])
export function ShowcaseTrend({ daily }: { daily: DailyStat[] }) {
  const ref = useRef<HTMLDivElement>(null)
  useEffect(() => {
    if (!ref.current || !daily.length) return
    const chart = echarts.init(ref.current)
    const draw = () => {
      const style = getComputedStyle(document.documentElement)
      const color = (name: string) => style.getPropertyValue(name).trim()
      chart.setOption({
        animation: false, textStyle: { color: color('--foreground') },
        legend: { top: 0, bottom: 'auto', textStyle: { color: color('--foreground') }, data: ['有效曝光', '有效点击'] },
        grid: { left: 16, right: 16, top: 42, bottom: 24, containLabel: true },
        xAxis: { type: 'category', data: daily.map(day => day.stat_date), axisLabel: { color: color('--muted-foreground') } },
        yAxis: { type: 'value', minInterval: 1, axisLabel: { color: color('--muted-foreground') }, splitLine: { lineStyle: { color: color('--border') } } },
        tooltip: { trigger: 'axis', renderMode: 'richText' },
        series: [
          { ...operationLineSeries(daily.map(day => ({ value: day.valid_impressions })), color('--primary')), name: '有效曝光' },
          { ...operationLineSeries(daily.map(day => ({ value: day.qualified_clicks })), color('--success')), name: '有效点击', symbol: 'rect' },
        ],
      })
    }
    draw()
    const resize = () => chart.resize()
    const observer = typeof ResizeObserver !== 'undefined' ? new ResizeObserver(resize) : null
    observer?.observe(ref.current); window.addEventListener('resize', resize)
    const theme = new MutationObserver(draw); theme.observe(document.documentElement, { attributes: true, attributeFilter: ['class', 'style', 'data-theme'] })
    return () => { observer?.disconnect(); theme.disconnect(); window.removeEventListener('resize', resize); chart.dispose() }
  }, [daily])
  return daily.length ? <div ref={ref} role="img" aria-label="有效曝光与有效点击每日趋势" className="h-64 min-w-0 w-full" /> : <EmptyState title="暂无统计趋势" />
}
export function ShowcaseStats({ workspace }: { workspace: CampaignWorkspace }) {
  const { params, update } = useShowcaseURL()
  const defaults = statsRange('30', workspace.starts_at)
  const from = params.get('from') ?? defaults.from; const to = params.get('to') ?? defaults.to
  const valid = validDateRange(from, to)
  const range = params.get('range') || (params.has('from') || params.has('to') ? 'custom' : '30')
  const query = useQuery({ queryKey: ['showcase', 'stats', workspace.id, from, to], queryFn: () => showcaseAPI.stats(workspace.id, from, to), enabled: valid })
  return <div className="grid gap-4"><Section title="统计" description={operatingTimezone} actions={valid && <a className="text-sm font-medium text-primary hover:underline" href={statsCSV(workspace.id, from, to)}>导出 CSV</a>}>
    <div className="mb-4 flex flex-wrap items-end gap-3"><FilterField label="统计范围"><Select ariaLabel="统计范围" value={range} onValueChange={value => {
      if (value === 'custom') update({ range: value, from, to })
      else update({ range: value, ...statsRange(value as '7' | '30' | 'all', workspace.starts_at) })
    }} options={[{ value: '7', label: '7日' }, { value: '30', label: '30日' }, { value: 'all', label: '全部（最多 366 天）' }, { value: 'custom', label: '自定义' }]} /></FilterField>
      <FilterField label="开始日期"><DatePicker ariaLabel="统计开始日期" value={from} onValueChange={value => update({ from: value || 'invalid', range: 'custom' })} /></FilterField>
      <FilterField label="结束日期"><DatePicker ariaLabel="统计结束日期" value={to} onValueChange={value => update({ to: value || 'invalid', range: 'custom' })} /></FilterField>
    </div>
    {!valid ? <Alert tone="danger">日期范围须有效、起止有序且最多 366 天。</Alert> : query.isLoading ? <LoadingState /> : query.error ? <ErrorState message={query.error.message} onRetry={() => void query.refetch()} /> : query.data && <>
      <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-4"><Kpi label="有效曝光" value={query.data.totals.valid_impressions.toLocaleString()} /><Kpi label="有效点击" value={query.data.totals.qualified_clicks.toLocaleString()} /><Kpi label="CTR" value={`${(query.data.totals.ctr * 100).toFixed(2)}%`} /><Kpi label="独立会话估算" value={query.data.totals.session_estimate.toLocaleString()} /></div>
      <p className="mt-3 text-xs text-muted-foreground">会话数为估算。多日总计为每日 HLL 估算之和，同一浏览会话可能跨日重复计入。</p>
    </>}
  </Section>
    {valid && query.data && <><Section title="每日趋势"><ShowcaseTrend daily={query.data.daily} /></Section><Section title="点击来源"><div className="grid gap-3 sm:grid-cols-2">{(['artwork', 'title', 'primary', 'secondary'] as const).map((source, index) => {
      const count = query.data.totals[`click_${source}`]
      const ratio = query.data.totals.qualified_clicks ? count / query.data.totals.qualified_clicks : 0
      return <div key={source}><p className="mb-1 flex justify-between text-sm"><span>{['Artwork', 'Title', 'Primary CTA', 'Secondary CTA'][index]}</span><span>{count.toLocaleString()} · {(ratio * 100).toFixed(1)}%</span></p><div className="h-1.5 overflow-hidden rounded bg-surface-muted"><div className="h-full bg-primary" style={{ width: `${Math.min(100, ratio * 100)}%` }} /></div></div>
    })}</div></Section></>}
  </div>
}
