package cmd

import (
	"os"

	"github.com/fatih/color"
	"github.com/seventv/helm-manager/v2/cmd/args"
	"github.com/seventv/helm-manager/v2/cmd/ui"
	"github.com/seventv/helm-manager/v2/logger"
	"github.com/seventv/helm-manager/v2/types"
	"github.com/seventv/helm-manager/v2/utils"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var Args = args.Args
var Manifest = types.GlobalManifest

const USAGE_EXTRA = "\nAll arguments are optional, if not provided, you will be prompted to enter them.\nAll arguments can be passed as kwargs."

func init() {
	rootCmd.PersistentFlags().BoolVar(&Args.Debug, "debug", false, "Enable debug mode")
	rootCmd.PersistentFlags().BoolVar(&Args.NonInteractive, "term", false, "Disable interactive mode")
	rootCmd.PersistentFlags().BoolVar(&Args.DryRun, "dry-run", false, "Dry run mode ( your actions wont be saved or applied )")
	rootCmd.PersistentFlags().BoolVar(&Args.Confirm, "confirm", false, "Confirm your actions")
	rootCmd.PersistentFlags().StringVar(&Args.EnvFile, "env", ".env", "Environment file to use")

	wd, _ := os.Getwd()
	rootCmd.PersistentFlags().StringVarP(&Args.Context, "context", "c", wd, "Context to use, working directory")

	cobra.OnInitialize(func() {
		Args.Context = utils.MergeRelativePath(wd, Args.Context)

		err := utils.ReadManifest(Args.Context)
		if err != nil {
			logger.Fatal(err)
		}

		if Manifest.Exists {
			EnvMapFuture.GetOrPanic()
		}

		if Args.Debug {
			logger.SetDebug(true)
		}
	})
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "helm-manager",
	Short: "Helm-Manager is a tool to manage helm charts and k8s manifests",
	Long:  `A tool to manage helm charts and k8s manifests, allowing you to easily install, upgrade, and delete charts and manifests.`,
	Args:  ui.SubCommandRequired(cobra.NoArgs),
	CompletionOptions: cobra.CompletionOptions{
		HiddenDefaultCmd: true,
	},
	Run: func(cmd *cobra.Command, _ []string) {
		zap.S().Infof("* %s *\r", color.CyanString("Helm Manager"))

		cmds := make([]ui.SelectableCommand, 0, len(cmd.Commands()))
		for _, cmd := range cmd.Commands() {
			if !cmd.Hidden && cmd.Name() != "help" {
				cmds = append(cmds, ui.CmdSelectable(cmd))
			}
		}

		ui.RunSubCommand(cmd, cmds)
	},
}
