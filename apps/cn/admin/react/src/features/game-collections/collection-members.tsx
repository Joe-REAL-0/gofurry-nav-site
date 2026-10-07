import { useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { RemoteSelect } from '../../components/admin/operations'
import { StatusBadge } from '../../components/admin/status'
import { Alert } from '../../components/ui/alert'
import { Button } from '../../components/ui/button'
import { Input } from '../../components/ui/input'
import { useCompositionSafeSearch } from '../../hooks/use-composition-safe-search'
import { errorMessage, getJSON } from '../../lib/api'
import type { GameWorkspace } from '../../lib/types'
import type { CollectionMember } from './types'

export function CollectionMembersEditor({ members, onChange, disabled, onLoadingChange }: { members: CollectionMember[]; onChange: (members: CollectionMember[]) => void; disabled: boolean; onLoadingChange: (loading: boolean) => void }) {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [keyword, setKeyword] = useState('')
  const [page, setPage] = useState(1)
  const search = useCompositionSafeSearch({ value: keyword, onCommit: value => { setKeyword(value); setPage(1) } })
  const filteredMembers = useMemo(() => {
    const query = keyword.trim().toLowerCase()
    return members.filter(member => [member.name, member.name_en, String(member.appid)].some(value => value.toLowerCase().includes(query)))
  }, [members, keyword])
  const pageSize = 20
  const pageCount = Math.max(1, Math.ceil(filteredMembers.length / pageSize))
  const currentPage = Math.min(page, pageCount)
  const pagedMembers = filteredMembers.slice((currentPage - 1) * pageSize, currentPage * pageSize)
  if (page > pageCount) setPage(pageCount)
  return <div className="grid gap-3">
    <p className="text-sm text-muted-foreground">游戏顺序由首次可用时间与发行事实自动计算，当前成员列表顺序不会影响公开时间线。</p>
    {!disabled && <RemoteSelect endpoint="/api/v1/options/games" value={null} placeholder="搜索并添加游戏…" pageSize={20} debounceMs={300} disabled={loading} excludeIDs={members.map(m => String(m.game_id))} onChange={async option => {
      if (!option || members.some(m => String(m.game_id) === option.id)) return
      setLoading(true); onLoadingChange(true); setError('')
      try {
        // Existing Game workspace owns the tag code and archived-tag semantics.
        const { game, tags } = await getJSON<GameWorkspace>(`/api/v1/game/games/${option.id}/workspace`)
        onChange([...members, { game_id: Number(game.id), name: game.name, name_en: game.name_en, appid: game.appid, adult: tags.some(tag => tag.code === 'adult') }])
      } catch (err) { setError(errorMessage(err)) }
      finally { setLoading(false); onLoadingChange(false) }
    }} />}
    <Input aria-label="搜索已收录游戏" placeholder="搜索已收录游戏…" {...search.inputProps} />
    {error && <Alert tone="danger">{error}</Alert>}
    {members.length === 0 ? <p className="text-sm text-muted-foreground">尚未收录游戏</p> : filteredMembers.length === 0 ? <p className="text-sm text-muted-foreground">未找到匹配的已收录游戏</p> : <ul aria-label="已收录游戏" className="divide-y">{pagedMembers.map(member => <li key={member.game_id} className="flex flex-wrap items-center gap-3 py-3">
      <Link className="min-w-0 flex-1 text-primary" to={`/game/games/${member.game_id}`}>{member.name || member.name_en}</Link>
      {member.adult && <StatusBadge>Adult</StatusBadge>}
      {!disabled && <Button variant="ghost" disabled={loading} aria-label={`移除 ${member.name}`} onClick={() => onChange(members.filter(m => m.game_id !== member.game_id))}>移除</Button>}
    </li>)}</ul>}
    <div className="flex flex-wrap items-center justify-between gap-3 text-xs text-muted-foreground" aria-label="成员分页">
      <span>显示 {filteredMembers.length === 0 ? 0 : (currentPage - 1) * pageSize + 1}–{Math.min(currentPage * pageSize, filteredMembers.length)}，共 {filteredMembers.length} 个</span>
      <div className="flex shrink-0 items-center gap-2">
        <Button variant="secondary" size="icon" aria-label="成员上一页" disabled={currentPage <= 1} onClick={() => setPage(currentPage - 1)}>‹</Button>
        <span>{currentPage} / {pageCount}</span>
        <Button variant="secondary" size="icon" aria-label="成员下一页" disabled={currentPage >= pageCount} onClick={() => setPage(currentPage + 1)}>›</Button>
      </div>
    </div>
  </div>
}
