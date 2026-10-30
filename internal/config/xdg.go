package config

import (
	"path/filepath"

	"github.com/adrg/xdg"
)

const (
	conmDirName              = "conm"
	postgresConfigFileName   = "postgres.toml"
	mysqlConfigFileName      = "mysql.toml"
	mssqlConfigFileName      = "mssql.toml"
	clickhouseConfigFileName = "clickhouse.toml"
	redisConfigFileName      = "redis.toml"
	mongodbConfigFileName    = "mongodb.toml"
	sshConfigFileName        = "ssh.toml"
	transportConfigFileName  = "transport.toml"
	conmConfigFileName       = "conm.toml"
	secretConfigFileName     = "secret.toml"
)

var (
	PostgresPath     = configPath(postgresConfigFileName)
	MySQLPath        = configPath(mysqlConfigFileName)
	MSSQLPath        = configPath(mssqlConfigFileName)
	ClickHousePath   = configPath(clickhouseConfigFileName)
	RedisPath        = configPath(redisConfigFileName)
	MongoDBPath      = configPath(mongodbConfigFileName)
	SSHPath          = configPath(sshConfigFileName)
	TransportPath    = configPath(transportConfigFileName)
	ConmPath         = configPath(conmConfigFileName)
	SecretConfigPath = dataPath(secretConfigFileName)
)

func configPath(fileName string) func() string {
	return func() string {
		return workspacePath(xdg.ConfigHome, fileName)
	}
}

func dataPath(fileName string) func() string {
	return func() string {
		return workspacePath(xdg.DataHome, fileName)
	}
}

func SecretConfigPathOf(ws string) string {
	return filepath.Join(workspaceDir(xdg.DataHome, ws), secretConfigFileName)
}

func workspacePath(home, fileName string) string {
	return filepath.Join(workspaceDir(home, workspace), fileName)
}

func workspaceDir(home, ws string) string {
	return filepath.Join(home, conmDirName, ws)
}
