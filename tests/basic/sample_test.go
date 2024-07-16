package tests

import (
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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
			startCmd := "pidof px"
			//Capturing PID pf PX before stopping PX pods
			for _, node := range nn.GetStorageNodes() {
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

			//Capturing PID pf PX after stopping PX pods
			processPidPostRestart := make(map[string]string)
			startCmd = "pidof px-ns"
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
			for nodeDetails, beforePID := range processPid {
				afterPID, ok := processPidPostRestart[nodeDetails]
				if !ok || beforePID != afterPID {
					log.FailOnError(err, fmt.Sprintf("Px process id %s seems to have been restarted as new process id observed on %s", afterPID, nodeDetails))
				}
				Expect(beforePID).To(Equal(afterPID))
			}
		})
	})
})
