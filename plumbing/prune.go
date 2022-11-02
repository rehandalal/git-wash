package plumbing

import (
	"github.com/AlecAivazis/survey/v2"
	"github.com/rehandalal/git-wash/helpers"
	"github.com/rehandalal/git-wash/types"
)

func PruneBranches(cwd string, options *types.RootOptions) error {
	shouldPrune := true
	if options.NoInput {
		helpers.PrintlnColorized("> Pruning remote branches...", "white+b")
	} else {
		prompt := &survey.Confirm{
			Message: "Do you want to prune remote branches that are deleted or merged?",
		}
		survey.AskOne(prompt, &shouldPrune)
	}
	if shouldPrune {
		helpers.ExecGitOutput(cwd, "fetch", "--prune")
		helpers.PrintlnColorized("Complete! 🎉", "green+b")
	} else {
		helpers.PrintlnColorized("! Skipping...", "yellow")
	}
	return nil
}
