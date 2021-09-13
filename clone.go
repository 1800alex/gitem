package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/1800alex/go-git-cmd-wrapper/v2/clone"
	"github.com/1800alex/go-git-cmd-wrapper/v2/git"
	"github.com/1800alex/go-git-cmd-wrapper/v2/types"

	"github.com/spf13/cobra"
)

func (g *Gitem) clone(repo RepoConfig) error {
	var ctx context.Context
	var cancel context.CancelFunc
	if repo.Timeout > 0 {
		ctx, cancel = context.WithDeadline(g.ctx, time.Now().Add(time.Duration(repo.Timeout)*time.Second))
	} else {
		ctx, cancel = context.WithCancel(g.ctx)
	}

	defer cancel()

	if g.config.Debug {
		g.Info.Fprintln(os.Stdout, "cloning", repo.Name)
	}

	var options []types.Option

	options = append(options, clone.Repository(repo.URL))
	options = append(options, clone.Directory(repo.Path))

	if g.config.Debug {
		options = append(options, git.Debugger(true))
	}

	if repo.Depth > 0 {
		options = append(options, clone.Depth(fmt.Sprintf("%d", repo.Depth)))
	}

	if g.config.Mock {
		options = append(options, git.CmdExecutor(cmdExecutorMock))
	}

	out, err := git.CloneWithContext(ctx, options...)

	if err != nil {
		return err
	}

	if g.config.Debug {
		g.Info.Fprintln(os.Stdout, repo.Name, "clone", out)
	}

	return nil
}

func (g *Gitem) cloneCmd() **cobra.Command {

	// var ip string
	// var timeout int

	var cmd = &cobra.Command{
		Use:   "clone",
		Short: "Retrieve the passwords for a BATS system",
		Long:  `password is for decoding the passwords on a BATS system.`,
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			for _, repo := range g.config.Repos {
				g.wg.Add(1)
				go func(repo2 RepoConfig) {
					err := g.clone(repo2)
					if err != nil {
						if err.Error() == "exit status 128" {
							g.Info.Fprintln(os.Stdout, "cloning", repo.Name, "failed because it exists already, trying to pull instead")
							err = g.pull(repo)

							if err != nil {
								g.Error.Fprintln(os.Stderr, repo.Name, "pull failed", err)
							}
							return
						}

						g.Error.Fprintln(os.Stderr, repo.Name, "stash failed", err)
					}
					g.wg.Done()
				}(repo)
			}
		},
	}

	// cmd.Flags().StringVarP(&ip, "ip", "i", "127.0.0.1", "the ip to port knock")
	// cmd.MarkFlagRequired("ip")
	// cmd.Flags().IntVarP(&timeout, "timeout", "t", 2000, "the connection timeout in ms")

	return &cmd
}
