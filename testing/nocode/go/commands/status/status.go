package status

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

func (c *Cmd) cmdStatus(cmd *cobra.Command, args []string) error {
	return c.gitm.NewWorker(0, func(ctx context.Context, repoConfig gitm.RepoConfig) error {
		return c.cmdStatusRepo(ctx, repoConfig)
	})
}

func (c *Cmd) cmdStatusRepo(ctx context.Context, repoConfig gitm.RepoConfig) error {
	if _, err := os.Stat(repoConfig.Path); err == nil {
		err := c.gitm.RunCommandWithOutputFormatting(ctx, "git", []string{"-C", repoConfig.Path, "status"})
		if err != nil {
			return fmt.Errorf("Failed to get repo status: %v", err)
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

	statusCmd := cobra.Command{
		Use:   "status",
		Short: "Get repo status",
		Run: func(cmd *cobra.Command, args []string) {
			if err := c.cmdStatus(cmd, args); err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(1)
			}
		},
	}

	c.root.AddCommand(&statusCmd)

	return &c
}
