package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func (gitm *Gitm) cmdFetch(cmd *cobra.Command, args []string) error {
	return gitm.cmdWorker(gitm.maxWorkers, func(ctx context.Context, repoConfig RepoConfig) error {
		return gitm.cmdFetchRepo(ctx, repoConfig)
	})
}

func (gitm *Gitm) cmdFetchRepo(ctx context.Context, repoConfig RepoConfig) error {
	if _, err := os.Stat(repoConfig.Path); err == nil {
		err := gitm.RunCommandWithOutputFormatting(ctx, "git", []string{"-C", repoConfig.Path, "fetch"})
		if err != nil {
			return fmt.Errorf("Failed to fetch repo: %v", err)
		}
	} else {
		return fmt.Errorf("Repo %s does not exist", gitm.repoName)
	}

	return nil
}
