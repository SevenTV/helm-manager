package add

import (
	"os"
	"path"

	"github.com/seventv/helm-manager/manager"
	"github.com/seventv/helm-manager/manager/types"
	"github.com/seventv/helm-manager/upgrade"
	"go.uber.org/zap"
)

func runAddChart(cfg types.Config) {
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
			zap.S().Fatalf("chart with name %s already exists", chart.Name)
		}
	}

	if !cfg.Arguments.Add.Chart.Overwrite {
		if _, err := os.Stat(chart.File); err == nil {
			zap.S().Fatalf("refusing to overwrite %s", chart.File)
		}
	}

	if inputFile != "" {
		if _, err := os.Stat(inputFile); err != nil {
			zap.S().Fatalf("could not find chart file: %s", chart.File)
		}

		data, err := os.ReadFile(inputFile)
		if err != nil {
			zap.S().Fatalf("failed to read chart input file %s: %v", inputFile, err)
		}

		if err = os.MkdirAll(path.Dir(chart.File), 0755); err != nil {
			zap.S().Fatalf("failed to create chart directory %s: %v", path.Dir(chart.File), err)
		}

		if err = os.WriteFile(chart.File, data, 0644); err != nil {
			zap.S().Fatalf("failed to write chart file %s: %v", chart.File, err)
		}
	} else {
		if _, err := os.Stat(chart.File); err == nil {
			err = os.Remove(chart.File)
			if err != nil {
				zap.S().Fatalf("failed to remove chart file %s: %v", chart.File, err)
			}
		}
	}

	_, success := upgrade.HandleChart(chart, map[string]string{}, true)
	if !success {
		os.Exit(1)
	}

	cfg.Charts = append(cfg.Charts, chart)

	manager.WriteConfig(cfg)

	zap.S().Infof("added chart %s, run `helm-manager upgrade` to install the chart", chart.Name)
}
