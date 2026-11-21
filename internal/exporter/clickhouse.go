package exporter

import (
	"bytes"
	"encoding/xml"
	"errors"
	"html"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"go.yaml.in/yaml/v3"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/importer"
)

func clickhouseSections(conn config.Connection, password string) ([]importer.Section, []string) {
	ch, _ := conn.(config.ClickHouse)

	var connection params
	connection.add("hostname", ch.Hostname)
	connection.addPort("port", ch.PortNumber)
	connection.add("user", ch.User)
	connection.add("password", password)
	connection.add("database", ch.DBName)
	if ch.Secure {
		connection.add("secure", strconv.FormatBool(true))
	}

	return []importer.Section{{Name: ch.Metadata.Name, Params: connection}}, transportWarning(conn)
}

const (
	clickhouseRootTag        = "config"
	clickhouseCredentialsTag = "connections_credentials"
	clickhouseConnectionTag  = "connection"
	clickhouseNameKey        = "name"
	xmlIndent                = "    "
	xmlClosing               = "</"
)

var (
	errNoXMLElement   = errors.New("no element to add the connections to")
	errXMLNotMerged   = errors.New("the connections can't be added without breaking the file")
	errNotConnections = errors.New("connection is neither a map nor a list")
)

var clickhouseYAMLExtensions = []string{".yaml", ".yml"}

type clickhouseXMLFile struct {
	Connections []struct{} `xml:"connections_credentials>connection"`
}

func clickhouseMerged(path string, raw []byte, sections []importer.Section) ([]byte, error) {
	isXML := !slices.Contains(clickhouseYAMLExtensions, filepath.Ext(path))
	if content := bytes.TrimSpace(raw); len(content) > 0 {
		isXML = content[0] == '<'
	}

	if !isXML {
		return clickhouseYAMLMerged(raw, sections)
	}

	merged, err := clickhouseXMLInserted(string(raw), sections)
	if err != nil {
		return nil, err
	}

	var before, after clickhouseXMLFile
	if len(bytes.TrimSpace(raw)) > 0 {
		if err := xml.Unmarshal(raw, &before); err != nil {
			return nil, err
		}
	}
	if err := xml.Unmarshal([]byte(merged), &after); err != nil || len(after.Connections) != len(before.Connections)+len(sections) {
		return nil, errXMLNotMerged
	}

	return []byte(merged), nil
}

func clickhouseXMLInserted(text string, sections []importer.Section) (string, error) {
	if strings.TrimSpace(text) == "" {
		credentials := xmlBlock(xmlIndent, clickhouseCredentialsTag, clickhouseXMLText(xmlIndent+xmlIndent, sections))

		return xmlBlock("", clickhouseRootTag, credentials), nil
	}

	inCredentials := true
	at := strings.LastIndex(text, xmlClosing+clickhouseCredentialsTag+">")
	if at < 0 {
		inCredentials, at = false, strings.LastIndex(text, xmlClosing)
	}
	if at < 0 {
		return "", errNoXMLElement
	}

	lineStart := strings.LastIndex(text[:at], "\n") + 1
	indent := text[lineStart:at]
	if strings.TrimSpace(indent) != "" {
		lineStart, indent = at, ""
	}

	added := clickhouseXMLText(indent+xmlIndent, sections)
	if !inCredentials {
		added = xmlBlock(indent+xmlIndent, clickhouseCredentialsTag, clickhouseXMLText(indent+xmlIndent+xmlIndent, sections))
	}

	return text[:lineStart] + added + text[lineStart:], nil
}

func clickhouseXMLText(indent string, sections []importer.Section) string {
	var text strings.Builder
	for _, section := range sections {
		var fields strings.Builder
		fields.WriteString(xmlField(indent+xmlIndent, clickhouseNameKey, section.Name))
		for _, param := range section.Params {
			fields.WriteString(xmlField(indent+xmlIndent, param.Key, param.Value))
		}
		text.WriteString(xmlBlock(indent, clickhouseConnectionTag, fields.String()))
	}

	return text.String()
}

func xmlBlock(indent, tag, body string) string {
	return indent + "<" + tag + ">\n" + body + indent + xmlClosing + tag + ">\n"
}

func xmlField(indent, tag, value string) string {
	return indent + "<" + tag + ">" + html.EscapeString(value) + xmlClosing + tag + ">\n"
}

func clickhouseYAMLMerged(raw []byte, sections []importer.Section) ([]byte, error) {
	doc, root, err := yamlMappingRoot(raw)
	if err != nil {
		return nil, err
	}

	credentials := yamlChild(root, clickhouseCredentialsTag, yaml.MappingNode)
	if credentials.Kind != yaml.MappingNode {
		return nil, errNotConnections
	}

	list := yamlChild(credentials, clickhouseConnectionTag, yaml.SequenceNode)
	if list.Kind == yaml.MappingNode {
		single := *list
		*list = yaml.Node{Kind: yaml.SequenceNode, Content: []*yaml.Node{&single}}
	}
	if list.Kind != yaml.SequenceNode {
		return nil, errNotConnections
	}

	for _, section := range sections {
		connection := &yaml.Node{Kind: yaml.MappingNode, Content: []*yaml.Node{yamlScalar(clickhouseNameKey), yamlScalar(section.Name)}}
		for _, param := range section.Params {
			connection.Content = append(connection.Content, yamlScalar(param.Key), yamlScalar(param.Value))
		}
		list.Content = append(list.Content, connection)
	}
	list.Style = 0

	return yamlEncoded(doc)
}
