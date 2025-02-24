package tests

import (
	"fmt"
	"github.com/pure-px/torpedo/drivers/volume"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/libopenstorage/openstorage/api"
	"github.com/portworx/sched-ops/k8s/talisman"
	"github.com/pure-px/sched-ops/k8s/core"
	"github.com/pure-px/sched-ops/k8s/storage"
	"github.com/pure-px/torpedo/pkg/log"
	"github.com/pure-px/torpedo/pkg/units"
	"github.com/pure-px/torpedo/pkg/vpsutil"
	corev1 "k8s.io/api/core/v1"
	storageApi "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pure-px/torpedo/pkg/testrailuttils"

	. "github.com/onsi/ginkgo/v2"

	"github.com/pure-px/torpedo/drivers/node"
	"github.com/pure-px/torpedo/drivers/scheduler"
	"github.com/pure-px/torpedo/drivers/scheduler/k8s"
	. "github.com/pure-px/torpedo/tests"
)

const (
	dfDefaultTimeout       = 1 * time.Minute
	driveFailTimeout       = 2 * time.Minute
	dfDefaultRetryInterval = 5 * time.Second
)

var _ = Describe("{DriveFailure}", Label("p1", "negative", "error_injection", "px_vol_ops", "drive_failure"), func() {
	var testrailID = 35265
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/cases/view/35265
	var runID int
	JustBeforeEach(func() {
		StartTorpedoTest("DriveFailure", "Validate PX after drive failure", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})
	var contexts []*scheduler.Context

	testName := "drivefailure"
	stepLog := "has to schedule apps and induce a drive failure on one of the nodes"
	It(stepLog, func() {
		log.InfoD(stepLog)
		var err error
		contexts = make([]*scheduler.Context, 0)

		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("%s-%d", testName, i))...)
		}

		ValidateApplications(contexts)

		stepLog = "get nodes for all apps in test and induce drive failure on one of the nodes"

		Step(stepLog, func() {
			log.InfoD(stepLog)
			for _, ctx := range contexts {
				var (
					drives        []string
					appNodes      []node.Node
					nodeWithDrive node.Node
				)

				stepLog = fmt.Sprintf("get nodes where %s app is running", ctx.App.Key)
				Step(stepLog, func() {
					log.InfoD(stepLog)
					appNodes, err = Inst().S.GetNodesForApp(ctx)
					log.FailOnError(err, "Failed to get nodes for app %s", ctx.App.Key)
					dash.VerifyFatal(len(appNodes) > 0, true, fmt.Sprintf("Found %d apps", len(appNodes)))

					nodeWithDrive = appNodes[0]
				})

				stepLog = fmt.Sprintf("get drive from node %v", nodeWithDrive)
				Step(stepLog, func() {
					log.InfoD(stepLog)
					drives, err = Inst().V.GetStorageDevices(nodeWithDrive)
					log.FailOnError(err, fmt.Sprintf("Failed to get storage devices for the node %s", nodeWithDrive.Name))
					dash.VerifyFatal(len(drives) > 0, true, fmt.Sprintf("Found drives length %d", len(appNodes)))
				})

				busInfoMap := make(map[string]string)
				stepLog := fmt.Sprintf("induce a failure on all drives on the node %v", nodeWithDrive)
				Step(stepLog, func() {
					log.InfoD(stepLog)
					for _, driveToFail := range drives {
						busID, err := Inst().N.YankDrive(nodeWithDrive, driveToFail, node.ConnectionOpts{
							Timeout:         dfDefaultTimeout,
							TimeBeforeRetry: dfDefaultRetryInterval,
						})
						busInfoMap[driveToFail] = busID
						log.FailOnError(err, "Failed to yank drive %s", driveToFail)

					}
					stepLog = "wait for the drives to fail"
					Step(stepLog, func() {
						log.InfoD(stepLog)
						time.Sleep(30 * time.Second)
					})

					Step(fmt.Sprintf("check if apps are running"), func() {
						ValidateContext(ctx)
					})

				})

				stepLog = "recover all drives and the storage driver"
				Step(stepLog, func() {
					log.InfoD(stepLog)
					for _, driveToFail := range drives {
						err = Inst().N.RecoverDrive(nodeWithDrive, driveToFail, busInfoMap[driveToFail], node.ConnectionOpts{
							Timeout:         driveFailTimeout,
							TimeBeforeRetry: dfDefaultRetryInterval,
						})
						dash.VerifyFatal(err, nil, fmt.Sprintf("Verify drive %s recovery init", driveToFail))

					}
					stepLog = "wait for the drives to recover"
					Step(stepLog, func() {
						log.InfoD(stepLog)
						time.Sleep(30 * time.Second)
					})

					err = Inst().V.RecoverDriver(nodeWithDrive)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Verify drive recovery in node %s", nodeWithDrive.Name))

				})

				stepLog = "check if volume driver is up"
				Step(stepLog, func() {
					err = Inst().V.WaitDriverUpOnNode(nodeWithDrive, Inst().DriverStartTimeout)
					dash.VerifyFatal(err, nil, "Validate volume driver is up")
				})
			}
		})

		ValidateAndDestroy(contexts, nil)
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})
})

var _ = Describe("{YankJournalWithPxRestart}", Label("p1", "hal_ops_disruption", "px_restart", "drive_failure", "YankJournalDrive", "functional"), func() {
	testName = "yank-journal-px-restart"
	testDescription = "Yank journal drive and restart PX"
	YankJournalTest(testName, testDescription)
})

var _ = Describe("{YankJournalWithNodeReboot}", Label("p1", "hal_ops_disruption", "node_reboot", "drive_failure", "YankJournalDrive", "functional"), func() {
	testName = "yank-journal-node-reboot"
	testDescription = "Yank journal drive and reboot node"
	YankJournalTest(testName, testDescription)
})

var _ = Describe("{YankJournalWithNodeMaintenanceCycle}", Label("p1", "hal_ops_disruption", "NodeMaintenance", "drive_failure", "YankJournalDrive", "functional"), func() {
	testName := "yank-journal-node-maintenance-cycle"
	testDescription = "Yank journal drive and do node maintenance"
	YankJournalTest(testName, testDescription)
})

var _ = Describe("{YankJournalWithIOsRunning}", Label("p1", "hal_ops_disruption", "px_restart", "drive_failure", "YankJournalDrive", "functional"), func() {
	testName := "yank-journal-io"
	testDescription = "Yank journal drive with IOs running"
	YankJournalTest(testName, testDescription)
})

func YankJournalTest(testName, testDesc string) {
	var (
		nodeSelected node.Node
		busID        string
		contexts     = make([]*scheduler.Context, 0)
	)

	JustBeforeEach(func() {
		StartTorpedoTest(testName, testDesc, nil, 0)
	})

	itLog := testDesc
	It(itLog, func() {
		log.InfoD(itLog)

		stepLog := "Schedule apps to perform IOs"
		Step(stepLog, func() {
			log.InfoD(stepLog)

			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				contexts = append(contexts, ScheduleApplications(fmt.Sprintf("yankjournal-%d", i))...)
			}
			ValidateApplications(contexts)
		})
		defer appsValidateAndDestroy(contexts)

		ctx := contexts[0]
		volumes, err := Inst().S.GetVolumes(ctx)
		log.FailOnError(err, "Failed while listing the volume with error")
		log.InfoD("Vol deatils %v", volumes)

		if len(volumes) == 0 {
			msg := fmt.Sprintf("There are no volumes associated with the app %v", ctx.App.Key)
			log.InfoD(msg)
			Skip(msg)
		}
		volumeSelected := volumes[0]

		stepLog = "Select the volume replica node"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			rsDetails, err := Inst().V.GetReplicaSets(volumeSelected)
			log.FailOnError(err, fmt.Sprintf("error getting replica sets for vol %s", volumeSelected.Name))
			log.InfoD("Volume Replica info %v", rsDetails)
			volReplicaNodeID := rsDetails[0].GetNodes()[0]

			storageNodes := node.GetStorageNodes()
			for _, nodeDetail := range storageNodes {
				if nodeDetail.Id == volReplicaNodeID {
					nodeSelected = nodeDetail
				}
			}
		})

		//add journal drive if it doesn't exists
		jDev, err := Inst().V.GetJournalDevicePath(&nodeSelected)
		log.FailOnError(err, fmt.Sprintf("error getting journal device path from node %s", nodeSelected.Name))

		if jDev == "" {
			stepLog = "enter pool maintenance mode"
			Step(stepLog, func() {
				log.InfoD(stepLog)
				err = EnterPoolMaintenance(nodeSelected)
				log.FailOnError(err, "Failed to enter maintenance mode")
				log.Info("enter pool maintenance mode succeed")
			})

			stepLog = "Add journal drive"
			Step(stepLog, func() {
				log.InfoD(stepLog)
				driveSpecs, err := GetCloudDriveDeviceSpecs()
				log.FailOnError(err, "Error getting cloud drive specs")
				deviceSpec := driveSpecs[0]
				devicespecjournal := deviceSpec + " --journal"
				systemOpts := node.SystemctlOpts{
					ConnectionOpts: node.ConnectionOpts{
						Timeout:         2 * time.Minute,
						TimeBeforeRetry: defaultRetryInterval,
					},
					Action: "start",
				}
				drivesMap, err := Inst().N.GetBlockDrives(nodeSelected, systemOpts)
				log.FailOnError(err, "error getting block drives from node %s", nodeSelected.Name)
				blockDeviceBefore := len(drivesMap)
				err = Inst().V.AddCloudDrive(&nodeSelected, devicespecjournal, -1)
				log.FailOnError(err, "journal add failed")
				drivesMapAfter, err := Inst().N.GetBlockDrives(nodeSelected, systemOpts)
				log.FailOnError(err, "error getting block drives from node %s", nodeSelected.Name)
				blockDeviceAfter := len(drivesMapAfter)
				dash.VerifyFatal(blockDeviceAfter > blockDeviceBefore, true, "adding cloud drive as journal successful")
			})

			stepLog = "Exit pool maintenance mode"
			Step(stepLog, func() {
				log.InfoD(stepLog)
				err = ExitPoolMaintenance(*&nodeSelected)
				log.FailOnError(err, "Failed to exit maintenance mode")
				log.Info("exit pool maintenance mode succeed")
			})

			//Get the newly added Journal device path
			jDev, err = Inst().V.GetJournalDevicePath(&nodeSelected)
			log.FailOnError(err, fmt.Sprintf("error getting journal device path from node %s", nodeSelected.Name))
			log.InfoD("Journal device path - %s", jDev)
		}

		cmd := fmt.Sprintf("lsblk -no pkname %s", jDev)
		parentDevPath, err := Inst().N.RunCommandWithNoRetry(nodeSelected, cmd, node.ConnectionOpts{
			Timeout:         2 * time.Minute,
			TimeBeforeRetry: 10 * time.Second,
		})
		log.FailOnError(err, "error occured running the command to identify the parent device path of the journal partition %s", jDev)
		log.InfoD("Parent device path of the journal device is %s", parentDevPath)
		parentDevPath = strings.TrimRight(parentDevPath, "\n")

		//random delay in secs
		sleepTime := rand.Intn(60-1) + 1
		time.Sleep(time.Second * (time.Duration(sleepTime)))

		stepLog = "Yank journal drive"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			busID, err = Inst().N.YankDrive(nodeSelected, parentDevPath, node.ConnectionOpts{
				Timeout:         dfDefaultTimeout,
				TimeBeforeRetry: dfDefaultRetryInterval,
			})
			log.FailOnError(err, fmt.Sprintf("failed to yank journal drive on node [%s]", nodeSelected.Name))
			log.InfoD("Bus id - %s", busID)
		})

		if testName == "yank-journal-node-reboot" {
			stepLog = "Reboot the node"
			Step(stepLog, func() {
				log.Info(stepLog)
				err = RebootNodeAndWaitForPxUp(nodeSelected)
				log.FailOnError(err, "Failed to reboot node and wait till it is up")
			})
		} else if testName == "yank-journal-px-restart" {
			stepLog = "Restart portworx and wait for it to come up"
			Step(stepLog, func() {
				log.Info(stepLog)
				Step(fmt.Sprintf("node with Px restart is: %s", nodeSelected.Name), func() {
					err := Inst().V.RestartDriver(nodeSelected, nil)
					log.FailOnError(err, fmt.Sprintf("Error occured while Restart PX on node:%v", nodeSelected.Name))
				})

				Step(fmt.Sprintf("wait for volume driver to restart on node: %v", nodeSelected.Name), func() {
					err := Inst().V.WaitForPxPodsToBeUp(nodeSelected)
					log.FailOnError(err, fmt.Sprintf("Error occured while Validating PX restart is done on node:%v", nodeSelected.Name))
				})
			})
		} else if testName == "yank-journal-node-maintenance-cycle" {
			stepLog = "Enter maintenance mode"
			Step(stepLog, func() {
				log.Info(stepLog)
				err = Inst().V.EnterMaintenance(nodeSelected)
				log.FailOnError(err, fmt.Sprintf("fail to enter node %s in maintenance mode", nodeSelected.Name))
				status, err := Inst().V.GetNodeStatus(nodeSelected)
				log.FailOnError(err, fmt.Sprintf("Error getting PX status of node %s", nodeSelected.Name))
				dash.VerifyFatal(*status, api.Status_STATUS_MAINTENANCE, fmt.Sprintf("Node %s Status not Online", nodeSelected.Name))
			})
		} else if testName == "yank-journal-io" {
			time.Sleep(2 * time.Minute)
		}

		stepLog = "Recover yank drive"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			err = Inst().N.RecoverDrive(nodeSelected, parentDevPath, busID, node.ConnectionOpts{
				Timeout:         driveFailTimeout,
				TimeBeforeRetry: dfDefaultRetryInterval,
			})
			log.FailOnError(err, fmt.Sprintf("failed to recover yank journal drive on node [%s]", nodeSelected.Name))
			log.InfoD("Verified recover yank drive")
		})

		if testName == "yank-journal-node-maintenance-cycle" {
			stepLog = "Exit maintenance mode"
			Step(stepLog, func() {
				err = Inst().V.ExitMaintenance(nodeSelected)
				log.FailOnError(err, fmt.Sprintf("fail to exit node %s in maintenance mode", nodeSelected.Name))
				status, err := Inst().V.GetNodeStatus(nodeSelected)
				log.FailOnError(err, fmt.Sprintf("Error getting PX status of node %s", nodeSelected.Name))
				dash.VerifyFatal(*status, api.Status_STATUS_OK, fmt.Sprintf("Node %s Status not Online", nodeSelected.Name))
			})
		}

		stepLog = "Verify Px Status"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			status, err := Inst().V.GetPxctlStatus(nodeSelected)
			log.FailOnError(err, fmt.Sprintf("failed to get pxctl status on node [%s]", nodeSelected.Name))
			dash.VerifyFatal(status == api.Status_STATUS_OK.String(), true, fmt.Sprintf("node [%s] status is up but PX cluster is not ok. Expected: %v Actual: %v",
				nodeSelected.Name, api.Status_STATUS_OK, status))
			log.InfoD("px status %v", status)
		})

		stepLog = "Do pool maintenance"
		Step(stepLog, func() {
			log.Info(stepLog)
			log.InfoD(fmt.Sprintf("Performing pool maintenance cycle on node %s", nodeSelected.Name))
			err = Inst().V.RecoverPool(nodeSelected)
			log.FailOnError(err, fmt.Sprintf("error performing pool maintenance cycle on node %s", nodeSelected.Name))
		})

		stepLog = "Verify pool status"
		Step(stepLog, func() {
			log.Info(stepLog)
			poolsStatus, err := Inst().V.GetNodePoolsStatus(nodeSelected)
			log.FailOnError(err, "error getting pool status on node %s", nodeSelected.Name)
			for poolID, status := range poolsStatus {
				dash.VerifyFatal(status, "Online", fmt.Sprintf("Pool %s Status not Online", poolID))
			}
		})

	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts)
	})
}

var _ = Describe("{YankPoolDriveWithPxRestart}", Label("p1", "hal_ops_disruption", "px_restart", "drive_failure", "YankPoolDrive", "functional"), func() {
	testName = "yank-pool-drive-px-restart"
	testDescription = "Yank pool drive and restart px"
	YankPoolDriveTest(testName, testDescription)
})

var _ = Describe("{YankPoolDriveWithNodeReboot}", Label("p1", "hal_ops_disruption", "node_reboot", "drive_failure", "YankPoolDrive", "functional"), func() {
	testName = "yank-pool-drive-node-reboot"
	testDescription = "Yank pool drive and reboot node"
	YankPoolDriveTest(testName, testDescription)
})

var _ = Describe("{YankPoolDriveWithNodeMaintenanceCycle}", Label("p1", "hal_ops_disruption", "NodeMaintenance", "drive_failure", "YankPoolDrive", "functional"), func() {
	testName = "yank-pool-drive-node-maintenance-cycle"
	testDescription = "Yank pool drive with node maintenance cycle"
	YankPoolDriveTest(testName, testDescription)
})

var _ = Describe("{YankPoolDriveWithIOs}", Label("p1", "hal_ops_disruption", "drive_failure", "YankPoolDrive", "functional"), func() {
	testName = "yank-pool-drive-IO"
	testDescription = "Yank pool drive with IOs running"
	YankPoolDriveTest(testName, testDescription)
})

var _ = Describe("{YankMetadataWithPxRestart}", Label("p1", "hal_ops_disruption", "px_restart", "drive_failure", "YankMetadataDrive", "functional"), func() {
	testName = "YankMetadataWithPxRestart"
	testDescription = "Yank metadata drive with restart PX"
	YankMetadataTest(testName, testDescription)
})
var _ = Describe("{YankMetadataWithNodeReboot}", Label("p1", "hal_ops_disruption", "node_reboot", "drive_failure", "YankMetadataDrive", "functional"), func() {
	testName = "YankMetadataWithNodeReboot"
	testDescription = "Yank metadata drive and reboot node"
	YankMetadataTest(testName, testDescription)
})
var _ = Describe("{YankMetadataWithNodeMaintenanceCycle}", Label("p1", "hal_ops_disruption", "NodeMaintenance", "drive_failure", "YankMetadataDrive", "functional"), func() {
	testName = "YankMetadataWithNodeMaintenanceCycle"
	testDescription = "Yank metadata drive and node maintenance cycle"
	YankMetadataTest(testName, testDescription)
})

var _ = Describe("{YankMetadataWithIOs}", func() {
	testName = "YankMetadataWithIOs"
	testDescription = "Yank metadata drive with IOs"
	YankMetadataTest(testName, testDescription)
})

func YankMetadataTest(testName, testDesc string) {
	var (
		nodeSelected node.Node
		busID, path  string
		kvdbNodesIDs []string
		contexts     = make([]*scheduler.Context, 0)
	)
	JustBeforeEach(func() {
		StartTorpedoTest(testName, testDesc, nil, 0)
	})

	itLog := testDesc
	It(itLog, func() {
		log.InfoD(itLog)

		stepLog := "Schedule apps to perform IOs"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				contexts = append(contexts, ScheduleApplications(fmt.Sprintf("yankmetadata-%d", i))...)
			}
			ValidateApplications(contexts)
		})
		defer appsValidateAndDestroy(contexts)

		ctx := contexts[0]
		volumes, err := Inst().S.GetVolumes(ctx)
		log.FailOnError(err, "Failed while listing the volume with error")
		log.InfoD("Vol deatils %v", volumes)

		if len(volumes) == 0 {
			msg := fmt.Sprintf("There are no volumes associated with the app %v", ctx.App.Key)
			log.InfoD(msg)
			Skip(msg)
		}
		volumeSelected := volumes[0]

		storageNodes := node.GetStorageNodes()
		index := rand.Intn(len(storageNodes))
		tNode := storageNodes[index]
		stepLog = "Get KVDB nodes"

		Step(stepLog, func() {
			log.InfoD(stepLog)
			kvdbMembers, err := Inst().V.GetKvdbMembers(tNode)
			log.FailOnError(err, "Error getting KVDB members")
			log.InfoD("kvdb members %+v", kvdbMembers)
			for _, n := range kvdbMembers {
				kvdbNodesIDs = append(kvdbNodesIDs, n.Name)
			}
		})
		stepLog = "Select the volume replica node"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			rsDetails, err := Inst().V.GetReplicaSets(volumeSelected)
			log.FailOnError(err, fmt.Sprintf("error getting replica sets for vol %s", volumeSelected.Name))
			log.InfoD("Volume Replica info %v", rsDetails)
			volReplicaNodeIDMap := map[string]bool{}

			for _, nodeId := range rsDetails[0].GetNodes() {
				volReplicaNodeIDMap[nodeId] = true
			}

			storageNodes := node.GetStorageNodes()
			for _, nodeDetail := range storageNodes {
				_, ok := volReplicaNodeIDMap[nodeDetail.Id]
				if ok && !Contains(kvdbNodesIDs, nodeDetail.Id) {
					nodeSelected = nodeDetail
				}
			}
		})

		//add metadata drive if it doesn't exists

		stepLog = "Check node has a metadata disk if not add"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			isDedicatedMetadataDiskExist := false
			//Check which node has metadata disk if not add one

			path, err = getMetaDataDiskPath(nodeSelected)
			log.FailOnError(err, "Failed to get metadata disk")
			if path != "" {
				log.InfoD("Metadata disk path: %v", path)
				isDedicatedMetadataDiskExist = true
			}

			if !isDedicatedMetadataDiskExist {
				deviceSpec := fmt.Sprintf("size=100 --metadata")
				log.InfoD("Initiate add cloud drive and validate")
				// enter pool maintenance mode

				stepLog := "Enter maintenance mode"
				Step(stepLog, func() {
					log.InfoD(stepLog)
					err = Inst().V.EnterPoolMaintenance(nodeSelected)
					log.FailOnError(err, "node: %v failed to transition to pool maintenance mode", nodeSelected.Name)
					log.Info("enter pool maintenance mode succeed")
				})

				stepLog = "Add metadata disk"
				Step(stepLog, func() {
					log.InfoD(stepLog)
					err := Inst().V.AddCloudDrive(&nodeSelected, deviceSpec, -1)
					log.FailOnError(err, "Failed to add metadata device on node : %s", nodeSelected.Name)
					log.InfoD("metadata disk added successfully on node [%s]", nodeSelected.Hostname)
				})

				// exit pool maintenance
				stepLog = "Exit pool maintenance mode"
				Step(stepLog, func() {
					log.InfoD(stepLog)
					err = Inst().V.ExitPoolMaintenance(nodeSelected)
					log.FailOnError(err, "Node: %v Failed to exit out of maintenance mode", nodeSelected.Name)
					log.Info("exit pool maintenance mode succeed")
				})
				path, err = getMetaDataDiskPath(nodeSelected)
				log.FailOnError(err, "Failed to get metadata disk")
			} else {
				log.InfoD("Metadata disk already exist: [%s]", path)
			}
			//check if selecteNode is empty or not
			if nodeSelected.Name == "" {
				log.FailOnError(fmt.Errorf("No node found with metadata disk or metadata disks cannot be added to any nodes"), "No node found with metadata disk ")
			}
		})

		path = strings.Trim(path, "/dev/")

		time.Sleep(time.Minute * 5)
		stepLog = "Yank metadata drive"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			busID, err = Inst().N.YankDrive(nodeSelected, path, node.ConnectionOpts{
				Timeout:         dfDefaultTimeout,
				TimeBeforeRetry: dfDefaultRetryInterval,
			})
			log.FailOnError(err, fmt.Sprintf("failed to yank metadata drive on node [%s]", nodeSelected.Name))
			log.InfoD("Bus id - %s", busID)
		})

		if testName == "YankMetadataWithNodeReboot" {
			stepLog = "Reboot the node"
			Step(stepLog, func() {
				log.Info(stepLog)
				err = RebootNodeAndWaitForPxUp(nodeSelected)
				log.FailOnError(err, "Failed to reboot node and wait till it is up")
			})
		} else if testName == "YankMetadataWithPxRestart" {
			stepLog = "Restart portworx and wait for it to come up"
			Step(stepLog, func() {
				log.Info(stepLog)
				Step(fmt.Sprintf("node with Px restart is: %s", nodeSelected.Name), func() {
					err := Inst().V.RestartDriver(nodeSelected, nil)
					log.FailOnError(err, fmt.Sprintf("Error occured while Restart PX on node:%v", nodeSelected.Name))
				})

				Step(fmt.Sprintf("wait for volume driver to restart on node: %v", nodeSelected.Name), func() {
					err := Inst().V.WaitForPxPodsToBeUp(nodeSelected)
					log.FailOnError(err, fmt.Sprintf("Error occured while Validating PX restart is done on node:%v", nodeSelected.Name))
				})
			})
		} else if testName == "YankMetadataWithNodeMaintenanceCycle" {
			stepLog = "Enter maintenance mode"
			Step(stepLog, func() {
				log.Info(stepLog)
				err = Inst().V.EnterMaintenance(nodeSelected)
				log.FailOnError(err, fmt.Sprintf("fail to enter node %s in maintenance mode", nodeSelected.Name))
				status, err := Inst().V.GetNodeStatus(nodeSelected)
				log.FailOnError(err, fmt.Sprintf("Error getting PX status of node %s", nodeSelected.Name))
				dash.VerifyFatal(*status, api.Status_STATUS_MAINTENANCE, fmt.Sprintf("Node %s Status not Online", nodeSelected.Name))
			})
		} else if testName == "YankMetadataWithIOs" {
			time.Sleep(time.Minute * 2)
		}

		stepLog = "Recover yank drive"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			err = Inst().N.RecoverDrive(nodeSelected, path, busID, node.ConnectionOpts{
				Timeout:         driveFailTimeout,
				TimeBeforeRetry: dfDefaultRetryInterval,
			})
			log.FailOnError(err, fmt.Sprintf("failed to recover yank metadata drive on node [%s]", nodeSelected.Name))
			log.InfoD("Verified recover yank drive")
		})

		if testName == "YankMetadataWithNodeMaintenanceCycle" {
			stepLog = "Exit maintenance mode"
			Step(stepLog, func() {
				err = Inst().V.ExitMaintenance(nodeSelected)
				log.FailOnError(err, fmt.Sprintf("fail to exit node %s in maintenance mode", nodeSelected.Name))
				status, err := Inst().V.GetNodeStatus(nodeSelected)
				log.FailOnError(err, fmt.Sprintf("Error getting PX status of node %s", nodeSelected.Name))
				dash.VerifyFatal(*status, api.Status_STATUS_OK, fmt.Sprintf("Node %s Status not Online", nodeSelected.Name))
			})
		}

		stepLog = fmt.Sprintf("Wait for driver to up")
		Step(stepLog, func() {
			log.InfoD(stepLog)
			err = Inst().V.WaitDriverUpOnNode(nodeSelected, Inst().DriverStartTimeout)
			log.FailOnError(err, fmt.Sprintf("Driver is down on node %s", nodeSelected.Name))
			log.Info("Driver is up")

		})

		stepLog = "Verify Px Status"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			status, err := Inst().V.GetPxctlStatus(nodeSelected)
			log.FailOnError(err, fmt.Sprintf("failed to get pxctl status on node [%s]", nodeSelected.Name))
			dash.VerifyFatal(status == api.Status_STATUS_OK.String(), true, fmt.Sprintf("node [%s] status is up but PX cluster is not ok. Expected: %v Actual: %v",
				nodeSelected.Name, api.Status_STATUS_OK, status))
			log.InfoD("px status %v", status)
		})

		stepLog = "Do pool maintenance"
		Step(stepLog, func() {
			log.Info(stepLog)
			log.InfoD(fmt.Sprintf("Performing pool maintenance cycle on node %s", nodeSelected.Name))
			err = Inst().V.RecoverPool(nodeSelected)
			log.FailOnError(err, fmt.Sprintf("error performing pool maintenance cycle on node %s", nodeSelected.Name))
		})

		stepLog = "Verify pool status"
		Step(stepLog, func() {
			log.Info(stepLog)
			poolsStatus, err := Inst().V.GetNodePoolsStatus(nodeSelected)
			log.FailOnError(err, "error getting pool status on node %s", nodeSelected.Name)
			for poolID, status := range poolsStatus {
				dash.VerifyFatal(status, "Online", fmt.Sprintf("Pool %s Status not Online", poolID))
			}
		})

	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts)
	})
}

func ExpandPoolWithAddDrive(testName, description string) {
	var (
		targetSizeGiB    uint64
		poolUUIDSelected string
		poolToResize     *api.StoragePool
		contexts         []*scheduler.Context
		nodeSelected     *node.Node
		poolDrive        string
		busID            string
	)

	JustBeforeEach(func() {
		StartTorpedoTest(testName, description, nil, 0)
	})

	It(testName, func() {
		stepLog := "Schedule Apps"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			contexts = scheduleApps()
			time.Sleep(5 * time.Minute)
			log.Info("schedule app succeeded")
		})

		stepLog = "Select a pool to expand"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			poolUUIDSelected = pickPoolToResize(contexts, api.SdkStoragePool_RESIZE_TYPE_ADD_DISK, 128)
			dash.VerifyFatal(len(poolUUIDSelected) > 0, true, fmt.Sprintf("Expected poolIDToResize to not be empty, pool id to resize %s", poolUUIDSelected))
			poolToResize = getStoragePool(poolUUIDSelected)
			log.InfoD(fmt.Sprintf("Pool going to resize is UUID: [%s]", poolUUIDSelected))
			nodeSelected, err = GetNodeWithGivenPoolID(poolUUIDSelected)
			log.FailOnError(err, "Failed to get node from pool id [%v]", poolUUIDSelected)

		})

		if testName == "ExpandPoolAddDriveAfterYankingDrive" {
			stepLog = "Select pool drive"
			Step(stepLog, func() {
				log.InfoD(stepLog)

				poolIDSelected, err := GetPoolIDFromPoolUUID(poolUUIDSelected)
				log.FailOnError(err, fmt.Sprintf("error getting pool id for the pool %s", poolUUIDSelected))

				jDev, err := Inst().V.GetJournalDevicePath(nodeSelected)
				log.FailOnError(err, fmt.Sprintf("error getting journal device path from node %s", nodeSelected.Name))
				log.InfoD("Journal device path - %s", jDev)

				cmd := fmt.Sprintf("lsblk -no pkname %s", jDev)
				journalParentDevPath, err := Inst().N.RunCommandWithNoRetry(*nodeSelected, cmd, node.ConnectionOpts{
					Timeout:         2 * time.Minute,
					TimeBeforeRetry: 10 * time.Second,
				})
				log.FailOnError(err, "error occured running the command to identify the parent device path of the journal partition %s", jDev)
				journalParentDevPath = strings.TrimRight(journalParentDevPath, "\n")
				log.InfoD("Parent device path of the journal device is %s", journalParentDevPath)

				driveMap, err := Inst().V.GetPoolDrives(nodeSelected)
				log.FailOnError(err, fmt.Sprintf("error getting pool drive details for the node %s", nodeSelected))

				drives := driveMap[strconv.Itoa(int(poolIDSelected))]

				for _, drive := range drives {
					cmd := fmt.Sprintf("lsblk -no pkname %s", drive.Device)
					poolDriveParentPath, err := Inst().N.RunCommandWithNoRetry(*nodeSelected, cmd, node.ConnectionOpts{
						Timeout:         2 * time.Minute,
						TimeBeforeRetry: 10 * time.Second,
					})
					log.FailOnError(err, "error occured running the command to identify the parent device path of the drive %s", drive.Device)
					poolDriveParentPath = strings.TrimRight(poolDriveParentPath, "\n")
					log.InfoD("Parent device path of the pool device is %s", poolDriveParentPath)

					if (journalParentDevPath != "") && (poolDriveParentPath != "") && (journalParentDevPath == poolDriveParentPath) {
						continue
					}

					poolDrive = drive.Device
					log.InfoD("Pool drive selected - %s", poolDrive)
					break
				}
			})

			poolDrive = strings.Trim(poolDrive, "/")
			poolDriveArr := strings.Split(poolDrive, "/")
			poolDrive = poolDriveArr[len(poolDriveArr)-1]
			stepLog = "Yank drive"
			Step(stepLog, func() {
				log.InfoD(stepLog)
				busID, err = Inst().N.YankDrive(*nodeSelected, poolDrive, node.ConnectionOpts{
					Timeout:         dfDefaultTimeout,
					TimeBeforeRetry: dfDefaultRetryInterval,
				})
				log.FailOnError(err, fmt.Sprintf("failed to yank journal drive on node [%s]", nodeSelected.Name))
				log.InfoD("Bus id - %s", busID)
			})

			time.Sleep(time.Minute * 2)

			stepLog = "Recover yank drive"
			Step(stepLog, func() {
				log.InfoD(stepLog)
				err = Inst().N.RecoverDrive(*nodeSelected, poolDrive, busID, node.ConnectionOpts{
					Timeout:         driveFailTimeout,
					TimeBeforeRetry: dfDefaultRetryInterval,
				})
				log.FailOnError(err, fmt.Sprintf("failed to recover yank journal drive on node [%s]", nodeSelected.Name))
				log.InfoD("Verified recover yank drive")
			})

			stepLog = fmt.Sprintf("Wait for driver to up and check px status")
			Step(stepLog, func() {
				log.InfoD(stepLog)
				err = Inst().V.WaitDriverUpOnNode(*nodeSelected, Inst().DriverStartTimeout)
				log.FailOnError(err, fmt.Sprintf("Driver is down on node %s", nodeSelected.Name))
				log.Info("Driver is up")

				status, err := Inst().V.GetPxctlStatus(*nodeSelected)
				log.FailOnError(err, fmt.Sprintf("failed to get pxctl status on node [%s]", nodeSelected.Name))
				dash.VerifyFatal(status == api.Status_STATUS_OK.String(), true, fmt.Sprintf("node [%s] status is up but PX cluster is not ok. Expected: %v Actual: %v",
					nodeSelected.Name, api.Status_STATUS_OK, status))
				log.InfoD("px status %v", status)
			})

		}

		originalSizeInBytes = poolToResize.TotalSize
		targetSizeInBytes = originalSizeInBytes + 128*units.GiB // getDesiredSize(originalSizeInBytes)
		targetSizeGiB = targetSizeInBytes / units.GiB

		stepLog = "Expanding pool by adding drive of 128GiB"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			log.InfoD("Current size of pool [%s] is [%d] GiB. Expand to [%v] GiB with type add-disk...",
				poolUUIDSelected, poolToResize.TotalSize/units.GiB, targetSizeGiB)
			triggerPoolExpansion(poolUUIDSelected, targetSizeGiB, api.SdkStoragePool_RESIZE_TYPE_ADD_DISK)
			log.InfoD(fmt.Sprintf("Pool expansion succeed [%s]", poolUUIDSelected))
		})

		if testName == "AddDriveWithNodeMaintenance" {
			//sleep for some random time
			sleepTime := rand.Intn(100) + 1
			time.Sleep(time.Second * (time.Duration(sleepTime)))

			//node maintenance cycle
			stepLog = "Perform node maintenance cycle"
			Step(stepLog, func() {
				log.InfoD(stepLog)
				log.InfoD(fmt.Sprintf("Performing node maintenance cycle on node %s", nodeSelected.Name))
				err = Inst().V.RecoverDriver(*nodeSelected)
				log.FailOnError(err, fmt.Sprintf("error performing maintenance cycle on node %s", nodeSelected.Name))
				log.Info("node maintenance cycle succeeded")
			})
		} else if testName == "PoolExpandAddDriveWithNodeReboot" {
			// sleep for some random time
			sleepTime := rand.Intn(100) + 1
			time.Sleep(time.Second * (time.Duration(sleepTime)))

			// node restart
			stepLog = fmt.Sprintf("Verify reboot after [%d] seconds", sleepTime)
			Step(stepLog, func() {
				log.InfoD(stepLog)
				log.Info(stepLog)
				err = Inst().N.RebootNodeAndWait(*nodeSelected)
				log.FailOnError(err, "Failed to reboot node and wait till it is up")
				log.InfoD("Verify reboot succeed")
				time.Sleep(2 * time.Minute)
			})

		} else if testName == "PoolExpandAddDriveWithPXRestart" {
			//sleep for some random time
			sleepTime := rand.Intn(100) + 1
			time.Sleep(time.Second * (time.Duration(sleepTime)))

			//Restart Portworx
			Step("Restart Portworx", func() {

				err := Inst().N.Systemctl(*nodeSelected, "portworx.service", node.SystemctlOpts{
					Action: "restart",
					ConnectionOpts: node.ConnectionOpts{
						Timeout:         5 * time.Minute,
						TimeBeforeRetry: 10 * time.Second,
					}})
				log.FailOnError(err, "failed to restart portworx on node [%v]", nodeSelected.Name)
				log.Info("portworx restart succeed")
			})

		}
		stepLog = fmt.Sprintf("Wait for driver to up")
		Step(stepLog, func() {
			log.InfoD(stepLog)
			err = Inst().V.WaitDriverUpOnNode(*nodeSelected, Inst().DriverStartTimeout)
			log.FailOnError(err, fmt.Sprintf("Driver is down on node %s", nodeSelected.Name))
			log.Info("Driver is up")

		})

		stepLog = fmt.Sprint("Wait for pool to expand")
		Step(stepLog, func() {
			log.InfoD(stepLog)
			resizeErr := waitForOngoingPoolExpansionToComplete(poolUUIDSelected)
			dash.VerifyFatal(resizeErr, nil, "Pool expansion does not result in error")
			log.Info(fmt.Sprintf("Pool expansion succeed [%s]", poolUUIDSelected))
		})

		//Check PX status
		stepLog = "Check PX status"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			status, err := Inst().V.GetPxctlStatus(*nodeSelected)
			log.FailOnError(err, fmt.Sprintf("failed to get pxctl status on node [%s]", nodeSelected.Name))
			dash.VerifyFatal(status == api.Status_STATUS_OK.String(), true, fmt.Sprintf("node [%s] status is up but PX cluster is not ok. Expected: %v Actual: %v",
				nodeSelected.Name, api.Status_STATUS_OK, status))
			log.Infof("px status %v", status)
		})

		//verifying pool resize
		stepLog = fmt.Sprintf("Verify pool resized")
		Step("Verify pool resized", func() {
			log.InfoD(stepLog)
			verifyPoolSizeEqualOrLargerThanExpected(poolUUIDSelected, targetSizeGiB)
			log.Info("verifying pool resize succeeded")
		})

	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts)
	})
}

var _ = Describe("{ExpandPoolAddDriveAfterYankingDrive}", func() {
	testName = "ExpandPoolAddDriveAfterYankingDrive"
	testDescription = "Yank pool drive with expand pool add drive"
	ExpandPoolWithAddDrive(testName, testDescription)
})
var _ = Describe("{YankPoolDriveWithNodeMaintenanceCycle}", func() {
	testName = "yank-pool-drive-node-maintenance-cycle"
	testDescription = "Yank pool drive with node maintenance cycle"
	YankPoolDriveTest(testName, testDescription)
})
var _ = Describe("{YankMetadataWithNodeMaintenanceCycle}", func() {
	testName = "YankMetadataWithNodeMaintenanceCycle"
	testDescription = "Yank metadata drive and node maintenance cycle"
	YankMetadataTest(testName, testDescription)
})
var _ = Describe("{DeletePoolAfterYankDriveWithPXRestart}", func() {
	testName = "delete-yank-drive-px-restart"
	testDescription = "Delete pool after yanking drive and restart PX"
	DeletePoolAfterYankDrive(testName, testDescription)
})

var _ = Describe("{DeletePoolAfterYankDriveWithNodeReboot}", func() {
	testName = "delete-yank-drive-node-reboot"
	testDescription = "Delete pool after yanking drive and reboot node"
	DeletePoolAfterYankDrive(testName, testDescription)
})

var _ = Describe("{DeletePoolAfterYankDriveWithNodeMaintenanceCycle}", func() {
	testName := "delete-yank-drive-node-maintenance-cycle"
	testDescription = "Delete pool after yanking drive and do node maintenance"
	DeletePoolAfterYankDrive(testName, testDescription)
})

func DeletePoolAfterYankDrive(testName, testDesc string) {

	var nodeSelected node.Node
	var busID, poolDrive string
	var poolIDSelected int

	JustBeforeEach(func() {
		StartTorpedoTest(testName, testDesc, nil, 0)
	})

	itLog := "Delete pool after yanking drive"
	It(itLog, func() {
		log.InfoD(itLog)

		stepLog = "Select pool to yank drive and delete"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			//select a storage node
			storageNodes := node.GetStorageNodes()
			index := rand.Intn(len(storageNodes))
			nodeSelected = storageNodes[index]
			log.Info("selected Node ID - %s , Name - %s", nodeSelected.Id, nodeSelected.Name)

			jDev, err := Inst().V.GetJournalDevicePath(&nodeSelected)
			log.FailOnError(err, fmt.Sprintf("error getting journal device path from node %s", nodeSelected.Name))
			log.InfoD("Journal device path - %s", jDev)

			cmd := fmt.Sprintf("lsblk -no pkname %s", jDev)
			journalParentDevPath, err := Inst().N.RunCommandWithNoRetry(nodeSelected, cmd, node.ConnectionOpts{
				Timeout:         2 * time.Minute,
				TimeBeforeRetry: 10 * time.Second,
			})
			log.FailOnError(err, "error occured running the command to identify the parent device path of the journal partition %s", jDev)
			journalParentDevPath = strings.TrimRight(journalParentDevPath, "\n")
			log.InfoD("Parent device path of the journal device is %s", journalParentDevPath)

			driveMap, err := Inst().V.GetPoolDrives(&nodeSelected)
			log.FailOnError(err, fmt.Sprintf("error getting pool drive details for the node %s", nodeSelected))

			for poolID, drives := range driveMap {
				if poolDrive == "" {
					for _, drive := range drives {
						if strings.Contains(drive.Device, journalParentDevPath) || drive.Device == "" {
							break
						}
						poolIDSelected, _ = strconv.Atoi(poolID)
						log.InfoD("Pool selected to delete is %d", int(poolIDSelected))
						poolDrive = drive.Device
						log.InfoD("Pool drive selected - %s", poolDrive)
						break
					}
				}
			}
		})

		poolDrive = strings.Trim(poolDrive, "/")
		poolDriveArr := strings.Split(poolDrive, "/")
		poolDrive = poolDriveArr[len(poolDriveArr)-1]
		stepLog = "Yank journal drive"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			busID, err = Inst().N.YankDrive(nodeSelected, poolDrive, node.ConnectionOpts{
				Timeout:         dfDefaultTimeout,
				TimeBeforeRetry: dfDefaultRetryInterval,
			})
			log.FailOnError(err, fmt.Sprintf("failed to yank journal drive on node [%s]", nodeSelected.Name))
			log.InfoD("Bus id - %s", busID)
		})

		time.Sleep(2 * time.Minute)

		stepLog = "Recover yank drive"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			err = Inst().N.RecoverDrive(nodeSelected, poolDrive, busID, node.ConnectionOpts{
				Timeout:         driveFailTimeout,
				TimeBeforeRetry: dfDefaultRetryInterval,
			})
			log.FailOnError(err, fmt.Sprintf("failed to recover yank journal drive on node [%s]", nodeSelected.Name))
			log.InfoD("Verified recover yank drive")
		})

		nodePoolMapBeforePoolDelete, err := Inst().V.GetNodePools(nodeSelected)
		log.FailOnError(err, fmt.Sprintf("Get Node pools failed on node %s", nodeSelected.Name))

		stepLog = "Delete the selected pool"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			err = DeletePoolAndValidate(nodeSelected, strconv.Itoa(int(poolIDSelected)))
			log.FailOnError(err, fmt.Sprintf("Error occured while Validating the deleted pool %d in the node %s", int(poolIDSelected), nodeSelected.Name))
		})

		if testName == "delete-yank-drive-px-restart" {
			stepLog = "Restart portworx and wait for it to come up"
			Step(stepLog, func() {
				log.Info(stepLog)
				Step(fmt.Sprintf("node with Px restart is: %s", nodeSelected.Name), func() {
					err := Inst().V.RestartDriver(nodeSelected, nil)
					log.FailOnError(err, fmt.Sprintf("Error occured while Restart PX on node:%v", nodeSelected.Name))
				})

				Step(fmt.Sprintf("wait for volume driver to restart on node: %v", nodeSelected.Name), func() {
					err := Inst().V.WaitForPxPodsToBeUp(nodeSelected)
					log.FailOnError(err, fmt.Sprintf("Error occured while Validating PX restart is done on node:%v", nodeSelected.Name))
				})
			})

		} else if testName == "delete-yank-drive-node-reboot" {
			stepLog = "Reboot the node"
			Step(stepLog, func() {
				log.Info(stepLog)
				err = RebootNodeAndWaitForPxUp(nodeSelected)
				log.FailOnError(err, "Failed to reboot node and wait till it is up")
			})
		} else if testName == "delete-yank-drive-node-maintenance-cycle" {
			stepLog := "start node maintenance cycle"
			Step(stepLog, func() {
				log.InfoD(stepLog)
				log.InfoD(fmt.Sprintf("Performing node maintenance cycle on node %s", nodeSelected.Name))
				err = Inst().V.RecoverDriver(nodeSelected)
				log.FailOnError(err, fmt.Sprintf("error performing maintenance cycle on node %s", nodeSelected.Name))

				err = Inst().V.WaitDriverUpOnNode(nodeSelected, 5*time.Minute)
				log.FailOnError(err, fmt.Sprintf("Driver is down on node %s", nodeSelected.Name))
				dash.VerifyFatal(err == nil, true, fmt.Sprintf("PX is up after maintenance cycle on node %s", nodeSelected.Name))
			})
		}

		stepLog = "Verify Px Status"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			err := Inst().V.WaitForPxPodsToBeUp(nodeSelected)
			log.FailOnError(err, fmt.Sprintf("Error occured while Validating PX restart is done on node:%v", nodeSelected.Name))
		})

		stepLog = "Verify pool is deleted"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			nodePoolMapAftrPoolDelete, err := Inst().V.GetNodePools(nodeSelected)
			log.FailOnError(err, fmt.Sprintf("Get Node pools failed on node %s", nodeSelected.Name))
			dash.VerifyFatal(len(nodePoolMapAftrPoolDelete) < len(nodePoolMapBeforePoolDelete), true, fmt.Sprintf("Verify pool is deleted"))
		})
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
	})
}

func YankPoolDriveTest(testName, testDesc string) {
	var (
		nodeSelected     *node.Node
		busID, poolDrive string
		contexts         = make([]*scheduler.Context, 0)
	)

	JustBeforeEach(func() {
		StartTorpedoTest(testName, testDesc, nil, 0)
	})

	itLog := testDesc
	It(itLog, func() {
		log.InfoD(itLog)

		stepLog := "Schedule apps to perform IOs"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				contexts = append(contexts, ScheduleApplications(fmt.Sprintf("yankjournal-%d", i))...)
			}
			ValidateApplications(contexts)
		})
		defer appsValidateAndDestroy(contexts)

		ctx := contexts[0]
		volumes, err := Inst().S.GetVolumes(ctx)
		log.FailOnError(err, "Failed while listing the volume with error")
		log.InfoD("Vol deatils %v", volumes)

		if len(volumes) == 0 {
			msg := fmt.Sprintf("There are no volumes associated with the app %v", ctx.App.Key)
			log.InfoD(msg)
			Skip(msg)
		}
		volumeSelected := volumes[0]

		stepLog = "Select the volume replica node and pool drive"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			rsDetails, err := Inst().V.GetReplicaSets(volumeSelected)
			log.FailOnError(err, fmt.Sprintf("error getting replica sets for vol %s", volumeSelected.Name))
			log.InfoD("Volume Replica info %v", rsDetails)

			poolUUIDs := rsDetails[0].GetPoolUuids()
			poolUUIDSelected := poolUUIDs[rand.Intn(len(poolUUIDs))]
			poolIDSelected, err := GetPoolIDFromPoolUUID(poolUUIDSelected)
			log.FailOnError(err, fmt.Sprintf("error getting pool id for the pool %s", poolUUIDSelected))

			nodeSelected, err = GetNodeWithGivenPoolID(poolUUIDSelected)
			log.FailOnError(err, fmt.Sprintf("error getting node detail for the pool %s", poolUUIDSelected))

			jDev, err := Inst().V.GetJournalDevicePath(nodeSelected)
			log.FailOnError(err, fmt.Sprintf("error getting journal device path from node %s", nodeSelected.Name))
			log.InfoD("Journal device path - %s", jDev)

			cmd := fmt.Sprintf("lsblk -no pkname %s", jDev)
			journalParentDevPath, err := Inst().N.RunCommandWithNoRetry(*nodeSelected, cmd, node.ConnectionOpts{
				Timeout:         2 * time.Minute,
				TimeBeforeRetry: 10 * time.Second,
			})
			log.FailOnError(err, "error occured running the command to identify the parent device path of the journal partition %s", jDev)
			journalParentDevPath = strings.TrimRight(journalParentDevPath, "\n")
			log.InfoD("Parent device path of the journal device is %s", journalParentDevPath)

			driveMap, err := Inst().V.GetPoolDrives(nodeSelected)
			log.FailOnError(err, fmt.Sprintf("error getting pool drive details for the node %s", nodeSelected))

			drives := driveMap[strconv.Itoa(int(poolIDSelected))]

			for _, drive := range drives {
				cmd := fmt.Sprintf("lsblk -no pkname %s", drive.Device)
				poolDriveParentPath, err := Inst().N.RunCommandWithNoRetry(*nodeSelected, cmd, node.ConnectionOpts{
					Timeout:         2 * time.Minute,
					TimeBeforeRetry: 10 * time.Second,
				})
				log.FailOnError(err, "error occured running the command to identify the parent device path of the drive %s", drive.Device)
				poolDriveParentPath = strings.TrimRight(poolDriveParentPath, "\n")
				log.InfoD("Parent device path of the pool device is %s", poolDriveParentPath)

				if (journalParentDevPath != "") && (poolDriveParentPath != "") && (journalParentDevPath == poolDriveParentPath) {
					continue
				}

				poolDrive = drive.Device
				log.InfoD("Pool drive selected - %s", poolDrive)
				break
			}
		})

		//random delay in secs
		time.Sleep(time.Minute * 5)

		poolDrive = strings.Trim(poolDrive, "/")
		poolDriveArr := strings.Split(poolDrive, "/")
		poolDrive = poolDriveArr[len(poolDriveArr)-1]
		stepLog = "Yank journal drive"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			busID, err = Inst().N.YankDrive(*nodeSelected, poolDrive, node.ConnectionOpts{
				Timeout:         dfDefaultTimeout,
				TimeBeforeRetry: dfDefaultRetryInterval,
			})
			log.FailOnError(err, fmt.Sprintf("failed to yank journal drive on node [%s]", nodeSelected.Name))
			log.InfoD("Bus id - %s", busID)
		})

		if testName == "yank-pool-drive-node-reboot" {
			stepLog = "Reboot the node"
			Step(stepLog, func() {
				log.Info(stepLog)
				err = RebootNodeAndWaitForPxUp(*nodeSelected)
				log.FailOnError(err, "Failed to reboot node and wait till it is up")
			})
		} else if testName == "yank-pool-drive-px-restart" {
			stepLog = "Restart portworx and wait for it to come up"
			Step(stepLog, func() {
				log.Info(stepLog)
				Step(fmt.Sprintf("node with Px restart is: %s", nodeSelected.Name), func() {
					err := Inst().V.RestartDriver(*nodeSelected, nil)
					log.FailOnError(err, fmt.Sprintf("Error occured while Restart PX on node:%v", nodeSelected.Name))
				})

				Step(fmt.Sprintf("wait for volume driver to restart on node: %v", nodeSelected.Name), func() {
					err := Inst().V.WaitForPxPodsToBeUp(*nodeSelected)
					log.FailOnError(err, fmt.Sprintf("Error occured while Validating PX restart is done on node:%v", nodeSelected.Name))
				})
			})
		} else if testName == "yank-pool-drive-node-maintenance-cycle" {
			stepLog = "Enter maintenance mode"
			Step(stepLog, func() {
				log.Info(stepLog)
				err = Inst().V.EnterMaintenance(*nodeSelected)
				log.FailOnError(err, fmt.Sprintf("fail to enter node %s in maintenance mode", nodeSelected.Name))
				status, err := Inst().V.GetNodeStatus(*nodeSelected)
				log.FailOnError(err, fmt.Sprintf("Error getting PX status of node %s", nodeSelected.Name))
				dash.VerifyFatal(*status, api.Status_STATUS_MAINTENANCE, fmt.Sprintf("Node %s Status not Online", nodeSelected.Name))
			})
		}

		if testName == "yank-pool-drive-IO" {
			//wait for two minutes before recovering pool drive
			time.Sleep(2 * time.Minute)
		}

		stepLog = "Recover yank drive"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			err = Inst().N.RecoverDrive(*nodeSelected, poolDrive, busID, node.ConnectionOpts{
				Timeout:         driveFailTimeout,
				TimeBeforeRetry: dfDefaultRetryInterval,
			})
			log.FailOnError(err, fmt.Sprintf("failed to recover yank journal drive on node [%s]", nodeSelected.Name))
			log.InfoD("Verified recover yank drive")
		})

		if testName == "yank-pool-drive-node-maintenance-cycle" {
			stepLog = "Exit maintenance mode"
			Step(stepLog, func() {
				err = Inst().V.ExitMaintenance(*nodeSelected)
				log.FailOnError(err, fmt.Sprintf("fail to exit node %s in maintenance mode", nodeSelected.Name))
				status, err := Inst().V.GetNodeStatus(*nodeSelected)
				log.FailOnError(err, fmt.Sprintf("Error getting PX status of node %s", nodeSelected.Name))
				dash.VerifyFatal(*status, api.Status_STATUS_OK, fmt.Sprintf("Node %s Status not Online", nodeSelected.Name))
			})
		}

		stepLog = "Verify Px Status"
		Step(stepLog, func() {
			log.InfoD(stepLog)

			err := Inst().V.WaitForPxPodsToBeUp(*nodeSelected)
			log.FailOnError(err, fmt.Sprintf("Error occured while Validating PX restart is done on node:%v", nodeSelected.Name))

			status, err := Inst().V.GetPxctlStatus(*nodeSelected)
			log.FailOnError(err, fmt.Sprintf("failed to get pxctl status on node [%s]", nodeSelected.Name))
			dash.VerifyFatal(status == api.Status_STATUS_OK.String(), true, fmt.Sprintf("node [%s] status is up but PX cluster is not ok. Expected: %v Actual: %v",
				nodeSelected.Name, api.Status_STATUS_OK, status))
			log.InfoD("px status %v", status)
		})

		stepLog = "Do pool maintenance"
		Step(stepLog, func() {
			log.Info(stepLog)
			log.InfoD(fmt.Sprintf("Performing pool maintenance cycle on node %s", nodeSelected.Name))
			err = Inst().V.RecoverPool(*nodeSelected)
			log.FailOnError(err, fmt.Sprintf("error performing pool maintenance cycle on node %s", nodeSelected.Name))
		})

		stepLog = "Verify pool status"
		Step(stepLog, func() {
			log.Info(stepLog)
			poolsStatus, err := Inst().V.GetNodePoolsStatus(*nodeSelected)
			log.FailOnError(err, "error getting pool status on node %s", nodeSelected.Name)
			for poolID, status := range poolsStatus {
				dash.VerifyFatal(status, "Online", fmt.Sprintf("Pool %s Status not Online", poolID))
			}
		})

	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts)
	})
}

var _ = Describe("{ExpandPoolResizeDriveAfterYankingDrive}", Label("p1", "PoolExpand", "YankPoolDrive"), func() {
	testName = "ExpandPoolResizeDriveAfterYankingDrive"
	testDescription = "Yank pool drive with expand pool resize drive"
	PoolResize(testName, testDescription)
})

var _ = Describe("{DeletePoolAfterYankDriveWithPXRestart}", Label("p1", "hal_ops_disruption", "px_restart", "PoolDelete", "drive_failure", "YankPoolDrive", "functional"), func() {
	testName = "yank-drive-pool-delete-px-restart"
	testDescription = "Delete pool after yanking drive and restart PX"
	PoolDeleteWithTimeInterval(testName, testDescription, "px_restart")
})

var _ = Describe("{DeletePoolAfterYankDriveWithNodeReboot}", Label("p1", "hal_ops_disruption", "node_reboot", "PoolDelete", "drive_failure", "YankPoolDrive", "functional"), func() {
	testName = "yank-drive-pool-delete-node-reboot"
	testDescription = "Delete pool after yanking drive and reboot node"
	PoolDeleteWithTimeInterval(testName, testDescription, "node_reboot")
})

var _ = Describe("{DeletePoolAfterYankDriveWithNodeMaintenanceCycle}", Label("p1", "hal_ops_disruption", "NodeMaintenance", "PoolDelete", "drive_failure", "YankPoolDrive", "functional"), func() {
	testName := "yank-drive-pool-delete-node-maintenance-cycle"
	testDescription = "Delete pool after yanking drive and do node maintenance"
	PoolDeleteWithTimeInterval(testName, testDescription, "node_maintenance_cycle")
})

var _ = Describe("{DeletePoolAndYankDriveWithPXRestart}", Label("p1", "hal_ops_disruption", "px_restart", "PoolDelete", "drive_failure", "YankPoolDrive", "functional"), func() {
	testName = "delete-pool-yank-drive-px-restart"
	testDescription = "Delete pool and yank drive with PX restart"
	PoolDeleteWithTimeInterval(testName, testDescription, "px_restart")
})

var _ = Describe("{DeletePoolAndYankDriveWithNodeReboot}", Label("p1", "hal_ops_disruption", "node_reboot", "PoolDelete", "drive_failure", "YankPoolDrive", "functional"), func() {
	testName = "delete-pool-yank-drive-node-reboot"
	testDescription = "Delete pool and yank drive with node reboot"
	PoolDeleteWithTimeInterval(testName, testDescription, "node_reboot")
})

var _ = Describe("{DeletePoolAndYankDriveWithNodeMaintenanceCycle}", Label("p1", "hal_ops_disruption", "NodeMaintenance", "PoolDelete", "drive_failure", "YankPoolDrive", "functional"), func() {
	testName := "delete-pool-yank-drive-node-maintenance-cycle"
	testDescription = "Delete pool and yank drive with node maintenance"
	PoolDeleteWithTimeInterval(testName, testDescription, "node_maintenance_cycle")
})

func PoolResize(testName, description string) {

	var (
		poolUUIDSelected              string
		poolToResize                  *api.StoragePool
		bufferSizeInGB, targetSizeGiB uint64
		isJournalEnabled              bool
		contexts                      []*scheduler.Context
		nodeSelected                  *node.Node
		poolDrive                     string
		busID                         string
	)

	JustBeforeEach(func() {
		StartTorpedoTest(testName, description, nil, 0)
		isJournalEnabled, err = IsJournalEnabled()
		log.FailOnError(err, "Failed to get journal enable or not")
		bufferSizeInGB = uint64(0)
		if isJournalEnabled {
			bufferSizeInGB = JournalDeviceSizeInGB
		}
	})

	It(testName, func() {
		poolDeleteMap := getDeletablePoolMap()
		stepLog = "Select a pool to resize"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for poolID, _ := range poolDeleteMap {
				n, err := GetNodeWithGivenPoolID(poolID)
				failOnError(err, "failed to get node details from PoolUUID [%v]", poolID)
				eligibilityMap, err := GetPoolExpansionEligibility(n, api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK, 100)
				if err != nil {
					log.Warnf("Error identifying pool expansion eligibility, Error: %v", err)
					continue
				}
				if eligibilityMap[n.Id] && eligibilityMap[poolID] {
					nodeSelected = n
					poolUUIDSelected = poolID
					break
				} else {
					log.Infof("Excluding pool [%s] from resize as it is on node [%s] as it is not eligible for expansion", poolID, n.Id)
				}
			}
			log.Info(fmt.Sprintf("Pool going to resize is UUID: %s", poolUUIDSelected))
		})

		dash.VerifyFatal(poolUUIDSelected != "", true, "no pool is found which is eligible for expansion and deletion")

		stepLog = "Get node details by pool uuid and add pool label"
		poolLabelToUpdate := make(map[string]string)
		labels := []string{"SSD"}
		Step(stepLog, func() {
			log.Infof(stepLog)
			poolLabelToUpdate["mediatype"] = labels[0]
			err = Inst().V.UpdatePoolLabels(*nodeSelected, poolUUIDSelected, poolLabelToUpdate)
			dash.VerifyFatal(err, nil, "Check if able to update the label on the pool")

		})

		stepLog = "Apply volume placement strategy"
		vpsName := fmt.Sprintf("mongo-vps-%v", time.Now().Unix())
		Step(stepLog, func() {
			log.Infof(stepLog)
			vpsSpec := vpsutil.ReplicaAffinityPool(vpsName)
			_, err = talisman.Instance().CreateVolumePlacementStrategy(&vpsSpec)
			dash.VerifyFatal(err, nil, "Check if able to apply volume placement strategy")
		})

		scName := fmt.Sprintf("mongo-sc-%v", time.Now().Unix())
		params := make(map[string]string)
		stepLog = "Apply storage class"
		k8sStorage := storage.Instance()
		Step(stepLog, func() {
			log.Infof(stepLog)
			params["repl"] = "1"
			params["placement_strategy"] = vpsName
			v1obj := metav1.ObjectMeta{
				Name: scName,
			}
			bindMode := storageApi.VolumeBindingImmediate
			scObj := storageApi.StorageClass{
				ObjectMeta:        v1obj,
				Provisioner:       k8s.CsiProvisioner,
				Parameters:        params,
				VolumeBindingMode: &bindMode,
			}
			_, err := k8sStorage.CreateStorageClass(&scObj)
			dash.VerifyFatal(err, nil, "Verifying creation of new storage class")
		})

		stepLog = "Apply persistent volume claim"
		pvcName := fmt.Sprintf("mongo-pvc-%v", time.Now().Unix())

		Step(stepLog, func() {
			log.Infof(stepLog)
			_, err := core.Instance().CreatePersistentVolumeClaim(&corev1.PersistentVolumeClaim{
				TypeMeta: metav1.TypeMeta{
					Kind: "PersistentVolumeClaim",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: pvcName,
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					AccessModes:      []corev1.PersistentVolumeAccessMode{corev1.ReadWriteMany},
					StorageClassName: &scName,
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceStorage: resource.MustParse("10Gi"),
						},
					},
				},
			})
			dash.VerifyFatal(err, nil, "Verifying creation of new storage class")
		})
		poolToResize = getStoragePool(poolUUIDSelected)
		originalSize := poolToResize.TotalSize

		stepLog := "Schedule Apps"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			contexts = scheduleApps()
			log.Info("schedule app succeed")
			time.Sleep(5 * time.Minute)
		})

		if testName == "ExpandPoolResizeDriveAfterYankingDrive" {
			stepLog = "Select pool drive"
			Step(stepLog, func() {
				log.InfoD(stepLog)
				poolIDSelected, err := GetPoolIDFromPoolUUID(poolUUIDSelected)
				log.FailOnError(err, fmt.Sprintf("error getting pool id for the pool %s", poolUUIDSelected))
				jDev, err := Inst().V.GetJournalDevicePath(nodeSelected)
				log.FailOnError(err, fmt.Sprintf("error getting journal device path from node %s", nodeSelected.Name))
				log.InfoD("Journal device path - %s", jDev)

				cmd := fmt.Sprintf("lsblk -no pkname %s", jDev)
				journalParentDevPath, err := Inst().N.RunCommandWithNoRetry(*nodeSelected, cmd, node.ConnectionOpts{
					Timeout:         2 * time.Minute,
					TimeBeforeRetry: 10 * time.Second,
				})
				log.FailOnError(err, "error occured running the command to identify the parent device path of the journal partition %s", jDev)
				journalParentDevPath = strings.TrimRight(journalParentDevPath, "\n")
				log.InfoD("Parent device path of the journal device is %s", journalParentDevPath)

				driveMap, err := Inst().V.GetPoolDrives(nodeSelected)
				log.FailOnError(err, fmt.Sprintf("error getting pool drive details for the node %s", nodeSelected))

				drives := driveMap[strconv.Itoa(int(poolIDSelected))]

				for _, drive := range drives {
					cmd := fmt.Sprintf("lsblk -no pkname %s", drive.Device)
					poolDriveParentPath, err := Inst().N.RunCommandWithNoRetry(*nodeSelected, cmd, node.ConnectionOpts{
						Timeout:         2 * time.Minute,
						TimeBeforeRetry: 10 * time.Second,
					})
					log.FailOnError(err, "error occured running the command to identify the parent device path of the drive %s", drive.Device)
					poolDriveParentPath = strings.TrimRight(poolDriveParentPath, "\n")
					log.InfoD("Parent device path of the pool device is %s", poolDriveParentPath)

					if (journalParentDevPath != "") && (poolDriveParentPath != "") && (journalParentDevPath == poolDriveParentPath) {
						continue
					}

					poolDrive = drive.Device
					log.InfoD("Pool drive selected - %s", poolDrive)
					break
				}
			})

			poolDrive = strings.Trim(poolDrive, "/")
			poolDriveArr := strings.Split(poolDrive, "/")
			poolDrive = poolDriveArr[len(poolDriveArr)-1]
			stepLog = "Yank drive"
			Step(stepLog, func() {
				log.InfoD(stepLog)
				busID, err = Inst().N.YankDrive(*nodeSelected, poolDrive, node.ConnectionOpts{
					Timeout:         dfDefaultTimeout,
					TimeBeforeRetry: dfDefaultRetryInterval,
				})
				log.FailOnError(err, fmt.Sprintf("failed to yank journal drive on node [%s]", nodeSelected.Name))
				log.InfoD("Bus id - %s", busID)
			})

			time.Sleep(time.Minute * 2)

			stepLog = "Recover yank drive"
			Step(stepLog, func() {
				log.InfoD(stepLog)
				err = Inst().N.RecoverDrive(*nodeSelected, poolDrive, busID, node.ConnectionOpts{
					Timeout:         driveFailTimeout,
					TimeBeforeRetry: dfDefaultRetryInterval,
				})
				log.FailOnError(err, fmt.Sprintf("failed to recover yank journal drive on node [%s]", nodeSelected.Name))
				log.InfoD("Verified recover yank drive")
			})

			stepLog = fmt.Sprintf("Wait for driver to up and check px status")
			Step(stepLog, func() {
				log.InfoD(stepLog)
				err = Inst().V.WaitDriverUpOnNode(*nodeSelected, Inst().DriverStartTimeout)
				log.FailOnError(err, fmt.Sprintf("Driver is down on node %s", nodeSelected.Name))
				log.Info("Driver is up")

				status, err := Inst().V.GetPxctlStatus(*nodeSelected)
				log.FailOnError(err, fmt.Sprintf("failed to get pxctl status on node [%s]", nodeSelected.Name))
				dash.VerifyFatal(status == api.Status_STATUS_OK.String(), true, fmt.Sprintf("node [%s] status is up but PX cluster is not ok. Expected: %v Actual: %v",
					nodeSelected.Name, api.Status_STATUS_OK, status))
				log.InfoD("px status %v", status)
			})

		}

		originalSizeInBytes := poolToResize.TotalSize
		targetSizeInBytes := originalSizeInBytes + 100*units.GiB // getDesiredSize(originalSizeInBytes)
		targetSizeGiB = targetSizeInBytes / units.GiB

		stepLog = "Expand it by 100 GiB with resize-disk type"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			log.InfoD("Current size of pool %s is %d GiB. Trying to expand to %v GiB",
				poolUUIDSelected, poolToResize.TotalSize/units.GiB, targetSizeGiB+bufferSizeInGB)
			triggerPoolExpansion(poolUUIDSelected, targetSizeGiB+bufferSizeInGB, api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK)
			log.Info(fmt.Sprintf("Pool expansion started [%s]", poolUUIDSelected))
		})

		if testName == "PoolResizeWithNodeMaintenanceCycleWithTimeInterval" {
			sleepTime := rand.Intn(100) + 1
			time.Sleep(time.Second * (time.Duration(sleepTime)))

			stepLog = fmt.Sprintf("Performing node maintenance cycle on node [%s] after [%d] seconds", nodeSelected.Name, sleepTime)
			Step(stepLog, func() {
				log.InfoD(stepLog)
				err = Inst().V.RecoverDriver(*nodeSelected)
				log.FailOnError(err, fmt.Sprintf("error performing maintenance cycle on node %s", nodeSelected.Name))
				dash.VerifyFatal(err == nil, true, fmt.Sprintf("PX is up after maintenance cycle on node %s", nodeSelected.Name))
			})
		} else if testName == "PoolResizeWithNodeRebootWithTimeInterval" {
			sleepTime := rand.Intn(100) + 1
			time.Sleep(time.Second * (time.Duration(sleepTime)))

			stepLog = fmt.Sprintf("Verify reboot after [%d] seconds", sleepTime)
			Step(stepLog, func() {
				log.InfoD(stepLog)
				err = Inst().N.RebootNodeAndWait(*nodeSelected)
				log.FailOnError(err, "Failed to reboot node and wait till it is up")
				log.Info("Verify reboot succeed")
			})

		} else if testName == "PoolResizeWithPXRestartWithTimeInterval" {
			sleepTime := rand.Intn(100) + 1
			time.Sleep(time.Second * (time.Duration(sleepTime)))

			stepLog = fmt.Sprintf("Restart Portworx after [%d] seconds", sleepTime)
			Step(stepLog, func() {
				log.InfoD(stepLog)
				err := Inst().N.Systemctl(*nodeSelected, "portworx.service", node.SystemctlOpts{
					Action: "restart",
					ConnectionOpts: node.ConnectionOpts{
						Timeout:         5 * time.Minute,
						TimeBeforeRetry: defaultRetryInterval,
					}})
				log.FailOnError(err, "failed to restart portworx on node [%v]", nodeSelected.Name)
				log.Infof("restarting portworx is done")
			})
		}

		stepLog = fmt.Sprintf("Wait for driver to up")
		Step(stepLog, func() {
			log.InfoD(stepLog)
			err = Inst().V.WaitDriverUpOnNode(*nodeSelected, Inst().DriverStartTimeout)
			log.FailOnError(err, fmt.Sprintf("Driver is down on node %s", nodeSelected.Name))
			log.Info("Driver is up")

		})

		stepLog = fmt.Sprint("Wait for pool to resize")
		Step(stepLog, func() {
			log.InfoD(stepLog)
			resizeErr := waitForOngoingPoolExpansionToComplete(poolUUIDSelected)
			dash.VerifyFatal(resizeErr, nil, "Pool expansion does not result in error")
			log.Info(fmt.Sprintf("Pool expansion succeed [%s]", poolUUIDSelected))
		})

		stepLog = fmt.Sprintf("Verify pool resized [%s]", poolUUIDSelected)
		Step(stepLog, func() {
			log.InfoD(stepLog)
			verifyPoolSizeEqualOrLargerThanExpected(poolUUIDSelected, targetSizeGiB)
			log.Info("Verify pool resized succeed")
		})

		stepLog = "Check px status"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			status, err := Inst().V.GetPxctlStatus(*nodeSelected)
			log.FailOnError(err, fmt.Sprintf("failed to get pxctl status on node [%s]", nodeSelected.Name))
			dash.VerifyFatal(status == api.Status_STATUS_OK.String(), true, fmt.Sprintf("node [%s] status is up but PX cluster is not ok. Expected: %v Actual: %v",
				nodeSelected.Name, api.Status_STATUS_OK, status))
			log.InfoD("px status %v", status)
		})

		stepLog = fmt.Sprintf("Check if apps are running")
		Step(stepLog, func() {
			log.InfoD(stepLog)
			ValidateApplications(contexts)
			log.Info("validate application succeed")
		})

		stepLog = fmt.Sprintf("Delete pool %s", poolToResize.Uuid)
		Step(stepLog, func() {
			log.InfoD(stepLog)
			err = DeletePoolAndValidate(*nodeSelected, strconv.Itoa(int(poolToResize.GetID())))
			log.FailOnError(err, fmt.Sprintf("Error occured while Validating the deleted pool %s in the node %s", poolToResize.Uuid, nodeSelected.Name))
			log.InfoD("pool [%d] delete succed", poolToResize.GetID())
		})

		stepLog = "Add new pools with original size"
		Step(stepLog, func() {
			log.Info(stepLog)
			driveSpecs, err := GetCloudDriveDeviceSpecs()
			log.FailOnError(err, "Error getting cloud drive specs")
			deviceSpec := driveSpecs[0]
			deviceSpecParams := strings.Split(deviceSpec, ",")

			paramsArr := make([]string, 0)
			for _, param := range deviceSpecParams {
				if strings.Contains(param, "size") {
					paramsArr = append(paramsArr, fmt.Sprintf("size=%d,", originalSize/units.GiB))
				} else {
					paramsArr = append(paramsArr, param)
				}
				//drive spec generated from actual cloudrive spec

			}
			newSpec := strings.Join(paramsArr, ",")

			err = Inst().V.AddCloudDrive(nodeSelected, newSpec, -1)
			log.FailOnError(err, fmt.Sprintf("Add cloud drive failed on node %s", nodeSelected.Name))
			log.InfoD("adding new pool was successful")
		})

	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		appsValidateAndDestroy(contexts)
		AfterEachTest(contexts)
	})
}

// Verify drive getting pulled when the node is booting up
var _ = Describe("{RebootNodeAndYankDriver}", Label("p1", "hal_ops_disruption", "node_reboot", "drive_failure", "YankPoolDrive", "staging"), func() {

	/*
		Jira-Link : https://purestorage.atlassian.net/browse/HAZEL-1002
		Reboot a node
		Yank a drive
		Once node comes up, validate apps
		Reboot that same node again
		Kill Px on the same node as step 4 and wait for it to come up
		Yank another drive on another pool
		Kill Px on the node from Step 6
		Validate apps
	*/

	JustBeforeEach(func() {
		StartTorpedoTest("RebootNodeAndYankDriver", "Verify drive getting pulled when the node is booting up", nil, 0)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})

	var (
		contexts         []*scheduler.Context
		nodeSelected     *node.Node
		busID, poolDrive string
	)

	itLog := "Verify drive getting pulled when the node is booting up"
	It(itLog, func() {
		log.InfoD(itLog)

		stepLog := "Scheduling apps to perform IOs"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				contexts = append(contexts, ScheduleApplications(fmt.Sprintf("yankdriver-%d", i))...)
			}
			ValidateApplications(contexts)
		})
		defer appsValidateAndDestroy(contexts)

		ctx := contexts[0]
		volumes, err := Inst().S.GetVolumes(ctx)
		log.FailOnError(err, "Failed while listing the volume with error")
		log.InfoD("Volume details %v", volumes)

		if len(volumes) == 0 || len(volumes) < 2 {
			msg := fmt.Sprintf("There are no enough volumes associated with the app %v, required volumes 2", ctx.App.Key)
			log.InfoD(msg)
			Skip(msg)
		}

		selectReplicaNodeAndPoolDrive := func(selectedVol *volume.Volume) {
			stepLog = "Select the volume replica node and pool drive"
			Step(stepLog, func() {
				log.InfoD(stepLog)
				rsDetails, err := Inst().V.GetReplicaSets(selectedVol)
				log.FailOnError(err, fmt.Sprintf("error getting replica sets for vol %s", selectedVol.Name))
				log.InfoD("Volume Replica info %v", rsDetails)

				poolUUIDs := rsDetails[0].GetPoolUuids()
				poolUUIDSelected := poolUUIDs[rand.Intn(len(poolUUIDs))]
				poolIDSelected, err := GetPoolIDFromPoolUUID(poolUUIDSelected)
				log.FailOnError(err, fmt.Sprintf("error getting pool id for the pool %s", poolUUIDSelected))

				nodeSelected, err = GetNodeWithGivenPoolID(poolUUIDSelected)
				log.FailOnError(err, fmt.Sprintf("error getting node detail for the pool %s", poolUUIDSelected))

				jDev, err := Inst().V.GetJournalDevicePath(nodeSelected)
				log.FailOnError(err, fmt.Sprintf("error getting journal device path from node %s", nodeSelected.Name))
				log.InfoD("Journal device path - %s", jDev)

				cmd := fmt.Sprintf("lsblk -no pkname %s", jDev)
				journalParentDevPath, err := Inst().N.RunCommandWithNoRetry(*nodeSelected, cmd, node.ConnectionOpts{
					Timeout:         2 * time.Minute,
					TimeBeforeRetry: 10 * time.Second,
				})
				log.FailOnError(err, "error occured running the command to identify the parent device path of the journal partition %s", jDev)
				journalParentDevPath = strings.TrimRight(journalParentDevPath, "\n")
				log.InfoD("Parent device path of the journal device is %s", journalParentDevPath)

				driveMap, err := Inst().V.GetPoolDrives(nodeSelected)
				log.FailOnError(err, fmt.Sprintf("error getting pool drive details for the node %s", nodeSelected))

				drives := driveMap[strconv.Itoa(int(poolIDSelected))]

				for _, drive := range drives {
					cmd := fmt.Sprintf("lsblk -no pkname %s", drive.Device)
					poolDriveParentPath, err := Inst().N.RunCommandWithNoRetry(*nodeSelected, cmd, node.ConnectionOpts{
						Timeout:         2 * time.Minute,
						TimeBeforeRetry: 10 * time.Second,
					})
					log.FailOnError(err, "error occured running the command to identify the parent device path of the drive %s", drive.Device)
					poolDriveParentPath = strings.TrimRight(poolDriveParentPath, "\n")
					log.InfoD("Parent device path of the pool device is %s", poolDriveParentPath)

					if (journalParentDevPath != "") && (poolDriveParentPath != "") && (journalParentDevPath == poolDriveParentPath) {
						continue
					}

					poolDrive = drive.Device
					log.InfoD("Pool drive selected - %s", poolDrive)
					break
				}
			})
		}

		rebootNode := func(nodeToRestart node.Node) {
			stepLog = "Reboot the node"
			Step(stepLog, func() {
				log.Info(stepLog)
				err = RebootNodeAndWaitForPxUp(nodeToRestart)
				log.FailOnError(err, "Failed to reboot node and wait till it is up")
			})
		}

		pxRestart := func(nodeToRestatPx node.Node) {
			stepLog = "Restart portworx and wait for it to come up"
			Step(stepLog, func() {
				log.Info(stepLog)
				Step(fmt.Sprintf("node with Px restart is: %s", nodeToRestatPx.Name), func() {
					err := Inst().V.RestartDriver(*nodeSelected, nil)
					log.FailOnError(err, fmt.Sprintf("Error occured while Restart PX on node:%v", nodeSelected.Name))
				})

				stepLog = fmt.Sprintf("Wait for driver to up")
				Step(stepLog, func() {
					log.InfoD(stepLog)
					err = Inst().V.WaitDriverUpOnNode(nodeToRestatPx, Inst().DriverStartTimeout)
					log.FailOnError(err, fmt.Sprintf("Driver is down on node %s", nodeSelected.Name))
					log.Infof("Driver is up")
				})

				Step(fmt.Sprintf("verify portworx after restart: %v", nodeToRestatPx.Name), func() {
					status, err := Inst().V.GetPxctlStatus(nodeToRestatPx)
					log.FailOnError(err, fmt.Sprintf("failed to get pxctl status on node [%s]", nodeToRestatPx.Name))
					log.InfoD("px status [%v]", status)
					dash.VerifyFatal(status == api.Status_STATUS_OK.String(), true, fmt.Sprintf("node [%s] status is up but PX cluster is not ok. Expected: %v Actual: %v",
						nodeToRestatPx.Name, api.Status_STATUS_OK, status))
				})
			})
		}

		yankDriver := func(nodeToYank node.Node, poolDrive string) {

			poolDrive = strings.Trim(poolDrive, "/")
			poolDriveArr := strings.Split(poolDrive, "/")
			poolDrive = poolDriveArr[len(poolDriveArr)-1]

			stepLog = "Yank journal drive"
			Step(stepLog, func() {
				log.InfoD(stepLog)
				busID, err = Inst().N.YankDrive(nodeToYank, poolDrive, node.ConnectionOpts{
					Timeout:         dfDefaultTimeout,
					TimeBeforeRetry: dfDefaultRetryInterval,
				})
				log.FailOnError(err, fmt.Sprintf("failed to yank journal drive on node [%s]", nodeToYank.Name))
				log.InfoD("Bus id - %s", busID)
			})

			stepLog = "Recover yank drive"
			Step(stepLog, func() {
				log.InfoD(stepLog)
				err = Inst().N.RecoverDrive(nodeToYank, poolDrive, busID, node.ConnectionOpts{
					Timeout:         driveFailTimeout,
					TimeBeforeRetry: dfDefaultRetryInterval,
				})
				log.FailOnError(err, fmt.Sprintf("failed to recover yank journal drive on node [%s]", nodeToYank.Name))
				log.InfoD("Verified recover yank drive")
			})
		}

		selectReplicaNodeAndPoolDrive(volumes[0])
		rebootNode(*nodeSelected)                 // Reboot a node
		yankDriver(*nodeSelected, poolDrive)      // Yank a drive
		ValidateApplications(contexts)            // Once node comes up, validate apps
		rebootNode(*nodeSelected)                 // Reboot that same node again
		pxRestart(*nodeSelected)                  // Kill Px on same node as step 4 and wait for it to come up
		selectReplicaNodeAndPoolDrive(volumes[1]) // Select another node and pool
		yankDriver(*nodeSelected, poolDrive)      // Yank another drive on another pool
		pxRestart(*nodeSelected)                  // Kill Px on the node from Step 6
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts)
	})
})

func getDeletablePoolMap() map[string]struct{} {
	var jrnlPartPoolID string
	deletablePoolMap := map[string]struct{}{}

	poolDeletableNodes := selectPoolDeletableNodes()
	for _, node := range poolDeletableNodes {
		nodePools := node.StoragePools
		log.Infof("selected node [%s]", node.Name)
		stepLog = "Selecting pool to delete"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			isjournal, err := IsJournalEnabled()
			log.FailOnError(err, "Failed to check if Journal enabled")

			if isjournal && len(nodePools) > 1 {
				jDev, err := Inst().V.GetJournalDevicePath(&node)
				log.FailOnError(err, fmt.Sprintf("error getting journal device path from node %s", node.Name))
				log.Infof("JournalDev: %s", jDev)
				if jDev == "" {
					log.FailOnError(fmt.Errorf("no journal device path found"), "error getting journal device path from storage spec")
				}
				drivesMap, err := Inst().V.GetPoolDrives(&node)
				jPath := jDev[:len(jDev)-1]
			outer:
				for k, v := range drivesMap {
					for _, dv := range v {
						if strings.Contains(dv.Device, jPath) {
							jrnlPartPoolID = k
							break outer
						}
					}
				}
				for _, nodePool := range nodePools {
					if strconv.Itoa(int(nodePool.ID)) != jrnlPartPoolID {
						deletablePoolMap[nodePool.Uuid] = struct{}{}
					}
				}
			} else {
				for _, nodePool := range nodePools {
					deletablePoolMap[nodePool.Uuid] = struct{}{}
				}
			}
		})
	}

	return deletablePoolMap
}
