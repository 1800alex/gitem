package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func (t *Tag) cmdTag(cmd *cobra.Command, args []string) error {
	return t.gitm.cmdWorker(t.gitm.maxWorkers, func(ctx context.Context, repoConfig RepoConfig) error {
		return t.cmdTagRepo(ctx, repoConfig, args)
	})
}

func (t *Tag) cmdTagRepo(ctx context.Context, repoConfig RepoConfig, args []string) error {
	cmdArgs := []string{"-C", repoConfig.Path, "tag"}

	cmdArgs = append(cmdArgs, args...)

	if _, err := os.Stat(repoConfig.Path); err == nil {
		if err := t.gitm.RunCommandWithOutputFormatting(ctx, "git", cmdArgs); err != nil {
			return fmt.Errorf("Failed to tag repo: %v", err)
		}

	} else {
		return fmt.Errorf("Repo %s does not exist", t.gitm.repoName)
	}

	return nil
}

type Tag struct {
	gitm *Gitm
	root *cobra.Command
}

// TODO move this into its own file

func (f *Tag) Init(gitm *Gitm, cmd *cobra.Command) error {
	f.gitm = gitm
	f.root = cmd

	tagCmd := cobra.Command{
		Use:   "tag",
		Short: "Tag",
		Run: func(cmd *cobra.Command, args []string) {
			if err := f.gitm.Load(cmd, args); err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(1)
			}

			if err := f.cmdTag(cmd, args); err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(1)
			}
		},
	}

	f.root.AddCommand(&tagCmd)

	return nil
}
