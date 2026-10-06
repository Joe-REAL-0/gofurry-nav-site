import { useQuery } from '@tanstack/react-query'
import { Link, useSearchParams } from 'react-router-dom'
import { DataTable, type AdminColumn } from '../../components/admin/data-table'
import { PageHeader, PageLayout } from '../../components/admin/page'
import { StatusBadge, TechnicalLabel } from '../../components/admin/status'
import { Button } from '../../components/ui/button'
import { Input } from '../../components/ui/input'
import { Select } from '../../components/ui/select'
import { useCompositionSafeSearch } from '../../hooks/use-composition-safe-search'
import { errorMessage, listJSON } from '../../lib/api'
import { formatDate } from '../../lib/utils'
import { useAuth } from '../auth/auth-context'
import { collectionEndpoint } from './api'
import { collectionStatusLabels, type GameCollection } from './types'

const columns: AdminColumn<GameCollection>[] = ([
  { key: 'name', header: '名称', render: c => <Link className="text-primary" to={`/game/collections/${c.id}`}>{c.name}</Link> },
  { key: 'code', header: 'Code', render: c => <TechnicalLabel>{c.code}</TechnicalLabel> },
  { key: 'status', header: '状态', render: c => <StatusBadge>{collectionStatusLabels[c.status]}</StatusBadge> },
  { key: 'member_count', header: '收录游戏' }, { key: 'sfw_member_count', header: 'SFW可见' },
  { key: 'home_slot', header: '首页', render: c => c.home_slot ? `第 ${c.home_slot} 位` : '—' },
  { key: 'updated_at', header: '更新时间', render: c => formatDate(c.updated_at) },
] satisfies AdminColumn<GameCollection>[]).map(column => ({ ...column, sortable: false }))

export function CollectionListPage() {
  const [params, setParams] = useSearchParams()
  const auth = useAuth()
  const page = Math.max(1, Number(params.get('page_num')) || 1)
  const keyword = params.get('keyword') ?? ''
  const status = params.get('status') ?? ''
  const set = (key: string, value: string) => {
    const next = new URLSearchParams(params); next.set(key, value)
    if (key !== 'page_num') next.set('page_num', '1')
    setParams(next, { replace: true })
  }
  const search = useCompositionSafeSearch({ value: keyword, onCommit: value => set('keyword', value) })
  const query = useQuery({ queryKey: ['game-collections', page, keyword, status], queryFn: () => listJSON<GameCollection>(`${collectionEndpoint}?status=${encodeURIComponent(status)}`, page, 50, keyword) })
  return <PageLayout>
    <PageHeader title="游戏分区" actions={<><Link to="/game/collections/home-curation"><Button variant="secondary">首页入口编排</Button></Link>{auth.can('content.write') && <Link to="/game/collections/new"><Button>新建游戏分区</Button></Link>}</>} />
    <DataTable data={query.data?.list ?? []} columns={columns} total={query.data?.total ?? 0} page={page} pageSize={50} search={keyword} searchable={false} onSearchChange={value => set('keyword', value)} onPageChange={value => set('page_num', String(value))} loading={query.isLoading} error={query.error ? errorMessage(query.error) : undefined} onRetry={() => void query.refetch()} toolbar={<>
      <Input aria-label="搜索游戏分区" placeholder="搜索名称或 Code…" className="max-w-sm" {...search.inputProps} />
      <Select ariaLabel="分区状态" value={status} onValueChange={value => set('status', value)} options={[{ value: '', label: '全部状态' }, ...Object.entries(collectionStatusLabels).map(([value, label]) => ({ value, label }))]} />
    </>} />
  </PageLayout>
}
