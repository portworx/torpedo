package tests

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/portworx/torpedo/drivers/backup/cluster_controller"
	"github.com/portworx/torpedo/drivers/backup/cluster_controller/cluster_utils"
	"github.com/portworx/torpedo/pkg/log"
	"time"
)

var _ = Describe("{ClusterControllerTest}", func() {
	JustBeforeEach(func() {})

	It("Cluster Controller Test", func() {
		clusterController := NewDefaultClusterController()
		logFilePath := cluster_utils.DefaultLogsLocation + fmt.Sprintf("px-backup-mongo-0-%v.log", time.Now().Unix())
		err := clusterController.Cluster("/tmp/source-config").Namespace("central").PodByName("pxc-backup-mongodb-0").CollectLogs(logFilePath)
		log.FailOnError(err, "failed to collect logs")
	})

	JustAfterEach(func() {})
})
