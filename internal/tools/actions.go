package tools

import (
	"context"
	"fmt"
	"io"

	"github.com/hexamatic/pihole-mcp/internal/pihole"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// RegisterActions registers action tools (gravity update, restart, flush).
func RegisterActions(s *server.MCPServer, c *pihole.Client) {
	addTool(s, mcp.NewTool("pihole_action_gravity_update",
		mcp.WithDescription("Re-download all configured blocklists and rebuild the gravity database. Takes 30+ seconds. Run after adding or removing lists."),
		mcp.WithOpenWorldHintAnnotation(true),
		mcp.WithIdempotentHintAnnotation(true),
	), actionGravityHandler(c))

	addTool(s, mcp.NewTool("pihole_action_restart_dns",
		mcp.WithDescription("Restart the FTL DNS resolver. Briefly interrupts DNS resolution for all clients on the network."),
		mcp.WithDestructiveHintAnnotation(true),
		mcp.WithIdempotentHintAnnotation(true),
	), actionRestartDNSHandler(c))

	addTool(s, mcp.NewTool("pihole_action_flush_logs",
		mcp.WithDescription("Permanently delete all DNS logs — purges last 24 hours from memory and database. Irreversible."),
		mcp.WithDestructiveHintAnnotation(true),
	), actionFlushLogsHandler(c))

	addTool(s, mcp.NewTool("pihole_action_flush_network",
		mcp.WithDescription("Permanently delete all network device records and associated addresses. Devices will be re-discovered over time."),
		mcp.WithDestructiveHintAnnotation(true),
	), actionFlushNetworkHandler(c))
}

func actionGravityHandler(c *pihole.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		resp, err := c.DoRaw(ctx, "POST", "/action/gravity", nil)
		if err != nil {
			return toolError("start gravity update", err), nil
		}
		defer func() { _ = resp.Body.Close() }()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to read gravity output: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("**Gravity update complete.**\n```\n%s\n```", string(body))), nil
	}
}

func actionRestartDNSHandler(c *pihole.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var result pihole.ActionResponse
		if err := c.Post(ctx, "/action/restartdns", nil, &result); err != nil {
			return toolError("restart DNS", err), nil
		}

		return mcp.NewToolResultText("**DNS restarted** successfully."), nil
	}
}

func actionFlushLogsHandler(c *pihole.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var result pihole.ActionResponse
		if err := c.Post(ctx, "/action/flush/logs", nil, &result); err != nil {
			return toolError("flush logs", err), nil
		}

		return mcp.NewToolResultText("**Logs flushed.** Last 24 hours purged from memory and database."), nil
	}
}

func actionFlushNetworkHandler(c *pihole.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var result pihole.ActionResponse
		if err := c.Post(ctx, "/action/flush/network", nil, &result); err != nil {
			return toolError("flush network table", err), nil
		}

		return mcp.NewToolResultText("**Network table flushed.** All device records removed."), nil
	}
}
