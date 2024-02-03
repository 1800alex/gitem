package gitm

type GitmOptions struct {
	ConfigFile string
	RepoName   string
	GroupName  string
	DebugMode  bool
	MaxWorkers int
}
