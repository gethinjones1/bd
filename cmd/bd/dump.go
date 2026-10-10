package main

import (
	"flag"
	"fmt"
	"io"

	"bd/internal/capture"
	"bd/internal/note"
)

type dumpOptions struct {
	vault   string
	daily   bool
	version bool
	words   []string
}

// runDump is the default command: capture text, decide which note it goes
// in, append it, and say where it went.
func runDump(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	opts, err := parseDumpFlags(args, stderr)
	if err != nil {
		return 2
	}
	if opts.version {
		fmt.Fprintf(stdout, "bd %s\n", version)
		return 0
	}

	text, err := capture.Read(opts.words, stdin, isTerminal(stdin))
	if err != nil {
		return fail(stderr, err)
	}
	v, err := openVault(opts.vault, stdout)
	if err != nil {
		return fail(stderr, err)
	}
	now, err := clock()
	if err != nil {
		return fail(stderr, err)
	}

	t := target{path: v.dailyNote(now)}
	if !opts.daily && interactive() {
		t = chooseTarget(v, t.path, stderr)
	}

	if err := note.Append(t.path, note.FormatEntry(now, text)); err != nil {
		return fail(stderr, err)
	}
	if t.isNew {
		fmt.Fprintf(stdout, "created %s\n", v.rel(t.path))
	} else {
		fmt.Fprintf(stdout, "added to %s\n", v.rel(t.path))
	}
	return 0
}

func parseDumpFlags(args []string, stderr io.Writer) (dumpOptions, error) {
	var opts dumpOptions
	fs := flag.NewFlagSet("bd", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.BoolVar(&opts.version, "version", false, "print version and exit")
	fs.StringVar(&opts.vault, "vault", "", "vault directory (overrides BD_VAULT and config)")
	fs.BoolVar(&opts.daily, "daily", false, "append straight to today's daily note, skipping the picker")
	err := fs.Parse(args)
	opts.words = fs.Args()
	return opts, err
}
