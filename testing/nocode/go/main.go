package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"gitm/internal/gitm"

	"gitm/commands/clone"
	"gitm/commands/fetch"
	"gitm/commands/find"
	"gitm/commands/foreach"
	"gitm/commands/pull"
	"gitm/commands/repos"
	"gitm/commands/run"
	"gitm/commands/status"
	"gitm/commands/tag"
)

// commit will need a way to present the user with a list of files that have changed which means parsing the output of git status

func cmdVersion(cmd *cobra.Command, args []string) {
	fmt.Printf("gitm version 0.0.1\n")
}

func cmdHelp(cmd *cobra.Command, args []string) {
	cmd.Usage()
}

func createSubcommands(rootCmd *cobra.Command) {
	// Create subcommands here based on AppConfig or any other dynamic logic
	subCmd1 := &cobra.Command{
		Use:   "subcommand1",
		Short: "Subcommand 1",
		Run: func(cmd *cobra.Command, args []string) {
			// Subcommand 1 logic using AppConfig
			fmt.Println("Subcommand 1 with args: ", args)
		},
	}

	subCmd2 := &cobra.Command{
		Use:   "subcommand2",
		Short: "Subcommand 2",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Subcommand 2 with args: ", args)
		},
	}

	// Add subcommands to the root command
	rootCmd.AddCommand(subCmd1, subCmd2)
}

func main() {
	var opts gitm.GitmOptions
	gitmObj := gitm.New(opts)

	var rootCmd *cobra.Command
	rootCmd = &cobra.Command{
		Use: "gitm",
		// PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// 	if err := gitmObj.Init(&opts, cmd, args); err != nil {
		// 		fmt.Fprintf(os.Stderr, "%v\n", err)
		// 		os.Exit(1)
		// 	}
		// },
	}

	// Check if GITM_CONFIG is set and use it as the config file
	if configFile, ok := os.LookupEnv("GITM_CONFIG"); ok {
		opts.ConfigFile = configFile
	}

	// Check if GITM_DEBUG is set and use it as the debug mode
	if debugMode, ok := os.LookupEnv("GITM_DEBUG"); ok {
		opts.DebugMode = debugMode == "true"
	}

	// Check if GITM_MAX_WORKERS is set and use it as the max workers
	if maxWorkers, ok := os.LookupEnv("GITM_MAX_WORKERS"); ok {
		// Parse the max workers as an int
		maxWorkersInt, err := strconv.Atoi(maxWorkers)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to parse GITM_MAX_WORKERS: %v\n", err)
			os.Exit(1)
		}

		opts.MaxWorkers = maxWorkersInt
	}

	// Check if GITM_REPO is set and use it as the repo name
	if repoName, ok := os.LookupEnv("GITM_REPO"); ok {
		opts.RepoName = repoName
	}

	// Check if GITM_GROUP is set and use it as the group name
	if groupName, ok := os.LookupEnv("GITM_GROUP"); ok {
		opts.GroupName = groupName
	}

	// Go ahead and load the config file and set the config
	if err := gitmObj.Init(&opts, rootCmd, []string{}); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	rootCmd.PersistentFlags().StringVarP(&opts.ConfigFile, "config", "c", "", "Specify config file")
	rootCmd.PersistentFlags().StringVarP(&opts.RepoName, "repo", "r", "", "Specify repo")
	rootCmd.PersistentFlags().StringVarP(&opts.GroupName, "group", "g", "", "Specify group")
	rootCmd.PersistentFlags().BoolVarP(&opts.DebugMode, "debug", "", false, "Enable debug mode")
	rootCmd.PersistentFlags().IntVarP(&opts.MaxWorkers, "max-workers", "", 0, "Specify max workers (default: number of CPUs)")

	clone.New(gitmObj, &opts, rootCmd)
	fetch.New(gitmObj, &opts, rootCmd)
	status.New(gitmObj, &opts, rootCmd)
	pull.New(gitmObj, &opts, rootCmd)
	find.New(gitmObj, &opts, rootCmd)
	tag.New(gitmObj, &opts, rootCmd)
	repos.New(gitmObj, &opts, rootCmd)
	foreach.New(gitmObj, &opts, rootCmd)
	run.New(gitmObj, &opts, rootCmd)

	rootCmd.AddCommand(
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
