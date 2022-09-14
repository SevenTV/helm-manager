package remove

import (
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/seventv/helm-manager/manager/types"
	"github.com/seventv/helm-manager/manager/utils"
	"go.uber.org/zap"
)

func runRemoveEnv(cfg types.Config) {
	if len(cfg.AllowedEnv) == 0 {
		utils.Fatal("No whitelisted environments added")
	}

	if cfg.Arguments.Remove.Env.Name == "" {
		if !cfg.Arguments.NonInteractive {
			prompt := promptui.Select{
				Label: "Name",
				Items: cfg.AllowedEnv,
				Templates: &promptui.SelectTemplates{
					Label:    "{{ . }}?",
					Active:   "➔ {{ . | cyan }}",
					Inactive: "  {{ . | cyan }}",
					Selected: `{{ "Name:" | faint }} {{ . }}`,
				},
			}

			i, _, err := prompt.Run()
			if err != nil {
				zap.S().Fatal(err)
			}

			cfg.Arguments.Remove.Env.Name = cfg.AllowedEnv[i]
		} else {
			utils.Fatal("Non-interactive mode requires an env variable name")
		}
	}

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
		utils.Fatal("env %s was not allowed to begin with", env)
	}

	cfg.AllowedEnv = newVars

	zap.S().Infof("Removed env from whitelist")

	utils.WriteConfig(cfg)
}
