package remove

import (
	"strings"

	"github.com/seventv/helm-manager/manager"
	"github.com/seventv/helm-manager/manager/types"
	"go.uber.org/zap"
)

func runRemoveChart(cfg types.Config) {
	chartName := cfg.Arguments.Remove.Chart.Name

	var (
		chart types.Chart
		idx   = -1
	)
	for idx, chart = range cfg.Charts {
		if chart.Name == chartName {
			break
		}
		idx = -1
	}

	if idx == -1 {
		zap.S().Fatalf("chart %s was not found", chartName)
	}

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

	zap.S().Infof("removing chart %s", chartName)

	data, err := manager.ExecuteCommand("helm", args...)
	strData := string(data)
	if strings.Contains(strData, "Release not loaded") {
		err = nil
	}

	if err != nil {
		zap.S().Fatalf("failed to delete chart %s %v\n%s", chart.Name, err, strData)
	}

	zap.S().Infof("chart %s deleted", chart.Name)

	if cfg.Arguments.Remove.Chart.DryRun {
		zap.S().Infof("dry-run: not writing config")
		return
	}

	manager.WriteConfig(cfg)
}
