package tests

import (
	"fmt"
	"github.com/libopenstorage/openstorage/api"
	. "github.com/onsi/ginkgo/v2"
	"github.com/portworx/sched-ops/k8s/core"
	"github.com/portworx/sched-ops/k8s/talisman"
	"github.com/portworx/sched-ops/task"
	"github.com/portworx/talisman/pkg/apis/portworx/v1beta1"
	"github.com/portworx/talisman/pkg/apis/portworx/v1beta2"
	"github.com/portworx/torpedo/drivers/node"
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/drivers/volume"
	"github.com/portworx/torpedo/pkg/log"
	"github.com/portworx/torpedo/pkg/testrailuttils"
	"github.com/portworx/torpedo/pkg/vpsutil"
	. "github.com/portworx/torpedo/tests"
	"math/rand"
	"time"
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

var _ = Describe("{VPSValueDoesntExist}", func() {

	JustBeforeEach(func() {
		StartTorpedoTest("VPSValueDoesntExist", "VPS test where vps rule where value is left empty", nil, 0)
	})

	itLog := "VPS test to check when vps value is left empty"
	It(itLog, func() {
		log.InfoD(itLog)

		var appList []string
		var contexts []*scheduler.Context

		stepLog := "Deploy vps with empty key"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			matchExpression := []*v1beta1.LabelSelectorRequirement{
				{
					Key:      "vpsValEmpty",
					Operator: v1beta1.LabelSelectorOpExists,
					Values:   []string{},
				},
			}
			vpsSpec := vpsutil.VolumeAffinityByMatchExpression("val-empty-vps", matchExpression)
			_, err := talisman.Instance().CreateVolumePlacementStrategy(&vpsSpec)
			log.FailOnError(err, "Failed to Deploy VPS Spec")

		})

		stepLog = "Label all the nodes with the vps key which was created in the previous step"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			storageNodes := node.GetStorageDriverNodes()
			for _, node := range storageNodes {
				err := Inst().S.AddLabelOnNode(node, "vpsValEmpty", "")
				log.FailOnError(err, "Failed to Label Node")
			}
		})

		stepLog = "Schedule applications with the label which was created in the previous step"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			appList = Inst().AppList
			defer func() {
				Inst().AppList = appList
			}()
			Inst().AppList = []string{"fio-vps-val-empty"}
			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				contexts = append(contexts, ScheduleApplications(fmt.Sprintf("fio-vps-val-empty-%d", i))...)
			}
		})

		stepLog = "PVC should not be scheduled on any node"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for _, ctx := range contexts {
				// get all pvc from the context
				pvcs, err := core.Instance().GetPersistentVolumeClaims(ctx.App.NameSpace, nil)
				log.FailOnError(err, "Failed to get PVCs")

				for _, pvc := range pvcs.Items {
					log.InfoD("PVC %s status is %s", pvc.Name, pvc.Status.Phase)
				}

				// get all pods from the context
				pods, err := core.Instance().GetPods(ctx.App.NameSpace, nil)
				log.FailOnError(err, "Failed to get Pods")

				for _, pod := range pods.Items {
					log.InfoD("Pod %s status is %s", pod.Name, pod.Status.Phase)
				}
			}
		})

		stepLog = "Remove the label from the nodes"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			storageNodes := node.GetStorageDriverNodes()
			for _, node := range storageNodes {
				err := Inst().S.RemoveLabelOnNode(node, "vpsValEmpty")
				log.FailOnError(err, "Failed to Remove Label from Node")
			}
		})

		stepLog = "After removing the label, the pvcs should be scheduled on the nodes"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			time.Sleep(1 * time.Minute)
			t := func() (interface{}, bool, error) {
				for _, ctx := range contexts {
					pvcs, err := core.Instance().GetPersistentVolumeClaims(ctx.App.NameSpace, nil)
					if err != nil {
						return nil, true, err
					}

					for _, pvc := range pvcs.Items {
						log.InfoD("PVC %s status is %s", pvc.Name, pvc.Status.Phase)
						if pvc.Status.Phase == "Pending" {
							return nil, false, nil
						}
					}
				}
				return nil, true, nil
			}
			_, err := task.DoRetryWithTimeout(t, 5*time.Minute, 10*time.Second)
			log.FailOnError(err, "Failed to get PVCs")

		})

	})

	JustAfterEach(func() {
		EndTorpedoTest()

	})

})
