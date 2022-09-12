package argparse

import (
	"fmt"
	"os"
	"strings"

	"github.com/jinzhu/copier"
)

type flag struct {
	name      string
	shorthand string

	value   any
	options *optionsCaster

	consumesValue bool
	action        func()
	positional    bool

	used       bool
	multi      bool
	hideGlobal bool

	owner *Cmd
}

func (f *flag) String() string {
	if f.name != "" {
		return fmt.Sprintf("--%s", f.name)
	}

	return fmt.Sprintf("-%s", f.shorthand)
}

type Command interface {
	parse(args []string) error

	Flag(shorthand string, name string, options *Options[bool]) *bool
	FlagCounter(shorthand string, name string, options *Options[int]) *int

	String(shorthand string, name string, options *Options[string]) *string
	Int(shorthand string, name string, options *Options[int]) *int
	Float(shorthand string, name string, options *Options[float64]) *float64

	StringList(shorthand string, name string, options *OptionsList[string]) *[]string
	IntList(shorthand string, name string, options *OptionsList[int]) *[]int
	FloatList(shorthand string, name string, options *OptionsList[float64]) *[]float64

	IntPositional(name string, options *Options[int]) *int
	FloatPositional(name string, options *Options[float64]) *float64
	StringPositional(name string, options *Options[string]) *string

	Happened() bool
	NewCommand(name string, help string) Command
	Usage(msg string) string
	Parent() Command
}

type Cmd struct {
	name string

	flagKwargs []*flag
	flags      map[string]*flag
	shortHand  map[string]*flag

	flagsArgs     []*flag
	positionalIdx int

	commands      map[string]*Cmd
	help          string
	happened      bool
	helpRequested bool
	parent        *Cmd

	treeFlagsCache       []*flag
	parentShorthandCache map[string]*flag
	parentFlagsCache     map[string]*flag
}

func (c *Cmd) path() string {
	if c.parent == nil {
		return c.name
	}

	return c.parent.path() + " " + c.name
}

func (c *Cmd) treeFlags() []*flag {
	if c.treeFlagsCache != nil {
		return c.treeFlagsCache
	}

	if c.parent == nil {
		return c.flagKwargs
	}

	cache := append(c.parent.treeFlags(), c.flagKwargs...)

	c.treeFlagsCache = cache

	return cache
}

func copy[T any](v T) T {
	var n T
	copier.Copy(&n, v)
	return n
}

func (c *Cmd) parentFlags() map[string]*flag {
	if c.parentFlagsCache != nil {
		return copy(c.parentFlagsCache)
	}

	if c.parent == nil {
		f := copy(c.flags)

		for k, v := range f {
			if v.hideGlobal {
				delete(f, k)
			}
		}

		c.parentFlagsCache = f

		return copy(f)
	}

	parentFlags := c.parent.parentFlags()
	for k, v := range c.flags {
		if !v.hideGlobal {
			parentFlags[k] = v
		}
	}

	return parentFlags
}

func (c *Cmd) parentShorthand() map[string]*flag {
	if c.parentShorthandCache != nil {
		return copy(c.parentShorthandCache)
	}

	if c.parent == nil {
		f := copy(c.shortHand)

		for k, v := range f {
			if v.hideGlobal {
				delete(f, k)
			}
		}

		c.parentShorthandCache = f

		return copy(f)
	}

	parentFlags := c.parent.parentShorthand()
	for k, v := range c.shortHand {
		if !v.hideGlobal {
			parentFlags[k] = v
		}
	}

	return parentFlags
}

func (c *Cmd) addFlag(shorthand string, name string, f *flag) {
	if name == "" && shorthand == "" {
		panic("flag must have a name or shorthand")
	}

	if name != "" {
		if c.parent != nil {
			if _, ok := c.parent.parentFlags()[name]; ok {
				panic(fmt.Sprintf("flag %s already exists in tree", name))
			}
		}

		if _, ok := c.flags[name]; ok {
			panic(fmt.Sprintf("flag %s already exists", name))
		} else {
			c.flags[name] = f
		}
	}

	if shorthand != "" {
		if len(shorthand) > 1 {
			panic("shorthand must be a single character")
		}

		if c.parent != nil {
			if _, ok := c.parent.parentShorthand()[shorthand]; ok {
				panic(fmt.Sprintf("shorthand %s already exists in tree", shorthand))
			}
		}

		if _, ok := c.shortHand[shorthand]; ok {
			panic(fmt.Sprintf("shorthand %s already exists", shorthand))
		} else {
			c.shortHand[shorthand] = f
		}
	}

	c.flagKwargs = append(c.flagKwargs, f)
}

func (c *Cmd) Flag(shorthand string, name string, options *Options[bool]) *bool {
	v := new(bool)

	var f *flag
	f = &flag{
		name:      name,
		shorthand: shorthand,
		value:     v,
		options:   options.toCaster(),
		owner:     c,
		action: func() {
			*v = true
		},
	}

	c.addFlag(shorthand, name, f)

	return v
}

func (c *Cmd) FlagCounter(shorthand string, name string, options *Options[int]) *int {
	v := new(int)

	f := &flag{
		name:      name,
		shorthand: shorthand,
		value:     v,
		options:   options.toCaster(),
		owner:     c,
		multi:     true,
		action: func() {
			*v++
		},
	}

	c.addFlag(shorthand, name, f)

	return v
}

func (c *Cmd) String(shorthand string, name string, options *Options[string]) *string {
	v := new(string)

	f := &flag{
		name:          name,
		shorthand:     shorthand,
		value:         v,
		options:       options.toCaster(),
		consumesValue: true,
		owner:         c,
	}

	c.addFlag(shorthand, name, f)

	return v
}

func (c *Cmd) Int(shorthand string, name string, options *Options[int]) *int {
	v := new(int)

	f := &flag{
		name:          name,
		shorthand:     shorthand,
		value:         v,
		options:       options.toCaster(),
		consumesValue: true,
		owner:         c,
	}

	c.addFlag(shorthand, name, f)

	return v
}

func (c *Cmd) Float(shorthand string, name string, options *Options[float64]) *float64 {
	v := new(float64)

	f := &flag{
		name:          name,
		shorthand:     shorthand,
		value:         v,
		options:       options.toCaster(),
		consumesValue: true,
		owner:         c,
	}

	c.addFlag(shorthand, name, f)

	return v
}

func (c *Cmd) StringList(shorthand string, name string, options *OptionsList[string]) *[]string {
	v := new([]string)

	f := &flag{
		name:          name,
		shorthand:     shorthand,
		value:         v,
		options:       options.toCaster(),
		consumesValue: true,
		multi:         true,
		owner:         c,
	}

	c.addFlag(shorthand, name, f)

	return v
}

func (c *Cmd) IntList(shorthand string, name string, options *OptionsList[int]) *[]int {
	v := new([]int)

	f := &flag{
		name:          name,
		shorthand:     shorthand,
		value:         v,
		options:       options.toCaster(),
		consumesValue: true,
		multi:         true,
		owner:         c,
	}

	c.addFlag(shorthand, name, f)

	return v
}

func (c *Cmd) FloatList(shorthand string, name string, options *OptionsList[float64]) *[]float64 {
	v := new([]float64)

	f := &flag{
		name:          name,
		shorthand:     shorthand,
		value:         v,
		options:       options.toCaster(),
		consumesValue: true,
		multi:         true,
		owner:         c,
	}

	c.addFlag(shorthand, name, f)

	return v
}

func (c *Cmd) IntPositional(name string, options *Options[int]) *int {
	v := new(int)

	f := &flag{
		name:       name,
		value:      v,
		options:    options.toCaster(),
		positional: true,
		owner:      c,
	}

	c.flagsArgs = append(c.flagsArgs, f)

	return v
}

func (c *Cmd) FloatPositional(name string, options *Options[float64]) *float64 {
	v := new(float64)

	f := &flag{
		name:       name,
		value:      v,
		options:    options.toCaster(),
		positional: true,
		owner:      c,
	}

	c.flagsArgs = append(c.flagsArgs, f)

	return v
}

func (c *Cmd) StringPositional(name string, options *Options[string]) *string {
	v := new(string)

	f := &flag{
		name:       name,
		value:      v,
		options:    options.toCaster(),
		positional: true,
		owner:      c,
	}

	c.flagsArgs = append(c.flagsArgs, f)

	return v
}

func (c *Cmd) parse(args []string) error {
	c.happened = true

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if arg == "" {
			// ignore empty arguments
			continue
		}

		if arg[0] == '-' {
			// this is a flag
			isLong := false
			argShift := arg[1:]
			if arg[1] == '-' {
				isLong = true
				argShift = argShift[1:]
			} else {
				// this is a short flag
				if len(arg) < 2 {
					return fmt.Errorf("invalid flag %s", arg)
				}
			}

			parts := strings.SplitN(argShift, "=", 2)
			key := parts[0]
			value := ""
			if len(parts) > 1 {
				value = parts[1]
			}

			var flag *flag
			if isLong {
				f, ok := c.flags[key]
				if !ok {
					if f, ok = c.parentFlags()[key]; !ok {
						return fmt.Errorf("unknown flag %s", arg)
					}
				}

				flag = f
			} else {
				f, ok := c.shortHand[key]
				if !ok {
					if f, ok = c.parentShorthand()[key]; !ok {
						return fmt.Errorf("unknown flag %s", arg)
					}
				}

				flag = f
			}

			if flag.used && !flag.multi {
				return fmt.Errorf("flag %s already been provided a value", arg)
			}

			if flag.consumesValue && value == "" {
				// this flag consumes a value, so we need to get the next argument
				i++
				if i >= len(args) {
					return fmt.Errorf("flag %s requires a value", arg)
				}

				value = args[i]
			} else if !flag.consumesValue && value != "" {
				return fmt.Errorf("flag %s does not consume a value", arg)
			}

			if flag.consumesValue {
				if err := flag.options.validate(value); err != nil {
					return err
				}

				if err := flag.options.cast(value, flag.value); err != nil {
					return err
				}
			} else if flag.action != nil {
				flag.action()
			}

			flag.used = true
		} else {
			// check if a subcommand is being called
			if cmd, ok := c.commands[arg]; ok {
				// this is a subcommand
				return cmd.parse(args[i+1:])
			}

			// this is a positional argument
			if len(c.flagsArgs) <= c.positionalIdx {
				return fmt.Errorf("unexpected positional argument %s", arg)
			}

			flag := c.flagsArgs[c.positionalIdx]
			c.positionalIdx++

			if err := flag.options.validate(arg); err != nil {
				return err
			}

			if err := flag.options.cast(arg, flag.value); err != nil {
				return err
			}
		}
	}

	if c.positionalIdx < len(c.flagsArgs) {
		return fmt.Errorf("missing positional argument [%s]", strings.ToUpper(c.flagsArgs[c.positionalIdx].name))
	}

	for _, f := range c.treeFlags() {
		if f.options.Required && !f.used {
			return fmt.Errorf("missing required flag: %s", f.String())
		}
		if f.options.Default != nil && !f.used {
			f.options.castValue(f.options.Default, f.value)
		}
	}

	return nil
}

func (c *Cmd) NewCommand(name string, help string) Command {
	cmd := newCmd(name, help, c)

	if _, ok := c.commands[name]; ok {
		panic(fmt.Sprintf("command %s already exists", name))
	} else {
		c.commands[name] = cmd
	}

	return cmd
}

func (c *Cmd) Happened() bool {
	return c.happened
}

func makePadding(n int) string {
	return strings.Repeat(" ", n)
}

func (c *Cmd) Usage(msg string) string {
	for _, child := range c.commands {
		if child.Happened() {
			return child.Usage(msg)
		}
	}

	var b strings.Builder

	if msg != "" {
		b.WriteString(msg)
		b.WriteString("\n\n")
	}
	b.WriteString("Usage: \n   ")
	b.WriteString(c.path())

	for _, f := range c.flagsArgs {
		b.WriteString(" [")
		b.WriteString(strings.ToUpper(f.name))
		b.WriteString("]")
	}

	if len(c.flagsArgs) > 0 {
		b.WriteString(" [flags]")
	} else if len(c.commands) > 0 {
		b.WriteString(" [command]")
	}

	b.WriteString("\n")
	b.WriteString(makePadding(len(c.path()) / 2))
	b.WriteString(c.help)
	b.WriteString("\n")

	if len(c.commands) > 0 {
		b.WriteString("\nAvailable Commands:\n")
		padding := ""
		for _, cmd := range c.commands {
			if len(cmd.name) > len(padding) {
				for i := len(padding); i < len(cmd.name); i++ {
					padding += " "
				}
			}
		}

		for _, cmd := range c.commands {
			b.WriteString("  ")
			b.WriteString(cmd.name)
			b.WriteString(padding[len(cmd.name):])
			b.WriteString("  ")
			b.WriteString(cmd.help)
			b.WriteString("\n")
		}
	}

	padding := ""

	flagsWriter := func(flags []*flag) {
		for _, f := range flags {
			offset := len(f.name + " " + f.options.strType())

			b.WriteString("   ")
			if f.shorthand != "" {
				b.WriteString("-")
				b.WriteString(f.shorthand)
				b.WriteString(", ")
			} else {
				b.WriteString("    ")
			}
			b.WriteString("--")
			b.WriteString(f.name)
			b.WriteString(" ")
			b.WriteString(f.options.strType())
			b.WriteString(padding[offset:])
			b.WriteString("  ")
			b.WriteString(f.options.Help)
			b.WriteString("\n")
		}
	}

	adjPadding := func(flags []*flag) {
		for _, f := range flags {
			offset := len(f.name + " " + f.options.strType())
			if offset > len(padding) {
				padding = makePadding(offset)
			}
		}
	}

	if len(c.flagKwargs) > 0 {
		adjPadding(c.flagKwargs)
	}

	var globalFlags []*flag
	if c.parent != nil {
		flags := c.parent.treeFlags()
		globalFlags = make([]*flag, 0, len(flags))
		for _, f := range flags {
			if !f.hideGlobal {
				globalFlags = append(globalFlags, f)
			}
		}

		if len(globalFlags) > 0 {
			adjPadding(globalFlags)
		}
	}

	if len(c.flagKwargs) > 0 {
		b.WriteString("\nFlags:\n")
		flagsWriter(c.flagKwargs)
	}

	if len(globalFlags) > 0 {
		b.WriteString("\nGlobal Flags:\n")
		flagsWriter(globalFlags)
	}

	return b.String()
}

func (c *Cmd) Parent() Command {
	return c.parent
}

func newCmd(name string, help string, parent *Cmd) *Cmd {
	cmd := &Cmd{
		name:      name,
		help:      help,
		flags:     map[string]*flag{},
		commands:  map[string]*Cmd{},
		shortHand: map[string]*flag{},
		parent:    parent,
	}

	cmd.addFlag("h", "help", &flag{
		name:          "help",
		shorthand:     "h",
		options:       &optionsCaster{Help: "Shows this help message"},
		consumesValue: false,
		hideGlobal:    true,
		action: func() {
			fmt.Fprintln(os.Stderr, cmd.Usage(""))
			os.Exit(1)
		},
	})

	return cmd
}
