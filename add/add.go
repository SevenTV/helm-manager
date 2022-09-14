package add

import (
	"github.com/fatih/color"
	"github.com/seventv/helm-manager/manager/cli"
	"github.com/seventv/helm-manager/manager/types"
	"github.com/seventv/helm-manager/manager/utils"
	"go.uber.org/zap"
)

var AddColor = color.New(color.Bold, color.FgGreen)

func Run(cfg types.Config) {
	if cfg.Arguments.Mode == cli.CommandModeAdd {
		if cfg.Arguments.InTerminal {
			cmd := utils.SelectCommand("Select a subcommand", []cli.Command{
				cli.AddChartCommand,
				cli.AddEnvCommand,
				cli.AddRepoCommand,
				cli.AddSingleCommand,
			})

			cfg.Arguments.Mode = cmd.Mode
		} else {
			zap.S().Infof("* %s *", AddColor.Sprint("Helm Manager Add"))
			zap.S().Fatal(cli.Parser.Usage(color.RedString("Non-interactive mode requires a subcommand")))
		}
	}

	switch cfg.Arguments.Mode {
	case cli.CommandModeAddChart:
		zap.S().Infof("* %s *", AddColor.Sprint("Helm Manager Add Chart"))
		runAddChart(cfg)
	case cli.CommandModeAddEnv:
		zap.S().Infof("* %s *", AddColor.Sprint("Helm Manager Add Env"))
		runAddEnv(cfg)
	case cli.CommandModeAddRepo:
		zap.S().Infof("* %s *", AddColor.Sprint("Helm Manager Add Repo"))
		runAddRepo(cfg)
	case cli.CommandModeAddSingle:
		zap.S().Infof("* %s *", AddColor.Sprint("Helm Manager Add Single"))
		runAddSingle(cfg)
	}
}
