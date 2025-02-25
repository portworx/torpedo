package tests

import (
	"fmt"
	"math/rand"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	storkv1 "github.com/pure-px/stork/pkg/apis/stork/v1alpha1"
	storkops "github.com/pure-px/stork/pkg/crud/stork"
	"github.com/pure-px/torpedo/drivers/scheduler/k8s"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/google/uuid"
	"github.com/libopenstorage/openstorage/api"
	opsapi "github.com/libopenstorage/openstorage/api"
	. "github.com/onsi/ginkgo/v2"
	"github.com/pure-px/sched-ops/task"
	"github.com/pure-px/torpedo/drivers/node"
	"github.com/pure-px/torpedo/drivers/scheduler"
	"github.com/pure-px/torpedo/drivers/volume"
	"github.com/pure-px/torpedo/pkg/log"
	"github.com/pure-px/torpedo/pkg/testrailuttils"
	"github.com/pure-px/torpedo/pkg/units"
	. "github.com/pure-px/torpedo/tests"
)

var (
	stepLog       string
	runID         int
	testrailID    int
	targetSizeGiB uint64
	storageNode   *node.Node
	err           error
)
var _ = Describe("{PoolExpandMultipleTimes}", Label("p0", "positive", "pool_ops", "px_ops", "PoolExpand", "AddDrive"), func() {

	var (
		poolToResize   *api.StoragePool
		poolIDToResize string
		contexts       = make([]*scheduler.Context, 0)
	)
	BeforeEach(func() {
		contexts = scheduleApps()
	})

	JustBeforeEach(func() {
		poolIDToResize = pickPoolToResize(contexts, api.SdkStoragePool_RESIZE_TYPE_ADD_DISK, 300)
		log.Infof("Picked pool %s to resize", poolIDToResize)
		poolToResize = getStoragePool(poolIDToResize)
	})

	JustAfterEach(func() {
		AfterEachTest(contexts)
	})

	AfterEach(func() {
		appsValidateAndDestroy(contexts)
		EndTorpedoTest()
	})

	It("Select a pool and expand it by 100 GiB 3 time with add-disk type. ", func() {
		// TestRail:https://portworx.testrail.net/index.php?/tests/view/86355092
		StartTorpedoTest("PoolExpandDiskAdd3Times",
			"Validate storage pool expansion 3 times with type=add-disk", nil, 86355092)
		for i := 0; i < 3; i++ {
			poolToResize = getStoragePool(poolIDToResize)
			originalSizeInBytes = poolToResize.TotalSize
			targetSizeInBytes = originalSizeInBytes + 100*units.GiB
			targetSizeGiB = targetSizeInBytes / units.GiB

			log.InfoD("Current Size of pool %s is %d GiB. Expand to %v GiB with type add-disk...",
				poolIDToResize, poolToResize.TotalSize/units.GiB, targetSizeGiB)
			triggerPoolExpansion(poolIDToResize, targetSizeGiB, api.SdkStoragePool_RESIZE_TYPE_ADD_DISK)
			resizeErr := waitForOngoingPoolExpansionToComplete(poolIDToResize)

			if isDMthin, _ := IsDMthin(); isDMthin {
				dash.VerifyFatal(resizeErr != nil, true,
					"Pool expansion request of add-disk type should be rejected with dmthin")
			} else {
				dash.VerifyFatal(resizeErr, nil, "Pool expansion does not result in error")
				verifyPoolSizeEqualOrLargerThanExpected(poolIDToResize, targetSizeGiB)
			}
		}
	})

	It("Select a pool and expand it by 100 GiB 3 times with resize-disk type. ", func() {
		// TestRail:https://portworx.testrail.net/index.php?/tests/view/86355067
		StartTorpedoTest("PoolExpandDiskResize3Times",
			"Validate storage pool expansion with type=resize-disk", nil, 86355067)
		for i := 0; i < 3; i++ {
			poolToResize = getStoragePool(poolIDToResize)
			originalSizeInBytes = poolToResize.TotalSize
			targetSizeInBytes = originalSizeInBytes + 100*units.GiB
			targetSizeGiB = targetSizeInBytes / units.GiB

			log.InfoD("Current Size of pool %s is %d GiB. Expand to %v GiB with type resize-disk...",
				poolIDToResize, poolToResize.TotalSize/units.GiB, targetSizeGiB)
			triggerPoolExpansion(poolIDToResize, targetSizeGiB, api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK)
			resizeErr := waitForOngoingPoolExpansionToComplete(poolIDToResize)
			dash.VerifyFatal(resizeErr, nil, "Pool expansion does not result in error")
			verifyPoolSizeEqualOrLargerThanExpected(poolIDToResize, targetSizeGiB)
		}
	})
})

var _ = Describe("{PoolExpandSmoke}", Label("p0", "positive", "pool_ops", "px_ops", "PoolExpand"), func() {
	var (
		poolToResize   *api.StoragePool
		poolIDToResize string
		contexts       = make([]*scheduler.Context, 0)
	)
	BeforeEach(func() {
		contexts = scheduleApps()
	})

	JustBeforeEach(func() {
		poolIDToResize = pickPoolToResize(contexts, api.SdkStoragePool_RESIZE_TYPE_ADD_DISK, 100)
		log.Infof("Picked pool %s to resize", poolIDToResize)
		poolToResize = getStoragePool(poolIDToResize)
	})

	JustAfterEach(func() {
		AfterEachTest(contexts)
	})

	AfterEach(func() {
		appsValidateAndDestroy(contexts)
		EndTorpedoTest()
	})

	It("Verify expanding pool with add-disk type is rejected with dmthin. ", func() {
		StartTorpedoTest("PoolExpandDiskAdd",
			"Validate storage pool expansion with type=add-disk", nil, 0)
		originalSizeInBytes = poolToResize.TotalSize
		targetSizeInBytes = originalSizeInBytes + 100*units.GiB
		targetSizeGiB = targetSizeInBytes / units.GiB

		log.InfoD("Current Size of the pool %s is %d GiB. Trying to expand to %v GiB with type add-disk",
			poolIDToResize, poolToResize.TotalSize/units.GiB, targetSizeGiB)
		triggerPoolExpansion(poolIDToResize, targetSizeGiB, api.SdkStoragePool_RESIZE_TYPE_ADD_DISK)
		if isDMthin, _ := IsDMthin(); !isDMthin {
			resizeErr := waitForOngoingPoolExpansionToComplete(poolIDToResize)
			dash.VerifyFatal(resizeErr, nil, "Pool expansion does not result in error")
			verifyPoolSizeEqualOrLargerThanExpected(poolIDToResize, targetSizeGiB)
		}
	})

	It("Select a pool and expand it by 100 GiB with resize-disk type. ", func() {
		StartTorpedoTest("PoolExpandDiskResize",
			"Validate storage pool expansion with type=resize-disk", nil, 0)
		originalSizeInBytes = poolToResize.TotalSize
		targetSizeInBytes = originalSizeInBytes + 100*units.GiB
		targetSizeGiB = targetSizeInBytes / units.GiB

		log.InfoD("Current Size of the pool %s is %d GiB. Trying to expand to %v GiB with type resize-disk",
			poolIDToResize, poolToResize.TotalSize/units.GiB, targetSizeGiB)
		triggerPoolExpansion(poolIDToResize, targetSizeGiB, api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK)
		resizeErr := waitForOngoingPoolExpansionToComplete(poolIDToResize)
		dash.VerifyFatal(resizeErr, nil, "Pool expansion does not result in error")
		verifyPoolSizeEqualOrLargerThanExpected(poolIDToResize, targetSizeGiB)
	})

	It("Select a pool and expand it by 100 GiB with auto type. ", func() {
		StartTorpedoTest("PoolExpandDiskAuto",
			"Validate storage pool expansion with type=auto ", nil, 0)
		originalSizeInBytes = poolToResize.TotalSize
		targetSizeInBytes = originalSizeInBytes + 100*units.GiB
		targetSizeGiB = targetSizeInBytes / units.GiB

		log.InfoD("Current Size of the pool %s is %d GiB. Trying to expand to %v GiB with type auto",
			poolIDToResize, poolToResize.TotalSize/units.GiB, targetSizeGiB)
		triggerPoolExpansion(poolIDToResize, targetSizeGiB, api.SdkStoragePool_RESIZE_TYPE_AUTO)
		resizeErr := waitForOngoingPoolExpansionToComplete(poolIDToResize)
		dash.VerifyFatal(resizeErr, nil, "Pool expansion does not result in error")
		verifyPoolSizeEqualOrLargerThanExpected(poolIDToResize, targetSizeGiB)
	})
})

var _ = Describe("{PoolExpandRejectConcurrentDiskResize}", Label("p1", "positive", "pool_ops", "px_ops", "PoolExpand", "ResizeDisk"), func() {

	var (
		poolToResize   *api.StoragePool
		poolIDToResize string
		contexts       = make([]*scheduler.Context, 0)
	)
	BeforeEach(func() {
		contexts = scheduleApps()
	})

	JustBeforeEach(func() {
		poolIDToResize = pickPoolToResize(contexts, api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK, 100)
		log.Infof("Picked pool %s to resize", poolIDToResize)
		poolToResize = getStoragePool(poolIDToResize)
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

	// test resizing all pools on one storage node concurrently and ensure only the first one makes progress
	It("Select all pools on a storage node and expand them concurrently. ", func() {
		// TestRail:https://portworx.testrail.net/index.php?/tests/view/86355074
		StartTorpedoTest("PoolExpandRejectConcurrentDiskResize",
			"Validate storage pool expansion rejects concurrent requests", nil, 86355074)
		var pools []*api.StoragePool
		Step("Verify multiple pools are present on this node", func() {
			// collect all pools available
			pools = append(pools, storageNode.Pools...)
			dash.VerifyFatal(len(pools) > 1, true, "This test requires more than 1 pool.")
		})

		Step("Expand all pools concurrently. ", func() {
			expandType := api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK
			var wg sync.WaitGroup
			for _, p := range pools {
				wg.Add(1)
				go func(p *api.StoragePool) {
					defer wg.Done()
					err = Inst().V.ExpandPool(p.Uuid, expandType, p.TotalSize/units.GiB+100, true)
				}(p)
			}
			wg.Wait()
		})

		Step("Verify only one expansion is making progress at any given time", func() {
			inProgressCount := 0
			startTime := time.Now()
			for time.Since(startTime) < 1*time.Minute {
				inProgressCount = 0
				time.Sleep(1 * time.Second)
				storageNode, err = GetNodeWithGivenPoolID(poolIDToResize)
				for _, p := range storageNode.Pools {
					if p.LastOperation.Status == api.SdkStoragePool_OPERATION_IN_PROGRESS {
						inProgressCount++
					}
					dash.VerifyFatal(inProgressCount <= 1, true, "Only one pool expansion should be in progress at any given time.")
				}
			}
		})
	})

	// test expansion request on a pool while a previous expansion is in progress is rejected
	It("Expand a pool while a previous expansion is in progress", func() {
		expandType := api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK
		targetSize := poolToResize.TotalSize/units.GiB + 100
		err = Inst().V.ExpandPool(poolIDToResize, expandType, targetSize, true)
		isExpandInProgress, expandErr := poolResizeIsInProgress(poolToResize)
		if expandErr != nil {
			log.Fatalf("Error checking if pool expansion is in progress: %v", expandErr)
		}
		if !isExpandInProgress {
			log.Warnf("Pool expansion already finished. Skipping this test. Use a testing app that writes " +
				"more data which may slow down resize-disk type expansion. ")
			return
		}
		expandResponse := Inst().V.ExpandPoolUsingPxctlCmd(*storageNode, poolToResize.Uuid, expandType, targetSize+100, true)
		dash.VerifyFatal(expandResponse != nil, true, "Pool expansion should fail when expansion is in progress")
		dash.VerifyFatal(strings.Contains(expandResponse.Error(), "is already in progress"), true,
			"Pool expansion failure reason should be communicated to the user	")
	})
})

var _ = Describe("{PoolExpandRejectConcurrentDiskAdd}", Label("p1", "positive", "pool_ops", "px_ops", "PoolExpand", "AddDrive"), func() {

	var (
		poolToResize   *api.StoragePool
		poolIDToResize string
		contexts       = make([]*scheduler.Context, 0)
	)
	BeforeEach(func() {
		contexts = scheduleApps()
	})

	JustBeforeEach(func() {
		poolIDToResize = pickPoolToResize(contexts, api.SdkStoragePool_RESIZE_TYPE_ADD_DISK, 100)
		log.Infof("Picked pool %s to resize", poolIDToResize)
		poolToResize = getStoragePool(poolIDToResize)
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

	// test resizing all pools on one storage node concurrently and ensure only the first one makes progress
	It("Select all pools on a storage node and expand them concurrently. ", func() {
		// TestRail:https://portworx.testrail.net/index.php?/tests/view/34542836
		StartTorpedoTest("PoolExpandRejectConcurrent",
			"Validate storage pool expansion rejects concurrent requests", nil, 34542836)
		var pools []*api.StoragePool
		Step("Verify multiple pools are present on this node", func() {
			// collect all pools available
			pools = append(pools, storageNode.Pools...)
			dash.VerifyFatal(len(pools) > 1, true, "This test requires more than 1 pool.")
		})

		Step("Expand all pools concurrently. ", func() {
			expandType := api.SdkStoragePool_RESIZE_TYPE_ADD_DISK
			var wg sync.WaitGroup
			for _, p := range pools {
				wg.Add(1)
				go func(p *api.StoragePool) {
					defer wg.Done()
					err = Inst().V.ExpandPool(p.Uuid, expandType, p.TotalSize/units.GiB+100, true)
				}(p)
			}
			wg.Wait()
		})

		Step("Verify only one expansion is making progress at any given time", func() {
			inProgressCount := 0
			startTime := time.Now()
			for time.Since(startTime) < 1*time.Minute {
				inProgressCount = 0
				time.Sleep(5 * time.Second)
				storageNode, err = GetNodeWithGivenPoolID(poolIDToResize)
				for _, p := range storageNode.Pools {
					if p.LastOperation.Status == api.SdkStoragePool_OPERATION_IN_PROGRESS {
						inProgressCount++
					}
					dash.VerifyFatal(inProgressCount <= 1, true, "Only one pool expansion should be in progress at any given time.")
				}
			}
		})
	})

	// test expansion request on a pool while a previous expansion is in progress is rejected
	It("Expand a pool while a previous expansion is in progress", func() {
		expandType := api.SdkStoragePool_RESIZE_TYPE_ADD_DISK
		targetSize := poolToResize.TotalSize/units.GiB + 100
		err = Inst().V.ExpandPool(poolIDToResize, expandType, targetSize, true)
		// wait for expansion to start
		// TODO: this is a hack to wait for expansion to start. The existing WaitForExpansionToStart() risks returning
		// when the expansion has already completed.
		time.Sleep(5 * time.Second)
		// verify pool expansion is in progress
		isExpandInProgress, expandErr := poolResizeIsInProgress(poolToResize)
		if expandErr != nil {
			log.Fatalf("Error checking if pool expansion is in progress: %v", expandErr)
		}
		if !isExpandInProgress {
			log.Warnf("Pool expansion already finished. Skipping this test. Using a testing app that writes " +
				"more data which may slow down add-disk type expansion. ")
			return
		}
		expandResponse := Inst().V.ExpandPoolUsingPxctlCmd(*storageNode, poolToResize.Uuid, expandType, targetSize+100, true)
		dash.VerifyFatal(expandResponse != nil, true, "Pool expansion should fail when expansion is in progress")
		dash.VerifyFatal(strings.Contains(expandResponse.Error(), "is already in progress"), true,
			"Pool expansion failure reason should be communicated to the user	")
	})
})

var _ = Describe("{PoolExpandDiskResizeWithReboot}", Label("p1", "negative", "pool_ops", "px_ops", "error_injection", "PoolExpand", "node_reboot", "ResizeDisk"), func() {

	var (
		poolToResize   *api.StoragePool
		poolIDToResize string
		contexts       = make([]*scheduler.Context, 0)
	)
	BeforeEach(func() {
		contexts = scheduleApps()
	})

	JustBeforeEach(func() {
		poolIDToResize = pickPoolToResize(contexts, api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK, 100)
		log.Infof("Picked pool %s to resize", poolIDToResize)
		poolToResize = getStoragePool(poolIDToResize)
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

	It("Initiate pool expansion using resize-disk and reboot node", func() {
		StartTorpedoTest("PoolExpandDiskResizeWithReboot", "Initiate pool expansion using resize-disk and reboot node", nil, 51309)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
		Step("Select a pool that has I/O and expand it by 100 GiB with resize-disk type. ", func() {
			originalSizeInBytes = poolToResize.TotalSize
			targetSizeInBytes = originalSizeInBytes + 100*units.GiB
			targetSizeGiB = targetSizeInBytes / units.GiB
			log.InfoD("Current Size of the pool %s is %d GiB. Trying to expand to %v GiB with type resize-disk",
				poolIDToResize, poolToResize.TotalSize/units.GiB, targetSizeGiB)
			triggerPoolExpansion(poolIDToResize, targetSizeGiB, api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK)
		})

		Step("Wait for expansion to start and reboot node", func() {
			err := WaitForExpansionToStart(poolIDToResize)
			log.FailOnError(err, "Timed out waiting for expansion to start")
			err = RebootNodeAndWaitForPxUp(*storageNode)
			log.FailOnError(err, "Failed to reboot node and wait till it is up")
		})

		log.Infof("Debug Pool %s", poolIDToResize)
		Step("Ensure pool has been expanded to the expected size", func() {
			_ = waitForOngoingPoolExpansionToComplete(poolIDToResize)
			verifyPoolSizeEqualOrLargerThanExpected(poolIDToResize, targetSizeGiB)
		})
		log.Infof("Debug Pool %s", poolIDToResize)
		Step("Pool has expanded. Reboot the node", func() {
			err = RebootNodeAndWaitForPxUp(*storageNode)
			log.FailOnError(err, "Failed to reboot node and wait till it is up")
			verifyPoolSizeEqualOrLargerThanExpected(poolIDToResize, targetSizeGiB)
		})
	})
})

var _ = Describe("{PoolExpandDiskAddWithReboot}", Label("p1", "negative", "pool_ops", "error_injection", "PoolExpand", "node_reboot", "AddDrive"), func() {

	var (
		poolToResize   *api.StoragePool
		poolIDToResize string
		contexts       = make([]*scheduler.Context, 0)
	)
	BeforeEach(func() {
		contexts = scheduleApps()
	})

	JustBeforeEach(func() {
		poolIDToResize = pickPoolToResize(contexts, api.SdkStoragePool_RESIZE_TYPE_ADD_DISK, 100)
		log.Infof("Picked pool %s to resize", poolIDToResize)
		poolToResize = getStoragePool(poolIDToResize)
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
			originalSizeInBytes = poolToResize.TotalSize
			targetSizeInBytes = originalSizeInBytes + 100*units.GiB
			targetSizeGiB = targetSizeInBytes / units.GiB
			log.InfoD("Current Size of the pool %s is %d GiB. Trying to expand to %v GiB with type add-disk",
				poolIDToResize, poolToResize.TotalSize/units.GiB, targetSizeGiB)
			triggerPoolExpansion(poolIDToResize, targetSizeGiB, api.SdkStoragePool_RESIZE_TYPE_ADD_DISK)
		})

		Step("Wait for expansion to start and reboot node", func() {
			err := WaitForExpansionToStart(poolIDToResize)
			log.FailOnError(err, "Timed out waiting for expansion to start")
			err = RebootNodeAndWaitForPxUp(*storageNode)
			log.FailOnError(err, "Failed to reboot node and wait till it is up")
		})

		Step("Ensure pool has been expanded to the expected size", func() {
			err = waitForOngoingPoolExpansionToComplete(poolIDToResize)
			dash.VerifyFatal(err, nil, "Pool expansion does not result in error")
			verifyPoolSizeEqualOrLargerThanExpected(poolIDToResize, targetSizeGiB)
		})

		Step("Pool has expanded. Reboot the node", func() {
			err = RebootNodeAndWaitForPxUp(*storageNode)
			log.FailOnError(err, "Failed to reboot node and wait till it is up")
			verifyPoolSizeEqualOrLargerThanExpected(poolIDToResize, targetSizeGiB)
		})
	})
})

var _ = Describe("{PoolExpandDiskResizePXRestart}", Label("p1", "negative", "pool_ops", "error_injection", "PoolExpand", "px_restart", "ResizeDisk"), func() {

	var (
		poolToResize   *api.StoragePool
		poolIDToResize string
		contexts       = make([]*scheduler.Context, 0)
	)
	BeforeEach(func() {
		contexts = scheduleApps()
	})

	JustBeforeEach(func() {
		poolIDToResize = pickPoolToResize(contexts, api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK, 100)
		log.Infof("Picked pool %s to resize", poolIDToResize)
		poolToResize = getStoragePool(poolIDToResize)
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

	It("Restart PX after pool expansion", func() {
		StartTorpedoTest("PoolExpandDiskResizePXRestart",
			"Restart PX after pool expansion", nil, testrailID)

		Step("Select a pool that has I/O and expand it by 100 GiB with resize-disk type. ", func() {
			originalSizeInBytes = poolToResize.TotalSize
			targetSizeInBytes = originalSizeInBytes + 100*units.GiB
			targetSizeGiB = targetSizeInBytes / units.GiB
			log.InfoD("Current Size of the pool %s is %d GiB. Trying to expand to %v GiB with type resize-disk",
				poolIDToResize, poolToResize.TotalSize/units.GiB, targetSizeGiB)
			triggerPoolExpansion(poolIDToResize, targetSizeGiB, api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK)
		})

		Step("Wait for expansion to finish and restart PX", func() {
			resizeErr := waitForOngoingPoolExpansionToComplete(poolIDToResize)
			dash.VerifyFatal(resizeErr, nil, "Pool expansion does not result in error")
			log.FailOnError(Inst().V.RestartDriver(*storageNode, nil),
				fmt.Sprintf("Error restarting px on node [%s]", storageNode.Name))
			log.FailOnError(Inst().V.WaitDriverUpOnNode(*storageNode, addDriveUpTimeOut),
				fmt.Sprintf("Timed out waiting for px to come up on node [%s]", storageNode.Name))
		})

		Step("Ensure pool is up and running", func() {
			// Ensure pool is up and running
			poolToResize = getStoragePool(poolIDToResize)
			// Ensure poolToResize is not nil
			dash.VerifyFatal(poolToResize != nil, true, "Pool is up and running after restart")
			verifyPoolSizeEqualOrLargerThanExpected(poolIDToResize, targetSizeGiB)
		})

		Step("Expansion successful. Restart PX", func() {
			err = Inst().V.RestartDriver(*storageNode, nil)
			log.FailOnError(err, fmt.Sprintf("Error restarting px on node [%s]", storageNode.Name))
			err = Inst().V.WaitDriverUpOnNode(*storageNode, addDriveUpTimeOut)
			log.FailOnError(err, fmt.Sprintf("Timed out waiting for px to come up on node [%s]", storageNode.Name))
			verifyPoolSizeEqualOrLargerThanExpected(poolIDToResize, targetSizeGiB)
		})
	})
})

var _ = Describe("{PoolExpandDiskAddPXRestart}", Label("p1", "negative", "pool_ops", "error_injection", "PoolExpand", "px_restart", "AddDrive"), func() {

	var (
		poolToResize   *api.StoragePool
		poolIDToResize string
		contexts       = make([]*scheduler.Context, 0)
	)
	BeforeEach(func() {
		contexts = scheduleApps()
	})

	JustBeforeEach(func() {
		poolIDToResize = pickPoolToResize(contexts, api.SdkStoragePool_RESIZE_TYPE_ADD_DISK, 100)
		log.Infof("Picked pool %s to resize", poolIDToResize)
		poolToResize = getStoragePool(poolIDToResize)
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

	It("Restart PX after pool expansion", func() {
		StartTorpedoTest("PoolExpandDiskAddPXRestart",
			"Restart PX after pool expansion", nil, testrailID)

		Step("Select a pool that has I/O and expand it by 100 GiB with add-disk type. ", func() {
			originalSizeInBytes = poolToResize.TotalSize
			targetSizeInBytes = originalSizeInBytes + 100*units.GiB
			targetSizeGiB = targetSizeInBytes / units.GiB
			log.InfoD("Current Size of the pool %s is %d GiB. Trying to expand to %v GiB with type add-disk",
				poolIDToResize, poolToResize.TotalSize/units.GiB, targetSizeGiB)
			triggerPoolExpansion(poolIDToResize, targetSizeGiB, api.SdkStoragePool_RESIZE_TYPE_ADD_DISK)
		})

		Step("Wait for expansion to finish and restart PX", func() {
			resizeErr := waitForOngoingPoolExpansionToComplete(poolIDToResize)
			dash.VerifyFatal(resizeErr, nil, "Pool expansion does not result in error")
			log.FailOnError(Inst().V.RestartDriver(*storageNode, nil),
				fmt.Sprintf("Error restarting px on node [%s]", storageNode.Name))
			log.FailOnError(Inst().V.WaitDriverUpOnNode(*storageNode, addDriveUpTimeOut),
				fmt.Sprintf("Timed out waiting for px to come up on node [%s]", storageNode.Name))
		})

		Step("Ensure pool is up and running", func() {
			// Ensure pool is up and running
			poolToResize = getStoragePool(poolIDToResize)
			// Ensure poolToResize is not nil
			dash.VerifyFatal(poolToResize != nil, true, "Pool is up and running after restart")
			verifyPoolSizeEqualOrLargerThanExpected(poolIDToResize, targetSizeGiB)
		})

		Step("Expansioin successful. Restart PX", func() {
			err = Inst().V.RestartDriver(*storageNode, nil)
			log.FailOnError(err, fmt.Sprintf("Error restarting px on node [%s]", storageNode.Name))
			err = Inst().V.WaitDriverUpOnNode(*storageNode, addDriveUpTimeOut)
			log.FailOnError(err, fmt.Sprintf("Timed out waiting for px to come up on node [%s]", storageNode.Name))
			verifyPoolSizeEqualOrLargerThanExpected(poolIDToResize, targetSizeGiB)
		})
	})
})

var _ = Describe("{PoolExpandInvalidSize}", Label("p2", "negative", "pool_ops", "PoolExpand"), func() {
	// TestrailId: https://portworx.testrail.net/index.php?/tests/view/34542945
	BeforeEach(func() {
		StartTorpedoTest("PoolExpansionDiskResizeInvalidSize",
			"Initiate pool expansion using invalid expansion size", nil, 34542945)
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
		storageNode, err = GetNodeWithGivenPoolID(pooltoPick)
		log.FailOnError(err, "Failed to get node with given pool ID")

		resizeErr := Inst().V.ExpandPool(pooltoPick, api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK, 30000000, true)
		dash.VerifyFatal(resizeErr, nil, "Verify error occurs with invalid Pool expansion size")
	})

})

var _ = Describe("{PoolExpandResizeInvalidPoolID}", Label("p2", "negative", "pool_ops", "PoolExpand"), func() {
	// TestrailID: https://portworx.testrail.net/index.php?/tests/view/34542946
	BeforeEach(func() {
		StartTorpedoTest("PoolExpandResizeInvalidPoolID",
			"Initiate pool expansion using invalid Id", nil, 34542946)
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
			resizeErr := Inst().V.ExpandPool(invalidPoolUUID, api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK, 100, true)
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

var _ = Describe("{PoolExpandDiskResizeAndVerifyFromOtherNode}", Label("p0", "positive", "pool_ops", "PoolExpand", "ResizeDisk"), func() {

	var (
		poolToResize   *api.StoragePool
		poolIDToResize string
		contexts       = make([]*scheduler.Context, 0)
	)
	BeforeEach(func() {
		StartTorpedoTest("PoolExpandDiskResizeAndVerifyFromOtherNode",
			"Initiate pool expansion and verify from other node", nil, 34542840)
		contexts = scheduleApps()
	})

	JustBeforeEach(func() {
		poolIDToResize = pickPoolToResize(contexts, api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK, 100)
		log.Infof("Picked pool %s to resize", poolIDToResize)
		poolToResize = getStoragePool(poolIDToResize)
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

	stepLog := "should get the existing pool and expand it by resizing a disk and verify from other node"
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

		originalSizeInBytes = poolToResize.TotalSize
		targetSizeInBytes = originalSizeInBytes + 100*units.GiB
		targetSizeGiB = targetSizeInBytes / units.GiB

		log.InfoD("Current Size of the pool %s is %d GiB. Trying to expand to %v GiB with type resize-disk",
			poolIDToResize, poolToResize.TotalSize/units.GiB, targetSizeGiB)
		triggerPoolExpansion(poolIDToResize, targetSizeGiB, api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK)

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

var _ = Describe("{PoolExpandDiskAddAndVerifyFromOtherNode}", Label("p0", "positive", "pool_ops", "PoolExpand", "AddDrive"), func() {
	// TestrailID: https://portworx.testrail.net/index.php?/tests/view/34542840
	var (
		poolToResize   *api.StoragePool
		poolIDToResize string
		contexts       = make([]*scheduler.Context, 0)
	)
	BeforeEach(func() {
		StartTorpedoTest("PoolExpandDiskAddAndVerifyFromOtherNode",
			"Initiate pool expansion and verify from other node", nil, 34542840)
		contexts = scheduleApps()
	})

	JustBeforeEach(func() {
		poolIDToResize = pickPoolToResize(contexts, api.SdkStoragePool_RESIZE_TYPE_ADD_DISK, 100)
		log.Infof("Picked pool %s to resize", poolIDToResize)
		poolToResize = getStoragePool(poolIDToResize)
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

		originalSizeInBytes = poolToResize.TotalSize
		targetSizeInBytes = originalSizeInBytes + 100*units.GiB
		targetSizeGiB = targetSizeInBytes / units.GiB

		log.InfoD("Current Size of the pool %s is %d GiB. Trying to expand to %v GiB with type add-disk",
			poolIDToResize, poolToResize.TotalSize/units.GiB, targetSizeGiB)
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

var _ = Describe("{PoolExpandResizeWithSameSize}", Label("p1", "positive", "pool_ops", "PoolExpand", "ResizeDisk"), func() {
	// TestrailId: https://portworx.testrail.net/index.php?/tests/view/34542944

	var poolToResize *api.StoragePool
	BeforeEach(func() {
		StartTorpedoTest("PoolExpandResizeWithSameSize",
			"Initiate pool expansion using same size", nil, 34542944)
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
		poolToResize = getStoragePool(pooltoPick)

		originalSizeGiB := poolToResize.TotalSize / units.GiB
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

var _ = Describe("{PoolExpandWhileResizeDiskInProgress}", Label("p1", "positive", "pool_ops", "PoolExpand", "ResizeDisk"), func() {

	var testrailID = 34542896
	// TestrailId: https://portworx.testrail.net/index.php?/tests/view/34542896
	var (
		poolToResize   *api.StoragePool
		poolIDToResize string
		contexts       = make([]*scheduler.Context, 0)
	)
	BeforeEach(func() {
		StartTorpedoTest("PoolExpandWhileResizeDiskInProgress",
			"Initiate pool expansion on a pool where one pool expansion is already in progress", nil, testrailID)
		contexts = scheduleApps()
	})

	JustBeforeEach(func() {
		poolIDToResize = pickPoolToResize(contexts, api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK, 100)
		log.Infof("Picked pool %s to resize", poolIDToResize)
		poolToResize = getStoragePool(poolIDToResize)
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

		originalSizeInBytes = poolToResize.TotalSize
		targetSizeInBytes = originalSizeInBytes + 100*units.GiB
		targetSizeGiB = targetSizeInBytes / units.GiB

		log.InfoD("Current Size of the pool %s is %d GiB. Trying to expand to %v GiB with type resize-disk",
			poolIDToResize, poolToResize.TotalSize/units.GiB, targetSizeGiB)
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
		dash.VerifyFatal(errMatch, nil, "Pool expand with one resize already in progress failed as expected.")

		Step("Ensure pool has been expanded to the expected size", func() {
			err = waitForOngoingPoolExpansionToComplete(poolIDToResize)
			dash.VerifyFatal(err, nil, "Pool expansion does not result in error")
			verifyPoolSizeEqualOrLargerThanExpected(poolIDToResize, targetSizeGiB)
		})

	})

})

var _ = Describe("{PoolExpandResizePoolMaintenanceCycle}", Label("p1", "negative", "pool_ops", "error_injection", "PoolExpand", "NodeMaintenance"), func() {
	var testrailID = 34542842
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/tests/view/34542842
	var (
		poolToResize   *api.StoragePool
		poolIDToResize string
		contexts       = make([]*scheduler.Context, 0)
	)
	BeforeEach(func() {
		StartTorpedoTest("PoolExpandResizePoolMaintenanceCycle",
			"Initiate pool expansion and do a maintenance cycle after resize", nil, testrailID)
		contexts = scheduleApps()
	})

	JustBeforeEach(func() {
		poolIDToResize = pickPoolToResize(contexts, api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK, 100)
		log.Infof("Picked pool %s to resize", poolIDToResize)
		poolToResize = getStoragePool(poolIDToResize)
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

	stepLog := "cycle through maintenance mode after pool expand is complete"
	It(stepLog, func() {
		log.InfoD(stepLog)
		originalSizeInBytes = poolToResize.TotalSize
		targetSizeInBytes = originalSizeInBytes + 100*units.GiB
		targetSizeGiB = targetSizeInBytes / units.GiB

		log.InfoD("Current Size of the pool %s is %d GiB. Trying to expand to %v GiB with type add-disk",
			poolIDToResize, poolToResize.TotalSize/units.GiB, targetSizeGiB)
		triggerPoolExpansion(poolIDToResize, targetSizeGiB, api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK)

		err = waitForOngoingPoolExpansionToComplete(poolIDToResize)
		dash.VerifyFatal(err, nil, "Pool expansion does not result in error")
		verifyPoolSizeEqualOrLargerThanExpected(poolIDToResize, targetSizeGiB)

		// Enter Maintenance Mode
		err = Inst().V.EnterMaintenance(*storageNode)
		log.FailOnError(err, fmt.Sprintf("fail to enter node %s in maintenance mode", storageNode.Name))
		status, err := Inst().V.GetNodeStatus(*storageNode)
		log.FailOnError(err, fmt.Sprintf("Error getting PX status of node %s", storageNode.Name))
		dash.VerifyFatal(*status, api.Status_STATUS_MAINTENANCE, fmt.Sprintf("Node %s Status not Online", storageNode.Name))

		// Exit Maintenance Mode
		err = Inst().V.ExitMaintenance(*storageNode)
		log.FailOnError(err, fmt.Sprintf("fail to exit node %s in maintenance mode", storageNode.Name))
		status, err = Inst().V.GetNodeStatus(*storageNode)
		log.FailOnError(err, fmt.Sprintf("Error getting PX status of node %s", storageNode.Name))
		dash.VerifyFatal(*status, api.Status_STATUS_OK, fmt.Sprintf("Node %s Status not Online", storageNode.Name))

		// verify pool size after maintenance cycle
		verifyPoolSizeEqualOrLargerThanExpected(poolIDToResize, targetSizeGiB)

		// check pool status is healthy after maintenance cycle
		poolsStatus, err := Inst().V.GetNodePoolsStatus(*storageNode)
		log.FailOnError(err, "error getting pool status on node %s", storageNode.Name)
		dash.VerifyFatal(poolsStatus[poolIDToResize], "Online", fmt.Sprintf("Pool %s Status not Online", poolIDToResize))
	})
})

var _ = Describe("{MaintenanceCycleDuringPoolExpandResizeDisk}", Label("p1", "negative", "pool_ops", "error_injection", "PoolExpand", "NodeMaintenance", "ResizeDisk"), func() {
	var testrailID = 34542902
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/tests/view/34542902

	/*
	   Steps:
	       1. Initiate pool expand with resize-disk operation.
	       2. Cycle the node through maintenance mode.
	       3. Verify pool expand operation goes to completion.
	*/
	var (
		poolToResize   *api.StoragePool
		poolIDToResize string
		contexts       = make([]*scheduler.Context, 0)
	)
	BeforeEach(func() {
		StartTorpedoTest("MaintenanceCycleDuringPoolExpandResizeDisk",
			"Perform maintenance cycle during pool expand with resize-disk operation", nil, testrailID)
		contexts = scheduleApps()
	})

	JustBeforeEach(func() {
		poolIDToResize = pickPoolToResize(contexts, api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK, 100)
		log.Infof("Picked pool %s to resize", poolIDToResize)
		poolToResize = getStoragePool(poolIDToResize)
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

	stepLog := "cycle through maintenance mode during pool expand with resize-disk"
	It(stepLog, func() {
		log.InfoD(stepLog)

		originalSizeInBytes = poolToResize.TotalSize
		targetSizeInBytes = originalSizeInBytes + 100*units.GiB
		targetSizeGiB = targetSizeInBytes / units.GiB

		log.InfoD("Current Size of the pool %s is %d GiB. Trying to expand to %v GiB with type resize-disk",
			poolIDToResize, poolToResize.TotalSize/units.GiB, targetSizeGiB)
		triggerPoolExpansion(poolIDToResize, targetSizeGiB, api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK)

		isjournal, err := IsJournalEnabled()
		log.FailOnError(err, "Failed to check is journal enabled")
		if isjournal {
			targetSizeGiB = targetSizeGiB - 3
		}

		// Enter Maintenance Mode
		err = Inst().V.EnterMaintenance(*storageNode)
		log.FailOnError(err, fmt.Sprintf("fail to enter node %s in maintenance mode", storageNode.Name))
		status, err := Inst().V.GetNodeStatus(*storageNode)
		log.FailOnError(err, fmt.Sprintf("Error getting PX status of node %s", storageNode.Name))
		dash.VerifyFatal(*status, api.Status_STATUS_MAINTENANCE, fmt.Sprintf("Node %s Status not Online", storageNode.Name))

		// Exit Maintenance Mode
		err = Inst().V.ExitMaintenance(*storageNode)
		log.FailOnError(err, fmt.Sprintf("fail to exit node %s in maintenance mode", storageNode.Name))
		status, err = Inst().V.GetNodeStatus(*storageNode)
		log.FailOnError(err, fmt.Sprintf("Error getting PX status of node %s", storageNode.Name))
		dash.VerifyFatal(*status, api.Status_STATUS_OK, fmt.Sprintf("Node %s Status not Online", storageNode.Name))

		err = waitForOngoingPoolExpansionToComplete(poolIDToResize)
		dash.VerifyFatal(err, nil, "Pool expansion does not result in error")

		// verify pool size after maintenance cycle
		verifyPoolSizeEqualOrLargerThanExpected(poolIDToResize, targetSizeGiB)

		// check pool status is healthy after maintenance cycle
		poolsStatus, err := Inst().V.GetNodePoolsStatus(*storageNode)
		log.FailOnError(err, "error getting pool status on node %s", storageNode.Name)
		dash.VerifyFatal(poolsStatus[poolIDToResize], "Online", fmt.Sprintf("Pool %s Status not Online", poolIDToResize))
	})
})

var _ = Describe("{PoolExpandResizeDiskInMaintenanceMode}", Label("p1", "negative", "pool_ops", "error_injection", "PoolExpand", "NodeMaintenance", "ResizeDisk"), func() {
	var testrailID = 34542861
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/tests/view/34542861

	/*
		Steps:
			1. Move a node to maintenance mode.
			2. Initiate pool expand with resize-disk operation.
			3. Exit out of maintenance mode (PX only performs pool expand in normal mode, not in maintenance mode)
			4. Verify pool expand operation goes to completion.
	*/
	var (
		poolToResize   *api.StoragePool
		poolIDToResize string
		contexts       = make([]*scheduler.Context, 0)
	)
	BeforeEach(func() {
		StartTorpedoTest("PoolExpandResizeDiskInMaintenanceMode",
			"Initiate pool expand with resize-disk when node is already in maintenance mode", nil, testrailID)
		contexts = scheduleApps()
	})

	JustBeforeEach(func() {
		poolIDToResize = pickPoolToResize(contexts, api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK, 100)
		log.Infof("Picked pool %s to resize", poolIDToResize)
		poolToResize = getStoragePool(poolIDToResize)
	})

	JustAfterEach(func() {
		AfterEachTest(contexts)
	})

	AfterEach(func() {
		appsValidateAndDestroy(contexts)
		EndTorpedoTest()
	})

	stepLog := "Start pool expand with resize-disk on node which is already in maintenance mode "
	It(stepLog, func() {
		log.InfoD(stepLog)
		var nodeDetail *node.Node
		var err error
		stepLog = "Move Pool to maintenance mode"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			nodeDetail, err = GetNodeWithGivenPoolID(poolToResize.Uuid)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Failed to get Node Details using PoolUUID [%v]", poolToResize.Uuid))

			log.InfoD("Bring Pool to Maintenance Mode")
			err = Inst().V.EnterPoolMaintenance(*nodeDetail)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Failed to shift Node [%s] to Mainteinance Mode", nodeDetail.Name))
		})

		stepLog = "Initiate pool expand with resize-disk operation"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			originalSizeInBytes = poolToResize.TotalSize
			targetSizeInBytes = originalSizeInBytes + 100*units.GiB
			targetSizeGiB = targetSizeInBytes / units.GiB

			log.InfoD("Current Size of the pool %s is %d GiB. Trying to expand to %v GiB with type resize-disk",
				poolIDToResize, poolToResize.TotalSize/units.GiB, targetSizeGiB)
			err := Inst().V.ExpandPool(poolIDToResize, api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK, targetSizeGiB, true)
			dash.VerifyFatal(err, nil, "pool expansion requested successfully")
		})

		stepLog = "Exit node out of maintenance mode"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			log.InfoD("Bring Pool out of Maintenance Mode")
			err = Inst().V.ExitPoolMaintenance(*nodeDetail)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Failed to shift Node [%s] out of Mainteinance Mode", nodeDetail.Name))
		})

		stepLog = "Verify pool expand completes successfully"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			resizeErr := waitForOngoingPoolExpansionToComplete(poolIDToResize)
			dash.VerifyFatal(resizeErr, nil, "Pool expansion does not result in error")
			verifyPoolSizeEqualOrLargerThanExpected(poolIDToResize, targetSizeGiB)
		})
	})
})

var _ = Describe("{PoolExpandAddDiskInMaintenanceMode}", Label("p1", "negative", "pool_ops", "error_injection", "PoolExpand", "NodeMaintenance", "AddDrive"), func() {
	var testrailID = 34542888
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/tests/view/34542888

	/*
		Steps:
			1. Move a node to maintenance mode.
			2. Initiate pool expand with add-disk operation.
			3. Exit out of maintenance mode (PX only performs pool expand in normal mode, not in maintenance mode)
			4. Verify pool expand operation goes to completion.
	*/
	var (
		poolToResize   *api.StoragePool
		poolIDToResize string
		contexts       = make([]*scheduler.Context, 0)
	)
	BeforeEach(func() {
		StartTorpedoTest("PoolExpandAddDiskInMaintenanceMode",
			"Initiate pool expand with add-disk when node is already in maintenance mode", nil, testrailID)
		contexts = scheduleApps()
	})

	JustBeforeEach(func() {
		poolIDToResize = pickPoolToResize(contexts, api.SdkStoragePool_RESIZE_TYPE_ADD_DISK, 100)
		log.Infof("Picked pool %s to resize", poolIDToResize)
		poolToResize = getStoragePool(poolIDToResize)
	})

	JustAfterEach(func() {
		AfterEachTest(contexts)
	})

	AfterEach(func() {
		appsValidateAndDestroy(contexts)
		EndTorpedoTest()
	})

	stepLog := "Start pool expand with add-disk on node which is already in maintenance mode "
	It(stepLog, func() {
		log.InfoD(stepLog)
		var nodeDetail *node.Node
		var err error
		stepLog = "Move pool to maintenance mode"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			nodeDetail, err = GetNodeWithGivenPoolID(poolToResize.Uuid)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Failed to get Node Details using PoolUUID [%v]", poolToResize.Uuid))

			log.InfoD("Bring Pool to Maintenance Mode")
			err = Inst().V.EnterPoolMaintenance(*nodeDetail)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Failed to shift Node [%s] to Mainteinance Mode", nodeDetail.Name))
		})

		stepLog = "Initiate pool expand with add-disk operation"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			originalSizeInBytes = poolToResize.TotalSize
			targetSizeInBytes = originalSizeInBytes + 100*units.GiB
			targetSizeGiB = targetSizeInBytes / units.GiB

			log.InfoD("Current Size of the pool %s is %d GiB. Trying to expand to %v GiB with type add-disk",
				poolIDToResize, poolToResize.TotalSize/units.GiB, targetSizeGiB)
			err := Inst().V.ExpandPool(poolIDToResize, api.SdkStoragePool_RESIZE_TYPE_ADD_DISK, targetSizeGiB, true)

			if isDMthin, _ := IsDMthin(); isDMthin {
				dash.VerifyFatal(err != nil, true,
					"Pool expansion request of add-disk type should be rejected with dmthin")
				dash.VerifyFatal(strings.Contains(err.Error(), "add-drive type expansion is not supported with px-storev2"), true, fmt.Sprintf("check error message: %v", err.Error()))
			} else {
				dash.VerifyFatal(err, nil, "pool expansion requested successfully")
			}
		})

		if isDMthin, _ := IsDMthin(); !isDMthin {
			stepLog = "Exit Pool out of maintenance mode"
			Step(stepLog, func() {
				log.InfoD(stepLog)
				log.InfoD("Bring Node out of Maintenance Mode")
				err = Inst().V.ExitPoolMaintenance(*nodeDetail)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Failed to shift Node [%s] out of Mainteinance Mode", nodeDetail.Name))
			})

			stepLog = "Verify pool expand completes successfully"
			Step(stepLog, func() {
				log.InfoD(stepLog)
				resizeErr := waitForOngoingPoolExpansionToComplete(poolIDToResize)
				dash.VerifyFatal(resizeErr, nil, "Pool expansion does not result in error")
				verifyPoolSizeEqualOrLargerThanExpected(poolIDToResize, targetSizeGiB)
			})
		}
	})
})

var _ = Describe("{StorageFullPoolExpansion}", Label("p0", "positive", "pool_ops", "PoolExpand", "Throttling"), func() {
	var (
		appList        []string
		selectedNode   *node.Node
		poolToResize   *api.StoragePool
		poolIDToResize string
		contexts       = make([]*scheduler.Context, 0)
	)

	BeforeEach(func() {
		Inst().AppList = []string{"fio-fastpath-repl1"}
		contexts = ScheduleApplications("storagefull-resize")
		appList = Inst().AppList
	})

	JustBeforeEach(func() {
		selectedNode = GetNodeWithLeastSize()
		_ = Inst().S.AddLabelOnNode(*selectedNode, k8s.NodeType, k8s.FastpathNodeType)
		log.FailOnError(err, fmt.Sprintf("Failed to add fastpath label on node %v", selectedNode.Name))
	})

	AfterEach(func() {
		Inst().AppList = appList
		appsValidateAndDestroy(contexts)
		_ = Inst().S.RemoveLabelOnNode(*selectedNode, k8s.NodeType)
	})

	It("Expand pool with resize-disk type after pool is down due to storage full", func() {
		// https://portworx.testrail.net/index.php?/cases/view/51280
		StartTorpedoTest("StorageFullPoolResize", "Feed a pool full, then expand the pool in type resize-disk", nil, 51280)
		Step("Prepare a full pool to expand", func() {
			err = WaitForPoolOffline(*selectedNode)
			log.FailOnError(err, fmt.Sprintf("Timed out waiting to load a pool and bring node %s storage down", selectedNode.Name))
			poolsStatus, err := Inst().V.GetNodePoolsStatus(*selectedNode)
			log.FailOnError(err, "error getting pool status on node %s", selectedNode.Name)
			for i, s := range poolsStatus {
				if s == "Offline" {
					poolIDToResize = i
					poolToResize, err = GetStoragePoolByUUID(poolIDToResize)
					log.FailOnError(err, "error getting pool with UUID [%s]", poolIDToResize)
					break
				}
			}
		})

		Step("Expand the full pool in type resize-disk", func() {
			targetSizeGiB = (poolToResize.TotalSize / units.GiB) * 2
			log.InfoD("Current Size of the pool %s is %d, trying to expand it to double the size", poolToResize.Uuid, poolToResize.TotalSize/units.GiB)
			err = Inst().V.ExpandPool(poolToResize.Uuid, api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK, targetSizeGiB, true)
			dash.VerifyFatal(err, nil, "Pool expansion init should be successful.")
		})

		Step("Verify that pool expansion is successful", func() {
			err = waitForOngoingPoolExpansionToComplete(poolToResize.Uuid)
			log.FailOnError(err, fmt.Sprintf("Error waiting for pool %s resize", poolToResize.Uuid))
			verifyPoolSizeEqualOrLargerThanExpected(poolIDToResize, targetSizeGiB)
			status, err := Inst().V.GetNodeStatus(*selectedNode)
			log.FailOnError(err, fmt.Sprintf("Error getting PX status of node %s", selectedNode.Name))
			dash.VerifySafely(*status, api.Status_STATUS_OK, fmt.Sprintf("validate PX status on node %s", selectedNode.Name))
		})
	})
})

var _ = Describe("{PoolExpandTestLimits}", Label("p1", "positive", "pool_ops", "PoolExpand"), func() {
	var (
		poolSizeInGiB  uint64
		poolToResize   *api.StoragePool
		poolIDToResize string
		contexts       = make([]*scheduler.Context, 0)
	)
	BeforeEach(func() {
		isDMthin, err := IsDMthin()
		dash.VerifyFatal(err, nil, "error verifying if set up is DMTHIN enabled")
		if !isDMthin {
			Skip("DMThin/PX-Storev2 is not enabled on underlaying PX cluster. Skipping `PoolExpandTestBeyond15TiBLimit` test.")
		}
		contexts = scheduleApps()
	})

	JustAfterEach(func() {
		AfterEachTest(contexts)
	})

	AfterEach(func() {
		appsValidateAndDestroy(contexts)
		EndTorpedoTest()
	})

	It("Initiate pool expansion (DMThin) to its limits (15 TiB)", func() {
		poolIDToResize = pickPoolToResize(contexts, api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK, 100)
		log.Infof("Picked pool %s to resize", poolIDToResize)
		poolToResize = getStoragePool(poolIDToResize)
		poolOriginalSize := poolToResize.GetTotalSize()
		poolSizeInGiB = poolOriginalSize / units.GiB
		log.InfoD("original size of the pool %s is %d GiB", poolIDToResize, poolSizeInGiB)

		var testrailID = 51292
		// testrailID corresponds to: https://portworx.testrail.net/index.php?/cases/view/51292

		StartTorpedoTest("PoolExpandTestWithin15TiBLimit",
			"Initiate pool expansion using resize-disk to 15 TiB target size", nil, testrailID)

		targetSizeTiB := uint64(15)
		targetSizeInBytes = targetSizeTiB * units.TiB
		targetSizeGiB = targetSizeInBytes / units.GiB

		log.InfoD("Current Size of the pool %s is %d GiB. Trying to expand to %v TiB with type resize-disk",
			poolIDToResize, targetSizeGiB, targetSizeTiB)
		triggerPoolExpansion(poolIDToResize, targetSizeGiB, api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK)
		resizeErr := waitForOngoingPoolExpansionToComplete(poolIDToResize)
		dash.VerifyFatal(resizeErr, nil, "Pool expansion should not result in error")
		verifyPoolSizeEqualOrLargerThanExpected(poolIDToResize, targetSizeGiB)

	})

	It("Expand pool to 20 TiB (beyond max supported capacity for DMThin) with resize-disk type. ", func() {
		// TestrailId: https://portworx.testrail.net/index.php?/cases/view/50643

		StartTorpedoTest("DMThinPoolExpandBeyond15TiBLimit",
			"Initiate pool expansion using resize-disk to 20 TiB target size", nil, testrailID)

		targetSizeTiB := uint64(20)
		targetSizeInBytes = targetSizeTiB * units.TiB
		targetSizeGiB = targetSizeInBytes / units.GiB
		log.InfoD("Trying to expand pool %s to %v TiB with type resize-disk",
			poolIDToResize, targetSizeTiB)
		err = Inst().V.ExpandPool(poolIDToResize, api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK, targetSizeGiB, true)
		dash.VerifyFatal(err != nil, true, "DMThin pool expansion to 20 TB should result in error")
	})

	It("Delete the expanded pool and recreate with original size", func() {
		nodeSelected, err := GetNodeFromPoolUUID(poolIDToResize)
		failOnError(err, fmt.Sprintf("failed to get node details from the pool id %s", poolIDToResize))

		poolIDSelected, err := GetPoolIDFromPoolUUID(poolIDToResize)
		failOnError(err, fmt.Sprintf("failed to get pool id for the pool %s", poolIDToResize))

		stepLog = "Delete the selected pool"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			err = DeletePoolAndValidate(*nodeSelected, strconv.Itoa(int(poolIDSelected)))
			log.FailOnError(err, fmt.Sprintf("Error occured while Validating the deleted pool %d in the node %s", int(poolIDSelected), nodeSelected.Name))
		})

		nodePoolMapBfrAddDrive, err := Inst().V.GetNodePools(*nodeSelected)
		log.FailOnError(err, fmt.Sprintf("Get Node pools failed on node %s", nodeSelected.Name))

		stepLog = "Create pool with original size"
		Step(stepLog, func() {
			log.InfoD("original size of the pool %d is %d GiB", poolIDSelected, poolSizeInGiB)

			//Get cloudrive spec
			driveSpecs, err := GetCloudDriveDeviceSpecs()
			log.FailOnError(err, "Error getting cloud drive specs")

			deviceSpec := driveSpecs[0]
			deviceSpecParams := strings.Split(deviceSpec, ",")
			paramsArr := make([]string, 0)
			for _, param := range deviceSpecParams {
				if strings.Contains(param, "size") {
					paramsArr = append(paramsArr, fmt.Sprintf("size=%d", poolSizeInGiB))
				} else {
					paramsArr = append(paramsArr, param)
				}
			}
			//drive spec generated from actual cloudrive spec
			newSpec := strings.Join(paramsArr, ",")

			stepLog = "Add drive to create newpool"
			Step(stepLog, func() {
				log.Info(stepLog)
				err = Inst().V.AddCloudDrive(nodeSelected, newSpec, -1)
				log.FailOnError(err, fmt.Sprintf("Add cloud drive failed on node %s", nodeSelected.Name))
			})

			nodePoolMapAftrAddDrive, err := Inst().V.GetNodePools(*nodeSelected)
			log.FailOnError(err, fmt.Sprintf("Get Node pools failed on node %s", nodeSelected.Name))
			dash.VerifyFatal(len(nodePoolMapAftrAddDrive) > len(nodePoolMapBfrAddDrive), true, fmt.Sprintf("verify pool is added successfully"))
		})

	})
})

var _ = Describe("{PoolExpandAndCheckAlertsUsingResizeDisk}", Label("p0", "positive", "pool_ops", "PoolExpand", "ResizeDisk"), func() {

	var testrailID = 34542894
	var (
		poolToResize   *api.StoragePool
		poolIDToResize string
		contexts       = make([]*scheduler.Context, 0)
	)
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/tests/view/34542894

	BeforeEach(func() {
		StartTorpedoTest("PoolExpandAndCheckAlertsUsingResizeDisk", "pool expansion using resize-disk and check alerts after each operation", nil, testrailID)
		contexts = scheduleApps()
	})
	JustBeforeEach(func() {
		poolIDToResize = pickPoolToResize(contexts, api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK, 100)
		log.Infof("Picked pool %s to resize", poolIDToResize)
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

	It("pool expansion using resize-disk and check alerts after each operation", func() {
		log.InfoD("Initiate pool expansion using resize-disk")
		poolToResize = getStoragePool(poolIDToResize)
		originalSizeInBytes = poolToResize.TotalSize
		targetSizeInBytes = originalSizeInBytes + 100*units.GiB
		targetSizeGiB = targetSizeInBytes / units.GiB
		log.InfoD("Current Size of the pool %s is %d GiB. Trying to expand to %v GiB with type resize-disk", poolIDToResize, poolToResize.TotalSize/units.GiB, targetSizeGiB)
		triggerPoolExpansion(poolIDToResize, targetSizeGiB, api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK)

		resizeErr := waitForOngoingPoolExpansionToComplete(poolIDToResize)
		dash.VerifyFatal(resizeErr, nil, "Pool expansion does not result in error")

		log.Infof("Check the alert for pool expand for pool uuid %s", poolIDToResize)
		alertExists, _ := checkAlertsForPoolExpansion(poolIDToResize, targetSizeGiB)
		dash.VerifyFatal(alertExists, true, "Verify Alert is Present")
	})

})

func checkAlertsForPoolExpansion(poolIDToResize string, targetSizeGiB uint64) (bool, error) {
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
	substr := "[0-9]+ GiB"
	re := regexp.MustCompile(substr)

	for _, l := range outLines {
		line := strings.Trim(l, " ")
		if strings.Contains(line, "PoolExpandSuccessful") && strings.Contains(line, poolIDToResize) {
			if re.MatchString(line) {
				matchedSize := re.FindStringSubmatch(line)[0]
				poolSize := matchedSize[:len(matchedSize)-4]
				poolSizeUint, _ := strconv.ParseUint(poolSize, 10, 64)
				if poolSizeUint >= targetSizeGiB {
					log.Infof("The Alert generated is %s", line)
					return true, nil
				}
			}
		}
	}
	return false, fmt.Errorf("Alert not found")
}

func checkPoolShowMessageOutput(n *node.Node) bool {
	// Get the node to check the pool show output and grep for Message
	cmd := "pxctl sv pool show | grep -e Message"
	// Execute the command and check the alerts of type POOL
	out, err := Inst().N.RunCommandWithNoRetry(*n, cmd, node.ConnectionOpts{
		Timeout:         2 * time.Minute,
		TimeBeforeRetry: 10 * time.Second,
	})
	log.FailOnError(err, "Unable to execute the pxctl show command")
	outLines := strings.Split(out, "\n")
	for _, l := range outLines {
		line := strings.Trim(l, " ")
		// the following error is expected cause we are trying to expand the pool beyond the limit
		if strings.Contains(line, "cannot be expanded beyond maximum size") {
			return true
		}
	}
	return false
}

var _ = Describe("{CheckPoolLabelsAfterResizeDisk}", Label("p0", "positive", "pool_ops", "PoolExpand", "ResizeDisk"), func() {

	var testrailID = 34542904
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/tests/view/34542904
	var (
		poolToResize   *api.StoragePool
		poolIDToResize string
		contexts       = make([]*scheduler.Context, 0)
	)
	BeforeEach(func() {
		StartTorpedoTest("CheckPoolLabelsAfterResizeDisk",
			"Initiate pool expansion and Newly set pool labels should persist post pool expand resize-disk operation", nil, testrailID)
		contexts = scheduleApps()
	})

	JustBeforeEach(func() {
		poolIDToResize = pickPoolToResize(contexts, api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK, 100)
		log.Infof("Picked pool %s to resize", poolIDToResize)
		poolToResize = getStoragePool(poolIDToResize)
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

	It("Initiate pool expansion and Newly set pool labels should persist post pool expand resize-disk operation", func() {
		log.InfoD("set pool label, before pool expand")
		poolLabelToUpdate := make(map[string]string)
		poolLabelToUpdate["cust-type"] = "test-label"
		// Update the pool label
		err = Inst().V.UpdatePoolLabels(*storageNode, poolIDToResize, poolLabelToUpdate)
		dash.VerifyFatal(err, nil, "Check if able to update the label on the pool")
		poolToResize = getStoragePool(poolIDToResize)
		labelBeforeExpand := poolToResize.Labels

		log.InfoD("expand pool using resize-disk")
		originalSizeInBytes = poolToResize.TotalSize
		targetSizeInBytes = originalSizeInBytes + 100*units.GiB
		targetSizeGiB = targetSizeInBytes / units.GiB

		log.InfoD("Current Size of the pool %s is %d GiB. Trying to expand to %v GiB with type resize-disk",
			poolIDToResize, poolToResize.TotalSize/units.GiB, targetSizeGiB)
		triggerPoolExpansion(poolIDToResize, targetSizeGiB, api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK)

		err = waitForOngoingPoolExpansionToComplete(poolIDToResize)
		dash.VerifyFatal(err, nil, "Pool expansion does not result in error")
		verifyPoolSizeEqualOrLargerThanExpected(poolIDToResize, targetSizeGiB)

		log.InfoD("check pool label, after pool expand")
		poolToResize = getStoragePool(poolIDToResize)
		labelAfterExpand := poolToResize.Labels

		result := reflect.DeepEqual(labelBeforeExpand, labelAfterExpand)
		dash.VerifyFatal(result, true, "Check if labels changed after pool expand")
	})

})

var _ = Describe("{CheckPoolLabelsAfterAddDisk}", Label("p0", "positive", "pool_ops", "PoolExpand", "AddDrive"), func() {

	var testrailID = 34542906
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/tests/view/34542906
	var (
		poolToResize   *api.StoragePool
		poolIDToResize string
		contexts       = make([]*scheduler.Context, 0)
	)
	BeforeEach(func() {
		StartTorpedoTest("CheckPoolLabelsAfterAddDisk",
			"Initiate pool expansion and Newly set pool labels should persist post pool expand add-disk operation", nil, testrailID)
		contexts = scheduleApps()
	})

	JustBeforeEach(func() {
		poolIDToResize = pickPoolToResize(contexts, api.SdkStoragePool_RESIZE_TYPE_ADD_DISK, 100)
		log.Infof("Picked pool %s to resize", poolIDToResize)
		poolToResize = getStoragePool(poolIDToResize)
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

	It("Initiate pool expansion and Newly set pool labels should persist post pool expand add-disk operation", func() {

		labelBeforeExpand := poolToResize.Labels

		stepLog = "set pool label, before pool expand"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			poolLabelToUpdate := make(map[string]string)
			poolLabelToUpdate["cust-type1"] = "add-disk-test-label"
			err = Inst().V.UpdatePoolLabels(*storageNode, poolIDToResize, poolLabelToUpdate)
			log.FailOnError(err, "Failed to update the label on the pool %s", poolIDToResize)
		})

		stepLog = "Move pool to maintenance mode"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			expectedStatus := "In Maintenance"
			log.InfoD(fmt.Sprintf("Entering pool maintenance mode on node %s", storageNode.Name))
			err = Inst().V.EnterPoolMaintenance(*storageNode)
			log.FailOnError(err, fmt.Sprintf("failed to enter node %s in maintenance mode", storageNode.Name))
			status, _ := Inst().V.GetNodeStatus(*storageNode)
			log.InfoD(fmt.Sprintf("Node %s status %s", storageNode.Name, status.String()))
			err := WaitForPoolStatusToUpdate(*storageNode, expectedStatus)
			dash.VerifyFatal(err, nil, "Pool now in maintenance mode")
		})

		stepLog = "Initiate pool expand with add-disk operation"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			originalSizeInBytes = poolToResize.TotalSize
			targetSizeInBytes = originalSizeInBytes + 100*units.GiB
			targetSizeGiB = targetSizeInBytes / units.GiB

			log.InfoD("Current Size of the pool %s is %d GiB. Trying to expand to %v GiB with type add-disk",
				poolIDToResize, poolToResize.TotalSize/units.GiB, targetSizeGiB)
			err := Inst().V.ExpandPool(poolIDToResize, api.SdkStoragePool_RESIZE_TYPE_ADD_DISK, targetSizeGiB, true)
			dash.VerifyFatal(err, nil, "pool expansion requested successfully")
			resizeErr := waitForOngoingPoolExpansionToComplete(poolIDToResize)
			dash.VerifyFatal(resizeErr, nil, "Pool expansion does not result in error")
		})

		stepLog = "Exit pool out of maintenance mode"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			expectedStatus := "Online"
			statusMap, err := Inst().V.GetNodePoolsStatus(*storageNode)
			log.FailOnError(err, "Failed to get map of pool UUID and status")
			log.InfoD(fmt.Sprintf("pool %s has status %s", storageNode.Name, statusMap[poolToResize.Uuid]))
			if statusMap[poolToResize.Uuid] == "In Maintenance" {
				log.InfoD(fmt.Sprintf("Exiting pool maintenance mode on node %s", storageNode.Name))
				err := Inst().V.ExitPoolMaintenance(*storageNode)
				log.FailOnError(err, "failed to exit pool maintenance mode")
			} else {
				dash.VerifyFatal(statusMap[poolToResize.Uuid], "In Maintenance", "Pool is not in Maintenance mode")
			}
			status, err := Inst().V.GetNodeStatus(*storageNode)
			log.FailOnError(err, "err getting node [%s] status", storageNode.Name)
			log.Infof(fmt.Sprintf("Node %s status %s after exit", storageNode.Name, status.String()))
			exitErr := WaitForPoolStatusToUpdate(*storageNode, expectedStatus)
			dash.VerifyFatal(exitErr, nil, "Pool is now online")
		})

		stepLog = "check pool label, after pool expand"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			labelAfterExpand := poolToResize.Labels
			result := reflect.DeepEqual(labelBeforeExpand, labelAfterExpand)
			dash.VerifyFatal(result, true, "Check if labels changed after pool expand")
		})
	})

})

var _ = Describe("{PoolExpandAndCheckAlertsUsingAddDisk}", Label("p1", "positive", "pool_ops", "PoolExpand", "AddDrive"), func() {

	var testrailID = 34542894
	var (
		poolToResize   *api.StoragePool
		poolIDToResize string
		contexts       = make([]*scheduler.Context, 0)
	)

	// testrailID corresponds to: https://portworx.testrail.net/index.php?/tests/view/34542894

	BeforeEach(func() {
		StartTorpedoTest("PoolExpandAndCheckAlertsUsingAddDisk", "pool expansion using add-disk and check alerts after each operation", nil, testrailID)
		contexts = scheduleApps()
		ValidateApplications(contexts)
	})
	JustBeforeEach(func() {
		poolIDToResize = pickPoolToResize(contexts, api.SdkStoragePool_RESIZE_TYPE_ADD_DISK, 100)
		log.Infof("Picked pool %s to resize", poolIDToResize)
		poolToResize = getStoragePool(poolIDToResize)
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

	It("pool expansion using add-disk and check alerts after each operation", func() {
		var err error

		stepLog = "Move pool to maintenance mode"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			expectedStatus := "In Maintenance"
			log.InfoD(fmt.Sprintf("Entering pool maintenance mode on node %s", storageNode.Name))
			err = Inst().V.EnterPoolMaintenance(*storageNode)
			log.FailOnError(err, fmt.Sprintf("failed to enter node %s in maintenance mode", storageNode.Name))
			status, _ := Inst().V.GetNodeStatus(*storageNode)
			log.InfoD(fmt.Sprintf("Node %s status %s", storageNode.Name, status.String()))
			err := WaitForPoolStatusToUpdate(*storageNode, expectedStatus)
			dash.VerifyFatal(err, nil, "Pool now in maintenance mode")
		})

		stepLog = "Initiate pool expand with add-disk operation"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			originalSizeInBytes = poolToResize.TotalSize
			targetSizeInBytes = originalSizeInBytes + 100*units.GiB
			targetSizeGiB = targetSizeInBytes / units.GiB

			log.InfoD("Current Size of the pool %s is %d GiB. Trying to expand to %v GiB with type add-disk",
				poolIDToResize, poolToResize.TotalSize/units.GiB, targetSizeGiB)
			err := Inst().V.ExpandPool(poolIDToResize, api.SdkStoragePool_RESIZE_TYPE_ADD_DISK, targetSizeGiB, true)
			dash.VerifyFatal(err, nil, "pool expansion requested successfully")
			resizeErr := waitForOngoingPoolExpansionToComplete(poolIDToResize)
			dash.VerifyFatal(resizeErr, nil, "Pool expansion does not result in error")

		})

		stepLog = "Exit pool out of maintenance mode"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			expectedStatus := "Online"
			statusMap, err := Inst().V.GetNodePoolsStatus(*storageNode)
			log.FailOnError(err, "Failed to get map of pool UUID and status")
			log.InfoD(fmt.Sprintf("pool %s has status %s", storageNode.Name, statusMap[poolToResize.Uuid]))
			if statusMap[poolToResize.Uuid] == "In Maintenance" {
				log.InfoD(fmt.Sprintf("Exiting pool maintenance mode on node %s", storageNode.Name))
				err := Inst().V.ExitPoolMaintenance(*storageNode)
				log.FailOnError(err, "failed to exit pool maintenance mode")
			} else {
				dash.VerifyFatal(statusMap[poolToResize.Uuid], "In Maintenance", "Pool is not in Maintenance mode")
			}
			status, err := Inst().V.GetNodeStatus(*storageNode)
			log.FailOnError(err, "err getting node [%s] status", storageNode.Name)
			log.Infof(fmt.Sprintf("Node %s status %s after exit", storageNode.Name, status.String()))
			exitErr := WaitForPoolStatusToUpdate(*storageNode, expectedStatus)
			dash.VerifyFatal(exitErr, nil, "Pool is now online")
		})

		stepLog = "Check the alerts for pool expand"
		Step(stepLog, func() {
			log.Infof("Check the alert for pool expand for pool uuid %s", poolIDToResize)
			alertExists, _ := checkAlertsForPoolExpansion(poolIDToResize, targetSizeGiB)
			dash.VerifyFatal(alertExists, true, "Verify Alert is Present")
		})
	})

})

var _ = Describe("{PoolVolUpdateResizeDisk}", Label("p0", "positive", "pool_ops", "PoolExpand", "ResizeDisk"), func() {

	// 1.The volumes originally have a HA of 2. We are adding a check at first to see if the HA is 3.
	// 2.If it is, we decrease it to 2 and then increase it back to 3 again.
	// 3.If not, then we are good to increase it by 1.
	// 4.While this increase is happening, we do the pool resize operation.

	var testrailID = 34542876
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/tests/view/34542876

	var (
		poolIDToResize string
		contexts       = make([]*scheduler.Context, 0)
	)
	BeforeEach(func() {
		StartTorpedoTest("PoolVolUpdateResizeDisk", "Increase the HA replica of the volume and expand pool using resize-disk during the increase", nil, testrailID)
		contexts = scheduleApps()
	})
	JustBeforeEach(func() {
		poolIDToResize = pickPoolToResize(contexts, api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK, 100)
		log.Infof("Picked pool %s to resize", poolIDToResize)
		//poolToResize = getStoragePool(poolIDToResize)
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
	It("Increase the HA replica of the volume and expand pool using resize-disk during the increase", func() {
		var (
			newRep       int64
			currRep      int64
			poolToResize *api.StoragePool
		)
		volSelected, err := GetVolumeWithMinimumSize(contexts, 10)
		log.FailOnError(err, "error identifying volume")
		opts := volume.Options{
			ValidateReplicationUpdateTimeout: replicationUpdateTimeout,
		}
		Step("Increase the HA replica of the volume", func() {
			currRep, err = Inst().V.GetReplicationFactor(volSelected)
			log.FailOnError(err, fmt.Sprintf("err getting repl factor for  vol : %s", volSelected.Name))
			newRep = currRep
			// If the HA is 3, reduce it by 1, so that we can increase it later
			if currRep == 3 {
				newRep = currRep - 1
				err = Inst().V.SetReplicationFactor(volSelected, newRep, nil, nil, true, opts)
				log.FailOnError(err, fmt.Sprintf("err setting repl factor  to %d for  vol : %s", newRep, volSelected.Name))

				decreasedRep, err := Inst().V.GetReplicationFactor(volSelected)
				log.FailOnError(err, fmt.Sprintf("err getting repl factor for  vol : %s", volSelected.Name))
				dash.VerifyFatal(decreasedRep == newRep, true, fmt.Sprintf("repl factor successfully decreased to %d for the vol : %s ", decreasedRep, volSelected.Name))
			}
			// Increase the HA by 1
			log.InfoD(fmt.Sprintf("setting repl factor to %d for vol : %s", newRep+1, volSelected.Name))
			// err = Inst().V.SetReplicationFactor(volSelected, newRep+1, []string{storageNode.Id}, []string{poolToResize.Uuid}, false, opts)
			err = Inst().V.SetReplicationFactor(volSelected, newRep+1, nil, nil, true, opts)
			log.FailOnError(err, fmt.Sprintf("err setting repl factor  to %d for  vol : %s", newRep+1, volSelected.Name))
			dash.VerifyFatal(err == nil, true, fmt.Sprintf("vol %s expansion triggered successfully on node %s", volSelected.Name, storageNode.Name))
		})

		Step("Initiate pool expansion using resize-disk while repl increase is in progress", func() {
			originalSizeInBytes = poolToResize.TotalSize
			targetSizeInBytes = originalSizeInBytes + 100*units.GiB
			targetSizeGiB = targetSizeInBytes / units.GiB
			log.InfoD("Current Size of the pool %s is %d GiB. Trying to expand to %v GiB with type resize-disk", poolIDToResize, poolToResize.TotalSize/units.GiB, targetSizeGiB)
			triggerPoolExpansion(poolIDToResize, targetSizeGiB, api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK)

			log.InfoD("Wait for expansion to finish")
			resizeErr := waitForOngoingPoolExpansionToComplete(poolIDToResize)
			dash.VerifyFatal(resizeErr, nil, "Pool expansion does not result in error")
			err = ValidateReplFactorUpdate(volSelected, newRep+1)
			log.FailOnError(err, "error validating repl factor for vol [%s]", volSelected.Name)

			//reverting the replication for volume validation
			if currRep < 3 {
				err = Inst().V.SetReplicationFactor(volSelected, currRep, nil, nil, true, opts)
				log.FailOnError(err, fmt.Sprintf("err setting repl factor to %d for vol : %s", newRep, volSelected.Name))
			}

		})
	})
})

var _ = Describe("{PoolExpandStorageFullPoolResize}", Label("p1", "positive", "pool_ops", "px_ops", "PoolExpand", "Throttling"), func() {

	//step1: feed p1 size GB I/O on the volume
	//step2: After I/O done p1 should be offline and full, expand the pool p1 using resize-disk
	//step4: validate the pool and the data

	var testrailID = 34542835
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/tests/view/34542835

	JustBeforeEach(func() {
		StartTorpedoTest("PoolExpandStorageFullPoolResize", "Feed a pool full, then expand the pool using resize-disk", nil, testrailID)
	})

	var contexts []*scheduler.Context

	stepLog := "Create vols and make pool full"
	It(stepLog, func() {
		log.InfoD(stepLog)
		selectedNode := GetNodeWithLeastSize()

		stNodes := node.GetStorageNodes()
		var secondReplNode node.Node
		for _, stNode := range stNodes {
			if stNode.Name != selectedNode.Name {
				secondReplNode = stNode
			}
		}

		applist := Inst().AppList
		var err error
		defer func() {
			Inst().AppList = applist
			err = Inst().S.RemoveLabelOnNode(*selectedNode, k8s.NodeType)
			log.FailOnError(err, "error removing label on node [%s]", selectedNode.Name)
			err = Inst().S.RemoveLabelOnNode(secondReplNode, k8s.NodeType)
			log.FailOnError(err, "error removing label on node [%s]", secondReplNode.Name)
		}()
		err = Inst().S.AddLabelOnNode(*selectedNode, k8s.NodeType, k8s.FastpathNodeType)
		log.FailOnError(err, fmt.Sprintf("Failed add label on node %s", selectedNode.Name))
		err = Inst().S.AddLabelOnNode(secondReplNode, k8s.NodeType, k8s.FastpathNodeType)
		log.FailOnError(err, fmt.Sprintf("Failed add label on node %s", secondReplNode.Name))

		isjournal, err := IsJournalEnabled()
		log.FailOnError(err, "is journal enabled check failed")

		err = adjustReplPools(*selectedNode, secondReplNode, isjournal)
		log.FailOnError(err, "Error setting pools for clean volumes")

		Inst().AppList = []string{"fio-fastpath"}
		contexts = make([]*scheduler.Context, 0)
		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("sfullrz-%d", i))...)
		}
		ValidateApplications(contexts)
		defer appsValidateAndDestroy(contexts)

		err = WaitForPoolOffline(*selectedNode)
		log.FailOnError(err, fmt.Sprintf("Failed to make node %s storage down", selectedNode.Name))

		poolsStatus, err := Inst().V.GetNodePoolsStatus(*selectedNode)
		log.FailOnError(err, "error getting pool status on node %s", selectedNode.Name)

		var offlinePoolUUID string
		for i, s := range poolsStatus {
			if s == "Offline" {
				offlinePoolUUID = i
				break
			}
		}
		selectedPool, err := GetStoragePoolByUUID(offlinePoolUUID)
		log.FailOnError(err, "error getting pool with UUID [%s]", offlinePoolUUID)

		var expandedExpectedPoolSize uint64
		Step(stepLog, func() {
			log.InfoD(stepLog)
			expandedExpectedPoolSize = (selectedPool.TotalSize / units.GiB) * 2
			log.InfoD("Current Size of the pool %s is %d", selectedPool.Uuid, selectedPool.TotalSize/units.GiB)
			err = Inst().V.ExpandPool(selectedPool.Uuid, api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK, expandedExpectedPoolSize, true)
			dash.VerifyFatal(err, nil, "Pool expansion init successful?")
		})
		stepLog = fmt.Sprintf("Ensure that pool %s expansion is successful", selectedPool.Uuid)
		Step(stepLog, func() {
			log.InfoD(stepLog)

			err = waitForPoolToBeResized(expandedExpectedPoolSize, selectedPool.Uuid, isjournal)
			log.FailOnError(err, fmt.Sprintf("Error waiting for poor %s resize", selectedPool.Uuid))
			resizedPool, err := GetStoragePoolByUUID(selectedPool.Uuid)
			log.FailOnError(err, fmt.Sprintf("error get pool using UUID %s", selectedPool.Uuid))
			newPoolSize := resizedPool.TotalSize / units.GiB
			isExpansionSuccess := false
			expectedSizeWithJournal := expandedExpectedPoolSize - 3

			if newPoolSize >= expectedSizeWithJournal {
				isExpansionSuccess = true
			}
			dash.VerifyFatal(isExpansionSuccess, true, fmt.Sprintf("expected new pool size to be %v or %v, got %v", expandedExpectedPoolSize, expectedSizeWithJournal, newPoolSize))
			status, err := Inst().V.GetNodeStatus(*selectedNode)
			log.FailOnError(err, fmt.Sprintf("Error getting PX status of node %s", selectedNode.Name))
			dash.VerifySafely(*status, api.Status_STATUS_OK, fmt.Sprintf("validate PX status on node %s", selectedNode.Name))
		})
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
		// Cores are expected with StoragePool full.
		// AfterEachTest will fail due to cores found during the test.
		// AfterEachTest(contexts, testrailID, runID)
	})
})

var _ = Describe("{DriveAddDifferentTypesAndResize}", Label("p1", "positive", "pool_ops", "px_ops", "PoolExpand", "AddDrive"), func() {

	var testrailID = 34542903
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/tests/view/34542903
	var (
		poolIDToResize string
		contexts       = make([]*scheduler.Context, 0)
	)
	BeforeEach(func() {
		StartTorpedoTest("DriveAddDifferentTypesAndResize",
			"Create pools with different types of drive and pool expand using resize-disk", nil, testrailID)
		contexts = scheduleApps()
	})

	JustBeforeEach(func() {
		poolIDToResize = pickPoolToResize(contexts, api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK, 100)
		log.Infof("Picked pool %s to resize", poolIDToResize)
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

	It("creating pools with different drive types and resizing them", func() {
		var (
			driveTypes   []string
			poolToResize *api.StoragePool
		)

		driveSize := "100"
		driveTypes, err = Inst().N.GetSupportedDriveTypes()
		log.FailOnError(err, "Error getting drive types for the provider")
		for i := 0; i < len(driveTypes); i++ {

			poolsBfr, err := Inst().V.ListStoragePools(metav1.LabelSelector{})
			log.FailOnError(err, "Failed to list storage pools")

			newDriveSpec := fmt.Sprintf("size=%s,type=%s", driveSize, driveTypes[i])
			err = Inst().V.AddCloudDrive(storageNode, newDriveSpec, -1)
			log.FailOnError(err, fmt.Sprintf("Add cloud drive failed on node %s", storageNode.Name))

			log.InfoD("Validate pool rebalance after drive add to the node %s", storageNode.Name)
			err = ValidateDriveRebalance(*storageNode)
			log.FailOnError(err, "pool re-balance failed on node %s", storageNode.Name)

			err = Inst().V.WaitDriverUpOnNode(*storageNode, addDriveUpTimeOut)
			log.FailOnError(err, "volume driver down on node %s", storageNode.Name)

			poolsAfr, err := Inst().V.ListStoragePools(metav1.LabelSelector{})
			log.FailOnError(err, "Failed to list storage pools")
			dash.VerifyFatal(len(poolsBfr)+1, len(poolsAfr), "verify new pool is created")

		}

		allPools, err := GetAllPoolsOnNode(storageNode.Id)
		log.FailOnError(err, "Unable to get pool drives on node %s", storageNode.Name)

		for i := 0; i < len(allPools); i++ {
			poolToResize = getStoragePool(allPools[i])
			originalSizeInBytes = poolToResize.TotalSize
			targetSizeInBytes = originalSizeInBytes + 100*units.GiB
			targetSizeGiB = targetSizeInBytes / units.GiB
			log.InfoD("Current Size of the pool %s is %d GiB. Trying to expand to %v GiB with type resize-disk", allPools[i], poolToResize.TotalSize/units.GiB, targetSizeGiB)
			triggerPoolExpansion(allPools[i], 1000, api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK)
			resizeErr := waitForOngoingPoolExpansionToComplete(allPools[i])
			dash.VerifyFatal(resizeErr, nil, "Pool expansion does not result in error")
		}

	})
})

var _ = Describe("{PoolResizeWithPXRestartWithTimeInterval}", Label("p1", "hal_ops_disruption", "px_restart", "PoolExpand", "functional"), func() {

	/*
	   1.  Create volume and do IOs / deploy apps to do IOs
	   2.  Select a pool to resize
	   3.  Expand selected pool with resize-disk type
	   4.  Restart Portworx with some random delay
	   5.  Wait for pool to expand
	   6.  Verify pool resized
	   7.  Check px status
	   8.  Check if apps are running
	*/

	testName = "PoolResizeWithPXRestartWithTimeInterval"
	testDescription = "Pool resize with px restart with random delay"
	PoolResize(testName, testDescription)
})

var _ = Describe("{AddNewPoolWithPxRestart}", Label("p1", "hal_ops_disruption", "px_restart", "AddNewPool", "functional"), func() {
	/*
			Install portworx in a 5 node cluster
		    Select a node and add drive to create new pool (pxctl sv drive add --newpool)
			With some random delay force restart the portworx (systemctl restart portworx)
		    After node comes up verify px status and check if pool is successfully added and apps are running.
	*/

	JustBeforeEach(func() {
		StartTorpedoTest("AddNewPoolWithPxRestart", "Automate drive add newpool with px restart", nil, 0)
	})

	var (
		storageNodes                                    []node.Node
		selectedNode                                    node.Node
		newPoolID                                       string
		nodePoolMapBfrAddDrive, nodePoolMapAftrAddDrive map[string]string
	)

	itLog := "AddNewPoolWithPxRestart"
	It(itLog, func() {

		storageNodes = node.GetStorageNodes()
		index := rand.Intn(len(storageNodes))
		selectedNode = storageNodes[index]

		log.Info("selected Node ID - %s , Name - %s", selectedNode.Id, selectedNode.Name)

		nodePoolMapBfrAddDrive, err = Inst().V.GetNodePools(selectedNode)
		log.FailOnError(err, fmt.Sprintf("Get Node pools failed on node %s", selectedNode.Name))

		log.Info("Number of Pools available - %d", len(nodePoolMapBfrAddDrive))

		if len(nodePoolMapBfrAddDrive) >= 6 {
			Skip("Skipping the test as there can be a maximum of 6 pools allowed to be present in a node")
		}

		log.Info("list Pool UUIDs - %v", nodePoolMapBfrAddDrive)

		//Get cloudrive spec
		driveSpecs, err := GetCloudDriveDeviceSpecs()
		log.FailOnError(err, "Error getting cloud drive specs")

		deviceSpec := driveSpecs[0]

		deviceSpecParams := strings.Split(deviceSpec, ",")
		paramsArr := make([]string, 0)
		for _, param := range deviceSpecParams {
			if strings.Contains(param, "size") {
				paramsArr = append(paramsArr, fmt.Sprintf("size=%d,", 128))
			} else {
				paramsArr = append(paramsArr, param)
			}
		}
		//drive spec generated from actual cloudrive spec
		newSpec := strings.Join(paramsArr, ",")

		stepLog = "Add drive to create newpool"
		Step(stepLog, func() {
			log.Info(stepLog)
			err = Inst().V.AddCloudDrive(&selectedNode, newSpec, -1)
			log.FailOnError(err, fmt.Sprintf("Add cloud drive failed on node %s", selectedNode.Name))
		})

		nodePoolMapAftrAddDrive, err = Inst().V.GetNodePools(selectedNode)
		log.FailOnError(err, fmt.Sprintf("Get Node pools failed on node %s", selectedNode.Name))

		//random delay in secs
		sleepTime := rand.Intn(60-1) + 1
		time.Sleep(time.Second * (time.Duration(sleepTime)))

		stepLog = "Restart Portworx"
		//Restart portworx and wait for it to come up
		Step(stepLog, func() {
			log.Info(stepLog)

			if selectedNode.IsStorageDriverInstalled {
				Step(fmt.Sprintf("node with Px restart is: %s", selectedNode.Name), func() {

					err := Inst().V.RestartDriver(selectedNode, nil)
					log.FailOnError(err, fmt.Sprintf("Error occured while Restart PX on node:%v", selectedNode.Name))
				})

				Step(fmt.Sprintf("wait for volume driver to restart on node: %v", selectedNode.Name), func() {
					err := Inst().V.WaitForPxPodsToBeUp(selectedNode)
					log.FailOnError(err, fmt.Sprintf("Error occured while Validating PX restart is done on node:%v", selectedNode.Name))
				})
			}

		})

		stepLog = "verify the Px status"
		Step(stepLog, func() {
			log.Info(stepLog)
			err = Inst().V.WaitDriverUpOnNode(selectedNode, 10*time.Minute)
			log.FailOnError(err, fmt.Sprintf("Driver is down on node %s", selectedNode.Name))
			pxReady := Inst().V.IsPxReadyOnNode(selectedNode)
			dash.VerifyFatal(pxReady, true, fmt.Sprintf("expected Px status response to be true but received false"))

		})

		stepLog = "verify pool is added to the node"
		Step(stepLog, func() {
			log.Info(stepLog)
			dash.VerifyFatal(len(nodePoolMapAftrAddDrive) > len(nodePoolMapBfrAddDrive), true, fmt.Sprintf("expecting the pool count to be greater than the count before adding new pool . Expected %d , but received - %d", len(nodePoolMapBfrAddDrive)+1, len(nodePoolMapAftrAddDrive)))

			//Identifying the newpool
			for uuid, id := range nodePoolMapAftrAddDrive {
				if _, ok := nodePoolMapBfrAddDrive[uuid]; !ok {
					newPoolID = id
				}
			}

			log.Info("Newly Added pool ID - %d ", newPoolID)
		})

	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
		if len(nodePoolMapAftrAddDrive) > len(nodePoolMapBfrAddDrive) {
			stepLog = "Delete the new pool"
			Step(stepLog, func() {
				log.Info(stepLog)
				err = DeletePoolAndValidate(selectedNode, newPoolID)
				log.FailOnError(err, fmt.Sprintf("Error occured while Validating the deleted pool %s in the node %s", newPoolID, selectedNode.Name))
			})
		}
	})
})

var _ = Describe("{DriveAddAsJournalWithNodeReboot}", Label("p1", "hal_ops_disreption", "node_reboot", "AddJournal", "functional"), func() {
	/*
		Install portworx in a 5 node cluster
		Deploy apps to do IOs
		Select a node and add journal drive
		With some random delay force reboot the node
		After node comes up verify px status and check if journal drive is successfully added and apps are running.
		Testrail test case
		https://portworx.testrail.net/index.php?/cases/view/301928
		https://portworx.testrail.net/index.php?/cases/view/301929
	*/

	var (
		nodeDetail        *node.Node
		poolUUID          string
		blockDeviceBefore int
		systemOpts        node.SystemctlOpts
		contexts          = make([]*scheduler.Context, 0)
	)

	JustBeforeEach(func() {
		StartTorpedoTest("DriveAddAsJournalWithNodeReboot",
			"Add drive when as journal and perform node reboot",
			nil, 0)
	})

	stepLog := "Automate drive add journal with node reboot"
	It(stepLog, func() {
		log.InfoD(stepLog)
		isjournal, err := IsJournalEnabled()
		log.FailOnError(err, "Error getting journal status")

		if isjournal {
			log.Info("Journal drive already exists")
			Skip("Skipping the test DriveAddAsJournalWithPXRestart as journal drive already exists")
		}

		stepLog = "Schedule apps"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			contexts = make([]*scheduler.Context, 0)
			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				contexts = append(contexts, ScheduleApplications(fmt.Sprintf("adddriveasjournal-%d", i))...)
			}
			ValidateApplications(contexts)
		})
		defer appsValidateAndDestroy(contexts)

		stepLog = "Select a node to add journal drive"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			// Get Pool with running IO on the cluster
			poolUUID = pickPoolToResize(contexts, api.SdkStoragePool_RESIZE_TYPE_AUTO, 0)
			log.InfoD("Pool UUID on which IO is running [%s]", poolUUID)

			// Get Node Details of the Pool with IO
			nodeDetail, err = GetNodeWithGivenPoolID(poolUUID)
			log.FailOnError(err, "Failed to get Node Details from PoolUUID [%v]", poolUUID)
			log.InfoD("Pool with UUID [%v] present in Node [%v]", poolUUID, nodeDetail.Name)
		})

		// Add cloud drive on the selected node
		driveSpecs, err := GetCloudDriveDeviceSpecs()
		log.FailOnError(err, "Error getting cloud drive specs")

		deviceSpec := driveSpecs[0]
		devicespecjournal := deviceSpec + " --journal"

		stepLog = "Add journal drive"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			//enter pool maintenance mode
			err = Inst().V.EnterPoolMaintenance(*nodeDetail)
			log.FailOnError(err, "Error Entering Maintenance mode on Node[%v]", nodeDetail.Name)
			log.InfoD("Enter pool Maintenance mode ")
			expectedStatus := "In Maintenance"

			log.FailOnError(WaitForPoolStatusToUpdate(*nodeDetail, expectedStatus),
				fmt.Sprintf("node %s pools are not in status %s", nodeDetail.Name, expectedStatus))

			systemOpts = node.SystemctlOpts{
				ConnectionOpts: node.ConnectionOpts{
					Timeout:         2 * time.Minute,
					TimeBeforeRetry: defaultRetryInterval,
				},
				Action: "start",
			}
			drivesMap, err := Inst().N.GetBlockDrives(*nodeDetail, systemOpts)
			log.FailOnError(err, "error getting block drives from node %s", nodeDetail.Name)
			blockDeviceBefore = len(drivesMap)
			err = Inst().V.AddCloudDrive(nodeDetail, devicespecjournal, -1)
			log.FailOnError(err, "journal add failed")
		})

		//random delay in secs
		sleepTime := rand.Intn(60-1) + 1
		time.Sleep(time.Second * (time.Duration(sleepTime)))

		stepLog := "Reboot the node"
		//Reboot node and wait for PX to come up
		Step(stepLog, func() {
			log.Info(stepLog)
			err = RebootNodeAndWaitForPxUp(*nodeDetail)
			log.FailOnError(err, "Failed to reboot node and wait till it is up")

		})

		stepLog = "Exit pool maintenance mode"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			err = ExitPoolMaintenance(*nodeDetail)
			log.FailOnError(err, "Failed to exit maintenance mode")
			log.Info("exit pool maintenance mode succeed")
		})

		stepLog = "Verify Px Status"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			status, err := Inst().V.GetPxctlStatus(*nodeDetail)
			log.FailOnError(err, fmt.Sprintf("failed to get pxctl status on node [%s]", nodeDetail.Name))
			dash.VerifyFatal(status == api.Status_STATUS_OK.String(), true, fmt.Sprintf("node [%s] status is up but PX cluster is not ok. Expected: %v Actual: %v",
				nodeDetail.Name, api.Status_STATUS_OK, status))
			log.InfoD("px status %v", status)
		})

		stepLog = "Verify if journal is enabled"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			jDev, err := Inst().V.GetJournalDevicePath(nodeDetail)
			log.FailOnError(err, fmt.Sprintf("error getting journal device path from node %s", nodeDetail.Name))
			if jDev == "" {
				log.FailOnError(fmt.Errorf("no journal device path found"), "error getting journal device path from storage spec")
			}
			log.Infof("JournalDev path : %s", jDev)

			err = Inst().V.RefreshDriverEndpoints()
			log.FailOnError(err, "error refreshing driver end points")
		})

		stepLog = "Verify drive is added successfully"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			drivesMap, err := Inst().N.GetBlockDrives(*nodeDetail, systemOpts)
			log.FailOnError(err, "error getting block drives from node %s", nodeDetail.Name)
			blockDeviceAfter := len(drivesMap)
			dash.VerifyFatal(blockDeviceAfter > blockDeviceBefore, true, "adding cloud drive as journal successful")
		})
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts)
	})
})

var _ = Describe("{DriveAddAsJournalWithNodeMaintenanceCycle}", Label("p1", "hal_ops_disreption", "NodeMaintenance", "AddJournal", "functional"), func() {
	/*
		Install portworx in a 5 node cluster
		Deploy apps to do IOs
		Select a node and add journal drive
		With some random delay start node maintenance cycle
		After node comes up verify px status and check if journal drive is successfully added and apps are running.
		Testrail test cases

		https://portworx.testrail.net/index.php?/cases/view/301930
		https://portworx.testrail.net/index.php?/cases/view/301931
	*/

	var (
		nodeDetail        *node.Node
		poolUUID          string
		blockDeviceBefore int
		systemOpts        node.SystemctlOpts
		contexts          = make([]*scheduler.Context, 0)
	)

	JustBeforeEach(func() {
		StartTorpedoTest("DriveAddAsJournalWithNodeMaintenanceCycle",
			"Add drive when as journal and perform node maintenance cycle",
			nil, 0)
	})

	stepLog := "Automate drive add journal with node maintenance cycle"
	It(stepLog, func() {
		log.InfoD(stepLog)

		isjournal, err := IsJournalEnabled()
		log.FailOnError(err, "Failed to check if Journal enabled")

		if isjournal {
			log.Info("Journal drive already exists")
			Skip("Skipping the test DriveAddAsJournalWithPXRestart as journal drive already exists")
		}

		stepLog = "Schedule apps"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			contexts = make([]*scheduler.Context, 0)
			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				contexts = append(contexts, ScheduleApplications(fmt.Sprintf("adddriveasjournal-%d", i))...)
			}
			ValidateApplications(contexts)
		})
		defer appsValidateAndDestroy(contexts)

		stepLog = "Select a node to add journal drive"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			// Get Pool with running IO on the cluster
			poolUUID = pickPoolToResize(contexts, api.SdkStoragePool_RESIZE_TYPE_AUTO, 0)
			log.InfoD("Pool UUID on which IO is running [%s]", poolUUID)

			// Get Node Details of the Pool with IO
			nodeDetail, err = GetNodeWithGivenPoolID(poolUUID)
			log.FailOnError(err, "Failed to get Node Details from PoolUUID [%v]", poolUUID)
			log.InfoD("Pool with UUID [%v] present in Node [%v]", poolUUID, nodeDetail.Name)
		})

		// Add cloud drive on the selected node
		driveSpecs, err := GetCloudDriveDeviceSpecs()
		log.FailOnError(err, "Error getting cloud drive specs")

		deviceSpec := driveSpecs[0]
		devicespecjournal := deviceSpec + " --journal"

		stepLog = "Add journal drive"
		Step(stepLog, func() {
			log.InfoD(stepLog)

			//enter pool maintenance mode
			err = Inst().V.EnterPoolMaintenance(*nodeDetail)
			log.FailOnError(err, "Error Entering Maintenance mode on Node[%v]", nodeDetail.Name)
			log.InfoD("Enter pool Maintenance mode ")
			expectedStatus := "In Maintenance"

			log.FailOnError(WaitForPoolStatusToUpdate(*nodeDetail, expectedStatus),
				fmt.Sprintf("node %s pools are not in status %s", nodeDetail.Name, expectedStatus))

			systemOpts = node.SystemctlOpts{
				ConnectionOpts: node.ConnectionOpts{
					Timeout:         2 * time.Minute,
					TimeBeforeRetry: defaultRetryInterval,
				},
				Action: "start",
			}
			drivesMap, err := Inst().N.GetBlockDrives(*nodeDetail, systemOpts)
			log.FailOnError(err, "error getting block drives from node %s", nodeDetail.Name)
			blockDeviceBefore = len(drivesMap)
			err = Inst().V.AddCloudDrive(nodeDetail, devicespecjournal, -1)
			log.FailOnError(err, "journal add failed")
		})

		//random delay in secs
		sleepTime := rand.Intn(60-1) + 1
		time.Sleep(time.Second * (time.Duration(sleepTime)))

		stepLog := "Perform node maintenance cycle"
		Step(stepLog, func() {
			log.Info(stepLog)

			log.InfoD(fmt.Sprintf("Performing node maintenance cycle on node %s", nodeDetail.Name))
			err = Inst().V.RecoverDriver(*nodeDetail)
			log.FailOnError(err, fmt.Sprintf("error performing maintenance cycle on node %s", nodeDetail.Name))

			err = Inst().V.WaitDriverUpOnNode(*nodeDetail, 5*time.Minute)
			log.FailOnError(err, fmt.Sprintf("Driver is down on node %s", nodeDetail.Name))
			dash.VerifyFatal(err == nil, true, fmt.Sprintf("PX is up after maintenance cycle on node %s", nodeDetail.Name))
		})

		stepLog = "Exit pool maintenance mode"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			err = ExitPoolMaintenance(*nodeDetail)
			log.FailOnError(err, "Failed to exit maintenance mode")
			log.Info("exit pool maintenance mode succeed")
		})

		stepLog = "Verify Px Status"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			status, err := Inst().V.GetPxctlStatus(*nodeDetail)
			log.FailOnError(err, fmt.Sprintf("failed to get pxctl status on node [%s]", nodeDetail.Name))
			dash.VerifyFatal(status == api.Status_STATUS_OK.String(), true, fmt.Sprintf("node [%s] status is up but PX cluster is not ok. Expected: %v Actual: %v",
				nodeDetail.Name, api.Status_STATUS_OK, status))
			log.InfoD("px status %v", status)
		})

		stepLog = "Verify if journal is enabled"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			jDev, err := Inst().V.GetJournalDevicePath(nodeDetail)
			log.FailOnError(err, fmt.Sprintf("error getting journal device path from node %s", nodeDetail.Name))
			if jDev == "" {
				log.FailOnError(fmt.Errorf("no journal device path found"), "error getting journal device path from storage spec")
			}
			log.Infof("JournalDev path : %s", jDev)

			err = Inst().V.RefreshDriverEndpoints()
			log.FailOnError(err, "error refreshing driver end points")
		})

		stepLog = "Verify drive is added successfully"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			drivesMap, err := Inst().N.GetBlockDrives(*nodeDetail, systemOpts)
			log.FailOnError(err, "error getting block drives from node %s", nodeDetail.Name)
			blockDeviceAfter := len(drivesMap)
			dash.VerifyFatal(blockDeviceAfter > blockDeviceBefore, true, "adding cloud drive as journal successful")
		})
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts)
	})
})

var _ = Describe("{PoolExpandAddDriveWithPXRestart}", func() {
	/*
		1. Create volume and do IOs / deploy apps to do IOs
		2. Expand pool by adding drive
		3. After some randon time delay Restart the Portworx
		4. Wait for node to comeup and check PX status
		5. Verify pool expanded successfully
	*/

	testName = "PoolExpandAddDriveWithPXRestart"
	testDescription = "Pool Expand Add Drive With PX Restart"
	ExpandPoolWithAddDrive(testName, testDescription)

})

var _ = Describe("{AddMetadataDriveWithPXRestart}", Label("p1", "hal_ops_disruption", "px_restart", "AddMetadata", "functional"), func() {

	var (
		selectedNode node.Node
		contexts     []*scheduler.Context
		kvdbNodesIDs []string
		path         string
	)

	JustBeforeEach(func() {
		StartTorpedoTest("AddMetadataDriveWithPXRestart", "Add Metadata Drive With PX Restart", nil, 0)
	})

	itLog := "AddMetadataDriveWithPXRestart"
	It(itLog, func() {
		stepLog := "Schedule Apps"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			contexts = scheduleApps()
			time.Sleep(5 * time.Minute)
			log.InfoD("schedule app succeed")
		})
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

		stepLog = "Check which node has a metadata disk if not add one"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			isDedicatedMetadataDiskExist := false
			//Check which node has metadata disk if not add one
			for _, storageNode := range storageNodes {

				path, err := getMetaDataDiskPath(storageNode)
				log.FailOnError(err, "Failed to get metadata disk")
				if path != "" {
					log.InfoD("Metadata disk path: %v", path)
					isDedicatedMetadataDiskExist = true
					tNode = storageNode
					break
				}
			}
			if !isDedicatedMetadataDiskExist {
				for _, storageNode := range storageNodes {
					deviceSpec := fmt.Sprintf("size=100 --metadata")
					log.InfoD("Initiate add cloud drive and validate")
					// enter pool maintenance mode
					if Contains(kvdbNodesIDs, storageNode.Id) {
						log.InfoD("[%s] is kvdb node", storageNode.Hostname)
						continue
					}

					stepLog := "Enter maintenance mode"
					Step(stepLog, func() {
						log.InfoD(stepLog)
						err = Inst().V.EnterPoolMaintenance(storageNode)
						log.FailOnError(err, "node: %v failed to transition to pool maintenance mode", storageNode.Name)
						log.Info("enter pool maintenance mode succeed")
					})

					stepLog = "Add metadata disk"
					Step(stepLog, func() {
						log.InfoD(stepLog)
						err := Inst().V.AddCloudDrive(&storageNode, deviceSpec, -1)
						log.FailOnError(err, "Failed to add metadata device on node : %s", storageNode.Name)
						log.InfoD("metadata disk added successfully on node [%s]", storageNode.Hostname)
					})
					sleepTime := rand.Intn(100) + 1
					time.Sleep(time.Second * (time.Duration(sleepTime)))

					stepLog = fmt.Sprintf("Restart Portworx after [%d] seconds", sleepTime)
					stepLog = "Restart Portworx"
					Step(stepLog, func() {
						log.InfoD(stepLog)
						err := Inst().N.Systemctl(storageNode, "portworx.service", node.SystemctlOpts{
							Action: "restart",
							ConnectionOpts: node.ConnectionOpts{
								Timeout:         5 * time.Minute,
								TimeBeforeRetry: 10 * time.Second,
							}})
						log.FailOnError(err, fmt.Sprintf("restart portworx failed on node [%v]", storageNode.Name))
						log.Info("portworx restart succeed")
						time.Sleep(2 * time.Minute)
					})

					// exit pool maintenance
					stepLog = "Exit pool maintenance mode"
					Step(stepLog, func() {
						log.InfoD(stepLog)
						err = Inst().V.ExitPoolMaintenance(storageNode)
						log.FailOnError(err, "Node: %v Failed to exit out of maintenance mode", storageNode.Name)
						log.Info("exit pool maintenance mode succeed")
					})

					selectedNode = storageNode
					break
				}
			} else {
				log.InfoD("Metadata disk already exist: [%s]", path)
				Skip("Metadata disk already exist")
			}
			//check if selecteNode is empty or not
			if selectedNode.Name == "" {
				log.FailOnError(fmt.Errorf("No node found with metadata disk or metadata disks cannot be added to any nodes"), "No node found with metadata disk ")
			}
		})

		//Check PX status
		stepLog = "Check PX status"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			status, err := Inst().V.GetPxctlStatus(selectedNode)
			log.FailOnError(err, fmt.Sprintf("failed to get pxctl status on node [%s]", selectedNode.Name))
			dash.VerifyFatal(status == api.Status_STATUS_OK.String(), true, fmt.Sprintf("node [%s] status is up but PX cluster is not ok. Expected: %v Actual: %v",
				selectedNode.Name, api.Status_STATUS_OK, status))
			log.Infof("px status [%v]", status)

		})

		stepLog = fmt.Sprintf("Check if apps are running")
		Step(stepLog, func() {
			log.InfoD(stepLog)
			ValidateApplications(contexts)
			log.Info("validate application succeed")
		})

		stepLog = "validate metadata device has been added successfully"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			//get metadata device path from pxctl status
			metadataDevicePath, err := getMetaDataDiskPath(selectedNode)
			log.FailOnError(err, fmt.Sprintf("failed to get pxctl status on node [%s]", selectedNode.Name))
			if metadataDevicePath == "" {
				log.FailOnError(fmt.Errorf("metadata device not added"), "metadata device not added")
			}
			log.InfoD("Metadata device path from pxctl : %v", metadataDevicePath)
		})

	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		appsValidateAndDestroy(contexts)
		AfterEachTest(contexts)
	})
})

var _ = Describe("{DriveAddAsJournalWithPXRestart}", Label("p1", "hal_ops_disreption", "px_restart", "AddJournal", "functional"), func() {
	/*
		1. Install portworx in a 5 node cluster
		2. Deploy apps to do IOs
		3. Select a node and add journal drive
		4. With some random delay force restart the portworx
		5. After node comes up verify px status and check if journal drive is successfully added and apps are running.

		Testrail test cases

		https://portworx.testrail.net/index.php?/cases/view/301926
		https://portworx.testrail.net/index.php?/cases/view/301927
	*/

	var (
		nodeDetail        *node.Node
		poolUUID          string
		blockDeviceBefore int
		systemOpts        node.SystemctlOpts
		contexts          = make([]*scheduler.Context, 0)
	)

	JustBeforeEach(func() {
		StartTorpedoTest("DriveAddAsJournalWithPXRestart",
			"Add drive when as journal",
			nil, 0)
	})

	stepLog := "Automate drive add journal with px restart"
	It(stepLog, func() {
		log.InfoD(stepLog)

		isjournal, err := IsJournalEnabled()
		log.FailOnError(err, "Failed to check if Journal enabled")

		if isjournal {
			log.Info("Journal drive already exists")
			Skip("Skipping the test DriveAddAsJournalWithPXRestart as journal drive already exists")
		}

		stepLog = "Schedule apps"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			contexts = make([]*scheduler.Context, 0)
			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				contexts = append(contexts, ScheduleApplications(fmt.Sprintf("adddriveasjournal-%d", i))...)
			}
			ValidateApplications(contexts)
		})
		defer appsValidateAndDestroy(contexts)

		stepLog = "Select a node to add journal drive"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			// Get Pool with running IO on the cluster
			poolUUID = pickPoolToResize(contexts, api.SdkStoragePool_RESIZE_TYPE_AUTO, 0)
			log.InfoD("Pool UUID on which IO is running [%s]", poolUUID)

			// Get Node Details of the Pool with IO
			nodeDetail, err = GetNodeWithGivenPoolID(poolUUID)
			log.FailOnError(err, "Failed to get Node Details from PoolUUID [%v]", poolUUID)
			log.InfoD("Pool with UUID [%v] present in Node [%v]", poolUUID, nodeDetail.Name)
		})

		// Add cloud drive on the selected node
		driveSpecs, err := GetCloudDriveDeviceSpecs()
		log.FailOnError(err, "Error getting cloud drive specs")

		deviceSpec := driveSpecs[0]
		devicespecjournal := deviceSpec + " --journal"

		stepLog = "Add journal drive"
		Step(stepLog, func() {
			log.InfoD(stepLog)

			//enter pool maintenance mode
			err = Inst().V.EnterPoolMaintenance(*nodeDetail)
			log.FailOnError(err, "Error Entering Maintenance mode on Node[%v]", nodeDetail.Name)
			log.InfoD("Enter pool Maintenance mode ")
			expectedStatus := "In Maintenance"

			log.FailOnError(WaitForPoolStatusToUpdate(*nodeDetail, expectedStatus),
				fmt.Sprintf("node %s pools are not in status %s", nodeDetail.Name, expectedStatus))

			systemOpts = node.SystemctlOpts{
				ConnectionOpts: node.ConnectionOpts{
					Timeout:         2 * time.Minute,
					TimeBeforeRetry: defaultRetryInterval,
				},
				Action: "start",
			}
			drivesMap, err := Inst().N.GetBlockDrives(*nodeDetail, systemOpts)
			log.FailOnError(err, "error getting block drives from node %s", nodeDetail.Name)
			blockDeviceBefore = len(drivesMap)
			err = Inst().V.AddCloudDrive(nodeDetail, devicespecjournal, -1)
			log.FailOnError(err, "journal add failed")
		})

		//random delay in secs
		sleepTime := rand.Intn(60-1) + 1
		time.Sleep(time.Second * (time.Duration(sleepTime)))

		stepLog = "Restart Portworx"
		//Restart portworx and wait for it to come up
		Step(stepLog, func() {
			log.Info(stepLog)
			Step(fmt.Sprintf("node with Px restart is: %s", nodeDetail.Name), func() {
				err := Inst().V.RestartDriver(*nodeDetail, nil)
				log.FailOnError(err, fmt.Sprintf("Error occured while Restart PX on node:%v", nodeDetail.Name))
			})

			Step(fmt.Sprintf("wait for volume driver to restart on node: %v", nodeDetail.Name), func() {
				err := Inst().V.WaitForPxPodsToBeUp(*nodeDetail)
				log.FailOnError(err, fmt.Sprintf("Error occured while Validating PX restart is done on node:%v", nodeDetail.Name))
			})
		})

		stepLog = "Exit pool maintenance mode"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			err = ExitPoolMaintenance(*nodeDetail)
			log.FailOnError(err, "Failed to exit maintenance mode")
			log.Info("exit pool maintenance mode succeed")
		})

		stepLog = "Verify Px Status"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			status, err := Inst().V.GetPxctlStatus(*nodeDetail)
			log.FailOnError(err, fmt.Sprintf("failed to get pxctl status on node [%s]", nodeDetail.Name))
			dash.VerifyFatal(status == api.Status_STATUS_OK.String(), true, fmt.Sprintf("node [%s] status is up but PX cluster is not ok. Expected: %v Actual: %v",
				nodeDetail.Name, api.Status_STATUS_OK, status))
			log.InfoD("px status %v", status)
		})

		stepLog = "Verify if journal is enabled"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			jDev, err := Inst().V.GetJournalDevicePath(nodeDetail)
			log.FailOnError(err, fmt.Sprintf("error getting journal device path from node %s", nodeDetail.Name))
			if jDev == "" {
				log.FailOnError(fmt.Errorf("no journal device path found"), "error getting journal device path from storage spec")
			}
			log.Infof("JournalDev path : %s", jDev)

			err = Inst().V.RefreshDriverEndpoints()
			log.FailOnError(err, "error refreshing driver end points")
		})

		stepLog = "Verify drive is added successfully"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			drivesMap, err := Inst().N.GetBlockDrives(*nodeDetail, systemOpts)
			log.FailOnError(err, "error getting block drives from node %s", nodeDetail.Name)
			blockDeviceAfter := len(drivesMap)
			dash.VerifyFatal(blockDeviceAfter > blockDeviceBefore, true, "adding cloud drive as journal successful")
		})
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts)
	})
})

var _ = Describe("{AddMetadataDriveWithNodeMaintenanceCycle}", Label("p1", "hal_ops_disruption", "NodeMaintenance", "AddMetadata", "functional"), func() {

	var (
		selectedNode node.Node
		contexts     []*scheduler.Context
		kvdbNodesIDs []string
		path         string
	)

	JustBeforeEach(func() {
		StartTorpedoTest("AddMetadataDriveWithNodeMaintenanceCycle", "Add Metadata Drive With node maintenance cycle", nil, 0)
	})

	itLog := "AddMetadataDriveWithNodeMaintenanceCycle"
	It(itLog, func() {
		stepLog := "Schedule Apps"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			contexts = scheduleApps()
			time.Sleep(5 * time.Minute)
			log.InfoD("schedule app succeed")
		})
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

		stepLog = "Check which node has a metadata disk if not add one"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			isDedicatedMetadataDiskExist := false
			//Check which node has metadata disk if not add one
			for _, storageNode := range storageNodes {

				path, err := getMetaDataDiskPath(storageNode)
				log.FailOnError(err, "Failed to get metadata disk")
				if path != "" {
					log.InfoD("Metadata disk path: %v", path)
					isDedicatedMetadataDiskExist = true
					tNode = storageNode
					break
				}
			}
			if !isDedicatedMetadataDiskExist {
				for _, storageNode := range storageNodes {
					deviceSpec := fmt.Sprintf("size=100 --metadata")
					log.InfoD("Initiate add cloud drive and validate")
					// enter pool maintenance mode
					if Contains(kvdbNodesIDs, storageNode.Id) {
						log.InfoD("[%s] is kvdb node", storageNode.Hostname)
						continue
					}

					stepLog := "Enter maintenance mode"
					Step(stepLog, func() {
						log.InfoD(stepLog)
						err = Inst().V.EnterPoolMaintenance(storageNode)
						log.FailOnError(err, "node: %v failed to transition to pool maintenance mode", storageNode.Name)
						log.Info("enter pool maintenance mode succeed")
					})

					stepLog = "Add metadata disk"
					Step(stepLog, func() {
						log.InfoD(stepLog)
						err := Inst().V.AddCloudDrive(&storageNode, deviceSpec, -1)
						log.FailOnError(err, "Failed to add metadata device on node : %s", storageNode.Name)
						log.InfoD("metadata disk added successfully on node [%s]", storageNode.Hostname)
					})

					sleepTime := rand.Intn(100) + 1
					time.Sleep(time.Second * (time.Duration(sleepTime)))
					stepLog = fmt.Sprintf("Performing node maintenance cycle on node [%s] after [%d] seconds", storageNode.Name, sleepTime)
					Step(stepLog, func() {
						log.InfoD(stepLog)
						err = Inst().V.RecoverDriver(storageNode)
						log.FailOnError(err, fmt.Sprintf("error performing maintenance cycle on node %s", storageNode.Name))
						err = Inst().V.WaitDriverUpOnNode(storageNode, 5*time.Minute)
						log.FailOnError(err, fmt.Sprintf("Driver is down on node %s", storageNode.Name))
						dash.VerifyFatal(err == nil, true, fmt.Sprintf("PX is up after maintenance cycle on node %s", storageNode.Name))
					})

					// exit pool maintenance
					stepLog = "Exit pool maintenance mode"
					Step(stepLog, func() {
						log.InfoD(stepLog)
						err = Inst().V.ExitPoolMaintenance(storageNode)
						log.FailOnError(err, "Node: %v Failed to exit out of maintenance mode", storageNode.Name)
						log.Info("exit pool maintenance mode succeed")
					})

					selectedNode = storageNode
					break
				}
			} else {
				log.InfoD("Metadata disk already exist: [%s]", path)
				Skip("Metadata disk already exist")
			}
			//check if selecteNode is empty or not
			if selectedNode.Name == "" {
				log.FailOnError(fmt.Errorf("No node found with metadata disk or metadata disks cannot be added to any nodes"), "No node found with metadata disk ")
			}
		})

		//Check PX status
		stepLog = "Check PX status"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			status, err := Inst().V.GetPxctlStatus(selectedNode)
			log.FailOnError(err, fmt.Sprintf("failed to get pxctl status on node [%s]", selectedNode.Name))
			dash.VerifyFatal(status == api.Status_STATUS_OK.String(), true, fmt.Sprintf("node [%s] status is up but PX cluster is not ok. Expected: %v Actual: %v",
				selectedNode.Name, api.Status_STATUS_OK, status))
			log.Infof("px status [%v]", status)

		})

		stepLog = fmt.Sprintf("Check if apps are running")
		Step(stepLog, func() {
			log.InfoD(stepLog)
			ValidateApplications(contexts)
			log.Info("validate application succeed")
		})

		stepLog = "validate metadata device has been added successfully"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			//get metadata device path from pxctl status
			metadataDevicePath, err := getMetaDataDiskPath(selectedNode)
			log.FailOnError(err, fmt.Sprintf("failed to get pxctl status on node [%s]", selectedNode.Name))
			if metadataDevicePath == "" {
				log.FailOnError(fmt.Errorf("metadata device not added"), "metadata device not added")
			}
			log.InfoD("Metadata device path from pxctl : %v", metadataDevicePath)
		})

	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		appsValidateAndDestroy(contexts)
		AfterEachTest(contexts)
	})
})

var _ = Describe("{AddMetadataDriveWithNodeReboot}", Label("p1", "hal_ops_disruption", "node_reboot", "AddMetadata", "functional"), func() {

	var (
		selectedNode node.Node
		contexts     []*scheduler.Context
		kvdbNodesIDs []string
		path         string
	)

	JustBeforeEach(func() {
		StartTorpedoTest("AddMetadataDriveWithNodeReboot", "Add Metadata Drive With Node Reboot", nil, 0)
	})

	itLog := "AddMetadataDriveWithNodeReboot"
	It(itLog, func() {
		stepLog := "Schedule Apps"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			contexts = scheduleApps()
			time.Sleep(5 * time.Minute)
			log.InfoD("schedule app succeed")
		})
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

		stepLog = "Check which node has a metadata disk if not add one"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			isDedicatedMetadataDiskExist := false
			//Check which node has metadata disk if not add one
			for _, storageNode := range storageNodes {

				path, err := getMetaDataDiskPath(storageNode)
				log.FailOnError(err, "Failed to get metadata disk")
				if path != "" {
					log.InfoD("Metadata disk path: %v", path)
					isDedicatedMetadataDiskExist = true
					tNode = storageNode
					break
				}
			}
			if !isDedicatedMetadataDiskExist {
				for _, storageNode := range storageNodes {
					deviceSpec := fmt.Sprintf("size=100 --metadata")
					log.InfoD("Initiate add cloud drive and validate")
					// enter pool maintenance mode
					if Contains(kvdbNodesIDs, storageNode.Id) {
						log.InfoD("[%s] is kvdb node", storageNode.Hostname)
						continue
					}

					stepLog := "Enter maintenance mode"
					Step(stepLog, func() {
						log.InfoD(stepLog)
						err = Inst().V.EnterPoolMaintenance(storageNode)
						log.FailOnError(err, "node: %v failed to transition to pool maintenance mode", storageNode.Name)
						log.Info("enter pool maintenance mode succeed")
					})

					stepLog = "Add metadata disk"
					Step(stepLog, func() {
						log.InfoD(stepLog)
						err := Inst().V.AddCloudDrive(&storageNode, deviceSpec, -1)
						log.FailOnError(err, "Failed to add metadata device on node : %s", storageNode.Name)
						log.InfoD("metadata disk added successfully on node [%s]", storageNode.Hostname)
					})
					// sleep for some random time
					sleepTime := rand.Intn(100) + 1
					time.Sleep(time.Second * (time.Duration(sleepTime)))

					// node restart
					stepLog = fmt.Sprintf("Verify reboot after [%d] seconds", sleepTime)
					Step(stepLog, func() {
						log.InfoD(stepLog)
						err = Inst().N.RebootNodeAndWait(storageNode)
						log.FailOnError(err, "Failed to reboot node and wait till it is up")
						log.InfoD("Verify reboot succeed")
					})

					// exit pool maintenance
					stepLog = "Exit pool maintenance mode"
					Step(stepLog, func() {
						log.InfoD(stepLog)
						err = Inst().V.ExitPoolMaintenance(storageNode)
						log.FailOnError(err, "Node: %v Failed to exit out of maintenance mode", storageNode.Name)
						log.Info("exit pool maintenance mode succeed")
					})

					selectedNode = storageNode
					break
				}
			} else {
				log.InfoD("Metadata disk already exist: [%s]", path)
				Skip("Metadata disk already exist")
			}
			//check if selecteNode is empty or not
			if selectedNode.Name == "" {
				log.FailOnError(fmt.Errorf("No node found with metadata disk or metadata disks cannot be added to any nodes"), "No node found with metadata disk ")
			}
		})

		//Check PX status
		stepLog = "Check PX status"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			status, err := Inst().V.GetPxctlStatus(selectedNode)
			log.FailOnError(err, fmt.Sprintf("failed to get pxctl status on node [%s]", selectedNode.Name))
			dash.VerifyFatal(status == api.Status_STATUS_OK.String(), true, fmt.Sprintf("node [%s] status is up but PX cluster is not ok. Expected: %v Actual: %v",
				selectedNode.Name, api.Status_STATUS_OK, status))
			log.Infof("px status [%v]", status)

		})

		stepLog = fmt.Sprintf("Check if apps are running")
		Step(stepLog, func() {
			log.InfoD(stepLog)
			ValidateApplications(contexts)
			log.Info("validate application succeed")
		})

		stepLog = "validate metadata device has been added successfully"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			//get metadata device path from pxctl status
			metadataDevicePath, err := getMetaDataDiskPath(selectedNode)
			log.FailOnError(err, fmt.Sprintf("failed to get pxctl status on node [%s]", selectedNode.Name))
			if metadataDevicePath == "" {
				log.FailOnError(fmt.Errorf("metadata device not added"), "metadata device not added")
			}
			log.InfoD("Metadata device path from pxctl : %v", metadataDevicePath)
		})

	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		appsValidateAndDestroy(contexts)
		AfterEachTest(contexts)
	})

})

var _ = Describe("{DriveScalingWithPxRestart}", Label("p1", "hal_ops_disruption", "px_restart", "AddMetadata", "functional"), func() {
	/*
		1. Select a storage node in the cluster
		2. Add meta drive if not already present
		3. Add new pools till the pool limit is hit
		4. For each pool, expand with add drive till drive limit is hit
		5. Schedule apps and wait for 5 mins
		6. Performing PX Restart and check the time it took for px to come up
		7. Validate apps are still running and then destroy them
		8. Delete all the pools on the selected node
		9. Recreate all the pools on the node which are there in the start of the test case
	*/
	performDriveScalingTest("DriveScalingWithPxRestart")
})

var _ = Describe("{DriveScalingWithNodeMaintenanceCycle}", Label("p1", "hal_ops_disruption", "NodeMaintenance", "AddMetadata", "functional"), func() {
	/*
	   1. Select a storage node in the cluster
	   2. Add meta drive if not already present
	   3. Add new pools till the pool limit is hit
	   4. For each pool, expand with add drive till drive limit is hit
	   5. Schedule apps and wait for 5 mins
	   6. Performing Maintenance cycle and check the time it took for px to come up
	   7. Validate apps are still running and then destroy them
	   8. Delete all the pools on the selected node
	   9. Recreate all the pools on the node which are there in the start of the test case
	*/
	performDriveScalingTest("DriveScalingWithNodeMaintenanceCycle")

})

var _ = Describe("{DriveScalingWithNodeReboot}", Label("p1", "hal_ops_disruption", "node_reboot", "AddMetadata", "functional"), func() {
	/*
		1. Select a storage node in the cluster
		2. Add meta drive if not already present
		3. Add new pools till the pool limit is hit
		4. For each pool, expand with add drive till drive limit is hit
		5. Schedule apps and wait for 5 mins
		6. Reboot node and check the time it took for px to come up
		7. Validate apps are still running and then destroy them
		8. Delete all the pools on the selected node
		9. Recreate all the pools on the node which are there in the start of the test case
	*/

	performDriveScalingTest("DriveScalingWithNodeReboot")
})

func performDriveScalingTest(testName string) {
	var (
		selectedNode            node.Node
		contexts                []*scheduler.Context
		kvdbNodesIDs            []string
		metaDataDiskPath        string
		expansionEligibilityMap map[string]bool
		jrnlPartPoolID          string
		isjournal               bool
		bufferSizeInGB          uint64
		testDes                 string
	)
	if testName == "DriveScalingWithNodeReboot" {
		testDes = "Drive scaling with node reboot"
	}

	if testName == "DriveScalingWithNodeMaintenanceCycle" {
		testDes = "Drive scaling with node maintenance cycle"
	}

	if testName == "DriveScalingWithPxRestart" {
		testDes = "Drive scaling with px restart"
	}

	JustBeforeEach(func() {
		StartTorpedoTest(testName, testDes, nil, 0)
	})

	itLog := testName
	It(itLog, func() {

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

		stepLog = "Check which node has a metadata disk if not add one"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			isDedicatedMetadataDiskExist := false
			//Check which node has metadata disk if not add one
			for _, storageNode := range storageNodes {

				path, err := getMetaDataDiskPath(storageNode)
				log.FailOnError(err, "Failed to get metadata disk")
				if Contains(kvdbNodesIDs, storageNode.Id) {
					log.InfoD("[%s] is kvdb node", storageNode.Hostname)
					continue
				}
				if path != "" {
					metaDataDiskPath = path
					log.InfoD("Metadata disk path: %v", path)
					isDedicatedMetadataDiskExist = true
					selectedNode = storageNode
					break
				}
			}

			if !isDedicatedMetadataDiskExist {
				for _, storageNode := range storageNodes {
					deviceSpec := fmt.Sprintf("size=100 --metadata")
					log.InfoD("Initiate add cloud drive and validate")
					// enter pool maintenance mode
					if Contains(kvdbNodesIDs, storageNode.Id) {
						log.InfoD("[%s] is kvdb node", storageNode.Hostname)
						continue
					}

					stepLog := "Enter maintenance mode"
					Step(stepLog, func() {
						log.InfoD(stepLog)
						err = Inst().V.EnterPoolMaintenance(storageNode)
						log.FailOnError(err, "node: %v failed to transition to pool maintenance mode", storageNode.Name)
						log.Info("enter pool maintenance mode succeed")
					})

					stepLog = "Add metadata disk"
					Step(stepLog, func() {
						log.InfoD(stepLog)
						err := Inst().V.AddCloudDrive(&storageNode, deviceSpec, -1)
						log.FailOnError(err, "Failed to add metadata device on node : %s", storageNode.Name)
						log.InfoD("metadata disk added successfully on node [%s]", storageNode.Hostname)
					})

					// exit pool maintenance
					stepLog = "Exit pool maintenance mode"
					Step(stepLog, func() {
						log.InfoD(stepLog)
						err = Inst().V.ExitPoolMaintenance(storageNode)
						log.FailOnError(err, "Node: %v Failed to exit out of maintenance mode", storageNode.Name)
						log.Info("exit pool maintenance mode succeed")
					})

					selectedNode = storageNode
					log.InfoD("selected node [%s] ", selectedNode.Name)

					break
				}
			} else {
				log.InfoD("Metadata disk already exist: [%s]", metaDataDiskPath)
			}
			//check if selecteNode is empty or not
			if selectedNode.Name == "" {
				log.FailOnError(fmt.Errorf("No node found with metadata disk or metadata disks cannot be added to any nodes"), "No node found with metadata disk ")
			}
		})

		poolsBfr, err := GetPoolsDetailsOnNode(&selectedNode)
		log.FailOnError(err, fmt.Sprintf("error getting pools on node %s", selectedNode.Name))
		driveSpecs, err := GetCloudDriveDeviceSpecs()
		log.FailOnError(err, "Error getting cloud drive specs")

		stepLog = "Add new pools till the pool limit is hit"
		Step(stepLog, func() {
			log.Info(stepLog)
			i := 1
			for {
				err = Inst().V.AddCloudDrive(&selectedNode, driveSpecs[0], -1)
				if err != nil && strings.Contains(err.Error(), "Maximum pools limit reached") {
					break
				}
				log.FailOnError(err, fmt.Sprintf("Add cloud drive failed on node %s", selectedNode.Name))
				log.InfoD("%d pool is added ", i)
				i++
				time.Sleep(time.Minute)
			}
			log.InfoD("New Pool added successfully")
		})
		poolsAft, err := GetPoolsDetailsOnNode(&selectedNode)
		log.FailOnError(err, fmt.Sprintf("error getting pools on node after adding metadata %s", selectedNode.Name))

		drvM, err := Inst().V.GetPoolDrives(&selectedNode)
		log.FailOnError(err, fmt.Sprintf("error getting pools drives on node %s", selectedNode.Name))

		isDMthin, err := IsDMthin()
		log.FailOnError(err, "Failed to check if the cluster is DMTHIN")
		isjournal, err = IsJournalEnabled()
		log.FailOnError(err, "Failed to check if Journal enabled")
		if isJournalEnabled {
			bufferSizeInGB = JournalDeviceSizeInGB
		}

		if !isDMthin {
			stepLog = fmt.Sprintf("Expand each pool on node [%s] with add drive till drive limit is hit ", selectedNode.Hostname)
			Step(stepLog, func() {
				log.InfoD(stepLog)
				expansionEligibilityMap, err = GetPoolExpansionEligibility(&selectedNode, api.SdkStoragePool_RESIZE_TYPE_ADD_DISK, 0)
				log.FailOnError(err, "error checking node [%s] expansion criteria", selectedNode.Name)
				poolExpansionCompleted := 0
				count := 1
			outer:
				for expansionEligibilityMap[selectedNode.Id] && poolExpansionCompleted < len(poolsAft) {
					poolExpansionCompleted = 0
					for _, pool := range poolsAft {
						if !expansionEligibilityMap[pool.Uuid] {
							log.Infof(fmt.Sprintf("Pool expansion completed on [%d]", pool.GetID()))
							poolExpansionCompleted++
							continue
						}
						d := drvM[fmt.Sprintf("%d", pool.ID)]
						log.Infof("Current size of pool %s is %d GiB. Expand to %v GiB with type add-disk...",
							pool.Uuid, pool.TotalSize/units.GiB, d[0].SizeInGib)
						targetSize := (pool.TotalSize / units.GiB) + (d[0].SizeInGib * uint64(count))
						triggerPoolExpansion(pool.GetUuid(), targetSize+bufferSizeInGB, api.SdkStoragePool_RESIZE_TYPE_ADD_DISK)
						resizeErr := waitForOngoingPoolExpansionToComplete(pool.GetUuid())
						if resizeErr != nil && strings.Contains(resizeErr.Error(), "node has reached it's maximum supported drive count") {
							break outer
						}
						dash.VerifyFatal(resizeErr, nil, fmt.Sprintf("Pool expansion for pool [%s] does not result in error", pool.GetUuid()))
						log.Infof(fmt.Sprintf("Pool expansion succeed [%s]", pool.Uuid))

					}
					count++
					d := drvM[fmt.Sprintf("%d", poolsAft[0].ID)]
					expansionEligibilityMap, err = GetPoolExpansionEligibility(&selectedNode, api.SdkStoragePool_RESIZE_TYPE_ADD_DISK, d[0].SizeInGib)
					log.FailOnError(err, "error checking node [%s] expansion criteria", selectedNode.Name)

				}
				log.Infof(fmt.Sprintf("Pools expansion succeed"))
			})
		}

		stepLog := "Schedule Apps"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			contexts = scheduleApps()
			time.Sleep(5 * time.Minute)
			log.InfoD("Scheduling the apps was successful")
		})

		if testName == "DriveScalingWithNodeReboot" {

			stepLog = "Reboot the node"
			//Reboot node and wait for PX to come up
			Step(stepLog, func() {
				t := time.Now()
				log.Info(stepLog)
				err = RebootNodeAndWaitForPxUp(selectedNode)
				log.FailOnError(err, "Failed to reboot node and wait till it is up")
				stepLog = "Check PX status"
				Step(stepLog, func() {
					log.InfoD(stepLog)
					status, err := Inst().V.GetPxctlStatus(selectedNode)
					log.FailOnError(err, fmt.Sprintf("failed to get pxctl status on node [%s]", selectedNode.Name))
					dash.VerifyFatal(status == api.Status_STATUS_OK.String(), true, fmt.Sprintf("node [%s] status is up but PX cluster is not ok. Expected: %v Actual: %v",
						selectedNode.Name, api.Status_STATUS_OK, status))
					log.Infof("px status [%v] [%f] seconds", status, time.Now().Sub(t).Seconds())

				})

			})
		} else if testName == "DriveScalingWithNodeMaintenanceCycle" {
			stepLog = fmt.Sprintf("Performing node maintenance cycle and checking px status on node [%s]", selectedNode.Name)
			Step(stepLog, func() {
				t := time.Now()
				log.InfoD(stepLog)
				err = Inst().V.RecoverDriver(selectedNode)
				log.FailOnError(err, fmt.Sprintf("error performing maintenance cycle on node %s", selectedNode.Name))
				err = Inst().V.WaitDriverUpOnNode(selectedNode, 10*time.Minute)
				log.FailOnError(err, fmt.Sprintf("Driver is down on node %s", selectedNode.Name))
				Step(stepLog, func() {
					log.InfoD(stepLog)
					status, err := Inst().V.GetPxctlStatus(selectedNode)
					log.FailOnError(err, fmt.Sprintf("failed to get pxctl status on node [%s]", selectedNode.Name))
					dash.VerifyFatal(status == api.Status_STATUS_OK.String(), true, fmt.Sprintf("node [%s] status is up but PX cluster is not ok. Expected: %v Actual: %v",
						selectedNode.Name, api.Status_STATUS_OK, status))
					log.Infof("px status [%v] [%f] seconds", status, time.Now().Sub(t).Seconds())

				})
			})
		} else if testName == "DriveScalingWithPxRestart" {

			stepLog = fmt.Sprintf("Performing Restart Portworx and checking px status on node [%s]", selectedNode.Name)
			//Restart portworx and wait for it to come up
			Step(stepLog, func() {
				log.Info(stepLog)
				t := time.Now()

				Step(fmt.Sprintf("node with Px restart is: %s", selectedNode.Name), func() {
					err := Inst().V.RestartDriver(selectedNode, nil)
					log.FailOnError(err, fmt.Sprintf("Error occured while Restart PX on node:%v", selectedNode.Name))
				})

				Step(fmt.Sprintf("wait for volume driver to restart on node: %v", selectedNode.Name), func() {
					err := Inst().V.WaitForPxPodsToBeUp(selectedNode)
					log.FailOnError(err, fmt.Sprintf("Error occured while Validating PX restart is done on node:%v", selectedNode.Name))
				})

				err = Inst().V.WaitDriverUpOnNode(selectedNode, 10*time.Minute)
				log.FailOnError(err, fmt.Sprintf("Driver is down on node %s", selectedNode.Name))

				stepLog = "Check PX status"
				Step(stepLog, func() {
					log.InfoD(stepLog)
					status, err := Inst().V.GetPxctlStatus(selectedNode)
					log.FailOnError(err, fmt.Sprintf("failed to get pxctl status on node [%s]", selectedNode.Name))
					dash.VerifyFatal(status == api.Status_STATUS_OK.String(), true, fmt.Sprintf("node [%s] status is up but PX cluster is not ok. Expected: %v Actual: %v",
						selectedNode.Name, api.Status_STATUS_OK, status))
					log.Infof("px status [%v] [%f] seconds", status, time.Now().Sub(t).Seconds())

				})
			})
		}
		stepLog = fmt.Sprintf("Validate apps and destroy")
		Step(stepLog, func() {
			log.InfoD(stepLog)
			appsValidateAndDestroy(contexts)
			log.Infof(fmt.Sprintf("Validated and destroyed apps successfully"))

		})

		stepLog = "Delete the new pool"
		Step(stepLog, func() {
			log.Info(stepLog)
			journalPoolId := ""

			stepLog = "Selecting journal pool"
			Step(stepLog, func() {
				log.InfoD(stepLog)

				nodePools := selectedNode.StoragePools
				if isjournal && len(nodePools) > 1 {
					jDev, err := Inst().V.GetJournalDevicePath(&selectedNode)
					log.FailOnError(err, fmt.Sprintf("error getting journal device path from node %s", selectedNode.Name))
					log.Infof("JournalDev: %s", jDev)
					if jDev == "" {
						log.FailOnError(fmt.Errorf("no journal device path found"), "error getting journal device path from storage spec")
					}
					drivesMap, err := Inst().V.GetPoolDrives(&selectedNode)
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

				}

			})

			for _, pool := range poolsAft {
				if strconv.Itoa(int(pool.GetID())) == jrnlPartPoolID {
					continue
				}
				err = DeletePoolAndValidate(selectedNode, strconv.Itoa(int(pool.GetID())))
				log.FailOnError(err, fmt.Sprintf("Error occured while Validating the deleted pool %s in the node %s", pool.Uuid, selectedNode.Name))
				log.InfoD("pool [%d] delete succed", pool.GetID())
			}
			if jrnlPartPoolID != "" {
				log.Info("Deleting the journal pool [%s]", jrnlPartPoolID)
				err = DeletePoolAndValidate(selectedNode, jrnlPartPoolID)
				log.FailOnError(err, fmt.Sprintf("Error occured while Validating the deleted journal pool %s in the node %s", journalPoolId, selectedNode.Name))
				log.InfoD("Journal pool delete succed")
			}
			log.InfoD("All pool deleted successfully")
		})

		stepLog = "Add new pools"
		Step(stepLog, func() {
			i := 1
			log.Info(stepLog)
			driveSpecs, err := GetCloudDriveDeviceSpecs()
			log.FailOnError(err, "Error getting cloud drive specs")
			for _, pool := range poolsBfr {

				deviceSpecParams := strings.Split(driveSpecs[0], ",")
				paramsArr := make([]string, 0)
				for _, param := range deviceSpecParams {
					if strings.Contains(param, "size") {
						paramsArr = append(paramsArr, fmt.Sprintf("size=%d,", pool.TotalSize/units.GiB))
					} else {
						paramsArr = append(paramsArr, param)
					}
				}
				//drive spec generated from actual cloudrive spec
				newSpec := strings.Join(paramsArr, ",")

				err = Inst().V.AddCloudDrive(&selectedNode, newSpec, -1)
				log.FailOnError(err, fmt.Sprintf("Add cloud drive failed on node %s", selectedNode.Name))
				time.Sleep(time.Minute)
				log.InfoD("%d pool is added ", i)
				i++
			}
			log.InfoD("Adding new pool was successful")
		})

	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts)
	})

}

var _ = Describe("{AddDataDriveWithMetadrive}", Label("staging", "p0", "postive", "AddDrive"), func() {
	/*
		ticket id : https://purestorage.atlassian.net/browse/HAZEL-1549
		Prerequsites:
		    At least one storageless nodes are required to run this test.
		step1 : Add drive on storeless node with adding metadrive
		step2: verify add data drive should happen
		step3: verify storage less node should convert to storage node
	*/
	JustBeforeEach(func() {
		StartTorpedoTest("AddDataDriveWithoutMetadrive", "Adding a drive to a storageless node with metadrive and verify node conversion.", nil, 0)
	})

	itLog := "Test adding a drive to a storageless node with metadrive, verify successful addition, and ensure node conversion to storage node"
	It(itLog, func() {
		log.InfoD(itLog)
		var (
			nodeSelected node.Node
		)
		storagelessNode := node.GetStorageLessNodes()
		log.InfoD("Checking number of storageless nodes: %d", len(storagelessNode))
		if len(storagelessNode) < 1 {
			Skip("At least one storageless nodes are required to run this test!...")
		}
		stepLog := "Check cluster type and attempt to add drive with metadata"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			log.InfoD("Check if the cluster is DMTHIN")
			isDmthin, _ := IsDMthin()
			if !isDmthin {
				Skip("Cluster is not DMTHIN, skipping metadata disk addition.")
			}
			nodeSelected = storagelessNode[0]
			log.InfoD("Selected a node to add a regular data drive to node [%s]", nodeSelected.Name)
			driveSpecs, err := GetCloudDriveDeviceSpecs()
			log.FailOnError(err, "Error getting cloud drive specs")
			deviceSpec := driveSpecs[0]
			err = AddMetadataDisk(nodeSelected)
			log.FailOnError(err, "error while adding metadata disk")
			log.InfoD("Add regular data drive")
			deviceSpecParams := strings.Split(deviceSpec, ",")
			paramsArr := make([]string, 0)
			for _, param := range deviceSpecParams {
				if strings.Contains(param, "size") {
					paramsArr = append(paramsArr, fmt.Sprintf("size=%d,", 70))
				} else {
					paramsArr = append(paramsArr, param)
				}
			}
			newSpec := strings.Join(paramsArr, ",")
			log.InfoD("Attempting to add a regular data drive of size  to node [%s]", nodeSelected.Name)
			err = Inst().V.AddCloudDrive(&nodeSelected, newSpec, -1)
			dash.VerifyFatal(err, nil, "Drive was not added successfully")
			log.InfoD("Successfully added a new data drive of size to node [%s]", nodeSelected.Name)

		})
		stepLog = "Verify storageless node is converted to a storage node"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			log.Info("Refresh the driver endpoints")
			err = Inst().S.RefreshNodeRegistry()
			log.FailOnError(err, "error refreshing node registry")
			err = Inst().V.RefreshDriverEndpoints()
			log.FailOnError(err, "error refreshing storage drive endpoints")
			storageNodesAfterstoragelessnodejoin := node.GetStorageNodes()
			log.InfoD("Current storage nodes: %+v", storageNodesAfterstoragelessnodejoin)
			found := false
			for _, nodeID := range storageNodesAfterstoragelessnodejoin {
				if nodeID.Id == nodeSelected.Id {
					found = true
					break
				}
			}
			dash.VerifyFatal(found, true, fmt.Sprintf("Storageless node with ID %s should be converted to a storage node.", nodeSelected.Id))
			log.InfoD("Storageless node with ID %s is successfully converted to a storage node.", nodeSelected.Id)
		})

	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
	})
})

var _ = Describe("{AddDataDriveWithoutMetadrive}", Label("staging", "p0", "postive", "adddrive"), func() {
	/*
		ticket id : https://purestorage.atlassian.net/browse/HAZEL-1549
		Prerequsites:
		    At least one storageless nodes are required to run this test.
		step1 : Add drive on storeless node without adding metadrive
		step2: verify add data drive should not happen and verify the msg
	*/
	JustBeforeEach(func() {
		StartTorpedoTest("AddDataDriveWithoutMetadrive", "Add a data drive on storage less node without adding metadrive", nil, 0)
	})

	itLog := "Add a drive on a storageless node without a metadrive and verify that data drive addition is blocked with an error message"
	It(itLog, func() {
		log.InfoD(itLog)
		const (
			expect_out = "Failed to get /adddrive: system metadata device not found, it is mandatory for this configuration"
		)
		storagelessNode := node.GetStorageLessNodes()
		log.InfoD("Checking number of storageless nodes: %d", len(storagelessNode))
		if len(storagelessNode) < 1 {
			Skip("At least one storageless nodes are required to run this test!...")
		}

		stepLog := "Check if the node has a metadata disk (DMTHIN cluster) or add a data drive (BTRFS cluster)"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			log.InfoD("Check if the cluster is DMTHIN")
			isDmthin, _ := IsDMthin()
			if !isDmthin {
				Skip("Cluster is not DMTHIN, skipping metadata disk addition.")
			}
			nodeSelected := storagelessNode[0]
			driveSpecs, err := GetCloudDriveDeviceSpecs()
			log.FailOnError(err, "Error getting cloud drive specs")
			deviceSpec := driveSpecs[0]
			deviceSpecParams := strings.Split(deviceSpec, ",")
			paramsArr := make([]string, 0)
			for _, param := range deviceSpecParams {
				if strings.Contains(param, "size") {
					paramsArr = append(paramsArr, fmt.Sprintf("size=%d,", 70))
				} else {
					paramsArr = append(paramsArr, param)
				}
			}
			newSpec := strings.Join(paramsArr, ",")
			log.InfoD("Attempting to add a regular data drive of size  to node [%s]", nodeSelected.Name)
			err = Inst().V.AddCloudDrive(&nodeSelected, newSpec, -1)
			dash.VerifyFatal(strings.Contains(err.Error(), expect_out), true, "Error message not as expected")

		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
	})

})

// Verify rebooting px doesn't enter into maintenance mode when a pool is offline.
var _ = Describe("{VerifyPxRebootAvoidsMaintenanceWithOfflinePool}", Label("p1", "px_restart", "pool_ops", "Throttling", "NodeMaintenance", "staging"), func() {
	/*
		Jira-ID :https://purestorage.atlassian.net/browse/HAZEL-1000
		Bring pool into offline state (
			step1: feed p1 size GB I/O on the volume
			step2: After I/O done p1 should be offline and full)
		Restart px
		Validate state of pool & node → It should not go into maintenance mode
	*/

	JustBeforeEach(func() {
		StartTorpedoTest("VerifyPxRebootAvoidsMaintenanceWithOfflinePool", "Verify rebooting px does not enter into maintenance mode when a pool is offline", nil, 0)
		// Remove if node-type label is set before the test
		err = RemoveLabelsAllNodes(k8s.NodeType, true, false)
		log.FailOnError(err, "error removing label on node ")
	})

	var (
		selectedNode   node.Node
		contexts       []*scheduler.Context
		appList        []string
		secondReplNode node.Node
		stNodes        []node.Node
		isjournal      bool
	)

	itLog := "Verify rebooting px does not enter into maintenance mode when a pool is offline"
	It(itLog, func() {
		stepLog := "Create vols and make pool full"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			selectedNode = *GetNodeWithLeastSize()
			if selectedNode.Name == "" {
				log.FailOnError(fmt.Errorf("unable get node with least size"), "error identifying the node with least size")
			}
			log.Infof(fmt.Sprintf("Node %s is marked for repl 1", selectedNode.Name))
			stNodes = node.GetStorageNodes()

			for _, stNode := range stNodes {
				if stNode.Name != selectedNode.Name {
					secondReplNode = stNode
				}
			}

			isjournal, err = IsJournalEnabled()
			log.FailOnError(err, "Failed to check if Journal enabled")

			err = adjustReplPools(selectedNode, secondReplNode, isjournal)
			log.FailOnError(err, "Error setting pools for clean volumes")

			appList = Inst().AppList

			err = Inst().S.AddLabelOnNode(selectedNode, k8s.NodeType, k8s.FastpathNodeType)
			log.FailOnError(err, fmt.Sprintf("Failed add label on node %s", selectedNode.Name))
			err = Inst().S.AddLabelOnNode(secondReplNode, k8s.NodeType, k8s.FastpathNodeType)
			log.FailOnError(err, fmt.Sprintf("Failed add label on node %s", secondReplNode.Name))

			Inst().AppList = []string{"fio-fastpath"}
			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				contexts = append(contexts, ScheduleApplications(fmt.Sprintf("nwplfull-%d", i))...)
			}
			ValidateApplications(contexts)
		})
		cleanup := func() {
			log.Info("Executing cleanup tasks")
			DestroyApps(contexts, nil)
			Inst().AppList = appList
			err := Inst().S.RemoveLabelOnNode(selectedNode, k8s.NodeType)
			log.FailOnError(err, "error removing label on node [%s]", selectedNode.Name)
			err = Inst().S.RemoveLabelOnNode(secondReplNode, k8s.NodeType)
			log.FailOnError(err, "error removing label on node [%s]", secondReplNode.Name)
		}
		defer cleanup()

		stepLog = "Checking Pool status before Px Restart"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			err := WaitForPoolOffline(selectedNode)
			log.FailOnError(err, fmt.Sprintf("Failed to make node %s storage down", selectedNode.Name))
			log.Infof("Waited for the pool to go offline...")
			poolsStatus, err := Inst().V.GetNodePoolsStatus(selectedNode)
			log.FailOnError(err, "error getting pool status on node %s", selectedNode.Name)
			log.Infof("Pool status on node %s is %v", selectedNode.Name, poolsStatus)
			for poolID, status := range poolsStatus {
				if status == "Offline" {
					log.Infof("The status of the poolID : [%s] is [%v]", poolID, status)
					break
				}
			}
		})

		stepLog = "Restart Portworx"
		//Restart portworx and wait for it to come up
		Step(stepLog, func() {
			log.Info(stepLog)
			log.FailOnError(Inst().V.RestartDriver(selectedNode, nil), fmt.Sprintf("Error restarting px on node [%s]", selectedNode.Name))
			err = Inst().V.WaitDriverUpOnNode(selectedNode, Inst().DriverStartTimeout)
			if err != nil {
				log.InfoD("Expanding storage full pools to bring storage down to online status")
				expandPoolSize := func(node node.Node) {
					poolStatusMap, err := Inst().V.GetNodePoolsStatus(node)

					for poolID, _ := range poolStatusMap {
						storagePool := getStoragePool(poolID)
						expectedSize := (storagePool.TotalSize / units.GiB) * 3

						log.InfoD("Current Size of the pool %s is %d", storagePool.Uuid, storagePool.TotalSize/units.GiB)
						err = Inst().V.ExpandPool(storagePool.Uuid, api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK, expectedSize, true)
						dash.VerifyFatal(err, nil, "Pool expansion init successful?")
						resizeErr := waitForPoolToBeResized(expectedSize, storagePool.Uuid, isjournal)
						dash.VerifyFatal(resizeErr, nil, fmt.Sprintf("Verify pool %s on node %s expansion using resize-disk", storagePool.Uuid, node.Name))
						status, err := Inst().V.GetNodeStatus(node)
						log.FailOnError(err, fmt.Sprintf("Error getting PX status of node %s", node.Name))
						dash.VerifySafely(*status, api.Status_STATUS_OK, fmt.Sprintf("validate PX status on node %s. Current status: [%s]", node.Name, status.String()))
					}
				}
				stepLog = "Expanding storage full pools and verify"
				Step(stepLog, func() {
					expandPoolSize(selectedNode)
				})
			}
			err = Inst().V.WaitDriverUpOnNode(selectedNode, Inst().DriverStartTimeout)
			log.FailOnError(err, fmt.Sprintf("Driver is down on node %s", selectedNode.Name))
		})

		stepLog = "Verify node status After Px Restart"
		Step(stepLog, func() {
			log.Info(stepLog)
			t := func() (interface{}, bool, error) {
				nodeStatus, err := Inst().V.GetNodeStatus(selectedNode)
				if err != nil {
					return nil, true, fmt.Errorf("error getting node status for the node %s", selectedNode.Name)
				}

				if *nodeStatus == opsapi.Status_STATUS_OK {
					log.Infof("The status of the node is online")
					return *nodeStatus, false, nil
				}

				return nil, true, fmt.Errorf("the status of the node is still not online")
			}
			nodeStatus, err := task.DoRetryWithTimeout(t, 30*time.Minute, 20*time.Second)
			log.FailOnError(err, "error getting node status on node %s", selectedNode.Name)
			dash.VerifyFatal(nodeStatus.(api.Status), opsapi.Status_STATUS_OK, fmt.Sprintf("validate PX status on node %s", selectedNode.Id))
		})

		stepLog = "Validate Applications"
		Step(stepLog, func() {
			ValidateApplications(contexts)
		})

		stepLog = "Expanding storage full pools and verify"
		Step(stepLog, func() {
			for _, node := range stNodes {
				poolStatusMap, err := Inst().V.GetNodePoolsStatus(node)
				for poolID, poolStatus := range poolStatusMap {
					if poolStatus == "Offline" {
						storagePool := getStoragePool(poolID)
						expectedSize := (storagePool.TotalSize / units.GiB) * 3

						log.InfoD("Current Size of the pool %s is %d", storagePool.Uuid, storagePool.TotalSize/units.GiB)
						err = Inst().V.ExpandPool(storagePool.Uuid, api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK, expectedSize, true)
						dash.VerifyFatal(err, nil, "Pool expansion init successful?")
						resizeErr := waitForPoolToBeResized(expectedSize, storagePool.Uuid, isjournal)
						dash.VerifyFatal(resizeErr, nil, fmt.Sprintf("Verify pool %s on node %s expansion using resize-disk", storagePool.Uuid, node.Name))
						status, err := Inst().V.GetNodeStatus(node)
						log.FailOnError(err, fmt.Sprintf("Error getting PX status of node %s", node.Name))
						dash.VerifySafely(*status, api.Status_STATUS_OK, fmt.Sprintf("validate PX status on node %s. Current status: [%s]", node.Name, status.String()))
					}
				}
			}
		})

		stepLog = "Validate Applications"
		Step(stepLog, func() {
			ValidateApplications(contexts)
		})
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})
})

var _ = Describe("{DMthinIncrementalPoolExpand}", Label("p1", "positive", "pool_ops", "px_ops", "PoolExpand", "staging"), func() {
	/*
		Try Incremental Pool expansion on the Pool till Max Supported size is reached , every time increase by 1T , and Max Limit is 15T .
		when trying pool expansion , we should have REPL1 , Repl2 , FastPath Volumes , with inflight IOs , and On going snapshots in progress.
	*/

	JustBeforeEach(func() {
		StartTorpedoTest("DMthinIncrementalPoolExpand", "Validate incremental pool expansion in DMTHIN", nil, 0)
	})

	var (
		contexts                []*scheduler.Context
		poolToResize            *api.StoragePool
		poolIDToResize          string
		targetSizeInBytes       uint64
		originalSizeInBytes     uint64
		targetSizeGiB           uint64
		resizeErr               error
		randomResizeMethodIndex int
		appList                 []string
		coordinatorNodes        []node.Node
	)

	stepLog := "DMthin incremental pool expansion"
	It(stepLog, func() {
		log.InfoD(stepLog)

		contexts = make([]*scheduler.Context, 0)
		retain := 3
		interval := 2

		// creating cloud credentials
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

		isDMthin, err := IsDMthin()
		log.FailOnError(err, "Failed to check if the cluster is DMTHIN")
		if !isDMthin {
			Skip("Cluster is not DMTHIN so skipping the test")
		}

		contexts = make([]*scheduler.Context, 0)
		policyNameList := []string{"localintervalpolicy", "intervalpolicy"}
		stepLog = "Create schedule policy"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for _, policyName := range policyNameList {
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
			}
		})

		appList = Inst().AppList
		Inst().AppList = []string{"fio-cloudsnap", "fio-fastpath-repl1"}
		stepLog = "schedule the applications and validate"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			contexts = make([]*scheduler.Context, 0)
			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				contexts = append(contexts, ScheduleApplications(fmt.Sprintf("dmthinincrementalexpand-%d", i))...)
			}
			ValidateApplications(contexts)
		})

		defer func() {
			Inst().AppList = appList
		}()

		stepLog = "Verify that snapshots are happening"
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
					dash.VerifyFatal(len(appVolumes) >= 1, true, "There should be atleast one volume to proceed with taking snapshot")
				})
				log.Infof("Got volume count : %v", len(appVolumes))

				for _, v := range appVolumes {
					snapshotScheduleName := v.Name + "-interval-schedule"
					log.InfoD("snapshotScheduleName : %v for volume: %s", snapshotScheduleName, v.Name)

					var latestSnapshot *storkv1.ScheduledVolumeSnapshotStatus
					// Polling for the new snapshot
					_, err = task.DoRetryWithTimeout(func() (interface{}, bool, error) {
						resp, err := storkops.Instance().GetSnapshotSchedule(snapshotScheduleName, appNamespace)
						if err != nil {
							return nil, false, fmt.Errorf("error getting snapshot schedule for %s, volume:%s in namespace %s", snapshotScheduleName, v.Name, v.Namespace)
						}
						if len(resp.Status.Items) == 0 {
							return nil, true, fmt.Errorf("waiting for new snapshot schedules for %s, volume:%s in namespace %s", snapshotScheduleName, v.Name, v.Namespace)
						}

						// Find the latest snapshot
						for _, item := range resp.Status.Items {
							for _, status := range item {
								if latestSnapshot == nil || status.CreationTimestamp.After(latestSnapshot.CreationTimestamp.Time) {
									latestSnapshot = status
								}
							}
						}

						if latestSnapshot != nil {
							return latestSnapshot, false, nil
						}
						return nil, true, fmt.Errorf("latest snapshot not found")
					}, time.Duration(15*interval)*defaultCommandTimeout, defaultReadynessTimeout)
					log.FailOnError(err, "Failed to get a latest snapshot")

					log.Infof("Latest snapshot %s for volume: %s", latestSnapshot.Name, v.Name)
				}
			}
		})

		stepLog = "Trigger the RESIZE_TYPE_RESIZE_DISK or RESIZE_TYPE_AUTO operation for incremental pool expansion"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			maxLimit := uint64(15)
			maxLimitSizeInBytes := maxLimit * units.TiB
			maxLimitSizeInGiB := maxLimitSizeInBytes / units.GiB

			// pick the coordinator nodes for the apps
			for _, ctx := range contexts {
				vols, err := Inst().S.GetVolumes(ctx)
				log.FailOnError(err, "Failed to get volumes for app %s", ctx.App.Key)
				vol := vols[0]
				attachedNode, err := Inst().V.GetNodeForVolume(vol, 1*time.Minute, 5*time.Second)
				log.FailOnError(err, "Failed to Get Attached node for volume [%v]", vol.ID)
				log.Infof("The coordinator node for app [%v] is [%s]", ctx.App.Key, attachedNode.Name)
				coordinatorNodes = append(coordinatorNodes, *attachedNode)
			}

			// randomly select a coordinator node to get one pool on it
			randGen := rand.New(rand.NewSource(time.Now().UnixNano()))
			randomIndex := randGen.Intn(len(coordinatorNodes))
			selectedNode := coordinatorNodes[randomIndex]
			log.Infof("Randomly selected coordinator node: %s", selectedNode.Name)

			// get pools of selected coordinator node
			poolIDs, err := GetAllPoolsOnNode(selectedNode.Id)
			log.FailOnError(err, "failed to get all pools on node [%s]", selectedNode.Id)
			log.Infof("Pool Ids of the node [%v]", poolIDs)
			// selecting one pool to resize
			poolIDToResize = poolIDs[0]
			log.Infof("Pool Id to resize [%v]", poolIDs)

			// map to store the method of resize on dmthin
			poolResizeMethod := map[api.SdkStoragePool_ResizeOperationType]string{
				api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK: "resize disk",
				api.SdkStoragePool_RESIZE_TYPE_AUTO:        "auto",
			}
			resizeMethodList := make([]api.SdkStoragePool_ResizeOperationType, 0, len(poolResizeMethod))
			for rm := range poolResizeMethod {
				resizeMethodList = append(resizeMethodList, rm)
			}

			for {
				// random pool expansion method on dmthin from the map
				randomResizeMethodIndex = rand.Intn(len(poolResizeMethod))
				poolToResize = getStoragePool(poolIDToResize)
				log.Infof(fmt.Sprintf("Pool going to resize is UUID: [%s]", poolIDToResize))
				originalSizeInBytes = poolToResize.TotalSize
				targetSizeInBytes = originalSizeInBytes + 1*units.TiB
				targetSizeGiB = targetSizeInBytes / units.GiB
				log.Infof("targetSizeGiB [%v] and resizeMethod [%v]", targetSizeGiB, poolResizeMethod[resizeMethodList[randomResizeMethodIndex]])

				if targetSizeGiB >= maxLimitSizeInGiB {
					// pool should not expand beyond maximum limit
					err := Inst().V.ExpandPool(poolIDToResize, resizeMethodList[randomResizeMethodIndex], targetSizeGiB, true)
					dash.VerifyFatal(strings.Contains(err.Error(), "cannot be expanded beyond maximum size 15 TiB"), true, fmt.Sprintf("Pool expansion beyond the maximum limit with Resize type %v failed?", poolResizeMethod[resizeMethodList[randomResizeMethodIndex]]))
					break
				}

				triggerPoolExpansion(poolIDToResize, targetSizeGiB, resizeMethodList[randomResizeMethodIndex])
				resizeErr = waitForOngoingPoolExpansionToComplete(poolIDToResize)
				dash.VerifyFatal(resizeErr, nil, fmt.Sprintf("Pool expansion with Resize type %v succeed?", poolResizeMethod[resizeMethodList[randomResizeMethodIndex]]))
				verifyPoolSizeEqualOrLargerThanExpected(poolIDToResize, targetSizeGiB)
			}
		})

		stepLog = "Validate and destroy applications"
		Step(stepLog, func() {
			for _, policyName := range policyNameList {
				err := storkops.Instance().DeleteSchedulePolicy(policyName)
				log.FailOnError(err, fmt.Sprintf("error deleting a SchedulePolicy [%s]", policyName))
			}
			appsValidateAndDestroy(contexts)
		})

	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts)
	})
})
