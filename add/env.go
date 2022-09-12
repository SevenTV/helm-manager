package add

import (
	"strings"

	"github.com/seventv/helm-manager/manager"
	"github.com/seventv/helm-manager/manager/types"
	"go.uber.org/zap"
)

func runAddEnv(cfg types.Config) {
	env := strings.ToUpper(cfg.Arguments.Add.Env.Name)

	for _, e := range cfg.AllowedEnv {
		if strings.ToUpper(e) == env {
			zap.S().Fatalf("env with name %s already exists", env)
		}
	}

	cfg.AllowedEnv = append(cfg.AllowedEnv, env)

	zap.S().Infof("added env to whitelist %s", env)

	manager.WriteConfig(cfg)
}
