package tools

import (
	"strings"
	"testing"
)

func TestQueriesSearch_Normal(t *testing.T) {
	c := newTestClient(t, piholeHandler(map[string]any{
		"/queries": map[string]any{
			"queries": []any{
				map[string]any{
					"id": 1, "time": 1700000000.0, "type": "A", "domain": "example.com",
					"cname": nil, "status": "FORWARDED",
					"client":   map[string]any{"ip": "192.168.1.10", "name": "desktop"},
					"dnssec":   "INSECURE",
					"reply":    map[string]any{"type": "IP", "time": 5.2},
					"list_id":  nil,
					"upstream": "1.1.1.1#53",
				},
			},
			"cursor":          100,
			"recordsTotal":    50000,
			"recordsFiltered": 150,
		},
	}))

	text := callTool(t, queriesSearchHandler, c, nil)
	if !strings.Contains(text, "1 of 150 queries") {
		t.Errorf("expected '1 of 150 queries', got: %s", text)
	}
	if !strings.Contains(text, "example.com") {
		t.Errorf("expected domain in output, got: %s", text)
	}
	if !strings.Contains(text, "FORWARDED") {
		t.Errorf("expected status in output, got: %s", text)
	}
	if !strings.Contains(text, "cursor=100") {
		t.Errorf("expected cursor pagination, got: %s", text)
	}
}

func TestQueriesSearch_Minimal(t *testing.T) {
	c := newTestClient(t, piholeHandler(map[string]any{
		"/queries": map[string]any{
			"queries": []any{
				map[string]any{
					"id": 1, "time": 1700000000.0, "type": "A", "domain": "example.com",
					"cname": nil, "status": "FORWARDED",
					"client":   map[string]any{"ip": "192.168.1.10", "name": "desktop"},
					"dnssec":   "INSECURE",
					"reply":    map[string]any{"type": "IP", "time": 5.2},
					"list_id":  nil,
					"upstream": "1.1.1.1#53",
				},
			},
			"cursor":          100,
			"recordsTotal":    50000,
			"recordsFiltered": 150,
		},
	}))

	text := callTool(t, queriesSearchHandler, c, map[string]any{"detail": "minimal"})
	if !strings.Contains(text, "1 of 150 queries.") {
		t.Errorf("expected count-only output, got: %s", text)
	}
	if !strings.Contains(text, "cursor=100") {
		t.Errorf("expected cursor in minimal output, got: %s", text)
	}
	// Minimal should not contain domain details.
	if strings.Contains(text, "example.com") {
		t.Errorf("minimal should not contain domain details, got: %s", text)
	}
}

func TestQueriesSearch_CSV(t *testing.T) {
	c := newTestClient(t, piholeHandler(map[string]any{
		"/queries": map[string]any{
			"queries": []any{
				map[string]any{
					"id": 1, "time": 1700000000.0, "type": "A", "domain": "example.com",
					"cname": nil, "status": "FORWARDED",
					"client":   map[string]any{"ip": "192.168.1.10", "name": "desktop"},
					"dnssec":   "INSECURE",
					"reply":    map[string]any{"type": "IP", "time": 5.2},
					"list_id":  nil,
					"upstream": "1.1.1.1#53",
				},
			},
			"cursor":          0,
			"recordsTotal":    50000,
			"recordsFiltered": 1,
		},
	}))

	text := callTool(t, queriesSearchHandler, c, map[string]any{"format": "csv"})
	if !strings.Contains(text, "Time,Type,Domain,Status,Client,Upstream") {
		t.Errorf("expected CSV headers, got: %s", text)
	}
	if !strings.Contains(text, "example.com") {
		t.Errorf("expected domain in CSV row, got: %s", text)
	}
}

func TestQueriesSearch_Pagination(t *testing.T) {
	c := newTestClient(t, piholeHandler(map[string]any{
		"/queries": map[string]any{
			"queries": []any{
				map[string]any{
					"id": 1, "time": 1700000000.0, "type": "A", "domain": "page1.com",
					"cname": nil, "status": "FORWARDED",
					"client":   map[string]any{"ip": "192.168.1.10", "name": "desktop"},
					"dnssec":   nil,
					"reply":    map[string]any{"type": "IP", "time": 3.1},
					"list_id":  nil,
					"upstream": "8.8.8.8#53",
				},
			},
			"cursor":          250,
			"recordsTotal":    50000,
			"recordsFiltered": 500,
		},
	}))

	text := callTool(t, queriesSearchHandler, c, nil)
	if !strings.Contains(text, "1 of 500 queries") {
		t.Errorf("expected '1 of 500 queries', got: %s", text)
	}
	if !strings.Contains(text, "cursor=250") {
		t.Errorf("expected next page cursor, got: %s", text)
	}
}

func TestQueriesSearch_Empty(t *testing.T) {
	c := newTestClient(t, piholeHandler(map[string]any{
		"/queries": map[string]any{
			"queries":         []any{},
			"cursor":          0,
			"recordsTotal":    0,
			"recordsFiltered": 0,
		},
	}))

	text := callTool(t, queriesSearchHandler, c, map[string]any{"detail": "minimal"})
	if !strings.Contains(text, "0 of 0 queries.") {
		t.Errorf("expected '0 of 0 queries.', got: %s", text)
	}
}

func TestQueriesSuggestions_Normal(t *testing.T) {
	c := newTestClient(t, piholeHandler(map[string]any{
		"/queries/suggestions": map[string]any{
			"suggestions": map[string]any{
				"domain":      []any{"example.com"},
				"client_ip":   []any{"192.168.1.10"},
				"client_name": []any{"desktop"},
				"upstream":    []any{"1.1.1.1#53"},
				"type":        []any{"A", "AAAA"},
				"status":      []any{"FORWARDED", "GRAVITY"},
				"reply":       []any{"IP", "NODATA"},
				"dnssec":      []any{"INSECURE"},
			},
		},
	}))

	text := callTool(t, queriesSuggestionsHandler, c, nil)
	if !strings.Contains(text, "Types:") {
		t.Errorf("expected 'Types:' section, got: %s", text)
	}
	if !strings.Contains(text, "A, AAAA") {
		t.Errorf("expected query types, got: %s", text)
	}
	if !strings.Contains(text, "Statuses:") {
		t.Errorf("expected 'Statuses:' section, got: %s", text)
	}
	if !strings.Contains(text, "FORWARDED") {
		t.Errorf("expected status values, got: %s", text)
	}
	if !strings.Contains(text, "Upstreams:") {
		t.Errorf("expected 'Upstreams:' section, got: %s", text)
	}
}
