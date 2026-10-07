import { useEffect } from 'react'
import { useSearchParams } from 'react-router-dom'
import { StatusBadge } from '../../components/admin/status'
import { Alert } from '../../components/ui/alert'
import { Select } from '../../components/ui/select'
import { useUnsavedChanges } from '../../hooks/use-unsaved-changes'
import type { CampaignWorkspace, DerivedStatus, Locale } from './types'

export const contentLabels = { game: '游戏', crowdfunding: '众筹项目', tabletop: '桌面游戏', merchandise: '收藏周边', other: '其他' }
export const reasonLabels = { editorial: 'GoFurry 精选', sponsored: '推广', upcoming: '即将发售', new_release: '新作推荐', trending: '热度上升' }
export const statusLabels = { draft: '草稿', scheduled: '已排期', active: '展示中', ended: '已结束', paused: '已暂停', archived: '已归档' }
export const contentOptions = Object.entries(contentLabels).map(([value, label]) => ({ value, label }))
export const localeOptions = [{ value: 'zh', label: '中文' }, { value: 'en', label: 'English' }]
export const primaryOptions = [{ value: '', label: '未配置' }, { value: 'game', label: '查看游戏' }, { value: 'project', label: '查看项目' }, { value: 'product', label: '查看产品' }, { value: 'website', label: '访问官网' }]
export const secondaryOptions = [{ value: '', label: '无' }, { value: 'steam', label: 'Steam' }, { value: 'kickstarter', label: 'Kickstarter' }, { value: 'website', label: '官方网站' }, { value: 'other', label: '其他外部链接' }]
export function CampaignStatus({ value }: { value: DerivedStatus }) {
  return <StatusBadge tone={value === 'active' ? 'success' : value === 'scheduled' ? 'info' : value === 'paused' ? 'warning' : 'neutral'}>{statusLabels[value] ?? value}</StatusBadge>
}
export function PublishedNotice({ workspace }: { workspace: CampaignWorkspace }) {
  return workspace.state === 'published' ? <Alert>当前活动已经发布，保存后将从后续 Showcase 请求开始生效。</Alert> : null
}
export type EditorProps = { workspace: CampaignWorkspace; canWrite: boolean; onBusyChange: (busy: boolean) => void }
export function useEditorGuard(dirty: boolean, busy: boolean, report: (busy: boolean) => void) {
  useUnsavedChanges(dirty || busy)
  useEffect(() => { report(dirty || busy); return () => report(false) }, [dirty, busy, report])
}
export function useShowcaseURL() {
  const [params, setParams] = useSearchParams()
  const update = (patch: Record<string, string | null>) => {
    const next = new URLSearchParams(params)
    Object.entries(patch).forEach(([key, value]) => { if (value) next.set(key, value); else next.delete(key) })
    setParams(next)
  }
  const locale: Locale = params.get('locale') === 'en' ? 'en' : 'zh'
  const rawPage = Number(params.get('page_num'))
  return { params, update, locale, page: Number.isSafeInteger(rawPage) && rawPage > 0 && rawPage <= 1000000 ? rawPage : 1, size: [20, 50, 100].includes(Number(params.get('page_size'))) ? Number(params.get('page_size')) : 20 }
}
export function LocaleSelect({ value, onChange }: { value: Locale; onChange: (value: string) => void }) {
  return <div className="w-40"><Select ariaLabel="展示语言" value={value} onValueChange={onChange} options={localeOptions} /></div>
}
