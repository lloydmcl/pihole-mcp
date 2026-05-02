package tools

import (
	"fmt"
	"strings"
	"time"

	"github.com/hexamatic/pihole-mcp/internal/pihole"
	"github.com/mark3labs/mcp-go/mcp"
)

// getTimeRange extracts from/until Unix timestamp parameters with a default window.
// If defaultWindow is non-zero, from defaults to (now - defaultWindow) and until defaults to now.
// Returns the formatted string values ready for query parameters.
func getTimeRange(req mcp.CallToolRequest, defaultWindow time.Duration) (from, until string) {
	now := float64(time.Now().Unix())
	var defaultFrom float64
	if defaultWindow > 0 {
		defaultFrom = now - defaultWindow.Seconds()
	}
	f := req.GetFloat("from", defaultFrom)
	u := req.GetFloat("until", now)
	return fmt.Sprintf("%.0f", f), fmt.Sprintf("%.0f", u)
}

// getCountCapped extracts an integer count parameter with a maximum cap.
func getCountCapped(req mcp.CallToolRequest, key string, defaultVal, maxVal int) int {
	v := int(req.GetFloat(key, float64(defaultVal)))
	if v > maxVal {
		return maxVal
	}
	return v
}

// writeProcessedResult writes bulk operation results to a string builder.
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
