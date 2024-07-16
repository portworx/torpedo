package tests

import (
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	nn "github.com/portworx/torpedo/drivers/node"
	"github.com/portworx/torpedo/pkg/log"
	"time"

	//"github.com/portworx/torpedo/drivers/scheduler"
	. "github.com/portworx/torpedo/tests"
)

// Arpit jira https://purestorage.atlassian.net/browse/PTX-24859

var _ = Describe("{VerifyNoNodeRestartUponPxPodRestart}", func() {

	JustBeforeEach(func() {
		StartTorpedoTest("VerifyNoNodeRestartUponPxPodRestart", "Verify that px serivce remain up even if px pod got deleted ", nil, 0)
	})

	It("has to setup, validate and teardown apps", func() {

		// Get uptime for px service on each node
		Step(stepLog, func() {
			log.InfoD(stepLog)
			processPid := make(map[string]string)

			for _, node := range nn.GetStorageNodes() {

				startCmd := "pidof px" //sudo systemctl status portworx
				output, _ := Inst().N.RunCommand(node, startCmd, nn.ConnectionOpts{
					Timeout:         30 * time.Second,
					TimeBeforeRetry: 20 * time.Second,
					Sudo:            true,
				})
				processPid[node.Id] = output
			}
			log.Infof(fmt.Sprintf("Process IDs for px before stopping portworx pod  %s", processPid))

			//Deleting px pods from all the node
			err = DeletePXPods("kube-system")

			processPidPostRestart := make(map[string]string)
			startCmd = "pidof px"
			for _, node := range nn.GetStorageNodes() {
				output, _ := Inst().N.RunCommand(node, startCmd, nn.ConnectionOpts{
					Timeout:         20 * time.Second,
					TimeBeforeRetry: 5 * time.Second,
					Sudo:            true,
				})
				processPidPostRestart[node.Id] = output
			}
			log.Infof(fmt.Sprintf("Process IDs for px after stopping portworx pod  %s", processPidPostRestart))
			//Verify PID before and after for PX process
			for key, value1 := range processPid {
				value2, ok := processPidPostRestart[key]
				if !ok || value1 != value2 {
					log.FailOnError(err, fmt.Sprintf("Px process seems to have restarted  as new process id observed on %s %s", key, value2))
				}
			}

		})
	})
})
