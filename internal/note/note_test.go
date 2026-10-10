package note

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

var at = time.Date(2026, 10, 10, 14, 32, 5, 0, time.Local)

func TestFormatEntry(t *testing.T) {
	got := FormatEntry(at, "kafka retries\nneed backoff")
	want := "## 2026-10-10 14:32\nkafka retries\nneed backoff\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestDailyPath(t *testing.T) {
	got := DailyPath("/vault", "Daily", at)
	if got != "/vault/Daily/2026-10-10.md" {
		t.Errorf("got %q", got)
	}
}

func TestAppend(t *testing.T) {
	entry := "## 2026-10-10 14:32\nnew\n"
	tests := []struct {
		name     string
		existing *string // nil = file does not exist
		want     string
	}{
		{name: "creates missing file and dirs", existing: nil, want: entry},
		{name: "empty file", existing: ptr(""), want: entry},
		{name: "separates with blank line", existing: ptr("# Note\nbody\n"), want: "# Note\nbody\n\n" + entry},
		{name: "adds missing trailing newline", existing: ptr("# Note\nbody"), want: "# Note\nbody\n\n" + entry},
		{name: "does not stack blank lines", existing: ptr("# Note\n\n\n"), want: "# Note\n\n" + entry},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "Daily", "2026-10-10.md")
			if tt.existing != nil {
				os.MkdirAll(filepath.Dir(path), 0o755)
				os.WriteFile(path, []byte(*tt.existing), 0o644)
			}
			if err := Append(path, entry); err != nil {
				t.Fatal(err)
			}
			got, _ := os.ReadFile(path)
			if string(got) != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func ptr(s string) *string { return &s }
