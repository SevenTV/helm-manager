package manager

import (
	"crypto/sha256"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/jinzhu/copier"
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
	zap.S().Debug("Updating repos")

	repoMap := CreateRepoMap(cfg)
	if len(repoMap) == 0 {
		zap.S().Warn("No repos to update")
		return
	}

	data, err := ExecuteCommand("helm", "repo", "list", "-o", "json")
	if err != nil {
		zap.S().Fatal("Failed to list helm repos, is helm installed?")
	}

	repos, err := ParseHelmRepos(data)
	if err != nil {
		zap.S().Fatal("Failed to parse helm repo list")
	}

	installedReposMap := map[string]types.Repo{}
	for _, repo := range repos {
		installedReposMap[repo.Name] = repo
	}

	for _, repo := range repoMap {
		installedRepo, ok := installedReposMap[repo.Name]
		if ok && installedRepo.URL != repo.URL {
			_, err = ExecuteCommand("helm", "repo", "remove", repo.Name)
			if err != nil {
				zap.S().Fatalf("Failed to remove repo %s", repo.Name)
			}
			ok = false
		}

		if !ok {
			_, err = ExecuteCommand("helm", "repo", "add", repo.Name, repo.URL)
			if err != nil {
				zap.S().Fatal("Failed to add helm repo %s %s", repo.Name, repo.URL)
			}
		}
	}

	_, err = ExecuteCommand("helm", "repo", "update")
	if err != nil {
		zap.S().Fatal("Failed to update helm repos")
	}

	zap.S().Debug("Updated repos")
}

func ToDocument(node yaml.Node) yaml.Node {
	var newNode yaml.Node
	copier.CopyWithOption(&newNode, &node, copier.Option{DeepCopy: true})

	newNode.Kind = yaml.DocumentNode
	newNode.Content = []*yaml.Node{&node}

	return newNode
}
