package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestVersion(t *testing.T) {
	h := newHarness(t)

	r := h.bd("--version")

	wantSuccess(t, r)
	wantStdout(t, r, "bd dev\n")
}

func TestDumpArgsGoToTodaysDailyNote(t *testing.T) {
	h := newHarness(t)
	vault := h.vault()

	r := h.bd("kafka", "retries", "need", "backoff")

	wantSuccess(t, r)
	wantStdout(t, r, "added to Daily/2026-10-10.md\n")
	wantFile(t, filepath.Join(vault, todaysNote),
		"## 2026-10-10 14:32\nkafka retries need backoff\n")
}

func TestDumpPipedStdinAppendsToDailyNote(t *testing.T) {
	h := newHarness(t)
	vault := h.vault()
	writeFile(t, filepath.Join(vault, todaysNote), "## 2026-10-10 09:00\nfirst\n")

	r := h.pipe("second thought\nfrom a pipe\n")

	wantSuccess(t, r)
	wantFile(t, filepath.Join(vault, todaysNote),
		"## 2026-10-10 09:00\nfirst\n\n## 2026-10-10 14:32\nsecond thought\nfrom a pipe\n")
}

func TestDumpVaultFlagBeatsEnv(t *testing.T) {
	h := newHarness(t)
	h.vault()
	other := filepath.Join(h.home, "other")
	os.Mkdir(other, 0o755)

	r := h.bd("--vault", other, "hello")

	wantSuccess(t, r)
	wantFile(t, filepath.Join(other, todaysNote), "## 2026-10-10 14:32\nhello\n")
}

func TestDumpEmptyInputFails(t *testing.T) {
	h := newHarness(t)
	vault := h.vault()

	r := h.pipe("  \n")

	wantFailure(t, r, "nothing to capture")
	wantNoFile(t, filepath.Join(vault, todaysNote))
}

func TestDumpWithoutVaultFails(t *testing.T) {
	h := newHarness(t)

	r := h.bd("hello")

	wantFailure(t, r, "no vault configured")
}

func TestDumpCreatesMissingVault(t *testing.T) {
	t.Run("from config", func(t *testing.T) {
		h := newHarness(t)
		dir := filepath.Join(h.home, "notes")
		wantSuccess(t, h.bd("config", "set", "vault", dir))
		os.RemoveAll(dir)

		r := h.bd("hello")

		wantSuccess(t, r)
		wantStdout(t, r, "created vault "+dir+"\nadded to Daily/2026-10-10.md\n")
		wantFile(t, filepath.Join(dir, todaysNote), "## 2026-10-10 14:32\nhello\n")
	})

	t.Run("from BD_VAULT", func(t *testing.T) {
		h := newHarness(t)
		dir := filepath.Join(h.home, "fromenv")
		h.env["BD_VAULT"] = dir

		wantSuccess(t, h.bd("hello"))
		wantFile(t, filepath.Join(dir, todaysNote), "## 2026-10-10 14:32\nhello\n")
	})

	t.Run("from --vault with parents", func(t *testing.T) {
		h := newHarness(t)
		dir := filepath.Join(h.home, "fromflag", "deep")

		wantSuccess(t, h.bd("--vault", dir, "hello"))
		wantFile(t, filepath.Join(dir, todaysNote), "## 2026-10-10 14:32\nhello\n")
	})
}

func TestDumpRejectsVaultThatIsAFile(t *testing.T) {
	h := newHarness(t)
	file := filepath.Join(h.home, "afile.md")
	writeFile(t, file, "x")

	r := h.bd("--vault", file, "nope")

	wantFailure(t, r, "not a directory")
}
