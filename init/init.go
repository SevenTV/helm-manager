package init

import (
	"os"
	"path"

	"github.com/seventv/helm-manager/manager/types"
	"github.com/seventv/helm-manager/manager/utils"
	"go.uber.org/zap"
)

func Run(cfg types.Config) {
	if err := os.MkdirAll(path.Join(cfg.Arguments.WorkingDir, "charts"), 0755); err != nil {
		utils.Fatal("Error creating charts directory")
	}

	if err := os.MkdirAll(path.Join(cfg.Arguments.WorkingDir, "singles"), 0755); err != nil {
		utils.Fatal("Error creating singles directory")
	}

	utils.WriteConfig(cfg)

	zap.S().Info("Sucessfully initialized Helm-Manger")
}
