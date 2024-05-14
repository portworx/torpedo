package tests

import (
	. "github.com/onsi/ginkgo/v2"
	pdslibs "github.com/portworx/torpedo/drivers/unifiedPlatform/pdsLibs"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/stworkflows/pds"
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

var _ = Describe("{ValidateWorkloads}", func() {

	JustBeforeEach(func() {
		tests.StartPDSTorpedoTest("ValidateWorkloads", "validate  workloads", nil, 0)
	})
	It("Validate Workloads", func() {
		//stepLog := "Running Workloads before upgrading the ds image"
		//Step(stepLog, func() {
		//	err := tests.WorkflowDataService.RunDataServiceWorkloads(tests.NewPdsParams, "postgresql")
		//	log.FailOnError(err, "Error while running workloads on ds")
		//})

	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
	})
})

var _ = Describe("{ValidateDnsEndPoint}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("ValidateDnsEndPoint", "validate dns endpoint", nil, 0)
	})

	var (
		workflowDataservice pds.WorkflowDataService
		err                 error
	)

	It("ValidateDnsEndPoint", func() {
		Step("validate dns endpoint", func() {
			depId := "dep:3a13954f-ae45-4223-8896-82029c90bca9"
			err = workflowDataservice.ValidateDNSEndpoint(depId, "Cassandra")
			log.FailOnError(err, "Error occurred while validating dns endpoint")
		})
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
	})
})
