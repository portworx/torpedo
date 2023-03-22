package tests

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	"github.com/portworx/sched-ops/k8s/talisman"
	"github.com/portworx/talisman/pkg/apis/portworx/v1beta1"
	"github.com/portworx/talisman/pkg/apis/portworx/v1beta2"
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/pkg/log"
	"github.com/portworx/torpedo/pkg/testrailuttils"
	"github.com/portworx/torpedo/pkg/vpsutil"
	. "github.com/portworx/torpedo/tests"
)

var _ = Describe("{VolumePlacementStrategyFunctional}", func() {
	var testrailID, runID int
	var contexts []*scheduler.Context
	var namespacePrefix string

	JustBeforeEach(func() {
		runID = testrailuttils.AddRunsToMilestone(testrailID)

		StartTorpedoTest("VolumePlacementStrategyFunctional", "Functional Tests for VPS", nil, testrailID)
	})

	Context("VolumePlacementStrategyValidation", func() {
		var vpsTestCase VolumePlaceMentStrategyTestCase

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
						contexts = append(contexts, ScheduleApplications(fmt.Sprintf("%s-%d", namespacePrefix, i))...)
					}
					log.InfoD("Validate Applications")
					ValidateApplications(contexts)
				})

				Step("Validate Deployment with VPS", func() {
					err := vpsTestCase.ValidateVPSDeployment(contexts)
					log.FailOnError(err, "Failed to Validate Deployments with respect to VPS")
				})

				Step("Destroy VPS Deployment", func() {
					err := vpsTestCase.DestroyVPSDeployment()
					log.FailOnError(err, "Failed to Destroy VPS Deployments")
				})

			})
		}

		// test mongo volume anti affinity
		Context("{VPSMongoAntiAffinity}", func() {
			BeforeEach(func() {
				namespacePrefix = "mongovpsantiaffinity"
				vpsTestCase = &mongoVPSAntiAffinity{}
			})
			testValidateVPS()
		})

		// test mongo volume anti affinity
		Context("{VPSMongoAffinity}", func() {
			BeforeEach(func() {
				namespacePrefix = "mongovpsaffinity"
				vpsTestCase = &mongoVPSAffinity{}
			})
			testValidateVPS()
		})
	})

	AfterEach(func() {
		Step("destroy apps", func() {
			log.InfoD("destroying apps")
			if CurrentGinkgoTestDescription().Failed {
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

type VolumePlaceMentStrategyTestCase interface {
	TestName() string
	DeployVPS() error
	DestroyVPSDeployment() error
	ValidateVPSDeployment(contexts []*scheduler.Context) error
}

type VolumePlacementStrategySpec struct {
	spec *v1beta2.VolumePlacementStrategy
}

type mongoVPSAntiAffinity struct {
	VolumePlacementStrategySpec
}

func (m *mongoVPSAntiAffinity) TestName() string {
	return "mongovpsantiaffinity"
}

func (m *mongoVPSAntiAffinity) DeployVPS() error {

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

func (m *mongoVPSAntiAffinity) DestroyVPSDeployment() error {
	return talisman.Instance().DeleteVolumePlacementStrategy(m.spec.Name)
}

// mongoVPSAntiAffinity is expecting to have deploy 2 replica of vol for each pod that has label [mongo-0, mongo-1]
// since this is antiaffinity, we are expecting that vol with the same labels are not deployed on the same pool.
// to validate that, we get the label from each deployed vol and extra the pool it's deployed on. if deployed correctly,
// there should be two pools per label.
func (m *mongoVPSAntiAffinity) ValidateVPSDeployment(contexts []*scheduler.Context) error {
	vols, err := Inst().S.GetVolumes(contexts[0])
	if err != nil {
		return err
	}

	resultMap := make(map[string][]string)
	for _, vol := range vols {

		vol, err := Inst().V.InspectVolume(vol.ID)
		if err != nil {
			return err
		}
		for key, value := range vol.Locator.VolumeLabels {
			if key == "px/statefulset-pod" {
				if Contains(resultMap[value], vol.ReplicaSets[0].PoolUuids[0]) {
					continue
				}
				resultMap[value] = append(resultMap[value], vol.ReplicaSets[0].PoolUuids[0])
				break
			}
		}
	}

	for _, value := range resultMap {
		if len(value) != 2 {
			return fmt.Errorf("failed to validate vps deployment, expecting label to exist in 2 pools, but got length %v and pools %v", len(value), value)
		}
	}
	return nil
}

type mongoVPSAffinity struct {
	VolumePlacementStrategySpec
}

func (m *mongoVPSAffinity) TestName() string {
	return "mongovpsaffinity"
}

func (m *mongoVPSAffinity) DeployVPS() error {

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

func (m *mongoVPSAffinity) DestroyVPSDeployment() error {
	return talisman.Instance().DeleteVolumePlacementStrategy(m.spec.Name)
}

// mongoVPSAffinity is expecting to have deploy 2 replica of vol for each pod that has label app=mongo-sts
// since this is affinity, we are expecting that vol with the same labels are not deployed on the same pool.
// to validate that, we get the label from each deployed vol and extra the pool it's deployed on. if deployed correctly,
// there should be one pools per label only.
func (m *mongoVPSAffinity) ValidateVPSDeployment(contexts []*scheduler.Context) error {
	vols, err := Inst().S.GetVolumes(contexts[0])
	if err != nil {
		return err
	}

	resultMap := make(map[string][]string)
	for _, vol := range vols {

		vol, err := Inst().V.InspectVolume(vol.ID)
		if err != nil {
			return err
		}
		for key, value := range vol.Locator.VolumeLabels {
			if key == "app" {
				if Contains(resultMap[value], vol.ReplicaSets[0].PoolUuids[0]) {
					continue
				}
				resultMap[value] = append(resultMap[value], vol.ReplicaSets[0].PoolUuids[0])
				break
			}
		}
	}

	for _, value := range resultMap {
		if len(value) != 1 {
			return fmt.Errorf("failed to validate vps deployment, expecting label to exist in 1 pools, but got length %v and pools %v", len(value), value)
		}
	}
	return nil
}
