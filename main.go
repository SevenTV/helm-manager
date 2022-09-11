package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

func HandleChart(chart Chart, cfg Config, lockMap map[string]ChartLock, envMap map[string]string) (*ChartUpgrade, bool) {
	values, err := ReadChartValues(chart)
	alwaysWrite := false

	if err == ErrorNotFound {
		err = nil
		alwaysWrite = true
		if _, ok := lockMap[chart.Name]; ok {
			zap.S().Warnf("Chart %s has no values file, but is in lock file", chart.Name)
			delete(lockMap, chart.Name)
		} else {
			zap.S().Infof("No values file found for %s, assuming first time running", chart.Name)
		}
	}
	if err != nil {
		zap.S().Error("Unable to parse values file for %s", chart.Name)
		return nil, false
	}

	nonDefaultChartValues := values

	// unmerge from the old version of the chart
	if chartLock, ok := lockMap[chart.Name]; ok {
		if chartLock.Version != chart.Version {
			zap.S().Infof("Chart %s version changed from %s to %s", chart.Name, chartLock.Version, chart.Version)
			c := chart
			c.Version = chartLock.Version
			vals := GetNonDefaultChartValues(c, nonDefaultChartValues)
			if vals.IsZero() {
				return nil, false
			}

			nonDefaultChartValues = vals
		}
	} else {
		zap.S().Infof("Chart %s not in lock file, assuming first time running", chart.Name)
	}

	// update from the new requested version
	defaultChartValues := GetDefaultChartValues(chart)
	if defaultChartValues.IsZero() {
		return nil, false
	}

	nonDefaultChartValues = PruneYaml(defaultChartValues, nonDefaultChartValues)

	// get the full version of the chart
	fullValues := MergeYaml(defaultChartValues, nonDefaultChartValues)
	// remove all comments from the full version
	fullValuesNoComments := RemoveYamlComments(fullValues)

	// marshal the full version without comments
	subbedValuesData, err := yaml.Marshal(&fullValuesNoComments)
	if err != nil {
		zap.S().Errorf("Failed to marshal values for %s", chart.Name)
		return nil, false
	}

	// substitute the env variables into the full version without comments
	for env, value := range envMap {
		subbedValuesData = bytes.ReplaceAll(subbedValuesData, []byte(fmt.Sprintf("${%s}", env)), []byte(value))
	}

	// create a new lock entry
	lock := ChartLock{
		Name:    chart.Name,
		Version: chart.Version,
		Chart:   chart.Chart,
		Hash:    hex.EncodeToString(Sum256(subbedValuesData)),
	}

	// allow for showing only the values that changed from the defaults.
	var commentedData []byte
	if chart.ShowDiff {
		commentedData, err = yaml.Marshal(nonDefaultChartValues)
	} else {
		commentedData, err = yaml.Marshal(fullValues)
	}
	if err != nil {
		zap.S().Errorf("Failed to marshal values for %s", chart.Name)
		return nil, false
	}

	if lock.Hash != lockMap[chart.Name].Hash || chart.AlwaysUpgrade {
		zap.S().Infof("Chart %s has changed, queued to upgrade.", chart.Name)
		return &ChartUpgrade{
			Chart:            chart,
			ChartLock:        lock,
			ValuesYaml:       commentedData,
			SubbedValuesYaml: subbedValuesData,
			AlwaysWrite:      alwaysWrite,
		}, true
	} else {
		zap.S().Infof("Chart %s has not changed, skipping.", chart.Name)
		err = os.WriteFile(chart.ValuesFile, commentedData, 0644)
		if err != nil {
			return nil, false
		}
	}

	return nil, true
}

func WriteLock(lockMap map[string]ChartLock) {
	newLock := ConfigLock{}
	for _, chart := range lockMap {
		newLock.Charts = append(newLock.Charts, chart)
	}

	buf := bytes.NewBuffer(nil)
	e := yaml.NewEncoder(buf)
	e.SetIndent(2)
	err := e.Encode(&newLock)
	if err != nil {
		zap.S().Fatal("Failed to marshal lock file")
	}
	err = ioutil.WriteFile("manifest-lock.yaml", buf.Bytes(), 0644)
	if err != nil {
		zap.S().Fatal("Failed to write lock file")
	}
}

func HandleUpgrade(upgrade ChartUpgrade, cfg Config) bool {
	if cfg.DryRun {
		zap.S().Infof("Dry run for %s", upgrade.Chart.Name)
	} else {
		zap.S().Infof("Upgrading %s", upgrade.Chart.Name)
	}

	chart := upgrade.Chart

	write := func() bool {
		err := os.MkdirAll(path.Dir(chart.ValuesFile), 0755)
		if err != nil {
			zap.S().Errorf("Failed to create directory for %s", chart.Name)
			return false
		}

		err = os.WriteFile(chart.ValuesFile, upgrade.ValuesYaml, 0644)
		if err != nil {
			zap.S().Errorf("Failed to write values file for %s", chart.Name)
			return false
		}

		err = os.WriteFile(upgrade.Chart.ValuesFile, upgrade.ValuesYaml, 0644)
		if err != nil {
			zap.S().Errorf("Failed to write values file for %s", upgrade.Chart.Name)
			return false
		}

		return true
	}

	var (
		output []byte
		err    error
	)
	if cfg.DryRun {
		output, err = ExecuteCommandStdin(bytes.NewReader(upgrade.SubbedValuesYaml), "helm", "upgrade", "--install", chart.Name, chart.Chart, "--namespace", chart.Namespace, "--values", "-", "--version", chart.Version, "--create-namespace", "--dry-run")
	} else {
		output, err = ExecuteCommandStdin(bytes.NewReader(upgrade.SubbedValuesYaml), "helm", "upgrade", "--install", chart.Name, chart.Chart, "--namespace", chart.Namespace, "--values", "-", "--version", chart.Version, "--create-namespace")
	}
	if err != nil {
		zap.S().Errorf("Failed to upgrade chart %s\n%s", chart.Name, output)
		if upgrade.AlwaysWrite {
			return write()
		}

		return false
	}

	zap.S().Infof("Successfully upgraded chart %s", chart.Name)

	return write()
}

func main() {
	cfg := GetConfig()

	zap.S().Infof("* Helm Manager Starting *")

	lockMap := GetLock()
	envMap := CreateEnvMap(cfg)

	ValidateCharts(cfg)
	UpdateRepos(cfg)

	newLockMap := map[string]ChartLock{}
	upgradeList := []ChartUpgrade{}

	if len(cfg.Charts) == 0 {
		zap.S().Warn("No charts to manage")
	} else {
		zap.S().Infof("%d charts to manage", len(cfg.Charts))
	}

	for _, chart := range cfg.Charts {
		chartUpgrade, success := HandleChart(chart, cfg, lockMap, envMap)
		if !success {
			if _, ok := lockMap[chart.Name]; ok {
				newLockMap[chart.Name] = lockMap[chart.Name]
			}
			continue
		} else if chartUpgrade != nil {
			newLockMap[chart.Name] = chartUpgrade.ChartLock
			upgradeList = append(upgradeList, *chartUpgrade)
		} else {
			newLockMap[chart.Name] = lockMap[chart.Name]
		}
	}

	for _, upgrade := range upgradeList {
		if !HandleUpgrade(upgrade, cfg) {
			delete(newLockMap, upgrade.Chart.Name)
		}
	}

	WriteLock(newLockMap)

	zap.S().Infof("* Helm Manager Finished *")
}
