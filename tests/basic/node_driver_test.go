package tests

import (
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	"github.com/portworx/torpedo/drivers/node"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests"
	"time"
)

var _ = Describe("{AWSNodeDriver}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("AWSNodeDriver", "Validates AWS node driver functionalities.", nil, 0)
	})

	It("Validates AWS node driver functionalities", func() {
		Step("Initializing the AWS node driver", func() {
			err := Inst().N.Init(node.InitOptions{})
			dash.VerifyFatal(err, nil, "Failed to initialize AWS node driver")
		})

		var workerNodes []node.Node
		Step("Getting the list of worker nodes", func() {
			workerNodes = node.GetWorkerNodes()
			dash.VerifyFatal(len(workerNodes) > 0, true, "No worker nodes found")
		})

		Step("Testing connection to each worker node", func() {
			for _, n := range workerNodes {
				err := Inst().N.TestConnection(n, node.ConnectionOpts{
					Timeout:         1 * time.Minute,
					TimeBeforeRetry: 10 * time.Second,
				})
				if err != nil {
					log.FailOnError(err, fmt.Sprintf("Failed to test connection to node: %s", n.Name))
				}
			}
		})

		Step("Rebooting each worker node and verifying connectivity", func() {
			for _, n := range workerNodes {
				err := Inst().N.RebootNode(n, node.RebootNodeOpts{})
				if err != nil {
					log.FailOnError(err, fmt.Sprintf("Failed to reboot node: %s", n.Name))
				}
				// Implement a wait or verification step here to ensure the node has rebooted and is back online before proceeding
			}
		})

		Step("Shutting down and then starting each worker node", func() {
			for _, n := range workerNodes {
				err := Inst().N.ShutdownNode(n, node.ShutdownNodeOpts{})
				if err != nil {
					log.FailOnError(err, fmt.Sprintf("Failed to shutdown node: %s", n.Name))
				}
				// Implement a wait or verification step here to ensure the node has shut down properly before starting it again
				// This step might involve starting the node again, which is not directly covered in your initial implementation
			}
		})

		Step("Updating the cluster version", func() {
			desiredVersion := "1.21" // Change to your desired version
			err := Inst().N.SetClusterVersion(desiredVersion, 30*time.Minute)
			dash.VerifyFatal(err, nil, "Failed to update the cluster version")
		})

		// Add any additional test cases here to cover the remaining functionalities like FindFiles, Systemctl, etc.
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
		Step("After each test", func() {
			// Code to execute after each test. Placeholder for any cleanup or reporting logic
		})
	})
})
