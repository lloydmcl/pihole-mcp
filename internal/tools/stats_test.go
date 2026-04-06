package tools

import (
	"strings"
	"testing"
)

func TestStatsSummary_Normal(t *testing.T) {
	c := newTestClient(t, piholeHandler(map[string]any{
		"/stats/summary": map[string]any{
			"queries": map[string]any{
				"total": 45231, "blocked": 12847, "percent_blocked": 28.4,
				"cached": 18492, "forwarded": 13892, "unique_domains": 445,
				"frequency": 1.1, "types": map[string]any{"A": 30000, "AAAA": 10000},
				"status": map[string]any{}, "replies": map[string]any{},
			},
			"clients": map[string]any{"active": 23, "total": 30},
			"gravity": map[string]any{"domains_being_blocked": 92277, "last_update": 1712345678},
		},
	}))

	text := callTool(t, statsSummaryHandler, c, nil)
	if !strings.Contains(text, "45,231") {
		t.Errorf("expected formatted query count, got: %s", text)
	}
	if !strings.Contains(text, "28.4%") {
		t.Errorf("expected blocking percentage, got: %s", text)
	}
	if !strings.Contains(text, "92,277") {
		t.Errorf("expected gravity count, got: %s", text)
	}
}

func TestStatsSummary_Minimal(t *testing.T) {
	c := newTestClient(t, piholeHandler(map[string]any{
		"/stats/summary": map[string]any{
			"queries": map[string]any{
				"total": 100, "blocked": 20, "percent_blocked": 20.0,
				"cached": 50, "forwarded": 30, "unique_domains": 10,
				"frequency": 0.5, "types": map[string]any{}, "status": map[string]any{}, "replies": map[string]any{},
			},
			"clients": map[string]any{"active": 5, "total": 5},
			"gravity": map[string]any{"domains_being_blocked": 1000, "last_update": 0},
		},
	}))

	text := callTool(t, statsSummaryHandler, c, map[string]any{"detail": "minimal"})
	// Minimal should be a single line.
	if strings.Count(text, "\n") > 1 {
		t.Errorf("minimal should be single-line, got: %s", text)
	}
	if !strings.Contains(text, "100") {
		t.Errorf("expected query count in minimal, got: %s", text)
	}
}

func TestStatsSummary_Full(t *testing.T) {
	c := newTestClient(t, piholeHandler(map[string]any{
		"/stats/summary": map[string]any{
			"queries": map[string]any{
				"total": 100, "blocked": 20, "percent_blocked": 20.0,
				"cached": 50, "forwarded": 30, "unique_domains": 10,
				"frequency": 0.5,
				"types":     map[string]any{"A": 80, "AAAA": 20},
				"status":    map[string]any{"FORWARDED": 30, "CACHE": 50, "GRAVITY": 20},
				"replies":   map[string]any{},
			},
			"clients": map[string]any{"active": 5, "total": 5},
			"gravity": map[string]any{"domains_being_blocked": 1000, "last_update": 0},
		},
	}))

	text := callTool(t, statsSummaryHandler, c, map[string]any{"detail": "full"})
	if !strings.Contains(text, "Status breakdown") {
		t.Errorf("full should include status breakdown, got: %s", text)
	}
	if !strings.Contains(text, "Unique domains") {
		t.Errorf("full should include unique domains, got: %s", text)
	}
}

func TestStatsTopDomains_CSV(t *testing.T) {
	c := newTestClient(t, piholeHandler(map[string]any{
		"/stats/top_domains": map[string]any{
			"domains":         []any{map[string]any{"domain": "example.com", "count": 100}},
			"total_queries":   1000,
			"blocked_queries": 200,
		},
	}))

	text := callTool(t, statsTopDomainsHandler, c, map[string]any{"format": "csv"})
	if !strings.Contains(text, "Rank,Domain,Queries") {
		t.Errorf("CSV should have header row, got: %s", text)
	}
	if !strings.Contains(text, "example.com") {
		t.Errorf("CSV should contain domain, got: %s", text)
	}
}
