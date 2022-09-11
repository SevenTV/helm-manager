package main

import (
	"crypto/sha256"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/jinzhu/copier"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

func NodeIsZero(n yaml.Node) bool {
	return n.Kind == 0 && n.Style == 0 && n.Tag == "" && n.Value == "" && n.Anchor == "" && n.Alias == nil && len(n.Content) == 0 &&
		n.HeadComment == "" && n.LineComment == "" && n.FootComment == "" && n.Line == 0 && n.Column == 0
}

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

func ParseHelmRepos(data []byte) ([]Repo, error) {
	repos := []Repo{}
	err := json.Unmarshal(data, &repos)
	if err != nil {
		return repos, err
	}

	return repos, nil
}

func ReadChartValues(chart Chart) (yaml.Node, error) {
	var chartValues yaml.Node
	values, err := os.ReadFile(chart.ValuesFile)
	if err != nil {
		return chartValues, ErrorNotFound
	}

	return ParseYaml(values)
}

func ParseYaml(data []byte) (yaml.Node, error) {
	var node yaml.Node
	err := yaml.Unmarshal(data, &node)
	if err != nil {
		return node, err
	}

	return node, nil
}

func OrStr(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

func ConvertDocument(node yaml.Node) yaml.Node {
	if node.Kind == yaml.DocumentNode {
		newNode := *node.Content[0]
		newNode.HeadComment = node.HeadComment
		newNode.LineComment = node.LineComment
		newNode.FootComment = node.FootComment
		return newNode
	}

	return node
}

func PruneYaml(defaultValues yaml.Node, chartValues yaml.Node) yaml.Node {
	var (
		prunedValues     yaml.Node
		newDefaultValues yaml.Node
	)

	copier.CopyWithOption(&newDefaultValues, &defaultValues, copier.Option{DeepCopy: true})
	copier.CopyWithOption(&prunedValues, &chartValues, copier.Option{DeepCopy: true})

	prunedValues = ConvertDocument(prunedValues)
	newDefaultValues = ConvertDocument(newDefaultValues)
	var pruneYaml func(*yaml.Node, *yaml.Node) *yaml.Node
	pruneYaml = func(newDefaultValues *yaml.Node, prunedValues *yaml.Node) *yaml.Node {
		if newDefaultValues.Kind != prunedValues.Kind {
			return prunedValues
		}

		switch newDefaultValues.Kind {
		case yaml.MappingNode:
			fastMap := make(map[string]**yaml.Node)
			for i := 0; i < len(prunedValues.Content); i += 2 {
				fastMap[prunedValues.Content[i].Value] = &prunedValues.Content[i+1]
			}

			for i := 0; i < len(newDefaultValues.Content); i += 2 {
				defaultKey := newDefaultValues.Content[i].Value
				if chartValue, ok := fastMap[defaultKey]; ok {
					defaultValue := newDefaultValues.Content[i+1]
					*chartValue = pruneYaml(defaultValue, *chartValue)
				}
			}

			newContent := []*yaml.Node{}
			for i := 0; i < len(prunedValues.Content); i += 2 {
				key := prunedValues.Content[i]
				value := prunedValues.Content[i+1]
				if value != nil {
					newContent = append(newContent, key, value)
				}
			}
			prunedValues.Content = newContent
			if len(prunedValues.Content) == 0 {
				return nil
			}
		case yaml.SequenceNode:
			if len(newDefaultValues.Content) == len(prunedValues.Content) {
				diff := false
				for i := 0; i < len(newDefaultValues.Content); i++ {
					defaultValue := newDefaultValues.Content[i]
					chartValue := prunedValues.Content[i]
					if pruneYaml(defaultValue, chartValue) != nil {
						diff = true
						break
					}
				}
				if !diff {
					return nil
				}
			}
		case yaml.ScalarNode:
			if newDefaultValues.Value == prunedValues.Value {
				return nil
			}
		}

		return prunedValues
	}

	ret := pruneYaml(&newDefaultValues, &prunedValues)
	if ret == nil {
		ret = &yaml.Node{}
	}

	if NodeIsZero(*ret) {
		*ret = yaml.Node{
			Kind:        newDefaultValues.Kind,
			Tag:         newDefaultValues.Tag,
			Content:     []*yaml.Node{},
			HeadComment: newDefaultValues.HeadComment,
			LineComment: newDefaultValues.LineComment,
			FootComment: newDefaultValues.FootComment,
			Style:       newDefaultValues.Style,
		}
	}

	return *ret
}

func MergeYaml(defaultValues yaml.Node, prunedChartValues yaml.Node) yaml.Node {
	var (
		newMergedValues      yaml.Node
		newPrunedChartValues yaml.Node
	)
	copier.Copy(&newMergedValues, &defaultValues)
	copier.Copy(&newPrunedChartValues, &prunedChartValues)

	newMergedValues = ConvertDocument(newMergedValues)
	newPrunedChartValues = ConvertDocument(newPrunedChartValues)

	var mergeYaml func(*yaml.Node, *yaml.Node) *yaml.Node
	mergeYaml = func(newMergedValues *yaml.Node, newPrunedChartValues *yaml.Node) *yaml.Node {
		if newMergedValues.Kind != newPrunedChartValues.Kind {
			if NodeIsZero(*newPrunedChartValues) {
				return newMergedValues
			}

			return newPrunedChartValues
		}

		switch newMergedValues.Kind {
		case yaml.MappingNode:
			fastMap := make(map[string]**yaml.Node)
			for i := 0; i < len(newMergedValues.Content); i += 2 {
				fastMap[newMergedValues.Content[i].Value] = &newMergedValues.Content[i+1]
			}

			for i := 0; i < len(newPrunedChartValues.Content); i += 2 {
				chartKey := newPrunedChartValues.Content[i].Value
				if defaultValue, ok := fastMap[chartKey]; ok {
					chartValue := newPrunedChartValues.Content[i+1]
					*defaultValue = mergeYaml(*defaultValue, chartValue)
				} else {
					newMergedValues.Content = append(newMergedValues.Content, newPrunedChartValues.Content[i], newPrunedChartValues.Content[i+1])
				}
			}
		case yaml.SequenceNode:
			return newPrunedChartValues
		case yaml.ScalarNode:
			if newMergedValues.Value != newPrunedChartValues.Value {
				return newPrunedChartValues
			}
		}

		return newMergedValues
	}

	return *mergeYaml(&newMergedValues, &newPrunedChartValues)
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

func GetDefaultChartValues(chart Chart) yaml.Node {
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

	return chartValues
}

func GetNonDefaultChartValues(chart Chart, values yaml.Node) yaml.Node {
	chartValues := GetDefaultChartValues(chart)
	if NodeIsZero(chartValues) {
		return yaml.Node{}
	}

	return PruneYaml(chartValues, values)
}

func UpdateRepos(cfg Config) {
	zap.S().Info("Updating repos")

	repoMap := CreateRepoMap(cfg)
	if len(repoMap) == 0 {
		zap.S().Warn("No repos to update")
		return
	}

	zap.S().Infof("%d repos to update", len(repoMap))

	data, err := ExecuteCommand("helm", "repo", "list", "-o", "json")
	if err != nil {
		zap.S().Fatal("Failed to list helm repos, is helm installed?")
	}

	repos, err := ParseHelmRepos(data)
	if err != nil {
		zap.S().Fatal("Failed to parse helm repo list")
	}

	installedReposMap := map[string]Repo{}
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
}
