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

func TestResolveVault(t *testing.T) {
	home := t.TempDir()
	flagDir, envDir, cfgDir := mkdir(t, home, "flag"), mkdir(t, home, "env"), mkdir(t, home, "cfg")
	notADir := filepath.Join(home, "file.md")
	os.WriteFile(notADir, nil, 0o644)

	tests := []struct {
		name, flag, env, cfg, want string
		wantErr                    error
	}{
		{name: "flag wins", flag: flagDir, env: envDir, cfg: cfgDir, want: flagDir},
		{name: "env beats config", env: envDir, cfg: cfgDir, want: envDir},
		{name: "config fallback", cfg: cfgDir, want: cfgDir},
		{name: "tilde expands to home", cfg: "~/cfg", want: cfgDir},
		{name: "nothing set", wantErr: ErrNoVault},
		{name: "missing dir", flag: filepath.Join(home, "missing"), wantErr: ErrBadVault},
		{name: "file not dir", flag: notADir, wantErr: ErrBadVault},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolveVault(tt.flag, tt.env, Config{Vault: tt.cfg}, home)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func mapEnv(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}

func mkdir(t *testing.T, parts ...string) string {
	t.Helper()
	p := filepath.Join(parts...)
	if err := os.MkdirAll(p, 0o755); err != nil {
		t.Fatal(err)
	}
	return p
}
