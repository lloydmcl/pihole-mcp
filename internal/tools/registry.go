// Package tools registers all MCP tool definitions for the Pi-hole MCP server.
package tools

import (
	"github.com/hexamatic/pihole-mcp/internal/pihole"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// RegisterAll registers every tool category on the MCP server.
// Each tool handler is wrapped with OpenTelemetry tracing (noop when OTel is not configured).
func RegisterAll(s *server.MCPServer, c *pihole.Client) {
	RegisterDNS(s, c)
	RegisterStats(s, c)
	RegisterInfo(s, c)
	RegisterQueries(s, c)
	RegisterHistory(s, c)
	RegisterSearch(s, c)
	RegisterDomains(s, c)
	RegisterGroups(s, c)
	RegisterClients(s, c)
	RegisterLists(s, c)
	RegisterConfig(s, c)
	RegisterActions(s, c)
	RegisterNetwork(s, c)
	RegisterDHCP(s, c)
	RegisterLogs(s, c)
	RegisterTeleporter(s, c)
	RegisterAuth(s, c)
}

// addTool registers a tool with tracing middleware.
func addTool(s *server.MCPServer, tool mcp.Tool, handler server.ToolHandlerFunc) {
	s.AddTool(tool, withTracing(tool.Name, handler))
}
