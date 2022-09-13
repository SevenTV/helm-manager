package upgrade

import (
	"bytes"
	"os"

	"github.com/seventv/helm-manager/manager"
	"github.com/seventv/helm-manager/manager/types"
	"go.uber.org/zap"
)

func Run(cfg types.Config) {
	envMap := manager.CreateEnvMap(cfg)

	manager.ValidateCharts(cfg)
	manager.UpdateRepos(cfg)

	newLockMap := map[string]types.ChartLock{}
	upgradeList := []types.ChartUpgrade{}

	if len(cfg.Charts) == 0 && len(cfg.Singles) == 0 {
		zap.S().Warn("No charts or singles to upgrade")
	}

	zap.S().Debug("Updating charts")

	for _, chart := range cfg.Charts {
		if cfg.Arguments.Upgrade.IgnoreChartsMap[chart.Name] {
			zap.S().Infof("Skipping %s, ignored", chart.Name)
			continue
		}

		if len(cfg.Arguments.Upgrade.ChartWhitelist) != 0 && !cfg.Arguments.Upgrade.ChartWhitelist[chart.Name] {
			zap.S().Infof("Skipping %s, not in whitelist", chart.Name)
			continue
		}

		chartUpgrade, success := HandleChart(chart, envMap, false)
		if !success {
			if cfg.Arguments.Upgrade.StopOnFirstError {
				zap.S().Fatalf("Failed to upgrade chart values %s", chart.Name)
			}

			continue
		} else {
			newLockMap[chart.Name] = chartUpgrade.ChartLock
			upgradeList = append(upgradeList, chartUpgrade)
		}
	}

	zap.S().Debug("Finished updating charts")

	if len(upgradeList) > 0 {
		zap.S().Debug("Updating cluster")

		for _, upgrade := range upgradeList {
			if !HandleUpgrade(cfg, upgrade) && cfg.Arguments.Upgrade.StopOnFirstError {
				zap.S().Fatalf("Failed to upgrade chart %s, failing fast", upgrade.Chart.Name)
			}
		}

		zap.S().Debug("Finished updating cluster")
	}

	zap.S().Debug("Updating singles")

	for _, single := range cfg.Singles {
		args := []string{"apply", "-n", single.Namespace, "-f", "-"}
		if cfg.Arguments.Upgrade.DryRun {
			args = append(args, "--dry-run")
		}

		data, err := os.ReadFile(single.File)
		if err != nil {
			if cfg.Arguments.Upgrade.StopOnFirstError {
				zap.S().Fatalf("Failed to read single file %s: %v", single.File, err)
			}

			zap.S().Errorf("Failed to read single file %s: %v", single.File, err)
			continue
		}

		if single.UseCreate {
			args[0] = "create"
		}

		_, err = manager.ExecuteCommandStdin(bytes.NewReader(data), "kubectl", args...)
		if err != nil {
			if cfg.Arguments.Upgrade.StopOnFirstError {
				zap.S().Fatalf("Failed to upgrade single %s: %v", single.Name, err)
			}

			zap.S().Errorf("Failed to upgrade single %s: %v", single.Name, err)
			continue
		}
	}

	zap.S().Debug("Finished updating singles")
}
