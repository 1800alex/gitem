package pull

import (
	"context"
	"fmt"
	"gitm/internal/gitm"
	"os"

	"github.com/spf13/cobra"
)

type Cmd struct {
	gitm *gitm.Gitm
	root *cobra.Command
}

func (c *Cmd) cmdPull(cmd *cobra.Command, args []string) error {
	return c.gitm.NewWorker(0, func(ctx context.Context, repoConfig gitm.RepoConfig) error {
		return c.cmdPullRepo(ctx, repoConfig)
	})
}

func (c *Cmd) cmdPullRepo(ctx context.Context, repoConfig gitm.RepoConfig) error {
	if _, err := os.Stat(repoConfig.Path); err == nil {
		if err := c.gitm.RunCommandWithOutputFormatting(ctx, "git", []string{"-C", repoConfig.Path, "pull"}); err != nil {
			return fmt.Errorf("Failed to pull repo: %v", err)
		}

		if repoConfig.UsesLFS {
			if err := c.gitm.RunCommandWithOutputFormatting(ctx, "git", []string{"-C", repoConfig.Path, "lfs", "pull"}); err != nil {
				return fmt.Errorf("Failed to pull lfs repo: %v", err)
			}
		}

		if repoConfig.UsesSubrepos {
			if err := c.gitm.RunCommandWithOutputFormatting(ctx, "git", []string{"-C", repoConfig.Path, "submodule", "update", "--init", "--recursive"}); err != nil {
				return fmt.Errorf("Failed to initialize submodules: %v", err)
			}
		}
	} else {
		return fmt.Errorf("Repo %s does not exist", repoConfig.Name)
	}

	return nil
}

// New creates a new gitm command
func New(gitm *gitm.Gitm, opts *gitm.GitmOptions, root *cobra.Command) *Cmd {
	c := Cmd{}

	c.gitm = gitm
	c.root = root

	pullCmd := cobra.Command{
		Use:   "pull",
		Short: "Pull repos",
		Run: func(cmd *cobra.Command, args []string) {
			if err := c.gitm.Init(opts, cmd, args); err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(1)
			}

			if err := c.cmdPull(cmd, args); err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(1)
			}
		},
	}

	c.root.AddCommand(&pullCmd)

	return &c
}