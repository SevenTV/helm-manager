package external

import (
	"bytes"
	"encoding/json"

	"github.com/seventv/helm-manager/types"
	"github.com/seventv/helm-manager/utils"
	"gopkg.in/yaml.v3"
)

type _helm struct{}

var Helm = _helm{}

func (_helm) AddRepo(repo types.ManifestRepo) ([]byte, error) {
	return utils.ExecuteCommand("helm", "repo", "add", repo.Name, repo.URL)
}

func (_helm) UpdateRepos() ([]byte, error) {
	return utils.ExecuteCommand("helm", "repo", "update")
}

func (_helm) ListRepos() ([]types.HelmRepo, error) {
	resp, err := utils.ExecuteCommand("helm", "repo", "list", "-o", "json")
	if err != nil {
		return nil, err
	}

	var repos []types.HelmRepo
	if err = json.Unmarshal(resp, &repos); err != nil {
		return nil, err
	}

	return repos, nil
}

func (_helm) ListReleases() ([]types.HelmRelease, error) {
	resp, err := utils.ExecuteCommand("helm", "list", "-o", "json", "--all-namespaces", "--deployed")
	if err != nil {
		return nil, err
	}

	var releases []types.HelmRelease
	if err = json.Unmarshal(resp, &releases); err != nil {
		return nil, err
	}

	return releases, nil
}

func (_helm) ListCharts() ([]types.HelmChartMulti, error) {
	resp, err := utils.ExecuteCommand("helm", "search", "repo", "--output", "json", "--versions")
	if err != nil {
		return nil, err
	}

	helmCharts := []types.HelmChart{}
	err = json.Unmarshal(resp, &helmCharts)
	if err != nil {
		return nil, err
	}

	helmChartMulti := map[string]*types.HelmChartMulti{}
	for _, helmChart := range helmCharts {
		if helmChartMulti[helmChart.RepoName] == nil {
			helmChartMulti[helmChart.RepoName] = &types.HelmChartMulti{
				HelmChart: helmChart,
			}
		}

		helmChartMulti[helmChart.RepoName].Versions = append(helmChartMulti[helmChart.RepoName].Versions, types.HelmChartMultiVersion(helmChart))
	}

	helmChartMultiList := make([]types.HelmChartMulti, 0, len(helmChartMulti))
	for _, helmChart := range helmChartMulti {
		helmChartMultiList = append(helmChartMultiList, *helmChart)
	}

	return helmChartMultiList, nil
}

func (_helm) GetReleaseValues(release types.HelmRelease) (*yaml.Node, error) {
	data, err := utils.ExecuteCommand("helm", "get", "values", release.Name, "-n", release.Namespace, "-o", "yaml")
	if err != nil {
		return nil, err
	}

	node, err := utils.ParseYaml(data)
	if err != nil {
		return nil, err
	}

	return node, err
}

func (_helm) GetDefaultChartValues(chart types.HelmChart) (*yaml.Node, error) {
	args := []string{
		"show",
		"values",
		chart.HelmName(),
	}
	if !chart.IsLocal && chart.Version != "" {
		args = append(args, "--version", chart.Version)
	}
	data, err := utils.ExecuteCommand("helm", args...)
	if err != nil {
		return nil, err
	}

	node, err := utils.ParseYaml(data)
	if err != nil {
		return nil, err
	}

	return node, err
}

func (_helm) GetDeployedReleaseValues(release types.HelmRelease) (*yaml.Node, error) {
	data, err := utils.ExecuteCommand("helm", "get", "values", release.Name, "-n", release.Namespace, "-o", "yaml", "--all")
	if err != nil {
		return nil, err
	}

	node, err := utils.ParseYaml(data)
	if err != nil {
		return nil, err
	}

	return node, err
}

func (_helm) UpgradeRelease(release types.ManifestRelease, chart types.HelmChart, values []byte, dryRun bool, debug bool) ([]byte, error) {
	args := []string{
		"upgrade",
		"--install",
		"--namespace",
		release.Namespace,
		"--create-namespace",
		release.Name,
		chart.HelmName(),
		"-f",
		"-",
	}
	if !chart.IsLocal && chart.Version != "" {
		args = append(args, "--version", chart.Version)
	}

	if dryRun {
		args = append(args, "--dry-run")
	}
	if debug {
		args = append(args, "--debug")
	}

	return utils.ExecuteCommandStdin(bytes.NewReader(values), "helm", args...)
}

func (_helm) UninstallRelease(release types.ManifestRelease, dryRun bool) ([]byte, error) {
	args := []string{
		"uninstall",
		"--namespace",
		release.Namespace,
		release.Name,
	}

	if dryRun {
		args = append(args, "--dry-run")
	}

	return utils.ExecuteCommand("helm", args...)
}

func (_helm) RemoveRepo(repo types.HelmRepo) ([]byte, error) {
	return utils.ExecuteCommand("helm", "repo", "remove", repo.Name)
}
