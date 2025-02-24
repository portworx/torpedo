package tests

import (
	"fmt"
	"strings"
	"time"

	"github.com/pure-px/sched-ops/k8s/core"
	"github.com/pure-px/sched-ops/k8s/storage"
	"github.com/pure-px/sched-ops/task"
	"github.com/pure-px/torpedo/pkg/log"
	corev1 "k8s.io/api/core/v1"
	storageApi "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pure-px/torpedo/drivers/node"
	"github.com/pure-px/torpedo/drivers/scheduler"
	"github.com/pure-px/torpedo/drivers/scheduler/k8s"
	"github.com/pure-px/torpedo/drivers/volume"
	"github.com/pure-px/torpedo/pkg/testrailuttils"
	. "github.com/pure-px/torpedo/tests"
)

const (
	nodeDeleteTimeoutMins = 7 * time.Minute
)

// This test performs multi volume mounts to a single deployment
var _ = Describe("{MultiVolumeMountsForSharedV4}", Label("p0", "positive", "px_vol_ops", "shared_v4", "MiniScale"), func() {
	var testrailID = 58846
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/cases/view/58846
	var runID int
	JustBeforeEach(func() {
		StartTorpedoTest("MultiVolumeMountsForSharedV4", "Validate mounting multiple SV4 volumes for one app", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})
	var contexts []*scheduler.Context

	stepLog := "has to create multiple sharedv4 volumes and mount to single pod"
	It(stepLog, func() {
		log.InfoD(stepLog)

		stepLog = "schedule application with multiple sharedv4 volumes attached"
		timeout := 5 * time.Minute

		Step(stepLog, func() {
			log.InfoD(stepLog)

			taskName := "sharedv4-multivol"

			log.Infof("Task name %s\n", taskName)

			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				newContexts := ScheduleApplications(taskName)
				contexts = append(contexts, newContexts...)
			}

			for _, ctx := range contexts {
				pvcs, err := k8sCore.GetPersistentVolumeClaims(ctx.App.NameSpace, nil)
				log.FailOnError(err, "Failed to get PVCs for app %s", ctx.App.Key)
				timeout = (15 * time.Duration(len(pvcs.Items)) * time.Minute) / 10
				log.InfoD("Number of Volumes to be mounted: %v", len(pvcs.Items))
				ctx.ReadinessTimeout = timeout
				ctx.SkipVolumeValidation = false
				ValidateContext(ctx)
			}
		})
		stepLog = "get nodes where volume is attached and restart volume driver"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for _, ctx := range contexts {
				appVolumes, err := Inst().S.GetVolumes(ctx)
				Expect(err).NotTo(HaveOccurred())
				for _, appVolume := range appVolumes {
					attachedNode, err := Inst().V.GetNodeForVolume(appVolume, defaultCommandTimeout, defaultCommandRetry)
					log.FailOnError(err, "Failed to get volume %s from node", appVolume.Name)
					stepLog = fmt.Sprintf("stop volume driver %s on app %s's node: %s",
						Inst().V.String(), ctx.App.Key, attachedNode.Name)
					Step(stepLog,
						func() {
							log.InfoD(stepLog)
							StopVolDriverAndWait([]node.Node{*attachedNode})
						})
					stepLog = fmt.Sprintf("starting volume %s driver on app %s's node %s",
						Inst().V.String(), ctx.App.Key, attachedNode.Name)
					Step(stepLog,
						func() {
							log.InfoD(stepLog)
							StartVolDriverAndWait([]node.Node{*attachedNode})
						})
					stepLog = "Giving few seconds for volume driver to stabilize"
					Step(stepLog, func() {
						log.InfoD(stepLog)
						time.Sleep(20 * time.Second)
					})
					stepLog = fmt.Sprintf("validate app %s", attachedNode.Name)
					Step(stepLog, func() {
						ctx.ReadinessTimeout = timeout
						ctx.SkipVolumeValidation = true
						ValidateContext(ctx)
					})
				}
			}
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})
})

// This test performs sharedv4 nfs server pod termination failover use case
var _ = Describe("{NFSServerNodeDelete}", Label("p0", "negative", "px_vol_ops", "node_ops", "shared_v4"), func() {
	JustBeforeEach(func() {
		StartTorpedoTest("NFSServerNodeDelete", "Vslidate NFS server delete", nil, 0)
	})

	var contexts []*scheduler.Context
	stepLog := "has to validate that the new pods started successfully after nfs server node is terminated"
	It(stepLog, func() {
		log.InfoD(stepLog)
		contexts = make([]*scheduler.Context, 0)
		var err error

		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("nodekill-%d", i))...)
		}

		ValidateApplications(contexts)
		for _, ctx := range contexts {
			var appVolumes []*volume.Volume
			stepLog = fmt.Sprintf("get volumes for %s app", ctx.App.Key)
			Step(stepLog, func() {
				log.InfoD(stepLog)
				appVolumes, err = Inst().S.GetVolumes(ctx)
				log.FailOnError(err, "Failed to get volumes")
				dash.VerifyFatal(len(appVolumes) > 0, 0, " App volumes are empty?")
			})
			for _, v := range appVolumes {
				stepLog = "get attached node and stop the instance"
				Step(stepLog, func() {
					log.InfoD(stepLog)
					currNodes := node.GetStorageDriverNodes()
					countOfCurrNodes := len(currNodes)

					attachedNode, err := Inst().V.GetNodeForVolume(v, defaultCommandTimeout, defaultCommandRetry)

					stepLog = fmt.Sprintf("delete node : %v having volume: %v attached", attachedNode.Name, v.Name)
					// Delete node and check Apps status
					Step(stepLog, func() {
						log.InfoD(stepLog)
						sv4KillANodeAndValidate(*attachedNode)
						stepLog = fmt.Sprintf("validate node: %v is deleted", attachedNode.Name)
						Step(stepLog, func() {
							log.InfoD(stepLog)
							currNodes = node.GetStorageDriverNodes()
							for _, currNode := range currNodes {
								if currNode.Name == attachedNode.Name {
									dash.VerifyFatal(currNode.Name, attachedNode.Name, fmt.Sprintf("Node: %v still exists?",
										attachedNode.Name))
									break
								}
							}
						})

						stepLog = fmt.Sprintf("validate applications after node [%v] deletion", attachedNode.Name)
						Step(stepLog, func() {
							log.InfoD(stepLog)
							for _, ctx := range contexts {
								ValidateContext(ctx)
							}
						})
						stepLog = fmt.Sprintf("wait to new instance to start scheduler: %s and volume driver: %s",
							Inst().S.String(), Inst().V.String())
						Step(stepLog, func() {
							log.InfoD(stepLog)
							time.Sleep(2 * time.Minute)
							currNodes = node.GetStorageDriverNodes()
							dash.VerifyFatal(countOfCurrNodes, len(currNodes), "Create new instance successful?")
							Expect(countOfCurrNodes).To(Equal(len(currNodes)))
							log.InfoD("Validating Node and Volume driver for all nodes")
							for _, n := range currNodes {

								err = Inst().S.IsNodeReady(n)
								log.FailOnError(err, "Node %s not ready", n.Name)

								err = Inst().V.WaitDriverUpOnNode(n, Inst().DriverStartTimeout)
								log.FailOnError(err, "Failed to wait for volume driver %s to be up", n.Name)
							}
						})

						Step("validate apps after new node is ready", func() {
							for _, ctx := range contexts {
								ValidateContext(ctx)
							}
						})

					})
				})
			}

		}

	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts)
	})
})

func sv4KillANodeAndValidate(nodeToKill node.Node) {
	steplog := fmt.Sprintf("Deleting node [%v]", nodeToKill.Name)
	Step(steplog, func() {
		log.InfoD(steplog)
		log.Infof("Instance is of %v ", Inst().S.String())
		err := Inst().S.DeleteNode(nodeToKill)
		dash.VerifyFatal(err, nil, "Validate node delete init")
	})
	steplog = fmt.Sprintf("Wait for node: %v to be deleted", nodeToKill.Name)
	Step(steplog, func() {
		log.InfoD(steplog)
		maxWait := 10
	OUTER:
		for maxWait > 0 {
			for _, currNode := range node.GetStorageDriverNodes() {
				if currNode.Name == nodeToKill.Name {
					log.Infof("Node %v still exists. Waiting for a minute to check again", nodeToKill.Name)
					maxWait--
					time.Sleep(1 * time.Minute)
					continue OUTER
				}
			}
			break
		}
	})

	err := Inst().S.RefreshNodeRegistry()
	dash.VerifyFatal(err, nil, "Validate node registry refresh")

	err = Inst().V.RefreshDriverEndpoints()
	dash.VerifyFatal(err, nil, "Validate volume driver end points refresh")
}

var _ = Describe("{PVCCreationWithRWXAccessModeAndSharedv4False}", Label("p0", "staging", "negative", "sharedv4"), func() {
	/*
		    Ticket id:https://purestorage.atlassian.net/browse/HAZEL-1087
			step1: set this parameter "sharedv4: false" in sc file
			step2: create pvc in have readwrite many.
			step3: pvc should in pending state with this error msg
	*/
	var (
		testrailID = 0
		runID      int
		contexts   = make([]*scheduler.Context, 0)
	)
	JustBeforeEach(func() {
		StartTorpedoTest("PVCCreationWithRWXAccessModeAndSharedv4False", "When PVC with RWX is ignored due to sharedv4:false", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})

	stepLog := "Verify sc in file sharedv4 is false pvc with RWX Should not happen"
	It(stepLog, func() {
		log.InfoD(stepLog)
		var (
			scName     = fmt.Sprintf("test-sv4-sc-svc%v", time.Now().Unix())
			pvcName    = fmt.Sprintf("test-sv4-pvc-svc-%v", time.Now().Unix())
			namespace  = "default"
			k8sStorage = storage.Instance()
		)

		stepLog = "Apply storage class with sharedv4:false"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			storageClass := storageApi.StorageClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: scName,
				},
				Provisioner: k8s.CsiProvisioner,
				Parameters: map[string]string{
					"sharedv4": "false",
				},
			}
			_, err := k8sStorage.CreateStorageClass(&storageClass)
			dash.VerifyFatal(err, nil, "Verifying creation of new storage class")
		})

		stepLog = "Apply persistent volume claim"
		Step(stepLog, func() {
			log.InfoD(stepLog)
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
			dash.VerifyFatal(err, nil, "Verifying creation of new pvc")
		})
		log.Infof("Starting to monitor PVC status")
		createdPVC, err := k8sCore.GetPersistentVolumeClaim(pvcName, namespace)
		log.FailOnError(err, "unable to get pvc %v and namespace %v", pvcName, namespace)
		if createdPVC.Status.Phase == "Pending" {
			log.Infof("PVC status: %v", createdPVC.Status.Phase)

			retryFunc := func() (interface{}, bool, error) {
				log.Infof("Fetching PVC events...")
				events := Inst().S.GetEvents()["PersistentVolumeClaim"]
				if len(events) == 0 {
					log.Infof("No events found for PersistentVolumeClaim, retrying...")
					return nil, true, fmt.Errorf("No events found for PersistentVolumeClaim, retrying...")
				}
				for _, event := range events {
					log.Infof("Checking event: %v", event)
					if strings.Contains(event.Message, "Failed to create volume: ReadWriteMany PVC has a StorageClass that disallows sharedv4 volumes; use a different StorageClass") {
						log.Infof("Volume creation error is: %v", event)
						dash.VerifyFatal(createdPVC.Status.Phase == "Pending", true, "Check if volume creation status is pending")
						errorMsg := "Failed to create volume: ReadWriteMany PVC has a StorageClass that disallows sharedv4 volumes; use a different StorageClass"
						dash.VerifyFatal(strings.Contains(event.Message, errorMsg), true, "Check if the volume was not created due to sharedv4 issue")
						return event, false, nil
					}
				}
				return nil, true, fmt.Errorf("Event not found, retrying...")
			}
			_, err := task.DoRetryWithTimeout(retryFunc, 3*time.Minute, 30*time.Second)
			log.FailOnError(err, "Retried for 3 minutes without finding the expected error message.")
			log.Infof("Verified the error message when sharedv4 is false ")
		} else {
			log.Infof("PVC should not be in the bound state as sharedv4 is false")
			err := fmt.Errorf("PVC should not be in the bound state because sharedv4 is false")
			log.FailOnError(err, "Verified the error message when sharedv4 is false ")
		}
		Step("Cleanup", func() {
			log.InfoD("Cleaning up resources")
			err = core.Instance().DeletePersistentVolumeClaim(pvcName, namespace)
			log.FailOnError(err, "Failed to delete PVC: %v", pvcName)
			err = k8sStorage.DeleteStorageClass(scName)
			log.FailOnError(err, "Failed to delete StorageClass: %v", scName)
		})
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})
})

var _ = Describe("{PVCCreationWithRWXAccessModeAndSharedv4True}", Label("p0", "staging", "positive", "sharedv4"), func() {
	/*
			step1: set this parameter "sharedv4: true" in sc file
		    step2: pvc is rwx and it's should be bound
	*/
	var (
		testrailID = 0
		runID      int
		contexts   = make([]*scheduler.Context, 0)
	)
	JustBeforeEach(func() {
		StartTorpedoTest("PVCCreationWithRWXAccessModeAndSharedv4True", "verifies PVC creation with RWX access mode and shared volume support for v4.", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})

	stepLog := "verifies PVC creation with RWX access mode and shared volume support for v4"
	It(stepLog, func() {
		log.InfoD(stepLog)
		var (
			scName     = fmt.Sprintf("test-sv4-sc-svc%v", time.Now().Unix())
			pvcName    = fmt.Sprintf("test-sv4-pvc-svc-%v", time.Now().Unix())
			namespace  = "default"
			k8sStorage = storage.Instance()
		)

		stepLog = "Apply storage class with sharedv4:true"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			storageClass := storageApi.StorageClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: scName,
				},
				Provisioner: k8s.CsiProvisioner,
				Parameters: map[string]string{
					"sharedv4": "true",
				},
			}
			_, err := k8sStorage.CreateStorageClass(&storageClass)
			dash.VerifyFatal(err, nil, "Verifying creation of new storage class with sharedv4=true")
		})

		stepLog = "Apply persistent volume claim with RWX access mode"
		Step(stepLog, func() {
			log.InfoD(stepLog)
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
			dash.VerifyFatal(err, nil, "Verifying creation of new PVC with RWX access mode")
		})
		log.Infof("Starting to monitor PVC status and PV attributes")
		retryFunc := func() (interface{}, bool, error) {
			createdPVC, err := core.Instance().GetPersistentVolumeClaim(pvcName, namespace)
			if err != nil {
				return nil, true, fmt.Errorf("failed to get PVC: %v", err)
			}
			if createdPVC.Status.Phase == "Pending" {
				log.Infof("PVC status: Pending, still waiting to be bound. Retrying...")
				return nil, true, fmt.Errorf("PVC status is neither Pending nor Bound, retrying...")
			}

			log.Infof("PVC successfully bound. PVC status: %v", createdPVC.Status.Phase)
			if createdPVC.Spec.VolumeName == "" {
				return nil, true, fmt.Errorf("PVC has no bound PersistentVolumeName. Retrying..")
			}

			pvName := createdPVC.Spec.VolumeName
			pv, err := core.Instance().GetPersistentVolume(pvName)
			if err != nil {
				return nil, true, fmt.Errorf("failed to get PV: %v", err)
			}

			log.Infof("Persistent Volume details: %v", pv)
			sharedv4Param, exists := pv.Spec.PersistentVolumeSource.CSI.VolumeAttributes["sharedv4"]
			dash.VerifyFatal(exists, true, "Verify if the PV has sharedv4 attribute")
			dash.VerifyFatal(sharedv4Param, "true", "Verify if sharedv4 attribute is set to true")

			return createdPVC, false, nil

		}
		_, err := task.DoRetryWithTimeout(retryFunc, 3*time.Minute, 30*time.Second)
		log.FailOnError(err, "Error checking PVC status and binding after retries")
		log.Infof("PVC check completed and verified the PVC bounding and PV attributes successfully")
		Step("Cleanup", func() {
			log.InfoD("Cleaning up resources")
			err = core.Instance().DeletePersistentVolumeClaim(pvcName, namespace)
			log.FailOnError(err, "Failed to delete PVC: %v", pvcName)
			err = k8sStorage.DeleteStorageClass(scName)
			log.FailOnError(err, "Failed to delete StorageClass: %v", scName)
		})
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})
})
