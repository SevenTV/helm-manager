package utils

import (
	"encoding/json"
	"sort"
	"time"

	"github.com/fatih/color"
	"github.com/seventv/helm-manager/manager/types"
	"go.uber.org/zap"
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
