package tools

import (
	"strings"
	"testing"
)

func TestListsList_Normal(t *testing.T) {
	c := newTestClient(t, piholeHandler(map[string]any{
		"/lists": map[string]any{
			"lists": []any{
				map[string]any{"address": "https://example.com/list.txt", "type": "block", "comment": "Ad list", "groups": []any{0}, "enabled": true, "id": 1, "number": 50000, "invalid_domains": 3, "date_added": 1700000000, "date_modified": 1700000000, "date_updated": 1700000000, "abp_entries": 0, "status": 0},
				map[string]any{"address": "https://example.com/allow.txt", "type": "allow", "comment": "", "groups": []any{0}, "enabled": true, "id": 2, "number": 100, "invalid_domains": 0, "date_added": 1700000000, "date_modified": 1700000000, "date_updated": 1700000000, "abp_entries": 0, "status": 0},
			},
		},
	}))

	text := callTool(t, listsListHandler, c, nil)
	if !strings.Contains(text, "2 lists") {
		t.Errorf("expected list count, got: %s", text)
	}
	if !strings.Contains(text, "50,000 domains") {
		t.Errorf("expected formatted domain count, got: %s", text)
	}
	if !strings.Contains(text, "https://example.com/list.txt") {
		t.Errorf("expected list URL, got: %s", text)
	}
}

func TestListsList_Minimal(t *testing.T) {
	c := newTestClient(t, piholeHandler(map[string]any{
		"/lists": map[string]any{
			"lists": []any{
				map[string]any{"address": "https://example.com/list.txt", "type": "block", "comment": "", "groups": []any{0}, "enabled": true, "id": 1, "number": 50000, "invalid_domains": 0, "date_added": 1700000000, "date_modified": 1700000000, "date_updated": 1700000000, "abp_entries": 0, "status": 0},
			},
		},
	}))

	text := callTool(t, listsListHandler, c, map[string]any{"detail": "minimal"})
	if !strings.Contains(text, "1 lists.") {
		t.Errorf("expected minimal count, got: %s", text)
	}
}

func TestListsList_Full(t *testing.T) {
	c := newTestClient(t, piholeHandler(map[string]any{
		"/lists": map[string]any{
			"lists": []any{
				map[string]any{"address": "https://example.com/list.txt", "type": "block", "comment": "", "groups": []any{0}, "enabled": true, "id": 7, "number": 50000, "invalid_domains": 3, "date_added": 1700000000, "date_modified": 1700000000, "date_updated": 1700000000, "abp_entries": 0, "status": 0},
			},
		},
	}))

	text := callTool(t, listsListHandler, c, map[string]any{"detail": "full"})
	if !strings.Contains(text, "id=7") {
		t.Errorf("expected id in full detail, got: %s", text)
	}
	if !strings.Contains(text, "updated=") {
		t.Errorf("expected updated timestamp in full detail, got: %s", text)
	}
	if !strings.Contains(text, "invalid=3") {
		t.Errorf("expected invalid count in full detail, got: %s", text)
	}
}

func TestListsList_CSV(t *testing.T) {
	c := newTestClient(t, piholeHandler(map[string]any{
		"/lists": map[string]any{
			"lists": []any{
				map[string]any{"address": "https://example.com/list.txt", "type": "block", "comment": "Ad list", "groups": []any{0}, "enabled": true, "id": 1, "number": 50000, "invalid_domains": 0, "date_added": 1700000000, "date_modified": 1700000000, "date_updated": 1700000000, "abp_entries": 0, "status": 0},
			},
		},
	}))

	text := callTool(t, listsListHandler, c, map[string]any{"format": "csv"})
	if !strings.Contains(text, "Address,Type,Domains,Enabled,Comment") {
		t.Errorf("expected CSV headers, got: %s", text)
	}
	if !strings.Contains(text, "https://example.com/list.txt") {
		t.Errorf("expected list address in CSV, got: %s", text)
	}
}

func TestListsList_Empty(t *testing.T) {
	c := newTestClient(t, piholeHandler(map[string]any{
		"/lists": map[string]any{"lists": []any{}},
	}))

	text := callTool(t, listsListHandler, c, nil)
	if text != "No lists found." {
		t.Errorf("expected empty message, got: %s", text)
	}
}

func TestListsAdd_Success(t *testing.T) {
	c := newTestClient(t, piholeHandler(map[string]any{
		"/lists": map[string]any{"lists": []any{}},
	}))

	text := callTool(t, listsAddHandler, c, map[string]any{
		"address": "https://example.com/new.txt", "type": "block",
	})
	if !strings.Contains(text, "Added") {
		t.Errorf("expected 'Added' message, got: %s", text)
	}
	if !strings.Contains(text, "gravity") {
		t.Errorf("expected gravity reminder, got: %s", text)
	}
}

func TestListsUpdate_Success(t *testing.T) {
	c := newTestClient(t, piholeHandler(map[string]any{
		"/lists/https://example.com/list.txt": map[string]any{"lists": []any{}},
	}))

	text := callTool(t, listsUpdateHandler, c, map[string]any{
		"address": "https://example.com/list.txt", "type": "block", "comment": "updated",
	})
	if !strings.Contains(text, "Updated") {
		t.Errorf("expected 'Updated' message, got: %s", text)
	}
	if !strings.Contains(text, "https://example.com/list.txt") {
		t.Errorf("expected list address in message, got: %s", text)
	}
}

func TestListsDelete_Success(t *testing.T) {
	c := newTestClient(t, piholeHandler(map[string]any{
		"/lists/https://example.com/list.txt": map[string]any{},
	}))

	text := callTool(t, listsDeleteHandler, c, map[string]any{
		"address": "https://example.com/list.txt", "type": "block",
	})
	if !strings.Contains(text, "Deleted") {
		t.Errorf("expected 'Deleted' message, got: %s", text)
	}
}

func TestListsBatchDelete_Success(t *testing.T) {
	c := newTestClient(t, piholeHandler(map[string]any{
		"/lists:batchDelete": map[string]any{},
	}))

	text := callTool(t, listsBatchDeleteHandler, c, map[string]any{
		"items": `[{"item":"https://example.com/list.txt","type":"block"}]`,
	})
	if !strings.Contains(text, "Batch delete completed") {
		t.Errorf("expected 'Batch delete completed' message, got: %s", text)
	}
}
