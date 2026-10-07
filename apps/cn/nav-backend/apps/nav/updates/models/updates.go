package models

import "time"

const (
	UpdatesSchemaVersion = 1

	UpdatesStateReady = "ready"
	UpdatesStateEmpty = "empty"
	UpdatesStateError = "error"

	TableNameGfnNavUpdateNotice = "gfn_nav_update_notice"
)

type UpdateNotice struct {
	Summary     string    `json:"summary"`
	Version     *string   `json:"version"`
	CommitSHA   *string   `json:"commit_sha"`
	SummaryEn   string    `json:"summary_en"`
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	TitleEn     string    `json:"title_en"`
	Body        string    `json:"body"`
	BodyEn      string    `json:"body_en"`
	PublishedAt time.Time `json:"published_at"`
	CreateTime  time.Time `json:"create_time"`
	UpdateTime  time.Time `json:"update_time"`
	Deleted     bool      `json:"deleted"`
}

func (*UpdateNotice) TableName() string {
	return TableNameGfnNavUpdateNotice
}

// Index is metadata only; full Markdown belongs to the detail projection.
type UpdateNoticeItem struct {
	Summary     string    `json:"summary"`
	Version     *string   `json:"version"`
	CommitSHA   *string   `json:"commit_sha"`
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	PublishedAt time.Time `json:"published_at"`
}

type UpdatesResponse struct {
	Page           int                `json:"page"`
	PageSize       int                `json:"page_size"`
	Total          int64              `json:"total"`
	HasMore        bool               `json:"has_more"`
	SchemaVersion  int                `json:"schema_version"`
	GeneratedAt    time.Time          `json:"generated_at"`
	State          string             `json:"state"`
	ReasonMessages []string           `json:"reason_messages,omitempty"`
	Items          []UpdateNoticeItem `json:"items"`
}

type UpdatePage struct{ Page, PageSize int }

// Detail uses an explicit public projection; no lifecycle or audit fields escape.
type ReleaseNote struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Summary     string    `json:"summary"`
	Body        string    `json:"body"`
	Version     *string   `json:"version"`
	CommitSHA   *string   `json:"commit_sha"`
	PublishedAt time.Time `json:"published_at"`
}
type ReleaseNoteNeighbor struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Version     *string   `json:"version"`
	PublishedAt time.Time `json:"published_at"`
}
type UpdateDetailResponse struct {
	SchemaVersion int                  `json:"schema_version"`
	GeneratedAt   time.Time            `json:"generated_at"`
	State         string               `json:"state"`
	Item          ReleaseNote          `json:"item"`
	Previous      *ReleaseNoteNeighbor `json:"previous"`
	Next          *ReleaseNoteNeighbor `json:"next"`
}
