package init

import (
	"os"
	"path"

	"github.com/seventv/helm-manager/manager"
	"github.com/seventv/helm-manager/manager/types"
	"go.uber.org/zap"
)

func Run(cfg types.Config) {
	if err := os.MkdirAll(path.Join(cfg.Arguments.WorkingDir, "charts"), 0755); err != nil {
		zap.S().Fatal("Error creating charts directory")
	}

	if err := os.MkdirAll(path.Join(cfg.Arguments.WorkingDir, "singles"), 0755); err != nil {
		zap.S().Fatal("Error creating singles directory")
	}

	manager.WriteConfig(cfg)

	zap.S().Info("Sucessfully initialized manifest.yaml")
}
