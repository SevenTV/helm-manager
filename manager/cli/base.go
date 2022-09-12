package cli

import (
	"path"

	"github.com/seventv/helm-manager/argparse"
)

type Arguments struct {
	WorkingDir   string
	ManifestFile string

	Debug   bool
	Mode    CommandMode
	Upgrade Upgrade
	Add     Add
	Remove  Remove
}

func BaseCli(parser argparse.Parser) Trigger {
	debugFlag := parser.Flag("", "debug", &argparse.Options[bool]{
		Required: false,
		Help:     "Enable debug logging",
	})
	manifestFileFlag := parser.String("m", "manifest", &argparse.Options[string]{
		Required: false,
		Help:     "The manifest file to use",
		Default:  "manifest.yaml",
	})

	return func(args *Arguments) {
		args.Debug = *debugFlag
		args.ManifestFile = *manifestFileFlag
		args.WorkingDir = path.Dir(args.ManifestFile)
	}
}
