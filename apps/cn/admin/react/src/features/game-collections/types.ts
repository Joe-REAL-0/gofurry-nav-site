export type CollectionStatus = 'draft' | 'published' | 'archived'
export type CollectionContent = { name: string; name_en: string; info: string; info_en: string }
export type GameCollection = CollectionContent & {
  id: number; code: string; status: CollectionStatus; version: number
  published_at: string | null; archived_at: string | null; created_at: string; updated_at: string
  member_count: number; sfw_member_count: number; home_slot: number | null
}
export type CollectionMember = { game_id: number; name: string; name_en: string; appid: number; adult: boolean }
export type CollectionMembers = { collection_id: number; version: number; members: CollectionMember[] }
export type CollectionWorkspace = { collection: GameCollection; members: CollectionMember[] }
export type CollectionHome = { revision: string; slots: { slot: number; collection: GameCollection | null }[] }
export const collectionStatusLabels: Record<CollectionStatus, string> = { draft: '草稿', published: '已发布', archived: '已归档' }
export const PUBLIC_REFRESH_NOTICE = '已保存；公开内容将在缓存刷新后更新。'
