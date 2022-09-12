package add

import (
	"github.com/seventv/helm-manager/manager/cli"
	"github.com/seventv/helm-manager/manager/types"
)

func Run(cfg types.Config) {
	switch cfg.Arguments.Mode {
	case cli.CommandModeAddChart:
		runAddChart(cfg)
	case cli.CommandModeAddEnv:
		runAddEnv(cfg)
	case cli.CommandModeAddRepo:
		runAddRepo(cfg)
	case cli.CommandModeAddSingle:
		runAddSingle(cfg)
	}
}
