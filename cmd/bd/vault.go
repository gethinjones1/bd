package main

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/gethinjones1/bd/internal/config"
	"github.com/gethinjones1/bd/internal/note"
)

// vaultDir is an opened vault: its root directory plus the loaded config.
type vaultDir struct {
	root string
	cfg  config.Config
}

// openVault loads the config, picks the vault (--vault, BD_VAULT, config),
// and creates its directory if missing.
func (a *app) openVault(flagValue string) (vaultDir, error) {
	cfg, err := config.Load(config.Path(a.getenv))
	if err != nil {
		return vaultDir{}, err
	}
	root, err := config.VaultPath(flagValue, a.getenv("BD_VAULT"), cfg, a.getenv("HOME"))
	if err != nil {
		return vaultDir{}, err
	}
	if err := a.ensureVault(root); err != nil {
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
func (a *app) ensureVault(dir string) error {
	created, err := config.EnsureDir(dir)
	if created {
		fmt.Fprintf(a.stdout, "created vault %s\n", dir)
	}
	return err
}
