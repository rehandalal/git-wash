package helpers

import (
	"errors"
	"os"
	"os/exec"
	"path"
	"path/filepath"
)

func ExecGitOutput(cwd string, args ...string) []byte {
	// Execute a Git command and return the output
	cmd := exec.Command("git", args...)
	cmd.Dir = cwd
	out, _ := cmd.Output()
	return out
}

func GetCurrentGitRepoRoot() (string, error) {
	// Find the nearest ancestor that is a git repo
	cwd, _ := os.Getwd()
	_, err := os.Stat(path.Join(cwd, ".git"))

	for errors.Is(err, os.ErrNotExist) && len([]rune(cwd)) > 1 {
		cwd = filepath.Dir(cwd)
		_, err = os.Stat(path.Join(cwd, ".git"))
	}

	if errors.Is(err, os.ErrNotExist) {
		return cwd, err
	}

	return cwd, nil
}

func IsGitWorkingTreeClean(cwd string) bool {
	// Check if the working tree is clean
	out := ExecGitOutput(cwd, "status", "-s")
	return len(out) == 0
}
