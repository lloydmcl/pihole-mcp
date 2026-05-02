package tools

import (
	"strings"
	"testing"
)

func TestDomainsList_Normal(t *testing.T) {
	c := newTestClient(t, piholeHandler(map[string]any{
		"/domains": map[string]any{
			"domains": []any{
				map[string]any{"domain": "example.com", "type": "deny", "kind": "exact", "enabled": true, "comment": "test", "id": 1, "groups": []any{0}, "date_added": 1700000000, "date_modified": 1700000000},
				map[string]any{"domain": "ads.net", "type": "deny", "kind": "regex", "enabled": false, "comment": "", "id": 2, "groups": []any{0, 1}, "date_added": 1700000000, "date_modified": 1700000000},
			},
		},
	}))

	text := callTool(t, domainsListHandler, c, nil)
	if !strings.Contains(text, "2 domains") {
		t.Errorf("expected domain count header, got: %s", text)
	}
	if !strings.Contains(text, "example.com") {
		t.Errorf("expected example.com in output, got: %s", text)
	}
	if !strings.Contains(text, "ads.net") {
		t.Errorf("expected ads.net in output, got: %s", text)
	}
}

func TestDomainsList_Minimal(t *testing.T) {
	c := newTestClient(t, piholeHandler(map[string]any{
		"/domains": map[string]any{
			"domains": []any{
				map[string]any{"domain": "example.com", "type": "deny", "kind": "exact", "enabled": true, "comment": "", "id": 1, "groups": []any{0}},
			},
		},
	}))

	text := callTool(t, domainsListHandler, c, map[string]any{"detail": "minimal"})
	if !strings.Contains(text, "1 domains.") {
		t.Errorf("expected single-line count, got: %s", text)
	}
}

func TestDomainsList_Full(t *testing.T) {
	c := newTestClient(t, piholeHandler(map[string]any{
		"/domains": map[string]any{
			"domains": []any{
				map[string]any{"domain": "example.com", "type": "deny", "kind": "exact", "enabled": true, "comment": "", "id": 5, "groups": []any{0, 2}},
			},
		},
	}))

	text := callTool(t, domainsListHandler, c, map[string]any{"detail": "full"})
	if !strings.Contains(text, "id=5") {
		t.Errorf("expected id in full detail, got: %s", text)
	}
	if !strings.Contains(text, "groups=") {
		t.Errorf("expected groups in full detail, got: %s", text)
	}
}

func TestDomainsList_CSV(t *testing.T) {
	c := newTestClient(t, piholeHandler(map[string]any{
		"/domains": map[string]any{
			"domains": []any{
				map[string]any{"domain": "example.com", "type": "deny", "kind": "exact", "enabled": true, "comment": "blocked", "id": 1, "groups": []any{0}},
			},
		},
	}))

	text := callTool(t, domainsListHandler, c, map[string]any{"format": "csv"})
	if !strings.Contains(text, "Domain,Type,Kind,Enabled,Comment") {
		t.Errorf("expected CSV header, got: %s", text)
	}
	if !strings.Contains(text, "example.com") {
		t.Errorf("expected domain in CSV row, got: %s", text)
	}
}

func TestDomainsList_Empty(t *testing.T) {
	c := newTestClient(t, piholeHandler(map[string]any{
		"/domains": map[string]any{"domains": []any{}},
	}))

	text := callTool(t, domainsListHandler, c, nil)
	if text != "No domains found." {
		t.Errorf("expected empty message, got: %s", text)
	}
}

func TestDomainsAdd_Success(t *testing.T) {
	c := newTestClient(t, piholeHandler(map[string]any{
		"/domains/deny/exact": map[string]any{
			"domains":   []any{},
			"processed": map[string]any{"success": []any{map[string]any{"item": "example.com"}}, "errors": []any{}},
		},
	}))

	text := callTool(t, domainsAddHandler, c, map[string]any{
		"type": "deny", "kind": "exact", "domain": "example.com",
	})
	if !strings.Contains(text, "Domain added") {
		t.Errorf("expected 'Domain added' message, got: %s", text)
	}
	if !strings.Contains(text, "example.com") {
		t.Errorf("expected processed domain in output, got: %s", text)
	}
}

func TestDomainsUpdate_Success(t *testing.T) {
	c := newTestClient(t, piholeHandler(map[string]any{
		"/domains/deny/exact/example.com": map[string]any{"domains": []any{}},
	}))

	text := callTool(t, domainsUpdateHandler, c, map[string]any{
		"type": "deny", "kind": "exact", "domain": "example.com", "comment": "updated",
	})
	if !strings.Contains(text, "Updated") {
		t.Errorf("expected 'Updated' message, got: %s", text)
	}
}

func TestDomainsDelete_Success(t *testing.T) {
	c := newTestClient(t, piholeHandler(map[string]any{
		"/domains/deny/exact/example.com": map[string]any{},
	}))

	text := callTool(t, domainsDeleteHandler, c, map[string]any{
		"type": "deny", "kind": "exact", "domain": "example.com",
	})
	if !strings.Contains(text, "Deleted") {
		t.Errorf("expected 'Deleted' message, got: %s", text)
	}
}

func TestDomainsBatchDelete_Success(t *testing.T) {
	c := newTestClient(t, piholeHandler(map[string]any{
		"/domains:batchDelete": map[string]any{},
	}))

	text := callTool(t, domainsBatchDeleteHandler, c, map[string]any{
		"items": `[{"item":"example.com","type":"deny","kind":"exact"}]`,
	})
	if !strings.Contains(text, "Batch delete completed") {
		t.Errorf("expected 'Batch delete completed' message, got: %s", text)
	}
}

func TestDomainsList_Error(t *testing.T) {
	c := newTestClient(t, piholeErrorServer(400, "bad_request", "Invalid filter", "Check parameters"))

	text := callToolExpectError(t, domainsListHandler, c, nil)
	if text == "" {
		t.Error("expected error text, got empty string")
	}
}
