package cmd

import (
	"os"
	"strings"

	"github.com/seventv/helm-manager/v2/external"
	"github.com/seventv/helm-manager/v2/logger"
	"github.com/seventv/helm-manager/v2/types"
	"github.com/seventv/helm-manager/v2/utils"
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
			LocalPath: utils.MergeRelativePath(Args.Context, string(pth)),
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
	allowedEnvMp["HELM_MANAGER_NAME"] = true
	allowedEnvMp["HELM_MANAGER_CONTEXT_NAME"] = true

	for _, env := range Manifest.AllowedEnv {
		if value, ok := os.LookupEnv(env.String()); ok {
			allowedEnvMp[env.String()] = false
			envMap[strings.ToUpper(env.String())] = value
		} else {
			allowedEnvMp[env.String()] = true
		}
	}

	envData, err := os.ReadFile(utils.MergeRelativePath(Args.Context, Args.EnvFile))
	if Args.EnvFile != ".env" && Args.EnvFile != "" && err != nil {
		logger.Fatalf("Failed to read env file \"%s\"\n  %v", Args.EnvFile, err)
	}

	for _, line := range strings.Split(string(envData), "\n") {
		if line != "" {
			parts := strings.SplitN(line, "=", 2)
			if val, ok := allowedEnvMp[strings.ToUpper(parts[0])]; !ok {
				logger.Warnf("Env variable %s is not allowed but specified in env file.", strings.ToUpper(parts[0]))
			} else {
				if !val {
					logger.Warnf("Env variable %s is specified multiple times.", strings.ToUpper(parts[0]))
				}

				allowedEnvMp[strings.ToUpper(parts[0])] = false
				envMap[strings.ToUpper(parts[0])] = parts[1]
			}
		}
	}

	delete(allowedEnvMp, "HELM_MANAGER_NAME")
	delete(allowedEnvMp, "HELM_MANAGER_CONTEXT_NAME")

	for env, unused := range allowedEnvMp {
		if unused {
			logger.Fatalf("Env variable %s is not specified.", strings.ToUpper(env))
		}
	}

	if Manifest.Name != "" && envMap["HELM_MANAGER_NAME"] != Manifest.Name {
		logger.Fatalf("Env variable HELM_MANAGER_NAME is not equal to manifest name.")
	} else if Manifest.Name == "" {
		logger.Warn("Manifest name is not specified in manifest file this is not recommended.")
	}

	context, err := external.Kubectl.GetCurrentContext()
	if err != nil {
		logger.Fatalf("Failed to get current kubectl context: %v", err)
	}

	if envMap["HELM_MANAGER_CONTEXT_NAME"] != "" && envMap["HELM_MANAGER_CONTEXT_NAME"] != context {
		logger.Fatalf("Env variable HELM_MANAGER_CONTEXT_NAME is not equal to current kubectl context.")
	} else if envMap["HELM_MANAGER_CONTEXT_NAME"] == "" {
		logger.Warn("Env variable HELM_MANAGER_CONTEXT_NAME is not specified this is not recommended.")
	}

	return envMap
})
