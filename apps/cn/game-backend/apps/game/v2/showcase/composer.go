package showcase

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"math"
	"sort"
	"time"
)

// WeightedOrder is a deterministic exponential race without replacement. Keys
// break floating-point ties, so input order and Go map iteration cannot change it.
func WeightedOrder(seed string, candidates []Candidate) []Candidate {
	out := append([]Candidate{}, candidates...)
	race := func(c Candidate) float64 {
		if c.Weight <= 0 {
			return math.Inf(1)
		}
		digest := sha256.Sum256([]byte(seed + "\x00" + c.Item.Key))
		u := (float64(binary.BigEndian.Uint64(digest[:8])) + 1) / 18446744073709551616.0
		return -math.Log(u) / float64(c.Weight)
	}
	sort.Slice(out, func(i, j int) bool {
		a, b := race(out[i]), race(out[j])
		if a == b {
			return out[i].Item.Key < out[j].Item.Key
		}
		return a < b
	})
	return out
}
func Compose(now time.Time, revision, lang, region string, input Inputs, hint HomeHint) Snapshot {
	now = now.UTC()
	hour, day := now.Format("2006-01-02T15"), now.Format(time.DateOnly)
	managedSeed := "showcase/v1/managed/" + lang + "/" + region + "/" + hour
	orderSeed := "showcase/v1/order/" + lang + "/" + region + "/" + hour
	result := Snapshot{SchemaVersion: SchemaVersion, GeneratedAt: now, ValidUntil: now.Add(CacheTTL), Items: []Item{}}
	for _, boundary := range []time.Time{now.Truncate(time.Hour).Add(time.Hour), input.Boundary} {
		if boundary.After(now) && boundary.Before(result.ValidUntil) {
			result.ValidUntil = boundary
		}
	}
	var slots [MaxItems]*Item
	used, usedGames := map[string]bool{}, map[string]bool{}
	managed, sponsored, total := 0, 0, 0
	pending := []Candidate{}
	canAdd := func(c Candidate) bool {
		return total < MaxItems && c.Weight > 0 && !used[c.Item.Key] && (c.Item.GameID == "" || !usedGames[c.Item.GameID]) && (!c.Item.Sponsored || sponsored < MaxSponsored)
	}
	add := func(c Candidate, pin bool) bool {
		if !canAdd(c) {
			return false
		}
		item := c.Item
		item.Tags = append([]string{}, item.Tags...)
		if pin {
			if c.Pin < 1 || c.Pin > MaxItems || slots[c.Pin-1] != nil {
				return false
			}
			slots[c.Pin-1] = &item
		} else {
			pending = append(pending, c)
		}
		used[item.Key] = true
		if item.GameID != "" {
			usedGames[item.GameID] = true
		}
		if item.Source == "managed" {
			managed++
		}
		if item.Sponsored {
			sponsored++
		}
		total++
		return true
	}
	pins, ordinary := []Candidate{}, []Candidate{}
	for _, c := range input.Candidates {
		if c.Item.Source == "managed" && c.Pin > 0 {
			pins = append(pins, c)
		} else {
			ordinary = append(ordinary, c)
		}
	}
	sort.Slice(pins, func(i, j int) bool {
		if pins[i].Pin == pins[j].Pin {
			return pins[i].Item.Key < pins[j].Item.Key
		}
		return pins[i].Pin < pins[j].Pin
	})
	for _, c := range pins {
		add(c, true)
	}
	managedLimit := MaxManaged
	if managed > managedLimit {
		managedLimit = managed
	}
	for _, wantSponsored := range []bool{true, false} {
		if !wantSponsored && managed >= managedLimit {
			break
		}
		if wantSponsored && sponsored > 0 {
			continue
		}
		editorial := false
		for _, p := range pins {
			if used[p.Item.Key] && !p.Item.Sponsored {
				editorial = true
			}
		}
		if !wantSponsored && editorial {
			continue
		}
		pool := []Candidate{}
		for _, c := range ordinary {
			if c.Item.Source == "managed" && c.Item.Sponsored == wantSponsored {
				pool = append(pool, c)
			}
		}
		for _, c := range WeightedOrder(managedSeed, pool) {
			if add(c, false) {
				// Several editorial pins can force the pinned balance exception
				// when the required Sponsored representative fills another slot.
				if managed > managedLimit {
					managedLimit = managed
				}
				break
			}
		}
	}
	representatives := []Candidate{}
	remaining := []Candidate{}
	for _, poolName := range []string{"upcoming", "new_release", "trending"} {
		pool, preferred := []Candidate{}, []Candidate{}
		for _, c := range ordinary {
			if c.Item.Source == "automatic" && c.Item.Reason == poolName && canAdd(c) {
				pool = append(pool, c)
				if !hint[poolName][c.Item.GameID] {
					preferred = append(preferred, c)
				}
			}
		}
		seed := "showcase/v1/automatic/" + poolName + "/" + lang + "/" + region + "/" + day
		pool = WeightedOrder(seed, pool)
		preferred = WeightedOrder(seed, preferred)
		if len(preferred) == 0 {
			preferred = pool
		}
		if len(preferred) > 0 {
			representatives = append(representatives, preferred[0])
		}
		remaining = append(remaining, pool...)
	}
	// Pool order is an equal-weight hourly permutation, independent of momentum.
	for _, c := range uniformOrder(orderSeed+"/pools", representatives) {
		add(c, false)
	}
	for _, c := range uniformOrder(orderSeed+"/backfill", remaining) {
		add(c, false)
	}
	if managed < managedLimit {
		for _, c := range WeightedOrder(managedSeed, ordinary) {
			if managed >= managedLimit {
				break
			}
			if c.Item.Source == "managed" {
				add(c, false)
			}
		}
	}
	pending = uniformOrder(orderSeed, pending)
	next := 0
	for i := range slots {
		if slots[i] == nil && next < len(pending) {
			item := pending[next].Item
			item.Tags = append([]string{}, item.Tags...)
			slots[i] = &item
			next++
		}
		if slots[i] != nil {
			slots[i].Position = i + 1
			result.Items = append(result.Items, *slots[i])
		}
	}
	keys := make([]string, 0, len(result.Items))
	for _, item := range result.Items {
		keys = append(keys, item.Key)
	}
	identity, _ := json.Marshal([]any{SchemaVersion, revision, lang, region, hour, day, keys})
	digest := sha256.Sum256(identity)
	result.SnapshotID = hex.EncodeToString(digest[:16])
	return result
}
func uniformOrder(seed string, items []Candidate) []Candidate {
	copy := append([]Candidate{}, items...)
	weights := map[string]int{}
	for i := range copy {
		weights[copy[i].Item.Key] = copy[i].Weight
		copy[i].Weight = 1
	}
	copy = WeightedOrder(seed, copy)
	for i := range copy {
		copy[i].Weight = weights[copy[i].Item.Key]
	}
	return copy
}
