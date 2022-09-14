package cli

import "github.com/seventv/helm-manager/argparse"

var InitCommand = Command{
	Name: "init",
	Help: "Initialize a new manifest file",
	Mode: CommandModeInit,
}

func InitCli(parser argparse.Command, args Arguments) Trigger {
	initCmd := parser.NewCommand(InitCommand.Name, InitCommand.Help)

	return func(args *Arguments) error {
		if !initCmd.Happened() {
			return nil
		}

		args.Mode = CommandModeInit

		return nil
	}
}
