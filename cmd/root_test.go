package cmd

import (
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/rehandalal/git-wash/helpers"
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
	os.Chdir(dir)
	out, err := execute(t, args...)
	os.Chdir(cwd)

	return out, err
}

func setupGitDirectory(t *testing.T) string {
	t.Helper()

	tmp, _ := os.MkdirTemp("", "*"+t.Name())
	helpers.ExecGitOutput(tmp, "init")
	helpers.ExecGitOutput(tmp, "config", "commit.gpgsign", "false")

	os.Create(path.Join(tmp, "first"))
	helpers.ExecGitOutput(tmp, "add", ".")
	helpers.ExecGitOutput(tmp, "commit", "-m", "'First commit'")

	return tmp
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
	assert.Contains(t, out, "Error: Repository does not exist.")
}

func Test_Dirty_Working_Tree(t *testing.T) {
	tmp := setupGitDirectory(t)

	helpers.ExecGitOutput(tmp, "init")
	os.Create(path.Join(tmp, "dirty"))

	out, err := executeWithDirectory(t, tmp)

	assert.EqualError(t, err, "working_tree_is_dirty")
	assert.Contains(t, out, "Error: Make sure your working tree is clean before attempting to run this script.")
}
