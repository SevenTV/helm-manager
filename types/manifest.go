package types

import (
	"fmt"
	"strings"
)

var GlobalManifest = &Manifest{}

type Manifest struct {
	Repos      []ManifestRepo     `yaml:"repos"`       // Helm repos
	AllowedEnv []SelectableString `yaml:"allowed_env"` // Allowed environment variables
	Releases   []ManifestRelease  `yaml:"releases"`    // Helm releases
	Singles    []ManifestSingle   `yaml:"singles"`     // Single files

	Exists bool `yaml:"-"` // Whether the manifest exists
}

func (m Manifest) Validate() bool {
	repoMap := make(map[string]bool)
	for _, repo := range m.Repos {
		if repoMap[repo.Name] {
			return false
		}

		repoMap[repo.Name] = true
	}

	releaseMap := make(map[string]bool)
	for _, release := range m.Releases {
		if releaseMap[release.Name] {
			return false
		}

		releaseMap[release.Name] = true
	}

	singleMap := make(map[string]bool)
	for _, single := range m.Singles {
		if singleMap[single.Name] {
			return false
		}

		singleMap[single.Name] = true
	}

	return true
}

func (m Manifest) RepoByName(name string) ManifestRepo {
	name = strings.ToLower(name)
	for _, repo := range m.Repos {
		if strings.ToLower(repo.Name) == name {
			return repo
		}
	}

	return ManifestRepo{}
}

func (m Manifest) ReleaseByName(name string) ManifestRelease {
	r, _ := m.ReleaseIdxByName(name)
	return r
}

func (m Manifest) ReleaseIdxByName(name string) (ManifestRelease, int) {
	name = strings.ToLower(name)

	for idx, release := range m.Releases {
		if strings.ToLower(release.Name) == name {
			return release, idx
		}
	}

	return ManifestRelease{}, -1
}

func (m Manifest) SingleByName(name string) ManifestSingle {
	for _, single := range m.Singles {
		if strings.ToLower(single.Name) == name {
			return single
		}
	}

	return ManifestSingle{}
}

type ManifestRepo struct {
	Name string `yaml:"name"` // Name of the repo
	URL  string `yaml:"url"`  // URL of the repo
}

func (m ManifestRepo) String() string {
	return m.Name
}

type ManifestRelease struct {
	Name      string        `yaml:"name"`      // the release name (required)
	Namespace string        `yaml:"namespace"` // the namespace the release is installed in (defaults to "default")
	Chart     ManifestChart `yaml:"chart"`     // The chart to install (required)
}

func (m ManifestRelease) String() string {
	return m.Name
}

type ManifestChart struct {
	Name       string `yaml:"name"`        // Name of the chart
	Version    string `yaml:"version"`     // Version of the chart
	AppVersion string `yaml:"app_version"` // AppVersion the chart covers
	Repo       string `yaml:"repo"`        // Repo the chart is in from the manifest
}

func (m ManifestChart) String() string {
	return m.Name
}

func (m ManifestChart) RepoName() string {
	return fmt.Sprintf("%s/%s", m.Repo, m.Name)
}

type ManifestSingle struct {
	Name      string `yaml:"name"`       // Name of the single
	UseCreate bool   `yaml:"use_create"` // Use create instead of apply
	Namespace string `yaml:"namespace"`  // Namespace to install the single in (optional)
}

func (m ManifestSingle) String() string {
	return m.Name
}

type ReleaseLock struct {
	Chart   string `yaml:"chart"`
	Version string `yaml:"version"`
}
