package remove

import (
	"bytes"
	"os"
	"path"

	"github.com/seventv/helm-manager/manager"
	"github.com/seventv/helm-manager/manager/types"
	"go.uber.org/zap"
)

func runRemoveSingle(cfg types.Config) {
	singleName := cfg.Arguments.Remove.Single.Name

	var (
		single types.Single
		idx    = -1
	)
	for idx, single = range cfg.Singles {
		if single.Name == singleName {
			break
		}
		idx = -1
	}

	if idx == -1 {
		zap.S().Fatalf("single %s was not found", singleName)
	}

	cfg.Singles = append(cfg.Singles[:idx], cfg.Singles[idx+1:]...)

	data, err := os.ReadFile(path.Join(cfg.Arguments.WorkingDir, "singles", single.File))
	if err != nil {
		zap.S().Fatalf("failed to read single file %s: %v", single.File, err)
	}

	data, err = manager.ExecuteCommandStdin(bytes.NewReader(data), "kubectl", "delete", "-n", single.Namespace, "-f", "-")
	if err != nil {
		zap.S().Fatalf("failed to delete single %s %v\n%s", single.Name, err, string(data))
	}

	zap.S().Infof("single %s deleted", single.Name)

	manager.WriteConfig(cfg)
}
