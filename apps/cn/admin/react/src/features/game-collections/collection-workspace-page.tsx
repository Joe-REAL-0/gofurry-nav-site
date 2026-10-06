import { zodResolver } from '@hookform/resolvers/zod'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useEffect, useState } from 'react'
import { Controller, useForm } from 'react-hook-form'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { z } from 'zod'
import { useToast } from '../../app/toast'
import { FormField, PageHeader, PageLayout, Section } from '../../components/admin/page'
import { StatusBadge } from '../../components/admin/status'
import { ErrorState, LoadingState } from '../../components/admin/states'
import { Alert } from '../../components/ui/alert'
import { Button } from '../../components/ui/button'
import { ConfirmAction } from '../../components/ui/dialog'
import { Input, Textarea } from '../../components/ui/input'
import { useUnsavedChanges } from '../../hooks/use-unsaved-changes'
import { ApiError, errorMessage, sendJSON } from '../../lib/api'
import { useAuth } from '../auth/auth-context'
import { collectionContentSchema, collectionEndpoint, contentOf, createCollectionSchema, loadCollectionWorkspace, workspaceKey } from './api'
import { CollectionMembersEditor } from './collection-members'
import { collectionStatusLabels, PUBLIC_REFRESH_NOTICE, type CollectionContent, type CollectionWorkspace, type GameCollection } from './types'

const fields = [{ key: 'name', label: '中文名称' }, { key: 'name_en', label: '英文名称' }, { key: 'info', label: '中文简介' }, { key: 'info_en', label: '英文简介' }] as const
const lifecycleLabels = { publish: '发布', unpublish: '取消发布', archive: '归档', restore: '恢复为草稿' }
type Lifecycle = keyof typeof lifecycleLabels
type Draft = { base: CollectionWorkspace; content: CollectionContent; members: CollectionWorkspace['members'] }
const memberIDs = (members: Draft['members']) => members.map(m => m.game_id).sort((a, b) => a - b)
const contentDirty = (draft: Draft) => JSON.stringify(draft.content) !== JSON.stringify(contentOf(draft.base.collection))
const membersDirty = (draft: Draft) => JSON.stringify(memberIDs(draft.members)) !== JSON.stringify(memberIDs(draft.base.members))
const dirtyDraft = (draft: Draft) => contentDirty(draft) || membersDirty(draft)

export function CreateCollectionPage() {
  const navigate = useNavigate()
  const { toast } = useToast()
  const canWrite = useAuth().can('content.write')
  const [error, setError] = useState('')
  const form = useForm<z.infer<typeof createCollectionSchema>>({ resolver: zodResolver(createCollectionSchema), defaultValues: { code: '', name: '', name_en: '', info: '', info_en: '' } })
  const mutation = useMutation({
    mutationFn: (values: z.infer<typeof createCollectionSchema>) => sendJSON<GameCollection>(collectionEndpoint, 'POST', values), retry: false,
    onSuccess: () => { form.reset(); toast(PUBLIC_REFRESH_NOTICE) },
    onError: err => setError(errorMessage(err)),
  })
  useUnsavedChanges(form.formState.isDirty && !mutation.isSuccess)
  // Navigate only after the saved render releases the unsaved-changes blocker.
  useEffect(() => {
    if (mutation.isSuccess) navigate(`/game/collections/${mutation.data.id}`, { replace: true })
  }, [mutation.isSuccess, mutation.data?.id, navigate])
  return <PageLayout><Link to="/game/collections">返回游戏分区</Link><PageHeader title="新建游戏分区" />
    {error && <Alert tone="danger">{error}</Alert>}
    <Section title="基本内容"><form className="grid gap-4" onSubmit={form.handleSubmit(values => mutation.mutate(values))}>
      <FormField label="Code" help="创建后不可更改。使用小写字母、数字及单个连字符。" error={form.formState.errors.code?.message}><Input {...form.register('code')} disabled={!canWrite || mutation.isPending} /></FormField>
      {fields.map(({ key, label }) => <FormField key={key} label={label} error={form.formState.errors[key]?.message}>{key.startsWith('info') ? <Textarea {...form.register(key)} disabled={!canWrite || mutation.isPending} /> : <Input {...form.register(key)} disabled={!canWrite || mutation.isPending} />}</FormField>)}
      <p className="text-sm text-muted-foreground">草稿可暂不填写简介和收录游戏；发布前须补全双语内容并收录至少两个游戏。</p>
      {canWrite && <Button type="submit" disabled={mutation.isPending}>创建草稿</Button>}
    </form></Section>
  </PageLayout>
}

export function CollectionWorkspacePage() {
  const { id = '' } = useParams()
  const query = useQuery({ queryKey: workspaceKey(id), queryFn: () => loadCollectionWorkspace(id) })
  if (query.isLoading) return <LoadingState />
  if (!query.data) return <ErrorState message={errorMessage(query.error)} onRetry={() => void query.refetch()} />
  return <CollectionWorkspaceEditor key={id} data={query.data} reload={async () => { const result = await query.refetch(); if (result.error) throw result.error; return result.data! }} />
}

export function CollectionWorkspaceEditor({ data, reload }: { data: CollectionWorkspace; reload: () => Promise<CollectionWorkspace> }) {
  const auth = useAuth()
  const client = useQueryClient()
  const { toast } = useToast()
  // Capture one base snapshot on the first edit. Refetch cannot advance its
  // version; only our successful write may rebase the other local draft.
  const [draft, setDraft] = useState<Draft | null>(null)
  const current: Draft = draft ?? { base: data, content: contentOf(data.collection), members: data.members }
  const collection = current.base.collection
  const [error, setError] = useState('')
  const [conflict, setConflict] = useState(false)
  const [confirm, setConfirm] = useState<Lifecycle | null>(null)
  const [loadingMember, setLoadingMember] = useState(false)
  const [reloading, setReloading] = useState(false)
  const dirty = dirtyDraft(current)
  useUnsavedChanges(dirty)
  const mutation = useMutation({
    retry: false,
    mutationFn: ({ kind, value }: { kind: 'content' | 'members' | Lifecycle; value: Draft }) => {
      const endpoint = `${collectionEndpoint}/${value.base.collection.id}`
      const version = value.base.collection.version
      if (kind === 'content') return sendJSON<GameCollection>(endpoint, 'PUT', { version, ...collectionContentSchema.parse(value.content) })
      if (kind === 'members') return sendJSON<GameCollection>(`${endpoint}/members`, 'PUT', { version, game_ids: memberIDs(value.members) })
      return sendJSON<GameCollection>(`${endpoint}/${kind}`, 'POST', { version })
    },
    onMutate: () => client.cancelQueries({ queryKey: workspaceKey(collection.id) }),
    onSuccess: (saved, { kind, value }) => {
      const base = { collection: saved, members: kind === 'members' ? value.members : value.base.members }
      const next: Draft = { base, content: kind === 'content' ? contentOf(saved) : value.content, members: value.members }
      client.setQueryData(workspaceKey(saved.id), base)
      setDraft(dirtyDraft(next) ? next : null)
      setError(''); setConflict(false); setConfirm(null)
      toast(value.base.collection.home_slot !== null && saved.home_slot === null ? `${PUBLIC_REFRESH_NOTICE} 该分区已同步撤出首页入口。` : PUBLIC_REFRESH_NOTICE)
      void client.invalidateQueries({ queryKey: ['game-collections'] })
      void client.invalidateQueries({ queryKey: ['game-collection-home'] })
    },
    onError: err => { setError(errorMessage(err)); setConflict(err instanceof ApiError && err.status === 409); setConfirm(null) },
  })
  const readonly = !auth.can('content.write') || collection.status === 'archived'
  const busy = mutation.isPending || loadingMember || reloading
  const actions: Lifecycle[] = collection.status === 'archived' ? ['restore'] : collection.status === 'published' ? ['unpublish', 'archive'] : ['publish', 'archive']
  const change = (next: Draft) => setDraft(dirtyDraft(next) ? next : null)
  const reloadExplicitly = async () => {
    if (dirty && !window.confirm('放弃未保存的内容和成员修改并重新加载？')) return
    setReloading(true)
    try { const fresh = await reload(); client.setQueryData(workspaceKey(collection.id), fresh); setDraft(null); setError(''); setConflict(false) }
    catch (err) { setError(errorMessage(err)) }
    finally { setReloading(false) }
  }
  return <PageLayout>
    <PageHeader title={collection.name} actions={<div className="flex flex-wrap gap-2">
      <Link to="/game/collections"><Button variant="secondary">返回</Button></Link>
      {auth.can('audit.read') && <Link to="/system/audit?resource=gfg_game_collection"><Button variant="secondary">操作审计</Button></Link>}
      {auth.can('content.write') && actions.map(action => <Button key={action} variant={action === 'archive' ? 'danger' : 'secondary'} disabled={dirty || busy || conflict} onClick={() => setConfirm(action)}>{lifecycleLabels[action]}</Button>)}
    </div>} />
    <div><StatusBadge>{collectionStatusLabels[collection.status]}</StatusBadge></div>
    {dirty && <p className="text-sm text-muted-foreground">请先保存或放弃修改，再执行发布、取消发布、归档或恢复。</p>}
    {error && <Alert tone="danger">{error}{conflict && ' 草稿已保留，请显式重新加载后重试。'}</Alert>}
    <div><Button variant="secondary" disabled={busy} onClick={() => void reloadExplicitly()}>重新加载 / 放弃修改</Button></div>
    <Section title="基本内容"><ContentForm code={collection.code} value={current.content} disabled={readonly || busy} saveDisabled={!contentDirty(current) || conflict} onChange={content => change({ ...current, content })} onSave={() => mutation.mutate({ kind: 'content', value: current })} readonly={readonly} /></Section>
    <Section title={`收录游戏（${current.members.length}）`}>
      <CollectionMembersEditor members={current.members} disabled={readonly || mutation.isPending || reloading} onLoadingChange={setLoadingMember} onChange={members => change({ ...current, members })} />
      {!readonly && <Button className="mt-3" disabled={!membersDirty(current) || busy || conflict} onClick={() => mutation.mutate({ kind: 'members', value: current })}>保存成员</Button>}
    </Section>
    <Section title="公开刷新说明"><p className="text-sm text-muted-foreground">保存后，公开页面将在最多约5分钟内刷新。仅已发布且至少有一个 SFW 可见游戏的分区可配置首页入口；移除最后一个 SFW 游戏时会自动撤下入口。</p></Section>
    <ConfirmAction open={confirm !== null} onOpenChange={open => { if (!open) setConfirm(null) }} title={`${confirm ? lifecycleLabels[confirm] : ''}游戏分区`} description={`确认对“${collection.name}”执行此操作？恢复只会回到草稿，不会恢复首页入口。`} confirmLabel="确认操作" variant={confirm === 'archive' ? 'danger' : 'primary'} busy={busy} onConfirm={() => { if (confirm && !dirty && !busy) mutation.mutate({ kind: confirm, value: current }) }} />
  </PageLayout>
}

function ContentForm({ code, value, disabled, saveDisabled, onChange, onSave, readonly }: { code: string; value: CollectionContent; disabled: boolean; saveDisabled: boolean; onChange: (value: CollectionContent) => void; onSave: () => void; readonly: boolean }) {
  const form = useForm<CollectionContent>({ resolver: zodResolver(collectionContentSchema), values: value })
  return <form className="grid gap-4" onSubmit={form.handleSubmit(onSave)}>
    <FormField label="Code" help="创建后不可更改"><Input value={code} readOnly /></FormField>
    {fields.map(({ key, label }) => <FormField key={key} label={label} error={form.formState.errors[key]?.message}><Controller name={key} control={form.control} render={({ field }) => {
      const props = { ...field, disabled, onChange: (event: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => { field.onChange(event); onChange({ ...value, [key]: event.currentTarget.value }) } }
      return key.startsWith('info') ? <Textarea {...props} /> : <Input {...props} />
    }} /></FormField>)}
    {!readonly && <Button type="submit" disabled={disabled || saveDisabled}>保存内容</Button>}
  </form>
}
