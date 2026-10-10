package main

import (
	"errors"
	"io"
	"io/fs"
	"os"

	"bd/internal/note"
	"bd/internal/picker"
	"bd/internal/vault"
)

// target is the note a dump will be appended to.
type target struct {
	path  string
	isNew bool // created by the user typing a new name in the picker
}

// chooseTarget asks the user where the dump goes. Any problem falls back to
// the daily note with a warning: a dump is never thrown away.
func chooseTarget(v vaultDir, dailyPath string, stderr io.Writer) target {
	daily := target{path: dailyPath}
	dailyRel := v.rel(dailyPath)

	notes, err := vault.ListNotes(v.root, dailyRel)
	if err != nil {
		warn(stderr, err)
		return daily
	}
	candidates := []picker.Candidate{{Path: dailyRel, Display: "📅 Today's daily note"}}
	for _, n := range notes {
		candidates = append(candidates, picker.Candidate{Path: n, Display: picker.DisplayName(n)})
	}

	choice, err := picker.FZF(fzfBin(), v.root, candidates)
	if err != nil {
		warn(stderr, err)
		return daily
	}
	if choice.Kind == picker.Daily {
		return daily
	}
	path, err := note.Resolve(v.root, choice.Path)
	if err != nil {
		warn(stderr, err)
		return daily
	}
	_, statErr := os.Stat(path)
	return target{path: path, isNew: errors.Is(statErr, fs.ErrNotExist)}
}
