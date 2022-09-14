package utils

import (
	"crypto/sha256"
	"io"
	"os/exec"
	"strings"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"github.com/seventv/helm-manager/manager/cli"
	"go.uber.org/zap"
)

func ExecuteCommand(name string, args ...string) ([]byte, error) {
	zap.S().Debugf("%s %s", name, strings.Join(args, " "))
	return exec.Command(name, args...).CombinedOutput()
}

func ExecuteCommandStdin(stdin io.Reader, name string, args ...string) ([]byte, error) {
	zap.S().Debugf("%s %s", name, strings.Join(args, " "))
	cmd := exec.Command(name, args...)

	cmd.Stdin = stdin

	return cmd.CombinedOutput()
}

func MergeStrings(lines ...string) string {
	prunedLines := []string{}
	for _, line := range lines {
		if line != "" {
			prunedLines = append(prunedLines, line)
		}
	}

	return strings.Join(prunedLines, "\n")
}

func OrStr(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

func Sum256(data []byte) []byte {
	h := sha256.New()
	h.Write(data)
	return h.Sum(nil)
}

func SelectCommand(prompt string, options []cli.Command) cli.Command {
	subcommandSelect := promptui.Select{
		Label:        prompt,
		HideSelected: true,
		HideHelp:     true,
		Templates: &promptui.SelectTemplates{
			Label:    "{{ . }}",
			Active:   "➔ {{ .Name | cyan }} {{ .Help | faint }}",
			Inactive: "  {{ .Name | cyan }} {{ .Help | faint }}",
			Selected: "{{ .Name }}",
		},
		Items: options,
	}

	idx, _, err := subcommandSelect.Run()
	if err != nil {
		zap.S().Fatal(err)
	}

	return options[idx]
}

func Fatal(msg string, args ...any) {
	zap.S().Fatalf("%s %s", color.New(color.Bold, color.FgBlack).Sprint("=>"), color.RedString(msg, args...))
}

func Error(msg string, args ...any) {
	zap.S().Errorf("%s %s", color.New(color.Bold, color.FgBlack).Sprint("=>"), color.RedString(msg, args...))
}

func Warn(msg string, args ...any) {
	zap.S().Warnf("%s %s", color.New(color.Bold, color.FgBlack).Sprint("->"), color.YellowString(msg, args...))
}

func Info(msg string, args ...any) {
	zap.S().Infof("%s %s", color.New(color.Bold, color.FgBlack).Sprint(">"), color.WhiteString(msg, args...))
}

func Debug(msg string, args ...any) {
	zap.S().Debugf("%s %s", color.New(color.Bold, color.FgBlack).Sprint(":>"), color.MagentaString(msg, args...))
}
