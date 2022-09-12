package cli

import (
	"fmt"
	"os"
	"path"

	"github.com/seventv/helm-manager/argparse"
)

type Trigger func(args *Arguments)

type CommandMode int

const (
	_ CommandMode = iota

	CommandModeInit

	CommandModeUpgrade

	CommandModeAddChart
	CommandModeAddSingle
	CommandModeAddRepo
	CommandModeAddEnv

	CommandModeRemoveChart
	CommandModeRemoveSingle
	CommandModeRemoveRepo
	CommandModeRemoveEnv
)

func ReadArguments() Arguments {
	var args Arguments

	parser := argparse.NewParser(path.Base(os.Args[0]), "Manage Helm Charts and their values")

	triggers := []Trigger{}
	triggers = append(triggers,
		BaseCli(parser),
		UpgradeCli(parser),
		AddCli(parser),
		RemoveCli(parser),
		InitCli(parser),
	)

	if err := parser.Parse(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, parser.Usage(err.Error()))
		os.Exit(1)
	}

	for _, trigger := range triggers {
		trigger(&args)
	}

	if args.Mode == 0 {
		fmt.Fprintln(os.Stderr, parser.Usage("no command specified"))
		os.Exit(1)
	}

	return args
}
