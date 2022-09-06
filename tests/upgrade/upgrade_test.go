package tests

import (
	"fmt"
	"github.com/portworx/sched-ops/k8s/apps"
	"github.com/portworx/torpedo/drivers/node"
	"github.com/portworx/torpedo/drivers/scheduler/k8s"
	"github.com/portworx/torpedo/drivers/volume"
	"github.com/portworx/torpedo/pkg/aetosutil"
	"github.com/sirupsen/logrus"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/portworx/torpedo/pkg/testrailuttils"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
	"github.com/portworx/torpedo/drivers/scheduler"
	. "github.com/portworx/torpedo/tests"
)

var tpLog *logrus.Logger
var testSet aetosutil.TestSet

var dash *aetosutil.Dashboard

var storkLabel = map[string]string{"name": "stork"}
var f *os.File

const (
	storkDeploymentName = "stork"
	storkNamespace      = "kube-system"
	pxctlCDListCmd      = "pxctl cd list"
)

func TestUpgrade(t *testing.T) {
	RegisterFailHandler(Fail)

	var specReporters []Reporter
	junitReporter := reporters.NewJUnitReporter("/testresults/junit_Upgrade.xml")
	specReporters = append(specReporters, junitReporter)
	RunSpecsWithDefaultAndCustomReporters(t, "Torpedo : Upgrade", specReporters)
}

var _ = BeforeSuite(func() {
	tpLog = Inst().Logger
	dash = Inst().Dash
	if dash.TestsetID == 0 {
		testSet = aetosutil.TestSet{
			CommitID:    "2.12.0-serfdf",
			User:        "lsrinivas",
			Product:     "PxEnp",
			Description: "Torpedo : Upgrade",
			Branch:      "master",
			TestType:    "SystemTest",
			Tags:        []string{"upgrade"},
			Status:      aetosutil.NOTSTARTED,
		}
		dash.TestSetBegin(&testSet)
	}

	InitInstance()
})

var _ = Describe("{SampleTest}", func() {

	var testrailID = 35269
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/cases/view/35269
	var runID int

	var f *os.File

	JustBeforeEach(func() {
		f = CreateLogFile("SampleTest.log")
		if f != nil {
			SetTorpedoFileOutput(tpLog, f)
		}

		dash.TestCaseBegin("upgrade: sample test", "validating logs in tests", "", nil)

		runID = testrailuttils.AddRunsToMilestone(testrailID)
		dash.Infof("runid: %d", runID)
	})

	It("upgrade volume driver and ensure everything is running fine", func() {

		dash.Info("Inside upgrade test")

		Step("start the upgrade of volume driver", func() {

			dash.Info("starting upgrade")
			dash.VerifySafely("2.12.0", "2.12.0", "validating PX version")
			Expect("test").To(BeEmpty())
		})

		Step("reinstall and validate all apps after upgrade", func() {
			dash.Info("Scheduling apps after upgrade")
			tpLog.Info("Apps Scheduled")

		})

		Step("destroy apps", func() {

			dash.Info("Destroying apps")
		})
	})
	JustAfterEach(func() {
		defer dash.TestCaseEnd()
		defer CloseLogFile(tpLog, f)
	})
})

var _ = Describe("{UpgradeVolumeDriver}", func() {
	var testrailID = 35269
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/cases/view/35269
	var runID int
	JustBeforeEach(func() {
		f = CreateLogFile("UpgradeVolumeDriver.log")
		if f != nil {
			SetTorpedoFileOutput(tpLog, f)
		}

		dash.TestCaseBegin("upgrade: UpgradeVolumeDriver", "validating volume driver upgrade", "", nil)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})
	var contexts []*scheduler.Context

	It("upgrade volume driver and ensure everything is running fine", func() {
		contexts = make([]*scheduler.Context, 0)

		storageNodes := node.GetStorageNodes()

		isCloudDrive, err := IsCloudDriveInitialised(storageNodes[0])
		dash.VerifyFatal(err, nil, "Validate cloud drive installation")
		Expect(err).NotTo(HaveOccurred())

		if !isCloudDrive {
			for _, storageNode := range storageNodes {
				err := Inst().V.AddBlockDrives(&storageNode, nil)
				if err != nil && strings.Contains(err.Error(), "no block drives available to add") {
					continue
				}
				dash.VerifyFatal(err, nil, "Verify adding block drive(s)")
				Expect(err).NotTo(HaveOccurred())
			}
		}
		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("upgradevolumedriver-%d", i))...)
		}

		ValidateApplications(contexts)
		var timeBeforeUpgrade time.Time
		var timeAfterUpgrade time.Time

		Step("start the upgrade of volume driver", func() {
			IsOperatorBasedInstall, _ := Inst().V.IsOperatorBasedInstall()
			if IsOperatorBasedInstall {
				timeBeforeUpgrade = time.Now()
				status, err := UpgradePxStorageCluster()
				timeAfterUpgrade = time.Now()
				if status {
					tpLog.Info("Volume Driver upgrade is successful")
				} else {
					tpLog.Error("Volume Driver upgrade failed")
				}
				dash.VerifyFatal(err, nil, "Verify volume drive upgrade for operator based set up")
				Expect(err).NotTo(HaveOccurred())

			} else {
				timeBeforeUpgrade = time.Now()
				err := Inst().V.UpgradeDriver(Inst().StorageDriverUpgradeEndpointURL,
					Inst().StorageDriverUpgradeEndpointVersion,
					false)
				timeAfterUpgrade = time.Now()
				dash.VerifyFatal(err, nil, "Verify volume drive upgrade for daemon set based set up")
				Expect(err).NotTo(HaveOccurred())
			}

			durationInMins := int(timeAfterUpgrade.Sub(timeBeforeUpgrade).Minutes())
			expectedUpgradeTime := 9 * len(node.GetStorageDriverNodes())
			dash.VerifySafely(durationInMins <= expectedUpgradeTime, true, "Verify volume drive upgrade within expected time")
			if durationInMins <= expectedUpgradeTime {
				tpLog.Infof("Upgrade successfully completed in %d minutes which is within %d minutes", durationInMins, expectedUpgradeTime)
			} else {
				tpLog.Errorf("Upgrade took %d minutes to completed which is greater than expected time %d minutee", durationInMins, expectedUpgradeTime)
				Expect(durationInMins <= expectedUpgradeTime).To(BeTrue())
			}
		})

		Step("reinstall and validate all apps after upgrade", func() {
			tpLog.Infof("Schedulings apps after upgrade")
			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				contexts = append(contexts, ScheduleApplications(fmt.Sprintf("upgradedvolumedriver-%d", i))...)
			}
			ValidateApplications(contexts)
		})

		Step("destroy apps", func() {
			opts := make(map[string]bool)
			opts[scheduler.OptionsWaitForResourceLeakCleanup] = true
			for _, ctx := range contexts {
				TearDownContext(ctx, opts)
			}
		})
	})
	JustAfterEach(func() {
		defer dash.TestCaseEnd()
		defer CloseLogFile(tpLog, f)
		AfterEachTest(contexts, testrailID, runID)
	})
})

var _ = Describe("{UpgradeStork}", func() {
	var testrailID = 11111
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/cases/view/35269
	var runID int
	var contexts []*scheduler.Context
	JustBeforeEach(func() {
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})

	for i := 0; i < Inst().GlobalScaleFactor; i++ {

		It("upgrade volume driver and ensure everything is running fine", func() {
			contexts = make([]*scheduler.Context, 0)
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("upgradestorkdeployment-%d", i))...)

			ValidateApplications(contexts)

			Step("start the upgrade of stork deployment", func() {
				err := Inst().V.UpgradeStork(Inst().StorageDriverUpgradeEndpointURL,
					Inst().StorageDriverUpgradeEndpointVersion)
				Expect(err).NotTo(HaveOccurred())
			})

			Step("validate all apps after upgrade", func() {
				for _, ctx := range contexts {
					ValidateContext(ctx)
				}
			})

			Step("destroy apps", func() {
				opts := make(map[string]bool)
				opts[scheduler.OptionsWaitForResourceLeakCleanup] = true
				for _, ctx := range contexts {
					TearDownContext(ctx, opts)
				}
			})

			Step("validate stork pods after upgrade", func() {
				k8sApps := apps.Instance()
				storkDeploy, err := k8sApps.GetDeployment(storkDeploymentName, storkNamespace)
				Expect(err).NotTo(HaveOccurred())
				err = k8sApps.ValidateDeployment(storkDeploy, k8s.DefaultTimeout, k8s.DefaultRetryInterval)
				Expect(err).NotTo(HaveOccurred())
			})

		})
	}
	JustAfterEach(func() {
		defer dash.TestCaseEnd()
		defer CloseLogFile(tpLog, f)
		AfterEachTest(contexts, testrailID, runID)
	})
})

func getImages(version string) []volume.Image {
	images := make([]volume.Image, 0)
	for _, imagestr := range strings.Split(version, ",") {
		image := strings.Split(imagestr, "=")
		if len(image) > 1 {
			images = append(images, volume.Image{Type: image[0], Version: image[1]})
		} else {
			images = append(images, volume.Image{Type: "", Version: image[0]})
		}

	}
	return images
}

/* We don't support downgrade volume drive, so comment it out
var _ = PDescribe("{UpgradeDowngradeVolumeDriver}", func() {
	It("upgrade and downgrade volume driver and ensure everything is running fine", func() {
		var contexts []*scheduler.Context
		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("upgradedowngradevolumedriver-%d", i))...)
		}

		ValidateApplications(contexts)

		Step("start the upgrade of volume driver", func() {
			images := getImages(Inst().StorageDriverUpgradeVersion)
			err := Inst().V.UpgradeDriver(images)
			Expect(err).NotTo(HaveOccurred())
		})

		Step("validate all apps after upgrade", func() {
			for _, ctx := range contexts {
				ValidateContext(ctx)
			}
		})

		Step("start the downgrade of volume driver", func() {
			images := getImages(Inst().StorageDriverBaseVersion)
			err := Inst().V.UpgradeDriver(images)
			Expect(err).NotTo(HaveOccurred())
		})

		opts := make(map[string]bool)
		opts[scheduler.OptionsWaitForResourceLeakCleanup] = true
		ValidateAndDestroy(contexts, opts)
	})
})
*/
var _ = AfterSuite(func() {
	defer dash.TestSetEnd()
	defer CloseLogFile(tpLog, nil)
	PerformSystemCheck()
	ValidateCleanup()
})

func TestMain(m *testing.M) {
	// call flag.Parse() here if TestMain uses flags
	ParseFlags()
	os.Exit(m.Run())
}
