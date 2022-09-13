package cli

import (
	"path"

	"github.com/seventv/helm-manager/argparse"
	"go.uber.org/zap"
)

type Upgrade struct {
	DryRun            bool
	GenerateTemplate  bool
	Wait              bool
	Atomic            bool
	StopOnFirstError  bool
	TemplateOutputDir string
	IgnoreChartsMap   map[string]bool
	ChartWhitelist    map[string]bool
}

func UpgradeCli(parser argparse.Parser) Trigger {
	cmd := parser.NewCommand("upgrade", "The upgrade subcommand is used to update the values files or the cluster")

	upgradeDryRunFlag := cmd.Flag("", "dry-run", &argparse.Options[bool]{
		Required: false,
		Help:     "Dry run the upgrade",
	})

	upgradeGenerateTemplateFlag := cmd.Flag("", "template", &argparse.Options[bool]{
		Required: false,
		Help:     "Generate a template file for the upgrade",
	})

	upgradeTemplateOutputDirFlag := cmd.String("o", "template-output", &argparse.Options[string]{
		Required: false,
		Help:     "The directory to output the generated template files to",
		Default:  "templates",
	})

	upgradeIgnoreChartsFlag := cmd.StringList("", "ignore", &argparse.OptionsList[string]{
		Required: false,
		Help:     "The charts to ignore",
	})

	upgradeOnlyChartsFlag := cmd.StringList("", "only", &argparse.OptionsList[string]{
		Required: false,
		Help:     "A whitelist of charts to upgrade",
	})

	upgradeWaitFlag := cmd.Flag("", "wait", &argparse.Options[bool]{
		Required: false,
		Help:     "Wait for the upgrade to complete",
	})

	upgradeAtomicFlag := cmd.Flag("", "atomic", &argparse.Options[bool]{
		Required: false,
		Help:     "Rollback the upgrade if it fails",
	})

	upgradeNoStopOnFirstErrorFlag := cmd.Flag("", "no-stop", &argparse.Options[bool]{
		Required: false,
		Help:     "Disable stopping on the first error",
	})

	return func(args *Arguments) {
		if !cmd.Happened() {
			return
		}

		args.Mode = CommandModeUpgrade
		args.Upgrade.DryRun = *upgradeDryRunFlag
		args.Upgrade.GenerateTemplate = *upgradeGenerateTemplateFlag
		args.Upgrade.TemplateOutputDir = *upgradeTemplateOutputDirFlag
		args.Upgrade.Wait = *upgradeWaitFlag
		args.Upgrade.Atomic = *upgradeAtomicFlag
		args.Upgrade.StopOnFirstError = !*upgradeNoStopOnFirstErrorFlag

		args.Upgrade.IgnoreChartsMap = map[string]bool{}
		args.Upgrade.ChartWhitelist = map[string]bool{}

		for _, name := range *upgradeIgnoreChartsFlag {
			args.Upgrade.IgnoreChartsMap[name] = true
		}

		if len(*upgradeOnlyChartsFlag) > 0 && len(args.Upgrade.IgnoreChartsMap) > 0 {
			zap.S().Fatalf("Invalid argument --only cannot be used with --ignore")
		}

		for _, name := range *upgradeOnlyChartsFlag {
			args.Upgrade.ChartWhitelist[name] = true
		}

		if !path.IsAbs(args.Upgrade.TemplateOutputDir) {
			args.Upgrade.TemplateOutputDir = path.Join(args.WorkingDir, args.Upgrade.TemplateOutputDir)
		}
	}
}
