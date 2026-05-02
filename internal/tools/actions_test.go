package tools

import (
	"strings"
	"testing"
)

func TestActionGravity_Success(t *testing.T) {
	c := newTestClient(t, piholeRawHandler(
		nil,
		map[string]string{
			"/action/gravity": "Downloading blocklist...\nDone.\n",
		},
	))

	text := callTool(t, actionGravityHandler, c, nil)
	if !strings.Contains(text, "Gravity update complete") {
		t.Errorf("expected gravity complete message, got: %s", text)
	}
	if !strings.Contains(text, "```") {
		t.Errorf("expected code block wrapping, got: %s", text)
	}
	if !strings.Contains(text, "Downloading blocklist") {
		t.Errorf("expected raw output in code block, got: %s", text)
	}
}

func TestActionRestartDNS_Success(t *testing.T) {
	c := newTestClient(t, piholeHandler(map[string]any{
		"/action/restartdns": map[string]any{"success": true},
	}))

	text := callTool(t, actionRestartDNSHandler, c, nil)
	if !strings.Contains(text, "DNS restarted") {
		t.Errorf("expected 'DNS restarted' message, got: %s", text)
	}
}

func TestActionFlushLogs_Success(t *testing.T) {
	c := newTestClient(t, piholeHandler(map[string]any{
		"/action/flush/logs": map[string]any{"success": true},
	}))

	text := callTool(t, actionFlushLogsHandler, c, nil)
	if !strings.Contains(text, "Logs flushed") {
		t.Errorf("expected 'Logs flushed' message, got: %s", text)
	}
}

func TestActionFlushNetwork_Success(t *testing.T) {
	c := newTestClient(t, piholeHandler(map[string]any{
		"/action/flush/network": map[string]any{"success": true},
	}))

	text := callTool(t, actionFlushNetworkHandler, c, nil)
	if !strings.Contains(text, "Network table flushed") {
		t.Errorf("expected 'Network table flushed' message, got: %s", text)
	}
}
