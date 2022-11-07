package wash

import (
	"fmt"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/rehandalal/git-wash/helpers"
)

func (rw RepoWasher) DeleteSquashMergedBranches() error {
	out, _ := rw.Repo.Exec("rev-parse", "--abbrev-ref", "HEAD")
	head := out.ToString()

	out, _ = rw.Repo.Exec("for-each-ref", "refs/heads/", "--format=%(refname:short)")
	branches := out.Lines()

	squashMergedBranches := []string{}
	for i := range branches {
		branch := branches[i]

		out, _ = rw.Repo.Exec("merge-base", head, branch)
		ancestor := out.ToString()

		out, _ = rw.Repo.Exec("rev-parse", fmt.Sprintf("%s^{tree}", branch))
		rp := out.ToString()

		out, _ = rw.Repo.Exec("commit-tree", rp, "-p", ancestor, "-m", "_")
		ct := out.ToString()

		out, _ = rw.Repo.Exec("cherry", head, ct)
		cherry := out.ToString()

		if strings.HasPrefix(cherry, "-") {
			squashMergedBranches = append(squashMergedBranches, branch)
		}
	}

	var removeBranches []string
	if rw.Options.NoInput {
		removeBranches = squashMergedBranches
	} else {
		removeBranches = []string{}
		prompt := &survey.MultiSelect{
			Message: "Which of these squash-merged branches would you like to delete:",
			Options: squashMergedBranches,
		}
		rw.Survey.AskOne(prompt, &removeBranches)
	}

	if rw.Options.NoInput || (!rw.Options.NoInput && len(squashMergedBranches) == 0) {
		helpers.PrintlnC("> Deleting squash-merged branches...", "white+b")
	}

	if len(removeBranches) == 0 {
		helpers.PrintlnC("! No squash-merged branches to delete.", "yellow")
	} else {
		for i := range removeBranches {
			branch := removeBranches[i]
			helpers.PrintC("> Deleting branch: ", "white+b")
			helpers.PrintlnC(branch, "white")

			rw.Repo.DeleteBranch(branch, true)
		}
	}

	if len(removeBranches) > 0 {
		helpers.PrintlnC("Complete! 🎉", "green+b")
	}

	return nil
}
