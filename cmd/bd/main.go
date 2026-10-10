package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"bd/internal/capture"
	"bd/internal/config"
	"bd/internal/note"
)

var version = "dev"

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) > 0 && args[0] == "config" {
		return runConfig(args[1:], stdout, stderr)
	}

	fs := flag.NewFlagSet("bd", flag.ContinueOnError)
	fs.SetOutput(stderr)
	showVersion := fs.Bool("version", false, "print version and exit")
	vaultFlag := fs.String("vault", "", "vault directory (overrides BD_VAULT and config)")
	fs.Bool("daily", false, "append straight to today's daily note")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *showVersion {
		fmt.Fprintf(stdout, "bd %s\n", version)
		return 0
	}

	text, err := capture.Read(fs.Args(), stdin, isTerminal(stdin))
	if err != nil {
		return fail(stderr, err)
	}
	cfg, err := config.Load(config.Path(os.Getenv))
	if err != nil {
		return fail(stderr, err)
	}
	home, _ := os.UserHomeDir()
	vault, err := config.ResolveVault(*vaultFlag, os.Getenv("BD_VAULT"), cfg, home)
	if err != nil {
		return fail(stderr, err)
	}
	now, err := clock()
	if err != nil {
		return fail(stderr, err)
	}

	target := note.DailyPath(vault, cfg.DailyDir, now)
	if err := note.Append(target, note.FormatEntry(now, text)); err != nil {
		return fail(stderr, err)
	}
	rel, _ := filepath.Rel(vault, target)
	fmt.Fprintf(stdout, "added to %s\n", rel)
	return 0
}

// clock returns the current time, or BD_NOW (RFC 3339) when set so tests
// can pin the date.
func clock() (time.Time, error) {
	if v := os.Getenv("BD_NOW"); v != "" {
		return time.Parse(time.RFC3339, v)
	}
	return time.Now(), nil
}

func isTerminal(r io.Reader) bool {
	f, ok := r.(*os.File)
	if !ok {
		return false
	}
	info, err := f.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

func runConfig(args []string, stdout, stderr io.Writer) int {
	if len(args) != 3 || args[0] != "set" {
		fmt.Fprintln(stderr, "usage: bd config set vault PATH")
		return 2
	}
	key, value := args[1], args[2]
	if key != "vault" {
		fmt.Fprintf(stderr, "bd: unknown config key %q\n", key)
		return 2
	}

	path := config.Path(os.Getenv)
	cfg, err := config.Load(path)
	if err != nil {
		return fail(stderr, err)
	}
	home, _ := os.UserHomeDir()
	vault, err := config.ResolveVault(value, "", cfg, home)
	if err != nil {
		return fail(stderr, err)
	}
	if cfg.Vault, err = filepath.Abs(vault); err != nil {
		return fail(stderr, err)
	}
	if err := config.Save(path, cfg); err != nil {
		return fail(stderr, err)
	}
	fmt.Fprintf(stdout, "vault set to %s\n", cfg.Vault)
	return 0
}

func fail(stderr io.Writer, err error) int {
	fmt.Fprintf(stderr, "bd: %v\n", err)
	return 1
}
