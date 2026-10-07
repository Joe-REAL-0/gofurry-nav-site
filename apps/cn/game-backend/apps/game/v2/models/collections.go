package models

import "time"

// CollectionRecord and CollectionMembership are internal batch projections, not public entities.
type CollectionRecord struct {
	ID                               int64
	Code, Name, NameEn, Info, InfoEn string
	PublishedAt                      time.Time
	Slot                             int16
	VisibleGameCount                 int64 // Home-only count, without loading game aggregates.
}

type CollectionMembership struct {
	CollectionID int64
	GameID       int64
}

type CollectionGames struct {
	Memberships []CollectionMembership
	Games       []CollectionProjectionGame
}

// CollectionProjectionGame contains only the fields needed by collection cards
// and chronology. It intentionally cannot carry prices, metrics or detail panels.
type CollectionProjectionGame struct {
	Site           GameV2SiteRecord
	Details        *GfgGameV2Details
	Localized      *GfgGameV2LocalizedDetails
	Media          []GfgGameV2Media
	Assets         []GfgGameV2Asset
	Adult          bool
	FirstAvailable *GameV2FirstAvailable
	ReleaseState   *GameV2ReleaseState
}

type CollectionQuery struct {
	Lang, Mode     string
	Q, Phase, Sort string
	Page, PageSize int64
}

type CollectionMetadata struct {
	SchemaVersion int       `json:"schema_version"`
	GeneratedAt   time.Time `json:"generated_at"`
	AsOfDate      string    `json:"as_of_date"`
}

type CollectionInfo struct {
	Code             string    `json:"code"`
	Name             string    `json:"name"`
	Info             string    `json:"info"`
	VisibleGameCount int       `json:"visible_game_count"`
	PublishedAt      time.Time `json:"published_at"`
}

type CollectionPreviewGame struct {
	GameID    string `json:"game_id"`
	Name      string `json:"name"`
	HeaderURL string `json:"header_url"`
}

type CollectionSummary struct {
	CollectionInfo
	PreviewGames []CollectionPreviewGame `json:"preview_games"`
}

type CollectionHomeSlot struct {
	Slot       int16             `json:"slot"`
	Collection CollectionSummary `json:"collection"`
}

type CollectionHome struct {
	CollectionMetadata
	Slots []CollectionHomeSlot `json:"slots"`
}

type CollectionIndex struct {
	CollectionMetadata
	Page     int64               `json:"page"`
	PageSize int64               `json:"page_size"`
	Total    int64               `json:"total"`
	HasMore  bool                `json:"has_more"`
	Items    []CollectionSummary `json:"items"`
}

type CollectionChronology struct {
	Source      string `json:"source"`
	Precision   string `json:"precision"`
	WindowStart string `json:"window_start"`
	WindowEnd   string `json:"window_end"`
	Inferred    bool   `json:"inferred"`
}

type CollectionTimelineItem struct {
	CollectionTimelineDecoration
	GameID     string                `json:"game_id"`
	Name       string                `json:"name"`
	Summary    string                `json:"summary"`
	HeaderURL  string                `json:"header_url"`
	Phase      string                `json:"phase"`
	Chronology *CollectionChronology `json:"chronology"`
}

// Detail-only metadata; the core projection and Index previews do not load it.
type CollectionTimelineDecoration struct {
	PrimaryTag     *CollectionTimelineTag    `json:"primary_tag"`
	SecondaryTag   *CollectionTimelineTag    `json:"secondary_tag"`
	Rating         *CollectionTimelineRating `json:"rating"`
	Online         *CollectionTimelineOnline `json:"online"`
	CommunityCount int32                     `json:"community_count"`
}

type CollectionTimelineTag struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type CollectionTimelineRating struct {
	Average float64 `json:"average"`
	Count   int64   `json:"count"`
}

type CollectionTimelineOnline struct {
	Count       int64     `json:"count"`
	CollectedAt time.Time `json:"collected_at"`
}

type CollectionDetail struct {
	CollectionMetadata
	Collection CollectionInfo           `json:"collection"`
	Items      []CollectionTimelineItem `json:"items"`
}

func (q CollectionQuery) IsDefaultBrowse() bool {
	return q.Q == "" && q.Phase == "all" && q.Sort == "published_desc"
}
