package repos

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"gitm/internal/gitm"
)

type Cmd struct {
	gitm *gitm.Gitm
	root *cobra.Command
}

// New creates a new gitm command
func New(gitm *gitm.Gitm, opts *gitm.GitmOptions, root *cobra.Command) *Cmd {
	c := Cmd{}

	c.gitm = gitm
	c.root = root

	cloneCmd := cobra.Command{
		Use:   "repos",
		Short: "List repos",
		Run: func(cmd *cobra.Command, args []string) {
			if err := c.gitm.Init(opts, cmd, args); err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(1)
			}

			for _, repo := range c.gitm.Config.Repos {
				fmt.Println(repo.Name)
			}
		},
	}

	c.root.AddCommand(&cloneCmd)

	return &c
}
