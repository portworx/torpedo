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

func (app *PostgresConfig) DefaultPort() int { return 5432 }

func (app *PostgresConfig) DefaultDBName() string { return "postgres" }

func (app *PostgresConfig) ExecuteCommand(commads []string, ctx context.Context) error {

	conn, err := app.GetConnection(ctx)
	if err != nil {
		return err
	}
	for _, eachCommand := range commads {
		_, err = conn.Exec(ctx, eachCommand)
		if err != nil {
			return err
		}
	}
	return nil
}

func (app *PostgresConfig) InsertBackupData(ctx context.Context) error {

	log.Infof("Inserting data")
	err := app.ExecuteCommand(app.SQLCommands["insert"], ctx)

	return err
}

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

func (app *PostgresConfig) UpdateBackupData(ctx context.Context) error {

	log.Infof("Running Update Queries")
	err := app.ExecuteCommand(app.SQLCommands["update"], ctx)

	return err
}

func (app *PostgresConfig) DeleteBackupData(ctx context.Context) error {

	log.Infof("Running Delete Queries")
	err := app.ExecuteCommand(app.SQLCommands["delete"], ctx)

	return err
}

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

func (app *PostgresConfig) startInsertingData(tableName string, ctx context.Context) (map[string][]string, error) {

	commandPair := GenerateSQLCommandPair(tableName, "postgres")

	err := app.ExecuteCommand(commandPair["insert"], ctx)
	if err != nil {
		return commandPair, err
	}

	return commandPair, nil
}
