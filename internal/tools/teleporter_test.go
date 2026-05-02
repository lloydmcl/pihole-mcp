package tools

import (
	"os"
	"strings"
	"testing"
)

func TestTeleporterExport_Success(t *testing.T) {
	c := newTestClient(t, piholeRawHandler(
		nil,
		map[string]string{
			"/teleporter": "PK\x03\x04fakezipdata",
		},
	))

	text := callTool(t, teleporterExportHandler, c, nil)
	if !strings.Contains(text, "Backup saved") {
		t.Errorf("expected 'Backup saved' message, got: %s", text)
	}
	if !strings.Contains(text, "bytes") {
		t.Errorf("expected file size in output, got: %s", text)
	}

	// Clean up the temp file created by the handler.
	// Extract the file path from the output (between "File: " and " (").
	if idx := strings.Index(text, "File: "); idx >= 0 {
		rest := text[idx+6:]
		if end := strings.Index(rest, " ("); end >= 0 {
			_ = os.Remove(rest[:end])
		}
	}
}

func TestTeleporterImport_MissingFile(t *testing.T) {
	c := newTestClient(t, piholeHandler(map[string]any{}))

	text := callToolExpectError(t, teleporterImportHandler, c, map[string]any{
		"file_path": "/nonexistent/path/to/backup.zip",
	})
	if !strings.Contains(text, "Failed to import") {
		t.Errorf("expected import failure message, got: %s", text)
	}
}
