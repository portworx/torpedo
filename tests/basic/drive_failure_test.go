package tests

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/libopenstorage/openstorage/api"
	"github.com/pure-px/torpedo/pkg/log"

	"github.com/pure-px/torpedo/pkg/testrailuttils"

	. "github.com/onsi/ginkgo/v2"

	"github.com/pure-px/torpedo/drivers/node"
	"github.com/pure-px/torpedo/drivers/scheduler"
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

var _ = Describe("{YankJournalWithPxRestart}", func() {
	testName = "yank-journal-px-restart"
	testDescription = "Yank journal drive and restart PX"
	YankJournalTest(testName, testDescription)
})

var _ = Describe("{YankJournalWithNodeReboot}", func() {
	testName = "yank-journal-node-reboot"
	testDescription = "Yank journal drive and reboot node"
	YankJournalTest(testName, testDescription)
})

var _ = Describe("{YankJournalWithNodeMaintenanceCycle}", func() {
	testName := "yank-journal-node-maintenance-cycle"
	testDescription = "Yank journal drive and do node maintenance"
	YankJournalTest(testName, testDescription)
})

var _ = Describe("{YankJournalWithIOsRunning}", func() {
	testName := "yank-journal-io"
	testDescription = "Yank journal drive with IOs running"
	YankJournalTest(testName, testDescription)
})

func YankJournalTest(testName, testDesc string) {
	var nodeSelected node.Node
	var busID string

	JustBeforeEach(func() {
		StartTorpedoTest(testName, testDesc, nil, 0)
	})

	itLog := testDesc
	It(itLog, func() {
		log.InfoD(itLog)

		stepLog := "Schedule apps to perform IOs"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			contexts = make([]*scheduler.Context, 0)
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

var _ = Describe("{YankPoolDriveWithPxRestart}", func() {
	testName = "yank-pool-drive-px-restart"
	testDescription = "Yank pool drive and restart px"
	YankPoolDriveTest(testName, testDescription)
})

var _ = Describe("{YankPoolDriveWithNodeReboot}", func() {
	testName = "yank-pool-drive-node-reboot"
	testDescription = "Yank pool drive and reboot node"
	YankPoolDriveTest(testName, testDescription)
})

var _ = Describe("{YankPoolDriveWithIOs}", func() {
	testName = "yank-pool-drive-IO"
	testDescription = "Yank pool drive with IOs running"
	YankPoolDriveTest(testName, testDescription)
})

var _ = Describe("{YankMetadataWithPxRestart}", func() {
	testName = "YankMetadataWithPxRestart"
	testDescription = "Yank metadata drive with restart PX"
	YankMetadataTest(testName, testDescription)
})
var _ = Describe("{YankMetadataWithNodeReboot}", func() {
	testName = "YankMetadataWithNodeReboot"
	testDescription = "Yank metadata drive and reboot node"
	YankMetadataTest(testName, testDescription)
})
var _ = Describe("{YankMetadataWithNodeMaintenanceCycle}", func() {
	testName = "YankMetadataWithNodeMaintenanceCycle"
	testDescription = "Yank metadata drive and node maintenance cycle"
	YankMetadataTest(testName, testDescription)
})

func YankMetadataTest(testName, testDesc string) {
	var (
		nodeSelected node.Node
		busID, path  string
		kvdbNodesIDs []string
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
			contexts = make([]*scheduler.Context, 0)
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

func YankPoolDriveTest(testName, testDesc string) {
	var nodeSelected *node.Node
	var busID, poolDrive string

	JustBeforeEach(func() {
		StartTorpedoTest(testName, testDesc, nil, 0)
	})

	itLog := testDesc
	It(itLog, func() {
		log.InfoD(itLog)

		stepLog := "Schedule apps to perform IOs"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			contexts = make([]*scheduler.Context, 0)
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
