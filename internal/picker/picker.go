// Package picker asks the user which note a dump belongs in, via fzf.
package picker

import (
	"fmt"
	"strings"
)

type Kind int

const (
	Daily    Kind = iota // fall back to today's daily note
	Existing             // Path is a vault-relative note that was selected
	New                  // Path is what the user typed, to be created
)

type Choice struct {
	Kind Kind
	Path string
}

// Parse interprets the output of `fzf --print-query` over lines of the form
// "relpath\tdisplay". Esc never discards the dump: it falls back to Daily.
func Parse(out string, exit int) (Choice, error) {
	switch exit {
	case 0, 1: // 0: selected; 1: no match
	case 130:
		return Choice{Kind: Daily}, nil
	default:
		return Choice{}, fmt.Errorf("fzf exited with status %d", exit)
	}
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) > 1 && lines[1] != "" {
		path, _, _ := strings.Cut(lines[1], "\t")
		return Choice{Kind: Existing, Path: path}, nil
	}
	if query := strings.TrimSpace(lines[0]); query != "" {
		return Choice{Kind: New, Path: query}, nil
	}
	return Choice{Kind: Daily}, nil
}
