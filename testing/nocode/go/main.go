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

	if gitm.groupName == "" && len(args) > 0 {
		gitm.groupName = args[0]
		args = args[1:]
	}

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

		if !ok {
			continue
		}
		repoConfig := gitm.config.Repos[i]
		f(ctx, repoConfig)
	}

	return worker.Run(ctx)
}

func (gitm *Gitm) cmdClone(cmd *cobra.Command, args []string) error {
	return gitm.cmdWorker(gitm.maxWorkers, func(ctx context.Context, repoConfig RepoConfig) error {
		return gitm.cmdCloneRepo(ctx, repoConfig)
	})
}

func (gitm *Gitm) cmdCloneRepo(ctx context.Context, repoConfig RepoConfig) error {
	if _, err := os.Stat(repoConfig.Path); err == nil {
		if err := gitm.runCommandWithOutputFormatting(ctx, "git", []string{"-C", repoConfig.Path, "pull"}); err != nil {
			return fmt.Errorf("Failed to pull repo: %v", err)
		}

	} else {
		if err := gitm.runCommandWithOutputFormatting(ctx, "git", []string{"clone", repoConfig.URL, repoConfig.Path}); err != nil {
			return fmt.Errorf("Failed to clone repo: %v", err)
		}
	}

	if repoConfig.UsesLFS {
		if err := gitm.runCommandWithOutputFormatting(ctx, "git", []string{"-C", repoConfig.Path, "lfs", "install"}); err != nil {
			return fmt.Errorf("Failed to install lfs repo: %v", err)
		}
	}

	if repoConfig.UsesSubrepos {
		if err := gitm.runCommandWithOutputFormatting(ctx, "git", []string{"-C", repoConfig.Path, "submodule", "update", "--init", "--recursive"}); err != nil {
			return fmt.Errorf("Failed to initialize submodules: %v", err)
		}
	}

	return nil
}

func (gitm *Gitm) cmdPull(cmd *cobra.Command, args []string) error {
	return gitm.cmdWorker(gitm.maxWorkers, func(ctx context.Context, repoConfig RepoConfig) error {
		return gitm.cmdPullRepo(ctx, repoConfig)
	})
}

func (gitm *Gitm) cmdPullRepo(ctx context.Context, repoConfig RepoConfig) error {
	if _, err := os.Stat(repoConfig.Path); err == nil {
		if err := gitm.runCommandWithOutputFormatting(ctx, "git", []string{"-C", repoConfig.Path, "pull"}); err != nil {
			return fmt.Errorf("Failed to pull repo: %v", err)
		}

		if repoConfig.UsesLFS {
			if err := gitm.runCommandWithOutputFormatting(ctx, "git", []string{"-C", repoConfig.Path, "lfs", "pull"}); err != nil {
				return fmt.Errorf("Failed to pull lfs repo: %v", err)
			}
		}

		if repoConfig.UsesSubrepos {
			if err := gitm.runCommandWithOutputFormatting(ctx, "git", []string{"-C", repoConfig.Path, "submodule", "update", "--init", "--recursive"}); err != nil {
				return fmt.Errorf("Failed to initialize submodules: %v", err)
			}
		}
	} else {
		return fmt.Errorf("Repo %s does not exist", gitm.repoName)
	}

	return nil
}

func (gitm *Gitm) cmdFetch(cmd *cobra.Command, args []string) error {
	return gitm.cmdWorker(gitm.maxWorkers, func(ctx context.Context, repoConfig RepoConfig) error {
		return gitm.cmdFetchRepo(ctx, repoConfig)
	})
}

func (gitm *Gitm) cmdFetchRepo(ctx context.Context, repoConfig RepoConfig) error {
	if _, err := os.Stat(repoConfig.Path); err == nil {
		err := gitm.runCommandWithOutputFormatting(ctx, "git", []string{"-C", repoConfig.Path, "fetch"})
		if err != nil {
			return fmt.Errorf("Failed to fetch repo: %v", err)
		}
	} else {
		return fmt.Errorf("Repo %s does not exist", gitm.repoName)
	}

	return nil
}

func (gitm *Gitm) cmdStatus(cmd *cobra.Command, args []string) error {
	return gitm.cmdWorker(1, func(ctx context.Context, repoConfig RepoConfig) error {
		return gitm.cmdStatusRepo(ctx, repoConfig)
	})
}

func (gitm *Gitm) cmdStatusRepo(ctx context.Context, repoConfig RepoConfig) error {
	if _, err := os.Stat(repoConfig.Path); err == nil {
		err := gitm.runCommandWithOutputFormatting(ctx, "git", []string{"-C", repoConfig.Path, "status"})
		if err != nil {
			return fmt.Errorf("Failed to get repo status: %v", err)
		}
	} else {
		return fmt.Errorf("Repo %s does not exist", gitm.repoName)
	}

	return nil
}

func cmdVersion(cmd *cobra.Command, args []string) {
	fmt.Printf("gitm version 0.0.1\n")
}

func cmdHelp(cmd *cobra.Command, args []string) {
	cmd.Usage()
}

func main() {
	var rootCmd = &cobra.Command{Use: "gitm"}
	var gitm Gitm

	rootCmd.PersistentFlags().StringVarP(&gitm.configFile, "config", "c", "", "Specify config file")
	rootCmd.PersistentFlags().StringVarP(&gitm.repoName, "repo", "r", "", "Specify repo")
	rootCmd.PersistentFlags().StringVarP(&gitm.groupName, "group", "g", "", "Specify group")
	rootCmd.PersistentFlags().BoolVarP(&gitm.debugMode, "debug", "", false, "Enable debug mode")
	rootCmd.PersistentFlags().IntVarP(&gitm.maxWorkers, "max-workers", "", 0, "Specify max workers (default: number of CPUs)")

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

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
