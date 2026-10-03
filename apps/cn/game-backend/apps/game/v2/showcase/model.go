// Package showcase owns the optional Games Home discovery slice. It never
// rebuilds the catalog Home cache and never creates automatic Campaign rows.
package showcase

import "time"

const (
	SchemaVersion = 1
	MaxItems      = 4
	MaxSponsored  = 1
	MaxManaged    = 2
	CacheTTL      = 5 * time.Minute
	TokenGrace    = 2 * time.Hour
	RevisionKey   = "game:v2:showcase:revision"
)

type Action struct {
	Type   string `json:"type"`
	GameID string `json:"game_id,omitempty"`
	Target string `json:"target,omitempty"`
}
type Artwork struct {
	Kind             string   `json:"kind"`
	URL              string   `json:"url,omitempty"`
	DesktopObjectKey string   `json:"desktop_object_key,omitempty"`
	MobileObjectKey  string   `json:"mobile_object_key,omitempty"`
	FocalX           *float64 `json:"focal_x,omitempty"`
	FocalY           *float64 `json:"focal_y,omitempty"`
}
type Release struct {
	Availability string  `json:"availability"`
	Precision    string  `json:"precision"`
	ExactDate    *string `json:"exact_date"`
	Year         *int32  `json:"year"`
	Month        *int32  `json:"month"`
	Quarter      *int32  `json:"quarter"`
	WindowStart  *string `json:"window_start"`
	WindowEnd    *string `json:"window_end"`
}
type Item struct {
	Key             string   `json:"key"`
	Source          string   `json:"source"`
	Reason          string   `json:"reason"`
	ContentType     string   `json:"content_type"`
	Sponsored       bool     `json:"sponsored"`
	CampaignID      string   `json:"campaign_id,omitempty"`
	GameID          string   `json:"game_id,omitempty"`
	Title           string   `json:"title"`
	Summary         string   `json:"summary"`
	Tags            []string `json:"tags"`
	EditorialNote   string   `json:"editorial_note,omitempty"`
	Artwork         Artwork  `json:"artwork"`
	PrimaryAction   Action   `json:"primary_action"`
	SecondaryAction *Action  `json:"secondary_action,omitempty"`
	Release         *Release `json:"release,omitempty"`
	Position        int      `json:"position"`
	TrackingToken   string   `json:"tracking_token"`
}
type Snapshot struct {
	SchemaVersion int       `json:"schema_version"`
	SnapshotID    string    `json:"snapshot_id"`
	GeneratedAt   time.Time `json:"generated_at"`
	ValidUntil    time.Time `json:"valid_until"`
	Items         []Item    `json:"items"`
}
type Candidate struct {
	Item   Item `json:"item"`
	Weight int  `json:"weight"`
	Pin    int  `json:"pin_position,omitempty"`
}
type Diagnostic struct {
	GameID           string   `json:"game_id"`
	Name             string   `json:"name"`
	ShowcaseEligible bool     `json:"showcase_eligible"`
	SFW              bool     `json:"sfw"`
	LocaleReady      bool     `json:"locale_ready"`
	ArtworkReady     bool     `json:"artwork_ready"`
	Release          *Release `json:"release"`
	Candidate        bool     `json:"candidate"`
	ExcludedReasons  []string `json:"excluded_reasons"`
	Trending         *Trend   `json:"trending,omitempty"`
}
type Inputs struct {
	Candidates  []Candidate
	Diagnostics map[string][]Diagnostic
	Boundary    time.Time
}
type HomeHint map[string]map[string]bool
