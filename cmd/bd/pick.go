package main

import (
	"errors"
	"io/fs"
	"os"

	"bd/internal/note"
	"bd/internal/picker"
	"bd/internal/vault"
)

// target is the note a dump will be appended to.
type target struct {
	path  string
	isNew bool // the user typed a new name in the picker
}

// chooseTarget asks the user where the dump goes. ok is false if they
// cancelled. Any other problem falls back to the daily note with a warning.
func (a *app) chooseTarget(v vaultDir, dailyPath string) (t target, ok bool) {
	daily := target{path: dailyPath}
	dailyRel := v.rel(dailyPath)

	notes, err := vault.ListNotes(v.root, dailyRel)
	if err != nil {
		a.warn(err)
		return daily, true
	}
	candidates := []picker.Candidate{{Path: dailyRel, Display: "📅 Today's daily note"}}
	for _, n := range notes {
		candidates = append(candidates, picker.Candidate{Path: n, Display: picker.DisplayName(n)})
	}

	choice, err := a.pick(v.root, candidates)
	if err != nil {
		a.warn(err)
		return daily, true
	}
	switch choice.Kind {
	case picker.Cancelled:
		return target{}, false
	case picker.Daily:
		return daily, true
	}
	path, err := note.Resolve(v.root, choice.Path)
	if err != nil {
		a.warn(err)
		return daily, true
	}
	_, statErr := os.Stat(path)
	isNew := choice.Kind == picker.New && errors.Is(statErr, fs.ErrNotExist)
	return target{path: path, isNew: isNew}, true
}
