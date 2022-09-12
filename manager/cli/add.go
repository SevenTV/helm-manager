package cli

import (
	"github.com/seventv/helm-manager/argparse"
)

type Add struct {
	Chart  AddChart
	Single AddSingle
	Repo   AddRepo
	Env    AddEnv
}

func AddCli(parser argparse.Command) Trigger {
	addCmd := parser.NewCommand("add", "Add a new chart, single, repo or env variable to the manifest")

	triggers := []Trigger{
		AddChartCli(addCmd),
		AddSingleCli(addCmd),
		AddRepoCli(addCmd),
		AddEnvCli(addCmd),
	}

	return func(args *Arguments) {
		if !addCmd.Happened() {
			return
		}

		for _, trigger := range triggers {
			trigger(args)
		}
	}
}

type AddChart struct {
	Name      string
	Chart     string
	Version   string
	Namespace string
	File      string
	Overwrite bool
}

func AddChartCli(addCmd argparse.Command) Trigger {
	addChartCmd := addCmd.NewCommand("chart", "Add a new chart to the manifest")

	chartNameFlag := addChartCmd.StringPositional("name", &argparse.Options[string]{
		Required: true,
		Help:     "The name of the chart",
	})

	chartChartFlag := addChartCmd.StringPositional("chart", &argparse.Options[string]{
		Required: true,
		Help:     "The chart to use",
	})

	chartVersionFlag := addChartCmd.String("v", "version", &argparse.Options[string]{
		Required: true,
		Help:     "The version of the chart",
	})

	chartNamespaceFlag := addChartCmd.String("n", "namespace", &argparse.Options[string]{
		Required: true,
		Help:     "The namespace to install the chart into",
	})

	chartValuesFileFlag := addChartCmd.String("f", "values", &argparse.Options[string]{
		Required: false,
		Help:     "The values file to use",
	})

	chartOverwriteFlag := addChartCmd.Flag("", "overwrite", &argparse.Options[bool]{
		Required: false,
		Help:     "Overwrite the chart file if it already exists",
	})

	return func(args *Arguments) {
		if !addChartCmd.Happened() {
			return
		}

		args.Mode = CommandModeAddChart
		args.Add.Chart.Chart = *chartChartFlag
		args.Add.Chart.Name = *chartNameFlag
		args.Add.Chart.Version = *chartVersionFlag
		args.Add.Chart.Namespace = *chartNamespaceFlag
		args.Add.Chart.File = *chartValuesFileFlag
		args.Add.Chart.Overwrite = *chartOverwriteFlag
	}
}

type AddSingle struct {
	Name      string
	Namespace string
	File      string
	UseCreate bool
}

func AddSingleCli(addCmd argparse.Command) Trigger {
	addSingleCmd := addCmd.NewCommand("single", "Add a new single to the manifest")

	singleNameFlag := addSingleCmd.StringPositional("name", &argparse.Options[string]{
		Required: true,
		Help:     "The name of the single",
	})

	singleNamespaceFlag := addSingleCmd.String("n", "namespace", &argparse.Options[string]{
		Required: true,
		Help:     "The namespace to install the single into",
	})

	singleValuesFileFlag := addSingleCmd.String("f", "values", &argparse.Options[string]{
		Required: false,
		Help:     "The single file to use",
	})

	useCreateFlag := addSingleCmd.Flag("", "use-create", &argparse.Options[bool]{
		Required: false,
		Help:     "Use the kubectl create command instead of apply",
	})

	return func(args *Arguments) {
		if !addSingleCmd.Happened() {
			return
		}

		args.Mode = CommandModeAddSingle
		args.Add.Single.Name = *singleNameFlag
		args.Add.Single.Namespace = *singleNamespaceFlag
		args.Add.Single.File = *singleValuesFileFlag
		args.Add.Single.UseCreate = *useCreateFlag
	}
}

type AddRepo struct {
	Name string
	URL  string
}

func AddRepoCli(addCmd argparse.Command) Trigger {
	addRepoCmd := addCmd.NewCommand("repo", "Add a new repo to the manifest")

	repoNameFlag := addRepoCmd.StringPositional("name", &argparse.Options[string]{
		Required: true,
		Help:     "The name of the repo",
	})

	repoURLFlag := addRepoCmd.StringPositional("url", &argparse.Options[string]{
		Required: true,
		Help:     "The URL of the repo",
	})

	return func(args *Arguments) {
		if !addRepoCmd.Happened() {
			return
		}

		args.Mode = CommandModeAddRepo
		args.Add.Repo.Name = *repoNameFlag
		args.Add.Repo.URL = *repoURLFlag
	}
}

type AddEnv struct {
	Name string
}

func AddEnvCli(addCmd argparse.Command) Trigger {
	addEnvCmd := addCmd.NewCommand("env", "Add a new env variable to the manifest")

	envNameFlag := addEnvCmd.StringPositional("name", &argparse.Options[string]{
		Required: true,
		Help:     "The name of the env variable",
	})

	return func(args *Arguments) {
		if !addEnvCmd.Happened() {
			return
		}

		args.Mode = CommandModeAddEnv
		args.Add.Env.Name = *envNameFlag
	}
}
