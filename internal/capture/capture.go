// Package capture reads the brain dump text from CLI args or piped stdin.
package capture

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

var ErrEmpty = errors.New("nothing to capture")

// Read returns the dump text. Args take precedence; stdin is only read when
// it is piped (not a terminal), so a bare `bd` never blocks waiting for input.
func Read(args []string, stdin io.Reader, stdinIsTTY bool) (string, error) {
	text := strings.Join(args, " ")
	if len(args) == 0 && !stdinIsTTY {
		b, err := io.ReadAll(stdin)
		if err != nil {
			return "", fmt.Errorf("reading stdin: %w", err)
		}
		text = string(b)
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return "", ErrEmpty
	}
	return text, nil
}
