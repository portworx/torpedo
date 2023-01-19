package tests

import (
	"fmt"
	"math/rand"

	"github.com/google/uuid"
	"github.com/portworx/torpedo/drivers/node"
	"github.com/portworx/torpedo/drivers/volume"

	"github.com/portworx/torpedo/pkg/log"

	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/portworx/torpedo/pkg/testrailuttils"

	"github.com/libopenstorage/openstorage/api"
	. "github.com/onsi/ginkgo"
	"github.com/portworx/sched-ops/task"
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/pkg/units"
	. "github.com/portworx/torpedo/tests"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const poolResizeTimeout = time.Minute * 180
const retryTimeout = time.Minute * 2

var _ = Describe("{StoragePoolExpandDiskResize}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("StoragePoolExpandDiskResize", "Validate storage pool expansion using resize-disk option", nil, 0)
	})

	var contexts []*scheduler.Context
	stepLog := "has to schedule apps, and expand it by resizing a disk"
	It(stepLog, func() {
		log.InfoD(stepLog)
		contexts = make([]*scheduler.Context, 0)

		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("poolexpand-%d", i))...)
		}

		ValidateApplications(contexts)

		var poolIDToResize string

		pools, err := Inst().V.ListStoragePools(metav1.LabelSelector{})
		log.FailOnError(err, "Failed to list storage pools")
		dash.VerifyFatal(len(pools) > 0, true, " Storage pools exist?")

		// pick a pool from a pools list and resize it
		poolIDToResize, err = GetPoolIDWithIOs()
		log.FailOnError(err, "error identifying pool to run test")
		dash.VerifyFatal(len(poolIDToResize) > 0, true, fmt.Sprintf("Expected poolIDToResize to not be empty, pool id to resize %s", poolIDToResize))

		poolToBeResized := pools[poolIDToResize]
		dash.VerifyFatal(poolToBeResized != nil, true, "Pool to be resized exist?")

		// px will put a new request in a queue, but in this case we can't calculate the expected size,
		// so need to wain until the ongoing operation is completed
		time.Sleep(time.Second * 60)
		stepLog = "Verify that pool resize is not in progress"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			if poolResizeIsInProgress(poolToBeResized) {
				// wait until resize is completed and get the updated pool again
				poolToBeResized, err = GetStoragePoolByUUID(poolIDToResize)
				log.FailOnError(err, "Failed to get pool using UUID ")
			}
		})

		var expectedSize uint64
		var expectedSizeWithJournal uint64
		stepLog = "Calculate expected pool size and trigger pool resize"
		Step(stepLog, func() {
			expectedSize = poolToBeResized.TotalSize * 2 / units.GiB

			isjournal, err := isJournalEnabled()
			log.FailOnError(err, "Failed to check if Journal enabled")

			//To-Do Need to handle the case for multiple pools
			expectedSizeWithJournal = expectedSize
			if isjournal {
				expectedSizeWithJournal = expectedSizeWithJournal - 3
			}

			log.InfoD("Current Size of the pool %s is %d", poolIDToResize, poolToBeResized.TotalSize/units.GiB)

			err = Inst().V.ExpandPool(poolIDToResize, api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK, expectedSize)
			dash.VerifyFatal(err, nil, "Pool expansion init successful?")

			resizeErr := waitForPoolToBeResized(expectedSize, poolIDToResize, isjournal)
			dash.VerifyFatal(resizeErr, nil, fmt.Sprintf("Expected new size to be '%d' or '%d'", expectedSize, expectedSizeWithJournal))
		})

		stepLog = "Ensure that new pool has been expanded to the expected size"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			ValidateApplications(contexts)

			resizedPool, err := GetStoragePoolByUUID(poolIDToResize)
			log.FailOnError(err, "Failed to get pool using UUID ")
			newPoolSize := resizedPool.TotalSize / units.GiB
			isExpansionSuccess := false
			if newPoolSize == expectedSize || newPoolSize == expectedSizeWithJournal {
				isExpansionSuccess = true
			}
			dash.VerifyFatal(isExpansionSuccess, true, fmt.Sprintf("Expected new pool size to be %v or %v, got %v", expectedSize, expectedSizeWithJournal, newPoolSize))
		})

		stepLog = "destroy apps"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			opts := make(map[string]bool)
			opts[scheduler.OptionsWaitForResourceLeakCleanup] = true
			for _, ctx := range contexts {
				TearDownContext(ctx, opts)
			}
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts)
	})
})

var _ = Describe("{StoragePoolExpandDiskAdd}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("StoragePoolExpandDiskAdd", "Validate storage pool expansion using add-disk option", nil, 0)
	})
	var contexts []*scheduler.Context

	stepLog := "should get the existing pool and expand it by adding a disk"
	It(stepLog, func() {
		log.InfoD(stepLog)
		contexts = make([]*scheduler.Context, 0)

		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("pooladddisk-%d", i))...)
		}

		ValidateApplications(contexts)

		var poolIDToResize string

		pools, err := Inst().V.ListStoragePools(metav1.LabelSelector{})
		log.FailOnError(err, "Failed to list storage pools")
		dash.VerifyFatal(len(pools) > 0, true, "Storage pools exist ?")

		// pick a pool from a pools list and resize it
		poolIDToResize, err = GetPoolIDWithIOs()
		log.FailOnError(err, "error identifying pool to run test")
		dash.VerifyFatal(len(poolIDToResize) > 0, true, fmt.Sprintf("Expected poolIDToResize to not be empty, pool id to resize %s", poolIDToResize))

		poolToBeResized := pools[poolIDToResize]
		dash.VerifyFatal(poolToBeResized != nil, true, "Pool to be resized exist?")

		// px will put a new request in a queue, but in this case we can't calculate the expected size,
		// so need to wain until the ongoing operation is completed
		stepLog = "Verify that pool resize is not in progress"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			if poolResizeIsInProgress(poolToBeResized) {
				// wait until resize is completed and get the updated pool again
				poolToBeResized, err = GetStoragePoolByUUID(poolIDToResize)
				log.FailOnError(err, "Failed to get pool using UUID ")
			}
		})

		var expectedSize uint64
		var expectedSizeWithJournal uint64

		stepLog = "Calculate expected pool size and trigger pool resize"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			expectedSize = poolToBeResized.TotalSize * 2 / units.GiB
			expectedSize = roundUpValue(expectedSize)
			isjournal, err := isJournalEnabled()
			log.FailOnError(err, "Failed to check is Journal enabled")

			//To-Do Need to handle the case for multiple pools
			expectedSizeWithJournal = expectedSize
			if isjournal {
				expectedSizeWithJournal = expectedSizeWithJournal - 3
			}

			log.InfoD("Current Size of the pool %s is %d", poolIDToResize, poolToBeResized.TotalSize/units.GiB)

			err = Inst().V.ExpandPool(poolIDToResize, api.SdkStoragePool_RESIZE_TYPE_ADD_DISK, expectedSize)
			dash.VerifyFatal(err, nil, "Pool expansion init successful?")

			resizeErr := waitForPoolToBeResized(expectedSize, poolIDToResize, isjournal)
			dash.VerifyFatal(resizeErr, nil, fmt.Sprintf("Expected new size to be '%d' or '%d' if pool has journal", expectedSize, expectedSizeWithJournal))
		})

		Step("Ensure that new pool has been expanded to the expected size", func() {
			ValidateApplications(contexts)

			resizedPool, err := GetStoragePoolByUUID(poolIDToResize)
			log.FailOnError(err, "Failed to get pool using UUID ")
			newPoolSize := resizedPool.TotalSize / units.GiB
			isExpansionSuccess := false
			if newPoolSize >= expectedSizeWithJournal {
				isExpansionSuccess = true
			}
			dash.VerifyFatal(isExpansionSuccess, true,
				fmt.Sprintf("expected new pool size to be %v or %v if pool has journal, got %v", expectedSize, expectedSizeWithJournal, newPoolSize))
		})
		stepLog = "destroy apps"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			opts := make(map[string]bool)
			opts[scheduler.OptionsWaitForResourceLeakCleanup] = true
			for _, ctx := range contexts {
				TearDownContext(ctx, opts)
			}
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts)
	})
})

var _ = Describe("{StoragePoolExpandDiskAuto}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("StoragePoolExpandDiskAuto", "Validate storage pool expansion using auto option", nil, 0)
	})

	var contexts []*scheduler.Context
	stepLog := "has to schedule apps, and expand it by resizing a disk"
	It(stepLog, func() {
		log.InfoD(stepLog)
		contexts = make([]*scheduler.Context, 0)

		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("poolexpandauto-%d", i))...)
		}

		ValidateApplications(contexts)

		var poolIDToResize string

		pools, err := Inst().V.ListStoragePools(metav1.LabelSelector{})
		log.FailOnError(err, "Failed to list storage pools")
		dash.VerifyFatal(len(pools) > 0, true, " Storage pools exist?")

		// pick a pool from a pools list and resize it
		poolIDToResize, err = GetPoolIDWithIOs()
		log.FailOnError(err, "error identifying pool to run test")
		dash.VerifyFatal(len(poolIDToResize) > 0, true, fmt.Sprintf("Expected poolIDToResize to not be empty, pool id to resize %s", poolIDToResize))

		poolToBeResized := pools[poolIDToResize]
		dash.VerifyFatal(poolToBeResized != nil, true, "Pool to be resized exist?")

		// px will put a new request in a queue, but in this case we can't calculate the expected size,
		// so need to wain until the ongoing operation is completed
		time.Sleep(time.Second * 60)
		stepLog = "Verify that pool resize is not in progress"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			if poolResizeIsInProgress(poolToBeResized) {
				// wait until resize is completed and get the updated pool again
				poolToBeResized, err = GetStoragePoolByUUID(poolIDToResize)
				log.FailOnError(err, "Failed to get pool using UUID ")
			}
		})

		var expectedSize uint64
		var expectedSizeWithJournal uint64
		stepLog = "Calculate expected pool size and trigger pool resize"
		Step(stepLog, func() {
			expectedSize = poolToBeResized.TotalSize * 2 / units.GiB

			isjournal, err := isJournalEnabled()
			log.FailOnError(err, "Failed to check if Journal enabled")

			//To-Do Need to handle the case for multiple pools
			expectedSizeWithJournal = expectedSize
			if isjournal {
				expectedSizeWithJournal = expectedSizeWithJournal - 3
			}
			log.InfoD("Current Size of the pool %s is %d", poolIDToResize, poolToBeResized.TotalSize/units.GiB)
			err = Inst().V.ExpandPool(poolIDToResize, api.SdkStoragePool_RESIZE_TYPE_AUTO, expectedSize)
			dash.VerifyFatal(err, nil, "Pool expansion init successful?")

			resizeErr := waitForPoolToBeResized(expectedSize, poolIDToResize, isjournal)
			dash.VerifyFatal(resizeErr, nil, fmt.Sprintf("Expected new size to be '%d' or '%d'", expectedSize, expectedSizeWithJournal))
		})

		stepLog = "Ensure that new pool has been expanded to the expected size"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			ValidateApplications(contexts)

			resizedPool, err := GetStoragePoolByUUID(poolIDToResize)
			log.FailOnError(err, "Failed to get pool using UUID ")
			newPoolSize := resizedPool.TotalSize / units.GiB
			isExpansionSuccess := false
			if newPoolSize == expectedSize || newPoolSize == expectedSizeWithJournal {
				isExpansionSuccess = true
			}
			dash.VerifyFatal(isExpansionSuccess, true, fmt.Sprintf("Expected new pool size to be %v or %v, got %v", expectedSize, expectedSizeWithJournal, newPoolSize))

		})
		stepLog = "destroy apps"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			opts := make(map[string]bool)
			opts[scheduler.OptionsWaitForResourceLeakCleanup] = true
			for _, ctx := range contexts {
				TearDownContext(ctx, opts)
			}
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts)
	})
})

var _ = Describe("{PoolResizeDiskReboot}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("PoolResizeDiskReboot", "Initiate pool expansion using resize-disk and reboot node", nil, 0)
	})

	var contexts []*scheduler.Context

	stepLog := "has to schedule apps, and expand it by resizing a disk"
	It(stepLog, func() {
		contexts = make([]*scheduler.Context, 0)

		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("poolresizediskreboot-%d", i))...)
		}

		ValidateApplications(contexts)

		var poolIDToResize string

		pools, err := Inst().V.ListStoragePools(metav1.LabelSelector{})
		dash.VerifyFatal(err, nil, "Validate list storage pools")
		dash.VerifyFatal(len(pools) > 0, true, "Validate storage pools exist")

		// pick a pool from a pools list and resize it
		poolIDToResize, err = GetPoolIDWithIOs()
		log.FailOnError(err, "error identifying pool to run test")
		dash.VerifyFatal(len(poolIDToResize) > 0, true, fmt.Sprintf("Expected poolIDToResize to not be empty, pool id to resize %s", poolIDToResize))

		poolToBeResized := pools[poolIDToResize]
		dash.VerifyFatal(poolToBeResized != nil, true, "Pool to be resized exist?")

		// px will put a new request in a queue, but in this case we can't calculate the expected size,
		// so need to wain until the ongoing operation is completed
		time.Sleep(time.Second * 60)
		stepLog = "Verify that pool resize is not in progress"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			if poolResizeIsInProgress(poolToBeResized) {
				// wait until resize is completed and get the updated pool again
				poolToBeResized, err = GetStoragePoolByUUID(poolIDToResize)
				log.FailOnError(err, "Failed to get pool using UUID ")
			}
		})

		var expectedSize uint64
		var expectedSizeWithJournal uint64

		stepLog = "Calculate expected pool size and trigger pool resize"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			expectedSize = poolToBeResized.TotalSize * 2 / units.GiB

			isjournal, err := isJournalEnabled()
			log.FailOnError(err, "Failed to check is journal enabled")

			//To-Do Need to handle the case for multiple pools
			expectedSizeWithJournal = expectedSize
			if isjournal {
				expectedSizeWithJournal = expectedSizeWithJournal - 3
			}
			log.InfoD("Current Size of the pool %s is %d", poolIDToResize, poolToBeResized.TotalSize/units.GiB)
			err = Inst().V.ExpandPool(poolIDToResize, api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK, expectedSize)
			dash.VerifyFatal(err, nil, "Pool expansion init successful ?")

			err = WaitForExpansionToStart(poolIDToResize)
			log.FailOnError(err, "Expansion is not started")

			storageNode, err := GetNodeWithGivenPoolID(poolIDToResize)
			log.FailOnError(err, "Failed to get pool using UUID ")
			err = RebootNodeAndWait(*storageNode)
			log.FailOnError(err, "Failed to reboot node and wait till it is up")
			resizeErr := waitForPoolToBeResized(expectedSize, poolIDToResize, isjournal)
			dash.VerifyFatal(resizeErr, nil, fmt.Sprintf("Expected new size to be '%d' or '%d'", expectedSize, expectedSizeWithJournal))
		})

		stepLog = "Ensure that new pool has been expanded to the expected size"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			ValidateApplications(contexts)

			resizedPool, err := GetStoragePoolByUUID(poolIDToResize)
			log.FailOnError(err, "Failed to get pool using UUID ")
			newPoolSize := resizedPool.TotalSize / units.GiB
			isExpansionSuccess := false
			if newPoolSize == expectedSize || newPoolSize == expectedSizeWithJournal {
				isExpansionSuccess = true
			}
			dash.VerifyFatal(isExpansionSuccess, true,
				fmt.Sprintf("Expected new pool size to be %v or %v, got %v", expectedSize, expectedSizeWithJournal, newPoolSize))
		})
		stepLog = "destroy apps"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			opts := make(map[string]bool)
			opts[scheduler.OptionsWaitForResourceLeakCleanup] = true
			for _, ctx := range contexts {
				TearDownContext(ctx, opts)
			}
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts)
	})
})

var _ = Describe("{PoolAddDiskReboot}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("PoolAddDiskReboot", "Initiate pool expansion using add-disk and reboot node", nil, 0)
	})
	var contexts []*scheduler.Context

	stepLog := "should get the existing pool and expand it by adding a disk"

	It(stepLog, func() {
		log.InfoD(stepLog)
		contexts = make([]*scheduler.Context, 0)

		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("pooladddiskreboot-%d", i))...)
		}

		ValidateApplications(contexts)

		var poolIDToResize string

		pools, err := Inst().V.ListStoragePools(metav1.LabelSelector{})
		log.FailOnError(err, "Failed to list storage pools")
		dash.VerifyFatal(len(pools) > 0, true, "Storage pools exist?")

		// pick a pool from a pools list and resize it
		poolIDToResize, err = GetPoolIDWithIOs()
		log.FailOnError(err, "error identifying pool to run test")
		dash.VerifyFatal(len(poolIDToResize) > 0, true, fmt.Sprintf("Expected poolIDToResize to not be empty, pool id to resize %s", poolIDToResize))

		poolToBeResized := pools[poolIDToResize]
		dash.VerifyFatal(poolToBeResized != nil, true, "Pool to be resized exist?")

		// px will put a new request in a queue, but in this case we can't calculate the expected size,
		// so need to wain until the ongoing operation is completed
		stepLog = "Verify that pool resize is not in progress"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			if poolResizeIsInProgress(poolToBeResized) {
				// wait until resize is completed and get the updated pool again
				poolToBeResized, err = GetStoragePoolByUUID(poolIDToResize)
				log.FailOnError(err, "Failed to get pool using UUID")
			}
		})

		var expectedSize uint64
		var expectedSizeWithJournal uint64

		stepLog = "Calculate expected pool size and trigger pool resize"
		Step(stepLog, func() {
			expectedSize = poolToBeResized.TotalSize * 2 / units.GiB
			expectedSize = roundUpValue(expectedSize)
			isjournal, err := isJournalEnabled()
			log.FailOnError(err, "Failed to check is journal enabled")

			//To-Do Need to handle the case for multiple pools
			expectedSizeWithJournal = expectedSize
			if isjournal {
				expectedSizeWithJournal = expectedSizeWithJournal - 3
			}
			log.InfoD("Current Size of the pool %s is %d", poolIDToResize, poolToBeResized.TotalSize/units.GiB)
			err = Inst().V.ExpandPool(poolIDToResize, api.SdkStoragePool_RESIZE_TYPE_ADD_DISK, expectedSize)
			dash.VerifyFatal(err, nil, "Pool expansion init successful?")

			err = WaitForExpansionToStart(poolIDToResize)
			log.FailOnError(err, "Failed while waiting for expansion to start")

			storageNode, err := GetNodeWithGivenPoolID(poolIDToResize)
			log.FailOnError(err, " Failed to get pool using UUID")
			err = RebootNodeAndWait(*storageNode)
			log.FailOnError(err, "Failed to reboot node and wait till it is up")
			resizeErr := waitForPoolToBeResized(expectedSize, poolIDToResize, isjournal)
			dash.VerifyFatal(resizeErr, nil, fmt.Sprintf("Expected new size to be '%d' or '%d' if pool has journal", expectedSize, expectedSizeWithJournal))
		})

		stepLog = "Ensure that new pool has been expanded to the expected size"
		Step(stepLog, func() {
			ValidateApplications(contexts)

			resizedPool, err := GetStoragePoolByUUID(poolIDToResize)
			log.FailOnError(err, " Failed to get pool using UUID")
			newPoolSize := resizedPool.TotalSize / units.GiB
			isExpansionSuccess := false
			if newPoolSize == expectedSize || newPoolSize == expectedSizeWithJournal {
				isExpansionSuccess = true
			}
			dash.VerifyFatal(isExpansionSuccess, true,
				fmt.Sprintf("Expected new pool size to be %v or %v if pool has journal, got %v", expectedSize, expectedSizeWithJournal, newPoolSize))
		})
		stepLog = "destroy apps"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			opts := make(map[string]bool)
			opts[scheduler.OptionsWaitForResourceLeakCleanup] = true
			for _, ctx := range contexts {
				TearDownContext(ctx, opts)
			}
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts)
	})
})

var _ = Describe("{NodePoolsResizeDisk}", func() {

	nodePoolsExpansion("NodePoolsResizeDisk")

})

var _ = Describe("{NodePoolsAddDisk}", func() {

	nodePoolsExpansion("NodePoolsAddDisk")

})

func nodePoolsExpansion(testName string) {

	var operation api.SdkStoragePool_ResizeOperationType
	var option string
	if testName == "NodePoolsResizeDisk" {
		operation = api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK
		option = "resize-disk"
	} else {
		operation = api.SdkStoragePool_RESIZE_TYPE_ADD_DISK
		option = "add-disk"
	}

	JustBeforeEach(func() {
		StartTorpedoTest(testName, fmt.Sprintf("Validate multi storage pools on the same node expansion  using %s option", option), nil, 0)
	})

	var contexts []*scheduler.Context
	stepLog := fmt.Sprintf("has to schedule apps, and expand it by %s", option)
	It(stepLog, func() {
		log.InfoD(stepLog)
		contexts = make([]*scheduler.Context, 0)

		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("nodepools-%s-%d", option, i))...)
		}

		ValidateApplications(contexts)

		var poolsToBeResized []*api.StoragePool

		stNodes := node.GetStorageNodes()
		var nodePoolToExpanded node.Node
		var nodePools []node.StoragePool
		for _, stNode := range stNodes {
			nodePools = stNode.StoragePools
			nodePoolToExpanded = stNode
			if len(nodePools) > 1 {
				break
			}
		}
		pools, err := Inst().V.ListStoragePools(metav1.LabelSelector{})
		log.FailOnError(err, "Failed to list storage pools")
		dash.VerifyFatal(len(nodePools) > 1, true, "Node has multiple storage pools?")

		for _, p := range nodePools {
			poolsToBeResized = append(poolsToBeResized, pools[p.Uuid])
		}

		dash.VerifyFatal(poolsToBeResized != nil, true, "Pools pending to be resized")

		// px will put a new request in a queue, but in this case we can't calculate the expected size,
		// so need to wait until the ongoing operation is completed
		stepLog = "Verify that pool resize is not in progress"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for _, poolToBeResized := range poolsToBeResized {
				poolIDToResize := poolToBeResized.Uuid
				if poolResizeIsInProgress(poolToBeResized) {
					// wait until resize is completed and get the updated pool again
					poolToBeResized, err = GetStoragePoolByUUID(poolIDToResize)
					log.FailOnError(err, fmt.Sprintf("Failed to get pool using UUID  %s", poolIDToResize))
				}
			}

		})

		var expectedSize uint64
		var expectedSizeWithJournal uint64
		poolsExpectedSizeMap := make(map[string]uint64)
		isjournal, err := isJournalEnabled()
		log.FailOnError(err, "Failed to check is Journal Enabled")
		stepLog = fmt.Sprintf("Calculate expected pool size and trigger pool resize for %s", nodePoolToExpanded.Name)
		Step(stepLog, func() {

			for _, poolToBeResized := range poolsToBeResized {
				expectedSize = poolToBeResized.TotalSize * 2 / units.GiB
				poolsExpectedSizeMap[poolToBeResized.Uuid] = expectedSize

				//To-Do Need to handle the case for multiple pools
				expectedSizeWithJournal = expectedSize
				if isjournal {
					expectedSizeWithJournal = expectedSizeWithJournal - 3
				}
				log.InfoD("Current Size of the pool %s is %d", poolToBeResized.Uuid, poolToBeResized.TotalSize/units.GiB)
				err = Inst().V.ExpandPool(poolToBeResized.Uuid, operation, expectedSize)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Pool %s expansion init succesful?", poolToBeResized.Uuid))
			}

			for poolUUID, expectedSize := range poolsExpectedSizeMap {
				resizeErr := waitForPoolToBeResized(expectedSize, poolUUID, isjournal)
				expectedSizeWithJournal = expectedSize
				if isjournal {
					expectedSizeWithJournal = expectedSizeWithJournal - 3
				}
				log.FailOnError(resizeErr, fmt.Sprintf("Expected new size to be '%d' or '%d'", expectedSize, expectedSizeWithJournal))
			}

		})

		stepLog = "Ensure that pools have been expanded to the expected size"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			ValidateApplications(contexts)
			for poolUUID, expectedSize := range poolsExpectedSizeMap {
				resizedPool, err := GetStoragePoolByUUID(poolUUID)
				log.FailOnError(err, fmt.Sprintf("Failed to get pool using UUID  %s", poolUUID))
				newPoolSize := resizedPool.TotalSize / units.GiB
				isExpansionSuccess := false
				expectedSizeWithJournal = expectedSize
				if isjournal {
					expectedSizeWithJournal = expectedSizeWithJournal - 3
				}
				if newPoolSize == expectedSize || newPoolSize == expectedSizeWithJournal {
					isExpansionSuccess = true
				}
				dash.VerifyFatal(isExpansionSuccess, true, fmt.Sprintf("Expected new pool size to be %v or %v, got %v", expectedSize, expectedSizeWithJournal, newPoolSize))
			}

		})
		stepLog = "destroy apps"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			opts := make(map[string]bool)
			opts[scheduler.OptionsWaitForResourceLeakCleanup] = true
			for _, ctx := range contexts {
				TearDownContext(ctx, opts)
			}
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts)
	})
}

var _ = Describe("{AddNewPoolWhileRebalance}", func() {
	//AddNewPoolWhileRebalance:
	//
	//step1: create volume repl=2, and get its pool P1 on n1 and p2 on n2
	//
	//step2: feed 10GB I/O on the volume
	//
	//step3: After I/O expand the pool p1 when p1 is rebalancing add a new drive with different size
	//so that a new pool would be created
	//
	//step4: validate the pool and the data
	var testrailID = 51441
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/cases/view/51441
	var runID int
	JustBeforeEach(func() {
		StartTorpedoTest("AddNewPoolWhileRebalance", "Validate adding nee storage pool while another pool rebalancing", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})

	var contexts []*scheduler.Context
	stepLog := "has to schedule apps, and expand it by resizing a disk"
	It(stepLog, func() {
		log.InfoD(stepLog)
		contexts = make([]*scheduler.Context, 0)

		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("addnewpoolrebal-%d", i))...)
		}

		ValidateApplications(contexts)

		var poolIDToResize string

		stNodes := node.GetStorageNodes()
		var poolToBeResized *api.StoragePool
		var currentTotalPoolSize uint64
		var err error
		var nodeSelected node.Node
		var pools map[string]*api.StoragePool
		var volSelected *volume.Volume
		volSelected, err = getVolumeWithMinimumSize(contexts, 10)
		log.FailOnError(err, "error identifying volume")
		log.Infof("%+v", volSelected)
		rs, err := Inst().V.GetReplicaSets(volSelected)
		log.FailOnError(err, fmt.Sprintf("error getting replica sets for vol %s", volSelected.Name))
		attachedNodeID := rs[0].Nodes[0]
		volumePools := rs[0].PoolUuids
		for _, stNode := range stNodes {
			if stNode.Id == attachedNodeID {
				nodeSelected = stNode
			}
		}

		if &nodeSelected == nil {
			dash.VerifyFatal(false, true, "unable to identify the node for add new pool")
		}
	poolloop:
		for _, volPool := range volumePools {
			for _, nodePool := range nodeSelected.Pools {
				if nodePool.Uuid == volPool {
					poolIDToResize = nodePool.Uuid
					break poolloop
				}
			}
		}
		dash.Infof("selected node %s, pool %s", nodeSelected.Name, poolIDToResize)
		poolToBeResized, err = GetStoragePoolByUUID(poolIDToResize)
		log.FailOnError(err, "unable to get pool using UUID")
		currentTotalPoolSize = poolToBeResized.TotalSize / units.GiB
		pools, err = Inst().V.ListStoragePools(metav1.LabelSelector{})
		log.FailOnError(err, "error getting storage pools")
		existingPoolsCount := len(pools)
		///creating a spec to perform add  drive
		driveSpecs, err := GetCloudDriveDeviceSpecs()
		log.FailOnError(err, "Error getting cloud drive specs")

		deviceSpec := driveSpecs[0]
		deviceSpecParams := strings.Split(deviceSpec, ",")
		var specSize uint64
		paramsArr := make([]string, 0)
		for _, param := range deviceSpecParams {
			if strings.Contains(param, "size") {
				val := strings.Split(param, "=")[1]
				specSize, err = strconv.ParseUint(val, 10, 64)
				log.FailOnError(err, "Error converting size to uint64")
				paramsArr = append(paramsArr, fmt.Sprintf("size=%d,", specSize/2))
			} else {
				paramsArr = append(paramsArr, param)
			}
		}
		newSpec := strings.Join(paramsArr, ",")
		expandedExpectedPoolSize := currentTotalPoolSize * 2

		stepLog = fmt.Sprintf("Verify that pool %s can be expanded", poolIDToResize)
		Step(stepLog, func() {
			log.InfoD(stepLog)
			isPoolHealthy := poolResizeIsInProgress(poolToBeResized)
			dash.VerifyFatal(isPoolHealthy, true, "Verfiy pool before expansion")
		})

		stepLog = fmt.Sprintf("Trigger pool %s resize by add-disk", poolIDToResize)
		Step(stepLog, func() {
			log.InfoD(stepLog)
			dash.VerifyFatal(err, nil, "Validate is journal enabled check")
			err = Inst().V.ExpandPool(poolIDToResize, api.SdkStoragePool_RESIZE_TYPE_ADD_DISK, expandedExpectedPoolSize)
			log.FailOnError(err, "failed to initiate pool expansion")
		})

		stepLog = fmt.Sprintf("Ensure that pool %s rebalance started and add new pool to the node %s", poolIDToResize, nodeSelected.Name)
		Step(stepLog, func() {
			log.InfoD(stepLog)
			t := func() (interface{}, bool, error) {
				expandedPool, err := GetStoragePoolByUUID(poolIDToResize)
				if err != nil {
					return nil, true, fmt.Errorf("error getting pool by using id %s", poolIDToResize)
				}

				if expandedPool == nil {
					return nil, false, fmt.Errorf("expanded pool value is nil")
				}
				if expandedPool.LastOperation != nil {
					log.Infof("Pool Resize Status : %v, Message : %s", expandedPool.LastOperation.Status, expandedPool.LastOperation.Msg)
					if expandedPool.LastOperation.Status == api.SdkStoragePool_OPERATION_IN_PROGRESS &&
						(strings.Contains(expandedPool.LastOperation.Msg, "Storage rebalance is running") || strings.Contains(expandedPool.LastOperation.Msg, "Rebalance in progress")) {
						return nil, false, nil
					}
					if expandedPool.LastOperation.Status == api.SdkStoragePool_OPERATION_FAILED {
						return nil, false, fmt.Errorf("PoolResize has failed. Error: %s", expandedPool.LastOperation)
					}

				}
				return nil, true, fmt.Errorf("pool status not updated")
			}
			_, err = task.DoRetryWithTimeout(t, 5*time.Minute, 10*time.Second)
			log.FailOnError(err, "Error checking pool rebalance")

			err = Inst().V.AddCloudDrive(&nodeSelected, newSpec, -1)
			log.FailOnError(err, fmt.Sprintf("Add cloud drive failed on node %s", nodeSelected.Name))

			log.InfoD("Validate pool rebalance after drive add")
			err = ValidatePoolRebalance()
			log.FailOnError(err, fmt.Sprintf("pool %s rebalance failed", poolIDToResize))
			isjournal, err := isJournalEnabled()
			log.FailOnError(err, "is journal enabled check failed")
			err = waitForPoolToBeResized(expandedExpectedPoolSize, poolIDToResize, isjournal)
			log.FailOnError(err, "Error waiting for poor resize")
			resizedPool, err := GetStoragePoolByUUID(poolIDToResize)
			log.FailOnError(err, fmt.Sprintf("error get pool using UUID %s", poolIDToResize))
			newPoolSize := resizedPool.TotalSize / units.GiB
			isExpansionSuccess := false
			expectedSizeWithJournal := expandedExpectedPoolSize - 3

			if newPoolSize >= expectedSizeWithJournal {
				isExpansionSuccess = true
			}
			dash.VerifyFatal(isExpansionSuccess, true, fmt.Sprintf("expected new pool size to be %v or %v, got %v", expandedExpectedPoolSize, expectedSizeWithJournal, newPoolSize))
			pools, err = Inst().V.ListStoragePools(metav1.LabelSelector{})
			log.FailOnError(err, "error getting storage pools")

			dash.VerifyFatal(len(pools), existingPoolsCount+1, "Validate new pool is created")
			ValidateApplications(contexts)
			for _, stNode := range stNodes {
				status, err := Inst().V.GetNodeStatus(stNode)
				log.FailOnError(err, fmt.Sprintf("Error getting PX status of node %s", stNode.Name))
				dash.VerifySafely(status, api.Status_STATUS_OK, fmt.Sprintf("validate PX status on node %s", stNode.Name))
			}
		})
		stepLog = "destroy apps"
		Step(stepLog, func() {
			log.InfoD(stepLog)
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

func roundUpValue(toRound uint64) uint64 {

	if toRound%10 == 0 {
		return toRound
	}
	rs := (10 - toRound%10) + toRound
	return rs

}

func poolResizeIsInProgress(poolToBeResized *api.StoragePool) bool {
	if poolToBeResized.LastOperation != nil {
		f := func() (interface{}, bool, error) {
			pools, err := Inst().V.ListStoragePools(metav1.LabelSelector{})
			if err != nil || len(pools) == 0 {
				return nil, true, fmt.Errorf("error getting pools list, err %v", err)
			}

			updatedPoolToBeResized := pools[poolToBeResized.Uuid]
			if updatedPoolToBeResized == nil {
				return nil, false, fmt.Errorf("error getting pool with given pool id %s", poolToBeResized.Uuid)
			}

			if updatedPoolToBeResized.LastOperation.Status != api.SdkStoragePool_OPERATION_SUCCESSFUL {
				log.Infof("Current pool status : %v", updatedPoolToBeResized.LastOperation)
				if updatedPoolToBeResized.LastOperation.Status == api.SdkStoragePool_OPERATION_FAILED {
					dash.VerifyFatal(updatedPoolToBeResized.LastOperation.Status, api.SdkStoragePool_OPERATION_SUCCESSFUL, fmt.Sprintf("PoolResize has failed. Error: %s", updatedPoolToBeResized.LastOperation))
					return nil, false, fmt.Errorf("PoolResize has failed. Error: %s", updatedPoolToBeResized.LastOperation)
				}
				err = ValidatePoolRebalance()
				if err != nil {
					return nil, true, fmt.Errorf("errorvalidatng  err %v", err)
				}
				log.Infof("Pool Resize is already in progress: %v", updatedPoolToBeResized.LastOperation)
				return nil, true, nil
			}
			return nil, false, nil
		}

		_, err := task.DoRetryWithTimeout(f, poolResizeTimeout, retryTimeout)
		if err != nil {
			dash.VerifyFatal(err, nil, "Verify pool status before expansion")
		}
		return true
	}
	return true
}

func waitForPoolToBeResized(expectedSize uint64, poolIDToResize string, isJournalEnabled bool) error {

	currentLastMsg := ""
	f := func() (interface{}, bool, error) {
		expandedPool, err := GetStoragePoolByUUID(poolIDToResize)
		if err != nil {
			return nil, true, fmt.Errorf("error getting pool by using id %s", poolIDToResize)
		}

		if expandedPool == nil {
			return nil, false, fmt.Errorf("expanded pool value is nil")
		}
		if expandedPool.LastOperation != nil {
			log.Infof("Pool Resize Status : %v, Message : %s", expandedPool.LastOperation.Status, expandedPool.LastOperation.Msg)
			if expandedPool.LastOperation.Status == api.SdkStoragePool_OPERATION_FAILED {
				return nil, false, fmt.Errorf("PoolResize has failed. Error: %s", expandedPool.LastOperation)
			}
			if expandedPool.LastOperation.Status == api.SdkStoragePool_OPERATION_IN_PROGRESS {
				if strings.Contains(expandedPool.LastOperation.Msg, "Rebalance in progress") {
					if currentLastMsg == expandedPool.LastOperation.Msg {
						return nil, false, fmt.Errorf("pool reblance is not progressing")
					}
					currentLastMsg = expandedPool.LastOperation.Msg
					return nil, true, fmt.Errorf("wait for pool rebalance to complete")
				}
				return nil, true, fmt.Errorf("waiting for pool status to update")
			}
		}
		newPoolSize := expandedPool.TotalSize / units.GiB
		err = ValidatePoolRebalance()
		if err != nil {
			return nil, true, fmt.Errorf("pool %s not been resized .Current size is %d,Error while pool rebalance: %v", poolIDToResize, newPoolSize, err)
		}
		expectedSizeWithJournal := expectedSize
		if isJournalEnabled {
			expectedSizeWithJournal = expectedSizeWithJournal - 3
		}
		if newPoolSize >= expectedSizeWithJournal {
			// storage pool resize has been completed
			return nil, false, nil
		}
		return nil, true, fmt.Errorf("pool has not been resized to %d or %d yet. Waiting...Current size is %d", expectedSize, expectedSizeWithJournal, newPoolSize)
	}

	_, err := task.DoRetryWithTimeout(f, poolResizeTimeout, retryTimeout)
	return err
}

func isJournalEnabled() (bool, error) {
	storageSpec, err := Inst().V.GetStorageSpec()
	if err != nil {
		return false, err
	}
	jDev := storageSpec.GetJournalDev()
	if jDev != "" {
		log.Infof("JournalDev: %s", jDev)
		return true, nil
	}
	return false, nil
}

var _ = Describe("{PoolAddDrive}", func() {
	var testrailID = 2017
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/cases/view/2017
	var runID int
	JustBeforeEach(func() {
		StartTorpedoTest("PoolAddDrive", "Initiate pool expansion using add-drive", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})
	var contexts []*scheduler.Context

	stepLog := "should get the existing storage node and expand the pool by adding a drive"

	It(stepLog, func() {
		log.InfoD(stepLog)
		contexts = make([]*scheduler.Context, 0)
		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("pooladddrive-%d", i))...)
		}
		ValidateApplications(contexts)

		stNodes := node.GetStorageNodes()
		if len(stNodes) == 0 {
			dash.VerifyFatal(len(stNodes) > 0, true, "Storage nodes found?")
		}
		stNode, err := GetRandomNodeWithPoolIOs(stNodes)
		log.FailOnError(err, "error identifying node to run test")
		err = addCloudDrive(stNode, -1)
		log.FailOnError(err, "error adding cloud drive")
		stepLog = "destroy apps"
		Step(stepLog, func() {
			log.InfoD(stepLog)
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

var _ = Describe("{AddDriveAndPXRestart}", func() {
	//1) Deploy px with cloud drive.
	//2) Create a volume on that pool and write some data on the volume.
	//3) Expand pool by adding cloud drives.
	//4) Restart px service where the pool is present.
	var testrailID = 2014
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/cases/view/2014
	var runID int

	JustBeforeEach(func() {
		StartTorpedoTest("AddDriveAndPXRestart", "Initiate pool expansion using add-drive and restart PX", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})
	var contexts []*scheduler.Context

	stepLog := "should get the existing storage node and expand the pool by adding a drive"

	It(stepLog, func() {
		log.InfoD(stepLog)
		contexts = make([]*scheduler.Context, 0)
		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("pladddrvrestrt-%d", i))...)
		}
		ValidateApplications(contexts)

		stNodes := node.GetStorageNodes()
		if len(stNodes) == 0 {
			dash.VerifyFatal(len(stNodes) > 0, true, "Storage nodes found?")
		}
		stNode, err := GetRandomNodeWithPoolIOs(stNodes)
		log.FailOnError(err, "error identifying node to run test")
		err = addCloudDrive(stNode, -1)
		log.FailOnError(err, "error adding cloud drive")
		stepLog = fmt.Sprintf("Restart PX on node %s", stNode.Name)
		Step(stepLog, func() {
			log.InfoD(stepLog)
			err := Inst().V.RestartDriver(stNode, nil)
			log.FailOnError(err, fmt.Sprintf("error restarting px on node %s", stNode.Name))
			err = Inst().V.WaitDriverUpOnNode(stNode, 2*time.Minute)
			log.FailOnError(err, fmt.Sprintf("Driver is down on node %s", stNode.Name))
			dash.VerifyFatal(err == nil, true, fmt.Sprintf("PX is up after restarting on node %s", stNode.Name))
		})
		opts := make(map[string]bool)
		opts[scheduler.OptionsWaitForResourceLeakCleanup] = true
		ValidateAndDestroy(contexts, opts)

	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})

})

var _ = Describe("{AddDriveWithPXRestart}", func() {
	//1) Deploy px with cloud drive.
	//2) Create a volume on that pool and write some data on the volume.
	//3) Expand pool by adding cloud drives.
	//4) Restart px service where the pool expansion is in-progress
	var testrailID = 50632
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/cases/view/50632
	var runID int

	JustBeforeEach(func() {
		StartTorpedoTest("AddDriveWithPXRestart", "Initiate pool expansion using add-drive and restart PX while it is in progress", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})
	var contexts []*scheduler.Context

	stepLog := "should get the existing storage node and expand the pool by adding a drive"

	It(stepLog, func() {
		log.InfoD(stepLog)
		contexts = make([]*scheduler.Context, 0)
		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("pladddrvwrst-%d", i))...)
		}
		ValidateApplications(contexts)

		stNodes := node.GetStorageNodes()
		if len(stNodes) == 0 {
			dash.VerifyFatal(len(stNodes) > 0, true, "Storage nodes found?")
		}
		stNode, err := GetRandomNodeWithPoolIOs(stNodes)
		log.FailOnError(err, "error identifying node to run test")
		pools, err := Inst().V.ListStoragePools(metav1.LabelSelector{})
		log.FailOnError(err, "error getting pools list")
		dash.VerifyFatal(len(pools) > 0, true, "Verify pools exist")

		var currentTotalPoolSize uint64
		var specSize uint64
		for _, pool := range pools {
			currentTotalPoolSize += pool.GetTotalSize() / units.GiB
		}

		driveSpecs, err := GetCloudDriveDeviceSpecs()
		log.FailOnError(err, "Error getting cloud drive specs")
		deviceSpec := driveSpecs[0]
		deviceSpecParams := strings.Split(deviceSpec, ",")

		for _, param := range deviceSpecParams {
			if strings.Contains(param, "size") {
				val := strings.Split(param, "=")[1]
				specSize, err = strconv.ParseUint(val, 10, 64)
				log.FailOnError(err, "Error converting size to uint64")
			}
		}
		expectedTotalPoolSize := currentTotalPoolSize + specSize

		stepLog := "Initiate add cloud drive and restart PX"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			err = Inst().V.AddCloudDrive(&stNode, deviceSpec, -1)
			log.FailOnError(err, fmt.Sprintf("Add cloud drive failed on node %s", stNode.Name))
			time.Sleep(3 * time.Second)
			err = Inst().V.RestartDriver(stNode, nil)
			log.FailOnError(err, fmt.Sprintf("error restarting px on node %s", stNode.Name))
			err = Inst().V.WaitDriverUpOnNode(stNode, 2*time.Minute)
			log.FailOnError(err, fmt.Sprintf("Driver is down on node %s", stNode.Name))
			log.InfoD("Validate pool rebalance after drive add")
			err = ValidatePoolRebalance()
			log.FailOnError(err, "Pool re-balance failed")
			dash.VerifyFatal(err == nil, true, "PX is up after add drive")

			var newTotalPoolSize uint64
			pools, err := Inst().V.ListStoragePools(metav1.LabelSelector{})
			log.FailOnError(err, "error getting pools list")
			dash.VerifyFatal(len(pools) > 0, true, "Verify pools exist")
			for _, pool := range pools {
				newTotalPoolSize += pool.GetTotalSize() / units.GiB
			}
			dash.VerifyFatal(newTotalPoolSize, expectedTotalPoolSize, fmt.Sprintf("Validate total pool size after add cloud drive on node %s", stNode.Name))
		})
		opts := make(map[string]bool)
		opts[scheduler.OptionsWaitForResourceLeakCleanup] = true
		ValidateAndDestroy(contexts, opts)

	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})

})

var _ = Describe("{PoolAddDriveVolResize}", func() {
	//1) Deploy px with cloud drive.
	//2) Create a volume on that pool and write some data on the volume.
	//3) Expand pool by adding cloud drives.
	//4) expand the volume to the resized pool
	var testrailID = 2018
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/cases/view/2018
	var runID int
	JustBeforeEach(func() {
		StartTorpedoTest("PoolAddDriveVolResize", "pool expansion using add-drive and expand volume to the pool", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})
	var contexts []*scheduler.Context

	stepLog := "should get the existing storage node and expand the pool by adding a drive"

	It(stepLog, func() {
		log.InfoD(stepLog)
		contexts = make([]*scheduler.Context, 0)
		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("pooladddrive-%d", i))...)
		}
		ValidateApplications(contexts)

		stNodes := node.GetStorageNodes()
		if len(stNodes) == 0 {
			dash.VerifyFatal(len(stNodes) > 0, true, "Storage nodes found?")
		}
		volSelected, err := getVolumeWithMinimumSize(contexts, 10)
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
		selectedPool := stNode.StoragePools[0]
		err = addCloudDrive(stNode, selectedPool.ID)
		log.FailOnError(err, "error adding cloud drive")
		stepLog = "Expand volume to the expanded pool"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			currRep, err := Inst().V.GetReplicationFactor(volSelected)
			log.FailOnError(err, fmt.Sprintf("err getting repl factor for  vol : %s", volSelected.Name))
			opts := volume.Options{
				ValidateReplicationUpdateTimeout: validateReplicationUpdateTimeout,
			}
			newRep := currRep
			if currRep == 3 {
				newRep = currRep - 1
				err = Inst().V.SetReplicationFactor(volSelected, newRep, nil, nil, true, opts)
				log.FailOnError(err, fmt.Sprintf("err setting repl factor  to %d for  vol : %s", newRep, volSelected.Name))
			}
			log.InfoD(fmt.Sprintf("setting repl factor  to %d for  vol : %s", newRep+1, volSelected.Name))
			err = Inst().V.SetReplicationFactor(volSelected, newRep+1, []string{stNode.Id}, []string{selectedPool.Uuid}, true, opts)
			log.FailOnError(err, fmt.Sprintf("err setting repl factor  to %d for  vol : %s", newRep+1, volSelected.Name))
			dash.VerifyFatal(err == nil, true, fmt.Sprintf("vol %s expanded successfully on node %s", volSelected.Name, stNode.Name))
		})
		for _, ctx := range contexts {
			ctx.SkipVolumeValidation = true
			ValidateContext(ctx)
		}
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

var _ = Describe("{AddDriveMaintenanceMode}", func() {
	var testrailID = 2013
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/cases/view/2013
	var runID int
	JustBeforeEach(func() {
		StartTorpedoTest("AddDriveMaintenanceMode", "pool expansion using add-drive when node is in maintenance mode", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})
	var contexts []*scheduler.Context

	stepLog := "should get the existing storage node and put it in maintenance mode"

	It(stepLog, func() {
		log.InfoD(stepLog)
		contexts = make([]*scheduler.Context, 0)
		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("adddrvmnt-%d", i))...)
		}
		ValidateApplications(contexts)

		stNodes := node.GetStorageNodes()
		if len(stNodes) == 0 {
			dash.VerifyFatal(len(stNodes) > 0, true, "Storage nodes found?")
		}
		stNode, err := GetRandomNodeWithPoolIOs(stNodes)
		log.FailOnError(err, "error identifying node to run test")
		err = Inst().V.EnterMaintenance(stNode)
		log.FailOnError(err, fmt.Sprintf("fail to enter node %s in maintenence mode", stNode.Name))
		status, err := Inst().V.GetNodeStatus(stNode)
		log.Infof(fmt.Sprintf("Node %s status %s", stNode.Name, status.String()))
		stepLog = fmt.Sprintf("add cloud drive to the node %s", stNode.Name)
		Step(stepLog, func() {
			log.InfoD(stepLog)
			err = addCloudDrive(stNode, -1)
			if strings.Contains(err.Error(), "command terminated with exit code 1") {
				dash.VerifySafely(true, true, "Add drive failed when node is in maintenance mode")
			} else {
				log.FailOnError(err, "add drive operation failed")
			}
		})
		t := func() (interface{}, bool, error) {
			if err := Inst().V.ExitMaintenance(stNode); err != nil {
				return nil, true, err
			}
			return nil, false, nil
		}

		_, err = task.DoRetryWithTimeout(t, 15*time.Minute, 2*time.Minute)
		log.FailOnError(err, fmt.Sprintf("fail to exit maintenence mode in node %s", stNode.Name))
		status, err = Inst().V.GetNodeStatus(stNode)
		log.Infof(fmt.Sprintf("Node %s status %s after exit", stNode.Name, status.String()))
	})
	opts := make(map[string]bool)
	opts[scheduler.OptionsWaitForResourceLeakCleanup] = true
	ValidateAndDestroy(contexts, opts)
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})
})

var _ = Describe("{AddDriveStoragelessAndResize}", func() {
	var testrailID = 50617
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/cases/view/2017
	var runID int
	JustBeforeEach(func() {
		StartTorpedoTest("AddDriveStorageless", "Initiate add-drive to storageless node and pool expansion", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})
	var contexts []*scheduler.Context

	stepLog := "should get the storageless node and add a drive"

	It(stepLog, func() {
		log.InfoD(stepLog)
		contexts = make([]*scheduler.Context, 0)
		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("adddrvsl-%d", i))...)
		}
		ValidateApplications(contexts)

		slNodes := node.GetStorageLessNodes()
		if len(slNodes) == 0 {
			dash.VerifyFatal(len(slNodes) > 0, true, "Storage less nodes found?")
		}
		slNode := GetRandomStorageLessNode(slNodes)
		err := addCloudDrive(slNode, -1)
		log.FailOnError(err, "error adding cloud drive")
		err = Inst().V.RefreshDriverEndpoints()
		log.FailOnError(err, "error refreshing end points")
		stNodes := node.GetStorageNodes()
		var stNode node.Node
		for _, n := range stNodes {
			if n.Id == slNode.Id {
				stNode = n
				break
			}
		}
		dash.VerifyFatal(stNode.Name != "", true, fmt.Sprintf("Verify node %s is converted to storage node", slNode.Name))

		poolToResize := stNode.Pools[0]

		dash.VerifyFatal(poolToResize != nil, true, fmt.Sprintf("Is pool identified frpm stroage node %s?", stNode.Name))

		pools, err := Inst().V.ListStoragePools(metav1.LabelSelector{})
		log.FailOnError(err, "error getting pools list")

		poolToBeResized := pools[poolToResize.Uuid]
		dash.VerifyFatal(poolToBeResized != nil, true, "Pool to be resized exist?")

		// px will put a new request in a queue, but in this case we can't calculate the expected size,
		// so need to wain until the ongoing operation is completed
		time.Sleep(time.Second * 60)
		stepLog = "Verify that pool resize is not in progress"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			dash.VerifyFatal(poolResizeIsInProgress(poolToBeResized), true, fmt.Sprintf("can pool %s expansion start?", poolToBeResized.Uuid))

		})

		var expectedSize uint64
		var expectedSizeWithJournal uint64

		stepLog = "Calculate expected pool size and trigger pool expansion by resize-disk "
		Step(stepLog, func() {
			log.InfoD(stepLog)
			expectedSize = poolToBeResized.TotalSize * 2 / units.GiB

			isjournal, err := isJournalEnabled()
			log.FailOnError(err, "Failed to check is journal enabled")

			//To-Do Need to handle the case for multiple pools
			expectedSizeWithJournal = expectedSize
			if isjournal {
				expectedSizeWithJournal = expectedSizeWithJournal - 3
			}
			log.InfoD("Current Size of the pool %s is %d", poolToBeResized.Uuid, poolToBeResized.TotalSize/units.GiB)
			err = Inst().V.ExpandPool(poolToBeResized.Uuid, api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK, expectedSize)
			log.FailOnError(err, fmt.Sprintf("Pool %s expansion init failed", poolToResize.Uuid))

			resizeErr := waitForPoolToBeResized(expectedSize, poolToResize.Uuid, isjournal)
			dash.VerifyFatal(resizeErr, nil, fmt.Sprintf("Expected new size to be '%d' or '%d'", expectedSize, expectedSizeWithJournal))
		})

		pools, err = Inst().V.ListStoragePools(metav1.LabelSelector{})
		log.FailOnError(err, "error getting pools list")

		poolToBeResized = pools[poolToResize.Uuid]

		stepLog = "Calculate expected pool size and trigger pool expansion by add-disk"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			expectedSize = poolToBeResized.TotalSize * 2 / units.GiB

			isjournal, err := isJournalEnabled()
			log.FailOnError(err, "Failed to check is journal enabled")

			//To-Do Need to handle the case for multiple pools
			expectedSizeWithJournal = expectedSize
			if isjournal {
				expectedSizeWithJournal = expectedSizeWithJournal - 3
			}
			log.InfoD("Current Size of the pool %s is %d", poolToBeResized.Uuid, poolToBeResized.TotalSize/units.GiB)
			err = Inst().V.ExpandPool(poolToBeResized.Uuid, api.SdkStoragePool_RESIZE_TYPE_ADD_DISK, expectedSize)
			log.FailOnError(err, fmt.Sprintf("Pool %s expansion init failed", poolToResize.Uuid))

			resizeErr := waitForPoolToBeResized(expectedSize, poolToResize.Uuid, isjournal)
			dash.VerifyFatal(resizeErr, nil, fmt.Sprintf("Expected new size to be '%d' or '%d'", expectedSize, expectedSizeWithJournal))
		})
		opts := make(map[string]bool)
		opts[scheduler.OptionsWaitForResourceLeakCleanup] = true
		ValidateAndDestroy(contexts, opts)

	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})
})

func addCloudDrive(stNode node.Node, poolID int32) error {
	driveSpecs, err := GetCloudDriveDeviceSpecs()
	if err != nil {
		return fmt.Errorf("error getting cloud drive specs, err: %v", err)
	}
	deviceSpec := driveSpecs[0]
	deviceSpecParams := strings.Split(deviceSpec, ",")
	var specSize uint64
	for _, param := range deviceSpecParams {
		if strings.Contains(param, "size") {
			val := strings.Split(param, "=")[1]
			specSize, err = strconv.ParseUint(val, 10, 64)
			if err != nil {
				return fmt.Errorf("error converting size to uint64, err: %v", err)
			}
		}
	}

	pools, err := Inst().V.ListStoragePools(metav1.LabelSelector{})
	if err != nil {
		return fmt.Errorf("error getting pools list, err: %v", err)
	}
	dash.VerifyFatal(len(pools) > 0, true, "Verify pools exist")

	var currentTotalPoolSize uint64

	for _, pool := range pools {
		currentTotalPoolSize += pool.GetTotalSize() / units.GiB
	}

	log.Info(fmt.Sprintf("current pool size: %d GiB", currentTotalPoolSize))

	expectedTotalPoolSize := currentTotalPoolSize + specSize

	log.InfoD("Initiate add cloud drive and validate")
	err = Inst().V.AddCloudDrive(&stNode, deviceSpec, poolID)
	if err != nil {
		return fmt.Errorf("add cloud drive failed on node %s, err: %v", stNode.Name, err)
	}
	log.InfoD("Validate pool rebalance after drive add")
	err = ValidatePoolRebalance()
	if err != nil {
		return fmt.Errorf("pool re-balance failed, err: %v", err)
	}
	err = Inst().V.WaitDriverUpOnNode(stNode, 5*time.Minute)
	if err != nil {
		return fmt.Errorf("volume driver is down on node %s, err: %v", stNode.Name, err)
	}
	dash.VerifyFatal(err == nil, true, "PX is up after add drive")

	var newTotalPoolSize uint64

	pools, err = Inst().V.ListStoragePools(metav1.LabelSelector{})
	if err != nil {
		return fmt.Errorf("error getting pools list, err: %v", err)
	}
	dash.VerifyFatal(len(pools) > 0, true, "Verify pools exist")
	for _, pool := range pools {
		newTotalPoolSize += pool.GetTotalSize() / units.GiB
	}
	isPoolSizeUpdated := false

	if newTotalPoolSize == expectedTotalPoolSize || newTotalPoolSize == (expectedTotalPoolSize-3) {
		isPoolSizeUpdated = true
	}
	log.Info(fmt.Sprintf("updated pool size: %d GiB", newTotalPoolSize))
	dash.VerifyFatal(isPoolSizeUpdated, true, fmt.Sprintf("Validate total pool size after add cloud drive on node %s", stNode.Name))
	return nil
}
func getVolumeWithMinimumSize(contexts []*scheduler.Context, size uint64) (*volume.Volume, error) {
	var volSelected *volume.Volume
	//waiting till one of the volume has enough IO and selecting pool and node  using the volume to run the test
	f := func() (interface{}, bool, error) {
		for _, ctx := range contexts {
			vols, err := Inst().S.GetVolumes(ctx)
			if err != nil {
				return nil, true, err
			}
			for _, vol := range vols {
				appVol, err := Inst().V.InspectVolume(vol.ID)
				if err != nil {
					return nil, true, err
				}
				usedBytes := appVol.GetUsage()
				usedGiB := usedBytes / units.GiB
				if usedGiB > size {
					volSelected = vol
					return nil, false, nil
				}
			}
		}
		return nil, true, fmt.Errorf("error getting volume with size atleast 10 GiB used")
	}
	_, err := task.DoRetryWithTimeout(f, 15*time.Minute, retryTimeout)
	return volSelected, err
}

var _ = Describe("{PoolResizeMul}", func() {
	//1) Deploy px with cloud drive.
	//2) Select a pool with iops happening.
	//3) Expand pool by adding cloud drives.
	//4) Expand pool again by adding cloud drives.
	//4) Expand pool again by pool expand auto.
	var testrailID = 2019
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/cases/view/2019
	var runID int
	JustBeforeEach(func() {
		StartTorpedoTest("PoolResizeMul", "Initiate pool resize multiple times", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})
	var contexts []*scheduler.Context

	stepLog := "should get the existing storage node and expand the pool multiple times"

	It(stepLog, func() {
		log.InfoD(stepLog)
		contexts = make([]*scheduler.Context, 0)

		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("poolresizemul-%d", i))...)
		}
		ValidateApplications(contexts)

		stNodes := node.GetStorageNodes()
		if len(stNodes) == 0 {
			dash.VerifyFatal(len(stNodes) > 0, true, "Storage nodes found?")
		}
		var selectedNode node.Node
		var err error
		var selectedPool *api.StoragePool
		for _, stNode := range stNodes {
			selectedPool, err = GetPoolWithIOsInGivenNode(stNode)
			if selectedPool != nil {
				selectedNode = stNode
				break
			}
		}
		log.FailOnError(err, "error identifying node to run test")
		stepLog = fmt.Sprintf("Adding drive to the node %s and pool UUID: %s, Id:%d", selectedNode.Name, selectedPool.Uuid, selectedPool.ID)
		Step(stepLog, func() {
			err = addCloudDrive(selectedNode, selectedPool.ID)
			log.FailOnError(err, "error adding cloud drive")
		})
		stepLog = fmt.Sprintf("Adding drive again to the node %s and pool UUID: %s, Id:%d", selectedNode.Name, selectedPool.Uuid, selectedPool.ID)
		Step(stepLog, func() {
			err = addCloudDrive(selectedNode, selectedPool.ID)
			log.FailOnError(err, "error adding cloud drive")
		})

		stepLog = fmt.Sprintf("Expanding pool  on node %s and pool UUID: %s using auto", selectedNode.Name, selectedPool.Uuid)
		Step(stepLog, func() {
			poolToBeResized, err := GetStoragePoolByUUID(selectedPool.Uuid)
			log.FailOnError(err, "Failed to get pool using UUID ")
			expectedSize := poolToBeResized.TotalSize * 2 / units.GiB

			isjournal, err := isJournalEnabled()
			log.FailOnError(err, "Failed to check if Journal enabled")

			log.InfoD("Current Size of the pool %s is %d", selectedPool.Uuid, poolToBeResized.TotalSize/units.GiB)
			err = Inst().V.ExpandPool(selectedPool.Uuid, api.SdkStoragePool_RESIZE_TYPE_AUTO, expectedSize)
			dash.VerifyFatal(err, nil, "Pool expansion init successful?")

			resizeErr := waitForPoolToBeResized(expectedSize, selectedPool.Uuid, isjournal)
			dash.VerifyFatal(resizeErr, nil, fmt.Sprintf("Verify pool %s on node %s expansion using auto", selectedPool.Uuid, selectedNode.Name))
		})
		opts := make(map[string]bool)
		opts[scheduler.OptionsWaitForResourceLeakCleanup] = true
		ValidateAndDestroy(contexts, opts)

	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})
})

var _ = Describe("{MultiDriveResizeDisk}", func() {
	//Select Pool with multiple drives
	//While IO is going onto repl=3 vols on all the pools on that system, Add drive to the pool using ""pxctl sv pool expand-u <uuid> -s <size> -o resize-disk"
	var testrailID = 51266
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/cases/view/51266
	var runID int
	JustBeforeEach(func() {
		StartTorpedoTest("MultiDriveResizeDisk", "Initiate pool resize multiple drive", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})
	var contexts []*scheduler.Context

	stepLog := "should get the existing storage node with multi drives and resize-disk"

	It(stepLog, func() {
		log.InfoD(stepLog)
		var err error
		contexts = make([]*scheduler.Context, 0)

		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("muldrvresize-%d", i))...)
		}
		ValidateApplications(contexts)

		stNodes := node.GetStorageNodes()
		if len(stNodes) == 0 {
			dash.VerifyFatal(len(stNodes) > 0, true, "Storage nodes found?")
		}
		isjournal, err := isJournalEnabled()
		log.FailOnError(err, "Failed to check if Journal enabled")
		minDiskCount := 1
		if isjournal {
			minDiskCount = 2
		}

		nodesWithMultiDrives := make([]node.Node, 0)
		for _, n := range stNodes {
			pxNode, err := Inst().V.GetDriverNode(&n)
			log.FailOnError(err, "Error getting PX node")
			log.Infof("PX node %s , Disks:%+v", pxNode.Hostname, pxNode.Disks)
			if len(pxNode.Disks) > minDiskCount {
				nodesWithMultiDrives = append(nodesWithMultiDrives, n)
			}
		}
		dash.VerifyFatal(len(nodesWithMultiDrives) > 0, true, "nodes with multiple disks exist?")
		var selectedNode node.Node

		var selectedPool *api.StoragePool
		for _, stNode := range nodesWithMultiDrives {
			selectedPool, err = GetPoolWithIOsInGivenNode(stNode)
			if selectedPool != nil {
				selectedNode = stNode
				break
			}
		}
		log.FailOnError(err, "error identifying node to run test")

		stepLog = fmt.Sprintf("Expanding pool  on node %s and pool UUID: %s using resize-disk", selectedNode.Name, selectedPool.Uuid)
		Step(stepLog, func() {
			poolToBeResized, err := GetStoragePoolByUUID(selectedPool.Uuid)
			log.FailOnError(err, "Failed to get pool using UUID ")
			expectedSize := poolToBeResized.TotalSize * 2 / units.GiB

			log.InfoD("Current Size of the pool %s is %d", selectedPool.Uuid, poolToBeResized.TotalSize/units.GiB)
			err = Inst().V.ExpandPool(selectedPool.Uuid, api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK, expectedSize)
			dash.VerifyFatal(err, nil, "Pool expansion init successful?")

			resizeErr := waitForPoolToBeResized(expectedSize, selectedPool.Uuid, isjournal)
			dash.VerifyFatal(resizeErr, nil, fmt.Sprintf("Verify pool %s on node %s expansion using resize-disk", selectedPool.Uuid, selectedNode.Name))
		})
		opts := make(map[string]bool)
		opts[scheduler.OptionsWaitForResourceLeakCleanup] = true
		ValidateAndDestroy(contexts, opts)

	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})
})

var _ = Describe("{ResizeWithPXRestart}", func() {
	//1) Deploy px with cloud drive.
	//2) Create a volume on that pool and write some data on the volume.
	//3) Expand pool by resize-disk
	//4) Restart px service where the pool expansion is in-progress
	var testrailID = 51281
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/cases/view/51281
	var runID int

	JustBeforeEach(func() {
		StartTorpedoTest("ResizeWithPXRestart", "Initiate pool expansion using resize-disk and restart PX while it is in progress", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})
	var contexts []*scheduler.Context

	stepLog := "should get the existing storage node and expand the pool by resize-disk"

	It(stepLog, func() {
		log.InfoD(stepLog)
		contexts = make([]*scheduler.Context, 0)
		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("rsizedrvwrst-%d", i))...)
		}
		ValidateApplications(contexts)

		stNodes := node.GetStorageNodes()
		if len(stNodes) == 0 {
			dash.VerifyFatal(len(stNodes) > 0, true, "Storage nodes found?")
		}
		stNode, err := GetRandomNodeWithPoolIOs(stNodes)
		log.FailOnError(err, "error identifying node to run test")
		selectedPool, err := GetPoolWithIOsInGivenNode(stNode)
		log.FailOnError(err, "error identifying pool to run test")

		stepLog := "Initiate pool expansion drive and restart PX"
		Step(stepLog, func() {
			log.InfoD(stepLog)

			poolToBeResized, err := GetStoragePoolByUUID(selectedPool.Uuid)
			log.FailOnError(err, "Failed to get pool using UUID ")
			expectedSize := poolToBeResized.TotalSize * 2 / units.GiB

			isjournal, err := isJournalEnabled()
			log.FailOnError(err, "Failed to check if Journal enabled")

			log.InfoD("Current Size of the pool %s is %d", selectedPool.Uuid, poolToBeResized.TotalSize/units.GiB)
			err = Inst().V.ExpandPool(selectedPool.Uuid, api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK, expectedSize)
			dash.VerifyFatal(err, nil, "Pool expansion init successful?")

			time.Sleep(3 * time.Second)
			err = Inst().V.RestartDriver(stNode, nil)
			log.FailOnError(err, fmt.Sprintf("error restarting px on node %s", stNode.Name))

			resizeErr := waitForPoolToBeResized(expectedSize, selectedPool.Uuid, isjournal)
			dash.VerifyFatal(resizeErr, nil, fmt.Sprintf("Verify pool %s on node %s expansion using auto", selectedPool.Uuid, stNode.Name))

		})
		opts := make(map[string]bool)
		opts[scheduler.OptionsWaitForResourceLeakCleanup] = true
		ValidateAndDestroy(contexts, opts)

	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})

})

var _ = Describe("{ResizeDiskVolUpdate}", func() {
	//1) Deploy px with cloud drive.
	//2) Create a volume on that pool and write some data on the volume.
	//3) Expand pool by resize-disk.
	//4) expand the volume to the resized pool
	var testrailID = 51290
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/cases/view/51290
	var runID int
	JustBeforeEach(func() {
		StartTorpedoTest("ResizeDiskVolUpdate", "pool expansion using resize-disk and expand volume to the pool", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
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

		stNodes := node.GetStorageNodes()
		if len(stNodes) == 0 {
			dash.VerifyFatal(len(stNodes) > 0, true, "Storage nodes found?")
		}
		volSelected, err := getVolumeWithMinimumSize(contexts, 10)
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
			log.FailOnError(err, "Failed to get pool using UUID ")
			expectedSize := poolToBeResized.TotalSize * 2 / units.GiB

			isjournal, err := isJournalEnabled()
			log.FailOnError(err, "Failed to check if Journal enabled")

			log.InfoD("Current Size of the pool %s is %d", selectedPool.Uuid, poolToBeResized.TotalSize/units.GiB)
			err = Inst().V.ExpandPool(selectedPool.Uuid, api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK, expectedSize)
			dash.VerifyFatal(err, nil, "Pool expansion init successful?")

			resizeErr := waitForPoolToBeResized(expectedSize, selectedPool.Uuid, isjournal)
			dash.VerifyFatal(resizeErr, nil, fmt.Sprintf("Verify pool %s on node %s expansion using resize-disk", selectedPool.Uuid, stNode.Name))

		})
		stepLog = "Expand volume to the expanded pool"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			currRep, err := Inst().V.GetReplicationFactor(volSelected)
			log.FailOnError(err, fmt.Sprintf("err getting repl factor for  vol : %s", volSelected.Name))
			opts := volume.Options{
				ValidateReplicationUpdateTimeout: validateReplicationUpdateTimeout,
			}
			newRep := currRep
			if currRep == 3 {
				newRep = currRep - 1
				err = Inst().V.SetReplicationFactor(volSelected, newRep, nil, nil, true, opts)
				log.FailOnError(err, fmt.Sprintf("err setting repl factor  to %d for  vol : %s", newRep, volSelected.Name))
			}
			log.InfoD(fmt.Sprintf("setting repl factor  to %d for  vol : %s", newRep+1, volSelected.Name))
			err = Inst().V.SetReplicationFactor(volSelected, newRep+1, []string{stNode.Id}, []string{poolToBeResized.Uuid}, true, opts)
			log.FailOnError(err, fmt.Sprintf("err setting repl factor  to %d for  vol : %s", newRep+1, volSelected.Name))
			dash.VerifyFatal(err == nil, true, fmt.Sprintf("vol %s expanded successfully on node %s", volSelected.Name, stNode.Name))
		})

		for _, ctx := range contexts {
			ctx.SkipVolumeValidation = true
			ValidateContext(ctx)
		}
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

var _ = Describe("{VolUpdateResizeDisk}", func() {
	//1) Deploy px with cloud drive.
	//2) Create a volume on that pool and write some data on the volume.
	//3) expand the volume to the pool
	//4) perform resize disk operation on the pool
	var testrailID = 51284
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/cases/view/51284
	var runID int
	JustBeforeEach(func() {
		StartTorpedoTest("VolUpdateResizeDisk", "expand volume to the pool and pool expansion using resize-disk", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})
	var contexts []*scheduler.Context

	stepLog := "should get the existing storage node and expand the pool by resize-disk"

	It(stepLog, func() {
		log.InfoD(stepLog)
		contexts = make([]*scheduler.Context, 0)
		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("volupdtplrsz-%d", i))...)
		}
		//	ValidateApplications(contexts)

		stNodes := node.GetStorageNodes()
		if len(stNodes) == 0 {
			dash.VerifyFatal(len(stNodes) > 0, true, "Storage nodes found?")
		}
		volSelected, err := getVolumeWithMinimumSize(contexts, 10)
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
		log.FailOnError(err, "Failed to get pool using UUID ")

		stepLog = "Expand volume to the expanded pool"
		var newRep int64
		Step(stepLog, func() {
			log.InfoD(stepLog)
			currRep, err := Inst().V.GetReplicationFactor(volSelected)
			log.FailOnError(err, fmt.Sprintf("err getting repl factor for  vol : %s", volSelected.Name))
			opts := volume.Options{
				ValidateReplicationUpdateTimeout: validateReplicationUpdateTimeout,
			}
			newRep = currRep
			if currRep == 3 {
				newRep = currRep - 1
				err = Inst().V.SetReplicationFactor(volSelected, newRep, nil, nil, true, opts)
				log.FailOnError(err, fmt.Sprintf("err setting repl factor  to %d for  vol : %s", newRep, volSelected.Name))
			}
			log.InfoD(fmt.Sprintf("setting repl factor  to %d for  vol : %s", newRep+1, volSelected.Name))
			err = Inst().V.SetReplicationFactor(volSelected, newRep+1, []string{stNode.Id}, []string{poolToBeResized.Uuid}, false, opts)
			log.FailOnError(err, fmt.Sprintf("err setting repl factor  to %d for  vol : %s", newRep+1, volSelected.Name))
			dash.VerifyFatal(err == nil, true, fmt.Sprintf("vol %s expansion triggered successfully on node %s", volSelected.Name, stNode.Name))
		})

		stepLog := "Initiate pool expansion using resize-disk"
		Step(stepLog, func() {
			log.InfoD(stepLog)

			expectedSize := poolToBeResized.TotalSize * 2 / units.GiB

			isjournal, err := isJournalEnabled()
			log.FailOnError(err, "Failed to check if Journal enabled")

			log.InfoD("Current Size of the pool %s is %d", selectedPool.Uuid, poolToBeResized.TotalSize/units.GiB)
			err = Inst().V.ExpandPool(selectedPool.Uuid, api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK, expectedSize)
			dash.VerifyFatal(err, nil, "Pool expansion init successful?")

			resizeErr := waitForPoolToBeResized(expectedSize, selectedPool.Uuid, isjournal)
			dash.VerifyFatal(resizeErr, nil, fmt.Sprintf("Verify pool %s on node %s expansion using resize-disk", selectedPool.Uuid, stNode.Name))

		})
		ValidateReplFactorUpdate(volSelected, newRep+1)

		for _, ctx := range contexts {
			ctx.SkipVolumeValidation = true
			ValidateContext(ctx)
		}
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

var _ = Describe("{VolUpdateAddDrive}", func() {
	//1) Deploy px with cloud drive.
	//2) Create a volume on that pool and write some data on the volume.
	//3) expand the volume to the pool
	//4) perform add drive on the pool
	var testrailID = 50635
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/cases/view/50635
	var runID int
	JustBeforeEach(func() {
		StartTorpedoTest("VolUpdateAddDrive", "expand volume to the pool and pool expansion using add drive", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
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

		stNodes := node.GetStorageNodes()
		if len(stNodes) == 0 {
			dash.VerifyFatal(len(stNodes) > 0, true, "Storage nodes found?")
		}
		volSelected, err := getVolumeWithMinimumSize(contexts, 10)
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
		log.FailOnError(err, "Failed to get pool using UUID ")

		var newRep int64
		Step("Expand volume to the expanded pool", func() {
			currRep, err := Inst().V.GetReplicationFactor(volSelected)
			log.FailOnError(err, fmt.Sprintf("err getting repl factor for  vol : %s", volSelected.Name))
			opts := volume.Options{
				ValidateReplicationUpdateTimeout: validateReplicationUpdateTimeout,
			}
			newRep = currRep
			if currRep == 3 {
				newRep = currRep - 1
				err = Inst().V.SetReplicationFactor(volSelected, newRep, nil, nil, true, opts)
				log.FailOnError(err, fmt.Sprintf("err setting repl factor  to [%d] for  vol : [%s]", newRep, volSelected.Name))
			}
			log.InfoD(fmt.Sprintf("setting repl factor  to [%d] for  vol : [%s]", newRep+1, volSelected.Name))
			err = Inst().V.SetReplicationFactor(volSelected, newRep+1, []string{stNode.Id}, []string{poolToBeResized.Uuid}, false, opts)
			log.FailOnError(err, fmt.Sprintf("err setting repl factor  to [%d] for  vol : [%s]", newRep+1, volSelected.Name))
			dash.VerifyFatal(err == nil, true, fmt.Sprintf("vol [%s] expansion triggered successfully on node [%s]", volSelected.Name, stNode.Name))
		})

		Step("Initiate pool expansion using add drive", func() {
			err = addCloudDrive(stNode, poolToBeResized.ID)
			log.FailOnError(err, "error adding cloud drive")
			dash.VerifyFatal(err == nil, true, fmt.Sprintf("Verify pool [%s] on node [%s] expansion using add drive", poolToBeResized.Uuid, stNode.Name))

		})
		ValidateReplFactorUpdate(volSelected, newRep+1)
		for _, ctx := range contexts {
			ctx.SkipVolumeValidation = true
			ValidateContext(ctx)
		}
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

var _ = Describe("{StoPoolExpMulPools}", func() {
	var testrailID = 51298
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/cases/view/51298
	// testrailID Description : Having multiple pools and resize only one pool
	var runID int
	JustBeforeEach(func() {
		StartTorpedoTest("StoPoolExpMulPools", "Validate storage pool expansion using resize-disk option when multiple pools are present on the cluster", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})

	var contexts []*scheduler.Context

	It("Has to schedule apps, and expand it by resizing a pool", func() {

		contexts = make([]*scheduler.Context, 0)

		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("poolexpand-%d", i))...)
		}
		ValidateApplications(contexts)

		// Get all the storage Nodes present in the system
		stNodes := node.GetStorageNodes()
		if len(stNodes) == 0 {
			dash.VerifyFatal(len(stNodes) > 0, true, "Storage nodes found!")
		}
		log.InfoD("All Storage Nodes present on the kubernetes cluster [%s]", stNodes)

		/* Validate if the Node with Multiple pools are available ,
		if, any node has multiple pools present , then use that Node for expanding
		else, Fail the test case
		*/
		var selectedNode node.Node
		isMultiPoolNode := false
		for _, selNode := range stNodes {
			log.InfoD("Validating Node [%s] for multipool configuraitons", selNode.Name)
			if len(selNode.StoragePools) > 1 {
				isMultiPoolNode = true
				selectedNode = selNode
				break
			}
		}

		dash.VerifyFatal(isMultiPoolNode, true, "Failed as Multipool configuration doesnot exists!")

		// Selecting Storage pool based on Pools present on the Node with IO running
		selectedPool, err := GetPoolWithIOsInGivenNode(selectedNode)
		log.FailOnError(err, "error while selecting the pool [%s]", selectedPool)

		stepLog := fmt.Sprintf("Expanding pool on node [%s] and pool UUID: [%s] using auto", selectedNode.Name, selectedPool.Uuid)
		Step(stepLog, func() {
			poolToBeResized, err := GetStoragePoolByUUID(selectedPool.Uuid)
			log.FailOnError(err, "Failed to get pool using UUID [%s]", selectedPool.Uuid)
			expectedSize := poolToBeResized.TotalSize * 2 / units.GiB

			isjournal, err := isJournalEnabled()
			log.FailOnError(err, "Failed to check if Journal enabled")

			log.InfoD("Current Size of the pool %s is %d", selectedPool.Uuid, poolToBeResized.TotalSize/units.GiB)
			err = Inst().V.ExpandPool(selectedPool.Uuid, api.SdkStoragePool_RESIZE_TYPE_AUTO, expectedSize)
			dash.VerifyFatal(err, nil, "Pool expansion init successful?")

			resizeErr := waitForPoolToBeResized(expectedSize, selectedPool.Uuid, isjournal)
			dash.VerifyFatal(resizeErr, nil, fmt.Sprintf("Verify pool [%s] on node [%s] expansion using auto", selectedPool.Uuid, selectedNode.Name))
		})

		opts := make(map[string]bool)
		opts[scheduler.OptionsWaitForResourceLeakCleanup] = true
		ValidateAndDestroy(contexts, opts)

		Step("destroy all the applications created before test runs", func() {
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

var _ = Describe("{CreateSnapshotsPoolResize}", func() {
	var testrailID = 50652
	// Testrail Description : Try pool resize when lot of snapshots are created on the volume
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/cases/view/50652
	var runID int

	JustBeforeEach(func() {
		StartTorpedoTest("CreateSnapshotsPoolResize", "Validate storage pool expansion when lots of snapshots present on the system", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})

	var contexts []*scheduler.Context
	var totalSnapshotsPerVol int = 60

	snapshotList := make(map[string][]string)
	var selectedNode node.Node
	var pickNode string

	// Try pool resize when ot of snapshots are created on the volume
	stepLog := "should get the existing storage node and expand the pool by resize-disk"
	It(stepLog, func() {

		contexts = make([]*scheduler.Context, 0)

		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("snapcreateresizepool-%d", i))...)
		}
		ValidateApplications(contexts)

		var stNode node.Node

		// Get List of Volumes presnet in the Node

		stNodes := node.GetStorageNodes()
		if len(stNodes) == 0 {
			dash.VerifyFatal(len(stNodes) > 0, true, "No Storage Node exists !! ")
		}
		log.InfoD("List of Nodes present [%s]", stNodes)

		for _, each := range contexts {
			log.InfoD("Getting context Info [%v]", each)
			Volumes, err := Inst().S.GetVolumes(each)
			log.FailOnError(err, "Listing Volumes Failed ")

			log.InfoD("Get all the details of Volumes Present")
			for _, vol := range Volumes {
				log.InfoD("List of Volumes to inspect [%T] , [%s]", vol, vol.ID)
				volInspect, err := Inst().V.InspectVolume(vol.ID)
				log.FailOnError(err, "Failed to Inpect volumes present Err : [%s]", volInspect)
				selectedNode := volInspect.ReplicaSets[0].Nodes
				randomIndex := rand.Intn(len(selectedNode))
				pickNode = selectedNode[randomIndex]

				for _, n := range stNodes {
					if n.Id == pickNode {
						stNode = n
					}
				}
				for snap := 0; snap < totalSnapshotsPerVol; snap++ {
					uuidCreated := uuid.New()
					snapshotName := fmt.Sprintf("snapshot_%s_%s", vol.ID, uuidCreated.String())
					snapshotResponse, err := Inst().V.CreateSnapshot(vol.ID, snapshotName)
					log.FailOnError(err, "error identifying volume [%s]", vol.ID)
					snapshotList[vol.ID] = append(snapshotList[vol.ID], snapshotName)
					log.InfoD("Snapshot [%s] created with ID [%s]", snapshotName, snapshotResponse.GetSnapshotId())
				}
				break

			}
		}

		// Selecting Storage pool based on Pools present on the Node
		selectedPool, err := GetPoolWithIOsInGivenNode(stNode)
		log.FailOnError(err, "error identifying pool running IO [%s]", stNode.Name)

		stepLog = fmt.Sprintf("Expanding pool on node [%s] and pool UUID: [%s] using auto", selectedNode.Name, selectedPool.Uuid)
		Step(stepLog, func() {
			poolToBeResized, err := GetStoragePoolByUUID(selectedPool.Uuid)
			log.FailOnError(err, "Failed to get pool using UUID [%s]", selectedPool.Uuid)
			expectedSize := poolToBeResized.TotalSize * 2 / units.GiB

			isjournal, err := isJournalEnabled()
			log.FailOnError(err, "Failed to check if Journal enabled")

			log.InfoD("Current Size of the pool [%s] is [%d]", selectedPool.Uuid, poolToBeResized.TotalSize/units.GiB)
			err = Inst().V.ExpandPool(selectedPool.Uuid, api.SdkStoragePool_RESIZE_TYPE_AUTO, expectedSize)
			dash.VerifyFatal(err, nil, "Pool expansion init successful?")

			resizeErr := waitForPoolToBeResized(expectedSize, selectedPool.Uuid, isjournal)
			dash.VerifyFatal(resizeErr, nil, fmt.Sprintf("Verify pool [%s] on node [%s] expansion using auto", selectedPool.Uuid, selectedNode.Name))
		})

		opts := make(map[string]bool)
		opts[scheduler.OptionsWaitForResourceLeakCleanup] = true
		ValidateAndDestroy(contexts, opts)

		Step("destroy all the applications created before test runs", func() {
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

func unique(arrayEle []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range arrayEle {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func inResync(vol string) bool {
	volDetails, err := Inst().V.InspectVolume(vol)
	if err != nil {
		log.Error("not in Resync State")
		return false
	}
	for _, v := range volDetails.RuntimeState {
		log.InfoD("RuntimeState is in state %s", v.GetRuntimeState()["RuntimeState"])
		if v.GetRuntimeState()["RuntimeState"] != "resync" {
			return false
		}
	}
	return true
}

func WaitTillVolumeInResync(vol string) bool {
	now := time.Now()
	targetTime := now.Add(30 * time.Minute)

	for {
		if now.After(targetTime) {
			log.Error("Failed as the timeout of 0 Min is reached before resync triggered ")
			return false
		} else {
			if inResync(vol) {
				return true
			}
		}
	}
}

var _ = Describe("{PoolResizeVolumesResync}", func() {
	var testrailID = 51301
	// Testrail Description : Try pool resize when lot of volumes are in resync state
	// Testrail Corresponds : https://portworx.testrail.net/index.php?/cases/view/51301
	var runID int

	JustBeforeEach(func() {
		StartTorpedoTest("PoolResizeVolumesResync", "Validate Pool resize when lots of volumes are in resync state", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})

	var contexts []*scheduler.Context
	var vol_ids []string

	It("should get the existing storage node and expand the pool by resize-disk", func() {
		contexts = make([]*scheduler.Context, 0)

		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("snapcreateresizepool-%d", i))...)
		}
		ValidateApplications(contexts)

		time.Sleep(5 * time.Second)
		for _, each := range contexts {
			Volumes, err := Inst().S.GetVolumes(each)
			log.FailOnError(err, "Failed while listing the volume with error [%s]", err)

			// Appending all the volume IDs to array so that one random volume can be picked for resizeing
			for _, vol := range Volumes {
				vol_ids = append(vol_ids, vol.ID)
			}

			// Select Random Volumes for pool Expand
			randomIndex := rand.Intn(len(vol_ids))
			randomVolIDs := vol_ids[randomIndex]

			// From each volume pick the random pool and restart pxdriver
			poolUUIDs, err := GetPoolIDsFromVolName(randomVolIDs)
			log.InfoD("List of pool IDs %v", poolUUIDs)
			log.FailOnError(err, "Failed to get Pool IDs from the volume [%s]", poolUUIDs)

			// Select the random pools from UUIDs for PxDriver Restart
			randomIndex = rand.Intn(len(poolUUIDs))
			rebootPoolID := poolUUIDs[randomIndex]

			// Rebooting Node
			log.InfoD("Get the Node for Restart %v", rebootPoolID)
			restartDriver, err := GetNodeWithGivenPoolID(rebootPoolID)
			log.FailOnError(err, "Geting Node Driver for restart failed")

			isjournal, err := isJournalEnabled()
			log.FailOnError(err, "Failed to check if Journal enabled")

			poolToBeResized, err := GetStoragePoolByUUID(rebootPoolID)
			log.InfoD("Pool to be resized %v", poolToBeResized)
			log.FailOnError(err, "Failed to get pool using UUID [%s]", rebootPoolID)
			expectedSize := poolToBeResized.TotalSize * 2 / units.GiB

			log.InfoD("Restarting the Driver on Node [%s]", restartDriver.Name)
			err = Inst().N.RebootNode(*restartDriver, node.RebootNodeOpts{
				Force: true,
				ConnectionOpts: node.ConnectionOpts{
					Timeout:         1 * time.Minute,
					TimeBeforeRetry: 5 * time.Second,
				},
			})
			log.FailOnError(err, "Rebooting Node failed ?")

			log.InfoD("Waiting till Volume is In Resync Mode ")
			if WaitTillVolumeInResync(randomVolIDs) == false {
				log.InfoD("Failed to get Volume in Resync state [%s] ", randomVolIDs)
			}

			log.InfoD("Current Size of the pool %s is %d", rebootPoolID, poolToBeResized.TotalSize/units.GiB)
			err = Inst().V.ExpandPool(rebootPoolID, api.SdkStoragePool_RESIZE_TYPE_AUTO, expectedSize)
			dash.VerifyFatal(err, nil, "Pool expansion init successful?")

			resizeErr := waitForPoolToBeResized(expectedSize, rebootPoolID, isjournal)
			dash.VerifyFatal(resizeErr, nil, fmt.Sprintf("Verify pool [%s] on node [%s] expansion using auto", rebootPoolID, restartDriver.Name))

		}
		opts := make(map[string]bool)
		opts[scheduler.OptionsWaitForResourceLeakCleanup] = true
		ValidateAndDestroy(contexts, opts)

		Step("destroy all the applications created before test runs", func() {
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

var _ = Describe("{PoolIncreaseSize20TB}", func() {
	var testrailID = 51292
	// Testrail Description : Resize a pool of capacity of 100GB to 20TB
	// Testrail Corresponds : https://portworx.testrail.net/index.php?/cases/view/51292
	var runID int

	JustBeforeEach(func() {
		StartTorpedoTest("PoolIncreaseSize20TB", "Resize a pool of capacity of 100GB to 20TB", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})

	var contexts []*scheduler.Context
	//var vol_ids []string
	stepLog := "should get the existing storage node and expand the pool by resize-disk"
	It(stepLog, func() {

		contexts = make([]*scheduler.Context, 0)

		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("snapcreateresizepool-%d", i))...)
		}
		ValidateApplications(contexts)

		pools, err := Inst().V.ListStoragePools(metav1.LabelSelector{})
		log.FailOnError(err, "Failed to list storage pools")
		dash.VerifyFatal(len(pools) > 0, true, "Storage pools exist ?")

		// pick a pool from a pools list and resize it
		poolIDToResize, err := GetPoolIDWithIOs()
		log.FailOnError(err, "error identifying pool to run test")
		dash.VerifyFatal(len(poolIDToResize) > 0, true, fmt.Sprintf("Expected poolIDToResize to not be empty, pool id to resize [%s]", poolIDToResize))

		poolToBeResized := pools[poolIDToResize]
		dash.VerifyFatal(poolToBeResized != nil, true, "Pool to be resized exist?")

		// px will put a new request in a queue, but in this case we can't calculate the expected size,
		// so need to wain until the ongoing operation is completed
		stepLog = "Verify that pool resize is not in progress"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			if poolResizeIsInProgress(poolToBeResized) {
				// wait until resize is completed and get the updated pool again
				poolToBeResized, err = GetStoragePoolByUUID(poolIDToResize)
				log.FailOnError(err, "Failed to get pool using UUID [%s]", poolIDToResize)
			}
		})

		var expectedSize uint64
		var expectedSizeWithJournal uint64

		// Marking the expected size to be 2TB
		expectedSize = (2048 * 1024 * 1024 * 1024 * 1024) / units.TiB

		stepLog = "Calculate expected pool size and trigger pool resize"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			isjournal, err := isJournalEnabled()
			log.FailOnError(err, "Failed to check is Journal enabled")

			//To-Do Need to handle the case for multiple pools
			expectedSizeWithJournal = expectedSize
			if isjournal {
				expectedSizeWithJournal = expectedSizeWithJournal - 3
			}
			err = Inst().V.ExpandPool(poolIDToResize, api.SdkStoragePool_RESIZE_TYPE_ADD_DISK, expectedSize)
			dash.VerifyFatal(err, nil, "Pool expansion init successful?")

			resizeErr := waitForPoolToBeResized(expectedSize, poolIDToResize, isjournal)
			dash.VerifyFatal(resizeErr, nil, fmt.Sprintf("Expected new size to be [%d] or [%d] if pool has journal", expectedSize, expectedSizeWithJournal))
		})

		Step("Ensure that new pool has been expanded to the expected size", func() {
			ValidateApplications(contexts)

			resizedPool, err := GetStoragePoolByUUID(poolIDToResize)
			log.FailOnError(err, "Failed to get pool using UUID [%s]", poolIDToResize)
			newPoolSize := resizedPool.TotalSize / units.GiB
			isExpansionSuccess := false
			if newPoolSize >= expectedSizeWithJournal {
				isExpansionSuccess = true
			}
			dash.VerifyFatal(isExpansionSuccess, true,
				fmt.Sprintf("expected new pool size to be [%v] or [%v] if pool has journal, got [%v]", expectedSize, expectedSizeWithJournal, newPoolSize))
		})

		opts := make(map[string]bool)
		opts[scheduler.OptionsWaitForResourceLeakCleanup] = true
		ValidateAndDestroy(contexts, opts)

		Step("destroy all the applications created before test runs", func() {
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

func addDiskToSpecificPool(node node.Node, sizeOfDisk uint64, poolID int32) bool {
	// Get the Spec to add the disk to the Node
	//  if the diskSize ( sizeOfDisK ) is 0 , then Disk of default spec size will be picked
	driveSpecs, err := GetCloudDriveDeviceSpecs()
	log.FailOnError(err, fmt.Sprintf("Error getting cloud drive specs : [%v]", err))
	log.InfoD("Cloud Drive Spec %s", driveSpecs)

	// Update the device spec to update the disk size
	deviceSpec := driveSpecs[0]
	deviceSpecParams := strings.Split(deviceSpec, ",")
	paramsArr := make([]string, 0)
	for _, param := range deviceSpecParams {
		if strings.Contains(param, "size") {
			if sizeOfDisk == 0 {
				var specSize uint64
				val := strings.Split(param, "=")[1]
				specSize, err = strconv.ParseUint(val, 10, 64)
				log.FailOnError(err, "Error converting size [%v] to uint64", val)
				paramsArr = append(paramsArr, fmt.Sprintf("size=%d,", specSize))
			} else {
				paramsArr = append(paramsArr, fmt.Sprintf("size=%d", sizeOfDisk))
			}
		} else {
			paramsArr = append(paramsArr, param)
		}
	}
	newSpec := strings.Join(paramsArr, ",")
	log.InfoD("New Spec Details %v", newSpec)

	// Add Drive to the Volume
	err = Inst().V.AddCloudDrive(&node, newSpec, poolID)
	if err != nil {
		// Regex to check if the error message is reported
		re := regexp.MustCompile(`Drive not compatible with specified pool.*`)
		if re.MatchString(fmt.Sprintf("%v", err)) {
			log.InfoD("Error while adding Disk %v", err)
			return false
		}
	}
	return true
}

var _ = Describe("{ResizePoolDrivesInDifferentSize}", func() {
	var testrailID = 51320
	// Testrail Description : Resizing the pool should fail when drives in the pool have been resized to different size
	// Testrail Corresponds : https://portworx.testrail.net/index.php?/cases/view/51320
	var runID int

	JustBeforeEach(func() {
		StartTorpedoTest("ResizePoolDrivesInDifferentSize",
			"Resizing the pool should fail when drives in the pool have been resized to different size",
			nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})

	var contexts []*scheduler.Context
	It("should get the existing storage node and expand the pool by resize-disk", func() {

		contexts = make([]*scheduler.Context, 0)
		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("resizepooldrivesdiffsize-%d", i))...)
		}
		ValidateApplications(contexts)

		// Select a Pool with IO Runing poolID returns UUID ( String )
		var poolID int32

		poolUUID, err := GetPoolIDWithIOs()
		log.InfoD("Pool UUID on which IO is running [%s]", poolUUID)
		log.FailOnError(err, "Failed to get pool using UUID [%v]", poolID)

		allPools, _ := Inst().V.ListStoragePools(metav1.LabelSelector{})
		log.InfoD("List of all the Pools present in the system [%s]", allPools)

		// Get Pool ID of pool selected for Resize
		for uuid, each := range allPools {
			if uuid == poolUUID {
				poolID = each.ID
				break
			}

		}
		log.InfoD("Getting Pool with ID [%v] and UUID [%v] for Drive Addition", poolID, poolUUID)

		// Get the Node from the PoolID (nodeDetails returns node.Node)
		nodeDetails, err := GetNodeWithGivenPoolID(poolUUID)
		log.FailOnError(err, "Getting NodeID from the given poolUUID [%v] Failed", poolUUID)
		log.InfoD("Node Details %v", nodeDetails)

		// Add disk to the Node
		var diskSize uint64
		minDiskSize := 50
		maxDiskSize := 150
		size := rand.Intn(maxDiskSize-minDiskSize) + minDiskSize
		diskSize = (uint64(size) * 1024 * 1024 * 1024) / units.GiB

		log.InfoD("Adding New Disk with Size [%v]", diskSize)
		response := addDiskToSpecificPool(*nodeDetails, diskSize, poolID)
		dash.VerifyFatal(response, false,
			fmt.Sprintf("Pool expansion with Disk Resize with Disk size [%v GiB] Succeeded?", diskSize))

		log.InfoD("Attempt Adding Disk with size same as pool size")
		response = addDiskToSpecificPool(*nodeDetails, 0, poolID)
		dash.VerifyFatal(response, true,
			fmt.Sprintf("Pool expansion with Disk size same as pool size [%v GiB] Succeeded?", diskSize))

		opts := make(map[string]bool)
		opts[scheduler.OptionsWaitForResourceLeakCleanup] = true
		ValidateAndDestroy(contexts, opts)

		Step("destroy all the applications created before test runs", func() {
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
