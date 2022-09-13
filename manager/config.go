package manager

import (
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/mattn/go-colorable"
	"github.com/seventv/helm-manager/manager/cli"
	"github.com/seventv/helm-manager/manager/types"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v2"
)

var ErrorNotFound = errors.New("not found")

func ReadConfig(path string) (types.Config, error) {
	config := types.Config{}
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

func SetupLogger(debug bool) {
	cfg := zap.NewProductionConfig()

	cfg.Encoding = "console"
	cfg.EncoderConfig = zap.NewDevelopmentEncoderConfig()
	cfg.EncoderConfig.CallerKey = ""
	cfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05,000")
	cfg.EncoderConfig.ConsoleSeparator = " "
	cfg.EncoderConfig.StacktraceKey = ""
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	lvl := zap.NewAtomicLevelAt(zapcore.InfoLevel)
	if debug {
		lvl = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	} else {
		cfg.EncoderConfig.TimeKey = ""
	}

	logger := zap.New(zapcore.NewCore(
		zapcore.NewConsoleEncoder(cfg.EncoderConfig),
		zapcore.AddSync(colorable.NewColorableStderr()),
		lvl,
	))

	colorable.EnableColorsStdout(nil)

	zap.ReplaceGlobals(logger)
}

func GetConfig() types.Config {
	args := cli.ReadArguments()

	SetupLogger(args.Debug)

	config, err := ReadConfig(args.ManifestFile)
	if err != nil && err != ErrorNotFound {
		zap.S().Fatalf("Error reading config file: %s", err)
	} else if err != ErrorNotFound {
		config.Exists = true
	}

	config.Arguments = args

	ValidateSingles(config)
	ValidateCharts(config)
	ValidateRepos(config)
	ValidateEnv(config)

	return config
}

func WriteConfig(cfg types.Config) {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		zap.S().Fatal("Error marshalling config file")
	}

	err = os.WriteFile(cfg.Arguments.ManifestFile, data, 0644)
	if err != nil {
		zap.S().Fatal("Error writing config file")
	}
}

func CreateRepoMap(cfg types.Config) map[string]types.Repo {
	repos := map[string]types.Repo{}
	for _, repo := range cfg.Repos {
		if _, ok := repos[repo.Name]; ok {
			zap.S().Fatalf("Duplicate repo name in manifest file: %s", repo.Name)
		}

		repos[repo.Name] = repo
	}

	return repos
}

func ValidateCharts(cfg types.Config) {
	charts := map[string]struct{}{}
	for idx, chart := range cfg.Charts {
		if _, ok := charts[chart.Name]; ok {
			zap.S().Fatalf("Duplicate chart name in manifest file: %s", chart.Name)
		}

		chart.File = path.Join(cfg.Arguments.WorkingDir, "charts", fmt.Sprintf("%s.yaml", chart.Name))

		cfg.Charts[idx] = chart

		charts[chart.Name] = struct{}{}
	}
}

func ValidateEnv(cfg types.Config) {
	mp := map[string]struct{}{}
	for _, env := range cfg.AllowedEnv {
		if _, ok := mp[env]; ok {
			zap.S().Fatalf("Duplicate allowed env variable in manifest file: %s", env)
		}

		mp[env] = struct{}{}
	}
}

func CreateSinglesMap(cfg types.Config) map[string]types.Single {
	singles := map[string]types.Single{}
	for _, single := range cfg.Singles {
		if _, ok := singles[single.Name]; ok {
			zap.S().Fatalf("Duplicate single name in manifest file: %s", single.Name)
		}

		singles[single.Name] = single
	}

	return singles
}

func ValidateRepos(cfg types.Config) {
	CreateRepoMap(cfg)
}

func ValidateSingles(cfg types.Config) {
	CreateSinglesMap(cfg)
}

func CreateEnvMap(cfg types.Config) map[string]string {
	envMap := map[string]string{}
	for _, env := range cfg.AllowedEnv {
		if val, ok := os.LookupEnv(env); ok {
			envMap[env] = val
		}
	}
	return envMap
}
