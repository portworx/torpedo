package tests

import (
	. "github.com/onsi/ginkgo/v2"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/stworkflows/pds"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/stworkflows/platform"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests"
	tests "github.com/portworx/torpedo/tests/unifiedPlatform"
)

var _ = Describe("{DummyBackupTest}", func() {

	var (
		workflowDataservice pds.WorkflowDataService
		workflowBackup      pds.WorkflowPDSBackup
		workflowRestore     pds.WorkflowPDSRestore
		workflowProject     platform.WorkflowProject
		deploymentName      string
		latestBackupUid     string
	)

	JustBeforeEach(func() {
		StartTorpedoTest("DummyBackupTest", "DummyBackupTest", nil, 0)
		log.Infof("We are here")
		workflowDataservice.DataServiceDeployment = make(map[string]string)
		log.Infof("We are here")
		deploymentName = "samore-pg-test-1"
		log.Infof("We are here")
		workflowDataservice.DataServiceDeployment[deploymentName] = "dep:fa70e52d-0563-4258-b96b-7d6ca6ed4799"
		log.Infof("We are here")
		workflowProject.ProjectId = "prj:2eee1079-83f0-4b88-b61d-d8782c2bdbd4"

		workflowRestore.Destination = tests.WorkflowNamespace
		workflowRestore.WorkflowProject = workflowProject
	})

	It("Dummy to verify backup and restore creation", func() {

		Step("Get latest backup from a backup config", func() {
			log.Infof("We are here")
			backupResponse, err := workflowBackup.GetLatestBackup(deploymentName)
			log.Infof("We are here")
			log.FailOnError(err, "Error occured while creating backup")
			log.Infof("We are here")
			latestBackupUid = *backupResponse.Meta.Uid
			log.Infof("Latest backup ID [%s], Name [%s]", *backupResponse.Meta.Uid, *backupResponse.Meta.Name)
		})

		Step("Get all backup from a backup config", func() {
			backupResponse, err := workflowBackup.ListAllBackups(deploymentName)
			log.FailOnError(err, "Error occured while creating backup")
			log.Infof("Number of backups - [%d]", len(backupResponse))

			log.Infof("Listing all backups \n\n")
			for _, backup := range backupResponse {
				log.Infof("Latest backup ID [%s], Name [%s]", *backup.Meta.Uid, *backup.Meta.Name)
			}

		})

		Step("Create Restore", func() {
			restoreName := "testing_restore_" + RandomString(5)
			_, err := workflowRestore.CreateRestore(restoreName, latestBackupUid, tests.PDS_DEFAULT_NAMESPACE)
			log.FailOnError(err, "Restore Failed")

			log.Infof("Restore created successfully with ID - [%s]", workflowRestore.Restores[restoreName].Meta.Uid)
		})

	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
	})
})

func PointerTo[T ~string](s T) *T {
	return &s
}
