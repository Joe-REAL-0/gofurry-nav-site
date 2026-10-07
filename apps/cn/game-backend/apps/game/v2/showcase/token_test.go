package showcase

import (
	"strings"
	"testing"
	"time"
)

func testSigner() Signer { return Signer{strings.Repeat("t", 32), strings.Repeat("h", 32)} }
func signedTestSnapshot(t *testing.T) Snapshot {
	t.Helper()
	s := Compose(testNow, "0", "zh", "CN", Inputs{Candidates: []Candidate{{Item: Item{Key: "campaign:1", CampaignID: "1", Source: "managed", Reason: "editorial", PrimaryAction: Action{Type: "website", Target: "https://example.test"}, Tags: []string{}}, Weight: 100}}}, nil)
	signed, e := testSigner().SignSnapshot(s, "zh", "CN")
	if e != nil {
		t.Fatal(e)
	}
	return signed
}
func TestTrackingToken(t *testing.T) {
	s := testSigner()
	snap := signedTestSnapshot(t)
	token := snap.Items[0].TrackingToken
	c, e := s.Verify(token, testNow)
	if e != nil {
		t.Fatal(e)
	}
	for _, bad := range []string{"", token + "x", "x" + token, strings.Replace(token, ".", ".x", 1)} {
		if _, e := s.Verify(bad, testNow); e == nil {
			t.Fatal("tampered token accepted")
		}
	}
	if _, e := s.Verify(token, snap.ValidUntil.Add(TokenGrace)); e == nil {
		t.Fatal("expired accepted")
	}
	for _, change := range []func(*Claims){func(c *Claims) { c.Locale = "xx" }, func(c *Claims) { c.Region = "unknown" }, func(c *Claims) { c.Position = 5 }, func(c *Claims) { c.SubjectKey = "campaign:2" }, func(c *Claims) { c.GameID = "-1" }, func(c *Claims) { c.Actions = 0 }, func(c *Claims) { c.ExpiresAt = c.IssuedAt + int64((24*time.Hour)/time.Second) }} {
		copy := c
		change(&copy)
		if _, e := s.Verify(s.sign(copy), testNow); e == nil {
			t.Fatal("invalid binding accepted", copy)
		}
	}
	if _, e := (Signer{}).SignSnapshot(snap, "zh", "CN"); e == nil {
		t.Fatal("unsigned populated Showcase")
	}
}
func TestEventValidationAndPrivacy(t *testing.T) {
	e := Event{TrackingToken: "token", SessionID: "11111111-1111-4111-8111-111111111111", Event: "click", Source: "primary"}
	if !ValidEventShape(e) {
		t.Fatal(e)
	}
	for _, change := range []func(*Event){func(e *Event) { e.Event = "purchase" }, func(e *Event) { e.Source = "x" }, func(e *Event) { e.SessionID = "person" }, func(e *Event) { e.Event = "impression" }} {
		copy := e
		change(&copy)
		if ValidEventShape(copy) {
			t.Fatal("invalid event shape")
		}
	}
	if !originAllowed("https://example.test", []string{"https://example.test"}) {
		t.Fatal("valid Origin")
	}
	for _, o := range []string{"", "null", "https://evil.test", "https://example.test/path", "https://example.test@evil.test"} {
		if originAllowed(o, []string{"https://example.test"}) {
			t.Fatal("invalid Origin", o)
		}
	}
	for _, ua := range append([]string{"", "short"}, botIndicators...) {
		if plausibleUA("Mozilla/5.0 "+ua) && ua != "" && ua != "short" {
			t.Fatal("bot allowed", ua)
		}
	}
	if plausibleUA("") || !plausibleUA("Mozilla/5.0 Chrome/120.0 Safari/537.36") {
		t.Fatal("UA contract")
	}
	s := testSigner()
	if s.hash("192.0.2.1|2026-10-03") == s.hash("192.0.2.1|2026-10-04") || strings.Contains(s.hash(e.SessionID), e.SessionID) {
		t.Fatal("identity privacy")
	}
}
