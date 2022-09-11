package main

import (
	"errors"
	"os"

	"github.com/mattn/go-colorable"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v2"
)

var ErrorNotFound = errors.New("not found")

type Config struct {
	LogLevel   string   `yaml:"log_level"`
	DryRun     bool     `yaml:"dry_run"`
	Repos      []Repo   `yaml:"repos"`
	AllowedEnv []string `yaml:"allowed_env"`
	Charts     []Chart  `yaml:"charts"`
}

type Repo struct {
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
}

type Chart struct {
	Name          string `yaml:"name"`
	Chart         string `yaml:"chart"`
	Namespace     string `yaml:"namespace"`
	Version       string `yaml:"version"`
	ValuesFile    string `yaml:"values_file"`
	ShowDiff      bool   `yaml:"show_diff"`
	AlwaysUpgrade bool   `yaml:"always_upgrade"`
}

type ChartUpgrade struct {
	Chart            Chart
	ChartLock        ChartLock
	AlwaysWrite      bool
	ValuesYaml       []byte
	SubbedValuesYaml []byte
}

type ConfigLock struct {
	Charts []ChartLock `yaml:"charts"`
}

type ChartLock struct {
	Name    string `yaml:"name"`
	Chart   string `yaml:"chart"`
	Version string `yaml:"version"`
	Hash    string `yaml:"hash"`
}

func ReadConfig(path string) (Config, error) {
	config := Config{}
	data, err := os.ReadFile(path)
	if err != nil {
		return config, ErrorNotFound
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return config, err
	}

	return config, nil
}

func ReadConfigLock(path string) (ConfigLock, error) {
	config := ConfigLock{}
	data, err := os.ReadFile(path)
	if err != nil {
		return ConfigLock{}, ErrorNotFound
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return ConfigLock{}, err
	}

	return config, nil
}

func SetupLogger(config Config) {
	cfg := zap.NewProductionConfig()
	lvl, err := zap.ParseAtomicLevel(config.LogLevel)
	if err != nil {
		lvl = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	}

	cfg.Level = lvl
	cfg.Encoding = "console"
	cfg.EncoderConfig = zap.NewDevelopmentEncoderConfig()
	cfg.EncoderConfig.CallerKey = ""
	cfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05,000")
	cfg.EncoderConfig.ConsoleSeparator = " "
	cfg.EncoderConfig.StacktraceKey = ""
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	logger := zap.New(zapcore.NewCore(
		zapcore.NewConsoleEncoder(cfg.EncoderConfig),
		zapcore.AddSync(colorable.NewColorableStdout()),
		cfg.Level,
	))

	colorable.EnableColorsStdout(nil)

	zap.ReplaceGlobals(logger)
}

func GetConfig() Config {
	SetupLogger(Config{LogLevel: "info"})

	config, err := ReadConfig("manifest.yaml")
	if err == ErrorNotFound {
		config = Config{LogLevel: "info"}
		zap.S().Warn("No config file found, using defaults")
		err = nil
		WriteConfig(config)
	}
	if err != nil {
		zap.S().Fatal(err)
	}

	SetupLogger(config)

	return config
}

func WriteConfig(cfg Config) {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		zap.S().Fatal(err)
	}

	err = os.WriteFile("manifest.yaml", data, 0644)
	if err != nil {
		zap.S().Fatal(err)
	}
}

func GetLock() map[string]ChartLock {
	lock, err := ReadConfigLock("manifest-lock.yaml")
	if err == ErrorNotFound {
		zap.S().Warn("No lock file found, first time running?")
		err = nil
	}

	if err != nil {
		zap.S().Fatal(err)
	}

	lockMap := map[string]ChartLock{}
	for _, chart := range lock.Charts {
		if _, ok := lockMap[chart.Name]; ok {
			zap.S().Fatalf("Duplicate chart name in lock file %s", chart.Name)
		}

		lockMap[chart.Name] = chart
	}

	return lockMap
}

func CreateRepoMap(cfg Config) map[string]Repo {
	repos := map[string]Repo{}
	for _, repo := range cfg.Repos {
		repos[repo.Name] = repo
	}

	return repos
}

func ValidateCharts(cfg Config) {
	charts := map[string]Chart{}
	for _, chart := range cfg.Charts {
		if _, ok := charts[chart.Name]; ok {
			zap.S().Fatalf("Duplicate chart name in config file %s", chart.Name)
		}

		charts[chart.Name] = chart
	}
}

func CreateEnvMap(cfg Config) map[string]string {
	envMap := map[string]string{}
	for _, env := range cfg.AllowedEnv {
		if val, ok := os.LookupEnv(env); ok {
			envMap[env] = val
		}
	}
	return envMap
}
