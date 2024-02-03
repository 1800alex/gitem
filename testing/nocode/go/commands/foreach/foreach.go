package foreach

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

func (c *Cmd) cmdForEach(cmd *cobra.Command, args []string) error {
	return c.gitm.NewWorker(0, func(ctx context.Context, repoConfig gitm.RepoConfig) error {
		return c.cmdForEachRepo(ctx, repoConfig, args)
	})
}

func (c *Cmd) cmdForEachRepo(ctx context.Context, repoConfig gitm.RepoConfig, args []string) error {
	if _, err := os.Stat(repoConfig.Path); err == nil {
		execOpts := gitm.ExecOptions{
			Command:   args[0],
			Dir:       &repoConfig.Path,
			LogStdout: true,
			LogStderr: true,
		}

		if len(args) > 1 {
			execOpts.Args = args[1:]
		}

		if err := c.gitm.Exec(ctx, execOpts); err != nil {
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

	foreachCmd := cobra.Command{
		Use:   "foreach",
		Short: "ForEach",
		Run: func(cmd *cobra.Command, args []string) {
			if err := c.gitm.Init(opts, cmd, args); err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(1)
			}

			if err := c.cmdForEach(cmd, args); err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(1)
			}
		},
	}

	c.root.AddCommand(&foreachCmd)

	return &c
}
