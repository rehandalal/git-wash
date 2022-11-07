package wash

import (
	"github.com/AlecAivazis/survey/v2"
	"github.com/rehandalal/git-wash/git"
	"github.com/rehandalal/git-wash/types"
)

type Surveyor interface {
	AskOne(p survey.Prompt, response interface{}, opts ...survey.AskOpt) error
}

type RepoWasher struct {
	Repo    *git.GitRepo
	Options *types.RootOptions
	Survey  Surveyor
}
