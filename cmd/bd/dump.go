package main

import (
	"errors"
	"flag"
	"fmt"
	"io"

	"github.com/gethinjones1/bd/internal/capture"
	"github.com/gethinjones1/bd/internal/note"
)

type dumpOptions struct {
	vault   string
	daily   bool
	version bool
	words   []string
}

// runDump is the default command: capture text, decide which note it goes
// in, append it, and say where it went.
func (a *app) runDump(args []string) int {
	opts, err := parseDumpFlags(args)
	if errors.Is(err, flag.ErrHelp) {
		a.printUsage()
		return 0
	}
	if err != nil {
		fmt.Fprintf(a.stderr, "bd: %v\nrun 'bd --help' for usage\n", err)
		return 2
	}
	if opts.version {
		fmt.Fprintf(a.stdout, "bd %s\n", version)
		return 0
	}

	text, err := capture.Read(opts.words, a.stdin, a.stdinIsTTY)
	if err != nil {
		return a.fail(err)
	}
	v, err := a.openVault(opts.vault)
	if err != nil {
		return a.fail(err)
	}
	now := a.now()

	t := target{path: v.dailyNote(now)}
	if !opts.daily && a.interactive() {
		var ok bool
		if t, ok = a.chooseTarget(v, t.path); !ok {
			fmt.Fprintln(a.stderr, "bd: cancelled, nothing saved")
			return 130
		}
	}

	if err := note.Append(t.path, note.FormatEntry(now, text)); err != nil {
		return a.fail(err)
	}
	if t.isNew {
		fmt.Fprintf(a.stdout, "created %s\n", v.rel(t.path))
	} else {
		fmt.Fprintf(a.stdout, "added to %s\n", v.rel(t.path))
	}
	return 0
}

func parseDumpFlags(args []string) (dumpOptions, error) {
	var opts dumpOptions
	fs := newDumpFlags(&opts)
	err := fs.Parse(args)
	opts.words = fs.Args()
	return opts, err
}

// newDumpFlags defines bd's flags. Errors are reported by the caller, so
// the flag package is told to print nothing itself.
func newDumpFlags(opts *dumpOptions) *flag.FlagSet {
	fs := flag.NewFlagSet("bd", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.BoolVar(&opts.version, "version", false, "print version and exit")
	fs.StringVar(&opts.vault, "vault", "", "use the vault at `PATH` (overrides BD_VAULT and config)")
	fs.BoolVar(&opts.daily, "daily", false, "append straight to today's daily note, skipping the picker")
	return fs
}
