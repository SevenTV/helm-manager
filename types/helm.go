package types

import "strings"

type HelmChartArray []HelmChart

type HelmChartMultiArray []HelmChartMulti

func (h HelmChartMultiArray) FindChart(chart string) HelmChartMulti {
	chart = strings.ToLower(chart)

	for _, c := range h {
		if strings.ToLower(c.RepoName) == chart {
			return c
		}
	}

	return HelmChartMulti{}
}

func (h HelmChartArray) ToHelmChartMulti() HelmChartMultiArray {
	mp := make(map[string]*HelmChartMulti)

	for _, chart := range h {
		if _, ok := mp[chart.RepoName]; !ok {
			mp[chart.RepoName] = &HelmChartMulti{
				HelmChart: chart,
			}
		}

		mp[chart.RepoName].Versions = append(mp[chart.RepoName].Versions, HelmChartMultiVersion(chart))
	}

	result := make([]HelmChartMulti, 0, len(mp))
	for _, chart := range mp {
		result = append(result, *chart)
	}

	return result
}

type HelmChartMulti struct {
	HelmChart
	Versions []HelmChartMultiVersion
}

type HelmChartMultiVersion HelmChart

func (h HelmChartMulti) FindVersion(version string) HelmChartMultiVersion {
	for _, v := range h.Versions {
		if v.Version == version {
			return v
		}
	}

	return HelmChartMultiVersion{}
}

type HelmChart struct {
	RepoName    string `json:"name"`
	Version     string `json:"version"`
	AppVersion  string `json:"app_version"`
	Description string `json:"description"`
}

func (h HelmChart) Repo() string {
	split := strings.SplitN(h.RepoName, "/", 2)
	if len(split) == 2 {
		return split[0]
	}

	return ""
}

func (h HelmChart) Name() string {
	split := strings.SplitN(h.RepoName, "/", 2)
	if len(split) == 2 {
		return split[1]
	}

	return split[0]
}

func (h HelmChart) String() string {
	return h.RepoName
}

func (h HelmChartMultiVersion) Repo() string {
	return HelmChart(h).Repo()
}

func (h HelmChartMultiVersion) Name() string {
	return HelmChart(h).Name()
}

type HelmRepo struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type HelmRelease struct {
	Name         string `json:"name"`
	Namespace    string `json:"namespace"`
	Revision     string `json:"revision"`
	Updated      string `json:"updated"`
	Status       string `json:"status"`
	ChartVersion string `json:"chart"`
	AppVersion   string `json:"app_version"`
}

func (h HelmRelease) String() string {
	return h.Name
}

func (h HelmRelease) Chart() string {
	idx := strings.LastIndex(h.ChartVersion, "-")
	if idx == -1 {
		return h.ChartVersion
	}

	return h.ChartVersion[:idx]
}

func (h HelmRelease) Version() string {
	idx := strings.LastIndex(h.ChartVersion, "-")
	if idx == -1 {
		return ""
	}

	return h.ChartVersion[idx+1:]
}
