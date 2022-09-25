package cmd

import (
	"os"
	"path"

	"github.com/fatih/color"
	"github.com/seventv/helm-manager/v2/logger"
	"github.com/seventv/helm-manager/v2/utils"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func init() {
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new manifest",
	Long:  `Initialize a new manifest`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, _ []string) {
		zap.S().Infof("* %s *", color.BlueString("Helm Manager Init"))

		if Manifest.Exists {
			logger.Fatal("manifest already exists")
		}

		utils.WriteManifest(Args.Context)
		err := os.MkdirAll(path.Join(Args.Context, "releases"), 0755)
		if err == nil {
			err = os.MkdirAll(path.Join(Args.Context, "singles"), 0755)
		}
		if err != nil {
			logger.Fatal("failed to create directories, ", err)
		}

		logger.Info("manifest initialized")
	},
}
