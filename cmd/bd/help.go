package main

import (
	"flag"
	"fmt"
	"strings"
)

const usageTop = `bd - brain dump into your Obsidian vault

Usage:
  bd [flags] TEXT...          save TEXT (or piped stdin) to a note you pick
  bd ls                       list the notes in your vault
  bd config set vault PATH    set your vault directory (created if missing)
  bd help                     show this help

Picker:
  enter                save to the highlighted note (today's daily note)
  type, then enter     save to the best match, or create a new note if nothing matches
  esc                  cancel, nothing is saved

Flags:
`

const usageBottom = `
Environment:
  BD_VAULT          vault directory (--vault wins over it, it wins over config)
  BD_INTERACTIVE    1 or 0 to force the picker on or off
  BD_FZF            picker binary to run instead of fzf

Examples:
  bd kafka retries need exponential backoff
  pbpaste | bd
  bd --daily standup: finished the ls command
`

// printUsage writes the full help to stdout. The flag list comes from the
// real flag definitions so it never drifts.
func (a *app) printUsage() {
	fmt.Fprint(a.stdout, usageTop)
	newDumpFlags(&dumpOptions{}).VisitAll(func(f *flag.Flag) {
		arg, usage := flag.UnquoteUsage(f)
		fmt.Fprintf(a.stdout, "  %-19s %s\n", strings.TrimSpace("--"+f.Name+" "+arg), usage)
	})
	fmt.Fprint(a.stdout, usageBottom)
}
