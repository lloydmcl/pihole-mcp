package tools

import (
	"strings"
	"testing"
)

func TestDNSGetBlocking_Enabled(t *testing.T) {
	c := newTestClient(t, piholeHandler(map[string]any{
		"/dns/blocking": map[string]any{"blocking": "enabled"},
	}))

	text := callTool(t, dnsGetBlockingHandler, c, nil)
	if !strings.Contains(text, "enabled") {
		t.Errorf("expected 'enabled' in response, got: %s", text)
	}
}

func TestDNSGetBlocking_WithTimer(t *testing.T) {
	timer := 45.0
	c := newTestClient(t, piholeHandler(map[string]any{
		"/dns/blocking": map[string]any{"blocking": "disabled", "timer": timer},
	}))

	text := callTool(t, dnsGetBlockingHandler, c, nil)
	if !strings.Contains(text, "disabled") {
		t.Errorf("expected 'disabled' in response, got: %s", text)
	}
	if !strings.Contains(text, "45") {
		t.Errorf("expected timer '45' in response, got: %s", text)
	}
}

func TestDNSSetBlocking(t *testing.T) {
	c := newTestClient(t, piholeHandler(map[string]any{
		"/dns/blocking": map[string]any{"blocking": "disabled", "timer": 60.0},
	}))

	text := callTool(t, dnsSetBlockingHandler, c, map[string]any{
		"blocking": false,
		"timer":    60.0,
	})
	if !strings.Contains(text, "disabled") {
		t.Errorf("expected 'disabled' in response, got: %s", text)
	}
}
