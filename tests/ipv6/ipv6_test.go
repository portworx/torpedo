package tests

import (
	"fmt"
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
	"github.com/portworx/torpedo/drivers/node"
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/pkg/ipv6util"
	"github.com/portworx/torpedo/pkg/testrailuttils"
	. "github.com/portworx/torpedo/tests"
)

func TestBasic(t *testing.T) {
	RegisterFailHandler(Fail)

	var specReporters []Reporter
	junitReporter := reporters.NewJUnitReporter("/testresults/junit_ipv6.xml")
	specReporters = append(specReporters, junitReporter)
	RunSpecsWithDefaultAndCustomReporters(t, "Torpedo : Ipv6", specReporters)
}

var _ = BeforeSuite(func() {
	InitInstance()
})

// Simple single test
var _ = Describe("{PxctlStatusTest}", func() {
	// update the testrailID with your test
	var testrailID = 9695443
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/tests/view/9695443
	var runID int
	JustBeforeEach(func() {
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})
	var contexts []*scheduler.Context

	It("has to run pxctl status and checks for valid ipv6 addresses", func() {
		var status string
		var err error
		var ips []string

		nodes := node.GetWorkerNodes()
		Step(fmt.Sprintln("run pxctl status"), func() {
			status, err = Inst().V.GetPxctlStatus(nodes[0], false)
			Expect(err).NotTo(HaveOccurred(), status)
		})

		Step(fmt.Sprintln("parse address from pxctl status"), func() {
			ips = ipv6util.ParseIpv6AddressInPxctlStatus(status, len(nodes))
			// number of ips are the number of nodes + 1 (the node IP where the status command is run on)
			Expect(len(ips)).To(Equal(len(nodes)+1), "unexpected ip count, should be one more than to number of worker nodes")
		})

		Step(fmt.Sprintln("validate the address are ipv6"), func() {
			isIpv6 := ipv6util.AreAddressesIPv6(ips)
			Expect(isIpv6).To(BeTrue(), "addresses in pxctl status are expected to be IPv6, parsed ips: %v", ips)
		})

	})
	JustAfterEach(func() {
		AfterEachTest(contexts, testrailID, runID)
	})
})

var _ = AfterSuite(func() {
	PerformSystemCheck()
	ValidateCleanup()
})

func TestMain(m *testing.M) {
	// call flag.Parse() here if TestMain uses flags
	ParseFlags()
	os.Exit(m.Run())
}
