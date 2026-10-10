package main

import (
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"bd/internal/picker"
)

func TestPickerOffersDailyNoteFirstThenNotes(t *testing.T) {
	h := newHarness(t)
	vault := h.vault()
	writeFile(t, filepath.Join(vault, "Projects/Kafka.md"), "# Kafka\n")
	writeFile(t, filepath.Join(vault, ".obsidian/hidden.md"), "nope\n")
	h.interactive = true
	h.pickFn = pressEnter

	wantSuccess(t, h.bd("hello"))

	want := []picker.Candidate{
		{Path: "Daily/2026-10-10.md", Display: "📅 Today's daily note"},
		{Path: "Projects/Kafka.md", Display: "Projects/Kafka"},
	}
	if !slices.Equal(h.offered, want) {
		t.Errorf("picker offered %+v, want %+v", h.offered, want)
	}
}

func TestPickerSelectingANoteAppendsToIt(t *testing.T) {
	h := newHarness(t)
	vault := h.vault()
	writeFile(t, filepath.Join(vault, "Projects/Kafka.md"), "# Kafka\n")
	h.interactive = true
	h.pickFn = selectNote("Projects/Kafka.md")

	r := h.bd("retries", "need", "backoff")

	wantSuccess(t, r)
	wantStdout(t, r, "added to Projects/Kafka.md\n")
	wantFile(t, filepath.Join(vault, "Projects/Kafka.md"),
		"# Kafka\n\n## 2026-10-10 14:32\nretries need backoff\n")
}

func TestPickerTypingANewNameCreatesIt(t *testing.T) {
	h := newHarness(t)
	vault := h.vault()
	h.interactive = true
	h.pickFn = typeNewName("Ideas/Rust CLI")

	r := h.bd("build", "it", "in", "go")

	wantSuccess(t, r)
	wantStdout(t, r, "created Ideas/Rust CLI.md\n")
	wantFile(t, filepath.Join(vault, "Ideas/Rust CLI.md"), "## 2026-10-10 14:32\nbuild it in go\n")
}

func TestPickerEnterOnEmptyPromptSavesToDailyNote(t *testing.T) {
	h := newHarness(t)
	vault := h.vault()
	h.interactive = true
	h.pickFn = pressEnter

	r := h.bd("quick", "thought")

	wantSuccess(t, r)
	wantStdout(t, r, "added to Daily/2026-10-10.md\n")
	wantFile(t, filepath.Join(vault, todaysNote), "## 2026-10-10 14:32\nquick thought\n")
}

func TestPickerEscCancelsWithoutSaving(t *testing.T) {
	h := newHarness(t)
	vault := h.vault()
	h.interactive = true
	h.pickFn = pressEsc

	r := h.bd("oops", "wrong", "thing")

	wantFailure(t, r, "cancelled, nothing saved")
	if r.code != 130 {
		t.Errorf("exit code = %d, want 130", r.code)
	}
	wantStdout(t, r, "")
	wantNoFile(t, filepath.Join(vault, todaysNote))
}

func TestPickerNameOutsideVaultFallsBackToDailyNote(t *testing.T) {
	h := newHarness(t)
	vault := h.vault()
	h.interactive = true
	h.pickFn = typeNewName("../escape")

	r := h.bd("sneaky")

	wantSuccess(t, r)
	if !strings.Contains(r.stderr, "inside the vault") {
		t.Errorf("stderr = %q, want a warning about the vault", r.stderr)
	}
	wantFile(t, filepath.Join(vault, todaysNote), "## 2026-10-10 14:32\nsneaky\n")
	wantNoFile(t, filepath.Join(h.home, "escape.md"))
}

func TestPickerIsSkipped(t *testing.T) {
	// pickFn is left nil: the harness fails the test if the picker is shown.
	tests := []struct {
		name        string
		interactive bool
		env         string
		args        []string
	}{
		{name: "with --daily", interactive: true, args: []string{"--daily", "straight", "in"}},
		{name: "when stdout is not a terminal", interactive: false, args: []string{"scripted"}},
		{name: "when BD_INTERACTIVE=0", interactive: true, env: "0", args: []string{"forced", "off"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHarness(t)
			h.vault()
			h.interactive = tt.interactive
			h.env["BD_INTERACTIVE"] = tt.env

			r := h.bd(tt.args...)

			wantSuccess(t, r)
			wantStdout(t, r, "added to Daily/2026-10-10.md\n")
		})
	}
}

func TestPickerForcedOnWithBDInteractive(t *testing.T) {
	h := newHarness(t)
	h.vault()
	h.env["BD_INTERACTIVE"] = "1"
	h.pickFn = pressEnter

	wantSuccess(t, h.bd("hello"))

	if h.offered == nil {
		t.Error("picker was not shown")
	}
}
