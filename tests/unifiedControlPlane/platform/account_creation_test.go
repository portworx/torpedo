package platform

import (
	. "github.com/onsi/ginkgo"
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

			})
		})
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
	})
})
