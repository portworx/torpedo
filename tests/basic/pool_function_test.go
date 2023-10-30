package tests

import (
	"fmt"
	"strings"
	"time"

	"regexp"

	"github.com/google/uuid"
	"github.com/libopenstorage/openstorage/api"
	. "github.com/onsi/ginkgo"
	"github.com/portworx/torpedo/drivers/node"
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/drivers/volume"
	"github.com/portworx/torpedo/pkg/log"
	"github.com/portworx/torpedo/pkg/testrailuttils"
	"github.com/portworx/torpedo/pkg/units"
	. "github.com/portworx/torpedo/tests"
)

var (
	stepLog       string
	runID         int
	testrailID    int
	targetSizeGiB uint64
	storageNode   *node.Node
	err           error
)
var _ = Describe("{PoolExpandMultipleTimes}", func() {
	BeforeEach(func() {
		contexts = scheduleApps()
	})

	JustBeforeEach(func() {
		poolIDToResize = pickPoolToResize()
		log.Infof("Picked pool %s to resize", poolIDToResize)
		poolToBeResized = getStoragePool(poolIDToResize)
	})

	JustAfterEach(func() {
		AfterEachTest(contexts)
	})

	AfterEach(func() {
		appsValidateAndDestroy(contexts)
		EndTorpedoTest()
	})

	It("Select a pool and expand it by 100 GiB 3 time with add-disk type. ", func() {
		StartTorpedoTest("PoolExpandDiskAdd3Times",
			"Validate storage pool expansion 3 times with type=add-disk", nil, 0)
		for i := 0; i < 3; i++ {
			poolToBeResized = getStoragePool(poolIDToResize)
			originalSizeInBytes = poolToBeResized.TotalSize
			targetSizeInBytes = originalSizeInBytes + 100*units.GiB
			targetSizeGiB = targetSizeInBytes / units.GiB

			log.InfoD("Current Size of pool %s is %d GiB. Expand to %v GiB with type add-disk...",
				poolIDToResize, poolToBeResized.TotalSize/units.GiB, targetSizeGiB)
			triggerPoolExpansion(poolIDToResize, targetSizeGiB, api.SdkStoragePool_RESIZE_TYPE_ADD_DISK)
			resizeErr := waitForOngoingPoolExpansionToComplete(poolIDToResize)
			dash.VerifyFatal(resizeErr, nil, "Pool expansion does not result in error")
			verifyPoolSizeEqualOrLargerThanExpected(poolIDToResize, targetSizeGiB)
		}
	})

	It("Select a pool and expand it by 100 GiB 3 times with resize-disk type. ", func() {
		StartTorpedoTest("PoolExpandDiskResize3Times",
			"Validate storage pool expansion with type=resize-disk", nil, 0)
		for i := 0; i < 3; i++ {
			originalSizeInBytes = poolToBeResized.TotalSize
			targetSizeInBytes = originalSizeInBytes + 100*units.GiB
			targetSizeGiB = targetSizeInBytes / units.GiB

			log.InfoD("Current Size of pool %s is %d GiB. Expand to %v GiB with type resize-disk...",
				poolIDToResize, poolToBeResized.TotalSize/units.GiB, targetSizeGiB)
			triggerPoolExpansion(poolIDToResize, targetSizeGiB, api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK)
			resizeErr := waitForOngoingPoolExpansionToComplete(poolIDToResize)
			dash.VerifyFatal(resizeErr, nil, "Pool expansion does not result in error")
			verifyPoolSizeEqualOrLargerThanExpected(poolIDToResize, targetSizeGiB)
		}
	})
})

var _ = Describe("{PoolExpandSmoky}", func() {
	BeforeEach(func() {
		contexts = scheduleApps()
	})

	JustBeforeEach(func() {
		poolIDToResize = pickPoolToResize()
		log.Infof("Picked pool %s to resize", poolIDToResize)
		poolToBeResized = getStoragePool(poolIDToResize)
	})

	JustAfterEach(func() {
		AfterEachTest(contexts)
	})

	AfterEach(func() {
		appsValidateAndDestroy(contexts)
		EndTorpedoTest()
	})

	It("Select a pool and expand it by 100 GiB with add-disk type. ", func() {
		StartTorpedoTest("PoolExpandDiskAdd",
			"Validate storage pool expansion with type=add-disk", nil, 0)
		originalSizeInBytes = poolToBeResized.TotalSize
		targetSizeInBytes = originalSizeInBytes + 100*units.GiB
		targetSizeGiB = targetSizeInBytes / units.GiB

		log.InfoD("Current Size of the pool %s is %d GiB. Trying to expand to %v GiB with type add-disk",
			poolIDToResize, poolToBeResized.TotalSize/units.GiB, targetSizeGiB)
		triggerPoolExpansion(poolIDToResize, targetSizeGiB, api.SdkStoragePool_RESIZE_TYPE_ADD_DISK)
		resizeErr := waitForOngoingPoolExpansionToComplete(poolIDToResize)
		dash.VerifyFatal(resizeErr, nil, "Pool expansion does not result in error")
		verifyPoolSizeEqualOrLargerThanExpected(poolIDToResize, targetSizeGiB)
	})

	It("Select a pool and expand it by 100 GiB with resize-disk type. ", func() {
		StartTorpedoTest("PoolExpandDiskResize",
			"Validate storage pool expansion with type=resize-disk", nil, 0)
		originalSizeInBytes = poolToBeResized.TotalSize
		targetSizeInBytes = originalSizeInBytes + 100*units.GiB
		targetSizeGiB = targetSizeInBytes / units.GiB

		log.InfoD("Current Size of the pool %s is %d GiB. Trying to expand to %v GiB with type resize-disk",
			poolIDToResize, poolToBeResized.TotalSize/units.GiB, targetSizeGiB)
		triggerPoolExpansion(poolIDToResize, targetSizeGiB, api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK)
		resizeErr := waitForOngoingPoolExpansionToComplete(poolIDToResize)
		dash.VerifyFatal(resizeErr, nil, "Pool expansion does not result in error")
		verifyPoolSizeEqualOrLargerThanExpected(poolIDToResize, targetSizeGiB)
	})

	It("Select a pool and expand it by 100 GiB with auto type. ", func() {
		StartTorpedoTest("PoolExpandDiskAuto",
			"Validate storage pool expansion with type=auto ", nil, 0)
		originalSizeInBytes = poolToBeResized.TotalSize
		targetSizeInBytes = originalSizeInBytes + 100*units.GiB
		targetSizeGiB = targetSizeInBytes / units.GiB

		log.InfoD("Current Size of the pool %s is %d GiB. Trying to expand to %v GiB with type auto",
			poolIDToResize, poolToBeResized.TotalSize/units.GiB, targetSizeGiB)
		triggerPoolExpansion(poolIDToResize, targetSizeGiB, api.SdkStoragePool_RESIZE_TYPE_AUTO)
		resizeErr := waitForOngoingPoolExpansionToComplete(poolIDToResize)
		dash.VerifyFatal(resizeErr, nil, "Pool expansion does not result in error")
		verifyPoolSizeEqualOrLargerThanExpected(poolIDToResize, targetSizeGiB)
	})
})

var _ = Describe("{PoolExpandWithReboot}", func() {
	BeforeEach(func() {
		contexts = scheduleApps()
	})

	JustBeforeEach(func() {
		poolIDToResize = pickPoolToResize()
		log.Infof("Picked pool %s to resize", poolIDToResize)
		poolToBeResized = getStoragePool(poolIDToResize)
		storageNode, err = GetNodeWithGivenPoolID(poolIDToResize)
		log.FailOnError(err, "Failed to get node with given pool ID")
	})

	JustAfterEach(func() {
		AfterEachTest(contexts)
	})

	AfterEach(func() {
		appsValidateAndDestroy(contexts)
		EndTorpedoTest()
	})

	It("Initiate pool expansion using add-disk and reboot node", func() {
		StartTorpedoTest("PoolExpandDiskAddWithReboot", "Initiate pool expansion using add-disk and reboot node", nil, 51309)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
		Step("Select a pool that has I/O and expand it by 100 GiB with add-disk type. ", func() {
			originalSizeInBytes = poolToBeResized.TotalSize
			targetSizeInBytes = originalSizeInBytes + 100*units.GiB
			targetSizeGiB = targetSizeInBytes / units.GiB
			log.InfoD("Current Size of the pool %s is %d GiB. Trying to expand to %v GiB with type add-disk",
				poolIDToResize, poolToBeResized.TotalSize/units.GiB, targetSizeGiB)
			triggerPoolExpansion(poolIDToResize, targetSizeGiB, api.SdkStoragePool_RESIZE_TYPE_ADD_DISK)
		})

		Step("Wait for expansion to start and reboot node", func() {
			err := WaitForExpansionToStart(poolIDToResize)
			log.FailOnError(err, "Timed out waiting for expansion to start")
			err = RebootNodeAndWait(*storageNode)
			log.FailOnError(err, "Failed to reboot node and wait till it is up")
		})

		Step("Ensure pool has been expanded to the expected size", func() {
			err = waitForOngoingPoolExpansionToComplete(poolIDToResize)
			dash.VerifyFatal(err, nil, "Pool expansion does not result in error")
			verifyPoolSizeEqualOrLargerThanExpected(poolIDToResize, targetSizeGiB)
		})
	})
})

var _ = Describe("{PoolExpandWithPXRestart}", func() {
	BeforeEach(func() {
		contexts = scheduleApps()
	})

	JustBeforeEach(func() {
		poolIDToResize = pickPoolToResize()
		log.Infof("Picked pool %s to resize", poolIDToResize)
		poolToBeResized = getStoragePool(poolIDToResize)
		storageNode, err = GetNodeWithGivenPoolID(poolIDToResize)
		log.FailOnError(err, "Failed to get node with given pool ID")
	})

	JustAfterEach(func() {
		AfterEachTest(contexts)
	})

	AfterEach(func() {
		appsValidateAndDestroy(contexts)
		EndTorpedoTest()
	})

	It("Initiate pool expansion using add-drive and restart PX", func() {
		StartTorpedoTest("PoolExpandAddDiskAndPXRestart",
			"Initiate pool expansion using add-drive and restart PX", nil, testrailID)

		Step("Select a pool that has I/O and expand it by 100 GiB with add-disk type. ", func() {
			originalSizeInBytes = poolToBeResized.TotalSize
			targetSizeInBytes = originalSizeInBytes + 100*units.GiB
			targetSizeGiB = targetSizeInBytes / units.GiB
			log.InfoD("Current Size of the pool %s is %d GiB. Trying to expand to %v GiB with type add-disk",
				poolIDToResize, poolToBeResized.TotalSize/units.GiB, targetSizeGiB)
			triggerPoolExpansion(poolIDToResize, targetSizeGiB, api.SdkStoragePool_RESIZE_TYPE_ADD_DISK)
		})

		Step("Wait for expansion to start and reboot node", func() {
			err := WaitForExpansionToStart(poolIDToResize)
			log.FailOnError(err, "Timed out waiting for expansion to start")
			err = Inst().V.RestartDriver(*storageNode, nil)
			log.FailOnError(err, fmt.Sprintf("Error restarting px on node [%s]", storageNode.Name))
			err = Inst().V.WaitDriverUpOnNode(*storageNode, addDriveUpTimeOut)
			log.FailOnError(err, fmt.Sprintf("Timed out waiting for px to come up on node [%s]", storageNode.Name))
		})

		Step("Ensure pool has been expanded to the expected size", func() {
			resizeErr := waitForOngoingPoolExpansionToComplete(poolIDToResize)
			dash.VerifyFatal(resizeErr, nil, "Pool expansion does not result in error")
			verifyPoolSizeEqualOrLargerThanExpected(poolIDToResize, targetSizeGiB)
		})
	})
})

var _ = Describe("{PoolVolUpdateResizeDisk}", func() {
	//1) Deploy px with cloud drive.
	//2) Create a volume on that pool and write some data on the volume.
	//3) expand the volume to the pool
	//4) perform resize disk operation on the pool while volume update is in-progress
	JustBeforeEach(func() {
		StartTorpedoTest("PoolVolUpdateResizeDisk", "expand volume to the pool and pool expansion using resize-disk", nil, 0)
	})
	var contexts []*scheduler.Context

	stepLog := "should get the existing storage node and expand the pool by resize-disk"

	It(stepLog, func() {
		log.InfoD(stepLog)
		contexts = make([]*scheduler.Context, 0)
		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("volupdtplrsz-%d", i))...)
		}
		ValidateApplications(contexts)
		defer appsValidateAndDestroy(contexts)

		stNodes := node.GetStorageNodes()
		if len(stNodes) == 0 {
			dash.VerifyFatal(len(stNodes) > 0, true, "Storage nodes found?")
		}
		volSelected, err := GetVolumeWithMinimumSize(contexts, 10)
		log.FailOnError(err, "error identifying volume")
		appVol, err := Inst().V.InspectVolume(volSelected.ID)
		log.FailOnError(err, fmt.Sprintf("err inspecting vol : %s", volSelected.ID))
		volNodes := appVol.ReplicaSets[0].Nodes
		var stNode node.Node
		for _, n := range stNodes {
			nodeExist := false
			for _, vn := range volNodes {
				if n.Id == vn {
					nodeExist = true
				}
			}
			if !nodeExist {
				stNode = n
				break
			}
		}
		selectedPool := stNode.Pools[0]
		var poolToBeResized *api.StoragePool
		poolToBeResized, err = GetStoragePoolByUUID(selectedPool.Uuid)
		log.FailOnError(err, fmt.Sprintf("Failed to get pool using UUID %s", selectedPool.Uuid))

		stepLog = "Expand volume to the expanded pool"
		var newRep int64
		opts := volume.Options{
			ValidateReplicationUpdateTimeout: replicationUpdateTimeout,
		}
		var currRep int64
		Step(stepLog, func() {
			log.InfoD(stepLog)
			currRep, err = Inst().V.GetReplicationFactor(volSelected)
			log.FailOnError(err, fmt.Sprintf("err getting repl factor for  vol : %s", volSelected.Name))

			newRep = currRep
			if currRep == 3 {
				newRep = currRep - 1
				err = Inst().V.SetReplicationFactor(volSelected, newRep, nil, nil, true, opts)
				log.FailOnError(err, fmt.Sprintf("err setting repl factor  to %d for  vol : %s", newRep, volSelected.Name))
			}
			log.InfoD(fmt.Sprintf("setting repl factor to %d for vol : %s", newRep+1, volSelected.Name))
			err = Inst().V.SetReplicationFactor(volSelected, newRep+1, []string{stNode.Id}, []string{poolToBeResized.Uuid}, false, opts)
			log.FailOnError(err, fmt.Sprintf("err setting repl factor  to %d for  vol : %s", newRep+1, volSelected.Name))
			dash.VerifyFatal(err == nil, true, fmt.Sprintf("vol %s expansion triggered successfully on node %s", volSelected.Name, stNode.Name))
		})
		isjournal, err := IsJournalEnabled()
		log.FailOnError(err, "Failed to check if Journal enabled")

		stepLog := "Initiate pool expansion using resize-disk while repl increase is in progress"
		Step(stepLog, func() {
			log.InfoD(stepLog)

			drvSize, err := getPoolDiskSize(poolToBeResized)
			log.FailOnError(err, "error getting drive size for pool [%s]", poolToBeResized.Uuid)
			expectedSize := (poolToBeResized.TotalSize / units.GiB) + drvSize

			log.InfoD("Current Size of the pool %s is %d", poolToBeResized.Uuid, poolToBeResized.TotalSize/units.GiB)
			err = Inst().V.ExpandPool(poolToBeResized.Uuid, api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK, expectedSize, false)
			if err != nil {
				if strings.Contains(fmt.Sprintf("%v", err), "Please re-issue expand with force") {
					err = Inst().V.ExpandPool(poolToBeResized.Uuid, api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK, expectedSize, true)
				}
			}
			dash.VerifyFatal(err, nil, "Pool expansion init successful?")

			resizeErr := waitForPoolToBeResized(expectedSize, selectedPool.Uuid, isjournal)
			dash.VerifyFatal(resizeErr, nil, fmt.Sprintf("Verify pool %s on node %s expansion using resize-disk", selectedPool.Uuid, stNode.Name))

		})
		err = ValidateReplFactorUpdate(volSelected, newRep+1)
		log.FailOnError(err, "error validating repl factor for vol [%s]", volSelected.Name)

		stepLog = "Initiate pool expansion using resize-disk after rsync is successfull"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			poolToBeResized, err = GetStoragePoolByUUID(selectedPool.Uuid)
			log.FailOnError(err, fmt.Sprintf("error getting pool using UUID [%s]", selectedPool.Uuid))

			drvSize, err := getPoolDiskSize(poolToBeResized)
			log.FailOnError(err, "error getting drive size for pool [%s]", poolToBeResized.Uuid)
			expectedSize := (poolToBeResized.TotalSize / units.GiB) + drvSize

			log.InfoD("Current Size of the pool %s is %d", selectedPool.Uuid, poolToBeResized.TotalSize/units.GiB)
			err = Inst().V.ExpandPool(selectedPool.Uuid, api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK, expectedSize, false)
			if err != nil {
				if strings.Contains(fmt.Sprintf("%v", err), "Please re-issue expand with force") {
					err = Inst().V.ExpandPool(selectedPool.Uuid, api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK, expectedSize, true)
				}
			}
			dash.VerifyFatal(err, nil, "Pool expansion init successful?")

			resizeErr := waitForPoolToBeResized(expectedSize, selectedPool.Uuid, isjournal)
			dash.VerifyFatal(resizeErr, nil, fmt.Sprintf("Verify pool %s on node %s expansion using resize-disk", selectedPool.Uuid, stNode.Name))
		})

		//reverting the replication for volume validation
		if currRep < 3 {
			err = Inst().V.SetReplicationFactor(volSelected, currRep, nil, nil, true, opts)
			log.FailOnError(err, fmt.Sprintf("err setting repl factor to %d for vol : %s", newRep, volSelected.Name))
		}

	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})
})

var _ = Describe("{PoolExpandAndCheckAlerts}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("PoolExpandAndCheckAlerts", "pool expansion using resize-disk and add-disk and check alerts after each operation", nil, 0)
	})
	var contexts []*scheduler.Context

	stepLog := "should get the existing storage node and expand the pool by resize-disk"

	It(stepLog, func() {
		log.InfoD(stepLog)
		contexts = make([]*scheduler.Context, 0)
		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("plrszvolupdt-%d", i))...)
		}
		ValidateApplications(contexts)
		defer appsValidateAndDestroy(contexts)

		stNodes := node.GetStorageNodes()
		if len(stNodes) == 0 {
			dash.VerifyFatal(len(stNodes) > 0, true, "Storage nodes found?")
		}
		volSelected, err := GetVolumeWithMinimumSize(contexts, 10)
		log.FailOnError(err, "error identifying volume")
		appVol, err := Inst().V.InspectVolume(volSelected.ID)
		log.FailOnError(err, fmt.Sprintf("err inspecting vol : %s", volSelected.ID))
		volNodes := appVol.ReplicaSets[0].Nodes
		var stNode node.Node
		for _, n := range stNodes {
			nodeExist := false
			for _, vn := range volNodes {
				if n.Id == vn {
					nodeExist = true
				}
			}
			if !nodeExist {
				stNode = n
				break
			}
		}
		selectedPool := stNode.Pools[0]
		var poolToBeResized *api.StoragePool

		stepLog := "Initiate pool expansion using resize-disk"
		Step(stepLog, func() {
			log.InfoD(stepLog)

			poolToBeResized, err = GetStoragePoolByUUID(selectedPool.Uuid)
			log.FailOnError(err, fmt.Sprintf("Failed to get pool using UUID %s", selectedPool.Uuid))

			drvSize, err := getPoolDiskSize(poolToBeResized)
			log.FailOnError(err, "error getting drive size for pool [%s]", poolToBeResized.Uuid)
			expectedSize := (poolToBeResized.TotalSize / units.GiB) + drvSize

			isjournal, err := IsJournalEnabled()
			log.FailOnError(err, "Failed to check if Journal enabled")
			expectedSizeWithJournal := expectedSize
			if isjournal {
				expectedSizeWithJournal = expectedSizeWithJournal - 3
			}
			log.InfoD("Current Size of the pool %s is %d", selectedPool.Uuid, poolToBeResized.TotalSize/units.GiB)
			err = Inst().V.ExpandPool(selectedPool.Uuid, api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK, expectedSize, false)
			dash.VerifyFatal(err, nil, "Pool expansion init successful?")

			resizeErr := waitForPoolToBeResized(expectedSize, selectedPool.Uuid, isjournal)
			dash.VerifyFatal(resizeErr, nil, fmt.Sprintf("Verify pool %s on node %s expansion using resize-disk", selectedPool.Uuid, stNode.Name))

			stepLog = "Ensure that new pool has been expanded to the expected size and also check the pool expand alert"
			Step(stepLog, func() {
				log.InfoD(stepLog)
				log.Infof("Check the alert for pool expand for pool uuid %s", poolIDToResize)
				// Get the node to check the pool show output
				n := node.GetStorageDriverNodes()[0]
				// Below command to change when PWX-28484 is fixed
				cmd := "pxctl alerts show| grep -e POOL"

				// Execute the command and check the alerts of type POOL
				out, err := Inst().N.RunCommandWithNoRetry(n, cmd, node.ConnectionOpts{
					Timeout:         2 * time.Minute,
					TimeBeforeRetry: 10 * time.Second,
				})

				log.FailOnError(err, "Unable to execute the alerts show command")

				outLines := strings.Split(out, "\n")
				var alertExist bool
				alertExist = false
				for _, l := range outLines {
					line := strings.Trim(l, " ")
					if strings.Contains(line, "PoolExpandSuccessful") && strings.Contains(line, poolIDToResize) {
						if strings.Contains(line, fmt.Sprintf("%d", expectedSize)) || strings.Contains(line, fmt.Sprintf("%d", expectedSizeWithJournal)) {
							alertExist = true
							log.Infof("The Alert generated is %s", line)
							break
						}
					}
				}
				dash.VerifyFatal(alertExist, true, "Verify Alert is Present")
			})
		})

		stepLog = "Initiate pool expansion using add-disk"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			poolToBeResized, err = GetStoragePoolByUUID(selectedPool.Uuid)
			log.FailOnError(err, fmt.Sprintf("Failed to get pool using UUID %s", selectedPool.Uuid))

			drvSize, err := getPoolDiskSize(poolToBeResized)
			log.FailOnError(err, "error getting drive size for pool [%s]", poolToBeResized.Uuid)
			expectedSize := (poolToBeResized.TotalSize / units.GiB) + drvSize

			isjournal, err := IsJournalEnabled()
			log.FailOnError(err, "Failed to check if Journal enabled")
			expectedSizeWithJournal := expectedSize
			if isjournal {
				expectedSizeWithJournal = expectedSizeWithJournal - 3
			}
			log.InfoD("Current Size of the pool %s is %d", selectedPool.Uuid, poolToBeResized.TotalSize/units.GiB)
			err = Inst().V.ExpandPool(selectedPool.Uuid, api.SdkStoragePool_RESIZE_TYPE_ADD_DISK, expectedSize, false)
			dash.VerifyFatal(err, nil, "Pool expansion init successful?")

			resizeErr := waitForPoolToBeResized(expectedSize, selectedPool.Uuid, isjournal)
			dash.VerifyFatal(resizeErr, nil, fmt.Sprintf("Verify pool %s on node %s expansion using add-disk", selectedPool.Uuid, stNode.Name))

			stepLog = "Ensure that new pool has been expanded to the expected size and also check the pool expand alert"
			Step(stepLog, func() {
				log.InfoD(stepLog)
				log.Infof("Check the alert for pool expand for pool uuid %s", poolIDToResize)
				// Get the node to check the pool show output
				n := node.GetStorageDriverNodes()[0]
				// Below command to change when PWX-28484 is fixed
				cmd := "pxctl alerts show| grep -e POOL"

				// Execute the command and check the alerts of type POOL
				out, err := Inst().N.RunCommandWithNoRetry(n, cmd, node.ConnectionOpts{
					Timeout:         2 * time.Minute,
					TimeBeforeRetry: 10 * time.Second,
				})

				log.FailOnError(err, "Unable to execute the alerts show command")

				outLines := strings.Split(out, "\n")
				var alertExist bool
				alertExist = false
				for _, l := range outLines {
					line := strings.Trim(l, " ")
					if strings.Contains(line, "PoolExpandSuccessful") && strings.Contains(line, poolIDToResize) {
						if strings.Contains(line, fmt.Sprintf("%d", expectedSize)) || strings.Contains(line, fmt.Sprintf("%d", expectedSizeWithJournal)) {
							alertExist = true
							log.Infof("The Alert generated is %s", line)
							break
						}
					}
				}
				dash.VerifyFatal(alertExist, true, "Verify Alert is Present")
			})
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})
})

var _ = Describe("{PoolExpandResizeInvalidPoolID}", func() {

	var testrailID = 34542946
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/tests/view/34542946

	BeforeEach(func() {
		StartTorpedoTest("PoolExpandResizeInvalidPoolID",
			"Initiate pool expansion using invalid Id", nil, testrailID)
	})

	AfterEach(func() {
		EndTorpedoTest()
	})

	stepLog := "Resize with invalid pool ID"
	It(stepLog, func() {
		log.InfoD(stepLog)
		// invalidPoolUUID Generation
		invalidPoolUUID := uuid.New().String()

		// Resize Pool with Invalid Pool ID
		stepLog = fmt.Sprintf("Expanding pool on Node UUID [%s] using auto", invalidPoolUUID)
		Step(stepLog, func() {
			resizeErr := Inst().V.ExpandPool(invalidPoolUUID, api.SdkStoragePool_RESIZE_TYPE_AUTO, 100, true)
			dash.VerifyFatal(resizeErr != nil, true, "Verify error occurs with invalid Pool UUID")
			// Verify error on pool expansion failure
			var errMatch error
			re := regexp.MustCompile(fmt.Sprintf(".*failed to find storage pool with UID.*%s.*",
				invalidPoolUUID))
			if !re.MatchString(fmt.Sprintf("%v", resizeErr)) {
				errMatch = fmt.Errorf("failed to verify failure using invalid PoolUUID [%v]", invalidPoolUUID)
			}
			dash.VerifyFatal(errMatch, nil, "Pool expand with invalid PoolUUID failed as expected.")
		})
	})

})

var _ = Describe("{PoolExpandDiskAddAndVerifyFromOtherNode}", func() {

	var testrailID = 34542840
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/tests/view/34542840

	BeforeEach(func() {
		StartTorpedoTest("PoolExpandDiskAddAndVerifyFromOtherNode",
			"Initiate pool expansion and verify from other node", nil, testrailID)
		contexts = scheduleApps()
	})

	JustBeforeEach(func() {
		poolIDToResize = pickPoolToResize()
		log.Infof("Picked pool %s to resize", poolIDToResize)
		poolToBeResized = getStoragePool(poolIDToResize)
		storageNode, err = GetNodeWithGivenPoolID(poolIDToResize)
		log.FailOnError(err, "Failed to get node with given pool ID")
	})

	JustAfterEach(func() {
		AfterEachTest(contexts)
	})

	AfterEach(func() {
		appsValidateAndDestroy(contexts)
		EndTorpedoTest()
	})

	stepLog := "should get the existing pool and expand it by adding a disk and verify from other node"
	It(stepLog, func() {
		log.InfoD(stepLog)
		// get original total size
		provisionStatus, err := GetClusterProvisionStatusOnSpecificNode(*storageNode)
		var orignalTotalSize float64
		for _, pstatus := range provisionStatus {
			if pstatus.NodeUUID == storageNode.Id {
				orignalTotalSize += pstatus.TotalSize
			}
		}

		originalSizeInBytes = poolToBeResized.TotalSize
		targetSizeInBytes = originalSizeInBytes + 100*units.GiB
		targetSizeGiB = targetSizeInBytes / units.GiB

		log.InfoD("Current Size of the pool %s is %d GiB. Trying to expand to %v GiB with type add-disk",
			poolIDToResize, poolToBeResized.TotalSize/units.GiB, targetSizeGiB)
		triggerPoolExpansion(poolIDToResize, targetSizeGiB, api.SdkStoragePool_RESIZE_TYPE_ADD_DISK)

		Step("Ensure pool has been expanded to the expected size", func() {
			err = waitForOngoingPoolExpansionToComplete(poolIDToResize)
			dash.VerifyFatal(err, nil, "Pool expansion does not result in error")
			verifyPoolSizeEqualOrLargerThanExpected(poolIDToResize, targetSizeGiB)
		})

		stNodes, err := GetStorageNodes()
		log.FailOnError(err, "Unable to get the storage nodes")
		var verifyNode node.Node
		for _, node := range stNodes {
			status, _ := IsPxRunningOnNode(&node)
			if node.Id != storageNode.Id && status {
				verifyNode = node
				break
			}
		}

		// get final total size
		provisionStatus, err = GetClusterProvisionStatusOnSpecificNode(verifyNode)
		var finalTotalSize float64
		for _, pstatus := range provisionStatus {
			if pstatus.NodeUUID == storageNode.Id {
				finalTotalSize += pstatus.TotalSize
			}
		}
		dash.VerifyFatal(finalTotalSize > orignalTotalSize, true, "Pool expansion failed, pool size is not greater than pool size before expansion")

	})

})

var _ = Describe("{PoolExpansionDiskResizeInvalidSize}", func() {

	var testrailID = 34542945
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/tests/view/34542945

	BeforeEach(func() {
		StartTorpedoTest("PoolExpansionDiskResizeInvalidSize",
			"Initiate pool expansion using invalid expansion size", nil, testrailID)
	})

	AfterEach(func() {
		EndTorpedoTest()
	})

	stepLog := "select a pool and expand it by 30000000 GiB with resize-disk type"
	It(stepLog, func() {
		log.InfoD(stepLog)
		// pick pool to resize
		pools, err := GetAllPoolsPresent()
		log.FailOnError(err, "Unable to get the storage Pools")
		pooltoPick := pools[0]

		resizeErr := Inst().V.ExpandPool(pooltoPick, api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK, 30000000, true)
		dash.VerifyFatal(resizeErr != nil, true, "Verify error occurs with invalid Pool expansion size")

		// Verify error on pool expansion failure
		var errMatch error
		re := regexp.MustCompile(`.*cannot be expanded beyond maximum size.*`)
		if !re.MatchString(fmt.Sprintf("%v", resizeErr)) {
			errMatch = fmt.Errorf("failed to verify failure using invalid Pool size")
		}
		dash.VerifyFatal(errMatch, nil, "Pool expand with invalid PoolUUID failed as expected.")
	})

})

var _ = Describe("{PoolExpandResizeWithSameSize}", func() {

	var testrailID = 34542944
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/tests/view/34542944

	BeforeEach(func() {
		StartTorpedoTest("PoolExpandResizeWithSameSize",
			"Initiate pool expansion using same size", nil, testrailID)
	})

	AfterEach(func() {
		EndTorpedoTest()
	})

	stepLog := "select a pool and expand it by same pool size with resize-disk type"
	It(stepLog, func() {
		log.InfoD(stepLog)
		// pick pool to resize
		pools, err := GetAllPoolsPresent()
		log.FailOnError(err, "Unable to get the storage Pools")
		pooltoPick := pools[0]
		poolToBeResized = getStoragePool(pooltoPick)

		originalSizeGiB := poolToBeResized.TotalSize / units.GiB
		targetSizeGiB = originalSizeGiB
		resizeErr := Inst().V.ExpandPool(pooltoPick, api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK, targetSizeGiB, true)
		dash.VerifyFatal(resizeErr != nil, true, "Verify error occurs with same pool size")

		// Verify error on pool expansion failure
		var errMatch error
		re := regexp.MustCompile(`.*already at a size.*`)
		if !re.MatchString(fmt.Sprintf("%v", resizeErr)) {
			errMatch = fmt.Errorf("failed to verify failure using same Pool size")
		}
		dash.VerifyFatal(errMatch, nil, "Pool expand with Same Pool Size failed as expected.")
	})
})

var _ = Describe("{PoolExpandWhileResizeDiskInProgress}", func() {

	var testrailID = 34542896
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/tests/view/34542896

	BeforeEach(func() {
		StartTorpedoTest("PoolExpandWhileResizeDiskInProgress",
			"Initiate pool expansion on a pool where one pool expansion is already in progress", nil, testrailID)
		contexts = scheduleApps()
	})

	JustBeforeEach(func() {
		poolIDToResize = pickPoolToResize()
		log.Infof("Picked pool %s to resize", poolIDToResize)
		poolToBeResized = getStoragePool(poolIDToResize)
		storageNode, err = GetNodeWithGivenPoolID(poolIDToResize)
		log.FailOnError(err, "Failed to get node with given pool ID")
	})

	JustAfterEach(func() {
		AfterEachTest(contexts)
	})

	AfterEach(func() {
		appsValidateAndDestroy(contexts)
		EndTorpedoTest()
	})

	stepLog := "should get the existing pool and expand it by initiating a resize-disk and again trigger pool expand on same pool"
	It(stepLog, func() {
		log.InfoD(stepLog)

		originalSizeInBytes = poolToBeResized.TotalSize
		targetSizeInBytes = originalSizeInBytes + 100*units.GiB
		targetSizeGiB = targetSizeInBytes / units.GiB

		log.InfoD("Current Size of the pool %s is %d GiB. Trying to expand to %v GiB with type resize-disk",
			poolIDToResize, poolToBeResized.TotalSize/units.GiB, targetSizeGiB)
		triggerPoolExpansion(poolIDToResize, targetSizeGiB, api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK)

		// we are using pxctl command direclty as we dont want retries and Inst().V.ExpandPool does not returns required error
		pxctlCmdFull := fmt.Sprintf("pxctl sv pool expand -u %s -s %d -o resize-disk ", poolIDToResize, targetSizeGiB)

		// Execute the command and check the alerts of type POOL
		_, err := Inst().N.RunCommandWithNoRetry(*storageNode, pxctlCmdFull, node.ConnectionOpts{
			Timeout:         1 * time.Minute,
			TimeBeforeRetry: 10 * time.Second,
			IgnoreError:     false,
		})

		// Verify error on pool expansion failure
		var errMatch error
		re := regexp.MustCompile(`.*already in progress.*`)
		if !re.MatchString(fmt.Sprintf("%v", err)) {
			errMatch = fmt.Errorf("failed to verify pool expand when one already in progress")
		}
		dash.VerifyFatal(errMatch, nil, "Pool expand with one resize already in Porgress failed as expected.")

		Step("Ensure pool has been expanded to the expected size", func() {
			err = waitForOngoingPoolExpansionToComplete(poolIDToResize)
			dash.VerifyFatal(err, nil, "Pool expansion does not result in error")
			verifyPoolSizeEqualOrLargerThanExpected(poolIDToResize, targetSizeGiB)
		})

	})

})
