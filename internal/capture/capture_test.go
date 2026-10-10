package capture

import (
	"errors"
	"strings"
	"testing"
)

func TestRead(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		stdin      string
		stdinIsTTY bool
		want       string
		wantErr    error
	}{
		{name: "single arg", args: []string{"kafka retries"}, stdinIsTTY: true, want: "kafka retries"},
		{name: "unquoted args are joined", args: []string{"kafka", "retries"}, stdinIsTTY: true, want: "kafka retries"},
		{name: "piped stdin", stdin: "line one\nline two\n", want: "line one\nline two"},
		{name: "arg wins over stdin", args: []string{"from arg"}, stdin: "from stdin", want: "from arg"},
		{name: "tty stdin is not read", stdin: "should not be read", stdinIsTTY: true, wantErr: ErrEmpty},
		{name: "whitespace only is empty", args: []string{"  \n\t"}, stdinIsTTY: true, wantErr: ErrEmpty},
		{name: "empty pipe is empty", stdin: "", wantErr: ErrEmpty},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Read(tt.args, strings.NewReader(tt.stdin), tt.stdinIsTTY)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
