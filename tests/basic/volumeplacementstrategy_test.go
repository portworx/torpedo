package tests

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/libopenstorage/openstorage/api"
	. "github.com/onsi/ginkgo/v2"
	"github.com/portworx/sched-ops/k8s/core"
	"github.com/portworx/sched-ops/k8s/storage"
	"github.com/portworx/sched-ops/k8s/talisman"
	"github.com/portworx/talisman/pkg/apis/portworx/v1beta1"
	"github.com/portworx/talisman/pkg/apis/portworx/v1beta2"
	"github.com/portworx/torpedo/drivers/node"
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/drivers/scheduler/k8s"
	"github.com/portworx/torpedo/drivers/volume"
	"github.com/portworx/torpedo/pkg/log"
	"github.com/portworx/torpedo/pkg/testrailuttils"
	"github.com/portworx/torpedo/pkg/vpsutil"
	. "github.com/portworx/torpedo/tests"
	corev1 "k8s.io/api/core/v1"
	storageApi "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("{VolumePlacementStrategyFunctional}", func() {
	var testrailID, runID int
	var contexts []*scheduler.Context

	JustBeforeEach(func() {
		runID = testrailuttils.AddRunsToMilestone(testrailID)

		StartTorpedoTest("VolumePlacementStrategyFunctional", "Functional Tests for VPS", nil, testrailID)
	})

	Context("VolumePlacementStrategyValidation", func() {
		var vpsTestCase vpsutil.VolumePlaceMentStrategyTestCase

		testValidateVPS := func() {
			It("has to deploy VPS and validate the scheduled application follow specified rules", func() {
				Step("Deploying VPS", func() {
					log.InfoD("Deploy VPS for %v", vpsTestCase.TestName())
					err := vpsTestCase.DeployVPS()
					log.FailOnError(err, "Failed to Deploy VPS Spec")
				})

				Step("Deploy and Validate Applications", func() {
					log.InfoD("Deploy Applications")
					contexts = make([]*scheduler.Context, 0)
					for i := 0; i < Inst().GlobalScaleFactor; i++ {
						contexts = append(contexts, ScheduleApplications(fmt.Sprintf("%s-%d", vpsTestCase.TestName(), i))...)
					}
					log.InfoD("Validate Applications")
					ValidateApplications(contexts)
				})

				Step("Validate Deployment with VPS", func() {
					log.InfoD("Validate Deployment with VPS")
					err := vpsTestCase.ValidateVPSDeployment(contexts)
					log.FailOnError(err, "Failed to Validate Deployments with respect to VPS")
				})

				Step("Destroy VPS Deployment", func() {
					log.InfoD("Destroy VPS Deployment")
					err := vpsTestCase.DestroyVPSDeployment()
					log.FailOnError(err, "Failed to Destroy VPS Deployments")
				})

			})
		}

		// test mongo volume anti affinity
		Context("{VPSMongoVolumeAntiAffinity}", func() {
			BeforeEach(func() {
				vpsTestCase = &mongoVolumeAntiAffinity{}
			})
			testValidateVPS()
		})

		// test mongo volume affinity with dynamic labels
		Context("{VPSMongoVolumeAntiAffinityDynamicLabels}", func() {
			BeforeEach(func() {
				vpsTestCase = &mongoVolumeAntiAffinityDynamicLabels{}
			})
			testValidateVPS()
		})

		// test mongo volume affinity
		Context("{VPSMongVolumeoAffinity}", func() {
			BeforeEach(func() {
				vpsTestCase = &mongoVolumeAffinity{}
			})
			testValidateVPS()
		})

		// test mongo replica affinity
		Context("{VPSMongoReplicaAffinity}", func() {
			BeforeEach(func() {
				vpsTestCase = &mongoVPSReplicaAffinity{}
			})
			testValidateVPS()
		})

		// test mongo replica anti affinity
		Context("{VPSMongoReplicaAntiAffinity}", func() {
			BeforeEach(func() {
				vpsTestCase = &mongoVPSReplicaAntiAffinity{}
			})
			testValidateVPS()
		})
	})

	AfterEach(func() {
		Step("destroy apps", func() {
			log.InfoD("destroying apps")
			if CurrentSpecReport().Failed() {
				log.InfoD("not destroying apps because the test failed\n")
				return
			}
			for _, ctx := range contexts {
				TearDownContext(ctx, map[string]bool{scheduler.OptionsWaitForResourceLeakCleanup: true})
			}

		})
	})

	AfterEach(func() {
		AfterEachTest(contexts, testrailID, runID)
		defer EndTorpedoTest()
	})
})

type VolumePlacementStrategySpec struct {
	spec *v1beta2.VolumePlacementStrategy
}

type mongoVolumeAntiAffinity struct {
	VolumePlacementStrategySpec
}

func (m *mongoVolumeAntiAffinity) TestName() string {
	return "mongovolumeantiaffinity"
}

func (m *mongoVolumeAntiAffinity) DeployVPS() error {

	matchExpression := []*v1beta1.LabelSelectorRequirement{
		{
			Key:      "px/statefulset-pod",
			Operator: v1beta1.LabelSelectorOpIn,
			Values:   []string{"${pvc.statefulset-pod}"},
		},
		{
			Key:      "app",
			Operator: v1beta1.LabelSelectorOpIn,
			Values:   []string{"mongo-sts"},
		},
	}

	vpsSpec := vpsutil.VolumeAntiAffinityByMatchExpression("mongo-vps", matchExpression)
	_, err := talisman.Instance().CreateVolumePlacementStrategy(&vpsSpec)
	m.spec = &vpsSpec
	return err
}

func (m *mongoVolumeAntiAffinity) DestroyVPSDeployment() error {
	return talisman.Instance().DeleteVolumePlacementStrategy(m.spec.Name)
}

// mongoVPSAntiAffinity is expecting to have deploy 2 replica of vol for each pod that has label [mongo-0, mongo-1]
// since this is antiaffinity, we are expecting that vol with the same labels are not deployed on the same pool/node.
// To validate that, we get the label from each deployed vol and extract the pool it's deployed on. if deployed correctly,
// there should be two pools per label.
func (m *mongoVolumeAntiAffinity) ValidateVPSDeployment(contexts []*scheduler.Context) error {
	vols, err := Inst().S.GetVolumes(contexts[0])
	if err != nil {
		return err
	}
	apiVols, err := getApiVols(vols)
	if err != nil {
		return err
	}

	volumeLabelKey := "px/statefulset-pod"
	expectedNodeLength := 2

	return vpsutil.ValidateVolumeAntiAffinityByNode(apiVols, volumeLabelKey, expectedNodeLength)
}

type mongoVolumeAntiAffinityDynamicLabels struct {
	VolumePlacementStrategySpec
}

func (m *mongoVolumeAntiAffinityDynamicLabels) TestName() string {
	return "mongovolumeantiaffinitydl"
}

func (m *mongoVolumeAntiAffinityDynamicLabels) DeployVPS() error {

	matchExpression := []*v1beta1.LabelSelectorRequirement{
		{
			Key:      "dynamiclabel",
			Operator: v1beta1.LabelSelectorOpIn,
			Values:   []string{"${pvc.labels.dynamiclabel}"},
		},
	}

	vpsSpec := vpsutil.VolumeAntiAffinityByMatchExpression("mongo-vps", matchExpression)
	_, err := talisman.Instance().CreateVolumePlacementStrategy(&vpsSpec)
	m.spec = &vpsSpec
	return err
}

func (m *mongoVolumeAntiAffinityDynamicLabels) DestroyVPSDeployment() error {
	return talisman.Instance().DeleteVolumePlacementStrategy(m.spec.Name)
}

// mongoVPSAntiAffinityDynamicLabels is expecting to have deploy 2 replica of vol for each pod that has label [mongo-0, mongo-1]
// since this is antiaffinity, we are expecting that vol with the same labels are not deployed on the same pool/node.
// To validate that, we get the label from each deployed vol and extract the pool it's deployed on. if deployed correctly,
// there should be two pools per label.
func (m *mongoVolumeAntiAffinityDynamicLabels) ValidateVPSDeployment(contexts []*scheduler.Context) error {
	vols, err := Inst().S.GetVolumes(contexts[0])
	if err != nil {
		return err
	}
	apiVols, err := getApiVols(vols)
	if err != nil {
		return err
	}

	volumeLabelKey := "dynamiclabel"
	expectedNodeLength := 2

	return vpsutil.ValidateVolumeAntiAffinityByNode(apiVols, volumeLabelKey, expectedNodeLength)
}

type mongoVolumeAffinity struct {
	VolumePlacementStrategySpec
}

func (m *mongoVolumeAffinity) TestName() string {
	return "mongovolumeaffinity"
}

func (m *mongoVolumeAffinity) DeployVPS() error {

	matchExpression := []*v1beta1.LabelSelectorRequirement{
		{
			Key:      "app",
			Operator: v1beta1.LabelSelectorOpIn,
			Values:   []string{"mongo-sts"},
		},
	}

	vpsSpec := vpsutil.VolumeAffinityByMatchExpression("mongo-vps", matchExpression)
	_, err := talisman.Instance().CreateVolumePlacementStrategy(&vpsSpec)
	m.spec = &vpsSpec
	return err
}

func (m *mongoVolumeAffinity) DestroyVPSDeployment() error {
	return talisman.Instance().DeleteVolumePlacementStrategy(m.spec.Name)
}

// mongoVolumeAffinity is expecting to have deploy 2 replica of vol for each pod that has label app=mongo-sts
// since this is affinity, we are expecting that vol with the same labels are not deployed on the same pool/node.
// to validate that, we get the label from each deployed vol and extracts the pool it's deployed on. if deployed correctly,
// there should be one pools per label only.
func (m *mongoVolumeAffinity) ValidateVPSDeployment(contexts []*scheduler.Context) error {
	vols, err := Inst().S.GetVolumes(contexts[0])
	if err != nil {
		return err
	}

	apiVols, err := getApiVols(vols)
	if err != nil {
		return err
	}

	volumeLabelKey := "app"

	return vpsutil.ValidateVolumeAffinityByNode(apiVols, volumeLabelKey)
}

type mongoVPSReplicaAffinity struct {
	VolumePlacementStrategySpec
	deployedNode node.Node
}

func (m *mongoVPSReplicaAffinity) TestName() string {
	return "mongovpsreplicaaffinity"
}

func putNodeLabels(node node.Node, nodeLabels map[string]string) error {
	for key, value := range nodeLabels {
		err := Inst().S.AddLabelOnNode(node, key, value)
		if err != nil {
			return err
		}
	}
	return nil
}

func removeNodeLabels(node node.Node, nodeLabels map[string]string) error {
	for key := range nodeLabels {
		err := Inst().S.RemoveLabelOnNode(node, key)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *mongoVPSReplicaAffinity) getNodeLabels() map[string]string {
	return map[string]string{
		"nodelabel": "affinity",
	}
}

func (m *mongoVPSReplicaAffinity) DeployVPS() error {
	storageNodes := node.GetStorageDriverNodes()
	m.deployedNode = storageNodes[rand.Intn(len(storageNodes))]

	err := putNodeLabels(m.deployedNode, m.getNodeLabels())
	if err != nil {
		return err
	}

	matchExpression := []*v1beta1.LabelSelectorRequirement{
		{
			Key:      "nodelabel",
			Operator: v1beta1.LabelSelectorOpIn,
			Values:   []string{"affinity"},
		},
	}

	vpsSpec := vpsutil.ReplicaAffinityByMatchExpression("mongo-vps", matchExpression)
	_, err = talisman.Instance().CreateVolumePlacementStrategy(&vpsSpec)
	m.spec = &vpsSpec
	return err
}

func (m *mongoVPSReplicaAffinity) ValidateVPSDeployment(contexts []*scheduler.Context) error {
	vols, err := Inst().S.GetVolumes(contexts[0])
	if err != nil {
		return err
	}

	apiVols, err := getApiVols(vols)
	if err != nil {
		return err
	}

	return vpsutil.ValidateReplicaAffinityByNode(apiVols, m.deployedNode)
}

func (m *mongoVPSReplicaAffinity) DestroyVPSDeployment() error {
	err := removeNodeLabels(m.deployedNode, m.getNodeLabels())
	if err != nil {
		return err
	}
	return talisman.Instance().DeleteVolumePlacementStrategy(m.spec.Name)
}

type mongoVPSReplicaAntiAffinity struct {
	VolumePlacementStrategySpec
	deployedNode node.Node
}

func (m *mongoVPSReplicaAntiAffinity) TestName() string {
	return "mongovpsreplicaantiaffinity"
}

func (m *mongoVPSReplicaAntiAffinity) getNodeLabels() map[string]string {
	return map[string]string{
		"nodelabel": "antiaffinity",
	}
}

func (m *mongoVPSReplicaAntiAffinity) DeployVPS() error {
	storageNodes := node.GetStorageDriverNodes()
	m.deployedNode = storageNodes[rand.Intn(len(storageNodes))]

	for _, labeledNode := range storageNodes {
		if labeledNode.Name != m.deployedNode.Name {
			err := putNodeLabels(labeledNode, m.getNodeLabels())
			if err != nil {
				return err
			}
		}
	}

	matchExpression := []*v1beta1.LabelSelectorRequirement{
		{
			Key:      "nodelabel",
			Operator: v1beta1.LabelSelectorOpIn,
			Values:   []string{"antiaffinity"},
		},
	}
	vpsSpec := vpsutil.ReplicaAntiAffinityByMatchExpression("mongo-vps", matchExpression)
	_, err := talisman.Instance().CreateVolumePlacementStrategy(&vpsSpec)
	m.spec = &vpsSpec
	return err
}

func (m *mongoVPSReplicaAntiAffinity) ValidateVPSDeployment(contexts []*scheduler.Context) error {
	vols, err := Inst().S.GetVolumes(contexts[0])
	if err != nil {
		return err
	}

	apiVols, err := getApiVols(vols)
	if err != nil {
		return err
	}

	return vpsutil.ValidateReplicaAffinityByNode(apiVols, m.deployedNode)
}

func (m *mongoVPSReplicaAntiAffinity) DestroyVPSDeployment() error {
	storageNodes := node.GetStorageDriverNodes()

	for _, labeledNode := range storageNodes {
		if labeledNode.Name != m.deployedNode.Name {
			err := removeNodeLabels(labeledNode, m.getNodeLabels())
			if err != nil {
				return err
			}
		}
	}
	return talisman.Instance().DeleteVolumePlacementStrategy(m.spec.Name)
}

func getApiVols(vols []*volume.Volume) ([]*api.Volume, error) {
	var apiVols []*api.Volume
	for _, vol := range vols {
		vol, err := Inst().V.InspectVolume(vol.ID)
		if err != nil {
			return nil, err
		}
		apiVols = append(apiVols, vol)
	}
	return apiVols, nil
}

var _ = Describe("{ValidateVPSruleWithPoolRestriction}", func() {
	/*
		1.	select two random pools and get their uid
		2.	label them mediatype=SSD and mediatype=SATA
		3.	Apply the vps,storage class, and pvc
		4.	check if replica 1 created in pool labeled SSD and replica 2 on pool labeled SATA
		5.  remove the labels and delete the pvc, stc, vps
	*/
	var testrailID = 0
	var runID int
	JustBeforeEach(func() {
		StartTorpedoTest("ValidateVPSruleWithPoolRestriction", "Validate behavior when VPS rules are used in conjunction with storage pool restrictions", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})

	stepLog := "Test adding label for two pools and apply vps rule to create volumes in the labeled pool"
	It(stepLog, func() {

		var selectedPools []string
		stepLog = "Select two random pools uuid"
		Step(stepLog, func() {
			log.Infof(stepLog)
			poolsAvailable, err := Inst().V.ListStoragePools(metav1.LabelSelector{})
			log.FailOnError(err, "Failed to list storage pools")
			log.Infof("List of pools present in the cluster [%v]", poolsAvailable)
			var poolIDs []string
			for _, v := range poolsAvailable {
				log.Infof("each pool details [%v]", v)
				log.Infof("pool id is [%v]", v.Uuid)
				poolIDs = append(poolIDs, v.Uuid)
			}
			if len(poolIDs) < 2 {
				log.FailOnError(fmt.Errorf("Not enough pools to select from."), "Failled to select random pools")
			}
			selectedPools = append(selectedPools, poolIDs[0], poolIDs[1])
			log.Infof("Selected pools for add label [%v]", selectedPools)
		})

		stepLog = "Get node details by pool uuid and add pool label"
		poolLabelToUpdate := make(map[string]string)
		labels := []string{"SSD", "SATA"}
		Step(stepLog, func() {
			log.Infof(stepLog)
			for i, poolID := range selectedPools {
				storageNode, err := GetNodeWithGivenPoolID(poolID)
				log.FailOnError(err, "Failed to get node with given pool ID")
				poolLabelToUpdate["mediatype"] = labels[i]
				err = Inst().V.UpdatePoolLabels(*storageNode, poolID, poolLabelToUpdate)
				dash.VerifyFatal(err, nil, "Check if able to update the label on the pool")
			}
		})

		stepLog = "Apply volume placement strategy"
		vpsName := fmt.Sprintf("mongo-vps-%v", time.Now().Unix())
		Step(stepLog, func() {
			log.Infof(stepLog)
			vpsSpec := vpsutil.ReplicaAffinityPool(vpsName)
			_, err = talisman.Instance().CreateVolumePlacementStrategy(&vpsSpec)
			dash.VerifyFatal(err, nil, "Check if able to apply volume placement strategy")
		})

		scName := fmt.Sprintf("mongo-sc-%v", time.Now().Unix())
		params := make(map[string]string)
		stepLog = "Apply storage class"
		k8sStorage := storage.Instance()
		Step(stepLog, func() {
			log.Infof(stepLog)
			params["repl"] = "2"
			params["placement_strategy"] = vpsName
			v1obj := metav1.ObjectMeta{
				Name: scName,
			}
			bindMode := storageApi.VolumeBindingImmediate
			scObj := storageApi.StorageClass{
				ObjectMeta:        v1obj,
				Provisioner:       k8s.CsiProvisioner,
				Parameters:        params,
				VolumeBindingMode: &bindMode,
			}
			_, err := k8sStorage.CreateStorageClass(&scObj)
			dash.VerifyFatal(err, nil, "Verifying creation of new storage class")
		})

		stepLog = "Apply persistent volume claim"
		pvcName := fmt.Sprintf("mongo-pvc-%v", time.Now().Unix())
		namespace := "default"
		Step(stepLog, func() {
			log.Infof(stepLog)
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
			dash.VerifyFatal(err, nil, "Verifying creation of new storage class")
		})

		_, err := k8sCore.GetPersistentVolumeClaim(pvcName, namespace)
		log.FailOnError(err, "Failed to get pvc")
		err = Inst().S.WaitForSinglePVCToBound(pvcName, namespace, 3)
		log.FailOnError(err, "Failed to wait for pvc to bound")

		stepLog = "Validate replicas created on the given pools"
		Step(stepLog, func() {
			log.Infof(stepLog)
			var poolsFromRepl []string

			log.Infof("Get volume list")
			volIDs, err := Inst().V.ListAllVolumes()
			log.FailOnError(err, "Failed to get volumes")
			log.Infof("volume IDs list: %v", volIDs)

			log.Infof("Get volume details by ID")
			apiVol, err := Inst().V.InspectVolume(volIDs[0])
			log.FailOnError(err, "Failed to inspect volume details")
			log.Infof("Volume created by pvc: %v", apiVol)

			log.Infof("Get replicas from the volume")
			replicaSets := apiVol.ReplicaSets
			log.Infof("Replica details for the volume: %v, %v", volIDs[0], replicaSets)

			for _, rel := range replicaSets {
				poolsFromRepl = append(poolsFromRepl, rel.PoolUuids...)
			}
			log.Infof("Created replicasets pool IDs: %v", poolsFromRepl)

			if len(poolsFromRepl) != 2 {
				log.FailOnError(fmt.Errorf("Created replicas count not matching"), "Failed to compare replica count : %v", poolsFromRepl)
			}

			for _, poolID := range poolsFromRepl {
				if poolID != selectedPools[0] && poolID != selectedPools[1] {
					log.FailOnError(fmt.Errorf("Replicas not created on the given pools"), "Failed to compare replicas %v with [%v or %v]", poolsFromRepl, selectedPools[0], selectedPools[1])
				}
				dash.VerifyFatal(poolID == selectedPools[0] || poolID == selectedPools[1], true, fmt.Sprintf("Replica should created in the given pools %v or %v", selectedPools[0], selectedPools[1]))
			}

		})

		stepLog = "Remove all newly created specs"
		Step(stepLog, func() {
			log.Infof(stepLog)
			err = core.Instance().DeletePersistentVolumeClaim(pvcName, namespace)
			log.FailOnError(err, "Failed to remove pvc: %v", pvcName)

			log.Infof("Deleting the newly created storage class")
			err = k8sStorage.DeleteStorageClass(scName)
			log.FailOnError(err, "Failed to remove storage class: %v", scName)

			log.Infof("Deleting the newly created VPS")
			err = talisman.Instance().DeleteVolumePlacementStrategy(vpsName)
			log.FailOnError(err, "Failed to remove VPS: %v", vpsName)

			log.Infof("Deleting the newly created labels on the selected pools")
			for _, poolID := range selectedPools {
				storageNode, err := GetNodeWithGivenPoolID(poolID)
				log.FailOnError(err, "Failed to get node with given pool ID")
				poolLabelToUpdate["mediatype"] = ""
				err = Inst().V.UpdatePoolLabels(*storageNode, poolID, poolLabelToUpdate)
				dash.VerifyFatal(err, nil, "Check if able to delete the label on the pool")
			}
		})
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})
})

var _ = Describe("{ValidateVPSAffinityAndAntiAffinityConflict}", func() {
	/*
		1.	Label all node zone=A
		2.	Apply the volume placement strategy,storage class, and pvc
		3.	Verify volume is not provisioned.
		4.  remove all labels and delete the pvc, stc, vps
	*/
	var testrailID = 0
	var runID int
	JustBeforeEach(func() {
		StartTorpedoTest("ValidateVPSAffinityAndAntiAffinityConflict", "Validate VPS handles conflicting affinity and anti-affinity rules.", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})

	stepLog := "Test conflict for apply affinity and anti-affinity on the same node"
	It(stepLog, func() {
		log.InfoD(stepLog)
		var (
			scName     = fmt.Sprintf("mongo-sc-%v", time.Now().Unix())
			vpsName    = fmt.Sprintf("mongo-vps-%v", time.Now().Unix())
			pvcName    = fmt.Sprintf("mongo-pvc-%v", time.Now().Unix())
			namespace  = "default"
			params     = make(map[string]string)
			k8sStorage = storage.Instance()
			nodes      []node.Node
		)
		stepLog = "Get all nodes and add label"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			nodes = node.GetWorkerNodes()
			log.Infof("node list: %v", nodes)
			for _, node := range nodes {
				log.Infof("Adding label for the node: %v", node.Name)
				err := k8sCore.AddLabelOnNode(node.Name, "zone", "A")
				log.FailOnError(err, "Failed to add label for the node: %v", node)
			}
		})

		stepLog = "Apply volume placement strategy"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			vpsSpec := vpsutil.ReplicaAffinityandAntiAffinity(vpsName)
			_, err = talisman.Instance().CreateVolumePlacementStrategy(&vpsSpec)
			dash.VerifyFatal(err, nil, "Check if able to apply volume placement strategy")
		})

		stepLog = "Apply storage class"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			params["repl"] = "2"
			params["placement_strategy"] = vpsName
			v1obj := metav1.ObjectMeta{
				Name: scName,
			}
			bindMode := storageApi.VolumeBindingImmediate
			scObj := storageApi.StorageClass{
				ObjectMeta:        v1obj,
				Provisioner:       k8s.CsiProvisioner,
				Parameters:        params,
				VolumeBindingMode: &bindMode,
			}
			_, err := k8sStorage.CreateStorageClass(&scObj)
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
			dash.VerifyFatal(err, nil, "Verifying creation of new storage class")
		})

		time.Sleep(1 * time.Minute)
		log.Infof("Waiting for a minute to get pvc status")

		createdPVC, err := k8sCore.GetPersistentVolumeClaim(pvcName, namespace)
		log.FailOnError(err, "Failed to get pvc")
		log.Infof("Created PVC details %v", createdPVC.Status)

		if createdPVC.Status.Phase == "Pending" {
			log.Infof("PVC status: %v", createdPVC.Status.Phase)
			for _, event := range Inst().S.GetEvents()["PersistentVolumeClaim"] {
				log.Infof("PVC Event: %v", event)
				if strings.Contains(event.Message, "Waiting for a volume to be created") {
					log.Infof("Volume creation error is: %v", event)
					dash.VerifyFatal(strings.Contains(event.Message, "Waiting for a volume to be created"), true, "Check if volume creation status pending reason")
				}
				if strings.Contains(event.Message, "failed to provision volume with StorageClass") {
					log.Infof("Volume creation error is: %v", event)
					dash.VerifyFatal(createdPVC.Status.Phase == "Pending", true, "Check if volume creation status is pending")

					errorMsg := fmt.Sprintf("pools could not be selected because they did not satisfy the following requirement: placement rule: enforcement: required expressions: label key=zone In values [A];")

					dash.VerifyFatal(strings.Contains(event.Message, errorMsg), true, "Check volume not created")
				}
			}
		}

		stepLog = "Remove all newly created specs"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			err = core.Instance().DeletePersistentVolumeClaim(pvcName, namespace)
			log.FailOnError(err, "Failed to remove pvc: %v", pvcName)

			log.Infof("Deleting the newly created storage class")
			err = k8sStorage.DeleteStorageClass(scName)
			log.FailOnError(err, "Failed to remove storage class: %v", scName)

			log.Infof("Deleting the newly created VPS")
			err = talisman.Instance().DeleteVolumePlacementStrategy(vpsName)
			log.FailOnError(err, "Failed to remove VPS: %v", vpsName)

			log.Infof("Deleting the newly created labels from node")
			for _, node := range nodes {
				err = Inst().S.RemoveLabelOnNode(node, "zone")
				log.FailOnError(err, "Failed to remove label from node")
			}
		})
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})
})
