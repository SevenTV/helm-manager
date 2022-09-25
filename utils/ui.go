package utils

import (
	"fmt"
	"sort"

	"github.com/manifoldco/promptui"
	"github.com/seventv/helm-manager/logger"
	"github.com/seventv/helm-manager/types"
)

func Selector(short string, labelLong string, help bool, search bool, options []types.Selectable) int {
	type selection struct {
		Label    string
		Selected string
		Details  string
		idx      int

		Match func(input string) bool
	}

	items := make([]selection, 0, len(options))

	for idx, s := range options {
		items = append(items, selection{
			Label:    s.Label(),
			Selected: s.Selected(),
			Details:  s.Details(),
			Match:    s.Match,
			idx:      idx,
		})
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].Label < items[j].Label
	})

	selected := fmt.Sprintf(`{{ "%s:" | faint }} {{ .Selected }}`, short)
	if short == "" {
		selected = ""
	}

	prompt := promptui.Select{
		Label:        labelLong,
		Items:        items,
		HideSelected: short == "",
		HideHelp:     !help,
		Templates: &promptui.SelectTemplates{
			Label:    "{{ .Label }}",
			Active:   "➔ {{ .Label }}",
			Inactive: "  {{ .Label }}",
			Selected: selected,
			Details:  "{{ .Details }}",
		},
	}

	if search {
		prompt.Searcher = func(input string, index int) bool {
			return items[index].Match(input)
		}
	}

	i, _, err := prompt.Run()
	if err != nil {
		logger.Fatal(err)
	}

	return items[i].idx
}

type PromptMessage[T comparable] struct {
	Label       string
	IsConfirm   bool
	HideEntered bool
	Validate    types.Validator[T]
	Default     string
	Templates   *promptui.PromptTemplates
}

func Prompt[T comparable](p PromptMessage[T]) (string, error) {
	valid := func(s string) error {
		t, err := p.Validate.Convert(s)
		if err != nil {
			return err
		}

		return p.Validate.Validate(t)
	}

	prompt := promptui.Prompt{
		Label:       p.Label,
		IsConfirm:   p.IsConfirm,
		Templates:   p.Templates,
		Default:     p.Default,
		HideEntered: p.HideEntered,
	}

	if !p.IsConfirm {
		prompt.Validate = valid
	}

	result, err := prompt.Run()
	if err != nil && err != promptui.ErrAbort {
		logger.Fatal(err)
	} else if err == promptui.ErrAbort {
		return "", err
	}

	return result, err
}
