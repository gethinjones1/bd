// Package config stores bd settings and resolves which vault to write to.
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrNoVault  = errors.New("no vault configured (use --vault, BD_VAULT, or `bd config set vault PATH`)")
	ErrBadVault = errors.New("vault is not a directory")
)

const defaultDailyDir = "Daily"

type Config struct {
	Vault    string `json:"vault,omitempty"`
	DailyDir string `json:"daily_dir,omitempty"`
}

// Path returns the config file location, honouring XDG_CONFIG_HOME.
func Path(getenv func(string) string) string {
	base := getenv("XDG_CONFIG_HOME")
	if base == "" {
		base = filepath.Join(getenv("HOME"), ".config")
	}
	return filepath.Join(base, "bd", "config.json")
}

// Load reads the config at path. A missing file is not an error.
func Load(path string) (Config, error) {
	cfg := Config{}
	b, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return cfg, err
	}
	if err == nil {
		if err := json.Unmarshal(b, &cfg); err != nil {
			return cfg, fmt.Errorf("parsing %s: %w", path, err)
		}
	}
	if cfg.DailyDir == "" {
		cfg.DailyDir = defaultDailyDir
	}
	return cfg, nil
}

func Save(path string, cfg Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(b, '\n'), 0o644)
}

// ResolveVault picks the vault from flag, then env, then config, and checks
// that it is an existing directory.
func ResolveVault(flag, env string, cfg Config, home string) (string, error) {
	vault := firstNonEmpty(flag, env, cfg.Vault)
	if vault == "" {
		return "", ErrNoVault
	}
	if vault == "~" || strings.HasPrefix(vault, "~/") {
		vault = filepath.Join(home, vault[1:])
	}
	info, err := os.Stat(vault)
	if err != nil || !info.IsDir() {
		return "", fmt.Errorf("%w: %s", ErrBadVault, vault)
	}
	return vault, nil
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
