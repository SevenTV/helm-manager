package add

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"github.com/seventv/helm-manager/manager/cli"
	"github.com/seventv/helm-manager/manager/types"
	"github.com/seventv/helm-manager/manager/utils"
	"go.uber.org/zap"
)

func runAddSingle(cfg types.Config) {
	singleMp := map[string]bool{}
	for _, s := range cfg.Singles {
		singleMp[s.Name] = true
	}

	if cfg.Arguments.Add.Single.Name == "" {
		if !cfg.Arguments.NonInteractive {
			prompt := promptui.Prompt{
				Label: "Name",
				Validate: func(input string) error {
					if input == "" {
						return errors.New("name cannot be empty")
					}

					if strings.Contains(input, " ") {
						return errors.New("name cannot contain spaces")
					}

					if singleMp[input] {
						return errors.New("single with name already exists")
					}

					return nil
				},
			}

			result, err := prompt.Run()
			if err != nil {
				zap.S().Fatal(err)
			}

			cfg.Arguments.Add.Single.Name = result

		} else {
			utils.Fatal("Non-interactive mode requires a name")
		}
	}

	if cfg.Arguments.Add.Single.Namespace == "" {
		if !cfg.Arguments.NonInteractive {
			prompt := promptui.Prompt{
				Label: "Namespace",
				Validate: func(input string) error {
					if input == "" {
						return errors.New("namespace cannot be empty")
					}

					if strings.Contains(input, " ") {
						return errors.New("namespace cannot contain spaces")
					}

					return nil
				},
			}

			result, err := prompt.Run()
			if err != nil {
				zap.S().Fatal(err)
			}

			cfg.Arguments.Add.Single.Namespace = result
		} else {
			utils.Fatal("Non-interactive mode requires a namespace")
		}
	}

	if cfg.Arguments.Add.Single.File == "" {
		if !cfg.Arguments.NonInteractive {
			prompt := promptui.Prompt{
				Label: "File",
				Validate: func(input string) error {
					if input == "" {
						return errors.New("file cannot be empty")
					}

					if s, err := os.Stat(input); err != nil || s.IsDir() {
						return errors.New("file does not exist")
					}

					return nil
				},
			}

			result, err := prompt.Run()
			if err != nil {
				zap.S().Fatal(err)
			}

			cfg.Arguments.Add.Single.File = result
		} else {
			utils.Fatal("Non-interactive mode requires a file")
		}
	}

	if !cfg.Arguments.Add.Single.UseCreate {
		if !cfg.Arguments.NonInteractive {
			prompt := promptui.Select{
				Label: "Use create instead of apply",
				Items: []string{"true", "false"},
			}

			_, result, err := prompt.Run()
			if err != nil {
				zap.S().Fatal(err)
			}

			cfg.Arguments.Add.Single.UseCreate = result == "true"

			zap.S().Infof("Use create: %s", color.GreenString(result))
		} else {
			cfg.Arguments.Add.Single.UseCreate = false
		}
	}

	single := types.Single{
		Name:      cfg.Arguments.Add.Single.Name,
		Namespace: cfg.Arguments.Add.Single.Namespace,
		UseCreate: cfg.Arguments.Add.Single.UseCreate,
	}

	for _, c := range cfg.Singles {
		if c.Name == single.Name {
			utils.Fatal("single with name %s already exists", single.Name)
		}
	}

	pth := path.Join(cfg.Arguments.WorkingDir, "singles", fmt.Sprintf("%s.yaml", single.Name))
	if _, err := os.Stat(pth); err == nil && !cfg.Arguments.Add.Single.Overwrite && pth != cfg.Arguments.Add.Single.File {
		if !cfg.Arguments.NonInteractive {
			prompt := promptui.Prompt{
				Label:     "Single already exists, overwrite",
				IsConfirm: true,
			}

			_, err := prompt.Run()
			if err != nil && err.Error() != "" {
				zap.S().Fatal(err)
			} else if err != nil {
				utils.Fatal("Aborted")
			}
		} else {
			utils.Fatal("Single already exists, overwrite with --overwrite")
		}
	}

	data, err := os.ReadFile(cfg.Arguments.Add.Single.File)
	if err != nil {
		utils.Fatal("failed to read single file: %s", err)
	}

	if err := os.WriteFile(path.Join(cfg.Arguments.WorkingDir, "singles", fmt.Sprintf("%s.yaml", single.Name)), data, 0644); err != nil {
		utils.Fatal("failed to write single file: %s", err)
	}

	cfg.Singles = append(cfg.Singles, single)

	utils.WriteConfig(cfg)

	zap.S().Infof("%s Added single to manifest", color.GreenString("✓"))
	zap.S().Infof("To deploy the single, run: %s", color.YellowString("%s upgrade", cli.BaseCommand.Name))
}
