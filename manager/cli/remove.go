package cli

import (
	"errors"

	"github.com/seventv/helm-manager/argparse"
)

type Remove struct {
	Chart  RemoveChart
	Single RemoveSingle
	Repo   RemoveRepo
	Env    RemoveEnv
}

var RemoveCommand = Command{
	Name: "remove",
	Help: "Remove a chart, single, repo or env variable from the manifest",
	Mode: CommandModeRemove,
}

// const RemoveCommandHelp = "Remove a chart, single, repo or env variable from the manifest"
// const RemoveCommand = "remove"

func RemoveCli(parser argparse.Parser, args Arguments) Trigger {
	removeCmd := parser.NewCommand(RemoveCommand.Name, RemoveCommand.Help)

	triggers := []Trigger{
		RemoveChartCli(removeCmd, args),
		RemoveSingleCli(removeCmd, args),
		RemoveRepoCli(removeCmd, args),
		RemoveEnvCli(removeCmd, args),
	}

	return func(args *Arguments) error {
		if !removeCmd.Happened() {
			return nil
		}

		args.Mode = CommandModeRemove

		for _, trigger := range triggers {
			if err := trigger(args); err != nil {
				return err
			}
		}

		return nil
	}
}

type RemoveChart struct {
	Name    string
	Wait    bool
	DryRun  bool
	Delete  bool
	Confirm bool
}

var RemoveChartCommand = Command{
	Name: "chart",
	Help: "Remove a chart from the manifest",
	Mode: CommandModeRemoveChart,
}

func RemoveChartCli(removeCmd argparse.Command, args Arguments) Trigger {
	removeChartCmd := removeCmd.NewCommand(RemoveChartCommand.Name, RemoveChartCommand.Help)

	chartNamePos := removeChartCmd.StringPositional("name", &argparse.Options[string]{
		Required: false,
		Help:     "The name of the chart",
	})

	chartNameFlag := removeChartCmd.String("", "name", &argparse.Options[string]{
		Required: false,
		Help:     "The name of the chart",
	})

	chartWaitFlag := removeChartCmd.Flag("", "wait", &argparse.Options[bool]{
		Required: false,
		Help:     "Wait for the chart to be removed",
	})

	chartDryRunFlag := removeChartCmd.Flag("", "dry-run", &argparse.Options[bool]{
		Required: false,
		Help:     "Dry run the chart removal",
	})

	chartDeleteFlag := removeChartCmd.Flag("", "delete", &argparse.Options[bool]{
		Required: false,
		Help:     "Delete the chart file",
	})

	chartConfirmFlag := removeChartCmd.Flag("y", "confirm", &argparse.Options[bool]{
		Required: false,
		Help:     "Confirm the chart removal",
	})

	return func(args *Arguments) error {
		if !removeChartCmd.Happened() {
			return nil
		}

		args.Mode = CommandModeRemoveChart
		args.Remove.Chart.Name = *chartNameFlag
		if args.Remove.Chart.Name == "" {
			args.Remove.Chart.Name = *chartNamePos
		} else if *chartNamePos != "" {
			return errors.New("chart name cannot be specified twice")
		}

		if args.Remove.Chart.Name == "" && !args.InTerminal {
			return errors.New("chart name is required")
		}

		args.Remove.Chart.Wait = *chartWaitFlag
		args.Remove.Chart.DryRun = *chartDryRunFlag
		args.Remove.Chart.Delete = *chartDeleteFlag
		args.Remove.Chart.Confirm = *chartConfirmFlag

		return nil
	}
}

type RemoveSingle struct {
	Name    string
	Delete  bool
	DryRun  bool
	Confirm bool
}

var RemoveSingleCommand = Command{
	Name: "single",
	Help: "Remove a single from the manifest",
	Mode: CommandModeRemoveSingle,
}

func RemoveSingleCli(removeCmd argparse.Command, args Arguments) Trigger {
	removeSingleCmd := removeCmd.NewCommand(RemoveSingleCommand.Name, RemoveSingleCommand.Help)

	singleNamePos := removeSingleCmd.StringPositional("name", &argparse.Options[string]{
		Required: false,
		Help:     "The name of the single",
	})

	singleNameFlag := removeSingleCmd.String("", "name", &argparse.Options[string]{
		Required: false,
		Help:     "The name of the single",
	})

	singleDeleteFlag := removeSingleCmd.Flag("", "delete", &argparse.Options[bool]{
		Required: false,
		Help:     "Delete the single file",
	})

	singleDryRunFlag := removeSingleCmd.Flag("", "dry-run", &argparse.Options[bool]{
		Required: false,
		Help:     "Dry run the single removal",
	})

	singleConfirmFlag := removeSingleCmd.Flag("y", "confirm", &argparse.Options[bool]{
		Required: false,
		Help:     "Confirm the single removal",
	})

	return func(args *Arguments) error {
		if !removeSingleCmd.Happened() {
			return nil
		}

		args.Mode = CommandModeRemoveSingle
		args.Remove.Single.Name = *singleNameFlag
		if args.Remove.Single.Name == "" {
			args.Remove.Single.Name = *singleNamePos
		} else if *singleNamePos != "" {
			return errors.New("single name cannot be specified twice")
		}

		if args.Remove.Single.Name == "" && !args.InTerminal {
			return errors.New("single name is required")
		}

		args.Remove.Single.Delete = *singleDeleteFlag
		args.Remove.Single.DryRun = *singleDryRunFlag
		args.Remove.Single.Confirm = *singleConfirmFlag

		return nil
	}
}

type RemoveRepo struct {
	Name string
}

var RemoveRepoCommand = Command{
	Name: "repo",
	Help: "Remove a repo from the manifest",
	Mode: CommandModeRemoveRepo,
}

func RemoveRepoCli(removeCmd argparse.Command, args Arguments) Trigger {
	removeRepoCmd := removeCmd.NewCommand(RemoveRepoCommand.Name, RemoveRepoCommand.Help)

	repoNamePos := removeRepoCmd.StringPositional("name", &argparse.Options[string]{
		Required: false,
		Help:     "The name of the repo",
	})

	repoNameFlag := removeRepoCmd.String("", "name", &argparse.Options[string]{
		Required: false,
		Help:     "The name of the repo",
	})

	return func(args *Arguments) error {
		if !removeRepoCmd.Happened() {
			return nil
		}

		args.Mode = CommandModeRemoveRepo
		args.Remove.Repo.Name = *repoNameFlag
		if args.Remove.Repo.Name == "" {
			args.Remove.Repo.Name = *repoNamePos
		} else if *repoNamePos != "" {
			return errors.New("repo name cannot be specified twice")
		}

		if args.Remove.Repo.Name == "" && !args.InTerminal {
			return errors.New("repo name is required")
		}

		return nil
	}
}

type RemoveEnv struct {
	Name string
}

var RemoveEnvCommand = Command{
	Name: "env",
	Help: "Remove an env variable from the manifest",
	Mode: CommandModeRemoveEnv,
}

func RemoveEnvCli(removeCmd argparse.Command, args Arguments) Trigger {
	removeEnvCmd := removeCmd.NewCommand(RemoveEnvCommand.Name, RemoveEnvCommand.Help)

	envNamePos := removeEnvCmd.StringPositional("name", &argparse.Options[string]{
		Required: false,
		Help:     "The name of the env variable",
	})

	envNameFlag := removeEnvCmd.String("", "name", &argparse.Options[string]{
		Required: false,
		Help:     "The name of the env variable",
	})

	return func(args *Arguments) error {
		if !removeEnvCmd.Happened() {
			return nil
		}

		args.Mode = CommandModeRemoveEnv
		args.Remove.Env.Name = *envNameFlag
		if args.Remove.Env.Name == "" {
			args.Remove.Env.Name = *envNamePos
		} else if *envNamePos != "" {
			return errors.New("env name cannot be specified twice")
		}

		if args.Remove.Env.Name == "" && !args.InTerminal {
			return errors.New("env name is required")
		}

		return nil
	}
}
