package add

import (
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

func runAddChart(cfg types.Config) {
	charts := utils.FetchHelmCharts(cfg)

	if cfg.Arguments.Add.Chart.Name == "" {
		if !cfg.Arguments.NonInteractive {
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

	isLocal := false
	if cfg.Arguments.Add.Chart.Chart == "" {
		if !cfg.Arguments.NonInteractive {
			isLocalPrompt := promptui.Prompt{
				Label:     "Is this a local chart",
				IsConfirm: true,
			}

			_, err := isLocalPrompt.Run()
			if err != nil && err != promptui.ErrAbort {
				zap.S().Fatal(err)
			}

			isLocal = err == nil

			if isLocal {
				chartPrompt := promptui.Prompt{
					Label: "Path to chart",
					Validate: func(s string) error {
						if s == "" {
							return errors.New("a path is required")
						}

						if stat, err := os.Stat(s); err != nil {
							return errors.New("path does not exist")
						} else if !stat.IsDir() {
							return errors.New("path is not a directory")
						} else if stat, err = os.Stat(path.Join(s, "Chart.yaml")); err != nil || stat.IsDir() {
							return errors.New("path does not contain a Chart.yaml file")
						}

						return nil
					},
				}

				result, err := chartPrompt.Run()
				if err != nil {
					zap.S().Fatal(err)
				}

				cfg.Arguments.Add.Chart.Chart = result
			} else {
				chartPrompt := promptui.Select{
					Label:             "Chart",
					Items:             charts,
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
						chart := charts[index]
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

				cfg.Arguments.Add.Chart.Chart = charts[idx].Name
			}

		} else {
			zap.S().Fatal(cli.Parser.Usage(color.RedString("Non-interactive mode requires a chart argument")))
		}
	}

	var helmChart utils.HelmChartParsed
	for _, c := range charts {
		if c.Name == cfg.Arguments.Add.Chart.Chart {
			helmChart = c
			break
		}
	}

	if helmChart.Name == "" {
		utils.Fatal("Chart not found")
	}

	if cfg.Arguments.Add.Chart.Version == "" {
		if !cfg.Arguments.NonInteractive {
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
		if !cfg.Arguments.NonInteractive {
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
			if !cfg.Arguments.NonInteractive {
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

	if inputFile == "" && !cfg.Arguments.NonInteractive {
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
	zap.S().Infof("To deploy the chart, run: %s", color.YellowString("%s upgrade charts --only %s", cli.BaseCommand.Name, chart.Name))
}
