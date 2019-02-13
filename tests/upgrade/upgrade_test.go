package tests

import (
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/portworx/torpedo/drivers/scheduler"
	. "github.com/portworx/torpedo/tests"
)

func TestUpgrade(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Torpedo : Upgrade")
}

var _ = BeforeSuite(func() {
	InitInstance()
})

var _ = Describe("{UpgradeVolumeDriver}", func() {
	var contexts []*scheduler.Context

	It("upgrade volume driver and ensure everything is running fine", func() {
		for i := 0; i < Inst().ScaleFactor; i++ {
			contexts = append(contexts, ScheduleAndValidate(fmt.Sprintf("upgradevolumedriver-%d", i))...)
		}

		Step("start the upgrade of volume driver", func() {
			err := Inst().V.UpgradeDriver(Inst().StorageDriverUpgradeVersion)
			Expect(err).NotTo(HaveOccurred())
		})

		Step("validate all apps after upgrade", func() {
			for _, ctx := range contexts {
				ValidateContext(ctx)
			}
		})

	})
	AfterEach(func() {
		TearDownAfterEachSpec(contexts)
	})

	JustAfterEach(func() {
		DescribeNamespaceJustAfterEachSpec(contexts)
	})
})

var _ = PDescribe("{UpgradeDowngradeVolumeDriver}", func() {
	var contexts []*scheduler.Context

	It("upgrade and downgrade volume driver and ensure everything is running fine", func() {
		for i := 0; i < Inst().ScaleFactor; i++ {
			contexts = append(contexts, ScheduleApps(fmt.Sprintf("upgradedowngradevolumedriver-%d", i))...)
		}
		ValidateApps(fmt.Sprintf("validate apps for %s", CurrentGinkgoTestDescription().TestText), contexts)

		Step("start the upgrade of volume driver", func() {
			err := Inst().V.UpgradeDriver(Inst().StorageDriverUpgradeVersion)
			Expect(err).NotTo(HaveOccurred())
		})

		Step("validate all apps after upgrade", func() {
			for _, ctx := range contexts {
				ValidateContext(ctx)
			}
		})

		Step("start the downgrade of volume driver", func() {
			err := Inst().V.UpgradeDriver(Inst().StorageDriverBaseVersion)
			Expect(err).NotTo(HaveOccurred())
		})

		opts := make(map[string]bool)
		opts[scheduler.OptionsWaitForResourceLeakCleanup] = true
		ValidateApps(fmt.Sprintf("validate apps for %s", CurrentGinkgoTestDescription().TestText), contexts)
	})
	AfterEach(func() {
		TearDownAfterEachSpec(contexts)
	})

	JustAfterEach(func() {
		DescribeNamespaceJustAfterEachSpec(contexts)
	})
})

var _ = AfterSuite(func() {
	PerformSystemCheck()
	CollectSupport()
	ValidateCleanup()
})

func init() {
	ParseFlags()
}
