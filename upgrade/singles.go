package upgrade

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"github.com/seventv/helm-manager/manager/types"
	"github.com/seventv/helm-manager/manager/utils"
	"go.uber.org/zap"
)

func runUpgradeSingles(cfg types.Config) {
	if len(cfg.Singles) == 0 {
		utils.Fatal("No singles found in manifest")
	}

	envMap := utils.CreateEnvMap(cfg)

	if !cfg.Arguments.Upgrade.DryRun && !cfg.Arguments.Upgrade.Singles.Deploy && !cfg.Arguments.NonInteractive {
		prompt := promptui.Prompt{
			Label:     "Are you sure you want to deploy these changes",
			IsConfirm: true,
		}

		_, err := prompt.Run()
		if err != nil && err != promptui.ErrAbort {
			zap.S().Fatal(err)
		}

		cfg.Arguments.Upgrade.Singles.Deploy = err == nil
		if !cfg.Arguments.Upgrade.Singles.Deploy {
			utils.Fatal("Aborted")
		}

	} else if !cfg.Arguments.Upgrade.DryRun && !cfg.Arguments.Upgrade.Singles.Deploy {
		utils.Fatal("Non-interactive mode requires --deploy")
	}

	if cfg.Arguments.Upgrade.Singles.Deploy {
		for _, single := range cfg.Singles {
			if !HandleSingle(cfg, single, envMap) && cfg.Arguments.Upgrade.StopOnFirstError {
				utils.Fatal("Failed to upgrade single %s, failing fast", single.Name)
			}
		}
	}
}

func HandleSingle(cfg types.Config, single types.Single, envMap map[string]string) bool {
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
					zap.S().Infof("%s [%s]\r", color.YellowString("Upgrading single %s", single.Name), color.CyanString("%s", stages[i%len(stages)]))
					i++
				case success := <-waiting:
					if success {
						zap.S().Infof("%s Single upgraded %s", color.GreenString("✓"), single.Name)
					} else {
						zap.S().Infof("%s Failed to upgrade single %s", color.RedString("✗"), single.Name)
					}
					return
				}
			}
		} else {
			utils.Info("Upgrading single %s...", single.Name)
			if <-waiting {
				utils.Info("Single %s upgrade", single.Name)
			} else {
				utils.Info("Failed to upgrade single %s", single.Name)
			}
		}
	}()

	args := []string{"apply"}
	if single.Namespace != "" {
		args = append(args, "--namespace", single.Namespace)
	}
	if cfg.Arguments.Upgrade.DryRun {
		args = append(args, "--dry-run")
	}
	args = append(args, "-f", "-")

	data, err := os.ReadFile(path.Join(cfg.Arguments.WorkingDir, "singles", fmt.Sprintf("%s.yaml", single.Name)))
	if err != nil {
		waiting <- false
		<-finished
		utils.Error("Failed to read single file %s: %v", single.Name, err)
		return false
	}

	if single.UseCreate {
		args[0] = "create"
	}

	for env, value := range envMap {
		data = bytes.ReplaceAll(data, []byte(fmt.Sprintf("${%s}", env)), []byte(value))
	}

	output, err := utils.ExecuteCommandStdin(bytes.NewReader(data), "kubectl", args...)
	if err != nil {
		waiting <- false
		<-finished
		zap.S().Errorf("Failed to upgrade single %s: %v\n%s", single.Name, err, output)
		return false
	}

	waiting <- true
	<-finished

	return true
}
