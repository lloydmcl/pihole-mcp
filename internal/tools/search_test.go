package tools

import (
	"strings"
	"testing"
)

func TestSearchDomains_Found(t *testing.T) {
	c := newTestClient(t, piholeHandler(map[string]any{
		"/search/ads.example.com": map[string]any{
			"search": map[string]any{
				"domains": []any{
					map[string]any{"domain": "ads.example.com", "type": "deny", "kind": "exact", "enabled": true, "comment": "", "id": 1, "date_added": 1700000000, "date_modified": 1700000000, "groups": []any{0}},
				},
				"gravity": []any{
					map[string]any{"domain": "ads.example.com", "address": "https://blocklist.example.com/list.txt", "comment": "", "enabled": true, "type": "block", "id": 1, "date_added": 1700000000, "date_modified": 1700000000, "date_updated": 1700000000, "number": 50000, "status": 0, "groups": []any{0}},
				},
				"parameters": map[string]any{"partial": false, "N": 20, "domain": "ads.example.com", "debug": false},
				"results": map[string]any{
					"domains": map[string]any{"exact": 1, "regex": 0},
					"gravity": map[string]any{"allow": 0, "block": 1},
					"total":   2,
				},
			},
		},
	}))

	text := callTool(t, searchDomainsHandler, c, map[string]any{
		"domain": "ads.example.com",
	})
	if !strings.Contains(text, "2 matches") {
		t.Errorf("expected '2 matches', got: %s", text)
	}
	if !strings.Contains(text, "Domain list matches") {
		t.Errorf("expected domain list section, got: %s", text)
	}
	if !strings.Contains(text, "Gravity matches") {
		t.Errorf("expected gravity section, got: %s", text)
	}
	if !strings.Contains(text, "ads.example.com") {
		t.Errorf("expected domain name in output, got: %s", text)
	}
	if !strings.Contains(text, "blocklist.example.com") {
		t.Errorf("expected gravity source in output, got: %s", text)
	}
}

func TestSearchDomains_NotFound(t *testing.T) {
	c := newTestClient(t, piholeHandler(map[string]any{
		"/search/nonexistent.com": map[string]any{
			"search": map[string]any{
				"domains":    []any{},
				"gravity":    []any{},
				"parameters": map[string]any{"partial": false, "N": 20, "domain": "nonexistent.com", "debug": false},
				"results": map[string]any{
					"domains": map[string]any{"exact": 0, "regex": 0},
					"gravity": map[string]any{"allow": 0, "block": 0},
					"total":   0,
				},
			},
		},
	}))

	text := callTool(t, searchDomainsHandler, c, map[string]any{
		"domain": "nonexistent.com",
	})
	if !strings.Contains(text, "0 matches") {
		t.Errorf("expected '0 matches', got: %s", text)
	}
	if strings.Contains(text, "Domain list matches") {
		t.Errorf("should not show domain list section when empty, got: %s", text)
	}
	if strings.Contains(text, "Gravity matches") {
		t.Errorf("should not show gravity section when empty, got: %s", text)
	}
}

func TestSearchDomains_Partial(t *testing.T) {
	c := newTestClient(t, piholeHandler(map[string]any{
		"/search/ads.example.com": map[string]any{
			"search": map[string]any{
				"domains": []any{
					map[string]any{"domain": "ads.example.com", "type": "deny", "kind": "exact", "enabled": true, "comment": "", "id": 1, "date_added": 1700000000, "date_modified": 1700000000, "groups": []any{0}},
				},
				"gravity":    []any{},
				"parameters": map[string]any{"partial": true, "N": 20, "domain": "ads.example.com", "debug": false},
				"results": map[string]any{
					"domains": map[string]any{"exact": 1, "regex": 0},
					"gravity": map[string]any{"allow": 0, "block": 0},
					"total":   1,
				},
			},
		},
	}))

	text := callTool(t, searchDomainsHandler, c, map[string]any{
		"domain":  "ads.example.com",
		"partial": true,
	})
	if !strings.Contains(text, "1 matches") {
		t.Errorf("expected '1 matches', got: %s", text)
	}
	if !strings.Contains(text, "1 exact") {
		t.Errorf("expected exact match count, got: %s", text)
	}
}
