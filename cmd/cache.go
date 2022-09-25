package cmd

import (
	"os"
	"strings"

	"github.com/seventv/helm-manager/external"
	"github.com/seventv/helm-manager/logger"
	"github.com/seventv/helm-manager/types"
	"github.com/seventv/helm-manager/utils"
)

var NamespaceFuture = types.FutureFromFuncErr(func() ([]string, error) {
	done := utils.Loader(utils.LoaderOptions{
		FetchingText: "Fetching k8s namespaces",
		SuccessText:  "Fetched k8s namespaces",
		FailureText:  "Failed to fetch k8s namespaces",
	})

	namespaces, err := external.Kubectl.GetNamespaces()
	done(err == nil)

	return namespaces, err
})

var HelmRepoFuture = types.FutureFromFuncErr(func() ([]types.HelmRepo, error) {
	done := utils.Loader(utils.LoaderOptions{
		FetchingText: "Fetching helm repos",
		SuccessText:  "Fetched helm repos",
		FailureText:  "Failed to fetch helm repos",
	})
	repos, err := external.Helm.ListRepos()
	done(err == nil)
	return repos, err
})

var HelmReleaseFuture = types.FutureFromFuncErr(func() ([]types.HelmRelease, error) {
	done := utils.Loader(utils.LoaderOptions{
		FetchingText: "Fetching helm releases",
		SuccessText:  "Fetched helm releases",
		FailureText:  "Failed to fetch helm releases",
	})
	releases, err := external.Helm.ListReleases()
	done(err == nil)
	return releases, err
})

var HelmChartsFuture = types.FutureFromFuncErr(func() ([]types.HelmChartMulti, error) {
	_, err := UpdateHelmRepoFuture.Get()
	if err != nil {
		return nil, err
	}

	done := utils.Loader(utils.LoaderOptions{
		FetchingText: "Fetching helm charts",
		SuccessText:  "Fetched helm charts",
		FailureText:  "Failed to fetch helm charts",
	})
	charts, err := external.Helm.ListCharts()
	done(err == nil)

	charts = append(charts, LocalChartsFuture.GetOrPanic()...)

	return charts, err
})

var LocalChartsFuture = types.FutureFromFunc(func() []types.HelmChartMulti {
	chartMp := make(map[string]*types.HelmChartMulti, len(Manifest.LocalCharts))

	for _, pth := range Manifest.LocalCharts {
		chart := types.HelmChart{
			LocalPath: string(pth),
			IsLocal:   true,
		}

		if err := utils.ParseLocalChartYaml(&chart); err != nil {
			logger.Errorf("Failed to parse local chart %s: %s", pth, err)
			continue
		}

		if c, ok := chartMp[chart.RepoName]; ok {
			c.Versions = append(c.Versions, types.HelmChartMultiVersion(chart))
		} else {
			chartMp[chart.RepoName] = &types.HelmChartMulti{
				HelmChart: chart,
				Versions:  []types.HelmChartMultiVersion{types.HelmChartMultiVersion(chart)},
			}
		}
	}

	charts := make([]types.HelmChartMulti, 0, len(chartMp))
	for _, c := range chartMp {
		charts = append(charts, *c)
	}

	return charts
})

var UpdateHelmRepoFuture = types.FutureFromFuncErr(func() (bool, error) {
	done := utils.Loader(utils.LoaderOptions{
		FetchingText: "Updating helm repos",
		SuccessText:  "Updated helm repos",
		FailureText:  "Failed to update helm repos",
	})
	_, err := external.Helm.UpdateRepos()
	done(err == nil)
	return err == nil, err
})

type HelmReleaseChart struct {
	types.HelmRelease
	Chart types.HelmChartMulti
}

var HelmReleaseChartFuture = types.FutureFromFuncErr(func() ([]HelmReleaseChart, error) {
	charts, err := HelmChartsFuture.Get()
	if err != nil {
		return nil, err
	}

	releases, err := HelmReleaseFuture.Get()
	if err != nil {
		return nil, err
	}

	ret := []HelmReleaseChart{}

	for _, release := range releases {
		multiChart := types.HelmChartMultiArray(charts).FindChart(release.Chart())
		chart := multiChart.FindVersion(release.Version())
		if chart.Version != "" {
			multiChart.HelmChart = types.HelmChart(chart)
			ret = append(ret, HelmReleaseChart{
				HelmRelease: release,
				Chart:       multiChart,
			})
		}
	}

	return ret, nil
})

var EnvMapFuture = types.FutureFromFunc(func() map[string]string {
	envMap := map[string]string{}
	allowedEnvMp := map[string]bool{}
	for _, env := range Manifest.AllowedEnv {
		allowedEnvMp[env.String()] = true
		if value, ok := os.LookupEnv(env.String()); ok {
			envMap[strings.ToUpper(env.String())] = value
		}
	}

	envData, err := os.ReadFile(Args.EnvFile)
	if Args.EnvFile != ".env" && Args.EnvFile != "" && err != nil {
		logger.Fatalf("Failed to read env file \"%s\"\n  %v", Args.EnvFile, err)
	}

	for _, line := range strings.Split(string(envData), "\n") {
		if line != "" {
			parts := strings.SplitN(line, "=", 2)
			if !allowedEnvMp[strings.ToUpper(parts[0])] {
				logger.Warnf("Env variable %s is not allowed but specified in env file", strings.ToUpper(parts[0]))
			} else {
				envMap[strings.ToUpper(parts[0])] = parts[1]
			}
		}
	}

	return envMap
})
