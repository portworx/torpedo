package tests

import (
	. "github.com/onsi/ginkgo/v2"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/platformLibs"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/stworkflows"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests"
)

var tc platformLibs.TargetCluster

var _ = Describe("{TenantsCRUD}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("TenantsCRUD", "Create and Get the Tenant", pdsLabels, 0)
	})

	It("Tenants", func() {
		steplog := "Tenants CRUD"
		Step(steplog, func() {
			log.InfoD(steplog)
			var tc platformLibs.TargetCluster
			var tenantId string
			accID := "acc:2199f82a-9c39-4070-a431-4a8c8b1c2ca7"
			//accList, err := platformLibs.GetAccountListv1()
			//log.FailOnError(err, "error while getting account list")
			//accID := platformLibs.GetPlatformAccountID(accList, defaultTestAccount)
			//log.Infof("account ID [%s]", accID)
			tenantList, err := platformLibs.GetTenantListV1(accID)
			log.FailOnError(err, "error while getting tenant list")
			for _, tenant := range tenantList {
				log.Infof("Available tenant's %s under the account id %s", *tenant.Meta.Name, accID)
				tenantId = *tenant.Meta.Uid
				break
			}
			log.Infof("TenantID [%s]", tenantId)
			clusterId, err := tc.RegisterToControlPlane("1.0.0", tenantId)
			if err != nil {
				log.FailOnError(err, "Failed to register Target Cluster to Control plane")
			}
			log.Infof("Registered Cluster ID is: %v\n", clusterId)
		})
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
	})
})

//var _ = Describe("{WhoamI}", func() {
//	steplog := "WhoamI"
//	JustBeforeEach(func() {
//		StartTorpedoTest("WhoAmI", "get actor id", nil, 0)
//	})
//
//	Step(steplog, func() {
//		log.InfoD(steplog)
//		It("WhoAmI", func() {
//			Step("create accounts", func() {
//				whoAmIResp, err := platformLibs.WhoAmIV1()
//				log.FailOnError(err, "error while creating account")
//				log.Infof("Actor ID %s", whoAmIResp.GetId())
//			})
//		})
//	})
//
//	JustAfterEach(func() {
//		defer EndTorpedoTest()
//	})
//})

var _ = Describe("{DeployDataServicesOnDemand}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("DeployDataService", "Deploy data services", nil, 0)
	})

	It("Deploy and Validate DataService", func() {
		for _, ds := range NewPdsParams.DataServiceToTest {
			_, err := stworkflows.DeployDataservice(ds)
			log.FailOnError(err, "Error while deploying ds")
		}
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
	})
})

var _ = Describe("{CreateAccount}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("ListAccounts", "Create and List Accounts", nil, 0)
	})

	It("Accounts", func() {
		Step("Create and List Accounts", func() {
			workflowResponse, err := stworkflows.WorkflowCreateAndListAccounts()
			log.FailOnError(err, "Some error occurred while running WorkflowCreateAndListAccounts")
			accountList := workflowResponse[stworkflows.GetAccountListv1]
			for _, account := range accountList {
				log.Infof("Found %s as part of result", account.Meta.Name)
			}
		})
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
	})
})

var _ = Describe("{ListAccounts}", func() {
	steplog := "List Accounts"
	JustBeforeEach(func() {
		StartTorpedoTest("ListAccounts", "Create and List Accounts", nil, 0)
	})

	log.InfoD(steplog)
	It("Accounts", func() {
		var accountID string
		steplog = "ListAccounts"
		Step(steplog, func() {
			log.InfoD(steplog)
			accList, err := platformLibs.GetAccountListv1()
			log.FailOnError(err, "error while getting account list")
			for _, acc := range accList {
				if *acc.Meta.Name == defaultTestAccount {
					log.Infof("Available account %s", *acc.Meta.Name)
					log.Infof("Available account ID %s", *acc.Meta.Uid)
					accountID = *acc.Meta.Uid
				}
			}
		})
		steplog = "GetAccounts"
		Step(steplog, func() {
			log.InfoD(steplog)
			acc, err := platformLibs.GetAccount(accountID)
			log.FailOnError(err, "error while getting account info")
			log.Infof("account name is %s", *acc.Meta.Name)
		})
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
	})
})
