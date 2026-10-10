package main

import (
	"os"
	"path/filepath"
	"testing"

	"bd/internal/config"
)

func TestConfigSetVaultSavesIt(t *testing.T) {
	h := newHarness(t)
	dir := filepath.Join(h.home, "vault")
	os.Mkdir(dir, 0o755)

	r := h.bd("config", "set", "vault", dir)

	wantSuccess(t, r)
	wantStdout(t, r, "vault set to "+dir+"\n")
	if got := savedVault(t, h); got != dir {
		t.Errorf("saved vault = %q, want %q", got, dir)
	}
}

func TestConfigSetVaultCreatesMissingDir(t *testing.T) {
	h := newHarness(t)
	dir := filepath.Join(h.home, "new", "nested", "vault")

	r := h.bd("config", "set", "vault", dir)

	wantSuccess(t, r)
	wantStdout(t, r, "created vault "+dir+"\nvault set to "+dir+"\n")
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		t.Errorf("vault dir not created: %v", err)
	}
}

func TestConfigSetVaultExpandsTilde(t *testing.T) {
	h := newHarness(t)

	wantSuccess(t, h.bd("config", "set", "vault", "~/notes"))

	want := filepath.Join(h.home, "notes")
	if got := savedVault(t, h); got != want {
		t.Errorf("saved vault = %q, want %q", got, want)
	}
}

func TestConfigSetVaultRejectsFile(t *testing.T) {
	h := newHarness(t)
	file := filepath.Join(h.home, "afile.md")
	writeFile(t, file, "x")

	wantFailure(t, h.bd("config", "set", "vault", file), "not a directory")
}

func TestConfigRejectsUnknownKey(t *testing.T) {
	h := newHarness(t)

	wantFailure(t, h.bd("config", "set", "colour", "blue"), "unknown config key")
}

func savedVault(t *testing.T, h *harness) string {
	t.Helper()
	cfg, err := config.Load(filepath.Join(h.home, ".config", "bd", "config.json"))
	if err != nil {
		t.Fatal(err)
	}
	return cfg.Vault
}
