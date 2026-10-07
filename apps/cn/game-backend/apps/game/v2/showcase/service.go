package showcase

import (
	"context"
	"encoding/json"
	"errors"
	"regexp"
	"strconv"
	"time"

	v2models "github.com/gofurry/gofurry-game-backend/apps/game/v2/models"
	"github.com/redis/go-redis/v9"
)

type Service struct {
	Reader Reader
	Redis  redis.UniversalClient
	Now    func() time.Time
	gate   chan struct{}
}

func New(reader Reader, redisClient redis.UniversalClient) *Service {
	return &Service{Reader: reader, Redis: redisClient, Now: time.Now, gate: make(chan struct{}, 1)}
}

var regionPattern = regexp.MustCompile(`^[A-Z]{2}$`)

func ValidateScope(lang, region string) bool {
	return (lang == "zh" || lang == "en") && regionPattern.MatchString(region)
}
func (s *Service) Snapshot(ctx context.Context, lang, region string) (Snapshot, error) {
	if !ValidateScope(lang, region) {
		return Snapshot{}, errors.New("invalid Showcase locale or region")
	}
	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	now := s.Now().UTC()
	revision := "0"
	if s.Redis != nil {
		v, e := s.Redis.Get(ctx, RevisionKey).Result()
		if e == nil {
			if n, e := strconv.ParseUint(v, 10, 64); e == nil {
				revision = strconv.FormatUint(n, 10)
			}
		}
	}
	key := "game:v2:showcase:" + revision + ":" + lang + ":" + region + ":" + now.Format("2006-01-02T15")
	cached := func() (Snapshot, bool) {
		var snap Snapshot
		if s.Redis == nil {
			return snap, false
		}
		data, e := s.Redis.Get(ctx, key).Bytes()
		if e != nil || json.Unmarshal(data, &snap) != nil || snap.SchemaVersion != SchemaVersion || snap.Items == nil || len(snap.Items) > MaxItems || !s.Now().Before(snap.ValidUntil) || snap.ValidUntil.Sub(snap.GeneratedAt) > CacheTTL {
			return snap, false
		}
		return snap, true
	}
	if snap, ok := cached(); ok {
		return snap, nil
	}
	select {
	case s.gate <- struct{}{}:
		defer func() { <-s.gate }()
	case <-ctx.Done():
		return Snapshot{}, ctx.Err()
	}
	if snap, ok := cached(); ok {
		return snap, nil
	}
	now = s.Now().UTC()
	input, e := s.Reader.Read(ctx, now, lang, region)
	if e != nil {
		return Snapshot{}, e
	}
	snap := Compose(now, revision, lang, region, input, s.homeHint(ctx, lang, region))
	if s.Redis != nil {
		data, e := json.Marshal(snap)
		if e == nil {
			ttl := snap.ValidUntil.Sub(s.Now())
			if ttl > 0 {
				_ = s.Redis.Set(ctx, key, data, ttl).Err()
			}
		}
	}
	return snap, nil
}
func (s *Service) homeHint(ctx context.Context, lang, region string) HomeHint {
	hints := HomeHint{}
	if s.Redis == nil {
		return hints
	}
	data, e := s.Redis.Get(ctx, "game:v2:home:"+lang+":"+region).Bytes()
	if e != nil {
		return hints
	}
	var home v2models.GameV2HomeReadModel
	if json.Unmarshal(data, &home) != nil {
		return hints
	}
	groups := map[string][]v2models.GameV2ListItem{"upcoming": home.Panel.UpdatedGames, "new_release": home.Panel.LatestGames, "trending": home.Panel.PopularGames}
	for pool, items := range groups {
		hints[pool] = map[string]bool{}
		for i, item := range items {
			if i >= 8 {
				break
			}
			hints[pool][item.ID] = true
		}
	}
	return hints
}
