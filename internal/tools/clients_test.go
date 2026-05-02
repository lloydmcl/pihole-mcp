package tools

import (
	"strings"
	"testing"
)

func TestClientsList_Normal(t *testing.T) {
	c := newTestClient(t, piholeHandler(map[string]any{
		"/clients": map[string]any{
			"clients": []any{
				map[string]any{"client": "192.168.1.10", "name": "desktop", "comment": "Main PC", "groups": []any{0, 1}, "id": 1},
				map[string]any{"client": "192.168.1.20", "name": "laptop", "comment": "", "groups": []any{0}, "id": 2},
			},
		},
	}))

	text := callTool(t, clientsListHandler, c, nil)
	if !strings.Contains(text, "2 clients") {
		t.Errorf("expected client count, got: %s", text)
	}
	if !strings.Contains(text, "192.168.1.10") {
		t.Errorf("expected first client IP, got: %s", text)
	}
	if !strings.Contains(text, "desktop") {
		t.Errorf("expected client name, got: %s", text)
	}
}

func TestClientsList_CSV(t *testing.T) {
	c := newTestClient(t, piholeHandler(map[string]any{
		"/clients": map[string]any{
			"clients": []any{
				map[string]any{"client": "192.168.1.10", "name": "desktop", "comment": "Main PC", "groups": []any{0}, "id": 1},
			},
		},
	}))

	text := callTool(t, clientsListHandler, c, map[string]any{"format": "csv"})
	if !strings.Contains(text, "Client,Name,Comment,Groups") {
		t.Errorf("expected CSV headers, got: %s", text)
	}
	if !strings.Contains(text, "192.168.1.10") {
		t.Errorf("expected client in CSV, got: %s", text)
	}
}

func TestClientsList_Empty(t *testing.T) {
	c := newTestClient(t, piholeHandler(map[string]any{
		"/clients": map[string]any{"clients": []any{}},
	}))

	text := callTool(t, clientsListHandler, c, nil)
	if text != "No configured clients." {
		t.Errorf("expected empty message, got: %s", text)
	}
}

func TestClientsSuggestions_Normal(t *testing.T) {
	hwaddr := "AA:BB:CC:DD:EE:FF"
	vendor := "Apple"
	addr := "192.168.1.50"
	names := "macbook"

	c := newTestClient(t, piholeHandler(map[string]any{
		"/clients/_suggestions": map[string]any{
			"clients": []any{
				map[string]any{"hwaddr": hwaddr, "macVendor": vendor, "lastQuery": 1700000000, "addresses": addr, "names": names},
				map[string]any{"lastQuery": 1700000000, "addresses": "10.0.0.5"},
			},
		},
	}))

	text := callTool(t, clientsSuggestionsHandler, c, nil)
	if !strings.Contains(text, "2 unconfigured clients") {
		t.Errorf("expected suggestion count, got: %s", text)
	}
	if !strings.Contains(text, "AA:BB:CC:DD:EE:FF") {
		t.Errorf("expected MAC address, got: %s", text)
	}
	if !strings.Contains(text, "Apple") {
		t.Errorf("expected vendor, got: %s", text)
	}
	if !strings.Contains(text, "unknown MAC") {
		t.Errorf("expected 'unknown MAC' for nil hwaddr, got: %s", text)
	}
}

func TestClientsSuggestions_NilFields(t *testing.T) {
	c := newTestClient(t, piholeHandler(map[string]any{
		"/clients/_suggestions": map[string]any{
			"clients": []any{
				map[string]any{"lastQuery": 1700000000},
			},
		},
	}))

	text := callTool(t, clientsSuggestionsHandler, c, nil)
	if !strings.Contains(text, "unknown MAC") {
		t.Errorf("expected 'unknown MAC' for nil hwaddr, got: %s", text)
	}
	if !strings.Contains(text, "no IPs") {
		t.Errorf("expected 'no IPs' for nil addresses, got: %s", text)
	}
}

func TestClientsSuggestions_Empty(t *testing.T) {
	c := newTestClient(t, piholeHandler(map[string]any{
		"/clients/_suggestions": map[string]any{"clients": []any{}},
	}))

	text := callTool(t, clientsSuggestionsHandler, c, nil)
	if text != "No unconfigured clients found." {
		t.Errorf("expected empty message, got: %s", text)
	}
}

func TestClientsAdd_Success(t *testing.T) {
	c := newTestClient(t, piholeHandler(map[string]any{
		"/clients": map[string]any{"clients": []any{}},
	}))

	text := callTool(t, clientsAddHandler, c, map[string]any{"client": "192.168.1.30"})
	if !strings.Contains(text, "Added") {
		t.Errorf("expected 'Added' message, got: %s", text)
	}
}

func TestClientsUpdate_Success(t *testing.T) {
	c := newTestClient(t, piholeHandler(map[string]any{
		"/clients/192.168.1.10": map[string]any{"clients": []any{}},
	}))

	text := callTool(t, clientsUpdateHandler, c, map[string]any{"client": "192.168.1.10", "comment": "updated"})
	if !strings.Contains(text, "Updated") {
		t.Errorf("expected 'Updated' message, got: %s", text)
	}
	if !strings.Contains(text, "192.168.1.10") {
		t.Errorf("expected client identifier in message, got: %s", text)
	}
}

func TestClientsDelete_Success(t *testing.T) {
	c := newTestClient(t, piholeHandler(map[string]any{
		"/clients/192.168.1.30": map[string]any{},
	}))

	text := callTool(t, clientsDeleteHandler, c, map[string]any{"client": "192.168.1.30"})
	if !strings.Contains(text, "Deleted") {
		t.Errorf("expected 'Deleted' message, got: %s", text)
	}
}

func TestClientsBatchDelete_Success(t *testing.T) {
	c := newTestClient(t, piholeHandler(map[string]any{
		"/clients:batchDelete": map[string]any{},
	}))

	text := callTool(t, clientsBatchDeleteHandler, c, map[string]any{
		"items": `["192.168.1.10","192.168.1.20"]`,
	})
	if !strings.Contains(text, "Batch delete completed") {
		t.Errorf("expected 'Batch delete completed' message, got: %s", text)
	}
}
