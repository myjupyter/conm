package spec

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/secret"
)

const (
	sshExampleHost         = "bastion.internal.corp"
	sshExamplePort         = "22"
	sshExampleUsername     = "deploy"
	sshExampleJump         = "deploy@bastion.internal.corp"
	sshExampleKeepAlive    = "seconds between probes"
	sshExampleLocalForward = "5432:localhost:5432"
	sshExampleName         = "bastion-eu"
	sshExampleDescription  = "Entry point for every eu-west box"
	sshExampleTags         = "prod,bastion"
)

const sshDefaultPort = 22

const (
	sshAdvancedSectionTitle = "advanced"
	sshAdvancedSectionNote  = "all optional"
)

const (
	SSHForwardAgentNo  FormFieldValue = "no"
	SSHForwardAgentYes FormFieldValue = "yes"
)

type sshFormField = string

const (
	sshFormFieldName         sshFormField = "Name"
	sshFormFieldDescription  sshFormField = "Description"
	sshFormFieldTags         sshFormField = "Tags"
	sshFormFieldHost         sshFormField = "Host"
	sshFormFieldPort         sshFormField = "Port"
	sshFormFieldUsername     sshFormField = "Username"
	sshFormFieldAuth         sshFormField = "Auth"
	sshFormFieldJump         sshFormField = "Proxy jump"
	sshFormFieldForwardAgent sshFormField = "Forward agent"
	sshFormFieldKeepAlive    sshFormField = "Keepalive"
	sshFormFieldLocalForward sshFormField = "Local forward"
)

var SSHAuthOrder = []config.SSHAuth{
	config.SSHAuthAgent,
	config.SSHAuthKey,
	config.SSHAuthPassword,
}

var SSHSecretProvidersOrder = []string{
	secret.None,
	secret.Literal,
	secret.Keyring,
	secret.Filepath,
}

var sshSecretProviderField = FormField{
	Key:          SecretProviderKey,
	Label:        secretProviderField.Label,
	Kind:         SelectFieldKind,
	DefaultValue: secret.None,
	Options:      SSHSecretProvidersOrder,
	OptionsFunc:  sshSecretProviders,
}

func sshSecretProviders(values map[FormFieldKey]FormFieldValue) []string {
	switch values[strings.ToLower(sshFormFieldAuth)] {
	case config.SSHAuthKey:
		return []string{secret.Filepath, secret.Literal, secret.Keyring}
	case config.SSHAuthPassword:
		return []string{secret.Literal, secret.Keyring, secret.None}
	default:
		return []string{secret.None}
	}
}

var SSHFormFields = []FormField{
	{
		Key:          strings.ToLower(sshFormFieldHost),
		Label:        sshFormFieldHost,
		Example:      sshExampleHost,
		ValidateFunc: config.ValidateSSHHost,
	},
	{
		Key:          strings.ToLower(sshFormFieldPort),
		Label:        sshFormFieldPort,
		Example:      sshExamplePort,
		Kind:         IntFieldKind,
		Property:     OptionalFieldProperty,
		DefaultValue: strconv.Itoa(sshDefaultPort),
		ValidateFunc: func(value string) error {
			if value == "" {
				return nil
			}
			port, err := strconv.Atoi(value)
			if err != nil {
				return errors.New("port must be a number")
			}
			return config.ValidateSSHPort(port)
		},
	},
	{
		Key:          strings.ToLower(sshFormFieldUsername),
		Label:        sshFormFieldUsername,
		Example:      sshExampleUsername,
		ValidateFunc: config.ValidateSSHUsername,
	},
	{
		Key:          strings.ToLower(sshFormFieldAuth),
		Label:        sshFormFieldAuth,
		Kind:         SelectFieldKind,
		DefaultValue: config.SSHAuthAgent,
		Options:      SSHAuthOrder,
		ValidateFunc: config.ValidateSSHAuth,
	},
	sshSecretProviderField,
	passwordField,
	{
		Key:          strings.ToLower(sshFormFieldJump),
		Label:        sshFormFieldJump,
		Example:      sshExampleJump,
		Property:     OptionalFieldProperty,
		ValidateFunc: config.ValidateSSHJump,
	},
	{
		Key:          strings.ToLower(sshFormFieldForwardAgent),
		Label:        sshFormFieldForwardAgent,
		Kind:         SelectFieldKind,
		DefaultValue: SSHForwardAgentNo,
		Options:      []string{SSHForwardAgentNo, SSHForwardAgentYes},
	},
	{
		Key:      strings.ToLower(sshFormFieldKeepAlive),
		Label:    sshFormFieldKeepAlive,
		Example:  sshExampleKeepAlive,
		Kind:     IntFieldKind,
		Property: OptionalFieldProperty,
		ValidateFunc: func(value string) error {
			if value == "" {
				return nil
			}
			seconds, err := strconv.Atoi(value)
			if err != nil {
				return errors.New("keepalive must be a number of seconds")
			}
			return config.ValidateSSHKeepAlive(seconds)
		},
	},
	{
		Key:          strings.ToLower(sshFormFieldLocalForward),
		Label:        sshFormFieldLocalForward,
		Example:      sshExampleLocalForward,
		Property:     OptionalFieldProperty,
		ValidateFunc: config.ValidateSSHLocalForward,
	},
	{
		Key:          strings.ToLower(sshFormFieldName),
		Label:        sshFormFieldName,
		Example:      sshExampleName,
		Property:     OptionalFieldProperty,
		ValidateFunc: config.ValidateSSHName,
	},
	{
		Key:          strings.ToLower(sshFormFieldDescription),
		Label:        sshFormFieldDescription,
		Example:      sshExampleDescription,
		Property:     OptionalFieldProperty,
		ValidateFunc: config.ValidateSSHDescription,
	},
	{
		Key:      strings.ToLower(sshFormFieldTags),
		Label:    sshFormFieldTags,
		Example:  sshExampleTags,
		Property: OptionalFieldProperty,
		ValidateFunc: func(value string) error {
			if len(value) > tagsMaxLength {
				return fmt.Errorf("tags must be at most %d characters long", tagsMaxLength)
			}
			return nil
		},
	},
}

var SSHFormSpec = FormSpec[config.Connection]{
	AddTitle:  "Add a new SSH connection",
	EditTitle: "Edit an SSH connection",
	Fields:    SSHFormFields,
	Sections: []FormSection{
		{
			Title: connectionSectionTitle,
			Fields: []FormFieldKey{
				strings.ToLower(sshFormFieldHost),
				strings.ToLower(sshFormFieldPort),
				strings.ToLower(sshFormFieldUsername),
				strings.ToLower(sshFormFieldAuth),
				SecretProviderKey,
				SecretValueKey,
			},
		},
		{
			Title: sshAdvancedSectionTitle,
			Note:  sshAdvancedSectionNote,
			Fields: []FormFieldKey{
				strings.ToLower(sshFormFieldJump),
				strings.ToLower(sshFormFieldForwardAgent),
				strings.ToLower(sshFormFieldKeepAlive),
				strings.ToLower(sshFormFieldLocalForward),
			},
		},
		metadataSection(
			strings.ToLower(sshFormFieldName),
			strings.ToLower(sshFormFieldDescription),
			strings.ToLower(sshFormFieldTags),
		),
	},
	SeedFunc: func(c config.Connection) map[FormFieldKey]FormFieldValue {
		s, ok := c.(config.SSH)
		if !ok {
			return nil
		}
		mode, value := SplitSecret(s.Password)
		keepAlive := ""
		if s.KeepAlive > 0 {
			keepAlive = strconv.Itoa(s.KeepAlive)
		}
		forwardAgent := SSHForwardAgentNo
		if s.ForwardAgent {
			forwardAgent = SSHForwardAgentYes
		}
		return map[FormFieldKey]FormFieldValue{
			strings.ToLower(sshFormFieldHost):         s.Hostname,
			strings.ToLower(sshFormFieldPort):         strconv.Itoa(s.PortNumber),
			strings.ToLower(sshFormFieldUsername):     s.User,
			strings.ToLower(sshFormFieldAuth):         s.Auth,
			SecretProviderKey:                         mode,
			SecretValueKey:                            value,
			strings.ToLower(sshFormFieldJump):         s.Jump,
			strings.ToLower(sshFormFieldForwardAgent): forwardAgent,
			strings.ToLower(sshFormFieldKeepAlive):    keepAlive,
			strings.ToLower(sshFormFieldLocalForward): s.LocalForward,
			strings.ToLower(sshFormFieldName):         s.Metadata.Name,
			strings.ToLower(sshFormFieldDescription):  s.Metadata.Description,
			strings.ToLower(sshFormFieldTags):         strings.Join(s.Metadata.Tags, ", "),
		}
	},
	BuildFunc: func(values map[FormFieldKey]FormFieldValue) (config.Connection, error) {
		port, err := intValue(values, strings.ToLower(sshFormFieldPort), sshDefaultPort)
		if err != nil {
			return nil, err
		}

		keepAlive, err := intValue(values, strings.ToLower(sshFormFieldKeepAlive), 0)
		if err != nil {
			return nil, err
		}

		return config.SSH{
			Metadata: buildMeta(
				values,
				strings.ToLower(sshFormFieldName),
				strings.ToLower(sshFormFieldDescription),
				strings.ToLower(sshFormFieldTags),
			),
			Hostname:   values[strings.ToLower(sshFormFieldHost)],
			PortNumber: port,
			User:       values[strings.ToLower(sshFormFieldUsername)],
			Auth:       values[strings.ToLower(sshFormFieldAuth)],
			Password: JoinSecret(
				values[SecretProviderKey],
				values[SecretValueKey],
			),
			Jump:         values[strings.ToLower(sshFormFieldJump)],
			ForwardAgent: values[strings.ToLower(sshFormFieldForwardAgent)] == SSHForwardAgentYes,
			KeepAlive:    keepAlive,
			LocalForward: values[strings.ToLower(sshFormFieldLocalForward)],
		}, nil
	},
}
