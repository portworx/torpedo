package tests

import (
	. "github.com/onsi/ginkgo/v2"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/stworkflows/pds"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests"
	tests "github.com/portworx/torpedo/tests/unifiedPlatform"
)

var _ = Describe("{DummyBackupTest}", func() {

	var (
		workflowDataservice pds.WorkflowDataService
		workflowBackup      pds.WorkflowPDSBackup
		workflowRestore     pds.WorkflowPDSRestore
		deploymentName      string
	)

	JustBeforeEach(func() {
		StartTorpedoTest("DummyBackupTest", "DummyBackupTest", nil, 0)
		workflowDataservice.DataServiceDeployment = make(map[string]string)
		deploymentName = "test"
		workflowDataservice.DataServiceDeployment[deploymentName] = "dep:cb47ce73-5304-40b0-93a7-849a2b33735a"

		workflowRestore.Destination = tests.WorkflowNamespace
		workflowRestore.WorkflowProject = tests.WorkflowProject
	})

	It("Dummy to verify backup and restore creation", func() {

		Step("Get latest backup from a backup config", func() {
			backupResponse, err := workflowBackup.GetLatestBackup(deploymentName)
			log.FailOnError(err, "Error occured while creating backup")
			log.Infof("Latest backup ID [%s], Name [%s]", *backupResponse.Meta.Uid, *backupResponse.Meta.Name)
			err = workflowBackup.WaitForBackupToComplete(*backupResponse.Meta.Uid)
			log.FailOnError(err, "Error occured while waiting for backup to complete")
		})

	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
	})
})

func PointerTo[T ~string](s T) *T {
	return &s
}
