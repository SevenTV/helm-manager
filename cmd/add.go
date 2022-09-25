package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/fatih/color"
	"github.com/seventv/helm-manager/v2/cmd/ui"
	"github.com/seventv/helm-manager/v2/constants"
	"github.com/seventv/helm-manager/v2/external"
	"github.com/seventv/helm-manager/v2/logger"
	"github.com/seventv/helm-manager/v2/types"
	"github.com/seventv/helm-manager/v2/utils"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func init() {
	rootCmd.AddCommand(addCmd)

	{
		addCmd.AddCommand(addRepoCmd)
		addRepoCmd.Flags().StringVar(&Args.Name, "name", "", "Name of the repo")
		addRepoCmd.Flags().StringVar(&Args.AddRepoCmd.URL, "url", "", "URL of the repo")
	}

	{
		addCmd.AddCommand(addSingleCmd)
		addSingleCmd.Flags().StringVar(&Args.Name, "name", "", "Name of the single")
		addSingleCmd.Flags().StringVar(&Args.File, "file", "", "File to create the single from")
		addSingleCmd.Flags().StringVarP(&Args.Namespace, "namespace", "n", "", "Namespace to use")
		addSingleCmd.Flags().BoolVar(&Args.Deploy, "deploy", false, "Deploy the single after creation")
		addSingleCmd.Flags().BoolVar(&Args.Force, "force", false, "Force overwrite the single if it already exists")
	}

	{
		addCmd.AddCommand(addReleaseCmd)
		addReleaseCmd.Flags().StringVar(&Args.Name, "name", "", "Name of the release")
		addReleaseCmd.Flags().StringVar(&Args.AddReleaseCmd.Chart, "chart", "", "Chart of the release")
		addReleaseCmd.Flags().StringVar(&Args.AddReleaseCmd.Version, "version", "", "Version of the release")
		addReleaseCmd.Flags().StringVarP(&Args.File, "file", "f", "", "File to create the release from")
		addReleaseCmd.Flags().StringVarP(&Args.Namespace, "namespace", "n", "", "Namespace to use")
		addReleaseCmd.Flags().BoolVar(&Args.Deploy, "deploy", false, "Deploy the release after creation")
		addReleaseCmd.Flags().BoolVar(&Args.Force, "force", false, "Force overwrite the release if it already exists")
	}

	{
		addCmd.AddCommand(addEnvCmd)
		addEnvCmd.Flags().StringVar(&Args.Name, "name", "", "Name of the env variable to be whitelisted")
	}

	{
		addCmd.AddCommand(addLocalChartCmd)
		addLocalChartCmd.Flags().StringVar(&Args.File, "path", "", "Path to the local chart")
	}
}

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new release, single, repo or env variable to the manifest",
	Long:  "Add a new release, single, repo or env variable to the manifest",
	Args:  ui.SubCommandRequired(cobra.NoArgs),
	Run: func(cmd *cobra.Command, args []string) {
		zap.S().Infof("* %s *\r", color.GreenString("Helm Manager Add"))

		ManifestExist(cmd)

		cmds := make([]ui.SelectableCommand, 0, len(cmd.Commands()))
		for _, cmd := range cmd.Commands() {
			if !cmd.Hidden && cmd.Name() != "help" {
				cmds = append(cmds, ui.CmdSelectable(cmd))
			}
		}

		ui.RunSubCommand(cmd, cmds)
	},
}

var addReleaseCmd = &cobra.Command{
	Use:     "release",
	Short:   "Add a new release to the manifest",
	Long:    "Add a new release to the manifest",
	Example: "   helm-manager add chart [NAME] [CHART]\n   helm-manager add chart mongo bitnami/mongodb --namespace mongo --version 10.0.0" + USAGE_EXTRA,
	Args: ui.PositionalArgs([]ui.RequiredArg{
		ui.Arg[string]{
			Name:       "name",
			Ptr:        &Args.Name,
			Positional: true,
			Validator: types.MultiValidator(
				types.NameValidator("name", false),
				types.NotEqualValidator(
					types.ToStringer(`"%s" is already used in the manifest`),
					types.FutureFromStringers(types.FutureFromPtr(&Manifest.Releases)),
				),
			),
			UI: ui.PromptUiFunc[string]("Name"),
			Callback: func(n string) error {
				Args.Name = strings.ToLower(n)

				return nil
			},
		},
		ui.Arg[string]{
			Name:       "chart",
			Ptr:        &Args.AddReleaseCmd.Chart,
			Positional: true,
			Validator: types.MultiValidator(
				types.EmptyValidator[string]("chart", false),
				types.EqualValidator(
					types.ToStringer("\"%s\" is not a valid chart name in the added helm repositories"),
					types.FutureFromStringers(HelmChartsFuture),
				),
			),
			UI: ui.PromptUiSelectorFunc[string]("Chart", "", func(i int) error {
				charts, err := HelmChartsFuture.Get()
				if err != nil {
					return err
				}

				Args.AddReleaseCmd.Chart = charts[i].RepoName

				return nil
			}, types.FutureInterfacerArray[types.HelmChartMulti, types.Selectable](HelmChartsFuture)),
			Callback: func(s string) error {
				Args.AddReleaseCmd.Chart = strings.ToLower(s)

				return nil
			},
		},
		ui.Arg[string]{
			Name: "namespace",
			Ptr:  &Args.Namespace,
			Callback: func(ns string) error {
				Args.Namespace = strings.ToLower(ns)

				releases, err := HelmReleaseFuture.Get()
				if err != nil {
					return err
				}

				for _, release := range releases {
					if strings.ToLower(release.Name) == Args.Name && strings.ToLower(release.Namespace) == Args.Namespace {
						return fmt.Errorf(`release "%s" already exists in the namespace "%s" on the cluster`, Args.Name, ns)
					}
				}

				return nil
			},
			Validator: types.NameValidator("namespace", false),
			UI: ui.PromptUiSelectorNewFunc[string]("Namespace", "", func(i int) error {
				ns, err := NamespaceFuture.Get()
				if err != nil {
					return err
				}

				Args.Namespace = ns[i]

				return nil
			}, types.FutureFromFunc(func() []types.Selectable {
				ns, err := NamespaceFuture.Get()
				if err != nil {
					logger.Errorf("Failed to get namespaces, ", err)
					return nil
				}

				namespaces := make([]types.Selectable, len(ns))
				for i, n := range ns {
					namespaces[i] = types.SelectableString(n)
				}

				return namespaces
			})),
		},
		ui.Arg[string]{
			Name: "version",
			Ptr:  &Args.AddReleaseCmd.Version,
			Validator: types.EqualValidator(
				types.StringerFunc(func() string {
					return fmt.Sprintf(`"%s" is not a version of "%s"`, "%s", Args.AddReleaseCmd.Chart)
				}),
				types.FutureFromFuncErr(func() ([]string, error) {
					charts, err := HelmChartsFuture.Get()
					if err != nil {
						return nil, err
					}

					for _, c := range charts {
						if c.RepoName == Args.AddReleaseCmd.Chart {
							versions := make([]string, len(c.Versions))
							for i, v := range c.Versions {
								versions[i] = v.Version
							}

							return versions, nil
						}
					}

					return nil, fmt.Errorf("chart not found")
				}),
			),
			UI: ui.PromptUiSelectorFunc[string]("Version", "", func(i int) error {
				charts, err := HelmChartsFuture.Get()
				if err != nil {
					return err
				}

				for _, c := range charts {
					if c.RepoName == Args.AddReleaseCmd.Chart {
						Args.AddReleaseCmd.Version = c.Versions[i].Version

						return nil
					}
				}

				return nil
			}, types.FutureFromFuncErr(func() ([]types.Selectable, error) {
				chart, err := HelmChartsFuture.Get()
				if err != nil {
					return nil, err
				}

				for _, c := range chart {
					if c.RepoName == Args.AddReleaseCmd.Chart {
						versions := make([]types.Selectable, len(c.Versions))
						for i, v := range c.Versions {
							versions[i] = v
						}

						return versions, nil
					}
				}

				return nil, fmt.Errorf("chart not found")
			})),
		},
		ui.Arg[string]{
			Name:      "values",
			Ptr:       &Args.File,
			Validator: types.PathValidator("values", true, false, true, constants.StdinUsed()),
			UI:        ui.PromptUiFunc[string]("Do you want to provide input values"),
		},
		ui.Arg[bool]{
			Name: "force",
			Ptr:  &Args.Force,
			Disabled: types.FutureFromFunc(func() bool {
				_, err := os.Stat(ReleasePath(Args.Name))
				return err != nil
			}),
			UI: ui.PromptUiConfirmFunc("The release file already exists, do you want to overwrite it", false),
			Callback: func(force bool) error {
				_, err := os.Stat(ReleasePath(Args.Name))
				if !force && err == nil {
					logger.Fatal("Release file already exists, use --force to overwrite it")
				}

				return nil
			},
		},
		ui.Arg[bool]{
			Name: "deploy",
			Ptr:  &Args.Deploy,
			UI:   ui.PromptUiConfirmFunc("Do you want to deploy changes to the cluster", true),
		},
	}, func(cmd *cobra.Command) {
		zap.S().Infof("* %s *", color.GreenString("Helm Manager Add Release"))
		ManifestExist(cmd)
	}),
	Run: func(cmd *cobra.Command, args []string) {
		charts, err := HelmChartsFuture.Get()
		if err != nil {
			logger.Fatalf("failed to get helm charts: %s", err)
		}

		mutliChart := types.HelmChartMultiArray(charts).FindChart(Args.AddReleaseCmd.Chart)
		chart := mutliChart.FindVersion(Args.AddReleaseCmd.Version)
		if chart.Version == "" {
			logger.Fatalf("no version %s found for %s", Args.AddReleaseCmd.Version, Args.AddReleaseCmd.Chart)
		}

		release := types.ManifestRelease{
			Name:      Args.Name,
			Namespace: Args.Namespace,
			Chart: types.ManifestChart{
				Name:       chart.Name(),
				Version:    chart.Version,
				AppVersion: chart.AppVersion,
				Repo:       chart.Repo(),
			},
		}

		Manifest.Releases = append(Manifest.Releases, release)

		data, err := utils.ReadFile(Args.File)
		if err != nil {
			logger.Fatalf("Failed to read file: %s", err)
		}

		mutliChart.HelmChart = types.HelmChart(chart)
		result, err := UpgradeDocument(data, mutliChart, true)
		if err != nil {
			logger.Fatal(err)
		}

		if !Args.DryRun {
			if err = os.WriteFile(ReleasePath(Args.Name), result.Document, 0644); err != nil {
				logger.Fatalf("failed to write release file: %s", err)
			}
			utils.WriteManifest(Args.Context)
		} else {
			logger.Info("Dry run, not writing manifest or release file")
		}

		if Args.Deploy {
			err = DeployRelease(release, types.HelmChart(chart), result.EnvSubbedValues, result.EnvSubbedDocument)
			if err != nil {
				logger.Fatal(err)
			}
		}
	},
}

var addSingleCmd = &cobra.Command{
	Use:     "single",
	Short:   "Add a new single to the manifest",
	Long:    "Add a new single to the manifest",
	Example: "   helm-manager add single [NAME] [FILE]\n   helm-manager add single xyz deploy.yaml\n   helm-manager add single xyz -" + USAGE_EXTRA,
	Args: ui.PositionalArgs([]ui.RequiredArg{
		ui.Arg[string]{
			Name:       "name",
			Ptr:        &Args.Name,
			Positional: true,
			Validator: types.MultiValidator(
				types.NameValidator("name", false),
				types.NotEqualValidator(
					types.ToStringer(`"%s" is already used in the manifest`),
					types.FutureFromStringers(types.FutureFromPtr(&Manifest.Singles)),
				),
			),
			UI: ui.PromptUiFunc[string]("Name"),
		},
		ui.Arg[string]{
			Name:       "file",
			Ptr:        &Args.File,
			Positional: true,
			Validator:  types.PathValidator("file", false, false, true, constants.StdinUsed()),
			UI:         ui.PromptUiFunc[string]("File"),
		},
		ui.Arg[string]{
			Name:      "namespace",
			Ptr:       &Args.Namespace,
			Validator: types.NameValidator("namespace", true),
			UI: ui.PromptUiSelectorNewFunc[string]("Namespace", "", func(i int) error {
				if i == 0 {
					Args.Namespace = ""
				} else {
					ns, err := NamespaceFuture.Get()
					if err != nil {
						return err
					}

					Args.Namespace = ns[i-1]
				}

				return nil
			}, types.FutureFromFunc(func() []types.Selectable {
				ns, err := NamespaceFuture.Get()
				if err != nil {
					logger.Error("Failed to get namespaces", zap.Error(err))
				}

				namespaces := make([]types.Selectable, len(ns)+1)
				namespaces[0] = types.SelectableString(color.RedString("* No namespace"))
				for i, n := range ns {
					namespaces[i+1] = types.SelectableString(n)
				}

				return namespaces
			})),
		},
		ui.Arg[bool]{
			Name:       "create",
			Ptr:        &Args.AddSingleCmd.Create,
			Positional: false,
			UI:         ui.PromptUiConfirmFunc("Use kubectl create instead of apply", false),
		},
		ui.Arg[bool]{
			Name:       "deploy",
			Ptr:        &Args.Deploy,
			Positional: false,
			UI:         ui.PromptUiConfirmFunc("Do you want to apply the creation to the cluster", false),
		},
		ui.Arg[bool]{
			Name: "force",
			Ptr:  &Args.Force,
			Disabled: types.FutureFromFunc(func() bool {
				_, err := os.Stat(SinglePath(Args.Name))
				return err != nil
			}),
			UI: ui.PromptUiConfirmFunc("The single file already exists, do you want to overwrite it", false),
			Callback: func(force bool) error {
				_, err := os.Stat(SinglePath(Args.Name))
				if !force && err == nil {
					logger.Fatal("Single file exists, use --force aborted")
				}

				return nil
			},
		},
	}, func(cmd *cobra.Command) {
		zap.S().Infof("* %s *", color.GreenString("Helm Manager Add Single"))
		ManifestExist(cmd)
	}),
	Run: func(cmd *cobra.Command, _ []string) {
		single := types.ManifestSingle{
			Name:      Args.Name,
			Namespace: Args.Namespace,
			UseCreate: Args.AddSingleCmd.Create,
		}

		var (
			data []byte
			err  error
		)
		if Args.File == "-" {
			data, err = ioutil.ReadAll(os.Stdin)
		} else {
			data, err = os.ReadFile(Args.File)
		}
		if err != nil {
			logger.Fatalf("Failed to read file: %v", err)
		}

		if err = os.WriteFile(SinglePath(single.Name), data, 0644); err != nil {
			logger.Fatalf("Failed to write file: %v", err)
		}

		Manifest.Singles = append(Manifest.Singles, single)

		utils.WriteManifest(Args.Context)

		logger.Info("Single added to the manifest")

		if Args.Deploy {
			err = DeploySingle(single, data)
			if err != nil {
				logger.Fatal(err)
			}
		}
	},
}

var addRepoCmd = &cobra.Command{
	Use:     "repo",
	Short:   "Add a new repo to the manifest",
	Long:    "Add a new repo to the manifest",
	Example: "   helm-manager add repo [REPO] [URL]\n   helm manager add repo xyz https://charts.xyz.com" + USAGE_EXTRA,
	Args: ui.PositionalArgs([]ui.RequiredArg{
		ui.Arg[string]{
			Name:       "name",
			Ptr:        &Args.Name,
			Positional: true,
			Validator: types.MultiValidator(
				types.NameValidator("name", false),
				types.NotEqualValidator(
					types.ToStringer(`"%s" is already used in the manifest`),
					types.FutureFromStringers(types.FutureFromPtr(&Manifest.Repos)),
				),
			),
			UI: ui.PromptUiFunc[string]("Name"),
		},
		ui.Arg[string]{
			Name:       "url",
			Ptr:        &Args.AddRepoCmd.URL,
			Positional: true,
			Validator:  types.UrlValidator("url", false),
			UI:         ui.PromptUiFunc[string]("URL"),
		},
	}, func(cmd *cobra.Command) {
		zap.S().Infof("* %s *", color.GreenString("Helm Manager Add Repo"))
		ManifestExist(cmd)
	}),
	Run: func(cmd *cobra.Command, _ []string) {
		repos, err := HelmRepoFuture.Get()
		if err != nil {
			logger.Error("Failed to get helm repos", zap.Error(err))
			return
		}

		exists := false
		for _, r := range repos {
			if r.Name == Args.Name {
				exists = true
				if r.URL != Args.AddRepoCmd.URL {
					logger.Fatalf("Repo name already used: %s", r.Name)
				}
			}
		}

		repo := types.ManifestRepo{
			Name: Args.Name,
			URL:  Args.AddRepoCmd.URL,
		}
		Manifest.Repos = append(Manifest.Repos)

		done := utils.Loader(utils.LoaderOptions{
			FetchingText: "Adding repo to manifest",
			SuccessText:  "Repo added to manifest",
			FailureText:  "Failed to add repo to manifest",
		})

		if !exists {
			resp, err := external.Helm.AddRepo(repo)
			if err != nil {
				done(false)
				logger.Fatalf("Failed to execute helm command: %v\n%s", err, resp)
			}
		}

		resp, err := external.Helm.UpdateRepos()
		if err != nil {
			done(false)
			logger.Fatalf("Failed to execute helm command: %v\n%s", err, resp)
		}

		done(true)

		utils.WriteManifest(Args.Context)
	},
}

var addEnvCmd = &cobra.Command{
	Use:     "env",
	Short:   "Add a new env variable to the manifest",
	Long:    "Add a new env variable to the manifest",
	Example: "   helm-manager add env [NAME]\n   helm-manager add env xyz" + USAGE_EXTRA,
	Args: ui.PositionalArgs([]ui.RequiredArg{
		ui.Arg[string]{
			Name:       "name",
			Ptr:        &Args.Name,
			Positional: true,
			Validator: types.MultiValidator(
				types.EnvValidator(),
				types.NotEqualValidator(
					types.ToStringer(`"%s" is already used in the manifest`),
					types.FutureFromStringers(types.FutureFromPtr(&Manifest.AllowedEnv)),
				),
			),
			UI: ui.PromptUiFunc[string]("Name"),
		},
	}, func(cmd *cobra.Command) {
		zap.S().Infof("* %s *", color.GreenString("Helm Manager Add Env"))
		ManifestExist(cmd)
	}),
	Run: func(cmd *cobra.Command, _ []string) {
		env := types.SelectableString(strings.ToUpper(Args.Name))
		Manifest.AllowedEnv = append(Manifest.AllowedEnv, env)

		if !Args.DryRun {
			utils.WriteManifest(Args.Context)
		} else {
			logger.Info("Dry run mode, not writing manifest")
		}

		logger.Infof("Added %s to the env variable whitelist", color.GreenString(string(env)))
	},
}

var addLocalChartCmd = &cobra.Command{
	Use:     "local-chart",
	Short:   "Add a local chart to the manifest",
	Long:    "Add a local chart to the manifest",
	Example: "   helm-manager add local-chart [PATH]\n   helm-manager add local-chart /path/to/chart" + USAGE_EXTRA,
	Args: ui.PositionalArgs([]ui.RequiredArg{
		ui.Arg[string]{
			Name:       "path",
			Ptr:        &Args.File,
			Positional: true,
			Validator: types.ValidatorFunction[string](func(pth string) error {
				chart := types.HelmChart{
					IsLocal:   true,
					LocalPath: pth,
				}

				if err := utils.ParseLocalChartYaml(&chart); err != nil {
					return fmt.Errorf("failed to parse Chart.yaml at %s", path.Join(pth, "Chart.yaml"))
				}

				charts, err := LocalChartsFuture.Get()
				if err != nil {
					return fmt.Errorf("failed to get local charts: %v", err)
				}

				existing := types.HelmChartMultiArray(charts).FindChart(chart.RepoName).FindVersion(chart.Version)
				if existing.Version != "" {
					return fmt.Errorf("chart already added to manifest")
				}

				return nil
			}),
			UI: ui.PromptUiFunc[string]("Path"),
		},
	}, func(cmd *cobra.Command) {
		zap.S().Infof("* %s *", color.GreenString("Helm Manager Add Local Chart"))
		ManifestExist(cmd)
	}),
	Run: func(cmd *cobra.Command, _ []string) {
		Manifest.LocalCharts = append(Manifest.LocalCharts, types.SelectableString(Args.File))

		if !Args.DryRun {
			utils.WriteManifest(Args.Context)
		} else {
			logger.Info("Dry run mode, not writing manifest")
		}

		logger.Infof("Added local chart \"%s\" to the manifest", color.GreenString(Args.File))
	},
}
