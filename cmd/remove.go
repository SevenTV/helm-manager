package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/seventv/helm-manager/cmd/ui"
	"github.com/seventv/helm-manager/external"
	"github.com/seventv/helm-manager/logger"
	"github.com/seventv/helm-manager/types"
	"github.com/seventv/helm-manager/utils"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func init() {
	rootCmd.AddCommand(removeCmd)

	{
		removeCmd.AddCommand(removeRepoCmd)
		removeRepoCmd.Flags().StringVar(&Args.Name, "name", "", "Name of the repo")
		removeRepoCmd.Flags().BoolVar(&Args.Delete, "delete", false, "Delete the repo from the system")
	}

	{
		removeCmd.AddCommand(removeSingleCmd)
		removeSingleCmd.Flags().StringVar(&Args.Name, "name", "", "Name of the single")
		removeSingleCmd.Flags().BoolVar(&Args.Delete, "delete", false, "Delete the single file from the system")
		removeSingleCmd.Flags().BoolVar(&Args.Deploy, "deploy", false, "Apply deletion of the single to the cluster")
	}

	{
		removeCmd.AddCommand(removeReleaseCmd)
		removeReleaseCmd.Flags().StringVar(&Args.Name, "name", "", "Name of the release")
		removeReleaseCmd.Flags().BoolVar(&Args.Delete, "delete", false, "Delete the release file from the system")
		removeReleaseCmd.Flags().BoolVar(&Args.Deploy, "deploy", false, "Apply deletion of the release to the cluster")
	}

	{
		removeCmd.AddCommand(removeEnvCmd)
		removeEnvCmd.Flags().StringVar(&Args.Name, "name", "", "Name of the env variable to be whitelisted")
	}
}

var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove an existing release, single, repo or env variable from the manifest",
	Long:  "Remove an existing release, single, repo or env variable from the manifest",
	Args:  ui.SubCommandRequired(cobra.NoArgs),
	Run: func(cmd *cobra.Command, args []string) {
		zap.S().Infof("* %s *\r", color.RedString("Helm Manager Remove"))

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

var removeReleaseCmd = &cobra.Command{
	Use:   "release",
	Short: "Remove a release from the manifest",
	Long:  "Remove a release from the manifest",
	Args: ui.PositionalArgs([]ui.RequiredArg{
		ui.Arg[string]{
			Name:       "name",
			Ptr:        &Args.Name,
			Positional: true,
			Validator:  types.EqualValidator(types.ToStringer(`"%s" is not a release name in the manifest`), types.FutureFromStringers(types.FutureFromPtr(&Manifest.Releases))),
			UI: ui.PromptUiSelectorFunc[string]("Release", "Which release do you want to remove?", func(i int) error {
				Args.Name = Manifest.Releases[i].Name
				return nil
			}, types.FutureInterfacerArray[types.ManifestRelease, types.Selectable](types.FutureFromPtr(&Manifest.Releases))),
			Callback: func(s string) error {
				Args.Name = strings.ToLower(s)

				return nil
			},
		},
		ui.Arg[bool]{
			Name:       "delete",
			Ptr:        &Args.Delete,
			Positional: false,
			UI:         ui.PromptUiConfirmFunc("Do you want to delete the release file", false),
		},
		ui.Arg[bool]{
			Name:       "deploy",
			Ptr:        &Args.Deploy,
			Positional: false,
			UI:         ui.PromptUiConfirmFunc("Do you want to apply the deletion of the release to the cluster", false),
		},
		ui.Arg[bool]{
			Name: "confirm",
			Ptr:  &Args.Confirm,
			Disabled: types.FutureFromFunc(func() bool {
				return (!Args.Delete && !Args.Deploy) || Args.DryRun
			}),
			UI: ui.PromptUiConfirmFunc("Are you sure you want to remove this release", false),
			Callback: func(b bool) error {
				if !b {
					return fmt.Errorf("Aborted")
				}

				return nil
			},
		},
	}, func(cmd *cobra.Command) {
		zap.S().Infof("* %s *", color.RedString("Helm Manager Remove Release"))
		ManifestExist(cmd)

		if len(Manifest.Releases) == 0 {
			logger.Fatal("no releases added to the manifest")
		}
	}),
	Run: func(cmd *cobra.Command, args []string) {
		var (
			release types.ManifestRelease
			i       int
		)
		for i, release = range Manifest.Releases {
			if strings.ToLower(string(release.Name)) == Args.Name {
				Manifest.Releases = append(Manifest.Releases[:i], Manifest.Releases[i+1:]...)
				break
			}
		}

		if Args.Delete {
			if err := os.Remove(ReleasePath(release.Name)); err != nil {
				logger.Fatal("failed to delete release file", zap.Error(err))
			}
		}

		if Args.Deploy {
			if Args.DryRun {
				logger.Info("Running in dry run mode, not actually deleting the release")
			}

			done := utils.Loader(utils.LoaderOptions{
				FetchingText: "Deleting release",
				SuccessText:  "Deleted release",
				FailureText:  "Failed to delete release",
			})

			resp, err := external.Helm.UninstallRelease(release.Name, release.Namespace, Args.DryRun)
			done(err == nil)
			if err != nil {
				logger.Fatalf("Failed to delete release: %s\n%s", err, resp)
			}
		}

		if !Args.DryRun {
			utils.WriteManifest(Args.Context)
		} else {
			logger.Info("Dry run mode, not writing manifest")
		}

		logger.Infof("Removed %s release from the manifest", color.RedString(string(release.Name)))
	},
}

var removeSingleCmd = &cobra.Command{
	Use:   "single",
	Short: "Remove a new single from the manifest",
	Long:  "Remove a new single from the manifest",
	Args: ui.PositionalArgs([]ui.RequiredArg{
		ui.Arg[string]{
			Name:       "name",
			Ptr:        &Args.Name,
			Positional: true,
			Validator:  types.EqualValidator(types.ToStringer(`"%s" is not a single name in the manifest`), types.FutureFromStringers(types.FutureFromPtr(&Manifest.Singles))),
			UI: ui.PromptUiSelectorFunc[string]("Single", "Which single do you want to remove?", func(i int) error {
				Args.Name = string(Manifest.Singles[i].Name)
				return nil
			}, types.FutureInterfacerArray[types.ManifestSingle, types.Selectable](types.FutureFromPtr(&Manifest.Singles))),
		},
		ui.Arg[bool]{
			Name: "delete",
			Ptr:  &Args.Delete,
			UI:   ui.PromptUiConfirmFunc("Do you want to delete the single file", false),
		},
		ui.Arg[bool]{
			Name: "deploy",
			Ptr:  &Args.Deploy,
			UI:   ui.PromptUiConfirmFunc("Do you want to apply the deletion of the single to the cluster", false),
		},
		ui.Arg[bool]{
			Name: "confirm",
			Ptr:  &Args.Confirm,
			Disabled: types.FutureFromFunc(func() bool {
				return (!Args.Delete && !Args.Deploy) || Args.DryRun
			}),
			UI: ui.PromptUiConfirmFunc("Are you sure you want to remove this single", false),
			Callback: func(b bool) error {
				if !b {
					return fmt.Errorf("Aborted")
				}

				return nil
			},
		},
	}, func(cmd *cobra.Command) {
		zap.S().Infof("* %s *", color.RedString("Helm Manager Remove Single"))
		ManifestExist(cmd)

		if len(Manifest.Singles) == 0 {
			logger.Fatal("no singles added to the manifest")
		}
	}),
	Run: func(cmd *cobra.Command, _ []string) {
		singleName := strings.ToLower(Args.Name)

		var (
			single types.ManifestSingle
			i      int
		)
		for i, single = range Manifest.Singles {
			if strings.ToLower(string(single.Name)) == singleName {
				Manifest.Singles = append(Manifest.Singles[:i], Manifest.Singles[i+1:]...)
				break
			}
		}

		if Args.Delete {
			done := utils.Loader(utils.LoaderOptions{
				FetchingText: "Deleting single",
				SuccessText:  "Deleted single",
				FailureText:  "Failed to delete single",
			})

			values, err := os.ReadFile(SinglePath(single.Name))
			if err != nil {
				done(false)
				logger.Fatal("failed to read single file", zap.Error(err))
			}

			resp, err := external.Kubectl.Delete(values, Args.Namespace, Args.DryRun)
			done(err == nil)
			if err != nil {
				logger.Fatalf("Failed to delete single: %s\n%s", err, resp)
			}
		}

		if !Args.DryRun {
			utils.WriteManifest(Args.Context)
		} else {
			logger.Info("Dry run mode, not writing manifest")
		}

		logger.Infof("Removed %s single from the manifest", color.RedString(string(singleName)))
	},
}

var removeRepoCmd = &cobra.Command{
	Use:     "repo",
	Short:   "Remove a new repo to from manifest",
	Long:    "Remove a new repo to from manifest",
	Example: "helm-manager remove repo [REPO] [URL]",
	Args: ui.PositionalArgs([]ui.RequiredArg{
		ui.Arg[string]{
			Name:       "name",
			Ptr:        &Args.Name,
			Positional: true,
			Validator:  types.EqualValidator(types.ToStringer(`"%s" is not a repo name in the manifest`), types.FutureFromStringers(types.FutureFromPtr(&Manifest.Repos))),
			UI: ui.PromptUiSelectorFunc[string]("Repo", "Which repo do you want to remove?", func(i int) error {
				Args.Name = string(Manifest.Repos[i].Name)
				return nil
			}, types.FutureInterfacerArray[types.ManifestRepo, types.Selectable](types.FutureFromPtr(&Manifest.Repos))),
		},
		ui.Arg[bool]{
			Name: "delete",
			Ptr:  &Args.Delete,
			UI:   ui.PromptUiConfirmFunc("Do you want to delete this repo from the system", false),
		},
		ui.Arg[bool]{
			Name: "confirm",
			Ptr:  &Args.Confirm,
			UI:   ui.PromptUiConfirmFunc("Are you sure you want to remove this repo", false),
			Disabled: types.FutureFromFunc(func() bool {
				return !Args.Delete || Args.DryRun
			}),
			Callback: func(b bool) error {
				if !b {
					return fmt.Errorf("Aborted")
				}

				return nil
			},
		},
	}, func(cmd *cobra.Command) {
		zap.S().Infof("* %s *", color.RedString("Helm Manager Remove Repo"))
		ManifestExist(cmd)

		if len(Manifest.Repos) == 0 {
			logger.Fatal("no repos added to the manifest")
		}
	}),
	Run: func(cmd *cobra.Command, _ []string) {
		repo := strings.ToLower(Args.Name)

		for i, r := range Manifest.Repos {
			if strings.ToLower(string(r.Name)) == repo {
				Manifest.Repos = append(Manifest.Repos[:i], Manifest.Repos[i+1:]...)
				break
			}
		}

		if Args.Delete {
			logger.Infof("Deleting %s repo", color.RedString(string(repo)))
			if !Args.DryRun {
				if resp, err := external.Helm.RemoveRepo(string(repo)); err != nil {
					logger.Error("failed to delete repo %v\n%s", err, resp)
				}
			} else {
				logger.Info("Dry run mode, not deleting repo")
			}
		}

		if !Args.DryRun {
			utils.WriteManifest(Args.Context)
		} else {
			logger.Info("Dry run mode, not writing manifest")
		}

		logger.Infof("Removed %s repo from the manifest", color.RedString(string(repo)))
	},
}

var removeEnvCmd = &cobra.Command{
	Use:   "env",
	Short: "Remove a new env variable to from manifest",
	Long:  "Remove a new env variable to from manifest",
	Args: ui.PositionalArgs([]ui.RequiredArg{
		ui.Arg[string]{
			Name:       "name",
			Ptr:        &Args.Name,
			Positional: true,
			Validator:  types.EqualValidator(types.ToStringer(`"%s" is not a whitelisted env variable in the manifest`), types.FutureFromStringers(types.FutureFromPtr(&Manifest.AllowedEnv))),
			UI: ui.PromptUiSelectorFunc[string]("Env", "Which environment veriable do you want to remove?", func(i int) error {
				Args.Name = string(Manifest.AllowedEnv[i])
				return nil
			}, types.FutureInterfacerArray[types.SelectableString, types.Selectable](types.FutureFromPtr(&Manifest.AllowedEnv))),
		},
	}, func(cmd *cobra.Command) {
		zap.S().Infof("* %s *", color.RedString("Helm Manager Remove Env"))
		ManifestExist(cmd)

		if len(Manifest.AllowedEnv) == 0 {
			logger.Fatal("no env variables added to the manifest")
		}
	}),
	Run: func(cmd *cobra.Command, _ []string) {
		env := strings.ToLower(Args.Name)

		for i, e := range Manifest.AllowedEnv {
			if strings.ToLower(string(e)) == env {
				Manifest.AllowedEnv = append(Manifest.AllowedEnv[:i], Manifest.AllowedEnv[i+1:]...)
				break
			}
		}

		if !Args.DryRun {
			utils.WriteManifest(Args.Context)
		} else {
			logger.Info("Dry run mode, not writing manifest")
		}

		logger.Infof("Removed %s to the env variable whitelist", color.RedString(string(env)))
	},
}
