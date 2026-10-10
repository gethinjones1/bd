package main

import (
	"fmt"
	"path/filepath"
	"testing"
)

func TestLsReturnsNotesFromVault(t *testing.T) {
	h := newHarness(t)
	vault := h.vault()
	fileName := "Projects/Kafka.md"
	anotherFileName := "Projects/hello_world.md"
	writeFile(t, filepath.Join(vault, fileName), "# Kafka\n")
	writeFile(t, filepath.Join(vault, anotherFileName), "# Hello World\n")
	result := h.bd("ls")
	wantSuccess(t, result)
	wantStdout(t, result, fmt.Sprintf("%s\n%s\n", anotherFileName, fileName))
}
