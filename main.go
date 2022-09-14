package main

import (
	"github.com/fatih/color"
	"github.com/seventv/helm-manager/add"
	i "github.com/seventv/helm-manager/init"
	"github.com/seventv/helm-manager/manager"
	"github.com/seventv/helm-manager/manager/cli"
	"github.com/seventv/helm-manager/manager/utils"
	"github.com/seventv/helm-manager/remove"
	"github.com/seventv/helm-manager/update"
	"github.com/seventv/helm-manager/upgrade"
	"go.uber.org/zap"
)

func main() {
	cfg := manager.GetConfig()

	if cfg.Arguments.Mode == cli.CommandModeBase {
		zap.S().Infof("* %s *\r", color.New(color.Bold, color.FgCyan).Sprint("Helm Manager"))
		if cfg.Arguments.InTerminal {
			cmd := utils.SelectCommand("Select a command", []cli.Command{
				cli.AddCommand,
				cli.RemoveCommand,
				cli.UpgradeCommand,
				cli.InitCommand,
				cli.UpdateCommand,
			})

			cfg.Arguments.Mode = cmd.Mode
		} else {
			zap.S().Fatal(cli.Parser.Usage("Non-interactive mode requires a command"))
		}
	}

	switch cfg.Arguments.Mode {
	case cli.CommandModeInit:
		zap.S().Infof("* %s *", color.New(color.Bold, color.FgBlue).Sprint("Helm Manager Init"))
		if cfg.Exists {
			utils.Fatal("manifest.yaml alredy exists, cannot re-initialize")
		}

		i.Run(cfg)
	case cli.CommandModeUpgrade:
		zap.S().Infof("* %s *", color.New(color.Bold, color.FgMagenta).Sprint("Helm Manager Upgrade"))
		if !cfg.Exists {
			utils.Fatal("manifest.yaml not found, please run '%s' first", color.YellowString("%s init", cli.BaseCommand.Name))
		}

		upgrade.Run(cfg)
	case cli.CommandModeAdd, cli.CommandModeAddChart, cli.CommandModeAddEnv, cli.CommandModeAddRepo, cli.CommandModeAddSingle:
		if !cfg.Exists {
			zap.S().Infof("* %s *", add.AddColor.Sprint("Helm Manager Add"))
			utils.Fatal("manifest.yaml not found, please run '%s' first", color.YellowString("%s init", cli.BaseCommand.Name))
		} else if cfg.Arguments.InTerminal {
			zap.S().Infof("* %s *\r", add.AddColor.Sprint("Helm Manager Add"))
		}

		add.Run(cfg)
	case cli.CommandModeRemove, cli.CommandModeRemoveChart, cli.CommandModeRemoveEnv, cli.CommandModeRemoveRepo, cli.CommandModeRemoveSingle:
		if !cfg.Exists {
			zap.S().Infof("* %s *", remove.RemoveColor.Sprint("Helm Manager Remove"))
			utils.Fatal("manifest.yaml not found, please run '%s' first", color.YellowString("%s init", cli.BaseCommand.Name))
		} else if cfg.Arguments.InTerminal {
			zap.S().Infof("* %s *\r", remove.RemoveColor.Sprint("Helm Manager Remove"))
		}

		remove.Run(cfg)
	case cli.CommandModeUpdate:
		zap.S().Infof("* %s *", update.UpdateColor.Sprint("Helm Manager Update"))
		if !cfg.Exists {
			utils.Fatal("manifest.yaml not found, please run '%s' first", color.YellowString("%s init", cli.BaseCommand.Name))
		}

		update.Run(cfg)
	}
}
