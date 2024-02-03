package run

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

func (c *Cmd) cmdRun(cmd *cobra.Command, command gitm.CommandConfig, args []string) error {
	return c.gitm.NewWorker(0, func(ctx context.Context, repoConfig gitm.RepoConfig) error {
		return c.cmdRunRepo(ctx, repoConfig, command, args)
	})
}

func (c *Cmd) cmdRunRepo(ctx context.Context, repoConfig gitm.RepoConfig, command gitm.CommandConfig, args []string) error {
	cmdArgs := []string{"sh", "-c", command.Command}
	cmdArgs = append(cmdArgs, args...)

	allowed := false
	if len(command.Group) == 0 {
		allowed = true
	} else {
		for _, group := range command.Group {
			for _, repoGroup := range repoConfig.Group {
				if group == repoGroup {
					allowed = true
					break
				}
			}

			if allowed {
				break
			}
		}
	}

	if !allowed {
		return nil
	}

	if _, err := os.Stat(repoConfig.Path); err == nil {
		execOpts := gitm.ExecOptions{
			Command:   cmdArgs[0],
			Dir:       &repoConfig.Path,
			LogStdout: true,
			LogStderr: true,
		}

		if len(cmdArgs) > 1 {
			execOpts.Args = cmdArgs[1:]
		}

		if err := c.gitm.Exec(ctx, execOpts); err != nil {
			return fmt.Errorf("Failed to run command %s in repo: %v", command.Name, err)
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

	runCmd := cobra.Command{
		Use:   "run",
		Short: "Runs a command",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintf(os.Stderr, "Command not found: %s\n", args[0])
			cmd.Usage()
			os.Exit(1)
		},
	}

	for _, command := range c.gitm.Config.Commands {
		runCmd.AddCommand(&cobra.Command{
			Use:   command.Name,
			Short: command.Description,
			Run: func(cmd *cobra.Command, args []string) {
				fmt.Println("Running command: ", command.Name, " with args: ", args)
				if err := c.cmdRun(cmd, command, args); err != nil {
					fmt.Fprintf(os.Stderr, "%v\n", err)
					os.Exit(1)
				}

				return
			},
		})
	}

	c.root.AddCommand(&runCmd)

	return &c
}
