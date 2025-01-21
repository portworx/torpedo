package tests

// Code temporarily commented out due to pending vendor update.
// Updates are required to incorporate support for CSI node placement and Portworx API functions.
// Once the vendor dependencies are updated to include these capabilities,
// we can uncomment this section and proceed with testing the functionality.

/*
import (
	"github.com/pure-px/sched-ops/k8s/core"
	"github.com/pure-px/sched-ops/k8s/operator"

	"strings"
	"time"

	opcorev1 "github.com/libopenstorage/operator/pkg/apis/core/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/pure-px/torpedo/drivers/node"
	"github.com/pure-px/torpedo/drivers/scheduler"
	"github.com/pure-px/torpedo/pkg/log"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/pure-px/torpedo/tests"
)

var _ = Describe("{OperatorTestAddToleration}", Label("staging", "p1", "negative", "AsyncDR"), func() {

		// https://purestorage.atlassian.net/browse/HAZEL-224

		// 1. Deploy cluster with multiple nodes
		// 2. Add tolerations for different components on some nodes
		// 3. Validate the resources if they followed the tolerations specified.

	JustBeforeEach(func() {
		StartTorpedoTest("OperatorTestAddToleration", "Add tolerations for different components on some nodes and Validate the resources if they followed the tolerations specified", nil, 0)
	})

	var contexts []*scheduler.Context
	itLog := "Add tolerations for different components"
	It(itLog, func() {
		log.InfoD(itLog)
		const (
			kvdbAnnotation = "pod/portworx-kvdb"
			csiAnnotation  = "deployment/px-csi-ext"
			apiAnnotation  = "daemonset/portworx-api"
		)
		var (
			k8sCore            = core.Instance()
			pxStc, backupPxStc *opcorev1.StorageCluster
			newPods            []corev1.Pod
			labels             = []string{"px/api=svc", "px/csi=enabled", "px/kvdb=internal"}
			workerNodes        []node.Node
			nodeLabelMap       = make(map[string]string)
		)

		stepLog := "Get worker nodes and add labels"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			workerNodes = node.GetWorkerNodes()
			for i, node := range workerNodes {
				label := labels[i%len(labels)]
				keyval := strings.Split(label, "=")
				err := k8sCore.AddLabelOnNode(node.Name, keyval[0], keyval[1])
				log.FailOnError(err, "Failed to add label %s for the node: %v", label, node.Name)
				nodeLabelMap[node.Name] = keyval[0]
				log.Infof("Label %v added to the node: %v", label, node.Name)
			}
			log.Infof("Added labels for the nodes: %v", workerNodes)
		})

		stepLog = "Get storage cluster and tolerations"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			pxStc, err = Inst().V.GetDriver()
			log.FailOnError(err, "error while getting storage cluster spec")
			log.InfoD("storage cluster %v for the namespace %v", pxStc, pxStc.Namespace)

			backupPxStc = pxStc.DeepCopy()

			stepLog = "Add annotations in the storage cluster"
			Step(stepLog, func() {
				log.InfoD(stepLog)
				if pxStc.Spec.Metadata == nil {
					pxStc.Spec.Metadata = &opcorev1.Metadata{}
					pxStc.Spec.Metadata.Annotations = make(map[string]map[string]string)
				}
				if pxStc.Spec.Metadata.Annotations == nil {
					pxStc.Spec.Metadata.Annotations = make(map[string]map[string]string)
				}
				if _, exists := pxStc.Spec.Metadata.Annotations[kvdbAnnotation]; !exists {
					pxStc.Spec.Metadata.Annotations[kvdbAnnotation] = make(map[string]string)
				}
				if _, exists := pxStc.Spec.Metadata.Annotations[csiAnnotation]; !exists {
					pxStc.Spec.Metadata.Annotations[csiAnnotation] = make(map[string]string)
				}
				if _, exists := pxStc.Spec.Metadata.Annotations[apiAnnotation]; !exists {
					pxStc.Spec.Metadata.Annotations[apiAnnotation] = make(map[string]string)
				}
				pxStc.Spec.Metadata.Annotations[kvdbAnnotation]["px/kvdb"] = "internal"
				pxStc.Spec.Metadata.Annotations[csiAnnotation]["px/csi"] = "enabled"
				pxStc.Spec.Metadata.Annotations[apiAnnotation]["px/api"] = "svc"
				log.Infof("Added annotations for kvdb, csi and api pods %v", pxStc.Spec.Metadata.Annotations)
			})

			stepLog = "Add the limits & requests in the storage cluster"
			Step(stepLog, func() {
				log.InfoD(stepLog)
				resources := corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceMemory: resource.MustParse("64Mi"),
						corev1.ResourceCPU:    resource.MustParse("250m"),
					},
					Limits: corev1.ResourceList{
						corev1.ResourceMemory: resource.MustParse("128Mi"),
						corev1.ResourceCPU:    resource.MustParse("500m"),
					},
				}
				if pxStc.Spec.PortworxAPI == nil {
					pxStc.Spec.PortworxAPI = &opcorev1.PortworxAPISpec{}
				}
				pxStc.Spec.PortworxAPI.Resources = &resources
				pxStc.Spec.Kvdb.Resources = &resources

				if pxStc.Spec.CSI.ExternalProvisioner == nil {
					pxStc.Spec.CSI.ExternalProvisioner = &opcorev1.CSIExternalProvisionerSpec{}
				}
				pxStc.Spec.CSI.ExternalProvisioner.Resources = &resources

				if pxStc.Spec.CSI.Resizer == nil {
					pxStc.Spec.CSI.Resizer = &opcorev1.CSIResizerSpec{}
				}
				pxStc.Spec.CSI.Resizer.Resources = &resources

				if pxStc.Spec.CSI.Snapshotter == nil {
					pxStc.Spec.CSI.Snapshotter = &opcorev1.CSISnapshotterSpec{}
				}
				pxStc.Spec.CSI.Snapshotter.Resources = &resources

				if pxStc.Spec.CSI.SnapshotController == nil {
					pxStc.Spec.CSI.SnapshotController = &opcorev1.CSISnapshotControllerSpec{}
				}
				pxStc.Spec.CSI.SnapshotController.Resources = &resources

				log.Infof("Added resources for kvdb, csi and api pods")
			})

			stepLog = "Add nodeaffinity for csi in the storage cluster"
			Step(stepLog, func() {
				log.InfoD(stepLog)
				selectorRequirements := []v1.NodeSelectorRequirement{
					{
						Key:      "px/csi",
						Operator: v1.NodeSelectorOpExists,
					},
				}
				newMatch := v1.NodeSelectorTerm{
					MatchExpressions: selectorRequirements,
				}

				if pxStc.Spec.CSI.Placement == nil {
					pxStc.Spec.CSI.Placement = &opcorev1.PlacementSpec{}
				}

				if pxStc.Spec.CSI.Placement.NodeAffinity == nil {
					pxStc.Spec.CSI.Placement.NodeAffinity = &v1.NodeAffinity{}
				}

				if pxStc.Spec.CSI.Placement.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution == nil {
					pxStc.Spec.CSI.Placement.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution = &v1.NodeSelector{}
				}

				pxStc.Spec.CSI.Placement.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms = []v1.NodeSelectorTerm{newMatch}
				log.Infof("Added placement rule for the csi pods")
			})
			_, err = operator.Instance().UpdateStorageCluster(pxStc)
			log.FailOnError(err, "error while updating stc")

		})

		log.Infof("waiting for storage cluster get updated")
		time.Sleep(5 * time.Minute)

		stepLog = "validate annotation and tollerations are added in the pods"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			podList, err := k8sCore.GetPods(pxStc.Namespace, nil)
			log.FailOnError(err, "error while get pods for the namespace %v", pxStc.Namespace)
			log.InfoD("Pod list %v for the namespace %v", podList, pxStc.Namespace)
			for _, pod := range podList.Items {
				if strings.Contains(pod.Name, "api") || strings.Contains(pod.Name, "kvdb") || strings.Contains(pod.Name, "csi") {
					newPods = append(newPods, pod)
				}
			}
			log.Infof("filtered pod list: %v namespace: %v", newPods, pxStc.Namespace)

			for _, pod := range newPods {
				nodeName := pod.Spec.NodeName
				nodeLabels, err := k8sCore.GetLabelsOnNode(nodeName)
				log.FailOnError(err, "unable to find the node from the pod")

				if strings.Contains(pod.Name, "api") {
					_, exists := pod.Annotations["px/api"]
					dash.VerifyFatal(exists, true, "check if api annotation added")
					isResourceAdded := false
					for _, container := range pod.Spec.Containers {
						if container.Name == "portworx-api" {
							if *container.Resources.Requests.Memory() == resource.MustParse("64Mi") {
								isResourceAdded = true
								break
							}
						}
					}
					dash.VerifyFatal(isResourceAdded, true, "check if api resource added")
				} else if strings.Contains(pod.Name, "kvdb") {
					_, exists := pod.Annotations["px/kvdb"]
					dash.VerifyFatal(exists, true, "check if kvdb annotation added")
					isResourceAdded := false
					for _, container := range pod.Spec.Containers {
						if container.Name == "portworx-kvdb" {
							if *container.Resources.Requests.Memory() == resource.MustParse("64Mi") {
								isResourceAdded = true
								break
							}
						}
					}
					dash.VerifyFatal(isResourceAdded, true, "check if kvdb resource added")
				} else if strings.Contains(pod.Name, "csi") {
					_, isRunningOnNode := nodeLabels["px/csi"]
					dash.VerifyFatal(isRunningOnNode, true, "check if csi pod running on labeled node")

					_, exists := pod.Annotations["px/csi"]
					dash.VerifyFatal(exists, true, "check if csi annotation added")

					isResourceAdded := false
					for _, container := range pod.Spec.Containers {
						if container.Name == "csi-external-provisioner" {
							if *container.Resources.Requests.Memory() == resource.MustParse("64Mi") {
								isResourceAdded = true
								break
							}
						}
					}
					dash.VerifyFatal(isResourceAdded, true, "check if csi resource added")
				}
			}
		})

		stepLog = "Remove all the labels and annotation changes"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			newPxStc, err := Inst().V.GetDriver()
			log.FailOnError(err, "error while getting storage cluster spec")

			newPxStc.Spec = backupPxStc.Spec

			_, err = operator.Instance().UpdateStorageCluster(newPxStc)
			log.FailOnError(err, "error while updating stc")

			log.Infof("waiting for storage cluster get updated")
			time.Sleep(3 * time.Minute)

			for nodeName, labelKey := range nodeLabelMap {
				err := k8sCore.RemoveLabelOnNode(nodeName, labelKey)
				log.FailOnError(err, "Failed to remove label for the node: %v", nodeName)
				log.Infof("Label %v removed from the node: %v", labelKey, nodeName)
			}
		})

	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts)
	})
})
*/
