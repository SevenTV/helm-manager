package add

import (
	"errors"
	"strings"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"github.com/seventv/helm-manager/manager/types"
	"github.com/seventv/helm-manager/manager/utils"
	"go.uber.org/zap"
)

func runAddEnv(cfg types.Config) {
	envMp := map[string]bool{}
	for _, e := range cfg.AllowedEnv {
		envMp[e] = true
	}

	if cfg.Arguments.Add.Env.Name == "" {
		if !cfg.Arguments.NonInteractive {
			prompt := promptui.Prompt{
				Label: "Name",
				Validate: func(input string) error {
					input = strings.ToUpper(input)

					if input == "" {
						return errors.New("Env Variable cannot be empty")
					}

					if strings.Contains(input, " ") {
						return errors.New("Env Variable cannot contain spaces")
					}

					if envMp[input] {
						return errors.New("Env Variable already whitelisted")
					}

					return nil
				},
			}

			result, err := prompt.Run()
			if err != nil {
				zap.S().Fatal(err)
			}

			result = strings.ToUpper(result)

			cfg.Arguments.Add.Env.Name = result
		} else {
			utils.Fatal("No environment variable specified")
		}

	}

	env := strings.ToUpper(cfg.Arguments.Add.Env.Name)

	if envMp[env] {
		utils.Fatal("Env Variable already whitelisted")
	}

	cfg.AllowedEnv = append(cfg.AllowedEnv, env)

	utils.WriteConfig(cfg)

	zap.S().Infof("Env Variable %s added to whitelist", color.GreenString(env))
}
