package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

type FetchCmd struct {
	gitm *Gitm
	root *cobra.Command
}

func (c *FetchCmd) cmdFetch(cmd *cobra.Command, args []string) error {
	return c.gitm.NewWorker(0, func(ctx context.Context, repoConfig RepoConfig) error {
		return c.cmdFetchRepo(ctx, repoConfig)
	})
}

func (c *FetchCmd) cmdFetchRepo(ctx context.Context, repoConfig RepoConfig) error {
	if _, err := os.Stat(repoConfig.Path); err == nil {
		err := c.gitm.RunCommandWithOutputFormatting(ctx, "git", []string{"-C", repoConfig.Path, "fetch"})
		if err != nil {
			return fmt.Errorf("Failed to fetch repo: %v", err)
		}
	} else {
		return fmt.Errorf("Repo %s does not exist", c.gitm.repoName)
	}

	return nil
}

func (c *FetchCmd) Init(gitm *Gitm, cmd *cobra.Command) error {
	c.gitm = gitm
	c.root = cmd

	fetchCmd := cobra.Command{
		Use:   "fetch",
		Short: "Fetch repos",
		Run: func(cmd *cobra.Command, args []string) {
			if err := c.gitm.Load(cmd, args); err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(1)
			}

			if err := c.cmdFetch(cmd, args); err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(1)
			}
		},
	}

	c.root.AddCommand(&fetchCmd)

	return nil
}
