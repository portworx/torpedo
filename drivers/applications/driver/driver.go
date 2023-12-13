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

	UpdateSQLCommands(count int)

	InsertBackupData(ctx context.Context) error

	InsertPostBackupData(ctx context.Context) error

	GetPreBackupData() []string

	GetPostBackupData() []string
}

// Returns struct of appType provided as input
func GetApplicationDriver(appType string, hostname string, user string,
	password string, port int, dbname string, ctx context.Context, nodePort int) (ApplicationDriver, error) {

	switch appType {
	case Postgres:
		return &PostgresConfig{
			Hostname:              hostname,
			User:                  user,
			Password:              password,
			Port:                  port,
			DBName:                dbname,
			SQLCommands:           GenerateRandomSQLCommands(20, appType),
			SQLCommandsPostBackup: GenerateRandomSQLCommands(20, appType),
			NodePort:              nodePort,
		}, nil
	case MySql:
		return &MySqlConfig{
			Hostname:              hostname,
			User:                  user,
			Password:              password,
			Port:                  port,
			DBName:                dbname,
			SQLCommands:           GenerateRandomSQLCommands(20, appType),
			SQLCommandsPostBackup: GenerateRandomSQLCommands(20, appType),
			NodePort:              nodePort,
		}, nil
	default:
		return &PostgresConfig{
			Hostname:              hostname,
			User:                  user,
			Password:              password,
			Port:                  port,
			DBName:                dbname,
			SQLCommands:           GenerateRandomSQLCommands(20, appType),
			SQLCommandsPostBackup: GenerateRandomSQLCommands(20, appType),
			NodePort:              nodePort,
		}, nil

	}
}
