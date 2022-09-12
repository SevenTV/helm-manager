package main

import (
	"github.com/seventv/helm-manager/add"
	"github.com/seventv/helm-manager/manager"
	"github.com/seventv/helm-manager/manager/cli"
	"github.com/seventv/helm-manager/remove"
	"github.com/seventv/helm-manager/upgrade"
	"go.uber.org/zap"
)

func main() {
	cfg := manager.GetConfig()

	switch cfg.Arguments.Mode {
	case cli.CommandModeInit:
		manager.WriteConfig(cfg)
	case cli.CommandModeUpgrade:
		zap.S().Info("* Helm Manager Upgrade *")
		if !cfg.Exists {
			zap.S().Fatalf("manifest.yaml not found, please run 'helm-manager init' first")
		}

		upgrade.Run(cfg)
	case cli.CommandModeAddChart, cli.CommandModeAddEnv, cli.CommandModeAddRepo, cli.CommandModeAddSingle:
		zap.S().Info("* Helm Manager Add *")
		if !cfg.Exists {
			zap.S().Fatalf("manifest.yaml not found, please run 'helm-manager init' first")
		}

		add.Run(cfg)
	case cli.CommandModeRemoveChart, cli.CommandModeRemoveEnv, cli.CommandModeRemoveRepo, cli.CommandModeRemoveSingle:
		zap.S().Info("* Helm Manager Remove *")
		if !cfg.Exists {
			zap.S().Fatalf("manifest.yaml not found, please run 'helm-manager init' first")
		}

		remove.Run(cfg)
	}
}
