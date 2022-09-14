package upgrade

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/fatih/color"
	"github.com/seventv/helm-manager/manager/types"
	"github.com/seventv/helm-manager/manager/utils"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

func HandleChart(cfg types.Config, chart types.Chart, envMap map[string]string) (types.ChartUpgrade, bool) {
	const (
		LOCK_IDX     = 0
		VALUES_IDX   = 1
		DEFAULTS_IDX = 2
	)

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
					zap.S().Infof("%s [%s]\r", color.YellowString("Updating chart %s", chart.Name), color.CyanString("%s", stages[i%len(stages)]))
					i++
				case success := <-waiting:
					if success {
						zap.S().Infof("%s Chart updated %s", color.GreenString("✓"), chart.Name)
					} else {
						zap.S().Infof("%s Failed to update chart %s", color.RedString("✗"), chart.Name)
					}
					return
				}
			}
		} else {
			utils.Info("Updating Chart %s...", chart.Name)
			if <-waiting {
				utils.Info("Chart %s updated", chart.Name)
			} else {
				utils.Info("Failed to update chart %s", chart.Name)
			}
		}
	}()

	values, err := utils.ReadChartValues(chart)
	if err == utils.ErrorNotFound {
		err = nil
	}
	if err != nil {
		waiting <- false
		<-finished
		utils.Error("Unable to parse values file for %s", chart.Name)
		return types.ChartUpgrade{}, false
	}

	if !values.IsZero() && values.Kind != yaml.DocumentNode {
		waiting <- false
		<-finished
		utils.Error("Invalid values file for %s", chart.Name)
		return types.ChartUpgrade{}, false
	}

	defaultChartValues := utils.GetDefaultChartValues(chart)
	if defaultChartValues.IsZero() {
		waiting <- false
		<-finished
		utils.Error("No default values found for %s", chart.Name)
		return types.ChartUpgrade{}, false
	}

	if len(values.Content) == 0 {
		values = yaml.Node{
			Kind: yaml.DocumentNode,
			Content: []*yaml.Node{{
				Kind:    yaml.MappingNode,
				Content: []*yaml.Node{},
				Tag:     "!!map",
			}, {
				Kind:    yaml.MappingNode,
				Content: []*yaml.Node{},
				Tag:     "!!map",
			}, &defaultChartValues},
		}
	} else if len(values.Content) == 1 {
		merged := utils.RemoveYamlComments(*values.Content[0])

		values.Content = []*yaml.Node{{
			Kind:    yaml.MappingNode,
			Content: []*yaml.Node{},
			Tag:     "!!map",
		}, &merged, &defaultChartValues}
	} else if len(values.Content) == 2 {
		values.Content = append(values.Content, &defaultChartValues)
	}

	if len(values.Content) != 3 {
		waiting <- false
		<-finished
		utils.Error("Invalid values file for %s", chart.Name)
		return types.ChartUpgrade{}, false
	}

	oldLock := types.ChartLock{}
	values.Content[LOCK_IDX].Decode(&oldLock)

	if oldLock.Version != "" && oldLock.Version != chart.Version {
		// the file is outdated we need to first prune off the old values
		// and then add the new ones
		c := chart
		c.Version = oldLock.Version
		oldValues := utils.GetDefaultChartValues(c)
		if err != nil {
			waiting <- false
			<-finished
			utils.Error("Unable to read values file for %s @ ", chart.Name, oldLock.Version)
		}

		merged := utils.PruneYaml(oldValues, utils.MergeYaml(utils.RemoveYamlComments(*values.Content[DEFAULTS_IDX]), *values.Content[VALUES_IDX]))

		values.Content[VALUES_IDX] = &merged
		values.Content[DEFAULTS_IDX] = &defaultChartValues
	}

	{
		merged := utils.PruneYaml(defaultChartValues, utils.MergeYaml(utils.RemoveYamlComments(*values.Content[DEFAULTS_IDX]), *values.Content[VALUES_IDX]))
		merged.HeadComment = "## This section contains the non-default values for this chart.\n## If you want to change a value, add it here.\n## If you want to reset a value to default, remove it here.\n## If you want to reset all values to default, delete this entire section.\n## You can also modify the section below, any changes there will be reset however they will be copied into this section.\n\n"

		values.Content[VALUES_IDX] = &merged
		values.Content[DEFAULTS_IDX] = &defaultChartValues
	}

	// marshal the full version without comments
	var envSubbedChartValuesData []byte
	{
		// remove all comments from the full version
		envSubbedChartValuesData, err = utils.MarshalYaml(utils.ToDocument(utils.RemoveYamlComments(utils.MergeYaml(*values.Content[DEFAULTS_IDX], *values.Content[VALUES_IDX]))))
		if err != nil {
			waiting <- false
			<-finished
			utils.Error("Failed to marshal values for %s", chart.Name)
			return types.ChartUpgrade{}, false
		}

		// substitute the env variables into the full version without comments
		for env, value := range envMap {
			envSubbedChartValuesData = bytes.ReplaceAll(envSubbedChartValuesData, []byte(fmt.Sprintf("${%s}", env)), []byte(value))
		}
	}

	// create a new lock entry
	lock := types.ChartLock{
		Version: chart.Version,
		Chart:   chart.Chart,
	}

	{ // marshal the lock entry
		lockData, err := yaml.Marshal(lock)
		if err != nil {
			waiting <- false
			<-finished
			utils.Error("Failed to marshal lock for %s", chart.Name)
			return types.ChartUpgrade{}, false
		}

		lockNode := yaml.Node{}
		err = yaml.Unmarshal(lockData, &lockNode)
		if err != nil {
			waiting <- false
			<-finished
			utils.Error("Failed to unmarshal lock for %s", chart.Name)
			return types.ChartUpgrade{}, false
		}

		lockNode = utils.ConvertDocument(lockNode)
		lockNode.HeadComment = "## This section is automatically generated by helm-manager. DO NOT EDIT.\n\n"
		values.Content[LOCK_IDX] = &lockNode
	}

	// allow for showing only the values that changed from the defaults.
	chartValuesData, err := utils.MarshalYaml(values)
	if err != nil {
		waiting <- false
		<-finished
		utils.Error("Failed to marshal values for %s", chart.Name)
		return types.ChartUpgrade{}, false
	}

	// make sure the folder exists
	err = os.MkdirAll(path.Dir(chart.File), 0755)
	if err != nil {
		waiting <- false
		<-finished
		utils.Error("Failed to create folder for %s", chart.Name)
		return types.ChartUpgrade{}, false
	}

	err = os.WriteFile(chart.File, chartValuesData, 0644)
	if err != nil {
		waiting <- false
		<-finished
		utils.Error("Failed to write values for %s to %s", chart.Name, chart.File)
		return types.ChartUpgrade{}, false
	}

	waiting <- true
	<-finished

	return types.ChartUpgrade{
		Chart:            chart,
		ChartLock:        lock,
		OldLock:          oldLock,
		ValuesYaml:       chartValuesData,
		SubbedValuesYaml: envSubbedChartValuesData,
	}, true
}

func HandleUpgrade(cfg types.Config, upgrade types.ChartUpgrade) bool {
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
					zap.S().Infof("%s [%s]\r", color.YellowString("Upgrading chart %s", upgrade.Chart.Name), color.CyanString("%s", stages[i%len(stages)]))
					i++
				case success := <-waiting:
					if success {
						if cfg.Arguments.Upgrade.GenerateTemplate {
							zap.S().Infof("%s Chart template generated %s", color.GreenString("✓"), upgrade.Chart.Name)
						} else {
							zap.S().Infof("%s Chart upgraded %s", color.GreenString("✓"), upgrade.Chart.Name)
						}
					} else {
						if cfg.Arguments.Upgrade.GenerateTemplate {
							zap.S().Infof("%s Failed to generate chart template %s", color.RedString("✗"), upgrade.Chart.Name)
						} else {
							zap.S().Infof("%s Failed to upgrade for chart %s", color.RedString("✗"), upgrade.Chart.Name)
						}
					}
					return
				}
			}
		} else {
			utils.Info("Upgrading Chart %s...", upgrade.Chart.Name)
			if <-waiting {
				if cfg.Arguments.Upgrade.GenerateTemplate {
					utils.Info("Chart template generated %s", upgrade.Chart.Name)
				} else {
					utils.Info("Chart %s upgraded", upgrade.Chart.Name)
				}
			} else {
				if cfg.Arguments.Upgrade.GenerateTemplate {
					utils.Info("Failed to generate templates for chart %s", upgrade.Chart.Name)
				} else {
					utils.Info("Failed to upgrade chart %s", upgrade.Chart.Name)
				}
			}
		}
	}()

	chart := upgrade.Chart

	var args []string
	if cfg.Arguments.Upgrade.GenerateTemplate {
		args = []string{
			"template",
			chart.Name, chart.Chart,
			"--namespace", chart.Namespace,
			"--version", chart.Version,
			"--values", "-",
			"--create-namespace",
			"--include-crds",
		}
	} else {
		args = []string{
			"upgrade", "--install",
			chart.Name, chart.Chart,
			"--namespace", chart.Namespace,
			"--values", "-",
			"--version", chart.Version,
			"--create-namespace",
		}

		if cfg.Arguments.Upgrade.Wait {
			args = append(args, "--wait")
		}

		if cfg.Arguments.Upgrade.Atomic {
			args = append(args, "--atomic")
		}
	}

	if cfg.Arguments.Upgrade.DryRun {
		args = append(args, "--dry-run")
	}

	output, err := utils.ExecuteCommandStdin(bytes.NewReader(upgrade.SubbedValuesYaml), "helm", args...)
	if err != nil {
		waiting <- false
		<-finished
		utils.Error("Failed to upgrade chart %s\n%s", chart.Name, output)
		return false
	}

	if cfg.Arguments.Upgrade.GenerateTemplate {
		if cfg.Arguments.Upgrade.TemplateOutputDir != "" {
			err = os.MkdirAll(cfg.Arguments.Upgrade.TemplateOutputDir, 0755)
			if err != nil {
				waiting <- false
				<-finished
				utils.Error("Failed to create directory generated templates")
				return false
			}
		}

		err = os.WriteFile(path.Join(cfg.Arguments.Upgrade.TemplateOutputDir, fmt.Sprintf("%s-template.yaml", chart.Name)), output, 0644)
		if err != nil {
			waiting <- false
			<-finished
			utils.Error("Failed to write template file for %s", chart.Name)
			return false
		}
	}

	waiting <- true
	<-finished

	return true
}

func HandleSingle(cfg types.Config, single types.Single) bool {
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

	args := []string{"apply", "-n", single.Namespace, "-f", "-"}
	if cfg.Arguments.Upgrade.DryRun {
		args = append(args, "--dry-run")
	}

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

	_, err = utils.ExecuteCommandStdin(bytes.NewReader(data), "kubectl", args...)
	if err != nil {
		waiting <- false
		<-finished
		zap.S().Errorf("Failed to upgrade single %s: %v", single.Name, err)
		return false
	}

	waiting <- true
	<-finished

	return true
}
