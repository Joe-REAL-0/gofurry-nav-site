import { zodResolver } from '@hookform/resolvers/zod'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useEffect } from 'react'
import { Controller, useForm, type Resolver } from 'react-hook-form'
import { z } from 'zod'
import { useToast } from '../../app/toast'
import { FormField, Section } from '../../components/admin/page'
import { Alert } from '../../components/ui/alert'
import { Button } from '../../components/ui/button'
import { DateTimePicker } from '../../components/ui/date-picker'
import { Input } from '../../components/ui/input'
import { Select } from '../../components/ui/select'
import { invalidateShowcase, showcaseAPI } from './api'
import { PublishedNotice, useEditorGuard, type EditorProps } from './showcase-shared'
import { chinaWallTimeToRFC3339, operatingTimezone, rfc3339ToChinaWallTime } from './showcase-time'
import type { CampaignWorkspace } from './types'

const wall = z.string().refine(value => { try { return !!chinaWallTimeToRFC3339(value) } catch { return false } }, '请选择有效的上海日期和时间')
export const scheduleSchema = z.object({ starts_at: wall, ends_at: wall, weight: z.coerce.number().int().min(1).max(10000), pin_position: z.enum(['', '1', '2', '3', '4']) }).refine(value => value.starts_at < value.ends_at, { path: ['ends_at'], message: '结束时间必须晚于开始时间' })
type Values = z.infer<typeof scheduleSchema>
const values = (w: CampaignWorkspace): Values => ({ starts_at: rfc3339ToChinaWallTime(w.starts_at), ends_at: rfc3339ToChinaWallTime(w.ends_at), weight: w.weight, pin_position: (w.pin_position ? String(w.pin_position) : '') as Values['pin_position'] })
export function ShowcaseSchedule({ workspace, canWrite, onBusyChange }: EditorProps) {
  const client = useQueryClient(); const { toast } = useToast()
  const form = useForm<Values>({ resolver: zodResolver(scheduleSchema) as Resolver<Values>, defaultValues: values(workspace) })
  useEffect(() => { if (!form.formState.isDirty) form.reset(values(workspace)) }, [workspace, form])
  const mutation = useMutation({ mutationFn: (v: Values) => showcaseAPI.schedule(workspace.id, { starts_at: chinaWallTimeToRFC3339(v.starts_at), ends_at: chinaWallTimeToRFC3339(v.ends_at), weight: v.weight, pin_position: v.pin_position ? Number(v.pin_position) : null }), onSuccess: async saved => {
    form.reset(values(saved)); client.setQueryData(['showcase', 'campaign', workspace.id], saved); await invalidateShowcase(client, workspace.id); toast('排期与展示已保存')
  } })
  useEditorGuard(form.formState.isDirty, mutation.isPending, onBusyChange)
  const disabled = !canWrite || mutation.isPending
  return <Section title="排期与展示" description={operatingTimezone}><form className="grid gap-5" onSubmit={form.handleSubmit(v => { if (canWrite) mutation.mutate(v) })}>
    <PublishedNotice workspace={workspace} />{mutation.error && <Alert tone="danger">{mutation.error.message}</Alert>}
    <div className="grid gap-4 md:grid-cols-2">{(['starts_at', 'ends_at'] as const).map((name, index) => <FormField key={name} label={index ? '结束时间' : '开始时间'} required error={form.formState.errors[name]?.message}><Controller name={name} control={form.control} render={({ field }) => <DateTimePicker disabled={disabled} ariaLabel={index ? '结束时间' : '开始时间'} value={field.value} onValueChange={field.onChange} />} /></FormField>)}</div>
    <FormField label="Weight" error={form.formState.errors.weight?.message} help="Weight 影响同类候选在稳定编排中的选择概率，不保证固定位置。"><Input disabled={disabled} type="number" min={1} max={10000} {...form.register('weight')} /></FormField>
    <FormField label="Pin Position" help="固定位置仅用于特殊策展或有明确位置约定的推广活动。"><Controller name="pin_position" control={form.control} render={({ field }) => <Select ariaLabel="Pin Position" disabled={disabled} value={field.value} onValueChange={field.onChange} options={[{ value: '', label: '不固定' }, ...[1, 2, 3, 4].map(n => ({ value: String(n), label: `#${n}` }))]} />} /></FormField>
    {canWrite && <div className="flex justify-end"><Button disabled={!form.formState.isDirty || mutation.isPending}>保存排期与展示</Button></div>}
  </form></Section>
}
