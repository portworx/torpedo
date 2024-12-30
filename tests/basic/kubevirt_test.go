package tests

import (
	context1 "context"
	"fmt"
	kubevirtdy "github.com/portworx/sched-ops/k8s/kubevirt-dynamic"
	"github.com/portworx/sched-ops/task"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	kubevirtv1 "kubevirt.io/api/core/v1"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/portworx/sched-ops/k8s/core"

	apapi "github.com/libopenstorage/autopilot-api/pkg/apis/autopilot/v1alpha1"
	oputil "github.com/pure-px/px-operator/pkg/util/test"

	. "github.com/onsi/ginkgo/v2"
	"github.com/pure-px/torpedo/drivers/node"
	"github.com/pure-px/torpedo/drivers/scheduler"
	"github.com/pure-px/torpedo/drivers/volume"
	"github.com/pure-px/torpedo/pkg/aututils"
	"github.com/pure-px/torpedo/pkg/log"
	"github.com/pure-px/torpedo/pkg/units"
	. "github.com/pure-px/torpedo/tests"
)

var _ = Describe("{AddNewDiskToKubevirtVM}", Label("p0", "positive", "kubevirt"), func() {
	JustBeforeEach(func() {
		StartTorpedoTest("AddNewDiskToKubevirtVM", "Add a new disk to a kubevirtVM", nil, 0)
	})
	var appCtxs []*scheduler.Context
	var namespace string
	itLog := "Add a new disk to a kubevirtVM"
	It(itLog, func() {
		pxNs, err := Inst().V.GetVolumeDriverNamespace()
		log.FailOnError(err, "Failed to get volume driver namespace")
		defer ListEvents(pxNs)

		appList := Inst().AppList
		defer func() {
			Inst().AppList = appList
		}()
		numberOfVolumes := 1
		Inst().AppList = []string{"kubevirt-debian-template"}
		stepLog := "Setting up Boot PVC Template"
		Step(stepLog, func() {
			template := ScheduleApplications("template")
			ValidateApplications(template)
		})

		Inst().AppList = []string{"kubevirt-debian-fio-minimal"}
		stepLog = "schedule a kubevirtVM"
		Step(stepLog, func() {
			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				namespace = fmt.Sprintf("kubevirt-%v", time.Now().Unix())
				appCtxs = append(appCtxs, ScheduleApplicationsOnNamespace(namespace, "test")...)
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
			_, err := AddDisksToKubevirtVM(appCtxs, numberOfVolumes, "10Gi")
			log.FailOnError(err, "Failed to add disks to kubevirt VM")
			dash.VerifyFatal(true, true, "Failed to add disks to kubevirt VM?")
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
		stepLog = "Destroy Applications"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			DestroyApps(appCtxs, nil)
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(appCtxs)
	})
})

var _ = Describe("{KubeVirtLiveMigration}", Label("p0", "positive", "kubevirt", "LiveMigration"), func() {
	JustBeforeEach(func() {
		StartTorpedoTest("KubeVirtLiveMigration", "Live migrate a kubevirtVM", nil, 0)
	})
	var appCtxs []*scheduler.Context
	var namespace string
	itLog := "Live migrate a kubevirtVM"
	It(itLog, func() {
		pxNs, err := Inst().V.GetVolumeDriverNamespace()
		log.FailOnError(err, "Failed to get volume driver namespace")
		defer ListEvents(pxNs)

		namespace = fmt.Sprintf("kubevirt-%v", time.Now().Unix())
		log.InfoD(stepLog)
		appList := Inst().AppList
		defer func() {
			Inst().AppList = appList
		}()
		Inst().AppList = []string{"kubevirt-debian-fio-minimal"}
		stepLog := "schedule a kubevirt VM"
		Step(stepLog, func() {
			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				taskName := fmt.Sprintf("test-%v", i)
				appCtxs = append(appCtxs, ScheduleApplicationsOnNamespace(namespace, taskName)...)
			}
		})
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
		stepLog = "Destroy Applications"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			DestroyApps(appCtxs, nil)
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(appCtxs)
	})
})

var _ = Describe("{PxKillBeforeAddDiskToVM}", Label("p1", "negative", "kubevirt", "error_injection", "px_crash"), func() {
	JustBeforeEach(func() {
		StartTorpedoTest("PxKillBeforeAddDiskToVM", "Kill Px on host node of Kubuevirt VM and then Add a disk", nil, 0)
	})
	var appCtxs []*scheduler.Context
	var nodes []string
	var namespace string
	itLog := "Kill Px then Add disk to Kubevirt VM"
	It(itLog, func() {
		pxNs, err := Inst().V.GetVolumeDriverNamespace()
		log.FailOnError(err, "Failed to get volume driver namespace")
		defer ListEvents(pxNs)

		namespace = fmt.Sprintf("kubevirt-%v", time.Now().Unix())
		appList := Inst().AppList
		defer func() {
			Inst().AppList = appList
		}()
		numberOfVolumes := 1
		Inst().AppList = []string{"kubevirt-debian-fio-minimal"}
		stepLog := "schedule a kubevirtVM"
		Step(stepLog, func() {
			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				appCtxs = append(appCtxs, ScheduleApplicationsOnNamespace(namespace, "test")...)
			}
		})
		ValidateApplications(appCtxs)
		for _, appCtx := range appCtxs {
			bindMount, err := IsVMBindMounted(appCtx, false)
			log.FailOnError(err, "Failed to verify bind mount")
			dash.VerifyFatal(bindMount, true, "Failed to verify bind mount")
		}
		stepLog = "Kill Px on node hosting VM"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			// Collect all nodes to restart Px on
			for _, appCtx := range appCtxs {
				vms, err := GetAllVMsFromScheduledContexts([]*scheduler.Context{appCtx})
				log.FailOnError(err, "Failed to get VMs from context")
				for _, vm := range vms {
					nodeName, err := GetNodeOfVM(vm)
					log.FailOnError(err, "Failed to get node of vm %v", vm.Name)
					nodes = append(nodes, nodeName)
				}
			}
			// Restart Px on all relevant nodes one by one
			for _, appNode := range node.GetStorageDriverNodes() {
				for _, vmNode := range nodes {
					if vmNode == appNode.Name {
						stepLog = fmt.Sprintf("stop volume driver %s on node: %s",
							Inst().V.String(), appNode.Name)
						Step(stepLog,
							func() {
								log.InfoD(stepLog)
								StopVolDriverAndWait([]node.Node{appNode})
							})

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
				}
			}
		})
		stepLog = "Add one disk to the kubevirt VM"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			_, err := AddDisksToKubevirtVM(appCtxs, numberOfVolumes, "10Gi")
			log.FailOnError(err, "Failed to add disks to kubevirt VM")
			dash.VerifyFatal(true, true, "Failed to add disks to kubevirt VM?")
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
		stepLog = "Destroy Applications"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			DestroyApps(appCtxs, nil)
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(appCtxs)
	})
})

var _ = Describe("{PxKillAfterAddDiskToVM}", Label("p1", "negative", "kubevirt", "error_injection", "px_crash"), func() {
	JustBeforeEach(func() {
		StartTorpedoTest("PxKillAfterAddDiskToVM", "Add a disk to Kubevirt VM, kill Px, Add another disk and validate the VM", nil, 0)
	})

	var appCtxs []*scheduler.Context
	var nodes []string
	var namespace string

	itLog := "Add disk to Kubevirt VM, Kill Px and then add another disk"
	It(itLog, func() {
		pxNs, err := Inst().V.GetVolumeDriverNamespace()
		log.FailOnError(err, "Failed to get volume driver namespace")
		defer ListEvents(pxNs)

		namespace = fmt.Sprintf("kubevirt-%v", time.Now().Unix())
		appList := Inst().AppList
		defer func() {
			Inst().AppList = appList
		}()
		numberOfVolumes := 1
		Inst().AppList = []string{"kubevirt-debian-fio-minimal"}
		stepLog := "schedule a kubevirtVM"
		Step(stepLog, func() {
			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				appCtxs = append(appCtxs, ScheduleApplicationsOnNamespace(namespace, "test")...)
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
			success, err := AddDisksToKubevirtVM(appCtxs, numberOfVolumes, "10Gi")
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
		stepLog = "Kill Px on node hosting VM"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			// Collect all nodes to restart Px on
			for _, appCtx := range appCtxs {
				vms, err := GetAllVMsFromScheduledContexts([]*scheduler.Context{appCtx})
				log.FailOnError(err, "Failed to get VMs from context")
				for _, vm := range vms {
					nodeName, err := GetNodeOfVM(vm)
					log.FailOnError(err, "Failed to get node of vm %v", vm.Name)
					nodes = append(nodes, nodeName)
				}
			}
			// Restart Px on all relevant nodes one by one
			for _, appNode := range node.GetStorageDriverNodes() {
				for _, vmNode := range nodes {
					if vmNode == appNode.Name {
						stepLog = fmt.Sprintf("stop volume driver %s on node: %s",
							Inst().V.String(), appNode.Name)
						Step(stepLog,
							func() {
								log.InfoD(stepLog)
								StopVolDriverAndWait([]node.Node{appNode})
							})

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
				}
			}
		})
		stepLog = "Verify the disks added are still bind mounted"
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
		stepLog = "Destroy Applications"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			DestroyApps(appCtxs, nil)
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(appCtxs)
	})
})

var _ = Describe("{KubevirtVMVolHaIncrease}", Label("p0", "positive", "kubevirt", "HA_Increase_Decrease"), func() {
	JustBeforeEach(func() {
		StartTorpedoTest("KubevirtVMVolHaIncrease", "Increase the volume HA of a kubevirt VM", nil, 0)
	})
	var appCtxs []*scheduler.Context
	var namespace string
	itLog := "Increase the volume HA of a kubevirt VM"
	It(itLog, func() {
		pxNs, err := Inst().V.GetVolumeDriverNamespace()
		log.FailOnError(err, "Failed to get volume driver namespace")
		defer ListEvents(pxNs)

		namespace = fmt.Sprintf("kubevirt-%v", time.Now().Unix())
		appList := Inst().AppList
		defer func() {
			Inst().AppList = appList
		}()
		Inst().AppList = []string{"kubevirt-debian-fio-low-ha"}
		stepLog := "schedule a kubevirt VM"
		Step(stepLog, func() {
			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				appCtxs = append(appCtxs, ScheduleApplicationsOnNamespace(namespace, "test")...)
			}
		})
		ValidateApplications(appCtxs)
		for _, appCtx := range appCtxs {
			bindMount, err := IsVMBindMounted(appCtx, false)
			log.FailOnError(err, "Failed to verify bind mount")
			dash.VerifyFatal(bindMount, true, "Failed to verify bind mount")
		}
		stepLog = "Increase the volume HA of the kubevirt VM Volumes"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for _, appCtx := range appCtxs {
				vols, err := Inst().S.GetVolumes(appCtx)
				log.FailOnError(err, "Failed to get volumes of kubevirt VM")
				for _, vol := range vols {
					currRep, err := Inst().V.GetReplicationFactor(vol)
					log.FailOnError(err, "Failed to get Repl factor for vil %s", vol.Name)

					if currRep < 3 {
						opts := volume.Options{
							ValidateReplicationUpdateTimeout: validateReplicationUpdateTimeout,
						}
						err = Inst().V.SetReplicationFactor(vol, currRep+1, nil, nil, true, opts)
						dash.VerifyFatal(err, nil, fmt.Sprintf("Validate set repl factor to %d", currRep+1))
					} else {
						log.Warnf("Volume %s has reached maximum replication factor", vol.Name)
					}
				}
			}
		})
		stepLog = "Verify if VM's are still bind mounted even after HA increase"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for _, appCtx := range appCtxs {
				isVmBindMounted, err := IsVMBindMounted(appCtx, true)
				log.FailOnError(err, "Failed to run vm bind mount check")
				if !isVmBindMounted {
					log.Errorf("The newly added replication to vm %s is not bind mounted", appCtx.App.Key)
				}
			}
		})
		stepLog = "Destroy Applications"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			DestroyApps(appCtxs, nil)
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(appCtxs)
	})
})

var _ = Describe("{KubevirtVMVolHaDecrease}", Label("p0", "positive", "kubevirt", "HA_Increase_Decrease"), func() {
	JustBeforeEach(func() {
		StartTorpedoTest("KubevirtVMVolHaDecrease", "Decrease the replication factor of kubevirt Vms", nil, 0)
	})
	var appCtxs []*scheduler.Context
	var namespace string
	itLog := "Decrease the volume HA of a kubevirt VM"
	It(itLog, func() {
		pxNs, err := Inst().V.GetVolumeDriverNamespace()
		log.FailOnError(err, "Failed to get volume driver namespace")
		defer ListEvents(pxNs)

		namespace = fmt.Sprintf("kubevirt-%v", time.Now().Unix())
		log.InfoD(stepLog)
		appList := Inst().AppList
		defer func() {
			Inst().AppList = appList
		}()

		Inst().AppList = []string{"kubevirt-debian-fio-minimal"}
		stepLog := "schedule a kubevirt VM"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				appCtxs = append(appCtxs, ScheduleApplicationsOnNamespace(namespace, "test")...)
			}
		})
		ValidateApplications(appCtxs)
		for _, appCtx := range appCtxs {
			bindMount, err := IsVMBindMounted(appCtx, false)
			log.FailOnError(err, "Failed to verify bind mount")
			dash.VerifyFatal(bindMount, true, "Failed to verify bind mount")
		}

		stepLog = "Decrease the volume HA of the kubevirt VM Volumes"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for _, appCtx := range appCtxs {
				vols, err := Inst().S.GetVolumes(appCtx)
				log.FailOnError(err, "Failed to get volumes of kubevirt VM")
				for _, vol := range vols {
					currRep, err := Inst().V.GetReplicationFactor(vol)
					log.FailOnError(err, "Failed to get Repl factor for vil %s", vol.Name)

					if currRep > 1 {
						opts := volume.Options{
							ValidateReplicationUpdateTimeout: validateReplicationUpdateTimeout,
						}
						err = Inst().V.SetReplicationFactor(vol, currRep-1, nil, nil, true, opts)
						dash.VerifyFatal(err, nil, fmt.Sprintf("Validate set repl factor to %d", currRep-1))
					} else {
						log.Warnf("Volume %s has reached maximum replication factor", vol.Name)
					}
				}
			}
		})
		stepLog = "Verify if VM's are still bind mounted even after HA Decrease"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for _, appCtx := range appCtxs {
				isVmBindMounted, err := IsVMBindMounted(appCtx, true)
				log.FailOnError(err, "Failed to run vm bind mount check")
				if !isVmBindMounted {
					log.Errorf("The newly added disk to vm %s is not bind mounted", appCtx.App.Key)
				}
			}
		})
		stepLog = "Destroy Applications"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			DestroyApps(appCtxs, nil)
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(appCtxs)
	})
})

var _ = Describe("{LiveMigrationBeforeAddDisk}", Label("p0", "positive", "kubevirt", "LiveMigration"), func() {
	JustBeforeEach(func() {
		StartTorpedoTest("LiveMigrationBeforeAddDisk", "Live Migrate a VM Before Adding a new disk to a kubevirtVM", nil, 0)
	})
	var appCtxs []*scheduler.Context
	var namespace string
	itLog := "Live Migrate a VM and then add a new disk to a kubevirtVM"
	It(itLog, func() {
		pxNs, err := Inst().V.GetVolumeDriverNamespace()
		log.FailOnError(err, "Failed to get volume driver namespace")
		defer ListEvents(pxNs)

		namespace = fmt.Sprintf("kubevirt-%v", time.Now().Unix())
		appList := Inst().AppList
		defer func() {
			Inst().AppList = appList
		}()
		numberOfVolumes := 1
		Inst().AppList = []string{"kubevirt-debian-fio-minimal"}
		stepLog := "schedule a kubevirtVM"
		Step(stepLog, func() {
			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				appCtxs = append(appCtxs, ScheduleApplicationsOnNamespace(namespace, "test")...)
			}
		})
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
		stepLog = "Add one disk to the kubevirt VM"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			success, err := AddDisksToKubevirtVM(appCtxs, numberOfVolumes, "10Gi")
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
		stepLog = "Destroy Applications"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			DestroyApps(appCtxs, nil)
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(appCtxs)
	})
})

var _ = Describe("{AddDiskAndLiveMigrate}", Label("p0", "positive", "kubevirt", "LiveMigration"), func() {
	JustBeforeEach(func() {
		StartTorpedoTest("AddDiskAndLiveMigrate", "Live Migrate a VM After Adding a new disk to a kubevirtVM", nil, 0)
	})
	var appCtxs []*scheduler.Context
	var namespace string
	itLog := "Add a new disk to a kubevirtVM and then Live Migrate"
	It(itLog, func() {
		pxNs, err := Inst().V.GetVolumeDriverNamespace()
		log.FailOnError(err, "Failed to get volume driver namespace")
		defer ListEvents(pxNs)

		namespace = fmt.Sprintf("kubevirt-%v", time.Now().Unix())
		appList := Inst().AppList
		defer func() {
			Inst().AppList = appList
		}()
		numberOfVolumes := 1
		Inst().AppList = []string{"kubevirt-debian-fio-minimal"}
		stepLog := "schedule a kubevirtVM"
		Step(stepLog, func() {
			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				appCtxs = append(appCtxs, ScheduleApplicationsOnNamespace(namespace, "test")...)
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
			success, err := AddDisksToKubevirtVM(appCtxs, numberOfVolumes, "10Gi")
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
		stepLog = "Live migrate the kubevirt VM"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for _, appCtx := range appCtxs {
				err := StartAndWaitForVMIMigration(appCtx, context1.TODO())
				log.FailOnError(err, "Failed to live migrate kubevirt VM")
			}
		})
		stepLog = "Destroy Applications"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			DestroyApps(appCtxs, nil)
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(appCtxs)
	})
})

var _ = Describe("{KubeVirtPvcAndPoolExpandWithAutopilot}", Label("p0", "positive", "kubevirt", "autopilot", "PvcResize", "PoolExpand"), func() {
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
		pxNs, err := Inst().V.GetVolumeDriverNamespace()
		log.FailOnError(err, "Failed to get volume driver namespace")
		defer ListEvents(pxNs)

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
		Step("Verify bind mount after volume resize", func() {
			log.InfoD("Verify bind mount after volume resize")
			for _, ctx := range contexts {
				isVmBindMounted, err := IsVMBindMounted(ctx, true)
				log.FailOnError(err, fmt.Sprintf("failed to verify bind mount for app [%s]", ctx.App.Key))
				dash.VerifyFatal(isVmBindMounted, true, fmt.Sprintf("failed to verify bind mount for app [%s]", ctx.App.Key))
			}
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

var _ = Describe("{UpgradeOCPAndValidateKubeVirtApps}", Label("p1", "positive", "kubevirt", "Upgrade"), func() {
	JustBeforeEach(func() {
		StartTorpedoTest("UpgradeClusterAndValidateKubeVirt", "Upgrade OCP cluster and validate kubevirt apps", nil, 0)
	})

	var appCtxs []*scheduler.Context

	itLog := "Upgrade OCP cluster and validate kubevirt apps"
	It(itLog, func() {
		pxNs, err := Inst().V.GetVolumeDriverNamespace()
		log.FailOnError(err, "Failed to get volume driver namespace")
		defer ListEvents(pxNs)

		stepLog := "schedule kubevirt VMs"
		Step(stepLog, func() {
			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				taskName := fmt.Sprintf("test-%v", i)
				appCtxs = append(appCtxs, ScheduleApplications(taskName)...)
			}
		})
		stepLog = "validate kubevirt apps before upgrade"
		Step(stepLog, func() {
			ValidateApplications(appCtxs)
			for _, appCtx := range appCtxs {
				isVmBindMounted, err := IsVMBindMounted(appCtx, false)
				log.FailOnError(err, "Failed to verify bind mount")
				dash.VerifyFatal(isVmBindMounted, true, "Failed to verify bind mount?")
			}
		})

		var versions []string
		if len(Inst().SchedUpgradeHops) > 0 {
			versions = strings.Split(Inst().SchedUpgradeHops, ",")
		}
		if len(versions) == 0 {
			log.Fatalf("No versions to upgrade")
			return
		}
		for _, version := range versions {
			Step(fmt.Sprintf("start [%s] scheduler upgrade to version [%s]", Inst().S.String(), version), func() {
				stopSignal := make(chan struct{})

				var mError error
				opver, err := oputil.GetPxOperatorVersion()
				if err == nil && opver.GreaterThanOrEqual(PDBValidationMinOpVersion) {
					go DoPDBValidation(stopSignal, &mError)
					defer func() {
						close(stopSignal)
					}()
				} else {
					log.Warnf("PDB validation skipped. Current Px-Operator version: [%s], minimum required: [%s]. Error: [%v].", opver, PDBValidationMinOpVersion, err)
				}

				err = Inst().S.UpgradeScheduler(version)
				dash.VerifyFatal(mError, nil, "validation of PDB of px-storage during cluster upgrade successful")
				dash.VerifyFatal(err, nil, fmt.Sprintf("verify [%s] upgrade to [%s] is successful", Inst().S.String(), version))

				PrintK8sClusterInfo()
			})

			Step("validate storage components", func() {
				urlToParse := fmt.Sprintf("%s/%s", Inst().StorageDriverUpgradeEndpointURL, Inst().StorageDriverUpgradeEndpointVersion)
				u, err := url.Parse(urlToParse)
				log.FailOnError(err, fmt.Sprintf("error parsing PX version the url [%s]", urlToParse))
				err = Inst().V.ValidateDriver(u.String(), true)
				dash.VerifyFatal(err, nil, fmt.Sprintf("verify volume driver after upgrade to %s", version))

				// Printing cluster node info after the upgrade
				PrintK8sClusterInfo()
			})

			Step("update node drive endpoints", func() {
				// Update NodeRegistry, this is needed as node names and IDs might change after upgrade
				err = Inst().S.RefreshNodeRegistry()
				log.FailOnError(err, "Refresh Node Registry failed")

				// Refresh Driver Endpoints
				err = Inst().V.RefreshDriverEndpoints()
				log.FailOnError(err, "Refresh Driver Endpoints failed")

				// Printing pxctl status after the upgrade
				PrintPxctlStatus()
			})

			stepLog = "validate kubevirt apps after upgrade and destroy"
			Step(stepLog, func() {
				ValidateApplications(appCtxs)
				for _, ctx := range appCtxs {
					isVmBindMounted, err := IsVMBindMounted(ctx, false)
					log.FailOnError(err, "Failed to verify bind mount")
					dash.VerifyFatal(isVmBindMounted, true, "Failed to verify bind mount?")
				}
				DestroyApps(appCtxs, nil)
			})
		}
	})
})

var _ = Describe("{RebootRootDiskAttachedNode}", Label("p1", "negative", "error_injection", "kubevirt", "node_reboot"), func() {
	JustBeforeEach(func() {
		StartTorpedoTest("RebootRootDiskAttachedNode", "Reboot the node where VMs root disk is attached", nil, 0)
		DeployVMTemplatesAndValidate()
	})
	var appCtxs []*scheduler.Context

	itLog := "Reboot node where Kubevirt VMs root disk is attached"
	It(itLog, func() {
		pxNs, err := Inst().V.GetVolumeDriverNamespace()
		log.FailOnError(err, "Failed to get volume driver namespace")
		defer ListEvents(pxNs)

		appList := Inst().AppList
		defer func() {
			Inst().AppList = appList
		}()
		Inst().AppList = []string{"kubevirt-debian-fio-minimal"}
		stepLog := "schedule a kubevirtVM"
		Step(stepLog, func() {
			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				appCtxs = append(appCtxs, ScheduleApplications("reboot")...)
			}
		})
		defer DestroyApps(appCtxs, nil)
		ValidateApplications(appCtxs)
		for _, appCtx := range appCtxs {
			bindMount, err := IsVMBindMounted(appCtx, false)
			log.FailOnError(err, "Failed to verify bind mount after initial deploy")
			dash.VerifyFatal(bindMount, true, "Failed to verify bind mount after intial deploy")
		}

		stepLog = "Get node where VM's root disk is attached and reboot that node"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for _, virtualMachineCtx := range appCtxs {
				bindMount, err := IsVMBindMounted(virtualMachineCtx, false)
				log.FailOnError(err, "Failed to verify bind mount pre node reboot in namespace: %s", virtualMachineCtx.App.NameSpace)
				dash.VerifyFatal(bindMount, true, "Failed to verify bind mount pre node reboot")

				vms, err := GetAllVMsFromScheduledContexts([]*scheduler.Context{virtualMachineCtx})
				log.FailOnError(err, "Failed to get VMs from scheduled contexts")
				dash.VerifyFatal(len(vms) > 0, true, "Failed to to get VMs from scheduled contexts")

				for _, vm := range vms {
					nodeName, err := GetNodeOfVM(vm)
					log.FailOnError(err, "Failed to get node name for VM: %s", vm.Name)
					log.Infof("Pre-reboot VM [%s] in namespace [%s] is scheduled on node [%s]. Rebooting it.", vm.Name, vm.Namespace, nodeName)
					nodeObj, err := node.GetNodeByName(nodeName)
					log.FailOnError(err, "Failed to get node obj for node name: %s", nodeName)
					err = Inst().N.RebootNodeAndWait(nodeObj)
					log.FailOnError(err, "Failed to reboot  node: %s", nodeObj.Name)
					log.Infof("Succesfully rebooted node: %s", nodeObj.Name)
				}
				ValidateApplications(appCtxs)
				// Get updated VM list and validate bind mount again
				// TODO: PTX-23439 Add validation that VM started on a different node than it's original node
				vms, err = GetAllVMsFromScheduledContexts([]*scheduler.Context{virtualMachineCtx})
				log.FailOnError(err, "Failed to get VMs from scheduled contexts")
				dash.VerifyFatal(len(vms) > 0, true, "Failed to to get VMs from scheduled contexts")
				for _, vm := range vms {
					nodeName, err := GetNodeOfVM(vm)
					log.FailOnError(err, "Failed to get node name for VM: %s", vm.Name)
					log.Infof("Post reboot VM [%s] in namespace [%s] is scheduled on node [%s]", vm.Name, vm.Namespace, nodeName)
				}
				bindMount, err = IsVMBindMounted(virtualMachineCtx, false)
				log.FailOnError(err, "Failed to verify bind mount post node reboot in namespace: %s", virtualMachineCtx.App.NameSpace)
				dash.VerifyFatal(bindMount, true, "Failed to verify bind mount pre node reboot")
			}
			ValidateApplications(appCtxs)
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(appCtxs)
	})
})

var _ = Describe("{ParallelAddDiskToVM}", Label("p1", "postive", "kubevirt", "shared_v4", "MiniScale"), func() {
	JustBeforeEach(func() {
		StartTorpedoTest("ParallelAddDiskToVM", "Add a new disk to multiple kubevirtVM parallely", nil, 0)
	})
	var appCtxs []*scheduler.Context
	var namespace string
	var wg sync.WaitGroup

	itLog := "Add a new disk to multiple kubevirtVM"
	It(itLog, func() {
		pxNs, err := Inst().V.GetVolumeDriverNamespace()
		log.FailOnError(err, "Failed to get volume driver namespace")
		defer ListEvents(pxNs)

		appList := Inst().AppList
		defer func() {
			Inst().AppList = appList
		}()
		numberOfVolumes := 1
		Inst().AppList = []string{"kubevirt-debian-template"}
		stepLog := "Setting up Boot PVC Template"
		Step(stepLog, func() {
			template := ScheduleApplications("template")
			ValidateApplications(template)
		})

		Inst().AppList = []string{"kubevirt-debian-fio-minimal"}
		stepLog = "schedule a kubevirtVM"
		Step(stepLog, func() {
			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				namespace = fmt.Sprintf("kubevirt-%v", time.Now().Unix())
				appCtxs = append(appCtxs, ScheduleApplicationsOnNamespace(namespace, "test")...)
			}
		})
		ValidateApplications(appCtxs)
		for _, appCtx := range appCtxs {
			bindMount, err := IsVMBindMounted(appCtx, false)
			log.FailOnError(err, "Failed to verify bind mount")
			dash.VerifyFatal(bindMount, true, "Failed to verify bind mount")
		}

		stepLog = "Add one disk to multiple kubevirt VM at the same time"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for _, appCtx := range appCtxs {
				wg.Add(1)
				go func(appCtx *scheduler.Context) {
					defer GinkgoRecover()
					defer wg.Done()
					_, err := AddDisksToKubevirtVM([]*scheduler.Context{appCtx}, numberOfVolumes, "10Gi")
					log.FailOnError(err, "Failed to add disks to kubevirt VM")
					dash.VerifyFatal(true, true, "Failed to add disks to kubevirt VM?")
				}(appCtx)
			}
		})
		wg.Wait()

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
		stepLog = "Destroy Applications"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			DestroyApps(appCtxs, nil)
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(appCtxs)
	})
})

var _ = Describe("{MultipleKubeVirtLiveMigration}", Label("p0", "postive", "kubevirt", "MiniScale", "LiveMigration"), func() {
	JustBeforeEach(func() {
		StartTorpedoTest("MultipleKubeVirtLiveMigration", "Live migrate multiple kubevirtVM's parallely", nil, 0)
	})
	var appCtxs []*scheduler.Context
	var namespace string
	var wg sync.WaitGroup
	var canSsh bool

	itLog := "Live migrate multiple kubevirtVM's parallely"
	It(itLog, func() {
		pxNs, err := Inst().V.GetVolumeDriverNamespace()
		log.FailOnError(err, "Failed to get volume driver namespace")
		defer ListEvents(pxNs)
		canSsh = false
		log.InfoD(stepLog)
		appList := Inst().AppList
		defer func() {
			Inst().AppList = appList
		}()
		Inst().AppList = []string{"kubevirt-fada-raw-fio"}
		Inst().CsiAppList = []string{"kubevirt-fada-raw-fio"}
		stepLog := "schedule a kubevirt VM"
		Step(stepLog, func() {
			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				namespace = fmt.Sprintf("kubevirt-%v", time.Now().Unix())
				appCtxs = append(appCtxs, ScheduleApplicationsOnNamespace(namespace, "test")...)
			}
		})
		ValidateApplications(appCtxs)
		for _, appCtx := range appCtxs {
			bindMount, err := IsVMBindMounted(appCtx, false)
			log.FailOnError(err, "Failed to verify bind mount")
			dash.VerifyFatal(bindMount, true, "Failed to verify bind mount")
		}

		log.Infof("Hard Sleep for 2 minutes to let VMs come up")
		time.Sleep(2 * time.Minute)

		stepLog = "Create SSH Pod"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			err = CreateSSHPod()
			if err == nil {
				canSsh = true
			}
		})
		if canSsh {
			stepLog = "Validate fio is running in the VM or not"
			Step(stepLog, func() {
				log.InfoD(stepLog)
				for _, appCtx := range appCtxs {
					wg.Add(1)
					go func(appCtx *scheduler.Context) {
						// Get the VMs from the appCtx
						vms, err := GetAllVMsFromScheduledContexts([]*scheduler.Context{appCtx})
						log.FailOnError(err, "Failed to get VMs from appCtx")
						defer GinkgoRecover()
						defer wg.Done()
						for _, vm := range vms {
							err = CheckFioIsRunningInVM(vm)
							log.FailOnError(err, "Failed to validate fio in VM %s", vm.Name)
						}
					}(appCtx)
				}
			})
		}
		wg.Wait()
		stepLog = "Live migrate the kubevirt VM"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for _, appCtx := range appCtxs {
				wg.Add(1)
				go func(appCtx *scheduler.Context) {
					defer GinkgoRecover()
					defer wg.Done()
					err := StartAndWaitForVMIMigration(appCtx, context1.TODO())
					log.FailOnError(err, "Failed to live migrate kubevirt VM")
				}(appCtx)
			}
		})
		wg.Wait()
		if canSsh {
			stepLog = "Validate fio is running in the VM or not"
			Step(stepLog, func() {
				log.InfoD(stepLog)
				for _, appCtx := range appCtxs {
					wg.Add(1)
					go func(appCtx *scheduler.Context) {
						// Get the VMs from the appCtx
						vms, err := GetAllVMsFromScheduledContexts([]*scheduler.Context{appCtx})
						log.FailOnError(err, "Failed to get VMs from appCtx")
						defer GinkgoRecover()
						defer wg.Done()
						for _, vm := range vms {
							err = CheckFioIsRunningInVM(vm)
							log.FailOnError(err, "Failed to validate fio in VM %s", vm.Name)
						}
					}(appCtx)
				}
			})
		}
		wg.Wait()
		stepLog = "Destroy Applications"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			DestroyApps(appCtxs, nil)
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(appCtxs)
	})
})

var _ = Describe("{AddDiskAndLiveMigrateMultipleVm}", Label("p1", "postive", "kubevirt", "MiniScale", "LiveMigration"), func() {
	JustBeforeEach(func() {
		StartTorpedoTest("AddDiskAndLiveMigrateMultipleVm", "Live Migrate multiple VM's After Adding a new disk to a kubevirtVM", nil, 0)
	})
	var appCtxs []*scheduler.Context
	var namespace string
	var wg sync.WaitGroup

	itLog := "Add a new disk to multiple kubevirtVM and then Live Migrate them parallely"
	It(itLog, func() {
		pxNs, err := Inst().V.GetVolumeDriverNamespace()
		log.FailOnError(err, "Failed to get volume driver namespace")
		defer ListEvents(pxNs)

		appList := Inst().AppList
		defer func() {
			Inst().AppList = appList
		}()
		numberOfVolumes := 1
		Inst().AppList = []string{"kubevirt-debian-fio-minimal"}
		stepLog := "schedule a kubevirtVM"
		Step(stepLog, func() {
			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				namespace = fmt.Sprintf("kubevirt-%v", time.Now().Unix())
				appCtxs = append(appCtxs, ScheduleApplicationsOnNamespace(namespace, "test")...)
			}
		})
		ValidateApplications(appCtxs)
		for _, appCtx := range appCtxs {
			bindMount, err := IsVMBindMounted(appCtx, false)
			log.FailOnError(err, "Failed to verify bind mount")
			dash.VerifyFatal(bindMount, true, "Failed to verify bind mount")
		}

		stepLog = "Add one disk to the kubevirt VM's and check if new added disk is bind mounted and live migrate the vms parallely"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for _, appCtx := range appCtxs {
				wg.Add(1)
				go func(appCtx *scheduler.Context) {
					defer GinkgoRecover()
					defer wg.Done()

					success, err := AddDisksToKubevirtVM([]*scheduler.Context{appCtx}, numberOfVolumes, "10Gi")
					log.FailOnError(err, "Failed to add disks to kubevirt VM")
					dash.VerifyFatal(success, true, "Failed to add disks to kubevirt VM?")

					isVmBindMounted, err := IsVMBindMounted(appCtx, true)
					log.FailOnError(err, "Failed to verify disks in kubevirt VM")
					if !isVmBindMounted {
						log.Errorf("The newly added disk to vm %s is not bind mounted", appCtx.App.Key)
					}
					err = StartAndWaitForVMIMigration(appCtx, context1.TODO())
					log.FailOnError(err, "Failed to live migrate kubevirt VM")
				}(appCtx)
			}
		})
		wg.Wait()

		stepLog = "Destroy Applications"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			DestroyApps(appCtxs, nil)
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(appCtxs)
	})
})

var _ = Describe("{LiveMigrationBeforeAddDiskMultipleVm}", Label("p1", "postive", "kubevirt", "LiveMigration"), func() {
	JustBeforeEach(func() {
		StartTorpedoTest("LiveMigrationBeforeAddDiskMultipleVm", "Live Migrate multiple VM's Before Adding a new disk to a kubevirtVM parallely", nil, 0)
	})
	var appCtxs []*scheduler.Context
	var namespace string
	var wg sync.WaitGroup

	itLog := "Live Migrate multiple VM's and then add a new disk to a kubevirtVM"
	It(itLog, func() {
		pxNs, err := Inst().V.GetVolumeDriverNamespace()
		log.FailOnError(err, "Failed to get volume driver namespace")
		defer ListEvents(pxNs)

		appList := Inst().AppList
		defer func() {
			Inst().AppList = appList
		}()
		numberOfVolumes := 1
		Inst().AppList = []string{"kubevirt-debian-fio-minimal"}
		stepLog := "schedule a kubevirtVM"
		Step(stepLog, func() {
			for i := 0; i < Inst().GlobalScaleFactor; i++ {

				namespace = fmt.Sprintf("kubevirt-%v", time.Now().Unix())
				appCtxs = append(appCtxs, ScheduleApplicationsOnNamespace(namespace, "test")...)
			}
		})
		ValidateApplications(appCtxs)
		for _, appCtx := range appCtxs {
			bindMount, err := IsVMBindMounted(appCtx, false)
			log.FailOnError(err, "Failed to verify bind mount")
			dash.VerifyFatal(bindMount, true, "Failed to verify bind mount")
		}
		stepLog = "Live migrate the kubevirt VM's,Add one disk to the kubevirt VM and verify the new disk added is also bind mounted"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for _, appCtx := range appCtxs {
				wg.Add(1)
				go func(appCtx *scheduler.Context) {
					defer GinkgoRecover()
					defer wg.Done()
					err := StartAndWaitForVMIMigration(appCtx, context1.TODO())
					log.FailOnError(err, "Failed to live migrate kubevirt VM")

					success, err := AddDisksToKubevirtVM([]*scheduler.Context{appCtx}, numberOfVolumes, "10Gi")
					log.FailOnError(err, "Failed to add disks to kubevirt VM")
					dash.VerifyFatal(success, true, "Failed to add disks to kubevirt VM?")

					isVmBindMounted, err := IsVMBindMounted(appCtx, true)
					log.FailOnError(err, "Failed to verify disks in kubevirt VM")
					if !isVmBindMounted {
						log.Errorf("The newly added disk to vm %s is not bind mounted", appCtx.App.Key)
					}
				}(appCtx)
			}
		})
		stepLog = "Destroy Applications"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			DestroyApps(appCtxs, nil)
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(appCtxs)
	})
})

var _ = Describe("{MultipleVMVolHaIncrease}", Label("p1", "postive", "kubevirt", "HA_Increase_Decrease", "MiniScale"), func() {
	JustBeforeEach(func() {
		StartTorpedoTest("MultipleVMVolHaIncrease", "Increase the volume HA of multiple kubevirt VM parallely", nil, 0)
	})
	var appCtxs []*scheduler.Context
	var namespace string
	var wg sync.WaitGroup

	itLog := "Increase the volume HA of multiple kubevirt VM parallely"
	It(itLog, func() {
		pxNs, err := Inst().V.GetVolumeDriverNamespace()
		log.FailOnError(err, "Failed to get volume driver namespace")
		defer ListEvents(pxNs)

		appList := Inst().AppList
		defer func() {
			Inst().AppList = appList
		}()
		Inst().AppList = []string{"kubevirt-debian-fio-low-ha"}
		stepLog := "schedule a kubevirt VM"
		Step(stepLog, func() {
			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				namespace = fmt.Sprintf("kubevirt-%v", time.Now().Unix())
				appCtxs = append(appCtxs, ScheduleApplicationsOnNamespace(namespace, "test")...)
			}
		})
		ValidateApplications(appCtxs)
		for _, appCtx := range appCtxs {
			bindMount, err := IsVMBindMounted(appCtx, false)
			log.FailOnError(err, "Failed to verify bind mount")
			dash.VerifyFatal(bindMount, true, "Failed to verify bind mount")
		}
		stepLog = "Increase the volume HA of the multiple kubevirt VM Volumes and check if they are bind mounted parallely"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for _, appCtx := range appCtxs {
				wg.Add(1)
				go func(appCtx *scheduler.Context) {
					defer GinkgoRecover()
					defer wg.Done()
					vols, err := Inst().S.GetVolumes(appCtx)
					log.FailOnError(err, "Failed to get volumes of kubevirt VM")
					for _, vol := range vols {
						currRep, err := Inst().V.GetReplicationFactor(vol)
						log.FailOnError(err, "Failed to get Repl factor for vil %s", vol.Name)

						if currRep < 3 {
							opts := volume.Options{
								ValidateReplicationUpdateTimeout: validateReplicationUpdateTimeout,
							}
							err = Inst().V.SetReplicationFactor(vol, currRep+1, nil, nil, true, opts)
							dash.VerifyFatal(err, nil, fmt.Sprintf("Validate set repl factor to %d", currRep+1))
						} else {
							log.Warnf("Volume %s has reached maximum replication factor", vol.Name)
						}
					}
					isVmBindMounted, err := IsVMBindMounted(appCtx, true)
					log.FailOnError(err, "Failed to run vm bind mount check")
					if !isVmBindMounted {
						log.Errorf("The newly added replication to vm %s is not bind mounted", appCtx.App.Key)
					}
				}(appCtx)
			}
		})

		wg.Wait()
		stepLog = "Destroy Applications"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			DestroyApps(appCtxs, nil)
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(appCtxs)
	})
})

var _ = Describe("{MultipleVMVolHaDecrease}", Label("p1", "postive", "kubevirt", "HA_Increase_Decrease", "MiniScale"), func() {
	JustBeforeEach(func() {
		StartTorpedoTest("MultipleVMVolHaDecrease", "Decrease the replication factor of multiple kubevirt Vms paralley", nil, 0)
	})
	var appCtxs []*scheduler.Context
	var namespace string
	var wg sync.WaitGroup

	itLog := "Decrease the replication factor of multiple kubevirt Vms paralley"
	It(itLog, func() {
		pxNs, err := Inst().V.GetVolumeDriverNamespace()
		log.FailOnError(err, "Failed to get volume driver namespace")
		defer ListEvents(pxNs)

		namespace = fmt.Sprintf("kubevirt-%v", time.Now().Unix())
		log.InfoD(stepLog)
		appList := Inst().AppList
		defer func() {
			Inst().AppList = appList
		}()

		Inst().AppList = []string{"kubevirt-debian-fio-minimal"}
		stepLog := "schedule a kubevirt VM"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				namespace = fmt.Sprintf("kubevirt-%v", time.Now().Unix())
				appCtxs = append(appCtxs, ScheduleApplicationsOnNamespace(namespace, "test")...)
			}
		})
		ValidateApplications(appCtxs)
		for _, appCtx := range appCtxs {
			bindMount, err := IsVMBindMounted(appCtx, false)
			log.FailOnError(err, "Failed to verify bind mount")
			dash.VerifyFatal(bindMount, true, "Failed to verify bind mount")
		}

		stepLog = "Decrease the volume HA of the multiple kubevirt VM Volumes and check if they are bind mounted"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for _, appCtx := range appCtxs {
				wg.Add(1)
				go func(appCtx *scheduler.Context) {
					defer GinkgoRecover()
					defer wg.Done()
					vols, err := Inst().S.GetVolumes(appCtx)
					log.FailOnError(err, "Failed to get volumes of kubevirt VM")
					for _, vol := range vols {
						currRep, err := Inst().V.GetReplicationFactor(vol)
						log.FailOnError(err, "Failed to get Repl factor for vil %s", vol.Name)

						if currRep > 1 {
							opts := volume.Options{
								ValidateReplicationUpdateTimeout: validateReplicationUpdateTimeout,
							}
							err = Inst().V.SetReplicationFactor(vol, currRep-1, nil, nil, true, opts)
							dash.VerifyFatal(err, nil, fmt.Sprintf("Validate set repl factor to %d", currRep-1))
						} else {
							log.Warnf("Volume %s has reached maximum replication factor", vol.Name)
						}
					}
					isVmBindMounted, err := IsVMBindMounted(appCtx, true)
					log.FailOnError(err, "Failed to run vm bind mount check")
					if !isVmBindMounted {
						log.Errorf("The newly added disk to vm %s is not bind mounted", appCtx.App.Key)
					}
				}(appCtx)
			}
			wg.Wait()
		})

		stepLog = "Destroy Applications"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			DestroyApps(appCtxs, nil)
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(appCtxs)
	})
})

var _ = Describe("{LiveMigrateWhileNodeInMaintenance}", Label("p2", "negative", "kubevirt", "error_injection", "LiveMigration", "NodeMaintenance"), func() {
	/*
		            1. Put replica nodes in maintenance mode
			    2. Initiate Live Migration of VM
			    3. Verify VM is migrated to different node
			    4. Exit maintenance mode of the node

	*/

	JustBeforeEach(func() {
		StartTorpedoTest("LiveMigrateWhileNodeInMaintenance", "Live Migrate VM while node is in maintenance mode", nil, 0)

	})

	var appCtxs []*scheduler.Context
	var wg sync.WaitGroup

	itLog := "Live Migrate VM while replica nodes are in maintenance mode"
	It(itLog, func() {
		log.InfoD(itLog)
		pxNs, err := Inst().V.GetVolumeDriverNamespace()
		log.FailOnError(err, "Failed to get volume driver namespace")
		defer ListEvents(pxNs)

		namespace := fmt.Sprintf("kubevirt-%v", time.Now().Unix())
		log.InfoD(stepLog)
		appList := Inst().AppList
		defer func() {
			Inst().AppList = appList
		}()
		Inst().AppList = []string{"kubevirt-debian-fio-minimal"}
		stepLog := "schedule a kubevirt VM"
		Step(stepLog, func() {
			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				taskName := fmt.Sprintf("test-%v", i)
				appCtxs = append(appCtxs, ScheduleApplicationsOnNamespace(namespace, taskName)...)
			}
			ValidateApplications(appCtxs)
			for _, appCtx := range appCtxs {
				bindMount, err := IsVMBindMounted(appCtx, false)
				log.FailOnError(err, "Failed to verify bind mount")
				dash.VerifyFatal(bindMount, true, "Failed to verify bind mount")
			}
		})

		stepLog = "Put replica nodes in maintenance mode and Live migrate the kubevirt VM"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for _, appCtx := range appCtxs {
				ReplicaNodes, err := GetReplicaNodesOfVM(appCtx)
				log.FailOnError(err, "Failed to get non replica nodes of VM")

				vms, err := GetAllVMsFromScheduledContexts([]*scheduler.Context{appCtx})

				for _, vm := range vms {
					nodeVMProvisionedOn, err := GetNodeOfVM(vm)
					log.InfoD("Node VM provisioned on: %s", nodeVMProvisionedOn)
					defer func() {
						var wg sync.WaitGroup
						for _, ReplicaNode := range ReplicaNodes {
							if nodeVMProvisionedOn != ReplicaNode {
								wg.Add(1)
								go func(nonReplicaNode string) {
									defer wg.Done()
									n, err := node.GetNodeByName(nonReplicaNode)
									err = Inst().V.ExitMaintenance(n)
									log.FailOnError(err, "Failed to exit node: %s from maintenance mode", nonReplicaNode)
									log.Infof("Succesfully exited node: %s from maintenance mode", nonReplicaNode)
								}(ReplicaNode)
							}
						}
						wg.Wait()
					}()

					for _, ReplicaNode := range ReplicaNodes {
						if nodeVMProvisionedOn != ReplicaNode {
							wg.Add(1)
							go func(nonReplicaNode string) {
								defer wg.Done()
								n, err := node.GetNodeByName(nonReplicaNode)
								err = Inst().V.EnterMaintenance(n)
								log.FailOnError(err, "Failed to put node: %s in maintenance mode", nonReplicaNode)
								log.Infof("Succesfully put node: %s in maintenance mode", nonReplicaNode)
							}(ReplicaNode)
						}
					}
					wg.Wait()
					err = StartAndWaitForVMIMigration(appCtx, context1.TODO())
					log.FailOnError(err, "Failed to live migrate kubevirt VM")
				}
			}
		})

		stepLog = "Destroy Applications"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			DestroyApps(appCtxs, nil)
		})
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(appCtxs)
	})
})

var _ = Describe("{LiveMigrateCordonNonReplicaNode}", Label("p2", "negative", "kubevirt", "error_injection", "LiveMigration", "RecycleNode"), func() {

	/*
		                1. Schedule a kubevirt VM
				2. Cordon the non replica nodes
				3. Put the replica nodes on maintenance mode
			        4. initiate Live Migration of VM
			        5. Verify VM is migrated to different node
			        6. Exit maintenance mode
				7. Put the replica node on maintenance mode and again initiate live migration of VM
				8. Verify VM is migrated to different node
				9. Uncordon the nodes

	*/

	JustBeforeEach(func() {
		StartTorpedoTest("LiveMigrateCordonNonReplicaNode", "Live Migrate VM while node is in maintenance mode", nil, 0)
	})

	var appCtxs []*scheduler.Context
	var wg sync.WaitGroup

	itLog := "Live Migrate VM while node is in maintenance mode"
	It(itLog, func() {
		log.InfoD(itLog)
		namespace := fmt.Sprintf("kubevirt-%v", time.Now().Unix())
		log.InfoD(stepLog)
		appList := Inst().AppList
		defer func() {
			Inst().AppList = appList
		}()
		Inst().AppList = []string{"kubevirt-debian-fio-minimal"}
		stepLog := "schedule a kubevirt VM"
		Step(stepLog, func() {
			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				taskName := fmt.Sprintf("test-%v", i)
				appCtxs = append(appCtxs, ScheduleApplicationsOnNamespace(namespace, taskName)...)
			}
		})
		stepLog = "Check if vm is bind mount"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for _, appCtx := range appCtxs {
				bindMount, err := IsVMBindMounted(appCtx, false)
				log.FailOnError(err, "Failed to verify bind mount")
				dash.VerifyFatal(bindMount, true, "Failed to verify bind mount")
			}
		})

		for _, appCtx := range appCtxs {
			nonReplicaNodes, err := GetNonReplicaNodesOfVM(appCtx)
			log.FailOnError(err, "Failed to get non replica nodes of VM")

			defer func() {
				//Uncordon the nodes
				for _, nonReplicaNode := range nonReplicaNodes {
					err = core.Instance().UnCordonNode(nonReplicaNode, defaultCommandTimeout, defaultCommandRetry)
					log.FailOnError(err, "Failed to uncordon the node")
				}
			}()
			// cordon the non replica nodes
			for _, nonReplicaNode := range nonReplicaNodes {
				err = core.Instance().CordonNode(nonReplicaNode, defaultCommandTimeout, defaultCommandRetry)
				log.FailOnError(err, "Failed to cordon the node")
			}

			vms, err := GetAllVMsFromScheduledContexts([]*scheduler.Context{appCtx})
			log.FailOnError(err, "Failed to get VMs from scheduled contexts")
			dash.VerifyFatal(len(vms) > 0, true, "Failed to get VMs from scheduled contexts")

			for _, vm := range vms {
				//Put the node on maintenance mode where VM is provisioned
				stepLog = "Put the replica nodes on maintenance mode and live migrate the VM"
				Step(stepLog, func() {
					log.InfoD(stepLog)
					nodeVMProvisionedOn, err := GetNodeOfVM(vm)
					log.InfoD("Node VM provisioned on: %s", nodeVMProvisionedOn)
					defer func() {
						var wg sync.WaitGroup
						for _, nonReplicaNode := range nonReplicaNodes {
							if nodeVMProvisionedOn != nonReplicaNode {
								wg.Add(1)
								go func(nonReplicaNode string) {
									defer wg.Done()
									n, err := node.GetNodeByName(nonReplicaNode)
									err = Inst().V.ExitMaintenance(n)
									log.FailOnError(err, "Failed to exit node: %s from maintenance mode", nonReplicaNode)
									log.Infof("Succesfully exited node: %s from maintenance mode", nonReplicaNode)
								}(nonReplicaNode)
							}
						}
						wg.Wait()
					}()

					for _, nonReplicaNode := range nonReplicaNodes {
						if nodeVMProvisionedOn != nonReplicaNode {
							wg.Add(1)
							go func(nonReplicaNode string) {
								defer wg.Done()
								n, err := node.GetNodeByName(nonReplicaNode)
								err = Inst().V.EnterMaintenance(n)
								log.FailOnError(err, "Failed to put node: %s in maintenance mode", nonReplicaNode)
								log.Infof("Succesfully put node: %s in maintenance mode", nonReplicaNode)
							}(nonReplicaNode)
						}
					}
					wg.Wait()
					err = StartAndWaitForVMIMigration(appCtx, context1.TODO())
					log.FailOnError(err, "Failed to live migrate kubevirt VM")
				})

				//Put the node on maintenance mode and again initiate live migration of VM
				stepLog = "Put the replica nodes on maintenance mode and again initiate live migration of VM"
				Step(stepLog, func() {
					log.InfoD(stepLog)
					nodeVMProvisionedOn, err := GetNodeOfVM(vm)
					log.InfoD("Node VM provisioned on: %s", nodeVMProvisionedOn)
					defer func() {
						var wg sync.WaitGroup
						for _, nonReplicaNode := range nonReplicaNodes {
							if nodeVMProvisionedOn != nonReplicaNode {
								wg.Add(1)
								go func(nonReplicaNode string) {
									defer wg.Done()
									n, err := node.GetNodeByName(nonReplicaNode)
									err = Inst().V.ExitMaintenance(n)
									log.FailOnError(err, "Failed to exit node: %s from maintenance mode", nonReplicaNode)
									log.Infof("Succesfully exited node: %s from maintenance mode", nonReplicaNode)
								}(nonReplicaNode)
							}
						}
						wg.Wait()
					}()

					for _, nonReplicaNode := range nonReplicaNodes {
						if nodeVMProvisionedOn != nonReplicaNode {
							wg.Add(1)
							go func(nonReplicaNode string) {
								defer wg.Done()
								n, err := node.GetNodeByName(nonReplicaNode)
								err = Inst().V.EnterMaintenance(n)
								log.FailOnError(err, "Failed to put node: %s in maintenance mode", nonReplicaNode)
								log.Infof("Succesfully put node: %s in maintenance mode", nonReplicaNode)
							}(nonReplicaNode)
						}
					}
					wg.Wait()
					err = StartAndWaitForVMIMigration(appCtx, context1.TODO())
					log.FailOnError(err, "Failed to live migrate kubevirt VM")
				})
			}
		}
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(appCtxs)
	})

})

var _ = Describe("{StopPxOnNodeWhereVMIsProvisioned}", Label("p1", "negative", "kubevirt", "error_injection", "px_crash"), func() {
	/*
			1. Schedule a kubevirt VM
		`	2. Stop PX on the node where VM is provisioned for 15mins
		        3. Start PX on the node where VM is provisioned
		        4. Verify VM is running fine

			https://portworx.testrail.net/index.php?/cases/view/296893

	*/

	JustBeforeEach(func() {
		StartTorpedoTest("StopPxOnNodeWhereVMIsProvisioned", "Stop PX on node where VM is provisioned", nil, 296893)
	})

	var appCtxs []*scheduler.Context

	itLog := "Stop PX on node where VM is provisioned"
	It(itLog, func() {
		log.InfoD(itLog)
		pxNs, err := Inst().V.GetVolumeDriverNamespace()
		log.FailOnError(err, "Failed to get volume driver namespace")
		defer ListEvents(pxNs)

		appList := Inst().AppList
		defer func() {
			Inst().AppList = appList
		}()
		Inst().AppList = []string{"kubevirt-debian-fio-minimal"}

		stepLog := "schedule a kubevirt VM"
		Step(stepLog, func() {
			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				test := fmt.Sprintf("test-%v", time.Now().Unix())
				appCtxs = append(appCtxs, ScheduleApplications(test)...)
			}
			ValidateApplications(appCtxs)
			for _, appCtx := range appCtxs {
				bindMount, err := IsVMBindMounted(appCtx, false)
				log.FailOnError(err, "Failed to verify bind mount")
				dash.VerifyFatal(bindMount, true, "Failed to verify bind mount")
			}
		})

		for _, appCtx := range appCtxs {
			vms, err := GetAllVMsFromScheduledContexts([]*scheduler.Context{appCtx})
			log.FailOnError(err, "Failed to get VMs from scheduled contexts")
			dash.VerifyFatal(len(vms) > 0, true, "Failed to get VMs from scheduled contexts")
			for _, vm := range vms {
				nodeName, err := GetNodeOfVM(vm)
				log.FailOnError(err, "Failed to get node name for VM: %s", vm.Name)
				nodeObj, err := node.GetNodeByName(nodeName)
				log.FailOnError(err, "Failed to get node obj for node name: %s", nodeName)
				stepLog = "Stop PX on the node where VM is provisioned for 15 mins"
				Step(stepLog, func() {
					log.InfoD(stepLog)
					err := Inst().V.StopDriver([]node.Node{nodeObj}, false, nil)
					log.FailOnError(err, "Failed to stop PX on the node: %s", nodeObj.Name)
					log.Infof("Succesfully stopped PX on the node: %s", nodeObj.Name)
					time.Sleep(15 * time.Minute)
				})

				stepLog = "Start PX on the node where VM is provisioned"
				Step(stepLog, func() {
					log.InfoD(stepLog)
					err := Inst().V.StartDriver(nodeObj)
					log.FailOnError(err, "Failed to start PX on the node: %s", nodeObj.Name)
					log.Infof("Succesfully started PX on the node: %s", nodeObj.Name)
				})
			}
		}

		stepLog = "Validate vm after stopping and starting px"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			ValidateApplications(appCtxs)
			for _, appCtx := range appCtxs {
				bindMount, err := IsVMBindMounted(appCtx, false)
				log.FailOnError(err, "Failed to verify bind mount")
				dash.VerifyFatal(bindMount, true, "Failed to verify bind mount")
			}
		})

	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(appCtxs)
	})
})

var _ = Describe("{RestartPXAndCheckIfVmBindMount}", Label("p1", "negative", "error_injection", "kubevirt", "px_restart"), func() {

	/*
			1. Schedule a kubevirt VM
			2. Cordon the non replica nodes
		        3. initiate Live Migration of VM
		        4. Verify VM is migrated to different node
			5. Restart portworx on the node where VM was provisioned
			6. The volume now should be locally attached to the node where the vm has been migrated
			7. Uncordon the nodes
	*/

	JustBeforeEach(func() {
		StartTorpedoTest("RestartPXAndCheckIfVmBindMount", "Restart PX and check if VM is bind mounted", nil, 0)

	})

	var appCtxs []*scheduler.Context

	itLog := "Live migrate,Restart PX and check if VM is bind mounted"
	It(itLog, func() {
		log.InfoD(itLog)
		namespace := fmt.Sprintf("kubevirt-%v", time.Now().Unix())
		log.InfoD(stepLog)
		appList := Inst().AppList
		defer func() {
			Inst().AppList = appList
		}()
		Inst().AppList = []string{"kubevirt-debian-fio-minimal"}
		stepLog := "schedule a kubevirt VM"
		Step(stepLog, func() {
			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				taskName := fmt.Sprintf("test-%v", i)
				appCtxs = append(appCtxs, ScheduleApplicationsOnNamespace(namespace, taskName)...)
			}
		})
		stepLog = "Check if vm is bind mount"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for _, appCtx := range appCtxs {
				bindMount, err := IsVMBindMounted(appCtx, false)
				log.FailOnError(err, "Failed to verify bind mount")
				dash.VerifyFatal(bindMount, true, "Failed to verify bind mount")
			}
		})

		for _, appCtx := range appCtxs {
			nonReplicaNodes, err := GetNonReplicaNodesOfVM(appCtx)
			log.FailOnError(err, "Failed to get non replica nodes of VM")

			defer func() {
				//Uncordon the nodes
				for _, nonReplicaNode := range nonReplicaNodes {
					err = core.Instance().UnCordonNode(nonReplicaNode, defaultCommandTimeout, defaultCommandRetry)
					log.FailOnError(err, "Failed to uncordon the node")
				}
			}()
			// cordon the non replica nodes
			for _, nonReplicaNode := range nonReplicaNodes {
				err = core.Instance().CordonNode(nonReplicaNode, defaultCommandTimeout, defaultCommandRetry)
				log.FailOnError(err, "Failed to cordon the node")
			}

			vms, err := GetAllVMsFromScheduledContexts([]*scheduler.Context{appCtx})
			log.FailOnError(err, "Failed to get VMs from scheduled contexts")
			dash.VerifyFatal(len(vms) > 0, true, "Failed to get VMs from scheduled contexts")

			for _, vm := range vms {
				nodeVMProvisionedOn, err := GetNodeOfVM(vm)
				log.FailOnError(err, "Failed to get node name for VM: %s", vm.Name)
				log.InfoD("Node VM provisioned on: %s", nodeVMProvisionedOn)

				nodeObj, err := node.GetNodeByName(nodeVMProvisionedOn)
				log.FailOnError(err, "Failed to get node obj for node name: %s", nodeVMProvisionedOn)

				stepLog = "live migrate the VM"
				Step(stepLog, func() {
					log.InfoD(stepLog)
					err = StartAndWaitForVMIMigration(appCtx, context1.TODO())
					log.FailOnError(err, "Failed to live migrate kubevirt VM")
				})

				stepLog = "Restart the PX on the node where VM was provisioned"
				Step(stepLog, func() {
					log.InfoD(stepLog)
					StopVolDriverAndWait([]node.Node{nodeObj})
					StartVolDriverAndWait([]node.Node{nodeObj})
					log.InfoD("Succesfully restarted PX on the node: %s", nodeObj.Name)
				})

				stepLog = "After restart of the px on previously provisioned node the volume should be locally attached to the node where the vm has been migrated"
				Step(stepLog, func() {
					log.InfoD(stepLog)
					isVmBindMounted, err := IsVMBindMounted(appCtx, true)
					log.FailOnError(err, "Failed to run vm bind mount check")
					dash.VerifyFatal(isVmBindMounted, true, "Failed to verify bind mount?")
				})
			}
			stepLog = "Destroy the applications"
			Step(stepLog, func() {
				log.InfoD(stepLog)
				DestroyApps([]*scheduler.Context{appCtx}, nil)
			})
		}
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(appCtxs)
	})
})

var _ = Describe("{FillVMRootDisk}", Label("p2", "negative", "error_injection", "kubevirt", "Throttling"), func() {
	/*
			https://purestorage.atlassian.net/browse/PTX-24760
		        https://portworx.testrail.net/index.php?/cases/view/296895
			1. Schedule a kubevirt VM
			2. Fill the root disk of the VM
			3. Verify the VM does not get effected
		        4. Perform live migration of the VM and verify the VM is migrated successfully
			5. Restart VM and verify the VM is running fine
	*/

	JustBeforeEach(func() {
		StartTorpedoTest("FillVMRootDisk", "Fill the root disk of the VM and verify the VM is running fine", nil, 296895)
	})

	var appCtxs []*scheduler.Context

	itLog := "Fill the root disk of the VM and verify the VM is running fine"
	It(itLog, func() {
		stepLog := "schedule a kubevirt VM"
		namespace := fmt.Sprintf("kubevirt-%v", time.Now().Unix())

		appList := Inst().AppList
		Inst().AppList = []string{"kubevirt-debian-fio-minimal"}
		defer func() {
			Inst().AppList = appList
		}()

		Step(stepLog, func() {
			log.InfoD(stepLog)
			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				taskName := fmt.Sprintf("test-%v", i)
				appCtxs = append(appCtxs, ScheduleApplicationsOnNamespace(namespace, taskName)...)
			}
			ValidateApplications(appCtxs)
			for _, appCtx := range appCtxs {
				bindMount, err := IsVMBindMounted(appCtx, false)
				log.FailOnError(err, "Failed to verify bind mount")
				dash.VerifyFatal(bindMount, true, "Failed to verify bind mount")
			}

		})

		stepLog = "Fill the root disk of the VM"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for _, appCtx := range appCtxs {
				err := FillRootDiskOfVM(appCtx)
				log.FailOnError(err, "Failed to fill the root disk of the VM")
			}
		})

		stepLog = "Verify the VM is running fine"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			ValidateApplications(appCtxs)
			for _, appCtx := range appCtxs {
				bindMount, err := IsVMBindMounted(appCtx, false)
				log.FailOnError(err, "Failed to verify bind mount")
				dash.VerifyFatal(bindMount, true, "Failed to verify bind mount")
			}
		})

		stepLog = "Perform live migration of the VM and verify the VM is migrated successfully"
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

func CreateSSHPodAndSetCanSsh() bool {
	stepLog := "Create SSH Pod"
	var canSsh bool = false
	Step(stepLog, func() {
		log.InfoD(stepLog)
		err := CreateSSHPod()
		if err == nil {
			canSsh = true
		} else {
			canSsh = false
		}
	})
	return canSsh
}

func ValidateFioInVMs(appCtxs []*scheduler.Context, canSsh bool) {
	if canSsh {
		stepLog := "Validate fio is running in the VMs"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			var wg sync.WaitGroup
			for _, appCtx := range appCtxs {
				wg.Add(1)
				go func(appCtx *scheduler.Context) {
					defer GinkgoRecover()
					defer wg.Done()
					vms, err := GetAllVMsFromScheduledContexts([]*scheduler.Context{appCtx})
					log.FailOnError(err, "Failed to get VMs from appCtx")
					for _, vm := range vms {
						err = CheckFioIsRunningInVM(vm)
						log.FailOnError(err, "Failed to validate fio in VM %s", vm.Name)
					}
				}(appCtx)
			}
			wg.Wait()
		})
	}
}

func ValidateVMUptime(appCtxs []*scheduler.Context, canSsh bool, initialUptime map[string]time.Duration) {
	if canSsh {
		stepLog := "Validate VMs have not restarted"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			var wg sync.WaitGroup
			for _, appCtx := range appCtxs {
				wg.Add(1)
				go func(appCtx *scheduler.Context) {
					defer GinkgoRecover()
					defer wg.Done()
					vms, err := GetAllVMsFromScheduledContexts([]*scheduler.Context{appCtx})
					log.FailOnError(err, "Failed to get VMs from appCtx")
					for _, vm := range vms {
						err = CheckVMUptime(vm, initialUptime)
						log.FailOnError(err, "Failed to validate uptime in VM %s", vm.Name)
					}
				}(appCtx)
			}
			wg.Wait()
		})
	}
}

func GetDynamicKubeClient() (dynamic.Interface, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	return dynamic.NewForConfig(config)
}

var _ = Describe("{SingleVMLiveMigration}", Label("p0", "positive", "kubevirt", "MiniScale", "LiveMigration"), func() {
	var app, volType string
	var present bool
	JustBeforeEach(func() {
		StartTorpedoTest("SingleVMLiveMigration", "Live migrate single kubevirt VM", nil, 0)
		volType, present = os.LookupEnv("KUBEVIRT_VOL_TYPE")
		if !present {
			app = "kubevirt-debian-fio-minimal"
		}
		if volType == "pxe-raw" {
			app = "kubevirt-raw-vol"
		} else if volType == "fada-raw" {
			app = "kubevirt-fada-raw-fio"
		}
		log.InfoD("Setting app for this test to be : %s", app)
	})
	var appCtxs []*scheduler.Context
	var namespace string
	var wg sync.WaitGroup
	var canSsh bool
	var initialUptime map[string]time.Duration
	var vm kubevirtv1.VirtualMachine
	var vmNodeName string

	itLog := "Live migrate single kubevirt VM"
	It(itLog, func() {
		pxNs, err := Inst().V.GetVolumeDriverNamespace()
		log.FailOnError(err, "Failed to get volume driver namespace")
		defer ListEvents(pxNs)
		canSsh = false
		stepLog := "Schedule a KubeVirt VM"
		log.InfoD(stepLog)
		appList := Inst().AppList
		defer func() {
			Inst().AppList = appList
		}()
		Inst().AppList = []string{app}
		Inst().CsiAppList = []string{app}

		stepLog = "Schedule a kubevirt VM"
		Step(stepLog, func() {
			namespace = fmt.Sprintf("kubevirt-%v", time.Now().Unix())
			appCtxs = append(appCtxs, ScheduleApplicationsOnNamespace(namespace, "test")...)
		})
		ValidateApplications(appCtxs)
		if !present {
			for _, appCtx := range appCtxs {
				bindMount, err := IsVMBindMounted(appCtx, false)
				log.FailOnError(err, "Failed to verify bind mount")
				dash.VerifyFatal(bindMount, true, "Failed to verify bind mount")
			}
		}
		log.Infof("Hard Sleep for 2 minutes to let VMs come up")
		time.Sleep(2 * time.Minute)

		canSsh = CreateSSHPodAndSetCanSsh()
		ValidateFioInVMs(appCtxs, canSsh)

		initialUptime = make(map[string]time.Duration)
		stepLog = "Get initial uptime of VMs and current node"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			var wg sync.WaitGroup
			for _, appCtx := range appCtxs {
				wg.Add(1)
				go func(appCtx *scheduler.Context) {
					defer GinkgoRecover()
					defer wg.Done()
					vms, err := GetAllVMsFromScheduledContexts([]*scheduler.Context{appCtx})
					log.FailOnError(err, "Failed to get VMs from appCtx")
					for _, vm := range vms {
						uptime, err := GetVMUptime(vm)
						log.FailOnError(err, "Failed to get uptime from VM %s", vm.Name)
						vmKey := fmt.Sprintf("%s/%s", vm.Namespace, vm.Name)
						initialUptime[vmKey] = uptime
						log.Infof("Initial uptime for VM %s is %v", vmKey, uptime)

						vmNodeName, err = GetNodeOfVM(vm)
						log.FailOnError(err, "Failed to get node of VM %v", vm.Name)
						log.Infof("VM %s is currently running on node %s", vm.Name, vmNodeName)
					}
				}(appCtx)
			}
			wg.Wait()
		})

		vms, err := GetAllVMsFromScheduledContexts(appCtxs)
		log.FailOnError(err, "Failed to get VMs from context")

		if len(vms) == 0 {
			log.FailOnError(fmt.Errorf("No VMs found"), "No VMs found in context")
		} else {
			vm = vms[0]
		}

		stepLog = "Live migrate the kubevirt VM"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for _, appCtx := range appCtxs {
				wg.Add(1)
				go func(appCtx *scheduler.Context) {
					defer GinkgoRecover()
					defer wg.Done()
					err := StartAndWaitForVMIMigration(appCtx, context1.TODO())
					log.FailOnError(err, "Failed to live migrate kubevirt VM")
				}(appCtx)
			}
		})
		wg.Wait()

		stepLog = "Get VM node after migration"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			newNodeName, err := GetNodeOfVM(vm)
			log.FailOnError(err, "Failed to get node of VM %v after migration", vm.Name)
			log.Infof("VM %s is now running on node %s after migration", vm.Name, newNodeName)
			if newNodeName == vmNodeName {
				log.FailOnError(fmt.Errorf("VM did not migrate to a different node"), "VM is still on node %s after migration", vmNodeName)
			} else {
				log.Infof("VM successfully migrated from node %s to node %s", vmNodeName, newNodeName)
			}
		})

		ValidateVMUptime(appCtxs, canSsh, initialUptime)

		ValidateFioInVMs(appCtxs, canSsh)

		stepLog = "Destroy Applications"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			DestroyApps(appCtxs, nil)
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(appCtxs)
	})
})

var _ = Describe("{MultipleParallelLiveMigration}", Label("p0", "positive", "kubevirt", "MiniScale", "LiveMigration"), func() {
	var app, volType string
	var present bool
	JustBeforeEach(func() {
		StartTorpedoTest("MultipleParallelLiveMigration", "Live migrate multiple kubevirt VMs in parallel", nil, 0)
		volType, present = os.LookupEnv("KUBEVIRT_VOL_TYPE")
		if !present {
			app = "kubevirt-debian-fio-minimal"
		}
		if volType == "pxe-raw" {
			app = "kubevirt-raw-vol"
		} else if volType == "fada-raw" {
			app = "kubevirt-fada-raw-fio"
		}
		log.InfoD("Setting app for this test to be : %s", app)
	})
	var appCtxs []*scheduler.Context
	var namespace string
	var wg sync.WaitGroup
	var canSsh bool
	var initialUptime map[string]time.Duration
	var initialNodeName map[string]string
	var failure bool = false

	itLog := "Live migrate multiple kubevirt VMs in parallel"
	It(itLog, func() {
		pxNs, err := Inst().V.GetVolumeDriverNamespace()
		log.FailOnError(err, "Failed to get volume driver namespace")
		defer ListEvents(pxNs)
		canSsh = false
		stepLog := "Schedule kubevirt VMs"
		log.InfoD(stepLog)
		appList := Inst().AppList
		defer func() {
			Inst().AppList = appList
		}()
		Inst().AppList = []string{app}
		Inst().CsiAppList = []string{app}

		stepLog = "Schedule kubevirt VMs"
		Step(stepLog, func() {
			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				namespace = fmt.Sprintf("kubevirt-%v", time.Now().Unix())
				appCtxs = append(appCtxs, ScheduleApplicationsOnNamespace(namespace, "test")...)
			}
		})
		ValidateApplications(appCtxs)
		if !present {
			for _, appCtx := range appCtxs {
				bindMount, err := IsVMBindMounted(appCtx, false)
				log.FailOnError(err, "Failed to verify bind mount")
				dash.VerifyFatal(bindMount, true, "Failed to verify bind mount")
			}
		}

		log.Infof("Hard Sleep for 2 minutes to let VMs come up")
		time.Sleep(2 * time.Minute)

		canSsh = CreateSSHPodAndSetCanSsh()
		ValidateFioInVMs(appCtxs, canSsh)

		initialUptime = make(map[string]time.Duration)
		initialNodeName = make(map[string]string)
		var mu sync.Mutex
		stepLog = "Get initial uptime of VMs and current nodes"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			var wg sync.WaitGroup
			for _, appCtx := range appCtxs {
				wg.Add(1)
				go func(appCtx *scheduler.Context) {
					defer GinkgoRecover()
					defer wg.Done()
					vms, err := GetAllVMsFromScheduledContexts([]*scheduler.Context{appCtx})
					log.FailOnError(err, "Failed to get VMs from appCtx")
					for _, vm := range vms {
						vmKey := fmt.Sprintf("%s/%s", vm.Namespace, vm.Name)
						uptime, err := GetVMUptime(vm)
						log.FailOnError(err, "Failed to get uptime from VM %s", vm.Name)
						log.Infof("Initial uptime for VM %s is %v", vmKey, uptime)

						nodeName, err := GetNodeOfVM(vm)
						log.FailOnError(err, "Failed to get node of VM %v", vm.Name)
						log.Infof("VM %s is currently running on node %s", vm.Name, nodeName)

						mu.Lock()
						initialUptime[vmKey] = uptime
						initialNodeName[vmKey] = nodeName
						mu.Unlock()
					}
				}(appCtx)
			}
			wg.Wait()
		})

		wg = sync.WaitGroup{}
		stepLog = "Live migrate the kubevirt VMs"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for _, appCtx := range appCtxs {
				wg.Add(1)
				go func(appCtx *scheduler.Context) {
					defer GinkgoRecover()
					defer wg.Done()
					err := StartAndWaitForVMIMigration(appCtx, context1.TODO())
					log.FailOnError(err, "Failed to live migrate kubevirt VM")
				}(appCtx)
			}
			wg.Wait()
		})

		stepLog = "Validate VMs after migration"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			var wg sync.WaitGroup
			for _, appCtx := range appCtxs {
				wg.Add(1)
				go func(appCtx *scheduler.Context) {
					defer GinkgoRecover()
					defer wg.Done()
					vms, err := GetAllVMsFromScheduledContexts([]*scheduler.Context{appCtx})
					log.FailOnError(err, "Failed to get VMs from appCtx")
					for _, vm := range vms {
						vmKey := fmt.Sprintf("%s/%s", vm.Namespace, vm.Name)
						newNodeName, err := GetNodeOfVM(vm)
						log.FailOnError(err, "Failed to get node of VM %s after migration", vm.Name)
						log.Infof("VM %s is now running on node %s after migration", vm.Name, newNodeName)
						mu.Lock()
						initialNode := initialNodeName[vmKey]
						mu.Unlock()
						if newNodeName == initialNode {
							failure = true
							log.FailOnError(fmt.Errorf("VM %s did not migrate to a different node", vm.Name), "VM is still on node %s after migration", initialNode)
						} else {
							log.Infof("VM %s successfully migrated from node %s to node %s", vm.Name, initialNode, newNodeName)
						}

						err = CheckVMUptime(vm, initialUptime)
						log.FailOnError(err, "Failed to validate uptime in VM %s", vm.Name)
					}
				}(appCtx)
			}
			wg.Wait()
		})

		ValidateFioInVMs(appCtxs, canSsh)

		stepLog = "Destroy Applications"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			if !failure {
				DestroyApps(appCtxs, nil)
			}
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(appCtxs)
	})
})

var _ = Describe("{LiveMigrationsOfVMsInALoop}", Label("p0", "positive", "kubevirt", "MiniScale", "LiveMigration"), func() {
	var app, volType string
	var present bool
	JustBeforeEach(func() {
		StartTorpedoTest("LiveMigrationsOfVMsInALoop", "Live migrate kubevirt VMs multiple times", nil, 0)
		volType, present = os.LookupEnv("KUBEVIRT_VOL_TYPE")
		if !present {
			app = "kubevirt-debian-fio-minimal"
		}
		if volType == "pxe-raw" {
			app = "kubevirt-raw-vol"
		} else if volType == "fada-raw" {
			app = "kubevirt-fada-raw-fio"
		}
		log.InfoD("Setting app for this test to be : %s", app)
	})
	var appCtxs []*scheduler.Context
	var namespace string
	var canSsh bool
	var numMigrations int = 10
	var initialUptime map[string]time.Duration
	var initialNodeName map[string]string

	itLog := "Live migrate kubevirt VMs multiple times"
	It(itLog, func() {
		pxNs, err := Inst().V.GetVolumeDriverNamespace()
		log.FailOnError(err, "Failed to get volume driver namespace")
		defer ListEvents(pxNs)
		canSsh = false
		appList := Inst().AppList
		defer func() {
			Inst().AppList = appList
		}()
		Inst().AppList = []string{app}
		Inst().CsiAppList = []string{app}

		stepLog := "Schedule kubevirt VMs"
		Step(stepLog, func() {
			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				namespace = fmt.Sprintf("kubevirt-%v", time.Now().Unix())
				appCtxs = append(appCtxs, ScheduleApplicationsOnNamespace(namespace, "test")...)
			}
		})
		ValidateApplications(appCtxs)
		if !present {
			for _, appCtx := range appCtxs {
				bindMount, err := IsVMBindMounted(appCtx, false)
				log.FailOnError(err, "Failed to verify bind mount")
				dash.VerifyFatal(bindMount, true, "Failed to verify bind mount")
			}
		}

		log.Infof("Sleeping for 2 minutes to let VMs come up")
		time.Sleep(2 * time.Minute)

		canSsh = CreateSSHPodAndSetCanSsh()
		ValidateFioInVMs(appCtxs, canSsh)

		initialUptime = make(map[string]time.Duration)
		initialNodeName = make(map[string]string)
		var mu sync.Mutex
		stepLog = "Get initial uptime of VMs and current nodes"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			var wg sync.WaitGroup
			for _, appCtx := range appCtxs {
				wg.Add(1)
				go func(appCtx *scheduler.Context) {
					defer GinkgoRecover()
					defer wg.Done()
					vms, err := GetAllVMsFromScheduledContexts([]*scheduler.Context{appCtx})
					log.FailOnError(err, "Failed to get VMs from appCtx")
					for _, vm := range vms {
						vmKey := fmt.Sprintf("%s/%s", vm.Namespace, vm.Name)
						uptime, err := GetVMUptime(vm)
						log.FailOnError(err, "Failed to get uptime from VM %s", vm.Name)
						log.Infof("Initial uptime for VM %s is %v", vmKey, uptime)

						nodeName, err := GetNodeOfVM(vm)
						log.FailOnError(err, "Failed to get node of VM %v", vm.Name)
						log.Infof("VM %s is currently running on node %s", vm.Name, nodeName)

						mu.Lock()
						initialUptime[vmKey] = uptime
						initialNodeName[vmKey] = nodeName
						mu.Unlock()
					}
				}(appCtx)
			}
			wg.Wait()
		})

		for migrationCount := 1; migrationCount <= numMigrations; migrationCount++ {
			stepLog = fmt.Sprintf("Live migrate the kubevirt VMs - iteration %d", migrationCount)
			Step(stepLog, func() {
				log.InfoD(stepLog)
				var wg sync.WaitGroup
				for _, appCtx := range appCtxs {
					wg.Add(1)
					go func(appCtx *scheduler.Context) {
						defer GinkgoRecover()
						defer wg.Done()
						err := StartAndWaitForVMIMigration(appCtx, context1.TODO())
						log.FailOnError(err, "Failed to live migrate kubevirt VM")
					}(appCtx)
				}
				wg.Wait()

				stepLog := fmt.Sprintf("Validate VMs after migration - iteration %d", migrationCount)
				Step(stepLog, func() {
					log.InfoD(stepLog)
					var wg sync.WaitGroup
					for _, appCtx := range appCtxs {
						wg.Add(1)
						go func(appCtx *scheduler.Context) {
							defer GinkgoRecover()
							defer wg.Done()
							vms, err := GetAllVMsFromScheduledContexts([]*scheduler.Context{appCtx})
							log.FailOnError(err, "Failed to get VMs from appCtx")
							for _, vm := range vms {
								vmKey := fmt.Sprintf("%s/%s", vm.Namespace, vm.Name)
								newNodeName, err := GetNodeOfVM(vm)
								log.FailOnError(err, "Failed to get node of VM %s after migration", vm.Name)
								log.Infof("VM %s is now running on node %s after migration", vm.Name, newNodeName)
								mu.Lock()
								initialNode := initialNodeName[vmKey]
								mu.Unlock()
								if newNodeName == initialNode {
									log.FailOnError(fmt.Errorf("VM %s did not migrate to a different node", vm.Name), "VM is still on node %s after migration", initialNode)
								} else {
									log.Infof("VM %s successfully migrated from node %s to node %s", vm.Name, initialNode, newNodeName)
									mu.Lock()
									initialNodeName[vmKey] = newNodeName
									mu.Unlock()
								}

								err = CheckVMUptime(vm, initialUptime)
								log.FailOnError(err, "Failed to validate uptime in VM %s", vm.Name)
							}
						}(appCtx)
					}
					wg.Wait()
				})
				ValidateFioInVMs(appCtxs, canSsh)
			})
		}

		stepLog = "Destroy Applications"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			DestroyApps(appCtxs, nil)
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(appCtxs)
	})
})

var _ = Describe("{AddNewRawDiskToKubevirtVM}", Label("p0", "positive", "kubevirt"), func() {
	var app, volType string
	var present bool
	JustBeforeEach(func() {
		StartTorpedoTest("AddNewRawDiskToKubevirtVM", "Add a new raw disk to a kubevirtVM", nil, 0)
		volType, present = os.LookupEnv("KUBEVIRT_VOL_TYPE")
		if !present {
			app = "kubevirt-debian-fio-minimal"
		}
		if volType == "pxe-raw" {
			app = "kubevirt-raw-vol"
		} else if volType == "fada-raw" {
			app = "kubevirt-fada-raw-fio"
		}
		log.InfoD("Setting app for this test to be : %s", app)
	})
	var appCtxs []*scheduler.Context
	var namespace string
	var canSsh bool

	itLog := "Add a new raw disk to a kubevirtVM"
	It(itLog, func() {
		pxNs, err := Inst().V.GetVolumeDriverNamespace()
		log.FailOnError(err, "Failed to get volume driver namespace")
		defer ListEvents(pxNs)

		appList := Inst().AppList
		defer func() {
			Inst().AppList = appList
		}()
		numberOfVolumes := 1

		Inst().AppList = []string{app}
		Inst().CsiAppList = []string{app}
		stepLog := "Schedule a kubevirt VM"
		Step(stepLog, func() {
			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				namespace = fmt.Sprintf("kubevirt-%v", time.Now().Unix())
				appCtxs = append(appCtxs, ScheduleApplicationsOnNamespace(namespace, "test")...)
			}
		})
		ValidateApplications(appCtxs)
		if !present {
			for _, appCtx := range appCtxs {
				bindMount, err := IsVMBindMounted(appCtx, false)
				log.FailOnError(err, "Failed to verify bind mount")
				dash.VerifyFatal(bindMount, true, "Failed to verify bind mount")
			}
		}

		log.Infof("Sleeping for 2 minutes to let VMs come up")
		time.Sleep(2 * time.Minute)

		canSsh = CreateSSHPodAndSetCanSsh()
		ValidateFioInVMs(appCtxs, canSsh)

		stepLog = "Add one raw disk to the kubevirt VM"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			_, err := AddRawBlockDriveToKubevirtVM(appCtxs, numberOfVolumes, "10Gi")
			log.FailOnError(err, "Failed to add disks to kubevirt VM")
			dash.VerifyFatal(true, true, "Disk added to kubevirt VM")
		})
		ValidateFioInVMs(appCtxs, canSsh)
		stepLog = "Destroy Applications"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			DestroyApps(appCtxs, nil)
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(appCtxs)
	})
})

var _ = Describe("{LMAfterRawDiskAddToVM}", Label("p0", "positive", "kubevirt", "LiveMigration"), func() {
	var app, volType string
	var present bool
	JustBeforeEach(func() {
		StartTorpedoTest("LMAfterRawDiskAddToVM", "Add a raw disk to KubeVirt VM and then live migrate it", nil, 0)
		volType, present = os.LookupEnv("KUBEVIRT_VOL_TYPE")
		if !present {
			app = "kubevirt-debian-fio-minimal"
		}
		if volType == "pxe-raw" {
			app = "kubevirt-raw-vol"
		} else if volType == "fada-raw" {
			app = "kubevirt-fada-raw-fio"
		}
		log.InfoD("Setting app for this test to be : %s", app)
	})
	var appCtxs []*scheduler.Context
	var namespace string
	var wg sync.WaitGroup
	var canSsh bool
	var initialUptime map[string]time.Duration
	var initialNodeName map[string]string

	itLog := "Add a disk to KubeVirt VM and then live migrate it"
	It(itLog, func() {
		pxNs, err := Inst().V.GetVolumeDriverNamespace()
		log.FailOnError(err, "Failed to get volume driver namespace")
		defer ListEvents(pxNs)
		canSsh = false
		appList := Inst().AppList
		defer func() {
			Inst().AppList = appList
		}()
		numberOfVolumes := 1

		Inst().AppList = []string{app}
		Inst().CsiAppList = []string{app}
		stepLog := "Schedule a KubeVirt VM"
		Step(stepLog, func() {
			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				namespace = fmt.Sprintf("kubevirt-%v", time.Now().Unix())
				appCtxs = append(appCtxs, ScheduleApplicationsOnNamespace(namespace, "test")...)
			}
		})
		ValidateApplications(appCtxs)
		if !present {
			for _, appCtx := range appCtxs {
				bindMount, err := IsVMBindMounted(appCtx, false)
				log.FailOnError(err, "Failed to verify bind mount")
				dash.VerifyFatal(bindMount, true, "Failed to verify bind mount")
			}
		}

		log.Infof("Sleeping for 2 minutes to let VMs come up")
		time.Sleep(2 * time.Minute)

		canSsh = CreateSSHPodAndSetCanSsh()
		ValidateFioInVMs(appCtxs, canSsh)

		stepLog = "Add one disk to the KubeVirt VM"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			_, err := AddRawBlockDriveToKubevirtVM(appCtxs, numberOfVolumes, "10Gi")
			log.FailOnError(err, "Failed to add disks to KubeVirt VM")
			dash.VerifyFatal(true, true, "Disk added to KubeVirt VM")
		})

		initialUptime = make(map[string]time.Duration)
		initialNodeName = make(map[string]string)
		var mu sync.Mutex
		stepLog = "Get initial uptime of VMs and current nodes"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			var wg sync.WaitGroup
			for _, appCtx := range appCtxs {
				wg.Add(1)
				go func(appCtx *scheduler.Context) {
					defer GinkgoRecover()
					defer wg.Done()
					vms, err := GetAllVMsFromScheduledContexts([]*scheduler.Context{appCtx})
					log.FailOnError(err, "Failed to get VMs from appCtx")
					for _, vm := range vms {
						vmKey := fmt.Sprintf("%s/%s", vm.Namespace, vm.Name)
						uptime, err := GetVMUptime(vm)
						log.FailOnError(err, "Failed to get uptime from VM %s", vm.Name)
						log.Infof("Initial uptime for VM %s is %v", vmKey, uptime)

						nodeName, err := GetNodeOfVM(vm)
						log.FailOnError(err, "Failed to get node of VM %v", vm.Name)
						log.Infof("VM %s is currently running on node %s", vm.Name, nodeName)

						mu.Lock()
						initialUptime[vmKey] = uptime
						initialNodeName[vmKey] = nodeName
						mu.Unlock()
					}
				}(appCtx)
			}
			wg.Wait()
		})

		stepLog = "Live migrate the KubeVirt VM"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for _, appCtx := range appCtxs {
				wg.Add(1)
				go func(appCtx *scheduler.Context) {
					defer GinkgoRecover()
					defer wg.Done()
					err := StartAndWaitForVMIMigration(appCtx, context1.TODO())
					log.FailOnError(err, "Failed to live migrate KubeVirt VM")
				}(appCtx)
			}
			wg.Wait()
		})

		stepLog = "Validate VMs after migration"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			var wg sync.WaitGroup
			for _, appCtx := range appCtxs {
				wg.Add(1)
				go func(appCtx *scheduler.Context) {
					defer GinkgoRecover()
					defer wg.Done()
					vms, err := GetAllVMsFromScheduledContexts([]*scheduler.Context{appCtx})
					log.FailOnError(err, "Failed to get VMs from appCtx")
					for _, vm := range vms {
						vmKey := fmt.Sprintf("%s/%s", vm.Namespace, vm.Name)
						newNodeName, err := GetNodeOfVM(vm)
						log.FailOnError(err, "Failed to get node of VM %s after migration", vm.Name)
						log.Infof("VM %s is now running on node %s after migration", vm.Name, newNodeName)
						mu.Lock()
						initialNode := initialNodeName[vmKey]
						mu.Unlock()
						if newNodeName == initialNode {
							log.FailOnError(fmt.Errorf("VM %s did not migrate to a different node", vm.Name), "VM is still on node %s after migration", initialNode)
						} else {
							log.Infof("VM %s successfully migrated from node %s to node %s", vm.Name, initialNode, newNodeName)
							mu.Lock()
							initialNodeName[vmKey] = newNodeName
							mu.Unlock()
						}

						err = CheckVMUptime(vm, initialUptime)
						log.FailOnError(err, "Failed to validate uptime in VM %s", vm.Name)
					}
				}(appCtx)
			}
			wg.Wait()
		})

		ValidateFioInVMs(appCtxs, canSsh)

		stepLog = "Destroy Applications"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			DestroyApps(appCtxs, nil)
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(appCtxs)
	})
})

var _ = Describe("{LMBeforeRawDiskAddToVM}", Label("p0", "positive", "kubevirt", "LiveMigration"), func() {
	var app, volType string
	var present bool
	JustBeforeEach(func() {
		StartTorpedoTest("LMBeforeRawDiskAddToVM", "Live migrate KubeVirt VM and then add a disk", nil, 0)
		volType, present = os.LookupEnv("KUBEVIRT_VOL_TYPE")
		if !present {
			app = "kubevirt-debian-fio-minimal"
		}
		if volType == "pxe-raw" {
			app = "kubevirt-raw-vol"
		} else if volType == "fada-raw" {
			app = "kubevirt-fada-raw-fio"
		}
		log.InfoD("Setting app for this test to be : %s", app)
	})
	var appCtxs []*scheduler.Context
	var namespace string
	var canSsh bool
	var initialUptime map[string]time.Duration
	var initialNodeName map[string]string
	var mu sync.Mutex

	itLog := "Live migrate KubeVirt VM and then add a disk"
	It(itLog, func() {
		pxNs, err := Inst().V.GetVolumeDriverNamespace()
		log.FailOnError(err, "Failed to get volume driver namespace")
		defer ListEvents(pxNs)
		canSsh = false
		appList := Inst().AppList
		defer func() {
			Inst().AppList = appList
		}()
		numberOfVolumes := 1

		Inst().AppList = []string{app}
		Inst().CsiAppList = []string{app}
		stepLog := "Schedule a KubeVirt VM"
		Step(stepLog, func() {
			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				namespace = fmt.Sprintf("kubevirt-%v", time.Now().Unix())
				appCtxs = append(appCtxs, ScheduleApplicationsOnNamespace(namespace, "test")...)
			}
		})
		ValidateApplications(appCtxs)
		if !present {
			for _, appCtx := range appCtxs {
				bindMount, err := IsVMBindMounted(appCtx, false)
				log.FailOnError(err, "Failed to verify bind mount")
				dash.VerifyFatal(bindMount, true, "Failed to verify bind mount")
			}
		}
		log.Infof("Sleeping for 2 minutes to let VMs come up")
		time.Sleep(2 * time.Minute)

		canSsh = CreateSSHPodAndSetCanSsh()

		ValidateFioInVMs(appCtxs, canSsh)

		initialUptime = make(map[string]time.Duration)
		initialNodeName = make(map[string]string)
		stepLog = "Get initial uptime of VMs and current nodes"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			var wg sync.WaitGroup
			for _, appCtx := range appCtxs {
				wg.Add(1)
				go func(appCtx *scheduler.Context) {
					defer GinkgoRecover()
					defer wg.Done()
					vms, err := GetAllVMsFromScheduledContexts([]*scheduler.Context{appCtx})
					log.FailOnError(err, "Failed to get VMs from appCtx")
					for _, vm := range vms {
						vmKey := fmt.Sprintf("%s/%s", vm.Namespace, vm.Name)
						uptime, err := GetVMUptime(vm)
						log.FailOnError(err, "Failed to get uptime from VM %s", vm.Name)
						log.Infof("Initial uptime for VM %s is %v", vmKey, uptime)

						nodeName, err := GetNodeOfVM(vm)
						log.FailOnError(err, "Failed to get node of VM %v", vm.Name)
						log.Infof("VM %s is currently running on node %s", vm.Name, nodeName)

						mu.Lock()
						initialUptime[vmKey] = uptime
						initialNodeName[vmKey] = nodeName
						mu.Unlock()
					}
				}(appCtx)
			}
			wg.Wait()
		})

		stepLog = "Live migrate the KubeVirt VM"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			var wg sync.WaitGroup
			for _, appCtx := range appCtxs {
				wg.Add(1)
				go func(appCtx *scheduler.Context) {
					defer GinkgoRecover()
					defer wg.Done()
					err := StartAndWaitForVMIMigration(appCtx, context1.TODO())
					log.FailOnError(err, "Failed to live migrate KubeVirt VM")
				}(appCtx)
			}
			wg.Wait()
		})

		stepLog = "Validate VMs after migration"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			var wg sync.WaitGroup
			for _, appCtx := range appCtxs {
				wg.Add(1)
				go func(appCtx *scheduler.Context) {
					defer GinkgoRecover()
					defer wg.Done()
					vms, err := GetAllVMsFromScheduledContexts([]*scheduler.Context{appCtx})
					log.FailOnError(err, "Failed to get VMs from appCtx")
					for _, vm := range vms {
						vmKey := fmt.Sprintf("%s/%s", vm.Namespace, vm.Name)
						newNodeName, err := GetNodeOfVM(vm)
						log.FailOnError(err, "Failed to get node of VM %s after migration", vm.Name)
						log.Infof("VM %s is now running on node %s after migration", vm.Name, newNodeName)
						mu.Lock()
						initialNode := initialNodeName[vmKey]
						mu.Unlock()
						if newNodeName == initialNode {
							log.FailOnError(fmt.Errorf("VM %s did not migrate to a different node", vm.Name), "VM is still on node %s after migration", initialNode)
						} else {
							log.Infof("VM %s successfully migrated from node %s to node %s", vm.Name, initialNode, newNodeName)
							mu.Lock()
							initialNodeName[vmKey] = newNodeName
							mu.Unlock()
						}

						err = CheckVMUptime(vm, initialUptime)
						log.FailOnError(err, "Failed to validate uptime in VM %s", vm.Name)
					}
				}(appCtx)
			}
			wg.Wait()
		})

		stepLog = "Add one disk to the KubeVirt VM"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			_, err := AddRawBlockDriveToKubevirtVM(appCtxs, numberOfVolumes, "10Gi")
			log.FailOnError(err, "Failed to add disks to KubeVirt VM")
			dash.VerifyFatal(true, true, "Disk added to KubeVirt VM")
		})

		ValidateFioInVMs(appCtxs, canSsh)

		stepLog = "Destroy Applications"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			DestroyApps(appCtxs, nil)
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(appCtxs)
	})
})

var _ = Describe("{LMAndAddRawDiskToVMInALoop}", Label("p0", "positive", "kubevirt", "LiveMigration"), func() {
	var app, volType string
	var present bool
	JustBeforeEach(func() {
		StartTorpedoTest("LMAndAddRawDiskToVMInALoop", "Live migrate and add raw disk to KubeVirt VM multiple times", nil, 0)
		volType, present = os.LookupEnv("KUBEVIRT_VOL_TYPE")
		if !present {
			app = "kubevirt-debian-fio-minimal"
		}
		if volType == "pxe-raw" {
			app = "kubevirt-raw-vol"
		} else if volType == "fada-raw" {
			app = "kubevirt-fada-raw-fio"
		}
		log.InfoD("Setting app for this test to be : %s", app)
	})
	var appCtxs []*scheduler.Context
	var namespace string
	var canSsh bool
	var iterations int = 5
	var numberOfVolumes int = 1
	var initialUptime map[string]time.Duration
	var initialNodeName map[string]string
	var mu sync.Mutex
	var failure bool = false

	itLog := "Live migrate and add disk to KubeVirt VM multiple times"
	It(itLog, func() {
		pxNs, err := Inst().V.GetVolumeDriverNamespace()
		log.FailOnError(err, "Failed to get volume driver namespace")
		defer ListEvents(pxNs)
		canSsh = false
		appList := Inst().AppList
		defer func() {
			Inst().AppList = appList
		}()

		Inst().AppList = []string{app}
		Inst().CsiAppList = []string{app}
		stepLog := "Schedule a KubeVirt VM"
		Step(stepLog, func() {
			namespace = fmt.Sprintf("kubevirt-%v", time.Now().Unix())
			appCtxs = append(appCtxs, ScheduleApplicationsOnNamespace(namespace, "test")...)
		})
		ValidateApplications(appCtxs)
		if !present {
			for _, appCtx := range appCtxs {
				bindMount, err := IsVMBindMounted(appCtx, false)
				log.FailOnError(err, "Failed to verify bind mount")
				dash.VerifyFatal(bindMount, true, "Failed to verify bind mount")
			}
		}
		log.Infof("Sleeping for 2 minutes to let VMs come up")
		time.Sleep(2 * time.Minute)

		canSsh = CreateSSHPodAndSetCanSsh()

		ValidateFioInVMs(appCtxs, canSsh)

		initialUptime = make(map[string]time.Duration)
		initialNodeName = make(map[string]string)
		stepLog = "Get initial uptime of VMs and current nodes"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			var wg sync.WaitGroup
			for _, appCtx := range appCtxs {
				wg.Add(1)
				go func(appCtx *scheduler.Context) {
					defer GinkgoRecover()
					defer wg.Done()
					vms, err := GetAllVMsFromScheduledContexts([]*scheduler.Context{appCtx})
					log.FailOnError(err, "Failed to get VMs from appCtx")
					for _, vm := range vms {
						vmKey := fmt.Sprintf("%s/%s", vm.Namespace, vm.Name)
						uptime, err := GetVMUptime(vm)
						log.FailOnError(err, "Failed to get uptime from VM %s", vm.Name)
						log.Infof("Initial uptime for VM %s is %v", vmKey, uptime)

						nodeName, err := GetNodeOfVM(vm)
						log.FailOnError(err, "Failed to get node of VM %v", vm.Name)
						log.Infof("VM %s is currently running on node %s", vm.Name, nodeName)

						mu.Lock()
						initialUptime[vmKey] = uptime
						initialNodeName[vmKey] = nodeName
						mu.Unlock()
					}
				}(appCtx)
			}
			wg.Wait()
		})

		for i := 1; i <= iterations; i++ {
			stepLog = fmt.Sprintf("Iteration %d: Live migrate the KubeVirt VM", i)
			Step(stepLog, func() {
				log.InfoD(stepLog)
				var wg sync.WaitGroup
				for _, appCtx := range appCtxs {
					wg.Add(1)
					go func(appCtx *scheduler.Context) {
						defer GinkgoRecover()
						defer wg.Done()
						err := StartAndWaitForVMIMigration(appCtx, context1.TODO())
						log.FailOnError(err, "Failed to live migrate KubeVirt VM")
					}(appCtx)
				}
				wg.Wait()
			})

			stepLog = fmt.Sprintf("Iteration %d: Validate VMs after migration", i)
			Step(stepLog, func() {
				log.InfoD(stepLog)
				var wg sync.WaitGroup
				for _, appCtx := range appCtxs {
					wg.Add(1)
					go func(appCtx *scheduler.Context) {
						defer GinkgoRecover()
						defer wg.Done()
						vms, err := GetAllVMsFromScheduledContexts([]*scheduler.Context{appCtx})
						log.FailOnError(err, "Failed to get VMs from appCtx")
						for _, vm := range vms {
							vmKey := fmt.Sprintf("%s/%s", vm.Namespace, vm.Name)
							newNodeName, err := GetNodeOfVM(vm)
							log.FailOnError(err, "Failed to get node of VM %s after migration", vm.Name)
							log.Infof("VM %s is now running on node %s after migration", vm.Name, newNodeName)
							mu.Lock()
							initialNode := initialNodeName[vmKey]
							mu.Unlock()
							if newNodeName == initialNode {
								failure = true
								log.FailOnError(fmt.Errorf("VM %s did not migrate to a different node", vm.Name), "VM is still on node %s after migration", initialNode)
							} else {
								log.Infof("VM %s successfully migrated from node %s to node %s", vm.Name, initialNode, newNodeName)
								mu.Lock()
								initialNodeName[vmKey] = newNodeName
								mu.Unlock()
							}

							err = CheckVMUptime(vm, initialUptime)
							log.FailOnError(err, "Failed to validate uptime in VM %s", vm.Name)
						}
					}(appCtx)
				}
				wg.Wait()
			})

			ValidateFioInVMs(appCtxs, canSsh)

			stepLog = fmt.Sprintf("Iteration %d: Add a disk to the KubeVirt VM", i)
			Step(stepLog, func() {
				log.InfoD(stepLog)
				_, err := AddRawBlockDriveToKubevirtVM(appCtxs, numberOfVolumes, "10Gi")
				log.FailOnError(err, "Failed to add disks to KubeVirt VM")
				dash.VerifyFatal(true, true, "Disk added to KubeVirt VM")
			})

			ValidateFioInVMs(appCtxs, canSsh)
		}

		stepLog = "Destroy Applications"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			if !failure {
				DestroyApps(appCtxs, nil)
			}
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(appCtxs)
	})
})

var _ = Describe("{PxKillAfterAddRawDiskToVM}", Label("p1", "negative", "kubevirt", "error_injection", "px_crash"), func() {
	var app, volType string
	var present bool
	JustBeforeEach(func() {
		StartTorpedoTest("PxKillAfterAddRawDiskToVM", "Add a disk to KubeVirt VM, kill Px, add another disk and validate the VM", nil, 0)
		volType, present = os.LookupEnv("KUBEVIRT_VOL_TYPE")
		if !present {
			app = "kubevirt-debian-fio-minimal"
		}
		if volType == "pxe-raw" {
			app = "kubevirt-raw-vol"
		} else if volType == "fada-raw" {
			app = "kubevirt-fada-raw-fio"
		}
		log.InfoD("Setting app for this test to be : %s", app)
	})

	var appCtxs []*scheduler.Context
	var nodes []string
	var namespace string
	var canSsh bool = false
	var initialUptime map[string]time.Duration
	var initialNodeName map[string]string
	var mu sync.Mutex

	itLog := "Add Fada Raw disk to KubeVirt VM, Kill Px and then add another disk"
	It(itLog, func() {
		pxNs, err := Inst().V.GetVolumeDriverNamespace()
		log.FailOnError(err, "Failed to get volume driver namespace")
		defer ListEvents(pxNs)

		namespace = fmt.Sprintf("kubevirt-%v", time.Now().Unix())
		appList := Inst().AppList
		defer func() {
			Inst().AppList = appList
		}()
		numberOfVolumes := 1
		Inst().AppList = []string{app}
		Inst().CsiAppList = []string{app}

		stepLog := "Schedule a KubeVirt VM"
		Step(stepLog, func() {
			appCtxs = append(appCtxs, ScheduleApplicationsOnNamespace(namespace, "test")...)
		})
		ValidateApplications(appCtxs)
		if !present {
			for _, appCtx := range appCtxs {
				bindMount, err := IsVMBindMounted(appCtx, false)
				log.FailOnError(err, "Failed to verify bind mount")
				dash.VerifyFatal(bindMount, true, "Failed to verify bind mount")
			}
		}
		log.Infof("Sleeping for 2 minutes to let VMs come up")
		time.Sleep(2 * time.Minute)

		canSsh = CreateSSHPodAndSetCanSsh()
		ValidateFioInVMs(appCtxs, canSsh)

		stepLog = "Add one disk to the KubeVirt VM"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			_, err := AddRawBlockDriveToKubevirtVM(appCtxs, numberOfVolumes, "10Gi")
			log.FailOnError(err, "Failed to add disks to KubeVirt VM")
			dash.VerifyFatal(true, true, "Disk added to KubeVirt VM")
		})

		initialUptime = make(map[string]time.Duration)
		initialNodeName = make(map[string]string)
		Step("Get initial uptime of VMs and current nodes", func() {
			log.InfoD("Get initial uptime of VMs and current nodes")
			var wg sync.WaitGroup
			for _, appCtx := range appCtxs {
				wg.Add(1)
				go func(appCtx *scheduler.Context) {
					defer GinkgoRecover()
					defer wg.Done()
					vms, err := GetAllVMsFromScheduledContexts([]*scheduler.Context{appCtx})
					log.FailOnError(err, "Failed to get VMs from appCtx")
					for _, vm := range vms {
						vmKey := fmt.Sprintf("%s/%s", vm.Namespace, vm.Name)
						uptime, err := GetVMUptime(vm)
						log.FailOnError(err, "Failed to get uptime from VM %s", vm.Name)
						log.Infof("Initial uptime for VM %s is %v", vmKey, uptime)

						nodeName, err := GetNodeOfVM(vm)
						log.FailOnError(err, "Failed to get node of VM %v", vm.Name)
						log.Infof("VM %s is currently running on node %s", vm.Name, nodeName)

						mu.Lock()
						initialUptime[vmKey] = uptime
						initialNodeName[vmKey] = nodeName
						mu.Unlock()
					}
				}(appCtx)
			}
			wg.Wait()
		})

		stepLog = "Kill Px on node hosting VM"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for _, appCtx := range appCtxs {
				vms, err := GetAllVMsFromScheduledContexts([]*scheduler.Context{appCtx})
				log.FailOnError(err, "Failed to get VMs from context")
				for _, vm := range vms {
					nodeName, err := GetNodeOfVM(vm)
					log.FailOnError(err, "Failed to get node of vm %v", vm.Name)
					nodes = append(nodes, nodeName)
				}
			}
			for _, appNode := range node.GetStorageDriverNodes() {
				for _, vmNode := range nodes {
					if vmNode == appNode.Name {
						stepLog = fmt.Sprintf("Stop volume driver %s on node: %s", Inst().V.String(), appNode.Name)
						Step(stepLog, func() {
							log.InfoD(stepLog)
							StopVolDriverAndWait([]node.Node{appNode})
						})

						stepLog = fmt.Sprintf("Start volume driver %s on node %s", Inst().V.String(), appNode.Name)
						Step(stepLog, func() {
							log.InfoD(stepLog)
							StartVolDriverAndWait([]node.Node{appNode})
						})

						stepLog = "Giving few seconds for volume driver to stabilize"
						Step(stepLog, func() {
							log.InfoD(stepLog)
							time.Sleep(20 * time.Second)
						})
					}
				}
			}
		})

		ValidateVMUptime(appCtxs, canSsh, initialUptime)

		stepLog = "Add another disk to the KubeVirt VM"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			_, err := AddRawBlockDriveToKubevirtVM(appCtxs, numberOfVolumes, "10Gi")
			log.FailOnError(err, "Failed to add disks to KubeVirt VM")
			dash.VerifyFatal(true, true, "Disk added to KubeVirt VM")
		})

		ValidateFioInVMs(appCtxs, canSsh)

		stepLog = "Destroy Applications"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			DestroyApps(appCtxs, nil)
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(appCtxs)
	})
})

var _ = Describe("{AddDiskKillPxLMAgainAddDisk}", Label("p1", "negative", "kubevirt", "error_injection", "px_crash"), func() {
	var app, volType string
	var present bool
	JustBeforeEach(func() {
		StartTorpedoTest("AddDiskKillPxLMAgainAddDisk", "Add a Fada raw disk to KubeVirt VM, live migrate, kill Px, and validate the VM", nil, 0)
		volType, present = os.LookupEnv("KUBEVIRT_VOL_TYPE")
		if !present {
			app = "kubevirt-debian-fio-minimal"
		}
		if volType == "pxe-raw" {
			app = "kubevirt-raw-vol"
		} else if volType == "fada-raw" {
			app = "kubevirt-fada-raw-fio"
		}
		log.InfoD("Setting app for this test to be : %s", app)
	})

	var appCtxs []*scheduler.Context
	var nodes []string
	var namespace string
	var canSsh bool = false
	var wg sync.WaitGroup
	var initialUptime map[string]time.Duration
	var initialNodeName map[string]string
	var mu sync.Mutex

	itLog := "Add Fada Raw disk to KubeVirt VM, live migrate, then kill Px, and add another disk"
	It(itLog, func() {
		pxNs, err := Inst().V.GetVolumeDriverNamespace()
		log.FailOnError(err, "Failed to get volume driver namespace")
		defer ListEvents(pxNs)

		namespace = fmt.Sprintf("kubevirt-%v", time.Now().Unix())
		appList := Inst().AppList
		defer func() {
			Inst().AppList = appList
		}()
		numberOfVolumes := 1
		Inst().AppList = []string{app}
		Inst().CsiAppList = []string{app}

		stepLog := "Schedule a KubeVirt VM"
		Step(stepLog, func() {
			appCtxs = append(appCtxs, ScheduleApplicationsOnNamespace(namespace, "test")...)
		})
		ValidateApplications(appCtxs)
		for _, appCtx := range appCtxs {
			bindMount, err := IsVMBindMounted(appCtx, false)
			log.FailOnError(err, "Failed to verify bind mount")
			dash.VerifyFatal(bindMount, true, "Failed to verify bind mount")
		}
		log.Infof("Sleeping for 2 minutes to let VMs come up")
		time.Sleep(2 * time.Minute)

		canSsh = CreateSSHPodAndSetCanSsh()
		ValidateFioInVMs(appCtxs, canSsh)

		stepLog = "Add one disk to the KubeVirt VM"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			_, err := AddRawBlockDriveToKubevirtVM(appCtxs, numberOfVolumes, "10Gi")
			log.FailOnError(err, "Failed to add disks to KubeVirt VM")
			dash.VerifyFatal(true, true, "Disk added to KubeVirt VM")
		})

		initialUptime = make(map[string]time.Duration)
		initialNodeName = make(map[string]string)
		stepLog = "Get initial uptime of VMs and current nodes"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			var wg sync.WaitGroup
			for _, appCtx := range appCtxs {
				wg.Add(1)
				go func(appCtx *scheduler.Context) {
					defer GinkgoRecover()
					defer wg.Done()
					vms, err := GetAllVMsFromScheduledContexts([]*scheduler.Context{appCtx})
					log.FailOnError(err, "Failed to get VMs from appCtx")
					for _, vm := range vms {
						vmKey := fmt.Sprintf("%s/%s", vm.Namespace, vm.Name)
						uptime, err := GetVMUptime(vm)
						log.FailOnError(err, "Failed to get uptime from VM %s", vm.Name)
						log.Infof("Initial uptime for VM %s is %v", vmKey, uptime)

						nodeName, err := GetNodeOfVM(vm)
						log.FailOnError(err, "Failed to get node of VM %v", vm.Name)
						log.Infof("VM %s is currently running on node %s", vm.Name, nodeName)

						mu.Lock()
						initialUptime[vmKey] = uptime
						initialNodeName[vmKey] = nodeName
						mu.Unlock()
					}
				}(appCtx)
			}
			wg.Wait()
		})

		stepLog = "Kill Px on node hosting VM"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for _, appCtx := range appCtxs {
				vms, err := GetAllVMsFromScheduledContexts([]*scheduler.Context{appCtx})
				log.FailOnError(err, "Failed to get VMs from context")
				for _, vm := range vms {
					nodeName, err := GetNodeOfVM(vm)
					log.FailOnError(err, "Failed to get node of VM %v", vm.Name)
					nodes = append(nodes, nodeName)
				}
			}
			for _, appNode := range node.GetStorageDriverNodes() {
				for _, vmNode := range nodes {
					if vmNode == appNode.Name {
						stepLog = fmt.Sprintf("Stop volume driver %s on node: %s",
							Inst().V.String(), appNode.Name)
						Step(stepLog, func() {
							log.InfoD(stepLog)
							StopVolDriverAndWait([]node.Node{appNode})
						})

						stepLog = fmt.Sprintf("Start volume driver %s on node %s",
							Inst().V.String(), appNode.Name)
						Step(stepLog, func() {
							log.InfoD(stepLog)
							StartVolDriverAndWait([]node.Node{appNode})
						})

						stepLog = "Giving few seconds for volume driver to stabilize"
						Step(stepLog, func() {
							log.InfoD(stepLog)
							time.Sleep(20 * time.Second)
						})
					}
				}
			}
		})

		ValidateVMUptime(appCtxs, canSsh, initialUptime)

		stepLog = "Live migrate the KubeVirt VM"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for _, appCtx := range appCtxs {
				wg.Add(1)
				go func(appCtx *scheduler.Context) {
					defer GinkgoRecover()
					defer wg.Done()
					err := StartAndWaitForVMIMigration(appCtx, context1.TODO())
					log.FailOnError(err, "Failed to live migrate KubeVirt VM")
				}(appCtx)
			}
			wg.Wait()
		})

		stepLog = "Validate VMs after migration"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			var wg sync.WaitGroup
			for _, appCtx := range appCtxs {
				wg.Add(1)
				go func(appCtx *scheduler.Context) {
					defer GinkgoRecover()
					defer wg.Done()
					vms, err := GetAllVMsFromScheduledContexts([]*scheduler.Context{appCtx})
					log.FailOnError(err, "Failed to get VMs from appCtx")
					for _, vm := range vms {
						vmKey := fmt.Sprintf("%s/%s", vm.Namespace, vm.Name)
						newNodeName, err := GetNodeOfVM(vm)
						log.FailOnError(err, "Failed to get node of VM %s after migration", vm.Name)
						log.Infof("VM %s is now running on node %s after migration", vm.Name, newNodeName)
						mu.Lock()
						initialNode := initialNodeName[vmKey]
						mu.Unlock()
						if newNodeName == initialNode {
							log.FailOnError(fmt.Errorf("VM %s did not migrate to a different node", vm.Name), "VM is still on node %s after migration", initialNode)
						} else {
							log.Infof("VM %s successfully migrated from node %s to node %s", vm.Name, initialNode, newNodeName)
							mu.Lock()
							initialNodeName[vmKey] = newNodeName
							mu.Unlock()
						}

						err = CheckVMUptime(vm, initialUptime)
						log.FailOnError(err, "Failed to validate uptime in VM %s", vm.Name)
					}
				}(appCtx)
			}
			wg.Wait()
		})

		stepLog = "Add another disk to the KubeVirt VM"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			_, err := AddRawBlockDriveToKubevirtVM(appCtxs, numberOfVolumes, "10Gi")
			log.FailOnError(err, "Failed to add disks to KubeVirt VM")
			dash.VerifyFatal(true, true, "Disk added to KubeVirt VM")
		})

		ValidateFioInVMs(appCtxs, canSsh)

		stepLog = "Destroy Applications"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			DestroyApps(appCtxs, nil)
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(appCtxs)
	})
})

var _ = Describe("{KillPxOnSourceNodeDuringMigration}", Label("p1", "negative", "kubevirt", "error_injection", "px_crash"), func() {
	var app, volType string
	var present bool
	JustBeforeEach(func() {
		StartTorpedoTest("KillPxOnSourceNodeDuringMigration", "Live migrate KubeVirt VM and kill Px on source node during migration", nil, 0)
		volType, present = os.LookupEnv("KUBEVIRT_VOL_TYPE")
		if !present {
			app = "kubevirt-debian-fio-minimal"
		}
		if volType == "pxe-raw" {
			app = "kubevirt-raw-vol"
		} else if volType == "fada-raw" {
			app = "kubevirt-fada-raw-fio"
		}
		log.InfoD("Setting app for this test to be : %s", app)
	})

	var appCtxs []*scheduler.Context
	var namespace string
	var migration *kubevirtdy.VirtualMachineInstanceMigration
	var vm kubevirtv1.VirtualMachine
	var vmNodeName string
	var canSsh bool
	var initialUptime map[string]time.Duration
	var initialNodeName map[string]string
	var mu sync.Mutex

	itLog := "Live migrate KubeVirt VM and kill Px during migration"
	It(itLog, func() {
		pxNs, err := Inst().V.GetVolumeDriverNamespace()
		log.FailOnError(err, "Failed to get volume driver namespace")
		defer ListEvents(pxNs)

		namespace = fmt.Sprintf("kubevirt-%v", time.Now().Unix())
		appList := Inst().AppList
		defer func() {
			Inst().AppList = appList
		}()
		Inst().AppList = []string{app}
		Inst().CsiAppList = []string{app}

		stepLog := "Schedule a KubeVirt VM"
		Step(stepLog, func() {
			appCtxs = append(appCtxs, ScheduleApplicationsOnNamespace(namespace, "test")...)
		})
		ValidateApplications(appCtxs)

		log.Infof("Sleeping for 2 minutes to let VMs come up")
		time.Sleep(2 * time.Minute)

		canSsh = CreateSSHPodAndSetCanSsh()
		ValidateFioInVMs(appCtxs, canSsh)

		initialUptime = make(map[string]time.Duration)
		initialNodeName = make(map[string]string)
		stepLog = "Get initial uptime of VMs and current nodes"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			vms, err := GetAllVMsFromScheduledContexts(appCtxs)
			log.FailOnError(err, "Failed to get VMs from context")

			if len(vms) == 0 {
				log.FailOnError(fmt.Errorf("No VMs found"), "No VMs found in context")
			} else {
				vm = vms[0]
				vmKey := fmt.Sprintf("%s/%s", vm.Namespace, vm.Name)
				uptime, err := GetVMUptime(vm)
				log.FailOnError(err, "Failed to get uptime from VM %s", vm.Name)
				log.Infof("Initial uptime for VM %s is %v", vmKey, uptime)

				vmNodeName, err = GetNodeOfVM(vm)
				log.FailOnError(err, "Failed to get node of VM %v", vm.Name)
				log.Infof("VM %s is running on node %s", vm.Name, vmNodeName)

				mu.Lock()
				initialUptime[vmKey] = uptime
				initialNodeName[vmKey] = vmNodeName
				mu.Unlock()
			}
		})

		stepLog = "Start live migration of the KubeVirt VM"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			migration, err = kubevirtdy.Instance().CreateVirtualMachineInstanceMigration(context1.TODO(), vm.Namespace, vm.Name)
			log.FailOnError(err, "Failed to create VM migration for VM %s", vm.Name)
			log.Infof("Migration %s created for VM %s", migration.Name, vm.Name)
		})

		migrationCreatedChan := make(chan struct{})
		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer GinkgoRecover()
			defer wg.Done()

			close(migrationCreatedChan)

			stepLog = "Wait for migration to complete"
			Step(stepLog, func() {
				log.InfoD(stepLog)
				t := func() (interface{}, bool, error) {
					migr, err := kubevirtdy.Instance().GetVirtualMachineInstanceMigration(context1.TODO(), vm.Namespace, migration.Name)
					if err != nil {
						log.Infof("Error getting migration: %v", err)
						return nil, true, err
					}
					if migr.Phase == string(kubevirtv1.MigrationSucceeded) {
						log.Infof("Migration has succeeded")
						return nil, false, nil
					} else if migr.Phase == string(kubevirtv1.MigrationFailed) {
						return nil, false, fmt.Errorf("Migration has failed")
					}
					return nil, true, fmt.Errorf("Migration not yet completed")
				}
				_, err := task.DoRetryWithTimeout(t, triggerCheckTimeout, triggerCheckInterval)
				log.FailOnError(err, "Failed to wait for migration to complete")
			})

			stepLog = "Check the state of the VM after migration"
			Step(stepLog, func() {
				log.InfoD(stepLog)
				err := CheckVMState(vm)
				log.FailOnError(err, "Failed to verify VM state after migration")
			})

			stepLog = "Validate VM has migrated to a different node"
			Step(stepLog, func() {
				log.InfoD(stepLog)
				newNodeName, err := GetNodeOfVM(vm)
				log.FailOnError(err, "Failed to get node of VM %s after migration", vm.Name)
				log.Infof("VM %s is now running on node %s after migration", vm.Name, newNodeName)
				mu.Lock()
				initialNode := initialNodeName[fmt.Sprintf("%s/%s", vm.Namespace, vm.Name)]
				mu.Unlock()
				if newNodeName == initialNode {
					log.FailOnError(fmt.Errorf("VM %s did not migrate to a different node", vm.Name), "VM is still on node %s after migration", initialNode)
				} else {
					log.Infof("VM %s successfully migrated from node %s to node %s", vm.Name, initialNode, newNodeName)
					mu.Lock()
					initialNodeName[fmt.Sprintf("%s/%s", vm.Namespace, vm.Name)] = newNodeName
					mu.Unlock()
				}
			})

			ValidateVMUptime([]*scheduler.Context{appCtxs[0]}, canSsh, initialUptime)

			ValidateFioInVMs(appCtxs, canSsh)

			stepLog = "Add one disk to the KubeVirt VM"
			Step(stepLog, func() {
				log.InfoD(stepLog)
				numberOfVolumes := 1
				_, err := AddRawBlockDriveToKubevirtVM(appCtxs, numberOfVolumes, "10Gi")
				log.FailOnError(err, "Failed to add disks to KubeVirt VM")
				dash.VerifyFatal(true, true, "Disk added to KubeVirt VM")
			})

			ValidateVMUptime([]*scheduler.Context{appCtxs[0]}, canSsh, initialUptime)

			ValidateFioInVMs(appCtxs, canSsh)
		}()

		go func() {
			defer GinkgoRecover()
			defer wg.Done()

			<-migrationCreatedChan

			t := func() (interface{}, bool, error) {
				migr, err := kubevirtdy.Instance().GetVirtualMachineInstanceMigration(context1.TODO(), vm.Namespace, migration.Name)
				if err != nil {
					log.Infof("Error getting migration: %v", err)
					return nil, true, err
				}
				phase := migr.Phase
				if phase == string(kubevirtv1.MigrationRunning) {
					log.Infof("Migration is in progress")
					return nil, false, nil
				} else if phase == string(kubevirtv1.MigrationSucceeded) {
					log.Infof("Migration has already succeeded")
					return nil, false, nil
				} else if phase == string(kubevirtv1.MigrationFailed) {
					log.Infof("Migration has failed")
					return nil, false, fmt.Errorf("Migration has failed")
				}
				return nil, true, fmt.Errorf("Migration not yet in progress, current phase: %s", phase)
			}
			_, err := task.DoRetryWithTimeout(t, triggerCheckTimeout, triggerCheckInterval)
			log.FailOnError(err, "Failed to wait for migration to be in progress")

			appNode, err := node.GetNodeByName(vmNodeName)
			log.FailOnError(err, "Failed to get node object for node %s", vmNodeName)

			stepLog = fmt.Sprintf("Stop volume driver %s on node: %s", Inst().V.String(), appNode.Name)
			Step(stepLog, func() {
				log.InfoD(stepLog)
				StopVolDriverAndWait([]node.Node{appNode})
			})

			stepLog = fmt.Sprintf("Start volume driver %s on node: %s", Inst().V.String(), appNode.Name)
			Step(stepLog, func() {
				log.InfoD(stepLog)
				StartVolDriverAndWait([]node.Node{appNode})
			})

			stepLog = "Giving few seconds for volume driver to stabilize"
			Step(stepLog, func() {
				log.InfoD(stepLog)
				time.Sleep(20 * time.Second)
			})
		}()

		wg.Wait()

		stepLog = "Destroy Applications"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			DestroyApps(appCtxs, nil)
		})
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(appCtxs)
	})
})

var _ = Describe("{KillPxOnDestNodeDuringMigration}", Label("p1", "negative", "kubevirt", "error_injection", "px_crash"), func() {
	var app, volType string
	var present bool
	JustBeforeEach(func() {
		StartTorpedoTest("KillPxOnDestNodeDuringMigration", "Live migrate KubeVirt VM and kill Px on destination node during migration", nil, 0)
		volType, present = os.LookupEnv("KUBEVIRT_VOL_TYPE")
		if !present {
			app = "kubevirt-debian-fio-minimal"
		}
		if volType == "pxe-raw" {
			app = "kubevirt-raw-vol"
		} else if volType == "fada-raw" {
			app = "kubevirt-fada-raw-fio"
		}
		log.InfoD("Setting app for this test to be : %s", app)
	})

	var appCtxs []*scheduler.Context
	var namespace string
	var migration *kubevirtdy.VirtualMachineInstanceMigration
	var vm kubevirtv1.VirtualMachine
	var vmNodeName string
	var destNodeName string
	var initialUptime map[string]time.Duration

	itLog := "Live migrate KubeVirt VM and kill Px on destination node during migration"
	It(itLog, func() {
		pxNs, err := Inst().V.GetVolumeDriverNamespace()
		log.FailOnError(err, "Failed to get volume driver namespace")
		defer ListEvents(pxNs)

		namespace = fmt.Sprintf("kubevirt-%v", time.Now().Unix())
		appList := Inst().AppList
		defer func() {
			Inst().AppList = appList
		}()
		Inst().AppList = []string{app}
		Inst().CsiAppList = []string{app}

		stepLog := "Schedule a KubeVirt VM"
		Step(stepLog, func() {
			appCtxs = append(appCtxs, ScheduleApplicationsOnNamespace(namespace, "test")...)
		})
		ValidateApplications(appCtxs)

		log.Infof("Sleeping for 2 minutes to let VMs come up")
		time.Sleep(2 * time.Minute)

		canSsh := CreateSSHPodAndSetCanSsh()

		ValidateFioInVMs(appCtxs, canSsh)

		initialUptime = make(map[string]time.Duration)
		stepLog = "Get initial uptime of VMs"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			var wg sync.WaitGroup
			for _, appCtx := range appCtxs {
				wg.Add(1)
				go func(appCtx *scheduler.Context) {
					defer GinkgoRecover()
					defer wg.Done()
					vms, err := GetAllVMsFromScheduledContexts([]*scheduler.Context{appCtx})
					log.FailOnError(err, "Failed to get VMs from appCtx")
					for _, vm := range vms {
						uptime, err := GetVMUptime(vm)
						log.FailOnError(err, "Failed to get uptime from VM %s", vm.Name)
						vmKey := fmt.Sprintf("%s/%s", vm.Namespace, vm.Name)
						initialUptime[vmKey] = uptime
						log.Infof("Initial uptime for VM %s is %v", vmKey, uptime)
					}
				}(appCtx)
			}
			wg.Wait()
		})

		vms, err := GetAllVMsFromScheduledContexts(appCtxs)
		log.FailOnError(err, "Failed to get VMs from context")

		if len(vms) == 0 {
			log.FailOnError(fmt.Errorf("No VMs found"), "No VMs found in context")
		} else {
			vm = vms[0]
			vmNodeName, err = GetNodeOfVM(vm)
			log.FailOnError(err, "Failed to get node of VM %v", vm.Name)
			log.Infof("VM %s is running on node %s", vm.Name, vmNodeName)
		}

		stepLog = "Start live migration of the KubeVirt VM"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			migration, err = kubevirtdy.Instance().CreateVirtualMachineInstanceMigration(context1.TODO(), vm.Namespace, vm.Name)
			log.FailOnError(err, "Failed to create VM migration for VM %s", vm.Name)
			log.Infof("Migration %s created for VM %s", migration.Name, vm.Name)
		})

		migrationCreatedChan := make(chan struct{})
		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer GinkgoRecover()
			defer wg.Done()

			close(migrationCreatedChan)

			stepLog = "Wait for migration to complete"
			Step(stepLog, func() {
				log.InfoD(stepLog)
				t := func() (interface{}, bool, error) {
					migr, err := kubevirtdy.Instance().GetVirtualMachineInstanceMigration(context1.TODO(), vm.Namespace, migration.Name)
					if err != nil {
						log.Infof("Error getting migration: %v", err)
						return nil, true, err
					}
					if migr.Phase == string(kubevirtv1.MigrationSucceeded) {
						log.Infof("Migration has succeeded")
						return nil, false, nil
					} else if migr.Phase == string(kubevirtv1.MigrationFailed) {
						return nil, false, fmt.Errorf("Migration has failed")
					}
					return nil, true, fmt.Errorf("Migration not yet completed")
				}
				_, err := task.DoRetryWithTimeout(t, triggerCheckTimeout, triggerCheckInterval)
				log.FailOnError(err, "Failed to wait for migration to complete")
			})

			stepLog = "Check the state of the VM after migration"
			Step(stepLog, func() {
				log.InfoD(stepLog)
				err := CheckVMState(vm)
				log.FailOnError(err, "Failed to verify VM state after migration")
			})

			ValidateVMUptime(appCtxs, canSsh, initialUptime)

			ValidateFioInVMs(appCtxs, canSsh)

			stepLog = "Add one disk to the KubeVirt VM"
			Step(stepLog, func() {
				log.InfoD(stepLog)
				numberOfVolumes := 1
				_, err := AddRawBlockDriveToKubevirtVM(appCtxs, numberOfVolumes, "10Gi")
				log.FailOnError(err, "Failed to add disks to KubeVirt VM")
				dash.VerifyFatal(true, true, "Disk added to KubeVirt VM")
			})

			ValidateFioInVMs(appCtxs, canSsh)
		}()

		go func() {
			defer GinkgoRecover()
			defer wg.Done()

			<-migrationCreatedChan

			var migr *kubevirtdy.VirtualMachineInstanceMigration
			t := func() (interface{}, bool, error) {
				tempMigr, err := kubevirtdy.Instance().GetVirtualMachineInstanceMigration(context1.TODO(), vm.Namespace, migration.Name)
				if err != nil {
					log.Infof("Error getting migration: %v", err)
					return nil, true, err
				}
				phase := tempMigr.Phase
				if phase == string(kubevirtv1.MigrationRunning) {
					log.Infof("Migration is in progress")
					return tempMigr, false, nil
				} else if phase == string(kubevirtv1.MigrationSucceeded) {
					log.Infof("Migration has already succeeded")
					return tempMigr, false, nil
				} else if phase == string(kubevirtv1.MigrationFailed) {
					log.Infof("Migration has failed")
					return tempMigr, false, fmt.Errorf("Migration has failed")
				}
				return nil, true, fmt.Errorf("Migration not yet in progress, current phase: %s", phase)
			}
			result, err := task.DoRetryWithTimeout(t, triggerCheckTimeout, triggerCheckInterval)
			log.FailOnError(err, "Failed to wait for migration to be in progress")

			t = func() (interface{}, bool, error) {
				tempMigr, err := kubevirtdy.Instance().GetVirtualMachineInstanceMigration(context1.TODO(), vm.Namespace, migration.Name)
				if err != nil {
					log.Infof("Error getting migration: %v", err)
					return nil, true, err
				}
				if tempMigr.TargetNode != "" {
					return tempMigr, false, nil
				}
				log.Infof("Waiting for TargetNode to be populated")
				return nil, true, fmt.Errorf("TargetNode not yet set")
			}
			result, err = task.DoRetryWithTimeout(t, triggerCheckTimeout, triggerCheckInterval)
			log.FailOnError(err, "Failed to get destination node from migration")
			migr = result.(*kubevirtdy.VirtualMachineInstanceMigration)
			destNodeName = migr.TargetNode
			log.Infof("Destination node for migration is %s", destNodeName)

			appNode, err := node.GetNodeByName(destNodeName)
			log.FailOnError(err, "Failed to get node object for node %s", destNodeName)

			stepLog = fmt.Sprintf("Stop volume driver %s on destination node: %s", Inst().V.String(), appNode.Name)
			Step(stepLog, func() {
				log.InfoD(stepLog)
				StopVolDriverAndWait([]node.Node{appNode})
			})

			stepLog = fmt.Sprintf("Start volume driver %s on destination node: %s", Inst().V.String(), appNode.Name)
			Step(stepLog, func() {
				log.InfoD(stepLog)
				StartVolDriverAndWait([]node.Node{appNode})
			})

			stepLog = "Giving few seconds for volume driver to stabilize"
			Step(stepLog, func() {
				log.InfoD(stepLog)
				time.Sleep(20 * time.Second)
			})
		}()

		wg.Wait()

		// Get and log the node where the VM is running after migration
		stepLog = "Get VM node after migration"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			newNodeName, err := GetNodeOfVM(vm)
			log.FailOnError(err, fmt.Sprintf("Failed to fetch VM %v Node after migration", vm.Name))
			log.Infof("VM %v is now running on Node %v", vm.Name, newNodeName)
			if newNodeName == vmNodeName {
				log.InfoD("VM did not migrate to a different node. VM is still on node %s after migration", vmNodeName)
				log.InfoD("Let's try to migrate it again")
				stepLog = "Live migrate the KubeVirt VM"
				Step(stepLog, func() {
					log.InfoD(stepLog)
					for _, appCtx := range appCtxs {
						wg.Add(1)
						go func(appCtx *scheduler.Context) {
							defer GinkgoRecover()
							defer wg.Done()
							err := StartAndWaitForVMIMigration(appCtx, context1.TODO())
							log.FailOnError(err, "Failed to live migrate KubeVirt VM")
						}(appCtx)
					}
					wg.Wait()
				})
				nextNewNodeName, err := GetNodeOfVM(vm)
				log.FailOnError(err, fmt.Sprintf("Failed to fetch VM %v Node after migration", vm.Name))
				log.Infof("VM %v is now running on Node %v", vm.Name, nextNewNodeName)
				if nextNewNodeName == vmNodeName {
					log.FailOnError(fmt.Errorf("VM did not move out of original node"), "VM Migration failed")
				} else {
					log.Infof("VM successfully migrated from node %s to node %s", vmNodeName, nextNewNodeName)
					ValidateVMUptime(appCtxs, canSsh, initialUptime)
				}
			} else {
				log.Infof("VM successfully migrated from node %s to node %s", vmNodeName, newNodeName)
			}
		})

		stepLog = "Destroy Applications"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			DestroyApps(appCtxs, nil)
		})
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(appCtxs)
	})
})

var _ = Describe("{KillPxOnSrcAndDestNodesDuringLM}", Label("p1", "negative", "kubevirt", "error_injection", "px_crash"), func() {
	var app, volType string
	var present bool
	JustBeforeEach(func() {
		StartTorpedoTest("KillPxOnSrcAndDestNodesDuringLM", "Live migrate KubeVirt VM and kill Px on both source and destination nodes during migration", nil, 0)
		volType, present = os.LookupEnv("KUBEVIRT_VOL_TYPE")
		if !present {
			app = "kubevirt-debian-fio-minimal"
		}
		if volType == "pxe-raw" {
			app = "kubevirt-raw-vol"
		} else if volType == "fada-raw" {
			app = "kubevirt-fada-raw-fio"
		}
		log.InfoD("Setting app for this test to be : %s", app)
	})

	var appCtxs []*scheduler.Context
	var namespace string
	var migration *kubevirtdy.VirtualMachineInstanceMigration
	var vm kubevirtv1.VirtualMachine
	var vmNodeName string
	var destNodeName string
	var canSsh bool
	var initialUptime map[string]time.Duration
	var initialNodeName map[string]string
	var mu sync.Mutex

	itLog := "Live migrate KubeVirt VM and kill Px on both source and destination nodes during migration"
	It(itLog, func() {
		pxNs, err := Inst().V.GetVolumeDriverNamespace()
		log.FailOnError(err, "Failed to get volume driver namespace")
		defer ListEvents(pxNs)

		namespace = fmt.Sprintf("kubevirt-%v", time.Now().Unix())
		appList := Inst().AppList
		defer func() {
			Inst().AppList = appList
		}()
		Inst().AppList = []string{app}
		Inst().CsiAppList = []string{app}

		stepLog := "Schedule a KubeVirt VM"
		Step(stepLog, func() {
			appCtxs = append(appCtxs, ScheduleApplicationsOnNamespace(namespace, "test")...)
		})
		ValidateApplications(appCtxs)

		log.Infof("Sleeping for 2 minutes to let VMs come up")
		time.Sleep(2 * time.Minute)

		canSsh = CreateSSHPodAndSetCanSsh()
		ValidateFioInVMs(appCtxs, canSsh)

		initialUptime = make(map[string]time.Duration)
		initialNodeName = make(map[string]string)
		stepLog = "Get initial uptime of VMs and current nodes"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			vms, err := GetAllVMsFromScheduledContexts(appCtxs)
			log.FailOnError(err, "Failed to get VMs from context")

			if len(vms) == 0 {
				log.FailOnError(fmt.Errorf("No VMs found"), "No VMs found in context")
			} else {
				vm = vms[0]
				vmKey := fmt.Sprintf("%s/%s", vm.Namespace, vm.Name)
				uptime, err := GetVMUptime(vm)
				log.FailOnError(err, "Failed to get uptime from VM %s", vm.Name)
				log.Infof("Initial uptime for VM %s is %v", vmKey, uptime)

				vmNodeName, err = GetNodeOfVM(vm)
				log.FailOnError(err, "Failed to get node of VM %v", vm.Name)
				log.Infof("VM %s is running on node %s", vm.Name, vmNodeName)

				mu.Lock()
				initialUptime[vmKey] = uptime
				initialNodeName[vmKey] = vmNodeName
				mu.Unlock()
			}
		})

		stepLog = "Start live migration of the KubeVirt VM"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			migration, err = kubevirtdy.Instance().CreateVirtualMachineInstanceMigration(context1.TODO(), vm.Namespace, vm.Name)
			log.FailOnError(err, "Failed to create VM migration for VM %s", vm.Name)
			log.Infof("Migration %s created for VM %s", migration.Name, vm.Name)
		})

		migrationCreatedChan := make(chan struct{})
		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer GinkgoRecover()
			defer wg.Done()

			close(migrationCreatedChan)

			stepLog = "Wait for migration to complete"
			Step(stepLog, func() {
				log.InfoD(stepLog)
				t := func() (interface{}, bool, error) {
					migr, err := kubevirtdy.Instance().GetVirtualMachineInstanceMigration(context1.TODO(), vm.Namespace, migration.Name)
					if err != nil {
						log.Infof("Error getting migration: %v", err)
						return nil, true, err
					}
					if migr.Phase == string(kubevirtv1.MigrationSucceeded) {
						log.Infof("Migration has succeeded")
						return nil, false, nil
					} else if migr.Phase == string(kubevirtv1.MigrationFailed) {
						return nil, false, fmt.Errorf("Migration has failed")
					}
					return nil, true, fmt.Errorf("Migration not yet completed")
				}
				_, err := task.DoRetryWithTimeout(t, triggerCheckTimeout, triggerCheckInterval)
				log.FailOnError(err, "Failed to wait for migration to complete")
			})

			stepLog = "Check the state of the VM after migration"
			Step(stepLog, func() {
				log.InfoD(stepLog)
				err := CheckVMState(vm)
				log.FailOnError(err, "Failed to verify VM state after migration")
			})

			stepLog = "Validate VM has migrated to a different node"
			Step(stepLog, func() {
				log.InfoD(stepLog)
				newNodeName, err := GetNodeOfVM(vm)
				log.FailOnError(err, "Failed to get node of VM %s after migration", vm.Name)
				log.Infof("VM %s is now running on node %s after migration", vm.Name, newNodeName)
				mu.Lock()
				initialNode := initialNodeName[fmt.Sprintf("%s/%s", vm.Namespace, vm.Name)]
				mu.Unlock()
				if newNodeName == initialNode {
					log.Infof("VM did not migrate to a different node. VM is still on node %s after migration", initialNode)
				} else {
					log.Infof("VM successfully migrated from node %s to node %s", initialNode, newNodeName)
					mu.Lock()
					initialNodeName[fmt.Sprintf("%s/%s", vm.Namespace, vm.Name)] = newNodeName
					mu.Unlock()
				}
			})

			ValidateVMUptime([]*scheduler.Context{appCtxs[0]}, canSsh, initialUptime)

			ValidateFioInVMs(appCtxs, canSsh)
		}()

		go func() {
			defer GinkgoRecover()
			defer wg.Done()

			<-migrationCreatedChan

			t := func() (interface{}, bool, error) {
				migr, err := kubevirtdy.Instance().GetVirtualMachineInstanceMigration(context1.TODO(), vm.Namespace, migration.Name)
				if err != nil {
					log.Infof("Error getting migration: %v", err)
					return nil, true, err
				}
				phase := migr.Phase
				if phase == string(kubevirtv1.MigrationRunning) {
					log.Infof("Migration is in progress")
					return migr, false, nil
				} else if phase == string(kubevirtv1.MigrationSucceeded) {
					log.Infof("Migration has already succeeded")
					return migr, false, nil
				} else if phase == string(kubevirtv1.MigrationFailed) {
					log.Infof("Migration has failed")
					return migr, false, fmt.Errorf("Migration has failed")
				}
				return nil, true, fmt.Errorf("Migration not yet in progress, current phase: %s", phase)
			}
			result, err := task.DoRetryWithTimeout(t, triggerCheckTimeout, triggerCheckInterval)
			log.FailOnError(err, "Failed to wait for migration to be in progress")
			migr := result.(*kubevirtdy.VirtualMachineInstanceMigration)

			t = func() (interface{}, bool, error) {
				tempMigr, err := kubevirtdy.Instance().GetVirtualMachineInstanceMigration(context1.TODO(), vm.Namespace, migration.Name)
				if err != nil {
					log.Infof("Error getting migration: %v", err)
					return nil, true, err
				}
				if tempMigr.TargetNode != "" {
					return tempMigr, false, nil
				}
				log.Infof("Waiting for TargetNode to be populated")
				return nil, true, fmt.Errorf("TargetNode not yet set")
			}
			result, err = task.DoRetryWithTimeout(t, triggerCheckTimeout, triggerCheckInterval)
			log.FailOnError(err, "Failed to get destination node from migration")
			migr = result.(*kubevirtdy.VirtualMachineInstanceMigration)
			destNodeName = migr.TargetNode
			log.Infof("Destination node for migration is %s", destNodeName)

			// Get node objects
			sourceNode, err := node.GetNodeByName(vmNodeName)
			log.FailOnError(err, "Failed to get node object for source node %s", vmNodeName)

			destNode, err := node.GetNodeByName(destNodeName)
			log.FailOnError(err, "Failed to get node object for destination node %s", destNodeName)

			stepLog = fmt.Sprintf("Stop volume driver %s on source node: %s and destination node: %s", Inst().V.String(), sourceNode.Name, destNode.Name)
			Step(stepLog, func() {
				log.InfoD(stepLog)
				StopVolDriverAndWait([]node.Node{sourceNode, destNode})
			})

			stepLog = fmt.Sprintf("Start volume driver %s on source node: %s and destination node: %s", Inst().V.String(), sourceNode.Name, destNode.Name)
			Step(stepLog, func() {
				log.InfoD(stepLog)
				StartVolDriverAndWait([]node.Node{sourceNode, destNode})
			})

			stepLog = "Giving few seconds for volume driver to stabilize"
			Step(stepLog, func() {
				log.InfoD(stepLog)
				time.Sleep(20 * time.Second)
			})
		}()

		wg.Wait()

		stepLog = "Get VM node after migration"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			newNodeName, err := GetNodeOfVM(vm)
			log.FailOnError(err, fmt.Sprintf("Failed to fetch VM %v Node after migration", vm.Name))
			log.Infof("VM %v is now running on Node %v", vm.Name, newNodeName)
			mu.Lock()
			initialNode := initialNodeName[fmt.Sprintf("%s/%s", vm.Namespace, vm.Name)]
			mu.Unlock()
			if newNodeName == initialNode {
				log.InfoD("VM did not migrate to a different node. VM is still on node %s after migration", initialNode)
				log.InfoD("Let's try to migrate it again")

				stepLog = "Live migrate the KubeVirt VM again"
				Step(stepLog, func() {
					log.InfoD(stepLog)
					err := StartAndWaitForVMIMigration(appCtxs[0], context1.TODO())
					log.FailOnError(err, "Failed to live migrate KubeVirt VM")
				})

				newNodeName, err = GetNodeOfVM(vm)
				log.FailOnError(err, fmt.Sprintf("Failed to fetch VM %v Node after migration", vm.Name))
				log.Infof("VM %v is now running on Node %v", vm.Name, newNodeName)
				if newNodeName == initialNode {
					log.FailOnError(fmt.Errorf("VM did not move out of original node"), "VM Migration failed")
				} else {
					log.Infof("VM successfully migrated from node %s to node %s", initialNode, newNodeName)
					mu.Lock()
					initialNodeName[fmt.Sprintf("%s/%s", vm.Namespace, vm.Name)] = newNodeName
					mu.Unlock()

					ValidateVMUptime([]*scheduler.Context{appCtxs[0]}, canSsh, initialUptime)
				}
			} else {
				log.Infof("VM successfully migrated from node %s to node %s", initialNode, newNodeName)
				mu.Lock()
				initialNodeName[fmt.Sprintf("%s/%s", vm.Namespace, vm.Name)] = newNodeName
				mu.Unlock()

				ValidateVMUptime([]*scheduler.Context{appCtxs[0]}, canSsh, initialUptime)
			}
		})

		ValidateFioInVMs(appCtxs, canSsh)

		stepLog = "Destroy Applications"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			DestroyApps(appCtxs, nil)
		})
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(appCtxs)
	})
})

var _ = Describe("{DeleteMigrationObjectDuringMigration}", Label("p0", "positive", "kubevirt", "LiveMigration"), func() {
	var app, volType string
	var present bool
	JustBeforeEach(func() {
		StartTorpedoTest("DeleteMigrationObjectDuringMigration", "Live migrate multiple KubeVirt VMs in parallel and delete migration objects during migration", nil, 0)
		volType, present = os.LookupEnv("KUBEVIRT_VOL_TYPE")
		if !present {
			app = "kubevirt-debian-fio-minimal"
		}
		if volType == "pxe-raw" {
			app = "kubevirt-raw-vol"
		} else if volType == "fada-raw" {
			app = "kubevirt-fada-raw-fio"
		}
		log.InfoD("Setting app for this test to be : %s", app)
	})
	var appCtxs []*scheduler.Context
	var namespace string
	var canSsh bool
	var initialUptime map[string]time.Duration
	var initialNodeName map[string]string
	var mu sync.Mutex

	itLog := "Live migrate multiple KubeVirt VMs in parallel and delete migration objects during migration"
	It(itLog, func() {
		pxNs, err := Inst().V.GetVolumeDriverNamespace()
		log.FailOnError(err, "Failed to get volume driver namespace")
		defer ListEvents(pxNs)
		canSsh = false

		appList := Inst().AppList
		defer func() {
			Inst().AppList = appList
		}()
		Inst().AppList = []string{app}
		Inst().CsiAppList = []string{app}

		stepLog := "Schedule KubeVirt VMs"
		Step(stepLog, func() {
			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				namespace = fmt.Sprintf("kubevirt-%v", time.Now().Unix())
				appCtxs = append(appCtxs, ScheduleApplicationsOnNamespace(namespace, "test")...)
			}
		})
		ValidateApplications(appCtxs)

		log.Infof("Sleeping for 2 minutes to let VMs come up")
		time.Sleep(2 * time.Minute)

		canSsh = CreateSSHPodAndSetCanSsh()
		ValidateFioInVMs(appCtxs, canSsh)

		initialUptime = make(map[string]time.Duration)
		initialNodeName = make(map[string]string)
		stepLog = "Get initial uptime of VMs and current nodes"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			var wg sync.WaitGroup
			for _, appCtx := range appCtxs {
				wg.Add(1)
				go func(appCtx *scheduler.Context) {
					defer GinkgoRecover()
					defer wg.Done()
					vms, err := GetAllVMsFromScheduledContexts([]*scheduler.Context{appCtx})
					log.FailOnError(err, "Failed to get VMs from appCtx")
					for _, vm := range vms {
						vmKey := fmt.Sprintf("%s/%s", vm.Namespace, vm.Name)
						uptime, err := GetVMUptime(vm)
						log.FailOnError(err, "Failed to get uptime from VM %s", vm.Name)
						log.Infof("Initial uptime for VM %s is %v", vmKey, uptime)

						nodeName, err := GetNodeOfVM(vm)
						log.FailOnError(err, "Failed to get node of VM %v", vm.Name)
						log.Infof("VM %s is currently running on node %s", vm.Name, nodeName)

						mu.Lock()
						initialUptime[vmKey] = uptime
						initialNodeName[vmKey] = nodeName
						mu.Unlock()
					}
				}(appCtx)
			}
			wg.Wait()
		})

		var wg sync.WaitGroup
		migrationMap := make(map[string]*kubevirtdy.VirtualMachineInstanceMigration)
		stepLog = "Start live migration of the KubeVirt VMs"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for _, appCtx := range appCtxs {
				wg.Add(1)
				go func(appCtx *scheduler.Context) {
					defer GinkgoRecover()
					defer wg.Done()
					vms, err := GetAllVMsFromScheduledContexts([]*scheduler.Context{appCtx})
					log.FailOnError(err, "Failed to get VMs from appCtx")
					for _, vm := range vms {
						migration, err := kubevirtdy.Instance().CreateVirtualMachineInstanceMigration(context1.TODO(), vm.Namespace, vm.Name)
						log.FailOnError(err, "Failed to create VM migration for VM %s", vm.Name)
						log.Infof("Migration %s created for VM %s", migration.Name, vm.Name)
						vmKey := fmt.Sprintf("%s/%s", vm.Namespace, vm.Name)
						mu.Lock()
						migrationMap[vmKey] = migration
						mu.Unlock()
					}
				}(appCtx)
			}
			wg.Wait()
		})

		stepLog = "Delete migration objects during migration"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			var wg sync.WaitGroup
			for vmKey, migration := range migrationMap {
				wg.Add(1)
				go func(vmKey string, migration *kubevirtdy.VirtualMachineInstanceMigration) {
					defer GinkgoRecover()
					defer wg.Done()
					vmNamespace, vmName := migration.NameSpace, migration.VMIName

					t := func() (interface{}, bool, error) {
						migr, err := kubevirtdy.Instance().GetVirtualMachineInstanceMigration(context1.TODO(), vmNamespace, migration.Name)
						if err != nil {
							log.Infof("Error getting migration: %v", err)
							return nil, true, err
						}
						phase := migr.Phase
						if phase == string(kubevirtv1.MigrationRunning) {
							log.Infof("Migration %s for VM %s is in progress", migration.Name, vmName)
							return migr, false, nil
						} else if phase == string(kubevirtv1.MigrationSucceeded) || phase == string(kubevirtv1.MigrationFailed) {
							log.Infof("Migration %s for VM %s has already completed with phase %s", migration.Name, vmName, phase)
							return migr, false, nil
						}
						return nil, true, fmt.Errorf("Migration %s for VM %s not yet in progress, current phase: %s", migration.Name, vmName, phase)
					}
					_, err := task.DoRetryWithTimeout(t, triggerCheckTimeout, triggerCheckInterval)
					log.FailOnError(err, "Failed to wait for migration to be in progress")

					dynamicClient, err := GetDynamicKubeClient()
					log.FailOnError(err, "Failed to get dynamic client")

					vmimGVR := schema.GroupVersionResource{
						Group:    "kubevirt.io",
						Version:  "v1",
						Resource: "virtualmachineinstancemigrations",
					}

					err = dynamicClient.Resource(vmimGVR).Namespace(vmNamespace).Delete(context1.TODO(), migration.Name, metav1.DeleteOptions{})
					log.FailOnError(err, "Failed to delete migration %s for VM %s", migration.Name, vmName)
					log.Infof("Deleted migration %s for VM %s", migration.Name, vmName)
				}(vmKey, migration)
			}
			wg.Wait()
		})

		stepLog = "Validate VMs after deleting migration objects"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			var wg sync.WaitGroup
			for _, appCtx := range appCtxs {
				wg.Add(1)
				go func(appCtx *scheduler.Context) {
					defer GinkgoRecover()
					defer wg.Done()
					vms, err := GetAllVMsFromScheduledContexts([]*scheduler.Context{appCtx})
					log.FailOnError(err, "Failed to get VMs from appCtx")
					for _, vm := range vms {
						vmKey := fmt.Sprintf("%s/%s", vm.Namespace, vm.Name)
						err := CheckVMState(vm)
						log.FailOnError(err, "Failed to verify VM state after deleting migration object")

						newNodeName, err := GetNodeOfVM(vm)
						log.FailOnError(err, "Failed to get node of VM %s", vm.Name)
						mu.Lock()
						initialNode := initialNodeName[vmKey]
						mu.Unlock()
						if newNodeName != initialNode {
							log.FailOnError(fmt.Errorf("VM %s migrated unexpectedly to node %s", vm.Name, newNodeName), "VM should remain on node %s", initialNode)
						} else {
							log.Infof("VM %s is still on node %s as expected", vm.Name, initialNode)
						}

						err = CheckVMUptime(vm, initialUptime)
						log.FailOnError(err, "Failed to validate uptime in VM %s", vm.Name)
					}
				}(appCtx)
			}
			wg.Wait()
		})

		ValidateFioInVMs(appCtxs, canSsh)

		stepLog = "Attempt live migration again"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for _, appCtx := range appCtxs {
				wg.Add(1)
				go func(appCtx *scheduler.Context) {
					defer GinkgoRecover()
					defer wg.Done()
					err := StartAndWaitForVMIMigration(appCtx, context1.TODO())
					log.FailOnError(err, "Failed to live migrate KubeVirt VM")
				}(appCtx)
			}
			wg.Wait()
		})

		stepLog = "Validate VMs after successful migration"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			var wg sync.WaitGroup
			for _, appCtx := range appCtxs {
				wg.Add(1)
				go func(appCtx *scheduler.Context) {
					defer GinkgoRecover()
					defer wg.Done()
					vms, err := GetAllVMsFromScheduledContexts([]*scheduler.Context{appCtx})
					log.FailOnError(err, "Failed to get VMs from appCtx")
					for _, vm := range vms {
						vmKey := fmt.Sprintf("%s/%s", vm.Namespace, vm.Name)
						newNodeName, err := GetNodeOfVM(vm)
						log.FailOnError(err, "Failed to get node of VM %s after migration", vm.Name)
						log.Infof("VM %s is now running on node %s after migration", vm.Name, newNodeName)
						mu.Lock()
						initialNode := initialNodeName[vmKey]
						mu.Unlock()
						if newNodeName == initialNode {
							log.FailOnError(fmt.Errorf("VM %s did not migrate to a different node", vm.Name), "VM is still on node %s after migration", initialNode)
						} else {
							log.Infof("VM %s successfully migrated from node %s to node %s", vm.Name, initialNode, newNodeName)
							mu.Lock()
							initialNodeName[vmKey] = newNodeName
							mu.Unlock()
						}

						err = CheckVMUptime(vm, initialUptime)
						log.FailOnError(err, "Failed to validate uptime in VM %s", vm.Name)
					}
				}(appCtx)
			}
			wg.Wait()
		})

		ValidateFioInVMs(appCtxs, canSsh)
		stepLog = "Destroy Applications"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			DestroyApps(appCtxs, nil)
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(appCtxs)
	})
})

var _ = Describe("{RepeatedDeleteMigrationObjectDuringMigration}", Label("p0", "positive", "kubevirt", "LiveMigration"), func() {
	var app, volType string
	var present bool
	JustBeforeEach(func() {
		StartTorpedoTest("RepeatedDeleteMigrationObjectDuringMigration", "Repeatedly delete migration objects during migration and finally migrate VMs", nil, 0)
		volType, present = os.LookupEnv("KUBEVIRT_VOL_TYPE")
		if !present {
			app = "kubevirt-debian-fio-minimal"
		}
		if volType == "pxe-raw" {
			app = "kubevirt-raw-vol"
		} else if volType == "fada-raw" {
			app = "kubevirt-fada-raw-fio"
		}
		log.InfoD("Setting app for this test to be : %s", app)
	})
	var appCtxs []*scheduler.Context
	var namespace string
	var canSsh bool
	var initialUptime map[string]time.Duration
	var initialNodeName map[string]string
	var mu sync.Mutex

	itLog := "Repeatedly delete migration objects during migration and finally migrate VMs"
	It(itLog, func() {
		pxNs, err := Inst().V.GetVolumeDriverNamespace()
		log.FailOnError(err, "Failed to get volume driver namespace")
		defer ListEvents(pxNs)
		canSsh = false

		appList := Inst().AppList
		defer func() {
			Inst().AppList = appList
		}()
		Inst().AppList = []string{app}
		Inst().CsiAppList = []string{app}

		stepLog := "Schedule KubeVirt VMs"
		Step(stepLog, func() {
			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				namespace = fmt.Sprintf("kubevirt-%v", time.Now().Unix())
				appCtxs = append(appCtxs, ScheduleApplicationsOnNamespace(namespace, "test")...)
			}
		})
		ValidateApplications(appCtxs)

		log.Infof("Sleeping for 2 minutes to let VMs come up")
		time.Sleep(2 * time.Minute)

		canSsh = CreateSSHPodAndSetCanSsh()
		ValidateFioInVMs(appCtxs, canSsh)

		initialUptime = make(map[string]time.Duration)
		initialNodeName = make(map[string]string)
		stepLog = "Get initial uptime of VMs and current nodes"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			var wg sync.WaitGroup
			for _, appCtx := range appCtxs {
				wg.Add(1)
				go func(appCtx *scheduler.Context) {
					defer GinkgoRecover()
					defer wg.Done()
					vms, err := GetAllVMsFromScheduledContexts([]*scheduler.Context{appCtx})
					log.FailOnError(err, "Failed to get VMs from appCtx")
					for _, vm := range vms {
						vmKey := fmt.Sprintf("%s/%s", vm.Namespace, vm.Name)
						uptime, err := GetVMUptime(vm)
						log.FailOnError(err, "Failed to get uptime from VM %s", vm.Name)
						log.Infof("Initial uptime for VM %s is %v", vmKey, uptime)

						nodeName, err := GetNodeOfVM(vm)
						log.FailOnError(err, "Failed to get node of VM %v", vm.Name)
						log.Infof("VM %s is currently running on node %s", vm.Name, nodeName)

						mu.Lock()
						initialUptime[vmKey] = uptime
						initialNodeName[vmKey] = nodeName
						mu.Unlock()
					}
				}(appCtx)
			}
			wg.Wait()
		})

		for i := 0; i < 5; i++ {
			var wg sync.WaitGroup
			migrationMap := make(map[string]*kubevirtdy.VirtualMachineInstanceMigration)
			stepLog = fmt.Sprintf("Iteration %d: Start live migration of the KubeVirt VMs", i+1)
			Step(stepLog, func() {
				log.InfoD(stepLog)
				for _, appCtx := range appCtxs {
					wg.Add(1)
					go func(appCtx *scheduler.Context) {
						defer GinkgoRecover()
						defer wg.Done()
						vms, err := GetAllVMsFromScheduledContexts([]*scheduler.Context{appCtx})
						log.FailOnError(err, "Failed to get VMs from appCtx")
						for _, vm := range vms {
							migration, err := kubevirtdy.Instance().CreateVirtualMachineInstanceMigration(context1.TODO(), vm.Namespace, vm.Name)
							log.FailOnError(err, "Failed to create VM migration for VM %s", vm.Name)
							log.Infof("Migration %s created for VM %s", migration.Name, vm.Name)
							vmKey := fmt.Sprintf("%s/%s", vm.Namespace, vm.Name)
							mu.Lock()
							migrationMap[vmKey] = migration
							mu.Unlock()
						}
					}(appCtx)
				}
				wg.Wait()
			})

			stepLog = fmt.Sprintf("Iteration %d: Delete migration objects during migration", i+1)
			Step(stepLog, func() {
				log.InfoD(stepLog)
				var wg sync.WaitGroup
				for vmKey, migration := range migrationMap {
					wg.Add(1)
					go func(vmKey string, migration *kubevirtdy.VirtualMachineInstanceMigration) {
						defer GinkgoRecover()
						defer wg.Done()
						vmNamespace, vmName := migration.NameSpace, migration.VMIName

						// Wait for migration to be in progress
						t := func() (interface{}, bool, error) {
							migr, err := kubevirtdy.Instance().GetVirtualMachineInstanceMigration(context1.TODO(), vmNamespace, migration.Name)
							if err != nil {
								log.Infof("Error getting migration: %v", err)
								return nil, true, err
							}
							phase := migr.Phase
							if phase == string(kubevirtv1.MigrationRunning) {
								log.Infof("Migration %s for VM %s is in progress", migration.Name, vmName)
								return migr, false, nil
							} else if phase == string(kubevirtv1.MigrationSucceeded) || phase == string(kubevirtv1.MigrationFailed) {
								log.Infof("Migration %s for VM %s has already completed with phase %s", migration.Name, vmName, phase)
								return migr, false, nil
							}
							return nil, true, fmt.Errorf("Migration %s for VM %s not yet in progress, current phase: %s", migration.Name, vmName, phase)
						}
						_, err := task.DoRetryWithTimeout(t, triggerCheckTimeout, triggerCheckInterval)
						log.FailOnError(err, "Failed to wait for migration to be in progress")

						// Delete the migration object using dynamic client
						dynamicClient, err := GetDynamicKubeClient()
						log.FailOnError(err, "Failed to get dynamic client")

						vmimGVR := schema.GroupVersionResource{
							Group:    "kubevirt.io",
							Version:  "v1",
							Resource: "virtualmachineinstancemigrations",
						}

						err = dynamicClient.Resource(vmimGVR).Namespace(vmNamespace).Delete(context1.TODO(), migration.Name, metav1.DeleteOptions{})
						log.FailOnError(err, "Failed to delete migration %s for VM %s", migration.Name, vmName)
						log.Infof("Deleted migration %s for VM %s", migration.Name, vmName)
					}(vmKey, migration)
				}
				wg.Wait()
			})

			stepLog = fmt.Sprintf("Iteration %d: Validate VMs after deleting migration objects", i+1)
			Step(stepLog, func() {
				log.InfoD(stepLog)
				var wg sync.WaitGroup
				for _, appCtx := range appCtxs {
					wg.Add(1)
					go func(appCtx *scheduler.Context) {
						defer GinkgoRecover()
						defer wg.Done()
						vms, err := GetAllVMsFromScheduledContexts([]*scheduler.Context{appCtx})
						log.FailOnError(err, "Failed to get VMs from appCtx")
						for _, vm := range vms {
							vmKey := fmt.Sprintf("%s/%s", vm.Namespace, vm.Name)
							err := CheckVMState(vm)
							log.FailOnError(err, "Failed to verify VM state after deleting migration object")

							newNodeName, err := GetNodeOfVM(vm)
							log.FailOnError(err, "Failed to get node of VM %s", vm.Name)
							mu.Lock()
							initialNode := initialNodeName[vmKey]
							mu.Unlock()
							if newNodeName != initialNode {
								log.FailOnError(fmt.Errorf("VM %s migrated unexpectedly to node %s", vm.Name, newNodeName), "VM should remain on node %s", initialNode)
							} else {
								log.Infof("VM %s is still on node %s as expected", vm.Name, initialNode)
							}

							err = CheckVMUptime(vm, initialUptime)
							log.FailOnError(err, "Failed to validate uptime in VM %s", vm.Name)
						}
					}(appCtx)
				}
				wg.Wait()
			})
			log.Infof("Sleeping for 20 seconds between each migration iteration")
			time.Sleep(20 * time.Second)
		}

		stepLog = "Attempt live migration after 5 iterations"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			var wg sync.WaitGroup
			for _, appCtx := range appCtxs {
				wg.Add(1)
				go func(appCtx *scheduler.Context) {
					defer GinkgoRecover()
					defer wg.Done()
					err := StartAndWaitForVMIMigration(appCtx, context1.TODO())
					log.FailOnError(err, "Failed to live migrate KubeVirt VM")
				}(appCtx)
			}
			wg.Wait()
		})

		stepLog = "Validate VMs after successful migration"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			var wg sync.WaitGroup
			for _, appCtx := range appCtxs {
				wg.Add(1)
				go func(appCtx *scheduler.Context) {
					defer GinkgoRecover()
					defer wg.Done()
					vms, err := GetAllVMsFromScheduledContexts([]*scheduler.Context{appCtx})
					log.FailOnError(err, "Failed to get VMs from appCtx")
					for _, vm := range vms {
						vmKey := fmt.Sprintf("%s/%s", vm.Namespace, vm.Name)
						newNodeName, err := GetNodeOfVM(vm)
						log.FailOnError(err, "Failed to get node of VM %s after migration", vm.Name)
						log.Infof("VM %s is now running on node %s after migration", vm.Name, newNodeName)
						mu.Lock()
						initialNode := initialNodeName[vmKey]
						mu.Unlock()
						if newNodeName == initialNode {
							log.FailOnError(fmt.Errorf("VM %s did not migrate to a different node", vm.Name), "VM is still on node %s after migration", initialNode)
						} else {
							log.Infof("VM %s successfully migrated from node %s to node %s", vm.Name, initialNode, newNodeName)
							mu.Lock()
							initialNodeName[vmKey] = newNodeName
							mu.Unlock()
						}

						err = CheckVMUptime(vm, initialUptime)
						log.FailOnError(err, "Failed to validate uptime in VM %s", vm.Name)
					}
				}(appCtx)
			}
			wg.Wait()
		})

		ValidateFioInVMs(appCtxs, canSsh)

		stepLog = "Destroy Applications"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			DestroyApps(appCtxs, nil)
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(appCtxs)
	})
})

var _ = Describe("{AddNewMixedDiskToKubevirtVMAndLM}", Label("p0", "positive", "kubevirt"), func() {
	var app, volType string
	var present bool
	JustBeforeEach(func() {
		StartTorpedoTest("AddNewMixedDiskToKubevirtVMAndLM", "Add a new sv4 disk to a raw block kubevirtVM and then do LM", nil, 0)
		volType, present = os.LookupEnv("KUBEVIRT_VOL_TYPE")
		if !present {
			app = "kubevirt-debian-fio-minimal"
		}
		if volType == "pxe-raw" {
			app = "kubevirt-raw-vol"
		} else if volType == "fada-raw" {
			app = "kubevirt-fada-raw-fio"
		}
		log.InfoD("Setting app for this test to be : %s", app)
	})
	var appCtxs []*scheduler.Context
	var namespace string
	var canSsh bool
	var initialUptime map[string]time.Duration
	var mu sync.Mutex
	var wg sync.WaitGroup
	var initialNodeName map[string]string

	itLog := "Add a new fada disk to a kubevirtVM"
	It(itLog, func() {
		pxNs, err := Inst().V.GetVolumeDriverNamespace()
		log.FailOnError(err, "Failed to get volume driver namespace")
		defer ListEvents(pxNs)

		appList := Inst().AppList
		defer func() {
			Inst().AppList = appList
		}()
		numberOfVolumes := 1

		Inst().AppList = []string{app}
		Inst().CsiAppList = []string{app}
		stepLog := "Schedule a kubevirt VM"
		Step(stepLog, func() {
			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				namespace = fmt.Sprintf("kubevirt-%v", time.Now().Unix())
				appCtxs = append(appCtxs, ScheduleApplicationsOnNamespace(namespace, "test")...)
			}
		})
		ValidateApplications(appCtxs)
		if !present {
			for _, appCtx := range appCtxs {
				bindMount, err := IsVMBindMounted(appCtx, false)
				log.FailOnError(err, "Failed to verify bind mount")
				dash.VerifyFatal(bindMount, true, "Failed to verify bind mount")
			}
		}

		log.Infof("Sleeping for 2 minutes to let VMs come up")
		time.Sleep(2 * time.Minute)

		canSsh = CreateSSHPodAndSetCanSsh()
		ValidateFioInVMs(appCtxs, canSsh)

		if !present {
			stepLog = "Add one raw block disk to the kubevirt VM"
			Step(stepLog, func() {
				log.InfoD(stepLog)
				_, err := AddRawBlockDriveToKubevirtVM(appCtxs, numberOfVolumes, "10Gi")
				log.FailOnError(err, "Failed to add disks to kubevirt VM")
				dash.VerifyFatal(true, true, "Disk added to kubevirt VM")
			})
		} else {
			stepLog = "Add one sv4 svc disk to the kubevirt VM having raw disks"
			Step(stepLog, func() {
				log.InfoD(stepLog)
				for _, appCtx := range appCtxs {
					vms, err := GetAllVMsFromScheduledContexts([]*scheduler.Context{appCtx})
					log.FailOnError(err, "Failed to get all VMs contexts")
					for _, v := range vms {
						initialDiskCount, err := GetNumberOfDrivesInVM(v)
						log.FailOnError(err, fmt.Sprintf("failed to get initial number of disks in VM [%s]: %v", v.Name, err))
						log.Infof("Initial number of disks in VM [%s]: %d", v.Name, initialDiskCount)
						pvcs, err := CreatePVCsForVM(v, 1, "sc-sharedv4svc", "10Gi")
						log.FailOnError(err, "Failed to create sv4 pvc")
						specListInterfaces := make([]interface{}, len(pvcs))
						for i, pvc := range pvcs {
							specListInterfaces[i] = pvc
						}
						appCtx.App.SpecList = append(appCtx.App.SpecList, specListInterfaces...)

						err = AddPVCsToVirtualMachine(v, pvcs)
						log.FailOnError(err, "Failed to add pvc to VM")
						err = RestartKubevirtVM(v.Name, v.Namespace, true)
						log.FailOnError(err, "Failed to restart VM after addition of PVC")
						err = WaitForVMToBeReady(v.Name, v.Namespace)
						log.FailOnError(err, "VM Did not come up")
						expectedDiskCount := initialDiskCount + 1
						t := func() (interface{}, bool, error) {
							newDiskCount, err := GetNumberOfDrivesInVM(v)
							if err != nil {
								log.Warnf("Failed to get number of disks in VM [%s]: %v", v.Name, err)
								return nil, true, err
							}
							if newDiskCount != expectedDiskCount {
								err := fmt.Errorf("number of disks in VM [%s] is %d; expected %d", v.Name, newDiskCount, expectedDiskCount)
								log.Warnf(err.Error())
								return nil, true, err
							}
							return newDiskCount, false, nil
						}
						_, err = task.DoRetryWithTimeout(t, 5*time.Minute, 30*time.Second)
						log.FailOnError(err, fmt.Sprintf("VM in namespace %v did not come up after addition of disk", v.Namespace))
						log.Infof("Successfully verified number of disks in VM [%s] in namespace %s", v.Name, v.Namespace)
					}
				}
			})
		}
		ValidateFioInVMs(appCtxs, canSsh)

		initialUptime = make(map[string]time.Duration)
		initialNodeName = make(map[string]string)
		stepLog = "Get initial uptime of VMs and current nodes"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			var wg sync.WaitGroup
			for _, appCtx := range appCtxs {
				wg.Add(1)
				go func(appCtx *scheduler.Context) {
					defer GinkgoRecover()
					defer wg.Done()
					vms, err := GetAllVMsFromScheduledContexts([]*scheduler.Context{appCtx})
					log.FailOnError(err, "Failed to get VMs from appCtx")
					for _, vm := range vms {
						vmKey := fmt.Sprintf("%s/%s", vm.Namespace, vm.Name)
						uptime, err := GetVMUptime(vm)
						log.FailOnError(err, "Failed to get uptime from VM %s", vm.Name)
						log.Infof("Initial uptime for VM %s is %v", vmKey, uptime)

						nodeName, err := GetNodeOfVM(vm)
						log.FailOnError(err, "Failed to get node of VM %v", vm.Name)
						log.Infof("VM %s is currently running on node %s", vm.Name, nodeName)

						mu.Lock()
						initialUptime[vmKey] = uptime
						initialNodeName[vmKey] = nodeName
						mu.Unlock()
					}
				}(appCtx)
			}
			wg.Wait()
		})

		wg = sync.WaitGroup{}
		stepLog = "Live migrate the kubevirt VMs"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for _, appCtx := range appCtxs {
				wg.Add(1)
				go func(appCtx *scheduler.Context) {
					defer GinkgoRecover()
					defer wg.Done()
					err := StartAndWaitForVMIMigration(appCtx, context1.TODO())
					log.FailOnError(err, "Failed to live migrate kubevirt VM")
				}(appCtx)
			}
			wg.Wait()
		})

		stepLog = "Validate VMs after migration"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			var wg sync.WaitGroup
			for _, appCtx := range appCtxs {
				wg.Add(1)
				go func(appCtx *scheduler.Context) {
					defer GinkgoRecover()
					defer wg.Done()
					vms, err := GetAllVMsFromScheduledContexts([]*scheduler.Context{appCtx})
					log.FailOnError(err, "Failed to get VMs from appCtx")
					for _, vm := range vms {
						vmKey := fmt.Sprintf("%s/%s", vm.Namespace, vm.Name)
						newNodeName, err := GetNodeOfVM(vm)
						log.FailOnError(err, "Failed to get node of VM %s after migration", vm.Name)
						log.Infof("VM %s is now running on node %s after migration", vm.Name, newNodeName)
						mu.Lock()
						initialNode := initialNodeName[vmKey]
						mu.Unlock()
						if newNodeName == initialNode {
							log.FailOnError(fmt.Errorf("VM %s did not migrate to a different node", vm.Name), "VM is still on node %s after migration", initialNode)
						} else {
							log.Infof("VM %s successfully migrated from node %s to node %s", vm.Name, initialNode, newNodeName)
						}

						err = CheckVMUptime(vm, initialUptime)
						log.FailOnError(err, "Failed to validate uptime in VM %s", vm.Name)
					}
				}(appCtx)
			}
			wg.Wait()
		})
		ValidateFioInVMs(appCtxs, canSsh)
		stepLog = "Destroy Applications"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			DestroyApps(appCtxs, nil)
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(appCtxs)
	})
})

var _ = Describe("{ResizePvcAndLiveMigrateVMs}", Label("p0", "positive", "kubevirt", "LiveMigration", "Resize"), func() {
	var app, volType string
	var present bool
	JustBeforeEach(func() {
		StartTorpedoTest("ResizePvcAndLiveMigrateVMs", "Resize PVCs attached to VMs, live migrate them, and add new disks in a loop", nil, 0)
		volType, present = os.LookupEnv("KUBEVIRT_VOL_TYPE")
		if !present {
			app = "kubevirt-debian-fio-minimal"
		}
		if volType == "pxe-raw" {
			app = "kubevirt-raw-vol"
		} else if volType == "fada-raw" {
			app = "kubevirt-fada-raw-fio"
		}
		log.InfoD("Setting app for this test to be : %s", app)
	})

	var appCtxs []*scheduler.Context
	var namespace string
	var canSsh bool
	var initialUptime map[string]time.Duration
	var initialNodeName map[string]string
	var mu sync.Mutex
	type PVCDetails struct {
		pvc *v1.PersistentVolumeClaim
		vm  kubevirtv1.VirtualMachine
	}

	itLog := "Resize PVCs attached to VMs, live migrate them, and add new disks in a loop"
	It(itLog, func() {
		pxNs, err := Inst().V.GetVolumeDriverNamespace()
		log.FailOnError(err, "Failed to get volume driver namespace")
		defer ListEvents(pxNs)
		canSsh = false

		appList := Inst().AppList
		defer func() {
			Inst().AppList = appList
		}()
		Inst().AppList = []string{app}
		Inst().CsiAppList = []string{app}

		stepLog := "Schedule KubeVirt VMs"
		Step(stepLog, func() {
			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				namespace = fmt.Sprintf("kubevirt-%v", time.Now().Unix())
				appCtxs = append(appCtxs, ScheduleApplicationsOnNamespace(namespace, "test")...)
			}
		})
		ValidateApplications(appCtxs)

		log.Infof("Sleeping for 2 minutes to let VMs come up")
		time.Sleep(2 * time.Minute)

		canSsh = CreateSSHPodAndSetCanSsh()
		ValidateFioInVMs(appCtxs, canSsh)

		initialUptime = make(map[string]time.Duration)
		initialNodeName = make(map[string]string)
		stepLog = "Get initial uptime of VMs and current nodes"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			var wg sync.WaitGroup
			for _, appCtx := range appCtxs {
				wg.Add(1)
				go func(appCtx *scheduler.Context) {
					defer GinkgoRecover()
					defer wg.Done()
					vms, err := GetAllVMsFromScheduledContexts([]*scheduler.Context{appCtx})
					log.FailOnError(err, "Failed to get VMs from appCtx")
					for _, vm := range vms {
						vmKey := fmt.Sprintf("%s/%s", vm.Namespace, vm.Name)
						uptime, err := GetVMUptime(vm)
						log.FailOnError(err, "Failed to get uptime from VM %s", vm.Name)
						log.Infof("Initial uptime for VM %s is %v", vmKey, uptime)

						nodeName, err := GetNodeOfVM(vm)
						log.FailOnError(err, "Failed to get node of VM %v", vm.Name)
						log.Infof("VM %s is currently running on node %s", vm.Name, nodeName)

						mu.Lock()
						initialUptime[vmKey] = uptime
						initialNodeName[vmKey] = nodeName
						mu.Unlock()
					}
				}(appCtx)
			}
			wg.Wait()
		})

		for i := 0; i < 5; i++ {
			var pvcDetailsList []*PVCDetails
			stepLog = fmt.Sprintf("Iteration %d: Collect PVCs attached to VMs", i+1)
			Step(stepLog, func() {
				log.InfoD(stepLog)
				var wg sync.WaitGroup
				for _, appCtx := range appCtxs {
					wg.Add(1)
					go func(appCtx *scheduler.Context) {
						defer GinkgoRecover()
						defer wg.Done()
						vms, err := GetAllVMsFromScheduledContexts([]*scheduler.Context{appCtx})
						log.FailOnError(err, "Failed to get VMs from appCtx")
						for _, vm := range vms {
							pvcNames := GetPVCsAttachedToVM(vm)
							for _, pvcName := range pvcNames {
								pvc, err := k8sCore.GetPersistentVolumeClaim(vm.Namespace, pvcName)
								log.FailOnError(err, "Failed to get PVC %s", pvcName)
								mu.Lock()
								pvcDetailsList = append(pvcDetailsList, &PVCDetails{
									pvc: pvc,
									vm:  vm,
								})
								mu.Unlock()
							}
						}
					}(appCtx)
				}
				wg.Wait()
			})

			stepLog = fmt.Sprintf("Iteration %d: Resize PVCs attached to VMs", i+1)
			Step(stepLog, func() {
				log.InfoD(stepLog)
				var wg sync.WaitGroup
				for _, pvcDetail := range pvcDetailsList {
					wg.Add(1)
					go func(pvcDetail *PVCDetails) {
						defer GinkgoRecover()
						defer wg.Done()
						pvc := pvcDetail.pvc
						currentSize := pvc.Spec.Resources.Requests[v1.ResourceStorage]
						newSize := currentSize.DeepCopy()
						newSize.Add(resource.MustParse("10Gi"))
						pvc.Spec.Resources.Requests[v1.ResourceStorage] = newSize

						log.Infof("Resizing PVC %s/%s from %v to %v", pvc.Namespace, pvc.Name, currentSize.String(), newSize.String())

						updatedPVC, err := k8sCore.UpdatePersistentVolumeClaim(pvc)
						log.FailOnError(err, "Failed to update PVC %s/%s", pvc.Namespace, pvc.Name)

						t := func() (interface{}, bool, error) {
							updatedPVC, err = k8sCore.GetPersistentVolumeClaim(pvc.Namespace, pvc.Name)
							if err != nil {
								return nil, true, err
							}
							capacity := updatedPVC.Status.Capacity[v1.ResourceStorage]
							if capacity.Cmp(newSize) >= 0 {
								log.Infof("PVC %s/%s resized successfully to %v", pvc.Namespace, pvc.Name, capacity.String())
								return nil, false, nil
							}
							return nil, true, fmt.Errorf("PVC %s/%s not yet resized. Current size: %v", pvc.Namespace, pvc.Name, capacity.String())
						}
						_, err = task.DoRetryWithTimeout(t, 5*time.Minute, 10*time.Second)
						log.FailOnError(err, "Failed to wait for PVC %s/%s to be resized", pvc.Namespace, pvc.Name)
					}(pvcDetail)
				}
				wg.Wait()
			})

			stepLog = fmt.Sprintf("Iteration %d: Validate VMs after resizing PVCs", i+1)
			Step(stepLog, func() {
				log.InfoD(stepLog)
				ValidateVMUptime(appCtxs, canSsh, initialUptime)
			})
			ValidateFioInVMs(appCtxs, canSsh)

			stepLog = fmt.Sprintf("Iteration %d: Live migrate VMs", i+1)
			Step(stepLog, func() {
				log.InfoD(stepLog)
				var wg sync.WaitGroup
				for _, appCtx := range appCtxs {
					wg.Add(1)
					go func(appCtx *scheduler.Context) {
						defer GinkgoRecover()
						defer wg.Done()
						err := StartAndWaitForVMIMigration(appCtx, context1.TODO())
						log.FailOnError(err, "Failed to live migrate KubeVirt VM")
					}(appCtx)
				}
				wg.Wait()
			})

			stepLog = fmt.Sprintf("Iteration %d: Validate VMs after live migration", i+1)
			Step(stepLog, func() {
				log.InfoD(stepLog)
				var wg sync.WaitGroup
				for _, appCtx := range appCtxs {
					wg.Add(1)
					go func(appCtx *scheduler.Context) {
						defer GinkgoRecover()
						defer wg.Done()
						vms, err := GetAllVMsFromScheduledContexts([]*scheduler.Context{appCtx})
						log.FailOnError(err, "Failed to get VMs from appCtx")
						for _, vm := range vms {
							vmKey := fmt.Sprintf("%s/%s", vm.Namespace, vm.Name)
							newNodeName, err := GetNodeOfVM(vm)
							log.FailOnError(err, "Failed to get node of VM %s after migration", vm.Name)
							log.Infof("VM %s is now running on node %s after migration", vm.Name, newNodeName)
							mu.Lock()
							initialNode := initialNodeName[vmKey]
							mu.Unlock()
							if newNodeName == initialNode {
								log.FailOnError(fmt.Errorf("VM %s did not migrate to a different node", vm.Name), "VM is still on node %s after migration", initialNode)
							} else {
								log.Infof("VM %s successfully migrated from node %s to node %s", vm.Name, initialNode, newNodeName)
								mu.Lock()
								initialNodeName[vmKey] = newNodeName
								mu.Unlock()
							}

							err = CheckVMUptime(vm, initialUptime)
							log.FailOnError(err, "Failed to validate uptime in VM %s", vm.Name)
						}
					}(appCtx)
				}
				wg.Wait()
			})
			ValidateFioInVMs(appCtxs, canSsh)

			stepLog = fmt.Sprintf("Iteration %d: Add new disk to VMs", i+1)
			Step(stepLog, func() {
				log.InfoD(stepLog)
				numberOfVolumes := 1
				_, err := AddRawBlockDriveToKubevirtVM(appCtxs, numberOfVolumes, "10Gi")
				log.FailOnError(err, "Failed to add disks to KubeVirt VMs")
				dash.VerifyFatal(true, true, "Disk added to KubeVirt VMs")
			})

			stepLog = fmt.Sprintf("Iteration %d: Validate VMs after adding new disk", i+1)
			Step(stepLog, func() {
				log.InfoD(stepLog)
				ValidateVMUptime(appCtxs, canSsh, initialUptime)
			})
			ValidateFioInVMs(appCtxs, canSsh)
		}

		stepLog = "Destroy Applications"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			DestroyApps(appCtxs, nil)
		})
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(appCtxs)
	})
})

var _ = Describe("{AddNewHotPlugDiskToKubevirtVM}", Label("p0", "positive", "kubevirt"), func() {
	var (
		app, volType  string
		present       bool
		appCtxs       []*scheduler.Context
		namespace     string
		canSsh        bool
		volumeMode    string
		initialUptime map[string]time.Duration
		vmNodeName    string
	)
	JustBeforeEach(func() {
		StartTorpedoTest("AddNewHotPlugDiskToKubevirtVM", "Add a new raw block disk to a running KubeVirt VM via hot-plug", nil, 0)
		volType, present = os.LookupEnv("KUBEVIRT_VOL_TYPE")
		if !present {
			app = "kubevirt-debian-fio-minimal"
		}
		if volType == "pxe-raw" {
			app = "kubevirt-raw-vol"
		} else if volType == "fada-raw" {
			app = "kubevirt-fada-raw-fio"
		} else {
			app = "kubevirt-debian-fio-minimal"
		}
		log.InfoD("Setting app for this test to be : %s", app)
	})

	It("hot-plug a new disk to a running KubeVirt VM", func() {
		pxNs, err := Inst().V.GetVolumeDriverNamespace()
		log.FailOnError(err, "Failed to get volume driver namespace")
		defer ListEvents(pxNs)

		appList := Inst().AppList
		defer func() {
			Inst().AppList = appList
		}()

		numberOfVolumes := 1

		Inst().AppList = []string{app}
		Inst().CsiAppList = []string{app}

		stepLog := "Schedule a kubevirt VM"
		Step(stepLog, func() {
			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				namespace = fmt.Sprintf("kubevirt-%v", time.Now().Unix())
				appCtxs = append(appCtxs, ScheduleApplicationsOnNamespace(namespace, "test")...)
			}
		})
		ValidateApplications(appCtxs)

		if !present {
			for _, appCtx := range appCtxs {
				bindMount, err := IsVMBindMounted(appCtx, false)
				log.FailOnError(err, "Failed to verify bind mount")
				dash.VerifyFatal(bindMount, true, "Failed to verify bind mount")
			}
		}
		
		if app == "kubevirt-debian-fio-minimal" {
			volumeMode = ""
		} else {
			volumeMode = "Block"
		}

		log.Infof("Sleeping for 2 minutes to let VMs come up fully")
		time.Sleep(2 * time.Minute)

		canSsh = CreateSSHPodAndSetCanSsh()
		ValidateFioInVMs(appCtxs, canSsh)
		initialUptime = make(map[string]time.Duration)
		stepLog = "Get initial uptime of VMs and current node"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			var wg sync.WaitGroup
			for _, appCtx := range appCtxs {
				wg.Add(1)
				go func(appCtx *scheduler.Context) {
					defer GinkgoRecover()
					defer wg.Done()
					vms, err := GetAllVMsFromScheduledContexts([]*scheduler.Context{appCtx})
					log.FailOnError(err, "Failed to get VMs from appCtx")
					for _, vm := range vms {
						uptime, err := GetVMUptime(vm)
						log.FailOnError(err, "Failed to get uptime from VM %s", vm.Name)
						vmKey := fmt.Sprintf("%s/%s", vm.Namespace, vm.Name)
						initialUptime[vmKey] = uptime
						log.Infof("Initial uptime for VM %s is %v", vmKey, uptime)

						vmNodeName, err = GetNodeOfVM(vm)
						log.FailOnError(err, "Failed to get node of VM %v", vm.Name)
						log.Infof("VM %s is currently running on node %s", vm.Name, vmNodeName)
					}
				}(appCtx)
			}
			wg.Wait()
		})

		stepLog = "Hot-plug one raw disk (DataVolume) to the running KubeVirt VM"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			_, err := HotPlugDataVolumesToKubevirtVM(appCtxs, numberOfVolumes, "50Gi", volumeMode)
			log.FailOnError(err, "Failed to hot-plug DataVolume to KubeVirt VM")
			dash.VerifyFatal(true, true, "DataVolume hot-plugged to KubeVirt VM")
		})

		ValidateFioInVMs(appCtxs, canSsh)
		ValidateVMUptime(appCtxs, canSsh, initialUptime)

		stepLog = "Destroy Applications"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			DestroyApps(appCtxs, nil)
		})
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(appCtxs)
	})
})

var _ = Describe("{RebootNodeAfterAddNewHotPlugDiskToKubevirtVM}", Label("p1", "negative", "kubevirt", "node_reboot", "staging"), func() {
	/*
		Step 1: Create a VM
		Step 2: Add Hot plug disk to a VM
		Step 3: Add hot plug disk and then reboot the VM node.
		Step 4: Validate VM moves out of the node and hot plug volume is added inside the disk.

		JIRA ID: https://purestorage.atlassian.net/browse/HAZEL-1688
	*/
	var (
		app, volType    string
		present         bool
		appCtxs         []*scheduler.Context
		namespace       string
		canSsh          bool
		volumeMode      string
		initialUptime   map[string]time.Duration
		initialNodeName map[string]string
		vmNodeName      string
		bindMount       bool
		vmDiskCount     map[string]int
		newDiskCount    int
	)
	JustBeforeEach(func() {
		StartTorpedoTest("RebootNodeAfterAddNewHotPlugDiskToKubevirtVM", "Add a new raw disk to a running KubeVirt VM via hot-plug", nil, 0)
		volType, present = os.LookupEnv("KUBEVIRT_VOL_TYPE")
		if !present {
			app = "kubevirt-debian-fio-minimal"
		}
		if volType == "pxe-raw" {
			app = "kubevirt-raw-vol"
		} else if volType == "fada-raw" {
			app = "kubevirt-fada-raw-fio"
		} else {
			app = "kubevirt-debian-fio-minimal"
		}
		log.InfoD("Setting app for this test to be : %s", app)
	})

	It("hot-plug a new disk to a running KubeVirt VM", func() {
		pxNs, err := Inst().V.GetVolumeDriverNamespace()
		log.FailOnError(err, "Failed to get volume driver namespace")
		defer ListEvents(pxNs)

		appList := Inst().AppList
		defer func() {
			Inst().AppList = appList
		}()

		numberOfVolumes := 1

		Inst().AppList = []string{app}
		Inst().CsiAppList = []string{app}

		stepLog := "Schedule a kubevirt VM"
		Step(stepLog, func() {
			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				namespace = fmt.Sprintf("kubevirt-%v", time.Now().Unix())
				appCtxs = append(appCtxs, ScheduleApplicationsOnNamespace(namespace, "test")...)
			}
		})
		ValidateApplications(appCtxs)

		if !present {
			for _, appCtx := range appCtxs {
				bindMount, err = IsVMBindMounted(appCtx, false)
				log.FailOnError(err, "Failed to verify bind mount")
				dash.VerifyFatal(bindMount, true, "VM bind mount verified")
			}
		}

		if app == "kubevirt-debian-fio-minimal" {
			volumeMode = ""
		} else {
			volumeMode = "Block"
		}

		log.Infof("Sleeping for 2 minutes to let VMs come up fully")
		time.Sleep(2 * time.Minute)

		canSsh = CreateSSHPodAndSetCanSsh()
		ValidateFioInVMs(appCtxs, canSsh)
		initialUptime = make(map[string]time.Duration)
		stepLog = "Get initial uptime of VMs and current node"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			var wg sync.WaitGroup
			for _, appCtx := range appCtxs {
				wg.Add(1)
				go func(appCtx *scheduler.Context) {
					defer GinkgoRecover()
					defer wg.Done()
					vms, err := GetAllVMsFromScheduledContexts([]*scheduler.Context{appCtx})
					log.FailOnError(err, "Failed to get VMs from appCtx")
					for _, vm := range vms {
						uptime, err := GetVMUptime(vm)
						log.FailOnError(err, "Failed to get uptime from VM %s", vm.Name)
						vmKey := fmt.Sprintf("%s/%s", vm.Namespace, vm.Name)
						initialUptime[vmKey] = uptime
						log.Infof("Initial uptime for VM %s is %v", vmKey, uptime)

						vmNodeName, err = GetNodeOfVM(vm)
						log.FailOnError(err, "Failed to get node of VM %v", vm.Name)
						log.Infof("VM %s is currently running on node %s", vm.Name, vmNodeName)
					}
				}(appCtx)
			}
			wg.Wait()
		})

		stepLog = "Hot-plug one raw disk (DataVolume) to the running KubeVirt VM"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			hotPlugDisk, err := HotPlugDataVolumesToKubevirtVM(appCtxs, numberOfVolumes, "50Gi", volumeMode)
			log.FailOnError(err, "Failed to hot-plug DataVolume to KubeVirt VM")
			dash.VerifyFatal(hotPlugDisk, true, "DataVolume hot-plugged to KubeVirt VM")
		})

		initialNodeName = make(map[string]string)
		vmDiskCount = make(map[string]int)
		stepLog = "Reboot the VM node"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for _, virtualMachineCtx := range appCtxs {
				vms, err := GetAllVMsFromScheduledContexts([]*scheduler.Context{virtualMachineCtx})
				log.FailOnError(err, "Failed to get VMs from scheduled contexts")
				dash.VerifyFatal(len(vms) > 0, true, "Failed to to get VMs from scheduled contexts")

				for _, vm := range vms {
					nodeName, err := GetNodeOfVM(vm)
					log.FailOnError(err, "Failed to get node name for VM: %s", vm.Name)
					initialNodeName[vm.Name] = nodeName
					log.Infof("Pre-reboot VM [%s] in namespace [%s] is scheduled on node [%s]. Rebooting it.", vm.Name, vm.Namespace, nodeName)
					newDiskCount, err = GetNumberOfDrivesInVM(vm)
					log.FailOnError(err, "Failed to get disk count of vm %v", vm.Name)
					vmDiskCount[vm.Name] = newDiskCount
					nodeObj, err := node.GetNodeByName(nodeName)
					log.FailOnError(err, "Failed to get node obj for node name: %s", nodeName)
					err = Inst().N.RebootNodeAndWait(nodeObj)
					log.FailOnError(err, "Failed to reboot node: %s", nodeObj.Name)
					log.Infof("Succesfully rebooted node: %s", nodeObj.Name)
				}
				ValidateApplications(appCtxs)
				log.Infof("Sleeping for 2 minutes to stabilize")
				time.Sleep(2 * time.Minute)
				vms, err = GetAllVMsFromScheduledContexts([]*scheduler.Context{virtualMachineCtx})
				log.FailOnError(err, "Failed to get VMs from scheduled contexts")
				dash.VerifyFatal(len(vms) > 0, true, "VMs from scheduled contexts")
				for _, vm := range vms {
					nodeName, err := GetNodeOfVM(vm)
					log.FailOnError(err, "Failed to get node name for VM: %s", vm.Name)
					dash.VerifyFatal(nodeName != initialNodeName[vm.Name], true, "VM moved out of Node?")
					log.Infof("Post reboot VM [%s] in namespace [%s] is scheduled on node [%s]", vm.Name, vm.Namespace, nodeName)
				}
			}
			ValidateApplications(appCtxs)
		})

		ValidateFioInVMs(appCtxs, canSsh)

		stepLog = "Validate hot plug volume inside VM"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for _, appCtx := range appCtxs {
				vms, err := GetAllVMsFromScheduledContexts([]*scheduler.Context{appCtx})
				log.FailOnError(err, "Failed to get VMs from context")
				for _, vm := range vms {
					diskCountAfterNodeReboot, err := GetNumberOfDrivesInVM(vm)
					log.FailOnError(err, "Failed to get disk count of vm %v", vm.Name)
					dash.VerifyFatal(diskCountAfterNodeReboot == vmDiskCount[vm.Name], true, "Validate the number of disk same after Node reboot to ensure hot plug disk is present")
				}
			}
		})

		stepLog = "Destroy Applications"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			DestroyApps(appCtxs, nil)
		})
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(appCtxs)
	})
})

var _ = Describe("{PxRestartAfterAddNewHotPlugDiskToKubevirtVM}", Label("p1", "negative", "kubevirt", "px_crash", "staging"), func() {
	/*
		Step 1: Create a VM
		Step 2: Add Hot plug disk to a VM
		Step 3: Add hot plug volume and then kill Px on source node
		Step 4: Validate hot plug volume inside VM

		JIRA ID: https://purestorage.atlassian.net/browse/HAZEL-1687
	*/
	var (
		app, volType  string
		present       bool
		appCtxs       []*scheduler.Context
		namespace     string
		canSsh        bool
		volumeMode    string
		initialUptime map[string]time.Duration
		vmNodeName    string
		bindMount     bool
		nodes         []string
		vmDiskCount   map[string]int
		newDiskCount  int
	)
	JustBeforeEach(func() {
		StartTorpedoTest("PxRestartAfterAddNewHotPlugDiskToKubevirtVM", "Add a new raw disk to a running KubeVirt VM via hot-plug", nil, 0)
		volType, present = os.LookupEnv("KUBEVIRT_VOL_TYPE")
		if !present {
			app = "kubevirt-debian-fio-minimal"
		}
		if volType == "pxe-raw" {
			app = "kubevirt-raw-vol"
		} else if volType == "fada-raw" {
			app = "kubevirt-fada-raw-fio"
		} else {
			app = "kubevirt-debian-fio-minimal"
		}
		log.InfoD("Setting app for this test to be : %s", app)
	})

	It("hot-plug a new disk to a running KubeVirt VM", func() {
		pxNs, err := Inst().V.GetVolumeDriverNamespace()
		log.FailOnError(err, "Failed to get volume driver namespace")
		defer ListEvents(pxNs)

		appList := Inst().AppList
		defer func() {
			Inst().AppList = appList
		}()

		numberOfVolumes := 1

		Inst().AppList = []string{app}
		Inst().CsiAppList = []string{app}

		stepLog := "Schedule a kubevirt VM"
		Step(stepLog, func() {
			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				namespace = fmt.Sprintf("kubevirt-%v", time.Now().Unix())
				appCtxs = append(appCtxs, ScheduleApplicationsOnNamespace(namespace, "test")...)
			}
		})
		ValidateApplications(appCtxs)

		if !present {
			for _, appCtx := range appCtxs {
				bindMount, err = IsVMBindMounted(appCtx, false)
				log.FailOnError(err, "Failed to verify bind mount")
				dash.VerifyFatal(bindMount, true, "VM bind mount verified")
			}
		}

		if app == "kubevirt-debian-fio-minimal" {
			volumeMode = ""
		} else {
			volumeMode = "Block"
		}

		log.Infof("Sleeping for 2 minutes to let VMs come up fully")
		time.Sleep(2 * time.Minute)

		canSsh = CreateSSHPodAndSetCanSsh()
		ValidateFioInVMs(appCtxs, canSsh)
		initialUptime = make(map[string]time.Duration)
		stepLog = "Get initial uptime of VMs and current node"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			var wg sync.WaitGroup
			for _, appCtx := range appCtxs {
				wg.Add(1)
				go func(appCtx *scheduler.Context) {
					defer GinkgoRecover()
					defer wg.Done()
					vms, err := GetAllVMsFromScheduledContexts([]*scheduler.Context{appCtx})
					log.FailOnError(err, "Failed to get VMs from appCtx")
					for _, vm := range vms {
						uptime, err := GetVMUptime(vm)
						log.FailOnError(err, "Failed to get uptime from VM %s", vm.Name)
						vmKey := fmt.Sprintf("%s/%s", vm.Namespace, vm.Name)
						initialUptime[vmKey] = uptime
						log.Infof("Initial uptime for VM %s is %v", vmKey, uptime)

						vmNodeName, err = GetNodeOfVM(vm)
						log.FailOnError(err, "Failed to get node of VM %v", vm.Name)
						log.Infof("VM %s is currently running on node %s", vm.Name, vmNodeName)
					}
				}(appCtx)
			}
			wg.Wait()
		})

		stepLog = "Hot-plug one raw disk (DataVolume) to the running KubeVirt VM"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			hotPlugDisk, err := HotPlugDataVolumesToKubevirtVM(appCtxs, numberOfVolumes, "50Gi", volumeMode)
			log.FailOnError(err, "Failed to hot-plug DataVolume to KubeVirt VM")
			dash.VerifyFatal(hotPlugDisk, true, "DataVolume hot-plugged to KubeVirt VM")
		})

		vmDiskCount = make(map[string]int)
		stepLog = "Kill Px on node hosting VM"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for _, appCtx := range appCtxs {
				vms, err := GetAllVMsFromScheduledContexts([]*scheduler.Context{appCtx})
				log.FailOnError(err, "Failed to get VMs from context")
				for _, vm := range vms {
					nodeName, err := GetNodeOfVM(vm)
					log.FailOnError(err, "Failed to get node of vm %v", vm.Name)
					nodes = append(nodes, nodeName)
					newDiskCount, err = GetNumberOfDrivesInVM(vm)
					log.FailOnError(err, "Failed to get disk count of vm %v", vm.Name)
					vmDiskCount[vm.Name] = newDiskCount
				}
			}
			for _, appNode := range node.GetStorageDriverNodes() {
				for _, vmNode := range nodes {
					if vmNode == appNode.Name {
						stepLog = fmt.Sprintf("Stop volume driver %s on node: %s", Inst().V.String(), appNode.Name)
						Step(stepLog, func() {
							log.InfoD(stepLog)
							StopVolDriverAndWait([]node.Node{appNode})
						})

						stepLog = fmt.Sprintf("Start volume driver %s on node %s", Inst().V.String(), appNode.Name)
						Step(stepLog, func() {
							log.InfoD(stepLog)
							StartVolDriverAndWait([]node.Node{appNode})
						})

						stepLog = "Giving few seconds for volume driver to stabilize"
						Step(stepLog, func() {
							log.InfoD(stepLog)
							time.Sleep(20 * time.Second)
						})
					}
				}
			}
		})

		ValidateFioInVMs(appCtxs, canSsh)
		ValidateVMUptime(appCtxs, canSsh, initialUptime)

		stepLog = "Validate hot plug volume inside VM"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for _, appCtx := range appCtxs {
				vms, err := GetAllVMsFromScheduledContexts([]*scheduler.Context{appCtx})
				log.FailOnError(err, "Failed to get VMs from context")
				for _, vm := range vms {
					diskCountAfterPxRestart, err := GetNumberOfDrivesInVM(vm)
					log.FailOnError(err, "Failed to get disk count of vm %v", vm.Name)
					dash.VerifyFatal(diskCountAfterPxRestart == vmDiskCount[vm.Name], true, "Validate the number of disk same after PX restart to ensure hot plug disk is present")
				}
			}
		})

		stepLog = "Destroy Applications"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			DestroyApps(appCtxs, nil)
		})
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(appCtxs)
	})
})
