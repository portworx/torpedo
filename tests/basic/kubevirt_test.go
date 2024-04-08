package tests

import (
	context1 "context"
	"fmt"
	apapi "github.com/libopenstorage/autopilot-api/pkg/apis/autopilot/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	"github.com/portworx/torpedo/drivers/node"
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/drivers/volume"
	"github.com/portworx/torpedo/pkg/aututils"
	"github.com/portworx/torpedo/pkg/log"
	"github.com/portworx/torpedo/pkg/units"
	. "github.com/portworx/torpedo/tests"
	"time"
)

var _ = Describe("{AddNewDiskToKubevirtVM}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("AddNewDiskToKubevirtVM", "Add a new disk to a kubevirtVM", nil, 0)
	})
	var appCtxs []*scheduler.Context

	itLog := "Add a new disk to a kubevirtVM"
	It(itLog, func() {
		appList := Inst().AppList
		defer func() {
			Inst().AppList = appList
		}()
		numberOfVolumes := 1
		Inst().AppList = []string{"kubevirt-cirros-live-migration"}
		stepLog := "schedule a kubevirtVM"
		Step(stepLog, func() {
			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				appCtxs = append(appCtxs, ScheduleApplications("test")...)
			}
		})
		ValidateApplications(appCtxs)
		for _, appCtx := range appCtxs {
			bindMount, err := IsVMBindMounted(appCtx, false)
			log.FailOnError(err, "Failed to verify bind mount")
			dash.VerifyFatal(bindMount, true, "Failed to verify bind mount")
		}

		stepLog = "Add one disk to the kubevirt VM"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			success, err := AddDisksToKubevirtVM(appCtxs, numberOfVolumes, "0.5Gi")
			log.FailOnError(err, "Failed to add disks to kubevirt VM")
			dash.VerifyFatal(success, true, "Failed to add disks to kubevirt VM?")
		})

		stepLog = "Verify the new disk added is also bind mounted"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for _, appCtx := range appCtxs {
				isVmBindMounted, err := IsVMBindMounted(appCtx, true)
				log.FailOnError(err, "Failed to verify disks in kubevirt VM")
				if !isVmBindMounted {
					log.Errorf("The newly added disk to vm %s is not bind mounted", appCtx.App.Key)
				}
			}
		})

	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(appCtxs)
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
		Inst().AppList = []string{"kubevirt-cirros-live-migration"}

		stepLog := "schedule a kubevirt VM"
		Step(stepLog, func() {
			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				taskName := fmt.Sprintf("test-%v", i)
				appCtxs = append(appCtxs, ScheduleApplications(taskName)...)
			}
		})
		defer DestroyApps(appCtxs, nil)
		ValidateApplications(appCtxs)

		for _, appCtx := range appCtxs {
			bindMount, err := IsVMBindMounted(appCtx, false)
			log.FailOnError(err, "Failed to verify bind mount")
			dash.VerifyFatal(bindMount, true, "Failed to verify bind mount")
		}

		stepLog = "Live migrate the kubevirt VM"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for _, appCtx := range appCtxs {
				err := StartAndWaitForVMIMigration(appCtx, context1.TODO())
				log.FailOnError(err, "Failed to live migrate kubevirt VM")
			}
		})

	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(appCtxs)
	})
})

var _ = Describe("{KubeVirtPvcAndPoolExpandWithAutopilot}", func() {

	/*
		PWX:
			https://purestorage.atlassian.net/browse/PWX-36709
		TestRail:
			https://portworx.testrail.net/index.php?/cases/view/93652
			https://portworx.testrail.net/index.php?/cases/view/93653
	*/

	var (
		testName                string
		contexts                []*scheduler.Context
		pvcLabelSelector        = make(map[string]string)
		poolLabelSelector       = make(map[string]string)
		pvcAutoPilotRules       []apapi.AutopilotRule
		poolAutoPilotRules      []apapi.AutopilotRule
		selectedStorageNode     node.Node
		preResizeVolumeMap      = make(map[string]*volume.Volume)
		postResizeVolumeMap     = make(map[string]*volume.Volume)
		stopWaitForRunningChan  = make(chan struct{})
		waitForRunningErrorChan = make(chan error)
	)

	JustBeforeEach(func() {
		testName = "kv-pvc-pool-ap"
		tags := map[string]string{"poolChange": "true", "volumeChange": "true"}
		StartTorpedoTest("KubeVirtPvcAndPoolExpandWithAutopilot", "Kubevirt PVC and Pool expand test with autopilot", tags, 93652)
	})

	It("has to fill up the volume completely, resize the volumes and storage pool(s), validate and teardown apps", func() {
		log.InfoD("filling up the volume completely, resizing the volumes and storage pool(s), validating and tearing down apps")

		Step("Create autopilot rules for PVC and pool expand", func() {
			log.InfoD("Creating autopilot rules for PVC and pool expand")
			selectedStorageNode = node.GetStorageDriverNodes()[0]
			log.Infof("Selected storage node: %s", selectedStorageNode.Name)
			pvcLabelSelector = map[string]string{"autopilot": "pvc-expand"}
			pvcAutoPilotRules = []apapi.AutopilotRule{
				aututils.PVCRuleByUsageCapacity(5, 100, "100"),
			}
			poolLabelSelector = map[string]string{"autopilot": "adddisk"}
			poolAutoPilotRules = []apapi.AutopilotRule{
				aututils.PoolRuleByTotalSize((getTotalPoolSize(selectedStorageNode)/units.GiB)+1, 10, aututils.RuleScaleTypeAddDisk, poolLabelSelector),
			}
		})

		Step("schedule applications for PVC expand", func() {
			log.Infof("Scheduling apps with autopilot rules for PVC expand")
			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				for id, apRule := range pvcAutoPilotRules {
					taskName := fmt.Sprintf("%s-%d-aprule%d", testName, i, id)
					apRule.Name = fmt.Sprintf("%s-%d", apRule.Name, i)
					apRule.Spec.ActionsCoolDownPeriod = int64(60)
					context, err := Inst().S.Schedule(taskName, scheduler.ScheduleOptions{
						AppKeys:            Inst().AppList,
						StorageProvisioner: Inst().Provisioner,
						AutopilotRule:      apRule,
						Labels:             pvcLabelSelector,
					})
					log.FailOnError(err, "failed to schedule app [%s] with autopilot rule [%s]", taskName, apRule.Name)
					contexts = append(contexts, context...)
				}
			}
		})

		Step("Schedule apps with autopilot rules for pool expand", func() {
			log.InfoD("Scheduling apps with autopilot rules for pool expand")
			log.Infof("Adding labels [%s] on node: %s", poolLabelSelector, selectedStorageNode.Name)
			err := AddLabelsOnNode(selectedStorageNode, poolLabelSelector)
			log.FailOnError(err, "failed to add labels [%s] on node: %s", poolLabelSelector, selectedStorageNode.Name)
			contexts = scheduleAppsWithAutopilot(testName, Inst().GlobalScaleFactor, poolAutoPilotRules, scheduler.ScheduleOptions{PvcSize: 20 * units.GiB})
		})

		Step("Wait until workload completes on volume", func() {
			log.InfoD("Waiting for workload to complete on volume")
			for _, ctx := range contexts {
				err := Inst().S.WaitForRunning(ctx, workloadTimeout, retryInterval)
				log.FailOnError(err, "failed to wait for workload by app [%s] to be running", ctx.App.Key)
			}
			for _, ctx := range contexts {
				isVmBindMounted, err := IsVMBindMounted(ctx, true)
				log.FailOnError(err, fmt.Sprintf("failed to verify bind mount for app [%s]", ctx.App.Key))
				dash.VerifyFatal(isVmBindMounted, true, fmt.Sprintf("failed to verify bind mount for app [%s]", ctx.App.Key))
			}
			for _, ctx := range contexts {
				vols, err := Inst().S.GetVolumes(ctx)
				log.FailOnError(err, "failed to get volumes for app [%s]", ctx.App.Key)
				for _, vol := range vols {
					if vol.ID == "" {
						log.FailOnError(err, "failed to get volume ID for app [%s]", ctx.App.Key)
					}
					preResizeVolumeMap[vol.ID] = vol
				}
			}
		})

		Step("Ensure the app is running while resizing the volumes", func() {
			log.Infof("Ensuring the app is running while resizing the volumes")
			for _, ctx := range contexts {
				go func(ctx *scheduler.Context) {
					defer GinkgoRecover()
					for {
						select {
						case <-stopWaitForRunningChan:
							log.Infof("Stopping wait for running goroutine for app [%s]", ctx.App.Key)
							return
						default:
							err := Inst().S.WaitForRunning(ctx, workloadTimeout, retryInterval)
							if err != nil {
								err = fmt.Errorf("failed to wait for app [%s] to be running at [%v]. Err: [%v]", ctx.App.Key, time.Now(), err)
								waitForRunningErrorChan <- err
							}
						}
						time.Sleep(60 * time.Second)
					}
				}(ctx)
			}
		})

		Step("Validating volumes and verifying size of volumes", func() {
			log.InfoD("Validating volumes and verifying size of volumes")
			for _, ctx := range contexts {
				ValidateVolumes(ctx)
			}
		})

		Step("Validate storage pools", func() {
			log.InfoD("Validating storage pools")
			ValidateStoragePools(contexts)
		})

		Step("Wait for unscheduled resize of volume", func() {
			log.InfoD("Waiting for unscheduled resize of volume for [%v]", unscheduledResizeTimeout)
			time.Sleep(unscheduledResizeTimeout)
		})

		Step("Validating volumes and verifying size of volumes", func() {
			log.Infof("Validating volumes and verifying size of volumes")
			for _, ctx := range contexts {
				ValidateVolumes(ctx)
			}
		})

		Step("Validate storage pools", func() {
			log.InfoD("Validating storage pools")
			ValidateStoragePools(contexts)
			for _, ctx := range contexts {
				vols, err := Inst().S.GetVolumes(ctx)
				log.FailOnError(err, "failed to get volumes for app [%s]", ctx.App.Key)
				for _, vol := range vols {
					if vol.ID == "" {
						log.FailOnError(err, "failed to get volume ID for app [%s]", ctx.App.Key)
					}
					postResizeVolumeMap[vol.ID] = vol
				}
			}
			resizedVolumeCount := 0
			for preVolID, preVol := range preResizeVolumeMap {
				for postVolID, postVol := range postResizeVolumeMap {
					if preVolID == postVolID {
						if postVol.Size > preVol.Size {
							resizedVolumeCount += 1
						}
					}
				}
			}
			dash.VerifyFatal(resizedVolumeCount > 0, true, "No volumes resized")
		})
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts)
		log.InfoD("Destroying apps")
		log.InfoD("Closing stopWaitForRunningChan and waitForRunningErrorChan")
		close(stopWaitForRunningChan)
		close(waitForRunningErrorChan)
		var waitForRunningErrorList []error
		for err := range waitForRunningErrorChan {
			waitForRunningErrorList = append(waitForRunningErrorList, err)
		}
		dash.VerifyFatal(len(waitForRunningErrorList) == 0, true, fmt.Sprintf("Verifying if the app [%s] is running during resizing failed with errors: %v", testName, waitForRunningErrorList))
		opts := make(map[string]bool)
		opts[scheduler.OptionsWaitForResourceLeakCleanup] = true
		for _, ctx := range contexts {
			TearDownContext(ctx, opts)
		}
		log.InfoD("Removing autopilot rules and node labels")
		for _, apRule := range pvcAutoPilotRules {
			log.Infof("Deleting pvc autopilot rule [%s]", apRule.Name)
			err := Inst().S.DeleteAutopilotRule(apRule.Name)
			log.FailOnError(err, "failed to delete autopilot rule [%s]", apRule.Name)
		}
		for _, apRule := range poolAutoPilotRules {
			log.Infof("Deleting pool autopilot rule [%s]", apRule.Name)
			err := Inst().S.DeleteAutopilotRule(apRule.Name)
			log.FailOnError(err, "failed to delete pool autopilot rule [%s]", apRule.Name)
		}
		for k := range poolLabelSelector {
			log.Infof("Removing label [%s] on node: %s", k, selectedStorageNode.Name)
			err := Inst().S.RemoveLabelOnNode(selectedStorageNode, k)
			log.FailOnError(err, "failed to remove label [%s] on node: %s", k, selectedStorageNode.Name)
		}
	})
})
