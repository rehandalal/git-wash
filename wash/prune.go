package wash

import (
	"github.com/AlecAivazis/survey/v2"
	"github.com/rehandalal/git-wash/helpers"
)

func (rw RepoWasher) PruneBranches() error {
	shouldPrune := true
	if rw.Options.NoInput {
		helpers.PrintlnC("> Pruning remote branches...", "white+b")
	} else {
		prompt := &survey.Confirm{
			Message: "Do you want to prune remote branches that are deleted or merged?",
		}
		survey.AskOne(prompt, shouldPrune)
	}
	if shouldPrune {
		rw.Repo.Exec("fetch", "--prune")
		helpers.PrintlnC("Complete! 🎉", "green+b")
	} else {
		helpers.PrintlnC("! Skipping...", "yellow")
	}
	return nil
}
