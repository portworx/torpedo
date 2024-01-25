package platform

import (
	. "github.com/onsi/ginkgo"
	pdslib "github.com/portworx/torpedo/drivers/pds/lib"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests"
)

var _ = Describe("{ListAccounts}", func() {
	steplog := "ListAccounts"
	JustBeforeEach(func() {
		StartTorpedoTest("ListAccounts", "validate dns endpoitns", nil, 0)
	})

	Step(steplog, func() {
		log.InfoD(steplog)
		It("ListAccounts", func() {
			steplog = "ListAccounts"
			Step(steplog, func() {
				log.InfoD(steplog)
				accList, err := pdslib.GetAccountListV2()
				log.FailOnError(err, "error while getting account list")
				for _, acc := range accList.Accounts {
					log.Infof("Available account %s", *acc.Meta.Name)
				}
			})
		})
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
	})
})
