package platform

import (
	. "github.com/onsi/ginkgo"
	platformUtils "github.com/portworx/torpedo/drivers/unifiedPlatform/platform/platformUtils"
	tc "github.com/portworx/torpedo/drivers/unifiedPlatform/platform/targetcluster"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests"
)

var targetCluster tc.TargetCluster

var _ = Describe("{TenantsCRUD}", func() {
	steplog := "Tenants CRUD"
	JustBeforeEach(func() {
		StartTorpedoTest("ListTenants", "Create and Get the Tenant", nil, 0)
	})

	Step(steplog, func() {
		log.InfoD(steplog)
		var tenantId string
		It("Tenants", func() {
			steplog = "ListTenants"
			Step(steplog, func() {
				log.InfoD(steplog)
				accList, err := platformUtils.GetPlatformAccountListV1()
				log.FailOnError(err, "error while getting account list")
				accID := platformUtils.GetPlatformAccountID(accList, defaultTestAccount)
				log.Infof("account ID [%s]", accID)
				tenantList, err := platformUtils.GetPlatformTenantListV1(accID)
				log.FailOnError(err, "error while getting tenant list")
				for _, tenant := range tenantList {
					log.Infof("Available tenant's %s under the account id %s", *tenant.Meta.Name, accID)
					tenantId = *tenant.Meta.Uid
					break
				}
				err = targetCluster.RegisterToControlPlane("1.0.0", tenantId, "")
				if err != nil {
					log.FailOnError(err, "Failed to register Target Cluster to Control plane")
				}
			})
		})
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
	})
})

var _ = Describe("{WhoamI}", func() {
	steplog := "WhoamI"
	JustBeforeEach(func() {
		StartTorpedoTest("WhoAmI", "get actor id", nil, 0)
	})

	Step(steplog, func() {
		log.InfoD(steplog)
		It("WhoAmI", func() {
			Step("create accounts", func() {
				whoAmIResp, err := platformUtils.WhoAmIV1()
				log.FailOnError(err, "error while creating account")
				log.Infof("Actor ID %s", whoAmIResp.GetId())
			})
		})
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
	})
})

var _ = Describe("{AccountsCRUD}", func() {
	steplog := "Accounts CRUD"
	JustBeforeEach(func() {
		StartTorpedoTest("ListAccounts", "Create and List Accounts", nil, 0)
	})

	Step(steplog, func() {
		log.InfoD(steplog)
		It("Accounts", func() {
			Step("create accounts", func() {
				acc, err := platformUtils.CreatePlatformAccountV1(envPlatformAccountName, envAccountDisplayName, envUserMailId)
				log.FailOnError(err, "error while creating account")
				log.Infof("created account with name %s", *acc.Meta.Name)
			})
			steplog = "ListAccounts"
			Step(steplog, func() {
				log.InfoD(steplog)
				accList, err := platformUtils.GetPlatformAccountListV1()
				log.FailOnError(err, "error while getting account list")
				for _, acc := range accList {
					log.Infof("Available account %s", *acc.Meta.Name)
					log.Infof("Available account ID %s", *acc.Meta.Uid)
				}
			})
		})
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
	})
})
