// Package note formats dump entries and appends them to Markdown notes.
package note

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// FormatEntry renders a dump as a timestamped Markdown section.
func FormatEntry(t time.Time, text string) string {
	return "## " + t.Format("2006-01-02 15:04") + "\n" + strings.TrimRight(text, "\n") + "\n"
}

// DailyPath is the Obsidian-style daily note for t, e.g. <vault>/Daily/2026-10-10.md.
func DailyPath(vault, dailyDir string, t time.Time) string {
	return filepath.Join(vault, dailyDir, t.Format("2006-01-02")+".md")
}

// Append adds entry to the note at path, creating it (and parent dirs) if
// needed. Existing content is separated from the entry by one blank line.
func Append(path, entry string) error {
	existing, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	body := strings.TrimRight(string(existing), "\n")
	if body != "" {
		body += "\n\n"
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(body+entry), 0o644)
}

var ErrOutsideVault = errors.New("note path must be inside the vault")

// Resolve turns a vault-relative note name (as typed by the user) into an
// absolute path, adding .md if missing and refusing anything outside vault.
func Resolve(vault, rel string) (string, error) {
	rel = strings.TrimSpace(rel)
	if !filepath.IsLocal(rel) {
		return "", fmt.Errorf("%w: %q", ErrOutsideVault, rel)
	}
	rel = filepath.Clean(rel)
	if filepath.Ext(rel) != ".md" {
		rel += ".md"
	}
	return filepath.Join(vault, rel), nil
}
