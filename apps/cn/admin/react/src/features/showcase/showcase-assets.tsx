import { zodResolver } from '@hookform/resolvers/zod'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useEffect, useState } from 'react'
import { useForm, type Resolver } from 'react-hook-form'
import { z } from 'zod'
import { useToast } from '../../app/toast'
import { FilePicker } from '../../components/admin/file-picker'
import { Section } from '../../components/admin/page'
import { TechnicalLabel } from '../../components/admin/status'
import { Alert } from '../../components/ui/alert'
import { Button } from '../../components/ui/button'
import { ConfirmAction } from '../../components/ui/dialog'
import { buildShowcaseContentPayload, invalidateShowcase, showcaseAPI } from './api'
import { PublishedNotice, useEditorGuard, type EditorProps } from './showcase-shared'

type Variant = 'desktop' | 'mobile'
const focalSchema = z.object({ focal_x: z.coerce.number().min(0).max(1), focal_y: z.coerce.number().min(0).max(1) })
type Focal = z.infer<typeof focalSchema>
const focalPoints = [
  { label: '左上', symbol: '↖', x: 0, y: 0 }, { label: '上方', symbol: '↑', x: 0.5, y: 0 }, { label: '右上', symbol: '↗', x: 1, y: 0 },
  { label: '左侧', symbol: '←', x: 0, y: 0.5 }, { label: '居中', symbol: '●', x: 0.5, y: 0.5 }, { label: '右侧', symbol: '→', x: 1, y: 0.5 },
  { label: '左下', symbol: '↙', x: 0, y: 1 }, { label: '下方', symbol: '↓', x: 0.5, y: 1 }, { label: '右下', symbol: '↘', x: 1, y: 1 },
]
const clampFocal = (value: number) => Math.max(0, Math.min(1, value))
export function artworkFileError(file: File) {
  if (!file.size || file.size > 5 * 1024 * 1024) return 'AVIF 文件必须非空且不超过 5 MiB'
  if ((file.type && file.type !== 'image/avif') || !/\.avif$/i.test(file.name)) return '请选择原始 AVIF 文件'
  return ''
}
function usePreview(file: File | null) {
  const [url, setURL] = useState('')
  useEffect(() => { if (!file) { setURL(''); return }; const next = URL.createObjectURL(file); setURL(next); return () => URL.revokeObjectURL(next) }, [file])
  return url
}
export function ShowcaseAssets({ workspace, canWrite, onBusyChange }: EditorProps) {
  const client = useQueryClient(); const { toast } = useToast()
  const [files, setFiles] = useState<Record<Variant, File | null>>({ desktop: null, mobile: null })
  const [fileError, setFileError] = useState(''); const [clear, setClear] = useState<Variant | null>(null)
  const desktop = usePreview(files.desktop); const mobile = usePreview(files.mobile)
  const form = useForm<Focal>({ resolver: zodResolver(focalSchema) as Resolver<Focal>, defaultValues: { focal_x: workspace.focal_x, focal_y: workspace.focal_y } })
  useEffect(() => { if (!form.formState.isDirty) form.reset({ focal_x: workspace.focal_x, focal_y: workspace.focal_y }) }, [workspace, form])
  const saveFocal = useMutation({ mutationFn: (focal: Focal) => showcaseAPI.content(workspace.id, buildShowcaseContentPayload(workspace, focal)), onSuccess: async saved => {
    form.reset({ focal_x: saved.focal_x, focal_y: saved.focal_y }); client.setQueryData(['showcase', 'campaign', workspace.id], saved)
    await invalidateShowcase(client, workspace.id); toast('焦点已保存')
  } })
  const upload = useMutation({ mutationFn: ({ variant, file }: { variant: Variant; file: File }) => showcaseAPI.upload(workspace.id, variant, file), onSuccess: async (result, { variant }) => {
    setFiles(current => ({ ...current, [variant]: null }))
    toast(result.warnings?.length || result.mirror !== 'ready' ? `已发布到 COS；R2 Mirror 同步失败，可稍后通过云资源修复。${result.warnings?.join('；') ?? ''}` : '素材已发布', result.warnings?.length || result.mirror !== 'ready' ? 'info' : 'success')
    await invalidateShowcase(client, workspace.id)
  } })
  const clearing = useMutation({ mutationFn: (variant: Variant) => showcaseAPI.clear(workspace.id, variant), onSuccess: async () => { setClear(null); await invalidateShowcase(client, workspace.id); toast('素材引用已清除') }, onError: () => setClear(null) })
  const busy = upload.isPending || clearing.isPending || saveFocal.isPending
  const x = form.watch('focal_x'), y = form.watch('focal_y')
  function selectFocal(x: number, y: number) {
    if (!canWrite || busy) return
    form.setValue('focal_x', clampFocal(x), { shouldDirty: true, shouldValidate: true })
    form.setValue('focal_y', clampFocal(y), { shouldDirty: true, shouldValidate: true })
  }
  useEditorGuard(form.formState.isDirty || !!files.desktop || !!files.mobile, busy, onBusyChange)
  return <div className="grid min-w-0 gap-4"><PublishedNotice workspace={workspace} />{[fileError, upload.error?.message, clearing.error?.message, saveFocal.error?.message].filter(Boolean).map((error, index) => <Alert key={index} tone="danger">{error}</Alert>)}
    <div className="grid min-w-0 gap-4 xl:grid-cols-2">{(['desktop', 'mobile'] as const).map(variant => {
      const label = variant === 'desktop' ? 'Desktop' : 'Mobile'; const url = variant === 'desktop' ? desktop : mobile
      const objectKey = workspace[`${variant}_object_key`]
      return <Section key={variant} title={`${label} Artwork`} description={`${variant === 'desktop' ? '1600 × 800 · 发布前必需' : '1200 × 675 · 可选'} · AVIF · ≤5 MiB`}>
        <div className="grid min-w-0 gap-3">
          {url ? <button type="button" disabled={!canWrite || busy} aria-label={`在 ${label} 预览中选择焦点（键盘按下设为居中）`} className={`relative w-full cursor-crosshair overflow-hidden rounded border focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:cursor-default ${variant === 'desktop' ? 'aspect-[2/1]' : 'aspect-video'}`} onClick={event => {
            const rect = event.currentTarget.getBoundingClientRect()
            if (event.detail === 0) { selectFocal(0.5, 0.5); return }
            if (rect.width > 0 && rect.height > 0) selectFocal((event.clientX - rect.left) / rect.width, (event.clientY - rect.top) / rect.height)
          }}><img src={url} alt={`${label} 本地预览`} draggable={false} className="h-full w-full object-cover" style={{ objectPosition: `${x * 100}% ${y * 100}%` }} /><span aria-hidden="true" className="pointer-events-none absolute size-4 -translate-x-1/2 -translate-y-1/2 rounded-full border-2 border-primary-foreground bg-primary ring-2 ring-primary" style={{ left: `${x * 100}%`, top: `${y * 100}%` }} /></button> : <div className="rounded border bg-surface-muted p-5 text-sm text-muted-foreground">{objectKey ? '已配置 Managed Artwork' : '尚未配置素材'}<p className="mt-2">重载后显示对象标识；此处不解析 CDN 地址。</p></div>}
          {objectKey && <code className="break-all text-xs">{objectKey}</code>}
          {canWrite && <><FilePicker label={`选择 ${label} AVIF`} accept="image/avif,.avif" file={files[variant]} disabled={busy} help="上传原始文件，不进行裁剪或转换；尺寸由后端权威校验。" onSelect={file => {
            setFileError(''); if (file) { const error = artworkFileError(file); if (error) { setFileError(error); return } }
            setFiles(current => ({ ...current, [variant]: file }))
          }} /><div className="flex flex-wrap gap-2">
            <Button disabled={!files[variant] || busy} onClick={() => { const file = files[variant]; if (file) upload.mutate({ variant, file }) }}>上传 {label}</Button>
            {files[variant] && <Button variant="ghost" disabled={busy} onClick={() => setFiles(current => ({ ...current, [variant]: null }))}>取消 {label} 文件</Button>}
            {objectKey && <Button variant="secondary" disabled={busy} onClick={() => setClear(variant)}>清除 {label}</Button>}
          </div></>}
        </div>
      </Section>
    })}</div>
    <Section title="素材焦点" description="选择图片中最需要保留的主体位置，响应式裁切时会优先保留该区域。"><form className="grid gap-4" onSubmit={form.handleSubmit(value => { if (canWrite) saveFocal.mutate(value) })}>
      <div role="group" aria-label="素材焦点位置" className="grid w-fit grid-cols-3 gap-1">{focalPoints.map(point => <Button key={point.label} type="button" size="icon" className="size-11" variant={Math.round(x * 2) / 2 === point.x && Math.round(y * 2) / 2 === point.y ? 'primary' : 'secondary'} aria-label={point.label} aria-pressed={Math.round(x * 2) / 2 === point.x && Math.round(y * 2) / 2 === point.y} disabled={!canWrite || busy} onClick={() => selectFocal(point.x, point.y)}><span aria-hidden="true">{point.symbol}</span></Button>)}</div>
      <TechnicalLabel>focal_x {x.toFixed(2)} · focal_y {y.toFixed(2)}</TechnicalLabel>
      {canWrite && <div className="flex justify-end"><Button disabled={!form.formState.isDirty || busy}>保存焦点</Button></div>}
    </form></Section>
    <ConfirmAction open={!!clear} onOpenChange={open => { if (!open && !clearing.isPending) setClear(null) }} title={`清除 ${clear === 'desktop' ? 'Desktop' : 'Mobile'} 素材`} description="只清除活动的素材引用，保留云端文件。已发布活动的必要素材由后端校验。" confirmLabel="确认清除" busy={clearing.isPending} onConfirm={() => { if (clear && canWrite) clearing.mutate(clear) }} />
  </div>
}
