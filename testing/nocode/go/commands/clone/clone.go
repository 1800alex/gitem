package clone

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"gitm/internal/gitm"
)

type Cmd struct {
	gitm *gitm.Gitm
	root *cobra.Command
}

func (c *Cmd) cmdClone(cmd *cobra.Command, args []string) error {
	return c.gitm.NewWorker(0, func(ctx context.Context, repoConfig gitm.RepoConfig) error {
		return c.cmdCloneRepo(ctx, repoConfig)
	})
}

func (c *Cmd) cmdCloneRepo(ctx context.Context, repoConfig gitm.RepoConfig) error {
	if _, err := os.Stat(repoConfig.Path); err == nil {
		if err := c.gitm.RunCommandWithOutputFormatting(ctx, "git", []string{"-C", repoConfig.Path, "pull"}); err != nil {
			return fmt.Errorf("Failed to pull repo: %v", err)
		}

	} else {
		if err := c.gitm.RunCommandWithOutputFormatting(ctx, "git", []string{"clone", repoConfig.URL, repoConfig.Path}); err != nil {
			return fmt.Errorf("Failed to clone repo: %v", err)
		}
	}

	if repoConfig.UsesLFS {
		if err := c.gitm.RunCommandWithOutputFormatting(ctx, "git", []string{"-C", repoConfig.Path, "lfs", "install"}); err != nil {
			return fmt.Errorf("Failed to install lfs repo: %v", err)
		}
	}

	if repoConfig.UsesSubrepos {
		if err := c.gitm.RunCommandWithOutputFormatting(ctx, "git", []string{"-C", repoConfig.Path, "submodule", "update", "--init", "--recursive"}); err != nil {
			return fmt.Errorf("Failed to initialize submodules: %v", err)
		}
	}

	return nil
}

// New creates a new gitm command
func New(gitm *gitm.Gitm, opts *gitm.GitmOptions, root *cobra.Command) *Cmd {
	c := Cmd{}

	c.gitm = gitm
	c.root = root

	cloneCmd := cobra.Command{
		Use:   "clone",
		Short: "Clone repos",
		Run: func(cmd *cobra.Command, args []string) {
			if err := c.cmdClone(cmd, args); err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(1)
			}
		},
	}

	c.root.AddCommand(&cloneCmd)

	return &c
}
