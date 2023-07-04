package tests

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	"github.com/portworx/torpedo/drivers/backup_controller"
	"github.com/portworx/torpedo/drivers/backup_controller/cluster_controller/cluster_utils"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests"
	"time"
)

var _ = Describe("{ClusterControllerTest}", func() {
	var (
		backupController *backup_controller.BackupController
	)

	JustBeforeEach(func() {
		backupController = backup_controller.NewDefaultBackupController()
	})

	It("Cluster Controller Test", func() {
		Step("Collect PxBackup MongoDB logs", func() {
			clusterController := backupController.GetClusterController()
			logFilePath := cluster_utils.DefaultLogsLocation + fmt.Sprintf("px-backup-mongo-0-%v.log", time.Now().Unix())
			err := clusterController.Cluster("/tmp/source-config").Namespace("central").PodByName("pxc-backup-mongodb-0").CollectLogs(logFilePath)
			log.FailOnError(err, "failed to collect logs")
		})
		Step("Create AWS Storage Location", func() {
			storageLocationController := backupController.GetStorageLocationController()
			err := storageLocationController.AWSStorageLocation("test-torpedo-storage-location-controller").Create(false, 0, "")
			log.FailOnError(err, "failed to create AWS storage location")
		})
	})

	JustAfterEach(func() {})
})
