package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func (gitm *Gitm) cmdClone(cmd *cobra.Command, args []string) error {
	return gitm.cmdWorker(gitm.maxWorkers, func(ctx context.Context, repoConfig RepoConfig) error {
		return gitm.cmdCloneRepo(ctx, repoConfig)
	})
}

func (gitm *Gitm) cmdCloneRepo(ctx context.Context, repoConfig RepoConfig) error {
	if _, err := os.Stat(repoConfig.Path); err == nil {
		if err := gitm.RunCommandWithOutputFormatting(ctx, "git", []string{"-C", repoConfig.Path, "pull"}); err != nil {
			return fmt.Errorf("Failed to pull repo: %v", err)
		}

	} else {
		if err := gitm.RunCommandWithOutputFormatting(ctx, "git", []string{"clone", repoConfig.URL, repoConfig.Path}); err != nil {
			return fmt.Errorf("Failed to clone repo: %v", err)
		}
	}

	if repoConfig.UsesLFS {
		if err := gitm.RunCommandWithOutputFormatting(ctx, "git", []string{"-C", repoConfig.Path, "lfs", "install"}); err != nil {
			return fmt.Errorf("Failed to install lfs repo: %v", err)
		}
	}

	if repoConfig.UsesSubrepos {
		if err := gitm.RunCommandWithOutputFormatting(ctx, "git", []string{"-C", repoConfig.Path, "submodule", "update", "--init", "--recursive"}); err != nil {
			return fmt.Errorf("Failed to initialize submodules: %v", err)
		}
	}

	return nil
}
