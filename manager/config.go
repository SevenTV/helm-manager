package manager

import (
	"github.com/seventv/helm-manager/manager/cli"
	"github.com/seventv/helm-manager/manager/types"
	"github.com/seventv/helm-manager/manager/utils"
)

func GetConfig() types.Config {
	args := cli.ReadArguments()

	utils.SetupLogger(args.Debug, args.InTerminal)

	config, err := utils.ReadConfig(args.ManifestFile)
	if err != nil && err != utils.ErrorNotFound {
		utils.Fatal("Error reading config file: %s", err)
	} else if err != utils.ErrorNotFound {
		config.Exists = true
	}

	config.Arguments = &args

	utils.ValidateSingles(config)
	utils.ValidateCharts(config)
	utils.ValidateRepos(config)
	utils.ValidateEnv(config)

	return config
}
