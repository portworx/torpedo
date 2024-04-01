package tests

import (
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/stworkflows"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests"
)

var _ = Describe("{MultiplyNumDuringSummation}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("MultiplyNumDuringSummation", "TestResiliencyDummy", nil, 0)
	})
	var (
		workflowResiliency  stworkflows.WorkflowResiliency
		workflowDataservice stworkflows.WorkflowDataService
	)
	It("Deploy and DS and Stop Px During Storage/PVC Resize", func() {
		Step("Create a PDS Namespace", func() {
			//Mark testcase as Resiliency
			workflowResiliency.MarkResiliencyTC(true)

			log.InfoD("Deploy dataservice")
		})

		stepLog := "Running Workloads before taking backups"
		Step(stepLog, func() {
			log.InfoD("Run Workloads")
		})

		Step("Induce errors while some PDS operation is going on", func() {
			err := workflowResiliency.InduceFailureAndExecuteResiliencyScenario(workflowDataservice.NamespaceName, MultiplyNumDuringSummation, PerformMultiplication())
			log.FailOnError(err, fmt.Sprintf("Error happened"))

		})

		Step("Running Workloads", func() {
			log.InfoD("Run Workloads")
		})
		Step("Clean up workload deployments", func() {
			log.InfoD("Cleanup Deployed dataservice")
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
	})
})
