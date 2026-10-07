package service

import (
	"sort"
	"strconv"
	"time"

	v2models "github.com/gofurry/gofurry-game-backend/apps/game/v2/models"
)

// resolveCollectionChronology consumes canonical facts only. asOfDate is supplied
// once by the request owner; release text and the wall clock are never parsed here.
func resolveCollectionChronology(first *v2models.GameV2FirstAvailable, release *v2models.GameV2ReleaseState, asOfDate time.Time) (string, *v2models.CollectionChronology) {
	if first != nil {
		return "released", &v2models.CollectionChronology{
			Source: "first_available", Precision: first.Precision,
			WindowStart: first.WindowStart, WindowEnd: first.WindowEnd, Inferred: first.Inferred,
		}
	}
	if release == nil || (release.Availability != "available" && release.Availability != "upcoming") {
		return "unknown", nil
	}
	chronology := collectionReleaseWindow(release)
	if release.Availability == "available" {
		if chronology == nil {
			return "released_unknown", nil
		}
		return "released", chronology
	}
	if chronology == nil {
		return "upcoming_tba", nil
	}
	if chronology.WindowEnd < asOfDate.UTC().Format(time.DateOnly) {
		return "upcoming_overdue", chronology
	}
	return "upcoming", chronology
}

func collectionReleaseWindow(release *v2models.GameV2ReleaseState) *v2models.CollectionChronology {
	switch release.Precision {
	case "day", "month", "quarter", "year":
	default:
		return nil
	}
	if release.WindowStart == nil || release.WindowEnd == nil {
		return nil
	}
	start, err := time.Parse(time.DateOnly, *release.WindowStart)
	if err != nil {
		return nil
	}
	end, err := time.Parse(time.DateOnly, *release.WindowEnd)
	if err != nil || end.Before(start) {
		return nil
	}
	return &v2models.CollectionChronology{Source: "release", Precision: release.Precision, WindowStart: *release.WindowStart, WindowEnd: *release.WindowEnd}
}

func collectionPhaseRank(phase string) int {
	switch phase {
	case "released":
		return 0
	case "released_unknown":
		return 1
	case "upcoming_overdue":
		return 2
	case "upcoming":
		return 3
	case "upcoming_tba":
		return 4
	default:
		return 5
	}
}

func sortCollectionTimeline(items []v2models.CollectionTimelineItem) {
	sort.Slice(items, func(i, j int) bool {
		a, b := items[i], items[j]
		if a.Phase != b.Phase {
			return collectionPhaseRank(a.Phase) < collectionPhaseRank(b.Phase)
		}
		if a.Chronology != nil && b.Chronology != nil {
			if a.Chronology.WindowStart != b.Chronology.WindowStart {
				return a.Chronology.WindowStart < b.Chronology.WindowStart
			}
			if a.Chronology.WindowEnd != b.Chronology.WindowEnd {
				return a.Chronology.WindowEnd < b.Chronology.WindowEnd
			}
		}
		aID, _ := strconv.ParseInt(a.GameID, 10, 64)
		bID, _ := strconv.ParseInt(b.GameID, 10, 64)
		return aID < bID
	})
}

func collectionPreview(items []v2models.CollectionTimelineItem) []v2models.CollectionPreviewGame {
	result := make([]v2models.CollectionPreviewGame, 0, 3)
	indices := []int{}
	switch len(items) {
	case 0:
		return result
	case 1:
		indices = []int{0}
	case 2:
		indices = []int{0, 1}
	default:
		indices = []int{0, len(items) / 2, len(items) - 1}
	}
	for _, index := range indices {
		item := items[index]
		result = append(result, v2models.CollectionPreviewGame{GameID: item.GameID, Name: item.Name, HeaderURL: item.HeaderURL})
	}
	return result
}
