package spec

import (
	"errors"
	"strconv"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/transport"
)

const (
	TransportTypeKey FormFieldKey = "transport type"
	TransportKey     FormFieldKey = "transport"
)

const (
	TransportDirect      = "direct"
	TransportTypesSource = "transport types"

	transportSectionTitle = "transport"
	tunnelSectionTitle    = "tunnel"
)

const (
	TransportNameKey                 FormFieldKey = "name"
	transportFormFieldKind           FormFieldKey = "kind"
	TransportViaKey                  FormFieldKey = "via"
	transportFormFieldLocalPort      FormFieldKey = "local port"
	transportFormFieldConnectTimeout FormFieldKey = "connect timeout"
	transportFormFieldKeepAlive      FormFieldKey = "keepalive"
	transportFormFieldDescription    FormFieldKey = "description"
)

const transportDefaultConnectTimeout = "10"

var transportTypeField = FormField{
	Key:          TransportTypeKey,
	Label:        "Type",
	Kind:         SourceFieldKind,
	DefaultValue: TransportDirect,
	Source:       &FieldSource{Name: TransportTypesSource},
}

var transportField = FormField{
	Key:     TransportKey,
	Label:   "Transport",
	Kind:    SourceFieldKind,
	Example: "connects directly",
	Source: &FieldSource{NameFunc: func(values map[FormFieldKey]FormFieldValue) string {
		kind := values[TransportTypeKey]
		if kind == "" || kind == TransportDirect {
			return ""
		}
		return TransportsSourceOf(kind)
	}},
}

var transportSection = FormSection{
	Title:  transportSectionTitle,
	Note:   "host and port above are the address as the transport sees it",
	Fields: []FormFieldKey{TransportTypeKey, TransportKey},
}

func TransportsSourceOf(kind config.TransportKind) string {
	return "transports " + kind
}

func JoinTransport(kind config.TransportKind, name string) string {
	if kind == "" || kind == TransportDirect || name == "" {
		return ""
	}
	return transport.Ref(kind, name)
}

func SplitTransport(ref string) (config.TransportKind, string) {
	kind, name, ok := transport.ParseRef(ref)
	if !ok {
		return TransportDirect, ""
	}
	return kind, name
}

func transportTypeOf(ref string) FormFieldValue {
	kind, _ := SplitTransport(ref)
	return kind
}

func transportNameOf(ref string) FormFieldValue {
	_, name := SplitTransport(ref)
	return name
}

var TransportKindsOrder = []config.TransportKind{config.SSHTransportKind}

var TransportFormSpec = FormSpec[config.Transport]{
	AddTitle:  "Add a transport",
	EditTitle: "Edit a transport",
	Fields: []FormField{
		{
			Key:          TransportNameKey,
			Label:        "Name",
			Example:      "prod-bastion",
			ValidateFunc: config.ValidateTransportName,
		},
		{
			Key:     transportFormFieldKind,
			Label:   "Kind",
			Kind:    SelectFieldKind,
			Options: TransportKindsOrder,
		},
		{
			Key:    TransportViaKey,
			Label:  "Via",
			Kind:   SourceFieldKind,
			Source: &FieldSource{Name: config.SSHConnType.String()},
		},
		{
			Key:          transportFormFieldLocalPort,
			Label:        "Local port",
			Kind:         IntFieldKind,
			Property:     OptionalFieldProperty,
			Example:      "picked automatically",
			ValidateFunc: intValidator(config.ValidateTransportLocalPort),
		},
		{
			Key:          transportFormFieldConnectTimeout,
			Label:        "Connect timeout",
			Kind:         IntFieldKind,
			Property:     OptionalFieldProperty,
			DefaultValue: transportDefaultConnectTimeout,
			Example:      "seconds",
			ValidateFunc: intValidator(config.ValidateTransportConnectTimeout),
		},
		{
			Key:          transportFormFieldKeepAlive,
			Label:        "Keepalive",
			Kind:         IntFieldKind,
			Property:     OptionalFieldProperty,
			Example:      "seconds between probes",
			ValidateFunc: intValidator(config.ValidateTransportKeepAlive),
		},
		{
			Key:          transportFormFieldDescription,
			Label:        "Description",
			Property:     OptionalFieldProperty,
			Example:      "Tunnel into the eu-west private network",
			ValidateFunc: config.ValidateTransportDescription,
		},
	},
	Sections: []FormSection{
		{
			Title:  transportSectionTitle,
			Fields: []FormFieldKey{TransportNameKey, transportFormFieldKind, TransportViaKey},
		},
		{
			Title:  tunnelSectionTitle,
			Fields: []FormFieldKey{transportFormFieldLocalPort, transportFormFieldConnectTimeout, transportFormFieldKeepAlive},
		},
		{
			Title:  MetadataSectionTitle,
			Fields: []FormFieldKey{transportFormFieldDescription},
		},
	},
	SeedFunc: func(t config.Transport) map[FormFieldKey]FormFieldValue {
		return map[FormFieldKey]FormFieldValue{
			TransportNameKey:                 t.Name,
			transportFormFieldKind:           t.Kind,
			TransportViaKey:                  t.Via,
			transportFormFieldLocalPort:      optionalInt(t.LocalPort),
			transportFormFieldConnectTimeout: optionalInt(t.ConnectTimeout),
			transportFormFieldKeepAlive:      optionalInt(t.KeepAlive),
			transportFormFieldDescription:    t.Description,
		}
	},
	BuildFunc: func(values map[FormFieldKey]FormFieldValue) (config.Transport, error) {
		localPort, err := intValue(values, transportFormFieldLocalPort, 0)
		if err != nil {
			return config.Transport{}, err
		}
		connectTimeout, err := intValue(values, transportFormFieldConnectTimeout, 0)
		if err != nil {
			return config.Transport{}, err
		}
		keepAlive, err := intValue(values, transportFormFieldKeepAlive, 0)
		if err != nil {
			return config.Transport{}, err
		}

		return config.Transport{
			Name:           values[TransportNameKey],
			Kind:           values[transportFormFieldKind],
			Via:            values[TransportViaKey],
			LocalPort:      localPort,
			ConnectTimeout: connectTimeout,
			KeepAlive:      keepAlive,
			Description:    values[transportFormFieldDescription],
		}, nil
	},
}

func intValidator(validate func(int) error) func(string) error {
	return func(raw string) error {
		if raw == "" {
			return nil
		}
		value, err := strconv.Atoi(raw)
		if err != nil {
			return errors.New("must be a number")
		}
		return validate(value)
	}
}

func optionalInt(value int) FormFieldValue {
	if value == 0 {
		return ""
	}
	return strconv.Itoa(value)
}
