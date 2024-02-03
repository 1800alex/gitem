package main

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

type FindCmd struct {
	gitm *Gitm
	root *cobra.Command

	opts FindOpts
}

type FindOpts struct {
	Search *regexp.Regexp
	Oldest bool
}

func (c *FindCmd) cmdFindTag(cmd *cobra.Command, args []string) error {
	return c.gitm.NewWorker(0, func(ctx context.Context, repoConfig RepoConfig) error {
		return c.cmdFindTagRepo(ctx, repoConfig)
	})
}

// TODO make this output a struct instead of writing to stdout

func (c *FindCmd) cmdFindTagRepo(ctx context.Context, repoConfig RepoConfig) error {
	if _, err := os.Stat(repoConfig.Path); err == nil {
		// TODO which one to use?
		// -sort=committerdate
		// -sort=creatordate

		cmdArgs := []string{"-C", repoConfig.Path, "for-each-ref"}

		cmdArgs = append(cmdArgs, "--sort=creatordate")
		cmdArgs = append(cmdArgs, "--format=%(refname) %(creatordate)")
		cmdArgs = append(cmdArgs, "refs/tags")

		res, err := c.gitm.RunCommandAndCaptureOutput(ctx, "git", cmdArgs)
		if err != nil {
			return fmt.Errorf("Failed to search repo: %v", err)
		}

		found := false
		var tag string
		var date time.Time

		for _, line := range res.Stdout {
			if c.opts.Search.MatchString(line) {
				// parse line 'refs/tags/v0.0.1 Sun Jan 28 11:10:14 2024 +0000'
				tag = strings.TrimPrefix(strings.Split(line, " ")[0], "refs/tags/")
				dateStr := strings.Join(strings.Split(line, " ")[1:], " ")

				date, err = time.Parse("Mon Jan 2 15:04:05 2006 -0700", dateStr)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Failed to parse date: %v\n", err)
					continue
				}

				found = true

				if c.opts.Oldest {
					// Oldest tag is first in list
					break
				}
			}
		}

		if !found {
			return fmt.Errorf("Failed to find tag matching regex %s", c.opts.Search.String())
		}
		fmt.Printf("%s: %v: %s\n", repoConfig.Name, date, tag)

	} else {
		return fmt.Errorf("Repo %s does not exist", c.gitm.repoName)
	}

	return nil
}

func (c *FindCmd) Init(gitm *Gitm, cmd *cobra.Command) error {
	c.gitm = gitm
	c.root = cmd

	findCmd := cobra.Command{
		Use:   "find",
		Short: "Find related commands",
	}

	// Add sub command to find a tag that matches a regex
	findTagRegexString := ""
	findTagCmd := cobra.Command{
		Use:   "tag",
		Short: "Find tags",
		Run: func(cmd *cobra.Command, args []string) {
			if findTagRegexString == "" && len(args) > 0 {
				findTagRegexString = args[0]
				args = args[1:]
			}

			if findTagRegexString == "" {
				fmt.Fprintf(os.Stderr, "%v\n", fmt.Errorf("Must specify a regex to search for"))
				os.Exit(1)
			}

			searchRe, err := regexp.Compile(findTagRegexString)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to compile regex: %v\n", err)
				os.Exit(1)
			}

			c.opts.Search = searchRe

			if err := c.gitm.Load(cmd, args); err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(1)
			}

			if err := c.cmdFindTag(cmd, args); err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(1)
			}
		},
	}

	findTagCmd.Flags().StringVarP(&findTagRegexString, "find-tag", "", "", "Specify tag to search for")
	findTagCmd.Flags().BoolVarP(&c.opts.Oldest, "find-oldest", "", false, "Find oldest tag (default: newest)")

	findCmd.AddCommand(&findTagCmd)

	c.root.AddCommand(&findCmd)

	return nil
}
