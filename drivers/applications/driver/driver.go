package driver

import (
	"context"

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
}

func GetApplicationDriver(appType string, hostname string, user string,
	password string, port int, dbname string, ctx context.Context) (ApplicationDriver, error) {

	switch appType {
	case "postgres":
		return &PostgresConfig{
			Hostname:    hostname,
			User:        user,
			Password:    password,
			Port:        port,
			DBName:      dbname,
			SQLCommands: GenerateRandomSQLCommands(20),
		}, nil
	case "mysql":
		return &MySqlConfig{
			Hostname:    hostname,
			User:        user,
			Password:    password,
			Port:        port,
			DBName:      dbname,
			SQLCommands: GenerateRandomSQLCommands(20),
		}, nil
	default:
		return &PostgresConfig{
			Hostname:    hostname,
			User:        user,
			Password:    password,
			Port:        port,
			DBName:      dbname,
			SQLCommands: GenerateRandomSQLCommands(20),
		}, nil

	}
}
