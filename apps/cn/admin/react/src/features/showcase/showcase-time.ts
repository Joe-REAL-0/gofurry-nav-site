const offset = 8 * 60 * 60 * 1000
export const operatingTimezone = '运营时区：Asia/Shanghai (UTC+08:00)'

export function chinaWallTimeToRFC3339(value: string): string {
  if (!/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}$/.test(value)) throw new Error('请输入完整的上海日期和时间')
  const canonical = new Date(`${value}:00Z`)
  if (!Number.isFinite(canonical.getTime()) || canonical.toISOString().slice(0, 16) !== value) throw new Error('日期或时间无效')
  return `${value}:00+08:00`
}
export function rfc3339ToChinaWallTime(value: string | null): string {
  if (!value) return ''
  // Never interpret an unzoned server value using the operator's timezone.
  if (!/(Z|[+-]\d{2}:\d{2})$/i.test(value)) return ''
  const timestamp = Date.parse(value)
  return Number.isFinite(timestamp) ? new Date(timestamp + offset).toISOString().slice(0, 16) : ''
}
export const chinaDate = (now = new Date()) => new Date(now.getTime() + offset).toISOString().slice(0, 10)
export const dateBefore = (date: string, days: number) => new Date(Date.parse(`${date}T00:00:00Z`) - days * 86400000).toISOString().slice(0, 10)
export function validDateRange(from: string, to: string) {
  const valid = (value: string) => /^\d{4}-\d{2}-\d{2}$/.test(value) && Number.isFinite(Date.parse(`${value}T00:00:00Z`)) && new Date(`${value}T00:00:00Z`).toISOString().slice(0, 10) === value
  return valid(from) && valid(to) && from <= to && Date.parse(`${to}T00:00:00Z`) - Date.parse(`${from}T00:00:00Z`) <= 365 * 86400000
}
export function statsRange(preset: '7' | '30' | 'all', starts: string | null, today = chinaDate()) {
  const start = rfc3339ToChinaWallTime(starts).slice(0, 10)
  const from = preset === 'all' && start ? [dateBefore(today, 365), start > today ? today : start].sort().at(-1)! : dateBefore(today, preset === '7' ? 6 : 29)
  return { from, to: today }
}
export const displayTime = (value: string | null) => rfc3339ToChinaWallTime(value).replace('T', ' ') || '未设置'
