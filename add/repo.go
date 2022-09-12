package add

import (
	"github.com/seventv/helm-manager/manager"
	"github.com/seventv/helm-manager/manager/types"
	"go.uber.org/zap"
)

func runAddRepo(cfg types.Config) {
	repo := types.Repo{
		Name: cfg.Arguments.Add.Repo.Name,
		URL:  cfg.Arguments.Add.Repo.URL,
	}

	for _, r := range cfg.Repos {
		if r.Name == repo.Name {
			zap.S().Fatalf("repo with name %s already exists", repo.Name)
		}
	}

	cfg.Repos = append(cfg.Repos, repo)

	zap.S().Infof("adding repo %s", repo.Name)

	manager.UpdateRepos(cfg)

	zap.S().Infof("added repo %s", repo.Name)

	manager.WriteConfig(cfg)
}
