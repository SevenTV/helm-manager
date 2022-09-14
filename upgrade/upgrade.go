package upgrade

import (
	"github.com/seventv/helm-manager/manager/types"
	"github.com/seventv/helm-manager/manager/utils"
)

func Run(cfg types.Config) {
	envMap := utils.CreateEnvMap(cfg)

	utils.UpdateRepos(cfg)

	newLockMap := map[string]types.ChartLock{}
	upgradeList := []types.ChartUpgrade{}

	if len(cfg.Charts) == 0 && len(cfg.Singles) == 0 {
		utils.Fatal("No charts or singles found in manifest")
	}

	for _, chart := range cfg.Charts {
		if cfg.Arguments.Upgrade.IgnoreChartsMap[chart.Name] {
			utils.Info("Skipping %s, ignored", chart.Name)
			continue
		}

		if len(cfg.Arguments.Upgrade.ChartWhitelist) != 0 && !cfg.Arguments.Upgrade.ChartWhitelist[chart.Name] {
			utils.Info("Skipping %s, not in whitelist", chart.Name)
			continue
		}

		chartUpgrade, success := HandleChart(cfg, chart, envMap)
		if !success {
			if cfg.Arguments.Upgrade.StopOnFirstError {
				utils.Fatal("Failed to upgrade chart values %s", chart.Name)
			}

			continue
		} else {
			newLockMap[chart.Name] = chartUpgrade.ChartLock
			upgradeList = append(upgradeList, chartUpgrade)
		}
	}

	if len(upgradeList) > 0 {
		for _, upgrade := range upgradeList {
			if !HandleUpgrade(cfg, upgrade) && cfg.Arguments.Upgrade.StopOnFirstError {
				utils.Fatal("Failed to upgrade chart %s, failing fast", upgrade.Chart.Name)
			}
		}
	}

	for _, single := range cfg.Singles {
		if !HandleSingle(cfg, single) && cfg.Arguments.Upgrade.StopOnFirstError {
			utils.Fatal("Failed to upgrade single %s, failing fast", single.Name)
		}
	}
}
