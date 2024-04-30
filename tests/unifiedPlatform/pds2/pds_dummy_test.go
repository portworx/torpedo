package tests

import (
	. "github.com/onsi/ginkgo/v2"
	pdslibs "github.com/portworx/torpedo/drivers/unifiedPlatform/pdsLibs"
	pds2 "github.com/portworx/torpedo/drivers/unifiedPlatform/stworkflows/pds"
	"github.com/portworx/torpedo/pkg/aetosutil"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests"
	tests "github.com/portworx/torpedo/tests/unifiedPlatform"
)

var dash *aetosutil.Dashboard

var _ = Describe("{CleanUpDeployments}", func() {
	It("Delete all deployments", func() {
		err := pdslibs.DeleteAllDeployments(tests.ProjectId)
		log.Errorf("ERROR WHILE DELETING DEPLOYMENT [%v]", err)
		//log.FailOnError(err, "error while deleting deployment")

	})
})

var _ = Describe("{ValidateDnsEndPoint}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("ValidateDnsEndPoint", "validate dns endpoint", nil, 0)
	})

	var (
		workflowDataservice pds2.WorkflowDataService
		err                 error
	)

	It("ValidateDnsEndPoint", func() {
		Step("validate dns endpoint", func() {
			depId := "dep:8965032c-c8e3-447f-af05-9b7a4badf5a9"
			err = workflowDataservice.ValidateDNSEndpoint(depId)
			log.FailOnError(err, "Error occurred while validating dns endpoint")
		})
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
	})
})

var _ = Describe("{DummyBackupTest}", func() {

	var (
		workflowDataservice pds2.WorkflowDataService
		workflowBackup      pds2.WorkflowPDSBackup
		workflowRestore     pds2.WorkflowPDSRestore
		deploymentName      string
		latestBackupUid     string
	)

	JustBeforeEach(func() {
		StartTorpedoTest("DummyBackupTest", "DummyBackupTest", nil, 0)
		workflowDataservice.DataServiceDeployment = make(map[string]string)
		deploymentName = "samore-pg-test-1"
		workflowDataservice.DataServiceDeployment[deploymentName] = "dep:fa70e52d-0563-4258-b96b-7d6ca6ed4799"

		workflowRestore.Destination = tests.WorkflowNamespace
		workflowRestore.WorkflowProject = tests.WorkflowProject
	})

	It("Dummy to verify backup and restore creation", func() {

		Step("Get latest backup from a backup config", func() {
			backupResponse, err := workflowBackup.GetLatestBackup(deploymentName)
			log.FailOnError(err, "Error occured while creating backup")
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
