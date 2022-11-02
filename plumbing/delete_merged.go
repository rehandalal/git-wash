package plumbing

import (
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/rehandalal/git-wash/helpers"
	"github.com/rehandalal/git-wash/types"
)

func DeleteMergedBranches(cwd string, options *types.RootOptions) error {
	// Get a list of merged branches
	out := helpers.ExecGitOutput(cwd, "branch", "--merged")
	branches := strings.Split(string(out), "\n")

	// Filter the current branch out
	mergedBranches := []string{}
	for i := range branches {
		branch := branches[i]
		// The current branch will have an asterisk marking it
		if !strings.HasPrefix(branch, "*") && len(branch) > 0 {
			mergedBranches = append(mergedBranches, branch[2:])
		}
	}

	// Get the list of branches to delete
	var deleteBranches []string
	if options.NoInput {
		deleteBranches = mergedBranches
	} else {
		deleteBranches = []string{}
		prompt := &survey.MultiSelect{
			Message: "Which of these merged branches would you like to delete:",
			Options: mergedBranches,
		}
		survey.AskOne(prompt, &deleteBranches)
	}

	if options.NoInput || (!options.NoInput && len(mergedBranches) == 0) {
		helpers.PrintlnColorized("> Deleting merged branches...", "white+b")
	}

	if len(deleteBranches) == 0 {
		helpers.PrintlnColorized("! No merged branches to delete.", "yellow")
	} else {
		// Delete all the branches
		for i := range deleteBranches {
			branch := deleteBranches[i]
			helpers.PrintColorized("> Deleting branch: ", "white+b")
			helpers.PrintlnColorized(branch, "white")

			helpers.ExecGitOutput(cwd, "branch", "-d", branch)
		}
	}

	if len(deleteBranches) > 0 {
		helpers.PrintlnColorized("Complete! 🎉", "green+b")
	}

	return nil
}
