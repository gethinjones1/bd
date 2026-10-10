package vault

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"
)

func TestListNotes(t *testing.T) {
	root := t.TempDir()
	base := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	write := func(rel string, age time.Duration) {
		p := filepath.Join(root, rel)
		os.MkdirAll(filepath.Dir(p), 0o755)
		os.WriteFile(p, []byte("x"), 0o644)
		os.Chtimes(p, base.Add(-age), base.Add(-age))
	}
	write("Old.md", 3*time.Hour)
	write("Projects/Kafka.md", time.Hour)
	write("Areas/Infra.md", 2*time.Hour)
	write("Daily/2026-10-10.md", 0)
	write("image.png", 0)
	write(".obsidian/workspace.md", 0)
	write(".trash/Deleted.md", 0)

	got, err := ListNotes(root, "Daily/2026-10-10.md")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"Projects/Kafka.md", "Areas/Infra.md", "Old.md"}
	if !slices.Equal(got, want) {
		t.Errorf("got %q, want %q (most recent first, no hidden dirs, no non-md, excluded skipped)", got, want)
	}
}
