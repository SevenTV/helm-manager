package ui

import (
	"errors"
	"os"
	"strings"

	"github.com/fatih/color"
	cmdErrors "github.com/seventv/helm-manager/v2/cmd/errors"
	"github.com/seventv/helm-manager/v2/logger"
	"github.com/seventv/helm-manager/v2/types"
	"github.com/seventv/helm-manager/v2/utils"
	"github.com/spf13/cobra"
)

var ErrIgnoreValue = errors.New("ignore value")

func SubCommandRequired(next ...cobra.PositionalArgs) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if !UseInteractive() {
			return cmdErrors.ErrSubcommandRequired
		}

		for _, f := range next {
			if err := f(cmd, args); err != nil {
				return err
			}
		}

		return nil
	}
}

func RunSubCommand(cmd *cobra.Command, cmds []SelectableCommand) {
	items := make([]types.Selectable, len(cmds))
	for i, c := range cmds {
		items[i] = c
	}

	idx := utils.Selector("", "Select a command", false, false, items)

	c := cmds[idx].Command()

	os.Args = append(os.Args, c.Use)
	cmd.Root().SetArgs(nil)

	if err := cmd.Root().Execute(); err != nil {
		os.Exit(1)
	}
}

type UiFunc[T comparable] func(types.Validator[T]) (T, error)

type Arg[T comparable] struct {
	Name       string
	Ptr        *T
	Disabled   types.Future[bool]
	Positional bool
	Validator  types.Validator[T]
	UI         UiFunc[T]
	Callback   func(T) error
}

func (p Arg[T]) ToRequiredArg() rArgInterface {
	validator := p.Validator
	if validator == nil {
		// this is so we get the type converter from the validator
		validator = types.ValidatorFunction[T](func(v T) error {
			return nil
		})
	}

	return rArg[T]{
		name:       p.Name,
		ptr:        p.Ptr,
		ui:         p.UI,
		positional: p.Positional,
		valid:      validator,
		disabled:   p.Disabled,
		callback:   p.Callback,
	}
}

type rArg[T comparable] struct {
	name string
	ptr  *T

	positional bool
	valid      types.Validator[T]
	ui         UiFunc[T]
	disabled   types.Future[bool]
	callback   func(T) error
}

func (r rArg[T]) Disabled() bool {
	if r.disabled == nil {
		return false
	}

	v, _ := r.disabled.Get()
	return v
}

func (r rArg[T]) Empty() bool {
	var zero T
	return r.ptr == nil || *r.ptr == zero
}

func (r rArg[T]) Valid() error {
	return r.valid.Validate(*r.ptr)
}

func (r rArg[T]) Set(arg string) error {
	t, err := r.valid.Convert(arg)
	if err != nil {
		return err
	}

	old := *r.ptr
	*r.ptr = t

	err = r.Valid()
	if err != nil {
		*r.ptr = old
	}

	return err
}

func (r rArg[T]) Positional() bool {
	return r.positional
}

func (r rArg[T]) Name() string {
	return r.name
}

func (r rArg[T]) HasUI() bool {
	return r.ui != nil
}

func (r rArg[T]) UI() error {
	res, err := r.ui(r.valid)
	if err != nil && err != ErrIgnoreValue {
		return err
	} else if err == nil {
		*r.ptr = res
	}

	return nil
}

func (r rArg[T]) ToRequiredArg() rArgInterface {
	return r
}

func (r rArg[T]) Callback() error {
	if r.callback == nil {
		return nil
	}

	return r.callback(*r.ptr)
}

type RequiredArg interface {
	ToRequiredArg() rArgInterface
}

type rArgInterface interface {
	ToRequiredArg() rArgInterface

	Name() string
	Empty() bool
	Positional() bool
	Set(string) error
	Valid() error
	Callback() error
	HasUI() bool
	UI() error
	Disabled() bool
}

var colorReset = color.New(color.Reset)

func fatalErr(err error, cmd *cobra.Command) {
	errMsg := err.Error()
	if strings.Contains(errMsg, "$USAGE") {
		logger.Fatalf(strings.Replace(errMsg, "$USAGE", "%s", -1), colorReset.Sprint(cmd.UsageString()))
	}

	logger.Fatal(err)
}

func PositionalArgs(aargs []RequiredArg, preHook func(cmd *cobra.Command)) cobra.PositionalArgs {
	rargs := make([]rArgInterface, len(aargs))
	for i, a := range aargs {
		rargs[i] = a.ToRequiredArg()
	}

	return func(cmd *cobra.Command, args []string) error {
		if preHook != nil {
			preHook(cmd)
		}

		for _, parg := range rargs {
			if len(args) == 0 {
				break
			}

			if !parg.Empty() || !parg.Positional() || parg.Disabled() {
				continue
			}

			arg := args[0]
			args = args[1:]

			if err := parg.Set(arg); err != nil {
				fatalErr(err, cmd)
			}
		}

		if len(args) > 0 {
			return cmdErrors.ErrUnexpectedArgs
		}

		for _, parg := range rargs {
			if parg.Disabled() {
				continue
			}

			if parg.Empty() && UseInteractive() && parg.HasUI() {
				err := parg.UI()
				if err != nil {
					fatalErr(err, cmd)
				}
			}

			if err := parg.Valid(); err != nil {
				fatalErr(err, cmd)
			}

			if err := parg.Callback(); err != nil {
				fatalErr(err, cmd)
			}
		}

		return nil
	}
}
