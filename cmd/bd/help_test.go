package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestHelpPrintsFullUsage(t *testing.T) {
	for _, arg := range []string{"--help", "-h", "help"} {
		t.Run(arg, func(t *testing.T) {
			h := newHarness(t)
			vault := h.vault()

			r := h.bd(arg)

			wantSuccess(t, r)
			for _, want := range []string{
				"Usage:",
				"bd [flags] TEXT...", // dumping
				"bd ls",              // subcommands
				"bd config set vault PATH",
				"esc",    // picker keys
				"-daily", // flags
				"-vault",
				"BD_VAULT",     // environment
				"pbpaste | bd", // examples
			} {
				if !strings.Contains(r.stdout, want) {
					t.Errorf("help is missing %q:\n%s", want, r.stdout)
				}
			}
			if r.stderr != "" {
				t.Errorf("stderr = %q, want help on stdout only", r.stderr)
			}
			wantNoFile(t, filepath.Join(vault, todaysNote)) // "help" is not saved as a dump
		})
	}
}

func TestUnknownFlagPointsToHelp(t *testing.T) {
	h := newHarness(t)

	r := h.bd("--nope")

	if r.code != 2 {
		t.Errorf("exit code = %d, want 2", r.code)
	}
	wantStdout(t, r, "")
	if want := "bd: flag provided but not defined: -nope\nrun 'bd --help' for usage\n"; r.stderr != want {
		t.Errorf("stderr = %q, want %q", r.stderr, want)
	}
}
