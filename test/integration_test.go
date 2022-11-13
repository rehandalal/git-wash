package test

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/rehandalal/git-wash/git"
	expect "github.com/rehandalal/goexpect"
	"github.com/stretchr/testify/assert"
)

const binaryName string = "git-wash"

const (
	dirtyScenario = 1 << iota
	pruneScenario
	mergedScenario
	squashMergedScenario
)

// Make a temporary directory to store the new binary
var tmpDir, _ = os.MkdirTemp("", "*")
var TestBinaryPath = filepath.Join(tmpDir, binaryName)

var update = flag.Bool("update", false, "update golden files")

type ExpectInput struct {
	R *regexp.Regexp
	S string
}

func TestCLI(t *testing.T) {
	tests := []struct {
		name                   string
		args                   []string
		input                  []ExpectInput
		gitScenario            int
		branchesShouldExist    []string
		branchesShouldNotExist []string
	}{
		{"help flag", []string{"--help"}, []ExpectInput{}, 0, []string{}, []string{}},
		{"help flag short", []string{"-h"}, []ExpectInput{}, 0, []string{}, []string{}},
		{"version flag", []string{"--version"}, []ExpectInput{}, 0, []string{}, []string{}},
		{"version flag short", []string{"-v"}, []ExpectInput{}, 0, []string{}, []string{}},
		{"no git repo", []string{}, []ExpectInput{}, 0, []string{}, []string{}},
		{"working tree is dirty", []string{}, []ExpectInput{}, dirtyScenario | pruneScenario, []string{"remotes/origin/prune-me"}, []string{}},
		//{"prune", []string{}, []string{"y"}, pruneScenario, []string{}, []string{"remotes/origin/prune-me"}},
		//{"decline prune", []string{}, []string{"n"}, pruneScenario, []string{"remotes/origin/prune-me"}, []string{}},
		{"skip prune", []string{"--skip-prune"}, []ExpectInput{}, pruneScenario, []string{"remotes/origin/prune-me"}, []string{}},
		{"prune with no input", []string{"--no-input"}, []ExpectInput{}, pruneScenario, []string{"remotes/origin/prune-me"}, []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tName := strings.ReplaceAll(t.Name(), string(os.PathSeparator), "_")
			testDir, err := os.MkdirTemp("", "*"+tName)
			if err != nil {
				t.Fatal(err)
			}

			var remote git.GitRepo
			var local git.GitRepo

			// Check if a git scenario is configured and execute accordingly
			if tt.gitScenario != 0 {
				// Create the remote repo
				tmp, err := os.MkdirTemp("", "*"+tName+"-remote.git")
				if err != nil {
					t.Fatal(err)
				}
				remote = git.GitRepo{Root: tmp}
				remote.Exec("init", "--bare", "-b", "main")
				remote.Exec("config", "commit.gpgsign", "false")

				// Create the local repo
				local = git.GitRepo{Root: testDir}
				local.Exec("clone", remote.Root, local.Root)
				local.Exec("config", "commit.gpgsign", "false")

				// Create the first commit and push it
				os.Create(filepath.Join(local.Root, "test"))
				local.Exec("add", ".")
				local.Exec("commit", "-m", "'First commit'")
				local.Exec("push", "origin", "main")

				assert.Equal(t, "", testDir)

				// Set up dirty working tree
				if (tt.gitScenario & dirtyScenario) != 0 {
					_, err := os.CreateTemp(testDir, "*")
					if err != nil {
						t.Fatal(err)
					}
				}

				// Set up the prune scenario
				if (tt.gitScenario & pruneScenario) != 0 {
					local.Exec("checkout", "-b", "prune-me")
					local.Exec("push", "origin", "prune-me")

					local.Exec("checkout", "main")
					local.Exec("branch", "-d", "prune-me")

					gitOut, _ := local.Exec("branch")
					assert.NotContains(t, gitOut.ToString(), "prune-me")
					gitOut, _ = local.Exec("branch", "-a")
					assert.Contains(t, gitOut.ToString(), "remotes/origin/prune-me")

					remote.Exec("branch", "-d", "prune-me")
				}

				// Validate branches
				if len(tt.branchesShouldExist) > 0 || len(tt.branchesShouldNotExist) > 0 {
					out, _ := local.Exec("branch", "-a")

					for _, b := range tt.branchesShouldExist {
						assert.Contains(t, out.ToString(), b)
					}

					for _, b := range tt.branchesShouldNotExist {
						assert.NotContains(t, out.ToString(), b)
					}
				}
			}

			// Set up CLI execution
			exp, _, err := expect.SpawnWithArgs(
				append([]string{TestBinaryPath}, tt.args...),
				(30 * time.Second),
				expect.Dir(testDir),
			)
			if err != nil {
				t.Fatal(err)
			}

			time.Sleep(100 * time.Millisecond)

			// Configure CLI input
			for _, i := range tt.input {
				exp.Expect(i.R, (30 * time.Second))
				exp.Send(i.S + "\n")
			}

			time.Sleep(100 * time.Millisecond)

			// Get CLI output
			output_string := ""
			buffer := make([]byte, 1)
			for {
				_, err := exp.Read(buffer)
				if err != nil {
					if err == io.EOF {
						break
					}
					t.Fatal(err)
				}
				buffer_string := string(buffer)
				output_string += buffer_string
			}
			output := []byte(output_string)

			// Update snapshot if required
			if *update {
				WriteTestData(t, output)
			}

			// Check snapshot
			actualTestData := string(output)
			expectedTestData := ReadTestData(t)
			assert.Equal(t, expectedTestData, actualTestData)
		})
	}
}

func TestMain(m *testing.M) {
	// Get the working directory
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("could not get working directory: %v", err)
	}

	// Change to project root
	_, filename, _, _ := runtime.Caller(0)
	path := filepath.Dir(filepath.Dir(filename))
	err = os.Chdir(path)
	if err != nil {
		fmt.Printf("could not change directory: %v", err)
		os.Exit(1)
	}

	// Build the git-wash binary
	build := exec.Command("go", "build", "-o", TestBinaryPath)
	_, err = build.Output()
	if err != nil {
		fmt.Printf("could not make binary for %s: %v", binaryName, err)
		os.Exit(1)
	}

	// Restore the original working directory
	os.Chdir(cwd)

	// Run the tests
	os.Exit(m.Run())
}
