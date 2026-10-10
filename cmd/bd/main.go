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
)

var version = "dev"

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) > 0 {
		switch args[0] {
		case "config":
			return runConfig(args[1:], stdout, stderr)
		}
	}
	return runDump(args, stdin, stdout, stderr)
}

func warn(stderr io.Writer, err error) {
	fmt.Fprintf(stderr, "bd: %v (saving to daily note instead)\n", err)
}

func fail(stderr io.Writer, err error) int {
	fmt.Fprintf(stderr, "bd: %v\n", err)
	return 1
}
