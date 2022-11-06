package git

import (
	"errors"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

type GitOutput struct {
	Bytes []byte
}

func (out GitOutput) Length() int {
	return len(out.Bytes)
}

func (out GitOutput) ToString() string {
	return strings.TrimSpace(string(out.Bytes))
}

func (out GitOutput) Lines() []string {
	return strings.Split(out.ToString(), "\n")
}

type GitRepo struct {
	Root string
}

func (repo *GitRepo) Exec(args ...string) (GitOutput, error) {
	// Execute a Git command and return the output
	cmd := exec.Command("git", args...)
	cmd.Dir = repo.Root
	out, err := cmd.Output()
	return GitOutput{Bytes: out}, err
}

func (repo *GitRepo) IsClean() bool {
	// Check if the repos working tree is clean
	out, err := repo.Exec("status", "-s")
	if err != nil {
		return false
	}
	return out.Length() == 0
}

func (repo GitRepo) DeleteBranch(name string, force bool) {
	// Delete a branch from a repo
	deleteFlag := "-d"
	if force {
		deleteFlag = "-D"
	}
	repo.Exec("branch", deleteFlag, name)
}

func GetClosestGitRepo(dir string) (*GitRepo, error) {
	// Find the nearest ancestor that is a git repo
	_, err := os.Stat(path.Join(dir, ".git"))

	if errors.Is(err, os.ErrNotExist) {
		if len([]rune(dir)) > 1 {
			return GetClosestGitRepo(filepath.Dir(dir))
		} else {
			return &GitRepo{Root: dir}, err
		}
	}

	return &GitRepo{Root: dir}, nil
}
