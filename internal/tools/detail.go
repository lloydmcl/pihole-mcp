package tools

import "github.com/mark3labs/mcp-go/mcp"

// detailParam is the standard detail parameter added to verbose tools.
var detailParam = mcp.WithString("detail",
	mcp.Description("Response detail: minimal, normal (default), or full."),
	mcp.Enum("minimal", "normal", "full"),
)

// formatParam is the standard output format parameter for tabular tools.
var formatParam = mcp.WithString("format",
	mcp.Description("Output format: text (default) or csv."),
	mcp.Enum("text", "csv"),
)

// getDetail extracts the detail level from a tool request.
func getDetail(req mcp.CallToolRequest) string {
	d := req.GetString("detail", "normal")
	if d != "minimal" && d != "full" {
		return "normal"
	}
	return d
}

// wantCSV returns true if the user requested CSV output format.
func wantCSV(req mcp.CallToolRequest) bool {
	return req.GetString("format", "text") == "csv"
}
