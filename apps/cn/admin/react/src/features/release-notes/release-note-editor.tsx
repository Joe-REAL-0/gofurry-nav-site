import { zodResolver } from '@hookform/resolvers/zod'
import { ArrowLeft } from '@phosphor-icons/react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useEffect, useState } from 'react'
import { Controller, useForm, useWatch } from 'react-hook-form'
import { Link, useNavigate } from 'react-router-dom'
import { useToast } from '../../app/toast'
import { FormField, PageHeader, PageLayout, Section } from '../../components/admin/page'
import { Alert } from '../../components/ui/alert'
import { Button } from '../../components/ui/button'
import { DateTimePicker } from '../../components/ui/date-picker'
import { ConfirmAction } from '../../components/ui/dialog'
import { Input, Textarea } from '../../components/ui/input'
import { useUnsavedChanges } from '../../hooks/use-unsaved-changes'
import { errorMessage, sendJSON } from '../../lib/api'
import { useAuth } from '../auth/auth-context'
import { MarkdownEditor } from './markdown-editor'
import { ReleaseStatus } from './release-status'
import { isFutureReleaseTime, releaseNoteForm, releaseNoteKey, releaseNotePayload, releaseNoteSchema, releaseNotesEndpoint, releaseNotesListKey, releasePickerNow, releaseWallTime, type ReleaseNote, type ReleaseNoteForm } from './release-note-model'

type Action = 'save' | 'publish' | 'schedule' | 'unpublish' | 'delete'
const actionLabels = { save: '保存', publish: '立即发布', schedule: '定时发布', unpublish: '取消发布', delete: '删除公告' }

export function ReleaseNoteEditor({ record }: { record: ReleaseNote | null }) {
  const canWrite = useAuth().can('content.write')
  const navigate = useNavigate()
  const client = useQueryClient()
  const { toast } = useToast()
  const [language, setLanguage] = useState<'zh' | 'en'>('zh')
  const [confirmation, setConfirmation] = useState<Exclude<Action, 'save'> | null>(null)
  const [destination, setDestination] = useState<string | null>(null)
  const [error, setError] = useState('')
  const form = useForm<ReleaseNoteForm>({ resolver: zodResolver(releaseNoteSchema), defaultValues: releaseNoteForm(record) })
  const { isDirty, errors } = form.formState
  useUnsavedChanges(isDirty, '存在未保存的修改，确定离开吗？')
  // Let RHF reset and the router blocker observe the clean form before leaving.
  useEffect(() => { if (destination && !isDirty) void navigate(destination, { replace: true }) }, [destination, isDirty, navigate])

  const accept = (saved: ReleaseNote) => {
    client.setQueryData(releaseNoteKey(saved.id), saved)
    form.reset(releaseNoteForm(saved))
    void client.invalidateQueries({ queryKey: releaseNotesListKey })
  }
  const mutation = useMutation({
    mutationFn: async ({ action, values }: { action: Action; values: ReleaseNoteForm }) => {
      if (!canWrite) throw new Error('当前账号只有只读权限。')
      if (action === 'schedule' && !isFutureReleaseTime(values.published_at)) throw new Error('定时发布需要选择未来时间（Asia/Shanghai）。')
      if ((action === 'publish' || action === 'schedule') && (!values.title.trim() || !values.body.trim())) throw new Error('发布需要中文标题和正文。')
      if (action === 'unpublish' && isDirty) throw new Error('请先保存或放弃未保存的修改，再取消发布。')
      let saved = record
      try {
        if (action === 'delete') {
          if (!saved) throw new Error('公告尚未保存。')
          await sendJSON(`${releaseNotesEndpoint}/${saved.id}`, 'DELETE')
          form.reset(form.getValues())
          client.removeQueries({ queryKey: releaseNoteKey(saved.id) })
          void client.invalidateQueries({ queryKey: releaseNotesListKey })
          setDestination('/nav/update-notices')
          return
        }
        if (action === 'save' || ((action === 'publish' || action === 'schedule') && (!saved || isDirty))) {
          saved = await sendJSON<ReleaseNote>(saved ? `${releaseNotesEndpoint}/${saved.id}` : releaseNotesEndpoint, saved ? 'PUT' : 'POST', releaseNotePayload(values))
          accept(saved)
        }
        if (!saved) throw new Error('公告尚未保存。')
        if (action === 'publish' || action === 'schedule') {
          saved = action === 'publish'
            ? await sendJSON<ReleaseNote>(`${releaseNotesEndpoint}/${saved.id}/publish`, 'POST')
            : await sendJSON<ReleaseNote>(`${releaseNotesEndpoint}/${saved.id}/publish`, 'POST', { published_at: values.published_at })
          accept(saved)
        } else if (action === 'unpublish') {
          saved = await sendJSON<ReleaseNote>(`${releaseNotesEndpoint}/${saved.id}/unpublish`, 'POST')
          accept(saved)
        }
      } finally {
        // A failed publish after successful creation still has a real, saved Draft.
        if (!record && saved) setDestination(`/nav/update-notices/${saved.id}`)
      }
    },
    onSuccess: (_result, { action }) => { setError(''); setConfirmation(null); toast(`${actionLabels[action]}成功`) },
    onError: err => { setError(errorMessage(err)); setConfirmation(null); toast(errorMessage(err), 'danger') },
  })
  const busy = mutation.isPending || form.formState.isSubmitting
  const submit = (action: Action) => {
    if (busy) return
    if (action === 'delete' || action === 'unpublish') mutation.mutate({ action, values: form.getValues() })
    else void form.handleSubmit(values => mutation.mutate({ action, values }), () => { setConfirmation(null); setError('请检查表单中的错误（包括另一种语言）。') })()
  }
  const titleField = language === 'zh' ? 'title' : 'title_en'
  const summaryField = language === 'zh' ? 'summary' : 'summary_en'
  const bodyField = language === 'zh' ? 'body' : 'body_en'
  const languageLabel = language === 'zh' ? '中文' : 'English'
  const values = useWatch({ control: form.control })
  const description = confirmation === 'delete' ? '删除后该公告会从管理列表和公开页面移除。'
    : confirmation === 'unpublish' ? '取消发布后，公开页面将不再显示该公告。'
      : `即将${confirmation === 'schedule' ? `在 ${releaseWallTime(values.published_at) || '所选时间'}（Asia/Shanghai）` : '立即'}公开 ${values.version || values.title || '这篇 Release Note'}。${isDirty ? '将先保存当前编辑内容，保存失败时不会发布。' : ''}`
  return <PageLayout>
    <Link className="inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-primary" to="/nav/update-notices"><ArrowLeft aria-hidden="true" className="size-4" />返回更新公告</Link>
    <PageHeader title={record ? '编辑 Release Note' : '新建 Release Note'} actions={<ReleaseStatus record={record ?? { publication_state: 'draft', published_at: null }} />} />
    {error && <Alert tone="danger">{error}</Alert>}
    {!canWrite && <Alert tone="info">只读模式，可以切换语言和预览正文。</Alert>}
    <form className="grid min-w-0 gap-4" onSubmit={event => { event.preventDefault(); submit('save') }}>
      <Section title="Release Metadata" description="时间为 Asia/Shanghai（UTC+8）；版本与 Commit SHA 可留空。">
        <fieldset disabled={busy || !canWrite} className="grid gap-4 border-0 p-0 md:grid-cols-3">
          <FormField label="Version" error={errors.version?.message}><Input {...form.register('version')} /></FormField>
          <FormField label="Commit SHA" error={errors.commit_sha?.message}><Input className="font-mono" {...form.register('commit_sha')} /></FormField>
          <FormField label="发布时间" help="草稿可留空；立即发布由服务端确定时间。" error={errors.published_at?.message}><Controller name="published_at" control={form.control} render={({ field }) => <DateTimePicker value={field.value} onValueChange={field.onChange} ariaLabel="发布时间（Asia/Shanghai）" now={releasePickerNow} disabled={busy || !canWrite} />} /></FormField>
        </fieldset>
      </Section>
      <Section title="公告内容" actions={<div className="flex gap-1" aria-label="编辑语言">{(['zh', 'en'] as const).map(locale => <Button type="button" key={locale} size="sm" variant={language === locale ? 'primary' : 'ghost'} aria-pressed={language === locale} disabled={busy} onClick={() => setLanguage(locale)}>{locale === 'zh' ? '中文' : 'English'}</Button>)}</div>}>
        <div className="grid min-w-0 gap-4">
          <FormField label={`${languageLabel}标题`} error={errors[titleField]?.message}><Input key={titleField} readOnly={!canWrite} disabled={busy} {...form.register(titleField)} /></FormField>
          <FormField label={`${languageLabel}摘要`}><Textarea key={summaryField} readOnly={!canWrite} disabled={busy} {...form.register(summaryField)} /></FormField>
          <Controller name={bodyField} control={form.control} render={({ field }) => <MarkdownEditor value={field.value} onChange={field.onChange} label={`${languageLabel}正文`} readOnly={!canWrite} busy={busy} />} />
        </div>
      </Section>
      {canWrite && <div className="flex flex-wrap items-center justify-between gap-3">
        <div className="flex flex-wrap gap-2">{record && <Button type="button" variant="danger" disabled={busy} onClick={() => setConfirmation('delete')}>删除公告</Button>}{record?.publication_state === 'published' && <Button type="button" variant="secondary" disabled={busy || isDirty} title={isDirty ? '请先保存或放弃修改' : undefined} onClick={() => setConfirmation('unpublish')}>取消发布</Button>}</div>
        <div className="flex flex-wrap items-center gap-2"><span className="text-xs text-muted-foreground" role="status">{isDirty ? '有未保存的修改' : record ? '已保存' : '尚未创建'}</span><Button type="submit" variant="secondary" disabled={busy}>{record ? '保存修改' : '保存草稿'}</Button><Button type="button" disabled={busy} onClick={() => setConfirmation('publish')}>立即发布</Button><Button type="button" variant="secondary" disabled={busy} onClick={() => setConfirmation('schedule')}>定时发布</Button></div>
      </div>}
    </form>
    <ConfirmAction open={confirmation !== null} onOpenChange={open => { if (!open && !busy) setConfirmation(null) }} title={confirmation ? actionLabels[confirmation] : ''} description={description} confirmLabel={confirmation ? `确认${actionLabels[confirmation]}` : '确认'} variant={confirmation === 'publish' || confirmation === 'schedule' ? 'primary' : 'danger'} busy={busy} onConfirm={() => { if (confirmation) submit(confirmation) }} />
  </PageLayout>
}
