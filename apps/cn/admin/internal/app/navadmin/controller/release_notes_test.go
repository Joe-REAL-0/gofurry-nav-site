package controller

import (
	"github.com/gofurry/gofurry-admin/internal/app/navadmin/models"
	"strings"
	"testing"
	"time"
)

func TestReleaseDraftValidationAndMetadataNormalization(t *testing.T) {
	version, sha := "  September preview / 10  ", "  ABCDEF0123  "
	req, at, err := normalizeUpdateNoticePayload(models.UpdateNoticePayload{Version: &version, CommitSHA: &sha})
	if err != nil || !at.IsZero() || *req.Version != "September preview / 10" || *req.CommitSHA != "abcdef0123" {
		t.Fatalf("draft: %+v %v %v", req, at, err)
	}
	blank := "  "
	req, _, err = normalizeUpdateNoticePayload(models.UpdateNoticePayload{Version: &blank, CommitSHA: &blank})
	if err != nil || req.Version != nil || req.CommitSHA != nil {
		t.Fatal("optional metadata must normalize to null")
	}
	for _, sha := range []string{"abc123", "xyz1234", strings.Repeat("a", 65)} {
		if _, _, err := normalizeUpdateNoticePayload(models.UpdateNoticePayload{CommitSHA: &sha}); err == nil {
			t.Fatalf("accepted invalid SHA: %q", sha)
		}
	}
	long := strings.Repeat("v", 65)
	if _, _, err := normalizeUpdateNoticePayload(models.UpdateNoticePayload{Version: &long}); err == nil {
		t.Fatal("accepted oversized version")
	}
	if _, _, err := normalizeUpdateNoticePayload(models.UpdateNoticePayload{PublishedAt: "nonsense"}); err == nil {
		t.Fatal("accepted invalid date")
	}
	for _, tc := range []struct{ title, body string }{{"", "Body"}, {"Title", " "}} {
		if validateUpdateNoticePublication(tc.title, tc.body) == nil {
			t.Fatal("publish requires Chinese title/body")
		}
	}
	if err := validateUpdateNoticePublication("标题", "正文"); err != nil {
		t.Fatal(err)
	}
}

func TestReleasePublicationTimeUsesChinaWallTime(t *testing.T) {
	for _, input := range []string{"2026-09-28T10:30:00+08:00", "2026-09-28T02:30:00Z", "2026-09-28 10:30:00", "2026-09-28T10:30"} {
		at, err := parseReleasePublicationTime(input)
		if err != nil || at.Format("2006-01-02 15:04:05") != "2026-09-28 10:30:00" {
			t.Fatalf("%s: %v %v", input, at, err)
		}
		_, offset := at.Zone()
		if offset != 8*int(time.Hour/time.Second) {
			t.Fatal("timezone")
		}
	}
}
