package tests

import (
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	"github.com/portworx/torpedo/drivers/node"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	pdsResLib "github.com/portworx/torpedo/drivers/unifiedPlatform/resiliency"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/stworkflows/pds"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests"
	. "github.com/portworx/torpedo/tests/unifiedPlatform"
)

var _ = Describe("{KillAgentDuringDeployment}", func() {
	var (
		deployment *automationModels.PDSDeploymentResponse
		err        error
	)

	JustBeforeEach(func() {
		StartPDSTorpedoTest("KillAgentDuringDeployment", "Kill PDS Agent Pod when a DS Deployment is happening", nil, 0)
		WorkflowDataService.SkipValidatation[pds.ValidatePdsDeployment] = true
		WorkflowDataService.SkipValidatation[pds.ValidatePdsWorkloads] = true
	})

	It("Kill PDS Agent Pod when a DS Deployment is happening", func() {
		for _, ds := range NewPdsParams.DataServiceToTest {

			Step("Deploy DataService", func() {
				deployment, err = WorkflowDataService.DeployDataService(ds, ds.Image, ds.Version, PDS_DEFAULT_NAMESPACE)
				log.FailOnError(err, "Deployment failed")
				log.Debugf("Source Deployment Id: [%s]", *deployment.Create.Meta.Uid)

			})

			Step("Delete PDSPods while deployment", func() {
				log.InfoD("Delete PDSPods while deployment")
				// Global Resiliency TC marker
				pdsResLib.MarkResiliencyTC(true)
				// Type of failure that this TC needs to cover
				failuretype := pdsResLib.TypeOfFailure{
					Type: pdsResLib.KillAgentPodDuringDeployment,
					Method: func() error {
						return WorkflowDataService.DeletePDSPods([]string{"pds-deployments", "pds-target"}, PlatformNamespace)
					},
				}

				pdsResLib.DefineFailureType(failuretype)
				err = pdsResLib.InduceFailureAfterWaitingForCondition(&deployment.Create, PDS_DEFAULT_NAMESPACE, 1, ds)
				log.FailOnError(err, fmt.Sprintf("Error happened while executing Kill Agent Pod test for data service %v", *deployment.Create.Status.CustomResourceName))
			})

			Step("Validate Data Service to after px-agent reboot", func() {
				log.InfoD("Validate Data Service to after pds-agent reboot")
				err = WorkflowDataService.ValidatePdsDataServiceDeployments(*deployment.Create.Meta.Uid, ds, ds.Replicas, WorkflowDataService.PDSTemplates.ResourceTemplateId, WorkflowDataService.PDSTemplates.StorageTemplateId, PDS_DEFAULT_NAMESPACE, ds.Version, ds.Image)
				log.FailOnError(err, "Error while Validating dataservice after px-agent reboot")
			})

			stepLog := "Running Workloads after px-agent reboot"
			Step(stepLog, func() {
				_, err := WorkflowDataService.RunDataServiceWorkloads(*deployment.Create.Meta.Uid)
				log.FailOnError(err, "Error while running workloads on ds")
			})
		}
	})

	JustAfterEach(func() {
		defer EndPDSTorpedoTest()
	})
})

var _ = Describe("{RebootAllWorkerNodesDuringDeployment}", func() {
	var (
		deployment *automationModels.PDSDeploymentResponse
		err        error
	)

	JustBeforeEach(func() {
		StartPDSTorpedoTest("RebootAllWorkerNodesDuringDeployment", "Reboots all worker nodes while a data service pod is coming up", nil, 0)
		WorkflowDataService.SkipValidatation[pds.ValidatePdsDeployment] = true
		WorkflowDataService.SkipValidatation[pds.ValidatePdsWorkloads] = true
	})

	It("Reboots all worker nodes while a data service pod is coming up", func() {

		nodesToReboot := node.GetWorkerNodes()

		for _, ds := range NewPdsParams.DataServiceToTest {

			Step("Deploy DataService", func() {
				deployment, err = WorkflowDataService.DeployDataService(ds, ds.Image, ds.Version, PDS_DEFAULT_NAMESPACE)
				log.FailOnError(err, "Deployment failed")
				log.Debugf("Source Deployment Id: [%s]", *deployment.Create.Meta.Uid)

			})

			Step("Delete PDSPods while deployment", func() {
				log.InfoD("Delete PDSPods while deployment")
				// Global Resiliency TC marker
				pdsResLib.MarkResiliencyTC(true)
				// Type of failure that this TC needs to cover
				failuretype := pdsResLib.TypeOfFailure{
					Type: pdsResLib.RebootNodesDuringDeployment,
					Method: func() error {
						return RebootNodes(nodesToReboot)
					},
				}

				pdsResLib.DefineFailureType(failuretype)
				err = pdsResLib.InduceFailureAfterWaitingForCondition(&deployment.Create, PDS_DEFAULT_NAMESPACE, 1, ds)
				log.FailOnError(err, fmt.Sprintf("Error happened while executing Reboot Nodes during deployment test for data service %v", *deployment.Create.Status.CustomResourceName))
			})

			Step("Validate Data Service to after px-agent reboot", func() {
				log.InfoD("Validate Data Service to after pds-agent reboot")
				err = WorkflowDataService.ValidatePdsDataServiceDeployments(*deployment.Create.Meta.Uid, ds, ds.Replicas, WorkflowDataService.PDSTemplates.ResourceTemplateId, WorkflowDataService.PDSTemplates.StorageTemplateId, PDS_DEFAULT_NAMESPACE, ds.Version, ds.Image)
				log.FailOnError(err, "Error while Validating dataservice after px-agent reboot")
			})

			stepLog := "Running Workloads after px-agent reboot"
			Step(stepLog, func() {
				_, err := WorkflowDataService.RunDataServiceWorkloads(*deployment.Create.Meta.Uid)
				log.FailOnError(err, "Error while running workloads on ds")
			})
		}
	})

	JustAfterEach(func() {
		defer EndPDSTorpedoTest()
	})
})

var _ = Describe("{KillPdsAgentPodDuringAppScaleUp}", func() {
	var (
		deploymentAfterUpdate automationModels.V1Deployment
		deployment            *automationModels.PDSDeploymentResponse
		err                   error
	)

	JustBeforeEach(func() {
		StartPDSTorpedoTest("KillPdsAgentPodDuringAppScaleUp", "Kill PDS-Agent Pod during application is scaled up", nil, 0)
	})

	It("Kill PDS-Agent Pod during application is scaled up", func() {
		for _, ds := range NewPdsParams.DataServiceToTest {
			Step("Deploy DataService", func() {
				deployment, err = WorkflowDataService.DeployDataService(ds, ds.Image, ds.Version, PDS_DEFAULT_NAMESPACE)
				log.FailOnError(err, "Error while deploying ds")
				log.Debugf("Source Deployment Id: [%s]", *deployment.Create.Meta.Uid)
				WorkflowDataService.SkipValidatation[pds.ValidatePdsDeployment] = true
				WorkflowDataService.SkipValidatation[pds.ValidatePdsWorkloads] = true
			})

			Step("ScaleUp DataService", func() {
				log.InfoD("Scaling Up dataServices...")
				updateDeployment, err := WorkflowDataService.UpdateDataService(ds, *deployment.Create.Meta.Uid, ds.Image, ds.Version)
				log.FailOnError(err, "Error while updating ds")
				log.Debugf("Updated Deployment Id: [%s]", *updateDeployment.Update.Meta.Uid)
				deploymentAfterUpdate, err = WorkflowDataService.GetDeployment(*deployment.Create.Meta.Uid)
				log.FailOnError(err, "Error while fetching the deployment")
			})

			Step("Delete PDSPods while scaling up the data service", func() {
				log.InfoD("Delete PDSPods while deployment")
				// Global Resiliency TC marker
				pdsResLib.MarkResiliencyTC(true)
				// Type of failure that this TC needs to cover
				failuretype := pdsResLib.TypeOfFailure{
					Type: pdsResLib.KillPdsAgentPodDuringAppScaleUp,
					Method: func() error {
						return WorkflowDataService.DeletePDSPods([]string{"pds-deployments", "pds-target"}, PlatformNamespace)
					},
				}

				pdsResLib.DefineFailureType(failuretype)
				err = pdsResLib.InduceFailureAfterWaitingForCondition(&deploymentAfterUpdate, PDS_DEFAULT_NAMESPACE, int32(ds.ScaleReplicas), ds)
				log.FailOnError(err, fmt.Sprintf("Error happened while executing Reboot Nodes during deployment test for data service %v", *deployment.Create.Status.CustomResourceName))
			})

			Step("Validate Data Service to after Scale Up", func() {
				log.InfoD("Validate Data Service to after Scale Up")
				err = WorkflowDataService.ValidatePdsDataServiceDeployments(*deployment.Create.Meta.Uid, ds, ds.ScaleReplicas, WorkflowDataService.PDSTemplates.ResourceTemplateId, WorkflowDataService.PDSTemplates.StorageTemplateId, PDS_DEFAULT_NAMESPACE, ds.Version, ds.Image)
				log.FailOnError(err, "Error while Validating dataservice after scale up")
			})

			stepLog := "Running Workloads after ScaleUp of DataService"
			Step(stepLog, func() {
				_, err := WorkflowDataService.RunDataServiceWorkloads(*deployment.Create.Meta.Uid)
				log.FailOnError(err, "Error while running workloads on ds")
			})
		}
	})

	JustAfterEach(func() {
		defer EndPDSTorpedoTest()
	})
})

var _ = Describe("{StopPXDuringStorageResize}", func() {
	var (
		deployment *automationModels.PDSDeploymentResponse
		err        error
	)

	JustBeforeEach(func() {
		StartPDSTorpedoTest("StopPXDuringStorageResize", "Stop PX on a node during application's storage is resized", nil, 0)
	})

	It("Kill PDS Agent Pod when a DS Deployment is happening", func() {
		var volNodesWithPx []node.Node
		for _, ds := range NewPdsParams.DataServiceToTest {

			Step("Deploy DataService", func() {
				deployment, err = WorkflowDataService.DeployDataService(ds, ds.Image, ds.Version, PDS_DEFAULT_NAMESPACE)
				log.FailOnError(err, "Deployment failed")
				log.Debugf("Source Deployment Id: [%s]", *deployment.Create.Meta.Uid)

				//Update Ds With New Values of Resource Templates
				resConfigIdUpdated, err := WorkflowPDSTemplate.CreateResourceTemplateWithCustomValue(NewPdsParams)
				log.FailOnError(err, "Unable to create Custom Templates for PDS")
				log.InfoD("Updated Resource Template ID- [updated- %v]", resConfigIdUpdated)
				log.Infof("Associate newly created template to the project")
				err = WorkflowProject.Associate(
					[]string{},
					[]string{},
					[]string{},
					[]string{},
					[]string{resConfigIdUpdated},
					[]string{},
				)
				log.FailOnError(err, "Unable to associate Templates to Project")
				log.Infof("Associated Resources - [%+v]", WorkflowProject.AssociatedResources)
				pdsResLib.UpdateTemplate = resConfigIdUpdated
			})

			Step("Fetch Volume Nodes on which PX is Running", func() {
				volNodesWithPx = GetVolumeNodesOnWhichPxIsRunning()
				log.InfoD("volume nodes list calculated is- %v", volNodesWithPx)
			})

			Step("Stop Px on Ds Node and replica node while storage size increase", func() {
				log.InfoD("Stop Px on Ds Node and replica node while storage size increase")
				// Global Resiliency TC marker
				pdsResLib.MarkResiliencyTC(true)
				// Type of failure that this TC needs to cover
				failuretype := pdsResLib.TypeOfFailure{
					Type: pdsResLib.StopPXDuringStorageResize,
					Method: func() error {
						return StopPxOnReplicaVolumeNode(volNodesWithPx)
					},
				}

				pdsResLib.DefineFailureType(failuretype)
				pdsResLib.AccountID = AccID
				err = pdsResLib.InduceFailureAfterWaitingForCondition(&deployment.Create, PDS_DEFAULT_NAMESPACE, 1, ds)
				log.FailOnError(err, fmt.Sprintf("Error happened while executing Kill Agent Pod test for data service %v", *deployment.Create.Status.CustomResourceName))
			})

			Step("Start PX on the same node after volume resize", func() {
				StartPxOnReplicaVolumeNode(volNodesWithPx)
			})

			Step("Validate Data Service to after px restart", func() {
				log.InfoD("Validate Data Service to after px restart")
				err = WorkflowDataService.ValidatePdsDataServiceDeployments(*deployment.Create.Meta.Uid, ds, ds.ScaleReplicas, WorkflowDataService.PDSTemplates.ResourceTemplateId, WorkflowDataService.PDSTemplates.StorageTemplateId, PDS_DEFAULT_NAMESPACE, ds.Version, ds.Image)
				log.FailOnError(err, "Error while Validating dataservice after px-agent reboot")
			})

			stepLog := "Running Workloads after px-agent reboot"
			Step(stepLog, func() {
				_, err := WorkflowDataService.RunDataServiceWorkloads(*deployment.Create.Meta.Uid)
				log.FailOnError(err, "Error while running workloads on ds")
			})
		}
	})

	JustAfterEach(func() {
		defer EndPDSTorpedoTest()
	})
})

var _ = Describe("{RestartPXDuringAppScaleUp}", func() {
	var (
		deployment *automationModels.PDSDeploymentResponse
		err        error
	)

	JustBeforeEach(func() {
		StartPDSTorpedoTest("RestartPXDuringAppScaleUp", "Restart PX on a node during application is scaled up", nil, 0)
	})

	It("Restart PX on a node during application is scaled up", func() {
		for _, ds := range NewPdsParams.DataServiceToTest {
			Step("Deploy DataService", func() {
				deployment, err = WorkflowDataService.DeployDataService(ds, ds.Image, ds.Version, PDS_DEFAULT_NAMESPACE)
				log.FailOnError(err, "Error while deploying ds")
				log.Debugf("Source Deployment Id: [%s]", *deployment.Create.Meta.Uid)
			})

			Step("Create and associate update template to the projectß", func() {
				resConfigIdUpdated, err := WorkflowPDSTemplate.CreateResourceTemplateWithCustomValue(NewPdsParams)
				log.FailOnError(err, "Unable to create Custom Templates for PDS")
				log.InfoD("Updated Resource Template ID- [updated- %v]", resConfigIdUpdated)
				log.Infof("Associate newly created template to the project")
				err = WorkflowProject.Associate(
					[]string{},
					[]string{},
					[]string{},
					[]string{},
					[]string{resConfigIdUpdated},
					[]string{},
				)
				log.FailOnError(err, "Unable to associate Templates to Project")
			})

			Step("Restart PX while data service is scaling up", func() {
				log.InfoD("Restart PX while data service is scaling up")
				// Global Resiliency TC marker
				pdsResLib.UpdateTemplate = WorkflowPDSTemplate.UpdateResourceTemplateId
				pdsResLib.MarkResiliencyTC(true)
				log.Infof("Update Id: [%s]", WorkflowDataService.PDSTemplates.UpdateResourceTemplateId)
				// Type of failure that this TC needs to cover
				failuretype := pdsResLib.TypeOfFailure{
					Type: pdsResLib.RestartPxDuringDSScaleUp,
					Method: func() error {
						return pdsResLib.RestartPXDuringDSScaleUp(PDS_DEFAULT_NAMESPACE, *deployment.Create.Status.CustomResourceName)
					},
				}

				pdsResLib.DefineFailureType(failuretype)
				err = pdsResLib.InduceFailureAfterWaitingForCondition(&deployment.Create, PDS_DEFAULT_NAMESPACE, int32(ds.Replicas), ds)
				log.FailOnError(err, fmt.Sprintf("Error happened while executing restarting PX %v", *deployment.Create.Status.CustomResourceName))
			})

			Step("Validate Data Service after Scale Up", func() {
				log.InfoD("Validate Data Service to after Scale Up")
				err = WorkflowDataService.ValidatePdsDataServiceDeployments(*deployment.Create.Meta.Uid, ds, ds.ScaleReplicas, WorkflowDataService.PDSTemplates.ResourceTemplateId, WorkflowDataService.PDSTemplates.StorageTemplateId, PDS_DEFAULT_NAMESPACE, ds.Version, ds.Image)
				log.FailOnError(err, "Error while Validating dataservice after scale up")
			})

			stepLog := "Running Workloads after ScaleUp of DataService"
			Step(stepLog, func() {
				_, err := WorkflowDataService.RunDataServiceWorkloads(*deployment.Create.Meta.Uid)
				log.FailOnError(err, "Error while running workloads on ds")
			})
		}
	})

	JustAfterEach(func() {
		defer EndPDSTorpedoTest()
	})
})

var _ = Describe("{RebootNodeDuringAppResourceUpdate}", func() {
	var (
		deployment *automationModels.PDSDeploymentResponse
		err        error
	)

	JustBeforeEach(func() {
		StartPDSTorpedoTest("RebootNodeDuringAppResourceUpdate", "Reboot active node during application resource Update, example increase the CPU/Mem limits ", nil, 0)
	})

	It("Reboot active node during application resource Update, example increase the CPU/Mem limits ", func() {
		for _, ds := range NewPdsParams.DataServiceToTest {

			Step("Deploy DataService", func() {
				deployment, err = WorkflowDataService.DeployDataService(ds, ds.Image, ds.Version, PDS_DEFAULT_NAMESPACE)
				log.FailOnError(err, "Deployment failed")
				log.Debugf("Source Deployment Id: [%s]", *deployment.Create.Meta.Uid)

				//Update Ds With New Values of Resource Templates
				resConfigIdUpdated, err := WorkflowPDSTemplate.CreateResourceTemplateWithCustomValue(NewPdsParams)
				log.FailOnError(err, "Unable to create Custom Templates for PDS")
				log.InfoD("Updated Resource Template ID- [updated- %v]", resConfigIdUpdated)
				log.Infof("Associate newly created template to the project")
				err = WorkflowProject.Associate(
					[]string{},
					[]string{},
					[]string{},
					[]string{},
					[]string{resConfigIdUpdated},
					[]string{},
				)
				log.FailOnError(err, "Unable to associate Templates to Project")
				log.Infof("Associated Resources - [%+v]", WorkflowProject.AssociatedResources)
				pdsResLib.UpdateTemplate = resConfigIdUpdated
			})

			Step("Reboot nodes while app resource are updated", func() {
				log.InfoD("Reboot nodes while app resource are updated")
				// Global Resiliency TC marker
				pdsResLib.MarkResiliencyTC(true)
				// Type of failure that this TC needs to cover
				failuretype := pdsResLib.TypeOfFailure{
					Type: pdsResLib.RebootNodeDuringAppResourceUpdate,
					Method: func() error {
						return pdsResLib.RebootActiveNodeDuringDeployment(PDS_DEFAULT_NAMESPACE, *deployment.Create.Status.CustomResourceName, 1)
					},
				}

				pdsResLib.DefineFailureType(failuretype)
				pdsResLib.AccountID = AccID
				err = pdsResLib.InduceFailureAfterWaitingForCondition(&deployment.Create, PDS_DEFAULT_NAMESPACE, 1, ds)
				log.FailOnError(err, fmt.Sprintf("Error happened while executing node reboot %v", *deployment.Create.Status.CustomResourceName))
			})

			Step("Validate Data Service to after node reboot", func() {
				log.InfoD("Validate Data Service to after node reboot")
				err = WorkflowDataService.ValidatePdsDataServiceDeployments(*deployment.Create.Meta.Uid, ds, ds.ScaleReplicas, WorkflowDataService.PDSTemplates.ResourceTemplateId, WorkflowDataService.PDSTemplates.StorageTemplateId, PDS_DEFAULT_NAMESPACE, ds.Version, ds.Image)
				log.FailOnError(err, "Error while Validating dataservice after px-agent reboot")
			})

			stepLog := "Running Workloads after node reboot"
			Step(stepLog, func() {
				_, err := WorkflowDataService.RunDataServiceWorkloads(*deployment.Create.Meta.Uid)
				log.FailOnError(err, "Error while running workloads on ds")
			})
		}
	})

	JustAfterEach(func() {
		defer EndPDSTorpedoTest()
	})
})

var _ = Describe("{RestartAppDuringResourceUpdate}", func() {
	var (
		deployment *automationModels.PDSDeploymentResponse
		err        error
	)

	JustBeforeEach(func() {
		StartPDSTorpedoTest("RestartAppDuringResourceUpdate", "Restart application pod during resource update", nil, 0)
	})

	It("Restart application pod during resource update", func() {
		for _, ds := range NewPdsParams.DataServiceToTest {

			Step("Deploy DataService", func() {
				deployment, err = WorkflowDataService.DeployDataService(ds, ds.Image, ds.Version, PDS_DEFAULT_NAMESPACE)
				log.FailOnError(err, "Deployment failed")
				log.Debugf("Source Deployment Id: [%s]", *deployment.Create.Meta.Uid)

				//Update Ds With New Values of Resource Templates
				resConfigIdUpdated, err := WorkflowPDSTemplate.CreateResourceTemplateWithCustomValue(NewPdsParams)
				log.FailOnError(err, "Unable to create Custom Templates for PDS")
				log.InfoD("Updated Resource Template ID- [updated- %v]", resConfigIdUpdated)
				log.Infof("Associate newly created template to the project")
				err = WorkflowProject.Associate(
					[]string{},
					[]string{},
					[]string{},
					[]string{},
					[]string{resConfigIdUpdated},
					[]string{},
				)
				log.FailOnError(err, "Unable to associate Templates to Project")
				log.Infof("Associated Resources - [%+v]", WorkflowProject.AssociatedResources)
				pdsResLib.UpdateTemplate = resConfigIdUpdated
			})

			Step("Restart applications during resource update", func() {
				log.InfoD("Restart applications during resource update")
				// Global Resiliency TC marker
				pdsResLib.MarkResiliencyTC(true)
				// Type of failure that this TC needs to cover
				failuretype := pdsResLib.TypeOfFailure{
					Type: pdsResLib.RestartAppDuringResourceUpdate,
					Method: func() error {
						return pdsResLib.RestartApplicationDuringResourceUpdate(PDS_DEFAULT_NAMESPACE, *deployment.Create.Status.CustomResourceName)
					},
				}

				pdsResLib.DefineFailureType(failuretype)
				pdsResLib.AccountID = AccID
				err = pdsResLib.InduceFailureAfterWaitingForCondition(&deployment.Create, PDS_DEFAULT_NAMESPACE, 1, ds)
				log.FailOnError(err, fmt.Sprintf("Error happened while executing node reboot %v", *deployment.Create.Status.CustomResourceName))
			})

			Step("Validate Data Service to after node reboot", func() {
				log.InfoD("Validate Data Service to after node reboot")
				err = WorkflowDataService.ValidatePdsDataServiceDeployments(*deployment.Create.Meta.Uid, ds, ds.ScaleReplicas, WorkflowDataService.PDSTemplates.ResourceTemplateId, WorkflowDataService.PDSTemplates.StorageTemplateId, PDS_DEFAULT_NAMESPACE, ds.Version, ds.Image)
				log.FailOnError(err, "Error while Validating dataservice after px-agent reboot")
			})

			stepLog := "Running Workloads after node reboot"
			Step(stepLog, func() {
				_, err := WorkflowDataService.RunDataServiceWorkloads(*deployment.Create.Meta.Uid)
				log.FailOnError(err, "Error while running workloads on ds")
			})
		}
	})

	JustAfterEach(func() {
		defer EndPDSTorpedoTest()
	})
})

var _ = Describe("{RebootActiveNodeMultipleTimesDuringDeployment}", func() {
	var (
		deployment *automationModels.PDSDeploymentResponse
		err        error
		num_reboot int
	)

	JustBeforeEach(func() {
		StartPDSTorpedoTest("RebootActiveNodeMultipleTimesDuringDeployment", "Reboots a Node multiple times onto which a pod is coming up", nil, 0)
		num_reboot = 3
		WorkflowDataService.SkipValidatation[pds.ValidatePdsDeployment] = true
		WorkflowDataService.SkipValidatation[pds.ValidatePdsWorkloads] = true
	})

	It("Reboots a Node multiple times onto which a pod is coming up", func() {
		for _, ds := range NewPdsParams.DataServiceToTest {

			Step("Deploy DataService", func() {
				deployment, err = WorkflowDataService.DeployDataService(ds, ds.Image, ds.Version, PDS_DEFAULT_NAMESPACE)
				log.FailOnError(err, "Deployment failed")
				log.Debugf("Source Deployment Id: [%s]", *deployment.Create.Meta.Uid)

			})

			Step("Reboot nodes while app resource are updated", func() {
				log.InfoD("Reboot nodes while app resource are updated")
				// Global Resiliency TC marker
				pdsResLib.MarkResiliencyTC(true)
				// Type of failure that this TC needs to cover
				failuretype := pdsResLib.TypeOfFailure{
					Type: pdsResLib.ActiveNodeRebootDuringDeployment,
					Method: func() error {
						return pdsResLib.RebootActiveNodeDuringDeployment(PDS_DEFAULT_NAMESPACE, *deployment.Create.Status.CustomResourceName, num_reboot)
					},
				}

				pdsResLib.DefineFailureType(failuretype)
				pdsResLib.AccountID = AccID
				err = pdsResLib.InduceFailureAfterWaitingForCondition(&deployment.Create, PDS_DEFAULT_NAMESPACE, 1, ds)
				log.FailOnError(err, fmt.Sprintf("Error happened while rebooting nodes %v", *deployment.Create.Status.CustomResourceName))
			})

			Step("Validate Data Service to after node reboot", func() {
				log.InfoD("Validate Data Service to after node reboot")
				err = WorkflowDataService.ValidatePdsDataServiceDeployments(*deployment.Create.Meta.Uid, ds, ds.Replicas, WorkflowDataService.PDSTemplates.ResourceTemplateId, WorkflowDataService.PDSTemplates.StorageTemplateId, PDS_DEFAULT_NAMESPACE, ds.Version, ds.Image)
				log.FailOnError(err, "Error while Validating dataservice after px-agent reboot")
			})

			stepLog := "Running Workloads after node reboot"
			Step(stepLog, func() {
				_, err := WorkflowDataService.RunDataServiceWorkloads(*deployment.Create.Meta.Uid)
				log.FailOnError(err, "Error while running workloads on ds")
			})
		}
	})

	JustAfterEach(func() {
		defer EndPDSTorpedoTest()
	})
})
