package databases

import (
	"context"
	"github.com/portworx/torpedo/drivers/node"
)

type DatabaseDriver interface {
	// DefaultPort returns the default port for the application
	DefaultPort() int

	// ExecuteCommand executes a command on the application
	ExecuteCommand(commands []string, ctx context.Context) ([]string, error)

	// StartData starts injecting continous data to the application
	StartData(command <-chan string, ctx context.Context) error

	// CheckDataPresent checks if the passed data is present in the app or not
	CheckDataPresent(selectQueries []string, ctx context.Context) error

	// UpdateDataCommands updates the SQL queries/data injection commands for the app
	UpdateDataCommands(count int, identifier string)

	// InsertBackupData inserts data before and after backup
	InsertBackupData(ctx context.Context, identifier string, commands []string) error

	// GetBackupData gets the sql queries inserted before or after backup
	GetBackupData(identifier string) []string

	// CopyBackupData copies the table and saves as csv file
	CopyBackupData(ctx context.Context, identifier string) error

	// GetRandomDataCommands creates random CRUD queries for an application
	GetRandomDataCommands(count int) map[string][]string

	// AddDataCommands adds CRUD queries to the application struct
	AddDataCommands(identifier string, commands map[string][]string)

	// GetApplicationType retruns the application type
	GetApplicationType() string

	// GetNamespace returns the application namespace
	GetNamespace() string

	// GetReplicaDetails returns details of all replicas
	GetReplicaDetails(context.Context) (ReplicaDetails, error)

	//AddDatabaseToCluster add specified database to the cluster
	AddDatabaseToCluster(ctx context.Context, databaseName string, clusterName string) error

	// Fetch Default database name for the engine
	DefaultDBName() string

	// Creates new database on the given dataservice
	CreateNewDatabase(ctx context.Context, databaseName string, databaseDetails DefaultDatabase) error

	// Fetches default new database configuration
	GetDefaultNewDatabaseConfig(databaseName string) DefaultDatabase

	// Executes command on primary replica
	ExecuteCommandOnPrimary(commands []string, ctx context.Context) ([]string, error)

	// Failover triggers fail over on the application
	Failover(ctx context.Context) error

	// GetDataserviceConfigurations returns the configuration of the dataservice
	GetDataserviceConfigurations(ctx context.Context) (DatabaseConfiguration, error)

	// GetDatabaseDetails returns the database details
	GetDatabaseDetails() Database
}

// GetApplicationDriver returns struct of appType provided as input
func GetDatabaseDriver(appType string, hostname string, user string,
	password string, port int, dbname string, nodePort int, namespace string, IPAddress string,
	nodeDriver node.Driver, stsName string) (DatabaseDriver, error) {

	switch appType {
	case Postgres:
		return &PostgresConfig{
			Database: Database{
				Hostname: hostname,
				User:     user,
				Password: password,
				Port:     port,
				DBName:   dbname,
				SQLCommands: map[string]map[string][]string{
					"default": GenerateRandomSQLCommands(20, appType),
				},
				NodePort:       nodePort,
				Namespace:      namespace,
				DeploymentName: stsName,
			},
		}, nil
	case MySql:
		return &MySqlConfig{
			Database: Database{
				Hostname: hostname,
				User:     user,
				Password: password,
				Port:     port,
				DBName:   dbname,
				SQLCommands: map[string]map[string][]string{
					"default": GenerateRandomSQLCommands(20, appType),
				},
				NodePort:       nodePort,
				Namespace:      namespace,
				DeploymentName: stsName,
			},
		}, nil

	case SqlServer:
		return &SqlServerConfig{
			Database: Database{
				Hostname: hostname,
				User:     user,
				Password: password,
				Port:     port,
				DBName:   dbname,
				SQLCommands: map[string]map[string][]string{
					"default": GenerateRandomSQLCommands(20, appType),
				},
				NodePort:       nodePort,
				Namespace:      namespace,
				DeploymentName: stsName,
			},
		}, nil
	case Neo4j:
		return &Neo4jConfig{
			Hostname: hostname,
			User:     user,
			Password: password,
			Port:     port,
			DBName:   dbname,
			CypherCommands: map[string]map[string][]string{
				"default": GenerateRandomSQLCommands(20, appType),
			},
			NodePort:  nodePort,
			Namespace: namespace,
		}, nil

	default:
		return &PostgresConfig{
			Database: Database{
				Hostname: hostname,
				User:     user,
				Password: password,
				Port:     port,
				DBName:   dbname,
				SQLCommands: map[string]map[string][]string{
					"default": GenerateRandomSQLCommands(20, appType),
				},
				NodePort:       nodePort,
				Namespace:      namespace,
				DeploymentName: stsName,
			},
		}, nil

	}
}
