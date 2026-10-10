package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestPath(t *testing.T) {
	env := map[string]string{"HOME": "/home/g"}
	if got := Path(mapEnv(env)); got != "/home/g/.config/bd/config.json" {
		t.Errorf("got %q", got)
	}
	env["XDG_CONFIG_HOME"] = "/xdg"
	if got := Path(mapEnv(env)); got != "/xdg/bd/config.json" {
		t.Errorf("with XDG got %q", got)
	}
}

func TestLoadMissingFileGivesDefaults(t *testing.T) {
	cfg, err := Load(filepath.Join(t.TempDir(), "nope.json"))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Vault != "" || cfg.DailyDir != "Daily" {
		t.Errorf("got %+v", cfg)
	}
}

func TestSaveThenLoadRoundTrips(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "config.json")
	if err := Save(path, Config{Vault: "/v", DailyDir: "Journal"}); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Vault != "/v" || cfg.DailyDir != "Journal" {
		t.Errorf("got %+v", cfg)
	}
}

func TestVaultPath(t *testing.T) {
	tests := []struct {
		name, flag, env, cfg, want string
		wantErr                    error
	}{
		{name: "flag wins", flag: "/f", env: "/e", cfg: "/c", want: "/f"},
		{name: "env beats config", env: "/e", cfg: "/c", want: "/e"},
		{name: "config fallback", cfg: "/c", want: "/c"},
		{name: "tilde expands to home", cfg: "~/notes", want: "/home/g/notes"},
		{name: "nothing set", wantErr: ErrNoVault},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := VaultPath(tt.flag, tt.env, Config{Vault: tt.cfg}, "/home/g")
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestEnsureDir(t *testing.T) {
	root := t.TempDir()

	t.Run("creates missing dir and parents", func(t *testing.T) {
		dir := filepath.Join(root, "a", "b")
		created, err := EnsureDir(dir)
		if err != nil || !created {
			t.Fatalf("created=%v err=%v", created, err)
		}
		if info, err := os.Stat(dir); err != nil || !info.IsDir() {
			t.Errorf("dir not created: %v", err)
		}
	})

	t.Run("existing dir is left alone", func(t *testing.T) {
		created, err := EnsureDir(root)
		if err != nil || created {
			t.Errorf("created=%v err=%v", created, err)
		}
	})

	t.Run("file is refused", func(t *testing.T) {
		file := filepath.Join(root, "file.md")
		os.WriteFile(file, nil, 0o644)
		if _, err := EnsureDir(file); !errors.Is(err, ErrBadVault) {
			t.Errorf("err = %v, want ErrBadVault", err)
		}
	})
}

func mapEnv(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}
