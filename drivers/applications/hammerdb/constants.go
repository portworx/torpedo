package hammerdb

import "time"

// Database related constants
const (
	MySql     = "mysql"
	Postgres  = "postgressql"
	SqlServer = "sqlserver"
	Neo4j     = "neo4j"
)

const (
	defaultPath = "./"
)

const (
	hammerDBPodTimeout       = 5 * time.Minute
	hammerDBPodRetryInterval = 10 * time.Second
)
