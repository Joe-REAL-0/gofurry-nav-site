package service

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

// Exercise the real go-redis command boundary without a network dependency.
type limiterRedisHook struct {
	counts      map[string]int64
	expirations []int64
}

func (h *limiterRedisHook) DialHook(redis.DialHook) redis.DialHook {
	return func(context.Context, string, string) (net.Conn, error) {
		return nil, fmt.Errorf("unexpected Redis network dial")
	}
}
func (h *limiterRedisHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return next
}
func (h *limiterRedisHook) ProcessHook(_ redis.ProcessHook) redis.ProcessHook {
	return func(_ context.Context, cmd redis.Cmder) error {
		switch cmd.Name() {
		case "incr":
			key := cmd.Args()[1].(string)
			h.counts[key]++
			cmd.(*redis.IntCmd).SetVal(h.counts[key])
		case "expire":
			h.expirations = append(h.expirations, cmd.Args()[2].(int64))
			cmd.(*redis.BoolCmd).SetVal(true)
		default:
			return fmt.Errorf("unexpected Redis command: %s", cmd.Name())
		}
		return nil
	}
}

func TestSuggestionRedisLimiterAllows300PerTenMinuteWindow(t *testing.T) {
	client := redis.NewClient(&redis.Options{Addr: "unused.invalid:6379"})
	defer client.Close()
	hook := &limiterRedisHook{counts: map[string]int64{}}
	client.AddHook(hook)
	now := time.Date(2026, 10, 6, 0, 0, 0, 0, time.UTC)
	limiter := &redisSuggestionRateLimiter{client: client, now: func() time.Time { return now }}
	for i := 1; i <= 300; i++ {
		if allowed, _ := limiter.Allow("192.0.2.1"); !allowed {
			t.Fatalf("request %d rejected", i)
		}
	}
	if allowed, retry := limiter.Allow("192.0.2.1"); allowed || retry != 600 {
		t.Fatalf("request 301 allowed=%v retry=%d", allowed, retry)
	}
	if allowed, _ := limiter.Allow("192.0.2.2"); !allowed {
		t.Fatal("different client shared limit")
	}
	now = now.Add(10 * time.Minute)
	if allowed, _ := limiter.Allow("192.0.2.1"); !allowed {
		t.Fatal("next window blocked")
	}
	if len(hook.counts) != 3 || len(hook.expirations) != 3 {
		t.Fatal("Redis window/expiry contract lost")
	}
	for _, ttl := range hook.expirations {
		if ttl != 660 {
			t.Fatalf("window key TTL=%d", ttl)
		}
	}
}
