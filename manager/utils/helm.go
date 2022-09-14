package utils

import (
	"encoding/json"
	"sort"
	"time"

	"github.com/fatih/color"
	"github.com/seventv/helm-manager/manager/types"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

type HelmChart struct {
	Name       string `json:"name"`
	Version    string `json:"version"`
	AppVersion string `json:"app_version"`
	Desctipton string `json:"description"`
}

type HelmChartParsed struct {
	Name       string
	Version    string
	AppVersion string
	Desctipton string

	Children []HelmChart
}

func FetchHelmCharts(cfg types.Config) []HelmChartParsed {
	downloading := make(chan bool)
	finished := make(chan struct{})
	go func() {
		defer close(downloading)
		defer close(finished)

		if cfg.Arguments.InTerminal {
			t := time.NewTicker(200 * time.Millisecond)
			defer t.Stop()
			i := 0
			stages := []string{"\\", "|", "/", "-"}
			for {
				select {
				case <-t.C:
					zap.S().Infof("%s [%s]\r", color.YellowString("Fetching Helm Charts"), color.CyanString("%s", stages[i%len(stages)]))
					i++
				case success := <-downloading:
					if success {
						zap.S().Infof("%s Feteched helm charts", color.GreenString("✓"))
					} else {
						zap.S().Infof("%s Failed to fetch helm charts", color.RedString("✗"))
					}
					return
				}
			}
		} else {
			Info("Downloading Helm Charts...")
			if <-downloading {
				Info("Finished downloading Helm Charts")
			} else {
				Info("Failed to download Helm Charts")
			}
		}
	}()

	repos, err := ExecuteCommand("helm", "search", "repo", "--output", "json", "--versions")
	downloading <- err == nil
	<-finished

	if err != nil {
		Fatal("failed to get helm chart list: ", err)
	}

	charts := []HelmChart{}
	if err := json.Unmarshal(repos, &charts); err != nil {
		Fatal("failed to unmarshal helm chart list: ", err)
	}

	chartMap := map[string]*HelmChartParsed{}
	for _, chart := range charts {
		if _, ok := chartMap[chart.Name]; !ok {
			chartMap[chart.Name] = &HelmChartParsed{
				Name:       chart.Name,
				Version:    chart.Version,
				Desctipton: chart.Desctipton,
				AppVersion: chart.AppVersion,
				Children:   []HelmChart{chart},
			}
		} else {
			chartMap[chart.Name].Children = append(chartMap[chart.Name].Children, chart)
		}
	}

	chartsParsed := []HelmChartParsed{}
	for _, chart := range chartMap {
		chartsParsed = append(chartsParsed, *chart)
	}

	sort.Slice(chartsParsed, func(i, j int) bool {
		return chartsParsed[i].Name < chartsParsed[j].Name
	})

	return chartsParsed
}

func GetHelmChartDefaultValues(chart types.Chart) yaml.Node {
	args := []string{"show", "values", chart.Chart}
	if chart.Version != "" {
		args = append(args, "--version", chart.Version)
	}

	out, err := ExecuteCommand("helm", args...)
	if err != nil {
		Error("Failed to get values for %s", chart.Name)
		return yaml.Node{}
	}

	chartValues, err := ParseYaml(out)
	if err != nil {
		Error("Failed to parse values for %s", chart.Name)
		return yaml.Node{}
	}

	return ConvertDocument(chartValues)
}

func UpdateRepos(cfg types.Config) {
	repoMap := CreateRepoMap(cfg)
	if len(repoMap) == 0 {
		return
	}

	updating := make(chan bool)
	finished := make(chan struct{})
	go func() {
		defer close(finished)
		defer close(updating)
		if cfg.Arguments.InTerminal {
			t := time.NewTicker(200 * time.Millisecond)
			defer t.Stop()
			i := 0
			stages := []string{"\\", "|", "/", "-"}
			for {
				select {
				case <-t.C:
					zap.S().Infof("%s [%s]\r", color.YellowString("Updating Helm Repos"), color.CyanString("%s", stages[i%len(stages)]))
					i++
				case success := <-updating:
					if success {
						zap.S().Infof("%s Updated Helm Repos", color.GreenString("✓"))
					} else {
						zap.S().Infof("%s Failed to update Helm Repos", color.RedString("✗"))
					}
					return
				}
			}
		} else {
			Info("Updating Helm Repos...")
			if <-updating {
				Info("Updated Helm Repos")
			} else {
				Error("Failed to update Helm Repos")
			}
		}
	}()

	data, err := ExecuteCommand("helm", "repo", "list", "-o", "json")
	if err != nil {
		updating <- false
		<-finished
		Fatal("Failed to list helm repos, is helm installed?\n", data)
	}

	repos, err := ParseHelmRepos(data)
	if err != nil {
		updating <- false
		<-finished

		Fatal("Failed to parse helm repo list")
	}

	installedReposMap := map[string]types.Repo{}
	for _, repo := range repos {
		installedReposMap[repo.Name] = repo
	}

	for _, repo := range repoMap {
		installedRepo, ok := installedReposMap[repo.Name]
		if ok && installedRepo.URL != repo.URL {
			out, err := ExecuteCommand("helm", "repo", "remove", repo.Name)
			if err != nil {
				updating <- false
				<-finished
				Fatal("Failed to remove repo %s\n%s", repo.Name, out)
			}
			ok = false
		}

		if !ok {
			out, err := ExecuteCommand("helm", "repo", "add", repo.Name, repo.URL)
			if err != nil {
				updating <- false
				<-finished
				Fatal("Failed to add helm repo %s\n%s", repo.Name, out)
			}
		}
	}

	_, err = ExecuteCommand("helm", "repo", "update")
	updating <- err == nil
	<-finished
	if err != nil {
		Fatal("Failed to update helm repos")
	}
}

func ParseHelmRepos(data []byte) ([]types.Repo, error) {
	repos := []types.Repo{}
	err := json.Unmarshal(data, &repos)
	if err != nil {
		return repos, err
	}

	return repos, nil
}
