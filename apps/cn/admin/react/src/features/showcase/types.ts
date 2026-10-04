export type Locale = 'zh' | 'en'
export type ContentType = 'game' | 'crowdfunding' | 'tabletop' | 'merchandise' | 'other'
export type CampaignState = 'draft' | 'published' | 'paused' | 'archived'
export type DerivedStatus = 'draft' | 'scheduled' | 'active' | 'ended' | 'paused' | 'archived'
export type PrimaryType = 'game' | 'project' | 'product' | 'website'
export type SecondaryType = 'steam' | 'kickstarter' | 'website' | 'other'
export type Pool = 'upcoming' | 'new_release' | 'trending'
export type Lifecycle = 'publish' | 'pause' | 'resume' | 'archive' | 'delete'
export type CampaignLocale = { lang: Locale; enabled: boolean; title: string; summary: string; tags: string[]; editorial_note: string | null }
export type Campaign = {
  id: string; internal_name: string; content_type: ContentType; sponsored: boolean; linked_game_id: string | null
  state: CampaignState; derived_status: DerivedStatus; starts_at: string | null; ends_at: string | null
  weight: number; pin_position: number | null; desktop_object_key: string | null; mobile_object_key: string | null
  focal_x: number; focal_y: number; primary_action_type: PrimaryType | null; primary_target: string | null
  secondary_action_type: SecondaryType | null; secondary_target: string | null; created_at: string; updated_at: string
}
export type CampaignWorkspace = Campaign & { locales: CampaignLocale[]; ready_to_publish: boolean; publication_diagnostics: string[] }
export type ContentPayload = Pick<Campaign, 'internal_name' | 'content_type' | 'sponsored' | 'focal_x' | 'focal_y' | 'primary_action_type' | 'primary_target' | 'secondary_action_type' | 'secondary_target'> & { linked_game_id: number | null; locales: CampaignLocale[] }
export type SchedulePayload = Pick<Campaign, 'starts_at' | 'ends_at' | 'weight' | 'pin_position'>
export type Publication = { object_key: string; primary: string; mirror: string; warnings: string[] }
export type Release = { availability: string; precision: string; exact_date: string | null; year: number | null; month: number | null; quarter: number | null; window_start: string | null; window_end: string | null }
export type Action = { type: PrimaryType | SecondaryType; game_id?: string; target?: string }
export type ShowcaseItem = {
  key: string; source: 'managed' | 'automatic'; reason: 'editorial' | 'sponsored' | Pool; content_type: ContentType; sponsored: boolean
  campaign_id?: string; game_id?: string; title: string; summary: string; tags: string[]; editorial_note?: string
  artwork: { kind: 'steam' | 'managed'; url?: string; desktop_object_key?: string; mobile_object_key?: string; focal_x?: number; focal_y?: number }
  primary_action: Action; secondary_action?: Action; release?: Release; position: number; tracking_token: string
}
export type Composition = { schema_version: number; snapshot_id: string; generated_at: string; valid_until: string; items: ShowcaseItem[] }
export type TrendWindow = { observed_days: number; average: number; coverage: number | null }
export type Candidate = {
  game_id: string; name: string; showcase_eligible: boolean; sfw: boolean; locale_ready: boolean; artwork_ready: boolean
  release: Release | null; candidate: boolean; excluded_reasons: string[]
  trending?: { recent: TrendWindow; baseline: TrendWindow; momentum: number; eligible: boolean; weight: number }
  status: 'eligible' | 'pending_approval' | 'blocked'; first_available: string | null; pool_failures: string[]
}
export type CandidateStatus = Candidate['status'] | 'all'
export type CandidatePage = { total: number; items: Candidate[]; counts: Record<CandidateStatus, number> }
export type Counts = { valid_impressions: number; qualified_clicks: number; session_estimate: number; click_artwork: number; click_title: number; click_primary: number; click_secondary: number }
export type DailyStat = Counts & { stat_date: string; impression_position_1: number; impression_position_2: number; impression_position_3: number; impression_position_4: number }
export type Stats = { totals: Counts & { ctr: number; session_estimate_method: string }; daily: DailyStat[]; timezone: string }
export type QualityDay = { stat_date: string; invalid_token_events: number; invalid_origin_events: number; filtered_user_agent_events: number; duplicate_impressions: number; duplicate_clicks: number; session_rate_limited: number; ip_rate_limited: number; malformed_events: number }
