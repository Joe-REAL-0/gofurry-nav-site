import { zodResolver } from '@hookform/resolvers/zod'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useEffect, useState } from 'react'
import { Controller, useForm } from 'react-hook-form'
import { useNavigate } from 'react-router-dom'
import { z } from 'zod'
import { useToast } from '../../app/toast'
import { RemoteSelect } from '../../components/admin/operations'
import { FormField } from '../../components/admin/page'
import { Alert } from '../../components/ui/alert'
import { Button } from '../../components/ui/button'
import { ConfirmAction, Dialog } from '../../components/ui/dialog'
import { Input } from '../../components/ui/input'
import { Select } from '../../components/ui/select'
import { useUnsavedChanges } from '../../hooks/use-unsaved-changes'
import { invalidateShowcase, showcaseAPI } from './api'
import { linkedGameSchema } from './showcase-content'
import { contentOptions } from './showcase-shared'

const schema = z.object({ internal_name: z.string().trim().min(1, '请输入内部名称').refine(v => Array.from(v).length <= 160, '最多 160 个 Unicode 字符'), content_type: z.enum(['game', 'crowdfunding', 'tabletop', 'merchandise', 'other']), sponsored: z.boolean(), linked_game_id: linkedGameSchema })
export function ShowcaseCampaignCreate({ close }: { close: () => void }) {
  const navigate = useNavigate(); const client = useQueryClient(); const { toast } = useToast()
  const [discard, setDiscard] = useState(false)
  const [createdId, setCreatedId] = useState<string | null>(null)
  const form = useForm<z.infer<typeof schema>>({ resolver: zodResolver(schema), defaultValues: { internal_name: '', content_type: 'game', sponsored: false, linked_game_id: '' } })
  const mutation = useMutation({ mutationFn: (value: z.infer<typeof schema>) => showcaseAPI.create({ ...value, linked_game_id: value.linked_game_id ? Number(value.linked_game_id) : null }), onSuccess: saved => {
    form.reset(); client.setQueryData(['showcase', 'campaign', saved.id], saved); void invalidateShowcase(client)
    toast('活动草稿已创建'); setCreatedId(saved.id)
  } })
  useUnsavedChanges(form.formState.isDirty && createdId === null)
  useEffect(() => {
    if (createdId !== null) { close(); void navigate(`/game/showcase/${createdId}`) }
  }, [createdId, close, navigate])
  return <><Dialog open onOpenChange={open => { if (!open && !mutation.isPending) { if (form.formState.isDirty) setDiscard(true); else close() } }} title="新建活动" description="先创建草稿，再进入工作台编辑内容和上传素材。">
    <form className="grid gap-4" onSubmit={form.handleSubmit(value => mutation.mutate(value))}>
      {mutation.error && <Alert tone="danger">{mutation.error.message}</Alert>}
      <FormField label="内部名称" required error={form.formState.errors.internal_name?.message}><Input disabled={mutation.isPending} {...form.register('internal_name')} /></FormField>
      <FormField label="内容类型"><Controller name="content_type" control={form.control} render={({ field }) => <Select disabled={mutation.isPending} ariaLabel="内容类型" value={field.value} onValueChange={field.onChange} options={contentOptions} />} /></FormField>
      <label className="flex items-center gap-2 text-sm"><input type="checkbox" disabled={mutation.isPending} {...form.register('sponsored')} />商业推广</label>
      <FormField label="关联 GoFurry 游戏（可选）" error={form.formState.errors.linked_game_id?.message}><Controller name="linked_game_id" control={form.control} render={({ field }) => <RemoteSelect disabled={mutation.isPending} resultsLayout="inline" endpoint="/api/v1/options/games" value={field.value} onChange={option => field.onChange(option?.id ?? '')} pageSize={10} debounceMs={300} placeholder="搜索关联游戏…" />} /></FormField>
      <Button disabled={mutation.isPending}>创建草稿并进入</Button>
    </form>
  </Dialog><ConfirmAction open={discard} onOpenChange={setDiscard} title="放弃新建活动？" description="尚未提交的内容将被丢弃。" confirmLabel="放弃修改" onConfirm={close} /></>
}
