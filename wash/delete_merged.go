package wash

import (
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/rehandalal/git-wash/helpers"
)

func (rw RepoWasher) DeleteMergedBranches() error {
	// Get a list of merged branches
	out, _ := rw.Repo.Exec("branch", "--merged")
	branches := out.Lines()

	// Filter the current branch out
	mergedBranches := []string{}
	for i := range branches {
		branch := branches[i]
		// The current branch will have an asterisk marking it
		if !strings.HasPrefix(branch, "*") {
			mergedBranches = append(mergedBranches, branch[2:])
		}
	}

	// Get the list of branches to delete
	var deleteBranches []string
	if rw.Options.NoInput {
		deleteBranches = mergedBranches
	} else {
		deleteBranches = []string{}
		prompt := &survey.MultiSelect{
			Message: "Which of these merged branches would you like to delete:",
			Options: mergedBranches,
		}
		rw.Survey.AskOne(prompt, &deleteBranches)
	}

	if rw.Options.NoInput || (!rw.Options.NoInput && len(mergedBranches) == 0) {
		helpers.PrintlnC("> Deleting merged branches...", "white+b")
	}

	if len(deleteBranches) == 0 {
		helpers.PrintlnC("! No merged branches to delete.", "yellow")
	} else {
		// Delete all the branches
		for i := range deleteBranches {
			branch := deleteBranches[i]
			helpers.PrintC("> Deleting branch: ", "white+b")
			helpers.PrintlnC(branch, "white")

			rw.Repo.DeleteBranch(branch, false)
		}
	}

	if len(deleteBranches) > 0 {
		helpers.PrintlnC("Complete! 🎉", "green+b")
	}

	return nil
}
