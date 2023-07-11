package tests

import (
	"fmt"
	"github.com/portworx/torpedo/pkg/log"

	"strings"
	"time"

	"github.com/libopenstorage/openstorage/api"
	. "github.com/onsi/ginkgo"
	"github.com/portworx/sched-ops/task"
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/pkg/units"
	. "github.com/portworx/torpedo/tests"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	poolResizeTimeout                = time.Minute * 90
	poolExpansionStatusCheckInterval = time.Minute * 4
)

var _ = Describe("{StoragePoolExpandDiskResize}", func() {
	var contexts []*scheduler.Context
	var poolIDToResize string

	JustBeforeEach(func() {
		StartTorpedoTest("StoragePoolExpandDiskResize", "Validate storage pool expansion using resize-disk option", nil, 0)
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts)
	})

	It("should schedule apps, validate them, and expand the pool by resizing a disk", func() {
		contexts = initializeContexts()
		defer appsValidateAndDestroy(contexts)

		poolIDToResize = pickPoolToResize(contexts)
		poolToBeResized := ensurePoolExists(poolIDToResize)

		waitForPoolToBeReadyForExpansion(poolToBeResized)

		desiredSize := getDesiredSize(poolToBeResized.TotalSize)

		log.InfoD("Current Size of the pool %s is %d GB. Trying to expand to %v GB",
			poolIDToResize, poolToBeResized.TotalSize/units.GiB, desiredSize/units.GiB)

		triggerPoolExpansion(poolIDToResize, desiredSize, api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK)

		waitForOngoingPoolExpansionToComplete(poolIDToResize)

		verifyPoolSizeEqualOrLargerThanExpected(poolIDToResize, desiredSize)
	})

})

var _ = Describe("{StoragePoolExpandDiskAdd}", func() {
	var contexts []*scheduler.Context
	var poolIDToResize string

	JustBeforeEach(func() {
		StartTorpedoTest("StoragePoolExpandDiskAdd", "Validate storage pool expansion using add-disk option", nil, 0)
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts)
	})

	It("Should expand an existing pool by adding a disk", func() {
		contexts = initializeContexts()
		defer appsValidateAndDestroy(contexts)

		// pick a pool from a pools list to resize
		poolIDToResize = pickPoolToResize(contexts)
		poolToBeResized := ensurePoolExists(poolIDToResize)

		waitForPoolToBeReadyForExpansion(poolToBeResized)

		originalSize := poolToBeResized.TotalSize
		expectedSize := getDesiredSize(originalSize)
		triggerPoolExpansion(poolIDToResize, expectedSize, api.SdkStoragePool_RESIZE_TYPE_ADD_DISK)

		resizeErr := waitForOngoingPoolExpansionToComplete(poolIDToResize)
		dash.VerifyFatal(resizeErr, nil, "Pool expansion resulted in error")

		verifyPoolSizeEqualOrLargerThanExpected(poolIDToResize, expectedSize)
	})
})

func verifyPoolSizeEqualOrLargerThanExpected(poolIDToResize string, expectedSize uint64) {
	Step("Verify that pool has been expanded to the expected size", func() {
		resizedPool, err := GetStoragePoolByUUID(poolIDToResize)
		failOnError(err, "Failed to get pool using UUID %s", poolIDToResize)
		newPoolSize := resizedPool.TotalSize / units.GiB
		dash.VerifyFatal(newPoolSize >= expectedSize, true,
			fmt.Sprintf("Expected pool to have been expanded to %v, but got %v", expectedSize, newPoolSize))
	})
}

func triggerPoolExpansion(poolIDToResize string, expectedSize uint64, expandType api.SdkStoragePool_ResizeOperationType) {
	stepLog := "Trigger pool expansion"
	Step(stepLog, func() {
		log.InfoD(stepLog)
		err := Inst().V.ExpandPool(poolIDToResize, expandType, expectedSize, true)
		dash.VerifyFatal(err, nil, "Failed to init pool expansion")
	})
}

func waitForExistingExpansionToFinish(pool *api.StoragePool) {
	stepLog := "Verify that pool resize is not in progress"
	poolIDToResize := pool.GetUuid()
	Step(stepLog, func() {
		log.InfoD(stepLog)
		if val, err := waitForPoolToBeReadyForExpansion(pool); val {
			// wait until resize is completed and get the updated pool again
			pool, err = GetStoragePoolByUUID(poolIDToResize)
			failOnError(err, "failed to get pool using UUID %s", poolIDToResize)
		} else {
			failOnError(err, "pool %s cannot be expanded: %v", poolIDToResize, err)
		}
	})

}

func getDesiredSize(originalSize uint64) uint64 {
	expectedSize := roundUpValue(originalSize * 2 / units.GiB)
	isjournal, err := isJournalEnabled()
	failOnError(err, "Failed to check if Journal is enabled")
	if isjournal {
		expectedSize -= 3
	}
	return expectedSize
}

func ensurePoolExists(poolIDToResize string) *api.StoragePool {
	pool, err := GetStoragePoolByUUID(poolIDToResize)
	failOnError(err, "Failed to get pool using UUID %s", poolIDToResize)
	dash.VerifyFatal(pool != nil, true, "failed to find pool to resize")
	return pool
}
func pickPoolToResize(contexts []*scheduler.Context) string {
	poolIDToResize, err := GetPoolIDWithIOs(contexts)
	failOnError(err, "Error identifying pool to run test")
	verifyNonEmpty(poolIDToResize, "Expected poolIDToResize to not be empty, pool id to resize %s", poolIDToResize)
	return poolIDToResize
}

func verifyNonEmpty(value string, message string, args ...interface{}) {
	dash.VerifyFatal(len(value) > 0, true, message)
}

func failOnError(err error, message string, args ...interface{}) {
	if err != nil {
		log.FailOnError(err, message, args...)
	}
}

func initializeContexts() []*scheduler.Context {
	contexts := make([]*scheduler.Context, 0)
	for i := 0; i < Inst().GlobalScaleFactor; i++ {
		contexts = append(contexts, ScheduleApplications(fmt.Sprintf("pooladddisk-%d", i))...)
	}
	ValidateApplications(contexts)
	return contexts
}

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
		defer appsValidateAndDestroy(contexts)

		var poolIDToResize string

		pools, err := Inst().V.ListStoragePools(metav1.LabelSelector{})
		log.FailOnError(err, "Failed to list storage pools")
		dash.VerifyFatal(len(pools) > 0, true, " Storage pools exist?")

		// pick a pool from a pools list and resize it
		poolIDToResize, err = GetPoolIDWithIOs(contexts)
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
			if val, err := waitForPoolToBeReadyForExpansion(poolToBeResized); val {
				// wait until resize is completed and get the updated pool again
				poolToBeResized, err = GetStoragePoolByUUID(poolIDToResize)
				log.FailOnError(err, fmt.Sprintf("Failed to get pool using UUID %s", poolIDToResize))
			} else {
				log.FailOnError(err, fmt.Sprintf("pool [%s] cannot be expanded due to error: %v", poolIDToResize, err))
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
			err = Inst().V.ExpandPool(poolIDToResize, api.SdkStoragePool_RESIZE_TYPE_AUTO, expectedSize, false)
			dash.VerifyFatal(err, nil, "Pool expansion init successful?")

			resizeErr := waitForOngoingPoolExpansionToComplete(poolIDToResize)
			dash.VerifyFatal(resizeErr, nil, fmt.Sprintf("Expected new size to be '%d' or '%d'", expectedSize, expectedSizeWithJournal))
		})

		stepLog = "Ensure that new pool has been expanded to the expected size"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			ValidateApplications(contexts)

			resizedPool, err := GetStoragePoolByUUID(poolIDToResize)
			log.FailOnError(err, fmt.Sprintf("Failed to get pool using UUID %s", poolIDToResize))
			newPoolSize := resizedPool.TotalSize / units.GiB
			isExpansionSuccess := false
			if newPoolSize >= expectedSizeWithJournal {
				isExpansionSuccess = true
			}
			dash.VerifyFatal(isExpansionSuccess, true, fmt.Sprintf("Expected new pool size to be %v or %v, got %v", expectedSize, expectedSizeWithJournal, newPoolSize))

		})

	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts)
	})
})

func roundUpValue(toRound uint64) uint64 {
	if toRound%10 == 0 {
		return toRound
	}
	rs := (10 - toRound%10) + toRound
	return rs
}

// Wait for existing pool expansion to complete and pool status	to be online
func waitForPoolToBeReadyForExpansion(poolToBeResized *api.StoragePool) (bool, error) {
	if err := waitForPoolToBeOnline(poolToBeResized.GetUuid()); err != nil {
		return false, err
	}

	if err := waitForOngoingPoolExpansionToComplete(poolToBeResized.GetUuid()); err != nil {
		return false, err
	}

	return true, nil
}

func waitForPoolToBeOnline(poolID string) error {
	stNode, err := GetNodeWithGivenPoolID(poolID)
	if err != nil {
		return err
	}

	t := func() (interface{}, bool, error) {
		// status check
		status, err := Inst().V.GetNodePoolsStatus(*stNode)
		if err != nil {
			return "", false, err
		}
		currStatus := status[poolID]

		if currStatus == "Offline" {
			return "", true, fmt.Errorf("pool [%s] is offline [%s]. Waiting for pool to come up", poolID, currStatus)
		}
		return "", true, nil
	}

	_, err = task.DoRetryWithTimeout(t, 90*time.Minute, 30*time.Second)
	return nil
}

func waitForOngoingPoolExpansionToComplete(poolIDToResize string) error {
	currentLastMsg := ""
	f := func() (interface{}, bool, error) {
		expandedPool, err := GetStoragePoolByUUID(poolIDToResize)
		if err != nil {
			return nil, true, fmt.Errorf("error getting pool by using id %s", poolIDToResize)
		}
		if expandedPool == nil {
			return nil, false, fmt.Errorf("pool to expand not found")
		}
		if expandedPool.LastOperation == nil {
			return nil, false, fmt.Errorf("no pool resize operation in progress")
		}
		log.Infof("Pool Resize Status: %v, Message : %s", expandedPool.LastOperation.Status, expandedPool.LastOperation.Msg)
		switch expandedPool.LastOperation.Status {
		case api.SdkStoragePool_OPERATION_SUCCESSFUL:
			return nil, false, nil
		case api.SdkStoragePool_OPERATION_FAILED:
			return nil, false, fmt.Errorf("pool %s expansion failed: %s", poolIDToResize, expandedPool.LastOperation)
		case api.SdkStoragePool_OPERATION_PENDING:
			return nil, true, fmt.Errorf("pool %s expansion is pending", poolIDToResize)
		case api.SdkStoragePool_OPERATION_IN_PROGRESS:
			if strings.Contains(expandedPool.LastOperation.Msg, "Rebalance in progress") {
				if currentLastMsg == expandedPool.LastOperation.Msg {
					return nil, true, fmt.Errorf("pool rebalance is not progressing")
				}
				currentLastMsg = expandedPool.LastOperation.Msg
				return nil, true, fmt.Errorf("wait for pool rebalance to complete")
			}
			fallthrough
		default:
			return nil, true, fmt.Errorf("waiting for pool status to update")
		}
	}

	_, err := task.DoRetryWithTimeout(f, poolResizeTimeout, poolExpansionStatusCheckInterval)
	return err
}

func getPoolLastOperation(poolID string) (*api.StoragePoolOperation, error) {
	log.Infof(fmt.Sprintf("Getting pool status for %s", poolID))
	f := func() (interface{}, bool, error) {
		pool, err := GetStoragePoolByUUID(poolID)
		if err != nil {
			return nil, true, fmt.Errorf("error getting pool by using id %s", poolID)
		}

		if pool == nil {
			return nil, false, fmt.Errorf("pool value is nil")
		}
		if pool.LastOperation != nil {
			return pool.LastOperation, false, nil
		}
		return nil, true, fmt.Errorf("pool status not updated")
	}

	var poolLastOperation *api.StoragePoolOperation
	poolStatus, err := task.DoRetryWithTimeout(f, poolResizeTimeout, poolExpansionStatusCheckInterval)
	if err != nil {
		return nil, err
	}
	poolLastOperation = poolStatus.(*api.StoragePoolOperation)
	return poolLastOperation, err
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
func appsValidateAndDestroy(contexts []*scheduler.Context) {
	opts := make(map[string]bool)
	opts[scheduler.OptionsWaitForResourceLeakCleanup] = true

	Step("validate apps", func() {
		log.InfoD("Validating apps")
		for _, ctx := range contexts {
			ctx.ReadinessTimeout = 15 * time.Minute
			ValidateContext(ctx)
		}
	})

	Step("destroy apps", func() {
		log.InfoD("Destroying apps")
		for _, ctx := range contexts {
			TearDownContext(ctx, opts)
		}
	})
}
