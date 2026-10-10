package main

import (
	"os"
	"time"
)

// clock returns the current time, or BD_NOW (RFC 3339) when set so tests
// can pin the date.
func clock() (time.Time, error) {
	if v := os.Getenv("BD_NOW"); v != "" {
		return time.Parse(time.RFC3339, v)
	}
	return time.Now(), nil
}

// interactive reports whether to show the picker: BD_INTERACTIVE=1/0 forces
// it on/off, otherwise only when stdout is a terminal.
func interactive() bool {
	switch os.Getenv("BD_INTERACTIVE") {
	case "1":
		return true
	case "0":
		return false
	}
	return isTerminal(os.Stdout)
}

func isTerminal(r any) bool {
	f, ok := r.(*os.File)
	if !ok {
		return false
	}
	info, err := f.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

func fzfBin() string {
	if bin := os.Getenv("BD_FZF"); bin != "" {
		return bin
	}
	return "fzf"
}
