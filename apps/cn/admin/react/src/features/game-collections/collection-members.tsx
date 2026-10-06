import { useState } from 'react'
import { Link } from 'react-router-dom'
import { RemoteSelect } from '../../components/admin/operations'
import { StatusBadge } from '../../components/admin/status'
import { Alert } from '../../components/ui/alert'
import { Button } from '../../components/ui/button'
import { errorMessage, getJSON } from '../../lib/api'
import type { GameWorkspace } from '../../lib/types'
import type { CollectionMember } from './types'

export function CollectionMembersEditor({ members, onChange, disabled, onLoadingChange }: { members: CollectionMember[]; onChange: (members: CollectionMember[]) => void; disabled: boolean; onLoadingChange: (loading: boolean) => void }) {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
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
    {error && <Alert tone="danger">{error}</Alert>}
    {members.length === 0 ? <p className="text-sm text-muted-foreground">尚未收录游戏</p> : <ul className="divide-y">{members.map(member => <li key={member.game_id} className="flex flex-wrap items-center gap-3 py-3">
      <Link className="min-w-0 flex-1 text-primary" to={`/game/games/${member.game_id}`}>{member.name || member.name_en}</Link>
      {member.adult && <StatusBadge>Adult</StatusBadge>}
      {!disabled && <Button variant="ghost" disabled={loading} aria-label={`移除 ${member.name}`} onClick={() => onChange(members.filter(m => m.game_id !== member.game_id))}>移除</Button>}
    </li>)}</ul>}
  </div>
}
