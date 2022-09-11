package updater

import "time"

type Config struct {
	Repos      []Repo   `yaml:"repos"`
	AllowedEnv []string `yaml:"allowed_env"`
	Charts     []Chart  `yaml:"charts"`

	Arguments CommandArgs `yaml:"-"`
}

type Repo struct {
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
}

type Chart struct {
	Name       string `yaml:"name"`
	Chart      string `yaml:"chart"`
	Namespace  string `yaml:"namespace"`
	Version    string `yaml:"version"`
	ValuesFile string `yaml:"values_file"`
}

type ChartUpgrade struct {
	Chart     Chart
	ChartLock ChartLock
	OldLock   ChartLock

	ValuesYaml       []byte
	SubbedValuesYaml []byte
}

type ChartLock struct {
	Chart   string    `yaml:"chart"`
	Version string    `yaml:"version"`
	Hash    string    `yaml:"hash"`
	Time    time.Time `yaml:"time"`
}
