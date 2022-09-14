package upgrade

import (
	"github.com/fatih/color"
	"github.com/seventv/helm-manager/manager/cli"
	"github.com/seventv/helm-manager/manager/types"
	"github.com/seventv/helm-manager/manager/utils"
	"go.uber.org/zap"
)

var UpgradeColor = color.New(color.Bold, color.FgMagenta)

func Run(cfg types.Config) {
	if cfg.Arguments.Mode == cli.CommandModeUpgrade {
		if !cfg.Arguments.NonInteractive {
			cmd := utils.SelectCommand("Select a subcommand", []cli.Command{
				cli.UpgradeChartsCommand,
				cli.UpgradeSinglesCommand,
			})

			cfg.Arguments.Mode = cmd.Mode
		} else {
			zap.S().Infof("* %s *", UpgradeColor.Sprint("Helm Manager Upgrade"))
			zap.S().Fatal(cli.Parser.Usage(color.RedString("Non-interactive mode requires a subcommand")))
		}
	}

	switch cfg.Arguments.Mode {
	case cli.CommandModeUpgradeCharts:
		zap.S().Infof("* %s *", UpgradeColor.Sprint("Helm Manager Upgrade Charts"))
		runUpgradeCharts(cfg)
	case cli.CommandModeUpgradeSingles:
		zap.S().Infof("* %s *", UpgradeColor.Sprint("Helm Manager Upgrade Singles"))
		runUpgradeSingles(cfg)
	}
}
