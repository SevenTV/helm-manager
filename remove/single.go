package remove

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"github.com/seventv/helm-manager/manager/types"
	"github.com/seventv/helm-manager/manager/utils"
	"go.uber.org/zap"
)

func runRemoveSingle(cfg types.Config) {
	if len(cfg.Singles) == 0 {
		utils.Fatal("No singles added")
	}

	if cfg.Arguments.Remove.Single.Name == "" {
		if cfg.Arguments.InTerminal {
			prompt := promptui.Select{
				Label: "Name",
				Items: cfg.Singles,
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

			cfg.Arguments.Remove.Single.Name = cfg.Singles[i].Name
		} else {
			utils.Fatal("Non-interactive mode requires a single name")
		}
	}

	if cfg.Arguments.InTerminal && !cfg.Arguments.Remove.Single.DryRun {
		prompt := promptui.Prompt{
			Label:     "Do you want to do a dry run",
			IsConfirm: true,
		}

		_, err := prompt.Run()
		if err != nil && err != promptui.ErrAbort {
			zap.S().Fatal(err)
		}

		cfg.Arguments.Remove.Single.DryRun = err == nil
	}

	singleName := cfg.Arguments.Remove.Single.Name

	var (
		single types.Single
		idx    = -1
	)
	for idx, single = range cfg.Singles {
		if single.Name == singleName {
			break
		}
		idx = -1
	}

	if idx == -1 {
		utils.Fatal("single %s was not found", singleName)
	}

	cfg.Singles = append(cfg.Singles[:idx], cfg.Singles[idx+1:]...)

	data, err := os.ReadFile(path.Join(cfg.Arguments.WorkingDir, "singles", fmt.Sprintf("%s.yaml", single.Name)))
	if err != nil {
		utils.Fatal("failed to read single file %s: %v", single.Name, err)
	}

	args := []string{
		"delete",
		"-n", single.Namespace,
		"-f", "-",
	}

	if cfg.Arguments.Remove.Single.DryRun {
		args = append(args, "--dry-run")
	}

	if !cfg.Arguments.Remove.Single.Confirm && !cfg.Arguments.Remove.Single.DryRun {
		if cfg.Arguments.InTerminal {
			prompt := promptui.Prompt{
				Label:     "Are you sure you want to delete this single",
				IsConfirm: true,
			}

			_, err := prompt.Run()
			if err != nil && err != promptui.ErrAbort {
				zap.S().Fatal(err)
			}

			utils.Fatal("Aborted")
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
					zap.S().Infof("%s [%s]\r", color.YellowString("Removing Single"), color.CyanString("%s", stages[i%len(stages)]))
					i++
				case success := <-waiting:
					if success {
						zap.S().Infof("%s Single Removed", color.GreenString("✓"))
					} else {
						zap.S().Infof("%s Failed to remove single", color.RedString("✗"))
					}
					return
				}
			}
		} else {
			utils.Info("Removing Single...")
			if <-waiting {
				utils.Info("Single Removed")
			} else {
				utils.Info("Failed to remove Single")
			}
		}
	}()

	data, err = utils.ExecuteCommandStdin(bytes.NewReader(data), "kubectl", args...)
	waiting <- err == nil
	<-finished
	if err != nil {
		utils.Fatal("%s Failed to delete single\n%s", color.RedString("✗"), data)
	}

	zap.S().Infof("%s Single deleted", color.GreenString("✓"))

	if !cfg.Arguments.Remove.Single.DryRun {
		utils.WriteConfig(cfg)

		if cfg.Arguments.InTerminal && !cfg.Arguments.Remove.Single.Delete {
			prompt := promptui.Prompt{
				Label:     "Do you want to delete the single file",
				IsConfirm: true,
			}

			_, err := prompt.Run()
			if err != nil && err != promptui.ErrAbort {
				zap.S().Fatal(err)
			}

			cfg.Arguments.Remove.Single.Delete = err == nil
		}

		if cfg.Arguments.Remove.Single.Delete {
			if err := os.Remove(path.Join(cfg.Arguments.WorkingDir, "singles", fmt.Sprintf("%s.yaml", single.Name))); err != nil {
				zap.S().Infof("%s Single file removed", color.GreenString("✓"))
			} else {
				zap.S().Infof("%s Failed to remove single file", color.RedString("✗"))
			}
		}
	}
}
