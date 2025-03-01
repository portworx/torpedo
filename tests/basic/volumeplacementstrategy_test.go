package tests

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/libopenstorage/openstorage/api"
	. "github.com/onsi/ginkgo/v2"
	"github.com/portworx/sched-ops/k8s/talisman"
	"github.com/portworx/talisman/pkg/apis/portworx/v1beta1"
	"github.com/portworx/talisman/pkg/apis/portworx/v1beta2"
	"github.com/pure-px/sched-ops/k8s/core"
	"github.com/pure-px/sched-ops/k8s/storage"
	storkv1 "github.com/pure-px/stork/pkg/apis/stork/v1alpha1"
	storkops "github.com/pure-px/stork/pkg/crud/stork"
	"github.com/pure-px/torpedo/drivers/node"
	"github.com/pure-px/torpedo/drivers/scheduler"
	"github.com/pure-px/torpedo/drivers/scheduler/k8s"
	"github.com/pure-px/torpedo/drivers/volume"
	"github.com/pure-px/torpedo/pkg/log"
	"github.com/pure-px/torpedo/pkg/testrailuttils"
	"github.com/pure-px/torpedo/pkg/vpsutil"
	. "github.com/pure-px/torpedo/tests"
	corev1 "k8s.io/api/core/v1"
	storageApi "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("{VolumePlacementStrategyFunctional}", Label("p0", "positive", "VPS"), func() {
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

var _ = Describe("{ValidateVPSruleWithPoolRestriction}", Label("p0", "positive", "VPS"), func() {
	/*
		1.	select two random pools and get their uid
		2.	label them mediatype=SSD and mediatype=SATA
		3.	Apply the vps,storage class, and pvc
		4.	check if replica 1 created in pool labeled SSD and replica 2 on pool labeled SATA
		5.  remove the labels and delete the pvc, stc, vps
	*/
	var (
		testrailID = 0
		runID      int
		contexts   = make([]*scheduler.Context, 0)
	)
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

var _ = Describe("{ValidateVPSAffinityAndAntiAffinityConflict}", Label("p0", "positive", "VPS"), func() {
	/*
		1.	Label all node zone=A
		2.	Apply the volume placement strategy,storage class, and pvc
		3.	Verify volume is not provisioned.
		4.  remove all labels and delete the pvc, stc, vps
	*/
	var (
		testrailID = 0
		runID      int
		contexts   = make([]*scheduler.Context, 0)
	)
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
				log.FailOnError(err, "Failed to remove label from node.")
			}
		})
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})
})

var _ = Describe("{ValidateVPSFailOnInsufficientPools}", Label("p0", "VPS", "staging"), func() {
	/*
		1.	select two random pools and get their uid
		2.	label them mediatype=SSD
		3.	Apply the vps, storage class, and pvc with three replicas
		4.	check if volume creation should fail when VolumePlacementStrategy fails to find enough pools
		5.  remove the labels and delete the pvc, vps
	*/
	var (
		testrailID = 0
		runID      int
		contexts   = make([]*scheduler.Context, 0)
	)
	JustBeforeEach(func() {
		StartTorpedoTest("ValidateVPSFailOnInsufficientPools", "Validate Volume Creation behavior when VPS rules are providing insufficient pools", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})

	stepLog := "Test adding label for two pools and apply vps rule to create volumes with three replicas in the labeled pool"
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
		Step(stepLog, func() {
			log.Infof(stepLog)
			for _, poolID := range selectedPools {
				storageNode, err := GetNodeWithGivenPoolID(poolID)
				log.FailOnError(err, "Failed to get node with given pool ID")
				poolLabelToUpdate["mediatype"] = "SSD"
				err = Inst().V.UpdatePoolLabels(*storageNode, poolID, poolLabelToUpdate)
				dash.VerifyFatal(err, nil, "Check if able to update the label on the pool")
			}
		})

		stepLog = "Apply volume placement strategy"
		vpsName := fmt.Sprintf("mongo-vps-%v", time.Now().Unix())
		Step(stepLog, func() {
			log.Infof(stepLog)
			matchExpression := []*v1beta1.LabelSelectorRequirement{
				{
					Key:      "mediatype",
					Operator: v1beta1.LabelSelectorOpIn,
					Values:   []string{"SSD"},
				},
			}
			vpsSpec := vpsutil.ReplicaAffinityByMatchExpression(vpsName, matchExpression)
			_, err = talisman.Instance().CreateVolumePlacementStrategy(&vpsSpec)
			dash.VerifyFatal(err, nil, "Check if able to apply volume placement strategy")
		})

		scName := fmt.Sprintf("mongo-sc-%v", time.Now().Unix())
		params := make(map[string]string)
		stepLog = "Apply storage class"
		k8sStorage := storage.Instance()
		Step(stepLog, func() {
			log.Infof(stepLog)
			params["repl"] = "3"
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
		err = Inst().S.WaitForSinglePVCToBound(pvcName, namespace, 5)
		if err != nil {
			log.InfoD("Volume creation failed successfully when VolumePlacementStrategy fails to find enough pools")
		}

		stepLog = "Remove all newly created specs"
		Step(stepLog, func() {
			log.Infof(stepLog)
			err = core.Instance().DeletePersistentVolumeClaim(pvcName, namespace)
			log.FailOnError(err, "Failed to remove pvc: %v", pvcName)

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

var _ = Describe("{VolumeCloneWithDifferentPlacementStrategy}", Label("p0", "positive", "VPS", "staging"), func() {

	/*	1. Create 2 volume placement strategies, vps-1, vps-2
		2. create 2 storage classes using the sc-1(uses vps-1), sc-2 (uses vps-2) (make sure this sc is repl 3)
		3. Create a pvc
		4. validate if vps is applied
		5. Take snapshot of this pvc.
		6. Clone this snapshot and use the sc-2
		7. Now validate if vps-2 is applied */

	var (
		testrailID = 0
		runID      int
		namespace  = "default"
		contexts   = make([]*scheduler.Context, 0)
	)

	JustBeforeEach(func() {
		StartTorpedoTest("VolumeCloneWithDifferentPlacementStrategy", "Validate VPS when creating clones with different placement strategies", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})

	stepLog := "Volume Clone with Placement Strategy"
	It(stepLog, func() {
		stepLog = "Adding Labels on Nodes"
		var nodes []node.Node
		nodes = node.GetStorageNodes()
		zoneANodeCount := 0
		zoneBNodeCount := 0
		Step(stepLog, func() {
			log.InfoD(stepLog)
			// Iterate over the nodes and apply labels
			for i, node := range nodes {
				log.Infof("Adding labels for the node: %v", node.Name)
				if i%2 == 0 {
					// Apply 'zone=A' label for  nodes 1, 3, 5...
					err := k8sCore.AddLabelOnNode(node.Name, "zone", "A")
					log.FailOnError(err, "Failed to add label 'zone=A' for node: %v", node)
					log.Infof("Successfully added label 'zone=A' to node: %v", node.Name)
					zoneANodeCount++
				} else {
					// Apply 'zone=B' label for  nodes 2, 4,6...
					err := k8sCore.AddLabelOnNode(node.Name, "zone", "B")
					log.FailOnError(err, "Failed to add label 'zone=B' for node: %v", node)
					log.Infof("Successfully added label 'zone=B' to node: %v", node.Name)
					zoneBNodeCount++
				}
			}

		})

		stepLog = "Creating VPS-1 and VPS-2 with different placement strategies"
		var (
			vpsName1 = fmt.Sprintf("mongo-vps-1-%v", time.Now().Unix())
			vpsName2 = fmt.Sprintf("mongo-vps-2-%v", time.Now().Unix())
		)
		Step(stepLog, func() {
			log.InfoD(stepLog)
			// Define match expressions for VPS-1 and VPS-2
			matchExpression1 := []*v1beta1.LabelSelectorRequirement{
				{
					Key:      "zone",
					Operator: v1beta1.LabelSelectorOpIn,
					Values:   []string{"A"},
				},
			}
			matchExpression2 := []*v1beta1.LabelSelectorRequirement{
				{
					Key:      "zone",
					Operator: v1beta1.LabelSelectorOpIn,
					Values:   []string{"B"},
				},
			}
			// Create VPS-1 with a match expression for nodes with "zone=A"
			log.Infof("Creating VPS-1 with placement strategy")
			vpsSpec1 := vpsutil.ReplicaAffinityByMatchExpression(vpsName1, matchExpression1)
			_, err := talisman.Instance().CreateVolumePlacementStrategy(&vpsSpec1)
			log.FailOnError(err, "Failed to apply VPS-1 placement strategy")
			// Create VPS-2 with a match expression for nodes with "zone=B"
			log.Infof("Creating VPS-2 with placement strategy")
			vpsSpec2 := vpsutil.ReplicaAffinityByMatchExpression(vpsName2, matchExpression2)
			_, err = talisman.Instance().CreateVolumePlacementStrategy(&vpsSpec2)
			log.FailOnError(err, "Failed to apply VPS-2 placement strategy")
		})

		stepLog = "Creating Storage Classes using VPS-1 and VPS-2"
		var (
			scName1          = fmt.Sprintf("mongo-sc1-%v", time.Now().Unix())
			scName2          = fmt.Sprintf("mongo-sc2-%v", time.Now().Unix())
			params1, params2 = map[string]string{}, map[string]string{}
			k8sStorage       = storage.Instance()
			bindMode         = storageApi.VolumeBindingImmediate
		)
		Step(stepLog, func() {
			zoneARepl := "2"
			if zoneANodeCount < 2 {
				zoneARepl = "1"
			}
			zoneBRepl := "1"
			if zoneBNodeCount > 2 {
				zoneBRepl = "3"
			} else {
				zoneBRepl = fmt.Sprintf("%d", zoneBNodeCount)
			}
			log.InfoD(stepLog)
			// Create sc-1 with VPS-1
			log.Infof("Creating sc-1 using VPS-1")
			params1["repl"] = zoneARepl
			params1["placement_strategy"] = vpsName1
			scObj1 := storageApi.StorageClass{
				ObjectMeta:        metav1.ObjectMeta{Name: scName1},
				Provisioner:       k8s.CsiProvisioner,
				Parameters:        params1,
				VolumeBindingMode: &bindMode,
			}
			_, err := k8sStorage.CreateStorageClass(&scObj1)
			log.FailOnError(err, "Failed to create sc-1")
			// Create sc-2 with VPS-2
			log.Infof("Creating sc-2 using VPS-2")
			params2["repl"] = zoneBRepl
			params2["placement_strategy"] = vpsName2
			scObj2 := storageApi.StorageClass{
				ObjectMeta:        metav1.ObjectMeta{Name: scName2},
				Provisioner:       k8s.CsiProvisioner,
				Parameters:        params2,
				VolumeBindingMode: &bindMode,
			}
			_, err = k8sStorage.CreateStorageClass(&scObj2)
			log.FailOnError(err, "Failed to create sc-2")
		})

		stepLog = "Creating PVC using sc-1 (vps-1)"
		var pvcName = fmt.Sprintf("mongo-pvc-original-%v", time.Now().Unix())
		Step(stepLog, func() {
			log.InfoD(stepLog)
			_, err := core.Instance().CreatePersistentVolumeClaim(&corev1.PersistentVolumeClaim{
				TypeMeta:   metav1.TypeMeta{Kind: "PersistentVolumeClaim"},
				ObjectMeta: metav1.ObjectMeta{Name: pvcName},
				Spec: corev1.PersistentVolumeClaimSpec{
					AccessModes:      []corev1.PersistentVolumeAccessMode{corev1.ReadWriteMany},
					StorageClassName: &scName1,
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{corev1.ResourceStorage: resource.MustParse("1Gi")},
					},
				},
			})
			log.FailOnError(err, "Failed to create PVC with sc-1")
		})

		stepLog = "Check PVC status and verify volume replica placement on nodes with label 'zone=A'"
		Step(stepLog, func() {
			log.InfoD(stepLog)

			//Waiting for PVC to Bound
			err = Inst().S.WaitForSinglePVCToBound(pvcName, namespace, 3)
			log.FailOnError(err, "Failed to wait for pvc to bound")
			// Get the PVC that was created using sc-1
			createdPVC, err := k8sCore.GetPersistentVolumeClaim(pvcName, namespace)
			log.FailOnError(err, "Failed to get PVC")
			log.Infof("Created PVC details: %v", createdPVC.Status)

			// Check PVC status - Ensure PVC is bound
			dash.VerifyFatal(createdPVC.Status.Phase == "Bound", true,
				fmt.Sprintf("PVC should be in 'Bound' status'%v'", createdPVC.Status.Phase))

			// Get the PV bound to the PVC
			pv, err := core.Instance().GetPersistentVolume(createdPVC.Spec.VolumeName)
			log.FailOnError(err, "Failed to get PersistentVolume")
			log.Infof("Persistent Volume details: %v", pv)

			// List all volumes and check replica placement
			volIDs, err := Inst().V.ListAllVolumes()
			log.FailOnError(err, "Failed to get volumes")
			log.Infof("Volume IDs list: %v", volIDs)

			for _, volId := range volIDs {
				apiVol, err := Inst().V.InspectVolume(volId)
				log.FailOnError(err, "Failed to inspect volume details")
				log.Infof("Inspecting volume ID: %s", volId)
				log.Infof("Found volume %s for createdPVC PVC", createdPVC.Name)

				if apiVol.Locator.VolumeLabels["pvc"] == createdPVC.Name {
					var nodeList []string
					for _, replica := range apiVol.ReplicaSets {
						nodeList = append(nodeList, replica.Nodes...)
					}
					// check if volume replicas are placed on nodes with the label 'zone=A'
					for _, nodeName := range nodeList {
						nodeID, err := node.GetNodeDetailsByNodeID(nodeName)
						log.FailOnError(err, "unable to find ID")
						nodeLabels, err := k8sCore.GetLabelsOnNode(nodeID.Name)
						log.FailOnError(err, "unable to find the node")
						labelValue, exists := nodeLabels["zone"]
						if exists {
							log.Infof("Found zone label: %s", exists)
						}
						dash.VerifyFatal(labelValue == "A", true,
							fmt.Sprintf("Node '%s' has label 'zone=A'. Found label: '%s'", nodeName, labelValue))
					}

				}
			}

		})

		stepLog = "Create local snapshot schedule"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			policyName := "intervalpolicy"
			log.Infof("Creating a interval schedule policy %v with interval %v minutes", policyName, 5)
			log.InfoD(stepLog)
			schedPolicy, err := storkops.Instance().GetSchedulePolicy(policyName)
			retain := 8
			interval := 5
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
		})

		stepLog = "Clone the snapshot & Create PVC from snapshot"
		clonedPVCName := fmt.Sprintf("mongo-cloned-pvc-%v", time.Now().Unix())
		Step(stepLog, func() {
			log.InfoD(stepLog)

			// After snapshotName has been created earlier for PVC-1
			snapshotName := fmt.Sprintf("%s-snapshot", pvcName)
			apiGroup := "stork.libopenstorage.org/v1alpha1"

			// Create a new PVC from the snapshot of PVC-1
			_, err := core.Instance().CreatePersistentVolumeClaim(&corev1.PersistentVolumeClaim{
				TypeMeta: metav1.TypeMeta{
					Kind: "PersistentVolumeClaim",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: clonedPVCName,
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					AccessModes:      []corev1.PersistentVolumeAccessMode{corev1.ReadWriteMany},
					StorageClassName: &scName2,
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{corev1.ResourceStorage: resource.MustParse("3Gi")},
					},
					DataSource: &corev1.TypedLocalObjectReference{
						APIGroup: &apiGroup,
						Kind:     "VolumeSnapshot",
						Name:     snapshotName,
					},
				},
			})

			log.FailOnError(err, "Failed to create PVC from snapshot")
			log.Infof("Successfully created PVC %s from snapshot %s", clonedPVCName, snapshotName)
		})

		stepLog := "Validate VPS-2 is applied to the cloned PVC"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			//Waiting for PVC to Bound
			err = Inst().S.WaitForSinglePVCToBound(clonedPVCName, namespace, 3)
			log.FailOnError(err, "Failed to wait for pvc to bound")

			// Get the cloned PVC and log the details
			clonedPVC, err := k8sCore.GetPersistentVolumeClaim(clonedPVCName, namespace)
			log.FailOnError(err, "Failed to get cloned PVC")
			log.Infof("Cloned PVC details: %v, status: %v", clonedPVC, clonedPVC.Status.Phase)

			// Check if PVC is Bound
			dash.VerifyFatal(clonedPVC.Status.Phase == "Bound", true, fmt.Sprintf("Cloned PVC should be in 'Bound' status'%v'", clonedPVC.Status.Phase))

			// Get the PV bound to the cloned PVC
			pv, err := core.Instance().GetPersistentVolume(clonedPVC.Spec.VolumeName)
			log.FailOnError(err, "Failed to get PersistentVolume for cloned PVC")
			log.Infof("Persistent Volume details: %v", pv)

			// Check volume placement (replica node labels) for each volume
			volIDs, err := Inst().V.ListAllVolumes()
			log.FailOnError(err, "Failed to get volumes")
			log.Infof("Volume IDs list: %v", volIDs)

			for _, volId := range volIDs {
				apiVol, err := Inst().V.InspectVolume(volId)
				log.FailOnError(err, "Failed to inspect volume details")
				log.Infof("Inspecting volume ID: %s", volId)
				if apiVol.Locator.VolumeLabels["pvc"] == clonedPVC.Name {
					log.Infof("Inspecting volume: %s for cloned PVC", clonedPVC.Name)
					// Gather replica node names
					var nodeList []string
					for _, replica := range apiVol.ReplicaSets {
						nodeList = append(nodeList, replica.Nodes...)
					}
					for _, nodeName := range nodeList {
						nodeID, err := node.GetNodeDetailsByNodeID(nodeName)
						log.FailOnError(err, "unable to find ID")
						nodeLabels, err := k8sCore.GetLabelsOnNode(nodeID.Name)
						log.FailOnError(err, "unable to find the node")
						labelValue, exists := nodeLabels["zone"]
						if exists {
							log.Infof("Found zone label: %s", exists)
						}
						dash.VerifyFatal(labelValue == "B", true,
							fmt.Sprintf("Node '%s' has label 'zone=B'. Found label: '%s'", nodeName, labelValue))
					}
				}
			}

		})

		stepLog = "Remove all newly created specs"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			err = core.Instance().DeletePersistentVolumeClaim(pvcName, namespace)
			log.FailOnError(err, "Failed to remove pvc: %v", pvcName)
			err = core.Instance().DeletePersistentVolumeClaim(clonedPVCName, namespace)
			log.FailOnError(err, "Failed to remove pvc: %v", pvcName)

			log.Infof("Deleting the newly created storage class")
			err = k8sStorage.DeleteStorageClass(scName1)
			log.FailOnError(err, "Failed to remove storage class: %v", scName1)
			err = k8sStorage.DeleteStorageClass(scName2)
			log.FailOnError(err, "Failed to remove storage class: %v", scName2)

			log.Infof("Deleting the newly created VPS")
			err = talisman.Instance().DeleteVolumePlacementStrategy(vpsName1)
			log.FailOnError(err, "Failed to remove VPS: %v", vpsName1)
			err = talisman.Instance().DeleteVolumePlacementStrategy(vpsName2)
			log.FailOnError(err, "Failed to remove VPS: %v", vpsName2)

			for _, node := range nodes {
				log.Infof("Removing labels for the node: %v", node.Name)
				// Remove the 'zone' label from all nodes (whether it is 'A' or 'B')
				err := k8sCore.RemoveLabelOnNode(node.Name, "zone")
				log.FailOnError(err, "Failed to remove label 'zone' for node: %v", node)
				log.Infof("Successfully removed label 'zone' from node: %v", node.Name)
			}
		})
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()

		AfterEachTest(contexts, testrailID, runID)
	})
})

var _ = Describe("{ValidateVPSAffinityWithoutRequiredLabel}", Label("staging", "p0", "negative", "VPS"), func() {
	/*
		https://purestorage.atlassian.net/browse/HAZEL-931
		1. Create a vps rule where no node has that particular label
		2. Create a volume using that rule
		3. Volume should be in pending state because it couldn’t find any node with that label.
	*/
	var (
		testrailID = 0
		runID      int
		contexts   = make([]*scheduler.Context, 0)
	)
	JustBeforeEach(func() {
		StartTorpedoTest("ValidateVPSAffinityWithoutRequiredLabel", "VPS with nodes in a zone not having the required labels", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})

	stepLog := "Test VPS with nodes in a zone not having the required labels"
	It(stepLog, func() {
		log.InfoD(stepLog)
		var (
			scName     = fmt.Sprintf("mongo-sc-%v", time.Now().Unix())
			vpsName    = fmt.Sprintf("mongo-vps-%v", time.Now().Unix())
			pvcName    = fmt.Sprintf("mongo-pvc-%v", time.Now().Unix())
			namespace  = "default"
			params     = make(map[string]string)
			k8sStorage = storage.Instance()
		)

		stepLog = "Apply volume placement strategy"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			matchExpression := []*v1beta1.LabelSelectorRequirement{
				{
					Key:      "zone",
					Operator: v1beta1.LabelSelectorOpIn,
					Values:   []string{"affinity"},
				},
			}

			vpsSpec := vpsutil.ReplicaAffinityByMatchExpression(vpsName, matchExpression)
			_, err = talisman.Instance().CreateVolumePlacementStrategy(&vpsSpec)
			dash.VerifyFatal(err, nil, "Check if able to apply volume placement strategy")
		})

		stepLog = "Apply storage class"
		Step(stepLog, func() {
			log.InfoD(stepLog)
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
					log.Infof("Volume creation error is: %v", event.Message)
					dash.VerifyFatal(createdPVC.Status.Phase == "Pending", true, "Check if volume creation status is pending")

					errorMsg := fmt.Sprintf("pools could not be selected because they did not satisfy the following requirement: placement rule: enforcement: required expressions: label key=zone In values [affinity];")
					dash.VerifyFatal(strings.Contains(event.Message, errorMsg), true, "Check volume not created")
				}
			}
		} else {
			errMsg := fmt.Errorf("Expected status: Pending. Actual status: [%s]", createdPVC.Status.Phase)
			log.FailOnError(errMsg, "Failed to validate PVC: [%v] in the namespace [%v]", pvcName, namespace)
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
		})
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})
})

// Setup VPS and validate that volumes in trashcan aren't considered when anti-affinity rules are set.
var _ = Describe("{SetupVPSValidateTrashcanAntiAffinity}", Label("p0", "positive", "pure_ops", "px_vol_ops", "px_ops", "staging"), func() {
	/*
		Label nodes with some tag
		Enable trashcan
		Create VPS
		Create some volumes using VPS and validate VPS is being followed
		Delete subset of these volumes
		Validate they are moved to trashcan
		Restore volumes
	*/
	JustBeforeEach(func() {
		StartTorpedoTest("SetupVPSValidateTrashcanAntiAffinity", "Setup VPS and validate that volumes in trashcan are not considered when anti-affinity rules are set", nil, 0)
		// Remove if node-type label is set before the test
		err = RemoveLabelsAllNodes(k8s.NodeType, true, false)
		log.FailOnError(err, "error removing label on node ")
	})
	var (
		contexts = make([]*scheduler.Context, 0)
	)

	stepLog := "Setup VPS and validate that volumes in trashcan are not considered when anti-affinity rules are set"
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
			currNode   node.Node
		)

		stepLog = "Enable Trashcan"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			currNode = node.GetStorageDriverNodes()[0]
			err := Inst().V.SetClusterOptsWithConfirmation(currNode, map[string]string{
				"--volume-expiration-minutes": "600",
			})
			log.FailOnError(err, "error while enabling trashcan")
			log.InfoD("Trashcan is successfully enabled")
		})

		stepLog = "Get all nodes and add label for some nodes"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			nodes = node.GetWorkerNodes()
			log.Infof("node list: %v", nodes)
			for i, node := range nodes {
				log.Infof("Adding label for the node: %v", node.Name)
				err := k8sCore.AddLabelOnNode(node.Name, "zone", "A")
				log.FailOnError(err, "Failed to add label for the node: %v", node)
				if i >= len(nodes)/2 {
					log.Infof("Added label to half of the nodes")
					break
				}
			}
		})

		stepLog = "Apply volume placement strategy"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			matchExpression := []*v1beta1.LabelSelectorRequirement{
				{
					Key:      "zone",
					Operator: v1beta1.LabelSelectorOpIn,
					Values:   []string{"A"},
				},
			}
			vpsSpec := vpsutil.ReplicaAntiAffinityByMatchExpression(vpsName, matchExpression)
			_, err = talisman.Instance().CreateVolumePlacementStrategy(&vpsSpec)
			dash.VerifyFatal(err, nil, "Check if able to apply volume placement strategy")
		})

		stepLog = "Apply storage class"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			params["repl"] = "3"
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

		checkVolumePlacement := func() {

			log.Infof("Waiting for 3 minute to get pvc status")
			time.Sleep(3 * time.Minute)
			//Waiting for PVC to Bound
			err = Inst().S.WaitForSinglePVCToBound(pvcName, namespace, 3)
			log.FailOnError(err, "Failed to wait for pvc to bound")
			// Get the PVC that was created using sc-1
			createdPVC, err := k8sCore.GetPersistentVolumeClaim(pvcName, namespace)
			log.FailOnError(err, "Failed to get PVC")
			log.Infof("Created PVC details: %v", createdPVC.Status)

			// Check PVC status - Ensure PVC is bound
			dash.VerifyFatal(createdPVC.Status.Phase == "Bound", true,
				fmt.Sprintf("PVC should be in 'Bound' status'%v'", createdPVC.Status.Phase))

			// Get the PV bound to the PVC
			pv, err := core.Instance().GetPersistentVolume(createdPVC.Spec.VolumeName)
			log.FailOnError(err, "Failed to get PersistentVolume")
			log.Infof("Persistent Volume details: %v", pv)

			// List all volumes and check replica placement
			volIDs, err := Inst().V.ListAllVolumes()
			log.FailOnError(err, "Failed to get volumes")
			log.Infof("Volume IDs list: %v", volIDs)

			for _, volId := range volIDs {
				apiVol, err := Inst().V.InspectVolume(volId)
				log.FailOnError(err, "Failed to inspect volume details")
				log.Infof("Inspecting volume ID: %s", volId)
				log.Infof("Found volume %s for createdPVC PVC", createdPVC.Name)

				if apiVol.Locator.VolumeLabels["pvc"] == createdPVC.Name {
					var nodeList []string
					for _, replica := range apiVol.ReplicaSets {
						nodeList = append(nodeList, replica.Nodes...)
					}
					// check if volume replicas are placed on nodes without the label 'zone=A'
					for _, nodeName := range nodeList {
						nodeID, err := node.GetNodeDetailsByNodeID(nodeName)
						log.FailOnError(err, "unable to find ID")
						nodeLabels, err := k8sCore.GetLabelsOnNode(nodeID.Name)
						log.FailOnError(err, "unable to find the node")
						labelValue, exists := nodeLabels["zone"]
						if !exists {
							log.Infof("Zone label not found : %s", exists)
						}
						dash.VerifyFatal(labelValue != "A", true,
							fmt.Sprintf("Node '%s' has label 'zone=A'. Found label: '%s'", nodeName, labelValue))
					}
				}
			}
		}

		stepLog = "Check PVC status and verify volume replica placement on nodes without label 'zone=A'"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			checkVolumePlacement()
		})

		deleteVolumes := func() {
			volIDs, err := Inst().V.ListAllVolumes()
			log.FailOnError(err, "Failed to get volumes")
			log.Infof("Volume IDs list: %v", volIDs)
			for _, each := range volIDs {
				if IsVolumeExits(each) {
					log.InfoD(fmt.Sprintf("delete volume [%v]", each))
					err := Inst().V.DetachVolume(each)
					if err != nil {
						log.Errorf("Failed to detach volume [%v]", each)
						return
					}
					time.Sleep(500 * time.Millisecond)
					err = Inst().V.DeleteVolume(each)
					if err != nil {
						log.Errorf("Delete volume with ID [%v] failed", each)
						return
					}
				}
			}
		}

		stepLog = fmt.Sprintf("Deleting volumes")
		Step(stepLog, func() {
			log.InfoD(stepLog)
			deleteVolumes()
		})

		var trashcanVols []string
		stepLog = "validate volumes in trashcan"
		Step(stepLog, func() {
			// wait for a few seconds for pvc to get deleted and volume to get detached
			time.Sleep(10 * time.Second)
			node := node.GetStorageDriverNodes()[0]
			log.InfoD(stepLog)
			trashcanVols, err = Inst().V.GetTrashCanVolumeIds(node)
			log.FailOnError(err, "error While getting trashcan volumes")
			log.Infof("trashcan len: %d", len(trashcanVols))
			dash.VerifyFatal(len(trashcanVols) > 0, true, "validate volumes exist in trashcan")
		})

		stepLog = "Validating trashcan restore"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for _, tID := range trashcanVols {
				if tID != "" {
					vol, err := Inst().V.InspectVolume(tID)
					log.FailOnError(err, fmt.Sprintf("error inspecting volume %s", tID))
					err = trashcanRestore(vol.Id, vol.Locator.Name)
					log.FailOnError(err, fmt.Sprintf("error restoring volume %s from trashcan", vol.Id))
				}
			}
		})

		stepLog = "Check PVC status and verify volume replica placement on nodes without label 'zone=A'"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			checkVolumePlacement()
		})

		stepLog = "Disable Trashcan"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			err := Inst().V.SetClusterOptsWithConfirmation(currNode, map[string]string{
				"--volume-expiration-minutes": "0",
			})
			log.FailOnError(err, "error while enabling trashcan")
			log.InfoD("Trashcan is successfully Disabled")
		})

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
				log.FailOnError(err, "Failed to remove label from node.")
			}
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts)
	})
})

// Verify Volume Anti-Affinity topology keys with volume labels (with few nodes not set with the topology key)
var _ = Describe("{ValidateAntiAffinityTopologyVPS}", Label("p0", "positive", "VPS", "topology", "staging"), func() {
	/*
		1.	Label some node with topology label (topology.kubernetes.io/zone=zone1)
		2.  Create anti-affinity vps rule
		3.	Apply the volume placement strategy, storage class, and pvc
		4.	Verify volume is provisioned according to vps on the non-labeled nodes.
		5.  Remove all labels and delete the pvc, vps
	*/
	var (
		contexts   = make([]*scheduler.Context, 0)
		testrailID = 0
		runID      int
	)
	JustBeforeEach(func() {
		StartTorpedoTest("ValidateAntiAffinityTopologyVPS", "Validate Anti-Affinity Topology Volume Placement Strategy to create vol on non-labeled nodes", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})

	stepLog := "Validate anti-affinity vps rule for topology key on the nodes"
	It(stepLog, func() {
		log.InfoD(stepLog)
		var (
			vpsName      = fmt.Sprintf("mongo-vps-%v", time.Now().Unix())
			scName       = fmt.Sprintf("mongo-sc-%v", time.Now().Unix())
			pvcName      = fmt.Sprintf("mongo-pvc-%v", time.Now().Unix())
			namespace    = "default"
			params       = make(map[string]string)
			k8sStorage   = storage.Instance()
			nodes        []node.Node
			vpsSpec      v1beta2.VolumePlacementStrategy
			labeledNodes []node.Node
		)
		stepLog = "Get all nodes and add topology label on some nodes"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			nodes = node.GetWorkerNodes()
			log.Infof("node list: %v", nodes)
			for index, node := range nodes {
				if index%2 == 1 {
					log.Infof("Adding TopologyZone for the node: %v", node.Name)
					err := k8sCore.AddLabelOnNode(node.Name, "topology.kubernetes.io/zone", "zone1")
					log.FailOnError(err, "Failed to add label 'topology.kubernetes.io/zone=zone1' for node: %v", node)
					labeledNodes = append(labeledNodes, node)
				}
			}
		})

		stepLog = "Apply volume placement strategy"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			matchExpression := []*v1beta1.LabelSelectorRequirement{
				{
					Key:      "topology.kubernetes.io/zone",
					Operator: v1beta1.LabelSelectorOpIn,
					Values:   []string{"zone1"},
				},
			}
			vpsSpec = vpsutil.VolumeAntiAffinityByMatchExpression(vpsName, matchExpression)
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
		log.Infof("Waiting for a minute to get pvc status")
		time.Sleep(10 * time.Second)

		createdPVC, err := k8sCore.GetPersistentVolumeClaim(pvcName, namespace)
		log.FailOnError(err, "Failed to get pvc")
		log.Infof("Created PVC details %v", createdPVC)

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

		stepLog = "Verify volumes are not created in the labeled node"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			//Waiting for PVC to Bound
			err = Inst().S.WaitForSinglePVCToBound(pvcName, namespace, 6)
			log.FailOnError(err, "Failed to wait for pvc to bound")
			// Get the PVC that was created using sc-1
			createdPVC, err := k8sCore.GetPersistentVolumeClaim(pvcName, namespace)
			log.FailOnError(err, "Failed to get PVC")
			log.Infof("Created PVC details: %v", createdPVC.Status)

			// Check PVC status - Ensure PVC is bound
			dash.VerifyFatal(createdPVC.Status.Phase == "Bound", true,
				fmt.Sprintf("PVC should be in 'Bound' status'%v'", createdPVC.Status.Phase))

			// Get the PV bound to the PVC
			pv, err := core.Instance().GetPersistentVolume(createdPVC.Spec.VolumeName)
			log.FailOnError(err, "Failed to get PersistentVolume")
			log.Infof("Persistent Volume details: %v", pv)

			// List all volumes and check replica placement
			volIDs, err := Inst().V.ListAllVolumes()
			log.FailOnError(err, "Failed to get volumes")
			log.Infof("Volume IDs list: %v", volIDs)

			for _, volId := range volIDs {
				apiVol, err := Inst().V.InspectVolume(volId)
				log.FailOnError(err, "Failed to inspect volume details")
				log.Infof("Inspecting volume ID: %s", volId)
				log.Infof("Found volume %s for createdPVC PVC", createdPVC.Name)

				if apiVol.Locator.VolumeLabels["pvc"] == createdPVC.Name {
					var nodeList []string
					for _, replica := range apiVol.ReplicaSets {
						nodeList = append(nodeList, replica.Nodes...)
					}
					// check if volume replicas are placed on nodes without the label 'zone=A'
					for _, nodeName := range nodeList {
						nodeID, err := node.GetNodeDetailsByNodeID(nodeName)
						log.FailOnError(err, "unable to find ID")
						nodeLabels, err := k8sCore.GetLabelsOnNode(nodeID.Name)
						log.FailOnError(err, "unable to find the node")
						labelValue, exists := nodeLabels["topology.kubernetes.io/zone"]
						if !exists {
							log.Infof("Zone label not found : %s", exists)
						}
						dash.VerifyFatal(labelValue != "zone1", true,
							fmt.Sprintf("Node '%s' has label 'topology.kubernetes.io/zone=zone1'. Found label: '%s'", nodeName, labelValue))
					}
				}
			}
		})

		stepLog = "Remove all newly created specs and labels on the nodes"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			log.Infof("Deleting the newly created VPS")
			err = talisman.Instance().DeleteVolumePlacementStrategy(vpsName)
			log.FailOnError(err, "Failed to remove VPS: %v", vpsName)
			log.Infof("Deleting the newly created labels from node")
			for _, node := range nodes {
				err = Inst().S.RemoveLabelOnNode(node, "topology.kubernetes.io/zone")
				log.FailOnError(err, "Failed to remove label from node.")
			}
		})

	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})
})
