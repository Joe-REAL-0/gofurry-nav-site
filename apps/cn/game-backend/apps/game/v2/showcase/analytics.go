package showcase

import (
	"context"
	"errors"
	"net/netip"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
	_ "time/tzdata"

	"github.com/gofurry/gofurry-game-backend/common/log"
	gamesqlc "github.com/gofurry/gofurry-game-backend/internal/db/game/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redis/go-redis/v9"
)

const (
	SessionEventsPerMinute = 30
	IPEventsPerMinute      = 600
	AnalyticsRetention     = 72 * time.Hour
	DedupeRetention        = 48 * time.Hour
)

var businessZone = func() *time.Location {
	z, e := time.LoadLocation("Asia/Shanghai")
	if e != nil {
		panic(e)
	}
	return z
}()
var sessionPattern = regexp.MustCompile(`^[a-f0-9]{8}-[a-f0-9]{4}-4[a-f0-9]{3}-[89ab][a-f0-9]{3}-[a-f0-9]{12}$`)
var botIndicators = []string{"bot", "crawler", "spider", "slurp", "bingpreview", "headlesschrome", "phantomjs", "curl", "wget", "python-requests", "go-http-client"}

type Event struct {
	TrackingToken string `json:"tracking_token"`
	SessionID     string `json:"session_id"`
	Event         string `json:"event"`
	Source        string `json:"source,omitempty"`
}
type StatWriter interface {
	UpsertShowcaseDailyStat(context.Context, gamesqlc.UpsertShowcaseDailyStatParams) error
	UpsertShowcaseQuality(context.Context, gamesqlc.UpsertShowcaseQualityParams) error
}
type Analytics struct {
	Redis   redis.UniversalClient
	Writer  StatWriter
	Signer  Signer
	Origins []string
	Now     func() time.Time
}

func NewAnalytics(client redis.UniversalClient, writer StatWriter, signer Signer, origins []string) *Analytics {
	return &Analytics{client, writer, signer, origins, time.Now}
}
func ValidEventShape(e Event) bool {
	return len(e.TrackingToken) > 0 && len(e.TrackingToken) <= 4096 && sessionPattern.MatchString(e.SessionID) && ((e.Event == "impression" && e.Source == "") || (e.Event == "click" && (e.Source == "artwork" || e.Source == "title" || e.Source == "primary" || e.Source == "secondary")))
}
func originAllowed(origin string, origins []string) bool {
	parsed, e := url.Parse(origin)
	if e != nil || parsed.User != nil || parsed.Host == "" || parsed.Path != "" || parsed.RawQuery != "" || parsed.Fragment != "" || (parsed.Scheme != "https" && parsed.Scheme != "http") {
		return false
	}
	for _, v := range origins {
		if origin == strings.TrimSpace(v) {
			return true
		}
	}
	return false
}
func plausibleUA(ua string) bool {
	if len(ua) < 12 || len(ua) > 1024 || strings.ContainsAny(ua, "\r\n") {
		return false
	}
	ua = strings.ToLower(ua)
	for _, v := range botIndicators {
		if strings.Contains(ua, v) {
			return false
		}
	}
	return strings.Contains(ua, "mozilla/")
}

// Lua keeps dedupe and counter invariants atomic, including an implicit
// impression on the server's current business date when a click crosses midnight.
// Only signed subjects and ephemeral HMAC identities reach these keys.
const eventScript = `
local function quality(field)
 redis.call('HINCRBY',KEYS[3],field,1);redis.call('EXPIRE',KEYS[3],259200)
end
local sr=redis.call('INCR',KEYS[1]);redis.call('EXPIRE',KEYS[1],180)
local ir=redis.call('INCR',KEYS[2]);redis.call('EXPIRE',KEYS[2],180)
if sr>30 then quality('session_rate_limited');return 0 end
if ir>600 then quality('ip_rate_limited');return 0 end
if ARGV[1]~='' then quality(ARGV[1]);return 0 end
local imp=redis.call('GET',KEYS[4])
if ARGV[2]=='impression' and imp then quality('duplicate_impressions');return 0 end
if ARGV[2]=='click' and redis.call('EXISTS',KEYS[5])==1 then quality('duplicate_clicks');return 0 end
local fields={'valid_impressions','qualified_clicks','click_artwork','click_title','click_primary','click_secondary','impression_position_1','impression_position_2','impression_position_3','impression_position_4'}
for _,f in ipairs(fields) do local v=redis.call('HGET',KEYS[6],f);if v and not tonumber(v) then return redis.error_reply('invalid counter') end end
redis.call('SADD',KEYS[8],ARGV[6]);redis.call('EXPIRE',KEYS[8],259200)
redis.call('PFADD',KEYS[7],ARGV[5]);redis.call('EXPIRE',KEYS[7],259200)
redis.call('HSETNX',KEYS[6],'subject_kind',ARGV[7]);redis.call('HSETNX',KEYS[6],'campaign_id',ARGV[8])
redis.call('HSETNX',KEYS[6],'game_id',ARGV[9]);redis.call('HSETNX',KEYS[6],'reason',ARGV[10])
if ARGV[2]=='click' then redis.call('SET',KEYS[5],'1','EX',172800) end
if not imp or (ARGV[2]=='click' and imp~=ARGV[4]) then
 redis.call('SET',KEYS[4],ARGV[4],'EX',172800)
 redis.call('HINCRBY',KEYS[6],'valid_impressions',1)
 redis.call('HINCRBY',KEYS[6],'impression_position_'..ARGV[11],1)
end
if ARGV[2]=='click' then
 redis.call('HINCRBY',KEYS[6],'click_'..ARGV[3],1)
 redis.call('HINCRBY',KEYS[6],'qualified_clicks',1)
end
redis.call('EXPIRE',KEYS[6],259200)
return 1
`

func (a *Analytics) Submit(ctx context.Context, event Event, origin, ua, clientIP string) {
	if a == nil || !a.Signer.Ready() {
		return
	}
	if a.Redis == nil {
		log.Warn("Showcase analytics Redis unavailable; event discarded")
		return
	}
	if !ValidEventShape(event) {
		a.Malformed(ctx)
		return
	}
	now := a.Now()
	date := now.In(businessZone).Format(time.DateOnly)
	minute := strconv.FormatInt(now.Unix()/60, 10)
	ip, e := netip.ParseAddr(clientIP)
	if e != nil {
		a.Malformed(ctx)
		return
	}
	sessionHash := a.Signer.hash(event.SessionID)
	ipHash := a.Signer.hash(ip.Unmap().String() + "|" + date)
	claims, err := a.Signer.Verify(event.TrackingToken, now)
	reject := ""
	switch {
	case err != nil:
		reject = "invalid_token_events"
	case !originAllowed(origin, a.Origins):
		reject = "invalid_origin_events"
	case !plausibleUA(ua):
		reject = "filtered_user_agent_events"
	}
	mask := map[string]int{"artwork": ActionArtwork, "title": ActionTitle, "primary": ActionPrimary, "secondary": ActionSecondary}
	if reject == "" && event.Event == "click" && claims.Actions&mask[event.Source] == 0 {
		reject = "invalid_token_events"
	}
	prefix := "game:v2:showcase:"
	identity := claims.Snapshot + ":" + a.Signer.hash(claims.ItemKey) + ":" + sessionHash
	keys := []string{prefix + "rate:session:" + minute + ":" + sessionHash, prefix + "rate:ip:" + minute + ":" + ipHash, prefix + "quality:" + date,
		prefix + "dedupe:impression:" + identity, prefix + "dedupe:click:" + identity, prefix + "stat:" + date + ":" + claims.SubjectKey, prefix + "sessions:" + date + ":" + claims.SubjectKey, prefix + "subjects:" + date}
	bounded, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()
	if e := a.Redis.Eval(bounded, eventScript, keys, reject, event.Event, event.Source, date, sessionHash, claims.SubjectKey, claims.Kind, claims.CampaignID, claims.GameID, claims.Reason, strconv.Itoa(claims.Position)).Err(); e != nil {
		log.Warn("Showcase analytics Redis unavailable; event discarded")
	}
}
func (a *Analytics) Malformed(ctx context.Context) {
	if a == nil || a.Redis == nil {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, 300*time.Millisecond)
	defer cancel()
	key := "game:v2:showcase:quality:" + a.Now().In(businessZone).Format(time.DateOnly)
	if e := a.Redis.Eval(ctx, `redis.call('HINCRBY',KEYS[1],'malformed_events',1);redis.call('EXPIRE',KEYS[1],259200);return 1`, []string{key}).Err(); e != nil {
		log.Warn("Showcase analytics quality counter unavailable")
	}
}
func (a *Analytics) Aggregate(ctx context.Context) error {
	if a.Redis == nil || a.Writer == nil {
		return errors.New("Showcase aggregation unavailable")
	}
	now := a.Now().In(businessZone)
	for days := 0; days < 3; days++ {
		date := now.AddDate(0, 0, -days).Format(time.DateOnly)
		day, _ := time.Parse(time.DateOnly, date)
		pgdate := pgtype.Date{Time: day, Valid: true}
		prefix := "game:v2:showcase:"
		subjects, e := a.Redis.SMembers(ctx, prefix+"subjects:"+date).Result()
		if e != nil {
			return e
		}
		for _, subject := range subjects {
			values, e := a.Redis.HGetAll(ctx, prefix+"stat:"+date+":"+subject).Result()
			if e != nil {
				return e
			}
			if len(values) == 0 {
				continue
			}
			sessions, e := a.Redis.PFCount(ctx, prefix+"sessions:"+date+":"+subject).Result()
			if e != nil {
				return e
			}
			p, e := statParams(pgdate, subject, values, sessions)
			if e != nil {
				return e
			}
			if e = a.Writer.UpsertShowcaseDailyStat(ctx, p); e != nil {
				return e
			}
		}
		quality, e := a.Redis.HGetAll(ctx, prefix+"quality:"+date).Result()
		if e != nil {
			return e
		}
		if len(quality) == 0 {
			continue
		}
		n := func(k string) int64 { v, _ := strconv.ParseInt(quality[k], 10, 64); return v }
		for _, v := range quality {
			if _, e := strconv.ParseInt(v, 10, 64); e != nil {
				return errors.New("invalid Showcase quality counter")
			}
		}
		e = a.Writer.UpsertShowcaseQuality(ctx, gamesqlc.UpsertShowcaseQualityParams{StatDate: pgdate, InvalidTokenEvents: n("invalid_token_events"), InvalidOriginEvents: n("invalid_origin_events"), FilteredUserAgentEvents: n("filtered_user_agent_events"), DuplicateImpressions: n("duplicate_impressions"), DuplicateClicks: n("duplicate_clicks"), SessionRateLimited: n("session_rate_limited"), IpRateLimited: n("ip_rate_limited"), MalformedEvents: n("malformed_events")})
		if e != nil {
			return e
		}
	}
	return nil
}
func statParams(date pgtype.Date, subject string, values map[string]string, sessions int64) (gamesqlc.UpsertShowcaseDailyStatParams, error) {
	var malformed bool
	n := func(k string) int64 {
		if values[k] == "" {
			return 0
		}
		v, e := strconv.ParseInt(values[k], 10, 64)
		if e != nil || v < 0 {
			malformed = true
		}
		return v
	}
	id := func(k string) *int64 {
		if values[k] == "" {
			return nil
		}
		v := n(k)
		if v <= 0 {
			malformed = true
		}
		return &v
	}
	p := gamesqlc.UpsertShowcaseDailyStatParams{StatDate: date, SubjectKey: subject, SubjectKind: values["subject_kind"], CampaignID: id("campaign_id"), GameID: id("game_id"), Reason: values["reason"], ValidImpressions: n("valid_impressions"), QualifiedClicks: n("qualified_clicks"), ClickArtwork: n("click_artwork"), ClickTitle: n("click_title"), ClickPrimary: n("click_primary"), ClickSecondary: n("click_secondary"), ImpressionPosition1: n("impression_position_1"), ImpressionPosition2: n("impression_position_2"), ImpressionPosition3: n("impression_position_3"), ImpressionPosition4: n("impression_position_4"), SessionEstimate: sessions}
	if malformed || p.QualifiedClicks > p.ValidImpressions || p.QualifiedClicks != p.ClickArtwork+p.ClickTitle+p.ClickPrimary+p.ClickSecondary {
		return p, errors.New("invalid Showcase aggregate")
	}
	return p, nil
}
