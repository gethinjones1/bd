// Package vault reads the structure of an Obsidian vault.
package vault

import (
	"io/fs"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

// ListNotes returns vault-relative paths of Markdown notes, most recently
// modified first. Hidden directories (.obsidian, .trash, .git) are skipped,
// as are any paths in exclude.
func ListNotes(root string, exclude ...string) ([]string, error) {
	type entry struct {
		rel string
		mod time.Time
	}
	var notes []entry
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if path != root && strings.HasPrefix(d.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".md" {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if slices.Contains(exclude, rel) {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		notes = append(notes, entry{rel, info.ModTime()})
		return nil
	})
	if err != nil {
		return nil, err
	}
	slices.SortStableFunc(notes, func(a, b entry) int { return b.mod.Compare(a.mod) })
	out := make([]string, len(notes))
	for i, n := range notes {
		out[i] = n.rel
	}
	return out, nil
}
