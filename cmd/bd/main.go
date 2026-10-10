package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"bd/internal/capture"
	"bd/internal/config"
	"bd/internal/note"
	"bd/internal/picker"
	"bd/internal/vault"
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
	daily := fs.Bool("daily", false, "append straight to today's daily note, skipping the picker")
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
	vaultDir, err := config.VaultPath(*vaultFlag, os.Getenv("BD_VAULT"), cfg, home)
	if err != nil {
		return fail(stderr, err)
	}
	if err := ensureVault(vaultDir, stdout); err != nil {
		return fail(stderr, err)
	}
	now, err := clock()
	if err != nil {
		return fail(stderr, err)
	}

	dailyPath := note.DailyPath(vaultDir, cfg.DailyDir, now)
	target, isNew := dailyPath, false
	if !*daily && interactive() {
		target, isNew = pickTarget(vaultDir, dailyPath, stderr)
	}

	if err := note.Append(target, note.FormatEntry(now, text)); err != nil {
		return fail(stderr, err)
	}
	rel, _ := filepath.Rel(vaultDir, target)
	if isNew {
		fmt.Fprintf(stdout, "created %s\n", rel)
	} else {
		fmt.Fprintf(stdout, "added to %s\n", rel)
	}
	return 0
}

// pickTarget asks the user where the dump goes. Any problem falls back to
// the daily note with a warning: a dump is never thrown away.
func pickTarget(vaultDir, dailyPath string, stderr io.Writer) (path string, isNew bool) {
	dailyRel, _ := filepath.Rel(vaultDir, dailyPath)
	dailyRel = filepath.ToSlash(dailyRel)
	notes, err := vault.ListNotes(vaultDir, dailyRel)
	if err != nil {
		warn(stderr, err)
		return dailyPath, false
	}
	candidates := []picker.Candidate{{Path: dailyRel, Display: "📅 Today's daily note"}}
	for _, n := range notes {
		candidates = append(candidates, picker.Candidate{Path: n, Display: picker.DisplayName(n)})
	}

	choice, err := picker.FZF(fzfBin(), vaultDir, candidates)
	if err != nil {
		warn(stderr, err)
		return dailyPath, false
	}
	if choice.Kind == picker.Daily {
		return dailyPath, false
	}
	path, err = note.Resolve(vaultDir, choice.Path)
	if err != nil {
		warn(stderr, err)
		return dailyPath, false
	}
	_, statErr := os.Stat(path)
	return path, errors.Is(statErr, fs.ErrNotExist)
}

func fzfBin() string {
	if bin := os.Getenv("BD_FZF"); bin != "" {
		return bin
	}
	return "fzf"
}

// interactive reports whether to show the picker: BD_INTERACTIVE=1/0 forces
// it on/off, otherwise only when stdout is a terminal.
func interactive() bool {
	switch os.Getenv("BD_INTERACTIVE") {
	case "1":
		return true
	case "0":
		return false
	}
	return isTerminal(os.Stdout)
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
	vault, err := config.VaultPath(value, "", cfg, home)
	if err != nil {
		return fail(stderr, err)
	}
	if err := ensureVault(vault, stdout); err != nil {
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

// ensureVault creates the vault directory if it is missing, saying so.
func ensureVault(dir string, stdout io.Writer) error {
	created, err := config.EnsureDir(dir)
	if created {
		fmt.Fprintf(stdout, "created vault %s\n", dir)
	}
	return err
}

// clock returns the current time, or BD_NOW (RFC 3339) when set so tests
// can pin the date.
func clock() (time.Time, error) {
	if v := os.Getenv("BD_NOW"); v != "" {
		return time.Parse(time.RFC3339, v)
	}
	return time.Now(), nil
}

func isTerminal(r any) bool {
	f, ok := r.(*os.File)
	if !ok {
		return false
	}
	info, err := f.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

func warn(stderr io.Writer, err error) {
	fmt.Fprintf(stderr, "bd: %v (saving to daily note instead)\n", err)
}

func fail(stderr io.Writer, err error) int {
	fmt.Fprintf(stderr, "bd: %v\n", err)
	return 1
}
