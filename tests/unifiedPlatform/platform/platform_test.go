package tests

import (
	. "github.com/onsi/ginkgo/v2"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/stworkflows"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests"
)

var _ = Describe("{PlatformOnboardingTest}", func() {
	var (
		workflowPlatform      stworkflows.WorkflowPlatform
		workflowTargetCluster stworkflows.WorkflowTargetCluster
	)
	JustBeforeEach(func() {
		StartTorpedoTest("TestingWorkflowToOnboardAccounts", "Onboard Accounts", nil, 0)
		workflowPlatform.Accounts = map[string]map[string]string{
			"testAccount1": map[string]string{
				apiStructs.UserName:        "testAccount1",
				apiStructs.UserDisplayName: "testAccount1",
				apiStructs.UserEmail:       "atrivedi+1@purestorage.com",
			},
		}
	})

	It("Onboard accounts", func() {
		Step("Onboarding Accounts", func() {
			// TODO: This needs to be enabled once https://portworx.atlassian.net/browse/DS-8552 is fixed
			//workflowResponse, err := workflowPlatform.OnboardAccounts()
			//log.FailOnError(err, "Unable to create accounts")
			//log.Infof("All Account onboarded Successfully, Response - [%+v]", workflowResponse)
			log.Infof("Initilaising values for tenant")
			workflowPlatform.TenantInit()
		})

		Step("Register Target Cluster", func() {

			workflowTargetCluster.Platform = workflowPlatform
			log.Infof("Tenant ID [%s]", workflowTargetCluster.Platform.TenantId)
			workflowTargetCluster, err := workflowTargetCluster.RegisterToControlPlane()
			log.FailOnError(err, "Unable to register target cluster")
			log.Infof("Target cluster registered with uid - [%s]", workflowTargetCluster.ClusterUID)
		})

		Step("Install PDS Applications", func() {
			err := workflowTargetCluster.InstallPDSAppOnTC()
			log.FailOnError(err, "Install PDS apps failed on Target cluster")
			log.Infof("PDS Apps deployed successfully on the targte cluster")
		})

	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
	})
})