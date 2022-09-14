package remove

import (
	"bufio"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"github.com/seventv/helm-manager/manager/types"
	"github.com/seventv/helm-manager/manager/utils"
	"go.uber.org/zap"
)

func runRemoveChart(cfg types.Config) {
	if len(cfg.Charts) == 0 {
		utils.Fatal("No charts added")
	}

	chartMp := map[string]int{}

	for i, c := range cfg.Charts {
		chartMp[c.Name] = i
	}

	if cfg.Arguments.Remove.Chart.Name == "" {
		if !cfg.Arguments.NonInteractive {
			prompt := promptui.Select{
				Label: "Select chart to remove",
				Items: cfg.Charts,
				Templates: &promptui.SelectTemplates{
					Label:    "{{ . }}?",
					Active:   "➔ {{ .Name | cyan }} ({{ .Chart | red }})",
					Inactive: "  {{ .Name | cyan }} ({{ .Chart | red }})",
					Selected: `{{ "Chart:" | faint }} {{ .Name }}`,
					Details: `
--------- Chart ----------
{{ "Name:" | faint }}	{{ .Name }}
{{ "Chart:" | faint }}	{{ .Chart }}
{{ "Version:" | faint }}	{{ .Version }}
`,
				},
			}

			i, _, err := prompt.Run()
			if err != nil {
				zap.S().Fatal(err)
			}

			cfg.Arguments.Remove.Chart.Name = cfg.Charts[i].Name
		} else {
			utils.Fatal("Non-interactive mode requires a chart name")
		}
	}

	if !cfg.Arguments.Remove.Chart.Wait && !cfg.Arguments.NonInteractive {
		prompt := promptui.Prompt{
			Label:     "Do you want to wait for the chart to be removed",
			IsConfirm: true,
		}

		_, err := prompt.Run()
		if err != nil && err.Error() != "" {
			zap.S().Fatal(err)
		}

		cfg.Arguments.Remove.Chart.Wait = err == nil
	}

	if !cfg.Arguments.Remove.Chart.DryRun && !cfg.Arguments.NonInteractive {
		prompt := promptui.Prompt{
			Label:     "Do you want to do a dry run",
			IsConfirm: true,
		}

		_, err := prompt.Run()
		if err != nil && err.Error() != "" {
			zap.S().Fatal(err)
		}

		cfg.Arguments.Remove.Chart.DryRun = err == nil
	}

	chartName := cfg.Arguments.Remove.Chart.Name

	idx, ok := chartMp[chartName]
	if !ok {
		utils.Fatal("chart %s was not found", chartName)
	}

	chart := cfg.Charts[idx]

	cfg.Charts = append(cfg.Charts[:idx], cfg.Charts[idx+1:]...)

	args := []string{
		"uninstall", chart.Name,
		"-n", chart.Namespace,
	}

	if cfg.Arguments.Remove.Chart.Wait {
		args = append(args, "--wait")
	}

	if cfg.Arguments.Remove.Chart.DryRun {
		args = append(args, "--dry-run")
	}

	if !cfg.Arguments.Remove.Chart.Confirm && !cfg.Arguments.Remove.Chart.DryRun {
		if !cfg.Arguments.NonInteractive {
			prompt := promptui.Prompt{
				Label:     "Are you sure you want to delete this chart",
				IsConfirm: true,
			}

			_, err := prompt.Run()
			if err != nil && err != promptui.ErrAbort {
				zap.S().Fatal(err)
			}

			if err == promptui.ErrAbort {
				utils.Fatal("Aborted")
			}
		} else {
			reader := bufio.NewReader(os.Stdin)

			response, err := reader.ReadString('\n')
			if err != nil {
				zap.S().Fatal(err)
			}

			response = strings.ToLower(strings.TrimSpace(response))

			if response == "n" || response == "no" {
				utils.Fatal("Aborted")
			}
		}
	}

	waiting := make(chan bool)
	finished := make(chan struct{})
	go func() {
		defer close(waiting)
		defer close(finished)

		if cfg.Arguments.InTerminal {
			t := time.NewTicker(200 * time.Millisecond)
			defer t.Stop()
			i := 0
			stages := []string{"\\", "|", "/", "-"}
			for {
				select {
				case <-t.C:
					zap.S().Infof("%s [%s]\r", color.YellowString("Removing chart"), color.CyanString("%s", stages[i%len(stages)]))
					i++
				case success := <-waiting:
					if success {
						zap.S().Infof("%s Chart Removed", color.GreenString("✓"))
					} else {
						zap.S().Infof("%s Failed to remove chart", color.RedString("✗"))
					}
					return
				}
			}
		} else {
			utils.Info("Removing Chart...")
			if <-waiting {
				utils.Info("Chart Removed")
			} else {
				utils.Info("Failed to remove chart")
			}
		}
	}()

	data, err := utils.ExecuteCommand("helm", args...)
	strData := string(data)
	{
		d := strings.ToLower(strData)
		if strings.Contains(d, "release not loaded") || strings.Contains(d, "release: not found") {
			err = nil
		}
	}

	waiting <- err == nil
	<-finished

	if err != nil {
		utils.Fatal("%s Failed to delete chart\n%s", color.RedString("✗"), data)
	}

	if !cfg.Arguments.Remove.Chart.DryRun {
		utils.WriteConfig(cfg)

		if !cfg.Arguments.Remove.Chart.Delete && !cfg.Arguments.NonInteractive {
			prompt := promptui.Prompt{
				Label:     "Do you want to delete the chart file",
				IsConfirm: true,
			}

			_, err := prompt.Run()
			if err != nil && err.Error() != "" {
				zap.S().Fatal(err)
			}

			cfg.Arguments.Remove.Chart.Delete = err == nil
		}

		if cfg.Arguments.Remove.Chart.Delete {
			if err := os.Remove(chart.File); err != nil {
				zap.S().Infof("%s Failed to remove chart file", color.RedString("✗"))
			} else {
				zap.S().Infof("%s Chart file removed", color.GreenString("✓"))
			}
		}
	}
}
