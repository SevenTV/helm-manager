package remove

import (
	"github.com/fatih/color"
	"github.com/seventv/helm-manager/manager/cli"
	"github.com/seventv/helm-manager/manager/types"
	"github.com/seventv/helm-manager/manager/utils"
	"go.uber.org/zap"
)

var RemoveColor = color.New(color.Bold, color.FgRed)

func Run(cfg types.Config) {
	if cfg.Arguments.Mode == cli.CommandModeRemove {
		if cfg.Arguments.InTerminal {
			cmd := utils.SelectCommand("Select a subcommand", []cli.Command{
				cli.RemoveChartCommand,
				cli.RemoveEnvCommand,
				cli.RemoveRepoCommand,
				cli.RemoveSingleCommand,
			})

			cfg.Arguments.Mode = cmd.Mode
		} else {
			zap.S().Infof("* %s *", RemoveColor.Sprint("Helm Manager Remove"))
			zap.S().Fatal(cli.Parser.Usage(color.RedString("Non-interactive mode requires a subcommand")))
		}
	}

	switch cfg.Arguments.Mode {
	case cli.CommandModeRemoveChart:
		zap.S().Infof("* %s *", RemoveColor.Sprint("Helm Manager Remove Chart"))
		runRemoveChart(cfg)
	case cli.CommandModeRemoveEnv:
		zap.S().Infof("* %s *", RemoveColor.Sprint("Helm Manager Remove Env"))
		runRemoveEnv(cfg)
	case cli.CommandModeRemoveRepo:
		zap.S().Infof("* %s *", RemoveColor.Sprint("Helm Manager Remove Repo"))
		runRemoveRepo(cfg)
	case cli.CommandModeRemoveSingle:
		zap.S().Infof("* %s *", RemoveColor.Sprint("Helm Manager Remove Single"))
		runRemoveSingle(cfg)
	}
}
