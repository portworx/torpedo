package tests

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	"github.com/portworx/sched-ops/k8s/apps"
	"github.com/portworx/torpedo/drivers/node"
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/pkg/log"
	rest "github.com/portworx/torpedo/pkg/restutil"
	"github.com/portworx/torpedo/pkg/testrailuttils"
	. "github.com/portworx/torpedo/tests"
)

const (
	logglyIterateUrl = "https://pxlite.loggly.com/apiv2/events/iterate"
)

var (
	k8sApps = apps.Instance()
)

var _ = Describe("{PodMetricFunctional}", func() {
	var testrailID, runID int
	var contexts []*scheduler.Context
	var namespacePrefix string

	JustBeforeEach(func() {
		runID = testrailuttils.AddRunsToMilestone(testrailID)

		StartTorpedoTest("PodMetricFunctional", "Functional Tests for Pod Metrics", nil, testrailID)
	})

	Context("{PodMetricsSample}", func() {
		namespacePrefix = "podmetricsample"

		// shared test function for pod metric functional tests
		sharedTestFunction := func() {
			It("has to fetch the logs from loggly", func() {
				log.InfoD("Getting cluster ID")
				clusterUUID, err := getClusterID()
				log.FailOnError(err, "Failed to get cluster id data")

				log.InfoD("Fetching logs from loggly")
				meteringData, err := getMeteringData(clusterUUID)
				log.FailOnError(err, "Failed to get metering data")

				log.InfoD("Validating meteringData")
				dash.VerifyFatal(len(meteringData), 0, "there should be no data in loggly initially")

				log.InfoD("Deploy applications")
				contexts = make([]*scheduler.Context, 0)
				for i := 0; i < Inst().GlobalScaleFactor; i++ {
					contexts = append(contexts, ScheduleApplications(fmt.Sprintf("%s-%d", namespacePrefix, i))...)
				}

				log.InfoD("Validate applications")
				ValidateApplications(contexts)

				log.InfoD("Wait for a loggly interval to go through")
				time.Sleep(90 * time.Second)

				// expectedPodHourInMinutes is a rough estimate for the pod hour for deployed application
				expectedPodHourInMinutes := getExpectedPodHourInMinutes(contexts)
				log.InfoD("Final pod hour is %v", expectedPodHourInMinutes)
				log.InfoD("Check metering data is accurate")
				meteringData, err = getMeteringData(clusterUUID)
				log.FailOnError(err, "Failed to get metering data")
				for _, md := range meteringData {
					dash.VerifyFatal(md.ClusterUUID, clusterUUID, "this cluster should have data now")
					dash.VerifyFatal(md.VolumeCount, 3, "there should be three volume records")
				}
			})
		}

		// Sample pod metric tests
		Describe("{SamplePodMetricTest}", func() {
			JustBeforeEach(func() {
				// testrailID =
			})
			sharedTestFunction()
		})

	})

	AfterEach(func() {
		Step("destroy apps", func() {
			log.InfoD("destroying apps")
			if CurrentGinkgoTestDescription().Failed {
				log.InfoD("not destroying apps because the test failed\n")
				return
			}
			for _, ctx := range contexts {
				TearDownContext(ctx, map[string]bool{scheduler.OptionsWaitForResourceLeakCleanup: true})
			}
		})
	})

	AfterEach(func() {
		AfterEachTest(contexts, testrailID, runID)
		defer EndTorpedoTest()
	})
})

// CallhomeData is the latest json format for parsing loggly callhome data
type CallhomeData struct {
	ClusterUUID             string  `json:"cluster_uuid"`
	UsageType               string  `json:"usage_type"`
	StorageNodeCount        int     `json:"storage_node_count"`
	StoragelessNodeCount    int     `json:"storageless_node_count"`
	BaremetalNodeCount      int     `json:"baremetal_node_count"`
	VirtualMachineNodeCount int     `json:"virtual_machine_node_count"`
	VolumeCount             int     `json:"volume_count"`
	PodHour                 float64 `json:"pod_hour"`
	Volumes                 []struct {
		ID        string `json:"id"`
		SizeBytes int    `json:"size_bytes"`
		UsedBytes int    `json:"used_bytes,omitempty"`
		Shared    string `json:"shared"`
	} `json:"volumes"`
	SentToPure1  bool `json:"SentToPure1"`
	SentToLoggly bool `json:"SentToLoggly"`
}

// LogglyPayload is the payload we receive from loggly calls
type LogglyPayload struct {
	Events []*LogglyEvent `json:"events"`
}

// LogglyEvent is an individual metering event
type LogglyEvent struct {
	ID        string   `json:"id"`
	Timestamp int64    `json:"timestamp"`
	Raw       string   `json:"raw"`
	Tags      []string `json:"tags"`
}

func getLogglyData(clusterUUID string, fromTime string) ([]byte, int, error) {
	query := fmt.Sprintf("q=%s&from=%s&until=now", clusterUUID, fromTime)

	logglyToken, ok := os.LookupEnv("LOGGLY_API_TOKEN")
	dash.VerifyFatal(ok, true, "failed to fetch loggly api token")
	headers := make(map[string]string)
	headers["Authorization"] = fmt.Sprintf("Bearer %v", logglyToken)
	return rest.Get(fmt.Sprintf("%v?%v", logglyIterateUrl, query), nil, headers)
}

func getClusterID() (string, error) {
	workerNode := node.GetWorkerNodes()[0]
	clusterID, err := Inst().N.RunCommand(workerNode, fmt.Sprintf("cat %s", "/etc/pwx/cluster_uuid"), node.ConnectionOpts{
		IgnoreError:     false,
		TimeBeforeRetry: defaultRetryInterval,
		Timeout:         defaultTimeout,
		Sudo:            true,
	})
	if err != nil {
		return "", fmt.Errorf("failed to get pxctl status, Err: %v", err)
	}

	return clusterID, nil
}

func getMeteringData(clusterUUID string) ([]*CallhomeData, error) {
	log.InfoD("fetching logs from loggly")

	data, code, err := getLogglyData(clusterUUID, "-2m")
	if err != nil {
		return nil, err
	}
	dash.VerifyFatal(code, 200, fmt.Sprintf("loggly return code %v not equal to 200", code))
	dash.VerifyFatal(len(data) == 0, false, "loggy return empty response")

	log.InfoD("parsing logs from loggly")
	var logglyPayload LogglyPayload
	err = json.Unmarshal(data, &logglyPayload)
	if err != nil {
		return nil, err
	}

	var callhomeEvents []*CallhomeData
	for _, e := range logglyPayload.Events {
		chd := CallhomeData{}
		err = json.Unmarshal([]byte(e.Raw), &chd)
		if err != nil {
			return nil, err
		}
		callhomeEvents = append(callhomeEvents, &chd)
	}

	var meteringData []*CallhomeData
	for _, d := range callhomeEvents {
		if d.UsageType == "meteringData" {
			meteringData = append(meteringData, d)
		}
	}

	return meteringData, nil
}

// getExpectedPodHourInMinutes returns the estimate pod hour given that the metering interval is
// 1 min. it checks a list of volumes and the number of pods using it to esitimate the pod hour.
func getExpectedPodHourInMinutes(contexts []*scheduler.Context) int {
	var expectedPodHourInMinutes int
	for _, ctx := range contexts {
		log.InfoD("getting pod hour for context %v", ctx.App.Key)
		vols, err := Inst().S.GetVolumes(ctx)
		log.FailOnError(err, "Failed to get volumes to check pod hour")
		for _, vol := range vols {
			pods, err := Inst().S.GetPodsForPVC(vol.Name, vol.Namespace)
			log.FailOnError(err, "Failed to get pods from PVC")
			if vol.Shared {
				expectedPodHourInMinutes += len(pods)
				continue
			}
			expectedPodHourInMinutes += 1
		}
	}
	return expectedPodHourInMinutes
}
