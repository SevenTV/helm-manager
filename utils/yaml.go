package utils

import (
	"bytes"
	"fmt"
	"io"

	"github.com/jinzhu/copier"
	"github.com/seventv/helm-manager/v2/logger"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

func DefaultNode(node *yaml.Node) *yaml.Node {
	return &yaml.Node{
		Kind:        node.Kind,
		Tag:         node.Tag,
		Style:       node.Style,
		HeadComment: node.HeadComment,
		LineComment: node.LineComment,
		FootComment: node.FootComment,
		Content:     []*yaml.Node{},
	}
}

func NodeIsZero(n *yaml.Node) bool {
	if n == nil {
		return true
	}

	return n.Kind == 0 && n.Style == 0 && n.Tag == "" && n.Value == "" && n.Anchor == "" && n.Alias == nil && len(n.Content) == 0 &&
		n.HeadComment == "" && n.LineComment == "" && n.FootComment == "" && n.Line == 0 && n.Column == 0
}

func ParseYaml(data []byte) (*yaml.Node, error) {
	parent := yaml.Node{
		Kind: yaml.DocumentNode,
	}
	node := &yaml.Node{}

	dec := yaml.NewDecoder(bytes.NewReader(data))

	for {
		err := dec.Decode(node)
		if err != nil && err != io.EOF {
			return nil, err
		}

		if !NodeIsZero(node) {
			child := ConvertDocument(node)
			child.Style = yaml.DoubleQuotedStyle
			parent.Content = append(parent.Content, child)
			node = &yaml.Node{}
		}

		if err == io.EOF {
			break
		}
	}

	if len(parent.Content) == 0 && len(data) != 0 {
		parent.HeadComment = string(data)
	}

	return &parent, nil
}

func MarshalYaml(node *yaml.Node) ([]byte, error) {
	if node.Kind != yaml.DocumentNode {
		return nil, &yaml.TypeError{Errors: []string{"node is not a document"}}
	}

	newNode := &yaml.Node{}
	copier.CopyWithOption(newNode, node, copier.Option{DeepCopy: true})

	TraverseYamlNode(newNode, func(node *yaml.Node) {
		node.Style = 0
	})

	var buf bytes.Buffer

	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)

	for {
		if len(newNode.Content) == 0 {
			break
		}

		err := enc.Encode(newNode.Content[0])
		if err != nil {
			return nil, err
		}

		newNode.Content = newNode.Content[1:]
	}

	if err := enc.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func PruneYaml(defaultValues *yaml.Node, chartValues *yaml.Node) *yaml.Node {
	var (
		newDefaultValues = &yaml.Node{}
		newChartValues   = &yaml.Node{}
	)

	copier.CopyWithOption(newDefaultValues, defaultValues, copier.Option{DeepCopy: true})
	copier.CopyWithOption(newChartValues, chartValues, copier.Option{DeepCopy: true})

	if newDefaultValues.Kind == yaml.DocumentNode {
		zap.S().Fatal("Default values should not be a document")
	}
	if newChartValues.Kind == yaml.DocumentNode {
		zap.S().Fatal("Chart values should not be a document")
	}

	var pruneYaml func(*yaml.Node, *yaml.Node) *yaml.Node
	pruneYaml = func(newDefaultValues *yaml.Node, newChartValues *yaml.Node) *yaml.Node {
		if newDefaultValues.Kind != newChartValues.Kind {
			return newChartValues
		}

		switch newDefaultValues.Kind {
		case yaml.MappingNode:
			fastMap := make(map[string]**yaml.Node)
			for i := 0; i < len(newChartValues.Content); i += 2 {
				fastMap[newChartValues.Content[i].Value] = &newChartValues.Content[i+1]
			}

			for i := 0; i < len(newDefaultValues.Content); i += 2 {
				defaultKey := newDefaultValues.Content[i].Value
				if chartValue, ok := fastMap[defaultKey]; ok {
					defaultValue := newDefaultValues.Content[i+1]
					*chartValue = pruneYaml(defaultValue, *chartValue)
				}
			}

			newContent := []*yaml.Node{}
			for i := 0; i < len(newChartValues.Content); i += 2 {
				key := newChartValues.Content[i]
				value := newChartValues.Content[i+1]
				if value != nil {
					newContent = append(newContent, key, value)
				}
			}
			newChartValues.Content = newContent
			if len(newChartValues.Content) == 0 {
				return nil
			}
		case yaml.SequenceNode:
			if len(newDefaultValues.Content) == len(newChartValues.Content) {
				diff := false
				for i := 0; i < len(newDefaultValues.Content); i++ {
					defaultValue := newDefaultValues.Content[i]
					chartValue := newChartValues.Content[i]
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
			if newDefaultValues.Value == newChartValues.Value {
				return nil
			}
		}

		return newChartValues
	}

	ret := pruneYaml(newDefaultValues, newChartValues)
	if ret == nil {
		ret = &yaml.Node{}
	}

	if NodeIsZero(ret) {
		ret = DefaultNode(newDefaultValues)
	}

	return ret
}

func MergeYaml(defaultValues *yaml.Node, chartValues *yaml.Node) *yaml.Node {
	var (
		newDefaultValues = &yaml.Node{}
		newChartValues   = &yaml.Node{}
	)
	copier.CopyWithOption(newDefaultValues, defaultValues, copier.Option{DeepCopy: true})
	copier.CopyWithOption(newChartValues, chartValues, copier.Option{DeepCopy: true})

	if newDefaultValues.Kind == yaml.DocumentNode {
		zap.S().Fatal("Default values should not be a document")
	}
	if newChartValues.Kind == yaml.DocumentNode {
		zap.S().Fatal("Chart values should not be a document")
	}

	var mergeYaml func(*yaml.Node, *yaml.Node) *yaml.Node
	mergeYaml = func(newDefaultValues *yaml.Node, newChartValues *yaml.Node) *yaml.Node {
		if newDefaultValues.Kind != newChartValues.Kind {
			if NodeIsZero(newChartValues) {
				return newDefaultValues
			}

			return newChartValues
		}

		switch newDefaultValues.Kind {
		case yaml.MappingNode:
			fastMap := make(map[string]**yaml.Node)
			for i := 0; i < len(newDefaultValues.Content); i += 2 {
				fastMap[newDefaultValues.Content[i].Value] = &newDefaultValues.Content[i+1]
			}

			for i := 0; i < len(newChartValues.Content); i += 2 {
				chartKey := newChartValues.Content[i].Value
				if defaultValue, ok := fastMap[chartKey]; ok {
					chartValue := newChartValues.Content[i+1]
					*defaultValue = mergeYaml(*defaultValue, chartValue)
				} else {
					newDefaultValues.Content = append(newDefaultValues.Content, newChartValues.Content[i], newChartValues.Content[i+1])
				}
			}
		case yaml.SequenceNode:
			return newChartValues
		case yaml.ScalarNode:
			if newDefaultValues.Value != newChartValues.Value {
				return newChartValues
			}
		}

		newDefaultValues.HeadComment = OrStr(newChartValues.HeadComment, newDefaultValues.HeadComment)
		newDefaultValues.LineComment = OrStr(newChartValues.LineComment, newDefaultValues.LineComment)
		newDefaultValues.FootComment = OrStr(newChartValues.FootComment, newDefaultValues.FootComment)

		return newDefaultValues
	}

	return mergeYaml(newDefaultValues, newChartValues)
}

func IsDifferent(first *yaml.Node, second *yaml.Node) bool {
	var isDifferent func(string, *yaml.Node, *yaml.Node) bool
	isDifferent = func(path string, first *yaml.Node, second *yaml.Node) bool {
		if first.Kind != second.Kind {
			logger.Debugf("Different kind (%s): %s != %s", path, first.Kind, second.Kind)
			return true
		}

		switch first.Kind {
		case yaml.MappingNode:
			fastMap := make(map[string]**yaml.Node)
			for i := 0; i < len(first.Content); i += 2 {
				fastMap[first.Content[i].Value] = &first.Content[i+1]
			}

			for i := 0; i < len(second.Content); i += 2 {
				secondKey := second.Content[i].Value
				secondValue := second.Content[i+1]
				if firstValue, ok := fastMap[secondKey]; ok {
					if isDifferent(fmt.Sprintf("%s.%s", path, secondKey), *firstValue, secondValue) {
						return true
					}
				} else {
					if secondValue.Kind == yaml.ScalarNode && secondValue.Tag == "!!null" {
						continue
					}

					return true
				}
			}
		case yaml.SequenceNode:
			if len(first.Content) != len(second.Content) {
				return true
			}

			for i := 0; i < len(first.Content); i++ {
				if isDifferent(fmt.Sprintf("%s.%d", path, i), first.Content[i], second.Content[i]) {
					return true
				}
			}
		case yaml.ScalarNode:
			if first.Value != second.Value {
				logger.Debugf("Different Value (%s): %s != %s", path, first.Kind, second.Kind)
				return true
			}
		}

		return false
	}

	return isDifferent("", first, second)
}

func ConvertDocument(node *yaml.Node) *yaml.Node {
	copiedNode := &yaml.Node{}

	copier.CopyWithOption(copiedNode, node, copier.Option{DeepCopy: true})

	if copiedNode.Kind == yaml.DocumentNode {
		var newNode *yaml.Node
		if len(copiedNode.Content) > 0 {
			newNode = copiedNode.Content[0]
		} else {
			newNode = &yaml.Node{
				Kind: yaml.MappingNode,
				Tag:  "!!map",
			}
		}

		newNode.HeadComment = MergeStrings(copiedNode.HeadComment, newNode.HeadComment)
		newNode.LineComment = MergeStrings(copiedNode.LineComment, newNode.LineComment)
		newNode.FootComment = MergeStrings(copiedNode.FootComment, newNode.FootComment)

		return newNode
	}

	return copiedNode
}

func RemoveYamlComments(node *yaml.Node) *yaml.Node {
	if NodeIsZero(node) {
		return nil
	}

	newNode := &yaml.Node{}
	copier.CopyWithOption(newNode, node, copier.Option{DeepCopy: true})

	TraverseYamlNode(newNode, func(node *yaml.Node) {
		node.HeadComment = ""
		node.LineComment = ""
		node.FootComment = ""
	})

	return newNode
}

func TraverseYamlNode(node *yaml.Node, f func(node *yaml.Node)) {
	f(node)

	for _, child := range node.Content {
		TraverseYamlNode(child, f)
	}
}

func ToDocument(node *yaml.Node) *yaml.Node {
	newNode := &yaml.Node{}
	copier.CopyWithOption(newNode, node, copier.Option{DeepCopy: true})

	newDefault := DefaultNode(node)

	newDefault.Kind = yaml.DocumentNode
	newDefault.Content = []*yaml.Node{newNode}

	return newDefault
}
