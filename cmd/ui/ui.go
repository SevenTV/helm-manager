package ui

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"github.com/seventv/helm-manager/cmd/args"
	"github.com/seventv/helm-manager/constants"
	"github.com/seventv/helm-manager/types"
	"github.com/seventv/helm-manager/utils"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var faintColor = color.New(color.Faint)

func UseInteractive() bool {
	return !args.Args.NonInteractive && constants.InTerm() && !constants.StdinUsed()
}

func CmdSelectable(cmd *cobra.Command) SelectableCommand {
	return cmdSelectable{
		Cmd: cmd,
	}
}

type SelectableCommand interface {
	types.Selectable
	Command() *cobra.Command
}

type cmdSelectable struct {
	Cmd *cobra.Command
}

func (p cmdSelectable) Command() *cobra.Command {
	return p.Cmd
}

func (c cmdSelectable) Label() string {
	return fmt.Sprintf("%s %s", color.CyanString(c.Cmd.Name()), faintColor.Sprint(c.Cmd.Short))
}

func (c cmdSelectable) Selected() string {
	return c.Cmd.Name()
}

func (c cmdSelectable) Details() string {
	return ""
}

func (c cmdSelectable) Match(input string) bool {
	return false
}

func PromptUiFunc[T comparable](label string) func(validator types.Validator[T]) (string, error) {
	return func(validator types.Validator[T]) (string, error) {
		return utils.Prompt(utils.PromptMessage[T]{
			Label:    label,
			Validate: validator,
		})
	}
}

func PromptUiConfirmFunc(label string, defaultValue bool) func(validator types.Validator[bool]) (bool, error) {
	return func(validator types.Validator[bool]) (bool, error) {
		d := "y/N"
		dv := "n"
		if defaultValue {
			d = "Y/n"
			dv = "y"
		}

		v, err := utils.Prompt(utils.PromptMessage[bool]{
			Label:       label,
			Validate:    validator,
			IsConfirm:   true,
			HideEntered: true,
			Default:     dv,
			Templates: &promptui.PromptTemplates{
				Confirm: fmt.Sprintf("{{ . }}? [%s]: ", d),
			},
		})

		if err == promptui.ErrAbort {
			v = "n"
		}

		if v == "" {
			v = strings.ToUpper(dv)
		}

		zap.S().Infof("%s? : %s", color.New(color.Faint).Sprint(label), v)

		if err == promptui.ErrAbort {
			return false, nil
		}

		return err == nil, err
	}
}

func PromptUiSelectorFunc[T comparable](label string, longLabel string, result func(int) error, options types.Future[[]types.Selectable]) func(validator types.Validator[T]) (string, error) {
	if longLabel == "" {
		longLabel = label
	} else if label == "" {
		label = longLabel
	}

	if result == nil {
		panic("result cannot be nil")
	}
	if options == nil {
		panic("options cannot be nil")
	}

	return func(validator types.Validator[T]) (string, error) {
		items, err := options.Get()
		if err != nil {
			return "", err
		}

		err = result(utils.Selector(label, longLabel, true, true, items))
		if err != nil {
			return "", err
		}

		return "", ErrIgnoreValue
	}
}

func PromptUiSelectorNewFunc[T comparable](label string, longLabel string, result func(int) error, options types.Future[[]types.Selectable]) func(validator types.Validator[T]) (string, error) {
	if longLabel == "" {
		longLabel = label
	} else if label == "" {
		label = longLabel
	}

	if result == nil {
		panic("result cannot be nil")
	}
	if options == nil {
		panic("options cannot be nil")
	}

	return func(validator types.Validator[T]) (string, error) {
		items, err := options.Get()
		if err != nil {
			return "", err
		}

		options := append([]types.Selectable{types.BasicSelectable{
			LabelStr:    color.GreenString("* New"),
			SelectedStr: "New",
		}}, items...)

		idx := utils.Selector("", longLabel, true, true, options)
		if idx == 0 {
			return PromptUiFunc[T](label)(validator)
		} else {
			zap.S().Infof("%s: %s", faintColor.Sprint(label), options[idx].Selected())
			err = result(idx - 1)
			if err != nil {
				return "", err
			}
		}

		return "", ErrIgnoreValue
	}
}
