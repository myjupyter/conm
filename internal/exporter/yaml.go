package exporter

import (
	"bytes"
	"errors"
	"slices"

	"go.yaml.in/yaml/v3"
)

const (
	yamlIndent  = 2
	yamlNullTag = "!!null"
)

var errYAMLNotMapping = errors.New("the file is not a yaml mapping")

var yamlNullSpellings = []string{"", "~", "null", "Null", "NULL"}

func yamlMappingRoot(raw []byte) (doc, root *yaml.Node, err error) {
	doc = &yaml.Node{}
	if err := yaml.Unmarshal(raw, doc); err != nil {
		return nil, nil, err
	}

	if doc.Kind == 0 {
		doc = &yaml.Node{Kind: yaml.DocumentNode, Content: []*yaml.Node{{Kind: yaml.MappingNode}}}
	}

	root = doc.Content[0]
	if root.Kind != yaml.MappingNode {
		return nil, nil, errYAMLNotMapping
	}

	return doc, root, nil
}

func yamlChild(mapping *yaml.Node, key string, kind yaml.Kind) *yaml.Node {
	for i := 0; i+1 < len(mapping.Content); i += 2 {
		if mapping.Content[i].Value != key {
			continue
		}

		value := mapping.Content[i+1]
		if value.Tag == yamlNullTag {
			*value = yaml.Node{Kind: kind}
		}

		return value
	}

	value := &yaml.Node{Kind: kind}
	mapping.Content = append(mapping.Content, yamlScalar(key), value)

	return value
}

func yamlScalar(value string) *yaml.Node {
	node := &yaml.Node{Kind: yaml.ScalarNode, Value: value}
	if slices.Contains(yamlNullSpellings, value) {
		node.Style = yaml.DoubleQuotedStyle
	}

	return node
}

func yamlAppend(list *yaml.Node, item any) error {
	node := &yaml.Node{}
	if err := node.Encode(item); err != nil {
		return err
	}

	list.Content = append(list.Content, node)
	list.Style = 0

	return nil
}

func yamlEncoded(doc *yaml.Node) ([]byte, error) {
	var out bytes.Buffer

	encoder := yaml.NewEncoder(&out)
	encoder.SetIndent(yamlIndent)
	if err := encoder.Encode(doc); err != nil {
		return nil, err
	}

	return out.Bytes(), encoder.Close()
}
