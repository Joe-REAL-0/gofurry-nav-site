import { z } from 'zod'

export const releaseNotesEndpoint = '/api/v1/nav/update-notices'
export const releaseNotesListKey = ['release-notes', 'list'] as const
export const releaseNoteKey = (id: string | number) => ['release-notes', 'detail', String(id)] as const

export type ReleaseNote = {
  id: number; version: string | null; commit_sha: string | null
  title: string; title_en: string; summary: string; summary_en: string; body: string; body_en: string
  publication_state: 'draft' | 'published'; published_at: string | null
  create_time: string; update_time: string; deleted: boolean
}

export const releaseNoteSchema = z.object({
  version: z.string().refine(value => [...value.trim()].length <= 64, '版本最多 64 个字符'),
  commit_sha: z.string().refine(value => !value.trim() || /^[a-f\d]{7,64}$/i.test(value.trim()), 'Commit SHA 应为 7–64 位十六进制字符'),
  published_at: z.string().refine(value => !value || parseReleaseTime(value) !== null, '请选择有效的发布时间'),
  title: z.string().max(120, '标题最多 120 个字符'), title_en: z.string().max(120, '标题最多 120 个字符'),
  summary: z.string(), summary_en: z.string(), body: z.string(), body_en: z.string(),
})
export type ReleaseNoteForm = z.infer<typeof releaseNoteSchema>

export function releaseNoteForm(record?: ReleaseNote | null): ReleaseNoteForm {
  return {
    version: record?.version ?? '', commit_sha: record?.commit_sha ?? '', published_at: releaseWallTime(record?.published_at),
    title: record?.title ?? '', title_en: record?.title_en ?? '', summary: record?.summary ?? '', summary_en: record?.summary_en ?? '',
    body: record?.body ?? '', body_en: record?.body_en ?? '',
  }
}

export function releaseNotePayload(values: ReleaseNoteForm) {
  return { ...values, version: values.version.trim() || null, commit_sha: values.commit_sha.trim().toLowerCase() || null }
}

// P1 stores China-site wall timestamps. Never parse unzoned input in browser local time.
export function parseReleaseTime(value?: string | null): number | null {
  if (!value) return null
  const wall = /^(\d{4})-(\d{2})-(\d{2})[ T](\d{2}):(\d{2})(?::(\d{2}))?$/.exec(value)
  if (wall) {
    const [, year, month, day, hour, minute, second = '00'] = wall
    const date = new Date(`${year}-${month}-${day}T${hour}:${minute}:${second}Z`)
    if (!Number.isFinite(date.getTime()) || date.toISOString().slice(0, 19) !== `${year}-${month}-${day}T${hour}:${minute}:${second}`) return null
    return date.getTime() - 8 * 60 * 60 * 1000
  }
  if (!/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:Z|[+-]\d{2}:\d{2})$/i.test(value)) return null
  const time = Date.parse(value)
  return Number.isFinite(time) ? time : null
}

export function releaseWallTime(value?: string | null): string {
  const time = parseReleaseTime(value)
  return time === null ? '' : new Date(time + 8 * 60 * 60 * 1000).toISOString().slice(0, 19).replace('T', ' ')
}

// Presentation only. The backend remains authoritative for public visibility.
export function releaseDisplayStatus(record: Pick<ReleaseNote, 'publication_state' | 'published_at'>, now = Date.now()) {
  if (record.publication_state === 'draft') return 'Draft'
  return isFutureReleaseTime(record.published_at, now) ? 'Scheduled' : 'Published'
}

export function isFutureReleaseTime(value?: string | null, now = Date.now()) {
  const time = parseReleaseTime(value)
  return time !== null && time > now
}

// DateTimePicker reads local calendar components; supply China's wall components.
// This is a picker convenience only, never a Publish Now request timestamp.
export function releasePickerNow() {
  const wall = new Date(Date.now() + 8 * 60 * 60 * 1000)
  return new Date(wall.getUTCFullYear(), wall.getUTCMonth(), wall.getUTCDate(), wall.getUTCHours(), wall.getUTCMinutes())
}
