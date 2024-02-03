package gitm

import (
	"context"
	"gitm/worker"
)

// NewWorker will create a new worker and run the function f
// The function f will be called for each repo
// If maxWorkers is less than 1, the number of workers will be set to gitm.MaxWorkers()
func (gitm *Gitm) NewWorker(maxWorkers int, f func(context.Context, RepoConfig) error) error {
	ctx, cancel := GetOSContext()
	defer cancel()

	if maxWorkers < 1 {
		maxWorkers = gitm.MaxWorkers()
	}

	w := worker.NewWorker(maxWorkers, false)

	for i, _ := range gitm.Config.Repos {
		ok := true
		if gitm.groupName != "" {
			ok = false
			for _, group := range gitm.Config.Repos[i].Group {
				if group == gitm.groupName {
					ok = true
				}
			}
		}

		if gitm.repoName != "" {
			ok = false
			if gitm.Config.Repos[i].Name == gitm.repoName {
				ok = true
			}
		}

		if !ok {
			continue
		}
		repoConfig := gitm.Config.Repos[i]
		f(ctx, repoConfig)
	}

	return w.Run(ctx)
}
