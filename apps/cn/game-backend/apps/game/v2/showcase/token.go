package showcase

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"
)

const (
	ActionArtwork   = 1
	ActionTitle     = 2
	ActionPrimary   = 4
	ActionSecondary = 8
)

type Claims struct {
	Version    int    `json:"v"`
	Snapshot   string `json:"snapshot"`
	ItemKey    string `json:"item"`
	SubjectKey string `json:"subject"`
	Kind       string `json:"kind"`
	CampaignID string `json:"campaign,omitempty"`
	GameID     string `json:"game,omitempty"`
	Reason     string `json:"reason"`
	Position   int    `json:"position"`
	Locale     string `json:"locale"`
	Region     string `json:"region"`
	Actions    int    `json:"actions"`
	IssuedAt   int64  `json:"iat"`
	ExpiresAt  int64  `json:"exp"`
}
type Signer struct {
	TrackingSecret string
	HashSecret     string
}

var ErrTracking = errors.New("Showcase tracking is unavailable")

func (s Signer) Ready() bool {
	return len(s.TrackingSecret) >= 32 && len(s.HashSecret) >= 32 && s.TrackingSecret != s.HashSecret && !strings.Contains(s.TrackingSecret, "CHANGE_ME") && !strings.Contains(s.HashSecret, "CHANGE_ME")
}
func (s Signer) SignSnapshot(snapshot Snapshot, lang, region string) (Snapshot, error) {
	if len(snapshot.Items) == 0 {
		return snapshot, nil
	}
	if !s.Ready() {
		return Snapshot{}, ErrTracking
	}
	snapshot.Items = append([]Item{}, snapshot.Items...)
	for i, item := range snapshot.Items {
		mask := ActionArtwork | ActionTitle | ActionPrimary
		if item.SecondaryAction != nil {
			mask |= ActionSecondary
		}
		c := Claims{SchemaVersion, snapshot.SnapshotID, item.Key, item.Key, item.Source, item.CampaignID, item.GameID, item.Reason, item.Position, lang, region, mask, snapshot.GeneratedAt.Unix(), snapshot.ValidUntil.Add(TokenGrace).Unix()}
		if !validClaims(c) {
			return Snapshot{}, ErrTracking
		}
		snapshot.Items[i].TrackingToken = s.sign(c)
	}
	return snapshot, nil
}
func (s Signer) sign(c Claims) string {
	body, _ := json.Marshal(c)
	payload := base64.RawURLEncoding.EncodeToString(body)
	mac := hmac.New(sha256.New, []byte(s.TrackingSecret))
	mac.Write([]byte(payload))
	return payload + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
func (s Signer) Verify(token string, now time.Time) (Claims, error) {
	var c Claims
	if !s.Ready() || len(token) > 4096 {
		return c, ErrTracking
	}
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return c, ErrTracking
	}
	sig, e := base64.RawURLEncoding.DecodeString(parts[1])
	if e != nil {
		return c, ErrTracking
	}
	mac := hmac.New(sha256.New, []byte(s.TrackingSecret))
	mac.Write([]byte(parts[0]))
	if !hmac.Equal(sig, mac.Sum(nil)) {
		return c, ErrTracking
	}
	body, e := base64.RawURLEncoding.DecodeString(parts[0])
	if e != nil || json.Unmarshal(body, &c) != nil || !validClaims(c) || now.Unix() < c.IssuedAt || now.Unix() >= c.ExpiresAt {
		return Claims{}, ErrTracking
	}
	return c, nil
}
func validClaims(c Claims) bool {
	if c.Version != SchemaVersion || len(c.Snapshot) != 32 || !ValidateScope(c.Locale, c.Region) || c.Position < 1 || c.Position > MaxItems || c.ItemKey != c.SubjectKey || c.Actions&7 != 7 || c.Actions&^15 != 0 || c.ExpiresAt <= c.IssuedAt || c.ExpiresAt-c.IssuedAt > int64((CacheTTL+TokenGrace)/time.Second) {
		return false
	}
	if _, e := hex.DecodeString(c.Snapshot); e != nil {
		return false
	}
	positive := func(s string) bool {
		v, e := strconv.ParseInt(s, 10, 64)
		return e == nil && v > 0 && strconv.FormatInt(v, 10) == s
	}
	if c.GameID != "" && !positive(c.GameID) {
		return false
	}
	if c.Kind == "managed" {
		return positive(c.CampaignID) && c.ItemKey == "campaign:"+c.CampaignID && (c.Reason == "editorial" || c.Reason == "sponsored")
	}
	return c.Kind == "automatic" && c.CampaignID == "" && positive(c.GameID) && (c.Reason == "upcoming" || c.Reason == "new_release" || c.Reason == "trending") && c.ItemKey == "auto:"+c.Reason+":"+c.GameID
}
func (s Signer) hash(value string) string {
	mac := hmac.New(sha256.New, []byte(s.HashSecret))
	mac.Write([]byte(value))
	return hex.EncodeToString(mac.Sum(nil)[:16])
}
