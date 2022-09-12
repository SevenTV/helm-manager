package add

import (
	"github.com/seventv/helm-manager/manager"
	"github.com/seventv/helm-manager/manager/types"
	"go.uber.org/zap"
)

func runAddSingle(cfg types.Config) {
	single := types.Single{
		Name:      cfg.Arguments.Add.Single.Name,
		Namespace: cfg.Arguments.Add.Single.Namespace,
		File:      cfg.Arguments.Add.Single.File,
		UseCreate: cfg.Arguments.Add.Single.UseCreate,
	}

	for _, c := range cfg.Singles {
		if c.Name == single.Name {
			zap.S().Fatalf("single with name %s already exists", single.Name)
		}
	}

	cfg.Singles = append(cfg.Singles, single)

	manager.WriteConfig(cfg)
	zap.S().Infof("added single %s, run `helm-manager upgrade`", single.Name)
}
