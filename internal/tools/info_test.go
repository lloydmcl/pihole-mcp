package tools

import (
	"strings"
	"testing"
)

func TestInfoSystem_EmptyUnits(t *testing.T) {
	// Simulates Docker environment where Pi-hole returns empty units.
	c := newTestClient(t, piholeHandler(map[string]any{
		"/info/system": map[string]any{
			"system": map[string]any{
				"uptime": 3600,
				"load":   []any{0.5, 0.3, 0.1},
				"cpu":    map[string]any{"nprocs": 4, "perc": 15.5},
				"memory": map[string]any{
					"ram": map[string]any{
						"total": 8025148.0, "used": 543304.0, "free": 7481844.0,
						"perc": 6.8, "unit": "", // Empty unit — Docker quirk
					},
				},
				"disk": map[string]any{
					"total": 0.0, "used": 0.0, "free": 0.0,
					"perc": 0.0, "unit": "", // Empty unit
				},
				"dns": map[string]any{"running": false},
			},
		},
		"/info/host": map[string]any{
			"host": map[string]any{
				"name": "pihole-dev", "os": "Linux", "arch": "aarch64",
				"kernel": "6.8.0", "domain": "",
			},
		},
		"/info/sensors": map[string]any{
			"sensors": map[string]any{"list": []any{}},
		},
	}))

	text := callTool(t, infoSystemHandler, c, nil)

	// Should show auto-converted bytes, not "543304.0/8025148.0 "
	if strings.Contains(text, "543304") {
		t.Errorf("raw bytes should be formatted, got: %s", text)
	}
	if !strings.Contains(text, "KB") && !strings.Contains(text, "MB") && !strings.Contains(text, "GB") {
		t.Errorf("expected auto-formatted size unit, got: %s", text)
	}

	// DNS not running should have Docker context
	if strings.Contains(text, "DNS:** not running\n") {
		t.Errorf("DNS not running should have Docker context, got: %s", text)
	}
	if !strings.Contains(text, "expected in Docker") {
		t.Errorf("should mention Docker context for DNS, got: %s", text)
	}
}

func TestInfoSystem_Minimal(t *testing.T) {
	c := newTestClient(t, piholeHandler(map[string]any{
		"/info/system": map[string]any{
			"system": map[string]any{
				"uptime": 3600,
				"load":   []any{0.5, 0.3, 0.1},
				"cpu":    map[string]any{"nprocs": 4, "perc": 15.5},
				"memory": map[string]any{
					"ram": map[string]any{
						"total": 8000000.0, "used": 500000.0, "free": 7500000.0,
						"perc": 6.3, "unit": "kB",
					},
				},
				"disk": map[string]any{"total": 50.0, "used": 20.0, "free": 30.0, "perc": 40.0, "unit": "GB"},
				"dns":  map[string]any{"running": true},
			},
		},
	}))

	text := callTool(t, infoSystemHandler, c, map[string]any{"detail": "minimal"})
	if strings.Count(text, "\n") > 1 {
		t.Errorf("minimal should be single-line, got: %s", text)
	}
	if !strings.Contains(text, "Load:") {
		t.Errorf("minimal should contain load, got: %s", text)
	}
}

func TestInfoDatabase_EmptyFields(t *testing.T) {
	c := newTestClient(t, piholeHandler(map[string]any{
		"/info/database": map[string]any{
			"database": map[string]any{
				"size": 0.0, "unit": "", "queries": 0, "sqlite_version": "",
			},
		},
	}))

	text := callTool(t, infoDatabaseHandler, c, nil)
	if strings.Contains(text, "| **Queries:** 0 | **SQLite:**  ") {
		t.Errorf("empty fields should show N/A, got: %s", text)
	}
	if !strings.Contains(text, "N/A") {
		t.Errorf("expected N/A for empty SQLite version, got: %s", text)
	}
}
