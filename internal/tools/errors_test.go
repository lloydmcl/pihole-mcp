package tools

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hexamatic/pihole-mcp/internal/pihole"
	"github.com/mark3labs/mcp-go/mcp"
)

func TestToolError_AuthError(t *testing.T) {
	err := &pihole.AuthError{APIError: &pihole.APIError{StatusCode: 401, Message: "Unauthorized"}}
	result := toolError("get stats", err)
	if !result.IsError {
		t.Fatal("expected IsError to be true")
	}
	text := mustText(t, result)
	if !strings.Contains(text, "Failed to get stats") {
		t.Errorf("expected action in message, got: %s", text)
	}
	if !strings.Contains(text, "Check PIHOLE_PASSWORD") {
		t.Errorf("expected auth guidance in message, got: %s", text)
	}
}

func TestToolError_RateLimitError(t *testing.T) {
	err := &pihole.RateLimitError{APIError: &pihole.APIError{StatusCode: 429, Message: "Rate limit exceeded"}}
	result := toolError("list domains", err)
	if !result.IsError {
		t.Fatal("expected IsError to be true")
	}
	text := mustText(t, result)
	if !strings.Contains(text, "Failed to list domains") {
		t.Errorf("expected action in message, got: %s", text)
	}
	if !strings.Contains(text, "try again shortly") {
		t.Errorf("expected rate limit guidance in message, got: %s", text)
	}
}

func TestToolError_NotFoundError(t *testing.T) {
	err := &pihole.NotFoundError{APIError: &pihole.APIError{StatusCode: 404, Message: "Not found"}}
	result := toolError("get client", err)
	if !result.IsError {
		t.Fatal("expected IsError to be true")
	}
	text := mustText(t, result)
	if !strings.Contains(text, "Failed to get client") {
		t.Errorf("expected action in message, got: %s", text)
	}
	if !strings.Contains(text, "does not exist") {
		t.Errorf("expected not-found guidance in message, got: %s", text)
	}
}

func TestToolError_ValidationError(t *testing.T) {
	err := &pihole.ValidationError{APIError: &pihole.APIError{StatusCode: 400, Message: "Invalid domain format"}}
	result := toolError("add domain", err)
	if !result.IsError {
		t.Fatal("expected IsError to be true")
	}
	text := mustText(t, result)
	if !strings.Contains(text, "Failed to add domain") {
		t.Errorf("expected action in message, got: %s", text)
	}
	if !strings.Contains(text, "Invalid domain format") {
		t.Errorf("expected validation message, got: %s", text)
	}
}

func TestToolError_ValidationError_WithHint(t *testing.T) {
	err := &pihole.ValidationError{APIError: &pihole.APIError{
		StatusCode: 400,
		Message:    "Invalid domain format",
		Hint:       "Use a FQDN like example.com",
	}}
	result := toolError("add domain", err)
	text := mustText(t, result)
	if !strings.Contains(text, "Use a FQDN like example.com") {
		t.Errorf("expected hint in message, got: %s", text)
	}
}

func TestToolError_APIError(t *testing.T) {
	err := &pihole.APIError{StatusCode: 500, Message: "Internal server error"}
	result := toolError("restart DNS", err)
	if !result.IsError {
		t.Fatal("expected IsError to be true")
	}
	text := mustText(t, result)
	if !strings.Contains(text, "Failed to restart DNS") {
		t.Errorf("expected action in message, got: %s", text)
	}
	if !strings.Contains(text, "Internal server error") {
		t.Errorf("expected API error message, got: %s", text)
	}
}

func TestToolError_APIError_WithHint(t *testing.T) {
	err := &pihole.APIError{StatusCode: 500, Message: "Database locked", Hint: "Wait and retry"}
	result := toolError("get stats", err)
	text := mustText(t, result)
	if !strings.Contains(text, "Wait and retry") {
		t.Errorf("expected hint in message, got: %s", text)
	}
}

func TestToolError_GenericError(t *testing.T) {
	err := fmt.Errorf("connection refused")
	result := toolError("get stats", err)
	if !result.IsError {
		t.Fatal("expected IsError to be true")
	}
	text := mustText(t, result)
	if !strings.Contains(text, "Failed to get stats") {
		t.Errorf("expected action in message, got: %s", text)
	}
	if !strings.Contains(text, "connection refused") {
		t.Errorf("expected original error in message, got: %s", text)
	}
}

// mustText extracts the text content from a CallToolResult for test assertions.
func mustText(t *testing.T, result *mcp.CallToolResult) string {
	t.Helper()
	if len(result.Content) == 0 {
		t.Fatal("expected content in result")
	}
	tc, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}
	return tc.Text
}
