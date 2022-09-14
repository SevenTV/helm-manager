package add

import (
	"errors"
	"net/url"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/seventv/helm-manager/manager/types"
	"github.com/seventv/helm-manager/manager/utils"
	"go.uber.org/zap"
)

func runAddRepo(cfg types.Config) {
	repoMp := map[string]bool{}
	for _, r := range cfg.Repos {
		repoMp[r.Name] = true
	}

	if cfg.Arguments.Add.Repo.Name == "" {
		if cfg.Arguments.InTerminal {
			prompt := promptui.Prompt{
				Label: "Name",
				Validate: func(input string) error {
					if input == "" {
						return errors.New("Repo Name cannot be empty")
					}

					if strings.Contains(input, " ") {
						return errors.New("Repo Name cannot contain spaces")
					}

					if repoMp[input] {
						return errors.New("Repo Name already added")
					}

					return nil
				},
			}

			result, err := prompt.Run()
			if err != nil {
				zap.S().Fatal(err)
			}

			cfg.Arguments.Add.Repo.Name = result
		} else {
			utils.Fatal("No repo name specified")
		}
	}

	if cfg.Arguments.Add.Repo.URL == "" {
		if cfg.Arguments.InTerminal {
			prompt := promptui.Prompt{
				Label: "URL",
				Validate: func(input string) error {
					if input == "" {
						return errors.New("Repo URL cannot be empty")
					}

					if _, err := url.ParseRequestURI(input); err != nil {
						return errors.New("Repo URL is not a valid URL")

					}
					return nil
				},
			}

			result, err := prompt.Run()
			if err != nil {
				zap.S().Fatal(err)
			}

			cfg.Arguments.Add.Repo.URL = result
		} else {
			utils.Fatal("Non-interactve mode requires a Repo URL")
		}
	}

	repo := types.Repo{
		Name: cfg.Arguments.Add.Repo.Name,
		URL:  cfg.Arguments.Add.Repo.URL,
	}

	if repoMp[repo.Name] {
		utils.Fatal("repo with name %s already exists", repo.Name)
	}

	cfg.Repos = append(cfg.Repos, repo)

	utils.UpdateRepos(cfg)

	zap.S().Infof("Added repo to manifest")

	utils.WriteConfig(cfg)
}
