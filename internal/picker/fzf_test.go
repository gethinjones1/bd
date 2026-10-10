package picker

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fakeFZF writes a shell script that records what bd sent it (stdin, args,
// working dir) and then replies like fzf would.
func fakeFZF(t *testing.T, reply string, exit int) (bin, recordDir string) {
	t.Helper()
	recordDir = t.TempDir()
	bin = filepath.Join(recordDir, "fzf")
	script := "#!/bin/sh\n" +
		"cat > '" + recordDir + "/stdin'\n" +
		"printf '%s\\n' \"$@\" > '" + recordDir + "/args'\n" +
		"pwd > '" + recordDir + "/pwd'\n" +
		"printf '" + reply + "'\n" +
		"exit " + string(rune('0'+exit)) + "\n"
	if err := os.WriteFile(bin, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	return bin, recordDir
}

func recorded(t *testing.T, dir, name string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(dir, name))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestFZFSendsCandidatesAndParsesSelection(t *testing.T) {
	bin, rec := fakeFZF(t, `kaf\nProjects/Kafka.md\tProjects/Kafka\n`, 0)
	vault := t.TempDir()

	got, err := FZF(bin, vault, []Candidate{
		{Path: "Daily/2026-10-10.md", Display: "📅 Today's daily note"},
		{Path: "Projects/Kafka.md", Display: "Projects/Kafka"},
	})
	if err != nil {
		t.Fatal(err)
	}

	if want := (Choice{Kind: Existing, Path: "Projects/Kafka.md"}); got != want {
		t.Errorf("choice = %+v, want %+v", got, want)
	}
	if got, want := recorded(t, rec, "stdin"),
		"Daily/2026-10-10.md\t📅 Today's daily note\nProjects/Kafka.md\tProjects/Kafka\n"; got != want {
		t.Errorf("fzf stdin = %q, want %q", got, want)
	}
	if args := recorded(t, rec, "args"); !strings.Contains(args, "--print-query\n") {
		t.Errorf("fzf args missing --print-query:\n%s", args)
	}
	if pwd, _ := filepath.EvalSymlinks(strings.TrimSpace(recorded(t, rec, "pwd"))); !strings.HasSuffix(pwd, filepath.Base(vault)) {
		t.Errorf("fzf ran in %q, want the vault %q (so previews can read notes)", pwd, vault)
	}
}

func TestFZFNoMatchExitCodeMeansNewNote(t *testing.T) {
	bin, _ := fakeFZF(t, `Ideas/Rust CLI\n`, 1)

	got, err := FZF(bin, t.TempDir(), nil)
	if err != nil {
		t.Fatal(err)
	}

	if want := (Choice{Kind: New, Path: "Ideas/Rust CLI"}); got != want {
		t.Errorf("choice = %+v, want %+v", got, want)
	}
}

func TestFZFMissingBinaryIsAnError(t *testing.T) {
	if _, err := FZF(filepath.Join(t.TempDir(), "no-such-fzf"), t.TempDir(), nil); err == nil {
		t.Error("want error when fzf is not installed")
	}
}
