# bd: brain dump

A quick CLI for getting thoughts out of your head and into your [Obsidian](https://obsidian.md) vault before they're gone.

```sh
bd kafka retries need exponential backoff
```

bd opens a picker with **today's daily note** highlighted. Press **Enter** to save, or type a few letters to pick a different note.

## Install

**macOS and Linux, no Go needed:**

```sh
curl -fsSL https://raw.githubusercontent.com/gethinjones1/bd/main/install.sh | sh
```

This downloads the right binary for your machine, checks its checksum, and installs it to `~/.local/bin`. Options:

```sh
curl -fsSL .../install.sh | BD_VERSION=v0.1.0 sh            # a specific version
curl -fsSL .../install.sh | BD_INSTALL_DIR=/usr/local/bin sh # somewhere else
```

The picker needs [fzf](https://github.com/junegunn/fzf): `brew install fzf` or `sudo apt install fzf`. Without fzf, bd still works, but dumps go straight to your daily note.

You can also download a binary yourself from the [releases page](https://github.com/gethinjones1/bd/releases).

**With Go** (1.25+):

```sh
go install github.com/gethinjones1/bd/cmd/bd@latest   # installs to ~/go/bin
```

## Setup

Point bd at your vault (it's created if it doesn't exist):

```sh
bd config set vault ~/Obsidian/Main
```

This writes `~/.config/bd/config.json`.

## Usage

```sh
bd thought about kafka retries      # no quotes needed
bd "has punctuation? quote it!"     # quote anything your shell would trip on
pbpaste | bd                        # dump your clipboard
bd --daily standup notes            # skip the picker, go straight to the daily note
bd ls                               # list the notes in your vault
bd --help                           # full usage
```

Each dump is appended to the chosen note as a timestamped section:

```md
## 2026-10-10 14:32
thought about kafka retries
```

Daily notes live at `<vault>/Daily/YYYY-MM-DD.md` and are created when needed.

### The picker

| You do | Result |
|---|---|
| **Enter** | Save to today's daily note |
| Type `kaf` → **Enter** | Save to the best match, e.g. `Projects/Kafka.md` |
| Type a name that matches nothing → **Enter** | Create that note, e.g. `Ideas/Rust CLI.md` |
| **Esc** | Cancel. Nothing is saved |

Notes are listed most recently edited first, with a preview pane. `.obsidian`, `.trash` and other hidden folders are skipped. If you type a path outside the vault, bd refuses it and saves to the daily note instead. It does the same if fzf fails.

The picker only appears when bd is run from a terminal. In scripts, or when output is redirected, dumps go straight to the daily note.

## Configuration

bd picks the vault from the first of these that is set:

1. `--vault PATH`
2. `BD_VAULT`
3. `vault` in `~/.config/bd/config.json` (or `$XDG_CONFIG_HOME/bd/config.json`)

```json
{
  "vault": "/Users/you/Obsidian/Main",
  "daily_dir": "Daily"
}
```

`daily_dir` is the folder for daily notes, relative to the vault.

| Environment variable | Purpose |
|---|---|
| `BD_VAULT` | Vault directory |
| `BD_INTERACTIVE` | `1` or `0` to force the picker on or off |
| `BD_FZF` | Picker binary to run instead of `fzf` (e.g. `sk`) |

## Development

bd is built test-first: write a failing test, make it pass, then refactor.

```sh
make test                  # go test ./...
go test -count=1 ./...     # skip the test cache
go test ./cmd/bd -run TestPicker -v
make build                 # builds bin/bd
make test-install          # tests install.sh against a locally built fake release
```

### Releasing

Push a version tag:

```sh
git tag v0.1.0 && git push origin v0.1.0
```

`.github/workflows/release.yml` runs the tests, then [GoReleaser](https://goreleaser.com) (`.goreleaser.yaml`) builds macOS and Linux binaries for amd64 and arm64 and publishes them, with `checksums.txt`, as a GitHub Release. `install.sh` downloads from that release. To try a release build locally without publishing:

```sh
go run github.com/goreleaser/goreleaser/v2@latest release --snapshot --clean   # output in dist/
```

| Path | What it does |
|---|---|
| `cmd/bd/` | The CLI. `main.go` sends each subcommand to its own file (`dump.go`, `config.go`, `help.go`, …) |
| `internal/capture` | Reads the dump from arguments or piped stdin |
| `internal/config` | Config file, vault selection, creating the vault folder |
| `internal/vault` | Lists the notes in a vault |
| `internal/picker` | Runs fzf and interprets what you chose |
| `internal/note` | Formats entries, daily note paths, safe paths, appending |

The command tests (`cmd/bd/*_test.go`) run the real `run()` in-process against a fake world: a temp `HOME`, a fixed clock (10 Oct 2026, 14:32), and a scripted picker. The shared setup is in `cmd/bd/helpers_test.go`:

```go
h := newHarness(t)
vault := h.vault()
h.interactive = true
h.pickFn = selectNote("Projects/Kafka.md")

r := h.bd("retries", "need", "backoff")

wantSuccess(t, r)
wantStdout(t, r, "added to Projects/Kafka.md\n")
```

## Roadmap

- **Smarter suggestions:** pull keywords out of the dump and use ripgrep to rank the notes it most likely belongs in, just below the daily note

