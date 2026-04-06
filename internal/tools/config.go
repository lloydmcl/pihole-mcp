package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/lloydmcl/pihole-mcp/internal/pihole"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// RegisterConfig registers Pi-hole configuration tools.
func RegisterConfig(s *server.MCPServer, c *pihole.Client) {
	addTool(s, mcp.NewTool("pihole_config_get",
		mcp.WithDescription("Get Pi-hole configuration. Specify a section (dns, webserver, dhcp, etc.) for a subset, or omit for full config."),
		mcp.WithString("section", mcp.Description("Config section: dns, webserver, dhcp, files, misc, etc.")),
		detailParam,
		mcp.WithReadOnlyHintAnnotation(true),
	), configGetHandler(c))

	addTool(s, mcp.NewTool("pihole_config_set",
		mcp.WithDescription("Modify Pi-hole configuration. Provide nested JSON properties to change. Changes take effect immediately."),
		mcp.WithString("config", mcp.Required(), mcp.Description("JSON config object, e.g. {\"dns\":{\"blocking\":{\"active\":true}}}")),
		mcp.WithIdempotentHintAnnotation(true),
	), configSetHandler(c))
}

func configGetHandler(c *pihole.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		path := "/config"
		if section := req.GetString("section", ""); section != "" {
			path += "/" + section
		}

		var result pihole.ConfigResponse
		if err := c.Get(ctx, path, &result); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get config: %v", err)), nil
		}

		detail := getDetail(req)

		if detail == "minimal" {
			sections := make([]string, 0, len(result.Config))
			for k := range result.Config {
				sections = append(sections, k)
			}
			return mcp.NewToolResultText(fmt.Sprintf("Config sections: %s", strings.Join(sections, ", "))), nil
		}

		configJSON, err := json.MarshalIndent(result.Config, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to format config: %v", err)), nil
		}

		var b strings.Builder
		b.WriteString("```json\n")
		b.Write(configJSON)
		b.WriteString("\n```")

		return mcp.NewToolResultText(b.String()), nil
	}
}

func configSetHandler(c *pihole.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		configStr, err := req.RequireString("config")
		if err != nil {
			return mcp.NewToolResultError("Parameter 'config' is required (JSON object)"), nil
		}

		var result pihole.ConfigResponse
		if err := c.Do(ctx, "PATCH", "/config", rawJSON(configStr), &result); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to update config: %v", err)), nil
		}

		configJSON, _ := json.MarshalIndent(result.Config, "", "  ")

		var b strings.Builder
		b.WriteString("**Config updated.**\n```json\n")
		b.Write(configJSON)
		b.WriteString("\n```")

		return mcp.NewToolResultText(b.String()), nil
	}
}
