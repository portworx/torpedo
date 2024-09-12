package databases

import "time"

// Commands to start, pause or stop data
const (
	DataStart   = "Start"
	DataPause   = "Pause"
	DataStop    = "Stop"
	DataRestart = "Restart"
)

// Database related constants
const (
	MySql     = "mysql"
	Postgres  = "postgressql"
	SqlServer = "sqlserver"
	Neo4j    = "neo4j"
)

// New Database specific constants for SQL Server
const (
	defaultMDFSize       = 5
	defaultMDFMaxSize    = 10
	defaultMDFFileGrowth = 1
	defaultLDFSize       = 25
	defaultLDFMaxSize    = 50
	defaultLDFFileGrowth = 5
)

const (
	retryTimeout  = 2 * time.Minute
	retryInterval = 10 * time.Second
)
