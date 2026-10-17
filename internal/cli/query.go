package cli

import "strings"

var queryArgs = map[string]func(query []string) []string{
	Psql:       flagQuery("-c"),
	Pgcli:      flagQuery("-c"),
	USQL:       flagQuery("-c"),
	MySQL:      flagQuery("-e"),
	MyCLI:      flagQuery("-e"),
	SQLCmd:     flagQuery("-Q"),
	ClickHouse: flagQuery("--query"),
	Mongosh:    flagQuery("--eval"),
	Mongo:      flagQuery("--eval"),
	RedisCLI:   plainQuery,
	ValkeyCLI:  plainQuery,
	IRedis:     plainQuery,
	SSH:        plainQuery,
}

func flagQuery(flag string) func([]string) []string {
	return func(query []string) []string {
		return []string{flag, strings.Join(query, " ")}
	}
}

func plainQuery(query []string) []string {
	return query
}
