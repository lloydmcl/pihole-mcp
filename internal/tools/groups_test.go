package tools

import (
	"strings"
	"testing"
)

func TestGroupsList_Normal(t *testing.T) {
	c := newTestClient(t, piholeHandler(map[string]any{
		"/groups": map[string]any{
			"groups": []any{
				map[string]any{"name": "default", "comment": "", "enabled": true, "id": 0},
				map[string]any{"name": "kids", "comment": "Child devices", "enabled": true, "id": 1},
			},
		},
	}))

	text := callTool(t, groupsListHandler, c, nil)
	if !strings.Contains(text, "2 groups") {
		t.Errorf("expected group count, got: %s", text)
	}
	if !strings.Contains(text, "default") {
		t.Errorf("expected 'default' group, got: %s", text)
	}
	if !strings.Contains(text, "kids") {
		t.Errorf("expected 'kids' group, got: %s", text)
	}
	if !strings.Contains(text, "Child devices") {
		t.Errorf("expected comment, got: %s", text)
	}
}

func TestGroupsList_Empty(t *testing.T) {
	c := newTestClient(t, piholeHandler(map[string]any{
		"/groups": map[string]any{"groups": []any{}},
	}))

	text := callTool(t, groupsListHandler, c, nil)
	if text != "No groups found." {
		t.Errorf("expected empty message, got: %s", text)
	}
}

func TestGroupsAdd_Success(t *testing.T) {
	c := newTestClient(t, piholeHandler(map[string]any{
		"/groups": map[string]any{"groups": []any{}},
	}))

	text := callTool(t, groupsAddHandler, c, map[string]any{"name": "guests"})
	if !strings.Contains(text, "Created") {
		t.Errorf("expected 'Created' message, got: %s", text)
	}
	if !strings.Contains(text, "guests") {
		t.Errorf("expected group name in message, got: %s", text)
	}
}

func TestGroupsUpdate_Success(t *testing.T) {
	c := newTestClient(t, piholeHandler(map[string]any{
		"/groups/kids": map[string]any{"groups": []any{}},
	}))

	text := callTool(t, groupsUpdateHandler, c, map[string]any{"name": "kids", "comment": "updated"})
	if !strings.Contains(text, "Updated") {
		t.Errorf("expected 'Updated' message, got: %s", text)
	}
}

func TestGroupsDelete_Success(t *testing.T) {
	c := newTestClient(t, piholeHandler(map[string]any{
		"/groups/kids": map[string]any{},
	}))

	text := callTool(t, groupsDeleteHandler, c, map[string]any{"name": "kids"})
	if !strings.Contains(text, "Deleted") {
		t.Errorf("expected 'Deleted' message, got: %s", text)
	}
}

func TestGroupsBatchDelete_Success(t *testing.T) {
	c := newTestClient(t, piholeHandler(map[string]any{
		"/groups:batchDelete": map[string]any{},
	}))

	text := callTool(t, groupsBatchDeleteHandler, c, map[string]any{
		"items": `[{"item":"kids"}]`,
	})
	if !strings.Contains(text, "Batch delete completed") {
		t.Errorf("expected 'Batch delete completed' message, got: %s", text)
	}
}
