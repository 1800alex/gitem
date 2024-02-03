package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

type TagCmd struct {
	gitm *Gitm
	root *cobra.Command
}

func (c *TagCmd) cmdTag(cmd *cobra.Command, args []string) error {
	return c.gitm.NewWorker(0, func(ctx context.Context, repoConfig RepoConfig) error {
		return c.cmdTagRepo(ctx, repoConfig, args)
	})
}

func (c *TagCmd) cmdTagRepo(ctx context.Context, repoConfig RepoConfig, args []string) error {
	cmdArgs := []string{"-C", repoConfig.Path, "tag"}

	cmdArgs = append(cmdArgs, args...)

	if _, err := os.Stat(repoConfig.Path); err == nil {
		if err := c.gitm.RunCommandWithOutputFormatting(ctx, "git", cmdArgs); err != nil {
			return fmt.Errorf("Failed to tag repo: %v", err)
		}

	} else {
		return fmt.Errorf("Repo %s does not exist", c.gitm.repoName)
	}

	return nil
}

func (c *TagCmd) Init(gitm *Gitm, cmd *cobra.Command) error {
	c.gitm = gitm
	c.root = cmd

	tagCmd := cobra.Command{
		Use:   "tag",
		Short: "Tag",
		Run: func(cmd *cobra.Command, args []string) {
			if err := c.gitm.Load(cmd, args); err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(1)
			}

			if err := c.cmdTag(cmd, args); err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(1)
			}
		},
	}

	c.root.AddCommand(&tagCmd)

	return nil
}
