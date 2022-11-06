package cmd

import (
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/rehandalal/git-wash/git"
	"github.com/stretchr/testify/assert"
)

func execute(t *testing.T, args ...string) (string, error) {
	t.Helper()

	// Mock os.Stdout
	realStdout := os.Stdout
	fakeStdoutReader, fakeStdout, _ := os.Pipe()
	os.Stdout = fakeStdout

	rootCmd := RootCommand()

	// Add test specific config to RootCmd
	rootCmd.SetOut(os.Stdout)
	rootCmd.SetErr(os.Stdout)
	rootCmd.SetArgs(args)

	// Execute
	err := rootCmd.Execute()

	// Read stdout and restore mock
	fakeStdout.Close()
	out, _ := io.ReadAll(fakeStdoutReader)
	os.Stdout = realStdout

	return strings.TrimSpace(string(out)), err
}

func executeWithDirectory(t *testing.T, dir string, args ...string) (string, error) {
	t.Helper()

	cwd, _ := os.Getwd()
	defer os.Chdir(cwd)

	os.Chdir(dir)
	out, err := execute(t, args...)

	return out, err
}

func setupGitDirectory(t *testing.T) (git.GitRepo, git.GitRepo) {
	t.Helper()

	tmp, _ := os.MkdirTemp("", "*"+t.Name()+"-remote.git")
	remote := git.GitRepo{Root: tmp}
	remote.Exec("init", "--bare", "-b", "main")
	remote.Exec("config", "commit.gpgsign", "false")

	tmp, _ = os.MkdirTemp("", "*"+t.Name())
	local := git.GitRepo{Root: tmp}
	local.Exec("clone", remote.Root, local.Root)
	local.Exec("config", "commit.gpgsign", "false")

	os.Create(path.Join(local.Root, "first"))
	local.Exec("add", ".")
	local.Exec("commit", "-m", "'First commit'")
	local.Exec("push", "origin", "main")

	return local, remote
}

func Test_Prints_Help(t *testing.T) {
	out, err := execute(t, "--help")

	assert.Nil(t, err)
	assert.True(t, strings.Contains(out, "Usage:"))

	out, err = execute(t, "-h")

	assert.Nil(t, err)
	assert.True(t, strings.Contains(out, "Usage:"))
}

func Test_Prints_Version(t *testing.T) {
	currentVersion := fmt.Sprintf("%d.%d.%d", VersionMajor, VersionMinor, VersionRevision)

	out, err := execute(t, "--version")

	assert.Nil(t, err)
	assert.Equal(t, currentVersion, out)

	out, err = execute(t, "-v")

	assert.Nil(t, err)
	assert.Equal(t, currentVersion, out)
}

func Test_Repo_Doesnt_Exist(t *testing.T) {
	tmp, _ := os.MkdirTemp("", "*"+t.Name())
	out, err := executeWithDirectory(t, tmp)

	assert.EqualError(t, err, "repo_does_not_exist")
	assert.Contains(t, out, "Error: Not a git repository (or any of the parent directories).")
}

func Test_Dirty_Working_Tree(t *testing.T) {
	local, _ := setupGitDirectory(t)

	local.Exec("init")
	os.Create(path.Join(local.Root, "dirty"))

	out, err := executeWithDirectory(t, local.Root)

	assert.EqualError(t, err, "working_tree_is_dirty")
	assert.Contains(t, out, "Error: Make sure your working tree is clean before attempting to run this script.")
}

// TODO: Fix this test with appropriate mocking
// func Test_Prune_Branches(t *testing.T) {
// 	local, remote := setupGitDirectory(t)

// 	local.Exec("checkout", "-b", "prune-me")
// 	local.Exec("push", "origin", "prune-me")

// 	local.Exec("checkout", "main")
// 	local.Exec("branch", "-d", "prune-me")

// 	gitOut, _ := local.Exec("branch")
// 	assert.NotContains(t, gitOut.ToString(), "prune-me")
// 	gitOut, _ = local.Exec("branch", "-a")
// 	assert.Contains(t, gitOut.ToString(), "origin/prune-me")

// 	remote.Exec("branch", "-d", "prune-me")
// 	out, err := executeWithDirectory(t, local.Root)

// 	assert.Nil(t, err)
// 	assert.Equal(t, "", out)

// 	gitOut, _ = local.Exec("branch", "-a")
// 	assert.NotContains(t, gitOut.ToString(), "origin/prune-me")
// }
