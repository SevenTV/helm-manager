package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/seventv/helm-manager/v2/cmd/ui"
	"github.com/seventv/helm-manager/v2/external"
	"github.com/seventv/helm-manager/v2/logger"
	"github.com/seventv/helm-manager/v2/types"
	"github.com/seventv/helm-manager/v2/utils"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func init() {
	rootCmd.AddCommand(importCmd)

	{
		importCmd.AddCommand(importReleaseCmd)
		importReleaseCmd.Flags().StringVarP(&Args.Name, "name", "", "", "Name of the release to import")
		importReleaseCmd.Flags().StringVarP(&Args.Namespace, "namespace", "n", "", "Namespace of the release to import")
		importReleaseCmd.Flags().BoolVarP(&Args.ImportCmd.All, "all", "", false, "Import all releases")
		importReleaseCmd.Flags().BoolVarP(&Args.Force, "force", "", false, "Overwrite existing release files")
	}

	{
		importCmd.AddCommand(importRepoCmd)
		importRepoCmd.Flags().StringVarP(&Args.Name, "name", "", "", "Name of the release to import")
		importRepoCmd.Flags().BoolVarP(&Args.ImportCmd.All, "all", "", false, "Import all releases")
	}
}

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import releases, repositories from other sources",
	Long:  "Import releases, repositories from other sources",
	Args:  ui.SubCommandRequired(cobra.NoArgs),
	Run: func(cmd *cobra.Command, args []string) {
		zap.S().Infof("* %s *\r", color.MagentaString("Helm Manager Import"))

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

var importReleaseCmd = &cobra.Command{
	Use:   "release",
	Short: "Import releases from a cluster",
	Long:  "Import releases from a cluster",
	Args: ui.PositionalArgs([]ui.RequiredArg{
		ui.Arg[bool]{
			Name: "all",
			Ptr:  &Args.ImportCmd.All,
			Disabled: types.FutureFromFunc(func() bool {
				return Args.Name == "" && Args.Namespace == ""
			}),
			UI: ui.PromptUiConfirmFunc("Import all releases from the cluster", false),
		},
		ui.Arg[string]{
			Name:     "name",
			Ptr:      &Args.Name,
			Disabled: types.FutureFromPtr(&Args.ImportCmd.All),
			Validator: types.MultiValidator(
				types.EqualValidator(
					types.StringerFunc(func() string {
						if Args.Namespace == "" {
							return "The release \"%s\" was not found in the cluster"
						} else {
							return fmt.Sprintf("The release \"%s\" in namespace \"%s\" was not found in the cluster", "%s", Args.Namespace)
						}
					}),
					types.FutureFromFuncErr(func() ([]string, error) {
						values, err := HelmReleaseChartFuture.Get()
						if err != nil {
							return nil, err
						}

						ret := []string{}
						for _, v := range values {
							if Args.Namespace != "" && v.Namespace != Args.Namespace {
								continue
							}

							ret = append(ret, v.Name, fmt.Sprintf("%s/%s", v.Namespace, v.Name))
						}

						return ret, nil
					}),
				),
				types.NotEqualValidator(
					types.ToStringer("The release name (%s) is already in use in the manifest"),
					types.FutureFromFunc(func() []string {
						names := make([]string, len(Manifest.Releases))
						for i, v := range Manifest.Releases {
							names[i] = v.Name
						}

						return names
					}),
				),
			),
			Positional: true,
			UI: ui.PromptUiSelectorFunc[string]("Release", "", func(i int) error {
				releases, err := HelmReleaseChartFuture.Get()
				if err != nil {
					return err
				}

				idx := 0
				for _, v := range releases {
					if Args.Namespace != "" && strings.ToLower(v.Namespace) != strings.ToLower(Args.Namespace) {
						continue
					}

					if _, idx := Manifest.ReleaseIdxByName(v.Name); idx != -1 {
						continue
					}

					if idx == i {
						Args.Name = v.Name
						if Args.Namespace == "" {
							Args.Namespace = v.Namespace
						}
						idx = -1
						break
					}
					idx++
				}

				if idx != -1 {
					return fmt.Errorf("Could not find release")
				}

				return nil
			}, types.FutureFromFuncErr(func() ([]types.Selectable, error) {
				releases, err := HelmReleaseChartFuture.Get()
				if err != nil {
					return nil, err
				}

				ret := []types.Selectable{}
				for _, v := range releases {
					if Args.Namespace != "" && strings.ToLower(v.Namespace) != strings.ToLower(Args.Namespace) {
						continue
					}

					if _, idx := Manifest.ReleaseIdxByName(v.Name); idx != -1 {
						continue
					}

					ret = append(ret, v)
				}

				if len(ret) == 0 {
					return nil, fmt.Errorf("No unimported releases found")
				}

				return ret, nil
			})),
			Callback: func(name string) error {
				if Args.Namespace == "" {
					name := strings.Split(name, "/")
					if len(name) == 2 {
						Args.Namespace = name[0]
						Args.Name = name[1]
					}
				}

				return nil
			},
		},
		ui.Arg[string]{
			Name:     "namespace",
			Ptr:      &Args.Namespace,
			Disabled: types.FutureFromPtr(&Args.ImportCmd.All),
			Validator: types.MultiValidator(
				types.OptionalEmptyValidator[string](),
				types.EqualValidator(
					types.StringerFunc(func() string {
						return fmt.Sprintf("The release \"%s\" in namespace \"%s\" was not found in the cluster", Args.Name, "%s")
					}),
					types.FutureFromFuncErr(func() ([]string, error) {
						values, err := HelmReleaseChartFuture.Get()
						if err != nil {
							return nil, err
						}

						ret := []string{}
						for _, v := range values {
							if v.Name != Args.Name {
								continue
							}

							ret = append(ret, v.Namespace)
						}

						return ret, nil
					}),
				),
			),
			Positional: true,
			UI: ui.PromptUiSelectorFunc[string]("Release", "", func(i int) error {
				releases, err := HelmReleaseChartFuture.Get()
				if err != nil {
					return err
				}

				idx := 0
				for _, v := range releases {
					if v.Name != Args.Name {
						continue
					}

					if idx == i {
						Args.Namespace = v.Namespace
						idx = -1
						break
					}

					idx++
				}

				if idx != -1 {
					return fmt.Errorf("Could not find release")
				}

				return nil
			}, types.FutureFromFuncErr(func() ([]types.Selectable, error) {
				releases, err := HelmReleaseFuture.Get()
				if err != nil {
					return nil, err
				}

				ret := []types.Selectable{}
				for _, v := range releases {
					if v.Name != Args.Name {
						continue
					}

					ret = append(ret, v)
				}

				return ret, nil
			})),
			Callback: func(name string) error {
				Args.Name = strings.ToLower(Args.Name)
				Args.Namespace = strings.ToLower(Args.Namespace)
				return nil
			},
		},
	}, func(cmd *cobra.Command) {
		zap.S().Infof("* %s *", color.MagentaString("Helm Manager Import Release"))

		ManifestExist(cmd)
	}),
	Run: func(cmd *cobra.Command, args []string) {
		releases, err := HelmReleaseChartFuture.Get()
		if err != nil {
			logger.Fatalf("Failed to get releases: %s", err)
		}

		importRelease := func(release HelmReleaseChart) {
			releaseValues, err := external.Helm.GetReleaseValues(release.HelmRelease)
			if err != nil {
				logger.Fatalf("Could not get release values: %s", err)
			}

			values, err := utils.MarshalYaml(releaseValues)
			if err != nil {
				logger.Fatalf("Could not marshal release values: %s", err)
			}

			manifestRelease := types.ManifestRelease{
				Name:      release.Name,
				Namespace: release.Namespace,
				Chart: types.ManifestChart{
					Name:       release.Chart.Name(),
					Version:    release.Chart.Version,
					AppVersion: release.Chart.AppVersion,
					Repo:       release.Chart.Repo(),
				},
			}

			Manifest.Releases = append(Manifest.Releases, manifestRelease)

			result, err := UpgradeDocument(values, release.Chart, true)
			if err != nil {
				logger.Fatalf("Could not upgrade document: %s", err)
			}

			if !Args.Force {
				if _, err := os.Stat(ReleasePath(release.Name)); err == nil {
					logger.Fatal("Release file already exists, use --force to overwrite")
				}
			}

			if !Args.DryRun {
				if err = os.WriteFile(ReleasePath(release.Name), result.Document, 0644); err != nil {
					logger.Fatalf("failed to write release file: %s", err)
				}
				utils.WriteManifest(Args.Context)
			} else {
				logger.Info("Dry run, not writing manifest or release file")
			}

			logger.Infof("Successfully imported release \"%s\" from namespace \"%s\"", release.Name, release.Namespace)
		}

		if Args.ImportCmd.All {
			for _, v := range releases {
				if _, idx := Manifest.ReleaseIdxByName(v.Name); idx == -1 {
					importRelease(v)
				} else {
					logger.Infof("Release \"%s\" already exists in manifest, skipping", v.Name)
				}
			}
		} else {
			var release HelmReleaseChart
			for _, v := range releases {
				if v.Name == Args.Name && strings.ToLower(v.Namespace) == Args.Namespace {
					release = v
					break
				}
			}

			if release.Name == "" {
				logger.Fatalf("Could not find release \"%s\" in namespace \"%s\"", Args.Name, Args.Namespace)
			}

			importRelease(release)
		}
	},
}

var importRepoCmd = &cobra.Command{
	Use:   "repo",
	Short: "Import a helm repository",
	Args: ui.PositionalArgs([]ui.RequiredArg{
		ui.Arg[string]{
			Name:       "Name",
			Ptr:        &Args.Name,
			Positional: true,
			Validator: types.EqualValidator(
				types.ToStringer("\"%s\" is not a helm repo on the system"),
				types.FutureFromFuncErr(func() ([]string, error) {
					repos, err := HelmRepoFuture.Get()
					if err != nil {
						return nil, err
					}

					ret := []string{}
					for _, v := range repos {
						if Manifest.RepoByName(v.Name).Name != "" {
							continue
						}

						ret = append(ret, v.Name)
					}

					return ret, nil
				}),
			),
			UI: ui.PromptUiSelectorFunc[string]("Name", "", func(i int) error {
				repos, err := HelmRepoFuture.Get()
				if err != nil {
					return err
				}
				idx := 0
				for _, v := range repos {
					if Manifest.RepoByName(v.Name).Name != "" {
						continue
					}

					if idx == i {
						Args.Name = v.Name
						return nil
					}

					idx++
				}

				return fmt.Errorf("invalid index")

			}, types.FutureFromFuncErr(func() ([]types.Selectable, error) {
				repos, err := HelmRepoFuture.Get()
				if err != nil {
					return nil, err
				}

				ret := []types.Selectable{}
				for _, v := range repos {
					if Manifest.RepoByName(v.Name).Name != "" {
						continue
					}

					ret = append(ret, v)
				}

				if len(ret) == 0 {
					return nil, fmt.Errorf("no repos to import")
				}

				return ret, nil
			})),
			Callback: func(a string) error {
				Args.Name = strings.ToLower(a)
				return nil
			},
		},
	}, func(cmd *cobra.Command) {
		zap.S().Infof("* %s *", color.MagentaString("Helm Manager Import Repo"))

		ManifestExist(cmd)
	}),
	Run: func(cmd *cobra.Command, args []string) {
		repos, err := HelmRepoFuture.Get()
		if err != nil {
			logger.Fatalf("Could not get helm repos: %s", err)
		}

		var repo types.HelmRepo
		for _, v := range repos {
			if v.Name == Args.Name {
				repo = v
				break
			}
		}

		if repo.Name == "" {
			logger.Fatalf("Could not find repo \"%s\"", Args.Name)
		}

		Manifest.Repos = append(Manifest.Repos, types.ManifestRepo{
			Name: repo.Name,
			URL:  repo.URL,
		})

		if !Args.DryRun {
			utils.WriteManifest(Args.Context)
		} else {
			logger.Info("Dry run, not writing manifest")
		}

		logger.Infof("Successfully imported repo \"%s\"", repo.Name)
	},
}
