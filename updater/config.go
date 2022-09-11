package updater

import (
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/mattn/go-colorable"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v2"
)

var ErrorNotFound = errors.New("not found")

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

func SetupLogger(config Config) {
	cfg := zap.NewProductionConfig()

	lvl := zap.NewAtomicLevelAt(zapcore.InfoLevel)
	if config.Arguments.Debug {
		lvl = zap.NewAtomicLevelAt(zapcore.DebugLevel)
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
	SetupLogger(Config{})

	config, err := ReadConfig("manifest.yaml")
	if err != nil && err != ErrorNotFound {
		zap.S().Fatalf("Error reading config file: %s", err)
	}

	if err == ErrorNotFound {
		zap.S().Warn("No config file found, using defaults")
		WriteConfig(config)
	}

	config.Arguments = ReadArguments()
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

func CreateRepoMap(cfg Config) map[string]Repo {
	repos := map[string]Repo{}
	for _, repo := range cfg.Repos {
		repos[repo.Name] = repo
	}

	return repos
}

func ValidateCharts(cfg Config) {
	charts := map[string]struct{}{}
	for idx, chart := range cfg.Charts {
		if _, ok := charts[chart.Name]; ok {
			zap.S().Fatalf("Duplicate chart name in config file %s", chart.Name)
		}

		if chart.ValuesFile == "" {
			chart.ValuesFile = fmt.Sprintf("%s-values.yaml", chart.Name)
		}

		if !path.IsAbs(chart.ValuesFile) {
			chart.ValuesFile = path.Join(cfg.Arguments.ValuesDir, chart.ValuesFile)
		}

		cfg.Charts[idx] = chart

		charts[chart.Name] = struct{}{}
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
