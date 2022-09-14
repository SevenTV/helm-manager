package argparse

import "os"

type Parser interface {
	Command
	Parse(args []string) error
}

type RootParser struct {
	*Cmd
}

func NewParser(name string, help string) Parser {
	return &RootParser{
		Cmd: newCmd(name, help, nil),
	}
}

func (r *RootParser) Parse(args []string) error {
	if args[0] == os.Args[0] {
		args = args[1:]
	}

	return r.parse(args)
}
