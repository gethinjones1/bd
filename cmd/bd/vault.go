package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"bd/internal/config"
	"bd/internal/note"
)

// vaultDir is an opened vault: its root directory plus the loaded config.
type vaultDir struct {
	root string
	cfg  config.Config
}

// openVault loads the config, picks the vault (--vault, BD_VAULT, config),
// and creates its directory if missing.
func openVault(flagValue string, stdout io.Writer) (vaultDir, error) {
	cfg, err := config.Load(config.Path(os.Getenv))
	if err != nil {
		return vaultDir{}, err
	}
	home, _ := os.UserHomeDir()
	root, err := config.VaultPath(flagValue, os.Getenv("BD_VAULT"), cfg, home)
	if err != nil {
		return vaultDir{}, err
	}
	if err := ensureVault(root, stdout); err != nil {
		return vaultDir{}, err
	}
	return vaultDir{root: root, cfg: cfg}, nil
}

func (v vaultDir) dailyNote(t time.Time) string {
	return note.DailyPath(v.root, v.cfg.DailyDir, t)
}

// rel returns path relative to the vault root, with forward slashes.
func (v vaultDir) rel(path string) string {
	rel, err := filepath.Rel(v.root, path)
	if err != nil {
		return path
	}
	return filepath.ToSlash(rel)
}

// ensureVault creates the vault directory if it is missing, saying so.
func ensureVault(dir string, stdout io.Writer) error {
	created, err := config.EnsureDir(dir)
	if created {
		fmt.Fprintf(stdout, "created vault %s\n", dir)
	}
	return err
}
