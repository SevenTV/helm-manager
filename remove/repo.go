package remove

import (
	"github.com/seventv/helm-manager/manager"
	"github.com/seventv/helm-manager/manager/types"
	"go.uber.org/zap"
)

func runRemoveRepo(cfg types.Config) {
	repoName := cfg.Arguments.Remove.Repo.Name

	var (
		repo types.Repo
		idx  = -1
	)
	for idx, repo = range cfg.Repos {
		if repo.Name == repoName {
			break
		}
		idx = -1
	}

	if idx == -1 {
		zap.S().Fatalf("repo %s was not found", repoName)
	}

	cfg.Repos = append(cfg.Repos[:idx], cfg.Repos[idx+1:]...)

	zap.S().Infof("repo %s deleted", repo.Name)

	manager.WriteConfig(cfg)
}
