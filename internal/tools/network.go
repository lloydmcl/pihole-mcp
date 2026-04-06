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

// RegisterNetwork registers network information tools.
func RegisterNetwork(s *server.MCPServer, c *pihole.Client) {
	addTool(s, mcp.NewTool("pihole_network_devices",
		mcp.WithDescription("Devices seen on the network: MAC, IPs, hostnames, vendor, query count, and first/last seen timestamps. Returns 20 by default."),
		mcp.WithNumber("max_devices", mcp.Description("Max devices (default 20).")),
		mcp.WithNumber("max_addresses", mcp.Description("Max IPs per device (default 3).")),
		detailParam,
		formatParam,
		mcp.WithReadOnlyHintAnnotation(true),
	), networkDevicesHandler(c))

	addTool(s, mcp.NewTool("pihole_network_gateway",
		mcp.WithDescription("Network gateway details: address, interface, address family, and local interface IPs."),
		mcp.WithReadOnlyHintAnnotation(true),
	), networkGatewayHandler(c))

	addTool(s, mcp.NewTool("pihole_network_info",
		mcp.WithDescription("Network routing table and interface information including addresses, speeds, and traffic statistics."),
		mcp.WithReadOnlyHintAnnotation(true),
	), networkInfoHandler(c))
}

func networkDevicesHandler(c *pihole.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		maxDev := int(req.GetFloat("max_devices", 20))
		maxAddr := int(req.GetFloat("max_addresses", 3))

		path := fmt.Sprintf("/network/devices?max_devices=%d&max_addresses=%d", maxDev, maxAddr)
		var result pihole.NetworkDevicesResponse
		if err := c.Get(ctx, path, &result); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get devices: %v", err)), nil
		}

		if len(result.Devices) == 0 {
			return mcp.NewToolResultText("No network devices found."), nil
		}

		detail := getDetail(req)

		if detail == "minimal" {
			return mcp.NewToolResultText(fmt.Sprintf("%d network devices.", len(result.Devices))), nil
		}

		if wantCSV(req) {
			headers := []string{"MAC", "Vendor", "IPs", "Queries", "LastQuery"}
			rows := make([][]string, 0, len(result.Devices))
			for _, d := range result.Devices {
				ips := make([]string, 0, len(d.IPs))
				for _, ip := range d.IPs {
					ips = append(ips, ip.IP)
				}
				rows = append(rows, []string{d.HWAddr, format.StringOr(d.MacVendor, ""), strings.Join(ips, ";"), format.Number(d.NumQueries), format.Timestamp(float64(d.LastQuery))})
			}
			return mcp.NewToolResultText(format.CSV(headers, rows)), nil
		}

		var b strings.Builder
		fmt.Fprintf(&b, "**%d devices:**\n", len(result.Devices))
		for _, d := range result.Devices {
			vendor := format.StringOr(d.MacVendor, "unknown")
			ips := make([]string, 0, len(d.IPs))
			for _, ip := range d.IPs {
				name := format.StringOr(ip.Name, "")
				if name != "" {
					ips = append(ips, fmt.Sprintf("%s (%s)", ip.IP, name))
				} else {
					ips = append(ips, ip.IP)
				}
			}
			fmt.Fprintf(&b, "- %s [%s] — %s, %s queries, last %s\n",
				d.HWAddr, vendor, strings.Join(ips, ", "),
				format.Number(d.NumQueries), format.Timestamp(float64(d.LastQuery)))
		}

		return mcp.NewToolResultText(b.String()), nil
	}
}

func networkGatewayHandler(c *pihole.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var result pihole.GatewayResponse
		if err := c.Get(ctx, "/network/gateway", &result); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get gateway: %v", err)), nil
		}

		if len(result.Gateway) == 0 {
			return mcp.NewToolResultText("No gateway information available."), nil
		}

		var b strings.Builder
		b.WriteString("**Gateway:**\n")
		for _, g := range result.Gateway {
			fmt.Fprintf(&b, "- %s via %s on %s (local: %s)\n",
				g.Address, g.Family, g.Interface, strings.Join(g.Local, ", "))
		}

		return mcp.NewToolResultText(b.String()), nil
	}
}

func networkInfoHandler(c *pihole.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var routes map[string]any
		var interfaces map[string]any

		_ = c.Get(ctx, "/network/routes", &routes)
		_ = c.Get(ctx, "/network/interfaces", &interfaces)

		var b strings.Builder

		if routes != nil {
			if routeList, ok := routes["routes"].([]any); ok {
				fmt.Fprintf(&b, "**%d routes:**\n", len(routeList))
				for _, r := range routeList {
					if rm, ok := r.(map[string]any); ok {
						fmt.Fprintf(&b, "- %s via %s (%s)\n",
							mapGetStr(rm, "dst", "default"),
							mapGetStr(rm, "gateway", "direct"),
							mapGetStr(rm, "oif", ""))
					}
				}
			}
		}

		if interfaces != nil {
			if ifList, ok := interfaces["interfaces"].([]any); ok {
				fmt.Fprintf(&b, "**%d interfaces:**\n", len(ifList))
				for _, i := range ifList {
					if im, ok := i.(map[string]any); ok {
						fmt.Fprintf(&b, "- %s (%s, %s)\n",
							mapGetStr(im, "name", "?"),
							mapGetStr(im, "type", "?"),
							mapGetStr(im, "state", "?"))
					}
				}
			}
		}

		if b.Len() == 0 {
			return mcp.NewToolResultText("No network info available."), nil
		}

		return mcp.NewToolResultText(b.String()), nil
	}
}

func mapGetStr(m map[string]any, key, fallback string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok && s != "" {
			return s
		}
	}
	return fallback
}
