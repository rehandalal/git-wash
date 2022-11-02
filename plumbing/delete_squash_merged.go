package plumbing

import (
	"fmt"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/rehandalal/git-wash/helpers"
	"github.com/rehandalal/git-wash/types"
)

func DeleteSquashMergedBranches(cwd string, options *types.RootOptions) error {
	out := helpers.ExecGitOutput(cwd, "rev-parse", "--abbrev-ref", "HEAD")
	head := strings.ReplaceAll(string(out), "\n", "")

	out = helpers.ExecGitOutput(cwd, "for-each-ref", "refs/heads/", "--format=%(refname:short)")
	branches := strings.Split(string(out), "\n")

	squashMergedBranches := []string{}
	for i := range branches {
		branch := branches[i]

		out = helpers.ExecGitOutput(cwd, "merge-base", head, branch)
		ancestor := strings.ReplaceAll(string(out), "\n", "")

		out = helpers.ExecGitOutput(cwd, "rev-parse", fmt.Sprintf("%s^{tree}", branch))
		rp := strings.ReplaceAll(string(out), "\n", "")

		out = helpers.ExecGitOutput(cwd, "commit-tree", rp, "-p", ancestor, "-m", "_")
		ct := strings.ReplaceAll(string(out), "\n", "")

		out = helpers.ExecGitOutput(cwd, "cherry", head, ct)
		cherry := strings.ReplaceAll(string(out), "\n", "")

		if strings.HasPrefix(cherry, "-") {
			squashMergedBranches = append(squashMergedBranches, branch)
		}
	}

	var removeBranches []string
	if options.NoInput {
		removeBranches = squashMergedBranches
	} else {
		removeBranches = []string{}
		prompt := &survey.MultiSelect{
			Message: "Which of these squash-merged branches would you like to delete:",
			Options: squashMergedBranches,
		}
		survey.AskOne(prompt, &removeBranches)
	}

	if options.NoInput || (!options.NoInput && len(squashMergedBranches) == 0) {
		helpers.PrintlnColorized("> Deleting squash-merged branches...", "white+b")
	}

	if len(removeBranches) == 0 {
		helpers.PrintlnColorized("! No squash-merged branches to delete.", "yellow")
	} else {
		for i := range removeBranches {
			branch := removeBranches[i]
			helpers.PrintColorized("> Deleting branch: ", "white+b")
			helpers.PrintlnColorized(branch, "white")

			helpers.ExecGitOutput(cwd, "branch", "-D", branch)
		}
	}

	if len(removeBranches) > 0 {
		helpers.PrintlnColorized("Complete! 🎉", "green+b")
	}

	return nil
}
