package main

import (
	"fmt"
	"path/filepath"

	"bd/internal/config"
)

// runConfig handles `bd config set vault PATH`.
func (a *app) runConfig(args []string) int {
	if len(args) != 3 || args[0] != "set" {
		fmt.Fprintln(a.stderr, "usage: bd config set vault PATH")
		return 2
	}
	key, value := args[1], args[2]
	if key != "vault" {
		fmt.Fprintf(a.stderr, "bd: unknown config key %q\n", key)
		return 2
	}

	path := config.Path(a.getenv)
	cfg, err := config.Load(path)
	if err != nil {
		return a.fail(err)
	}
	dir := config.ExpandHome(value, a.getenv("HOME"))
	if err := a.ensureVault(dir); err != nil {
		return a.fail(err)
	}
	if cfg.Vault, err = filepath.Abs(dir); err != nil {
		return a.fail(err)
	}
	if err := config.Save(path, cfg); err != nil {
		return a.fail(err)
	}
	fmt.Fprintf(a.stdout, "vault set to %s\n", cfg.Vault)
	return 0
}
