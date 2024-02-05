package platform

import (
	. "github.com/onsi/ginkgo"
	pdslib "github.com/portworx/torpedo/drivers/pds/lib"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests"
)

var _ = Describe("{TenantsCRUD}", func() {
	steplog := "Tenants CRUD"
	JustBeforeEach(func() {
		StartTorpedoTest("ListTenants", "Create and Get the Tenant", nil, 0)
	})

	Step(steplog, func() {
		log.InfoD(steplog)
		It("Tenants", func() {
			steplog = "ListTenants"
			Step(steplog, func() {
				log.InfoD(steplog)
				accList, err := pdslib.GetAccountListV2()
				log.FailOnError(err, "error while getting account list")
				accID := pdslib.GetPlatformAccountID(accList, "demo-milestone-one")
				log.Infof("account ID [%s]", accID)
				tenantList, err := pdslib.GetTenantList(accID)
				log.FailOnError(err, "error while getting tenant list")
				for _, tenant := range tenantList {
					log.Infof("Available tenants under the account id %s", *tenant.Meta.Name)
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
				whoAmIResp, err := pdslib.WhoAmI()
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
				acc, err := pdslib.CreateAccountV2("test-account", "qa-test-automation-account", "marunachalam+2@purestorage.com")
				log.FailOnError(err, "error while creating account")
				log.Infof("created account with name %s", *acc.Meta.Name)
			})
			steplog = "ListAccounts"
			Step(steplog, func() {
				log.InfoD(steplog)
				accList, err := pdslib.GetAccountListV2()
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
