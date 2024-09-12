package databases

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/portworx/sched-ops/task"
	"strings"
	"time"

	. "github.com/portworx/torpedo/drivers/utilities"
	"github.com/portworx/torpedo/pkg/log"

	_ "github.com/denisenkom/go-mssqldb"
)

type SqlServerConfig struct {
	Database
	Replicas ReplicaDetails
}

const (
	// Query to fetch status and details of AG cluster and replicas
	queryToFetchReplicaDetails = "SELECT C.name, convert(nvarchar(50), C.group_id), convert(nvarchar(50), REPLICAS.replica_id), CS.replica_server_name, RS.role_desc, ISNULL(RS.operational_state_desc, ''), RS.connected_state_desc, RS.synchronization_health_desc, REPLICAS.availability_mode_desc, REPLICAS.failover_mode_desc FROM master.sys.availability_groups_cluster AS C INNER JOIN master.sys.dm_hadr_availability_replica_cluster_states AS CS ON CS.group_id = C.group_id INNER JOIN master.sys.dm_hadr_availability_replica_states AS RS ON RS.replica_id = CS.replica_id INNER JOIN master.sys.availability_replicas AS REPLICAS ON RS.replica_id = REPLICAS.replica_id"
	// Query to manually fail over the AAG replica
	queryToAddDatabaseToAAG = "USE master; ALTER AVAILABILITY GROUP [%s] ADD DATABASE [%s]"
	// Query to create new databse
	queryToCreateDB = "CREATE DATABASE %s ON (NAME = %s_dat, FILENAME = '%s', SIZE = %d, MAXSIZE = %d, FILEGROWTH = %d) LOG ON (NAME = %s_log, FILENAME = '%s', SIZE = %d, MAXSIZE = %d, FILEGROWTH = %d)"
	// Query to failover database
	failoverQuery = "EXEC sp_set_session_context @key = N'external_cluster', @value = N'yes';USE master;ALTER AVAILABILITY GROUP [%s] FAILOVER;"
	// Query to fetch recovery modes
	getAllRecoveryModes = "USE master;SELECT name, recovery_model_desc FROM sys.databases"
)

func (app *SqlServerConfig) DefaultPort() int { return 1433 }

func (app *SqlServerConfig) DefaultDBName() string { return app.DBName }

// GetConnection returns a connection object for SqlServer database
func (app *SqlServerConfig) GetConnection(ctx context.Context) (*sql.DB, error) {

	if app.Port == 0 {
		app.Port = app.DefaultPort()
	}

	if app.DBName == "" {
		app.DBName = app.DefaultDBName()
	}

	var url string

	if app.NodePort != 0 {
		// Connect with NodePort Service
		if app.DBName == "" {
			// Connect directly to the instance
			url = fmt.Sprintf("sqlserver://%s:%s@%s:%d",
				app.User, app.Password, app.Hostname, app.NodePort)
		} else {
			url = fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s",
				app.User, app.Password, app.Hostname, app.NodePort, app.DBName)
		}
	} else {
		// Connect with Cluster Service
		if app.DBName == "" {
			// Connect directly to the instance
			url = fmt.Sprintf("sqlserver://%s:%s@%s:%d",
				app.User, app.Password, app.Hostname, app.Port)
		} else {
			url = fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s",
				app.User, app.Password, app.Hostname, app.Port, app.DBName)
		}
	}

	conn, err := sql.Open("mssql", url)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %s", err)
	}

	waitForPing := func() (interface{}, bool, error) {
		err := conn.Ping()
		if err != nil {
			return nil, true, fmt.Errorf("some error occurred while pinging the database from GetConnection. Error - [%s]", err.Error())
		}
		return nil, false, nil
	}

	_, err = task.DoRetryWithTimeout(waitForPing, retryTimeout, retryInterval)

	if err != nil {
		return nil, fmt.Errorf("unable to get connection - [%s]", err.Error())
	}

	return conn, nil
}

// GetConnection returns a connection object for SqlServer database
func (app *SqlServerConfig) getPrimaryConnection(ctx context.Context) (*sql.DB, error) {

	var url string

	svcName := fmt.Sprintf("%s-%s-rw", app.DeploymentName, app.Namespace)

	primaryHostName := fmt.Sprintf("%s.%s.svc.cluster.local", svcName, app.Namespace)
	log.Infof("Primary Replica Conn String - [%s]", primaryHostName)

	// Connect directly to the instance
	url = fmt.Sprintf("sqlserver://%s:%s@%s:%d",
		app.User, app.Password, primaryHostName, app.Port)

	if app.DBName != "" {
		log.Infof("Connecting to [%s]", app.DBName)
		url = fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s",
			app.User, app.Password, primaryHostName, app.Port, app.DBName)
	}

	conn, err := sql.Open("mssql", url)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %s", err)
	}

	waitForPing := func() (interface{}, bool, error) {
		err := conn.Ping()
		if err != nil {
			return nil, true, fmt.Errorf("some error occurred while pinging the database from getPrimaryConnection. Error - [%s]", err.Error())
		}
		return nil, false, nil
	}

	_, err = task.DoRetryWithTimeout(waitForPing, retryTimeout, retryInterval)

	if err != nil {
		return nil, fmt.Errorf("unable to get connection - [%s]", err.Error())
	}

	app.PrimaryHostName = primaryHostName

	return conn, nil
}

// GetConnection returns a connection object for SqlServer database
func (app *SqlServerConfig) getSecondaryConnection(ctx context.Context) (*sql.DB, error) {

	var url string

	svcName := fmt.Sprintf("%s-%s-rr", app.DeploymentName, app.Namespace)

	primaryHostName := fmt.Sprintf("%s.%s.svc.cluster.local", svcName, app.Namespace)
	log.Infof("Primary Replica Conn String - [%s]", primaryHostName)

	// Connect directly to the instance
	url = fmt.Sprintf("sqlserver://%s:%s@%s:%d",
		app.User, app.Password, primaryHostName, app.Port)

	conn, err := sql.Open("mssql", url)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %s", err)
	}

	waitForPing := func() (interface{}, bool, error) {
		err := conn.Ping()
		if err != nil {
			return nil, true, fmt.Errorf("some error occurred while pinging the database from getPrimaryConnection. Error - [%s]", err.Error())
		}
		return nil, false, nil
	}

	_, err = task.DoRetryWithTimeout(waitForPing, retryTimeout, retryInterval)

	if err != nil {
		return nil, fmt.Errorf("unable to get connection - [%s]", err.Error())
	}

	return conn, nil
}

// ExecuteCommand executes a SQL command for SqlServer database
func (app *SqlServerConfig) ExecuteCommand(commands []string, ctx context.Context) ([]string, error) {

	var dummy []string
	conn, err := app.GetConnection(ctx)
	if err != nil {
		return dummy, err
	}

	defer conn.Close()

	for _, eachCommand := range commands {
		log.Infof("Command to be run - [%s]", eachCommand)
		_, err = conn.ExecContext(ctx, eachCommand)
		if err != nil {
			return dummy, err
		}
	}
	return dummy, nil
}

// ExecuteCommand executes a SQL command for Primary Replica SqlServer database
func (app *SqlServerConfig) ExecuteCommandOnPrimary(commands []string, ctx context.Context) ([]string, error) {

	var dummy []string

	log.Infof("Forcing to run this command on primary replica")

	conn, err := app.getPrimaryConnection(ctx)
	if err != nil {
		return dummy, err
	}

	defer conn.Close()

	for _, eachCommand := range commands {
		log.Infof("Command to be run - [%s]", eachCommand)
		_, err = conn.ExecContext(ctx, eachCommand)
		if err != nil {
			return dummy, err
		}
	}
	return dummy, nil
}

// InsertBackupData inserts the rows generated initially by utilities or rows passed
func (app *SqlServerConfig) InsertBackupData(ctx context.Context, identifier string, commnads []string) error {

	var err error
	log.InfoD("Inserting data")
	if len(commnads) == 0 {
		log.Infof("Inserting below data : %s", strings.Join(app.SQLCommands[identifier]["insert"], "\n"))
		_, err = app.ExecuteCommand(app.SQLCommands[identifier]["insert"], ctx)
	} else {
		log.Infof("Inserting below data : %s", strings.Join(commnads, "\n"))
		_, err = app.ExecuteCommand(commnads, ctx)
	}

	return err
}

// Return data inserted before backup
func (app *SqlServerConfig) GetBackupData(identifier string) []string {
	if _, ok := app.SQLCommands[identifier]; ok {
		return app.SQLCommands[identifier]["select"]
	} else {
		log.InfoD("%s not found in app sql command", identifier)
		log.Infof("All current SQL commands - %+v", app.SQLCommands)
		return nil
	}
}

// CheckDataPresent checks if the mentioned entry is present or not in the database
func (app *SqlServerConfig) CheckDataPresent(selectQueries []string, ctx context.Context) error {

	log.InfoD("Running Select Queries")

	conn, err := app.GetConnection(ctx)

	if err != nil {
		return err
	}

	defer conn.Close()

	var key string
	var value string
	var queryNotFoundList []string

	for _, eachQuery := range selectQueries {
		log.Debugf("Running select query - [%s]", eachQuery)
		currentRow := conn.QueryRowContext(ctx, eachQuery)
		err := currentRow.Scan(&key, &value)

		if err != nil {
			log.InfoD("Select query failed - [%s] Error - [%s]", eachQuery, err.Error())
			queryNotFoundList = append(queryNotFoundList, eachQuery)
		}
	}

	if len(queryNotFoundList) != 0 {
		errorMessage := strings.Join(queryNotFoundList, "\n")
		return fmt.Errorf("Below results not found in the table:\n %s", errorMessage)
	}
	return nil
}

// UpdateBackupData updates the rows generated initially by utilities
func (app *SqlServerConfig) UpdateBackupData(ctx context.Context, identifier string) error {

	log.InfoD("Running Update Queries")
	_, err := app.ExecuteCommand(app.SQLCommands[identifier]["update"], ctx)

	return err
}

// DeleteBackupData deletes the rows generated initially by utilities
func (app *SqlServerConfig) DeleteBackupData(ctx context.Context, identifier string) error {

	log.InfoD("Running Delete Queries")
	_, err := app.ExecuteCommand(app.SQLCommands[identifier]["delete"], ctx)

	return err
}

// StartData - Go routine to run parallal with app to keep injecting data every 2 seconds
func (app *SqlServerConfig) StartData(command <-chan string, ctx context.Context) error {
	var status = DataStart
	var allSelectCommands []string
	var allErrors []string
	var tableName = "sqlserver_table_" + RandomString(4)

	createTableQuery := fmt.Sprintf(`CREATE TABLE %s (
		dbkey varchar(45) NOT NULL,
		value varchar(45) NOT NULL
	  )`, tableName)
	_, err := app.ExecuteCommand([]string{createTableQuery}, ctx)
	if err != nil {
		allErrors = append(allErrors, fmt.Sprintf("Continuity Pipeline Error - [%s] at [%s]", err.Error(), time.Now().Format("2006-01-02 15:04:05")))
	}
	for {
		select {
		case cmd := <-command:
			switch cmd {
			case DataStop:
				if len(allErrors) != 0 {
					return fmt.Errorf(strings.Join(allErrors, "\n"))
				}
				err := app.CheckDataPresent(allSelectCommands, ctx)
				return err
			case DataRestart:

				allSelectCommands = []string{}

				createTableQuery := fmt.Sprintf(`CREATE TABLE %s (
					dbkey varchar(45) NOT NULL,
					value varchar(45) NOT NULL
	  				)`, tableName)

				_, err = app.ExecuteCommand([]string{createTableQuery}, ctx)

				if err != nil {
					allErrors = append(allErrors, fmt.Sprintf("Continuity Pipeline Error - [%s] at [%s]", err.Error(), time.Now().Format("2006-01-02 15:04:05")))
				}

				// Starting data again on new database
				status = DataStart

			case DataPause:
				status = DataPause
			default:
				status = DataStart
			}
		default:
			if status == DataStart {
				commandPair, err := app.startInsertingData(tableName, ctx)
				if err != nil {
					allErrors = append(allErrors, fmt.Sprintf("Continuity Pipeline Error - [%s] at [%s]", err.Error(), time.Now().Format("2006-01-02 15:04:05")))
				}
				allSelectCommands = append(allSelectCommands, commandPair["select"]...)
				time.Sleep(2 * time.Second)
			}
		}
	}
}

// startInsertingData is helper to insert generate rows and insert data parallely for SqlServer app
func (app *SqlServerConfig) startInsertingData(tableName string, ctx context.Context) (map[string][]string, error) {

	commandPair := GenerateSQLCommandPair(tableName, SqlServer)

	_, err := app.ExecuteCommandOnPrimary(commandPair["insert"], ctx)
	if err != nil {
		return commandPair, err
	}

	return commandPair, nil
}

// CopyBackupData copies the table and saves as csv file
func (app *SqlServerConfig) CopyBackupData(ctx context.Context, identifier string) error {
	return nil
}

// Update the existing SQL commands
func (app *SqlServerConfig) UpdateDataCommands(count int, identifier string) {
	app.SQLCommands[identifier] = GenerateRandomSQLCommands(count, SqlServer)
	log.InfoD("SQL Commands updated")
}

// Update the existing SQL commands
func (app *SqlServerConfig) AddDataCommands(identifier string, commands map[string][]string) {
	app.SQLCommands[identifier] = commands
	log.InfoD("Sql commands added")
}

// Generate and return random SQL commands
func (app *SqlServerConfig) GetRandomDataCommands(count int) map[string][]string {
	return GenerateRandomSQLCommands(count, SqlServer)
}

// Get the application type
func (app *SqlServerConfig) GetApplicationType() string {
	return SqlServer
}

// Get Namespace of the app
func (app *SqlServerConfig) GetNamespace() string {
	return app.Namespace
}

// GetReplicaDetails gets the replica details for given deployment
func (app *SqlServerConfig) GetReplicaDetails(ctx context.Context) (ReplicaDetails, error) {

	log.Debugf("Fetching SQL replica details")

	var replicaDetails = ReplicaDetails{
		Replicas: make([]Replica, 0),
	}
	conn, err := app.getPrimaryConnection(ctx)
	if err != nil {
		log.Errorf("Error - [%s]", err.Error())
		return replicaDetails, err
	}
	defer conn.Close()

	rows, err := conn.Query(queryToFetchReplicaDetails)
	if err != nil {
		log.Errorf("Error - [%s]", err.Error())
		return replicaDetails, err
	}

	defer rows.Close()

	for rows.Next() {
		currentreplicaDetail := Replica{}
		err = rows.Scan(
			&replicaDetails.AGName,
			&replicaDetails.AGId,
			&currentreplicaDetail.ReplicaId,
			&currentreplicaDetail.Name,
			&currentreplicaDetail.Role,
			&currentreplicaDetail.OperationalState,
			&currentreplicaDetail.ConnectionState,
			&currentreplicaDetail.Health,
			&currentreplicaDetail.AvailabilityMode,
			&currentreplicaDetail.FailoverMode)

		if err != nil {
			log.Errorf("Error - [%s]", err.Error())
			return replicaDetails, err
		}

		if currentreplicaDetail.Role == "PRIMARY" {
			replicaDetails.PrimaryReplica = currentreplicaDetail.Name
		}

		log.Debugf("Current Replica - [%v]", currentreplicaDetail)
		replicaDetails.Replicas = append(replicaDetails.Replicas, currentreplicaDetail)

		err = rows.Err()
		if err != nil {
			log.Errorf("Error - [%s]", err.Error())
			return replicaDetails, err
		}
	}

	log.InfoD("Total replicas found under [%s] - [%d]", replicaDetails.AGName, len(replicaDetails.Replicas))

	return replicaDetails, nil
}

// AddDatabaseToCluster Adds database to AAG Cluster for SQL Server
func (app *SqlServerConfig) AddDatabaseToCluster(ctx context.Context, databaseName string, clusterName string) error {

	query := fmt.Sprintf(queryToAddDatabaseToAAG, clusterName, databaseName)

	_, err := app.ExecuteCommandOnPrimary([]string{query}, ctx)

	return err
}

// CreateNewDatabase creates a new database on the given data service
func (app *SqlServerConfig) CreateNewDatabase(ctx context.Context, databaseName string, databaseDetails DefaultDatabase) error {

	// Setting DBName to "" to connect to the instance and create new database
	app.DBName = ""

	query := fmt.Sprintf(
		queryToCreateDB,
		databaseName,                  // Database Name
		databaseName,                  // MDF Name
		databaseDetails.MDF,           // MDF File Name
		databaseDetails.Size,          // MDF Size
		databaseDetails.MaxSize,       // MDF Max Size
		databaseDetails.FileGrowth,    // MDF File Growth
		databaseName,                  // LDF name
		databaseDetails.LDF,           // LDF File
		databaseDetails.LogSize,       // LDF Log Size
		databaseDetails.LogMaxSize,    // LDF Max Size
		databaseDetails.LogFileGrowth, // LDF File Growth
	)

	log.Infof("Query to be run - [%s]", query)

	_, err := app.ExecuteCommandOnPrimary([]string{query}, ctx)
	if err != nil {
		return err
	}

	// Sleeping explicitly for one minute after DB creation
	time.Sleep(1 * time.Minute)

	// Setting default database to the new database
	app.DBName = databaseName

	return nil
}

// GetDefaultNewDatabaseConfig will get default database config details
func (app *SqlServerConfig) GetDefaultNewDatabaseConfig(databaseName string) DefaultDatabase {

	return DefaultDatabase{
		MDF:           fmt.Sprintf("/var/opt/mssql/data/%s.mdf", databaseName),
		LDF:           fmt.Sprintf("/var/opt/mssql/data/%slog.ldf", databaseName),
		Size:          defaultMDFSize,
		MaxSize:       defaultMDFMaxSize,
		FileGrowth:    defaultMDFFileGrowth,
		LogSize:       defaultLDFSize,
		LogMaxSize:    defaultLDFMaxSize,
		LogFileGrowth: defaultLDFFileGrowth,
	}

}

// Failover will failover to secondary replica
func (app *SqlServerConfig) Failover(ctx context.Context) error {

	replicaDetails, err := app.GetReplicaDetails(ctx)
	if err != nil {
		return err
	}

	conn, err := app.getSecondaryConnection(ctx)
	if err != nil {
		return err
	}

	failoverQueryToBeRun := fmt.Sprintf(failoverQuery, replicaDetails.AGName)
	log.Infof("Failover query - [%s]", failoverQueryToBeRun)

	_, err = conn.ExecContext(ctx, failoverQueryToBeRun)

	return err
}

// GetDataserviceConfigurations returns the database configuration details
func (app *SqlServerConfig) GetDataserviceConfigurations(ctx context.Context) (DatabaseConfiguration, error) {

	replicaDetails, err := app.GetReplicaDetails(ctx)
	if err != nil {
		return DatabaseConfiguration{}, err
	}

	recoveryMode, err := app.getRecoveryModes(ctx)
	if err != nil {
		return DatabaseConfiguration{}, err
	}

	result := DatabaseConfiguration{
		ReplicaDetails:        replicaDetails,
		DatabaseRecoveryModes: recoveryMode,
	}

	return result, nil

}

func (app *SqlServerConfig) getRecoveryModes(ctx context.Context) (map[string]string, error) {

	conn, err := app.getPrimaryConnection(ctx)
	if err != nil {
		return nil, err
	}

	defer conn.Close()

	rows, err := conn.Query(getAllRecoveryModes)
	if err != nil {
		log.Errorf("Error - [%s]", err.Error())
		return nil, err
	}

	defer rows.Close()

	var databaseName string
	var recoveryMode string

	var recoveryModeMap = make(map[string]string)

	for rows.Next() {

		err = rows.Scan(
			&databaseName,
			&recoveryMode,
		)

		if err != nil {
			log.Errorf("Error - [%s]", err.Error())
			return nil, err
		}

		log.Debugf("Database Name - [%s], Recover Mode - [%s]", databaseName, recoveryMode)
		recoveryModeMap[databaseName] = recoveryMode

		err = rows.Err()
		if err != nil {
			log.Errorf("Error - [%s]", err.Error())
			return nil, err
		}
	}

	return recoveryModeMap, nil

}

// GetDatabaseDetails returns the database details
func (app *SqlServerConfig) GetDatabaseDetails() Database {
	return app.Database
}
