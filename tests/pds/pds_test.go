package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/portworx/torpedo/tests"

	"github.com/portworx/torpedo/pkg/pds"
)

func TestPDS(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Torpedo: PDS")
}

var (
	controlPlane  *pds.ControlPlane
	targetCluster *pds.TargetCluster
)

var _ = BeforeSuite(func() {
	cpKubeconfig, err := tests.GetSourceClusterConfigPath()
	Expect(err).NotTo(HaveOccurred())
	Expect(cpKubeconfig).To(BeAnExistingFile())
	controlPlane, err = pds.NewControlPlane(cpKubeconfig)
	Expect(err).NotTo(HaveOccurred())

	targetKubeconfig, err := tests.GetDestinationClusterConfigPath()
	Expect(err).NotTo(HaveOccurred())
	Expect(targetKubeconfig).To(BeAnExistingFile())
	targetCluster, err = pds.NewTargetCluster(targetKubeconfig)
	Expect(err).NotTo(HaveOccurred())
})

var _ = Describe("PDS", func() {
	var (
		startTime time.Time
	)

	BeforeEach(func() {
		startTime = time.Now()
	})

	AfterEach(func() {
		By("getting control plane logs", func() {
			logs, err := controlPlane.ComponentLogsSince(context.TODO(), startTime)
			Expect(err).NotTo(HaveOccurred())
			fmt.Fprint(GinkgoWriter, logs)
		})
	})

	Describe("Test logging", func() {
		When("test passes", func() {
			It("prints nothing", func() {})
		})

		When("test is skipped", func() {
			BeforeEach(func() {
				Skip("test skipped")
			})
			It("prints nothing", func() {})
		})

		When("test fails", func() {
			BeforeEach(func() {
				Fail("forced failure")
			})
			It("prints component logs", func() {
				// Expect component logs written to output.
			})
		})
	})
})
