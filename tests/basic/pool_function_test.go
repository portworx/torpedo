package tests

import (
	"github.com/libopenstorage/openstorage/api"
	. "github.com/onsi/ginkgo"
	"github.com/portworx/torpedo/drivers/node"
	"github.com/portworx/torpedo/pkg/log"
	"github.com/portworx/torpedo/pkg/testrailuttils"
	"github.com/portworx/torpedo/pkg/units"
	. "github.com/portworx/torpedo/tests"
)

var stepLog string
var runID int
var testrailID int
var targetSizeGiB uint64
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

	testName = "PoolExpandDiskAdd"
	testDescription = "Validate storage pool expansion using add-disk option"
	It("select a pool that has I/O and expand it by 100 GiB with add-disk type. ", func() {
		StartTorpedoTest(testName, testDescription, nil, 0)
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

	testName = "PoolExpandDiskResize"
	testDescription = "Validate storage pool expansion using resize-disk option"
	It("select a pool that has I/O and expand it by 100 GiB with resize-disk type. ", func() {
		StartTorpedoTest(testName, testDescription, nil, 0)
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

	testName = "PoolExpandDiskAuto"
	testDescription = "Validate storage pool expansion using auto option"
	It("select a pool that has I/O and expand it by 100 GiB with auto type. ", func() {
		StartTorpedoTest(testName, testDescription, nil, 0)
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

	testrailID = 51309
	testName = "PoolExpandDiskAddWithReboot"
	testDescription = "Initiate pool expansion using add-disk and reboot node"
	It(testName+" - "+testDescription, func() {
		StartTorpedoTest(testName, testDescription, nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
		var storageNode *node.Node
		var err error
		Step("Select a pool that has I/O and expand it by 100 GiB with add-disk type. ", func() {
			originalSizeInBytes = poolToBeResized.TotalSize
			targetSizeInBytes = originalSizeInBytes + 100*units.GiB
			targetSizeGiB = targetSizeInBytes / units.GiB
			log.InfoD("Current Size of the pool %s is %d GiB. Trying to expand to %v GiB with type auto",
				poolIDToResize, poolToBeResized.TotalSize/units.GiB, targetSizeGiB)
			storageNode, err = GetNodeWithGivenPoolID(poolIDToResize)
			log.FailOnError(err, "Failed to get node with given pool ID")
			triggerPoolExpansion(poolIDToResize, targetSizeGiB, api.SdkStoragePool_RESIZE_TYPE_AUTO)
		})

		Step("Wait for expansion to start and reboot node", func() {
			err := WaitForExpansionToStart(poolIDToResize)
			log.FailOnError(err, "Timed out waiting for expansion to start")
			err = RebootNodeAndWait(*storageNode)
			log.FailOnError(err, "Failed to reboot node and wait till it is up")
		})

		Step("Ensure that new pool has been expanded to the expected size", func() {
			resizeErr := waitForOngoingPoolExpansionToComplete(poolIDToResize)
			dash.VerifyFatal(resizeErr, nil, "Pool expansion does not result in error")
			verifyPoolSizeEqualOrLargerThanExpected(poolIDToResize, targetSizeGiB)
		})
	})
})
