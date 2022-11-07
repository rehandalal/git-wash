package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/rehandalal/git-wash/git"
	"github.com/rehandalal/git-wash/helpers"
	"github.com/rehandalal/git-wash/types"
	"github.com/rehandalal/git-wash/wash"
	"github.com/spf13/cobra"
)

const VersionMajor int = 1
const VersionMinor int = 0
const VersionRevision int = 0

func getOptions(cmd *cobra.Command) *types.RootOptions {
	// Returns a struct with the flags/options passed to the CLI

	version, _ := cmd.Flags().GetBool("version")
	noInput, _ := cmd.Flags().GetBool("no-input")
	skipPrune, _ := cmd.Flags().GetBool("skip-prune")
	skipMerged, _ := cmd.Flags().GetBool("skip-merged")
	skipSquashMerged, _ := cmd.Flags().GetBool("skip-squash-merged")

	return &types.RootOptions{
		Version:          version,
		NoInput:          noInput,
		SkipPrune:        skipPrune,
		SkipMerged:       skipMerged,
		SkipSquashMerged: skipSquashMerged,
	}
}

func getRootCommandRunE(survey wash.Surveyor) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		// Get the options for the CLI
		options := getOptions(cmd)

		// Show version info if flag was passed
		if options.Version {
			fmt.Printf("%d.%d.%d\n", VersionMajor, VersionMinor, VersionRevision)
			return nil
		}

		// Attempt to find the closest git repo
		cwd, _ := os.Getwd()

		repo, err := git.GetClosestGitRepo(cwd)
		if err != nil {
			helpers.PrintlnC(
				"Error: Not a git repository (or any of the parent directories).",
				"red+b",
			)
			return errors.New("repo_does_not_exist")
		}

		// Make sure the working tree is clean
		if !repo.IsClean() {
			helpers.PrintlnC(
				"Error: Make sure your working tree is clean before attempting to run this script.",
				"red+b",
			)
			return errors.New("working_tree_is_dirty")
		}

		// Create the repo washer
		rw := wash.RepoWasher{
			Repo:    repo,
			Options: options,
			Survey:  survey,
		}

		// Prune remote branches
		if !options.SkipPrune {
			err = rw.PruneBranches()
			if err != nil {
				return err
			}
		}

		// Delete merged branches
		if !options.SkipMerged {
			err = rw.DeleteMergedBranches()
			if err != nil {
				return err
			}
		}

		// Delete squash merged branches
		if !options.SkipSquashMerged {
			err = rw.DeleteSquashMergedBranches()
			if err != nil {
				return err
			}
		}

		return nil
	}
}

func RootCommand(survey wash.Surveyor) *cobra.Command {
	// Initialize the root command
	cmd := &cobra.Command{
		Use:           "git-wash",
		Short:         "A Git extension to clean up your repo",
		Long:          "A Git extension to clean up your repo",
		RunE:          getRootCommandRunE(survey),
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	// Initialize the flags
	cmd.Flags().BoolP("help", "h", false, "Show help for git-wash")
	cmd.Flags().BoolP("version", "v", false, "Show the version for git-wash")

	cmd.Flags().BoolP("no-input", "y", false, "Do not prompt for input")

	cmd.Flags().Bool("skip-prune", false, "Skips pruning of remote branches that are deleted or merged")
	cmd.Flags().Bool("skip-merged", false, "Skips deletion of branches that have been merged")
	cmd.Flags().Bool("skip-squash-merged", false, "Skips deletion of branches that have been squash merged")

	return cmd
}

type coreSurveyor struct{}

func (cs coreSurveyor) AskOne(p survey.Prompt, response interface{}, opts ...survey.AskOpt) error {
	return survey.AskOne(p, response, opts...)
}

func Execute() {
	cs := coreSurveyor{}
	rootCmd := RootCommand(cs)
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
