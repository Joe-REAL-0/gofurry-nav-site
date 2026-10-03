import { zodResolver } from '@hookform/resolvers/zod'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useEffect } from 'react'
import { Controller, useForm, type Resolver } from 'react-hook-form'
import { z } from 'zod'
import { useToast } from '../../app/toast'
import { RemoteSelect } from '../../components/admin/operations'
import { FormField, FormSection, Section } from '../../components/admin/page'
import { Alert } from '../../components/ui/alert'
import { Button } from '../../components/ui/button'
import { Input, Textarea } from '../../components/ui/input'
import { Select } from '../../components/ui/select'
import { buildShowcaseContentPayload, invalidateShowcase, showcaseAPI } from './api'
import { contentOptions, primaryOptions, PublishedNotice, secondaryOptions, useEditorGuard, type EditorProps } from './showcase-shared'
import type { CampaignWorkspace, ContentPayload } from './types'

const bounded = (max: number) => z.string().refine(value => Array.from(value).length <= max, `最多 ${max} 个 Unicode 字符`)
export const linkedGameSchema = z.string().refine(value => !value || (/^[1-9]\d*$/.test(value) && Number.isSafeInteger(Number(value))), '请选择有效的游戏')
export const contentSchema = z.object({
  internal_name: bounded(160).refine(value => !!value.trim(), '请输入内部名称'),
  content_type: z.enum(['game', 'crowdfunding', 'tabletop', 'merchandise', 'other']), sponsored: z.boolean(), linked_game_id: linkedGameSchema,
  primary_action_type: z.enum(['', 'game', 'project', 'product', 'website']), primary_target: z.string(),
  secondary_action_type: z.enum(['', 'steam', 'kickstarter', 'website', 'other']), secondary_target: z.string(),
  locales: z.array(z.object({ lang: z.enum(['zh', 'en']), enabled: z.boolean(), title: bounded(160), summary: bounded(400), tags: z.array(z.string()).max(3, '最多 3 个标签'), editorial_note: z.string() })).length(2),
}).superRefine((value, ctx) => {
  if (value.primary_action_type === 'game' && !value.linked_game_id) ctx.addIssue({ code: 'custom', path: ['linked_game_id'], message: '查看游戏需要关联 GoFurry 游戏' })
  const https = (target: string) => { try { const u = new URL(target); return u.protocol === 'https:' && !!u.hostname && !u.username && !u.password && !/[\r\n\t]/.test(target) } catch { return false } }
  if (value.primary_action_type && value.primary_action_type !== 'game' && !https(value.primary_target)) ctx.addIssue({ code: 'custom', path: ['primary_target'], message: '请输入不含账号密码的 HTTPS 地址' })
  if (value.secondary_action_type && !https(value.secondary_target)) ctx.addIssue({ code: 'custom', path: ['secondary_target'], message: '请输入不含账号密码的 HTTPS 地址' })
  if (!value.sponsored) value.locales.forEach((locale, index) => {
    if (Array.from(locale.editorial_note).length > 200) ctx.addIssue({ code: 'custom', path: ['locales', index, 'editorial_note'], message: '最多 200 个 Unicode 字符' })
  })
})
export type ContentValues = z.infer<typeof contentSchema>
export function contentValues(workspace: CampaignWorkspace): ContentValues {
  return {
    internal_name: workspace.internal_name, content_type: workspace.content_type, sponsored: workspace.sponsored,
    linked_game_id: workspace.linked_game_id ?? '', primary_action_type: workspace.primary_action_type ?? '', primary_target: workspace.primary_target ?? '',
    secondary_action_type: workspace.secondary_action_type ?? '', secondary_target: workspace.secondary_target ?? '',
    locales: (['zh', 'en'] as const).map(lang => { const locale = workspace.locales.find(row => row.lang === lang); return { lang, enabled: locale?.enabled ?? false, title: locale?.title ?? '', summary: locale?.summary ?? '', tags: Array.from({ length: 3 }, (_, i) => locale?.tags[i] ?? ''), editorial_note: locale?.editorial_note ?? '' } }),
  }
}
export function contentPatch(values: ContentValues): Partial<ContentPayload> {
  return { ...values, linked_game_id: values.linked_game_id ? Number(values.linked_game_id) : null,
    primary_action_type: values.primary_action_type || null, primary_target: values.primary_action_type ? values.primary_target.trim() || null : null,
    secondary_action_type: values.secondary_action_type || null, secondary_target: values.secondary_target.trim() || null,
    locales: values.locales.map(locale => ({ ...locale, tags: locale.tags.map(tag => tag.trim()).filter(Boolean), editorial_note: values.sponsored ? null : locale.editorial_note || null })),
  }
}

export function ShowcaseContent({ workspace, canWrite, onBusyChange }: EditorProps) {
  const client = useQueryClient(); const { toast } = useToast()
  const form = useForm<ContentValues>({ resolver: zodResolver(contentSchema) as Resolver<ContentValues>, defaultValues: contentValues(workspace) })
  useEffect(() => { if (!form.formState.isDirty) form.reset(contentValues(workspace)) }, [workspace, form])
  const mutation = useMutation({ mutationFn: (values: ContentValues) => showcaseAPI.content(workspace.id, buildShowcaseContentPayload(workspace, contentPatch(values))), onSuccess: async saved => {
    form.reset(contentValues(saved)); client.setQueryData(['showcase', 'campaign', workspace.id], saved)
    await invalidateShowcase(client, workspace.id, saved.linked_game_id); toast('活动内容已保存')
  } })
  useEditorGuard(form.formState.isDirty, mutation.isPending, onBusyChange)
  const disabled = !canWrite || mutation.isPending
  const sponsored = form.watch('sponsored'); const primary = form.watch('primary_action_type'); const secondary = form.watch('secondary_action_type')
  return <Section title="内容"><form className="grid gap-6" onSubmit={form.handleSubmit(values => { if (canWrite) mutation.mutate(values) })}>
    <PublishedNotice workspace={workspace} />{mutation.error && <Alert tone="danger">{mutation.error.message}</Alert>}
    <fieldset disabled={disabled} className="grid min-w-0 gap-6">
      <FormSection title="基础内容"><div className="grid gap-4 md:grid-cols-2">
        <FormField label="内部名称" required error={form.formState.errors.internal_name?.message}><Input {...form.register('internal_name')} /></FormField>
        <FormField label="内容类型"><Controller control={form.control} name="content_type" render={({ field }) => <Select ariaLabel="内容类型" disabled={disabled} value={field.value} onValueChange={field.onChange} options={contentOptions} />} /></FormField>
        <label className="flex items-center gap-2 text-sm"><input type="checkbox" {...form.register('sponsored')} />商业推广</label>
        <FormField label="关联 GoFurry 游戏" error={form.formState.errors.linked_game_id?.message}>{disabled ? <p>{workspace.linked_game_id ? `游戏 #${workspace.linked_game_id}` : '未关联'}</p> : <Controller control={form.control} name="linked_game_id" render={({ field }) => <RemoteSelect endpoint="/api/v1/options/games" pageSize={10} debounceMs={300} value={field.value ? { id: field.value, label: `游戏 #${field.value}` } : null} onChange={option => field.onChange(option?.id ?? '')} placeholder="搜索关联游戏…" />} />}</FormField>
      </div></FormSection>
      <div className="grid min-w-0 gap-6 lg:grid-cols-2">{(['中文', 'English'] as const).map((label, i) => <FormSection key={label} title={label}>
        <label className="flex items-center gap-2 text-sm"><input type="checkbox" aria-label={`启用 ${label}`} {...form.register(`locales.${i}.enabled`)} />启用 {label}</label>
        <FormField label={`${label}标题`} error={form.formState.errors.locales?.[i]?.title?.message}><Input {...form.register(`locales.${i}.title`)} /></FormField>
        <FormField label={`${label}摘要`} error={form.formState.errors.locales?.[i]?.summary?.message} help="最多 400 个 Unicode 字符；不会跨语言回退。"><Textarea {...form.register(`locales.${i}.summary`)} /></FormField>
        <div className="grid gap-2">{[0, 1, 2].map(index => <FormField key={index} label={`${label}标签 ${index + 1}`}><Input {...form.register(`locales.${i}.tags.${index}`)} /></FormField>)}</div>
        {!sponsored && <FormField label={`${label}编辑推荐语`} error={form.formState.errors.locales?.[i]?.editorial_note?.message}><Textarea {...form.register(`locales.${i}.editorial_note`)} /></FormField>}
      </FormSection>)}</div>
      <FormSection title="动作"><div className="grid gap-4 md:grid-cols-2">
        <FormField label="Primary 动作"><Controller control={form.control} name="primary_action_type" render={({ field }) => <Select ariaLabel="Primary 动作" disabled={disabled} value={field.value} onValueChange={field.onChange} options={primaryOptions} />} /></FormField>
        {primary && primary !== 'game' && <FormField label="Primary HTTPS 地址" error={form.formState.errors.primary_target?.message}><Input {...form.register('primary_target')} /></FormField>}
        <FormField label="Secondary 动作"><Controller control={form.control} name="secondary_action_type" render={({ field }) => <Select ariaLabel="Secondary 动作" disabled={disabled} value={field.value} onValueChange={field.onChange} options={secondaryOptions} />} /></FormField>
        {secondary && <FormField label="Secondary HTTPS 地址" error={form.formState.errors.secondary_target?.message}><Input {...form.register('secondary_target')} /></FormField>}
      </div></FormSection>
    </fieldset>
    {canWrite && <div className="flex justify-end"><Button disabled={!form.formState.isDirty || mutation.isPending}>保存内容</Button></div>}
  </form></Section>
}
