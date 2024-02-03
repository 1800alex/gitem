package tag

import (
	"context"
	"fmt"
	"os"

	"gitm/internal/gitm"

	"github.com/spf13/cobra"
)

type Cmd struct {
	gitm *gitm.Gitm
	root *cobra.Command
}

func (c *Cmd) cmdTag(cmd *cobra.Command, args []string) error {
	return c.gitm.NewWorker(0, func(ctx context.Context, repoConfig gitm.RepoConfig) error {
		return c.cmdTagRepo(ctx, repoConfig, args)
	})
}

func (c *Cmd) cmdTagRepo(ctx context.Context, repoConfig gitm.RepoConfig, args []string) error {
	cmdArgs := []string{"-C", repoConfig.Path, "tag"}

	cmdArgs = append(cmdArgs, args...)

	if _, err := os.Stat(repoConfig.Path); err == nil {
		if err := c.gitm.RunCommandWithOutputFormatting(ctx, "git", cmdArgs); err != nil {
			return fmt.Errorf("Failed to tag repo: %v", err)
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

	tagCmd := cobra.Command{
		Use:   "tag",
		Short: "Tag",
		Run: func(cmd *cobra.Command, args []string) {
			if err := c.gitm.Init(opts, cmd, args); err != nil {
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

	return &c
}
