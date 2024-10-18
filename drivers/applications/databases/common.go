package databases

import (
	"fmt"
	"strconv"

	. "github.com/pure-px/torpedo/drivers/utilities"
)

type Database struct {
	Hostname        string
	DeploymentName  string
	User            string
	Password        string
	Port            int
	NodePort        int
	DBName          string
	Namespace       string
	SQLCommands     map[string]map[string][]string
	PrimaryHostName string
}

type DatabaseConfiguration struct {
	ReplicaDetails        ReplicaDetails
	DatabaseRecoveryModes map[string]string
}

type ReplicaDetails struct {
	PrimaryReplica string
	Replicas       []Replica
	// AAG Specific parameters
	AGName string
	AGId   string
}

type Replica struct {
	Name             string
	ReplicaId        string
	Role             string
	OperationalState string
	ConnectionState  string
	Health           string
	AvailabilityMode string
	FailoverMode     string
}

const (
	filePath = "/srv/pds/test_table_export.csv"
)

func GenerateRandomCypherCommands(count int, appType string) map[string][]string {
	return nil
}

type DefaultDatabase struct {
	// SQL Server database Specific details
	MDF           string
	LDF           string
	Size          int
	MaxSize       int
	FileGrowth    int
	LogSize       int
	LogMaxSize    int
	LogFileGrowth int
}

// GenerateRandomSQLCommands generates pairs of INSERT, UPDATE, SELECT and DELETE queries for a database
func GenerateRandomSQLCommands(count int, appType string) map[string][]string {
	var randomSqlCommands = make(map[string][]string)
	var tableName string
	var insertCommands []string
	var selectCommands []string
	var deleteCommands []string
	var updateCommands []string
	var copyTableCommands []string

	if appType == Postgres {
		tableName = "pg_validation_" + RandomString(5)
		insertCommands = append(insertCommands, fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
		key varchar(45) NOT NULL,
		value varchar(45) NOT NULL
	  )`, tableName))
	} else if appType == MySql {
		tableName = "mysql_validation_" + RandomString(5)
		insertCommands = append(insertCommands, fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
			`+"`key` "+`VARCHAR(45) NOT NULL ,
			value VARCHAR(255)
		  )`, tableName))
	} else if appType == SqlServer {
		tableName = "sqlserver_validation_" + RandomString(5)
		insertCommands = append(insertCommands, fmt.Sprintf(`CREATE TABLE %s (
		dbkey varchar(45) NOT NULL,
		value varchar(45) NOT NULL
	  )`, tableName))
	}

	for counter := 0; counter < count; counter++ {
		currentCounter := strconv.Itoa(counter)
		randomValue := "Value-" + RandomString(10)
		updatedRandomValue := "Value-Updated-" + RandomString(10)
		insertCommands = append(insertCommands, fmt.Sprintf("INSERT INTO %s VALUES('%s', '%s')", tableName, currentCounter, randomValue))
		if appType == Postgres {
			selectCommands = append(selectCommands, fmt.Sprintf("SELECT * FROM %s WHERE key='%s'", tableName, currentCounter))
			updateCommands = append(updateCommands, fmt.Sprintf("UPDATE %s SET value='%s' WHERE key='%s'", tableName, updatedRandomValue, currentCounter))
			deleteCommands = append(deleteCommands, fmt.Sprintf("DELETE FROM %s WHERE key='%s'", tableName, currentCounter))
			copyTableCommands = append(copyTableCommands, fmt.Sprintf("COPY %s TO '%s' WITH (FORMAT CSV, HEADER)", tableName, filePath))
		} else if appType == MySql {
			selectCommands = append(selectCommands, fmt.Sprintf("SELECT * FROM %s WHERE `key`='%s'", tableName, currentCounter))
			updateCommands = append(updateCommands, fmt.Sprintf("UPDATE %s SET value='%s' WHERE `key`='%s'", tableName, updatedRandomValue, currentCounter))
			deleteCommands = append(deleteCommands, fmt.Sprintf("DELETE FROM %s WHERE `key`='%s'", tableName, currentCounter))
			copyTableCommands = append(copyTableCommands, fmt.Sprintf("COPY %s TO '%s' WITH (FORMAT CSV, HEADER)", tableName, filePath))
		} else if appType == SqlServer {
			selectCommands = append(selectCommands, fmt.Sprintf("SELECT * FROM %s WHERE dbkey='%s'", tableName, currentCounter))
			updateCommands = append(updateCommands, fmt.Sprintf("UPDATE %s SET value='%s' WHERE dbkey='%s'", tableName, updatedRandomValue, currentCounter))
			deleteCommands = append(deleteCommands, fmt.Sprintf("DELETE FROM %s WHERE dbkey='%s'", tableName, currentCounter))
			copyTableCommands = append(copyTableCommands, fmt.Sprintf("COPY %s TO '%s' WITH (FORMAT CSV, HEADER)", tableName, filePath))
		}

	}

	randomSqlCommands["insert"] = insertCommands
	randomSqlCommands["select"] = selectCommands
	randomSqlCommands["update"] = updateCommands
	randomSqlCommands["delete"] = deleteCommands
	randomSqlCommands["copy"] = copyTableCommands

	return randomSqlCommands

}

// GenerateSQLCommandPair generates pairs of INSERT and SELECT queries for a database
func GenerateSQLCommandPair(tableName string, appType string) map[string][]string {
	var sqlCommandMap = make(map[string][]string)
	var selectQuery string
	randomKey := "key-" + RandomString(10)
	randomValue := "value-" + RandomString(10)

	insertQuery := fmt.Sprintf("INSERT INTO %s VALUES('%s', '%s')", tableName, randomKey, randomValue)
	if appType == Postgres {
		selectQuery = fmt.Sprintf("SELECT * FROM %s WHERE key='%s'", tableName, randomKey)
	} else if appType == MySql {
		selectQuery = fmt.Sprintf("SELECT * FROM %s WHERE `key`='%s'", tableName, randomKey)
	} else if appType == SqlServer {
		selectQuery = fmt.Sprintf("SELECT * FROM %s WHERE dbkey='%s'", tableName, randomKey)
	}

	sqlCommandMap["insert"] = append(sqlCommandMap["insert"], insertQuery)
	sqlCommandMap["select"] = append(sqlCommandMap["select"], selectQuery)

	return sqlCommandMap
}
