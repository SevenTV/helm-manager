package utils

import (
	"crypto/sha256"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/jinzhu/copier"
	"github.com/manifoldco/promptui"
	"github.com/seventv/helm-manager/manager/cli"
	"github.com/seventv/helm-manager/manager/types"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
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

func ParseHelmRepos(data []byte) ([]types.Repo, error) {
	repos := []types.Repo{}
	err := json.Unmarshal(data, &repos)
	if err != nil {
		return repos, err
	}

	return repos, nil
}

func ReadChartValues(chart types.Chart) (yaml.Node, error) {
	values, err := os.ReadFile(chart.File)
	if err != nil {
		return yaml.Node{}, ErrorNotFound
	}

	return ParseYaml(values)
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

func Sum256(data []byte) []byte {
	h := sha256.New()
	h.Write(data)
	return h.Sum(nil)
}

func RemoveYamlComments(node yaml.Node) yaml.Node {
	if NodeIsZero(node) {
		return yaml.Node{}
	}

	var newNode yaml.Node
	copier.CopyWithOption(&newNode, &node, copier.Option{DeepCopy: true})

	var removeComments func(node *yaml.Node)
	removeComments = func(node *yaml.Node) {
		node.HeadComment = ""
		node.LineComment = ""
		node.FootComment = ""

		for _, child := range node.Content {
			removeComments(child)
		}
	}

	removeComments(&newNode)

	return newNode
}

func GetDefaultChartValues(chart types.Chart) yaml.Node {
	out, err := ExecuteCommand("helm", "show", "values", chart.Chart, "--version", chart.Version)
	if err != nil {
		zap.S().Errorf("Failed to get values for %s", chart.Name)
		return yaml.Node{}
	}

	chartValues, err := ParseYaml(out)
	if err != nil {
		zap.S().Errorf("Failed to parse values for %s", chart.Name)
		return yaml.Node{}
	}

	return ConvertDocument(chartValues)
}

func GetNonDefaultChartValues(chart types.Chart, values yaml.Node) yaml.Node {
	chartValues := GetDefaultChartValues(chart)
	if NodeIsZero(chartValues) {
		return yaml.Node{}
	}

	return PruneYaml(chartValues, values)
}

func UpdateRepos(cfg types.Config) {
	repoMap := CreateRepoMap(cfg)
	if len(repoMap) == 0 {
		return
	}

	updating := make(chan bool)
	finished := make(chan struct{})
	go func() {
		defer close(finished)
		defer close(updating)
		if cfg.Arguments.InTerminal {
			t := time.NewTicker(200 * time.Millisecond)
			defer t.Stop()
			i := 0
			stages := []string{"\\", "|", "/", "-"}
			for {
				select {
				case <-t.C:
					zap.S().Infof("%s [%s]\r", color.YellowString("Updating Helm Repos"), color.CyanString("%s", stages[i%len(stages)]))
					i++
				case success := <-updating:
					if success {
						zap.S().Infof("%s Updated Helm Repos", color.GreenString("✓"))
					} else {
						zap.S().Infof("%s Failed to update Helm Repos", color.RedString("✗"))
					}
					return
				}
			}
		} else {
			Info("Updating Helm Repos...")
			if <-updating {
				Info("Updated Helm Repos")
			} else {
				Error("Failed to update Helm Repos")
			}
		}
	}()

	data, err := ExecuteCommand("helm", "repo", "list", "-o", "json")
	if err != nil {
		updating <- false
		<-finished
		Fatal("Failed to list helm repos, is helm installed?\n", data)
	}

	repos, err := ParseHelmRepos(data)
	if err != nil {
		updating <- false
		<-finished

		Fatal("Failed to parse helm repo list")
	}

	installedReposMap := map[string]types.Repo{}
	for _, repo := range repos {
		installedReposMap[repo.Name] = repo
	}

	for _, repo := range repoMap {
		installedRepo, ok := installedReposMap[repo.Name]
		if ok && installedRepo.URL != repo.URL {
			out, err := ExecuteCommand("helm", "repo", "remove", repo.Name)
			if err != nil {
				updating <- false
				<-finished
				Fatal("Failed to remove repo %s\n%s", repo.Name, out)
			}
			ok = false
		}

		if !ok {
			out, err := ExecuteCommand("helm", "repo", "add", repo.Name, repo.URL)
			if err != nil {
				updating <- false
				<-finished
				Fatal("Failed to add helm repo %s\n%s", repo.Name, out)
			}
		}
	}

	_, err = ExecuteCommand("helm", "repo", "update")
	updating <- err == nil
	<-finished
	if err != nil {
		Fatal("Failed to update helm repos")
	}
}

func ToDocument(node yaml.Node) yaml.Node {
	var newNode yaml.Node
	copier.CopyWithOption(&newNode, &node, copier.Option{DeepCopy: true})

	newNode.Kind = yaml.DocumentNode
	newNode.Content = []*yaml.Node{&node}

	return newNode
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
