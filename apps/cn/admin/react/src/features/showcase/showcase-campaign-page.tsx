import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { useToast } from '../../app/toast'
import { Detail, DetailGrid, PageHeader, PageLayout, Section } from '../../components/admin/page'
import { ErrorState, LoadingState } from '../../components/admin/states'
import { TechnicalLabel } from '../../components/admin/status'
import { WorkspaceTabs } from '../../components/admin/workspace'
import { Alert } from '../../components/ui/alert'
import { Button } from '../../components/ui/button'
import { ConfirmAction } from '../../components/ui/dialog'
import { useAuth } from '../auth/auth-context'
import { invalidateShowcase, showcaseAPI } from './api'
import { ShowcaseAssets } from './showcase-assets'
import { ShowcaseContent } from './showcase-content'
import { ShowcaseSchedule } from './showcase-schedule'
import { CampaignStatus, contentLabels, useShowcaseURL } from './showcase-shared'
import { ShowcaseStats } from './showcase-stats'
import { displayTime, operatingTimezone } from './showcase-time'
import type { CampaignWorkspace, Lifecycle } from './types'

const tabs = [{ key: 'overview', label: '概览' }, { key: 'content', label: '内容' }, { key: 'assets', label: '素材' }, { key: 'schedule', label: '排期与展示' }, { key: 'stats', label: '统计' }, { key: 'history', label: '历史' }]
export const lifecycleLabels: Record<Lifecycle, string> = { publish: '发布', pause: '暂停', resume: '恢复', archive: '归档', delete: '删除草稿' }
function Overview({ workspace: w }: { workspace: CampaignWorkspace }) {
  return <Section title="活动概览" description={operatingTimezone}><div className="grid gap-5"><DetailGrid>
    <Detail label="状态"><CampaignStatus value={w.derived_status} /></Detail><Detail label="内容类型">{contentLabels[w.content_type]}</Detail><Detail label="商业属性">{w.sponsored ? 'Sponsored · 商业推广' : 'Editorial'}</Detail>
    <Detail label="启用语言">{w.locales.filter(l => l.enabled).map(l => l.lang === 'zh' ? '中文' : 'English').join(' / ') || '无'}</Detail>
    <Detail label="关联游戏">{w.linked_game_id ? <Link to={`/game/games/${w.linked_game_id}`} className="text-primary hover:underline">游戏 #{w.linked_game_id}</Link> : '未关联'}</Detail>
    <Detail label="开始时间">{displayTime(w.starts_at)}</Detail><Detail label="结束时间">{displayTime(w.ends_at)}</Detail><Detail label="Weight">{w.weight}</Detail><Detail label="Pin Position">{w.pin_position ? `#${w.pin_position}` : '不固定'}</Detail>
    <Detail label="Desktop Artwork">{w.desktop_object_key ? '已配置' : '未配置'}</Detail><Detail label="Mobile Artwork">{w.mobile_object_key ? '已配置' : '未配置（可选）'}</Detail><Detail label="Ready To Publish">{w.ready_to_publish ? '已满足发布条件' : '需要完成发布条件'}</Detail>
  </DetailGrid>{!w.ready_to_publish && <Alert tone="info"><p className="font-medium">后端发布诊断</p><ul className="mt-2 list-inside list-disc">{w.publication_diagnostics.map((message, i) => <li key={i}>{message}</li>)}</ul></Alert>}</div></Section>
}
function History({ id }: { id: string }) {
  const auth = useAuth()
  return <Section title="历史"><p className="mb-3 text-sm">Campaign 的操作历史由系统 Audit 持有。</p><TechnicalLabel>resource: gfg_showcase_campaign · target: #{id}</TechnicalLabel>
    {auth.can('audit.read') ? <div className="mt-4"><Link className="text-sm text-primary hover:underline" to="/system/audit?resource=gfg_showcase_campaign">打开操作审计</Link><p className="mt-2 text-xs text-muted-foreground">在系统审计中按 target #{id} 核对记录；这里不提供不完整的历史副本。</p></div> : <p className="mt-4 text-sm text-muted-foreground">当前账号没有 audit.read，请由具备操作审计权限的成员查看。</p>}
  </Section>
}
function Workspace({ id }: { id: string }) {
  const { params, update } = useShowcaseURL(); const auth = useAuth(); const client = useQueryClient(); const { toast } = useToast(); const navigate = useNavigate()
  const [editing, setEditing] = useState(false); const [confirmation, setConfirmation] = useState<Lifecycle | null>(null)
  const query = useQuery({ queryKey: ['showcase', 'campaign', id], queryFn: () => showcaseAPI.campaign(id), refetchOnWindowFocus: false })
  const mutation = useMutation({ mutationFn: (action: Lifecycle) => showcaseAPI.lifecycle(id, action), onSuccess: async (saved, action) => {
    setConfirmation(null)
    if (saved) client.setQueryData(['showcase', 'campaign', id], saved)
    if (action === 'delete') { client.removeQueries({ queryKey: ['showcase', 'campaign', id] }); navigate('/game/showcase?tab=campaigns', { replace: true }) }
    await invalidateShowcase(client, action === 'delete' ? undefined : id); toast(`活动已${lifecycleLabels[action]}`)
  }, onError: () => setConfirmation(null) })
  if (query.isLoading) return <LoadingState />
  if (query.error && !query.data) return <ErrorState message={query.error.message} onRetry={() => void query.refetch()} />
  if (!query.data) return null
  const workspace = query.data
  const canWrite = auth.can('content.write') && workspace.state !== 'archived'
  const active = tabs.some(tab => tab.key === params.get('tab')) ? params.get('tab')! : 'overview'
  const actions: Lifecycle[] = workspace.state === 'draft' ? ['publish', 'archive', 'delete'] : workspace.state === 'published' ? ['pause', 'archive'] : workspace.state === 'paused' ? ['resume', 'archive'] : []
  const props = { workspace, canWrite: canWrite && !mutation.isPending, onBusyChange: setEditing }
  return <PageLayout className="min-w-0"><PageHeader title={workspace.internal_name} />
    <div className="flex flex-wrap items-center justify-between gap-3"><div className="flex flex-wrap items-center gap-2"><TechnicalLabel>showcase.campaign · #{id}</TechnicalLabel><CampaignStatus value={workspace.derived_status} /></div>
      <div className="flex flex-wrap gap-2"><Button variant="secondary" onClick={() => navigate('/game/showcase?tab=campaigns')}>返回活动</Button>{canWrite && actions.map(action => <Button key={action} variant={action === 'delete' ? 'danger' : 'secondary'} disabled={editing || mutation.isPending || ((action === 'publish' || action === 'resume') && !workspace.ready_to_publish)} onClick={() => setConfirmation(action)}>{lifecycleLabels[action]}</Button>)}</div>
    </div>
    {editing && <p className="text-xs text-muted-foreground">请先保存或放弃当前编辑，再执行生命周期操作。</p>}{query.error && <ErrorState message={query.error.message} onRetry={() => void query.refetch()} />}{mutation.error && <Alert tone="danger">{mutation.error.message}</Alert>}
    <div className="overflow-x-auto"><div className="min-w-max"><WorkspaceTabs tabs={tabs} active={active} onChange={tab => update({ tab: tab === 'overview' ? null : tab })} /></div></div>
    {active === 'overview' && <Overview workspace={workspace} />}{active === 'content' && <ShowcaseContent {...props} />}{active === 'assets' && <ShowcaseAssets {...props} />}{active === 'schedule' && <ShowcaseSchedule {...props} />}{active === 'stats' && <ShowcaseStats workspace={workspace} />}{active === 'history' && <History id={id} />}
    <ConfirmAction open={!!confirmation} onOpenChange={open => { if (!open && !mutation.isPending) setConfirmation(null) }} title={confirmation ? `${lifecycleLabels[confirmation]}活动` : '活动操作'} description={`${workspace.internal_name} · #${id}。后端将校验当前活动状态和发布条件。`} confirmLabel={confirmation ? `确认${lifecycleLabels[confirmation]}` : '确认'} variant={confirmation === 'delete' ? 'danger' : 'primary'} busy={mutation.isPending} onConfirm={() => { if (confirmation && canWrite && !editing) mutation.mutate(confirmation) }} />
  </PageLayout>
}
export function ShowcaseCampaignPage() {
  const { id } = useParams()
  return id && /^[1-9]\d*$/.test(id) ? <Workspace key={id} id={id} /> : <ErrorState title="活动标识无效" message="请选择真实 Campaign ID。" />
}
