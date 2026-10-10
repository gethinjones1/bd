// Command bd captures brain dumps into an Obsidian vault.
//
// Each subcommand lives in its own file:
//
//	bd [flags] TEXT...   dump.go    capture a dump into a note
//	bd config set ...    config.go  change settings
package main

import (
	"fmt"
	"io"
	"os"
	"time"

	"bd/internal/picker"
)

var version = "dev"

// app is everything bd takes from the outside world. main fills it with the
// real thing; tests fill it with fakes.
type app struct {
	stdin       io.Reader
	stdout      io.Writer
	stderr      io.Writer
	getenv      func(string) string
	stdinIsTTY  bool
	stdoutIsTTY bool
	now         func() time.Time
	pick        func(vaultRoot string, candidates []picker.Candidate) (picker.Choice, error)
}

func main() {
	a := &app{
		stdin:       os.Stdin,
		stdout:      os.Stdout,
		stderr:      os.Stderr,
		getenv:      os.Getenv,
		stdinIsTTY:  isTerminal(os.Stdin),
		stdoutIsTTY: isTerminal(os.Stdout),
		now:         time.Now,
		pick:        fzfPicker(os.Getenv),
	}
	os.Exit(a.run(os.Args[1:]))
}

func (a *app) run(args []string) int {
	if len(args) > 0 {
		switch args[0] {
		case "config":
			return a.runConfig(args[1:])
		}
	}
	return a.runDump(args)
}

func (a *app) warn(err error) {
	fmt.Fprintf(a.stderr, "bd: %v (saving to daily note instead)\n", err)
}

func (a *app) fail(err error) int {
	fmt.Fprintf(a.stderr, "bd: %v\n", err)
	return 1
}
