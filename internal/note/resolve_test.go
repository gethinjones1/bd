package note

import (
	"errors"
	"testing"
)

func TestResolve(t *testing.T) {
	tests := []struct {
		name, rel, want string
		wantErr         error
	}{
		{name: "adds .md", rel: "Projects/Kafka", want: "/vault/Projects/Kafka.md"},
		{name: "keeps .md", rel: "Projects/Kafka.md", want: "/vault/Projects/Kafka.md"},
		{name: "spaces are fine", rel: "Ideas/Rust CLI", want: "/vault/Ideas/Rust CLI.md"},
		{name: "dots inside name", rel: "v1.2 notes", want: "/vault/v1.2 notes.md"},
		{name: "cleans inner dots", rel: "a/../b", want: "/vault/b.md"},
		{name: "escape with dotdot", rel: "../outside", wantErr: ErrOutsideVault},
		{name: "absolute path", rel: "/etc/passwd", wantErr: ErrOutsideVault},
		{name: "blank", rel: "  ", wantErr: ErrOutsideVault},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Resolve("/vault", tt.rel)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
