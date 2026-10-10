package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"bd/internal/config"
)

// runConfig handles `bd config set vault PATH`.
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
	dir := config.ExpandHome(value, home)
	if err := ensureVault(dir, stdout); err != nil {
		return fail(stderr, err)
	}
	if cfg.Vault, err = filepath.Abs(dir); err != nil {
		return fail(stderr, err)
	}
	if err := config.Save(path, cfg); err != nil {
		return fail(stderr, err)
	}
	fmt.Fprintf(stdout, "vault set to %s\n", cfg.Vault)
	return 0
}
