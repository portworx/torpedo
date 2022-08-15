package tests

import (
	"fmt"
	"math"
	"os"
	"strings"
	"testing"
	"time"

	storkv1 "github.com/libopenstorage/stork/pkg/apis/stork/v1alpha1"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
	"github.com/portworx/sched-ops/k8s/stork"
	"github.com/portworx/torpedo/drivers/node"
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/drivers/volume"
	"github.com/portworx/torpedo/pkg/testrailuttils"
	. "github.com/portworx/torpedo/tests"
	"github.com/sirupsen/logrus"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	storkops = stork.Instance()
)

const (
	dropPercentage      = 20
	delayInMilliseconds = 250
	//24 hours
	totalTimeInHours = 24
	// TODO need to make it 60 minutes
	errorPersistTimeInMinutes     = 10 * time.Minute
	snapshotScheduleRetryInterval = 10 * time.Second
	snapshotScheduleRetryTimeout  = 3 * time.Minute
)

func TestBasic(t *testing.T) {
	RegisterFailHandler(Fail)

	var specReporters []Reporter
	junitReporter := reporters.NewJUnitReporter("/testresults/junit_basic.xml")
	specReporters = append(specReporters, junitReporter)
	RunSpecsWithDefaultAndCustomReporters(t, "Torpedo : Basic", specReporters)
}

var _ = BeforeSuite(func() {
	InitInstance()
	// package installation hangs when using nsenter to run command on nodes
	if Inst().N.IsUsingSSH() {
		for _, anode := range node.GetWorkerNodes() {
			// TODO: support other OS'es
			logrus.Infof("installing tcpdump on node %s", anode.Name)
			cmd := "yum install -y tcpdump"
			_, err := Inst().N.RunCommandWithNoRetry(anode, cmd, node.ConnectionOpts{
				Timeout:         1 * time.Minute,
				TimeBeforeRetry: 5 * time.Second,
				Sudo:            true,
			})
			if err != nil {
				logrus.Warnf("failed to install tcpdump on node %s: %v", anode.Name, err)
				break
			}
		}
	}
})

// This test scales up and down an application and checks if app has actually scaled accordingly
var _ = Describe("{NetworkErrorInjection}", func() {
	var testrailID = 3526435
	injectionType := "drop"
	//TODO need to fix this issue later.
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/cases/view/35264
	var runID int
	JustBeforeEach(func() {
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})
	var contexts []*scheduler.Context
	It("Inject network error while applications are running", func() {
		contexts = make([]*scheduler.Context, 0)

		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			logrus.Infof("Iteration number %d", i)
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("applicationscaleup-%d", i))...)
		}
		currentTime := time.Now()
		timeToExecuteTest := time.Now().Local().Add(time.Hour*time.Duration(totalTimeInHours) +
			time.Minute*time.Duration(0) +
			time.Second*time.Duration(0))
		Step("Verify applications after deployment", func() {
			for _, ctx := range contexts {
				ValidateContext(ctx)
			}
		})

		Step("Create snapshot schedule policy", func() {
			createSnapshotSchedule(contexts)
		})

		for int64(timeToExecuteTest.Sub(currentTime).Seconds()) > 0 {
			// TODO core check
			//TODO Enable fstrim
			logrus.Infof("Remaining time to test in minutes : %d ", int64(timeToExecuteTest.Sub(currentTime).Seconds()/60))
			Step("Set packet loss on random nodes ", func() {
				//Get all nodes and set eth0
				nodes := node.GetWorkerNodes()
				numberOfNodes := int(math.Ceil(float64(0.40) * float64(len(nodes))))
				selectedNodes := nodes[:numberOfNodes]
				//nodes []Node, errorInjectionType string, operationType string,
				//dropPercentage int, delayInMilliseconds int
				logrus.Infof("Set network error injection")
				Inst().N.InjectNetworkError(selectedNodes, injectionType, "add", dropPercentage, delayInMilliseconds)
				logrus.Infof("Wait %d minutes before checking px status ", errorPersistTimeInMinutes/(time.Minute))
				time.Sleep(errorPersistTimeInMinutes)
				for _, n := range nodes {
					logrus.Infof("Check PX status on %v", n.Name)
					Inst().V.WaitForPxPodsToBeUp(n)
				}
				logrus.Infof("Clear network error injection ")
				Inst().N.InjectNetworkError(selectedNodes, injectionType, "delete", 0, 0)
				//Get kvdb members and
				if injectionType == "drop" {
					injectionType = "delay"
				} else {
					injectionType = "drop"
				}
			})
			logrus.Infof("Wait 5 minutes before checking px status ")
			time.Sleep(5 * time.Minute)
			Step("Verify application after clearing error", func() {
				for _, ctx := range contexts {
					ValidateContext(ctx)
				}
			})
			Step("Check KVDB memebers health", func() {
				nodes := node.GetWorkerNodes()
				kvdbMembers, err := Inst().V.GetKvdbMembers(nodes[0])
				if err != nil {
					err = fmt.Errorf("Error getting kvdb members using node %v. cause: %v", nodes[0].Name, err)
					Expect(err).NotTo(HaveOccurred())
				}
				err = ValidateKVDBMembers(kvdbMembers)
				Expect(err).NotTo(HaveOccurred())
			})
			Step("Check Cloudsnap status ", func() {
				verifyCloudSnaps(contexts)
			})
			Step("Check for crash and verify crash was found before ", func() {
				//TODO need to add this method in future.
			})
			logrus.Infof("Wait  %d minutes before starting next iteration ", errorPersistTimeInMinutes/(time.Minute))
			time.Sleep(errorPersistTimeInMinutes)
			currentTime = time.Now()
		}
		Step("teardown all apps", func() {
			for _, ctx := range contexts {
				TearDownContext(ctx, nil)
			}
		})
	})
	JustAfterEach(func() {
		AfterEachTest(contexts, testrailID, runID)
	})
})

// createSnapshotSchedule creating snapshot schedule
func createSnapshotSchedule(contexts []*scheduler.Context) {
	//Create snapshot schedule
	policyName := "intervalpolicy"
	interval := 30
	for _, ctx := range contexts {
		err := SchedulePolicy(ctx, policyName, interval)
		Expect(err).NotTo(HaveOccurred())
		if strings.Contains(ctx.App.Key, "cloudsnap") {
			appVolumes, err := Inst().S.GetVolumes(ctx)
			if err != nil {
				Expect(err).NotTo(HaveOccurred())
			}
			if len(appVolumes) == 0 {
				err = fmt.Errorf("found no volumes for app %s", ctx.App.Key)
				logrus.Warnf("No appvolumes found")
				Expect(err).NotTo(HaveOccurred())
			}
		}
	}
}

// verifyCloudSnaps check cloudsnaps are taken on scheduled time.
func verifyCloudSnaps(contexts []*scheduler.Context) {
	for _, ctx := range contexts {
		appVolumes, err := Inst().S.GetVolumes(ctx)
		if err != nil {
			logrus.Warnf("Error found while getting volumes %s ", err)
		}
		if len(appVolumes) == 0 {
			err = fmt.Errorf("found no volumes for app %s", ctx.App.Key)
			logrus.Warnf("No appvolumes found")
		}
		//Verify cloudsnap is continuing
		for _, v := range appVolumes {
			if strings.Contains(ctx.App.Key, "cloudsnap") == false {
				logrus.Infof("Apps are not cloudsnap supported %s ", v.Name)
				continue
			}
			// Skip cloud snapshot trigger for Pure DA volumes
			isPureVol, err := Inst().V.IsPureVolume(v)
			if err != nil {
				logrus.Warnf("No pure volumes found in %s ", ctx.App.Key)
			}
			if isPureVol {
				logrus.Warnf("Cloud snapshot is not supported for Pure DA volumes: [%s]", v.Name)
				continue
			}
			err = ValidateSnapshotSchedule(ctx, v)
			Expect(err).NotTo(HaveOccurred())
		}
	}
}

// ValidateKVDBMembers health and availability.
func ValidateKVDBMembers(kvdbMembers map[string]*volume.MetadataNode) error {
	logrus.Infof("Current KVDB members: %v", kvdbMembers)
	if len(kvdbMembers) < 3 {
		err := fmt.Errorf("No KVDB membes to validate or less than 3 members to validate")
		logrus.Warn(err.Error())
		return err
	}
	for id, m := range kvdbMembers {
		if !m.IsHealthy {
			err := fmt.Errorf("kvdb member node: %v is not healthy", id)
			logrus.Warn(err.Error())
			return err
		}
		logrus.Infof("KVDB member node %v is healthy", id)
	}
	return nil
}

// SchedulePolicy
func SchedulePolicy(ctx *scheduler.Context, policyName string, interval int) error {
	if strings.Contains(ctx.App.Key, "cloudsnap") {

		logrus.Infof("APP with cloudsnap key available %v ", ctx.App.Key)
		schedPolicy, err := storkops.GetSchedulePolicy(policyName)
		if err == nil {
			logrus.Infof("schedPolicy is %v already exists", schedPolicy.Name)
		} else {
			//Create snapshot schedule interval.
			retain := 2
			logrus.Infof("Creating a interval schedule policy %v with interval %v minutes", policyName, interval)
			schedPolicy = &storkv1.SchedulePolicy{
				ObjectMeta: meta_v1.ObjectMeta{
					Name: policyName,
				},
				Policy: storkv1.SchedulePolicyItem{
					Interval: &storkv1.IntervalPolicy{
						Retain:          storkv1.Retain(retain),
						IntervalMinutes: interval,
					},
				}}
			_, err = storkops.CreateSchedulePolicy(schedPolicy)
			if err != nil {
				return err
			}
		}
		logrus.Infof("Waiting for 10 mins for Snapshots to be completed")
		time.Sleep(10 * time.Minute)
	}
	return nil
}

// ValidateSnapshotSchedule
func ValidateSnapshotSchedule(ctx *scheduler.Context, appVolume *volume.Volume) error {
	//TODO If cloudsnap is enabled
	snapshotScheduleName := appVolume.Name + "-interval-schedule"
	logrus.Infof("snapshotScheduleName : %v for volume: %s", snapshotScheduleName, appVolume.Name)
	appNamespace := ctx.App.Key + "-" + ctx.UID
	logrus.Infof("Namespace : %v", appNamespace)
	snapStatuses, err := storkops.ValidateSnapshotSchedule(snapshotScheduleName,
		appNamespace,
		snapshotScheduleRetryTimeout,
		snapshotScheduleRetryInterval)
	if err == nil {
		for k, v := range snapStatuses {
			logrus.Infof("Policy Type: %v", k)
			for _, e := range v {
				logrus.Infof("ScheduledVolumeSnapShot Name: %v", e.Name)
				logrus.Infof("ScheduledVolumeSnapShot status: %v", e.Status)
			}
		}
	} else {
		logrus.Infof("Got error while getting volume snapshot status :%v", err.Error())
		return err
	}
	return nil
}

var _ = AfterSuite(func() {
	PerformSystemCheck()
	ValidateCleanup()
})

func TestMain(m *testing.M) {
	// call flag.Parse() here if TestMain uses flags
	ParseFlags()
	os.Exit(m.Run())
}
