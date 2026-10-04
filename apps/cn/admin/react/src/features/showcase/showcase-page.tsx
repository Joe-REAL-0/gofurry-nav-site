import { Plus } from '@phosphor-icons/react'
import { useQuery } from '@tanstack/react-query'
import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { DataTable, type AdminColumn } from '../../components/admin/data-table'
import { FilterField } from '../../components/admin/operations'
import { PageHeader, PageLayout } from '../../components/admin/page'
import { StatusBadge } from '../../components/admin/status'
import { WorkspaceTabs } from '../../components/admin/workspace'
import { Button } from '../../components/ui/button'
import { Select } from '../../components/ui/select'
import { useAuth } from '../auth/auth-context'
import { showcaseAPI } from './api'
import { ShowcaseCampaignCreate } from './showcase-campaign-create'
import { ShowcaseComposition } from './showcase-composition'
import { ShowcaseDiscovery } from './showcase-discovery'
import { CampaignStatus, contentLabels, LocaleSelect, statusLabels, useShowcaseURL } from './showcase-shared'
import { displayTime, operatingTimezone } from './showcase-time'
import type { Campaign } from './types'

const tabs = [{ key: 'composition', label: '当前编排' }, { key: 'campaigns', label: '活动' }, { key: 'discovery', label: '自动发现' }]
const columns: AdminColumn<Campaign>[] = [
  { key: 'internal_name', header: '内部名称', sortable: false, render: row => <Link className="inline-block max-w-64 truncate align-middle font-medium text-primary hover:underline" to={`/game/showcase/${row.id}`}>{row.internal_name}</Link> },
  { key: 'content_type', header: '类型', sortable: false, render: row => contentLabels[row.content_type] },
  { key: 'sponsored', header: '商业属性', sortable: false, render: row => <StatusBadge tone={row.sponsored ? 'info' : 'neutral'}>{row.sponsored ? '商业推广' : 'Editorial'}</StatusBadge> },
  { key: 'derived_status', header: '状态', sortable: false, render: row => <CampaignStatus value={row.derived_status} /> },
  { key: 'starts_at', header: '展示周期（上海时间）', sortable: false, render: row => <div className="py-2 text-xs"><p>{displayTime(row.starts_at)}</p><p>至 {displayTime(row.ends_at)}</p></div> },
  { key: 'weight', header: '选取权重', sortable: false }, { key: 'pin_position', header: '固定展示位置', sortable: false, render: row => row.pin_position ? `第 ${row.pin_position} 位` : '自动排序' }, { key: 'id', header: 'ID', sortable: false },
]
export function CampaignList() {
  const auth = useAuth(); const navigate = useNavigate(); const [creating, setCreating] = useState(false)
  const { params, update, page, size } = useShowcaseURL()
  const filters = new URLSearchParams({ page_num: String(page), page_size: String(size) })
  for (const key of ['keyword', 'state', 'derived_status', 'sponsored']) { const value = params.get(key); if (value) filters.set(key, value) }
  const query = useQuery({ queryKey: ['showcase', 'campaigns', filters.toString()], queryFn: () => showcaseAPI.campaigns(filters) })
  return <><DataTable columns={columns} data={query.data?.list ?? []} total={query.data?.total ?? 0} page={page} pageSize={size} search={params.get('keyword') ?? ''} onSearchChange={keyword => update({ keyword, page_num: '1' })} onPageChange={value => update({ page_num: String(value) })} onPageSizeChange={value => update({ page_num: '1', page_size: String(value) })} onRowClick={row => navigate(`/game/showcase/${row.id}`)} loading={query.isLoading} error={query.error?.message} onRetry={() => void query.refetch()} headerActions={auth.can('content.write') && <Button onClick={() => setCreating(true)}><Plus className="size-4" />新建活动</Button>} toolbar={<>
    <FilterField label="存储状态"><Select ariaLabel="存储状态" value={params.get('state') ?? ''} onValueChange={state => update({ state, page_num: '1' })} options={[{ value: '', label: '全部' }, { value: 'draft', label: '草稿' }, { value: 'published', label: '已发布' }, { value: 'paused', label: '已暂停' }, { value: 'archived', label: '已归档' }]} /></FilterField>
    <FilterField label="展示状态"><Select ariaLabel="展示状态" value={params.get('derived_status') ?? ''} onValueChange={derived_status => update({ derived_status, page_num: '1' })} options={[{ value: '', label: '全部' }, ...Object.entries(statusLabels).map(([value, label]) => ({ value, label }))]} /></FilterField>
    <FilterField label="商业属性"><Select ariaLabel="商业属性" value={params.get('sponsored') ?? ''} onValueChange={sponsored => update({ sponsored, page_num: '1' })} options={[{ value: '', label: '全部' }, { value: 'true', label: '商业推广' }, { value: 'false', label: 'Editorial' }]} /></FilterField>
  </>} />{creating && auth.can('content.write') && <ShowcaseCampaignCreate close={() => setCreating(false)} />}</>
}
export function ShowcasePage() {
  const { params, update, locale } = useShowcaseURL()
  const tab = tabs.some(item => item.key === params.get('tab')) ? params.get('tab')! : 'composition'
  return <PageLayout className="min-w-0"><PageHeader title="首页 Showcase" /><p className="text-xs text-muted-foreground">{operatingTimezone}</p>
    <div className="overflow-x-auto"><div className="min-w-max"><WorkspaceTabs tabs={tabs} active={tab} onChange={value => update({ tab: value === 'composition' ? null : value, page_num: null })} /></div></div>
    {tab !== 'campaigns' && <LocaleSelect value={locale} onChange={value => update({ locale: value, page_num: '1' })} />}
    {tab === 'composition' && <ShowcaseComposition locale={locale} />}{tab === 'campaigns' && <CampaignList />}{tab === 'discovery' && <ShowcaseDiscovery />}
  </PageLayout>
}
