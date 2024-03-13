package tests

import (
	. "github.com/onsi/ginkgo/v2"
	pdslib "github.com/portworx/torpedo/drivers/pds/lib"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/stworkflows"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests"
	"math/rand"
	"strconv"
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

var _ = Describe("{TestRbacForPds}", func() {
	var (
		pdsRbac  *stworkflows.UserWithRbac
		userName string
	)
	JustBeforeEach(func() {
		pdsparams := pdslib.GetAndExpectStringEnvVar("PDS_PARAM_CM")
		NewPdsParams, err := ReadNewParams(pdsparams)
		infraParams := NewPdsParams.InfraToTest
		pdsLabels["clusterType"] = infraParams.ClusterType
		rbacParams := NewPdsParams.RbacParams
		log.FailOnError(err, "Failed to read params from json file")
		if rbacParams.RunWithRbac == true {
			userName = "pdsUser-" + strconv.Itoa(rand.Int())
			err, _ := pdsRbac.CreateNewPdsUser(accID, userName, rbacParams.RoleName, rbacParams.ResourceId)
			if err != nil {
				return
			}
		}
		StartTorpedoTest("ListAccounts", "Create and List Accounts", nil, 0)
	})

	It("Accounts", func() {
		Step("Create and List Accounts", func() {
			pdsRbac.SwitchPdsUser(userName)
			//Perform any PDS workflow with Rbac enabled.
		})
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
	})
})
