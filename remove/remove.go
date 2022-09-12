package remove

import (
	"github.com/seventv/helm-manager/manager/cli"
	"github.com/seventv/helm-manager/manager/types"
)

func Run(cfg types.Config) {
	switch cfg.Arguments.Mode {
	case cli.CommandModeRemoveChart:
		runRemoveChart(cfg)
	case cli.CommandModeRemoveEnv:
		runRemoveEnv(cfg)
	case cli.CommandModeRemoveRepo:
		runRemoveRepo(cfg)
	case cli.CommandModeRemoveSingle:
		runRemoveSingle(cfg)
	}
}
