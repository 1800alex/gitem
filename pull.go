package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/1800alex/go-git-cmd-wrapper/v2/git"
	"github.com/1800alex/go-git-cmd-wrapper/v2/pull"
	"github.com/1800alex/go-git-cmd-wrapper/v2/types"

	"github.com/spf13/cobra"
)

func (g *Gitem) pull(repo RepoConfig) error {
	var ctx context.Context
	var cancel context.CancelFunc
	if repo.Timeout > 0 {
		ctx, cancel = context.WithDeadline(g.ctx, time.Now().Add(time.Duration(repo.Timeout)*time.Second))
	} else {
		ctx, cancel = context.WithCancel(g.ctx)
	}

	defer cancel()

	if g.config.Debug {
		g.Info.Fprintln(os.Stdout, "pulling", repo.Name)
	}

	var options []types.Option

	options = append(options, pull.Repository(repo.URL))
	options = append(options, git.WorkingDir(repo.Path))

	if g.config.Debug {
		options = append(options, git.Debugger(true))
	}

	if repo.Depth > 0 {
		options = append(options, pull.Depth(fmt.Sprintf("%d", repo.Depth)))
	}

	if g.config.Mock {
		options = append(options, git.CmdExecutor(cmdExecutorMock))
	}

	out, err := git.PullWithContext(ctx, options...)

	if err != nil {
		return err
	}

	if g.config.Debug {
		g.Info.Fprintln(os.Stdout, repo.Name, "pull", out)
	}

	return nil
}

func (g *Gitem) pullCmd() **cobra.Command {

	// var ip string
	// var timeout int

	var cmd = &cobra.Command{
		Use:   "pull",
		Short: "Retrieve the passwords for a BATS system",
		Long:  `password is for decoding the passwords on a BATS system.`,
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			for _, repo := range g.config.Repos {
				g.wg.Add(1)
				go func(repo2 RepoConfig) {
					// if repo.Autostash {
					// 	errStash := g.stash(repo2)
					// 	if errStash != nil {
					// 		g.Error.Fprintln(os.Stderr, repo.Name, "stash failed", errStash)
					// 	}
					// }

					err := g.pull(repo2)
					if err != nil {
						g.Error.Fprintln(os.Stderr, repo.Name, "pull failed", err)
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
