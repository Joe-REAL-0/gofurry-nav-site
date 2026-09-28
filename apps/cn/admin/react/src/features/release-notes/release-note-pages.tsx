import { Plus } from '@phosphor-icons/react'
import { useQuery } from '@tanstack/react-query'
import { Link, useNavigate, useParams, useSearchParams } from 'react-router-dom'
import { DataTable, type AdminColumn } from '../../components/admin/data-table'
import { PageHeader, PageLayout } from '../../components/admin/page'
import { ErrorState, LoadingState } from '../../components/admin/states'
import { Alert } from '../../components/ui/alert'
import { Button } from '../../components/ui/button'
import { ApiError, errorMessage, getJSON, listJSON } from '../../lib/api'
import { useAuth } from '../auth/auth-context'
import { ReleaseStatus } from './release-status'
import { ReleaseNoteEditor } from './release-note-editor'
import { releaseNoteKey, releaseNotesEndpoint, releaseNotesListKey, releaseWallTime, type ReleaseNote } from './release-note-model'

const columns: AdminColumn<ReleaseNote>[] = [
  { key: 'version', header: 'Version', render: note => note.version || '—' },
  { key: 'title', header: 'Title', render: note => <Link className="text-primary hover:underline" onClick={event => event.stopPropagation()} to={`/nav/update-notices/${note.id}`}>{note.title || note.title_en || '未命名公告'}</Link> },
  { key: 'publication_state', header: 'Status', render: note => <ReleaseStatus record={note} /> },
  { key: 'published_at', header: 'Published At', render: note => releaseWallTime(note.published_at) || '—' },
  { key: 'commit_sha', header: 'Commit', render: note => <span className="font-mono" title={note.commit_sha ?? undefined}>{note.commit_sha?.slice(0, 7) || '—'}</span> },
  { key: 'update_time', header: 'Last Updated', render: note => releaseWallTime(note.update_time) || '—' },
]

export function ReleaseNoteListPage() {
  const [params, setParams] = useSearchParams()
  const navigate = useNavigate()
  const canWrite = useAuth().can('content.write')
  const page = Math.max(1, Number(params.get('page')) || 1)
  const pageSize = [20, 50, 100].includes(Number(params.get('page_size'))) ? Number(params.get('page_size')) : 50
  const search = params.get('search') ?? ''
  const query = useQuery({ queryKey: [...releaseNotesListKey, page, pageSize, search], queryFn: () => listJSON<ReleaseNote>(releaseNotesEndpoint, page, pageSize, search) })
  const change = (key: string, value: string) => {
    const next = new URLSearchParams(params)
    if (value) next.set(key, value); else next.delete(key)
    if (key !== 'page') next.set('page', '1')
    setParams(next, { replace: true })
  }
  return <PageLayout>
    <PageHeader title="更新公告" actions={canWrite && <Button onClick={() => navigate('/nav/update-notices/new')}><Plus aria-hidden="true" className="size-4" />新建 Release Note</Button>} />
    <p className="text-xs text-muted-foreground">时间均为 Asia/Shanghai（UTC+8）。状态标记仅供管理参考，公开可见性由服务端决定。</p>
    <DataTable data={query.data?.list ?? []} columns={columns} total={query.data?.total ?? 0} page={page} pageSize={pageSize} search={search}
      onSearchChange={value => change('search', value)} onPageChange={value => change('page', String(value))} onPageSizeChange={value => change('page_size', String(value))}
      onRowClick={note => navigate(`/nav/update-notices/${note.id}`)} loading={query.isLoading} error={query.error?.message} onRetry={() => void query.refetch()} />
  </PageLayout>
}

export function ReleaseNoteEditorPage() {
  const { id } = useParams()
  const canWrite = useAuth().can('content.write')
  const validID = id === undefined || /^\d+$/.test(id) && Number(id) > 0
  const query = useQuery({ queryKey: releaseNoteKey(id ?? 'new'), queryFn: () => getJSON<ReleaseNote>(`${releaseNotesEndpoint}/${id}`), enabled: Boolean(id) && validID, refetchOnWindowFocus: false })
  if (!id && !canWrite) return <Alert tone="warning">当前账号只有只读权限。<Link to="/nav/update-notices">返回更新公告</Link></Alert>
  if (id && query.isLoading) return <LoadingState />
  if (!validID || query.error instanceof ApiError && query.error.status === 404) return <PageLayout><Alert tone="warning">未找到该公告，可能已被删除。</Alert><Link to="/nav/update-notices">返回更新公告</Link></PageLayout>
  if (id && !query.data) return <ErrorState message={errorMessage(query.error)} onRetry={() => void query.refetch()} />
  return <ReleaseNoteEditor key={id ?? 'new'} record={id ? query.data! : null} />
}
