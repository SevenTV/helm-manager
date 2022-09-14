package remove

import (
	"github.com/manifoldco/promptui"
	"github.com/seventv/helm-manager/manager/types"
	"github.com/seventv/helm-manager/manager/utils"
	"go.uber.org/zap"
)

func runRemoveRepo(cfg types.Config) {
	if len(cfg.Repos) == 0 {
		utils.Fatal("No repositories added")
	}

	if cfg.Arguments.Remove.Repo.Name == "" {
		if cfg.Arguments.InTerminal {
			prompt := promptui.Select{
				Label: "Name",
				Items: cfg.Repos,
				Templates: &promptui.SelectTemplates{
					Label:    "{{ .Name }}?",
					Active:   "➔ {{ .Name | cyan }} ( {{ .URL | red }} )",
					Inactive: "  {{ .Name | cyan }} ( {{ .URL | red }} )",
					Selected: `{{ "Name:" | faint }} {{ .Name }}`,
				},
			}

			i, _, err := prompt.Run()
			if err != nil {
				zap.S().Fatal(err)
			}

			cfg.Arguments.Remove.Repo.Name = cfg.Repos[i].Name
		} else {
			utils.Fatal("Non-interactive mode requires a repo name")
		}
	}

	repoName := cfg.Arguments.Remove.Repo.Name

	var (
		repo types.Repo
		idx  = -1
	)
	for idx, repo = range cfg.Repos {
		if repo.Name == repoName {
			break
		}
		idx = -1
	}

	if idx == -1 {
		utils.Fatal("repo %s was not found", repoName)
	}

	cfg.Repos = append(cfg.Repos[:idx], cfg.Repos[idx+1:]...)

	zap.S().Infof("Repo has been removed from the manifest %s", repo.Name)

	utils.WriteConfig(cfg)
}
