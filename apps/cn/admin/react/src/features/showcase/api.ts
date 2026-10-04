import type { QueryClient } from '@tanstack/react-query'
import { getJSON, sendForm, sendJSON } from '../../lib/api'
import type { PageResult } from '../../lib/types'
import type { Campaign, CampaignWorkspace, CandidatePage, Composition, ContentPayload, Lifecycle, Locale, Publication, QualityDay, SchedulePayload, Stats } from './types'

export const showcaseBase = '/api/v1/game/showcase'
export const campaignPath = (id: string) => `${showcaseBase}/campaigns/${encodeURIComponent(id)}`
export const showcaseAPI = {
  campaigns: (params: URLSearchParams) => getJSON<PageResult<Campaign>>(`${showcaseBase}/campaigns?${params}`),
  campaign: (id: string) => getJSON<CampaignWorkspace>(campaignPath(id)),
  create: (body: Pick<ContentPayload, 'internal_name' | 'content_type' | 'sponsored' | 'linked_game_id'>) => sendJSON<CampaignWorkspace>(`${showcaseBase}/campaigns`, 'POST', body),
  content: (id: string, body: ContentPayload) => sendJSON<CampaignWorkspace>(`${campaignPath(id)}/content`, 'PUT', body),
  schedule: (id: string, body: SchedulePayload) => sendJSON<CampaignWorkspace>(`${campaignPath(id)}/schedule`, 'PUT', body),
  lifecycle: async (id: string, action: Lifecycle): Promise<CampaignWorkspace | void> => action === 'delete' ? sendJSON<void>(campaignPath(id), 'DELETE') : sendJSON<CampaignWorkspace>(`${campaignPath(id)}/${action}`, 'POST'),
  upload: (id: string, variant: 'desktop' | 'mobile', file: File) => { const body = new FormData(); body.set('file', file); return sendForm<Publication>(`${campaignPath(id)}/artwork/${variant}`, body) },
  clear: (id: string, variant: 'desktop' | 'mobile') => sendJSON<void>(`${campaignPath(id)}/artwork/${variant}`, 'DELETE'),
  composition: (locale: Locale) => getJSON<Composition>(`${showcaseBase}/composition?lang=${locale}&region=CN`),
  candidates: (params: URLSearchParams) => getJSON<CandidatePage>(`${showcaseBase}/candidates?${params}`),
  stats: (id: string, from: string, to: string) => getJSON<Stats>(`${campaignPath(id)}/stats?${new URLSearchParams({ from, to })}`),
  quality: (from: string, to: string) => getJSON<{ daily: QualityDay[]; timezone: string }>(`${showcaseBase}/analytics/quality?${new URLSearchParams({ from, to })}`),
}
export const statsCSV = (id: string, from: string, to: string) => `${campaignPath(id)}/stats.csv?${new URLSearchParams({ from, to })}`

export async function invalidateShowcase(client: QueryClient, id?: string, gameID?: string | null) {
  await Promise.all([
    ...['campaigns', 'composition', 'candidates'].map(key => client.invalidateQueries({ queryKey: ['showcase', key] })),
    ...(id ? [client.invalidateQueries({ queryKey: ['showcase', 'campaign', id] })] : []),
    ...(gameID ? [client.invalidateQueries({ queryKey: ['game', Number(gameID)] })] : []),
  ])
}

// /content replaces the whole content contract, even when the UI only edits focal coordinates.
export function buildShowcaseContentPayload(workspace: CampaignWorkspace, patch: Partial<ContentPayload> = {}): ContentPayload {
  const body: ContentPayload = {
    internal_name: workspace.internal_name, content_type: workspace.content_type, sponsored: workspace.sponsored,
    linked_game_id: workspace.linked_game_id ? Number(workspace.linked_game_id) : null,
    focal_x: workspace.focal_x, focal_y: workspace.focal_y,
    primary_action_type: workspace.primary_action_type, primary_target: workspace.primary_target,
    secondary_action_type: workspace.secondary_action_type, secondary_target: workspace.secondary_target,
    locales: workspace.locales.map(locale => ({ ...locale, tags: [...locale.tags] })), ...patch,
  }
  if (body.sponsored) body.locales = body.locales.map(locale => ({ ...locale, editorial_note: null }))
  if (body.primary_action_type === 'game') body.primary_target = null
  if (!body.secondary_action_type) body.secondary_target = null
  return body
}
