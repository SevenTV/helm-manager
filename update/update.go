package update

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"github.com/seventv/helm-manager/manager/cli"
	"github.com/seventv/helm-manager/manager/types"
	"github.com/seventv/helm-manager/manager/utils"
	"github.com/seventv/helm-manager/upgrade"
	"go.uber.org/zap"
)

var UpdateColor = color.New(color.Bold, color.FgYellow)

func Run(cfg types.Config) {
	if len(cfg.Charts) == 0 {
		utils.Fatal("No charts added")
	}

	charts := utils.FetchHelmCharts(cfg)
	helmChartMp := map[string]utils.HelmChartParsed{}
	for _, c := range charts {
		helmChartMp[c.Name] = c
	}

	chartMp := map[string]int{}
	for idx, c := range cfg.Charts {
		chartMp[c.Name] = idx
	}

	if !cfg.Arguments.Update.List && !cfg.Arguments.NonInteractive && cfg.Arguments.Update.Name == "" && cfg.Arguments.Update.Version == "" {
		prompt := promptui.Prompt{
			Label:     "Do you want to list all updates",
			IsConfirm: true,
		}

		_, err := prompt.Run()
		if err != nil && err != promptui.ErrAbort {
			zap.S().Fatal(err)
		}

		cfg.Arguments.Update.List = err == nil
	}

	if cfg.Arguments.Update.Version != "" && cfg.Arguments.Update.List {
		utils.Fatal("Cannot list updates and update to a specific version at the same time")
	}

	if cfg.Arguments.Update.Name != "" && cfg.Arguments.Update.List {
		utils.Fatal("Cannot list updates and update a specific chart at the same time")
	}

	if cfg.Arguments.Update.Name == "" && !cfg.Arguments.Update.List {
		if !cfg.Arguments.NonInteractive {
			names := []string{}
			for _, chart := range cfg.Charts {
				names = append(names, chart.Name)
			}

			prompt := promptui.Select{
				Label: "Name",
				Items: names,
				Templates: &promptui.SelectTemplates{
					Label:    "{{ .Name }}?",
					Active:   "➔ {{ . | cyan }}",
					Inactive: "  {{ . | cyan }}",
					Selected: `{{ "Name:" | faint }} {{ .Name }}`,
				},
			}

			i, _, err := prompt.Run()
			if err != nil {
				zap.S().Fatal(err)
			}

			cfg.Arguments.Update.Name = cfg.Charts[i].Name
		} else {
			utils.Fatal("Non-interactive mode requires a chart name")
		}
	}

	if !cfg.Arguments.Update.List {
		idx, ok := chartMp[cfg.Arguments.Update.Name]
		if !ok {
			utils.Fatal("Chart %s not found", cfg.Arguments.Update.Name)
		}

		chart := cfg.Charts[idx]

		if cfg.Arguments.Update.Version == "" {
			if !cfg.Arguments.NonInteractive {
				versions := helmChartMp[chart.Chart].Children

				prompt := promptui.Select{
					Label: "Version (App Version)",
					Items: versions,
					Templates: &promptui.SelectTemplates{
						Label:    "{{ .Version }}?",
						Active:   "➔ {{ .Version | cyan }} ({{ .AppVersion | red }})",
						Inactive: "  {{ .Version | cyan }} ({{ .AppVersion | red }})",
						Selected: `{{ "Version:" | faint }} {{ .Version }}`,
					},
				}

				idx, _, err := prompt.Run()
				if err != nil {
					zap.S().Fatal(err)
				}

				cfg.Arguments.Update.Version = versions[idx].Version
			} else {
				utils.Fatal("Non-interactive mode requires a chart version to update to")
			}
		}

		found := false
		helmChart := helmChartMp[chart.Chart]
		for _, child := range helmChart.Children {
			if child.Version == cfg.Arguments.Update.Version {
				found = true
			}
		}

		if !found {
			utils.Fatal("Version %s not found for chart %s", cfg.Arguments.Update.Version, chart.Chart)
		}

		if cfg.Arguments.Update.Version == chart.Version {
			utils.Fatal("Chart %s is already at version %s", chart.Chart, cfg.Arguments.Update.Version)
		}

		utils.Info("Updating chart %s from %s to %s", chart.Chart, chart.Version, cfg.Arguments.Update.Version)

		chart.Version = cfg.Arguments.Update.Version

		_, success := upgrade.HandleChart(cfg, chart, map[string]string{})
		if !success {
			utils.Fatal("Failed to update chart %s", chart.Chart)
		}

		zap.S().Infof("%s Updated chart in manifest", color.GreenString("✓"))
		zap.S().Infof("To deploy the chart, run: %s", color.YellowString("%s upgrade charts --only %s", cli.BaseCommand.Name, chart.Name))

		cfg.Charts[idx] = chart

		utils.WriteConfig(cfg)
	} else {
		utils.Info("Listing all updates")
		i := 0
		for _, chart := range cfg.Charts {
			helmChart, ok := helmChartMp[chart.Chart]
			if !ok {
				if chart.Version != "" {
					utils.Warn("Chart %s not found in helm repo", chart.Chart)
				}

				continue
			}

			if helmChart.Version != chart.Version {
				i++
				var oldChart utils.HelmChart
				for _, oldChart = range helmChart.Children {
					if oldChart.Version == chart.Version {
						break
					}
				}

				versionChanged := ""
				if oldChart.AppVersion != helmChart.AppVersion {
					versionChanged = fmt.Sprintf("\n\tApp Version: %s -> %s", color.RedString(oldChart.AppVersion), color.GreenString(helmChart.AppVersion))
				}

				utils.Info("Chart %s has an update available from %s to %s%s\n\tYou can update this chart by typing: %s", chart.Chart, chart.Version, helmChart.Version, versionChanged, color.YellowString("%s update --name %s --version %s", cli.BaseCommand.Name, chart.Name, helmChart.Version))
			}
		}
		if i == 0 {
			utils.Info("No updates available")
		}
	}
}
