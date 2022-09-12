package cli

import "github.com/seventv/helm-manager/argparse"

func InitCli(parser argparse.Command) Trigger {
	initCmd := parser.NewCommand("init", "Initialize a new manifest")

	return func(args *Arguments) {
		if !initCmd.Happened() {
			return
		}

		args.Mode = CommandModeInit
	}
}
