package exporter

import (
	"bytes"
	"encoding/xml"
	"errors"
	"html"
	"io"
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

type clickhouseXMLSpotKind int

const (
	xmlUnderRoot clickhouseXMLSpotKind = iota
	xmlInCredentials
	xmlEmptyCredentials
)

type clickhouseXMLSpot struct {
	kind    clickhouseXMLSpotKind
	at, end int
}

func clickhouseXMLInserted(text string, sections []importer.Section) (string, error) {
	if strings.TrimSpace(text) == "" {
		credentials := xmlBlock(xmlIndent, clickhouseCredentialsTag, clickhouseXMLText(xmlIndent+xmlIndent, sections))

		return xmlBlock("", clickhouseRootTag, credentials), nil
	}

	spot, err := clickhouseXMLSpotOf(text)
	if err != nil {
		return "", err
	}

	lineStart := strings.LastIndex(text[:spot.at], "\n") + 1
	indent := text[lineStart:spot.at]
	if strings.TrimSpace(indent) != "" {
		lineStart, indent = spot.at, ""
	}

	switch spot.kind {
	case xmlInCredentials:
		return text[:lineStart] + clickhouseXMLText(indent+xmlIndent, sections) + text[lineStart:], nil
	case xmlEmptyCredentials:
		credentials := xmlBlock(indent, clickhouseCredentialsTag, clickhouseXMLText(indent+xmlIndent, sections))

		return text[:lineStart] + strings.TrimSuffix(credentials, "\n") + text[spot.end:], nil
	default:
		credentials := xmlBlock(indent+xmlIndent, clickhouseCredentialsTag, clickhouseXMLText(indent+xmlIndent+xmlIndent, sections))

		return text[:lineStart] + credentials + text[lineStart:], nil
	}
}

func clickhouseXMLSpotOf(text string) (clickhouseXMLSpot, error) {
	var (
		decoder                = xml.NewDecoder(strings.NewReader(text))
		credentials            *clickhouseXMLSpot
		rootClosing            = -1
		depth, opened, emptyAt int
	)
	for {
		at := int(decoder.InputOffset())
		token, err := decoder.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return clickhouseXMLSpot{}, err
		}

		switch token := token.(type) {
		case xml.StartElement:
			depth++
			if depth == 2 && token.Name.Local == clickhouseCredentialsTag {
				opened, emptyAt = at, int(decoder.InputOffset())
			}
		case xml.EndElement:
			if depth == 2 && token.Name.Local == clickhouseCredentialsTag {
				credentials = &clickhouseXMLSpot{kind: xmlInCredentials, at: at}
				if int(decoder.InputOffset()) == emptyAt {
					credentials = &clickhouseXMLSpot{kind: xmlEmptyCredentials, at: opened, end: emptyAt}
				}
			}
			if depth == 1 {
				rootClosing = at
			}
			depth--
		}
	}

	switch {
	case credentials != nil:
		return *credentials, nil
	case rootClosing >= 0:
		return clickhouseXMLSpot{kind: xmlUnderRoot, at: rootClosing}, nil
	default:
		return clickhouseXMLSpot{}, errNoXMLElement
	}
}

func clickhouseXMLText(indent string, sections []importer.Section) string {
	var text strings.Builder
	for _, section := range sections {
		var fields strings.Builder
		fields.WriteString(xmlField(indent+xmlIndent, importer.ClickHouseNameKey, section.Name))
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
		connection := &yaml.Node{Kind: yaml.MappingNode, Content: []*yaml.Node{yamlScalar(importer.ClickHouseNameKey), yamlScalar(section.Name)}}
		for _, param := range section.Params {
			connection.Content = append(connection.Content, yamlScalar(param.Key), yamlScalar(param.Value))
		}
		list.Content = append(list.Content, connection)
	}
	list.Style = 0

	return yamlEncoded(doc)
}
