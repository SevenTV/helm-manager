package cli

import (
	"errors"
	"path"

	"github.com/seventv/helm-manager/argparse"
)

type Upgrade struct {
	DryRun           bool
	Wait             bool
	StopOnFirstError bool

	IgnoreMap    map[string]bool
	WhitelistMap map[string]bool

	Charts  UpgradeCharts
	Singles UpgradeSingles
}

var UpgradeCommand = Command{
	Name: "upgrade",
	Help: "Upgrade cluster and charts",
	Mode: CommandModeUpgrade,
}

func UpgradeCli(parser argparse.Command, args Arguments) Trigger {
	upgradeCmd := parser.NewCommand(UpgradeCommand.Name, UpgradeCommand.Help)

	upgradeDryRunFlag := upgradeCmd.Flag("", "dry-run", &argparse.Options[bool]{
		Required: false,
		Help:     "Dry run the upgrade",
	})

	upgradeIgnoreFlag := upgradeCmd.StringList("", "ignore", &argparse.OptionsList[string]{
		Required: false,
		Help:     "The charts to ignore",
	})

	upgradeOnlyFlag := upgradeCmd.StringList("", "only", &argparse.OptionsList[string]{
		Required: false,
		Help:     "A whitelist of things to upgrade",
	})

	upgradeWaitFlag := upgradeCmd.Flag("", "wait", &argparse.Options[bool]{
		Required: false,
		Help:     "Wait for the upgrade to complete",
	})

	upgradeNoStopOnFirstErrorFlag := upgradeCmd.Flag("", "no-stop", &argparse.Options[bool]{
		Required: false,
		Help:     "Disable stopping on the first error",
	})

	triggers := []Trigger{
		UpgradeChartsCli(upgradeCmd, args),
		UpgradeSinglesCli(upgradeCmd, args),
	}

	return func(args *Arguments) error {
		if !upgradeCmd.Happened() {
			return nil
		}

		args.Mode = CommandModeUpgrade
		args.Upgrade.DryRun = *upgradeDryRunFlag
		args.Upgrade.Wait = *upgradeWaitFlag
		args.Upgrade.StopOnFirstError = !*upgradeNoStopOnFirstErrorFlag

		args.Upgrade.IgnoreMap = map[string]bool{}
		args.Upgrade.WhitelistMap = map[string]bool{}

		for _, name := range *upgradeIgnoreFlag {
			args.Upgrade.IgnoreMap[name] = true
		}

		if len(*upgradeOnlyFlag) > 0 && len(args.Upgrade.IgnoreMap) > 0 {
			return errors.New("Invalid argument --only cannot be used with --ignore")
		}

		for _, name := range *upgradeOnlyFlag {
			args.Upgrade.WhitelistMap[name] = true
		}

		for _, trigger := range triggers {
			err := trigger(args)
			if err != nil {
				return err
			}
		}

		return nil
	}
}

var UpgradeChartsCommand = Command{
	Name: "charts",
	Help: "Upgrade charts",
	Mode: CommandModeUpgradeCharts,
}

type UpgradeCharts struct {
	GenerateTemplate  bool
	TemplateOutputDir string
	Atomic            bool
	Deploy            bool
}

func UpgradeChartsCli(parser argparse.Command, arguments Arguments) Trigger {
	upgradeChartsCmd := parser.NewCommand(UpgradeChartsCommand.Name, UpgradeChartsCommand.Help)

	upgradeGenerateTemplateFlag := upgradeChartsCmd.Flag("", "template", &argparse.Options[bool]{
		Required: false,
		Help:     "Generate a template file for the upgrade",
	})

	upgradeTemplateOutputDirFlag := upgradeChartsCmd.String("o", "template-output", &argparse.Options[string]{
		Required: false,
		Help:     "The directory to output the generated template files to",
		Default:  "templates",
	})

	upgradeAtomicFlag := upgradeChartsCmd.Flag("", "atomic", &argparse.Options[bool]{
		Required: false,
		Help:     "Rollback the upgrade if it fails",
	})

	upgradeDeployFlag := upgradeChartsCmd.Flag("", "deploy", &argparse.Options[bool]{
		Required: false,
		Help:     "Deploy the upgrade",
	})

	return func(args *Arguments) error {
		if !upgradeChartsCmd.Happened() {
			return nil
		}

		args.Mode = CommandModeUpgradeCharts

		args.Upgrade.Charts.GenerateTemplate = *upgradeGenerateTemplateFlag
		args.Upgrade.Charts.TemplateOutputDir = *upgradeTemplateOutputDirFlag
		args.Upgrade.Charts.Atomic = *upgradeAtomicFlag
		args.Upgrade.Charts.Deploy = *upgradeDeployFlag

		if !path.IsAbs(args.Upgrade.Charts.TemplateOutputDir) {
			args.Upgrade.Charts.TemplateOutputDir = path.Join(args.WorkingDir, args.Upgrade.Charts.TemplateOutputDir)
		}

		return nil
	}
}

var UpgradeSinglesCommand = Command{
	Name: "singles",
	Help: "Upgrade singles",
	Mode: CommandModeUpgradeSingles,
}

type UpgradeSingles struct {
	Deploy bool
}

func UpgradeSinglesCli(parser argparse.Command, arguments Arguments) Trigger {
	cmd := parser.NewCommand(UpgradeSinglesCommand.Name, UpgradeSinglesCommand.Help)

	deployFlag := cmd.Flag("", "deploy", &argparse.Options[bool]{
		Required: false,
		Help:     "Deploy the upgrade",
	})

	return func(args *Arguments) error {
		if !cmd.Happened() {
			return nil
		}

		args.Mode = CommandModeUpgradeSingles

		args.Upgrade.Singles.Deploy = *deployFlag

		return nil
	}
}
