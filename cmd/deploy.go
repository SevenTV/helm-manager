package cmd

import (
	"os"

	"github.com/fatih/color"
	"github.com/seventv/helm-manager/cmd/args"
	"github.com/seventv/helm-manager/cmd/ui"
	"github.com/seventv/helm-manager/external"
	"github.com/seventv/helm-manager/logger"
	"github.com/seventv/helm-manager/types"
	"github.com/seventv/helm-manager/utils"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func init() {
	rootCmd.AddCommand(deployCmd)

	{
		deployCmd.AddCommand(deployReleaseCmd)
		deployReleaseCmd.Flags().StringVarP(&Args.Name, "name", "", "", "Release name")
		deployReleaseCmd.Flags().BoolVarP(&Args.DeployCmd.All, "all", "", false, "Deploy all releases")
	}

	{
		deployCmd.AddCommand(deploySingleCmd)
		deploySingleCmd.Flags().StringVarP(&Args.Name, "name", "", "", "single name")
		deploySingleCmd.Flags().BoolVarP(&Args.DeployCmd.All, "all", "", false, "Deploy all releases")
	}

	{
		deployCmd.AddCommand(deployRepoCmd)
	}

	{
		deployCmd.AddCommand(deployAllCmd)
	}
}

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy a release or single",
	Long:  "Deploy a release or single",
	Args:  ui.SubCommandRequired(cobra.NoArgs),
	Run: func(cmd *cobra.Command, _ []string) {
		zap.S().Infof("* %s *\r", color.BlueString("Helm Manager Deploy"))

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

func deployReleaseHelper(release types.ManifestRelease) error {
	data, err := utils.ReadFile(ReleasePath(release.Name))
	if err != nil {
		logger.Fatalf("failed to read release file: %s", err)
	}

	result, err := UpgradeDocument(data, types.HelmChart{
		Version:    release.Chart.Version,
		RepoName:   release.Chart.RepoName(),
		AppVersion: release.Chart.AppVersion,
	}, false)
	if err != nil {
		return err
	}

	if !Args.DryRun {
		if err = os.WriteFile(ReleasePath(release.Name), result.Document, 0644); err != nil {
			logger.Fatalf("failed to write release file: %s", err)
		}
	}

	return DeployRelease(release, result.EnvSubbedValues, result.EnvSubbedDocument)
}

var deployReleaseCmd = &cobra.Command{
	Use:   "release",
	Short: "Deploy a release",
	Long:  "Deploy a release",
	Args: ui.PositionalArgs([]ui.RequiredArg{
		ui.Arg[bool]{
			Name: "All",
			Ptr:  &Args.DeployCmd.All,
			Disabled: types.FutureFromFunc(func() bool {
				return Args.Name != ""
			}),
			UI: ui.PromptUiConfirmFunc("Do you want to deploy all releases", false),
		},
		ui.Arg[string]{
			Name:       "Name",
			Ptr:        &Args.Name,
			Positional: true,
			Disabled:   types.FutureFromPtr(&Args.DeployCmd.All),
			Validator: types.EqualValidator(types.ToStringer("\"%s\" is not a release in the manifest"), types.FutureFromFuncErr(func() ([]string, error) {
				names := make([]string, len(Manifest.Releases))
				for i, r := range Manifest.Releases {
					names[i] = r.Name
				}
				return names, nil
			})),
			UI: ui.PromptUiSelectorFunc[string]("Name", "", func(i int) error {
				Args.Name = Manifest.Releases[i].Name
				return nil
			}, types.FutureInterfacerArray[types.ManifestRelease, types.Selectable](types.FutureFromPtr(&Manifest.Releases))),
		},
	}, func(cmd *cobra.Command) {
		zap.S().Infof("* %s *", color.BlueString("Helm Manager Deploy Release"))
	}),
	Run: func(cmd *cobra.Command, _ []string) {
		if Args.DeployCmd.All {
			for _, release := range Manifest.Releases {
				if err := deployReleaseHelper(release); err != nil {
					logger.Fatalf("failed to deploy release: %s", err)
				}
			}
		} else {
			release := Manifest.ReleaseByName(Args.Name)
			if err := deployReleaseHelper(release); err != nil {
				logger.Fatalf("failed to deploy release: %s", err)
			}
		}

		logger.Info("releases deployed")
	},
}

func deploySingleHelper(single types.ManifestSingle) error {
	data, err := utils.ReadFile(SinglePath(single.Name))
	if err != nil {
		logger.Fatalf("failed to read release file: %s", err)
	}

	return DeploySingle(single, data)
}

func deployReposHelper() {
	repos, err := HelmRepoFuture.Get()
	if err != nil {
		logger.Fatalf("failed to get helm repos: %s", err)
	}

	repoMp := map[string]string{}
	for _, repo := range repos {
		repoMp[repo.Name] = repo.URL
	}

	var resp []byte
	for _, repo := range Manifest.Repos {
		if _, ok := repoMp[repo.Name]; !ok {
			resp, err = external.Helm.AddRepo(repo.Name, repo.URL)
			if err != nil {
				logger.Fatalf("failed to add repo: %s\n%s", err, resp)
			}
		} else {
			if repoMp[repo.Name] != repo.URL {
				if args.Args.Force {
					resp, err = external.Helm.RemoveRepo(repo.Name)
					if err != nil {
						logger.Fatalf("failed to remove repo: %s\n%s", err, resp)
					}

					resp, err = external.Helm.AddRepo(repo.Name, repo.URL)
					if err != nil {
						logger.Fatalf("failed to add repo: %s\n%s", err, resp)
					}
				} else {
					logger.Warnf("repo %s already exists with a different url", repo.Name)
				}
			}
		}
	}

	_, err = UpdateHelmRepoFuture.Get()
	if err != nil {
		logger.Fatalf("failed to update helm repos: %s", err)
	}

	logger.Infof("repos deployed")
}

var deploySingleCmd = &cobra.Command{
	Use:   "single",
	Short: "Deploy a single",
	Long:  "Deploy a single",
	Args: ui.PositionalArgs([]ui.RequiredArg{
		ui.Arg[bool]{
			Name: "All",
			Ptr:  &Args.DeployCmd.All,
			Disabled: types.FutureFromFunc(func() bool {
				return Args.Name != ""
			}),
			UI: ui.PromptUiConfirmFunc("Do you want to deploy all singles", false),
		},
		ui.Arg[string]{
			Name:       "Name",
			Ptr:        &Args.Name,
			Positional: true,
			Disabled:   types.FutureFromPtr(&Args.DeployCmd.All),
			Validator: types.EqualValidator(types.ToStringer("\"%s\" is not a single in the manifest"), types.FutureFromFuncErr(func() ([]string, error) {
				names := make([]string, len(Manifest.Singles))
				for i, r := range Manifest.Releases {
					names[i] = r.Name
				}
				return names, nil
			})),
			UI: ui.PromptUiSelectorFunc[string]("Name", "", func(i int) error {
				Args.Name = Manifest.Singles[i].Name
				return nil
			}, types.FutureInterfacerArray[types.ManifestSingle, types.Selectable](types.FutureFromPtr(&Manifest.Singles))),
		},
	}, func(cmd *cobra.Command) {
		zap.S().Infof("* %s *", color.BlueString("Helm Manager Deploy Single"))
	}),
	Run: func(cmd *cobra.Command, _ []string) {
		if Args.DeployCmd.All {
			for _, single := range Manifest.Singles {
				if err := deploySingleHelper(single); err != nil {
					logger.Fatalf("failed to deploy release: %s", err)
				}
			}
		} else {
			single := Manifest.SingleByName(Args.Name)
			if err := deploySingleHelper(single); err != nil {
				logger.Fatalf("failed to deploy release: %s", err)
			}
		}

		logger.Infof("singles deployed")
	},
}

var deployRepoCmd = &cobra.Command{
	Use:   "repo",
	Short: "Deploy a repo",
	Long:  "Deploy a repo",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, _ []string) {
		zap.S().Infof("* %s *", color.BlueString("Helm Manager Deploy Repo"))

		ManifestExist(cmd)

		deployReposHelper()
	},
}

var deployAllCmd = &cobra.Command{
	Use:   "all",
	Short: "Deploy all releases and singles",
	Long:  "Deploy all releases and singles",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, _ []string) {
		zap.S().Infof("* %s *", color.BlueString("Helm Manager Deploy All"))

		ManifestExist(cmd)

		deployReposHelper()

		for _, release := range Manifest.Releases {
			if err := deployReleaseHelper(release); err != nil {
				logger.Fatalf("failed to deploy release: %s", err)
			}
		}

		for _, single := range Manifest.Singles {
			if err := deploySingleHelper(single); err != nil {
				logger.Fatalf("failed to deploy release: %s", err)
			}
		}

		logger.Infof("all deployed")
	},
}