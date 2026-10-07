package models

import "time"

const (
	SearchSuggestionsSchemaVersion = 2

	SearchSuggestionsStateReady       = "ready"
	SearchSuggestionsStateEmpty       = "empty"
	SearchSuggestionsStateUnavailable = "unavailable"

	SearchSuggestionsCacheHit  = "hit"
	SearchSuggestionsCacheMiss = "miss"
)

type SearchSuggestionsResponse struct {
	SchemaVersion int       `json:"schema_version"`
	GeneratedAt   time.Time `json:"generated_at"`
	State         string    `json:"state"`
	Query         string    `json:"query"`
	Suggestions   []string  `json:"suggestions"`
	CacheState    string    `json:"cache_state"`
}
