package cli

import (
	"fmt"
	"os"
	"path"

	"github.com/seventv/helm-manager/argparse"
	"go.uber.org/zap"
)

type Trigger func(args *Arguments) error

type CommandMode int

type Command struct {
	Name string
	Help string
	Mode CommandMode
}

const (
	CommandModeBase CommandMode = iota

	CommandModeInit

	CommandModeUpgrade

	CommandModeAdd
	CommandModeAddChart
	CommandModeAddSingle
	CommandModeAddRepo
	CommandModeAddEnv

	CommandModeRemove
	CommandModeRemoveChart
	CommandModeRemoveSingle
	CommandModeRemoveRepo
	CommandModeRemoveEnv

	CommandModeUpdate
)

var BaseCommand = Command{
	Name: path.Base(os.Args[0]),
	Help: "Manage Helm Charts and their values",
	Mode: CommandModeBase,
}

type Arguments struct {
	WorkingDir   string
	ManifestFile string
	InTerminal   bool

	Debug   bool
	Mode    CommandMode
	Upgrade Upgrade
	Add     Add
	Remove  Remove
	Update  Update
}

func BaseCli(parser argparse.Parser, args Arguments) Trigger {
	debugFlag := parser.Flag("", "debug", &argparse.Options[bool]{
		Required: false,
		Help:     "Enable debug logging",
	})
	manifestFileFlag := parser.String("m", "manifest", &argparse.Options[string]{
		Required: false,
		Help:     "The manifest file to use",
		Default:  "manifest.yaml",
	})
	cli := parser.Flag("", "cli", &argparse.Options[bool]{
		Required: false,
		Help:     "CLI Mode Only (Non-Interactive Mode)",
	})

	return func(args *Arguments) error {
		args.Debug = *debugFlag
		args.ManifestFile = *manifestFileFlag
		args.WorkingDir = path.Dir(args.ManifestFile)
		if args.InTerminal {
			args.InTerminal = !*cli
		}

		args.Mode = CommandModeBase

		return nil
	}
}

var Parser = argparse.NewParser(BaseCommand.Name, BaseCommand.Help)

func ReadArguments() Arguments {
	var args Arguments

	args.InTerminal = CheckTerminal()

	triggers := []Trigger{}
	triggers = append(triggers,
		BaseCli(Parser, args),
		UpgradeCli(Parser, args),
		AddCli(Parser, args),
		RemoveCli(Parser, args),
		InitCli(Parser, args),
		UpdateCli(Parser, args),
	)

	if err := Parser.Parse(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, Parser.Usage(err.Error()))
		os.Exit(1)
	}

	for _, trigger := range triggers {
		if err := trigger(&args); err != nil {
			zap.S().Fatal(Parser.Usage(err.Error()))
		}
	}

	return args
}
