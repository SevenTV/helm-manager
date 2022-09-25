package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/seventv/helm-manager/cmd/ui"
	"github.com/seventv/helm-manager/logger"
	"github.com/seventv/helm-manager/types"
	"github.com/seventv/helm-manager/utils"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func init() {
	rootCmd.AddCommand(updateCmd)

	{
		updateCmd.Flags().StringVar(&Args.Name, "name", "", "Update the name of the release")
		updateCmd.Flags().StringVar(&Args.UpdateCmd.Version, "version", "", "Version to update to")
		updateCmd.Flags().BoolVar(&Args.Deploy, "deploy", false, "Deploy the updated release to the cluster")
		updateCmd.Flags().BoolVar(&Args.UpdateCmd.List, "list", false, "List all available versions")
	}
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update an existing release, or list available versions",
	Long:  "Update an existing release, or list available versions",
	Args: ui.PositionalArgs([]ui.RequiredArg{
		ui.Arg[bool]{
			Name: "list",
			Ptr:  &Args.UpdateCmd.List,
			UI:   ui.PromptUiConfirmFunc("Do you want to list all updates", true),
		},
		ui.Arg[string]{
			Name:       "name",
			Ptr:        &Args.Name,
			Positional: true,
			Disabled:   types.FutureFromPtr(&Args.UpdateCmd.List),
			Validator:  types.EqualValidator(types.ToStringer(`"%s" is not a release in the manifest`), types.FutureFromStringers(types.FutureFromPtr(&Manifest.Releases))),
			UI: ui.PromptUiSelectorFunc[string]("Release", "Which release would you like to update", func(i int) error {
				Args.Name = Manifest.Releases[i].Name
				return nil
			}, types.FutureInterfacerArray[types.ManifestRelease, types.Selectable](types.FutureFromPtr(&Manifest.Releases))),
			Callback: func(s string) error {
				Args.Name = strings.ToLower(s)

				return nil
			},
		},
		ui.Arg[string]{
			Name:       "version",
			Ptr:        &Args.UpdateCmd.Version,
			Positional: true,
			Disabled:   types.FutureFromPtr(&Args.UpdateCmd.List),
			Validator: types.EqualValidator(types.StringerFunc(func() string {
				return fmt.Sprintf("\"%s\" is not a valid version for the chart \"%s\"", "%s", Manifest.ReleaseByName(Args.Name).Chart.RepoName())
			}), types.FutureFromFuncErr(func() ([]string, error) {
				charts, err := HelmChartsFuture.Get()
				if err != nil {
					return nil, err
				}

				release := Manifest.ReleaseByName(Args.Name)
				for _, chart := range charts {
					if release.Chart.RepoName() == chart.RepoName {
						versions := make([]string, len(chart.Versions))
						for i, version := range chart.Versions {
							versions[i] = version.Version
						}

						return versions, nil
					}
				}
				return nil, fmt.Errorf("no versions found for %s", release.Chart.RepoName())
			})),
			UI: ui.PromptUiSelectorFunc[string]("Version", "Which version would you like to update to", func(i int) error {
				versions, err := HelmChartsFuture.Get()
				if err != nil {
					return err
				}

				release := Manifest.ReleaseByName(Args.Name)
				for _, chart := range versions {
					if release.Chart.RepoName() == chart.RepoName {
						Args.UpdateCmd.Version = chart.Versions[i].Version
						return nil
					}
				}

				return fmt.Errorf("no versions found for %s", release.Chart.RepoName())
			}, types.FutureInterfacerArray[types.HelmChartMultiVersion, types.Selectable](types.FutureFromFuncErr(func() ([]types.HelmChartMultiVersion, error) {
				charts, err := HelmChartsFuture.Get()
				if err != nil {
					return nil, err
				}

				release := Manifest.ReleaseByName(Args.Name)
				for _, chart := range charts {
					if release.Chart.RepoName() == chart.RepoName {
						return chart.Versions, nil
					}
				}
				return nil, fmt.Errorf("no versions found for %s", release.Chart.RepoName())
			}))),
		},
		ui.Arg[bool]{
			Name:     "deploy",
			Ptr:      &Args.Deploy,
			Disabled: types.FutureFromPtr(&Args.UpdateCmd.List),
			UI:       ui.PromptUiConfirmFunc("Do you want to deploy the update to the cluster", true),
		},
	}, func(cmd *cobra.Command) {
		zap.S().Infof("* %s *", color.YellowString("Helm Manager Update"))

		ManifestExist(cmd)

		if len(Manifest.Releases) == 0 {
			logger.Fatal("no releases found in manifest")
		}
	}),
	Run: func(cmd *cobra.Command, _ []string) {
		charts, err := HelmChartsFuture.Get()
		if err != nil {
			logger.Fatalf("failed to get helm charts: %s", err)
		}
		if Args.UpdateCmd.List {
			mapChart := map[string]types.HelmChartMulti{}
			for _, chart := range charts {
				mapChart[chart.RepoName] = chart
			}

			updates := false
			for _, release := range Manifest.Releases {
				chart, ok := mapChart[release.Chart.RepoName()]
				if !ok {
					logger.Warnf("no chart found for release %s", release.Name)
					continue
				}

				if chart.Version != release.Chart.Version {
					updates = true
					appVersionUpdate := ""
					if chart.AppVersion != release.Chart.AppVersion {
						appVersionUpdate = fmt.Sprintf("\n   App version from %s to %s", color.RedString(release.Chart.AppVersion), color.GreenString(chart.AppVersion))
					}
					logger.Infof(
						"Release %s (%s) has an update available from %s to %s%s\n   You can upgrade by typing `%s`",
						color.CyanString(release.Name),
						color.YellowString(release.Chart.RepoName()),
						color.RedString(release.Chart.Version),
						color.GreenString(chart.Version),
						appVersionUpdate,
						color.YellowString("helm-manager update %s --version %s", release.Name, chart.Version),
					)
				}
			}

			if !updates {
				logger.Info("No updates available")
			}

			return
		}

		release, idx := Manifest.ReleaseIdxByName(Args.Name)

		multiChart := types.HelmChartMultiArray(charts).FindChart(release.Chart.RepoName())
		chart := multiChart.FindVersion(Args.UpdateCmd.Version)
		if chart.Version == "" {
			logger.Fatalf("no version %s found for %s", Args.UpdateCmd.Version, release.Chart.RepoName())
		}

		data, err := utils.ReadFile(ReleasePath(Args.Name))
		if err != nil {
			logger.Fatalf("failed to read release file: %s", err)
		}

		multiChart.HelmChart = types.HelmChart(chart)
		result, err := UpgradeDocument(data, multiChart, true)
		if err != nil {
			logger.Fatal(err)
		}

		release.Chart = types.ManifestChart{
			Name:       chart.Name(),
			Version:    chart.Version,
			AppVersion: chart.AppVersion,
			Repo:       chart.Repo(),
		}

		Manifest.Releases[idx] = release

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

		logger.Infof("Release %s (%s) updated", Args.Name, release.Chart.RepoName())
	},
}
