package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func (gitm *Gitm) cmdPull(cmd *cobra.Command, args []string) error {
	return gitm.cmdWorker(gitm.maxWorkers, func(ctx context.Context, repoConfig RepoConfig) error {
		return gitm.cmdPullRepo(ctx, repoConfig)
	})
}

func (gitm *Gitm) cmdPullRepo(ctx context.Context, repoConfig RepoConfig) error {
	if _, err := os.Stat(repoConfig.Path); err == nil {
		if err := gitm.RunCommandWithOutputFormatting(ctx, "git", []string{"-C", repoConfig.Path, "pull"}); err != nil {
			return fmt.Errorf("Failed to pull repo: %v", err)
		}

		if repoConfig.UsesLFS {
			if err := gitm.RunCommandWithOutputFormatting(ctx, "git", []string{"-C", repoConfig.Path, "lfs", "pull"}); err != nil {
				return fmt.Errorf("Failed to pull lfs repo: %v", err)
			}
		}

		if repoConfig.UsesSubrepos {
			if err := gitm.RunCommandWithOutputFormatting(ctx, "git", []string{"-C", repoConfig.Path, "submodule", "update", "--init", "--recursive"}); err != nil {
				return fmt.Errorf("Failed to initialize submodules: %v", err)
			}
		}
	} else {
		return fmt.Errorf("Repo %s does not exist", gitm.repoName)
	}

	return nil
}
