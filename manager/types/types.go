package types

import (
	"github.com/seventv/helm-manager/manager/cli"
)

type Config struct {
	Repos      []Repo   `yaml:"repos"`
	AllowedEnv []string `yaml:"allowed_env"`
	Charts     []Chart  `yaml:"charts"`
	Singles    []Single `yaml:"singles"`

	Exists    bool          `yaml:"-"`
	Arguments cli.Arguments `yaml:"-"`
}

type Single struct {
	Name      string `yaml:"name"`
	Namespace string `yaml:"namespace"`
	File      string `yaml:"file"`
	UseCreate bool   `yaml:"use_create"`
}

type Repo struct {
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
}

type Chart struct {
	Name      string `yaml:"name"`
	Chart     string `yaml:"chart"`
	Namespace string `yaml:"namespace"`
	Version   string `yaml:"version"`

	File string `yaml:"-"`
}

type ChartUpgrade struct {
	Chart     Chart
	ChartLock ChartLock
	OldLock   ChartLock

	ValuesYaml       []byte
	SubbedValuesYaml []byte
}

type ChartLock struct {
	Chart   string `yaml:"chart"`
	Version string `yaml:"version"`
}
