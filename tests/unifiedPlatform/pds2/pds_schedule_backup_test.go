package tests

import (
	. "github.com/onsi/ginkgo/v2"
	"github.com/pure-px/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/pure-px/torpedo/pkg/log"
	. "github.com/pure-px/torpedo/tests"
	. "github.com/pure-px/torpedo/tests/unifiedPlatform"
	"time"
)

var _ = Describe("{ValidateScheduleBackupAndRetention}", func() {
	var (
		deployment          *automationModels.PDSDeploymentResponse
		pdsBackupConfigName string
		err                 error
		schedule            *automationModels.V1BackupPolicy
		allBackups          []automationModels.V1Backup
		intervalTime        time.Duration
		retainTime          string
	)

	JustBeforeEach(func() {
		StartPDSTorpedoTest("ValidateScheduleBackupAndRetention", "Validate schedule backup timing and retention count", nil, 0)
		intervalTime = 2 * time.Minute
		retainTime = "3"
	})

	It("Validate schedule backup timing and retention count", func() {
		for _, ds := range NewPdsParams.DataServiceToTest {

			Step("Deploy dataservice", func() {
				deployment, err = WorkflowDataService.DeployDataService(ds, ds.Image, ds.Version, PDS_DEFAULT_NAMESPACE)
				log.FailOnError(err, "Error while deploying ds")
				log.Infof("All deployments - [%+v]", WorkflowDataService.DataServiceDeployment)

			})

			Step("Create a schedule for the backup", func() {
				scheduleName := "schedule-" + RandomString(5)
				schedule, err = WorkflowSchedule.CreateSchedule(scheduleName, intervalTime, retainTime)
				log.FailOnError(err, "Unable to create schedule")
				log.InfoD("Schedule created with uid - [%s]", *schedule.Meta.Uid)
			})

			Step("Associate backup schedule to the project", func() {
				log.InfoD("Associate backup schedule to the project")
				err := WorkflowProject.Associate(
					[]string{},
					[]string{},
					[]string{},
					[]string{},
					[]string{},
					[]string{*schedule.Meta.Uid},
				)
				log.FailOnError(err, "Unable to associate Cluster to Project")
				log.Infof("Associated Resources - [%+v]", WorkflowProject.AssociatedResources)
			})

			Step("Create Adhoc backup config of the existing deployment", func() {
				pdsBackupConfigName = "pds-adhoc-backup-" + RandomString(5)
				bkpConfigResponse, err := WorkflowPDSBackupConfig.CreateScheduleBackup(pdsBackupConfigName, *deployment.Create.Meta.Uid, *schedule.Meta.Uid)
				log.FailOnError(err, "Error occured while creating backupConfig")
				log.Infof("BackupConfigName: [%s], BackupConfigId: [%s]", *bkpConfigResponse.Create.Meta.Name, *bkpConfigResponse.Create.Meta.Uid)
				log.Infof("All deployments - [%+v]", WorkflowDataService.DataServiceDeployment)
			})

			Step("Waiting for first backup to trigger", func() {
				timeTaken, err := WorkflowPDSBackup.CheckFirstBackupTime(*deployment.Create.Meta.Uid, intervalTime)
				log.FailOnError(err, "Error occured while waiting for first backup to trigger")
				log.InfoD("Time taken to create first backup - [%v] minutes", timeTaken.Minutes())
				dash.VerifyFatal(timeTaken > intervalTime, false, "Validating first schedule backup time")
			})

			Step("Waiting for schedule backup to trigger atleast 3 times", func() {
				log.InfoD("Sleeping for 10 minutes")
				time.Sleep(5 * intervalTime)
			})

			Step("Get all backups detail for the deployment", func() {
				allBackups, err = WorkflowPDSBackup.ListAllBackups(*deployment.Create.Meta.Uid)
				log.FailOnError(err, "Error occured while listing all backups")
				dash.VerifyFatal(len(allBackups) == 3, true, "Validating backup retention policy")
			})

			Step("Delete all scheduled backups", func() {
				for _, backup := range allBackups {
					log.InfoD("Deleting backup - [%s]", *backup.Meta.Uid)
					err = WorkflowPDSBackup.DeleteBackup(*backup.Meta.Uid)
					log.FailOnError(err, "Error occured while deleting backup")
				}
			})

			// this is a workaround to delete the backup config
			Step("Delete the data service", func() {
				err := WorkflowDataService.Purge(false)
				log.FailOnError(err, "Error occured while deleting dataservice")
			})

			Step("Delete backup config", func() {
				err = WorkflowPDSBackupConfig.Purge(true)
				log.FailOnError(err, "Error occured while deleting backupConfig")
			})

		}

	})

	JustAfterEach(func() {
		defer EndPDSTorpedoTest()
	})
})
