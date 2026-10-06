package models

import "time"

// CollectionRecord and CollectionMembership are internal batch projections, not public entities.
type CollectionRecord struct {
	ID                               int64
	Code, Name, NameEn, Info, InfoEn string
	PublishedAt                      time.Time
	Slot                             int16
}

type CollectionMembership struct {
	CollectionID int64
	GameID       int64
}

type CollectionGames struct {
	Memberships []CollectionMembership
	Games       []GameV2Aggregate
}

type CollectionQuery struct {
	Lang, Mode     string
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
	GameID     string                `json:"game_id"`
	Name       string                `json:"name"`
	Summary    string                `json:"summary"`
	HeaderURL  string                `json:"header_url"`
	Phase      string                `json:"phase"`
	Chronology *CollectionChronology `json:"chronology"`
}

type CollectionDetail struct {
	CollectionMetadata
	Collection CollectionInfo           `json:"collection"`
	Items      []CollectionTimelineItem `json:"items"`
}
