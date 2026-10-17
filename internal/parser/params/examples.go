package params

import "github.com/myjupyter/conm/internal/config"

var examples = map[config.ConnType][]string{
	config.PostgresConnType: {
		`psql -h localhost -p 5432 -U me mydb`,
		`pgcli --host localhost --dbname mydb --user me`,
		`postgres://me@localhost:5432/mydb?sslmode=require`,
	},
	config.MySQLConnType: {
		`mysql -h db.internal -P 3306 -u root shop`,
		`mysql://app@db.internal:3306/shop`,
	},
	config.MSSQLConnType: {
		`sqlcmd -S sql.internal,1433 -U sa -d app`,
		`sqlserver://sa@sql.internal:1433?database=app`,
	},
	config.ClickHouseConnType: {
		`clickhouse client --host ch.internal --user app --database metrics`,
		`clickhouse://app@ch.internal:9000/metrics`,
	},
	config.RedisConnType: {
		`redis-cli -h cache.internal -p 6379 -n 3`,
		`redis://default@cache.internal:6379/3`,
	},
	config.MongoDBConnType: {
		`mongosh --host mongo.internal --port 27017 -u app records`,
		`mongodb+srv://app@cluster.example.net/records`,
	},
	config.SSHConnType: {
		`ssh -p 2222 -i ~/.ssh/id_ed25519 deploy@bastion.internal`,
		`ssh://deploy@bastion.internal:2222`,
	},
}

func Examples(t config.ConnType) []string {
	return examples[t]
}
