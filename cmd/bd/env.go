package main

import (
	"os"

	"github.com/gethinjones1/bd/internal/picker"
)

// interactive reports whether to show the picker: BD_INTERACTIVE=1/0 forces
// it on/off, otherwise only when stdout is a terminal.
func (a *app) interactive() bool {
	switch a.getenv("BD_INTERACTIVE") {
	case "1":
		return true
	case "0":
		return false
	}
	return a.stdoutIsTTY
}

func isTerminal(f *os.File) bool {
	info, err := f.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

// fzfPicker runs fzf, or the binary named by BD_FZF (e.g. sk).
func fzfPicker(getenv func(string) string) func(string, []picker.Candidate) (picker.Choice, error) {
	bin := getenv("BD_FZF")
	if bin == "" {
		bin = "fzf"
	}
	return func(vaultRoot string, candidates []picker.Candidate) (picker.Choice, error) {
		return picker.FZF(bin, vaultRoot, candidates)
	}
}
