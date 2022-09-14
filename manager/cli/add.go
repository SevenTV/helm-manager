package cli

import (
	"errors"

	"github.com/seventv/helm-manager/argparse"
)

var AddCommand = Command{
	Name: "add",
	Help: "Add a new chart, single, repo or env variable to the manifest",
	Mode: CommandModeAdd,
}

type Add struct {
	Chart  AddChart
	Single AddSingle
	Repo   AddRepo
	Env    AddEnv
}

func AddCli(parser argparse.Command, args Arguments) Trigger {
	addCmd := parser.NewCommand(AddCommand.Name, AddCommand.Help)

	triggers := []Trigger{
		AddChartCli(addCmd, args),
		AddSingleCli(addCmd, args),
		AddRepoCli(addCmd, args),
		AddEnvCli(addCmd, args),
	}

	return func(args *Arguments) error {
		if !addCmd.Happened() {
			return nil
		}

		args.Mode = CommandModeAdd

		for _, trigger := range triggers {
			if err := trigger(args); err != nil {
				return err
			}
		}

		return nil
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

var AddChartCommand = Command{
	Name: "chart",
	Help: "Add a new chart to the manifest",
	Mode: CommandModeAddChart,
}

func AddChartCli(addCmd argparse.Command, args Arguments) Trigger {
	addChartCmd := addCmd.NewCommand(AddChartCommand.Name, AddChartCommand.Help)

	chartNamePos := addChartCmd.StringPositional("name", &argparse.Options[string]{
		Required: false,
		Help:     "The name of the chart",
	})

	chartNameFlag := addChartCmd.String("", "name", &argparse.Options[string]{
		Required: false,
		Help:     "The name of the chart",
	})

	chartChartPos := addChartCmd.StringPositional("chart", &argparse.Options[string]{
		Required: false,
		Help:     "The chart to use",
	})

	chartChartFlag := addChartCmd.String("c", "chart", &argparse.Options[string]{
		Required: false,
		Help:     "The chart to use",
	})

	chartVersionPos := addChartCmd.StringPositional("version", &argparse.Options[string]{
		Required: false,
		Help:     "The version of the chart to use",
	})

	chartVersionFlag := addChartCmd.String("v", "version", &argparse.Options[string]{
		Required: !args.NonInteractive,
		Help:     "The version of the chart",
	})

	chartNamespacePos := addChartCmd.StringPositional("namespace", &argparse.Options[string]{
		Required: false,
		Help:     "The namespace to install the chart into",
	})

	chartNamespaceFlag := addChartCmd.String("n", "namespace", &argparse.Options[string]{
		Required: !args.NonInteractive,
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

	return func(args *Arguments) error {
		if !addChartCmd.Happened() {
			return nil
		}

		args.Mode = CommandModeAddChart
		args.Add.Chart.Name = *chartNameFlag
		if args.Add.Chart.Name == "" {
			args.Add.Chart.Name = *chartNamePos
		} else if *chartNamePos != "" {
			return errors.New("name cannot be specified twice")
		}

		if args.Add.Chart.Name == "" && args.NonInteractive {
			return errors.New("no chart name provided")
		}

		args.Add.Chart.Chart = *chartChartFlag
		if args.Add.Chart.Chart == "" {
			args.Add.Chart.Chart = *chartChartPos
		} else if *chartChartPos != "" {
			return errors.New("chart cannot be specified twice")
		}

		if args.Add.Chart.Chart == "" && args.NonInteractive {
			return errors.New("no chart provided")
		}

		args.Add.Chart.Version = *chartVersionFlag
		if args.Add.Chart.Version == "" {
			args.Add.Chart.Version = *chartVersionPos
		} else if *chartVersionPos != "" {
			return errors.New("version cannot be specified twice")
		}

		if args.Add.Chart.Version == "" && args.NonInteractive {
			return errors.New("no chart version provided")
		}

		args.Add.Chart.Namespace = *chartNamespaceFlag
		if args.Add.Chart.Namespace == "" {
			args.Add.Chart.Namespace = *chartNamespacePos
		} else if *chartNamespacePos != "" {
			return errors.New("namespace cannot be specified twice")
		}

		if args.Add.Chart.Namespace == "" && args.NonInteractive {
			return errors.New("no chart namespace provided")
		}

		args.Add.Chart.File = *chartValuesFileFlag
		args.Add.Chart.Overwrite = *chartOverwriteFlag

		return nil
	}
}

type AddSingle struct {
	Name      string
	Namespace string
	File      string
	UseCreate bool
	Overwrite bool
}

var AddSingleCommand = Command{
	Name: "single",
	Help: "Add a new single file to the manifest",
	Mode: CommandModeAddSingle,
}

func AddSingleCli(addCmd argparse.Command, args Arguments) Trigger {
	addSingleCmd := addCmd.NewCommand(AddSingleCommand.Name, AddSingleCommand.Help)

	singleNamePos := addSingleCmd.StringPositional("name", &argparse.Options[string]{
		Required: false,
		Help:     "The name of the single",
	})

	singleNameFlag := addSingleCmd.String("", "name", &argparse.Options[string]{
		Required: false,
		Help:     "The name of the single",
	})

	singleNamespacePos := addSingleCmd.StringPositional("namespace", &argparse.Options[string]{
		Required: false,
		Help:     "The namespace to install the single into",
	})

	singleNamespaceFlag := addSingleCmd.String("n", "namespace", &argparse.Options[string]{
		Required: false,
		Help:     "The namespace to install the single into",
	})

	singleValuesFileFlag := addSingleCmd.String("f", "values", &argparse.Options[string]{
		Required: false,
		Help:     "The single file to use",
	})

	singleUseCreateFlag := addSingleCmd.Flag("", "use-create", &argparse.Options[bool]{
		Required: false,
		Help:     "Use the kubectl create command instead of apply",
	})

	singleOverwriteFlag := addSingleCmd.Flag("", "overwrite", &argparse.Options[bool]{
		Required: false,
		Help:     "Overwrite the single file if it already exists",
	})

	return func(args *Arguments) error {
		if !addSingleCmd.Happened() {
			return nil
		}

		args.Mode = CommandModeAddSingle
		args.Add.Single.Name = *singleNameFlag
		if args.Add.Single.Name == "" {
			args.Add.Single.Name = *singleNamePos
		} else if *singleNamePos != "" {
			return errors.New("name cannot be specified twice")
		}

		if args.Add.Single.Name == "" && args.NonInteractive {
			return errors.New("no single name provided")
		}

		args.Add.Single.Namespace = *singleNamespaceFlag
		if args.Add.Single.Namespace == "" {
			args.Add.Single.Namespace = *singleNamespacePos
		} else if *singleNamespacePos != "" {
			return errors.New("namespace cannot be specified twice")
		}

		if args.Add.Single.Namespace == "" && args.NonInteractive {
			return errors.New("no single namespace provided")
		}

		args.Add.Single.File = *singleValuesFileFlag
		if args.Add.Single.File == "" && args.NonInteractive {
			return errors.New("no single file provided")
		}

		args.Add.Single.UseCreate = *singleUseCreateFlag
		args.Add.Single.Overwrite = *singleOverwriteFlag

		return nil
	}
}

type AddRepo struct {
	Name string
	URL  string
}

var AddRepoCommand = Command{
	Name: "repo",
	Help: "Add a new repo to the manifest",
	Mode: CommandModeAddRepo,
}

func AddRepoCli(addCmd argparse.Command, args Arguments) Trigger {
	addRepoCmd := addCmd.NewCommand(AddRepoCommand.Name, AddRepoCommand.Help)

	repoNamePos := addRepoCmd.StringPositional("name", &argparse.Options[string]{
		Required: false,
		Help:     "The name of the repo",
	})

	repoNameFlag := addRepoCmd.String("", "name", &argparse.Options[string]{
		Required: false,
		Help:     "The name of the repo",
	})

	repoURLPos := addRepoCmd.StringPositional("url", &argparse.Options[string]{
		Required: false,
		Help:     "The URL of the repo",
	})

	repoURLFlag := addRepoCmd.String("", "url", &argparse.Options[string]{
		Required: false,
		Help:     "The URL of the repo",
	})

	return func(args *Arguments) error {
		if !addRepoCmd.Happened() {
			return nil
		}

		args.Mode = CommandModeAddRepo
		args.Add.Repo.Name = *repoNameFlag
		if args.Add.Repo.Name == "" {
			args.Add.Repo.Name = *repoNamePos
		} else if *repoNamePos != "" {
			return errors.New("name cannot be specified twice")
		}

		if args.Add.Repo.Name == "" && args.NonInteractive {
			return errors.New("no repo name provided")
		}

		args.Add.Repo.URL = *repoURLFlag
		if args.Add.Repo.URL == "" {
			args.Add.Repo.URL = *repoURLPos
		} else if *repoURLPos != "" {
			return errors.New("url cannot be specified twice")
		}

		if args.Add.Repo.URL == "" && args.NonInteractive {
			return errors.New("no repo url provided")
		}

		return nil
	}
}

type AddEnv struct {
	Name string
}

var AddEnvCommand = Command{
	Name: "env",
	Help: "Add a new whitelisted environment to the manifest",
	Mode: CommandModeAddEnv,
}

func AddEnvCli(addCmd argparse.Command, args Arguments) Trigger {
	addEnvCmd := addCmd.NewCommand(AddEnvCommand.Name, AddEnvCommand.Help)

	envNamePos := addEnvCmd.StringPositional("name", &argparse.Options[string]{
		Required: false,
		Help:     "The name of the env variable",
	})

	envNameFlag := addEnvCmd.String("", "name", &argparse.Options[string]{
		Required: false,
		Help:     "The name of the env variable",
	})

	return func(args *Arguments) error {
		if !addEnvCmd.Happened() {
			return nil
		}

		args.Mode = CommandModeAddEnv
		args.Add.Env.Name = *envNameFlag
		if args.Add.Env.Name == "" {
			args.Add.Env.Name = *envNamePos
		} else if *envNamePos != "" {
			return errors.New("name cannot be specified twice")
		}

		if args.Add.Env.Name == "" && args.NonInteractive {
			return errors.New("no env name provided")
		}

		return nil
	}
}
