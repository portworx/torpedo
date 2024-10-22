package tests

import (
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	"github.com/pure-px/torpedo/drivers/applications/databases"
	"github.com/pure-px/torpedo/drivers/applications/hammerdb"
	"github.com/pure-px/torpedo/drivers/unifiedPlatform/automationModels"
	pdslibs "github.com/pure-px/torpedo/drivers/unifiedPlatform/pdsLibs"
	"github.com/pure-px/torpedo/pkg/log"
	. "github.com/pure-px/torpedo/tests"
	. "github.com/pure-px/torpedo/tests/unifiedPlatform"
	"strconv"
	"time"
)

var _ = Describe("{BackupRestoreSQLAAGSynchronousCommitModeUnPlanned}", func() {
	var (
		deployment          *automationModels.PDSDeploymentResponse
		err                 error
		oldPrimary          string
		replicaDetails      databases.ReplicaDetails
		databaseDetails     pdslibs.DataServiceDetails
		aagDatabaseName     string
		latestBackupUid     string
		pdsBackupConfigName string
		restoreName         string
		dataForDatabase     map[string][]string
		restoreDriver       databases.DatabaseDriver
	)

	JustBeforeEach(func() {
		StartPDSTorpedoTest("BackupRestoreSQLAAGSynchronousCommitModeUnPlanned", "Test fail over, Backup and Restore in MSSQL AAG synchronous commit mode with un-planned fail over", nil, 0)
		aagDatabaseName = "aag_database_" + RandomString(5)
	})

	It("Test fail over, Backup and Restore in MSSQL AAG synchronous commit mode with unplanned fail over", func() {
		for _, ds := range NewPdsParams.DataServiceToTest {
			Step("Deploy MSSQL AAG DataService", func() {
				deployment, err = WorkflowDataService.DeployDataService(ds, ds.Image, ds.Version, PDS_DEFAULT_NAMESPACE)
				log.FailOnError(err, "Error while deploying ds")
				log.Debugf("Source Deployment Id: [%s]", *deployment.Create.Meta.Uid)
				databaseDetails = *WorkflowDataService.DataServiceDeployment[*deployment.Create.Meta.Uid]
			})

			Step("Create new database", func() {
				newDbConfig := databaseDetails.DatabaseDriver.GetDefaultNewDatabaseConfig(aagDatabaseName)
				err := databaseDetails.DatabaseDriver.CreateNewDatabase(ctx, aagDatabaseName, newDbConfig)
				dash.VerifyFatal(err, nil, "create new database")
			})

			Step("Change Availability mode to Full and take backup", func() {
				_, err := databaseDetails.DatabaseDriver.ExecuteCommandOnPrimary([]string{fmt.Sprintf("ALTER DATABASE [%s] SET RECOVERY FULL", aagDatabaseName)}, ctx)
				log.FailOnError(err, "unable to set recovery mode to full")
				_, err = databaseDetails.DatabaseDriver.ExecuteCommandOnPrimary([]string{fmt.Sprintf("BACKUP DATABASE [%s] TO DISK = 'NUL' WITH NOFORMAT, NOINIT, NAME = '%s full backup', SKIP, NOREWIND, NOUNLOAD,  STATS = 10", aagDatabaseName, aagDatabaseName)}, ctx)
				log.FailOnError(err, "unable to take backup")
			})

			Step("Get Replica Details", func() {
				replicaDetails, err = databaseDetails.DatabaseDriver.GetReplicaDetails(ctx)
				log.FailOnError(err, "unable to get replica details")
				log.Infof("AG Name - [%s], AG Id - [%s], Total Replicas [%s]", replicaDetails.AGName, replicaDetails.AGId, len(replicaDetails.Replicas))

				for _, replica := range replicaDetails.Replicas {
					log.Infof("============================================")
					log.Infof("Replica Name - [%s]", replica.Name)
					log.Infof("Replica Id - [%s]", replica.ReplicaId)
					log.Infof("Replica Role - [%s]", replica.Role)
					log.Infof("Replica Failover Mode - [%s]", replica.FailoverMode)
					log.Infof("Replica Availability Mode - [%s]", replica.AvailabilityMode)
					log.Infof("Replica Connection State - [%s]", replica.ConnectionState)
					log.Infof("Replica Health - [%s]", replica.Health)
					log.Infof("============================================")
				}
			})

			Step("Add database to AAG", func() {
				err := databaseDetails.DatabaseDriver.AddDatabaseToCluster(ctx, aagDatabaseName, replicaDetails.AGName)
				log.FailOnError(err, "Error occurred while adding database")
			})

			Step("Add data to database", func() {
				dataForDatabase = databaseDetails.DatabaseDriver.GetRandomDataCommands(50)
				_, err := databaseDetails.DatabaseDriver.ExecuteCommandOnPrimary(dataForDatabase["insert"], ctx)
				log.FailOnError(err, "unable to insert data into database")
			})

			Step("Verify AAG", func() {
				_, err := WorkflowDataService.ValidateSQLAAGReplicas(*deployment.Create.Meta.Uid)
				log.FailOnError(err, "replica health validation failed")
			})

			Step("Create Adhoc backup config of the existing deployment", func() {
				pdsBackupConfigName = "pds-adhoc-backup-" + RandomString(5)
				bkpConfigResponse, err := WorkflowPDSBackupConfig.CreateBackupConfig(pdsBackupConfigName, *deployment.Create.Meta.Uid)
				log.FailOnError(err, "Error occured while creating backupConfig")
				log.Infof("BackupConfigName: [%s], BackupConfigId: [%s]", *bkpConfigResponse.Create.Meta.Name, *bkpConfigResponse.Create.Meta.Uid)
				log.Infof("All deployments - [%+v]", WorkflowDataService.DataServiceDeployment)
			})

			Step("Get the latest backup detail for the deployment", func() {
				backupResponse, err := WorkflowPDSBackup.GetLatestBackup(*deployment.Create.Meta.Uid)
				log.FailOnError(err, "Error occurred while creating backup")
				latestBackupUid = *backupResponse.Meta.Uid
				log.Infof("Latest backup ID [%s], Name [%s]", *backupResponse.Meta.Uid, *backupResponse.Meta.Name)
				err = WorkflowPDSBackup.WaitForBackupToComplete(*backupResponse.Meta.Uid)
				log.FailOnError(err, "error occurred while waiting for backup to complete")
			})

			Step("Fail over AAG replicas by killing primary pod and validate", func() {
				WorkflowDataService.DeletePDSPods([]string{replicaDetails.PrimaryReplica}, PDS_DEFAULT_NAMESPACE)
				log.FailOnError(err, "unable to fail over replica manually")
				log.InfoD("Primary pod deleted successfully")
				err := WorkflowDataService.WaitForDeploymentToBeAvailable(*deployment.Create.Meta.Uid)
				log.FailOnError(err, "Deployment didn't come back online after pods restart")
				replicaDetails, err = databaseDetails.DatabaseDriver.GetReplicaDetails(ctx)
				log.FailOnError(err, "unable to get replica details after fail over")
				log.InfoD("Current primary replica after fail over - [%s]", replicaDetails.PrimaryReplica)
				dash.VerifyFatal(replicaDetails.PrimaryReplica != oldPrimary, true, "Validating primary replica fail over")
				_, err = WorkflowDataService.ValidateSQLAAGReplicas(*deployment.Create.Meta.Uid)
				log.FailOnError(err, "replica health validation failed post failover")
			})

			Step("Create Restore from the latest backup Id", func() {
				CheckforClusterSwitch()
				restoreName = "restore-" + RandomString(5)
				_, err := WorkflowPDSRestore.CreateRestore(restoreName, latestBackupUid, restoreName, *deployment.Create.Meta.Uid)
				log.FailOnError(err, "Restore Failed")
				log.Infof("All restores - [%+v]", WorkflowPDSRestore.Restores)
				log.Infof("Restore Created Name - [%s], UID - [%s]", *WorkflowPDSRestore.Restores[restoreName].Meta.Name, *WorkflowPDSRestore.Restores[restoreName].Meta.Uid)
			})

			Step("Verify AAG deployment post restore", func() {
				defer func() {
					err := SetSourceKubeConfig()
					log.FailOnError(err, "failed to switch context to source cluster")
				}()
				_, err := WorkflowPDSRestore.RestoredDeployments.ValidateSQLAAGReplicas(WorkflowPDSRestore.Restores[restoreName].Config.DestinationReferences.DataServiceDeploymentId)
				log.FailOnError(err, "replica health validation failed post restore")
				restoreDriver = WorkflowPDSRestore.RestoredDeployments.DataServiceDeployment[WorkflowPDSRestore.Restores[restoreName].Config.DestinationReferences.DataServiceDeploymentId].DatabaseDriver
			})

			Step("Validate Data After Restore", func() {
				_, err := restoreDriver.ExecuteCommand(dataForDatabase["select"], ctx)
				log.FailOnError(err, "unable to select data from database - Data validation failed")
			})

		}
	})

	JustAfterEach(func() {
		defer EndPDSTorpedoTest()
	})
})

var _ = Describe("{BackupRestoreSQLAAGSynchronousCommitModePlanned}", func() {
	var (
		deployment          *automationModels.PDSDeploymentResponse
		err                 error
		oldPrimary          string
		replicaDetails      databases.ReplicaDetails
		databaseDetails     pdslibs.DataServiceDetails
		aagDatabaseName     string
		latestBackupUid     string
		pdsBackupConfigName string
		restoreName         string
		dataForDatabase     map[string][]string
		restoreDriver       databases.DatabaseDriver
	)

	JustBeforeEach(func() {
		StartPDSTorpedoTest("BackupRestoreSQLAAGSynchronousCommitModePlanned", "Test fail over, Backup and Restore in MSSQL AAG synchronous commit mode with planned fail over", nil, 0)
		aagDatabaseName = "aag_database_" + RandomString(5)
	})

	It("Test fail over, Backup and Restore in MSSQL AAG synchronous commit mode with planned fail over", func() {
		for _, ds := range NewPdsParams.DataServiceToTest {
			Step("Deploy MSSQL AAG DataService", func() {
				deployment, err = WorkflowDataService.DeployDataService(ds, ds.Image, ds.Version, PDS_DEFAULT_NAMESPACE)
				log.FailOnError(err, "Error while deploying ds")
				log.Debugf("Source Deployment Id: [%s]", *deployment.Create.Meta.Uid)
				databaseDetails = *WorkflowDataService.DataServiceDeployment[*deployment.Create.Meta.Uid]
			})

			Step("Create new database", func() {
				newDbConfig := databaseDetails.DatabaseDriver.GetDefaultNewDatabaseConfig(aagDatabaseName)
				err := databaseDetails.DatabaseDriver.CreateNewDatabase(ctx, aagDatabaseName, newDbConfig)
				dash.VerifyFatal(err, nil, "create new database")
			})

			Step("Change Availability mode to Full and take backup", func() {
				_, err := databaseDetails.DatabaseDriver.ExecuteCommandOnPrimary([]string{fmt.Sprintf("ALTER DATABASE [%s] SET RECOVERY FULL", aagDatabaseName)}, ctx)
				log.FailOnError(err, "unable to set recovery mode to full")
				_, err = databaseDetails.DatabaseDriver.ExecuteCommandOnPrimary([]string{fmt.Sprintf("BACKUP DATABASE [%s] TO DISK = 'NUL' WITH NOFORMAT, NOINIT, NAME = '%s full backup', SKIP, NOREWIND, NOUNLOAD,  STATS = 10", aagDatabaseName, aagDatabaseName)}, ctx)
				log.FailOnError(err, "unable to take backup")
			})

			Step("Get Replica Details", func() {
				replicaDetails, err = databaseDetails.DatabaseDriver.GetReplicaDetails(ctx)
				log.FailOnError(err, "unable to get replica details")
				log.Infof("AG Name - [%s], AG Id - [%s], Total Replicas [%s]", replicaDetails.AGName, replicaDetails.AGId, len(replicaDetails.Replicas))

				for _, replica := range replicaDetails.Replicas {
					log.Infof("============================================")
					log.Infof("Replica Name - [%s]", replica.Name)
					log.Infof("Replica Id - [%s]", replica.ReplicaId)
					log.Infof("Replica Role - [%s]", replica.Role)
					log.Infof("Replica Failover Mode - [%s]", replica.FailoverMode)
					log.Infof("Replica Availability Mode - [%s]", replica.AvailabilityMode)
					log.Infof("Replica Connection State - [%s]", replica.ConnectionState)
					log.Infof("Replica Health - [%s]", replica.Health)
					log.Infof("============================================")
				}
			})

			Step("Add database to AAG", func() {
				err := databaseDetails.DatabaseDriver.AddDatabaseToCluster(ctx, aagDatabaseName, replicaDetails.AGName)
				log.FailOnError(err, "Error occurred while adding database")
			})

			Step("Add data to database", func() {
				dataForDatabase = databaseDetails.DatabaseDriver.GetRandomDataCommands(50)
				_, err := databaseDetails.DatabaseDriver.ExecuteCommandOnPrimary(dataForDatabase["insert"], ctx)
				log.FailOnError(err, "unable to insert data into database")
			})

			Step("Verify AAG", func() {
				_, err := WorkflowDataService.ValidateSQLAAGReplicas(*deployment.Create.Meta.Uid)
				log.FailOnError(err, "replica health validation failed")
			})

			Step("Create Adhoc backup config of the existing deployment", func() {
				pdsBackupConfigName = "pds-adhoc-backup-" + RandomString(5)
				bkpConfigResponse, err := WorkflowPDSBackupConfig.CreateBackupConfig(pdsBackupConfigName, *deployment.Create.Meta.Uid)
				log.FailOnError(err, "Error occured while creating backupConfig")
				log.Infof("BackupConfigName: [%s], BackupConfigId: [%s]", *bkpConfigResponse.Create.Meta.Name, *bkpConfigResponse.Create.Meta.Uid)
				log.Infof("All deployments - [%+v]", WorkflowDataService.DataServiceDeployment)
			})

			Step("Get the latest backup detail for the deployment", func() {
				backupResponse, err := WorkflowPDSBackup.GetLatestBackup(*deployment.Create.Meta.Uid)
				log.FailOnError(err, "Error occurred while creating backup")
				latestBackupUid = *backupResponse.Meta.Uid
				log.Infof("Latest backup ID [%s], Name [%s]", *backupResponse.Meta.Uid, *backupResponse.Meta.Name)
				err = WorkflowPDSBackup.WaitForBackupToComplete(*backupResponse.Meta.Uid)
				log.FailOnError(err, "error occurred while waiting for backup to complete")
			})

			Step("Fail over AAG replicas with ag-monitor command", func() {
				err := databaseDetails.DatabaseDriver.Failover(ctx)
				log.FailOnError(err, "unable to fail over replica manually")
				err = WorkflowDataService.WaitForDeploymentToBeAvailable(*deployment.Create.Meta.Uid)
				log.FailOnError(err, "Deployment didn't come back online after pods restart")
				replicaDetails, err = databaseDetails.DatabaseDriver.GetReplicaDetails(ctx)
				log.FailOnError(err, "unable to get replica details after fail over")
				log.InfoD("Current primary replica after fail over - [%s]", replicaDetails.PrimaryReplica)
				dash.VerifyFatal(replicaDetails.PrimaryReplica != oldPrimary, true, "Validating primary replica fail over")
				_, err = WorkflowDataService.ValidateSQLAAGReplicas(*deployment.Create.Meta.Uid)
				log.FailOnError(err, "replica health validation failed post failover")
			})

			Step("Create Restore from the latest backup Id", func() {

				CheckforClusterSwitch()
				restoreName = "restore-" + RandomString(5)
				_, err := WorkflowPDSRestore.CreateRestore(restoreName, latestBackupUid, restoreName, *deployment.Create.Meta.Uid)
				log.FailOnError(err, "Restore Failed")
				log.Infof("All restores - [%+v]", WorkflowPDSRestore.Restores)
				log.Infof("Restore Created Name - [%s], UID - [%s]", *WorkflowPDSRestore.Restores[restoreName].Meta.Name, *WorkflowPDSRestore.Restores[restoreName].Meta.Uid)
			})

			Step("Verify AAG deployment post restore", func() {
				defer func() {
					err := SetSourceKubeConfig()
					log.FailOnError(err, "failed to switch context to source cluster")
				}()
				_, err := WorkflowPDSRestore.RestoredDeployments.ValidateSQLAAGReplicas(WorkflowPDSRestore.Restores[restoreName].Config.DestinationReferences.DataServiceDeploymentId)
				log.FailOnError(err, "replica health validation failed post restore")
				restoreDriver = WorkflowPDSRestore.RestoredDeployments.DataServiceDeployment[WorkflowPDSRestore.Restores[restoreName].Config.DestinationReferences.DataServiceDeploymentId].DatabaseDriver
			})

			Step("Validate Data After Restore", func() {
				_, err := restoreDriver.ExecuteCommand(dataForDatabase["select"], ctx)
				log.FailOnError(err, "unable to select data from database - Data validation failed")
			})

		}
	})

	JustAfterEach(func() {
		defer EndPDSTorpedoTest()
	})
})

var _ = Describe("{MSSQLRecoveryModeTest}", func() {
	var (
		deployment                  *automationModels.PDSDeploymentResponse
		err                         error
		databaseDetails             pdslibs.DataServiceDetails
		dsConfig                    databases.DatabaseConfiguration
		fullRecoveryDatabaseName    string
		defaultRecoveryDatabaseName string
		latestBackupUid             string
		pdsBackupConfigName         string
		restoreName                 string
	)

	JustBeforeEach(func() {
		StartPDSTorpedoTest("MSSQLRecoveryModeTest", "Tests multiple database in different recovery modes", nil, 0)
		fullRecoveryDatabaseName = "full_recover_database" + RandomString(5)
		defaultRecoveryDatabaseName = "default_database" + RandomString(5)
	})

	It("Tests multiple database in different recovery modes", func() {
		for _, ds := range NewPdsParams.DataServiceToTest {
			Step("Deploy MSSQL DataService", func() {
				deployment, err = WorkflowDataService.DeployDataService(ds, ds.Image, ds.Version, PDS_DEFAULT_NAMESPACE)
				log.FailOnError(err, "Error while deploying ds")
				log.Debugf("Source Deployment Id: [%s]", *deployment.Create.Meta.Uid)
				databaseDetails = *WorkflowDataService.DataServiceDeployment[*deployment.Create.Meta.Uid]
			})

			Step("Create new databases", func() {
				newDbConfig := databaseDetails.DatabaseDriver.GetDefaultNewDatabaseConfig(fullRecoveryDatabaseName)
				err := databaseDetails.DatabaseDriver.CreateNewDatabase(ctx, fullRecoveryDatabaseName, newDbConfig)
				dash.VerifyFatal(err, nil, "create new database for full recovery")
				newDbConfig = databaseDetails.DatabaseDriver.GetDefaultNewDatabaseConfig(defaultRecoveryDatabaseName)
				err = databaseDetails.DatabaseDriver.CreateNewDatabase(ctx, defaultRecoveryDatabaseName, newDbConfig)
				dash.VerifyFatal(err, nil, "create new database for default mode")
			})

			Step("Change Availability mode to Full and take backup", func() {
				_, err := databaseDetails.DatabaseDriver.ExecuteCommandOnPrimary([]string{fmt.Sprintf("ALTER DATABASE [%s] SET RECOVERY FULL", fullRecoveryDatabaseName)}, ctx)
				log.FailOnError(err, "unable to set recovery mode to full")
				_, err = databaseDetails.DatabaseDriver.ExecuteCommandOnPrimary([]string{fmt.Sprintf("BACKUP DATABASE [%s] TO DISK = 'NUL' WITH NOFORMAT, NOINIT, NAME = '%s full backup', SKIP, NOREWIND, NOUNLOAD,  STATS = 10", fullRecoveryDatabaseName, fullRecoveryDatabaseName)}, ctx)
				log.FailOnError(err, "unable to take backup")
			})

			Step("Check Recovery Mode of Databases", func() {
				dsConfig, err = databaseDetails.DatabaseDriver.GetDataserviceConfigurations(ctx)
				log.FailOnError(err, "unable to get dataservice configurations")
				log.Infof("Dataservice Configurations (Recovery) - [%+v]", dsConfig.DatabaseRecoveryModes)
			})

			Step("Create Adhoc backup config of the existing deployment", func() {
				pdsBackupConfigName = "pds-adhoc-backup-" + RandomString(5)
				bkpConfigResponse, err := WorkflowPDSBackupConfig.CreateBackupConfig(pdsBackupConfigName, *deployment.Create.Meta.Uid)
				log.FailOnError(err, "Error occured while creating backupConfig")
				log.Infof("BackupConfigName: [%s], BackupConfigId: [%s]", *bkpConfigResponse.Create.Meta.Name, *bkpConfigResponse.Create.Meta.Uid)
				log.Infof("All deployments - [%+v]", WorkflowDataService.DataServiceDeployment)
			})

			Step("Get the latest backup detail for the deployment", func() {
				backupResponse, err := WorkflowPDSBackup.GetLatestBackup(*deployment.Create.Meta.Uid)
				log.FailOnError(err, "Error occurred while creating backup")
				latestBackupUid = *backupResponse.Meta.Uid
				log.Infof("Latest backup ID [%s], Name [%s]", *backupResponse.Meta.Uid, *backupResponse.Meta.Name)
				err = WorkflowPDSBackup.WaitForBackupToComplete(*backupResponse.Meta.Uid)
				log.FailOnError(err, "error occurred while waiting for backup to complete")
			})

			Step("Create Restore from the latest backup Id", func() {
				CheckforClusterSwitch()
				restoreName = "restore-" + RandomString(5)
				_, err := WorkflowPDSRestore.CreateRestore(restoreName, latestBackupUid, restoreName, *deployment.Create.Meta.Uid)
				log.FailOnError(err, "Restore Failed")
				log.Infof("All restores - [%+v]", WorkflowPDSRestore.Restores)
				log.Infof("Restore Created Name - [%s], UID - [%s]", *WorkflowPDSRestore.Restores[restoreName].Meta.Name, *WorkflowPDSRestore.Restores[restoreName].Meta.Uid)
			})

			Step("Check Recovery Mode of Databases", func() {
				defer func() {
					err := SetSourceKubeConfig()
					log.FailOnError(err, "failed to switch context to source cluster")
				}()

				restoreDsConfig, err := WorkflowPDSRestore.RestoredDeployments.DataServiceDeployment[WorkflowPDSRestore.Restores[restoreName].Config.DestinationReferences.DataServiceDeploymentId].DatabaseDriver.GetDataserviceConfigurations(ctx)
				log.FailOnError(err, "unable to get dataservice configurations")
				log.Infof("Restore Configurations (Recovery) - [%+v]", dsConfig.DatabaseRecoveryModes)
				for key, value := range dsConfig.DatabaseRecoveryModes {
					dash.VerifyFatal(value, restoreDsConfig.DatabaseRecoveryModes[key], fmt.Sprintf("Recovery mode of database [%s]", key))
				}

			})

		}
	})

	JustAfterEach(func() {
		defer EndPDSTorpedoTest()
	})
})

var _ = Describe("{SQLAAGStorageSizeIncrease}", func() {
	var (
		deployment                *automationModels.PDSDeploymentResponse
		err                       error
		replicaDetails            databases.ReplicaDetails
		databaseDetails           pdslibs.DataServiceDetails
		aagDatabaseName           string
		latestBackupUid           string
		pdsBackupConfigName       string
		restoreName               string
		restoreDriver             databases.DatabaseDriver
		allSelectQueries          []string
		baseStorageSizeForUpgrade int
	)

	JustBeforeEach(func() {
		StartPDSTorpedoTest("SQLAAGStorageSizeIncrease", "Increase storage size for SQL AAG Multiple times", nil, 0)
		aagDatabaseName = "aag_database_" + RandomString(5)
		allSelectQueries = make([]string, 0)
		baseStorageSizeForUpgrade = NewPdsParams.StorageConfigurationsSSIE.BaseStorageInGBForUpgrade
	})

	It("Increase storage size for SQL AAG Multiple times", func() {
		for _, ds := range NewPdsParams.DataServiceToTest {
			Step("Deploy MSSQL AAG DataService", func() {
				deployment, err = WorkflowDataService.DeployDataService(ds, ds.Image, ds.Version, PDS_DEFAULT_NAMESPACE)
				log.FailOnError(err, "Error while deploying ds")
				log.Debugf("Source Deployment Id: [%s]", *deployment.Create.Meta.Uid)
				databaseDetails = *WorkflowDataService.DataServiceDeployment[*deployment.Create.Meta.Uid]
			})

			Step("Create new database", func() {
				newDbConfig := databaseDetails.DatabaseDriver.GetDefaultNewDatabaseConfig(aagDatabaseName)
				err := databaseDetails.DatabaseDriver.CreateNewDatabase(ctx, aagDatabaseName, newDbConfig)
				dash.VerifyFatal(err, nil, "create new database")
			})

			Step("Change Availability mode to Full and take backup", func() {
				_, err := databaseDetails.DatabaseDriver.ExecuteCommandOnPrimary([]string{fmt.Sprintf("ALTER DATABASE [%s] SET RECOVERY FULL", aagDatabaseName)}, ctx)
				log.FailOnError(err, "unable to set recovery mode to full")
				_, err = databaseDetails.DatabaseDriver.ExecuteCommandOnPrimary([]string{fmt.Sprintf("BACKUP DATABASE [%s] TO DISK = 'NUL' WITH NOFORMAT, NOINIT, NAME = '%s full backup', SKIP, NOREWIND, NOUNLOAD,  STATS = 10", aagDatabaseName, aagDatabaseName)}, ctx)
				log.FailOnError(err, "unable to take backup")
			})

			Step("Get Replica Details", func() {
				replicaDetails, err = databaseDetails.DatabaseDriver.GetReplicaDetails(ctx)
				log.FailOnError(err, "unable to get replica details")
				log.Infof("AG Name - [%s], AG Id - [%s], Total Replicas [%s]", replicaDetails.AGName, replicaDetails.AGId, len(replicaDetails.Replicas))

				for _, replica := range replicaDetails.Replicas {
					log.Infof("============================================")
					log.Infof("Replica Name - [%s]", replica.Name)
					log.Infof("Replica Id - [%s]", replica.ReplicaId)
					log.Infof("Replica Role - [%s]", replica.Role)
					log.Infof("Replica Failover Mode - [%s]", replica.FailoverMode)
					log.Infof("Replica Availability Mode - [%s]", replica.AvailabilityMode)
					log.Infof("Replica Connection State - [%s]", replica.ConnectionState)
					log.Infof("Replica Health - [%s]", replica.Health)
					log.Infof("============================================")
				}
			})

			Step("Add database to AAG", func() {
				err := databaseDetails.DatabaseDriver.AddDatabaseToCluster(ctx, aagDatabaseName, replicaDetails.AGName)
				log.FailOnError(err, "Error occurred while adding database")
			})

			for i := 1; i <= NewPdsParams.StorageConfigurationsSSIE.Iterations; i++ {
				Step("Extend storage size of the AAG database", func() {

					log.InfoD("Extend storage size of the AAG database - Iteration - [%d]", i)

					baseStorageSizeForUpgrade += NewPdsParams.StorageConfigurationsSSIE.StorageIncreaseStepInGB

					NewPdsParams.ResourceConfiguration.New_Storage_Request = strconv.Itoa(baseStorageSizeForUpgrade) + "G"
					log.Debugf("Updating the Storage to: [%s]", NewPdsParams.ResourceConfiguration.New_Storage_Request)

					resConfigIdUpdated, err := WorkflowPDSTemplate.CreateResourceTemplateWithCustomValue(NewPdsParams)
					log.FailOnError(err, "Unable to create Custom Templates for PDS")
					log.InfoD("Updated Resource Template ID- [updated- %v]", resConfigIdUpdated)
					log.Infof("Associate newly created template to the project")
					err = WorkflowProject.Associate(
						[]string{},
						[]string{},
						[]string{},
						[]string{},
						[]string{resConfigIdUpdated},
						[]string{},
					)
					log.FailOnError(err, "Unable to associate Templates to Project")
					log.Infof("Associated Resources - [%+v]", WorkflowProject.AssociatedResources)

					beforeResizePodAge, err := WorkflowDataService.GetPodAgeForDeployment(*deployment.Create.Meta.Uid)
					log.FailOnError(err, "unable to get pods AGE before Storage resize")
					log.InfoD("Pods Age before storage resize is- [%v]Min", beforeResizePodAge)

					//sleeping 30sec before the next update
					time.Sleep(30 * time.Second)

					WorkflowDataService.UpdateDeploymentTemplates = true
					WorkflowDataService.PDSTemplates = WorkflowPDSTemplate
					_, err = WorkflowDataService.UpdateDataService(ds, *deployment.Create.Meta.Uid, ds.Image, ds.Version)
					log.FailOnError(err, "Error while updating ds")

					afterResizePodAge, err := WorkflowDataService.GetPodAgeForDeployment(*deployment.Create.Meta.Uid)
					log.FailOnError(err, "unable to get pods AGE before Storage resize")
					log.InfoD("Pods Age after storage resize is- [%v]Min", afterResizePodAge)

					dash.VerifyFatal(afterResizePodAge > beforeResizePodAge, true, "Validating if the pod restarted after storage size increase")
				})

				Step("Add data to database", func() {
					dataForDatabase := databaseDetails.DatabaseDriver.GetRandomDataCommands(50)
					_, err := databaseDetails.DatabaseDriver.ExecuteCommandOnPrimary(dataForDatabase["insert"], ctx)
					log.FailOnError(err, "unable to insert data into database")
					allSelectQueries = append(allSelectQueries, dataForDatabase["select"]...)

				})

				Step("Verify AAG", func() {
					_, err := WorkflowDataService.ValidateSQLAAGReplicas(*deployment.Create.Meta.Uid)
					log.FailOnError(err, "replica health validation failed")
				})

			}

			Step("Create Adhoc backup config of the existing deployment", func() {
				pdsBackupConfigName = "pds-adhoc-backup-" + RandomString(5)
				bkpConfigResponse, err := WorkflowPDSBackupConfig.CreateBackupConfig(pdsBackupConfigName, *deployment.Create.Meta.Uid)
				log.FailOnError(err, "Error occured while creating backupConfig")
				log.Infof("BackupConfigName: [%s], BackupConfigId: [%s]", *bkpConfigResponse.Create.Meta.Name, *bkpConfigResponse.Create.Meta.Uid)
				log.Infof("All deployments - [%+v]", WorkflowDataService.DataServiceDeployment)
			})

			Step("Get the latest backup detail for the deployment", func() {
				backupResponse, err := WorkflowPDSBackup.GetLatestBackup(*deployment.Create.Meta.Uid)
				log.FailOnError(err, "Error occurred while creating backup")
				latestBackupUid = *backupResponse.Meta.Uid
				log.Infof("Latest backup ID [%s], Name [%s]", *backupResponse.Meta.Uid, *backupResponse.Meta.Name)
				err = WorkflowPDSBackup.WaitForBackupToComplete(*backupResponse.Meta.Uid)
				log.FailOnError(err, "error occurred while waiting for backup to complete")
			})

			Step("Create Restore from the latest backup Id", func() {

				CheckforClusterSwitch()
				restoreName = "restore-" + RandomString(5)
				_, err := WorkflowPDSRestore.CreateRestore(restoreName, latestBackupUid, restoreName, *deployment.Create.Meta.Uid)
				log.FailOnError(err, "Restore Failed")
				log.Infof("All restores - [%+v]", WorkflowPDSRestore.Restores)
				log.Infof("Restore Created Name - [%s], UID - [%s]", *WorkflowPDSRestore.Restores[restoreName].Meta.Name, *WorkflowPDSRestore.Restores[restoreName].Meta.Uid)
			})

			Step("Verify AAG deployment post restore", func() {
				_, err := WorkflowPDSRestore.RestoredDeployments.ValidateSQLAAGReplicas(WorkflowPDSRestore.Restores[restoreName].Config.DestinationReferences.DataServiceDeploymentId)
				restoreDriver = WorkflowPDSRestore.RestoredDeployments.DataServiceDeployment[WorkflowPDSRestore.Restores[restoreName].Config.DestinationReferences.DataServiceDeploymentId].DatabaseDriver
				log.FailOnError(err, "replica health validation failed post restore")
			})

			Step("Validate Data After Restore", func() {
				defer func() {
					err := SetSourceKubeConfig()
					log.FailOnError(err, "failed to switch context to source cluster")
				}()
				_, err := restoreDriver.ExecuteCommand(allSelectQueries, ctx)
				log.FailOnError(err, "unable to select data from database - Data validation failed")
			})

		}
	})

	JustAfterEach(func() {
		defer EndPDSTorpedoTest()
	})
})

var _ = Describe("{MSSQLAAGVerticalScaleTest}", func() {
	var (
		deployment          *automationModels.PDSDeploymentResponse
		err                 error
		replicaDetails      databases.ReplicaDetails
		databaseDetails     pdslibs.DataServiceDetails
		aagDatabaseName     string
		latestBackupUid     string
		pdsBackupConfigName string
		restoreName         string
		hammerDBDriver      *hammerdb.HammerDB
	)

	JustBeforeEach(func() {
		StartPDSTorpedoTest("MSSQLAAGVerticalScaleTest", "Create AAG database and fill upto scale factor using hammer DB", nil, 0)
		aagDatabaseName = "aag_database_" + RandomString(5)
	})

	It("Create AAG database and fill upto scale factor using hammer DB", func() {
		for _, ds := range NewPdsParams.DataServiceToTest {
			Step("Deploy MSSQL AAG DataService", func() {
				deployment, err = WorkflowDataService.DeployDataService(ds, ds.Image, ds.Version, PDS_DEFAULT_NAMESPACE)
				log.FailOnError(err, "Error while deploying ds")
				log.Debugf("Source Deployment Id: [%s]", *deployment.Create.Meta.Uid)
				databaseDetails = *WorkflowDataService.DataServiceDeployment[*deployment.Create.Meta.Uid]
			})

			Step("Create new database", func() {
				newDbConfig := databaseDetails.DatabaseDriver.GetDefaultNewDatabaseConfig(aagDatabaseName)
				err := databaseDetails.DatabaseDriver.CreateNewDatabase(ctx, aagDatabaseName, newDbConfig)
				dash.VerifyFatal(err, nil, "create new database")
			})

			Step("Change Availability mode to Full and take backup", func() {
				_, err := databaseDetails.DatabaseDriver.ExecuteCommandOnPrimary([]string{fmt.Sprintf("ALTER DATABASE [%s] SET RECOVERY FULL", aagDatabaseName)}, ctx)
				log.FailOnError(err, "unable to set recovery mode to full")
				_, err = databaseDetails.DatabaseDriver.ExecuteCommandOnPrimary([]string{fmt.Sprintf("BACKUP DATABASE [%s] TO DISK = 'NUL' WITH NOFORMAT, NOINIT, NAME = '%s full backup', SKIP, NOREWIND, NOUNLOAD,  STATS = 10", aagDatabaseName, aagDatabaseName)}, ctx)
				log.FailOnError(err, "unable to take backup")
			})

			Step("Get Replica Details", func() {
				replicaDetails, err = databaseDetails.DatabaseDriver.GetReplicaDetails(ctx)
				log.FailOnError(err, "unable to get replica details")
				log.Infof("AG Name - [%s], AG Id - [%s], Total Replicas [%s]", replicaDetails.AGName, replicaDetails.AGId, len(replicaDetails.Replicas))

				for _, replica := range replicaDetails.Replicas {
					log.Infof("============================================")
					log.Infof("Replica Name - [%s]", replica.Name)
					log.Infof("Replica Id - [%s]", replica.ReplicaId)
					log.Infof("Replica Role - [%s]", replica.Role)
					log.Infof("Replica Failover Mode - [%s]", replica.FailoverMode)
					log.Infof("Replica Availability Mode - [%s]", replica.AvailabilityMode)
					log.Infof("Replica Connection State - [%s]", replica.ConnectionState)
					log.Infof("Replica Health - [%s]", replica.Health)
					log.Infof("============================================")
				}
			})

			Step("Inititalize HammerDB", func() {
				databaseConnDetails := databaseDetails.DatabaseDriver.GetDatabaseDetails()
				log.Infof("Primary Hostname - [%s]", databaseConnDetails.PrimaryHostName)

				hammerDBDriver = hammerdb.HammerDBInit(
					"hammerdb-"+RandomString(5),
					databaseConnDetails.PrimaryHostName,
					databaseConnDetails.User,
					databaseConnDetails.Password,
					aagDatabaseName,
					"mssql",
					NewPdsParams.SSIE.HammerDBScaleFactor,
				)
			})

			Step("Deploy and run hammerDB load", func() {
				err := hammerDBDriver.DeployHammerDBPod()
				log.FailOnError(err, "Error while deploying hammerdb pod")

				output, err := hammerDBDriver.RunHammerDBLoad()
				log.FailOnError(err, "Error while running hammerdb load")
				log.Infof("Hammerdb output - \n\n%s\n\n", output)
			})

			Step("Add database to AAG", func() {
				err := databaseDetails.DatabaseDriver.AddDatabaseToCluster(ctx, aagDatabaseName, replicaDetails.AGName)
				log.FailOnError(err, "Error occurred while adding database")
			})

			Step("Validate AAG state on source", func() {
				_, err := WorkflowDataService.ValidateSQLAAGReplicas(*deployment.Create.Meta.Uid)
				log.FailOnError(err, "replica health validation failed before hammerdb load")
				log.Infof("AAG replicas are healthy")
			})

			Step("Create Adhoc backup config of the existing deployment", func() {
				pdsBackupConfigName = "pds-adhoc-backup-" + RandomString(5)
				bkpConfigResponse, err := WorkflowPDSBackupConfig.CreateBackupConfig(pdsBackupConfigName, *deployment.Create.Meta.Uid)
				log.FailOnError(err, "Error occured while creating backupConfig")
				log.Infof("BackupConfigName: [%s], BackupConfigId: [%s]", *bkpConfigResponse.Create.Meta.Name, *bkpConfigResponse.Create.Meta.Uid)
				log.Infof("All deployments - [%+v]", WorkflowDataService.DataServiceDeployment)
			})

			Step("Get the latest backup detail for the deployment", func() {
				backupResponse, err := WorkflowPDSBackup.GetLatestBackup(*deployment.Create.Meta.Uid)
				log.FailOnError(err, "Error occurred while creating backup")
				latestBackupUid = *backupResponse.Meta.Uid
				log.Infof("Latest backup ID [%s], Name [%s]", *backupResponse.Meta.Uid, *backupResponse.Meta.Name)
				err = WorkflowPDSBackup.WaitForBackupToComplete(*backupResponse.Meta.Uid)
				log.FailOnError(err, "error occurred while waiting for backup to complete")
			})

			Step("Create Restore from the latest backup Id", func() {
				CheckforClusterSwitch()
				restoreName = "restore-" + RandomString(5)
				_, err := WorkflowPDSRestore.CreateRestore(restoreName, latestBackupUid, restoreName, *deployment.Create.Meta.Uid)
				log.FailOnError(err, "Restore Failed")
				log.Infof("All restores - [%+v]", WorkflowPDSRestore.Restores)
				log.Infof("Restore Created Name - [%s], UID - [%s]", *WorkflowPDSRestore.Restores[restoreName].Meta.Name, *WorkflowPDSRestore.Restores[restoreName].Meta.Uid)
			})

			Step("Verify AAG after hammerdb load run restore", func() {
				defer func() {
					err := SetSourceKubeConfig()
					log.FailOnError(err, "failed to switch context to source cluster")
				}()
				_, err := WorkflowPDSRestore.RestoredDeployments.ValidateSQLAAGReplicas(WorkflowPDSRestore.Restores[restoreName].Config.DestinationReferences.DataServiceDeploymentId)
				log.FailOnError(err, "replica health validation failed post restore")
			})

		}
	})

	JustAfterEach(func() {
		defer EndPDSTorpedoTest()
	})
})
