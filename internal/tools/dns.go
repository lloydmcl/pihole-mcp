package tools

import (
	"context"
	"fmt"

	"github.com/lloydmcl/pihole-mcp/internal/pihole"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// RegisterDNS registers DNS blocking control tools.
func RegisterDNS(s *server.MCPServer, c *pihole.Client) {
	addTool(s,
		mcp.NewTool("pihole_dns_get_blocking",
			mcp.WithDescription("Get the current DNS blocking status and any active timer for temporary blocking changes."),
			mcp.WithReadOnlyHintAnnotation(true),
		),
		dnsGetBlockingHandler(c),
	)

	addTool(s,
		mcp.NewTool("pihole_dns_set_blocking",
			mcp.WithDescription("Enable or disable DNS blocking. Set a timer (seconds) to automatically revert after a duration."),
			mcp.WithBoolean("blocking",
				mcp.Required(),
				mcp.Description("True to enable blocking, false to disable."),
			),
			mcp.WithNumber("timer",
				mcp.Description("Seconds until automatic revert. Omit for permanent change."),
			),
			mcp.WithDestructiveHintAnnotation(true),
			mcp.WithIdempotentHintAnnotation(true),
		),
		dnsSetBlockingHandler(c),
	)
}

func dnsGetBlockingHandler(c *pihole.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var status pihole.BlockingStatus
		if err := c.Get(ctx, "/dns/blocking", &status); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get blocking status: %v", err)), nil
		}

		text := fmt.Sprintf("**Blocking:** %s", status.Blocking)
		if status.Timer != nil {
			text += fmt.Sprintf(" (%.0fs remaining)", *status.Timer)
		}

		return mcp.NewToolResultText(text), nil
	}
}

func dnsSetBlockingHandler(c *pihole.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		blocking, err := req.RequireBool("blocking")
		if err != nil {
			return mcp.NewToolResultError("Parameter 'blocking' is required (true or false)"), nil
		}

		body := pihole.BlockingRequest{Blocking: blocking}

		timer := req.GetFloat("timer", 0)
		if timer > 0 {
			body.Timer = &timer
		}

		var status pihole.BlockingStatus
		if err := c.Post(ctx, "/dns/blocking", body, &status); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to set blocking: %v", err)), nil
		}

		action := "enabled"
		if !blocking {
			action = "disabled"
		}
		text := fmt.Sprintf("**Blocking %s.** Status: %s", action, status.Blocking)
		if status.Timer != nil {
			text += fmt.Sprintf(" (reverts in %.0fs)", *status.Timer)
		}

		return mcp.NewToolResultText(text), nil
	}
}
