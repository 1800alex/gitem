package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

type StatusCmd struct {
	gitm *Gitm
	root *cobra.Command
}

func (c *StatusCmd) cmdStatus(cmd *cobra.Command, args []string) error {
	return c.gitm.NewWorker(0, func(ctx context.Context, repoConfig RepoConfig) error {
		return c.cmdStatusRepo(ctx, repoConfig)
	})
}

func (c *StatusCmd) cmdStatusRepo(ctx context.Context, repoConfig RepoConfig) error {
	if _, err := os.Stat(repoConfig.Path); err == nil {
		err := c.gitm.RunCommandWithOutputFormatting(ctx, "git", []string{"-C", repoConfig.Path, "status"})
		if err != nil {
			return fmt.Errorf("Failed to get repo status: %v", err)
		}
	} else {
		return fmt.Errorf("Repo %s does not exist", c.gitm.repoName)
	}

	return nil
}

func (c *StatusCmd) Init(gitm *Gitm, cmd *cobra.Command) error {
	c.gitm = gitm
	c.root = cmd

	statusCmd := cobra.Command{
		Use:   "status",
		Short: "Get repo status",
		Run: func(cmd *cobra.Command, args []string) {
			if err := c.gitm.Load(cmd, args); err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(1)
			}

			if err := c.cmdStatus(cmd, args); err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(1)
			}
		},
	}

	c.root.AddCommand(&statusCmd)

	return nil
}
