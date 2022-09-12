package remove

import (
	"strings"

	"github.com/seventv/helm-manager/manager"
	"github.com/seventv/helm-manager/manager/types"
	"go.uber.org/zap"
)

func runRemoveEnv(cfg types.Config) {
	env := strings.ToUpper(cfg.Arguments.Remove.Env.Name)

	newVars := []string{}
	found := false
	for _, e := range cfg.AllowedEnv {
		if strings.ToUpper(e) != env {
			newVars = append(newVars, e)
		} else {
			found = true
		}
	}

	if !found {
		zap.S().Fatalf("env %s was not allowed to begin with", env)
	}

	cfg.AllowedEnv = newVars

	zap.S().Infof("removed env %s from whitelist", env)

	manager.WriteConfig(cfg)
}
