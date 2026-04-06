package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/lloydmcl/pihole-mcp/internal/format"
	"github.com/lloydmcl/pihole-mcp/internal/pihole"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// RegisterDomains registers domain management tools.
func RegisterDomains(s *server.MCPServer, c *pihole.Client) {
	addTool(s, mcp.NewTool("pihole_domains_list",
		mcp.WithDescription("List domains on allow/deny lists. Filter by type (allow/deny) and kind (exact/regex). Use pihole_search_domains for cross-list search."),
		mcp.WithString("type", mcp.Description("Filter: 'allow' or 'deny'."), mcp.Enum("allow", "deny")),
		mcp.WithString("kind", mcp.Description("Filter: 'exact' or 'regex'."), mcp.Enum("exact", "regex")),
		detailParam,
		formatParam,
		mcp.WithReadOnlyHintAnnotation(true),
	), domainsListHandler(c))

	addTool(s, mcp.NewTool("pihole_domains_add",
		mcp.WithDescription("Add domains to an allow or deny list. Supports bulk add via comma-separated domains. Use pihole_search_domains first to avoid duplicates."),
		mcp.WithString("type", mcp.Required(), mcp.Description("'allow' or 'deny'."), mcp.Enum("allow", "deny")),
		mcp.WithString("kind", mcp.Required(), mcp.Description("'exact' or 'regex'."), mcp.Enum("exact", "regex")),
		mcp.WithString("domain", mcp.Required(), mcp.Description("Domain(s) to add (comma-separated for bulk).")),
		mcp.WithString("comment", mcp.Description("Comment for the entry.")),
		mcp.WithBoolean("enabled", mcp.Description("Enabled state (default true).")),
		mcp.WithOpenWorldHintAnnotation(true),
	), domainsAddHandler(c))

	addTool(s, mcp.NewTool("pihole_domains_update",
		mcp.WithDescription("Update a domain entry's comment, enabled status, or move it between allow/deny lists."),
		mcp.WithString("type", mcp.Required(), mcp.Description("Current type: 'allow' or 'deny'."), mcp.Enum("allow", "deny")),
		mcp.WithString("kind", mcp.Required(), mcp.Description("Current kind: 'exact' or 'regex'."), mcp.Enum("exact", "regex")),
		mcp.WithString("domain", mcp.Required(), mcp.Description("Domain to update.")),
		mcp.WithString("comment", mcp.Description("Updated comment.")),
		mcp.WithBoolean("enabled", mcp.Description("Updated enabled status.")),
		mcp.WithIdempotentHintAnnotation(true),
	), domainsUpdateHandler(c))

	addTool(s, mcp.NewTool("pihole_domains_delete",
		mcp.WithDescription("Remove a domain from an allow or deny list. Requires the exact type, kind, and domain."),
		mcp.WithString("type", mcp.Required(), mcp.Description("'allow' or 'deny'."), mcp.Enum("allow", "deny")),
		mcp.WithString("kind", mcp.Required(), mcp.Description("'exact' or 'regex'."), mcp.Enum("exact", "regex")),
		mcp.WithString("domain", mcp.Required(), mcp.Description("Domain to remove.")),
		mcp.WithDestructiveHintAnnotation(true),
		mcp.WithIdempotentHintAnnotation(true),
	), domainsDeleteHandler(c))

	addTool(s, mcp.NewTool("pihole_domains_batch_delete",
		mcp.WithDescription("Remove multiple domains at once. Each item needs domain, type (allow/deny), and kind (exact/regex)."),
		mcp.WithString("items", mcp.Required(), mcp.Description("JSON array: [{\"item\":\"domain\",\"type\":\"deny\",\"kind\":\"exact\"}]")),
		mcp.WithDestructiveHintAnnotation(true),
	), domainsBatchDeleteHandler(c))
}

func domainsListHandler(c *pihole.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		path := "/domains"
		if t := req.GetString("type", ""); t != "" {
			path += "/" + t
			if k := req.GetString("kind", ""); k != "" {
				path += "/" + k
			}
		}

		var result pihole.DomainsResponse
		if err := c.Get(ctx, path, &result); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list domains: %v", err)), nil
		}

		if len(result.Domains) == 0 {
			return mcp.NewToolResultText("No domains found."), nil
		}

		detail := getDetail(req)

		if detail == "minimal" {
			return mcp.NewToolResultText(fmt.Sprintf("%d domains.", len(result.Domains))), nil
		}

		if wantCSV(req) {
			headers := []string{"Domain", "Type", "Kind", "Enabled", "Comment"}
			rows := make([][]string, 0, len(result.Domains))
			for _, d := range result.Domains {
				rows = append(rows, []string{d.Domain, d.Type, d.Kind, format.Bool(d.Enabled), d.Comment})
			}
			return mcp.NewToolResultText(format.CSV(headers, rows)), nil
		}

		var b strings.Builder
		fmt.Fprintf(&b, "**%d domains:**\n", len(result.Domains))
		for _, d := range result.Domains {
			status := "enabled"
			if !d.Enabled {
				status = "disabled"
			}
			fmt.Fprintf(&b, "- %s (%s/%s, %s)", d.Domain, d.Type, d.Kind, status)
			if d.Comment != "" {
				fmt.Fprintf(&b, " — %s", d.Comment)
			}
			if detail == "full" {
				fmt.Fprintf(&b, " [id=%d, groups=%v]", d.ID, d.Groups)
			}
			b.WriteString("\n")
		}

		return mcp.NewToolResultText(b.String()), nil
	}
}

func domainsAddHandler(c *pihole.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		t, _ := req.RequireString("type")
		k, _ := req.RequireString("kind")
		domain, _ := req.RequireString("domain")

		body := map[string]any{"domain": domain}
		if comment := req.GetString("comment", ""); comment != "" {
			body["comment"] = comment
		}
		if !req.GetBool("enabled", true) {
			body["enabled"] = false
		}

		path := fmt.Sprintf("/domains/%s/%s", t, k)
		var result pihole.DomainsResponse
		if err := c.Post(ctx, path, body, &result); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to add domain: %v", err)), nil
		}

		var b strings.Builder
		b.WriteString("**Domain added.**\n")
		writeProcessedResult(&b, result.Processed)

		return mcp.NewToolResultText(b.String()), nil
	}
}

func domainsUpdateHandler(c *pihole.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		t, _ := req.RequireString("type")
		k, _ := req.RequireString("kind")
		domain, _ := req.RequireString("domain")

		body := make(map[string]any)
		if comment := req.GetString("comment", ""); comment != "" {
			body["comment"] = comment
		}
		if enabled := req.GetBool("enabled", true); !enabled {
			body["enabled"] = false
		}

		path := fmt.Sprintf("/domains/%s/%s/%s", t, k, domain)
		var result pihole.DomainsResponse
		if err := c.Put(ctx, path, body, &result); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to update domain: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("**Updated** %s on %s/%s list.", domain, t, k)), nil
	}
}

func domainsDeleteHandler(c *pihole.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		t, _ := req.RequireString("type")
		k, _ := req.RequireString("kind")
		domain, _ := req.RequireString("domain")

		path := fmt.Sprintf("/domains/%s/%s/%s", t, k, domain)
		if err := c.Delete(ctx, path); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to delete domain: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("**Deleted** %s from %s/%s list.", domain, t, k)), nil
	}
}

func domainsBatchDeleteHandler(c *pihole.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		items, err := req.RequireString("items")
		if err != nil {
			return mcp.NewToolResultError("Parameter 'items' is required (JSON array)"), nil
		}

		if err := c.Post(ctx, "/domains:batchDelete", rawJSON(items), nil); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Batch delete failed: %v", err)), nil
		}

		return mcp.NewToolResultText("**Batch delete completed.**"), nil
	}
}

func writeProcessedResult(b *strings.Builder, p *pihole.ProcessedResult) {
	if p == nil {
		return
	}
	for _, s := range p.Success {
		fmt.Fprintf(b, "- Added: %s\n", s.Item)
	}
	for _, e := range p.Errors {
		fmt.Fprintf(b, "- Failed: %s (%s)\n", e.Item, e.Error)
	}
}

type rawJSON string

func (r rawJSON) MarshalJSON() ([]byte, error) {
	return []byte(r), nil
}
