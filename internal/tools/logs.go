package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/lloydmcl/pihole-mcp/internal/pihole"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// RegisterLogs registers log retrieval tools.
func RegisterLogs(s *server.MCPServer, c *pihole.Client) {
	for _, l := range []struct {
		name, endpoint, desc string
	}{
		{"pihole_logs_dns", "/logs/dnsmasq",
			"DNS resolver (dnsmasq) log. Use next_id for incremental polling to follow new entries."},
		{"pihole_logs_ftl", "/logs/ftl",
			"FTL engine log — internal Pi-hole diagnostics, database operations, and resolver events."},
		{"pihole_logs_webserver", "/logs/webserver",
			"Web server access log — HTTP requests to the Pi-hole admin interface and API."},
	} {
		addTool(s, mcp.NewTool(l.name,
			mcp.WithDescription(l.desc),
			mcp.WithNumber("next_id", mcp.Description("Only return lines after this ID (incremental polling).")),
			mcp.WithReadOnlyHintAnnotation(true),
		), logHandler(c, l.endpoint))
	}
}

func logHandler(c *pihole.Client, endpoint string) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		path := endpoint
		if nextID := req.GetFloat("next_id", 0); nextID > 0 {
			path += fmt.Sprintf("?nextID=%.0f", nextID)
		}

		var result pihole.LogResponse
		if err := c.Get(ctx, path, &result); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get logs: %v", err)), nil
		}

		if len(result.Log) == 0 {
			return mcp.NewToolResultText(fmt.Sprintf("No log entries. Next ID: %d", result.NextID)), nil
		}

		shown := min(50, len(result.Log))
		var b strings.Builder
		fmt.Fprintf(&b, "**%d lines** (showing %d):\n", len(result.Log), shown)
		for _, entry := range result.Log[:shown] {
			fmt.Fprintf(&b, "- %s\n", entry.Message)
		}
		fmt.Fprintf(&b, "Next ID: %d", result.NextID)

		return mcp.NewToolResultText(b.String()), nil
	}
}
