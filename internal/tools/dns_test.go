package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/lloydmcl/pihole-mcp/internal/pihole"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func newTestClient(t *testing.T, handler http.Handler) *pihole.Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return pihole.New(srv.URL, "test", pihole.WithHTTPClient(srv.Client()))
}

func piholeHandler(routes map[string]any) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/auth" {
			writeTestJSON(w, map[string]any{
				"session": map[string]any{"valid": true, "sid": "test-sid"},
			})
			return
		}
		path := strings.TrimPrefix(r.URL.Path, "/api")
		if resp, ok := routes[path]; ok {
			writeTestJSON(w, resp)
			return
		}
		http.NotFound(w, r)
	}
}

func writeTestJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func callTool(t *testing.T, handlerFn func(*pihole.Client) server.ToolHandlerFunc, c *pihole.Client, args map[string]any) string {
	t.Helper()
	h := handlerFn(c)
	req := mcp.CallToolRequest{}
	req.Params.Arguments = args
	result, err := h(context.Background(), req)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if result.IsError {
		t.Fatalf("tool error: %v", result.Content)
	}
	if len(result.Content) == 0 {
		return ""
	}
	if tc, ok := result.Content[0].(mcp.TextContent); ok {
		return tc.Text
	}
	return ""
}

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
