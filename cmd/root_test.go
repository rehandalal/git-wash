package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"reflect"
	"strings"
	"testing"

	"github.com/AlecAivazis/survey/v2"
	"github.com/rehandalal/git-wash/git"
	"github.com/stretchr/testify/assert"
)

type testSurveyor struct {
	input  []interface{}
	cursor int
}

func (ts *testSurveyor) AskOne(p survey.Prompt, response interface{}, opts ...survey.AskOpt) error {
	var input interface{} = nil
	if ts.cursor < len(ts.input) {
		input = ts.input[ts.cursor]
	}

	ts.cursor += 1
	switch p := p.(type) {
	case *survey.Confirm:
		fmt.Println(p.Message)
	case *survey.MultiSelect:
		if len(p.Options) > 0 {
			fmt.Println(p.Message)
		}
	}
	target := reflect.ValueOf(response)
	value := reflect.ValueOf(input)
	if target.Kind() != reflect.Ptr {
		return errors.New("you must pass a pointer as the target of a Write operation")
	}
	elem := target.Elem()
	if elem.Kind() != value.Kind() {
		return errors.New("invalid input")
	}
	if elem.Kind() == reflect.Bool {
		elem.Set(value)
	}
	return nil
}

type GitWashTest struct {
	T          *testing.T
	Stdout     *os.File
	LocalRepo  *git.GitRepo
	RemoteRepo *git.GitRepo
	Input      *[]interface{}
}

func (gwt *GitWashTest) Execute(args ...string) error {
	gwt.T.Helper()
	rootCmd := RootCommand(&testSurveyor{
		input:  *gwt.Input,
		cursor: 0,
	})

	// Mock Stdout
	realStdout := os.Stdout
	defer func() {
		os.Stdout = realStdout
	}()
	os.Stdout = gwt.Stdout

	// Add test specific config to RootCmd
	rootCmd.SetOut(gwt.Stdout)
	rootCmd.SetErr(gwt.Stdout)
	rootCmd.SetArgs(args)

	cwd, _ := os.Getwd()
	defer os.Chdir(cwd)
	os.Chdir(gwt.LocalRepo.Root)

	return rootCmd.Execute()
}

func (gwt GitWashTest) ReadAll() string {
	gwt.Stdout.Seek(0, 0)
	out, _ := io.ReadAll(gwt.Stdout)
	return strings.TrimSpace(string(out))
}

func (gwt *GitWashTest) Write(values ...interface{}) {
	gwt.Input = &values
}

func NewGitWashTest(t *testing.T) *GitWashTest {
	t.Helper()

	stdout, _ := os.CreateTemp("", "*"+t.Name()+"-stdout")

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

	return &GitWashTest{
		T:          t,
		Stdout:     stdout,
		LocalRepo:  &local,
		RemoteRepo: &remote,
		Input:      &[]interface{}{},
	}
}

func Test_Prints_Help(t *testing.T) {
	test := NewGitWashTest(t)
	err := test.Execute("--help")

	assert.Nil(t, err)
	assert.True(t, strings.Contains(test.ReadAll(), "Usage:"))

	test = NewGitWashTest(t)
	err = test.Execute("-h")

	assert.Nil(t, err)
	assert.True(t, strings.Contains(test.ReadAll(), "Usage:"))
}

func Test_Prints_Version(t *testing.T) {
	test := NewGitWashTest(t)
	currentVersion := fmt.Sprintf("%d.%d.%d", VersionMajor, VersionMinor, VersionRevision)

	err := test.Execute("--version")

	assert.Nil(t, err)
	assert.Equal(t, currentVersion, test.ReadAll())

	test = NewGitWashTest(t)
	err = test.Execute("-v")

	assert.Nil(t, err)
	assert.Equal(t, currentVersion, test.ReadAll())
}

func Test_Repo_Doesnt_Exist(t *testing.T) {
	test := NewGitWashTest(t)
	test.LocalRepo.Root = path.Dir(test.LocalRepo.Root)
	err := test.Execute()

	assert.EqualError(t, err, "repo_does_not_exist")
	assert.Contains(t, test.ReadAll(), "Error: Not a git repository (or any of the parent directories).")
}

func Test_Dirty_Working_Tree(t *testing.T) {
	test := NewGitWashTest(t)

	test.LocalRepo.Exec("init")
	os.Create(path.Join(test.LocalRepo.Root, "dirty"))

	err := test.Execute()

	assert.EqualError(t, err, "working_tree_is_dirty")
	assert.Contains(t, test.ReadAll(), "Error: Make sure your working tree is clean before attempting to run this script.")
}

func Test_Prune_Branches(t *testing.T) {
	test := NewGitWashTest(t)

	test.LocalRepo.Exec("checkout", "-b", "prune-me")
	test.LocalRepo.Exec("push", "origin", "prune-me")

	test.LocalRepo.Exec("checkout", "main")
	test.LocalRepo.Exec("branch", "-d", "prune-me")

	gitOut, _ := test.LocalRepo.Exec("branch")
	assert.NotContains(t, gitOut.ToString(), "prune-me")
	gitOut, _ = test.LocalRepo.Exec("branch", "-a")
	assert.Contains(t, gitOut.ToString(), "remotes/origin/prune-me")

	test.RemoteRepo.Exec("branch", "-d", "prune-me")

	test.Write(true)
	err := test.Execute()

	assert.Nil(t, err)
	assert.Contains(t, test.ReadAll(), "Complete!")
	assert.Contains(t, test.ReadAll(), "! No merged branches to delete.")
	assert.Contains(t, test.ReadAll(), "! No squash-merged branches to delete.")

	gitOut, _ = test.LocalRepo.Exec("branch", "-a")
	assert.NotContains(t, gitOut.ToString(), "remotes/origin/prune-me")
}
