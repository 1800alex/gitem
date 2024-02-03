package gitm

import (
	"fmt"
	"sync"

	"github.com/spf13/cobra"
)

// Gitm is the main object for gitm
type Gitm struct {
	config *GitmConfig

	maxWorkers int
	configFile string
	repoName   string
	groupName  string
	debugMode  bool

	logMutex sync.Mutex
}

// New will create a new Gitm object
func New(opts GitmOptions) *Gitm {
	return &Gitm{
		configFile: opts.ConfigFile,
		repoName:   opts.RepoName,
		groupName:  opts.GroupName,
		debugMode:  opts.DebugMode,
		maxWorkers: opts.MaxWorkers,
	}
}

// SetOptions will set the options
func (gitm *Gitm) SetOptions(opts GitmOptions) {
	gitm.configFile = opts.ConfigFile
	gitm.repoName = opts.RepoName
	gitm.groupName = opts.GroupName
	gitm.debugMode = opts.DebugMode
	gitm.maxWorkers = opts.MaxWorkers
}

// Load will load the config file and set the config
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

// Init will initialize the Gitm object
func (gitm *Gitm) Init(opts *GitmOptions, cmd *cobra.Command, args []string) error {
	if opts != nil {
		gitm.SetOptions(*opts)
	}
	return gitm.Load(cmd, args)
}

// MaxWorkers will return the number of workers to use
func (gitm *Gitm) MaxWorkers() int {
	if gitm.maxWorkers < 1 {
		return 1
	}

	return gitm.maxWorkers
}

// GitmCommand is an interface for gitm commands
type GitmCommand interface {
	Init(gitm *Gitm, opts *GitmOptions, cmd *cobra.Command) error
}
