package cli

import "github.com/seventv/helm-manager/argparse"

type Remove struct {
	Chart  RemoveChart
	Single RemoveSingle
	Repo   RemoveRepo
	Env    RemoveEnv
}

func RemoveCli(parser argparse.Parser) Trigger {
	removeCmd := parser.NewCommand("remove", "Remove a chart, single, repo or env variable from the manifest")

	triggers := []Trigger{
		RemoveChartCli(removeCmd),
		RemoveSingleCli(removeCmd),
		RemoveRepoCli(removeCmd),
		RemoveEnvCli(removeCmd),
	}

	return func(args *Arguments) {
		if !removeCmd.Happened() {
			return
		}

		for _, trigger := range triggers {
			trigger(args)
		}
	}
}

type RemoveChart struct {
	Name   string
	Wait   bool
	DryRun bool
}

func RemoveChartCli(removeCmd argparse.Command) Trigger {
	removeChartCmd := removeCmd.NewCommand("chart", "Remove a chart from the manifest")

	chartNameFlag := removeChartCmd.StringPositional("name", &argparse.Options[string]{
		Required: true,
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

	return func(args *Arguments) {
		if !removeChartCmd.Happened() {
			return
		}

		args.Mode = CommandModeRemoveChart
		args.Remove.Chart.Name = *chartNameFlag
		args.Remove.Chart.Wait = *chartWaitFlag
		args.Remove.Chart.DryRun = *chartDryRunFlag
	}
}

type RemoveSingle struct {
	Name string
}

func RemoveSingleCli(removeCmd argparse.Command) Trigger {
	removeSingleCmd := removeCmd.NewCommand("single", "Remove a single from the manifest")

	singleNameFlag := removeSingleCmd.StringPositional("name", &argparse.Options[string]{
		Required: true,
		Help:     "The name of the single",
	})

	return func(args *Arguments) {
		if !removeSingleCmd.Happened() {
			return
		}

		args.Mode = CommandModeRemoveSingle
		args.Remove.Single.Name = *singleNameFlag
	}
}

type RemoveRepo struct {
	Name string
}

func RemoveRepoCli(removeCmd argparse.Command) Trigger {
	removeRepoCmd := removeCmd.NewCommand("repo", "Remove a repo from the manifest")

	repoNameFlag := removeRepoCmd.StringPositional("name", &argparse.Options[string]{
		Required: true,
		Help:     "The name of the repo",
	})

	return func(args *Arguments) {
		if !removeRepoCmd.Happened() {
			return
		}

		args.Mode = CommandModeRemoveRepo
		args.Remove.Repo.Name = *repoNameFlag
	}
}

type RemoveEnv struct {
	Name string
}

func RemoveEnvCli(removeCmd argparse.Command) Trigger {
	removeEnvCmd := removeCmd.NewCommand("env", "Remove an env variable from the manifest")

	envNameFlag := removeEnvCmd.StringPositional("name", &argparse.Options[string]{
		Required: true,
		Help:     "The name of the env variable",
	})

	return func(args *Arguments) {
		if !removeEnvCmd.Happened() {
			return
		}

		args.Mode = CommandModeRemoveEnv
		args.Remove.Env.Name = *envNameFlag
	}
}
