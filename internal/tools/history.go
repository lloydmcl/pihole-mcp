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

// RegisterHistory registers activity history tools.
func RegisterHistory(s *server.MCPServer, c *pihole.Client) {
	addTool(s, mcp.NewTool("pihole_history_graph",
		mcp.WithDescription("Query activity over time: total, cached, blocked, and forwarded per slot. Add from/until for long-term database queries."),
		mcp.WithNumber("from", mcp.Description("Start Unix timestamp for long-term mode.")),
		mcp.WithNumber("until", mcp.Description("End Unix timestamp for long-term mode.")),
		mcp.WithReadOnlyHintAnnotation(true),
	), historyGraphHandler(c))

	addTool(s, mcp.NewTool("pihole_history_clients",
		mcp.WithDescription("Per-client query activity over time for top N clients. Add from/until for long-term database queries."),
		mcp.WithNumber("count", mcp.Description("Max clients to return (default 10, 0=all).")),
		mcp.WithNumber("from", mcp.Description("Start Unix timestamp for long-term mode.")),
		mcp.WithNumber("until", mcp.Description("End Unix timestamp for long-term mode.")),
		mcp.WithReadOnlyHintAnnotation(true),
	), historyClientsHandler(c))
}

func historyGraphHandler(c *pihole.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		from := req.GetFloat("from", 0)
		until := req.GetFloat("until", 0)

		var path string
		if from > 0 || until > 0 {
			params := make(map[string]string)
			if from > 0 {
				params["from"] = fmt.Sprintf("%.0f", from)
			}
			if until > 0 {
				params["until"] = fmt.Sprintf("%.0f", until)
			}
			path = "/history/database" + format.QueryParams(params)
		} else {
			path = "/history"
		}

		var result pihole.HistoryResponse
		if err := c.Get(ctx, path, &result); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get history: %v", err)), nil
		}

		if len(result.History) == 0 {
			return mcp.NewToolResultText("No history data available."), nil
		}

		var totalQ, totalBlocked, totalCached, totalFwd int
		for _, h := range result.History {
			totalQ += h.Total
			totalBlocked += h.Blocked
			totalCached += h.Cached
			totalFwd += h.Forwarded
		}

		var b strings.Builder
		fmt.Fprintf(&b, "**%d data points** (%s to %s)\n",
			len(result.History),
			format.Timestamp(result.History[0].Timestamp),
			format.Timestamp(result.History[len(result.History)-1].Timestamp))
		fmt.Fprintf(&b, "Total: %s queries, %s blocked, %s cached, %s forwarded\n",
			format.Number(totalQ), format.Number(totalBlocked),
			format.Number(totalCached), format.Number(totalFwd))

		return mcp.NewToolResultText(b.String()), nil
	}
}

func historyClientsHandler(c *pihole.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		count := int(req.GetFloat("count", 10))
		from := req.GetFloat("from", 0)
		until := req.GetFloat("until", 0)

		params := make(map[string]string)
		if count > 0 {
			params["N"] = fmt.Sprintf("%d", count)
		}

		var path string
		if from > 0 || until > 0 {
			if from > 0 {
				params["from"] = fmt.Sprintf("%.0f", from)
			}
			if until > 0 {
				params["until"] = fmt.Sprintf("%.0f", until)
			}
			path = "/history/database/clients" + format.QueryParams(params)
		} else {
			path = "/history/clients" + format.QueryParams(params)
		}

		var result pihole.ClientHistoryResponse
		if err := c.Get(ctx, path, &result); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get client history: %v", err)), nil
		}

		var b strings.Builder
		fmt.Fprintf(&b, "**%d clients:**\n", len(result.Clients))
		for ip, info := range result.Clients {
			name := format.StringOr(info.Name, "")
			if name != "" {
				fmt.Fprintf(&b, "- %s (%s) — %s queries\n", name, ip, format.Number(info.Total))
			} else {
				fmt.Fprintf(&b, "- %s — %s queries\n", ip, format.Number(info.Total))
			}
		}

		return mcp.NewToolResultText(b.String()), nil
	}
}
