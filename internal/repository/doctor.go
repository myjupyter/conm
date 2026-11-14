package repository

import (
	"context"

	"github.com/myjupyter/conm/internal/cli"
	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/transport"
)

type ClientCheck struct {
	Type    config.ConnType
	Client  cli.Info
	Invalid error
}

type FileCheck struct {
	Kind    string
	Err     error
	Records []RecordCheck
}

type RecordCheck struct {
	Name string
	Errs []error
}

func (c ClientCheck) Ready() bool {
	return c.Invalid == nil && c.Client.Installed
}

func CheckClients(ctx context.Context, cfg config.Conm) []ClientCheck {
	checks := make([]ClientCheck, 0, len(config.ConnTypes))
	for _, t := range config.ConnTypes {
		checks = append(checks, checkClient(ctx, cfg, t))
	}

	return checks
}

func checkClient(ctx context.Context, cfg config.Conm, t config.ConnType) ClientCheck {
	configured := cfg.CLI(t)
	check := ClientCheck{Type: t, Client: cli.Info{Name: cli.Resolve(t, configured)}}
	if configured != "" {
		check.Invalid = cli.Validate(t, configured)
	}

	for _, info := range cli.Detect(t) {
		if info.Name == check.Client.Name {
			check.Client = info
		}
	}
	if check.Client.Installed {
		check.Client.Version = cli.Version(ctx, check.Client.Name)
	}

	return check
}

func CheckFiles(cfg config.Conm) []FileCheck {
	trs, err := NewTransportRepository(nil, nil, nil)
	transportCheck := FileCheck{Kind: "transport", Err: err}

	var transports transport.Provider
	if err == nil {
		transports = trs
		for i := range trs.Len() {
			t, _ := trs.Get(i)
			transportCheck.Records = appendInvalid(transportCheck.Records, t.Name, t.Validate())
		}
	}

	checks := make([]FileCheck, 0, 1+len(config.ConnTypes))
	checks = append(checks, transportCheck)
	for _, t := range config.ConnTypes {
		checks = append(checks, checkConnections(cfg, t, transports))
	}

	return checks
}

func checkConnections(cfg config.Conm, t config.ConnType, transports transport.Provider) FileCheck {
	check := FileCheck{Kind: t.String()}

	repo, err := connectionRepositories[t](cfg, transports, nil, nil)
	if err != nil {
		check.Err = err
		return check
	}
	defer repo.Close()

	for i := range repo.Len() {
		c, _ := repo.ConfigAt(i)
		check.Records = appendInvalid(check.Records, c.Meta().Name, c.Validate())
	}

	return check
}

func appendInvalid(records []RecordCheck, name string, errs []error) []RecordCheck {
	if len(errs) == 0 {
		return records
	}

	return append(records, RecordCheck{Name: name, Errs: errs})
}
