package tests

import (
	"encoding/json"
	"errors"
	"fmt"
	"k8s.io/utils/strings/slices"
	"math"
	"math/rand"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	volsnapv1 "github.com/kubernetes-csi/external-snapshotter/client/v6/apis/volumesnapshot/v1"
	snapv1 "github.com/kubernetes-incubator/external-storage/snapshot/pkg/apis/crd/v1"
	"github.com/libopenstorage/openstorage/api"
	opsapi "github.com/libopenstorage/openstorage/api"
	. "github.com/onsi/ginkgo/v2"
	"github.com/portworx/sched-ops/k8s/core"
	csisnapshot "github.com/portworx/sched-ops/k8s/externalsnapshotter"
	"github.com/portworx/sched-ops/k8s/storage"
	"github.com/portworx/sched-ops/task"
	storkv1 "github.com/pure-px/stork/pkg/apis/stork/v1alpha1"
	storkops "github.com/pure-px/stork/pkg/crud/stork"
	"github.com/pure-px/torpedo/drivers/node"
	"github.com/pure-px/torpedo/drivers/scheduler"
	"github.com/pure-px/torpedo/drivers/scheduler/k8s"
	"github.com/pure-px/torpedo/drivers/scheduler/spec"
	"github.com/pure-px/torpedo/drivers/volume"
	"github.com/pure-px/torpedo/drivers/volume/portworx"
	"github.com/pure-px/torpedo/pkg/log"
	"github.com/pure-px/torpedo/pkg/osutils"
	"github.com/pure-px/torpedo/pkg/restutil"
	"github.com/pure-px/torpedo/pkg/testrailuttils"
	"github.com/pure-px/torpedo/pkg/units"
	. "github.com/pure-px/torpedo/tests"
	corev1 "k8s.io/api/core/v1"
	storageApi "k8s.io/api/storage/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	bandwidthMBps = 1
	// buffered BW = 1 MBps with 10% buffer speed in KBps
	bufferedBW = 1130
)
const (
	fio                     = "fio-throttle-io"
	fastpathAppName         = "fastpath"
	fioPVScheduleName       = "tc-cs-volsnapsched"
	fioOutputPVScheduleName = "tc-cs-volsnapsched-2"
	CloudSnapShotClass      = "cloud-snapshotclass"
)

type volumeDataMap struct {
	UsedSize         float64 `json:"UsedSize"`
	PoolID           int     `json:"PoolID"`
	ClusterID        string  `json:"ClusterID"`
	TotalRestoreSize float64 `json:"TotalRestoreSize"`
}

type CloudBackupSizeAPI struct {
	Size                  string `json:"size"`
	TotalDownloadBytes    string `json:"total_download_bytes"`
	CompressedObjectBytes string `json:"compressed_object_bytes"`
	Capacity              string `json:"capacity_required_for_restore"`
}

// Volume replication change
var _ = Describe("{VolumeUpdate}", Label("p0", "positive", "px_vol_ops"), func() {
	var testrailID = 35271
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/cases/view/35271
	var runID int
	JustBeforeEach(func() {
		StartTorpedoTest("VolumeUpdate", "Validate Volume update", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})
	var contexts []*scheduler.Context

	stepLog := "has to schedule apps and update replication factor and size on all volumes of the apps"
	It(stepLog, func() {
		log.InfoD(stepLog)
		var err error
		contexts = make([]*scheduler.Context, 0)
		expReplMap := make(map[*volume.Volume]int64)

		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("volupdate-%d", i))...)
		}

		ValidateApplications(contexts)

		stepLog = "get volumes for all apps in test and update replication factor and size"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for _, ctx := range contexts {
				var appVolumes []*volume.Volume
				stepLog = fmt.Sprintf("get volumes for %s app", ctx.App.Key)
				Step(stepLog, func() {
					log.InfoD(stepLog)
					appVolumes, err = Inst().S.GetVolumes(ctx)
					log.FailOnError(err, "Failed to get volumes for app %s", ctx.App.Key)
					dash.VerifyFatal(len(appVolumes) > 0, true, "App volumes exist?")
				})
				for _, v := range appVolumes {
					MaxRF := Inst().V.GetMaxReplicationFactor()
					MinRF := Inst().V.GetMinReplicationFactor()
					stepLog = fmt.Sprintf("repl decrease volume driver %s on app %s's volume: %v",
						Inst().V.String(), ctx.App.Key, v)
					Step(stepLog,
						func() {
							log.InfoD(stepLog)
							currRep, err := Inst().V.GetReplicationFactor(v)
							log.FailOnError(err, "Failed to get volume  %s repl factor", v.Name)
							expReplMap[v] = int64(math.Max(float64(MinRF), float64(currRep)-1))
							err = Inst().V.SetReplicationFactor(v, currRep-1, nil, nil, true)
							log.FailOnError(err, "Failed to set volume  %s repl factor", v.Name)
							dash.VerifyFatal(err == nil, true, fmt.Sprintf("Set volume  %s repl factor successful ?", v.Name))
						})
					stepLog = fmt.Sprintf("validate successful repl decrease on app %s's volume: %v",
						ctx.App.Key, v)
					Step(stepLog,
						func() {
							log.InfoD(stepLog)
							newRepl, err := Inst().V.GetReplicationFactor(v)
							log.FailOnError(err, "Failed to get volume  %s repl factor", v.Name)
							dash.VerifyFatal(newRepl, expReplMap[v], "Repl factor is as expected ?")
						})
					stepLog = fmt.Sprintf("repl increase volume driver %s on app %s's volume: %v",
						Inst().V.String(), ctx.App.Key, v)
					Step(stepLog,
						func() {
							log.InfoD(stepLog)
							currRep, err := Inst().V.GetReplicationFactor(v)
							log.FailOnError(err, "Failed to get volume  %s repl factor", v.Name)
							// GetMaxReplicationFactory is hardcoded to 3
							// if it increases repl 3 to an aggregated 2 volume, it will fail
							// because it would require 6 worker nodes, since
							// number of nodes required = aggregation level * replication factor
							currAggr, err := Inst().V.GetAggregationLevel(v)
							log.FailOnError(err, "Failed to get volume  %s aggregate level", v.Name)
							if currAggr > 1 {
								MaxRF = int64(len(node.GetStorageDriverNodes())) / currAggr
							}
							expReplMap[v] = int64(math.Min(float64(MaxRF), float64(currRep)+1))
							opts := volume.Options{
								ValidateReplicationUpdateTimeout: validateReplicationUpdateTimeout,
							}
							err = Inst().V.SetReplicationFactor(v, currRep+1, nil, nil, true, opts)
							log.FailOnError(err, "Failed to set volume  %s repl factor", v.Name)
							dash.VerifyFatal(err == nil, true, fmt.Sprintf("Repl factor set succesfully on volume  %s repl", v.Name))
						})
					stepLog = fmt.Sprintf("validate successful repl increase on app %s's volume: %v",
						ctx.App.Key, v)
					Step(stepLog,
						func() {
							newRepl, err := Inst().V.GetReplicationFactor(v)
							log.FailOnError(err, "Failed to get volume  %s repl factor", v.Name)
							dash.VerifyFatal(newRepl, expReplMap[v], "Repl factor is as expected?")
						})
				}
				var requestedVols []*volume.Volume
				stepLog = fmt.Sprintf("increase volume size %s on app %s's volumes: %v from %vGB to %vGB",
					Inst().V.String(), ctx.App.Key, appVolumes, appVolumes[0].Size/units.GiB, (appVolumes[0].Size/units.GiB)+1)
				Step(stepLog,
					func() {
						log.InfoD(stepLog)
						requestedVols, err = Inst().S.ResizeVolume(ctx, Inst().ConfigMap)
						log.FailOnError(err, "Volume resize successful ?")
					})

				stepLog = fmt.Sprintf("validate successful volume size increase on app %s's volumes: %v from %vGB to %vGB",
					ctx.App.Key, appVolumes, appVolumes[0].Size/units.GiB, (appVolumes[0].Size/units.GiB)+1)
				Step(stepLog,
					func() {
						log.InfoD(stepLog)
						for _, v := range requestedVols {
							// Need to pass token before validating volume
							params := make(map[string]string)
							if Inst().ConfigMap != "" {
								params["auth-token"], err = Inst().S.GetTokenFromConfigMap(Inst().ConfigMap)
								log.FailOnError(err, "Failed to get token from configMap")
							}
							err := Inst().V.ValidateUpdateVolume(v, params)
							dash.VerifyFatal(err, nil, "Validate volume update successful?")
						}
					})

			}
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
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})
})

// Volume IO Throttle change
var _ = Describe("{VolumeIOThrottle}", Label("p0", "positive", "px_vol_ops", "Throttling"), func() {
	var contexts []*scheduler.Context
	var namespace string
	var speedBeforeUpdate, speedAfterUpdate int
	var testrailID = 58504
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/cases/view/58504
	var runID int
	JustBeforeEach(func() {
		StartTorpedoTest("VolumeIOThrottle", "Validate volume IO throttle", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})
	stepLog := "has to schedule IOPs and limit them to a max bandwidth"
	It(stepLog, func() {
		log.InfoD(stepLog)
		contexts = make([]*scheduler.Context, 0)
		var err error
		taskNamePrefix := "io-throttle"
		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			taskName := fmt.Sprintf("%s-%d", taskNamePrefix, i)
			log.Debugf("Task name %s\n", taskName)
			appContexts := ScheduleApplications(taskName)
			contexts = append(contexts, appContexts...)
			namespace = appContexts[0].ScheduleOptions.Namespace
		}
		ValidateApplications(contexts)
		stepLog = "get the BW for volume without limiting bandwidth"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			log.Infof("waiting for 5 sec for the pod to stablize")
			time.Sleep(5 * time.Second)
			speedBeforeUpdate, err = Inst().S.GetIOBandwidth(fio, namespace)
			log.FailOnError(err, "Failed to get IO Bandwidth")
		})
		log.InfoD("BW before update %d", speedBeforeUpdate)

		stepLog = "updating the BW"
		Step(stepLog, func() {
			for _, ctx := range contexts {
				var appVolumes []*volume.Volume
				stepLog = fmt.Sprintf("get volumes for %s app", ctx.App.Key)
				Step(stepLog, func() {
					appVolumes, err = Inst().S.GetVolumes(ctx)
					log.FailOnError(err, "Failed to get volumes for app %s", ctx.App.Key)
					dash.VerifyFatal(len(appVolumes) > 0, true, "App volumes exist?")
				})
				log.InfoD("Volumes to be updated %s", appVolumes)
				for _, v := range appVolumes {
					err := Inst().V.SetIoBandwidth(v, bandwidthMBps, bandwidthMBps)
					log.FailOnError(err, "Failed to set IO bandwidth")
				}
			}
		})
		log.InfoD("waiting for the FIO to reduce the speed to take into account the IO Throttle")
		time.Sleep(60 * time.Second)
		stepLog = "get the BW for volume after limiting bandwidth"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			speedAfterUpdate, err = Inst().S.GetIOBandwidth(fio, namespace)
			log.FailOnError(err, "Failed to get IO bandwidth after update")
		})
		log.InfoD("BW after update %d", speedAfterUpdate)
		stepLog = "Validate speed reduction"
		Step(stepLog, func() {
			// We are setting the BW to 1 MBps so expecting the returned value to be in 10% buffer
			log.InfoD(stepLog)
			dash.VerifyFatal(speedAfterUpdate < bufferedBW, true, "Speed reduced below the buffer?")
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
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})
})

// Volume replication change
var _ = Describe("{VolumeUpdateForAttachedNode}", Label("p0", "positive", "px_vol_ops"), func() {
	var testrailID = 58838
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/cases/view/58838
	var runID int
	JustBeforeEach(func() {
		StartTorpedoTest("VolumeUpdateForAttachedNode", "Validate volume update for the attached node", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})
	var contexts []*scheduler.Context

	stepLog := "has to schedule apps and update replication factor for attached node"

	It(stepLog, func() {
		log.InfoD(stepLog)
		var err error
		contexts = make([]*scheduler.Context, 0)
		expReplMap := make(map[*volume.Volume]int64)

		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("volupdate-%d", i))...)
		}

		ValidateApplications(contexts)
		stepLog = "get volumes for all apps in test and update replication factor and size"
		Step(stepLog, func() {
			for _, ctx := range contexts {
				var appVolumes []*volume.Volume
				Step(fmt.Sprintf("get volumes for %s app", ctx.App.Key), func() {
					appVolumes, err = Inst().S.GetVolumes(ctx)
					log.FailOnError(err, "Failed to get volumes for app %s", ctx.App.Key)
					dash.VerifyFatal(len(appVolumes) > 0, true, "App volumes exist ?")
				})
				for _, v := range appVolumes {
					if _, ok := v.Labels[k8s.PureDAVolumeLabel]; ok {
						// This is a Pure Direct Access volume, which will not support repl updates.
						// Ensure the command fails.
						Step(fmt.Sprintf("ensure repl update fails on Pure Direct Access on app %s's volume: %v", ctx.App.Key, v), func() {
							appNodes, err := Inst().S.GetNodesForApp(ctx)
							log.FailOnError(err, "Failed to get nodes for app %s", ctx.App.Key)
							dash.VerifyFatal(len(appNodes) > 0, true, "App nodes exist ?")

							err = Inst().V.SetReplicationFactor(v, 3, []string{appNodes[0].VolDriverNodeID}, nil, true)
							dash.VerifyFatal(err != nil, true, "Repl update failed (as expected) for Pure DA volumes?")
							dash.VerifyFatal(strings.Contains(err.Error(), "not supported for Pure"), true, "Repl update error for Pure DA volumes contains proper string?")
						})
						continue
					}

					MaxRF := Inst().V.GetMaxReplicationFactor()
					MinRF := Inst().V.GetMinReplicationFactor()
					currReplicaSet := []string{}
					updateReplicaSet := []string{}
					expectedReplicaSet := []string{}
					stepLog = fmt.Sprintf("repl decrease volume driver %s on app %s's volume: %v",
						Inst().V.String(), ctx.App.Key, v)
					Step(stepLog,
						func() {
							log.InfoD(stepLog)
							currRep, err := Inst().V.GetReplicationFactor(v)
							log.FailOnError(err, "Failed to get vol %s repl factor", v.Name)
							attachedNode, err := Inst().V.GetNodeForVolume(v, defaultCommandTimeout, defaultCommandRetry)
							log.FailOnError(err, fmt.Sprintf("Failed to get node for vol %s", v.Name))

							replicaSets, err := Inst().V.GetReplicaSets(v)
							log.FailOnError(err, "Failed to get vol %s replica sets", v.Name)
							dash.VerifyFatal(len(replicaSets) > 0, true, fmt.Sprintf("Validate vol %s has replica sets", v.Name))

							for _, nID := range replicaSets[0].Nodes {
								currReplicaSet = append(currReplicaSet, nID)
							}

							log.InfoD("ReplicaSet of volume %v is: %v", v.Name, currReplicaSet)
							log.InfoD("Volume %v is attached to : %v", v.Name, attachedNode.Id)

							for _, n := range currReplicaSet {
								if n == attachedNode.Id {
									updateReplicaSet = append(updateReplicaSet, n)
								} else {
									expectedReplicaSet = append(expectedReplicaSet, n)
								}
							}

							if len(updateReplicaSet) == 0 {
								log.InfoD("Attached node in not part of ReplicatSet, choosing a random node part of set for setting replication factor")
								updateReplicaSet = append(updateReplicaSet, expectedReplicaSet[0])
								expectedReplicaSet = expectedReplicaSet[1:]
							}

							expReplMap[v] = int64(math.Max(float64(MinRF), float64(currRep)-1))
							err = Inst().V.SetReplicationFactor(v, currRep-1, updateReplicaSet, nil, true)
							log.FailOnError(err, "Failed to set repl factor")
							dash.VerifyFatal(err == nil, true, "Repl factor set successfully?")
						})
					stepLog = fmt.Sprintf("validate successful repl decrease on app %s's volume: %v",
						ctx.App.Key, v)
					Step(stepLog,
						func() {
							log.InfoD(stepLog)
							newRepl, err := Inst().V.GetReplicationFactor(v)
							log.InfoD("Got repl factor after update: %v", newRepl)
							log.FailOnError(err, "Failed to get vol %s repl factor", v.Name)
							dash.VerifyFatal(newRepl, expReplMap[v], "New repl factor is as expected?")
							currReplicaSets, err := Inst().V.GetReplicaSets(v)
							log.FailOnError(err, "Failed to get vol %s replica sets", v.Name)
							dash.VerifyFatal(len(currReplicaSet) > 0, true, fmt.Sprintf(" Vol %s repl sets exist?", v.Name))
							reducedReplicaSet := []string{}
							for _, nID := range currReplicaSets[0].Nodes {
								reducedReplicaSet = append(reducedReplicaSet, nID)
							}

							log.InfoD("ReplicaSet of volume %v is: %v", v.Name, reducedReplicaSet)
							log.InfoD("Expected ReplicaSet of volume %v is: %v", v.Name, expectedReplicaSet)
							res := reflect.DeepEqual(reducedReplicaSet, expectedReplicaSet)
							dash.VerifyFatal(res, true, "Reduced replica set is as expected?")
						})
					for _, ctx := range contexts {
						ctx.SkipVolumeValidation = true
					}
					ValidateApplications(contexts)
					for _, ctx := range contexts {
						ctx.SkipVolumeValidation = false
					}
					stepLog = fmt.Sprintf("repl increase volume driver %s on app %s's volume: %v",
						Inst().V.String(), ctx.App.Key, v)
					Step(stepLog,
						func() {
							log.InfoD(stepLog)
							currRep, err := Inst().V.GetReplicationFactor(v)
							log.FailOnError(err, "Failed to get vol %s repl factor", v.Name)
							// GetMaxReplicationFactory is hardcoded to 3
							// if it increases repl 3 to an aggregated 2 volume, it will fail
							// because it would require 6 worker nodes, since
							// number of nodes required = aggregation level * replication factor
							currAggr, err := Inst().V.GetAggregationLevel(v)
							log.FailOnError(err, "Failed to get vol %s aggregation level", v.Name)
							if currAggr > 1 {
								MaxRF = int64(len(node.GetStorageDriverNodes())) / currAggr
							}
							expReplMap[v] = int64(math.Min(float64(MaxRF), float64(currRep)+1))
							opts := volume.Options{
								ValidateReplicationUpdateTimeout: validateReplicationUpdateTimeout,
							}
							err = Inst().V.SetReplicationFactor(v, currRep+1, updateReplicaSet, nil, true, opts)
							log.FailOnError(err, "Failed to set vol %s repl factor", v.Name)
							dash.VerifyFatal(err == nil, true, fmt.Sprintf("Vol %s repl factor set as expected?", v.Name))
						})
					stepLog = fmt.Sprintf("validate successful repl increase on app %s's volume: %v",
						ctx.App.Key, v)
					Step(stepLog,
						func() {
							log.InfoD(stepLog)
							newRepl, err := Inst().V.GetReplicationFactor(v)
							log.FailOnError(err, "Failed to get vol %s repl factor", v.Name)
							dash.VerifyFatal(newRepl, expReplMap[v], "New repl factor is as expected?")
							currReplicaSets, err := Inst().V.GetReplicaSets(v)
							log.FailOnError(err, "Failed to get vol %s replica sets", v.Name)
							dash.VerifyFatal(len(currReplicaSets) > 0, true, "New repl factor is as expected?")
							dash.VerifyFatal(len(currReplicaSet) > 0, true, fmt.Sprintf("Vol %s repl sets exist?", v.Name))
							increasedReplicaSet := []string{}
							for _, nID := range currReplicaSets[0].Nodes {
								increasedReplicaSet = append(increasedReplicaSet, nID)
							}

							log.InfoD("ReplicaSet of volume %v is: %v", v.Name, increasedReplicaSet)
							res := reflect.DeepEqual(increasedReplicaSet, currReplicaSet)
							dash.VerifyFatal(res, true, "Validate increased replica set is as expected")
						})
					ValidateApplications(contexts)
				}

			}
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
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})
})

// Volume replication change
var _ = Describe("{CreateLargeNumberOfVolumes}", Label("p0", "positive", "px_vol_ops", "MiniScale"), func() {
	var testrailID = 0
	// JIRA ID :https://portworx.atlassian.net/browse/PWX-26820
	var runID int
	JustBeforeEach(func() {
		StartTorpedoTest("CreateLargeNumberOfVolumes", "Volumes more than 684 went into down state after creation", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})
	var contexts []*scheduler.Context
	var totalVolumesToCreate = 700
	var maxVolumesToAttach = 100
	var volumesCurrentlyAttached = 0
	var newVolumeIDs []string
	var attachedVolumes []string
	terminate := false

	stepLog := "has to schedule apps and update replication factor for attached node"
	It(stepLog, func() {
		log.InfoD(stepLog)
		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("createmaxvolume-%d", i))...)
		}
		deleteVolumes := func() {
			terminate = true
			for _, each := range newVolumeIDs {
				if IsVolumeExits(each) {
					log.InfoD(fmt.Sprintf("delete volume [%v]", each))
					log.FailOnError(Inst().V.DetachVolume(each), fmt.Sprintf("Failed to detach volume [%v]", each))
					time.Sleep(500 * time.Millisecond)
					log.FailOnError(Inst().V.DeleteVolume(each), fmt.Sprintf("Delete volume with ID [%v] failed", each))
				}
			}
		}

		ValidateApplications(contexts)
		defer appsValidateAndDestroy(contexts)
		defer deleteVolumes()

		// Get list of all volumes present in the cluster
		log.InfoD("Listing all the volumes present in the cluster")
		allVolumeIds, err := Inst().V.ListAllVolumes()

		log.FailOnError(err, "failed to list all the volume")
		log.Info(fmt.Sprintf("total number of volumes present in the cluster [%v]", len(allVolumeIds)))

		if len(allVolumeIds) >= totalVolumesToCreate {
			log.FailOnError(fmt.Errorf("exceeded total volume count limit.. exiting [%d]", len(allVolumeIds)),
				"Total volume count exceeded ")
		}

		// Get Total number of already attached volumes
		for _, each := range allVolumeIds {
			vol, err := Inst().V.InspectVolume(each)
			log.FailOnError(err, "inspect returned error ?")
			if vol.State.String() == "VOLUME_STATE_ATTACHED" {
				volumesCurrentlyAttached = volumesCurrentlyAttached + 1
			}
		}

		// Run inspect continuously in the background
		log.InfoD("start attach volume in the backend while more than 100 volumes got created")
		go func(volumeIds []string) {
			defer GinkgoRecover()
			attachedCount := 0
			for {
				if terminate == true {
					break
				}
				if len(newVolumeIDs) > 100 {
					for _, each := range newVolumeIDs {
						if attachedCount < (maxVolumesToAttach - volumesCurrentlyAttached) {
							_, err := Inst().V.AttachVolume(each)
							log.FailOnError(err, "attaching volume failed")
							attachedCount += 1
							attachedVolumes = append(attachedVolumes, each)
							time.Sleep(2 * time.Second)
						}
					}
				}
			}
		}(newVolumeIDs)

		volumesToBeCreated := totalVolumesToCreate - len(allVolumeIds)
		log.InfoD(fmt.Sprintf("Total number of new volumes to be created in the cluster [%v]", volumesToBeCreated))

		// Create volumes in the cluster till it reaches maximum count
		for initVol := 0; initVol < volumesToBeCreated; initVol++ {
			id := uuid.New()
			volName := fmt.Sprintf("volume_%s", id.String()[:8])
			log.InfoD(fmt.Sprintf("Volume [%v] will be created with name [%v]", initVol, volName))

			// get size of the volume from size 1GiB till 100GiB
			minSize := 1
			maxSize := 100
			randSize := uint64(rand.Intn(maxSize-minSize) + minSize)

			// Pick HA Update from 1 to 3
			haUpdate := int64(rand.Intn(3-1) + 1)

			volId, err := Inst().V.CreateVolume(volName, randSize, haUpdate)
			if err != nil {
				terminate = true
				log.FailOnError(err, fmt.Sprintf("Failed to create volume with vol Name [%v]", volName))
			}
			log.InfoD("Volume Created with ID [%v]", volId)
			newVolumeIDs = append(newVolumeIDs, volId)
		}

		// Validate Volume Attached status
		for _, eachVol := range attachedVolumes {
			vol, err := Inst().V.InspectVolume(eachVol)
			log.FailOnError(err, fmt.Sprintf("Inspect volume failed on volume [%v]", eachVol))
			dash.VerifyFatal(vol.State.String() == "VOLUME_STATE_ATTACHED", true,
				fmt.Sprintf(" volume [%v] state is [%v]", eachVol, vol.State.String()))
		}
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})

})

// Volume replication change
var _ = Describe("{CreateDeleteVolumeKillKVDBMaster}", Label("p1", "negative", "px_vol_ops", "error_injection", "kvdb_ops", "KVDBFailover"), func() {
	var testrailID = 0
	// JIRA ID :https://portworx.atlassian.net/browse/PTX-17728
	var runID int
	JustBeforeEach(func() {
		StartTorpedoTest("CreateDeleteVolumeKillKVDBMaster",
			"Create Delete volume in loop kill kvdb master node in random intervals", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})
	var contexts []*scheduler.Context

	stepLog := "Continuously creates and deletes volume while killing kvdb master node"
	It(stepLog, func() {

		var wg sync.WaitGroup
		numGoroutines := 2

		wg.Add(numGoroutines)
		terminate := false

		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("createmaxvolume-%d", i))...)
		}
		ValidateApplications(contexts)
		defer appsValidateAndDestroy(contexts)

		// Kill KVDB Master in regular interval
		kvdbMaster, err := GetKvdbMasterNode()
		log.FailOnError(err, "Getting KVDB Master Node details failed")
		log.InfoD("KVDB Master Node is [%v]", kvdbMaster.Name)

		// Go Routine to create volume continuously
		volumesCreated := []string{}

		stopRoutine := func() {
			if !terminate {
				terminate = true
				for _, each := range volumesCreated {
					if IsVolumeExits(each) {
						log.FailOnError(Inst().V.DeleteVolume(each), "volume deletion failed on the cluster with volume ID [%s]", each)
					}
				}
			}
		}
		defer stopRoutine()

		go func() {
			defer wg.Done()
			defer GinkgoRecover()
			for {
				if terminate {
					break
				}
				// Volume create continuously
				uuidObj := uuid.New()
				VolName := fmt.Sprintf("volume_%s", uuidObj.String())
				Size := uint64(rand.Intn(100) + 1)  // Size of the Volume between 1G to 100G
				haUpdate := int64(rand.Intn(2) + 1) // Size of the HA between 1 and 3

				volId, err := Inst().V.CreateVolume(VolName, Size, haUpdate)
				if err != nil {
					stopRoutine()
					log.FailOnError(err, "volume creation failed on the cluster with volume name [%s]", VolName)
				}

				volumesCreated = append(volumesCreated, volId)
			}
		}()

		// Go Routine to delete volume continuously in parallel to volume create
		go func() {
			defer wg.Done()
			defer GinkgoRecover()
			for {
				if terminate {
					break
				}
				if len(volumesCreated) > 5 {
					deleteVolume := volumesCreated[0]

					if IsVolumeExits(deleteVolume) {
						err := Inst().V.DeleteVolume(deleteVolume)
						if err != nil {
							stopRoutine()
							log.FailOnError(err,
								"volume deletion failed on the cluster with volume ID [%s]", deleteVolume)
						}
					}

					// Remove the first element
					for i := 0; i < len(volumesCreated)-1; i++ {
						volumesCreated[i] = volumesCreated[i+1]
					}
					// Resize the array by truncating the last element
					volumesCreated = volumesCreated[:len(volumesCreated)-1]
				}
			}

		}()

		// Run KVDB Master Terminate / Volume Create / Delete continuously in parallel for latest one hour
		for i := 0; i < 10; i++ {
			// Wait for KVDB Members to be online
			log.FailOnError(WaitForKVDBMembers(), "failed waiting for KVDB members to be active")

			// Kill KVDB Master Node
			masterNode, err := GetKvdbMasterNode()
			log.FailOnError(err, "failed getting details of KVDB master node")

			// Get KVDB Master PID
			pid, err := GetKvdbMasterPID(*masterNode)
			log.FailOnError(err, "failed getting PID of KVDB master node")

			log.InfoD("KVDB Master is [%v] and PID is [%v]", masterNode.Name, pid)

			// Kill kvdb master PID for regular intervals
			log.FailOnError(KillKvdbMemberUsingPid(*masterNode), "failed to kill KVDB Node")

			// Wait for some time after killing kvdb master Node
			time.Sleep(5 * time.Minute)
		}
		terminate = true
		wg.Wait()

	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})

})

var _ = Describe("{VolumeMultipleHAIncreaseVolResize}", Label("p1", "positive", "px_vol_ops", "MiniScale", "HA_Increase_Decrease"), func() {
	var testrailID = 0
	/*  Try Volume resize to 5 GB every time
	        Try HA Refactor of the volume
	        Try one HA node Reboot

	    	all the above 3 operations are done in parallel
	*/
	// JIRA ID :https://portworx.atlassian.net/browse/PWX-27123
	var runID int
	JustBeforeEach(func() {
		StartTorpedoTest("VolumeMultipleHAIncreaseVolResize",
			"Px crashes when we perform multiple HAUpdate in a loop", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})

	stepLog := "Px crashes when we perform multiple HAUpdate in a loop"
	It(stepLog, func() {
		var wg sync.WaitGroup
		var driverNode *node.Node

		volReplMap := make(map[string]int64)

		driverNode = nil

		contexts = make([]*scheduler.Context, 0)

		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("volmulhaupvolr-%d", i))...)
		}
		ValidateApplications(contexts)
		defer appsValidateAndDestroy(contexts)

		// Get a pool with running IO
		poolUUID := pickPoolToResize(contexts, api.SdkStoragePool_RESIZE_TYPE_AUTO, 0)
		log.InfoD("Pool UUID on which IO is running [%s]", poolUUID)

		// Get Node Details of the Pool with IO
		nodeDetail, err := GetNodeWithGivenPoolID(poolUUID)
		log.FailOnError(err, "Failed to get Node Details from PoolUUID [%v]", poolUUID)
		log.InfoD("Pool with UUID [%v] present in Node [%v]", poolUUID, nodeDetail.Name)

		// Get All Volumes from the pool
		volumes, err := GetVolumesFromPoolID(contexts, poolUUID)
		log.FailOnError(err, "Failed to get list of volumes from the poolIDs")

		for _, each := range volumes {
			replFactor, err := Inst().V.GetReplicationFactor(each)
			log.FailOnError(err, "failed to get replication factor for volume [%v]", each.Name)
			volReplMap[each.ID] = replFactor
		}

		terminate := false
		terminateflow := func() {
			terminate = true
		}

		waitTillDriverUp := func() {
			if driverNode != nil {
				err = Inst().V.WaitDriverUpOnNode(*nodeDetail, 10*time.Minute)
				if err != nil {
					terminateflow()
					log.FailOnError(err, fmt.Sprintf("Driver is down on node %s", nodeDetail.Name))
				}
			}
		}

		// Function to Set replication factor on Volume
		setReplOnVolume := func(volName *volume.Volume, replCount int64, waitToFinish bool) error {
			err = Inst().V.SetReplicationFactor(volName, replCount, nil, nil, waitToFinish)
			if err != nil {
				if strings.Contains(fmt.Sprintf("%v", err), "Another HA increase operation is in progress") {
					return nil
				} else if strings.Contains(fmt.Sprintf("%v", err), "Resource has not been initialized") {
					waitTillDriverUp()
					err = Inst().V.SetReplicationFactor(volName, replCount, nil, nil, waitToFinish)
					if err != nil {
						return err
					}
				} else {
					return err
				}

			}
			return nil
		}

		revertReplica := func() {
			log.Info("Reverting Replica on the volumes")
			waitTillDriverUp()
			for _, each := range volumes {
				for vID, replCount := range volReplMap {
					if each.ID == vID {
						replFactor, err := Inst().V.GetReplicationFactor(each)
						log.FailOnError(err, "failed to get replication factor for volume [%v]", each.Name)
						if replFactor != replCount {
							err := setReplOnVolume(each, replCount, true)
							log.FailOnError(err, "failed to set replication factor for volume [%v]", each.Name)
						}
					}
				}
			}
		}

		defer revertReplica()

		wg.Add(2)
		defer waitTillDriverUp()

		log.InfoD("Initiate Volume resize continuously")
		volumeResize := func(vol *volume.Volume) error {

			apiVol, err := Inst().V.InspectVolume(vol.ID)
			if err != nil {
				terminateflow()
				return err
			}

			curSize := apiVol.Spec.Size
			newSize := curSize + (uint64(5) * units.GiB)
			log.Infof("Initiating volume size increase on volume [%v] by size [%vGB] to [%vGB]",
				vol.ID, curSize/units.GiB, newSize/units.GiB)

			err = Inst().V.ResizeVolume(vol.ID, newSize)
			if err != nil {
				terminateflow()
				return err
			}

			// Wait for 2 seconds for Volume to update stats
			time.Sleep(2 * time.Second)
			volumeInspect, err := Inst().V.InspectVolume(vol.ID)
			if err != nil {
				terminateflow()
				return err
			}

			updatedSize := volumeInspect.Spec.Size
			if updatedSize <= curSize {
				terminateflow()
				return fmt.Errorf("volume did not update from [%vGB] to [%vGB] ",
					curSize/units.GiB, updatedSize/units.GiB)
			}

			return nil
		}

		go func() {
			defer wg.Done()
			defer GinkgoRecover()
			for {
				if terminate {
					break
				}
				for _, eachVol := range volumes {
					err := volumeResize(eachVol)
					if err != nil {
						if strings.Contains(fmt.Sprintf("%v", err), "Resource has not been initialized") {
							waitTillDriverUp()
						} else {
							terminateflow()
							log.FailOnError(err, "failed to resize Volume  [%v]", eachVol.Name)
						}
					}
				}
			}
		}()

		log.InfoD("Trigger test to change replication factor of the volume continuously")
		previousReplFactor := int64(1)
		go func(vol []*volume.Volume) {
			defer wg.Done()
			defer GinkgoRecover()
			for {
				if terminate {
					break
				}

				// Change replication factor of the volume continuously once volume is resized
				for _, each := range vol {
					log.Infof("Changing replication factor of volume [%v]", each.Name)
					setReplFactor := int64(1)
					currRepFactor, err := Inst().V.GetReplicationFactor(each)
					if err != nil {
						terminateflow()
						log.FailOnError(err, "failed to get replication factor for volume [%v]", each.Name)
					}
					// Do HA Update based on current replication factor for the node
					// if repl factor is 3 reduce it by 1
					// if previous repl factor is 3 and current repl factor is 2 reduce it by 1
					// else increase repl factor by 1
					if currRepFactor == 3 {
						setReplFactor = currRepFactor - 1
					} else if currRepFactor == 2 || previousReplFactor == 3 {
						setReplFactor = setReplFactor
					} else {
						setReplFactor = currRepFactor + 1
					}
					previousReplFactor = currRepFactor

					log.Infof("Setting replication factor on volume [%v] from [%v] to [%v]", each.Name, currRepFactor, setReplFactor)
					err = setReplOnVolume(each, setReplFactor, true)
					log.FailOnError(err, "failed to set replication factor for volume after waiting for Px online [%v]", each.Name)
				}
			}

		}(volumes)

		for i := 0; i < 5; i++ {
			// Pick a random volume
			randomIndex := rand.Intn(len(volumes))
			volPicked := volumes[randomIndex]

			// Pick a Node on which volume is placed and start rebooting the node
			poolIds, err := GetPoolIDsFromVolName(volPicked.ID)
			if err != nil {
				terminateflow()
				log.FailOnError(err, "failed to get pool details from the volume")
			}

			// select random pool and get the node associated with that pool
			randomIndex = rand.Intn(len(poolIds))
			poolPicked := poolIds[randomIndex]

			nodeDetail, err := GetNodeWithGivenPoolID(poolPicked)
			if err != nil {
				terminateflow()
				log.FailOnError(err, "error while fetching node details from pool ID")
			}
			log.InfoD("Restarting Px on Node [%v] and waiting for the Px to come back online", nodeDetail.Name)

			driverNode = nodeDetail
			err = Inst().V.RestartDriver(*nodeDetail, nil)
			if err != nil {
				terminateflow()
				log.FailOnError(err, fmt.Sprintf("error restarting px on node %s", nodeDetail.Name))
			}

			waitTillDriverUp()

			// flag is to make sure to wait for driver to be up and running when because
			// of some other process test terminates in middle
			driverNode = nil
		}
		terminateflow()
		wg.Wait()

	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})
})

var _ = Describe("{CloudsnapAndRestore}", Label("p0", "positive", "px_vol_ops", "px_ops", "CloudSnapAndRestore"), func() {
	JustBeforeEach(func() {
		StartTorpedoTest("CloudsnapAndRestore", "Validate cloudsnap creation and restore", nil, 0)
	})

	var contexts []*scheduler.Context
	stepLog := "has to schedule apps, create scheduled cloud snap and restore it"
	It(stepLog, func() {
		log.InfoD(stepLog)
		contexts = make([]*scheduler.Context, 0)
		retain := 8
		interval := 4

		err := CreatePXCloudCredential()
		log.FailOnError(err, "failed to create cloud credential")

		n := node.GetStorageDriverNodes()[0]
		uuidCmd := "pxctl cred list -j | grep uuid"
		output, err := runCmd(uuidCmd, n)
		log.FailOnError(err, "error getting uuid for cloudsnap credential")
		if output == "" {
			log.FailOnError(fmt.Errorf("cloud cred is not created"), "Check for cloud cred exists?")
		}

		credUUID := strings.Split(strings.TrimSpace(output), " ")[1]
		credUUID = strings.ReplaceAll(credUUID, "\"", "")
		log.Infof("Got Cred UUID: %s", credUUID)
		contexts = make([]*scheduler.Context, 0)
		policyName := "intervalpolicy"
		stepLog = fmt.Sprintf("create schedule policy %s", policyName)

		Step(stepLog, func() {
			log.InfoD(stepLog)

			schedPolicy, err := storkops.Instance().GetSchedulePolicy(policyName)
			if err != nil {

				log.InfoD("Creating a interval schedule policy %v with interval %v minutes", policyName, interval)
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

				_, err = storkops.Instance().CreateSchedulePolicy(schedPolicy)
				log.FailOnError(err, fmt.Sprintf("error creating a SchedulePolicy [%s]", policyName))
			}

			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				contexts = append(contexts, ScheduleApplications(fmt.Sprintf("cloudsnaprestore-%d", i))...)
			}

			ValidateApplications(contexts)

		})

		defer func() {
			err := storkops.Instance().DeleteSchedulePolicy(policyName)
			log.FailOnError(err, fmt.Sprintf("error deleting a SchedulePolicy [%s]", policyName))
		}()

		stepLog = "Verify that cloud snap status"
		Step(stepLog, func() {
			log.InfoD(stepLog)

			for _, ctx := range contexts {
				if !strings.Contains(ctx.App.Key, "cloudsnap") {
					continue
				}
				var appVolumes []*volume.Volume
				var err error
				appNamespace := ctx.App.Key + "-" + ctx.UID
				log.Infof("Namespace: %v", appNamespace)
				stepLog = fmt.Sprintf("Getting app volumes for volume %s", ctx.App.Key)
				Step(stepLog, func() {
					log.InfoD(stepLog)
					appVolumes, err = Inst().S.GetVolumes(ctx)
					log.FailOnError(err, "error getting volumes for [%s]", ctx.App.Key)

					if len(appVolumes) == 0 {
						log.FailOnError(fmt.Errorf("no volumes found for [%s]", ctx.App.Key), "error getting volumes for [%s]", ctx.App.Key)
					}
				})
				log.Infof("Got volume count : %v", len(appVolumes))
				scaleFactor := time.Duration(Inst().GlobalScaleFactor * len(appVolumes))
				err = Inst().S.ValidateVolumes(ctx, scaleFactor*4*time.Minute, defaultRetryInterval, nil)
				log.FailOnError(err, "error validating volumes for [%s]", ctx.App.Key)
				for _, v := range appVolumes {
					// Skip cloud snapshot trigger for Pure DA volumes
					isPureVol, err := Inst().V.IsPureVolume(v)
					log.FailOnError(err, "error checking if volume is pure volume")
					if isPureVol {
						log.Warnf("Cloud snapshot is not supported for Pure DA volumes: [%s],Skipping cloud snapshot trigger for pure volume.", v.Name)
						continue
					}

					snapshotScheduleName := v.Name + "-interval-schedule"
					log.InfoD("snapshotScheduleName : %v for volume: %s", snapshotScheduleName, v.Name)

					resp, err := storkops.Instance().GetSnapshotSchedule(snapshotScheduleName, appNamespace)
					log.FailOnError(err, fmt.Sprintf("error getting snapshot schedule for [%s], volume:[%s] in namespace [%s]", snapshotScheduleName, v.Name, v.Namespace))
					dash.VerifyFatal(len(resp.Status.Items) > 0, true, fmt.Sprintf("verify snapshots exists for [%s]", snapshotScheduleName))
					for _, snapshotStatuses := range resp.Status.Items {
						if len(snapshotStatuses) > 0 {
							status := snapshotStatuses[len(snapshotStatuses)-1]
							if status == nil {
								log.FailOnError(fmt.Errorf("SnapshotSchedule has an empty migration in it's most recent status"), fmt.Sprintf("error getting latest snapshot status for [%s]", snapshotScheduleName))
							}
							status, err = WaitForSnapShotToReady(snapshotScheduleName, status.Name, appNamespace)
							log.Infof("Snapshot [%s] has status [%v]", status.Name, status.Status)
							if status.Status == snapv1.VolumeSnapshotConditionError {
								resp, _ := storkops.Instance().GetSnapshotSchedule(snapshotScheduleName, appNamespace)
								log.Infof("SnapshotSchedule resp: %+v", resp)
								snapData, _ := Inst().S.GetSnapShotData(ctx, status.Name, appNamespace)
								if snapData != nil {
									log.Infof("snapData : %v", snapData)
								}

								volumeSnapshot, err := Inst().S.GetSnapShot(ctx, status.Name, appNamespace)
								if err != nil {
									log.Errorf("Error getting volume snapshot [%s] in namespace [%s]. Error: [%v]", status.Name, appNamespace, err)
								}
								if volumeSnapshot != nil {
									log.Errorf("volumeSnapshot : %+v", volumeSnapshot)
								}
								log.FailOnError(fmt.Errorf("snapshot: %s failed. status: [%v]", status.Name, status.Status), fmt.Sprintf("cloud snapshot for [%s] failed", snapshotScheduleName))
							}
							if status.Status == snapv1.VolumeSnapshotConditionPending {
								log.FailOnError(fmt.Errorf("snapshot: %s not completed. status: [%v]", status.Name, status.Status), fmt.Sprintf("cloud snapshot for [%s] stuck in pending state", snapshotScheduleName))
							}
							if status.Status == snapv1.VolumeSnapshotConditionReady {
								snapData, err := Inst().S.GetSnapShotData(ctx, status.Name, appNamespace)
								log.FailOnError(err, fmt.Sprintf("error getting snapshot data for [%s/%s]", appNamespace, status.Name))

								snapType := snapData.Spec.PortworxSnapshot.SnapshotType
								log.Infof("Snapshot Type: %v", snapType)
								if snapType != "cloud" {
									err = &scheduler.ErrFailedToGetVolumeParameters{
										App:   ctx.App,
										Cause: fmt.Sprintf("Snapshot Type: [%s] does not match", snapType),
									}
									log.FailOnError(err, fmt.Sprintf("error validating snapshot data for [%s/%s]", appNamespace, status.Name))
								}
								condition := snapData.Status.Conditions[0]
								dash.VerifyFatal(condition.Type == snapv1.VolumeSnapshotDataConditionReady, true, fmt.Sprintf("validate volume snapshot condition data for [%s] expteced: [%v], actual [%v]", status.Name, snapv1.VolumeSnapshotDataConditionReady, condition.Type))

								snapID := snapData.Spec.PortworxSnapshot.SnapshotID
								log.Infof("Snapshot ID: %v", snapID)
								if snapData.Spec.VolumeSnapshotDataSource.PortworxSnapshot == nil ||
									len(snapData.Spec.VolumeSnapshotDataSource.PortworxSnapshot.SnapshotID) == 0 {
									err = &scheduler.ErrFailedToGetVolumeParameters{
										App:   ctx.App,
										Cause: fmt.Sprintf("volumesnapshotdata: %s does not have portworx volume source set", snapData.Metadata.Name),
									}
									log.FailOnError(err, fmt.Sprintf("error validating snapshot data for [%s/%s]", appNamespace, status.Name))
								}
							}
						}
					}
				}
			}
		})

		stepLog = "Validating cloud snapshot backup size values"
		Step(stepLog, func() {
			for _, ctx := range contexts {
				if !strings.Contains(ctx.App.Key, "cloudsnap") {
					continue
				}
				// Validate the cloud snapshot backup size values [PTX-17342]
				log.Infof("Validating cloud snapshot backup size values for app [%s]", ctx.App.Key)
				vols, err := Inst().S.GetVolumeParameters(ctx)
				log.FailOnError(err, fmt.Sprintf("error getting volume params for [%s]", ctx.App.Key))
				for vol, params := range vols {
					dash.VerifyFatal(validateCloudSnapValues(credUUID, vol, params), true, fmt.Sprintf("validate cloud snap values for volume [%s]", vol))
				}
			}
		})

		isdmthin, err := IsDMthin()
		log.FailOnError(err, fmt.Sprintf("failed while checking for cluster type"))
		if !isdmthin {
			stepLog = "Update volume io_profiles on all volumes"
			Step(stepLog, func() {
				for _, ctx := range contexts {

					appVols, err := Inst().S.GetVolumes(ctx)
					log.FailOnError(err, "error getting volumes for [%s]", ctx.App.Key)

					for _, v := range appVols {
						var volumeSpec *api.VolumeSpecUpdate
						inspectVolume, err := Inst().V.InspectVolume(v.ID)
						log.FailOnError(err, fmt.Sprintf("error inspecting volume %s", v.ID))
						newIOProfile := api.IoProfile_IO_PROFILE_JOURNAL

						if inspectVolume.DerivedIoProfile != api.IoProfile_IO_PROFILE_JOURNAL {
							volumeSpec = &api.VolumeSpecUpdate{IoProfileOpt: &api.VolumeSpecUpdate_IoProfile{IoProfile: newIOProfile}}
						} else {
							newIOProfile = api.IoProfile_IO_PROFILE_AUTO
							volumeSpec = &api.VolumeSpecUpdate{IoProfileOpt: &api.VolumeSpecUpdate_IoProfile{IoProfile: newIOProfile}}
						}

						err = Inst().V.UpdateVolumeSpec(v, volumeSpec)
						log.FailOnError(err, fmt.Sprintf("failed to update io profile to %v for volume %s", newIOProfile, v.ID))
					}

					ctx.SkipVolumeValidation = true
					ValidateContext(ctx)

				}
			})
		}

		stepLog = "Verify cloud snap restore"
		Step(stepLog, func() {
			for _, ctx := range contexts {

				if strings.Contains(ctx.App.Key, "cloudsnap") {

					appNamespace := ctx.App.Key + "-" + ctx.UID
					snapSchedList, err := storkops.Instance().ListSnapshotSchedules(appNamespace)
					log.FailOnError(err, "error getting snapshot list")

					vols, err := Inst().S.GetVolumes(ctx)
					log.FailOnError(err, "error getting volumes")

					for _, vol := range vols {
						var snapshotScheduleName string
						for _, snap := range snapSchedList.Items {
							snapshotScheduleName = snap.Name
							if strings.Contains(snapshotScheduleName, vol.Name) {
								break
							}
						}
						resp, err := storkops.Instance().GetSnapshotSchedule(snapshotScheduleName, appNamespace)
						log.FailOnError(err, "error getting snapshot schedule for [%s] in namespace [%s]", snapshotScheduleName, appNamespace)
						var volumeSnapshotStatus *storkv1.ScheduledVolumeSnapshotStatus
					outer:
						for _, snapshotStatuses := range resp.Status.Items {
							for _, vsStatus := range snapshotStatuses {
								if vsStatus.Status == snapv1.VolumeSnapshotConditionReady {
									volumeSnapshotStatus = vsStatus
									break outer
								}
							}
						}
						if volumeSnapshotStatus != nil {
							restoreSpec := &storkv1.VolumeSnapshotRestore{ObjectMeta: meta_v1.ObjectMeta{
								Name:      vol.Name,
								Namespace: vol.Namespace,
							}, Spec: storkv1.VolumeSnapshotRestoreSpec{SourceName: volumeSnapshotStatus.Name, SourceNamespace: appNamespace, GroupSnapshot: false}}
							restore, err := storkops.Instance().CreateVolumeSnapshotRestore(restoreSpec)
							log.FailOnError(err, "error CreateVolumeSnapshotRestore")
							err = storkops.Instance().ValidateVolumeSnapshotRestore(restore.Name, restore.Namespace, snapshotScheduleRetryTimeout, snapshotScheduleRetryInterval)
							dash.VerifySafely(err, nil, fmt.Sprintf("validate snapshot restore source: %s , destnation: %s in namespace %s", restore.Name, vol.Name, vol.Namespace))
							if err == nil {
								err = storkops.Instance().DeleteVolumeSnapshotRestore(restore.Name, restore.Namespace)
								log.FailOnError(err, "error deleting volume snapshot restore object")
							}
						} else {
							log.FailOnError(fmt.Errorf("no snapshot with Ready status found for vol[%s] in namespace[%s]", vol.Name, vol.Namespace), "error getting volume snapshot")
						}
					}
				}
			}
		})

		stepLog = "Validating apps after cloudsnap restore"
		Step(stepLog, func() {

			for _, ctx := range contexts {
				ctx.ReadinessTimeout = 15 * time.Minute
				//skipping volume validation as ip_profiles are updated
				ctx.SkipVolumeValidation = true
				ValidateContext(ctx)
			}
		})

	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		bucketName, err := GetCloudsnapBucketName(contexts)
		log.FailOnError(err, "error getting cloud snap bucket name")
		opts := make(map[string]bool)
		DestroyApps(contexts, opts)
		DeleteCloudSnapBucket(bucketName)
		AfterEachTest(contexts)
	})
})

var _ = Describe("{LocalsnapAndRestore}", Label("p0", "positive", "px_vol_ops", "px_ops", "CloudSnapAndRestore"), func() {
	JustBeforeEach(func() {
		StartTorpedoTest("LocalsnapAndRestore", "Validate localsnap creation and restore", nil, 0)
	})

	var contexts []*scheduler.Context
	stepLog := "has to schedule apps, create scheduled local snap and restore it"
	It(stepLog, func() {
		log.InfoD(stepLog)
		contexts = make([]*scheduler.Context, 0)
		retain := 8
		interval := 3

		contexts = make([]*scheduler.Context, 0)
		policyName := "localintervalpolicy"
		stepLog = fmt.Sprintf("create schedule policy %s for local snapshots", policyName)

		Step(stepLog, func() {
			log.InfoD(stepLog)

			schedPolicy, err := storkops.Instance().GetSchedulePolicy(policyName)
			if err != nil {

				log.InfoD("Creating a interval schedule policy %v with interval %v minutes", policyName, interval)
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

				_, err = storkops.Instance().CreateSchedulePolicy(schedPolicy)
				log.FailOnError(err, fmt.Sprintf("error creating a SchedulePolicy [%s]", policyName))
			}

			appList := Inst().AppList

			defer func() {

				Inst().AppList = appList

			}()

			Inst().AppList = []string{"fio-localsnap"}

			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				contexts = append(contexts, ScheduleApplications(fmt.Sprintf("localsnaprestore-%d", i))...)
			}

			ValidateApplications(contexts)

		})
		volSnapMap := make(map[string]map[*volume.Volume]*storkv1.ScheduledVolumeSnapshotStatus)

		stepLog = "Verify that local snap status"
		Step(stepLog, func() {
			log.InfoD(stepLog)

			for _, ctx := range contexts {
				var appVolumes []*volume.Volume
				var err error
				appNamespace := ctx.App.Key + "-" + ctx.UID
				log.Infof("Namespace: %v", appNamespace)
				stepLog = fmt.Sprintf("Getting app volumes for volume %s", ctx.App.Key)
				Step(stepLog, func() {
					log.InfoD(stepLog)
					appVolumes, err = Inst().S.GetVolumes(ctx)
					log.FailOnError(err, "error getting volumes for [%s]", ctx.App.Key)

					if len(appVolumes) == 0 {
						log.FailOnError(fmt.Errorf("no volumes found for [%s]", ctx.App.Key), "error getting volumes for [%s]", ctx.App.Key)
					}
				})
				log.Infof("Got volume count : %v", len(appVolumes))
				scaleFactor := time.Duration(Inst().GlobalScaleFactor * len(appVolumes))
				err = Inst().S.ValidateVolumes(ctx, scaleFactor*4*time.Minute, defaultRetryInterval, nil)
				log.FailOnError(err, "error validating volumes for [%s]", ctx.App.Key)
				snapMap := make(map[*volume.Volume]*storkv1.ScheduledVolumeSnapshotStatus)
				for _, v := range appVolumes {

					isPureVol, err := Inst().V.IsPureVolume(v)
					log.FailOnError(err, "error checking if volume is pure volume")
					if isPureVol {
						log.Warnf("Cloud snapshot is not supported for Pure DA volumes: [%s],Skipping cloud snapshot trigger for pure volume.", v.Name)
						continue
					}

					snapshotScheduleName := v.Name + "-interval-schedule"
					log.InfoD("snapshotScheduleName : %v for volume: %s", snapshotScheduleName, v.Name)

					var volumeSnapshotStatus *storkv1.ScheduledVolumeSnapshotStatus
					checkSnapshotSchedules := func() (interface{}, bool, error) {
						resp, err := storkops.Instance().GetSnapshotSchedule(snapshotScheduleName, appNamespace)
						if err != nil {
							return "", false, fmt.Errorf("error getting snapshot schedule for %s, volume:%s in namespace %s", snapshotScheduleName, v.Name, v.Namespace)
						}
						if len(resp.Status.Items) == 0 {
							return "", false, fmt.Errorf("no snapshot schedules found for %s, volume:%s in namespace %s", snapshotScheduleName, v.Name, v.Namespace)
						}

						for _, snapshotStatuses := range resp.Status.Items {
							if len(snapshotStatuses) > 0 {
								volumeSnapshotStatus = snapshotStatuses[len(snapshotStatuses)-1]
								if volumeSnapshotStatus == nil {
									return "", true, fmt.Errorf("SnapshotSchedule has an empty migration in it's most recent status")
								}
								if volumeSnapshotStatus.Status == snapv1.VolumeSnapshotConditionReady {
									return nil, false, nil
								}
								if volumeSnapshotStatus.Status == snapv1.VolumeSnapshotConditionError {
									return nil, false, fmt.Errorf("volume snapshot: %s failed. status: %v", volumeSnapshotStatus.Name, volumeSnapshotStatus.Status)
								}
								if volumeSnapshotStatus.Status == snapv1.VolumeSnapshotConditionPending {
									return nil, true, fmt.Errorf("volume Sanpshot %s is still pending", volumeSnapshotStatus.Name)
								}
							}
						}
						return nil, true, fmt.Errorf("volume Sanpshots for %s is not found", v.Name)
					}
					_, err = task.DoRetryWithTimeout(checkSnapshotSchedules, time.Duration(5*15)*defaultCommandTimeout, defaultReadynessTimeout)
					log.FailOnError(err, "error validating volume snapshot for %s", v.Name)

					snapMap[v] = volumeSnapshotStatus

					snapData, err := Inst().S.GetSnapShotData(ctx, volumeSnapshotStatus.Name, appNamespace)
					log.FailOnError(err, fmt.Sprintf("error getting snapshot data for [%s/%s]", appNamespace, volumeSnapshotStatus.Name))

					snapType := snapData.Spec.PortworxSnapshot.SnapshotType
					log.Infof("Snapshot Type: %v", snapType)
					if snapType != "local" {
						err = &scheduler.ErrFailedToGetVolumeParameters{
							App:   ctx.App,
							Cause: fmt.Sprintf("Snapshot Type: %s does not match", snapType),
						}
						log.FailOnError(err, fmt.Sprintf("error validating snapshot data for [%s/%s]", appNamespace, volumeSnapshotStatus.Name))
					}
					condition := snapData.Status.Conditions[0]
					dash.VerifyFatal(condition.Type == snapv1.VolumeSnapshotDataConditionReady, true, fmt.Sprintf("validate volume snapshot condition data for %s expteced: %v, actual %v", volumeSnapshotStatus.Name, snapv1.VolumeSnapshotDataConditionReady, condition.Type))

					snapID := snapData.Spec.PortworxSnapshot.SnapshotID
					log.Infof("Snapshot ID: %v", snapID)
					if snapData.Spec.VolumeSnapshotDataSource.PortworxSnapshot == nil ||
						len(snapData.Spec.VolumeSnapshotDataSource.PortworxSnapshot.SnapshotID) == 0 {
						err = &scheduler.ErrFailedToGetVolumeParameters{
							App:   ctx.App,
							Cause: fmt.Sprintf("volumesnapshotdata: %s does not have portworx volume source set", snapData.Metadata.Name),
						}
						log.FailOnError(err, fmt.Sprintf("error validating snapshot data for [%s/%s]", appNamespace, volumeSnapshotStatus.Name))
					}

				}
				volSnapMap[appNamespace] = snapMap
			}
			log.Infof("waiting for 10 mins to create multiple local snaps")
			time.Sleep(10 * time.Minute)
		})

		stepLog = "Update volume io_profiles on all volumes"
		Step(stepLog, func() {
			for _, ctx := range contexts {

				appVols, err := Inst().S.GetVolumes(ctx)
				log.FailOnError(err, "error getting volumes for [%s]", ctx.App.Key)

				for _, v := range appVols {
					var volumeSpec *api.VolumeSpecUpdate
					inspectVolume, err := Inst().V.InspectVolume(v.ID)
					log.FailOnError(err, fmt.Sprintf("error inspecting volume %s", v.ID))
					newIOProfile := api.IoProfile_IO_PROFILE_JOURNAL
					if inspectVolume.DerivedIoProfile != api.IoProfile_IO_PROFILE_JOURNAL {
						volumeSpec = &api.VolumeSpecUpdate{IoProfileOpt: &api.VolumeSpecUpdate_IoProfile{IoProfile: newIOProfile}}
					} else {
						newIOProfile = api.IoProfile_IO_PROFILE_AUTO
						volumeSpec = &api.VolumeSpecUpdate{IoProfileOpt: &api.VolumeSpecUpdate_IoProfile{IoProfile: newIOProfile}}
					}
					err = Inst().V.UpdateVolumeSpec(v, volumeSpec)
					log.FailOnError(err, fmt.Sprintf("failed to update io profile to %v for volume %s", newIOProfile, v.ID))
				}

				ctx.SkipVolumeValidation = true
				ValidateContext(ctx)

			}

		})

		stepLog = "Verify local snap restore"
		Step(stepLog, func() {
			for ns, volSnap := range volSnapMap {
				for vol, snap := range volSnap {
					restoreSpec := &storkv1.VolumeSnapshotRestore{ObjectMeta: meta_v1.ObjectMeta{
						Name:      vol.Name,
						Namespace: vol.Namespace,
					}, Spec: storkv1.VolumeSnapshotRestoreSpec{SourceName: snap.Name, SourceNamespace: ns, GroupSnapshot: false}}
					restore, err := storkops.Instance().CreateVolumeSnapshotRestore(restoreSpec)
					log.FailOnError(err, fmt.Sprintf("error creating volume snapshot restore for %s", snap.Name))
					err = storkops.Instance().ValidateVolumeSnapshotRestore(restore.Name, restore.Namespace, time.Duration(5*15)*defaultCommandTimeout, defaultReadynessTimeout)
					dash.VerifyFatal(err, nil, fmt.Sprintf("validate snapshot restore source: %s , destination: %s in namespace %s", restore.Name, vol.Name, vol.Namespace))
				}
			}

		})
		stepLog = "Validating and Destroying apps"
		Step(stepLog, func() {
			for _, ctx := range contexts {
				ctx.SkipVolumeValidation = false
				ctx.ReadinessTimeout = 15 * time.Minute
				ValidateContext(ctx)
				opts := make(map[string]bool)
				opts[SkipClusterScopedObjects] = true
				DestroyApps(contexts, opts)
			}
		})

	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts)
	})
})

var _ = Describe("{CSIOnlyTestCloudSnapshot}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("CSICloudsnapAndRestore", "Validate cloud-snap creation and restore", nil, 0)
	})

	var contexts []*scheduler.Context
	var listRestoredPVC []*corev1.PersistentVolumeClaim
	stepLog := "has to schedule apps, create scheduled cloud snap and restore it"
	It(stepLog, func() {
		log.InfoD(stepLog)
		contexts = make([]*scheduler.Context, 0)
		volumeSnapshotMap := make(map[string]map[string]*volsnapv1.VolumeSnapshot)
		err := CreatePXCloudCredential()
		log.FailOnError(err, "failed to create cloud credential")

		n := node.GetStorageDriverNodes()[0]
		uuidCmd := "pxctl cred list -j | grep uuid"
		output, err := runCmd(uuidCmd, n)
		log.FailOnError(err, "error getting uuid for cloudsnap credential")
		if output == "" {
			log.FailOnError(fmt.Errorf("cloud cred is not created"), "Check for cloud cred exists?")
		}

		credUUID := strings.Split(strings.TrimSpace(output), " ")[1]
		credUUID = strings.ReplaceAll(credUUID, "\"", "")
		log.Infof("Got Cred UUID: %s", credUUID)
		contexts = make([]*scheduler.Context, 0)

		apps := Inst().CsiAppList
		if apps == nil || len(apps) == 0 {
			if Inst().Provisioner == k8s.CsiProvisioner {
				apps = Inst().AppList
			} else {
				log.FailOnError(fmt.Errorf("could not find any CSI Apps to test"), "No CSI apps present to test")
			}
		}
		appList := Inst().AppList
		defer func() {
			Inst().AppList = appList
		}()
		Inst().AppList = apps
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				contexts = append(contexts, ScheduleApplications(fmt.Sprintf("csi-cloudsnaprestore-%d", i))...)
			}
			ValidateApplications(contexts)

		})
		Step(stepLog, func() {
			log.InfoD(stepLog)
			snapShotClassName := CloudSnapShotClass
			volSnapshotClass, err := Inst().S.CreateCSISnapshotClass(scheduler.CSISnapshotClassCreateRequest{
				SnapClassName:  snapShotClassName,
				DeletionPolicy: "Delete",
				Parameters:     map[string]string{"csi.openstorage.org/snapshot-type": "cloud"},
			})
			if err != nil {
				isSnapshotClassExists := strings.Contains(err.Error(), "already exists")
				dash.VerifyFatal(isSnapshotClassExists, true, "Check if snapshot exists")
			} else {
				log.InfoD("Successfully created volume snapshot class: %v", volSnapshotClass.Name)
			}
		})
		stepLog = "Create cloud snapshot and verify status"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for _, ctx := range contexts {
				mapSnapshots := make(map[string]*volsnapv1.VolumeSnapshot)
				response, err := Inst().S.CreateCsiSnapsForVolumes(ctx, CloudSnapShotClass)
				log.FailOnError(err, "Failed to create the snapshots")
				for k, v := range response {
					mapSnapshots[k] = v
				}
				volumeSnapshotMap[ctx.UID] = mapSnapshots
				err = Inst().S.ValidateCsiSnapshots(ctx, response)
				log.FailOnError(err, "Failed to validate the snapshots")
			}
		})

		stepLog = "Validating cloud snapshot backup size values"
		Step(stepLog, func() {
			log.Infof("Validating cloudsnapshot")
			for _, ctx := range contexts {
				// Validate the cloud snapshot backup size values [PTX-17342]
				log.Infof("Validating cloud snapshot backup size values for app [%s]", ctx.App.Key)
				vols, err := Inst().S.GetVolumeParameters(ctx)
				log.FailOnError(err, fmt.Sprintf("error getting volume params for [%s]", ctx.App.Key))
				for vol, params := range vols {
					dash.VerifyFatal(validateCloudSnapValues(credUUID, vol, params), true, fmt.Sprintf("validate cloud snap values for volume [%s]", vol))
				}
			}
		})

		stepLog = "Verify cloud snap restore"
		Step(stepLog, func() {
			for _, ctx := range contexts {
				mapSnapshots := volumeSnapshotMap[ctx.UID]
				log.Infof("Validating cloudsnapshot restore")
				vols, err := Inst().S.GetVolumes(ctx)
				log.FailOnError(err, "error getting volumes")
				for _, vol := range vols {
					snapShot, ok := mapSnapshots[vol.VolumeName]
					if !ok {
						dash.VerifySafely(ok, true, "error getting volume snapshot for volume")
						continue
					}
					log.Infof("Volume snapshot found for volume %s", vol.Name)
					quantity, err := resource.ParseQuantity(strconv.FormatUint(vol.Size, 10))
					log.FailOnError(err, "failed to parse size")
					restoredPVCSpec, err := k8s.GeneratePVCRestoreSpec(quantity, vol.Namespace, vol.Name+"-restore", snapShot.Name, vol.StorageClassName)
					log.FailOnError(err, "failed to build restored PVC Spec")
					log.Infof("Generating PVC from snapshot source snapshot %s, pvc name %s", snapShot.Name, restoredPVCSpec.Name)
					_, err = k8sCore.CreatePersistentVolumeClaim(restoredPVCSpec)
					log.FailOnError(err, "failed to restore PVC")
					listRestoredPVC = append(listRestoredPVC, restoredPVCSpec)
					log.Infof("Validating pvc with name %s after restore", restoredPVCSpec.Name)
					err = k8sCore.ValidatePersistentVolumeClaim(restoredPVCSpec, 4*time.Minute, defaultRetryInterval)
					log.FailOnError(err, "failed to restore PVC")
					validatePodCreationWithPVCName(restoredPVCSpec)
				}
			}
		})

		stepLog = "Delete cloud snapshots"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for _, ctx := range contexts {
				mapSnapshots := volumeSnapshotMap[ctx.UID]
				for _, v := range mapSnapshots {
					err = Inst().S.DeleteCsiSnapshot(ctx, v.Name, v.Namespace)
					log.FailOnError(err, "Failed to delete the snapshots")
				}
			}
		})

		stepLog = "Validating apps after cloud snaphot restore"
		Step(stepLog, func() {

			for _, ctx := range contexts {
				ctx.ReadinessTimeout = 15 * time.Minute
				//skipping volume validation as ip_profiles are updated
				ctx.SkipVolumeValidation = true
				ValidateContext(ctx)
			}
		})

	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		bucketName, err := GetCloudsnapBucketName(contexts)
		log.FailOnError(err, "error getting cloud snap bucket name")
		opts := make(map[string]bool)
		DestroyApps(contexts, opts)
		DeleteCloudSnapBucket(bucketName)
		for _, pvc := range listRestoredPVC {
			err := k8sCore.DeletePersistentVolumeClaim(pvc.Name, pvc.Namespace)
			if err != nil {
				log.Warnf("Could not delete pvc %s", pvc.Name)
			}
		}
		AfterEachTest(contexts)
	})
})

// CSI snapshot tests with invalid credentials
var _ = Describe("{CSIOnlyTestCloudSnapshotInvalidCredentials}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("CSIOnlyTestCloudSnapshotInvalidCredentials", "Test create and restore snapshot with invalid creds", nil, 0)
	})

	var pvcName, ns, snapShotClassName, snapName, scName, secretName string
	context := &scheduler.Context{
		App: &spec.AppSpec{
			Key: "snapshot-invalid-cred-test",
		},
	}
	var pvc *corev1.PersistentVolumeClaim
	stepLog := "has to test with invalid credentials"
	It(stepLog, func() {
		log.InfoD(stepLog)
		err := CreatePXCloudCredential()
		log.FailOnError(err, "failed to create cloud credential")

		stepLog = "Create CSI storage class"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			scName = fmt.Sprintf("storage-class-for-invalid-creds-test")
			createStorageClass(scName, nil)
		})

		ns = fmt.Sprintf("csi-creds-test-ns-%v", time.Now().Unix())
		createNamespace(ns)

		secretName = "cred-secret"
		var data = make(map[string]string)
		data["snapshot-credential-id"] = "invalid"
		metaObj := metav1.ObjectMeta{
			Name:      secretName,
			Namespace: ns,
		}
		obj := &corev1.Secret{
			ObjectMeta: metaObj,
			StringData: data,
		}

		_, err = k8sCore.CreateSecret(obj)
		log.FailOnError(err, fmt.Sprintf("error creating secret [%s] failed [%v]", ns, err))

		stepLog = "Create invalid credentials storage class"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			snapShotClassName = fmt.Sprintf(CloudSnapShotClass+"-invalid-%v", time.Now().Unix())
			createVolumeSnapshotClass(snapShotClassName, map[string]string{
				"csi.storage.k8s.io/snapshotter-secret-name":      "cred-secret",
				"csi.storage.k8s.io/snapshotter-secret-namespace": ns,
				"csi.openstorage.org/snapshot-type":               "cloud",
			})
		})

		stepLog = "Create PVC"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			log.FailOnError(err, fmt.Sprintf("error creating namespace [%s] failed [%v]", ns, err))

			pvcName = fmt.Sprintf("csi-creds-test-%v", time.Now().Unix())
			pvc = create50GiReadWriteOncePVC(pvcName, ns, scName)
		})

		stepLog = "Create cloud-snap with invalid credentials"
		Step(stepLog, func() {
			log.Infof("create cloudsnapshot with invalid credentials")

			_, err := Inst().S.CreateCsiSnapshot(fmt.Sprintf("csi-creds-test-%v", time.Now().Unix()), ns, snapShotClassName, pvcName)
			log.FailOnNoError(err, "snapshot should have failed as credentials were invalid")
		})

		stepLog = "Create valid secret"
		Step(stepLog, func() {
			log.InfoD(stepLog)

			uuid, err := GetPXCloudCredential()
			log.FailOnError(err, "could not get cloudsnap credentials")

			err = Inst().S.DeleteSecret(ns, secretName)
			dash.VerifySafely(err, nil, fmt.Sprintf("Deleting snapshots in namespace [%s]", ns))

			data = make(map[string]string)
			data["snapshot-credential-id"] = uuid
			metaObj := metav1.ObjectMeta{
				Name:      secretName,
				Namespace: ns,
			}
			obj := &corev1.Secret{
				ObjectMeta: metaObj,
				StringData: data,
			}
			_, err = k8sCore.CreateSecret(obj)
			log.FailOnError(err, fmt.Sprintf("error creating secret [%s] failed [%v]", ns, err))
		})

		stepLog = "Create cloud-snap with valid credentials"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			snapName = fmt.Sprintf("csi-creds-test-%v", time.Now().Unix())

			_, err := Inst().S.CreateCsiSnapshot(snapName, ns, snapShotClassName, pvcName)
			log.FailOnError(err, "snapshot should not have failed")
		})

		stepLog = "Delete cloud credentials"
		Step(stepLog, func() {
			log.InfoD(stepLog)

			err := DeletePXCloudCredential()
			log.FailOnError(err, "snapshot should not have failed")
		})

		stepLog = "Restore snapshot should fail"
		Step(stepLog, func() {
			log.InfoD(stepLog)

			restoredPVCSpec, err := k8s.GeneratePVCRestoreSpec(resource.MustParse("5Gi"), pvc.Namespace, pvc.Name+"-restore", snapName, scName)
			log.FailOnError(err, "failed to build restored PVC Spec")
			log.Infof("Generating PVC from snapshot source snapshot %s, pvc name %s", snapName, restoredPVCSpec.Name)
			restoredPvc, err := k8sCore.CreatePersistentVolumeClaim(restoredPVCSpec)
			log.FailOnError(err, "failed to restore PVC")
			err = Inst().S.WaitForSinglePVCToBound(restoredPvc.Name, ns, 3)
			log.FailOnNoError(err, "restore should have failed as cloudsnap creds doesn't exist")
		})

		stepLog = "Delete snapshots, this should use updated credentials"
		Step(stepLog, func() {
			log.InfoD(stepLog)

			err := CreatePXCloudCredential()
			log.FailOnError(err, "could not create cloudsnap credentials")

			uuid, err := GetPXCloudCredential()
			log.FailOnError(err, "could not get cloudsnap credentials")

			err = Inst().S.DeleteSecret(ns, secretName)
			dash.VerifySafely(err, nil, fmt.Sprintf("Deleting snapshots in namespace [%s]", ns))

			// wait for some time for secret to be deleted
			time.Sleep(10 * time.Second)

			data = make(map[string]string)
			data["snapshot-credential-id"] = uuid
			metaObj := metav1.ObjectMeta{
				Name:      secretName,
				Namespace: ns,
			}
			obj := &corev1.Secret{
				ObjectMeta: metaObj,
				StringData: data,
			}
			_, err = k8sCore.CreateSecret(obj)
			log.FailOnError(err, fmt.Sprintf("error creating secret [%s] failed [%v]", ns, err))

			err = Inst().S.DeletePvcsFromNamespace(context, ns)
			dash.VerifySafely(err, nil, fmt.Sprintf("Deleting PVCs in namespace [%s]", ns))

			err = Inst().S.WaitForPvcsToBeDeleted(context, ns)
			dash.VerifySafely(err, nil, fmt.Sprintf("Waiting for PVCs to be deleted in namespace [%s]", ns))

			err = Inst().S.DeleteCsiSnapshot(context, snapName, ns)
			log.FailOnError(err, fmt.Sprintf("error deleting snapshot %s in namespace [%s]", snapName, ns))

			t := func() (interface{}, bool, error) {
				_, err = csisnapshot.Instance().GetSnapshot(snapName, ns)
				if err != nil && k8serrors.IsNotFound(err) {
					return "", false, nil
				}
				return "", true, fmt.Errorf("snapshot is not deleted")
			}
			if _, err := task.DoRetryWithTimeout(t, 5*time.Minute, 10*time.Second); err != nil {
				log.FailOnError(err, fmt.Sprintf("error deleting snapshot %s in namespace [%s]", snapName, ns))
			}
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()

		err = Inst().S.DeleteCsiSnapshotsFromNamespace(context, ns)
		dash.VerifySafely(err, nil, fmt.Sprintf("Deleting snapshots in namespace [%s]", ns))

		err = Inst().S.WaitForSnapshotsToBeDeleted(context, ns)
		dash.VerifySafely(err, nil, fmt.Sprintf("Waiting for snapshots to be deleted in namespace [%s]", ns))

		cleanupSnapshotests(context, ns)
		AfterEachTest(contexts)
	})
})

// Test CSI snapshot when volume is in HA update state
// Write lots of data, increase repl factor, then take snapshot
var _ = Describe("{CSIOnlyTestCloudSnapshotHAUpdateState}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("CSIOnlyTestCloudSnapshotHAUpdateState", "Test create and restore snapshot, volume in HA update", nil, 0)
	})

	var pvcName, ns, snapShotClassName, snapName, scName string
	context := &scheduler.Context{
		App: &spec.AppSpec{
			Key: "snapshot-ha-update-test",
		},
	}
	var pvc *corev1.PersistentVolumeClaim
	var pod *corev1.Pod
	stepLog := "Take snapshots of volume in HA update state mode"
	It(stepLog, func() {
		log.InfoD(stepLog)
		err := CreatePXCloudCredential()
		log.FailOnError(err, "failed to create cloud credential")

		ns = fmt.Sprintf("csi-snapshot-ha-update-test-ns-%v", time.Now().Unix())
		createNamespace(ns)

		stepLog = "Create CSI storage class"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			scName = fmt.Sprintf("csi-storage-class-ha-update")
			createStorageClass(scName, nil)
		})

		stepLog = "Create volume snapshot class"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			snapShotClassName = CloudSnapShotClass
			createVolumeSnapshotClass(snapShotClassName, map[string]string{"csi.openstorage.org/snapshot-type": "cloud"})
		})

		stepLog = "Create PVC"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			pvcName = fmt.Sprintf("csi-snapshot-ha-update-test-%v", time.Now().Unix())
			pvc = create50GiReadWriteOncePVC(pvcName, ns, scName)
		})

		stepLog = "Create Pod"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			pod = createPod(pvc)
		})

		stepLog = "Write data"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			writeCmd := []string{"bash", "-c", fmt.Sprintf("dd if=/dev/urandom of=%s/filename bs=1048576 count=4096", "/usr/share/nginx/html")}
			_, err := k8sCore.RunCommandInPod(writeCmd, pod.Name, pod.Spec.Containers[0].Name, pod.Namespace)
			log.FailOnError(err, "Unable to write data")
		})

		stepLog = "Update HA mode"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			pv := pvc.Spec.VolumeName
			pxNodes, err := GetStorageNodes()
			log.FailOnError(err, "Unable to get the storage nodes")
			pxNode := GetRandomNode(pxNodes)
			pxctlCmdFull := fmt.Sprintf("v ha-update --repl 2 %s", pv)
			_, err = Inst().V.GetPxctlCmdOutput(pxNode, pxctlCmdFull)
			log.FailOnError(err, fmt.Sprintf("error ha-updating legacy shared volume %s", pv))
		})

		stepLog = "Create cloud-snap with valid credentials"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			snapName = fmt.Sprintf("csi-snapshot-ha-update-test-snap-%v", time.Now().Unix())
			_, err := Inst().S.CreateCsiSnapshot(snapName, ns, snapShotClassName, pvcName)
			log.FailOnError(err, "snapshot should not have failed")
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()

		cleanupSnapshotests(context, ns)

		AfterEachTest(contexts)
	})
})

// Test Snapshot restore when PX driver is restarting
var _ = Describe("{CSIOnlyTestCloudSnapshotRestartPX}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("CSIOnlyTestCloudSnapshotRestartPX", "Test create snapshot, with restart px", nil, 0)
	})

	var pvcName, ns, snapShotClassName, snapName, scName string
	context := &scheduler.Context{
		App: &spec.AppSpec{
			Key: "snapshpot-restart-px-test",
		},
	}
	var pvc *corev1.PersistentVolumeClaim
	var pod *corev1.Pod
	stepLog := "Create snapshot, with restart px"
	It(stepLog, func() {
		log.InfoD(stepLog)
		err := CreatePXCloudCredential()
		log.FailOnError(err, "failed to create cloud credential")

		ns = fmt.Sprintf("csi-snapshot-restart-px-test-ns-%v", time.Now().Unix())
		createNamespace(ns)

		stepLog = "Create CSI storage class"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			scName = fmt.Sprintf("csi-storage-class-restart-px")
			createStorageClass(scName, nil)
		})

		stepLog = "Create volume snapshot class"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			snapShotClassName = CloudSnapShotClass
			createVolumeSnapshotClass(snapShotClassName, map[string]string{"csi.openstorage.org/snapshot-type": "cloud"})
		})

		stepLog = "Create PVC"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			pvcName = fmt.Sprintf("csi-snapshot-restart-px-test-%v", time.Now().Unix())
			pvc = create50GiReadWriteOncePVC(pvcName, ns, scName)
		})

		stepLog = "Create Pod"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			pod = createPod(pvc)
		})

		stepLog = "Write data"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			writeCmd := []string{"bash", "-c", fmt.Sprintf("dd if=/dev/urandom of=%s/filename bs=1048576 count=4096", "/usr/share/nginx/html")}
			_, err := k8sCore.RunCommandInPod(writeCmd, pod.Name, pod.Spec.Containers[0].Name, pod.Namespace)
			log.FailOnError(err, "Unable to write data")
		})

		stepLog = "Create cloud-snap, with px restart"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			log.Infof("create cloudsnapshot with valid credentials")
			snapName = fmt.Sprintf("csi-snapshot-restart-px-test-snap-%v", time.Now().Unix())

			v1obj := metav1.ObjectMeta{
				Name:      snapName,
				Namespace: ns,
			}
			source := volsnapv1.VolumeSnapshotSource{
				PersistentVolumeClaimName: &pvcName,
			}
			spec := volsnapv1.VolumeSnapshotSpec{
				VolumeSnapshotClassName: &snapShotClassName,
				Source:                  source,
			}
			snap := volsnapv1.VolumeSnapshot{
				ObjectMeta: v1obj,
				Spec:       spec,
			}
			_, err := csisnapshot.Instance().CreateSnapshot(&snap)
			log.FailOnError(err, "snapshot should not have failed")

			replicaSets, err := Inst().V.GetReplicaSets(&volume.Volume{
				ID: pvc.Spec.VolumeName,
			})
			var nodeForPxStop node.Node
			// Put the volume in Degraded state.
			replicasNodes := replicaSets[0].Nodes
			// Stop Driver on one of the replicas.
			storagenodes, err := GetStorageNodes()
			for _, n := range storagenodes {
				if n.Id == replicasNodes[0] {
					nodeForPxStop = n
					break
				}
			}
			restartPx(nodeForPxStop)

			err = k8s.WaitForCsiSnapToBeReady(snapName, ns)
			log.FailOnError(err, "snapshot should have been completed successfully")
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		cleanupSnapshotests(context, ns)
		AfterEachTest(contexts)
	})
})

// Test CSI snapshots with multiple Snapshot/Restore requests in parallel
var _ = Describe("{CSIOnlyTestCloudSnapshotMultipleSnapshotAndRestore}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("CSIOnlyTestCloudSnapshotMultipleSnapshotAndRestore", "Test create and restore multiple snapshots", nil, 0)
	})

	var ns, snapShotClassName, scName string
	context := &scheduler.Context{
		App: &spec.AppSpec{
			Key: "snapshot-multiple-test",
		},
	}
	stepLog := "Create multiple PVCS, snapshot and restore"
	count := 100
	It(stepLog, func() {
		log.InfoD(stepLog)
		err := CreatePXCloudCredential()
		log.FailOnError(err, "failed to create cloud credential")

		ns = fmt.Sprintf("csi-snapshot-multiple-test-ns-%v", time.Now().Unix())
		createNamespace(ns)

		stepLog = "Create CSI storage class"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			scName = fmt.Sprintf("csi-storage-class-multiple")
			createStorageClass(scName, nil)
		})

		stepLog = "Create volume snapshot class"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			snapShotClassName = CloudSnapShotClass
			createVolumeSnapshotClass(snapShotClassName, map[string]string{"csi.openstorage.org/snapshot-type": "cloud"})
		})

		timeNow := time.Now().Unix()
		stepLog = "Create PVCs"

		wg := sync.WaitGroup{}
		wg.Add(count)
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for i := 0; i < count; i++ {
				num := i
				go func() {
					defer wg.Done()
					defer GinkgoRecover()
					pvcName := fmt.Sprintf("csi-snapshot-multiple-test-%d-%v", num, timeNow)
					create50GiReadWriteOncePVC(pvcName, ns, scName)

				}()
			}
		})
		wg.Wait()

		wg = sync.WaitGroup{}
		wg.Add(count)
		stepLog = "Create snapshots in parallel"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for i := 0; i < count; i++ {
				num := i
				go func() {
					defer wg.Done()
					defer GinkgoRecover()
					pvcName := fmt.Sprintf("csi-snapshot-multiple-test-%d-%v", num, timeNow)
					log.Infof("Create cloudsnapshot with valid credentials for pvc %s", pvcName)
					snapName := fmt.Sprintf("csi-snapshot-multiple-test-snap-%d-%v", num, timeNow)
					_, err := Inst().S.CreateCsiSnapshot(snapName, ns, snapShotClassName, pvcName)
					log.FailOnError(err, "snapshot failed")
				}()
			}
		})
		wg.Wait()

		wg = sync.WaitGroup{}
		wg.Add(count)
		stepLog = "Create restores in parallel"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for i := 0; i < count; i++ {
				num := i
				go func() {
					defer wg.Done()
					defer GinkgoRecover()
					snapName := fmt.Sprintf("csi-snapshot-multiple-test-snap-%d-%v", num, timeNow)
					pvcName := fmt.Sprintf("csi-snapshot-multiple-test-%d-%v", num, timeNow)
					restoredPVCSpec, err := k8s.GeneratePVCRestoreSpec(resource.MustParse("5Gi"), ns, pvcName+"-restore", snapName, scName)
					log.FailOnError(err, "failed to build restored PVC Spec")
					log.Infof("Generating PVC from snapshot source snapshot %s, pvc name %s", snapName, restoredPVCSpec.Name)
					restoredPvc, err := k8sCore.CreatePersistentVolumeClaim(restoredPVCSpec)
					log.FailOnError(err, "failed to restore PVC")
					err = Inst().S.WaitForSinglePVCToBound(restoredPvc.Name, ns, 3)
					log.FailOnError(err, "restore failed with error")
				}()
			}
		})
		wg.Wait()
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
		cleanupSnapshotests(context, ns)
		AfterEachTest(contexts)
	})
})

// Test CSI snapshots when a node is restarted
var _ = Describe("{CSIOnlyTestCloudSnapshotRestartNode}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("CSIOnlyTestCloudSnapshotRestartNode", "Test create snapshot, with restart node", nil, 0)
	})

	var pvcName, ns, snapShotClassName, snapName, scName string
	context := &scheduler.Context{
		App: &spec.AppSpec{
			Key: "snapshot-restart-node-test",
		},
	}
	var pvc *corev1.PersistentVolumeClaim
	var pod *corev1.Pod
	stepLog := "Create snapshot, with restart node"
	It(stepLog, func() {
		log.InfoD(stepLog)
		err := CreatePXCloudCredential()
		log.FailOnError(err, "failed to create cloud credential")

		ns = fmt.Sprintf("csi-snapshot-restart-node-test-ns-%v", time.Now().Unix())
		createNamespace(ns)

		stepLog = "Create CSI storage class"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			scName = fmt.Sprintf("csi-storage-class-restart-node")
			createStorageClass(scName, nil)
		})

		stepLog = "Create volume snapshot class"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			snapShotClassName = CloudSnapShotClass
			createVolumeSnapshotClass(snapShotClassName, map[string]string{"csi.openstorage.org/snapshot-type": "cloud"})
		})

		stepLog = "Create PVC"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			pvcName = fmt.Sprintf("csi-snapshot-restart-node-test-%v", time.Now().Unix())
			pvc = create50GiReadWriteOncePVC(pvcName, ns, scName)
		})

		stepLog = "Create Pod"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			pod = createPod(pvc)
		})

		stepLog = "Write data"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			writeCmd := []string{"bash", "-c", fmt.Sprintf("dd if=/dev/urandom of=%s/filename bs=1048576 count=4096", "/usr/share/nginx/html")}
			_, err := k8sCore.RunCommandInPod(writeCmd, pod.Name, pod.Spec.Containers[0].Name, pod.Namespace)
			log.FailOnError(err, "Unable to write data")
		})

		stepLog = "Create cloud-snap, with px restart"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			log.Infof("create cloudsnapshot with valid credentials")
			snapName = fmt.Sprintf("csi-snapshot-restart-node-test-snap-%v", time.Now().Unix())

			v1obj := metav1.ObjectMeta{
				Name:      snapName,
				Namespace: ns,
			}
			source := volsnapv1.VolumeSnapshotSource{
				PersistentVolumeClaimName: &pvcName,
			}
			spec := volsnapv1.VolumeSnapshotSpec{
				VolumeSnapshotClassName: &snapShotClassName,
				Source:                  source,
			}
			snap := volsnapv1.VolumeSnapshot{
				ObjectMeta: v1obj,
				Spec:       spec,
			}
			_, err := csisnapshot.Instance().CreateSnapshot(&snap)
			log.FailOnError(err, "snapshot should not have failed")

			replicaSets, err := Inst().V.GetReplicaSets(&volume.Volume{
				ID: pvc.Spec.VolumeName,
			})
			var nodeForPxStop node.Node
			// Put the volume in Degraded state.
			replicasNodes := replicaSets[0].Nodes
			// Stop Driver on one of the replicas.
			storagenodes, err := GetStorageNodes()
			for _, n := range storagenodes {
				if n.Id == replicasNodes[0] {
					nodeForPxStop = n
					break
				}
			}
			err = Inst().N.RebootNodeAndWait(nodeForPxStop)
			dash.VerifyFatal(err == nil, true, fmt.Sprintf("Reboot node %s", nodeForPxStop.Name))

			err = Inst().V.WaitDriverUpOnNode(nodeForPxStop, 5*time.Minute)
			dash.VerifyFatal(err == nil, true, fmt.Sprintf("Wait for driver to start"))

			err = k8s.WaitForCsiSnapToBeReady(snapName, ns)
			log.FailOnError(err, "Snapshot creation failed")
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()

		cleanupSnapshotests(context, ns)

		AfterEachTest(contexts)
	})
})

// Test CSI snapshot restore when CSI pods are restarted
var _ = Describe("{CSIOnlyTestCloudSnapshotRestartCSIPods}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("CSIOnlyTestCloudSnapshotRestartCSIPods", "Test create snapshot, with restart csi pods", nil, 0)
	})

	var pvcName, ns, snapShotClassName, snapName, scName string
	context := &scheduler.Context{
		App: &spec.AppSpec{
			Key: "snapshot-restart-csi-pods-test",
		},
	}
	var pvc *corev1.PersistentVolumeClaim
	var pod *corev1.Pod
	stepLog := "Create snapshot with CSI pod restart"
	It(stepLog, func() {
		log.InfoD(stepLog)
		err := CreatePXCloudCredential()
		log.FailOnError(err, "failed to create cloud credential")

		ns = fmt.Sprintf("csi-snapshot-restart-csi-pods-test-ns-%v", time.Now().Unix())
		createNamespace(ns)

		stepLog = "Create CSI storage class"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			scName = fmt.Sprintf("csi-storage-class-restart-csi-pods-snapshot")
			createStorageClass(scName, nil)
		})

		stepLog = "Create volume snapshot class"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			snapShotClassName = CloudSnapShotClass
			createVolumeSnapshotClass(snapShotClassName, map[string]string{"csi.openstorage.org/snapshot-type": "cloud"})
		})

		stepLog = "Create PVC"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			pvcName = fmt.Sprintf("csi-snapshot-restart-csi-pods-%v", time.Now().Unix())
			pvc = create50GiReadWriteOncePVC(pvcName, ns, scName)
		})

		stepLog = "Create Pod"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			pod = createPod(pvc)
		})

		stepLog = "Write data"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			writeCmd := []string{"bash", "-c", fmt.Sprintf("dd if=/dev/urandom of=%s/filename bs=1048576 count=4096", "/usr/share/nginx/html")}
			_, err := k8sCore.RunCommandInPod(writeCmd, pod.Name, pod.Spec.Containers[0].Name, pod.Namespace)
			log.FailOnError(err, "Unable to write data")
		})

		stepLog = "Create cloud-snap, with px restart"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			log.Infof("create cloudsnapshot with valid credentials")
			snapName = fmt.Sprintf("csi-snapshot-restart-node-test-snap-%v", time.Now().Unix())

			v1obj := metav1.ObjectMeta{
				Name:      snapName,
				Namespace: ns,
			}
			source := volsnapv1.VolumeSnapshotSource{
				PersistentVolumeClaimName: &pvcName,
			}
			spec := volsnapv1.VolumeSnapshotSpec{
				VolumeSnapshotClassName: &snapShotClassName,
				Source:                  source,
			}
			snap := volsnapv1.VolumeSnapshot{
				ObjectMeta: v1obj,
				Spec:       spec,
			}
			_, err := csisnapshot.Instance().CreateSnapshot(&snap)
			log.FailOnError(err, "snapshot should not have failed")

			wg := sync.WaitGroup{}
			wg.Add(3)

			pods, err := core.Instance().ListPods(map[string]string{"app": "px-csi-driver"})
			dash.VerifyFatal(err == nil, true, fmt.Sprintf("Listed CSI pods"))

			for _, pod := range pods.Items {
				pod := pod
				go func() {
					defer wg.Done()
					defer GinkgoRecover()
					err = core.Instance().DeletePod(pod.Name, pod.Namespace, false)
					dash.VerifyFatal(err == nil, true, fmt.Sprintf("delete csi pod"))
				}()
			}
			wg.Wait()

			err = k8s.WaitForCsiSnapToBeReady(snapName, ns)
			log.FailOnError(err, "snapshot should have been completed successfully")
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		cleanupSnapshotests(context, ns)
		AfterEachTest(contexts)
	})
})

// Test CSI snapshot restore when CSI pods are restarted
var _ = Describe("{CSIOnlyTestCloudRestoreRestartCSIPods}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("CSIOnlyTestCloudRestoreRestartCSIPods", "Test create snapshot, with restart csi pods", nil, 0)
	})

	var pvcName, ns, snapShotClassName, snapName, scName string
	context := &scheduler.Context{
		App: &spec.AppSpec{
			Key: "snapshot-restart-csi-pods-test",
		},
	}
	var pvc *corev1.PersistentVolumeClaim
	var pod *corev1.Pod
	stepLog := "Create restore, with restart node"
	It(stepLog, func() {
		log.InfoD(stepLog)
		err := CreatePXCloudCredential()
		log.FailOnError(err, "failed to create cloud credential")

		ns = fmt.Sprintf("csi-restore-restart-csi-pods-test-ns-%v", time.Now().Unix())
		createNamespace(ns)

		stepLog = "Create CSI storage class"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			scName = fmt.Sprintf("csi-storage-class-restart-csi-pods-restore")
			createStorageClass(scName, nil)
		})

		stepLog = "Create volume snapshot class"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			snapShotClassName = CloudSnapShotClass
			createVolumeSnapshotClass(snapShotClassName, map[string]string{"csi.openstorage.org/snapshot-type": "cloud"})
		})

		stepLog = "Create PVC"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			pvcName = fmt.Sprintf("csi-restore-restart-csi-pods-%v", time.Now().Unix())
			pvc = create50GiReadWriteOncePVC(pvcName, ns, scName)
		})

		stepLog = "Create Pod"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			pod = createPod(pvc)
		})

		stepLog = "Write data"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			writeCmd := []string{"bash", "-c", fmt.Sprintf("dd if=/dev/urandom of=%s/filename bs=1048576 count=4096", "/usr/share/nginx/html")}
			_, err := k8sCore.RunCommandInPod(writeCmd, pod.Name, pod.Spec.Containers[0].Name, pod.Namespace)
			log.FailOnError(err, "Unable to write data")
		})

		stepLog = "Create cloud-snap"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			log.Infof("create cloudsnapshot with valid credentials")
			snapName = fmt.Sprintf("csi-snapshot-restart-node-test-snap-%v", time.Now().Unix())

			_, err = Inst().S.CreateCsiSnapshot(snapName, ns, snapShotClassName, pvcName)
			log.FailOnError(err, fmt.Sprintf("error creating snapshot [%v]", snapName))
		})
		stepLog = "Restore cloud-snap with CSI pod restart"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			restoredPVCSpec, err := k8s.GeneratePVCRestoreSpec(resource.MustParse("50Gi"), pvc.Namespace, pvc.Name+"-restore", snapName, scName)
			log.FailOnError(err, "failed to build restored PVC Spec")
			log.Infof("Generating PVC from snapshot source snapshot %s, pvc name %s", snapName, restoredPVCSpec.Name)
			restoredPvc, err := k8sCore.CreatePersistentVolumeClaim(restoredPVCSpec)
			log.FailOnError(err, "failed to restore PVC")

			pods, err := core.Instance().ListPods(map[string]string{"app": "px-csi-driver"})
			dash.VerifyFatal(err == nil, true, fmt.Sprintf("Listed CSI pods"))

			wg := sync.WaitGroup{}
			wg.Add(3)

			for _, pod := range pods.Items {
				pod := pod
				go func() {
					defer wg.Done()
					defer GinkgoRecover()
					err = core.Instance().DeletePod(pod.Name, pod.Namespace, false)
					dash.VerifyFatal(err == nil, true, fmt.Sprintf("delete csi pod"))
				}()
			}
			wg.Wait()

			err = Inst().S.WaitForSinglePVCToBound(restoredPvc.Name, ns, 3)
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		cleanupSnapshotests(context, ns)
		AfterEachTest(contexts)
	})
})

// Test CSI snapshot restore when node is restarted
var _ = Describe("{CSIOnlyTestCloudRestoreRestartNode}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("CSIOnlyTestCloudRestoreRestartNode", "Test create restore, with restart node", nil, 0)
	})

	var pvcName, ns, snapShotClassName, snapName, scName string
	context := &scheduler.Context{
		App: &spec.AppSpec{
			Key: "restore-restart-node-test",
		},
	}
	var pvc *corev1.PersistentVolumeClaim
	var pod *corev1.Pod
	stepLog := "Create restore, with restart node"
	It(stepLog, func() {
		log.InfoD(stepLog)
		err := CreatePXCloudCredential()
		log.FailOnError(err, "failed to create cloud credential")

		ns = fmt.Sprintf("csi-restore-restart-node-test-ns-%v", time.Now().Unix())
		createNamespace(ns)

		stepLog = "Create CSI storage class"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			scName = fmt.Sprintf("csi-storage-class-restart-node")
			createStorageClass(scName, nil)
		})

		stepLog = "Create volume snapshot class"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			snapShotClassName = CloudSnapShotClass
			createVolumeSnapshotClass(snapShotClassName, map[string]string{"csi.openstorage.org/snapshot-type": "cloud"})
		})

		stepLog = "Create PVC"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			pvcName = fmt.Sprintf("csi-snapshot-restart-node-test-%v", time.Now().Unix())
			pvc = create50GiReadWriteOncePVC(pvcName, ns, scName)
		})

		stepLog = "Create Pod"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			pod = createPod(pvc)
		})

		stepLog = "Write data"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			writeCmd := []string{"bash", "-c", fmt.Sprintf("dd if=/dev/urandom of=%s/filename bs=1048576 count=4096", "/usr/share/nginx/html")}
			_, err := k8sCore.RunCommandInPod(writeCmd, pod.Name, pod.Spec.Containers[0].Name, pod.Namespace)
			log.FailOnError(err, "Unable to write data")
		})

		stepLog = "Create cloud-snap"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			log.Infof("create cloudsnapshot with valid credentials")
			snapName = fmt.Sprintf("csi-snapshot-restart-node-test-snap-%v", time.Now().Unix())

			_, err = Inst().S.CreateCsiSnapshot(snapName, ns, snapShotClassName, pvcName)
			log.FailOnError(err, fmt.Sprintf("error creating snapshot [%v]", snapName))
		})
		stepLog = "Restore cloud-snap with Node restart"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			restoredPVCSpec, err := k8s.GeneratePVCRestoreSpec(resource.MustParse("50Gi"), pvc.Namespace, pvc.Name+"-restore", snapName, scName)
			log.FailOnError(err, "failed to build restored PVC Spec")
			log.Infof("Generating PVC from snapshot source snapshot %s, pvc name %s", snapName, restoredPVCSpec.Name)
			restoredPvc, err := k8sCore.CreatePersistentVolumeClaim(restoredPVCSpec)
			log.FailOnError(err, "failed to restore PVC")

			// Stop Driver on one of the replicas.
			storagenodes, err := GetStorageNodes()
			wg := sync.WaitGroup{}
			wg.Add(3)
			for _, n := range storagenodes {
				num := n
				go func() {
					defer wg.Done()
					defer GinkgoRecover()
					err = Inst().N.RebootNodeAndWait(num)
					dash.VerifyFatal(err == nil, true, fmt.Sprintf("Reboot node"))
					err = Inst().V.WaitDriverUpOnNode(num, 5*time.Minute)
					dash.VerifyFatal(err == nil, true, fmt.Sprintf("Wait for driver to start"))
				}()
			}
			wg.Wait()
			err = Inst().S.WaitForSinglePVCToBound(restoredPvc.Name, ns, 3)
		})
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
		cleanupSnapshotests(context, ns)
		AfterEachTest(contexts)
	})
})

// Test CSI snapshot restore when PX is restarted
var _ = Describe("{CSIOnlyTestCloudRestoreRestartPx}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("CSIOnlyTestCloudRestoreRestartPx", "Test create restore, with restart px", nil, 0)
	})

	var pvcName, ns, snapShotClassName, snapName, scName string
	context := &scheduler.Context{
		App: &spec.AppSpec{
			Key: "restore-restart-px-test",
		},
	}
	var pvc *corev1.PersistentVolumeClaim
	var pod *corev1.Pod
	stepLog := "Create restore, with restart px"
	It(stepLog, func() {
		log.InfoD(stepLog)
		err := CreatePXCloudCredential()
		log.FailOnError(err, "failed to create cloud credential")

		ns = fmt.Sprintf("csi-restore-restart-px-test-ns-%v", time.Now().Unix())
		createNamespace(ns)

		stepLog = "Create CSI storage class"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			scName = fmt.Sprintf("csi-storage-class-restart-node")
			createStorageClass(scName, nil)
		})

		stepLog = "Create volume snapshot class"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			snapShotClassName = CloudSnapShotClass
			createVolumeSnapshotClass(snapShotClassName, map[string]string{"csi.openstorage.org/snapshot-type": "cloud"})
		})

		stepLog = "Create PVC"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			pvcName = fmt.Sprintf("csi-snapshot-restart-px-test-%v", time.Now().Unix())
			pvc = create50GiReadWriteOncePVC(pvcName, ns, scName)
		})

		stepLog = "Create Pod"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			pod = createPod(pvc)
		})

		stepLog = "Write data"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			writeCmd := []string{"bash", "-c", fmt.Sprintf("dd if=/dev/urandom of=%s/filename bs=1048576 count=4096", "/usr/share/nginx/html")}
			_, err := k8sCore.RunCommandInPod(writeCmd, pod.Name, pod.Spec.Containers[0].Name, pod.Namespace)
			log.FailOnError(err, "Unable to write data")
		})

		stepLog = "Create cloud-snap"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			log.Infof("create cloudsnapshot with valid credentials")
			snapName = fmt.Sprintf("csi-snapshot-restart-px-test-snap-%v", time.Now().Unix())

			_, err = Inst().S.CreateCsiSnapshot(snapName, ns, snapShotClassName, pvcName)
			log.FailOnError(err, fmt.Sprintf("error creating snapshot [%v]", snapName))
		})

		stepLog = "Restore cloud-snap with Node restart"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			restoredPVCSpec, err := k8s.GeneratePVCRestoreSpec(resource.MustParse("50Gi"), pvc.Namespace, pvc.Name+"-restore", snapName, scName)
			log.FailOnError(err, "failed to build restored PVC Spec")
			log.Infof("Generating PVC from snapshot source snapshot %s, pvc name %s", snapName, restoredPVCSpec.Name)
			restoredPvc, err := k8sCore.CreatePersistentVolumeClaim(restoredPVCSpec)
			log.FailOnError(err, "failed to restore PVC")

			// Stop Driver on one of the replicas.
			storagenodes, err := GetStorageNodes()
			wg := sync.WaitGroup{}
			wg.Add(3)
			for _, n := range storagenodes {
				num := n
				go func() {
					defer wg.Done()
					defer GinkgoRecover()
					// restart portworx on node.
					restartPx(num)
				}()
			}
			wg.Wait()
			err = Inst().S.WaitForSinglePVCToBound(restoredPvc.Name, ns, 3)
		})
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
		cleanupSnapshotests(context, ns)
		AfterEachTest(contexts)
	})
})

// Test CSI snapshot restore after deleting bucket
var _ = Describe("{CSIOnlyTestCloudRestoreAfterBucketDelete}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("CSIOnlyTestCloudRestoreAfterBucketDelete", "Test create restore, after deleting bucket", nil, 0)
	})

	var pvcName, ns, snapShotClassName, snapName, scName string
	context := &scheduler.Context{
		App: &spec.AppSpec{
			Key: "restore-bucket-delete-test",
		},
	}
	var pvc *corev1.PersistentVolumeClaim
	stepLog := "Create restore, after deleting bucket"
	It(stepLog, func() {
		log.InfoD(stepLog)
		err := CreatePXCloudCredential()
		log.FailOnError(err, "failed to create cloud credential")

		ns = fmt.Sprintf("csi-restore-bucket-delete-ns-%v", time.Now().Unix())
		createNamespace(ns)

		stepLog = "Create CSI storage class"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			scName = fmt.Sprintf("csi-storage-class-bucket-delete")
			createStorageClass(scName, nil)
		})

		stepLog = "Create volume snapshot class"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			snapShotClassName = CloudSnapShotClass
			createVolumeSnapshotClass(snapShotClassName, map[string]string{"csi.openstorage.org/snapshot-type": "cloud"})
		})

		stepLog = "Create PVC"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			pvcName = fmt.Sprintf("csi-snapshot-bucket-delete-test-%v", time.Now().Unix())
			pvc = create50GiReadWriteOncePVC(pvcName, ns, scName)
		})

		stepLog = "Create cloud-snap"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			log.Infof("create cloudsnapshot with valid credentials")
			snapName = fmt.Sprintf("csi-snapshot-restart-px-test-snap-%v", time.Now().Unix())

			_, err = Inst().S.CreateCsiSnapshot(snapName, ns, snapShotClassName, pvcName)
			log.FailOnError(err, fmt.Sprintf("error creating snapshot [%v]", snapName))
		})

		stepLog = "Delete bucket"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			log.Infof("create cloudsnapshot with valid credentials")

			params := make(map[string]string)
			for k, v := range pvc.Annotations {
				params[k] = v
			}
			params[k8s.PvcNameKey] = pvc.GetName()
			params[k8s.PvcNamespaceKey] = pvc.GetNamespace()
			csBksps, err := Inst().V.GetCloudsnaps(pvc.Spec.VolumeName, params)
			log.FailOnError(err, "failed to get cloudsnaps")
			var bucketName string
			for _, csBksp := range csBksps {
				bkid := csBksp.GetId()
				bucketName = strings.Split(bkid, "/")[0]
				break
			}
			log.Infof("Got Bucket Name [%s]", bucketName)
			DeleteCloudSnapBucket(bucketName)
		})

		stepLog = "Restore cloud-snap"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			restoredPVCSpec, err := k8s.GeneratePVCRestoreSpec(resource.MustParse("50Gi"), pvc.Namespace, pvc.Name+"-restore", snapName, scName)
			log.FailOnError(err, "failed to build restored PVC Spec")
			log.Infof("Generating PVC from snapshot source snapshot %s, pvc name %s", snapName, restoredPVCSpec.Name)
			restoredPvc, err := k8sCore.CreatePersistentVolumeClaim(restoredPVCSpec)
			log.FailOnError(err, "failed to restore PVC")
			err = Inst().S.WaitForSinglePVCToBound(restoredPvc.Name, ns, 3)
			log.FailOnNoError(err, "restore should have failed as bucket is deleted")
		})
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
		cleanupSnapshotests(context, ns)
		AfterEachTest(contexts)
	})
})

// Test CSI snapshot when a volume is in degraded state.
// We do this by stopping one replica
var _ = Describe("{CSIOnlyTestCloudSnapshotDegradedState}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("CSIOnlyTestCloudSnapshotDegradedState", "Test create and restore snapshot, volume in degraded state", nil, 0)
	})

	var pvcName, ns, snapShotClassName, snapName, scName string
	context := &scheduler.Context{
		App: &spec.AppSpec{
			Key: "snapshot-degraded-stateƒIn",
		},
	}
	var pvc *corev1.PersistentVolumeClaim
	stepLog := "Take snapshot in volume degraded state"
	It(stepLog, func() {
		log.InfoD(stepLog)
		err := CreatePXCloudCredential()
		log.FailOnError(err, "failed to create cloud credential")

		ns = fmt.Sprintf("csi-snapshot-degraded-test-ns-%v", time.Now().Unix())
		createNamespace(ns)

		stepLog = "Create CSI storage class"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			scName = fmt.Sprintf("csi-storage-class-degraded")
			createStorageClass(scName, map[string]string{"repl": "2"})
		})

		stepLog = "Create volume snapshot class"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			snapShotClassName = CloudSnapShotClass
			createVolumeSnapshotClass(snapShotClassName, map[string]string{"csi.openstorage.org/snapshot-type": "cloud"})
		})

		stepLog = "Create PVC"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			pvcName = fmt.Sprintf("si-snapshot-degraded-test-%v", time.Now().Unix())
			pvc = create50GiReadWriteOncePVC(pvcName, ns, scName)
		})

		stepLog = "Get PV, put volume in degraded mode, and take snapshot"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			pv := pvc.Spec.VolumeName
			replicaSets, err := Inst().V.GetReplicaSets(&volume.Volume{
				ID: pv,
			})
			var nodeForPxStop node.Node
			// Put the volume in Degraded state.
			replicasNodes := replicaSets[0].Nodes
			// Stop Driver on one of the replicas.
			storagenodes, err := GetStorageNodes()
			for _, n := range storagenodes {
				if n.Id == replicasNodes[0] {
					nodeForPxStop = n
					break
				}
			}
			// Stop px, and try taking snapshot
			err = Inst().V.StopDriver([]node.Node{nodeForPxStop}, false, nil)
			dash.VerifyFatal(err == nil, true, fmt.Sprintf("Stop driver"))
			err = Inst().V.WaitDriverDownOnNode(nodeForPxStop)
			dash.VerifyFatal(err == nil, true, fmt.Sprintf("Wait for driver to sop"))
			// defer Restart
			defer func() {
				err = Inst().V.StartDriver(nodeForPxStop)
				dash.VerifyFatal(err == nil, true, fmt.Sprintf("Start driver"))
				err = Inst().V.WaitDriverUpOnNode(nodeForPxStop, 5*time.Minute)
				dash.VerifyFatal(err == nil, true, fmt.Sprintf("Wait for driver to start"))
			}()

			log.Infof("create cloudsnapshot with valid credentials")
			snapName = fmt.Sprintf("si-snapshot-degraded-test-snap-%v", time.Now().Unix())
			_, err = Inst().S.CreateCsiSnapshot(snapName, ns, snapShotClassName, pvcName)
			log.FailOnError(err, "snapshot should have failed")
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		cleanupSnapshotests(context, ns)
		AfterEachTest(contexts)
	})
})

// Test CSI snapshot when a volume is in out of quorum state.
// We do this by stopping all replicas
var _ = Describe("{CSIOnlyTestCloudSnapshotOutOfQuorum}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("CSIOnlyTestCloudSnapshotOutOfQuorum", "Test create and restore snapshot, volume in out of quorum state", nil, 0)
	})

	var pvcName, ns, snapShotClassName, snapName, scName string
	context := &scheduler.Context{
		App: &spec.AppSpec{
			Key: "snapshot-out-of-quorum",
		},
	}
	var pvc *corev1.PersistentVolumeClaim
	stepLog := "Take snapshot in volume out of quorum state"
	It(stepLog, func() {
		log.InfoD(stepLog)
		err := CreatePXCloudCredential()
		log.FailOnError(err, "failed to create cloud credential")

		ns = fmt.Sprintf("csi-snapshot-out-of-quorum-test-ns-%v", time.Now().Unix())
		createNamespace(ns)

		stepLog = "Create CSI storage class"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			scName = fmt.Sprintf("csi-storage-class-out-of-quorum")
			createStorageClass(scName, nil)
		})

		stepLog = "Create volume snapshot class"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			snapShotClassName = CloudSnapShotClass
			createVolumeSnapshotClass(snapShotClassName, map[string]string{"csi.openstorage.org/snapshot-type": "cloud"})
		})

		stepLog = "Create PVC"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			pvcName = fmt.Sprintf("csi-snapshot-out-of-quorum-test-%v", time.Now().Unix())
			pvc = create50GiReadWriteOncePVC(pvcName, ns, scName)
		})

		stepLog = "Get PV, put volume in degraded mode, and take snapshot"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			pv := pvc.Spec.VolumeName
			replicaSets, err := Inst().V.GetReplicaSets(&volume.Volume{
				ID: pv,
			})
			var nodeForPxStop node.Node
			// Put the volume in Degraded state.
			replicasNodes := replicaSets[0].Nodes
			// Stop Driver on one of the replicas.
			storagenodes, err := GetStorageNodes()
			for _, n := range storagenodes {
				if n.Id == replicasNodes[0] {
					nodeForPxStop = n
					break
				}
			}
			// Stop node and take snapshot
			err = Inst().V.StopDriver([]node.Node{nodeForPxStop}, false, nil)
			dash.VerifyFatal(err == nil, true, fmt.Sprintf("Stop driverr"))
			err = Inst().V.WaitDriverDownOnNode(nodeForPxStop)
			dash.VerifyFatal(err == nil, true, fmt.Sprintf("Wait for driver to sopt"))
			// defer Restart
			defer func() {
				err = Inst().V.StartDriver(nodeForPxStop)
				dash.VerifyFatal(err == nil, true, fmt.Sprintf("Start driver"))
				err = Inst().V.WaitDriverUpOnNode(nodeForPxStop, 5*time.Minute)
				dash.VerifyFatal(err == nil, true, fmt.Sprintf("Wait for driver to start"))
			}()

			log.Infof("create cloudsnapshot with valid credentials")
			snapName = fmt.Sprintf("csi-creds-test-%v", time.Now().Unix())
			_, err = Inst().S.CreateCsiSnapshot(snapName, ns, snapShotClassName, pvcName)
			log.FailOnNoError(err, "snapshot should have failed")
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()

		cleanupSnapshotests(context, ns)

		AfterEachTest(contexts)
	})
})

var _ = Describe("{ResizeVolumeAfterFull}", Label("p1", "positive", "px_vol_ops", "VolResize"), func() {
	/*
		https://portworx.atlassian.net/browse/PTX-18927
		Fill volumes completely , then resize volume by 50%, verify IO on volumes in Longevity
	*/
	JustBeforeEach(func() {
		StartTorpedoTest("ResizeVolumeAfterFull",
			"Fill volumes completely , then resize volume by 50%, verify IO on volumes in Longrun script",
			nil, 0)
	})

	var contexts []*scheduler.Context
	stepLog := "Fill volumes completely , then resize volume by 50%, verify IO on volumes in Longrun script"
	It(stepLog, func() {
		contexts = make([]*scheduler.Context, 0)
		currAppList := Inst().AppList

		revertAppList := func() {
			Inst().AppList = currAppList
		}
		defer revertAppList()

		Inst().AppList = []string{}
		var ioIntensiveApp = []string{"vdbench-heavyload"}

		for _, eachApp := range ioIntensiveApp {
			Inst().AppList = append(Inst().AppList, eachApp)
		}

		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("resizepoolfiftyper-%d", i))...)
		}
		ValidateApplications(contexts)
		defer appsValidateAndDestroy(contexts)

		log.Infof("Get all the list of available volumes with IO running")
		allVolumes, err := GetAllVolumesWithIO(contexts)
		log.FailOnError(err, "Failed to get volumes with IO Running")
		log.InfoD("List of all volumes with IO Running [%v]", allVolumes)

		// All Data volumes for Resize
		volumesToResize := []*volume.Volume{}
		for _, eachVol := range allVolumes {
			log.Infof("Checking volume with name [%v]", eachVol.Name)
			if eachVol.Name != "vdbench-pvc-output" {
				volumesToResize = append(volumesToResize, eachVol)
			}
		}
		dash.VerifyFatal(len(volumesToResize) > 0, true, "no volumes with IO for resize operations to continue")

		// Select Random Volumes for pool Expand
		randomIndex := rand.Intn(len(volumesToResize))
		randomVol := volumesToResize[randomIndex]

		waitForVolumeFull := func(volName *volume.Volume) error {
			waitTillVolume := func() (interface{}, bool, error) {
				volumeFull, err := IsVolumeFull(*volName)
				if err != nil {
					return nil, true, err
				}
				if volumeFull {
					return nil, false, nil
				}
				return nil, true, fmt.Errorf("Volume is still not full waiting.")
			}
			_, err := task.DoRetryWithTimeout(waitTillVolume, 2*time.Hour, 10*time.Second)
			return err
		}

		// Wait for Volume Full on the Node
		err = waitForVolumeFull(randomVol)
		log.FailOnError(err, "waiting for volume full on the node")

		// Expand Volume Size by 50%
		expectedSize := randomVol.Size + (randomVol.Size / 2)
		log.InfoD("Volume will be resized from [%v] to [%v]", randomVol.Size, expectedSize)
		log.FailOnError(Inst().V.ResizeVolume(randomVol.ID, expectedSize), "failed to Resize Volume")

		// Verify after Resize volume if IO is running
		isIOsInProgress, err := Inst().V.IsIOsInProgressForTheVolume(&node.GetStorageNodes()[0], randomVol.ID)
		log.FailOnError(err, "is io running on the volume?")
		dash.VerifyFatal(isIOsInProgress, true, fmt.Sprintf("no io running on the volume [%v] after resize", randomVol.Name))
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts)
	})

})

var _ = Describe("{CreateFastpathVolumeRebootNode}", Label("p1", "negative", "px_vol_ops", "error_injection", "px_ops", "node_reboot"), func() {
	var testrailID = 0
	// JIRA ID : https://portworx.atlassian.net/browse/PTX-15700
	var runID int
	JustBeforeEach(func() {
		StartTorpedoTest("CreateFastpathVolumeRebootNode",
			"Create fast path volume, reboot the node, check fastpath is active", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
		log.Infof("The runID  %v ", runID)
	})

	var pxNode node.Node
	var contexts []*scheduler.Context
	var volumrlidttr []*api.Volume

	stepLog := "Create fastpath Volume reboot node and check if fastpath is active"
	It(stepLog, func() {
		applist := Inst().AppList
		log.InfoD(stepLog)
		revertAppList := func() {
			Inst().AppList = applist
		}
		defer revertAppList()
		stepLog = "Step 1: Get all the Storage nodes and select a node for test"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			// Get all the Nodes
			pxNodes, err := GetStorageNodes()
			log.FailOnError(err, "Unable to get the storage nodes")

			// Select random Storage node for the test
			if len(pxNodes) > 0 {
				pxNode = GetRandomNode(pxNodes)
			} else {
				log.FailOnError(errors.New("No Storage Node Availiable"), "Error occured while selecting StorageNode")
			}

			log.Infof("The Selected node for Fast path label is %v : ", pxNode.Name)

			// Remove if node-type label is set before the test
			err = RemoveLabelsAllNodes(k8s.NodeType, true, false)
			log.FailOnError(err, "error removing label on node ")
		})

		stepLog = "Step 2: Schedule application and Add label on the selected storage node"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			var err error

			// Add label on the selected node
			err = Inst().S.AddLabelOnNode(pxNode, k8s.NodeType, k8s.FastpathNodeType)
			log.FailOnError(err, fmt.Sprintf("Failed add label on node %s", pxNode.Name))
			Inst().AppList = []string{"fio-fastpath-repl1"}
			contexts = make([]*scheduler.Context, 0)
			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				contexts = append(contexts, ScheduleApplications(fmt.Sprintf("fastpath-%d", i))...)
			}
			ValidateApplications(contexts)
		})

		stepLog = " Step 3: Get app volumes and Check fast path is active on the node"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			var err error
			for _, ctx := range contexts {
				var appVolumes []*volume.Volume
				stepLog = fmt.Sprintf("get volumes for %s app", ctx.App.Key)
				Step(stepLog, func() {
					log.InfoD(stepLog)
					appVolumes, err = Inst().S.GetVolumes(ctx)

					// Get the app volume details
					log.FailOnError(err, "Failed to get volumes for app %s", ctx.App.Key)
					dash.VerifyFatal(len(appVolumes) > 0, true, "App volumes exist?")
					log.Infof("App volumes details: %v ", appVolumes)

					// Store the volume details
					for _, vol := range appVolumes {
						apivol, err := Inst().V.InspectVolume(vol.ID)
						log.FailOnError(err, "failed to inspect volume %v", vol.ID)
						volumrlidttr = append(volumrlidttr, apivol)
					}

					// Loop through the apps and check if the volumes are fastpath active before reboot
					for _, appvolume := range appVolumes {
						log.Infof("current volume : %v", appvolume.Name)
						if strings.Contains(ctx.App.Key, fastpathAppName) {
							err := ValidateFastpathVolume(ctx, opsapi.FastpathStatus_FASTPATH_ACTIVE)
							log.FailOnError(err, "fastpath volume validation failed")
						}
					}
				})
			}
		})
		stepLog = " Step 4: Reboot the node wait for it to complete"
		Step(stepLog, func() {
			var err error
			var volExists bool
			log.Infof(" Before reboot check if the volumes are attached on the local node")
			for _, volumePtr := range volumrlidttr {
				volExists, err = Inst().V.IsVolumeAttachedOnNode(volumePtr, pxNode)
				log.FailOnError(err, "Volume attached on local node validation failed")
				if !volExists {
					log.FailOnError(errors.New("Error occured while inspecting volume "), " Volume attached on local node validation failed")
				}
			}
			log.Infof("The volumes were found attached on local node: %v", pxNode.Name)
			log.InfoD(stepLog)
			rebootNodeAndWaitForReady(&pxNode)
		})

		stepLog = " Step 5: Get app volumes and Check fast path is active on the node"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			var err error
			for _, ctx := range contexts {
				var appVolumes []*volume.Volume
				stepLog = fmt.Sprintf("get volumes for %s app", ctx.App.Key)
				Step(stepLog, func() {
					log.InfoD(stepLog)
					appVolumes, err = Inst().S.GetVolumes(ctx)

					// Get the app volume details
					log.FailOnError(err, "Failed to get volumes for app %s", ctx.App.Key)
					dash.VerifyFatal(len(appVolumes) > 0, true, "App volumes exist?")
					log.Infof("App volumes details: %v ", appVolumes)

					// Loop through the apps and check if the volumes are fastpath active after the reboot
					for _, appvolume := range appVolumes {
						log.Infof("current volume : %s", appvolume.Name)
						if strings.Contains(ctx.App.Key, fastpathAppName) {
							err := ValidateFastpathVolume(ctx, opsapi.FastpathStatus_FASTPATH_ACTIVE)
							log.FailOnError(err, "fastpath volume validation failed")
						}
					}

				})
			}

		})
		stepLog = " Step 6: Remove the Label from the selected node"
		Step(stepLog, func() {
			var err error
			log.InfoD(stepLog)
			err = Inst().S.RemoveLabelOnNode(pxNode, k8s.NodeType)
			log.FailOnError(err, "error removing label on node [%s]", pxNode.Name)
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts)
	})
})

var _ = Describe("{TrashcanRecoveryWithCloudsnap}", Label("p1", "positive", "px_vol_ops", "px_ops", "CloudSnapAndRestore"), func() {
	/*
		1) Create volumes
		2) Put the volume in resync state
		3) While some volumes are in resync, delete all volumes.
		4) Make sure all the volumes are in the trashcan.
		5) Recover all volumes from trashcan and Verify the volumes are restored correctly.
		6) Validate the data and verify cloudsnap schedules continues after restore

	*/
	JustBeforeEach(func() {
		StartTorpedoTest("TrashcanRecoveryWithCloudsnap", "Validate the successful restore from Trashcan when volumes got deleted in resync state", nil, 0)
	})

	stepLog := "Validate the successful restore from Trashcan of volume in resync"
	It(stepLog, func() {
		log.InfoD(stepLog)
		err := CreatePXCloudCredential()
		log.FailOnError(err, "failed to create cloud credential")

		stepLog = "Enable Trashcan"
		Step(stepLog,
			func() {
				log.InfoD(stepLog)
				currNode := node.GetStorageDriverNodes()[0]
				err := Inst().V.SetClusterOptsWithConfirmation(currNode, map[string]string{
					"--volume-expiration-minutes": "600",
				})
				log.FailOnError(err, "error while enabling trashcan")
				log.InfoD("Trashcan is successfully enabled")
			})
		stepLog = "Create schedule policy"
		policyName := "intervalpolicy"
		Step(stepLog, func() {
			log.InfoD(stepLog)

			schedPolicy, err := storkops.Instance().GetSchedulePolicy(policyName)
			if err != nil {
				retain := 3
				interval := 3
				log.InfoD("Creating a interval schedule policy %v with interval %v minutes", policyName, interval)
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

				_, err = storkops.Instance().CreateSchedulePolicy(schedPolicy)
				log.FailOnError(err, fmt.Sprintf("error creating a SchedulePolicy [%s]", policyName))
			}
		})

		defer func() {
			err := storkops.Instance().DeleteSchedulePolicy(policyName)
			log.FailOnError(err, fmt.Sprintf("error deleting a SchedulePolicy [%s]", policyName))

		}()
		fioPVC := "fio-pvc"
		fioPVName := "fio-pv"
		fioOutputPVC := "fio-output-pvc"
		fioOutputPVName := "fio-output-pv"

		appNamespace := fmt.Sprintf("tc-cs-%s", Inst().InstanceID)

		stepLog = fmt.Sprintf("create volumes %s and %s using volume request", fioPVName, fioOutputPVName)
		Step(stepLog, func() {
			log.InfoD(stepLog)
			log.Infof("Creating volume : %s", fioPVName)
			pxctlCmdFull := fmt.Sprintf("v c %s -s 500 -r 2", fioPVName)
			output, err := Inst().V.GetPxctlCmdOutput(node.GetStorageNodes()[0], pxctlCmdFull)
			log.FailOnError(err, fmt.Sprintf("error creating volume %s", fioPVName))
			log.Infof(output)

			log.Infof("Creating volume : %s", fioOutputPVName)
			pxctlCmdFull = fmt.Sprintf("v c %s -s 50 -r 2", fioOutputPVName)
			output, err = Inst().V.GetPxctlCmdOutput(node.GetStorageNodes()[0], pxctlCmdFull)
			log.FailOnError(err, fmt.Sprintf("error creating volume %s", fioOutputPVName))
			log.Infof(output)
		})
		appList := Inst().AppList
		defer func() {
			Inst().AppList = appList
		}()
		Inst().AppList = []string{"fio-pod"}

		contexts = make([]*scheduler.Context, 0)
		actRepls := make(map[*volume.Volume]int64)

		log.InfoD("scheduling apps ")
		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplicationsOnNamespace(appNamespace, fmt.Sprintf("trashrec-%d", i))...)
		}
		stepLog = fmt.Sprintf("Create volume snapshot schedule %s and %s", fioPVScheduleName, fioOutputPVScheduleName)
		Step(stepLog, func() {
			annotations := make(map[string]string)
			annotations["portworx/snapshot-type"] = "cloud"
			suspend := false
			log.InfoD(stepLog)

			tcVolSnapSched := &storkv1.VolumeSnapshotSchedule{
				ObjectMeta: meta_v1.ObjectMeta{
					Name:        fioPVScheduleName,
					Namespace:   appNamespace,
					Annotations: annotations,
				},
				Spec: storkv1.VolumeSnapshotScheduleSpec{
					SchedulePolicyName: policyName,
					Suspend:            &suspend,
					ReclaimPolicy:      storkv1.ReclaimPolicyDelete,
					Template: storkv1.VolumeSnapshotTemplateSpec{
						Spec: snapv1.VolumeSnapshotSpec{
							PersistentVolumeClaimName: fioPVC,
						},
					},
				},
			}

			volSnapshotSchedule, err := storkops.Instance().CreateSnapshotSchedule(tcVolSnapSched)
			log.FailOnError(err, fmt.Sprintf("error creating volume snapshot schedule %s", fioPVScheduleName))
			dash.VerifyFatal(volSnapshotSchedule.Name, fioPVScheduleName, "verify volume snapshot schedule is created successfully")

			tcVolSnapSched = &storkv1.VolumeSnapshotSchedule{
				ObjectMeta: meta_v1.ObjectMeta{
					Name:        fioOutputPVScheduleName,
					Namespace:   appNamespace,
					Annotations: annotations,
				},
				Spec: storkv1.VolumeSnapshotScheduleSpec{
					SchedulePolicyName: policyName,
					Suspend:            &suspend,
					ReclaimPolicy:      storkv1.ReclaimPolicyDelete,
					Template: storkv1.VolumeSnapshotTemplateSpec{
						Spec: snapv1.VolumeSnapshotSpec{
							PersistentVolumeClaimName: fioOutputPVC,
						},
					},
				},
			}

			volSnapshotSchedule, err = storkops.Instance().CreateSnapshotSchedule(tcVolSnapSched)
			log.FailOnError(err, fmt.Sprintf("error creating volume snapshot schedule %s", fioOutputPVScheduleName))
			dash.VerifyFatal(volSnapshotSchedule.Name, fioOutputPVScheduleName, "verify volume snapshot schedule is created successfully")
		})

		for _, ctx := range contexts {
			ctx.SkipVolumeValidation = true
			ValidateContext(ctx)
		}

		log.Infof("waiting for 10 mins for enough cloud snaps to be created.")
		time.Sleep(10 * time.Minute)
		initalSnaps, err := validateCloudSnaps(appNamespace)
		log.FailOnError(err, "error validating cloudsnaps")

		stepLog := "Scenario: Get volumes for cloudsnap apps in test,update replication factor,delete volumes then restore volumes from trashcan and validate cloudsnaps"
		Step(stepLog, func() {
			log.InfoD(stepLog)

			for _, ctx := range contexts {

				//waiting for the data to be written before performing ha-update
				_, err := GetVolumeWithMinimumSize([]*scheduler.Context{ctx}, 50)
				log.FailOnError(err, "error selecting volumes for resync")

				testVolumes, err := Inst().S.GetVolumes(ctx)
				log.FailOnError(err, "Failed to get volumes for app %s", ctx.App.Key)
				dash.VerifyFatal(len(testVolumes) > 0, true, "App volumes exist?")

				appVolumes := make([]*volume.Volume, 0)
				//Getting volumes for repl update having minimum size
				for _, vol := range testVolumes {
					appVol, err := Inst().V.InspectVolume(vol.ID)
					log.FailOnError(err, fmt.Sprintf("error inspecting volume %s", vol.Name))
					usedBytes := appVol.GetUsage()
					usedGiB := usedBytes / units.GiB
					if usedGiB >= 50 {
						appVolumes = append(appVolumes, vol)
					}
				}

				//Reducing the repl factor if volume as max replication factor enabled
				stepLog = fmt.Sprintf("Adjusting the volume replications before increasing the repls for %s", ctx.App.Key)
				Step(stepLog, func() {
					log.InfoD(stepLog)
					err = replAdjust(appVolumes, actRepls)
					log.FailOnError(err, "Failed to adjust the repls for volumes of the app %s", ctx.App.Key)
				})

				stepLog = fmt.Sprintf("Increasing the volume repls for %s", ctx.App.Key)
				Step(stepLog, func() {
					log.InfoD(stepLog)
					newRepls := make(map[*volume.Volume]int64)
					for _, v := range appVolumes {
						currRep, err := Inst().V.GetReplicationFactor(v)
						log.FailOnError(err, "Failed to get volume  %s repl factor", v.Name)
						currAggr, err := Inst().V.GetAggregationLevel(v)
						log.FailOnError(err, "Failed to get volume  %s aggregate level", v.Name)
						numStorageNodes := len(node.GetStorageNodes())
						numStorageNodesRequired := int(currAggr * (currRep + 1))

						if numStorageNodes < numStorageNodesRequired {
							log.Warnf("skipping volume %s repl increase as numStorageNodesRequired is %d where as numStorageNodes is %d", v.Name, numStorageNodesRequired, numStorageNodes)
							continue
						}
						newRepls[v] = currRep + 1
					}
					err = setVolumeRepl(newRepls, false)
					dash.VerifyFatal(err, nil, fmt.Sprintf("validate successful repl increase for %s", ctx.App.Key))
				})

				log.InfoD("waiting for volumes to be in resync state")
				time.Sleep(2 * time.Minute)

				//waiting for all the volume is resync state
				for _, v := range appVolumes {

					checkVolumeState := func() (interface{}, bool, error) {
						runTimeState, err := GetVolumeReplicationStatus(v)
						if err != nil {
							return "", false, fmt.Errorf("error getting run time state for volume:%s. App : %s", v.Name, ctx.App.Key)
						}
						if strings.ToLower(runTimeState) == "resync" {
							return "", false, nil
						}

						return nil, true, fmt.Errorf("waiting for volume %s run time state to change to resync, current state: %s", v.Name, runTimeState)
					}
					_, err = task.DoRetryWithTimeout(checkVolumeState, time.Duration(5)*defaultCommandTimeout, 1*time.Minute)
					log.FailOnError(err, "error validating volume runtime state for %s", v.Name)

				}

				stepLog = fmt.Sprintf("Deleting app %s", ctx.App.Key)
				Step(stepLog, func() {
					DestroyApps(contexts, nil)
					log.FailOnError(deletePXVolume(fioPVName), fmt.Sprintf("error deleting portworx volume %s", fioPVName))
					log.FailOnError(deletePXVolume(fioOutputPVName), fmt.Sprintf("error deleting portworx volume %s", fioOutputPVName))
				})

				var trashcanVols []string
				stepLog = "validate volumes in trashcan"
				Step(stepLog, func() {
					// wait for few seconds for pvc to get deleted and volume to get detached
					time.Sleep(10 * time.Second)
					node := node.GetStorageDriverNodes()[0]
					log.InfoD(stepLog)
					trashcanVols, err = Inst().V.GetTrashCanVolumeIds(node)
					log.FailOnError(err, "error While getting trashcan volumes")
					log.Infof("trashcan len: %d", len(trashcanVols))
					dash.VerifyFatal(len(trashcanVols) > 0, true, "validate volumes exist in trashcan")

				})

				stepLog = "Validating trashcan restore"
				Step(stepLog,
					func() {
						log.InfoD(stepLog)
						for _, tID := range trashcanVols {
							if tID != "" {
								vol, err := Inst().V.InspectVolume(tID)
								log.FailOnError(err, fmt.Sprintf("error inspecting volume %s", tID))
								if strings.Contains(vol.Locator.Name, "fio-output-pv") {
									err = trashcanRestore(vol.Id, "fio-output-pv")
									log.FailOnError(err, fmt.Sprintf("error restoring volume %s from trashcan", vol.Id))
								}
								if strings.Contains(vol.Locator.Name, "fio-pv") {
									err = trashcanRestore(vol.Id, "fio-pv")
									log.FailOnError(err, fmt.Sprintf("error restoring volume %s from trashcan", vol.Id))
								}
							}
						}

					})
				log.InfoD("scheduling apps ")
				for i := 0; i < Inst().GlobalScaleFactor; i++ {
					contexts = append(contexts, ScheduleApplicationsOnNamespace(appNamespace, fmt.Sprintf("trashrec-%d", i))...)
				}
				for _, ctx := range contexts {
					ctx.SkipVolumeValidation = true
					ValidateContext(ctx)
				}

				log.Infof("waiting for 10 mins for enough cloud snaps to be created.")
				time.Sleep(10 * time.Minute)
				restoredSnaps, err := validateCloudSnaps(appNamespace)
				log.FailOnError(err, "error validating cloudsnaps")
				dash.VerifyFatal(len(restoredSnaps) > 0, true, "verify cloudsnaps are created.")
				for k := range restoredSnaps {
					dash.VerifySafely(initalSnaps[k] != restoredSnaps[k], true, fmt.Sprintf("verfiy new snaps are created for schedule %s", k))
				}

			}

		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		bucketName, err := GetCloudsnapBucketName(contexts)
		log.FailOnError(err, "error getting cloud snap bucket name")
		opts := make(map[string]bool)
		DestroyApps(contexts, opts)
		DeleteCloudSnapBucket(bucketName)
		AfterEachTest(contexts)
	})
})

func replAdjust(appVolumes []*volume.Volume, actRepls map[*volume.Volume]int64) error {
	setRepls := make(map[*volume.Volume]int64)

	for _, v := range appVolumes {
		currRep, err := Inst().V.GetReplicationFactor(v)
		if err != nil {
			return err
		}
		actRepls[v] = currRep
		if currRep == 3 {
			setRepls[v] = currRep - 1
		}
	}
	err := setVolumeRepl(setRepls, true)
	return err
}

func setVolumeRepl(setRepls map[*volume.Volume]int64, waitToFinish bool) error {

	for v, r := range setRepls {
		log.InfoD("setting repl for volume %s to %d", v.ID, r)
		err := Inst().V.SetReplicationFactor(v, r, nil, nil, waitToFinish)
		if err != nil {
			return err
		}

	}

	return nil
}

func validateCloudSnaps(appNamespace string) (map[string]string, error) {

	log.Infof("Verify that cloud snap status")
	snapsMap := make(map[string]string, 0)

	for _, ctx := range contexts {
		if strings.Contains(ctx.App.Key, "cloudsnap") || strings.Contains(ctx.App.Key, "fastpath") {

			var appVolumes []*volume.Volume
			var err error

			appVolumes, err = Inst().S.GetVolumes(ctx)
			log.FailOnError(err, "error getting volumes for [%s]", ctx.App.Key)
			if err != nil {
				return snapsMap, err
			}

			if len(appVolumes) == 0 {
				return snapsMap, fmt.Errorf("no volumes found for [%s]", ctx.App.Key)
			}

			log.Infof("Got volume count : %v", len(appVolumes))

			err = Inst().S.ValidateVolumes(ctx, 4*time.Minute, defaultRetryInterval, nil)
			log.FailOnError(err, "error validating volumes for [%s]", ctx.App.Key)
			if err != nil {
				return snapsMap, err
			}

			for _, v := range appVolumes {
				isPureVol, err := Inst().V.IsPureVolume(v)
				if err != nil {
					return snapsMap, err
				}

				if isPureVol {
					log.Warnf("Cloud snapshot is not supported for Pure DA volumes: [%s],Skipping cloud snapshot trigger for pure volume.", v.Name)
					continue
				}
				snapshotScheduleName := ""
				if v.Name == "fio-pvc" {
					snapshotScheduleName = fioPVScheduleName
				} else if v.Name == "fio-output-pvc" {
					snapshotScheduleName = fioOutputPVScheduleName
				} else {
					snapshotScheduleName = v.Name + "-interval-schedule"
				}

				log.InfoD("snapshotScheduleName : %v for volume: %s", snapshotScheduleName, v.Name)
				resp, err := storkops.Instance().GetSnapshotSchedule(snapshotScheduleName, appNamespace)
				if err != nil {
					return snapsMap, err
				}

				dash.VerifyFatal(len(resp.Status.Items) > 0, true, fmt.Sprintf("verify snapshots exists for %s", snapshotScheduleName))
				for _, snapshotStatuses := range resp.Status.Items {
					if len(snapshotStatuses) > 0 {
						status := snapshotStatuses[len(snapshotStatuses)-1]
						if status == nil {
							return snapsMap, fmt.Errorf("snapshotSchedule has an empty migration in it's most recent status,Err: %v", err)
						}
						status, err = WaitForSnapShotToReady(snapshotScheduleName, status.Name, appNamespace)
						if err != nil {
							return snapsMap, err
						}
						log.Infof("Snapshot %s has status %v", status.Name, status.Status)

						if status.Status == snapv1.VolumeSnapshotConditionError {
							volumeSnapshot, err := Inst().S.GetSnapShot(ctx, status.Name, appNamespace)
							if err != nil {
								log.Errorf("Error getting volume snapshot [%s] in namespace [%s]. Error: [%v]", status.Name, appNamespace, err)
							}
							if volumeSnapshot != nil {
								log.Errorf("volumeSnapshot : %+v", volumeSnapshot)
							}
							return snapsMap, fmt.Errorf("snapshot: %s failed. status: %v", status.Name, status.Status)
						}

						if status.Status == snapv1.VolumeSnapshotConditionPending {
							return snapsMap, fmt.Errorf("snapshot: %s not completed. status: %v", status.Name, status.Status)
						}

						if status.Status == snapv1.VolumeSnapshotConditionReady {
							snapData, err := Inst().S.GetSnapShotData(ctx, status.Name, appNamespace)

							if err != nil {
								return snapsMap, err
							}

							snapType := snapData.Spec.PortworxSnapshot.SnapshotType
							log.Infof("Snapshot Type: %v", snapType)
							if snapType != "cloud" {
								err = &scheduler.ErrFailedToGetVolumeParameters{
									App:   ctx.App,
									Cause: fmt.Sprintf("Snapshot Type: %s does not match", snapType),
								}
								return snapsMap, err
							}

							snapID := snapData.Spec.PortworxSnapshot.SnapshotID
							log.Infof("Snapshot ID: %v", snapID)
							if snapData.Spec.VolumeSnapshotDataSource.PortworxSnapshot == nil ||
								len(snapData.Spec.VolumeSnapshotDataSource.PortworxSnapshot.SnapshotID) == 0 {
								err = &scheduler.ErrFailedToGetVolumeParameters{
									App:   ctx.App,
									Cause: fmt.Sprintf("volumesnapshotdata: %s does not have portworx volume source set", snapData.Metadata.Name),
								}
								return snapsMap, err
							}

						}
						snapsMap[snapshotScheduleName] = status.Name

					}
				}

			}
		}
	}
	return snapsMap, nil
}

func validateCloudSnapValues(credUUID string, volName string, params map[string]string) bool {
	log.InfoD("Validating snapshot values")
	cSnaps, err := Inst().V.GetCloudsnaps(volName, params)
	if err != nil || len(cSnaps) == 0 {
		log.FailOnError(err, "error getting cloudsnaps or no cloudsnaps found!")
		return false
	}
	for _, cSnap := range cSnaps {
		volData := cSnap.Metadata["volume"]
		log.Infof("Volume being validated, source  %s, target %s", cSnap.SrcVolumeName, volName)
		if cSnap.SrcVolumeName != volName {
			continue
		}
		log.Infof("Volume Data from SDK: %v", volData)
		var volumeData volumeDataMap
		err := json.Unmarshal([]byte(volData), &volumeData)
		if err != nil {
			log.FailOnError(err, "Error while unmarshalling volume data")
			return false
		}
		totalRestoreSize := strconv.FormatFloat(volumeData.TotalRestoreSize, 'f', -1, 64)
		compressedSizeBytes := cSnap.Metadata["compressedSizeBytes"]
		capacityRequiredForRestore := strconv.FormatFloat(volumeData.UsedSize, 'f', -1, 64)
		log.Infof("TotalRestoreSize: %v, CompressedObjectBytes: %v, CapacityRequiredForRestore: %v",
			totalRestoreSize, compressedSizeBytes, capacityRequiredForRestore)

		// API GET values from v1/cloudbackups/size
		url := fmt.Sprintf("http://%s:9021/v1/cloudbackups/size?credential_id=%s&backup_id=%s",
			node.GetStorageDriverNodes()[0].MgmtIp, credUUID, cSnap.Id)
		resp, respStatusCode, err := restutil.GET(url, nil, nil)
		if err != nil || respStatusCode != http.StatusOK || len(resp) == 0 {
			log.FailOnError(err, "Error in fetching cloud backup size, Cause: %v; Status code: %v "+
				"\n Or the data is empty", err, respStatusCode)
			return false
		}
		log.InfoD("Parsing output from cloud backup size API")
		var cloudBackupSize CloudBackupSizeAPI
		err = json.Unmarshal(resp, &cloudBackupSize)
		log.Infof("CloudBackupSize: %v", cloudBackupSize)

		if err != nil || cloudBackupSize.Size != totalRestoreSize ||
			cloudBackupSize.TotalDownloadBytes != totalRestoreSize ||
			cloudBackupSize.CompressedObjectBytes != compressedSizeBytes ||
			cloudBackupSize.Capacity != capacityRequiredForRestore {
			log.FailOnError(err, "Cloudsnap size mismatch: %v", cSnap.Id)
			return false
		}
	}

	return true
}

func trashcanRestore(volId, volName string) error {
	log.InfoD("Restoring vol [%v] from trashcan", volId)
	pxctlCmdFull := fmt.Sprintf("v r %s --trashcan %s", volName, volId)
	output, err := Inst().V.GetPxctlCmdOutput(node.GetStorageNodes()[0], pxctlCmdFull)
	if err != nil {
		return err
	}
	log.Infof("output: %v", output)
	if !strings.Contains(output, fmt.Sprintf("Successfully restored: %s", volName)) {
		err = fmt.Errorf("volume %v, restore from trashcan failed, Err: %v", volId, output)
		return err
	}
	return nil
}

func deletePXVolume(volName string) error {
	delVol := func() (interface{}, bool, error) {
		err := Inst().V.DeleteVolume(volName)
		if err != nil {
			return "", true, err
		}
		return nil, false, nil
	}
	_, err := task.DoRetryWithTimeout(delVol, time.Duration(60)*defaultCommandTimeout, 2*time.Minute)
	return err
}

var _ = Describe("{CloudSnapWithPXEvents}", Label("p0", "positive", "px_vol_ops", "px_ops", "CloudSnapAndRestore"), func() {
	var testrailID = 0
	var runID int
	JustBeforeEach(func() {
		StartTorpedoTest("CloudSnapWithPXEvents", "Validate cloudsnap during PX events", nil, 0)
		runID = testrailuttils.AddRunsToMilestone(0)

	})

	stepLog := "has to schedule apps with cloudsnaps and perform PX events"
	It(stepLog, func() {
		log.InfoD(stepLog)

		contexts = make([]*scheduler.Context, 0)

		stepLog = "validate cloud cred and create schedule policy"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			err := CreatePXCloudCredential()
			log.FailOnError(err, "failed to create cloud credential")
			contexts = make([]*scheduler.Context, 0)
			policyName := "intervalpolicy"

			stepLog = fmt.Sprintf("create schedule policy %s", policyName)
			Step(stepLog, func() {
				log.InfoD(stepLog)

				schedPolicy, err := storkops.Instance().GetSchedulePolicy(policyName)
				if err != nil {
					retain := 5
					interval := 1
					log.InfoD("Creating a interval schedule policy %v with interval %v minutes", policyName, interval)
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

					_, err = storkops.Instance().CreateSchedulePolicy(schedPolicy)
					log.FailOnError(err, fmt.Sprintf("error creating a SchedulePolicy [%s]", policyName))
				}
			})

			defer func() {
				err := storkops.Instance().DeleteSchedulePolicy(policyName)
				log.FailOnError(err, fmt.Sprintf("error deleting a SchedulePolicy [%s]", policyName))
			}()

			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				contexts = append(contexts, ScheduleApplications(fmt.Sprintf("cspxevents-%d", i))...)
			}

			areCloudsnapEnabledAppsDeployed := false

			for _, ctx := range contexts {
				if strings.Contains(ctx.App.Key, "cloudsnap") {
					areCloudsnapEnabledAppsDeployed = true
					break
				}
			}

			if !areCloudsnapEnabledAppsDeployed {
				log.FailOnError(fmt.Errorf("no cloudsnap enabled apps deployed"), "error validating apps for cloudsnaps events test")
			}

			ValidateApplications(contexts)
			nsList, err := core.Instance().ListNamespaces(map[string]string{"creator": "torpedo"})
			log.FailOnError(err, "error getting all namespaces")
			log.Infof("%v", nsList)
			appNamespaces := make([]string, 0)
			for _, ns := range nsList.Items {
				if strings.Contains(ns.Name, "cspxevents") {
					appNamespaces = append(appNamespaces, ns.Name)
				}
			}

			if len(appNamespaces) == 0 {
				log.FailOnError(fmt.Errorf("no namespaces found to validate cloudsnaps"), "error getting cloudsnap namespaces")
			}

			stepLog = "validate cloudsnaps"
			Step(stepLog, func() {
				log.InfoD(stepLog)
				for _, ns := range appNamespaces {
					_, err = validateCloudSnaps(ns)
					log.FailOnError(err, fmt.Sprintf("error validating cloudsnaps in namespace [%s]", ns))
				}
			})
			var csAppVolumes []*volume.Volume
			for _, ctx := range contexts {
				if strings.Contains(ctx.App.Key, "cloudsnap") {
					appVolumes, err := Inst().S.GetVolumes(ctx)
					log.FailOnError(err, "Failed to get volumes for app %s", ctx.App.Key)
					csAppVolumes = append(csAppVolumes, appVolumes...)
				}
			}

			stepLog = "trigger px restart event during cloudsnaps and validate cloudsnaps"
			Step(stepLog, func() {
				log.InfoD(stepLog)

				for _, csAppVol := range csAppVolumes {
					attachedNode, err := Inst().V.GetNodeForVolume(csAppVol, defaultCommandTimeout, defaultCommandRetry)
					dash.VerifySafely(err, nil, fmt.Sprintf("Verify Get nodes for vol %s", csAppVol.Name))
					stepLog = fmt.Sprintf("stop volume driver %s on node: %s",
						Inst().V.String(), attachedNode.Name)
					Step(stepLog,
						func() {
							StopVolDriverAndWait([]node.Node{*attachedNode})
						})

					log.Infof("wait for 10 mins for volumes to reallocate")
					time.Sleep(10 * time.Minute)

					stepLog = fmt.Sprintf("starting volume %s driver on node %s",
						Inst().V.String(), attachedNode.Name)
					Step(stepLog,
						func() {
							StartVolDriverAndWait([]node.Node{*attachedNode})
						})

					stepLog = "Giving few seconds for volume driver to stabilize"
					Step(stepLog, func() {
						log.InfoD("Giving few seconds for volume driver to stabilize")
						time.Sleep(20 * time.Second)
					})

				}
				for _, ctx := range contexts {
					ValidateContext(ctx)
				}
				stepLog = "validate cloudsnaps"
				Step(stepLog, func() {
					log.InfoD(stepLog)

					for _, ns := range appNamespaces {
						_, err = validateCloudSnaps(ns)
						log.FailOnError(err, fmt.Sprintf("error validating cloudsnaps in namespace [%s]", ns))
					}
				})
			})

			stepLog = "validate repl update during cloudsnaps"

			Step(stepLog, func() {
				stopValidation := make(chan bool, 1)

				log.InfoD(stepLog)

				actRepls := make(map[*volume.Volume]int64)
				//Reducing the repl factor if volume as max replication factor enabled
				stepLog = fmt.Sprintf("Adjusting the volume replications before increasing the repls for cloudsnap volumes")
				Step(stepLog, func() {
					log.InfoD(stepLog)
					err = replAdjust(csAppVolumes, actRepls)
					log.FailOnError(err, "Failed to adjust the repls for cloudsnap volumes")
				})

				stepLog = fmt.Sprintf("Increasing the repls for cloudsnap volumes")
				Step(stepLog, func() {
					log.InfoD(stepLog)
					newRepls := make(map[*volume.Volume]int64)
					for _, v := range csAppVolumes {
						currRep, err := Inst().V.GetReplicationFactor(v)
						log.FailOnError(err, "Failed to get volume  %s repl factor", v.Name)
						currAggr, err := Inst().V.GetAggregationLevel(v)
						log.FailOnError(err, "Failed to get volume  %s aggregate level", v.Name)
						numStorageNodes := len(node.GetStorageNodes())
						numStorageNodesRequired := int(currAggr * (currRep + 1))

						if numStorageNodes < numStorageNodesRequired {
							log.Warnf("skipping volume %s repl increase as numStorageNodesRequired is %d where as numStorageNodes is %d", v.Name, numStorageNodesRequired, numStorageNodes)
							continue
						}
						newRepls[v] = currRep + 1
					}
					// Create a channel to signal an error.
					errorChan := make(chan error, 2)
					// Create a WaitGroup to wait for both functions to finish.
					var wg sync.WaitGroup
					for v, r := range newRepls {
						log.InfoD("setting repl for volume %s to %d", v.Name, r)
						if err := Inst().V.SetReplicationFactor(v, r, nil, nil, false); err != nil {
							log.Errorf(fmt.Sprintf("got error while repl increase %v", err))

						}
					}

					//Go routine for cloudsnap validate
					wg.Add(1)
					go func() {
						defer GinkgoRecover()
						defer wg.Done()

						for {
							select {
							case <-errorChan:
								close(stopValidation)
								close(errorChan)
								return
							case <-stopValidation:
								close(stopValidation)
								close(errorChan)
								return
							default:
								for _, ns := range appNamespaces {
									if _, err := validateCloudSnaps(ns); err != nil {
										errorChan <- err
										break
									}
								}
								log.Infof("waiting for 3 mins for next validation")
								time.Sleep(3 * time.Minute)
							}
						}

					}()

					for v, r := range newRepls {
						log.InfoD(fmt.Sprintf("validating repl for %s shoulbe be %d", v.ID, r))
						if err = ValidateReplFactorUpdate(v, r); err != nil {
							errorChan <- err
							break
						}
					}
					stopValidation <- true

					wg.Wait()
					for e := range errorChan {
						dash.VerifySafely(e, nil, "validate cloudsnaps while repl increase")
					}

					for v, r := range actRepls {
						if newRepls[v] != r {
							log.InfoD("setting repl for volume %s to %d", v.Name, r)
							err = Inst().V.SetReplicationFactor(v, r, nil, nil, true)
							log.FailOnError(err, fmt.Sprintf("error setting repl for volume %s with value %d", v.ID, r))
						}

					}

				})

			})

			stepLog = "node de-comm and rejoin while cloudsnap in progress"
			Step(stepLog, func() {
				attachedNodes := make(map[string]bool, 0)
				for _, v := range csAppVolumes {
					attachedNode, err := Inst().V.GetNodeForVolume(v, 1*time.Minute, 5*time.Second)
					log.FailOnError(err, fmt.Sprintf("error getting attahced node for volume %s", v.Name))
					if attachedNode != nil {
						if _, ok := attachedNodes[attachedNode.Name]; !ok {
							attachedNodes[attachedNode.Name] = true
						}
					}
				}

				for attachedNode := range attachedNodes {
					nodeToDecommission, err := node.GetNodeByName(attachedNode)
					stepLog = fmt.Sprintf("decommission node %s", nodeToDecommission.Name)
					Step(stepLog, func() {
						log.InfoD(stepLog)
						err := Inst().S.PrepareNodeToDecommission(nodeToDecommission, Inst().Provisioner)
						dash.VerifyFatal(err, nil, "Validate node decommission preparation")
						err = Inst().V.DecommissionNode(&nodeToDecommission)
						dash.VerifyFatal(err, nil, fmt.Sprintf("Validate node [%s] decommission init", nodeToDecommission.Name))
						stepLog = fmt.Sprintf("check if node %s was decommissioned", nodeToDecommission.Name)
						Step(stepLog, func() {
							log.InfoD(stepLog)
							t := func() (interface{}, bool, error) {
								status, err := Inst().V.GetNodeStatus(nodeToDecommission)
								if err != nil {
									return false, true, err
								}
								if *status == api.Status_STATUS_NONE {
									return true, false, nil
								}
								return false, true, fmt.Errorf("node %s not decomissioned yet", nodeToDecommission.Name)
							}
							decommissioned, err := task.DoRetryWithTimeout(t, 15*time.Minute, defaultRetryInterval)
							log.FailOnError(err, "Failed to get decommissioned node status")
							dash.VerifyFatal(decommissioned.(bool), true, fmt.Sprintf("Validate node [%s] is decommissioned", nodeToDecommission.Name))
						})
					})
					stepLog = "validate cloudsnaps after node decomm"
					Step(stepLog, func() {
						log.InfoD(stepLog)
						for _, ns := range appNamespaces {
							_, err = validateCloudSnaps(ns)
							dash.VerifySafely(err, nil, fmt.Sprintf("validating cloudsnaps in namespace [%s]", ns))
						}
					})
					stepLog = fmt.Sprintf("Rejoin node %s", nodeToDecommission.Name)
					Step(stepLog, func() {
						log.InfoD(stepLog)
						//reboot required to remove encrypted dm devices if any
						err := Inst().N.RebootNode(nodeToDecommission, node.RebootNodeOpts{
							Force: true,
							ConnectionOpts: node.ConnectionOpts{
								Timeout:         defaultCommandTimeout,
								TimeBeforeRetry: defaultRetryInterval,
							},
						})
						log.FailOnError(err, fmt.Sprintf("error rebooting node %s", nodeToDecommission.Name))
						err = Inst().V.RejoinNode(&nodeToDecommission)
						dash.VerifyFatal(err, nil, "Validate node rejoin init")
						var rejoinedNode *api.StorageNode
						t := func() (interface{}, bool, error) {
							drvNodes, err := Inst().V.GetDriverNodes()
							if err != nil {
								return false, true, err
							}

							for _, n := range drvNodes {
								if n.Hostname == nodeToDecommission.Hostname {
									rejoinedNode = n
									return true, false, nil
								}
							}

							return false, true, fmt.Errorf("node %s not joined yet", nodeToDecommission.Name)
						}
						_, err = task.DoRetryWithTimeout(t, 15*time.Minute, defaultRetryInterval)
						log.FailOnError(err, fmt.Sprintf("error joining the node [%s]", nodeToDecommission.Name))
						dash.VerifyFatal(rejoinedNode != nil, true, fmt.Sprintf("verify node [%s] rejoined PX cluster", nodeToDecommission.Name))
						err = Inst().S.RefreshNodeRegistry()
						log.FailOnError(err, "error refreshing node registry")
						err = Inst().V.RefreshDriverEndpoints()
						log.FailOnError(err, "error refreshing storage drive endpoints")
						decommissionedNode := node.Node{}
						for _, n := range node.GetStorageDriverNodes() {
							if n.Name == rejoinedNode.Hostname {
								decommissionedNode = n
								break
							}
						}
						if decommissionedNode.Name == "" {
							log.FailOnError(fmt.Errorf("rejoined node not found"), fmt.Sprintf("node [%s] not found in the node registry", rejoinedNode.Hostname))
						}
						err = Inst().V.WaitDriverUpOnNode(decommissionedNode, Inst().DriverStartTimeout)
						dash.VerifyFatal(err, nil, fmt.Sprintf("Validate driver up on rejoined node [%s] after rejoining", decommissionedNode.Name))
					})

				}

			})
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
		defer EndTorpedoTest()
		bucketName, err := GetCloudsnapBucketName(contexts)
		log.FailOnError(err, "error getting cloud snap bucket name")
		opts := make(map[string]bool)
		DestroyApps(contexts, opts)
		DeleteCloudSnapBucket(bucketName)
		AfterEachTest(contexts, testrailID, runID)
	})
})

var _ = Describe("{PoolFullCloudsnap}", Label("p1", "positive", "px_vol_ops", "Throttling", "CloudSnapAndRestore"), func() {

	/*
			Priority: P1
		1. Selected a node and a pool
		2. Deploy cloudsnap apps and make sure volumes are attached to the pool selected above
		2. Fill up the pool and validate cloudsnaps
		3. Do pool expansion and validate cloudsnaps
	*/

	var testrailID = 0
	var runID int

	JustBeforeEach(func() {
		StartTorpedoTest("PoolFullCloudsnap",
			"Make pool full and validate cloudsnaps",
			nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})

	stepLog := "Make pool full and validate cloudsnaps"
	It(stepLog, func() {

		stepLog = "Create cloudsnap schedule and validate cloud cred"
		policyName := "intervalpolicy"
		Step(stepLog, func() {

			log.InfoD(stepLog)
			err := CreatePXCloudCredential()
			log.FailOnError(err, "failed to create cloud credential")
			contexts = make([]*scheduler.Context, 0)

			stepLog = fmt.Sprintf("create schedule policy %s", policyName)
			Step(stepLog, func() {
				log.InfoD(stepLog)

				schedPolicy, err := storkops.Instance().GetSchedulePolicy(policyName)
				if err != nil {
					retain := 4
					interval := 2
					log.InfoD("Creating a interval schedule policy %v with interval %v minutes", policyName, interval)
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

					_, err = storkops.Instance().CreateSchedulePolicy(schedPolicy)
					log.FailOnError(err, fmt.Sprintf("error creating a SchedulePolicy [%s]", policyName))
				}
			})

		})

		defer func() {
			err := storkops.Instance().DeleteSchedulePolicy(policyName)
			log.FailOnError(err, fmt.Sprintf("error deleting a SchedulePolicy [%s]", policyName))
		}()

		log.InfoD(stepLog)
		existingAppList := Inst().AppList

		selectedNode := GetNodeWithLeastSize()

		stNodes := node.GetStorageNodes()
		var secondReplNode node.Node
		for _, stNode := range stNodes {
			if stNode.Name != selectedNode.Name {
				secondReplNode = stNode
			}
		}

		if selectedNode.Name == "" {
			log.FailOnError(fmt.Errorf("no node with multiple pools exists"), "error identifying node with more than one pool")

		}
		log.Infof("Identified node [%s] for pool expansion", selectedNode.Name)

		repl1PoolUUID := selectedNode.Pools[0].Uuid

		repl1Pool, err := GetStoragePoolByUUID(repl1PoolUUID)
		log.FailOnError(err, "error getting storage pool with UUID [%s]", repl1PoolUUID)

		repl2Pool := secondReplNode.Pools[0]
		isjournal, err := IsJournalEnabled()
		log.FailOnError(err, "Failed to check if Journal enabled")

		//expanding to repl2 pool so that it won't go to storage down state
		if (repl2Pool.TotalSize / units.GiB) <= (repl1Pool.TotalSize/units.GiB)*2 {
			expectedSize := (repl2Pool.TotalSize / units.GiB) * 2
			log.InfoD("Current Size of the pool %s is %d", repl2Pool.Uuid, repl2Pool.TotalSize/units.GiB)
			err = Inst().V.ExpandPool(repl2Pool.Uuid, api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK, expectedSize, true)
			dash.VerifyFatal(err, nil, "Pool expansion init successful?")
			resizeErr := waitForPoolToBeResized(expectedSize, repl2Pool.Uuid, isjournal)
			dash.VerifyFatal(resizeErr, nil, fmt.Sprintf("Verify pool %s on node %s expansion using resize-disk", repl2Pool.Uuid, secondReplNode.Name))
		}

		stepLog = fmt.Sprintf("Fill up  pool [%s] in node [%s] and validate cloudsnaps", repl1Pool.Uuid, selectedNode.Name)
		Step(stepLog, func() {
			log.InfoD(stepLog)

			poolLabelToUpdate := make(map[string]string)
			nodesToDisableProvisioning := make([]string, 0)
			poolsToDisableProvisioning := make([]string, 0)

			defer func() {
				//Reverting the provisioning changes done for the test
				Inst().AppList = existingAppList
				err = Inst().V.SetClusterOpts(*selectedNode, map[string]string{
					"--disable-provisioning-labels": ""})
				log.FailOnError(err, fmt.Sprintf("error removing cluster options disable-provisioning-labels"))
				err = Inst().S.RemoveLabelOnNode(*selectedNode, k8s.NodeType)
				log.FailOnError(err, "error removing label on node [%s]", selectedNode.Name)
				err = Inst().S.RemoveLabelOnNode(secondReplNode, k8s.NodeType)
				log.FailOnError(err, "error removing label on node [%s]", secondReplNode.Name)

				poolLabelToUpdate[k8s.NodeType] = ""
				poolLabelToUpdate["provision"] = ""
				// Update the pool label
				for _, p := range selectedNode.Pools {
					err = Inst().V.UpdatePoolLabels(*selectedNode, p.Uuid, poolLabelToUpdate)
					log.FailOnError(err, "Failed to update the label [%v] on the pool [%s] on node [%s]", poolLabelToUpdate, repl1Pool.Uuid, selectedNode.Name)
				}

			}()

			//Disabling provisioning on the other nodes/pools  and enabling only on selected pools for making sure the metadata node is full
			err = Inst().S.AddLabelOnNode(*selectedNode, k8s.NodeType, k8s.FastpathNodeType)
			log.FailOnError(err, fmt.Sprintf("Failed add label on node %s", selectedNode.Name))
			err = Inst().S.AddLabelOnNode(secondReplNode, k8s.NodeType, k8s.FastpathNodeType)
			log.FailOnError(err, fmt.Sprintf("Failed add label on node %s", secondReplNode.Name))

			for _, n := range stNodes {
				if n.VolDriverNodeID != selectedNode.VolDriverNodeID && n.VolDriverNodeID != secondReplNode.VolDriverNodeID {
					nodesToDisableProvisioning = append(nodesToDisableProvisioning, n.VolDriverNodeID)
				}
			}

			for _, p := range selectedNode.Pools {
				if p.Uuid != repl1Pool.Uuid {
					poolsToDisableProvisioning = append(poolsToDisableProvisioning, p.Uuid)
				}

			}
			for _, p := range secondReplNode.Pools {
				if p.Uuid != repl2Pool.Uuid {
					poolsToDisableProvisioning = append(poolsToDisableProvisioning, p.Uuid)
				}

			}

			poolLabelToUpdate[k8s.NodeType] = ""
			poolLabelToUpdate["provision"] = "disable"
			for _, p := range selectedNode.Pools {
				if p.Uuid != repl1Pool.Uuid {
					err = Inst().V.UpdatePoolLabels(*selectedNode, p.Uuid, poolLabelToUpdate)
					log.FailOnError(err, "Failed to update the label [%v] on the pool [%s] on node [%s]", poolLabelToUpdate, repl1Pool.Uuid, selectedNode.Name)

				}
			}

			clusterOptsVal := fmt.Sprintf("\"node=%s;provision=disable\"", strings.Join(nodesToDisableProvisioning, ","))
			err = Inst().V.SetClusterOpts(*selectedNode, map[string]string{
				"--disable-provisioning-labels": clusterOptsVal})
			log.FailOnError(err, fmt.Sprintf("error update cluster options disable-provisioning-labels with value [%s]", clusterOptsVal))

			Inst().AppList = []string{"fio-fastpath"}
			contexts = make([]*scheduler.Context, 0)
			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				contexts = append(contexts, ScheduleApplications(fmt.Sprintf("plflcs-%d", i))...)
			}
			ValidateApplications(contexts)
			defer appsValidateAndDestroy(contexts)
			nsList, err := core.Instance().ListNamespaces(map[string]string{"creator": "torpedo"})
			log.FailOnError(err, "error getting all namespaces")
			log.Infof("%v", nsList)
			appNamespaces := make([]string, 0)
			for _, ns := range nsList.Items {
				if strings.Contains(ns.Name, "plflcs") {
					appNamespaces = append(appNamespaces, ns.Name)
				}
			}

			if len(appNamespaces) == 0 {
				log.FailOnError(fmt.Errorf("no namespaces found to validate cloudsnaps"), "error getting cloudsnap namespaces")
			}

			stepLog = "validate cloudsnaps"
			Step(stepLog, func() {
				log.InfoD(stepLog)
				for _, ns := range appNamespaces {
					_, err = validateCloudSnaps(ns)
					log.FailOnError(err, fmt.Sprintf("error validating cloudsnaps in namespace [%s]", ns))
				}
			})

			err = WaitForPoolOffline(*selectedNode)
			log.FailOnError(err, fmt.Sprintf("Failed to make pool [%s] offline", repl1Pool.Uuid))

			stepLog = "validate cloudsnaps after pool full"
			Step(stepLog, func() {
				log.InfoD(stepLog)
				log.Infof("waiting for 2 mins to create new cloudsnaps")
				time.Sleep(2 * time.Minute)
				for _, ns := range appNamespaces {
					_, err = validateCloudSnaps(ns)
					log.FailOnError(err, fmt.Sprintf("error validating cloudsnaps in namespace [%s] after pool full", ns))
				}
			})

			expectedSize := (repl1Pool.TotalSize / units.GiB) * 2

			log.InfoD("Current Size of the pool %s is %d", repl1Pool.Uuid, repl1Pool.TotalSize/units.GiB)
			err = Inst().V.ExpandPool(repl1Pool.Uuid, api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK, expectedSize, true)
			dash.VerifyFatal(err, nil, "Pool expansion init successful?")
			resizeErr := waitForPoolToBeResized(expectedSize, repl1Pool.Uuid, isjournal)
			dash.VerifyFatal(resizeErr, nil, fmt.Sprintf("Verify pool %s on node %s expansion using resize-disk", repl1Pool.Uuid, selectedNode.Name))
			status, err := Inst().V.GetNodeStatus(*selectedNode)
			log.FailOnError(err, fmt.Sprintf("Error getting PX status of node %s", selectedNode.Name))
			dash.VerifySafely(*status, api.Status_STATUS_OK, fmt.Sprintf("validate PX status on node %s. Current status: [%s]", selectedNode.Name, status.String()))

			stepLog = "validate cloudsnaps after pool resize"
			Step(stepLog, func() {
				log.InfoD(stepLog)
				log.Infof("waiting for 2 mins to create new cloudsnaps")
				time.Sleep(2 * time.Minute)
				for _, ns := range appNamespaces {
					_, err = validateCloudSnaps(ns)
					log.FailOnError(err, fmt.Sprintf("error validating cloudsnaps in namespace [%s] after pool resize", ns))
				}
			})

		})

	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
		bucketName, err := GetCloudsnapBucketName(contexts)
		log.FailOnError(err, "error getting cloud snap bucket name")
		opts := make(map[string]bool)
		DestroyApps(contexts, opts)
		DeleteCloudSnapBucket(bucketName)
		AfterEachTest(contexts, testrailID, runID)
	})
})

var _ = Describe("{NFSProxyVolumeValidation}", Label("p1", "positive", "px_vol_ops"), func() {
	var contexts []*scheduler.Context
	JustBeforeEach(func() {
		StartTorpedoTest("NFSProxyVolumeValidation", "Validate PX operations with NFS proxy volumes", nil, 0)
	})

	It("schedule proxy volumes on applications, run CRUD, tear down", func() {
		var masterNode node.Node
		stepLog = "setup proxy server necessary for proxy volume"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			masterNodes := node.GetMasterNodes()
			if len(masterNodes) == 0 {
				log.FailOnError(fmt.Errorf("no master nodes found"), "Identifying master node of proxy server failed")
			}

			masterNode = masterNodes[0]
			err = SetupProxyServer(masterNode)
			log.FailOnError(err, fmt.Sprintf("error setting up proxy server on master node %s", masterNode.Name))

		})
		stepLog = "create storage class for proxy volumes"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			addresses := masterNode.Addresses
			if len(addresses) == 0 {
				log.FailOnError(fmt.Errorf("no addresses found for node [%s]", masterNode.Name), "error getting ip addresses ")
			}
			err = CreateNFSProxyStorageClass("portworx-proxy-volume-volume", addresses[0], "/exports/testnfsexportdir")
			log.FailOnError(err, "error creating storage class for proxy volume")
		})

		stepLog = "create apps with proxy volumes"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			appList := Inst().AppList

			defer func() {
				Inst().AppList = appList
			}()

			Inst().AppList = []string{"nginx-proxy-deployment"}
			contexts = make([]*scheduler.Context, 0)

			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				contexts = append(contexts, ScheduleApplications(fmt.Sprintf("nfsproxytest-%d", i))...)
			}

			for _, ctx := range contexts {
				log.InfoD("Validating application [%s]", ctx.App.Key)
				ctx.SkipVolumeValidation = true //skipping as volume does not have the mount path inside the pod
				ValidateContext(ctx)
			}
		})

		stepLog = "restart PX on all nodes one by one"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for _, appNode := range node.GetStorageDriverNodes() {
				stepLog = fmt.Sprintf("stop volume driver %s on node: %s",
					Inst().V.String(), appNode.Name)
				Step(stepLog,
					func() {
						log.InfoD(stepLog)
						StopVolDriverAndWait([]node.Node{appNode})
					})
				time.Sleep(20 * time.Second)

				stepLog = fmt.Sprintf("starting volume %s driver on node %s",
					Inst().V.String(), appNode.Name)
				Step(stepLog,
					func() {
						log.InfoD(stepLog)
						StartVolDriverAndWait([]node.Node{appNode})
					})

				stepLog = "Giving few seconds for volume driver to stabilize"
				Step(stepLog, func() {
					log.InfoD(stepLog)
					time.Sleep(20 * time.Second)
				})
			}
			Step("validate apps after PX restart", func() {
				for _, ctx := range contexts {
					log.InfoD("Validating application [%s]", ctx.App.Key)
					ctx.SkipVolumeValidation = true //skipping as volume does not have the mount path inside the pod
					ValidateContext(ctx)
				}
			})
			PerformSystemCheck()

		})

		opts := make(map[string]bool)
		opts[scheduler.OptionsWaitForResourceLeakCleanup] = true

		for _, ctx := range contexts {
			TearDownContext(ctx, opts)
		}

	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts)
	})
})

var _ = Describe("{SharedVolFuseTest}", Label("p1", "positive", "px_vol_ops", "shared_v4"), func() {
	/*
					https://portworx.atlassian.net/browse/PWX-35639
				   https://portworx.atlassian.net/browse/PTX-21805



			  		1.Get the list of storage nodes where sv4 service and sv4 volumes are attached
					2.Stop/Start PX on each storage node filtered in step 1
					3.Validate PX on the node
			   		4.Repeat this in a loop for 10 iterations
		            5. Validate the applications
	*/
	var testrailID = 12133434
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/cases/view/35259
	var runID int
	var contexts []*scheduler.Context
	JustBeforeEach(func() {
		StartTorpedoTest("SharedVolFuseTest", "Validate PX operations after sharedv4 and sharedv4 svc volumes  failover multiple times", nil, 0)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})

	It("schedule sharedv4 and sharedv4_svc volumes and perform failover of the coordinator node", func() {

		appList := make([]string, 0)
		stepLog = "create sharedv4 and sharedv4_svc apps "
		Step(stepLog, func() {
			log.InfoD(stepLog)

			for _, appName := range Inst().AppList {

				if strings.Contains(appName, "shared") || strings.Contains(appName, "svc") {
					appList = append(appList, appName)
				}
			}

			if len(appList) == 0 {
				log.FailOnError(fmt.Errorf("sharedv4 or sharedv4 svc apps are mandatory for the test"), "no sharedv4 or sharedv4 svc apps found to deploy")
			}

			contexts = make([]*scheduler.Context, 0)

			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				contexts = append(contexts, ScheduleApplications(fmt.Sprintf("pxfusetest-%d", i))...)
			}

			for _, ctx := range contexts {
				log.InfoD("Validating application [%s]", ctx.App.Key)
				ValidateContext(ctx)
			}
		})

		stNodes := node.GetStorageNodes()

		nodesToRestart := make(map[string]bool)
		sharedVols := make([]*volume.Volume, 0)

		for _, stNode := range stNodes {
			nodesToRestart[stNode.Name] = false
		}

		stepLog = "restart PX on storage nodes"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			numIter := 10

			//Getting the volumes of sharedv4 and sharedv4 svc apps
			for _, ctx := range contexts {
				if Contains(appList, ctx.App.Key) {
					appVols, err := Inst().S.GetVolumes(ctx)
					log.FailOnError(err, fmt.Sprintf("error getting volumes for app [%s]", ctx.App.Key))
					sharedVols = append(sharedVols, appVols...)
				}
			}

			for i := 1; i <= numIter; i++ {

				log.Infof("Running Iteration: #%d", i)

				//Getting the coordinator nodes of sharedv4 and sharedv4 svc volumes
				for _, appVol := range sharedVols {
					attachedNode, err := Inst().V.GetNodeForVolume(appVol, 1*time.Minute, 5*time.Second)
					log.FailOnError(err, fmt.Sprintf("error getting attached node for volume [%s]", appVol.Name))
					nodesToRestart[attachedNode.Name] = true
				}

				for _, stNode := range stNodes {
					//Restarting Px only if sharedv4 or sharedv4 svc volume is attached to the provided node
					if nodesToRestart[stNode.Name] {
						StopVolDriverAndWait([]node.Node{stNode})
						log.Infof("waiting for 1 min before starting PX for volumes corordinator to failover")
						time.Sleep(1 * time.Second)
						StartVolDriverAndWait([]node.Node{stNode})
						status, err := IsPxRunningOnNode(&stNode)
						log.FailOnError(err, "error checking px status on node [%s]", stNode.Name)
						dash.VerifyFatal(status, true, fmt.Sprintf("verfiy px is running on node [%s]", stNode.Name))
					}
				}

				//Setting it false to obtain refreshed nodes once volumes failover
				for _, stNode := range stNodes {
					nodesToRestart[stNode.Name] = false
				}

			}

			Step("validate apps after all failovers", func() {
				for _, ctx := range contexts {
					log.InfoD("Validating application [%s]", ctx.App.Key)
					ValidateContext(ctx)
				}
			})
			PerformSystemCheck()

		})

		opts := make(map[string]bool)
		opts[scheduler.OptionsWaitForResourceLeakCleanup] = true

		for _, ctx := range contexts {
			TearDownContext(ctx, opts)
		}

	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})
})

var _ = Describe("{FioClonedVolumeFaultInjection}", Label("p1", "negative", "px_vol_ops", "error_injection", "shared_v4"), func() {
	/*
		https://portworx.atlassian.net/browse/PTX-15687
			1. Create 1 Volume,  Run fio
			2. Create the Clone of Volume in step 1 , Run Fio on the cloned volume
			3. Perform HA increase/Decrease
			4. Perform Volume resize
			4. Inject faults ( like portworx restart)
			6. Repeat step 1-5 for 10 iterations
	*/
	JustBeforeEach(func() {
		StartTorpedoTest("FioClonedVolumeFaultInjectoin", "Create fio clone volume and inject faults,HA increase and volume resize", nil, 0)
	})

	itLog := "FioClonedVolumeFaultInjectoin"
	It(itLog, func() {
		log.InfoD(itLog)

		var secretId = "secret"
		var secretValue = "password"
		selectedNode := node.GetStorageDriverNodes()[0]

		numberOfTotalVolumes := 10
		numberofVolumeCreationInParallel := 5

		numberofIterations := numberOfTotalVolumes / numberofVolumeCreationInParallel
		var Wg sync.WaitGroup

		for j := 0; j < numberofIterations; j++ {
			var once sync.Once
			stepLog := "Create a secret using pxctl secrets kvdb"
			Step(stepLog, func() {
				log.InfoD(stepLog)
				cmd := fmt.Sprintf("pxctl secrets kvdb login | pxctl secrets kvdb put-secret --secret_id %v --secret_value %v", secretId, secretValue)

				out, err := Inst().N.RunCommandWithNoRetry(selectedNode, cmd, node.ConnectionOpts{
					Timeout:         2 * time.Minute,
					TimeBeforeRetry: 10 * time.Second,
				})
				log.FailOnError(err, "Unable to execute the pxctl show command")
				log.InfoD("Succesfully created secrets for secure volume: %v", out)
			})

			for i := 0; i < numberofVolumeCreationInParallel; i++ {
				// Create a px secure volume using pxctl.
				volName := fmt.Sprintf("fio-clone-fault-injection-%d", i)
				var cloneVol string
				Wg.Add(1)
				go func(i int) {
					defer Wg.Done()
					defer GinkgoRecover()

					stepLog = "Create 1 Volume,  Run fio"
					Step(stepLog, func() {
						log.InfoD(stepLog)

						pxctlCreateVolumeCmd := fmt.Sprintf("volume create --secure --size 10 %v --secret_key %v", volName, secretId)
						output, err := runPxctlCommand(pxctlCreateVolumeCmd, selectedNode, nil)
						log.FailOnError(err, "Failed to create volume using pxctl")
						log.InfoD("Successfully created volume: %v", output)

						//attach volume to host
						attachCmd := fmt.Sprintf("pxctl host attach %s --secret_key %v", volName, secretId)
						cmdConnectionOpts := node.ConnectionOpts{
							Timeout:         15 * time.Second,
							TimeBeforeRetry: 5 * time.Second,
							Sudo:            true,
						}

						_, err = Inst().N.RunCommandWithNoRetry(selectedNode, attachCmd, cmdConnectionOpts)
						log.FailOnError(err, "Failed to attach volume to host")

						err = writeFioDataToVolume(volName, selectedNode, 5)
						log.FailOnError(err, "Failed to write data to volume")

					})

					stepLog = "Create a clone of volume and run fio on the cloned volume"
					Step(stepLog, func() {
						log.InfoD(stepLog)
						cloneVol, err = Inst().V.CloneVolume(volName)
						log.FailOnError(err, "Failed to clone volume")
						log.InfoD("successfully create clone of volume :%v -> %v", volName, cloneVol)

						//attach volume to host
						attachCmd := fmt.Sprintf("pxctl host attach %s --secret_key %v", cloneVol, secretId)
						cmdConnectionOpts := node.ConnectionOpts{
							Timeout:         15 * time.Second,
							TimeBeforeRetry: 5 * time.Second,
							Sudo:            true,
						}

						_, err = Inst().N.RunCommandWithNoRetry(selectedNode, attachCmd, cmdConnectionOpts)
						log.FailOnError(err, "Failed to attach volume to host")

						err = writeFioDataToVolume(cloneVol, selectedNode, 2)
						log.FailOnError(err, "Failed to write data to volume")
					})

					stepLog = "Perform HA Increase/Decrease"
					Step(stepLog, func() {
						log.InfoD(stepLog)

						// Increase replication factor to 2
						pxctlHAUpdateCmd := fmt.Sprintf("v ha-update --repl 2 %v", volName)
						_, err = runPxctlCommand(pxctlHAUpdateCmd, selectedNode, nil)
						log.FailOnError(err, "Failed to increase replication factor to 2")
						log.InfoD("Successfully increase replication factor to 2")

						time.Sleep(2 * time.Minute)

						// Decrease replication factor to 1
						pxctlHAUpdateCmd = fmt.Sprintf("v ha-update --repl 1 %v", volName)
						_, err = runPxctlCommand(pxctlHAUpdateCmd, selectedNode, nil)
						log.FailOnError(err, "Failed to increase replication factor to 1")
						log.InfoD("Successfully increase replication factor to 1")

						time.Sleep(30 * time.Second)
					})

					stepLog = "Perform volume resize"
					Step(stepLog, func() {
						log.InfoD(stepLog)

						pxctlVolResizeCmd := fmt.Sprintf("v update %v --size %v", volName, 20)
						_, err = runPxctlCommand(pxctlVolResizeCmd, selectedNode, nil)
						log.FailOnError(err, "Failed to resize volume: %v", volName)
						log.InfoD("Succesfully resized volume: %v", volName)

						volInspect, err := Inst().V.InspectVolume(volName)
						log.FailOnError(err, "Failed to inspect volume")
						log.InfoD("Volume size: %v", volInspect.Spec.Size)

						time.Sleep(30 * time.Second)

					})
					stepLog = "Delete the volume and clone of the volume"
					Step(stepLog, func() {
						//unmount volume
						pxctlUnmountCmd := fmt.Sprintf("host unmount --path /var/lib/osd/mounts/%s %s", volName, volName)
						_, err = runPxctlCommand(pxctlUnmountCmd, selectedNode, nil)
						log.FailOnError(err, "Failed to unmount volume: %v", volName)
						log.InfoD("Succesfully unmounted volume: %v", volName)

						err = Inst().V.DeleteVolume(volName)
						log.FailOnError(err, "Failed to delete volume:%v", volName)

						pxctlUnmountCmd = fmt.Sprintf("host unmount --path /var/lib/osd/mounts/%s %s", cloneVol, cloneVol)
						_, err = runPxctlCommand(pxctlUnmountCmd, selectedNode, nil)
						log.FailOnError(err, "Failed to unmount volume: %v", volName)
						log.InfoD("Succesfully unmounted volume: %v", volName)

						//Delete the clone volume
						err = Inst().V.DeleteVolume(cloneVol)
						log.FailOnError(err, "Failed to delete volume:%v", cloneVol)
					})
				}(i)
			}
			Wg.Wait()

			once.Do(func() {
				stepLog = "Restart portworx where the volume is attached"
				Step(stepLog, func() {
					log.InfoD(stepLog)
					log.Infof("Stop volume driver [%s] on node: [%s]", Inst().V.String(), selectedNode.Name)
					StopVolDriverAndWait([]node.Node{selectedNode})
					log.Infof("Starting volume driver [%s] on node [%s]", Inst().V.String(), selectedNode.Name)
					StartVolDriverAndWait([]node.Node{selectedNode})
					time.Sleep(30 * time.Second)
				})
			})
		}
	})
})
var _ = Describe("{VolumePreCheck}", Label("p0", "positive", "px_vol_ops"), func() {
	/*
			https://portworx.atlassian.net/browse/PTX-20557
			1. Deploy a basic volume on the cluster
			2. Now we need to test the ha-update of the volume with different cases
			scenarios :
		    1. Test the ha-update sources option with an uuid of other nodes on which node is not present
		    2. Test the ha-update sources option with an ip of nodes
			3. Test the ha-update sources option with an invalid uuid of the node on which the repl of the volume is present
		    4. Test the ha-update sources option with a valid uuid of the node on which the repl of the volume is present

	*/
	JustBeforeEach(func() {
		StartTorpedoTest("volumeprecheck", "Precheck for volume operations", nil, 0)
	})

	It("volumeprecheck", func() {
		log.InfoD("volumeprecheck")
		stepLog := "Create a volume and perform a precheck for sources options"
		Step(stepLog, func() {
			log.InfoD(stepLog)

			stNodes := node.GetStorageDriverNodes()
			nodesuuidWithoutReplica := make([]string, 0)
			nodesIP := make([]string, 0)

			index := rand.Intn(len(stNodes))
			selectedNode := &stNodes[index]

			var aggr_level int
			var repl_level int
			storageNodes := node.GetStorageNodes()
			if len(storageNodes) >= 4 && len(storageNodes) < 9 {
				aggr_level = 2
				repl_level = 2
			} else if len(storageNodes) >= 9 {
				aggr_level = 3
				repl_level = 3
			} else {
				aggr_level = 2
				repl_level = 1
			}
			log.InfoD("Setting the aggr_level to %d and repl_level to %d as storage nodes in the cluster are %d", aggr_level, repl_level, len(storageNodes))

			id := uuid.New()
			volName := fmt.Sprintf("volume_%s", id.String()[:8])
			log.InfoD("Create a volume with a min size on node [%s]", selectedNode.Name)
			basicVolumeCreate := fmt.Sprintf("volume create -a %d --repl %d  %s", aggr_level, repl_level, volName)
			_, err := runPxctlCommand(basicVolumeCreate, *selectedNode, nil)
			log.FailOnError(err, "volume creation failed on the cluster with volume name [%s] ", volName)
			log.InfoD("Base Volume creation with volume name %s successful", volName)
			//find on which node volume got created
			volInspect, err := Inst().V.InspectVolume(volName)
			log.FailOnError(err, "Failed to inspect volume")
			log.InfoD("Volume created on node: %s", volInspect.ReplicaSets[0].Nodes[0])

			selectedNodeId := volInspect.ReplicaSets[0].Nodes[0]
			listofNodesVolumePlaced := volInspect.ReplicaSets[0].Nodes

			log.InfoD("List of nodes on which volume is placed: %v", listofNodesVolumePlaced)
			for _, stNode := range stNodes {
				if !Contains(listofNodesVolumePlaced, stNode.VolDriverNodeID) {
					nodesuuidWithoutReplica = append(nodesuuidWithoutReplica, stNode.Id)
					nodesIP = append(nodesIP, stNode.Addresses[0])
				}
			}
			if len(nodesuuidWithoutReplica) == 0 && len(listofNodesVolumePlaced) == len(storageNodes) {
				log.InfoD("Volume Cannot be placed on other nodes as all the nodes are already used for the volume creation")
				return
			}

			log.InfoD("Test the ha-update sources option with a uuid of other nodes on which node is not present")
			wrongUuidcmd := fmt.Sprintf("v ha-update %s --repl %d --sources %s", volName, repl_level+1, nodesuuidWithoutReplica[rand.Intn(len(nodesuuidWithoutReplica))])
			_, err = runPxctlCommand(wrongUuidcmd, node.GetStorageDriverNodes()[0], nil)
			if err != nil {
				isExpectedError := strings.Contains(err.Error(), "does not belong to volume's replication set")
				dash.VerifyFatal(isExpectedError, true, fmt.Sprintf("Expected error: %v", err))

			}

			log.InfoD("Test the ha-update sources option with a ip of nodes ")
			wrongIPcmd := fmt.Sprintf("v ha-update %s --repl %d --sources %s", volName, repl_level+1, nodesIP[rand.Intn(len(nodesIP))])
			_, err = runPxctlCommand(wrongIPcmd, node.GetStorageDriverNodes()[0], nil)
			if err != nil {
				isExpectedError := strings.Contains(err.Error(), "could not find any node with id")
				dash.VerifyFatal(isExpectedError, true, fmt.Sprintf("Expected error: %v", err))

			}

			log.InfoD("Test the ha-update sources option with an invalid uuid of the node on which the repl of the volume is present")
			randomUUID := uuid.New()
			invalidUuidcmd := fmt.Sprintf("v ha-update %s --repl %d --sources %s", volName, repl_level+1, randomUUID)
			_, err = runPxctlCommand(invalidUuidcmd, node.GetStorageDriverNodes()[0], nil)
			if err != nil {
				isExpectedError := strings.Contains(err.Error(), "Failed to update volume: could not find any node with id")
				dash.VerifyFatal(isExpectedError, true, fmt.Sprintf("Expected error: %v", err))

			}

			log.InfoD("Test the ha-update sources option with a valid uuid of the node on which the repl of the volume is present")
			validUuidcmd := fmt.Sprintf("v ha-update %s --repl %d --sources %s", volName, repl_level+1, selectedNodeId)
			_, err = runPxctlCommand(validUuidcmd, node.GetStorageDriverNodes()[0], nil)
			log.FailOnError(err, "Failed to update volume: %v", volName)
			log.InfoD("Successfully updated volume: %v", volName)

			log.InfoD("Delete the volume that is created for the test")
			deleteVolumeCmd := fmt.Sprintf("volume delete %s", volName)
			_, err = runPxctlCommand(deleteVolumeCmd, *selectedNode, nil)
			log.FailOnError(err, "Failed to delete volume: %v", volName)
		})
	})
})

func writeFioDataToVolume(volName string, n node.Node, size int64) error {
	mountPath := fmt.Sprintf("/var/lib/osd/mounts/%s", volName)
	creatDir := fmt.Sprintf("mkdir %s", mountPath)

	cmdConnectionOpts := node.ConnectionOpts{
		Timeout:         15 * time.Second,
		TimeBeforeRetry: 5 * time.Second,
		Sudo:            true,
	}

	log.Infof("Running command %s on %s", creatDir, n.Name)
	_, err := Inst().N.RunCommandWithNoRetry(n, creatDir, cmdConnectionOpts)

	if err != nil {
		return err
	}

	mountCmd := fmt.Sprintf("pxctl host mount --path %s %s", mountPath, volName)
	log.Infof("Running command %s on %s", mountCmd, n.Name)
	_, err = Inst().N.RunCommandWithNoRetry(n, mountCmd, cmdConnectionOpts)

	if err != nil {
		return err
	}

	writeCmd := fmt.Sprintf("fio --name=%s --ioengine=libaio --rw=write --bs=4k --numjobs=1 --size=%vG --iodepth=256 --directory=%s --output=/tmp/vol_write.log --verify=meta --direct=1 --randrepeat=1 --verify_pattern=0xbeddacef --end_fsync=1", volName, size, mountPath)

	log.Infof("Running command %s on %s", writeCmd, n.Name)
	_, err = Inst().N.RunCommandWithNoRetry(n, writeCmd, cmdConnectionOpts)

	if err != nil {
		return err
	}

	return nil

}

var _ = Describe("{OverCommitVolumeTest}", Label("p1", "positive", "px_vol_ops"), func() {
	/*
						    https://portworx.atlassian.net/browse/PTX-19103
							Total 5 scenarios Tested
							1. Verify Thick Provisioning on Specific Nodes when resizing the volume are honoured
							2. Verify Thick Provisioning (Global) OverCommitPercent when resizing the volume are honoured
							3. Update the pxctl cluster with cluster option OverCommitPercent with 200(Enabeling Thick Provisioning) on a specific Node and thin provisioning on the other nodes
							4. Thin Provisioning with Global and Node Specific Settings [300% on a certain node and 200% over commit on the other nodes]
							5. Disable all imposed cluster options and try creating thin provisioned volumes
		                    Process:
							1. Update the pxctl cluster with cluster option OverCommitPercent
						    2. Check the overall storage pool capacity of a particular node
						    3. create a volume with max of target size of the pool
						    4. Now update the volume size more than size of the storage pool
						    5. Check vol size is successful or not
						    6. Also validate the creation of volume with size more than available capacity on the node


	*/
	JustBeforeEach(func() {
		StartTorpedoTest("OverCommitVolumeTest", "Validate Overcommit volume size", nil, 0)
	})
	// check the size left in the node
	itLog := "honor OverCommitPercent when resizing the volume"
	It(itLog, func() {
		log.InfoD(itLog)
		getRandomPoolandCalculateSize := func(snapshotPercent uint64) (selectedNode *node.Node, targetSizeGiB uint64) {
			stNodes := node.GetStorageDriverNodes()
			index := rand.Intn(len(stNodes))
			selectedNode = &stNodes[index]
			pools := selectedNode.Pools
			poolToResize := pools[rand.Intn(len(pools))]
			poolIDToResize := poolToResize.Uuid
			originalSizeInBytes := poolToResize.TotalSize
			log.InfoD("Original size of the pool %s is %d of node %s ", poolIDToResize, originalSizeInBytes, selectedNode.Name)
			SnapshotPercent := snapshotPercent
			SubtractSize := (SnapshotPercent * originalSizeInBytes) / 100
			targetSizeInBytes := originalSizeInBytes - SubtractSize
			targetSizeGiB = targetSizeInBytes / units.GiB
			log.InfoD("Target size of the pool %s is %d", poolIDToResize, targetSizeGiB)
			return selectedNode, targetSizeGiB
		}
		CreateVolumeandValidate := func(selectedNode *node.Node, multiple uint64, targetSizeGiB uint64) {
			id := uuid.New()
			VolName := fmt.Sprintf("volume_%s", id.String()[:8])
			log.InfoD("Create a volume with a min size on node [%s]", selectedNode.Name)
			basicVolumeCreate := fmt.Sprintf("volume create --nodes %s %s", selectedNode.Id, VolName)
			_, err := runPxctlCommand(basicVolumeCreate, *selectedNode, nil)
			log.FailOnError(err, "volume creation failed on the cluster with volume name [%s]", VolName)
			log.InfoD("Base Volume creation with volume name %s successful", VolName)
			log.InfoD("Now Resize volume with a size of %d %d time of targetsize of pool on node [%s] as %d overcommit percent imposed", multiple*targetSizeGiB, multiple, selectedNode.Name, multiple*100)
			resizeVolumeCmd := fmt.Sprintf("volume update --size %d %s", multiple*targetSizeGiB, VolName)
			_, volCreateErr := runPxctlCommand(resizeVolumeCmd, *selectedNode, nil)
			log.FailOnError(volCreateErr, "volume resize failed  on the cluster with volume name [%s]", VolName)
			log.InfoD("Volume [%s] resized to %d GiB", VolName, multiple*targetSizeGiB)
			log.InfoD("Resize the volume more than %d times available capacity on the node [%s]", multiple, selectedNode.Name)
			resizeVolumeGreaterThanPoolSizeCmd := fmt.Sprintf("volume update --size %d %s", (multiple+1)*targetSizeGiB, VolName)
			_, err = runPxctlCommand(resizeVolumeGreaterThanPoolSizeCmd, *selectedNode, nil)
			if err != nil {
				IsExpectederr := strings.Contains(err.Error(), "Failed to resize volume")
				dash.VerifyFatal(IsExpectederr, true, err.Error())
			}

			err = Inst().V.DeleteVolume(VolName)
			log.FailOnError(err, "Failed to delete volume [%s]", VolName)
			log.InfoD("Successfully deleted volume [%s]", VolName)
			log.InfoD("Try Creating a New Volume with size more than %d times available capacity on the node [%s]", multiple, selectedNode.Name)
			id = uuid.New()
			VolName = fmt.Sprintf("volume_%s", id.String()[:8])
			volCreatecmd := fmt.Sprintf("volume create --size %d --nodes %s %s", (multiple+1)*targetSizeGiB, selectedNode.Id, VolName)
			_, volerr := runPxctlCommand(volCreatecmd, *selectedNode, nil)
			IsExpectederr := strings.Contains(volerr.Error(), "pools must not over-commit provisioning space")
			if volerr != nil {
				dash.VerifyFatal(IsExpectederr, true, volerr.Error())
			} else {
				PrintInspectVolume(VolName)
				dash.VerifyFatal(volerr, fmt.Errorf("Volume Creation should be failed"), "Volume should not be created as we have imposed the cluster options")
			}
			DisableClusterOptionscmd := "cluster options update  --provisioning-commit-labels '[]'"
			_, disable_err := runPxctlCommand(DisableClusterOptionscmd, *selectedNode, nil)
			log.FailOnError(disable_err, "Failed to set cluster options")
			log.InfoD("Successfully set cluster options")

		}

		//Case 1: Verify Thick Provisioning on Specific Nodes when resizing the volume are honoured
		stepLog = "Verify Thick Provisioning on Specific Nodes when resizing the volume are honoured "
		Step(stepLog, func() {
			log.InfoD(stepLog)
			selectedNode, targetSizeGiB := getRandomPoolandCalculateSize(30)
			SetClusterOptionscmdOnNode := fmt.Sprintf("cluster options update  --provisioning-commit-labels '[{\"OverCommitPercent\": 100, \"SnapReservePercent\": 30,\"LabelSelector\": {\"node\": \"%s\"}} ]'", selectedNode.Id)
			_, err := runPxctlCommand(SetClusterOptionscmdOnNode, *selectedNode, nil)
			log.FailOnError(err, "Failed to set cluster options")
			log.InfoD("Successfully set cluster options")
			ClusterOptionsValidationcmd := "cluster options list -j | jq -r '.ProvisionCommitRule'"
			output, err := runPxctlCommand(ClusterOptionsValidationcmd, *selectedNode, nil)
			log.InfoD("The Current Cluster options: %v", output)
			CreateVolumeandValidate(selectedNode, 1, targetSizeGiB)
		})

		// Case 2: Verify Thick Provisioning (Global) OverCommitPercent when resizing the volume are honoured
		stepLog = "Verify Thick Provisioning (Global) OverCommitPercent when resizing the volume are honoured "
		Step(stepLog, func() {
			log.InfoD(stepLog)
			selectedNode, targetSizeGiB := getRandomPoolandCalculateSize(15)
			SetClusterOptionscmd := "cluster options update  --provisioning-commit-labels '[{\"OverCommitPercent\": 100, \"SnapReservePercent\": 15}]'"
			_, err := runPxctlCommand(SetClusterOptionscmd, *selectedNode, nil)
			log.FailOnError(err, "Failed to set cluster options")
			log.InfoD("Successfully set cluster options")
			ClusterOptionsValidationcmd := "cluster options list -j | jq -r '.ProvisionCommitRule'"
			output, err := runPxctlCommand(ClusterOptionsValidationcmd, *selectedNode, nil)
			log.InfoD("The Current Cluster options: %v", output)
			CreateVolumeandValidate(selectedNode, 1, targetSizeGiB)

		})

		//Case 3 : Verify Thin Provisioning 200% (Global)  when resizing the volume are honoured
		stepLog = "Update the pxctl cluster with cluster option OverCommitPercent with 200(Enabeling Thick Provisioning)"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			selectedNode, targetSizeGiB := getRandomPoolandCalculateSize(15)
			SetClusterOptionscmd := "cluster options update  --provisioning-commit-labels '[{\"OverCommitPercent\": 200, \"SnapReservePercent\": 15}]'"
			_, err := runPxctlCommand(SetClusterOptionscmd, *selectedNode, nil)
			log.FailOnError(err, "Failed to set cluster options")
			log.InfoD("Successfully set cluster options")
			ClusterOptionsValidationcmd := "cluster options list -j | jq -r '.ProvisionCommitRule'"
			output, err := runPxctlCommand(ClusterOptionsValidationcmd, *selectedNode, nil)
			log.InfoD("The Current Cluster options: %v", output)
			CreateVolumeandValidate(selectedNode, 2, targetSizeGiB)

		})

		//Case 4: Thin Provisioning with Global and Node Specific Settings
		stepLog = "Update the pxctl cluster with cluster option OverCommitPercent with 200(Enabeling Thick Provisioning) on a specific Node and thin provisioning on the other nodes"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			selectedNode, targetSizeGiB := getRandomPoolandCalculateSize(30)
			SetClusterOptionscmd := fmt.Sprintf("cluster options update  --provisioning-commit-labels '[{\"OverCommitPercent\": 300, \"SnapReservePercent\": 30,\"LabelSelector\": {\"node\": \"%s\"}}, {\"OverCommitPercent\": 200, \"SnapReservePercent\": 15}]'", selectedNode.Id)
			_, err := runPxctlCommand(SetClusterOptionscmd, *selectedNode, nil)
			log.FailOnError(err, "Failed to set cluster options")
			log.InfoD("Successfully set cluster options")
			ClusterOptionsValidationcmd := "cluster options list -j | jq -r '.ProvisionCommitRule'"
			output, err := runPxctlCommand(ClusterOptionsValidationcmd, *selectedNode, nil)
			log.InfoD("The Current Cluster options: %v", output)
			CreateVolumeandValidate(selectedNode, 3, targetSizeGiB)

			log.Info("Try Creating a New Volume with size more than 200% available capacity on any other node ")
			stNodes := node.GetStorageDriverNodes()
			for _, node := range stNodes {
				if node.Name != selectedNode.Name {
					SetClusterOptionscmd := fmt.Sprintf("cluster options update  --provisioning-commit-labels '[{\"OverCommitPercent\": 300, \"SnapReservePercent\": 30,\"LabelSelector\": {\"node\": \"%s\"}}, {\"OverCommitPercent\": 200, \"SnapReservePercent\": 15}]'", selectedNode.Id)
					_, err := runPxctlCommand(SetClusterOptionscmd, node, nil)
					log.FailOnError(err, "Failed to set cluster options")
					id := uuid.New()
					VolName := fmt.Sprintf("volume_%s", id.String()[:8])
					volCreatecmd := fmt.Sprintf("volume create --size %d --nodes %s %s", 3*targetSizeGiB, node.Id, VolName)
					_, volerr := runPxctlCommand(volCreatecmd, node, nil)
					if volerr != nil {
						IsExpectederr := strings.Contains(volerr.Error(), "pools must not over-commit provisioning space")
						dash.VerifyFatal(IsExpectederr, true, volerr.Error())

					}
					break
				}

			}

		})
		//Case 5 : Disable all imposed cluster options and try creating thin provisioned volumes
		stepLog = "Create a volume again with size greater than Storage pool size as we have disabled the thick provisioning (Should Be created)"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			selectedNode, targetSizeGiB := getRandomPoolandCalculateSize(0)
			SetClusterOptionscmd := "cluster options update  --provisioning-commit-labels '[]'"
			_, err := runPxctlCommand(SetClusterOptionscmd, *selectedNode, nil)
			log.FailOnError(err, "Failed to set cluster options")
			log.InfoD("Successfully set cluster options")
			ClusterOptionsValidationcmd := "cluster options list -j | jq -r '.ProvisionCommitRule'"
			output, err := runPxctlCommand(ClusterOptionsValidationcmd, *selectedNode, nil)
			log.InfoD("The Current Cluster options: %v", output)
			id := uuid.New()
			VolName := fmt.Sprintf("volume_%s", id.String()[:8])
			volerr := Inst().V.CreateVolumeUsingPxctlCmd(*selectedNode, VolName, 3*targetSizeGiB, 1)
			log.FailOnError(volerr, "volume creation failed on the cluster with volume name [%s]", VolName)
			log.InfoD("Volume created with name [%s]", VolName)
			//Delete the Volume , As we have created it only for Validation Purpose
			err = Inst().V.DeleteVolume(VolName)
			log.FailOnError(err, "Failed to delete volume [%s]", VolName)

		})

	})
})
var _ = Describe("{RestartPxandRestartNode}", Label("p1", "negative", "px_vol_ops", "error_injection", "node_reboot", "px_restart"), func() {
	/*
	   https://purestorage.atlassian.net/browse/PTX-24483
	   1.Deploy Applications
	   2.Validate Applications are Deployed
	   3.Restart Portworx service on few nodes
	   4.once portworx is up on the node,immediately reboot the node
	   5.Make sure Both portworx and node are up.
	   6.Validate the Applications are running
	*/
	var contexts []*scheduler.Context
	JustBeforeEach(func() {
		StartTorpedoTest("RestartPxandRestartNode",
			"Restart Portworx and Restart Node", nil, 0)
	})
	itLog := "RestartPxandRestartNode"
	It(itLog, func() {
		log.InfoD(itLog)
		pxNodes := node.GetStorageDriverNodes()
		selectedNodesForReboot := pxNodes[:len(pxNodes)/2]
		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			taskName := "restartpxandrebootnode"
			Provisioner := fmt.Sprintf("%v", portworx.PortworxCsi)
			context, err := Inst().S.Schedule(taskName, scheduler.ScheduleOptions{
				AppKeys:            Inst().AppList,
				StorageProvisioner: Provisioner,
				Namespace:          taskName,
			})
			log.FailOnError(err, "Failed to schedule application of %v namespace", taskName)
			contexts = append(contexts, context...)
		}
		ValidateApplications(contexts)
		defer DestroyApps(contexts, nil)
		stepLog := "Restart Portworx Service on few nodes and once portworx is up, immediately reboot the node"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for _, nodeToReboot := range selectedNodesForReboot {
				log.InfoD("Restarting portworx  Service on Node [%v]", nodeToReboot.Name)
				err := Inst().V.RestartDriver(nodeToReboot, nil)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Failed to restart portworx on node [%v]", nodeToReboot.Name))
				log.InfoD("Restarted portworx on  node %s", nodeToReboot.Name)
				err = Inst().N.RebootNode(nodeToReboot,
					node.RebootNodeOpts{
						Force: true,
						ConnectionOpts: node.ConnectionOpts{
							Timeout:         defaultCommandTimeout,
							TimeBeforeRetry: defaultCommandRetry,
						},
					})
				log.FailOnError(err, "Failed to reboot node %v", nodeToReboot.Name)
				nodeReadyStatus := func() (interface{}, bool, error) {
					err := Inst().S.IsNodeReady(nodeToReboot)
					if err != nil {
						return "", true, err
					}
					return "", false, nil
				}
				log.InfoD("wait for node: %s to be back up", nodeToReboot.Name)
				_, err = DoRetryWithTimeoutWithGinkgoRecover(nodeReadyStatus, 10*time.Minute, 35*time.Second)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying the status of rebooted node %s", nodeToReboot.Name))
				err = Inst().V.WaitDriverUpOnNode(nodeToReboot, Inst().DriverStartTimeout)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying the node driver status of rebooted node %s", nodeToReboot.Name))
			}

		})
		stepLog = "Validate the applications are in running state"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			ValidateApplications(contexts)
		})
	})
	JustAfterEach(func() {
		EndTorpedoTest()
		AfterEachTest(contexts)
	})
})

// Verify volume delete from the new node- Place volumes on the new node and enable trash can features on all volume.
var _ = Describe("{EnableTrashCanForvolume}", Label("p0", "positive", "px_vol_ops", "staging"), Label("p2", "positive", "px_vol_ops", "trashcan"), func() {
	/*
		Step1: Take storage node from cluster
		Step3: Create a  few Volumes on the new node.
		step4: Once volume are created on the node, enable Trashcan. ( Enabling trashcan is a cluster wide operation , value that you set for trashcan can be 10 min for Automation)
		Step5: we need wait for 10 min update to trashcan   ( wait for 10 min for Volumes to be in trashcan before deletion so that once volume is deleted it will go to trashcan. Volumes deleted will be there for 10 min as we did set the time for trashcan is 10 min ) [ NO Need to wait after updaing trashcan param ]
		Step 6: once you have trashcan enabled , delete a volume which was created earlier on the new Node [ Now you should be waiting for 10 min before checking if volume is deleted ]
		make sure the volume deleted should go to trashcan .
		Step7:waiting for 10 min should delete the volumes under trashcan.
	*/
	var testrailID = 0
	JustBeforeEach(func() {
		StartTorpedoTest("EnableTrashCanForvolume", "Create multiple volumes on node and delete and validate trashcan volume expiry", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})
	stepLog := "Create volumes on the storage node "
	It(stepLog, func() {
		log.InfoD(stepLog)
		contexts = make([]*scheduler.Context, 0)
		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("storagenodecreatevolume-%d", i))...)
		}
		ValidateApplications(contexts)
		defer appsValidateAndDestroy(contexts)
		log.InfoD("Pick the one Storage node in cluster")
		pickNode := node.GetStorageNodes()[0]
		stepLog = "Create volumes on cluster"
		var volumes []string
		number_of_volumes := 200
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for i := 0; i < number_of_volumes; i++ {
				volumeName := fmt.Sprintf("Trashcanvolume%d", i)
				log.Infof("Creating volume : %s", volumeName)
				pxctlCmdFull := fmt.Sprintf("v c %s -s 20 ", volumeName)
				output, err := Inst().V.GetPxctlCmdOutput(pickNode, pxctlCmdFull)
				volumes = append(volumes, volumeName)
				log.FailOnError(err, fmt.Sprintf("error creating volume %s", volumeName))
				log.Infof(output)
			}
			stepLog = "Enable Trashcan"
			Step(stepLog, func() {
				log.InfoD(stepLog)
				err := Inst().V.SetClusterOptsWithConfirmation(pickNode, map[string]string{
					"--volume-expiration-minutes": "10",
				})
				log.FailOnError(err, "error while enabling trashcan")
				log.InfoD("Trashcan is successfully enabled")
			})
			stepLog := "Delete volumes on the node"
			Step(stepLog, func() {
				log.InfoD("Deleting the Volumes")
				for _, volumeName := range volumes {
					err = deletePXVolume(volumeName)
					log.FailOnError(err, "Failed to delete volume [%v]", volumeName)
				}
			})
			trashcanVols := make([]string, 0)
			stepLog = "validate volumes in trashcan"
			Step(stepLog, func() {
				// wait for few seconds for pvc to get deleted and volume to get detached
				time.Sleep(20 * time.Second)
				log.InfoD(stepLog)
				trashcanVolsNew, err := Inst().V.GetTrashCanVolumeIds(pickNode)
				log.FailOnError(err, "error While getting trashcan volumes")
				for _, vol := range trashcanVolsNew {
					if vol = strings.ReplaceAll(vol, " ", ""); vol != "" {
						trashcanVols = append(trashcanVols, vol)
					}
				}
				log.InfoD("Listed the trashcan volumes in the node:%s", trashcanVols)
				log.Infof("trashcan len: %d", len(trashcanVols))
				dash.VerifyFatal(len(trashcanVols) > 0, true, "validate volumes exist in trashcan")

			})
			trashcanVolumes := make([]string, 0)
			stepLog = "validate volumes are permanently delete in trashcan"
			Step(stepLog, func() {
				log.InfoD(stepLog)
				// wait for expiration volume delete in the trashcan
				log.InfoD("waiting for volumes to delete from trashcan")
				expirationDuration := 15 * time.Minute
				time.Sleep(expirationDuration)
				log.InfoD("Get the volumes in trashcan")
				trashcanVolumesNew, err := Inst().V.GetTrashCanVolumeIds(pickNode)
				log.FailOnError(err, "error while getting trashcan volumes after expiry time")
				for _, vol := range trashcanVolumesNew {
					if vol = strings.ReplaceAll(vol, " ", ""); vol != "" {
						trashcanVols = append(trashcanVols, vol)
					}
				}
				log.Infof("trashcanvolume len: %d", len(trashcanVolumes))
				if len(trashcanVolumes) == 0 {
					log.Infof("Volumes are permanently delete from trashcan")
				} else {
					err := fmt.Errorf("Volumes are not deleted in trashcan")
					log.FailOnError(err, "Volumes are still in trashcan")
				}

			})
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})
})

// For each volume, get replica nodes, bring PX down, bring it back up, and ensure PX and pods are running
var _ = Describe("{BringVolumeoutofQuorum}", Label("p1", "positive", "px_vol_ops"), Label("p0", "negative", "px_ops"), func() {
	/*
		For each volume, get it's replicas
		Stop PX on all replica nodes
		Wait for 10 minutes to stop. 10 minutes is needed because if the volume goes out of quorum PX can hang for 10 minutes for the IO to abort
		Bring PX back up
		Ensure PX comes up
		Ensure app pods are fine
	*/
	var testrailID = 0
	JustBeforeEach(func() {
		StartTorpedoTest("BringVolumeoutofQuorum", "For each volume, get replica nodes, bring PX down, bring it back up, and ensure PX and pods are running", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})
	var contexts []*scheduler.Context
	stepLog := "For each volume, get replica nodes, bring PX down, bring it back up, and ensure PX and pods are running"
	It(stepLog, func() {
		log.InfoD(stepLog)
		var err error
		contexts = make([]*scheduler.Context, 0)
		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("volumeout-%d", i))...)
		}

		ValidateApplications(contexts)
		defer appsValidateAndDestroy(contexts)
		stepLog := "Get Replicaset for the volumes and  and then bring PX down,start PX service, ensure pods and PX are running"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for _, ctx := range contexts {
				var appVolumes []*volume.Volume
				stepLog = fmt.Sprintf("get volumes for %s app", ctx.App.Key)
				Step(stepLog, func() {
					log.InfoD(stepLog)
					appVolumes, err = Inst().S.GetVolumes(ctx)
					log.FailOnError(err, "Failed to get volumes for app %s", ctx.App.Key)
				})
				var replicaSets []*opsapi.ReplicaSet
				var nodes []string
				volume := appVolumes[0]
				log.Infof("selected volume from the context: %v", volume)
				replicaSets, err = Inst().V.GetReplicaSets(volume)
				log.FailOnError(err, "Failed to get Replica Sets for volume : %v", volume)
				for _, replicaSet := range replicaSets {
					nodes = append(nodes, replicaSet.Nodes...)
				}
				var nodeDetails []node.Node
				Step(fmt.Sprintf("Stop PX on the replica node for app %s", ctx.App.Key), func() {
					// Collect node details
					for _, nodeID := range nodes {
						nodeInfo, err := node.GetNodeDetailsByNodeID(nodeID)
						log.FailOnError(err, "Error getting node details for node [%s]", nodeID)
						nodeDetails = append(nodeDetails, nodeInfo)
					}
					// Stop the PX service on the collected node details
					log.InfoD("Stopping PX service on nodes: %+v", nodeDetails)
					StopVolDriverAndWait(nodeDetails[:len(nodeDetails)-1])
					log.InfoD("Waiting for 10 minutes for volume to be out of quorum. I/O abort by PX.")
					time.Sleep(10 * time.Minute)
					log.InfoD("Successfully stopped PX service on nodes: %+v", nodeDetails[:len(nodeDetails)-1])

				})
				var PxserviceNode node.Node
				Step("Check the replica status is not in quorum", func() {
					log.InfoD("Stopping PX service on nodes: %+v", nodeDetails[len(nodeDetails)-1:])
					StopVolDriverAndWait(nodeDetails[len(nodeDetails)-1:])
					log.InfoD("Waiting for 2 minutes after stopping PX service")
					time.Sleep(2 * time.Minute)
					log.InfoD("Successfully stopped PX service on nodes: %+v", nodeDetails[len(nodeDetails)-1:])
					log.InfoD("Starting Px service on nodes: %+v", nodeDetails[:1])
					StartVolDriverAndWait(nodeDetails[:1])
					log.InfoD("Successfully start PX service on nodes: %+v", nodeDetails[:1])
					PxserviceNode = nodeDetails[0]
					replStatus, err := GetVolumeReplicationStatusOnPxservicenode(PxserviceNode, volume)
					log.FailOnError(err, "Failed to get replication status for volume:%v", volume)
					log.Infof("Replication status for volume:%v", replStatus)
					dash.VerifyFatal(replStatus == "Not in quorum", true, "Verified status 'Not in quorum' for the volume")
				})
				Step("Initialize and start the Px service on all available nodes", func() {
					// Start the Px service on the collected node details.
					log.InfoD("Starting Px service on nodes: %+v", nodeDetails)
					StartVolDriverAndWait(nodeDetails[1:])
					log.InfoD("Successfully started Px service on nodes: %+v", nodeDetails)
				})
				Step("Check the replica status is up", func() {
					time.Sleep(5 * time.Minute)
					replStatus, err := GetVolumeReplicationStatus(volume)
					log.FailOnError(err, "Failed to get replication status for volume:%v", volume)
					log.Infof("Replication status for volume:%v", replStatus)
					log.InfoD(replStatus)
					dash.VerifyFatal(replStatus == "Up", true, "Verified status to be 'Up' for volume")
				})
				Step("Verify that the pod and Px are ready on all nodes", func() {
					for _, node := range nodeDetails {
						isPodReady := Inst().V.IsPxReadyOnNode(node)
						if isPodReady {
							log.InfoD("Pod and Px are running and healthy on node: %s", node)
						} else {
							err := fmt.Errorf("pod and Px are not running or not healthy on node: %s", node)
							log.FailOnError(err, "pod and Px verification failed for the node")
						}
					}

				})
			}

		})
		stepLog = "Validate the applications are in running state"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			ValidateApplications(contexts)
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})
})

var _ = Describe("{VerifySanpWhenVolumeDown}", Label("p0", "negative", "px_ops"), func() {
	/*
		Ticket ID:https://purestorage.atlassian.net/browse/HAZEL-156
		Steps:
		Create volume
		Write some data
		Create snapshot schedule
		Verify snap schedule is working
		Make volume down
		Expected:
		Scheduled snap should not happen.
	*/
	var testrailID = 0
	JustBeforeEach(func() {
		StartTorpedoTest("VerifySanpWhenVolumeDown", "For the volume, create a snapshot schedule, then bring the volume down. Verify that the snapshot should not occur again.", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})
	var contexts []*scheduler.Context
	stepLog := "For the volume, create a snapshot schedule, then bring the volume down. Verify that the snapshot should not occur again."
	It(stepLog, func() {
		log.InfoD(stepLog)
		var (
			err                     error
			appVolumes              []*volume.Volume
			nodeDetails             []node.Node
			replicaSets             []*opsapi.ReplicaSet
			nodes                   []string
			PxserviceNode           node.Node
			volume                  *volume.Volume
			snapshotScheduleName    string
			snapshotNames           []string
			snapshotNamesVolumedown []string
		)
		contexts = make([]*scheduler.Context, 0)
		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("volumeout-%d", i))...)
		}
		ValidateApplications(contexts)
		defer DestroyApps(contexts, nil)
		stepLog := "Get the volume from context, create a snapshot schedule for that volume"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for _, ctx := range contexts {
				appNamespace := ctx.App.Key + "-" + ctx.UID
				stepLog = fmt.Sprintf("get volumes for %s app", ctx.App.Key)
				Step(stepLog, func() {
					log.InfoD(stepLog)
					appVolumes, err = Inst().S.GetVolumes(ctx)
					volume = appVolumes[0]
					log.Infof("selected volume from the context: %v", volume)
					log.FailOnError(err, "Failed to get volumes for app %s", ctx.App.Key)
				})
				stepLog = fmt.Sprintf("create schedule policy for %s app", ctx.App.Key)
				Step(stepLog, func() {
					log.InfoD(stepLog)
					policyName := "localintervalpolicy"
					schedPolicy, err := storkops.Instance().GetSchedulePolicy(policyName)
					if err != nil {
						snapshotInterval := 15
						log.InfoD("Creating a interval schedule policy %v with interval %v minutes", policyName, snapshotInterval)
						schedPolicy = &storkv1.SchedulePolicy{
							ObjectMeta: metav1.ObjectMeta{
								Name: policyName,
							},
							Policy: storkv1.SchedulePolicyItem{
								Interval: &storkv1.IntervalPolicy{
									IntervalMinutes: snapshotInterval,
								},
							}}
						_, err = storkops.Instance().CreateSchedulePolicy(schedPolicy)
						log.FailOnError(err, "Unable to create schedule policy")
						log.Infof("Waiting for 15 mins for Snapshots to be completed")
						time.Sleep(15 * time.Minute)
					} else {
						log.Infof("schedPolicy is %v already exists", schedPolicy.Name)
					}
				})
				stepLog = "Scheduling snapshot creation for volume"
				stepLog = fmt.Sprintf("Stop PX on the replica node for app %s", ctx.App.Key)
				Step(stepLog, func() {
					replicaSets, err = Inst().V.GetReplicaSets(volume)
					log.FailOnError(err, "Failed to get Replica Sets for volume : %v", volume)
					for _, replicaSet := range replicaSets {
						nodes = append(nodes, replicaSet.Nodes...)
					}

					for _, nodeID := range nodes {
						nodeInfo, err := node.GetNodeDetailsByNodeID(nodeID)
						log.FailOnError(err, "Error getting node details for node [%s]", nodeID)
						nodeDetails = append(nodeDetails, nodeInfo)
					}
					// Stop the PX service on the collected node details
					log.InfoD("Stopping PX service on nodes: %+v", nodeDetails[:len(nodeDetails)-1])
					StopVolDriverAndWait(nodeDetails[:len(nodeDetails)-1])
					log.InfoD("Waiting for 10 minutes for volume to be out of quorum. I/O abort by PX.")
					time.Sleep(10 * time.Minute)
					log.InfoD("Successfully stopped PX service on nodes: %+v", nodeDetails[:len(nodeDetails)-1])

				})
				Step("Check the replica status is not in quorum", func() {
					log.InfoD("Stopping PX service on nodes: %+v", nodeDetails[len(nodeDetails)-1:])
					StopVolDriverAndWait(nodeDetails[len(nodeDetails)-1:])
					log.InfoD("Waiting for 2 minutes after stopping PX service")
					time.Sleep(2 * time.Minute)
					log.InfoD("Successfully stopped PX service on nodes: %+v", nodeDetails[len(nodeDetails)-1:])
					log.InfoD("Starting Px service on nodes: %+v", nodeDetails[:1])
					StartVolDriverAndWait(nodeDetails[:1])
					log.InfoD("Successfully start PX service on nodes: %+v", nodeDetails[:1])
					snapshotScheduleName = volume.Name + "-interval-schedule"
					snapMap := make(map[storkv1.SchedulePolicyType][]*storkv1.ScheduledVolumeSnapshotStatus)
					newSnapStatuses, err := storkops.Instance().GetSnapshotSchedule(snapshotScheduleName, appNamespace)
					log.FailOnError(err, "Failed to retrieve snapshot schedule")
					for v, snapshotStatuses := range newSnapStatuses.Status.Items {
						if len(snapshotStatuses) > 0 {
							var statuses []*storkv1.ScheduledVolumeSnapshotStatus
							for _, status := range snapshotStatuses {
								if status == nil {
									err := fmt.Errorf("Snapshot not found for volume")
									log.FailOnError(err, "Failed to find snapshot for the specified volume")
								}
								if status.Status == snapv1.VolumeSnapshotConditionReady {
									statuses = append(statuses, status)
								}
							}
							snapMap[v] = statuses
						}
					}
					for _, snapname := range snapMap["Interval"] {
						snapshotNames = append(snapshotNames, snapname.Name)
					}
					log.Infof("Snapshot before volume down %v", len(snapMap["Interval"]))
					dash.VerifyFatal(len(snapMap["Interval"]) > 0, true, "validate snapshot is created for that volume ")
					PxserviceNode = nodeDetails[0]
					log.InfoD("Volume name :%v", volume)
					replStatus, err := GetVolumeReplicationStatusOnPxservicenode(PxserviceNode, volume)
					log.FailOnError(err, "Failed to get replication status for volume:%v", volume)
					log.Infof("Replication status for volume:%v", replStatus)
					dash.VerifyFatal(replStatus == "Not in quorum", true, "Verified status 'Not in quorum' for the volume")
					log.Infof("Wait for 15 mins snapshot not take again")
					time.Sleep(15 * time.Minute)

				})
				stepLog = "Validate: Snapshot should not occur after volume is down."
				Step(stepLog, func() {
					snapMap := make(map[storkv1.SchedulePolicyType][]*storkv1.ScheduledVolumeSnapshotStatus)
					newSnapStatuses, err := storkops.Instance().GetSnapshotSchedule(snapshotScheduleName, appNamespace)
					log.FailOnError(err, "unable to get snapschedule")
					for v, snapshotStatuses := range newSnapStatuses.Status.Items {
						if len(snapshotStatuses) > 0 {
							var statuses []*storkv1.ScheduledVolumeSnapshotStatus
							for _, status := range snapshotStatuses {
								if status == nil {
									err := fmt.Errorf("Snapshot not found for volume")
									log.FailOnError(err, "Failed to find snapshot for the specified volume")
								}
								if status.Status == snapv1.VolumeSnapshotConditionReady {
									statuses = append(statuses, status)
								}
							}
							snapMap[v] = statuses
							log.Infof("Snapshot status for this volume: %+v", snapMap)
						}
					}
					for _, snapname := range snapMap["Interval"] {
						snapshotNamesVolumedown = append(snapshotNamesVolumedown, snapname.Name)

					}
					log.Infof("Sanpshot count before:%d volumes are down and snaphot names:%v", len(snapshotNames), snapshotNames)
					log.Infof("Sanpshot count after:%d volumes are down and snaphot names:%v", len(snapshotNamesVolumedown), snapshotNamesVolumedown)
					dash.VerifyFatal(len(snapshotNames) == len(snapshotNamesVolumedown), true, "snapshot is created after volume down")
					snapshotCount := make(map[string]int)
					for _, name := range snapshotNames {
						snapshotCount[name]++
					}
					for _, name := range snapshotNamesVolumedown {
						snapshotCount[name]--
					}

					ismatch := true
					for _, count := range snapshotCount {
						if count != 0 {
							ismatch = false
							break
						}
					}
					dash.VerifyFatal(ismatch, true, "Snapshots do not match after volume down")
				})

				Step("Initialize and start the Px service on all available nodes", func() {
					// Start the Px service on the collected node details.
					log.InfoD("Starting Px service on nodes: %+v", nodeDetails)
					StartVolDriverAndWait(nodeDetails[1:])
					log.InfoD("Successfully started Px service on nodes: %+v", nodeDetails)
				})
				Step("Verify that the pod and Px are ready on all nodes", func() {
					for _, node := range nodeDetails {
						isPodReady := Inst().V.IsPxReadyOnNode(node)
						if isPodReady {
							log.InfoD("Pod and Px are running and healthy on node: %s", node)
						} else {
							err := fmt.Errorf("pod and Px are not running or not healthy on node: %s", node)
							log.FailOnError(err, "pod and Px verification failed for the node")
						}
					}

				})
			}

		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})

})

var _ = Describe("{DetachVolSnapshotTest}", func() {
	/*
		1. Deploy Applications
		2. Validate Applications are Deployed
		3. Scale down the applications so that the volumes are detached
		4. Take snapshot of one volume
		5. Inspect the parent volume, make sure the labels are intact, e.g. namespace, pvc
		6. Inspect the labels on the snapshot are expected
	*/
	var contexts []*scheduler.Context
	var err error
	JustBeforeEach(func() {
		StartTorpedoTest("DetachVolSnapshotTest",
			"Validate labels are present on detached volumes after taking snapshot", nil, 0)
	})
	itLog := "Check labels after detached volumes getting snapshotted"
	It(itLog, func() {
		log.InfoD(itLog)

		stepLog := "Schedule apps"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			contexts = make([]*scheduler.Context, 0)
			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				contexts = append(contexts, ScheduleApplications(fmt.Sprintf("detached-vol-%d", i))...)
			}
			ValidateApplications(contexts)
		})

		stepLog = "Scale down apps to detach volumes"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			// scaleDownApp scales an app to zero replicas using the given context and waits for pods to terminate
			scaleDownApp := func(ctx *scheduler.Context) error {
				scaleApp(ctx, 0)
				waitForPodsToTerminate := func() (interface{}, bool, error) {
					vols, err := Inst().S.GetVolumes(ctx)
					if err != nil {
						return nil, false, err
					}
					podCount := 0
					for _, vol := range vols {
						if vol.ID == "" {
							return nil, false, fmt.Errorf("empty vol.ID in volume [%v]", vol)
						}
						pods, err := core.Instance().GetPodsUsingPV(vol.ID)
						if err != nil {
							return nil, false, err
						}
						podCount += len(pods)
					}
					if podCount > 0 {
						return nil, true, fmt.Errorf("expected no pods, but found [%d] remaining", podCount)
					}
					return nil, false, nil
				}
				_, err := task.DoRetryWithTimeout(waitForPodsToTerminate, 10*time.Minute, 30*time.Second)
				if err != nil {
					return fmt.Errorf("failed to scale down app [%s] and ensure all pods are deleted. Err: [%v]", ctx.App.Key, err)
				}
				return nil
			}
			for _, ctx := range contexts {
				log.InfoD("Scaling down app [%s]", ctx.App.Key)
				err := scaleDownApp(ctx)
				log.FailOnError(err, "failed to scale down app [%s]", ctx.App.Key)
			}
		})

		var volume *volume.Volume
		newLabels := map[string]string{"foo": "bar"}
		stepLog = "Get one detached volume and update it"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for _, ctx := range contexts {
				vols, err := Inst().S.GetVolumes(ctx)
				log.FailOnError(err, "failed to get volumes for app [%s]", ctx.App.Key)
				if len(vols) > 0 {
					volume = vols[0]
					break
				}
			}
			if volume == nil {
				err := fmt.Errorf("unable to find any volume in the cluster")
				log.FailOnError(err, err.Error())
			}

			err = Inst().V.UpdateVolumeLabels(volume, newLabels)
			log.FailOnError(err, fmt.Sprintf("failed to update labels %v for volume %s", newLabels, volume.ID))
		})

		stepLog = "Get parent volume labels before snapshotting"
		var parentVolumeLabels map[string]string
		Step(stepLog, func() {
			log.InfoD(stepLog)
			parentVolInspectResult, err := Inst().V.InspectVolume(volume.ID)
			log.FailOnError(err, "Failed to inspect volume %v", volume.ID)
			parentVolumeLabels = parentVolInspectResult.Locator.VolumeLabels
		})

		stepLog = "Make a snapshot out of the volume"
		var snapshotResponse *opsapi.SdkVolumeSnapshotCreateResponse
		Step(stepLog, func() {
			log.InfoD(stepLog)
			snapshotName := fmt.Sprintf("snapshot_%s", volume.ID)
			snapshotResponse, err = Inst().V.CreateSnapshot(volume.ID, snapshotName)
			log.FailOnError(err, "error creeating snapshot out of volume [%s]", volume.ID)
			log.InfoD("Snapshot [%s] created with ID [%s]", snapshotName, snapshotResponse.GetSnapshotId())
		})

		verifyVolumeLabels := func(volumeId string, expectedLabels map[string]string) bool {
			volInspect, err := Inst().V.InspectVolume(volumeId)
			log.FailOnError(err, "Failed to inspect volume %v", volumeId)
			pass := true
			for label, expectedValue := range expectedLabels {
				value, ok := volInspect.Locator.VolumeLabels[label]
				if !ok {
					log.Errorf("unable to find label [%s] in the labels from the volume inspect result", label)
					pass = false
				}
				if value != expectedValue {
					log.Errorf("label [%s] from the volume inspect result is not expected. Expected: [%s], found: [%s]", label, expectedValue, value)
					pass = false
				}
			}
			return pass

		}

		stepLog = "Verify parent volume having labels as expected"
		expectedParentVolumeLabels := make(map[string]string, 0)
		for _, label := range []string{
			"namespace",
			"pvc",
		} {
			expectedParentVolumeLabels[label] = parentVolumeLabels[label]
		}
		for label, value := range newLabels {
			expectedParentVolumeLabels[label] = value
		}
		Step(stepLog, func() {
			log.InfoD(stepLog)
			pass := verifyVolumeLabels(volume.ID, expectedParentVolumeLabels)
			dash.VerifyFatal(pass, true, "Volume labels of the volume is not as expected.")
		})

		stepLog = "Verify snapshot having labels as expected"
		expectedSnapshotLabels := make(map[string]string, 0)
		for label, value := range newLabels {
			expectedSnapshotLabels[label] = value
		}
		Step(stepLog, func() {
			log.InfoD(stepLog)
			pass := verifyVolumeLabels(snapshotResponse.GetSnapshotId(), expectedSnapshotLabels)
			dash.VerifyFatal(pass, true, "Volume labels of the snapshot is not as expected.")
		})
	})
	JustAfterEach(func() {
		EndTorpedoTest()
		AfterEachTest(contexts)
	})
})

// Verify Delete volumes from trashcan using Volume expiration minutes.
var _ = Describe("{DeleteVolFromTrashCanWithVEM}", Label("p0", "positive", "px_ops", "px_vol_ops", "pure_ops", "trashcan", "staging"), func() {
	/*
		Step1: Take storage node from cluster
		Step3: Scheduled Applications on the node
		Step4: Validated those Applications
		step5: Once volume are created on the node, enable Trashcan.
		Step6: we need wait for 10 min update to Trashcan.
		Step7: Once you have trashcan enabled , delete a volume which was created earlier on the new Node
		make sure the volume deleted should go to trashcan.
		Step8:waiting for 10 min should delete the volumes under trashcan.
	*/
	var testrailID = 86093862
	JustBeforeEach(func() {
		StartTorpedoTest("DeleteVolFromTrashCanWithVEM", "Create multiple volumes on node and delete and validate trashcan volume expiry", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})
	stepLog := "Create volumes on the storage node "
	It(stepLog, func() {
		log.InfoD(stepLog)
		contexts = make([]*scheduler.Context, 0)

		log.InfoD("Pick the one Storage node in cluster")
		pickNode := node.GetStorageNodes()[0]
		stepLog = "Create volumes on cluster"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			stepLog = "Scheduling the Applications"
			Step(stepLog, func() {
				log.InfoD(stepLog)
				for i := 0; i < Inst().GlobalScaleFactor; i++ {
					contexts = append(contexts, ScheduleApplications(fmt.Sprintf("storagenodecreatevolume-%d", i))...)
				}
				log.InfoD("Applications are scheduled successfully")
			})

			stepLog = "Validate the applications"
			Step(stepLog, func() {
				log.InfoD(stepLog)
				ValidateApplications(contexts)
				log.InfoD("scheduled Applications are validated")
			})

			stepLog = "Enable Trashcan"
			Step(stepLog, func() {
				log.InfoD(stepLog)
				err := Inst().V.SetClusterOptsWithConfirmation(pickNode, map[string]string{
					"--volume-expiration-minutes": "10",
				})
				log.FailOnError(err, "error while enabling trashcan")
				log.InfoD("Trashcan is successfully enabled")
			})
			stepLog := "Delete volumes on the node"
			Step(stepLog, func() {
				log.InfoD("Deleting the Volumes with destroy Apps")
				appsValidateAndDestroy(contexts)
			})
			trashcanVols := make([]string, 0)
			stepLog = "validate volumes in trashcan"
			Step(stepLog, func() {
				// wait for few seconds for pvc to get deleted and volume to get detached
				time.Sleep(20 * time.Second)
				log.InfoD(stepLog)
				trashcanVolsNew, err := Inst().V.GetTrashCanVolumeIds(pickNode)
				log.FailOnError(err, "error While getting trashcan volumes")
				for _, vol := range trashcanVolsNew {
					if vol = strings.ReplaceAll(vol, " ", ""); vol != "" {
						trashcanVols = append(trashcanVols, vol)
					}
				}
				log.InfoD("Listed the trashcan volumes in the node:%s", trashcanVols)
				log.Infof("trashcan len: %d", len(trashcanVols))
				dash.VerifyFatal(len(trashcanVols) > 0, true, "validate volumes exist in trashcan")

			})
			trashcanVolumes := make([]string, 0)
			stepLog = "validate volumes are permanently delete in trashcan"
			Step(stepLog, func() {
				log.InfoD(stepLog)
				// wait for expiration volume delete in the trashcan
				log.InfoD("waiting for volumes to delete from trashcan")
				expirationDuration := 11 * time.Minute
				time.Sleep(expirationDuration)
				log.InfoD("Get the volumes in trashcan")
				trashcanVolumesNew, err := Inst().V.GetTrashCanVolumeIds(pickNode)
				log.FailOnError(err, "error while getting trashcan volumes after expiry time")
				for _, vol := range trashcanVolumesNew {
					if vol = strings.ReplaceAll(vol, " ", ""); vol != "" {
						trashcanVols = append(trashcanVols, vol)
					}
				}
				log.Infof("trashcanvolume len: %d", len(trashcanVolumes))
				if len(trashcanVolumes) == 0 {
					log.Infof("Volumes are permanently deleted from trashcan")
				} else {
					err := fmt.Errorf("Volumes are not deleted in trashcan")
					log.FailOnError(err, "Volumes are still in trashcan")
				}

			})
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})
})

func restartPx(nodeForPxStop node.Node) {
	// restart portworx on node.
	err = Inst().V.StopDriver([]node.Node{nodeForPxStop}, false, nil)
	dash.VerifyFatal(err == nil, true, fmt.Sprintf("Stop px driver"))
	err = Inst().V.WaitDriverDownOnNode(nodeForPxStop)
	dash.VerifyFatal(err == nil, true, fmt.Sprintf("Wait for driver down"))
	err = Inst().V.StartDriver(nodeForPxStop)
	dash.VerifyFatal(err == nil, true, fmt.Sprintf("Start px driver"))
	err = Inst().V.WaitDriverUpOnNode(nodeForPxStop, 5*time.Minute)
	dash.VerifyFatal(err == nil, true, fmt.Sprintf("Wait for driver to start"))
}

func verifyNoError(err error, description string) {
	dash.VerifyFatal(err == nil, true, description)
}

func createPod(pvc *corev1.PersistentVolumeClaim) *corev1.Pod {
	podSpec := &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod-" + pvc.Name,
			Namespace: pvc.Namespace,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "nginx-container",
					Image: "nginx:latest",
					Ports: []corev1.ContainerPort{
						{
							ContainerPort: 80,
						},
					},
					VolumeMounts: []corev1.VolumeMount{
						{
							MountPath: "/usr/share/nginx/html",
							Name:      "nginx-volume",
						},
					},
				},
			},
			Volumes: []corev1.Volume{
				{
					Name: "nginx-volume",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: pvc.Name,
						},
					},
				},
			},
		},
	}
	log.Infof("Creating nginx pod from pvc")
	pod, err := k8sCore.CreatePod(podSpec)
	log.FailOnError(err, "Failed to create pod")

	t := func() (interface{}, bool, error) {
		pod, err := k8sCore.GetPodByName(pod.Name, pod.Namespace)
		if err != nil {
			return "", false, err
		}
		if !k8sCore.IsPodReady(*pod) {
			return "", true, fmt.Errorf("waiting for pod %s to be in running state", pod.Name)
		}
		return "", false, nil
	}
	_, err = task.DoRetryWithTimeout(t, 5*time.Minute, 30*time.Second)
	log.FailOnError(err, "Pod did not go to running state")
	return pod
}

func createStorageClass(scName string, parameters map[string]string) {
	reclaimPolicyDelete := corev1.PersistentVolumeReclaimDelete
	bindMode := storageApi.VolumeBindingImmediate

	v1obj := metav1.ObjectMeta{
		Name: scName,
	}
	scObj := storageApi.StorageClass{
		ObjectMeta:        v1obj,
		Provisioner:       k8s.CsiProvisioner,
		ReclaimPolicy:     &reclaimPolicyDelete,
		VolumeBindingMode: &bindMode,
		Parameters:        parameters,
	}

	k8sStorage := storage.Instance()
	_, err = k8sStorage.CreateStorageClass(&scObj)
	if err != nil {
		isStorageClassExists := strings.Contains(err.Error(), "already exists")
		dash.VerifyFatal(isStorageClassExists, true, "Check if storage class exists")
	} else {
		log.InfoD("Successfully created storage class: %v", scName)
	}
}

func createVolumeSnapshotClass(snapShotClassName string, params map[string]string) {
	volSnapshotClass, err := Inst().S.CreateCSISnapshotClass(scheduler.CSISnapshotClassCreateRequest{
		SnapClassName:  snapShotClassName,
		DeletionPolicy: "Delete",
		Parameters:     params,
	})
	if err != nil {
		isSnapshotClassExists := strings.Contains(err.Error(), "already exists")
		dash.VerifyFatal(isSnapshotClassExists, true, "Check if snapshot class exists")
	} else {
		log.InfoD("Successfully created volume snapshot class: %v", volSnapshotClass.Name)
	}
}

func create50GiReadWriteOncePVC(pvcName string, ns string, scName string) *corev1.PersistentVolumeClaim {
	log.InfoD("creating PVC [%s] in namespace [%s]", pvcName, ns)
	pvcObj := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pvcName,
			Namespace: ns,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse("50Gi"),
				},
			},
			StorageClassName: &scName,
		},
	}
	_, err := core.Instance().CreatePersistentVolumeClaim(pvcObj)
	dash.VerifyFatal(err, nil, fmt.Sprintf("Verify PVC [%s] is created successfully", pvcName))
	time.Sleep(10 * time.Second)
	pvc, err := core.Instance().GetPersistentVolumeClaim(pvcName, ns)
	log.FailOnError(err, "Failed to create PVC [%v]. Error : [%v]", pvcName, err)
	err = Inst().S.WaitForSinglePVCToBound(pvcName, ns, 3)
	dash.VerifyFatal(err, nil, fmt.Sprintf("Verify PVC [%s] got bound successfully.", pvc.Name))
	return pvc
}

func validatePodCreationWithPVCName(restoredPVCSpec *corev1.PersistentVolumeClaim) {
	podSpec := &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod-" + restoredPVCSpec.Name,
			Namespace: restoredPVCSpec.Namespace,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "nginx-container",
					Image: "nginx:latest",
					Ports: []corev1.ContainerPort{
						{
							ContainerPort: 80,
						},
					},
					VolumeMounts: []corev1.VolumeMount{
						{
							MountPath: "/usr/share/nginx/html",
							Name:      "nginx-volume",
						},
					},
				},
			},
			Volumes: []corev1.Volume{
				{
					Name: "nginx-volume",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: restoredPVCSpec.Name,
						},
					},
				},
			},
		},
	}
	log.Infof("Creating nginx pod from restored spec")
	pod, err := k8sCore.CreatePod(podSpec)
	if err != nil {
		log.FailOnError(err, "Failed to create pod from snapshot")
	}
	defer func() {
		err := k8sCore.DeletePod(pod.Name, pod.Namespace, false)
		if err != nil {
			log.Warnf("Failed to delete pod %s: %v", pod.Name, err)
		}
	}()

	t := func() (interface{}, bool, error) {
		pod, err := k8sCore.GetPodByName(pod.Name, pod.Namespace)
		if err != nil {
			return "", false, err
		}
		if !k8sCore.IsPodReady(*pod) {
			return "", true, fmt.Errorf("waiting for pod %s to be in running state", pod.Name)
		}
		return "", false, nil
	}
	_, err = task.DoRetryWithTimeout(t, 5*time.Minute, 30*time.Second)
	log.FailOnError(err, "Failed to create pods from snapshot")
}

func cleanupSnapshotests(context *scheduler.Context, ns string) {
	err = Inst().S.DeletePodsFromNamespace(context, ns)
	dash.VerifySafely(err, nil, fmt.Sprintf("Deleting pods in namespace [%s]", ns))

	err = Inst().S.DeletePvcsFromNamespace(context, ns)
	dash.VerifySafely(err, nil, fmt.Sprintf("Deleting PVCs in namespace [%s]", ns))

	err = Inst().S.DeleteCsiSnapshotsFromNamespace(context, ns)
	dash.VerifySafely(err, nil, fmt.Sprintf("Deleting snapshots in namespace [%s]", ns))

	log.Infof("Deleting namespace[%s]", ns)
	err = core.Instance().DeleteNamespace(ns)
	dash.VerifySafely(err, nil, fmt.Sprintf("Deleting namespace[%s]", ns))
}

func createNamespace(ns string) {
	nsName := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: ns,
		},
	}
	log.InfoD("Creating namespace %v", ns)
	_, err = k8sCore.CreateNamespace(nsName)
	log.FailOnError(err, "failed to create namespace")
}

var _ = Describe("{SnapValidateWithCredRecreate}", Label("staging", "p0", "negative", "px_ops"), func() {
	/*
		https://purestorage.atlassian.net/browse/HAZEL-290
		Step1: Depoly app
		Step2: create cloud cred and scheduled cloudsnapshot for app
		step3: Validate snapshot is create
		Step3: Delete creds and validate cloudsnap fails
		Step4: Create creds again and validate cloudsnaps working
	*/
	JustBeforeEach(func() {
		StartTorpedoTest("SnapValidateWithCredRecreate", "Validate cloudsnap status after cloud creds delete and check cloudsnap status is ready again new cred is created.", nil, 0)
	})

	var contexts []*scheduler.Context
	stepLog := "Schedule a cloud snapshot using credentials, delete old credentials, recreate new credentials, and verify continued snapshot access and status with the new credentials"
	It(stepLog, func() {
		log.InfoD(stepLog)
		contexts = make([]*scheduler.Context, 0)
		retain := 8
		interval := 5

		err := CreatePXCloudCredential()
		log.FailOnError(err, "failed to create cloud credential")
		defer DeletePXCloudCredential()
		n := node.GetStorageDriverNodes()[0]
		uuidCmd := "pxctl cred list -j | grep uuid"
		output, err := runCmd(uuidCmd, n)
		log.FailOnError(err, "error getting uuid for cloudsnap credential")
		if output == "" {
			log.FailOnError(fmt.Errorf("cloud cred is not created"), "Check for cloud cred exists?")
		}

		credUUID := strings.Split(strings.TrimSpace(output), " ")[1]
		credUUID = strings.ReplaceAll(credUUID, "\"", "")
		log.Infof("Got Cred UUID: %s", credUUID)
		contexts = make([]*scheduler.Context, 0)
		policyName := "intervalpolicy"
		stepLog = fmt.Sprintf("create schedule policy %s", policyName)

		Step(stepLog, func() {
			log.InfoD(stepLog)
			schedPolicy, err := storkops.Instance().GetSchedulePolicy(policyName)
			if err != nil {

				log.InfoD("Creating a interval schedule policy %v with interval %v minutes", policyName, interval)
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

				_, err = storkops.Instance().CreateSchedulePolicy(schedPolicy)
				log.FailOnError(err, fmt.Sprintf("error creating a SchedulePolicy [%s]", policyName))
			}

			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				contexts = append(contexts, ScheduleApplications(fmt.Sprintf("snapvalidate-%d", i))...)
			}

			ValidateApplications(contexts)

		})

		defer func() {
			err := storkops.Instance().DeleteSchedulePolicy(policyName)
			log.FailOnError(err, fmt.Sprintf("error deleting a SchedulePolicy [%s]", policyName))
		}()

		stepLog = "Verify that cloud snap status"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			err := ValidateSnapshot(contexts)
			log.FailOnError(err, "Error during snapshot validation")
			log.InfoD("Snapshot validation completed successfully")
		})
		stepLog = "Delete cloud credentials"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			credDeleteCmd := fmt.Sprintf("cred delete %s", credUUID)
			output, err = Inst().V.GetPxctlCmdOutputConnectionOpts(n, credDeleteCmd, node.ConnectionOpts{
				IgnoreError:     false,
				TimeBeforeRetry: defaultRetryInterval,
				Timeout:         defaultTimeout,
			}, false)

			if err != nil {
				err = fmt.Errorf("error deleting existing cred [%s], cause: %v", credUUID, err)
				log.FailOnError(err, "failed to delete cloud credentials")
			}

			log.Infof("Deleted cloud cred [%s] successfully", output)

		})
		stepLog = "Verify that the snapshot returns an error when attempting to access it after deleting the cloud credentials."
		Step(stepLog, func() {
			log.InfoD(stepLog)
			log.Info("Waiting for the next scheduled snapshot to be triggered.")
			time.Sleep(5 * time.Minute)
			for _, ctx := range contexts {
				if !strings.Contains(ctx.App.Key, "cloudsnap") {
					continue
				}
				var appVolumes []*volume.Volume
				var err error
				appNamespace := ctx.App.Key + "-" + ctx.UID
				log.Infof("Namespace: %v", appNamespace)
				stepLog = fmt.Sprintf("Getting app volumes for volume %s", ctx.App.Key)
				Step(stepLog, func() {
					log.InfoD(stepLog)
					appVolumes, err = Inst().S.GetVolumes(ctx)
					log.FailOnError(err, "error getting volumes for [%s]", ctx.App.Key)

					if len(appVolumes) == 0 {
						log.FailOnError(fmt.Errorf("no volumes found for [%s]", ctx.App.Key), "error getting volumes for [%s]", ctx.App.Key)
					}
				})
				log.Infof("Got volume count : %v", len(appVolumes))
				log.FailOnError(err, "error validating volumes for [%s]", ctx.App.Key)
				for _, v := range appVolumes {
					snapshotScheduleName := v.Name + "-interval-schedule"
					log.InfoD("snapshotScheduleName : %v for volume: %s", snapshotScheduleName, v.Name)

					resp, err := storkops.Instance().GetSnapshotSchedule(snapshotScheduleName, appNamespace)
					log.FailOnError(err, fmt.Sprintf("error getting snapshot schedule for [%s], volume:[%s] in namespace [%s]", snapshotScheduleName, v.Name, v.Namespace))
					dash.VerifyFatal(len(resp.Status.Items) > 0, true, fmt.Sprintf("verify snapshots exists for [%s]", snapshotScheduleName))
					snapshotstatuserror := false
					for _, snapshotStatuses := range resp.Status.Items {
						if len(snapshotStatuses) > 0 {
							status := snapshotStatuses[len(snapshotStatuses)-1]
							if status == nil {
								log.FailOnError(fmt.Errorf("SnapshotSchedule has an empty migration in it's most recent status"), fmt.Sprintf("error getting latest snapshot status for [%s]", snapshotScheduleName))
							}
							log.Infof("Snapshot [%s] has status [%v]", status.Name, status.Status)
							if status.Status == snapv1.VolumeSnapshotConditionError {
								snapshotstatuserror = true
								resp, _ := storkops.Instance().GetSnapshotSchedule(snapshotScheduleName, appNamespace)
								log.Infof("SnapshotSchedule resp: %+v", resp)
								snapData, _ := Inst().S.GetSnapShotData(ctx, status.Name, appNamespace)
								if snapData != nil {
									log.Infof("snapData : %v", snapData)
								}
								break

							}

						}

					}
					dash.VerifyFatal(snapshotstatuserror, true, "Volume snapshot is in error state?")
				}
			}
		})
		Step("Create cloud credentials", func() {
			log.InfoD(stepLog)
			err := CreatePXCloudCredential()
			log.FailOnError(err, "failed to create cloud credential")
			log.Info("Cloud credentials created successfully.")
			n := node.GetStorageDriverNodes()[0]
			uuidCmd := "pxctl cred list -j | grep uuid"
			output, err := runCmd(uuidCmd, n)
			log.FailOnError(err, "error getting uuid for cloudsnap credential.")
			if output == "" {
				log.FailOnError(fmt.Errorf("cloud cred is not created"), "Check for cloud cred exists?")
			}

			credUUID := strings.Split(strings.TrimSpace(output), " ")[1]
			credUUID = strings.ReplaceAll(credUUID, "\"", "")
			log.Infof("Got Cred UUID: %s", credUUID)

		})
		stepLog = "Verify that the cloud snapshot status is ready after creating the cloud credentials"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			log.Info("Waiting for the next scheduled snapshot to be triggered.")
			time.Sleep(5 * time.Minute)
			err := ValidateSnapshot(contexts)
			log.FailOnError(err, "Error during snapshot validation")
			log.InfoD("Snapshot validation completed successfully")

		})

	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		opts := make(map[string]bool)
		DestroyApps(contexts, opts)
		AfterEachTest(contexts)
	})
})

var _ = Describe("{GroupCloudSnapshot}", Label("staging", "p0", "positive", "px_ops"), func() {
	/*
	   Step1: Deploy app
	   step2: Label pvc reading the yaml file
	   Step3: create cloud credentials
	   step4: Create groupcloudsnapshot
	   step5: verify snapshot is done
	*/
	var testrailID = 0
	JustBeforeEach(func() {
		StartTorpedoTest("GroupCloudSnapshot", "Initiating Group CloudSnaphot for the volume. Verifying that snapshots are completed successfully.", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})
	stepLog := "Initiating Group CloudSnaphot for the volume. Verifying that snapshots are completed successfully."
	It(stepLog, func() {
		log.InfoD(stepLog)
		contexts = make([]*scheduler.Context, 0)
		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("groupsnapshot-%d", i))...)
		}
		ValidateApplications(contexts)
		defer DestroyApps(contexts, nil)
		stepLog = "Create cloud credentails"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			err := CreatePXCloudCredential()
			log.FailOnError(err, "failed to create cloud credential")
			n := node.GetStorageDriverNodes()[0]
			uuidCmd := "pxctl cred list -j | grep uuid"
			output, err := runCmd(uuidCmd, n)
			log.FailOnError(err, "error getting uuid for cloudsnap credential")
			if output == "" {
				log.FailOnError(fmt.Errorf("cloud cred is not created"), "Check for cloud cred exists?")
			}

		})
		for _, ctx := range contexts {
			var appVolumes []*volume.Volume
			log.InfoD(fmt.Sprintf("get volumes for %s app", ctx.App.Key))
			appVolumes, err = Inst().S.GetVolumes(ctx)
			log.FailOnError(err, "failed to get volumes for app [%s]", ctx.App.Key)
			log.Infof("List of app [%s] volumes [%v]", ctx.App.Key, appVolumes)
			dash.VerifyFatal(len(appVolumes) > 0, true, "App volumes exist?")
			groupSnapshot := &storkv1.GroupVolumeSnapshot{
				ObjectMeta: v1.ObjectMeta{
					Name:      ctx.App.Key + "-groupsnapshot",
					Namespace: ctx.App.NameSpace,
				},
				Spec: storkv1.GroupVolumeSnapshotSpec{
					PVCSelector: storkv1.PVCSelectorSpec{
						LabelSelector: v1.LabelSelector{
							MatchLabels: map[string]string{
								"app": ctx.App.Key,
							},
						},
					},
					RestoreNamespaces: []string{ctx.App.NameSpace},
					MaxRetries:        0,
					Options: map[string]string{
						"portworx/snapshot-type": "cloud",
					},
				},
			}
			stepLog := "Retrieving PVCs and adding labels based on the specified selector"
			Step(stepLog, func() {
				log.InfoD(stepLog)
				matchLabels := groupSnapshot.Spec.PVCSelector.LabelSelector.MatchLabels
				log.Infof("Match Labels: %v", matchLabels)
				for labelKey, labelValue := range matchLabels {
					log.Infof("Labelkey %s: LabelValue %s\n", labelKey, labelValue)
					pvcList, err := core.Instance().GetPersistentVolumeClaims(ctx.App.NameSpace, nil)
					log.FailOnError(err, "Failed to get PVCs from context")
					for _, pvc := range pvcList.Items {
						pvcPointer, err := core.Instance().GetPersistentVolumeClaim(pvc.Name, ctx.App.NameSpace)
						log.FailOnError(err, "Unable to get PVC for the namespace %v", ctx.App.NameSpace)
						err = AddLabelToResource(pvcPointer, labelKey, labelValue)
						log.FailOnError(err, fmt.Sprintf("Failed to add label %s: %s to PVC %s", labelKey, labelValue, pvc.Name))
						log.Infof("Label %s: %s added to PVC %s", labelKey, labelValue, pvc.Name)

					}
				}

			})
			stepLog = "Creating group volume snapshot"
			Step(stepLog, func() {
				log.InfoD(stepLog)
				_, err := storkops.Instance().CreateGroupSnapshot(groupSnapshot)
				log.FailOnError(err, "Unable to create group snaphot for cloud")
				msg := fmt.Sprintf("GroupVolumeSnapshot %v created successfully", groupSnapshot.Name)
				log.InfoD(msg)

			})
			log.Info("Waiting for 3 minutes for snapshot status to be updated...")
			time.Sleep(3 * time.Minute)
			stepLog = "Verify the cloud snapshot status"
			Step(stepLog, func() {
				log.InfoD(stepLog)
				var snapshotScheduleName string
				listgroupSnapshots, err := storkops.Instance().ListGroupSnapshots(ctx.App.NameSpace)
				log.FailOnError(err, "error getting snapshot list")
				log.Infof("groupsnapshot values %v", listgroupSnapshots)
				for _, snap := range listgroupSnapshots.Items {
					snapshotScheduleName = snap.Name
				}
				groupSnapshots, err := storkops.Instance().GetGroupSnapshot(snapshotScheduleName, ctx.App.NameSpace)
				log.FailOnError(err, "Error getting group snapshot for [%s] in namespace [%s]", snapshotScheduleName, ctx.App.NameSpace)
				log.Infof("Snapshot schedule response for [%s]: %+v", snapshotScheduleName, groupSnapshots)
				log.Infof("Checking snapshots for volumes: %v", appVolumes)
				for _, vol := range appVolumes {
					foundSnapshotandcheckthecondition := false
					apiVol, err := Inst().V.InspectVolume(vol.ID)
					log.FailOnError(err, "unable to inspect volume ")
					log.Infof("volume ids:%v", apiVol.Id)
					log.Infof("Checking for snapshot related to volume %s", vol.Name)
					log.Infof("Checking group snapshot: %s", groupSnapshots.Name)
					for _, volumeSnapshot := range groupSnapshots.Status.VolumeSnapshots {
						log.Infof("Checking volume snapshot: %s", volumeSnapshot.VolumeSnapshotName)
						parentVolumeID := volumeSnapshot.ParentVolumeID
						log.Infof("parent volume id:%v", parentVolumeID)
						log.Infof("volume id for app volume %v", apiVol.Id)
						if parentVolumeID == apiVol.Id {
							log.Infof("Found snapshot for volume %s in group snapshot %s", vol.Name, groupSnapshots.Name)
							for _, condition := range volumeSnapshot.Conditions {
								if condition.Type == "Ready" && condition.Status == "True" {
									foundSnapshotandcheckthecondition = true
									log.Infof("Volume groupsnapshot %s for volume %s is ready.", volumeSnapshot.VolumeSnapshotName, vol.Name)
								} else {
									log.Infof("Volume groupsnapshot %s for volume %s is not ready. Status: %s", volumeSnapshot.VolumeSnapshotName, vol.Name, condition.Status)
									err := osutils.Kubectl([]string{"-n", ctx.App.NameSpace, "describe", "volumesnapshot.volumesnapshot.external-storage.k8s.io", snapshotScheduleName})
									log.FailOnError(err, "Unable to get describe command to get groupsnapshot details")
									foundSnapshotandcheckthecondition = false
									break
								}
							}
							if !foundSnapshotandcheckthecondition {
								break
							}
						}
					}
					dash.VerifyFatal(foundSnapshotandcheckthecondition, true, "Snapshot is ready for that volume?")
				}

			})
			stepLog = "Deleting group volume snapshot"
			Step(stepLog, func() {
				log.InfoD(stepLog)
				err := storkops.Instance().DeleteGroupSnapshot(groupSnapshot.Name, ctx.App.NameSpace)
				log.FailOnError(err, "Unable to delete groupsnapshot: %v", groupSnapshot.Name)
				log.InfoD("GroupVolumeSnapshot %v is deleted successfully", groupSnapshot.Name)

			})
			stepLog = "Initiating cleanup of Persistent Volume Claims (PVCs)"
			Step(stepLog, func() {
				log.InfoD(stepLog)
				TearDownContext(ctx, nil)
				pvcList, err := core.Instance().GetPersistentVolumeClaims(ctx.App.NameSpace, nil)
				log.FailOnError(err, "Failed to get PVCs from context")
				for _, pvc := range pvcList.Items {
					log.Infof("Successfully initiated PVC deletion  '%s'", pvc.Name)
					err := core.Instance().DeletePersistentVolumeClaim(pvc.Name, ctx.App.NameSpace)
					if err != nil {
						log.Infof("Unable to delete PVC %s: %v", pvc.Name, err)
					} else {
						log.Infof("PVC %s deleted successfully", pvc.Name)
					}

				}

			})

		}

	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})
})

// Empty trashcan all volumes should get deleted immediately.
var _ = Describe("{EmptyTrashcanBeforeVEM}", Label("p0", "positive", "px_vol_ops", "trashcan", "staging"), func() {
	/*
		Step1: Take storage node from cluster
		Step3: Scheduled Applications on the node
		Step4: Validated those Applications
		step5: Once volume are created on the node, enable Trashcan.
		Step6: we need wait for 10 min update to Trashcan.
		Step7: Once you have trashcan enabled , delete a volume which was created earlier on the new Node
		make sure the volume deleted should go to trashcan.
		Step8: Delete the volumes under trashcan immediately
	*/
	var testrailID = 0
	JustBeforeEach(func() {
		StartTorpedoTest("EmptyTrashcanBeforeVEM", "delete Volumes in trashcan before volume expiry minutes", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})
	stepLog := "Delete Volumes on Trashcan before Volume Expiration Minutes"
	It(stepLog, func() {
		log.InfoD(stepLog)
		var (
			contexts                 = make([]*scheduler.Context, 0)
			originalTrashcanVols     []string
			originalTrashcanVolsLeng int
		)

		log.InfoD("Pick the one Storage node in cluster")
		pickNode := node.GetStorageNodes()[0]
		stepLog = "Create volumes on cluster with applications"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			stepLog = "Scheduling the Applications"
			Step(stepLog, func() {
				log.InfoD(stepLog)
				for i := 0; i < Inst().GlobalScaleFactor; i++ {
					contexts = append(contexts, ScheduleApplications(fmt.Sprintf("storagenodecreatevolume-%d", i))...)
				}
				log.InfoD("Applications are scheduled successfully")
			})

			stepLog = "Validate the applications"
			Step(stepLog, func() {
				log.InfoD(stepLog)
				ValidateApplications(contexts)
				log.InfoD("scheduled Applications are validated")
			})

			stepLog = "Enable Trashcan"
			Step(stepLog, func() {
				log.InfoD(stepLog)
				err := Inst().V.SetClusterOptsWithConfirmation(pickNode, map[string]string{
					"--volume-expiration-minutes": "10",
				})
				log.FailOnError(err, "error while enabling trashcan")
				log.InfoD("Trashcan is successfully enabled")
			})

			stepLog = "fetching all the originally existing volumes from the Trashcan"
			Step(stepLog, func() {
				log.InfoD(stepLog)
				trashcanVols, err := Inst().V.GetTrashCanVolumeIds(pickNode)
				log.FailOnError(err, "error while getting volumes from the trashcan")
				for _, vol := range trashcanVols {
					if vol = strings.ReplaceAll(vol, " ", ""); vol != "" {
						originalTrashcanVols = append(originalTrashcanVols, vol)
					}
				}
				originalTrashcanVolsLeng = len(originalTrashcanVols)
				log.InfoD("Volumes that are already exists in the trashcan [%v]", originalTrashcanVols)
			})

			stepLog := "Delete volumes on the node"
			Step(stepLog, func() {
				log.InfoD("Deleting the Volumes with destroy Apps")
				appsValidateAndDestroy(contexts)
			})
			trashcanVols := make([]string, 0)
			stepLog = "validate volumes in trashcan"
			Step(stepLog, func() {
				log.InfoD(stepLog)
				t := func() (interface{}, bool, error) {
					trashcanVolsNew, err := Inst().V.GetTrashCanVolumeIds(pickNode)
					if err != nil {
						return "", false, err
					}
					if len(trashcanVolsNew) <= originalTrashcanVolsLeng {
						return "", false, err
					}
					return trashcanVolsNew, true, nil
				}
				trashcanVolsNew, err := task.DoRetryWithTimeout(t, 5*time.Minute, 5*time.Second)
				log.FailOnError(err, "error While getting trashcan volumes")
				for _, vol := range trashcanVolsNew.([]string) {
					if vol = strings.ReplaceAll(vol, " ", ""); vol != "" {
						trashcanVols = append(trashcanVols, vol)
					}
				}
				log.InfoD("Listed the trashcan volumes in the node:%s", trashcanVols)
				log.Infof("trashcan len: %d", len(trashcanVols))
				dash.VerifyFatal(len(trashcanVols) > 0, true, "validate volumes exist in trashcan")
			})

			stepLog = "Delete volumes in trashcan"
			Step(stepLog, func() {
				log.InfoD(stepLog)
				trashcanVolsNew, err := Inst().V.GetTrashCanVolumeIds(pickNode)
				for _, vol := range trashcanVolsNew {
					log.Infof("Trashcan volume id: %s", vol)

					if !IsVolumeExits(vol) {
						if vol = strings.ReplaceAll(vol, " ", ""); vol == "" {
							continue
						} else if slices.Contains(originalTrashcanVols, vol) {
							log.Infof("Skipping the volume id: %s since it's already exists", vol)
							continue
						}
					}
					log.InfoD(fmt.Sprintf("detach volume [%v]", vol))
					err = Inst().V.DetachVolume(vol)
					log.FailOnError(err, fmt.Sprintf("Failed to detach volume [%v]", vol))
					time.Sleep(500 * time.Millisecond)
					log.InfoD(fmt.Sprintf("detele volume [%v]", vol))
					err = Inst().V.DeleteVolume(vol)
					log.FailOnError(err, fmt.Sprintf("Delete volume with ID [%v] failed", vol))
				}
			})

			trashcanVolumes := make([]string, 0)
			stepLog = "validate volumes are permanently delete in trashcan"
			Step(stepLog, func() {
				log.InfoD(stepLog)
				log.InfoD("Get the volumes in trashcan")
				trashcanVolumesNew, err := Inst().V.GetTrashCanVolumeIds(pickNode)
				log.FailOnError(err, "error while getting trashcan volumes after expiry time")
				for _, vol := range trashcanVolumesNew {
					if vol = strings.ReplaceAll(vol, " ", ""); vol != "" {
						trashcanVolumes = append(trashcanVolumes, vol)
					}
				}
				log.Infof("Trashcan Volume len: %d", len(trashcanVolumes))
				if len(trashcanVolumes) == originalTrashcanVolsLeng {
					log.Infof("Volumes are permanently deleted from trashcan")
				} else {
					err := fmt.Errorf("Volumes are not deleted in trashcan")
					log.FailOnError(err, "Volumes are still in trashcan")
				}
			})

			stepLog = "Disable Trashcan"
			Step(stepLog, func() {
				log.InfoD(stepLog)
				err := Inst().V.SetClusterOptsWithConfirmation(pickNode, map[string]string{
					"--volume-expiration-minutes": "0",
				})
				log.FailOnError(err, "error while enabling trashcan")
				log.InfoD("Trashcan is successfully Disabled")
			})
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})
})
