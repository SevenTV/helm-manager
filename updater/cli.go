package updater

import (
	"fmt"
	"os"
	"path"

	"github.com/akamensky/argparse"
	"go.uber.org/zap"
)

type CommandMode int

const (
	CommandModeUpdate CommandMode = 1 << iota
)

func (c CommandMode) Update() bool {
	return c&CommandModeUpdate != 0
}

type CommandArgs struct {
	WorkingDir   string
	ManifestFile string
	ValuesDir    string
	Debug        bool
	Mode         CommandMode
	UpdateArgs   CommandArgsUpdate
}

type CommandArgsUpdate struct {
	DryRun            bool
	GenerateTemplate  bool
	Wait              bool
	Atomic            bool
	StopOnFirstError  bool
	TemplateOutputDir string
	IgnoreChartsMap   map[string]bool
	ForceCharts       map[string]bool
}

func ReadArguments() CommandArgs {
	var args CommandArgs

	parser := argparse.NewParser(path.Base(os.Args[0]), "Manage Helm Charts and their values")

	debugFlag := parser.Flag("", "debug", &argparse.Options{
		Required: false,
		Help:     "Enable debug logging",
	})
	workingDirFlag := parser.String("d", "working-dir", &argparse.Options{
		Required: false,
		Help:     "The working directory to use",
		Default:  ".",
	})
	manifestFileFlag := parser.String("m", "manifest-file", &argparse.Options{
		Required: false,
		Help:     "The manifest file to use",
		Default:  "manifest.yaml",
	})
	valuesDirFlag := parser.String("v", "values-dir", &argparse.Options{
		Required: false,
		Help:     "The values directory to use",
		Default:  "values",
	})

	updateCommand := parser.NewCommand("update", "The update subcommand is used to update the values files or the cluster")

	updateDryRunFlag := updateCommand.Flag("", "dry-run", &argparse.Options{
		Required: false,
		Help:     "Dry run the upgrade",
	})

	updateGenerateTemplateFlag := updateCommand.Flag("t", "generate-template", &argparse.Options{
		Required: false,
		Help:     "Generate a template file for the upgrade",
	})

	updateTemplateOutputDirFlag := updateCommand.String("o", "template-output-dir", &argparse.Options{
		Required: false,
		Help:     "The directory to output the generated template files to",
		Default:  "templates",
	})

	updateIgnoreChartsFlag := updateCommand.List("i", "ignore-charts", &argparse.Options{
		Required: false,
		Help:     "The charts to ignore",
	})

	updateForceChartsFlag := updateCommand.List("f", "force-charts", &argparse.Options{
		Required: false,
		Help:     "The charts to force upgrade",
	})

	updateWaitFlag := updateCommand.Flag("w", "wait", &argparse.Options{
		Required: false,
		Help:     "Wait for the upgrade to complete",
	})

	updateAtomicFlag := updateCommand.Flag("a", "atomic", &argparse.Options{
		Required: false,
		Help:     "Rollback the upgrade if it fails",
	})

	updateNoStopOnFirstErrorFlag := updateCommand.Flag("", "no-stop", &argparse.Options{
		Required: false,
		Help:     "Disable stopping on the first error",
	})

	if err := parser.Parse(os.Args); err != nil {
		parser.Usage(err)
	}

	args.Debug = *debugFlag
	args.WorkingDir = *workingDirFlag
	args.ManifestFile = *manifestFileFlag
	args.ValuesDir = *valuesDirFlag

	if !path.IsAbs(args.ManifestFile) {
		args.ManifestFile = path.Join(args.WorkingDir, args.ManifestFile)
	}
	if !path.IsAbs(args.ValuesDir) {
		args.ValuesDir = path.Join(args.WorkingDir, args.ValuesDir)
	}

	if updateCommand.Happened() {
		args.Mode |= CommandModeUpdate
		args.UpdateArgs.DryRun = *updateDryRunFlag
		args.UpdateArgs.GenerateTemplate = *updateGenerateTemplateFlag
		args.UpdateArgs.TemplateOutputDir = *updateTemplateOutputDirFlag
		args.UpdateArgs.Wait = *updateWaitFlag
		args.UpdateArgs.Atomic = *updateAtomicFlag
		args.UpdateArgs.StopOnFirstError = !*updateNoStopOnFirstErrorFlag

		args.UpdateArgs.ForceCharts = map[string]bool{}
		args.UpdateArgs.IgnoreChartsMap = map[string]bool{}

		for _, name := range *updateIgnoreChartsFlag {
			args.UpdateArgs.IgnoreChartsMap[name] = true
		}
		for _, name := range *updateForceChartsFlag {
			if args.UpdateArgs.IgnoreChartsMap[name] {
				zap.S().Fatalf("Invalid argument chart %s is both ignored and forced", name)
			}

			args.UpdateArgs.ForceCharts[name] = true
		}

		if !path.IsAbs(args.UpdateArgs.TemplateOutputDir) {
			args.UpdateArgs.TemplateOutputDir = path.Join(args.WorkingDir, args.UpdateArgs.TemplateOutputDir)
		}
	}

	if args.Mode == 0 {
		fmt.Println(parser.Usage(nil))
		os.Exit(1)
	}

	// testing values
	args.Debug = true
	args.UpdateArgs.DryRun = true

	return args
}
