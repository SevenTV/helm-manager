package cli

import (
	"errors"

	"github.com/seventv/helm-manager/argparse"
)

var UpdateCommand = Command{
	Name: "update",
	Help: "Update the manifest with the latest versions of the charts",
	Mode: CommandModeUpdate,
}

type Update struct {
	Name    string
	Version string
	List    bool
}

func UpdateCli(parser argparse.Command, args Arguments) Trigger {
	updateCmd := parser.NewCommand(UpdateCommand.Name, UpdateCommand.Help)

	updateNameFlag := updateCmd.String("", "name", &argparse.Options[string]{
		Required: false,
		Help:     "Name of the chart to update",
	})

	updateNamePos := updateCmd.StringPositional("name", &argparse.Options[string]{
		Required: false,
		Help:     "Name of the chart to update",
	})

	updateVersionFlag := updateCmd.String("", "version", &argparse.Options[string]{
		Required: false,
		Help:     "Version to update the chart to",
	})

	updateVersionPos := updateCmd.StringPositional("version", &argparse.Options[string]{
		Required: false,
		Help:     "Version to update the chart to",
	})

	updateListFlag := updateCmd.Flag("", "list", &argparse.Options[bool]{
		Required: false,
		Help:     "List all available versions of the chart",
	})

	return func(args *Arguments) error {
		if !updateCmd.Happened() {
			return nil
		}

		args.Mode = CommandModeUpdate
		args.Update.List = *updateListFlag

		args.Update.Name = *updateNameFlag
		if args.Update.Name == "" {
			args.Update.Name = *updateNamePos
		} else if *updateNamePos != "" {
			return errors.New("cannot specify name twice")
		}

		args.Update.Version = *updateVersionFlag
		if args.Update.Version == "" {
			args.Update.Version = *updateVersionPos
		} else if *updateVersionPos != "" {
			return errors.New("cannot specify version twice")
		}

		if args.NonInteractive {
			if !args.Update.List {
				if args.Update.Name == "" {
					return errors.New("name is required")
				}

				if args.Update.Version == "" {
					return errors.New("version is required")
				}
			}
		}

		return nil
	}
}
