package wash

import (
	"github.com/rehandalal/git-wash/git"
	"github.com/rehandalal/git-wash/types"
)

type RepoWasher struct {
	Repo    *git.GitRepo
	Options *types.RootOptions
}
