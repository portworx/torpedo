package databases

import (
	"context"
	"fmt"
	"strings"

	"github.com/docker/docker/errdefs"
	neogo "github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/portworx/torpedo/pkg/log"
)

type Neo4jConfig struct {
	Hostname       string
	User           string
	Password       string
	Port           int
	NodePort       int
	DBName         string
	Namespace      string
	TLSEnabled     bool
	CypherCommands map[string]map[string][]string
}

// GetReplicaDetails gets the replica details for given deployment
func (app *Neo4jConfig) GetReplicaDetails(context.Context) (ReplicaDetails, error) {
	return ReplicaDetails{}, errdefs.NotImplemented(fmt.Errorf("GetReplicaDetails is not implemented for Neo4j"))
}

// AddDatabaseToCluster Adds database cluster
func (app *Neo4jConfig) AddDatabaseToCluster(ctx context.Context, databaseName string, clusterName string) error {
	return errdefs.NotImplemented(fmt.Errorf("AddDatabaseToCluster is not implemented for Neo4j"))
}

func (app *Neo4jConfig) DefaultPort() int { return 7687 }

func (app *Neo4jConfig) DefaultDBName() string { return "neo4j" }

func (app *Neo4jConfig) GetConnection(ctx context.Context) (neogo.DriverWithContext, error) {

	log.Debugf("Getting Connection")
	if app.Port == 0 {
		app.Port = app.DefaultPort()
	}

	if app.DBName == "" {
		app.DBName = app.DefaultDBName()
	}

	var neourl string
	if app.TLSEnabled {
		neourl = fmt.Sprintf("neo4j+ssc://%s:%d", app.Hostname, app.NodePort)
	} else {
		neourl = fmt.Sprintf("neo4j://%s:%d", app.Hostname, app.NodePort)
	}
	log.Debugf("Url-[%s]", neourl)
	log.Debugf("User-[%s]", app.User)
	log.Debugf("Password-[%s]", app.Password)

	driver, err := neogo.NewDriverWithContext(neourl, neogo.BasicAuth(app.User, app.Password, ""))
	if err != nil {
		return nil, err
	}

	return driver, nil
}

func (app *Neo4jConfig) ExecuteCommand(commands []string, ctx context.Context) ([]string, error) {
	var results []string
	log.Debugf("Executing command on database")

	driver, err := app.GetConnection(ctx)
	if err != nil {
		return results, err
	}
	defer driver.Close(ctx)

	// Create neo session
	session := driver.NewSession(ctx, neogo.SessionConfig{AccessMode: neogo.AccessModeWrite})
	defer session.Close(ctx)

	// Run Cypher Query
	for _, query := range commands {
		log.Debugf("Running Cypher Query...")
		resultWithContext, err := session.Run(ctx, query, nil)
		if err != nil {
			return results, fmt.Errorf("error while running the session %v", err)
		}

		records, err := resultWithContext.Collect(ctx)
		if err != nil {
			return results, fmt.Errorf("error while collecting result %v", err)
		}

		for _, record := range records {
			for _, value := range record.Values {
				// Convert the value to a string and append to resultStrings
				results = append(results, fmt.Sprintf("%v", value))
			}
		}
		log.Debugf("query result [%+v]", results[0])
	}
	return results, nil
}

func (app *Neo4jConfig) StartData(command <-chan string, ctx context.Context) error {
	return nil
}

func (app *Neo4jConfig) CheckDataPresent(selectQueries []string, ctx context.Context) error {
	return nil
}

// CopyBackupData copies the table and saves as csv file
func (app *Neo4jConfig) CopyBackupData(ctx context.Context, identifier string) error {
	return nil
}

func (app *Neo4jConfig) InsertBackupData(ctx context.Context, identifier string, commands []string) error {

	var err error
	log.InfoD("Inserting data")
	if len(commands) == 0 {
		_, err = app.ExecuteCommand(commands, ctx)
	} else {
		log.Infof("Inserting below data : %s", strings.Join(commands, "\n"))
		_, err = app.ExecuteCommand(commands, ctx)
	}

	return err
}

func (app *Neo4jConfig) GetBackupData(identifier string) []string {
	return nil
}

// Update the existing SQL commands
func (app *Neo4jConfig) UpdateDataCommands(count int, identifier string) {
	app.CypherCommands[identifier] = GenerateRandomCypherCommands(count, Neo4j)
	log.InfoD("SQL Commands updated")
}

// Update the existing SQL commands
func (app *Neo4jConfig) AddDataCommands(identifier string, commands map[string][]string) {
	app.CypherCommands[identifier] = commands
	log.InfoD("Sql commands added")
}

// Generate and return random SQL commands
func (app *Neo4jConfig) GetRandomDataCommands(count int) map[string][]string {
	return GenerateRandomCypherCommands(count, Neo4j)
}

// Get the application type
func (app *Neo4jConfig) GetApplicationType() string {
	return Neo4j
}

// Get Namespace of the app
func (app *Neo4jConfig) GetNamespace() string {
	return app.Namespace
}

// CreateNewDatabase creates a new database on the data service
func (app *Neo4jConfig) CreateNewDatabase(ctx context.Context, databaseName string, databaseDetails DefaultDatabase) error {
	return errdefs.NotImplemented(fmt.Errorf("CreateNewDatabase is not implemented for neo4j"))
}

// GetDefaultNewDatabaseConfig provides default inputs to create a new database
func (app *Neo4jConfig) GetDefaultNewDatabaseConfig(databaseName string) DefaultDatabase {
	log.Warnf("GetDefaultNewDatabaseConfig is not implemented for Neo4J")
	return DefaultDatabase{}
}

// ExecuteCommandOnPrimary executes a command on the primary replica
func (app *Neo4jConfig) ExecuteCommandOnPrimary(commands []string, ctx context.Context) ([]string, error) {
	return nil, errdefs.NotImplemented(fmt.Errorf("ExecuteCommandOnPrimary is not implemented for Neo4j"))
}

// Failover is not implemented for Neo4j
func (app *Neo4jConfig) Failover(ctx context.Context) error {
	return errdefs.NotImplemented(fmt.Errorf("Failover is not implemented for Neo4j"))
}

// GetDataserviceConfigurations returns the configuration of the dataservice
func (app *Neo4jConfig) GetDataserviceConfigurations(ctx context.Context) (DatabaseConfiguration, error) {
	return DatabaseConfiguration{}, errdefs.NotImplemented(fmt.Errorf("GetDataserviceConfigurations is not implemented for Neo4j"))
}

// GetDatabaseDetails returns the database details
func (app *Neo4jConfig) GetDatabaseDetails() Database {
	return Database{}
}
