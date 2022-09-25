package utils

import (
	"gopkg.in/yaml.v3"
)

type HelmChart struct {
	Name       string `json:"name"`
	Repo       string `json:"-"`
	Version    string `json:"version"`
	AppVersion string `json:"app_version"`
	Desctipton string `json:"description"`
}

func (h HelmChart) FullName() string {
	return h.Repo + "/" + h.Name
}

type HelmChartParsed struct {
	HelmChart

	Children []HelmChart
}

type DeployedHelmChart struct {
	Name       string `json:"name"`
	Namespace  string `json:"namespace"`
	Chart      string `json:"chart"`
	AppVersion string `json:"app_version"`

	Version string    `json:"-"`
	Values  yaml.Node `json:"-"`
}
