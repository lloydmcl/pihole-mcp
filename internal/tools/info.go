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

// RegisterInfo registers system information tools.
func RegisterInfo(s *server.MCPServer, c *pihole.Client) {
	addTool(s, mcp.NewTool("pihole_info_system",
		mcp.WithDescription("System health: hostname, OS, CPU/memory/disk usage, load averages, temperature, and DNS service status."),
		detailParam,
		mcp.WithReadOnlyHintAnnotation(true),
	), infoSystemHandler(c))

	addTool(s, mcp.NewTool("pihole_info_version",
		mcp.WithDescription("Pi-hole component versions: core, FTL engine, web interface, and Docker tag if applicable."),
		mcp.WithReadOnlyHintAnnotation(true),
	), infoVersionHandler(c))

	addTool(s, mcp.NewTool("pihole_info_database",
		mcp.WithDescription("Query database details: file size, total stored queries, and SQLite version."),
		mcp.WithReadOnlyHintAnnotation(true),
	), infoDatabaseHandler(c))

	addTool(s, mcp.NewTool("pihole_info_messages",
		mcp.WithDescription("FTL diagnostic messages — warnings about DNS resolution failures, database issues, or configuration problems."),
		mcp.WithReadOnlyHintAnnotation(true),
	), infoMessagesHandler(c))

	addTool(s, mcp.NewTool("pihole_info_client",
		mcp.WithDescription("Information about the requesting client's IP address and connection. Does not require authentication."),
		mcp.WithReadOnlyHintAnnotation(true),
	), infoClientHandler(c))
}

func infoSystemHandler(c *pihole.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var sysInfo pihole.SystemInfo
		var hostInfo pihole.HostInfo
		var sensors pihole.SensorsInfo

		if err := c.Get(ctx, "/info/system", &sysInfo); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get system info: %v", err)), nil
		}
		_ = c.Get(ctx, "/info/host", &hostInfo)
		_ = c.Get(ctx, "/info/sensors", &sensors)

		sys := sysInfo.System
		detail := getDetail(req)

		if detail == "minimal" {
			mem := format.SizeWithUnit(sys.Memory.RAM.Used, sys.Memory.RAM.Unit)
			dns := "up"
			if !sys.DNS.Running {
				dns = "down"
			}
			return mcp.NewToolResultText(fmt.Sprintf(
				"Load: %.2f | Memory: %s (%.0f%%) | DNS: %s | Uptime: %s",
				sys.Load[0], mem, sys.Memory.RAM.Perc, dns, format.Duration(float64(sys.Uptime)))), nil
		}

		var b strings.Builder
		if hostInfo.Host.Name != "" {
			fmt.Fprintf(&b, "**Host:** %s (%s %s)\n", hostInfo.Host.Name, hostInfo.Host.OS, hostInfo.Host.Arch)
		}
		fmt.Fprintf(&b, "**Uptime:** %s\n", format.Duration(float64(sys.Uptime)))
		fmt.Fprintf(&b, "**Load:** %.2f, %.2f, %.2f\n", sys.Load[0], sys.Load[1], sys.Load[2])
		fmt.Fprintf(&b, "**CPU:** %d cores, %.1f%% used\n", sys.CPU.Nprocs, sys.CPU.Perc)
		fmt.Fprintf(&b, "**Memory:** %s / %s (%.1f%%)\n",
			format.SizeWithUnit(sys.Memory.RAM.Used, sys.Memory.RAM.Unit),
			format.SizeWithUnit(sys.Memory.RAM.Total, sys.Memory.RAM.Unit),
			sys.Memory.RAM.Perc)
		fmt.Fprintf(&b, "**Disk:** %s / %s (%.1f%%)\n",
			format.SizeWithUnit(sys.Disk.Used, sys.Disk.Unit),
			format.SizeWithUnit(sys.Disk.Total, sys.Disk.Unit),
			sys.Disk.Perc)

		if sys.DNS.Running {
			b.WriteString("**DNS:** running\n")
		} else {
			b.WriteString("**DNS:** not running (expected in Docker — Pi-hole manages DNS internally)\n")
		}

		for _, t := range sensors.Sensors.Temperatures {
			fmt.Fprintf(&b, "**%s:** %.1f%s\n", t.Name, t.Value, t.Unit)
		}

		if detail == "full" && hostInfo.Host.Kernel != "" {
			fmt.Fprintf(&b, "**Kernel:** %s\n", hostInfo.Host.Kernel)
			fmt.Fprintf(&b, "**Domain:** %s\n", format.ValueOr(hostInfo.Host.Domain, "N/A"))
		}

		return mcp.NewToolResultText(b.String()), nil
	}
}

func infoVersionHandler(c *pihole.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var ver pihole.VersionInfo
		if err := c.Get(ctx, "/info/version", &ver); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get version: %v", err)), nil
		}

		var b strings.Builder
		fmt.Fprintf(&b, "**Core:** %s (%s)\n", ver.Version.Core.Local.Version, ver.Version.Core.Local.Branch)
		fmt.Fprintf(&b, "**FTL:** %s (%s)\n", ver.Version.FTL.Local.Version, ver.Version.FTL.Local.Branch)
		fmt.Fprintf(&b, "**Web:** %s (%s)\n", ver.Version.Web.Local.Version, ver.Version.Web.Local.Branch)
		if ver.Version.Docker.Local != "" {
			fmt.Fprintf(&b, "**Docker:** %s\n", ver.Version.Docker.Local)
		}

		return mcp.NewToolResultText(b.String()), nil
	}
}

func infoDatabaseHandler(c *pihole.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var db pihole.DatabaseInfo
		if err := c.Get(ctx, "/info/database", &db); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get database info: %v", err)), nil
		}

		size := format.SizeWithUnit(db.Database.Size, db.Database.Unit)
		sqlite := format.ValueOr(db.Database.SQLite, "N/A")

		return mcp.NewToolResultText(fmt.Sprintf(
			"**Size:** %s | **Queries:** %s | **SQLite:** %s",
			size, format.Number(db.Database.Queries), sqlite)), nil
	}
}

func infoMessagesHandler(c *pihole.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var msgs pihole.MessagesResponse
		if err := c.Get(ctx, "/info/messages", &msgs); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get messages: %v", err)), nil
		}

		if len(msgs.Messages) == 0 {
			return mcp.NewToolResultText("No diagnostic messages."), nil
		}

		var b strings.Builder
		fmt.Fprintf(&b, "**%d messages:**\n", len(msgs.Messages))
		for _, m := range msgs.Messages {
			fmt.Fprintf(&b, "- [%s] %s (%s)\n", m.Type, m.Message, format.Timestamp(float64(m.Timestamp)))
		}

		return mcp.NewToolResultText(b.String()), nil
	}
}

func infoClientHandler(c *pihole.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var info pihole.ClientInfo
		if err := c.Get(ctx, "/info/client", &info); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get client info: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf(
			"**Remote address:** %s | **HTTP:** %s | **Method:** %s",
			info.RemoteAddr, info.HTTPVersion, info.Method)), nil
	}
}
