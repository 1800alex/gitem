package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/spf13/cobra"
)

type Gitm struct {
	config *GitmConfig

	configFile string
	repoName   string
	groupName  string
	debugMode  bool

	logMutex sync.Mutex
}

func (gitm *Gitm) Load() error {
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

	return nil
}

func (gitm *Gitm) cmdRepos(cmd *cobra.Command, args []string) {
	for _, repo := range gitm.config.Repos {
		fmt.Println(repo.Name)
	}
}

func (gitm *Gitm) cmdClone(cmd *cobra.Command, args []string) {
	// TODO: support fetching all repos in parallel

	for _, repo := range gitm.config.Repos {
		err := gitm.cmdCloneRepo(repo)
		if err != nil {
			// TODO
			log.Fatalf("%v", err)
		}
	}
}

func (gitm *Gitm) cmdCloneRepo(repoConfig RepoConfig) error {
	if _, err := os.Stat(repoConfig.Path); err == nil {
		if err := gitm.runCommandWithOutputFormatting("git", []string{"-C", repoConfig.Path, "pull"}); err != nil {
			return fmt.Errorf("Failed to pull repo: %v", err)
		}

	} else {
		if err := gitm.runCommandWithOutputFormatting("git", []string{"clone", repoConfig.URL, repoConfig.Path}); err != nil {
			return fmt.Errorf("Failed to clone repo: %v", err)
		}
	}

	if repoConfig.UsesLFS {
		if err := gitm.runCommandWithOutputFormatting("git", []string{"-C", repoConfig.Path, "lfs", "install"}); err != nil {
			return fmt.Errorf("Failed to install lfs repo: %v", err)
		}
	}

	if repoConfig.UsesSubrepos {
		if err := gitm.runCommandWithOutputFormatting("git", []string{"-C", repoConfig.Path, "submodule", "update", "--init", "--recursive"}); err != nil {
			return fmt.Errorf("Failed to initialize submodules: %v", err)
		}
	}

	return nil
}

func (gitm *Gitm) cmdFetch(cmd *cobra.Command, args []string) {
	worker := NewWorker(10, false)

	for i, _ := range gitm.config.Repos {
		repoConfig := gitm.config.Repos[i]
		worker.Add(func() error {
			return gitm.cmdFetchRepo(repoConfig)
		})
	}

	err := worker.Run(context.TODO())
	if err != nil {
		log.Fatalf("%v", err)
	}
}

func (gitm *Gitm) cmdFetchRepo(repoConfig RepoConfig) error {
	if _, err := os.Stat(repoConfig.Path); err == nil {
		err := gitm.runCommandWithOutputFormatting("git", []string{"-C", repoConfig.Path, "fetch"})
		if err != nil {
			return fmt.Errorf("Failed to fetch repo: %v", err)
		}
	} else {
		return fmt.Errorf("Repo %s does not exist", gitm.repoName)
	}

	return nil
}

func (gitm *Gitm) cmdStatus(cmd *cobra.Command, args []string) {
	// TODO: support fetching all repos in parallel

	for _, repo := range gitm.config.Repos {
		err := gitm.cmdStatusRepo(repo)
		if err != nil {
			// TODO
			log.Fatalf("%v", err)
		}
	}
}

func (gitm *Gitm) cmdStatusRepo(repoConfig RepoConfig) error {
	if _, err := os.Stat(repoConfig.Path); err == nil {
		err := gitm.runCommandWithOutputFormatting("git", []string{"-C", repoConfig.Path, "status"})
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

	rootCmd.AddCommand(
		&cobra.Command{
			Use:   "repos",
			Short: "List repos",
			Run: func(cmd *cobra.Command, args []string) {
				if err := gitm.Load(); err != nil {
					log.Fatalf("%v", err)
				}

				gitm.cmdRepos(cmd, args)
			},
		},
		&cobra.Command{
			Use:   "clone",
			Short: "Clone repos",
			Run: func(cmd *cobra.Command, args []string) {
				if err := gitm.Load(); err != nil {
					log.Fatalf("%v", err)
				}

				gitm.cmdClone(cmd, args)
			},
		},
		&cobra.Command{
			Use:   "fetch",
			Short: "Fetch repos",
			Run: func(cmd *cobra.Command, args []string) {
				if err := gitm.Load(); err != nil {
					log.Fatalf("%v", err)
				}

				gitm.cmdFetch(cmd, args)
			},
		},
		&cobra.Command{
			Use:   "status",
			Short: "Get repo status",
			Run: func(cmd *cobra.Command, args []string) {
				if err := gitm.Load(); err != nil {
					log.Fatalf("%v", err)
				}

				gitm.cmdStatus(cmd, args)
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
		fmt.Println(err)
		os.Exit(1)
	}
}
