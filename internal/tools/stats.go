package tools

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/lloydmcl/pihole-mcp/internal/format"
	"github.com/lloydmcl/pihole-mcp/internal/pihole"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// RegisterStats registers statistics and metrics tools.
func RegisterStats(s *server.MCPServer, c *pihole.Client) {
	addTool(s, mcp.NewTool("pihole_stats_summary",
		mcp.WithDescription("Overview of Pi-hole activity: queries, blocking rate, active clients, gravity size, and top query types. Start here for a health check."),
		detailParam,
		mcp.WithReadOnlyHintAnnotation(true),
	), statsSummaryHandler(c))

	addTool(s, mcp.NewTool("pihole_stats_top_domains",
		mcp.WithDescription("Top queried or top blocked domains ranked by count. Returns 10 by default, max 50."),
		mcp.WithBoolean("blocked", mcp.Description("True for top blocked, false/omit for top permitted.")),
		mcp.WithNumber("count", mcp.Description("Number of results (default 10, max 50).")),
		formatParam,
		mcp.WithReadOnlyHintAnnotation(true),
	), statsTopDomainsHandler(c))

	addTool(s, mcp.NewTool("pihole_stats_top_clients",
		mcp.WithDescription("Most active network clients ranked by query count. Use blocked=true for clients with most blocked queries."),
		mcp.WithBoolean("blocked", mcp.Description("True for most-blocked clients.")),
		mcp.WithNumber("count", mcp.Description("Number of results (default 10, max 50).")),
		formatParam,
		mcp.WithReadOnlyHintAnnotation(true),
	), statsTopClientsHandler(c))

	addTool(s, mcp.NewTool("pihole_stats_upstreams",
		mcp.WithDescription("Upstream DNS server performance: query counts, average response time, and standard deviation per server."),
		mcp.WithReadOnlyHintAnnotation(true),
	), statsUpstreamsHandler(c))

	addTool(s, mcp.NewTool("pihole_stats_query_types",
		mcp.WithDescription("Distribution of DNS query types (A, AAAA, MX, PTR, HTTPS, etc.) with counts."),
		mcp.WithReadOnlyHintAnnotation(true),
	), statsQueryTypesHandler(c))

	addTool(s, mcp.NewTool("pihole_stats_recent_blocked",
		mcp.WithDescription("Most recently blocked domains — useful for spotting new tracking domains or false positives in real-time."),
		mcp.WithNumber("count", mcp.Description("Number of domains (default 10).")),
		mcp.WithReadOnlyHintAnnotation(true),
	), statsRecentBlockedHandler(c))

	addTool(s, mcp.NewTool("pihole_stats_database",
		mcp.WithDescription("Long-term database statistics for a time range. Returns totals for queries, blocked, and clients."),
		mcp.WithNumber("from", mcp.Description("Start Unix timestamp.")),
		mcp.WithNumber("until", mcp.Description("End Unix timestamp.")),
		mcp.WithReadOnlyHintAnnotation(true),
	), statsDatabaseHandler(c))
}

func statsSummaryHandler(c *pihole.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var stats pihole.StatsSummary
		if err := c.Get(ctx, "/stats/summary", &stats); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get stats: %v", err)), nil
		}

		q := stats.Queries
		detail := getDetail(req)

		if detail == "minimal" {
			return mcp.NewToolResultText(fmt.Sprintf(
				"Queries: %s | Blocked: %s (%s) | Clients: %d | Gravity: %s domains",
				format.Number(q.Total), format.Number(q.Blocked), format.Percent(q.PercentBlocked),
				stats.Clients.Active, format.Number(stats.Gravity.DomainsBeingBlocked))), nil
		}

		var b strings.Builder
		fmt.Fprintf(&b, "**Queries:** %s total, %s blocked (%s), %s cached, %s forwarded\n",
			format.Number(q.Total), format.Number(q.Blocked), format.Percent(q.PercentBlocked),
			format.Number(q.Cached), format.Number(q.Forwarded))
		fmt.Fprintf(&b, "**Clients:** %d active of %d total\n", stats.Clients.Active, stats.Clients.Total)
		fmt.Fprintf(&b, "**Gravity:** %s domains", format.Number(stats.Gravity.DomainsBeingBlocked))
		if stats.Gravity.LastUpdate > 0 {
			fmt.Fprintf(&b, " (updated %s)", format.Timestamp(float64(stats.Gravity.LastUpdate)))
		}
		b.WriteString("\n")

		if len(q.Types) > 0 {
			sorted := sortMapByValue(q.Types)
			limit := 5
			if detail == "full" {
				limit = len(sorted)
			}
			b.WriteString("**Query types:** ")
			parts := make([]string, 0, limit)
			for _, kv := range sorted[:min(limit, len(sorted))] {
				if kv.Value > 0 || detail == "full" {
					parts = append(parts, fmt.Sprintf("%s=%s", kv.Key, format.Number(kv.Value)))
				}
			}
			b.WriteString(strings.Join(parts, ", "))
			b.WriteString("\n")
		}

		if detail == "full" {
			if len(q.Status) > 0 {
				sorted := sortMapByValue(q.Status)
				b.WriteString("**Status breakdown:** ")
				parts := make([]string, 0, len(sorted))
				for _, kv := range sorted {
					if kv.Value > 0 {
						parts = append(parts, fmt.Sprintf("%s=%s", kv.Key, format.Number(kv.Value)))
					}
				}
				b.WriteString(strings.Join(parts, ", "))
				b.WriteString("\n")
			}
			fmt.Fprintf(&b, "**Unique domains:** %s | **Frequency:** %.1f queries/sec\n",
				format.Number(q.UniqueDomains), q.Frequency)
		}

		return mcp.NewToolResultText(b.String()), nil
	}
}

func statsTopDomainsHandler(c *pihole.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		count := int(req.GetFloat("count", 10))
		if count > 50 {
			count = 50
		}
		blocked := req.GetBool("blocked", false)

		path := fmt.Sprintf("/stats/top_domains?count=%d&blocked=%t", count, blocked)
		var result pihole.TopItems
		if err := c.Get(ctx, path, &result); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get top domains: %v", err)), nil
		}

		headers := []string{"Rank", "Domain", "Queries"}
		rows := make([][]string, 0, len(result.Domains))
		for i, d := range result.Domains {
			rows = append(rows, []string{fmt.Sprintf("%d", i+1), d.Domain, format.Number(d.Count)})
		}

		if wantCSV(req) {
			return mcp.NewToolResultText(format.CSV(headers, rows)), nil
		}

		label := "permitted"
		if blocked {
			label = "blocked"
		}
		var b strings.Builder
		fmt.Fprintf(&b, "**Top %s domains:**\n", label)
		for i, d := range result.Domains {
			fmt.Fprintf(&b, "%d. %s — %s\n", i+1, d.Domain, format.Number(d.Count))
		}
		return mcp.NewToolResultText(b.String()), nil
	}
}

func statsTopClientsHandler(c *pihole.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		count := int(req.GetFloat("count", 10))
		if count > 50 {
			count = 50
		}
		blocked := req.GetBool("blocked", false)

		path := fmt.Sprintf("/stats/top_clients?count=%d&blocked=%t", count, blocked)
		var result pihole.TopItems
		if err := c.Get(ctx, path, &result); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get top clients: %v", err)), nil
		}

		headers := []string{"Rank", "IP", "Name", "Queries"}
		rows := make([][]string, 0, len(result.Clients))
		for i, cl := range result.Clients {
			rows = append(rows, []string{fmt.Sprintf("%d", i+1), cl.IP, cl.Name, format.Number(cl.Count)})
		}

		if wantCSV(req) {
			return mcp.NewToolResultText(format.CSV(headers, rows)), nil
		}

		var b strings.Builder
		b.WriteString("**Top clients:**\n")
		for i, cl := range result.Clients {
			name := cl.Name
			if name == "" {
				name = cl.IP
			} else {
				name = fmt.Sprintf("%s (%s)", cl.Name, cl.IP)
			}
			fmt.Fprintf(&b, "%d. %s — %s\n", i+1, name, format.Number(cl.Count))
		}
		return mcp.NewToolResultText(b.String()), nil
	}
}

func statsUpstreamsHandler(c *pihole.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var result pihole.Upstreams
		if err := c.Get(ctx, "/stats/upstreams", &result); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get upstreams: %v", err)), nil
		}

		var b strings.Builder
		fmt.Fprintf(&b, "**Forwarded:** %s of %s total queries\n",
			format.Number(result.ForwardedQueries), format.Number(result.TotalQueries))
		for _, u := range result.Upstreams {
			name := format.StringOr(u.Name, format.StringOr(u.IP, "cache"))
			fmt.Fprintf(&b, "- %s:%d — %s queries, avg %s\n",
				name, u.Port, format.Number(u.Count), format.ResponseTime(u.Statistics.Response))
		}
		return mcp.NewToolResultText(b.String()), nil
	}
}

func statsQueryTypesHandler(c *pihole.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var result pihole.QueryTypes
		if err := c.Get(ctx, "/stats/query_types", &result); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get query types: %v", err)), nil
		}

		sorted := sortMapByValue(result.Types)
		var b strings.Builder
		b.WriteString("**Query types:**\n")
		for _, kv := range sorted {
			if kv.Value > 0 {
				fmt.Fprintf(&b, "- %s: %s\n", kv.Key, format.Number(kv.Value))
			}
		}
		return mcp.NewToolResultText(b.String()), nil
	}
}

func statsRecentBlockedHandler(c *pihole.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		count := int(req.GetFloat("count", 10))
		path := fmt.Sprintf("/stats/recent_blocked?count=%d", count)

		var result pihole.RecentBlocked
		if err := c.Get(ctx, path, &result); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get recent blocked: %v", err)), nil
		}

		if len(result.Blocked) == 0 {
			return mcp.NewToolResultText("No recently blocked domains."), nil
		}

		var b strings.Builder
		b.WriteString("**Recently blocked:**\n")
		for i, domain := range result.Blocked {
			fmt.Fprintf(&b, "%d. %s\n", i+1, domain)
		}
		return mcp.NewToolResultText(b.String()), nil
	}
}

func statsDatabaseHandler(c *pihole.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		now := float64(time.Now().Unix())
		from := req.GetFloat("from", now-86400) // Default: 24 hours ago
		until := req.GetFloat("until", now)

		params := map[string]string{
			"from":  fmt.Sprintf("%.0f", from),
			"until": fmt.Sprintf("%.0f", until),
		}

		path := "/stats/database/summary" + format.QueryParams(params)
		var result pihole.DatabaseSummary
		if err := c.Get(ctx, path, &result); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get database stats: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf(
			"**Database summary:** %s queries, %s blocked (%s), %d clients",
			format.Number(result.SumQueries), format.Number(result.SumBlocked),
			format.Percent(result.PercentBlocked), result.TotalClients)), nil
	}
}

type kv struct {
	Key   string
	Value int
}

func sortMapByValue(m map[string]int) []kv {
	sorted := make([]kv, 0, len(m))
	for k, v := range m {
		sorted = append(sorted, kv{k, v})
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Value > sorted[j].Value
	})
	return sorted
}
