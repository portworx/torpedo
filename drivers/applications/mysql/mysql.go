package applications

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	. "github.com/portworx/torpedo/drivers/applications/apptypes"
	. "github.com/portworx/torpedo/drivers/utilities"
	"github.com/portworx/torpedo/pkg/log"
)

type MySqlConfig struct {
	Hostname    string
	User        string
	Password    string
	Port        int
	DBName      string
	SQLCommands map[string][]string
}

// GetConnection returns a connection object for mysql database
func (app *MySqlConfig) GetConnection(ctx context.Context) (*sql.DB, error) {

	if app.Port == 0 {
		app.Port = app.DefaultPort()
	}

	if app.DBName == "" {
		app.DBName = app.DefaultDBName()
	}

	url := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
		app.User, app.Password, app.Hostname, app.Port, app.DBName)

	conn, err := sql.Open("mysql", url)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %s, Conn String - [%s]", err, url)
	}

	return conn, nil
}

// DefaultPort returns default port for mysql
func (app *MySqlConfig) DefaultPort() int { return 3306 }

// DefaultDBName returns default database name
func (app *MySqlConfig) DefaultDBName() string { return "mysql" }

// ExecuteCommand executes a SQL command for mysql database
func (app *MySqlConfig) ExecuteCommand(commands []string, ctx context.Context) error {

	conn, err := app.GetConnection(ctx)
	if err != nil {
		return err
	}

	defer conn.Close()

	for _, eachCommand := range commands {
		_, err = conn.ExecContext(ctx, eachCommand)
		if err != nil {
			return err
		}
	}
	return nil
}

// InsertBackupData inserts the rows generated initially by utilities
func (app *MySqlConfig) InsertBackupData(ctx context.Context) error {

	log.Infof("Inserting data")
	err := app.ExecuteCommand(app.SQLCommands["insert"], ctx)

	return err
}

// CheckDataPresent checks if the mentioned entry is present or not in the database
func (app *MySqlConfig) CheckDataPresent(selectQueries []string, ctx context.Context) error {

	log.Infof("Running Select Queries")

	conn, err := app.GetConnection(ctx)
	if err != nil {
		return err
	}

	defer conn.Close()

	var key string
	var value string
	var queryNotFoundList []string

	for _, eachQuery := range selectQueries {
		currentRow := conn.QueryRowContext(ctx, eachQuery)
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
func (app *MySqlConfig) UpdateBackupData(ctx context.Context) error {

	log.Infof("Running Update Queries")
	err := app.ExecuteCommand(app.SQLCommands["update"], ctx)

	return err
}

// DeleteBackupData deletes the rows generated initially by utilities
func (app *MySqlConfig) DeleteBackupData(ctx context.Context) error {

	log.Infof("Running Delete Queries")
	err := app.ExecuteCommand(app.SQLCommands["delete"], ctx)

	return err
}

// StartData - Go routine to run parallal with app to keep injecting data every 2 seconds
func (app *MySqlConfig) StartData(command <-chan string, ctx context.Context) error {
	var status = "Start"
	var allSelectCommands []string
	var allErrors []string
	var tableName = "table_" + RandomString(4)

	createTableQuery := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
		`+"`key` "+`VARCHAR(45) NOT NULL,
		value VARCHAR(255)
	  )`, tableName)
	err := app.ExecuteCommand([]string{createTableQuery}, ctx)
	if err != nil {
		log.Infof("Error while creating table - [%s]", err.Error())
		allErrors = append(allErrors, err.Error())
	}
	for {
		select {
		case cmd := <-command:
			switch cmd {
			case "Stop":
				if len(allErrors) != 0 {
					return fmt.Errorf(strings.Join(allErrors, "\n"))
				}
				log.Infof("All select commands - [%s]", strings.Join(allSelectCommands, "\n"))
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
				log.Infof("Running insert command for app")
				if err != nil {
					allErrors = append(allErrors, err.Error())
				}
				allSelectCommands = append(allSelectCommands, commandPair["select"]...)
				time.Sleep(2 * time.Second)
			}
		}
	}
}

// startInsertingData is helper to insert generate rows and insert data parallely for mysql app
func (app *MySqlConfig) startInsertingData(tableName string, ctx context.Context) (map[string][]string, error) {

	commandPair := GenerateSQLCommandPair(tableName, MySql)

	err := app.ExecuteCommand(commandPair["insert"], ctx)
	if err != nil {
		return commandPair, err
	}

	return commandPair, nil
}