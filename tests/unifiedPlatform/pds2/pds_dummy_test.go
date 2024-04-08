package tests

import (
	. "github.com/onsi/ginkgo/v2"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/stworkflows/pds"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests"
)

var _ = Describe("{DummyBackupTest}", func() {

	var (
		workflowBackUpConfig pds.WorkflowPDSBackupConfig
		workflowBackup       pds.WorkflowPDSBackup
		backupConfigName     string
	)

	JustBeforeEach(func() {
		StartTorpedoTest("DummyBackupTest", "DummyBackupTest", nil, 0)

		workflowBackUpConfig.Backups = make(map[string]automationModels.V1BackupConfig)
		backupConfigName = "pds-qa-bkp"
		dummyBackupConfig := automationModels.V1BackupConfig{
			Meta: &automationModels.Meta{
				Uid:         PointerTo("bkc:ab9aa7b5-2240-474f-822e-d7ed1f276d9a"),
				Name:        PointerTo(backupConfigName),
				Description: PointerTo(""),
			},
		}

		workflowBackUpConfig.Backups[backupConfigName] = dummyBackupConfig
		workflowBackup.WorkflowBackupConfig = workflowBackUpConfig
	})

	It("Dummy to verify backup and restore creation", func() {

		Step("Get latest backup from a backup config", func() {
			backupResponse, err := workflowBackup.GetLatestBackup(backupConfigName)
			log.FailOnError(err, "Error occured while creating backup")
			log.Infof("Latest backup ID [%s], Name [%s]", *backupResponse.Meta.Uid, *backupResponse.Meta.Name)
		})

		Step("Get all backup from a backup config", func() {
			backupResponse, err := workflowBackup.ListAllBackups(backupConfigName)
			log.FailOnError(err, "Error occured while creating backup")
			log.Infof("Number of backups - [%d]", len(backupResponse))

			log.Infof("Listing all backups \n\n")
			for _, backup := range backupResponse {
				log.Infof("Latest backup ID [%s], Name [%s]", *backup.Meta.Uid, *backup.Meta.Name)
			}

		})

	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
	})
})

func PointerTo[T ~string](s T) *T {
	return &s
}
