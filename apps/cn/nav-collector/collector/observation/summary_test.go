package observation

import (
	"encoding/json"
	"slices"
	"testing"
	"time"

	"github.com/gofurry/gofurry-nav-collector/roof/env"
)

func TestSummaryKeys(t *testing.T) {
	if got := TargetLatestKey(ProtocolHTTP, 123, "www.example.com"); got != "collector:v2:latest:http:123:www.example.com" {
		t.Fatalf("TargetLatestKey() = %q", got)
	}
	if got := TargetSummaryKey(123, "www.example.com"); got != "collector:v2:summary:target:123:www.example.com" {
		t.Fatalf("TargetSummaryKey() = %q", got)
	}
	if got := SiteSummaryKey(123); got != "collector:v2:summary:site:123" {
		t.Fatalf("SiteSummaryKey() = %q", got)
	}
	if got := SiteSummaryTargetsKey(123); got != "collector:v2:summary:site_targets:123" {
		t.Fatalf("SiteSummaryTargetsKey() = %q", got)
	}
}

func TestTargetFromSummaryKey(t *testing.T) {
	key := TargetSummaryKey(123, "www.example.com")
	if got := targetFromSummaryKey(123, key); got != "www.example.com" {
		t.Fatalf("targetFromSummaryKey() = %q", got)
	}
	if got := targetFromSummaryKey(456, key); got != "" {
		t.Fatalf("targetFromSummaryKey(site mismatch) = %q, want empty", got)
	}
}

func TestBuildTargetSummaryHTTPHealthyPingFailureInformational(t *testing.T) {
	now := time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)
	docs := map[string]LatestDocument{
		ProtocolHTTP: latestDoc(ProtocolHTTP, StatusSuccess, now.Add(-time.Minute), map[string]any{
			"tls_handshake": "not_tls",
		}),
		ProtocolPing: latestDoc(ProtocolPing, StatusFailure, now.Add(-time.Minute), nil),
		ProtocolDNS:  latestDoc(ProtocolDNS, StatusSuccess, now.Add(-time.Minute), nil),
	}

	summary := BuildTargetSummary(1, "example.com", docs, now)
	if summary.Status != StatusHealthy {
		t.Fatalf("Status = %q, want healthy, reasons=%v", summary.Status, summary.ReasonCodes)
	}
	if len(summary.ReasonCodes) != 0 || len(summary.ReasonMessages) != 0 {
		t.Fatalf("informational Ping failure must not become a health reason: %+v", summary)
	}
	if summary.Protocols[ProtocolPing].Status != StatusFailure || docs[ProtocolPing].Status != StatusFailure {
		t.Fatal("Ping failure evidence must remain unchanged")
	}
}

func TestBuildTargetSummaryHTTPOnlyHealthyWhenOtherProtocolsDisabled(t *testing.T) {
	now := time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)
	oldV2 := env.GetServerConfig().Collector.V2
	env.GetServerConfig().Collector.V2 = env.CollectorV2Config{}
	t.Cleanup(func() {
		env.GetServerConfig().Collector.V2 = oldV2
	})

	summary := BuildTargetSummary(1, "example.com", map[string]LatestDocument{
		ProtocolHTTP: latestDoc(ProtocolHTTP, StatusSuccess, now.Add(-time.Minute), map[string]any{
			"tls_handshake": "not_tls",
		}),
	}, now)
	if summary.Status != StatusHealthy {
		t.Fatalf("Status = %q, want healthy, reasons=%v", summary.Status, summary.ReasonCodes)
	}
}

func TestBuildTargetSummaryHTTPAndDNSFailureDown(t *testing.T) {
	now := time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)
	docs := map[string]LatestDocument{
		ProtocolHTTP: latestDoc(ProtocolHTTP, StatusFailure, now.Add(-time.Minute), nil),
		ProtocolDNS:  latestDoc(ProtocolDNS, StatusFailure, now.Add(-time.Minute), nil),
	}

	summary := BuildTargetSummary(1, "example.com", docs, now)
	if summary.Status != StatusDown {
		t.Fatalf("Status = %q, want down, reasons=%v", summary.Status, summary.ReasonCodes)
	}
	if !contains(summary.ReasonCodes, "http_failed") || !contains(summary.ReasonCodes, "dns_failed") {
		t.Fatalf("missing down reasons: %v", summary.ReasonCodes)
	}
}

func TestBuildTargetSummaryStaleHTTPUnknown(t *testing.T) {
	now := time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)
	oldHTTPInterval := env.GetServerConfig().Collector.Request.RequestInterval
	env.GetServerConfig().Collector.Request.RequestInterval = 1
	t.Cleanup(func() {
		env.GetServerConfig().Collector.Request.RequestInterval = oldHTTPInterval
	})
	docs := map[string]LatestDocument{
		ProtocolHTTP: latestDoc(ProtocolHTTP, StatusSuccess, now.Add(-3*time.Hour), nil),
	}

	summary := BuildTargetSummary(1, "example.com", docs, now)
	if summary.Status != StatusUnknown {
		t.Fatalf("Status = %q, want unknown, reasons=%v", summary.Status, summary.ReasonCodes)
	}
	if !summary.Protocols[ProtocolHTTP].Stale {
		t.Fatal("HTTP protocol summary should be stale")
	}
}

func TestBuildTargetSummaryTLSExpiringWarning(t *testing.T) {
	now := time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)
	docs := map[string]LatestDocument{
		ProtocolHTTP: latestDoc(ProtocolHTTP, StatusSuccess, now.Add(-time.Minute), map[string]any{
			"tls_handshake":  "collected",
			"cert_verified":  true,
			"cert_not_after": now.Add(15 * 24 * time.Hour).Format(time.RFC3339),
		}),
	}

	summary := BuildTargetSummary(1, "example.com", docs, now)
	if summary.Status != StatusWarning {
		t.Fatalf("Status = %q, want warning, reasons=%v", summary.Status, summary.ReasonCodes)
	}
	if !contains(summary.ReasonCodes, "tls_cert_expiring_soon") {
		t.Fatalf("missing TLS expiring reason: %v", summary.ReasonCodes)
	}
}

func TestBuildTargetSummaryTLSVerifyFailureDegraded(t *testing.T) {
	now := time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)
	docs := map[string]LatestDocument{
		ProtocolHTTP: latestDoc(ProtocolHTTP, StatusSuccess, now.Add(-time.Minute), map[string]any{
			"tls_handshake":         "collected",
			"cert_verified":         false,
			"verify_error_category": "hostname_mismatch",
		}),
	}

	summary := BuildTargetSummary(1, "example.com", docs, now)
	if summary.Status != StatusDegraded {
		t.Fatalf("Status = %q, want degraded, reasons=%v", summary.Status, summary.ReasonCodes)
	}
	if !contains(summary.ReasonCodes, "tls_verify_hostname_mismatch") {
		t.Fatalf("missing TLS verify reason: %v", summary.ReasonCodes)
	}
}

func TestBuildTargetSummaryDNSRiskClassification(t *testing.T) {
	now := time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		name       string
		flags      []string
		wantStatus string
		wantCodes  []string
	}{
		{"ptr empty", []string{"ptr_empty"}, StatusHealthy, nil},
		{"low ttl", []string{"low_ttl"}, StatusHealthy, nil},
		{"unknown flag", []string{"future_risk"}, StatusHealthy, nil},
		{"private ip", []string{"private_ip"}, StatusWarning, []string{"dns_risk_private_ip"}},
		{"nxdomain with answer", []string{"nxdomain_with_answer"}, StatusWarning, []string{"dns_risk_nxdomain_with_answer"}},
		{"mixed signals", []string{"ptr_empty", "private_ip", "low_ttl", "future_risk"}, StatusWarning, []string{"dns_risk_private_ip"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			decodedFlags := make([]any, len(tc.flags))
			for i, flag := range tc.flags {
				decodedFlags[i] = flag
			}
			for _, flags := range []any{tc.flags, decodedFlags} {
				dnsPayload := map[string]any{"risk_flags": flags, "reverse_ptr": ""}
				before, _ := json.Marshal(dnsPayload)
				docs := map[string]LatestDocument{
					ProtocolHTTP: latestDoc(ProtocolHTTP, StatusSuccess, now.Add(-time.Minute), nil),
					ProtocolDNS:  latestDoc(ProtocolDNS, StatusSuccess, now.Add(-time.Minute), dnsPayload),
				}
				summary := BuildTargetSummary(1, "example.com", docs, now)
				if summary.Status != tc.wantStatus || !slices.Equal(summary.ReasonCodes, tc.wantCodes) {
					t.Fatalf("Status=%s reasons=%v, want %s %v", summary.Status, summary.ReasonCodes, tc.wantStatus, tc.wantCodes)
				}
				var wantMessages []string
				for _, code := range tc.wantCodes {
					definition, _ := ReasonDefinitionByCode(code)
					wantMessages = append(wantMessages, definition.MessageZH)
				}
				if !slices.Equal(summary.ReasonMessages, wantMessages) {
					t.Fatalf("health messages=%v, want %v", summary.ReasonMessages, wantMessages)
				}
				after, _ := json.Marshal(dnsPayload)
				if string(after) != string(before) {
					t.Fatalf("raw DNS evidence changed: %+v", dnsPayload)
				}
			}
		})
	}
	if got := dnsRiskReasonCode("future_risk"); got != "dns_risk_other" {
		t.Fatalf("unknown flags must retain the stable reason code, got %q", got)
	}
}

func TestBuildTargetSummaryUnknownTLSVerifyCategoryFallsBackToOther(t *testing.T) {
	now := time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)
	docs := map[string]LatestDocument{
		ProtocolHTTP: latestDoc(ProtocolHTTP, StatusSuccess, now.Add(-time.Minute), map[string]any{
			"tls_handshake":         "collected",
			"cert_verified":         false,
			"verify_error_category": "new_tls_error",
		}),
	}

	summary := BuildTargetSummary(1, "example.com", docs, now)
	if !contains(summary.ReasonCodes, "tls_verify_other") {
		t.Fatalf("missing fallback TLS reason: %v", summary.ReasonCodes)
	}
}

func TestBuildTargetSummaryPreservesHealthAffectingFailures(t *testing.T) {
	now := time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)
	oldV2 := env.GetServerConfig().Collector.V2
	env.GetServerConfig().Collector.V2 = env.CollectorV2Config{
		Enabled: true, LatestRedis: true, Protocols: env.CollectorProtocols{HTTP: true, DNS: true, Ping: true},
	}
	t.Cleanup(func() { env.GetServerConfig().Collector.V2 = oldV2 })
	for _, tc := range []struct {
		name       string
		httpStatus string
		dnsStatus  string
		stale      string
		payload    map[string]any
		wantStatus string
		wantCodes  []string
	}{
		{"DNS failed", StatusSuccess, StatusFailure, "", nil, StatusWarning, []string{"dns_failed_but_http_ok"}},
		{"DNS missing", StatusSuccess, "", "", nil, StatusWarning, []string{"dns_missing_or_stale"}},
		{"DNS stale", StatusSuccess, StatusSuccess, ProtocolDNS, nil, StatusWarning, []string{"dns_missing_or_stale"}},
		{"HTTP failed", StatusFailure, StatusSuccess, "", nil, StatusDegraded, []string{"http_failed"}},
		{"HTTP and DNS failed", StatusFailure, StatusFailure, "", nil, StatusDown, []string{"http_failed", "dns_failed"}},
		{"HTTP missing", "", StatusSuccess, "", nil, StatusUnknown, []string{"http_missing_or_stale"}},
		{"HTTP stale", StatusSuccess, StatusSuccess, ProtocolHTTP, nil, StatusUnknown, []string{"http_missing_or_stale"}},
		{"TLS verify failed", StatusSuccess, StatusSuccess, "", map[string]any{
			"tls_handshake": "collected", "cert_verified": false, "verify_error_category": "hostname_mismatch",
		}, StatusDegraded, []string{"tls_verify_hostname_mismatch"}},
		{"certificate expired", StatusSuccess, StatusSuccess, "", map[string]any{
			"tls_handshake": "collected", "cert_verified": true, "cert_not_after": now.Add(-48 * time.Hour).Format(time.RFC3339),
		}, StatusDegraded, []string{"tls_cert_expired"}},
		{"certificate 30 days", StatusSuccess, StatusSuccess, "", map[string]any{
			"tls_handshake": "collected", "cert_verified": true, "cert_not_after": now.Add(30 * 24 * time.Hour).Format(time.RFC3339),
		}, StatusWarning, []string{"tls_cert_expiring_soon"}},
		{"certificate 31 days", StatusSuccess, StatusSuccess, "", map[string]any{
			"tls_handshake": "collected", "cert_verified": true, "cert_not_after": now.Add(31 * 24 * time.Hour).Format(time.RFC3339),
		}, StatusHealthy, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			docs := map[string]LatestDocument{
				ProtocolPing: latestDoc(ProtocolPing, StatusFailure, now.Add(-time.Minute), nil),
			}
			if tc.httpStatus != "" {
				docs[ProtocolHTTP] = latestDoc(ProtocolHTTP, tc.httpStatus, now.Add(-time.Minute), tc.payload)
			}
			if tc.dnsStatus != "" {
				docs[ProtocolDNS] = latestDoc(ProtocolDNS, tc.dnsStatus, now.Add(-time.Minute), map[string]any{
					"risk_flags": []string{"ptr_empty", "low_ttl", "future_risk"},
				})
			}
			if tc.stale != "" {
				doc := docs[tc.stale]
				doc.ObservedAt = now.Add(-staleAfterForProtocol(tc.stale) - time.Minute)
				docs[tc.stale] = doc
			}
			summary := BuildTargetSummary(1, "example.com", docs, now)
			if summary.Status != tc.wantStatus || !slices.Equal(summary.ReasonCodes, tc.wantCodes) {
				t.Fatalf("Status=%s reasons=%v, want %s %v", summary.Status, summary.ReasonCodes, tc.wantStatus, tc.wantCodes)
			}
			if len(summary.ReasonMessages) != len(tc.wantCodes) {
				t.Fatalf("unexpected health reason messages: %v", summary.ReasonMessages)
			}
		})
	}
}

func TestBuildSiteSummaryInformationalSignalsDoNotWarn(t *testing.T) {
	now := time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)
	var targets []TargetSummaryDocument
	for _, target := range []struct {
		host       string
		flags      []string
		pingStatus string
	}{
		{"a.example.com", []string{"ptr_empty"}, StatusSuccess},
		{"b.example.com", []string{"low_ttl"}, StatusSuccess},
		{"c.example.com", nil, StatusFailure},
	} {
		summary := BuildTargetSummary(1, target.host, map[string]LatestDocument{
			ProtocolHTTP: latestDoc(ProtocolHTTP, StatusSuccess, now.Add(-time.Minute), nil),
			ProtocolDNS: latestDoc(ProtocolDNS, StatusSuccess, now.Add(-time.Minute), map[string]any{
				"risk_flags": target.flags,
			}),
			ProtocolPing: latestDoc(ProtocolPing, target.pingStatus, now.Add(-time.Minute), nil),
		}, now)
		if summary.Status != StatusHealthy || len(summary.ReasonCodes) != 0 || len(summary.ReasonMessages) != 0 {
			t.Fatalf("informational-only target must be healthy: %+v", summary)
		}
		targets = append(targets, summary)
	}
	summary := BuildSiteSummary(1, targets, now)
	if summary.Status != StatusHealthy || summary.StatusCounts[StatusHealthy] != 3 || summary.StatusCounts[StatusWarning] != 0 {
		t.Fatalf("informational-only site must be healthy: %+v", summary)
	}
	if len(summary.ReasonCodes) != 0 || len(summary.ReasonMessages) != 0 {
		t.Fatalf("informational-only site must have no health reasons: %+v", summary)
	}
	warning := BuildTargetSummary(1, "d.example.com", map[string]LatestDocument{
		ProtocolHTTP: latestDoc(ProtocolHTTP, StatusSuccess, now.Add(-time.Minute), nil),
		ProtocolDNS: latestDoc(ProtocolDNS, StatusSuccess, now.Add(-time.Minute), map[string]any{
			"risk_flags": []string{"private_ip", "ptr_empty"},
		}),
	}, now)
	summary = BuildSiteSummary(1, append(targets, warning), now)
	if summary.Status != StatusWarning || summary.StatusCounts[StatusWarning] != 1 || summary.StatusCounts[StatusHealthy] != 3 {
		t.Fatalf("one true warning must still warn at site level: %+v", summary)
	}
	if !slices.Equal(summary.ReasonCodes, []string{"some_targets_warning"}) {
		t.Fatalf("missing site health explanation: %v", summary.ReasonCodes)
	}
}

func TestBuildSiteSummaryAggregatesTargetsConservatively(t *testing.T) {
	now := time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)
	summary := BuildSiteSummary(1, []TargetSummaryDocument{
		{SiteID: 1, Target: "a.example.com", Status: StatusHealthy},
		{SiteID: 1, Target: "b.example.com", Status: StatusDown},
	}, now)

	if summary.Status != StatusDegraded {
		t.Fatalf("Status = %q, want degraded, reasons=%v", summary.Status, summary.ReasonCodes)
	}
	if summary.StatusCounts[StatusHealthy] != 1 || summary.StatusCounts[StatusDown] != 1 {
		t.Fatalf("status counts wrong: %+v", summary.StatusCounts)
	}

	allDown := BuildSiteSummary(1, []TargetSummaryDocument{
		{SiteID: 1, Target: "a.example.com", Status: StatusDown},
		{SiteID: 1, Target: "b.example.com", Status: StatusDown},
	}, now)
	if allDown.Status != StatusDown {
		t.Fatalf("all down status = %q, want down", allDown.Status)
	}
}

func TestReasonDefinitionsAreStableAndComplete(t *testing.T) {
	definitions := AllReasonDefinitions()
	if len(definitions) == 0 {
		t.Fatal("reason definitions should not be empty")
	}
	seen := map[string]bool{}
	for _, definition := range definitions {
		if definition.Code == "" || definition.MessageZH == "" || definition.DescriptionZH == "" {
			t.Fatalf("reason definition has empty required field: %+v", definition)
		}
		if seen[definition.Code] {
			t.Fatalf("duplicate reason definition code: %s", definition.Code)
		}
		seen[definition.Code] = true
		if !validReasonSeverity(definition.Severity) {
			t.Fatalf("invalid reason severity: %+v", definition)
		}
		if definition.Scope != ReasonScopeTarget && definition.Scope != ReasonScopeSite {
			t.Fatalf("invalid reason scope: %+v", definition)
		}
		if _, ok := ReasonDefinitionByCode(definition.Code); !ok {
			t.Fatalf("ReasonDefinitionByCode(%q) missing", definition.Code)
		}
	}
	for _, code := range []string{
		"http_missing_or_stale",
		"http_failed",
		"dns_failed",
		"dns_missing_or_stale",
		"dns_failed_but_http_ok",
		"ping_failed_but_http_ok",
		"dns_risk_private_ip",
		"dns_risk_low_ttl",
		"dns_risk_nxdomain_with_answer",
		"dns_risk_ptr_empty",
		"dns_risk_other",
		"tls_verify_expired",
		"tls_verify_not_yet_valid",
		"tls_verify_hostname_mismatch",
		"tls_verify_unknown_authority",
		"tls_verify_incompatible_usage",
		"tls_verify_other",
		"tls_cert_expired",
		"tls_cert_expiring_soon",
		"no_target_summary",
		"all_targets_down",
		"all_targets_unknown",
		"some_targets_degraded",
		"some_targets_warning",
	} {
		if _, ok := ReasonDefinitionByCode(code); !ok {
			t.Fatalf("expected reason definition %q", code)
		}
	}
}

func TestAuxiliaryReasonClassification(t *testing.T) {
	for _, tc := range []struct {
		code          string
		severity      string
		affectsHealth bool
	}{
		{"ping_failed_but_http_ok", ReasonSeverityInfo, false},
		{"dns_risk_ptr_empty", ReasonSeverityInfo, false},
		{"dns_risk_low_ttl", ReasonSeverityInfo, false},
		{"dns_risk_other", ReasonSeverityInfo, false},
		{"dns_risk_private_ip", ReasonSeverityWarning, true},
		{"dns_risk_nxdomain_with_answer", ReasonSeverityWarning, true},
		{"dns_failed_but_http_ok", ReasonSeverityWarning, true},
		{"dns_missing_or_stale", ReasonSeverityWarning, true},
	} {
		t.Run(tc.code, func(t *testing.T) {
			definition, ok := ReasonDefinitionByCode(tc.code)
			if !ok || definition.Severity != tc.severity || definition.AffectsHealth != tc.affectsHealth {
				t.Fatalf("classification=%+v, want severity=%s affects_health=%t", definition, tc.severity, tc.affectsHealth)
			}
		})
	}
}

func TestSummaryReasonCodesHaveDefinitions(t *testing.T) {
	now := time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)
	targetSummaries := []TargetSummaryDocument{
		BuildTargetSummary(1, "missing-http.example.com", map[string]LatestDocument{}, now),
		BuildTargetSummary(1, "http-dns-failed.example.com", map[string]LatestDocument{
			ProtocolHTTP: latestDoc(ProtocolHTTP, StatusFailure, now.Add(-time.Minute), nil),
			ProtocolDNS:  latestDoc(ProtocolDNS, StatusFailure, now.Add(-time.Minute), nil),
		}, now),
		BuildTargetSummary(1, "warning.example.com", map[string]LatestDocument{
			ProtocolHTTP: latestDoc(ProtocolHTTP, StatusSuccess, now.Add(-time.Minute), map[string]any{
				"tls_handshake":         "collected",
				"cert_verified":         false,
				"verify_error_category": "unknown_authority",
				"cert_not_after":        now.Add(15 * 24 * time.Hour).Format(time.RFC3339),
			}),
			ProtocolDNS: latestDoc(ProtocolDNS, StatusSuccess, now.Add(-time.Minute), map[string]any{
				"risk_flags": []any{"ptr_empty"},
			}),
			ProtocolPing: latestDoc(ProtocolPing, StatusFailure, now.Add(-time.Minute), nil),
		}, now),
	}
	for _, summary := range targetSummaries {
		assertReasonCodesDefined(t, summary.ReasonCodes)
	}

	siteSummary := BuildSiteSummary(1, targetSummaries, now)
	assertReasonCodesDefined(t, siteSummary.ReasonCodes)
}

func latestDoc(protocol string, status string, observedAt time.Time, payload any) LatestDocument {
	if payload == nil {
		payload = map[string]any{}
	}
	return LatestDocument{
		SiteID:        1,
		Target:        "example.com",
		Protocol:      protocol,
		Status:        status,
		ObservedAt:    observedAt,
		DurationMS:    123,
		Payload:       payload,
		SchemaVersion: schemaVersion,
	}
}

func validReasonSeverity(value string) bool {
	switch value {
	case ReasonSeverityInfo, ReasonSeverityWarning, ReasonSeverityDegraded, ReasonSeverityDown, ReasonSeverityUnknown:
		return true
	default:
		return false
	}
}

func assertReasonCodesDefined(t *testing.T, codes []string) {
	t.Helper()
	for _, code := range codes {
		if definition, ok := ReasonDefinitionByCode(code); !ok {
			t.Fatalf("reason code %q has no definition", code)
		} else if !definition.AffectsHealth || definition.Severity == ReasonSeverityInfo {
			t.Fatalf("summary reason must explain a health-affecting conclusion: %+v", definition)
		}
	}
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
