package main

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/spf13/cobra"
)

type Gitm struct {
	config *GitmConfig

	maxWorkers int
	configFile string
	repoName   string
	groupName  string
	debugMode  bool

	logMutex sync.Mutex
}

func (gitm *Gitm) Load(cmd *cobra.Command, args []string) error {
	var err error
	if gitm.configFile == "" {
		gitm.configFile, err = cfgFind()
		if err != nil {
			return fmt.Errorf("Failed to find config file: %v", err)
		}
	}
	gitm.config, err = loadConfig(gitm.configFile)
	if err != nil {
		return fmt.Errorf("Failed to load config: %v", err)
	}

	// TODO: How do I know if the user passed in "--", cobra seems to remove it from args so we can't check for it

	// TODO: If we want to pass the group name as the first argument, we need to somehow make cobra aware of this
	// if gitm.groupName == "" && len(args) > 0 && args[0] != "--" {
	// 	gitm.groupName = args[0]
	// 	args = args[1:]
	// }

	// Print cobra options
	if gitm.debugMode {
		fmt.Printf("configFile: %s\n", gitm.configFile)
		fmt.Printf("repoName: %s\n", gitm.repoName)
		fmt.Printf("groupName: %s\n", gitm.groupName)
		fmt.Printf("debugMode: %v\n", gitm.debugMode)
		fmt.Printf("maxWorkers: %d\n", gitm.maxWorkers)
		fmt.Printf("config: %#v\n", gitm.config)
		fmt.Printf("args: %#v\n", args)
	}

	if gitm.groupName == "" && gitm.config.RequireGroup != nil && *gitm.config.RequireGroup {
		return fmt.Errorf("Group name is required when require-group is set to true")
	}

	return nil
}

func (gitm *Gitm) cmdRepos(cmd *cobra.Command, args []string) error {
	for _, repo := range gitm.config.Repos {
		fmt.Println(repo.Name)
	}

	return nil
}

func (gitm *Gitm) cmdWorker(maxWorkers int, f func(context.Context, RepoConfig) error) error {
	ctx, cancel := GetOSContext()
	defer cancel()

	worker := NewWorker(maxWorkers, false)

	for i, _ := range gitm.config.Repos {
		ok := true
		if gitm.groupName != "" {
			ok = false
			for _, group := range gitm.config.Repos[i].Group {
				if group == gitm.groupName {
					ok = true
				}
			}
		}

		if gitm.repoName != "" {
			ok = false
			if gitm.config.Repos[i].Name == gitm.repoName {
				ok = true
			}
		}

		if !ok {
			continue
		}
		repoConfig := gitm.config.Repos[i]
		f(ctx, repoConfig)
	}

	return worker.Run(ctx)
}

func (gitm *Gitm) cmdStatus(cmd *cobra.Command, args []string) error {
	return gitm.cmdWorker(1, func(ctx context.Context, repoConfig RepoConfig) error {
		return gitm.cmdStatusRepo(ctx, repoConfig)
	})
}

func (gitm *Gitm) cmdStatusRepo(ctx context.Context, repoConfig RepoConfig) error {
	if _, err := os.Stat(repoConfig.Path); err == nil {
		res, err := gitm.RunCommandAndCaptureOutput(ctx, "git", []string{"-C", repoConfig.Path, "status"})
		if err != nil {
			return fmt.Errorf("Failed to get repo status: %v", err)
		}

		fmt.Printf("%s\n", res.Stdout)

	} else {
		return fmt.Errorf("Repo %s does not exist", gitm.repoName)
	}

	return nil
}

// commit will need a way to present the user with a list of files that have changed which means parsing the output of git status

func cmdVersion(cmd *cobra.Command, args []string) {
	fmt.Printf("gitm version 0.0.1\n")
}

func cmdHelp(cmd *cobra.Command, args []string) {
	cmd.Usage()
}

type GitmCommand interface {
	Init(gitm *Gitm, cmd *cobra.Command) error
}

func main() {
	var rootCmd = &cobra.Command{Use: "gitm"}
	var gitm Gitm

	rootCmd.PersistentFlags().StringVarP(&gitm.configFile, "config", "c", "", "Specify config file")
	rootCmd.PersistentFlags().StringVarP(&gitm.repoName, "repo", "r", "", "Specify repo")
	rootCmd.PersistentFlags().StringVarP(&gitm.groupName, "group", "g", "", "Specify group")
	rootCmd.PersistentFlags().BoolVarP(&gitm.debugMode, "debug", "", false, "Enable debug mode")
	rootCmd.PersistentFlags().IntVarP(&gitm.maxWorkers, "max-workers", "", 0, "Specify max workers (default: number of CPUs)")

	find := Find{}
	find.Init(&gitm, rootCmd)

	tag := Tag{}
	tag.Init(&gitm, rootCmd)

	rootCmd.AddCommand(
		&cobra.Command{
			Use:   "repos",
			Short: "List repos",
			Run: func(cmd *cobra.Command, args []string) {
				if err := gitm.Load(cmd, args); err != nil {
					fmt.Fprintf(os.Stderr, "%v\n", err)
					os.Exit(1)
				}

				if err := gitm.cmdRepos(cmd, args); err != nil {
					fmt.Fprintf(os.Stderr, "%v\n", err)
					os.Exit(1)
				}
			},
		},
		&cobra.Command{
			Use:   "clone",
			Short: "Clone repos",
			Run: func(cmd *cobra.Command, args []string) {
				if err := gitm.Load(cmd, args); err != nil {
					fmt.Fprintf(os.Stderr, "%v\n", err)
					os.Exit(1)
				}

				if err := gitm.cmdClone(cmd, args); err != nil {
					fmt.Fprintf(os.Stderr, "%v\n", err)
					os.Exit(1)
				}
			},
		},
		&cobra.Command{
			Use:   "pull",
			Short: "Pull repos",
			Run: func(cmd *cobra.Command, args []string) {
				if err := gitm.Load(cmd, args); err != nil {
					fmt.Fprintf(os.Stderr, "%v\n", err)
					os.Exit(1)
				}

				if err := gitm.cmdPull(cmd, args); err != nil {
					fmt.Fprintf(os.Stderr, "%v\n", err)
					os.Exit(1)
				}
			},
		},
		&cobra.Command{
			Use:   "fetch",
			Short: "Fetch repos",
			Run: func(cmd *cobra.Command, args []string) {
				if err := gitm.Load(cmd, args); err != nil {
					fmt.Fprintf(os.Stderr, "%v\n", err)
					os.Exit(1)
				}

				if err := gitm.cmdFetch(cmd, args); err != nil {
					fmt.Fprintf(os.Stderr, "%v\n", err)
					os.Exit(1)
				}
			},
		},
		&cobra.Command{
			Use:   "status",
			Short: "Get repo status",
			Run: func(cmd *cobra.Command, args []string) {
				if err := gitm.Load(cmd, args); err != nil {
					fmt.Fprintf(os.Stderr, "%v\n", err)
					os.Exit(1)
				}

				if err := gitm.cmdStatus(cmd, args); err != nil {
					fmt.Fprintf(os.Stderr, "%v\n", err)
					os.Exit(1)
				}
			},
		},
		&cobra.Command{
			Use:   "version",
			Short: "Show version",
			Run:   cmdVersion,
		},
		&cobra.Command{
			Use:   "help",
			Short: "Show help",
			Run:   cmdHelp,
		},
	)

	// TODO: sync
	// TODO: checkout
	// TODO: branch
	// TODO: commit
	// TODO: push (with force, and with tags)
	// TODO: find (find branch, find tag, find commit, find file, find tag that matches regex
	// TODO: Multiple origins and call them by name
	// TODO: Auto stash changes before pull
	// TODO: Force reset option
	// TODO: Submodule commands (update, init, etc)
	// TODO: define config options on command line for logging, etc

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
