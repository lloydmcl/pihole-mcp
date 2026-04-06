// Package server constructs and configures the MCP server.
package server

import (
	"github.com/lloydmcl/pihole-mcp/internal/pihole"
	"github.com/lloydmcl/pihole-mcp/internal/prompts"
	"github.com/lloydmcl/pihole-mcp/internal/resources"
	"github.com/lloydmcl/pihole-mcp/internal/tools"
	"github.com/mark3labs/mcp-go/server"
)

// Version is set at build time via ldflags.
var Version = "dev"

// New creates a configured MCP server with all Pi-hole tools, resources,
// and prompts registered.
func New(client *pihole.Client) *server.MCPServer {
	s := server.NewMCPServer(
		"pihole-mcp",
		Version,
		server.WithToolCapabilities(false),
		server.WithResourceCapabilities(false, false),
		server.WithPromptCapabilities(false),
		server.WithLogging(),
		server.WithRecovery(),
		server.WithInstructions(
			"Pi-hole v6 DNS management server. "+
				"Start with pihole_stats_summary for a quick overview. "+
				"Use pihole_search_domains before pihole_domains_add to check for duplicates. "+
				"After pihole_lists_add or pihole_lists_delete, run pihole_action_gravity_update to apply changes. "+
				"Use pihole_queries_suggestions to discover valid filter values for pihole_queries_search. "+
				"Tools accept optional 'detail' (minimal/normal/full) and 'format' (text/csv) parameters.",
		),
	)

	tools.RegisterAll(s, client)
	resources.RegisterAll(s, client)
	prompts.RegisterAll(s)

	return s
}
