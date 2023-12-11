package applications

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v4"
	. "github.com/portworx/torpedo/drivers/utilities"
	"github.com/portworx/torpedo/pkg/log"
)

type PostgresConfig struct {
	Hostname    string
	User        string
	Password    string
	Port        int
	DBName      string
	SQLCommands map[string][]string
}

// GetConnection returns a connection object for postgres database
func (app *PostgresConfig) GetConnection(ctx context.Context) (*pgx.Conn, error) {

	if app.Port == 0 {
		app.Port = app.DefaultPort()
	}

	if app.DBName == "" {
		app.DBName = app.DefaultDBName()
	}

	url := fmt.Sprintf("postgres://%s:%s@%s:%d/%s",
		app.User, app.Password, app.Hostname, app.Port, app.DBName)
	conn, err := pgx.Connect(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %s, Conn String - [%s]", err, url)
	}

	return conn, nil
}

// DefaultPort returns default port for postgres
func (app *PostgresConfig) DefaultPort() int { return 5432 }

// DefaultDBName returns default database name
func (app *PostgresConfig) DefaultDBName() string { return "postgres" }

// ExecuteCommand executes a SQL command for postgres database
func (app *PostgresConfig) ExecuteCommand(commands []string, ctx context.Context) error {

	conn, err := app.GetConnection(ctx)
	if err != nil {
		return err
	}
	for _, eachCommand := range commands {
		_, err = conn.Exec(ctx, eachCommand)
		if err != nil {
			return err
		}
	}
	return nil
}

// InsertBackupData inserts the rows generated initially by utilities
func (app *PostgresConfig) InsertBackupData(ctx context.Context) error {

	log.Infof("Inserting data")
	err := app.ExecuteCommand(app.SQLCommands["insert"], ctx)

	return err
}

// CheckDataPresent checks if the mentioned entry is present or not in the database
func (app *PostgresConfig) CheckDataPresent(selectQueries []string, ctx context.Context) error {

	log.Infof("Running Select Queries")

	conn, err := app.GetConnection(ctx)
	if err != nil {
		return err
	}

	var key string
	var value string
	var queryNotFoundList []string

	for _, eachQuery := range selectQueries {
		currentRow := conn.QueryRow(ctx, eachQuery)
		err := currentRow.Scan(&key, &value)

		if err != nil {
			log.Infof("Select query failed - [%s] Error - [%s]", eachQuery, err.Error())
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
func (app *PostgresConfig) UpdateBackupData(ctx context.Context) error {

	log.Infof("Running Update Queries")
	err := app.ExecuteCommand(app.SQLCommands["update"], ctx)

	return err
}

// DeleteBackupData deletes the rows generated initially by utilities
func (app *PostgresConfig) DeleteBackupData(ctx context.Context) error {

	log.Infof("Running Delete Queries")
	err := app.ExecuteCommand(app.SQLCommands["delete"], ctx)

	return err
}

// StartData - Go routine to run parallal with app to keep injecting data every 2 seconds
func (app *PostgresConfig) StartData(command <-chan string, ctx context.Context) error {
	var status = "Start"
	var allSelectCommands []string
	var tableName = "table_" + RandomString(4)

	createTableQuery := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
		key varchar(45) NOT NULL,
		value varchar(45) NOT NULL
	  )`, tableName)
	err := app.ExecuteCommand([]string{createTableQuery}, ctx)
	if err != nil {
		return err
	}
	for {
		select {
		case cmd := <-command:
			switch cmd {
			case "Stop":
				log.Infof("All select command [%v] for [%s]", allSelectCommands, app.Hostname)
				err := app.CheckDataPresent(allSelectCommands, ctx)
				return err

			case "Pause":
				status = "Pause"
			default:
				status = "Start"
			}
		default:
			if status == "Start" {
				commandPair, err := app.startInsertingData(tableName, ctx)
				if err != nil {
					return err
				}
				allSelectCommands = append(allSelectCommands, commandPair["select"]...)
				time.Sleep(2 * time.Second)
			}
		}
	}
}

// startInsertingData is helper to insert generate rows and insert data parallely for postgres app
func (app *PostgresConfig) startInsertingData(tableName string, ctx context.Context) (map[string][]string, error) {

	commandPair := GenerateSQLCommandPair(tableName, "postgres")

	err := app.ExecuteCommand(commandPair["insert"], ctx)
	if err != nil {
		return commandPair, err
	}

	return commandPair, nil
}
