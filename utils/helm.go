package utils

import (
	"os"
	"path"

	"github.com/seventv/helm-manager/types"
	"gopkg.in/yaml.v3"
)

type helmChartYaml struct {
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	AppVersion  string `yaml:"appVersion"`
	Description string `yaml:"description"`
}

func ParseLocalChartYaml(chart *types.HelmChart) error {
	data, err := os.ReadFile(path.Join(chart.LocalPath, "Chart.yaml"))
	if err != nil {
		return err
	}

	values := helmChartYaml{}
	if err = yaml.Unmarshal(data, &values); err != nil {
		return err
	}

	chart.RepoName = values.Name
	chart.Version = values.Version
	chart.AppVersion = values.AppVersion
	chart.Description = values.Description

	return nil
}
