package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"gitm/internal/gitm"

	"gitm/commands/clone"
	"gitm/commands/fetch"
	"gitm/commands/find"
	"gitm/commands/pull"
	"gitm/commands/repos"
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

func main() {
	var rootCmd = &cobra.Command{Use: "gitm"}
	var opts gitm.GitmOptions

	rootCmd.PersistentFlags().StringVarP(&opts.ConfigFile, "config", "c", "", "Specify config file")
	rootCmd.PersistentFlags().StringVarP(&opts.RepoName, "repo", "r", "", "Specify repo")
	rootCmd.PersistentFlags().StringVarP(&opts.GroupName, "group", "g", "", "Specify group")
	rootCmd.PersistentFlags().BoolVarP(&opts.DebugMode, "debug", "", false, "Enable debug mode")
	rootCmd.PersistentFlags().IntVarP(&opts.MaxWorkers, "max-workers", "", 0, "Specify max workers (default: number of CPUs)")

	gitmObj := gitm.New(opts)

	clone.New(gitmObj, &opts, rootCmd)
	fetch.New(gitmObj, &opts, rootCmd)
	status.New(gitmObj, &opts, rootCmd)
	pull.New(gitmObj, &opts, rootCmd)
	find.New(gitmObj, &opts, rootCmd)
	tag.New(gitmObj, &opts, rootCmd)
	repos.New(gitmObj, &opts, rootCmd)

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
