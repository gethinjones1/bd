package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"bd/internal/picker"
)

// These tests run bd's real run() in-process against a fake outside world:
// a temp HOME, a fixed clock, and a scripted picker instead of fzf.

var testNow = time.Date(2026, 10, 10, 14, 32, 0, 0, time.UTC)

const todaysNote = "Daily/2026-10-10.md"

type harness struct {
	t           *testing.T
	home        string
	env         map[string]string
	interactive bool                 // pretend stdout is a terminal
	pickFn      func() picker.Choice // what the user does in the picker; nil = must not be shown
	offered     []picker.Candidate   // what the picker was shown
}

type result struct {
	stdout, stderr string
	code           int
}

func newHarness(t *testing.T) *harness {
	home := t.TempDir()
	return &harness{t: t, home: home, env: map[string]string{"HOME": home}}
}

// vault creates an empty vault in the temp home and points BD_VAULT at it.
func (h *harness) vault() string {
	dir := filepath.Join(h.home, "vault")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		h.t.Fatal(err)
	}
	h.env["BD_VAULT"] = dir
	return dir
}

// bd runs `bd args...` as if typed in a terminal (stdin is not read).
func (h *harness) bd(args ...string) result {
	return h.exec(nil, args)
}

// pipe runs `echo input | bd args...`.
func (h *harness) pipe(input string, args ...string) result {
	return h.exec(&input, args)
}

func (h *harness) exec(stdin *string, args []string) result {
	var stdout, stderr bytes.Buffer
	in := ""
	if stdin != nil {
		in = *stdin
	}
	a := &app{
		stdin:       strings.NewReader(in),
		stdout:      &stdout,
		stderr:      &stderr,
		getenv:      func(k string) string { return h.env[k] },
		stdinIsTTY:  stdin == nil,
		stdoutIsTTY: h.interactive,
		now:         func() time.Time { return testNow },
		pick: func(_ string, candidates []picker.Candidate) (picker.Choice, error) {
			h.offered = candidates
			if h.pickFn == nil {
				h.t.Fatal("picker was shown but this test expects it to be skipped")
			}
			return h.pickFn(), nil
		},
	}
	code := a.run(args)
	return result{stdout: stdout.String(), stderr: stderr.String(), code: code}
}

// What the user does in the picker.
func selectNote(rel string) func() picker.Choice {
	return func() picker.Choice { return picker.Choice{Kind: picker.Existing, Path: rel} }
}
func typeNewName(name string) func() picker.Choice {
	return func() picker.Choice { return picker.Choice{Kind: picker.New, Path: name} }
}
func pressEsc() picker.Choice { return picker.Choice{Kind: picker.Daily} }

// Assertions.

func wantSuccess(t *testing.T, r result) {
	t.Helper()
	if r.code != 0 {
		t.Fatalf("exit code %d, stderr: %s", r.code, r.stderr)
	}
}

func wantFailure(t *testing.T, r result, stderrContains string) {
	t.Helper()
	if r.code == 0 {
		t.Fatalf("expected failure, got success; stdout: %s", r.stdout)
	}
	if !strings.Contains(r.stderr, stderrContains) {
		t.Errorf("stderr = %q, want it to contain %q", r.stderr, stderrContains)
	}
}

func wantStdout(t *testing.T, r result, want string) {
	t.Helper()
	if r.stdout != want {
		t.Errorf("stdout = %q, want %q", r.stdout, want)
	}
}

func wantFile(t *testing.T, path, want string) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}
	if string(got) != want {
		t.Errorf("%s =\n%q\nwant\n%q", path, got, want)
	}
}

func wantNoFile(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err == nil {
		t.Errorf("%s exists, want it not to", path)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
