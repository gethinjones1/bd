package picker

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"strings"
)

// Candidate is one row in the picker: Path is vault-relative, Display is
// what the user sees and fuzzy-matches against.
type Candidate struct {
	Path, Display string
}

const header = "enter: save · no match + enter: create note · alt-enter: create typed name · esc: daily note"

// FZF runs the fzf binary bin over candidates with dir (the vault) as its
// working directory, so the preview can read notes by relative path.
func FZF(bin, dir string, candidates []Candidate) (Choice, error) {
	var in bytes.Buffer
	for _, c := range candidates {
		in.WriteString(c.Path + "\t" + c.Display + "\n")
	}
	cmd := exec.Command(bin,
		"--print-query",
		"--delimiter", "\t", "--with-nth", "2",
		"--layout", "reverse", "--height", "40%",
		"--prompt", "bd ▸ ", "--header", header,
		"--bind", "alt-enter:print-query",
		"--preview", "cat {1} 2>/dev/null", "--preview-window", "right,50%,border-left",
	)
	cmd.Dir = dir
	cmd.Stdin = &in
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	exit := 0
	if err != nil {
		var ee *exec.ExitError
		if !errors.As(err, &ee) {
			return Choice{}, err
		}
		exit = ee.ExitCode()
	}
	return Parse(string(out), exit)
}

// DisplayName is how a note path is shown in the picker.
func DisplayName(rel string) string {
	return strings.TrimSuffix(rel, ".md")
}
