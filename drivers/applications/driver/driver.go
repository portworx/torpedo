package driver

import (
	"context"

	. "github.com/portworx/torpedo/drivers/applications/apptypes"
	. "github.com/portworx/torpedo/drivers/applications/mysql"
	. "github.com/portworx/torpedo/drivers/applications/postgres"
	. "github.com/portworx/torpedo/drivers/utilities"
)

type ApplicationDriver interface {
	DefaultPort() int

	DefaultDBName() string

	ExecuteCommand(commands []string, ctx context.Context) error

	StartData(command <-chan string, ctx context.Context) error

	CheckDataPresent(selectQueries []string, ctx context.Context) error

	UpdateDataCommands(count int, identifier string)

	InsertBackupData(ctx context.Context, identifier string, commnads []string) error

	GetBackupData(identifier string) []string

	GetRandomDataCommands(count int) map[string][]string

	AddDataCommands(identifier string, commands map[string][]string)

	GetApplicationType() string

	GetNamespace() string
}

// Returns struct of appType provided as input
func GetApplicationDriver(appType string, hostname string, user string,
	password string, port int, dbname string, ctx context.Context, nodePort int, namespace string) (ApplicationDriver, error) {

	switch appType {
	case Postgres:
		return &PostgresConfig{
			Hostname: hostname,
			User:     user,
			Password: password,
			Port:     port,
			DBName:   dbname,
			SQLCommands: map[string]map[string][]string{
				"default": GenerateRandomSQLCommands(20, appType),
			},
			NodePort:  nodePort,
			Namespace: namespace,
		}, nil
	case MySql:
		return &MySqlConfig{
			Hostname: hostname,
			User:     user,
			Password: password,
			Port:     port,
			DBName:   dbname,
			SQLCommands: map[string]map[string][]string{
				"default": GenerateRandomSQLCommands(20, appType),
			},
			NodePort:  nodePort,
			Namespace: namespace,
		}, nil
	default:
		return &PostgresConfig{
			Hostname: hostname,
			User:     user,
			Password: password,
			Port:     port,
			DBName:   dbname,
			SQLCommands: map[string]map[string][]string{
				"default": GenerateRandomSQLCommands(20, appType),
			},
			NodePort:  nodePort,
			Namespace: namespace,
		}, nil

	}
}
