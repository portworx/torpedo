package tests

import (
	context1 "context"
	. "github.com/onsi/ginkgo/v2"
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests"
)

var _ = Describe("{AddNewDiskToKubevirtVM}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("AddNewDiskToKubevirtVM", "Add a new disk to a kubevirtVM", nil, 0)
	})
	var contexts []*scheduler.Context

	itLog := "Add a new disk to a kubevirtVM"
	It(itLog, func() {
		appList := Inst().AppList
		defer func() {
			Inst().AppList = appList
		}()
		numberOfVolumes := 1
		Inst().AppList = []string{"kubevirt-cirros-cd-with-pvc"}
		stepLog := "schedule a kubevirtVM"
		Step(stepLog, func() {
			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				contexts = append(contexts, ScheduleApplications("test")...)
			}
		})
		defer DestroyApps(contexts, nil)
		ValidateApplications(contexts)
		stepLog = "Add one disk to the kubevirt VM"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			success, err := AddDisksToKubevirtVM(contexts, numberOfVolumes, "0.5Gi")
			log.FailOnError(err, "Failed to add disks to kubevirt VM")
			dash.VerifyFatal(success, true, "Failed to add disks to kubevirt VM?")
		})

	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts)
	})
})

var _ = Describe("{KubeVirtLiveMigration}", func() {

	JustBeforeEach(func() {
		StartTorpedoTest("KubeVirtLiveMigration", "Live migrate a kubevirtVM", nil, 0)
	})

	var appCtxs []*scheduler.Context

	itLog := "Live migrate a kubevirtVM"
	It(itLog, func() {
		log.InfoD(stepLog)
		appList := Inst().AppList
		defer func() {
			Inst().AppList = appList
		}()

		Inst().AppList = []string{"kubevirt-cirros-cd-with-pvc"}

		stepLog := "schedule a kubevirtVM"
		Step(stepLog, func() {
			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				appCtxs = append(appCtxs, ScheduleApplications("test")...)
			}
		})
		defer DestroyApps(appCtxs, nil)
		ValidateApplications(appCtxs)

		// Check if the vm's are bind mounted or not
		for _, appCtx := range appCtxs {
			success, err := IsVMBindMounted(appCtx, true)
			log.FailOnError(err, "Failed to check if VM is bind mounted")
			dash.VerifyFatal(success, true, "Failed to check if VM is bind mounted?")
		}

		stepLog = "Live migrate the kubevirt VM"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for _, appCtx := range appCtxs {
				success, err := StartAndWaitForVMIMigration(appCtx, context1.TODO())
				log.FailOnError(err, "Failed to live migrate kubevirt VM")
				dash.VerifyFatal(success, true, "Failed to live migrate kubevirt VM?")
			}
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts)
	})
})
