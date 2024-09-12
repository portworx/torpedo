package pds

const (
	postgres = "postgresql"
	mssql    = "sqlserver"
	neo4j    = "neo4j"
)

// DATASERVICEWITHQUERYSUPPORT contains all data service for which query support is enabled
var DATASERVICEWITHQUERYSUPPORT = []string{
	postgres,
	mssql,
	neo4j,
}

// defaultDatabaseNames hold default database present on the database engine
var defaultDatabaseNames = map[string]string{
	postgres: "pds",
	mssql:    "master",
	neo4j:    "neo4j",
}
