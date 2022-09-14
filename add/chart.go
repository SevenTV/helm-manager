package add

import (
	"encoding/json"
	"errors"
	"html/template"
	"os"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/jinzhu/copier"
	"github.com/manifoldco/promptui"
	"github.com/seventv/helm-manager/manager/cli"
	"github.com/seventv/helm-manager/manager/types"
	"github.com/seventv/helm-manager/manager/utils"
	"github.com/seventv/helm-manager/upgrade"
	"go.uber.org/zap"
)

type HelmChart struct {
	Name       string `json:"name"`
	Version    string `json:"version"`
	Desctipton string `json:"description"`
}

type HelmChartParsed struct {
	Name       string
	Version    string
	Desctipton string

	Children []HelmChart
}

func runAddChart(cfg types.Config) {
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
			utils.Info("Downloading Helm Charts...")
			if <-downloading {
				utils.Info("Finished downloading Helm Charts")
			} else {
				utils.Info("Failed to download Helm Charts")
			}
		}
	}()

	repos, err := utils.ExecuteCommand("helm", "search", "repo", "--output", "json", "--versions")
	downloading <- err == nil
	<-finished

	if err != nil {
		utils.Fatal("failed to get helm chart list: ", err)
	}

	charts := []HelmChart{}
	if err := json.Unmarshal(repos, &charts); err != nil {
		utils.Fatal("failed to unmarshal helm chart list: ", err)
	}

	chartMap := map[string]*HelmChartParsed{}
	for _, chart := range charts {
		if _, ok := chartMap[chart.Name]; !ok {
			chartMap[chart.Name] = &HelmChartParsed{
				Name:       chart.Name,
				Version:    chart.Version,
				Desctipton: chart.Desctipton,
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

	if cfg.Arguments.Add.Chart.Name == "" {
		if cfg.Arguments.InTerminal {
			namePrompt := promptui.Prompt{
				Label: "Name",
				Validate: func(s string) error {
					if s == "" {
						return errors.New("a name is required")
					}

					if strings.Contains(s, " ") {
						return errors.New("a name cannot contain spaces")
					}

					for _, c := range cfg.Charts {
						if c.Name == s {
							return errors.New("a chart with this name already exists")
						}
					}

					return nil
				},
			}
			result, err := namePrompt.Run()
			if err != nil {
				zap.S().Fatal(err)
			}

			cfg.Arguments.Add.Chart.Name = result
		} else {
			zap.S().Fatal(cli.Parser.Usage(color.RedString("Non-interactive mode requires a name argument")))
		}
	}

	if cfg.Arguments.Add.Chart.Chart == "" {
		if cfg.Arguments.InTerminal {
			chartPrompt := promptui.Select{
				Label:             "Chart",
				Items:             chartsParsed,
				StartInSearchMode: true,
				Templates: &promptui.SelectTemplates{
					Label:    "{{ . }}?",
					Active:   "➔ {{ .Name | cyan }} ({{ .Version | red }})",
					Inactive: "  {{ .Name | cyan }} ({{ .Version | red }})",
					Selected: `{{ "Chart:" | faint }} {{ .Name }}`,
					Details: `
--------- Chart ----------
{{ "Name:" | faint }}	{{ .Name }}
{{ "Version:" | faint }}	{{ .Version }}
{{ "Description:" | faint }}	{{ .Desctipton }}
				`,
				},
				Searcher: func(input string, index int) bool {
					chart := chartsParsed[index]
					name := strings.Replace(strings.ToLower(chart.Name), " ", "", -1)
					input = strings.Replace(strings.ToLower(input), " ", "", -1)
					description := strings.Replace(strings.ToLower(chart.Desctipton), " ", "", -1)

					return strings.Contains(name, input) || strings.Contains(description, input)
				},
			}

			idx, _, err := chartPrompt.Run()
			if err != nil {
				zap.S().Fatal(err)
			}

			cfg.Arguments.Add.Chart.Chart = chartsParsed[idx].Name
		} else {
			zap.S().Fatal(cli.Parser.Usage(color.RedString("Non-interactive mode requires a chart argument")))
		}
	}

	helmChart := chartMap[cfg.Arguments.Add.Chart.Chart]
	if helmChart == nil {
		utils.Fatal("chart not found")
	}

	if cfg.Arguments.Add.Chart.Version == "" {
		if cfg.Arguments.InTerminal {
			versionPrompt := promptui.Select{
				Label: "Version",
				Items: helmChart.Children,
				Templates: &promptui.SelectTemplates{
					Label:    "{{ . }}?",
					Active:   "➔ {{ .Name | cyan }} ({{ .Version | red }})",
					Inactive: "  {{ .Name | cyan }} ({{ .Version | red }})",
					Selected: `{{ "Version:" | faint }} {{ .Version }}`,
					Details: `
	--------- Chart ----------
	{{ "Name:" | faint }}	{{ .Name }}
	{{ "Version:" | faint }}	{{ .Version }}
	{{ "Description:" | faint }}	{{ .Desctipton }}
					`,
				},
				Searcher: func(input string, index int) bool {
					chart := helmChart.Children[index]
					version := strings.Replace(strings.ToLower(chart.Version), " ", "", -1)

					return strings.Contains(version, input)
				},
			}

			cIdx, _, err := versionPrompt.Run()
			if err != nil {
				zap.S().Fatal(err)
			}

			cfg.Arguments.Add.Chart.Version = helmChart.Children[cIdx].Version
		} else {
			zap.S().Fatal(cli.Parser.Usage(color.RedString("Non-interactive mode requires a version argument")))
		}
	} else {
		found := false
		for _, c := range helmChart.Children {
			if c.Version == cfg.Arguments.Add.Chart.Version {
				found = true
				break
			}
		}

		if !found {
			zap.S().Fatal("version not found")
		}
	}

	if cfg.Arguments.Add.Chart.Namespace == "" {
		if cfg.Arguments.InTerminal {
			downloading := make(chan bool)
			finished := make(chan struct{})
			go func() {
				defer close(downloading)
				defer close(finished)

				t := time.NewTicker(200 * time.Millisecond)
				defer t.Stop()
				i := 0
				stages := []string{"\\", "|", "/", "-"}
				for {
					select {
					case <-t.C:
						zap.S().Infof("%s [%s]\r", color.YellowString("Fetching k8s namespaces"), color.CyanString("%s", stages[i%len(stages)]))
						i++
					case success := <-downloading:
						if success {
							zap.S().Infof("%s Fetched k8s namespaces", color.GreenString("✓"))
						} else {
							zap.S().Infof("%s Failed to fetch k8s namespaces", color.RedString("✗"))
						}
						return
					}
				}
			}()

			output, err := utils.ExecuteCommand("kubectl", "get", "ns", "-o=jsonpath={.items[*].metadata.name}")
			downloading <- err == nil
			<-finished

			namespaces := strings.Split(string(output), " ")

			sort.StringSlice(namespaces).Sort()

			const AddNamespace = "* Add new namespace"

			namespacesWithAdd := []string{AddNamespace}
			namespacesWithAdd = append(namespacesWithAdd, namespaces...)

			var funcMap template.FuncMap
			copier.Copy(&funcMap, &promptui.FuncMap)

			funcMap["custom"] = func(s string) string {
				if s == AddNamespace {
					return color.GreenString(s)
				}

				return color.CyanString(s)
			}

			namespacePrompt := promptui.Select{
				Label:        "Namespace",
				Items:        namespacesWithAdd,
				HideSelected: true,
				Templates: &promptui.SelectTemplates{
					FuncMap:  funcMap,
					Label:    "{{ . }}?",
					Active:   "➔ {{ . | custom }}",
					Inactive: "  {{ . | custom }}",
				},
			}

			_, result, err := namespacePrompt.Run()
			if err != nil {
				zap.S().Fatal(err)
			}

			if result == AddNamespace {
				namespacePrompt := promptui.Prompt{
					Label:       "Namespace",
					HideEntered: true,
					Validate: func(s string) error {
						if s == "" {
							return errors.New("namespace can not be empty")
						}

						for _, ns := range namespaces {
							if ns == s {
								return errors.New("namespace already exists")
							}
						}

						return nil
					},
				}

				result, err = namespacePrompt.Run()
				if err != nil {
					zap.S().Fatal(err)
				}
			}

			zap.S().Infof("%s %s", color.New(color.Faint).Sprint("Namespace:"), result)

			cfg.Arguments.Add.Chart.Namespace = result
		} else {
			zap.S().Fatal(cli.Parser.Usage(color.RedString("Non-interactive mode requires a namespace argument")))
		}
	}

	inputFile := cfg.Arguments.Add.Chart.File

	chart := types.Chart{
		Name:      cfg.Arguments.Add.Chart.Name,
		Chart:     cfg.Arguments.Add.Chart.Chart,
		Version:   cfg.Arguments.Add.Chart.Version,
		Namespace: cfg.Arguments.Add.Chart.Namespace,
		File:      path.Join(cfg.Arguments.WorkingDir, "charts", cfg.Arguments.Add.Chart.Name+".yaml"),
	}

	for _, c := range cfg.Charts {
		if c.Name == chart.Name {
			utils.Fatal("chart with name %s already exists", chart.Name)
		}
	}

	if !cfg.Arguments.Add.Chart.Overwrite {
		if _, err := os.Stat(chart.File); err == nil {
			if cfg.Arguments.InTerminal {
				prompt := promptui.Prompt{
					Label:     "Chart file already exists, overwrite",
					IsConfirm: true,
				}

				_, err := prompt.Run()
				if err != nil {
					zap.S().Fatal(color.RedString("✗ Aborted"))
				}
			} else {
				utils.Fatal("Chart file already exists, use --overwrite to overwrite")
			}
		}
	}

	if inputFile == "" && cfg.Arguments.InTerminal {
		prompt := promptui.Prompt{
			Label: "Template File",
			Validate: func(s string) error {
				if s == "" {
					return nil
				}

				if s, err := os.Stat(s); err != nil {
					return errors.New("file does not exist")
				} else if s.IsDir() {
					return errors.New("file is a directory")
				}

				return nil
			},
		}

		var err error
		inputFile, err = prompt.Run()
		if err != nil {
			zap.S().Fatal(err)
		}
	}

	if inputFile != "" {
		if _, err := os.Stat(inputFile); err != nil {
			utils.Fatal("input file does not exist")
		}

		data, err := os.ReadFile(inputFile)
		if err != nil {
			utils.Fatal("failed to read chart input file %s: %v", inputFile, err)
		}

		if err = os.MkdirAll(path.Dir(chart.File), 0755); err != nil {
			utils.Fatal("failed to create chart directory %s: %v", path.Dir(chart.File), err)
		}

		if err = os.WriteFile(chart.File, data, 0644); err != nil {
			utils.Fatal("failed to write chart file %s: %v", chart.File, err)
		}
	} else {
		if _, err := os.Stat(chart.File); err == nil {
			err = os.Remove(chart.File)
			if err != nil {
				utils.Fatal("failed to remove chart file %s: %v", chart.File, err)
			}
		}
	}

	_, success := upgrade.HandleChart(cfg, chart, map[string]string{})
	if !success {
		os.Exit(1)
	}

	cfg.Charts = append(cfg.Charts, chart)

	utils.WriteConfig(cfg)

	zap.S().Infof("%s Added chart to manifest", color.GreenString("✓"))
	zap.S().Infof("To deploy the chart, run: %s", color.YellowString("%s upgrade", cli.BaseCommand.Name))
}
