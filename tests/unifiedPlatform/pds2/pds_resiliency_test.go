package tests

import (
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	"github.com/pure-px/torpedo/drivers/node"
	"github.com/pure-px/torpedo/drivers/unifiedPlatform/automationModels"
	dslibs "github.com/pure-px/torpedo/drivers/unifiedPlatform/pdsLibs"
	pdsResLib "github.com/pure-px/torpedo/drivers/unifiedPlatform/resiliency"
	"github.com/pure-px/torpedo/drivers/unifiedPlatform/stworkflows/pds"
	"github.com/pure-px/torpedo/pkg/log"
	. "github.com/pure-px/torpedo/tests"
	. "github.com/pure-px/torpedo/tests/unifiedPlatform"
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
		WorkflowDataService.SkipValidatation[pds.ValidatePdsWorkloads] = true
	})

	It("Kill PDS-Agent Pod during application is scaled up", func() {
		for _, ds := range NewPdsParams.DataServiceToTest {
			Step("Deploy DataService", func() {
				deployment, err = WorkflowDataService.DeployDataService(ds, ds.Image, ds.Version, PDS_DEFAULT_NAMESPACE)
				log.FailOnError(err, "Error while deploying ds")
				log.Debugf("Source Deployment Id: [%s]", *deployment.Create.Meta.Uid)
				WorkflowDataService.SkipValidatation[pds.ValidatePdsDeployment] = true
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
		delete(WorkflowDataService.SkipValidatation, pds.ValidatePdsWorkloads)
		delete(WorkflowDataService.SkipValidatation, pds.ValidatePdsDeployment)
	})
})

var _ = Describe("{KillPxAgentDuringDeployment}", func() {
	var (
		deployment *automationModels.PDSDeploymentResponse
		err        error
	)

	JustBeforeEach(func() {
		StartPDSTorpedoTest("KillPxAgentDuringDeployment", "Kill Px Agent Pod when a DS Deployment is happening", nil, 0)
		WorkflowDataService.SkipValidatation[pds.ValidatePdsDeployment] = true
		WorkflowDataService.SkipValidatation[pds.ValidatePdsWorkloads] = true
	})

	It("Kill Px-Agent Pod during application is scaled up", func() {
		for _, ds := range NewPdsParams.DataServiceToTest {
			Step("Deploy DataService", func() {
				deployment, err = WorkflowDataService.DeployDataService(ds, ds.Image, ds.Version, PDS_DEFAULT_NAMESPACE)
				log.FailOnError(err, "Error while deploying ds")
				log.Debugf("Source Deployment Id: [%s]", *deployment.Create.Meta.Uid)
			})

			Step("Delete Px System Pods while scaling up the data service", func() {
				log.InfoD("Delete Px System Pods while deployment")
				// Global Resiliency TC marker
				pdsResLib.MarkResiliencyTC(true)
				// Type of failure that this TC needs to cover
				failuretype := pdsResLib.TypeOfFailure{
					Type: pdsResLib.KillPxAgentDuringDeployment,
					Method: func() error {
						return WorkflowDataService.DeletePDSPods([]string{"px-agent", "px-app-operator", "px-tc-operator"}, PlatformNamespace)
					},
				}

				pdsResLib.DefineFailureType(failuretype)
				err = pdsResLib.InduceFailureAfterWaitingForCondition(&deployment.Create, PDS_DEFAULT_NAMESPACE, int32(ds.Replicas), ds)
				log.FailOnError(err, fmt.Sprintf("Error happened while killing px pods during deployment test for data service %v", *deployment.Create.Status.CustomResourceName))
			})

			Step("Validate Data Service to after deployment", func() {
				log.InfoD("Validate Data Service after deployment")
				err = WorkflowDataService.ValidatePdsDataServiceDeployments(*deployment.Create.Meta.Uid, ds, ds.Replicas, WorkflowDataService.PDSTemplates.ResourceTemplateId, WorkflowDataService.PDSTemplates.StorageTemplateId, PDS_DEFAULT_NAMESPACE, ds.Version, ds.Image)
				log.FailOnError(err, "Error while Validating dataservice after deployment")
			})

			stepLog := "Running Workloads after deployment of DataService"
			Step(stepLog, func() {
				_, err := WorkflowDataService.RunDataServiceWorkloads(*deployment.Create.Meta.Uid)
				log.FailOnError(err, "Error while running workloads on ds")
			})
		}
	})

	JustAfterEach(func() {
		defer EndPDSTorpedoTest()
		delete(WorkflowDataService.SkipValidatation, pds.ValidatePdsWorkloads)
		delete(WorkflowDataService.SkipValidatation, pds.ValidatePdsDeployment)
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
				err = WorkflowDataService.ValidatePdsDataServiceDeployments(*deployment.Create.Meta.Uid, ds, ds.ScaleReplicas, pdsResLib.UpdateTemplate, WorkflowDataService.PDSTemplates.StorageTemplateId, PDS_DEFAULT_NAMESPACE, ds.Version, ds.Image)
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
				pdsResLib.UpdateTemplate = resConfigIdUpdated
			})

			Step("Restart PX while data service is scaling up", func() {
				log.InfoD("Restart PX while data service is scaling up")
				// Global Resiliency TC marker
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
				err = WorkflowDataService.ValidatePdsDataServiceDeployments(*deployment.Create.Meta.Uid, ds, ds.ScaleReplicas, pdsResLib.UpdateTemplate, WorkflowDataService.PDSTemplates.StorageTemplateId, PDS_DEFAULT_NAMESPACE, ds.Version, ds.Image)
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
		deployment         *automationModels.PDSDeploymentResponse
		err                error
		resConfigIdUpdated string
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
				resConfigIdUpdated, err = WorkflowPDSTemplate.CreateResourceTemplateWithCustomValue(NewPdsParams)
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
				err = WorkflowDataService.ValidatePdsDataServiceDeployments(*deployment.Create.Meta.Uid, ds, ds.ScaleReplicas, pdsResLib.UpdateTemplate, WorkflowDataService.PDSTemplates.StorageTemplateId, PDS_DEFAULT_NAMESPACE, ds.Version, ds.Image)
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
				err = WorkflowDataService.ValidatePdsDataServiceDeployments(*deployment.Create.Meta.Uid, ds, ds.ScaleReplicas, pdsResLib.UpdateTemplate, WorkflowDataService.PDSTemplates.StorageTemplateId, PDS_DEFAULT_NAMESPACE, ds.Version, ds.Image)
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

var _ = Describe("{RestoreDuringNodesAreRebooted}", func() {
	var (
		deployment          *automationModels.PDSDeploymentResponse
		restoreResponse     *automationModels.PDSRestoreResponse
		latestBackupUid     string
		pdsBackupConfigName string
		restoreNamespace    string
		restoreName         string
		err                 error
	)

	JustBeforeEach(func() {
		StartPDSTorpedoTest("RestoreDuringNodesAreRebooted", "Restore DataService during nodes are rebooted", nil, 0)
		WorkflowPDSRestore.SkipValidation[pds.ValidatePdsRestore] = true
		WorkflowPDSRestore.SkipValidation[pds.CheckDataAfterRestore] = true
	})

	It("Restore DataService during nodes are rebooted", func() {
		for _, ds := range NewPdsParams.DataServiceToTest {

			Step("Deploy dataservice", func() {
				deployment, err = WorkflowDataService.DeployDataService(ds, ds.Image, ds.Version, PDS_DEFAULT_NAMESPACE)
				log.FailOnError(err, "Error while deploying ds")
				log.Infof("All deployments - [%+v]", WorkflowDataService.DataServiceDeployment)

			})

			Step("Create Adhoc backup config of the existing deployment", func() {
				pdsBackupConfigName = "pds-adhoc-backup-" + RandomString(5)
				bkpConfigResponse, err := WorkflowPDSBackupConfig.CreateBackupConfig(pdsBackupConfigName, *deployment.Create.Meta.Uid)
				log.FailOnError(err, "Error occured while creating backupConfig")
				log.Infof("BackupConfigName: [%s], BackupConfigId: [%s]", *bkpConfigResponse.Create.Meta.Name, *bkpConfigResponse.Create.Meta.Uid)
				log.Infof("All deployments - [%+v]", WorkflowDataService.DataServiceDeployment)
			})

			Step("Get the latest backup detail for the deployment", func() {
				backupResponse, err := WorkflowPDSBackup.GetLatestBackup(*deployment.Create.Meta.Uid)
				log.FailOnError(err, "Error occured while creating backup")
				latestBackupUid = *backupResponse.Meta.Uid
				log.Infof("Latest backup ID [%s], Name [%s]", *backupResponse.Meta.Uid, *backupResponse.Meta.Name)
				err = WorkflowPDSBackup.WaitForBackupToComplete(*backupResponse.Meta.Uid)
				log.FailOnError(err, "Error occured while waiting for backup to complete")
			})

			Step("Create Restore from the latest backup Id", func() {
				defer func() {
					err := SetSourceKubeConfig()
					log.FailOnError(err, "failed to switch context to source cluster")
				}()
				CheckforClusterSwitch()
				restoreNamespace = "restore-" + RandomString(5)
				restoreName = "restore-" + RandomString(5)
				restoreResponse, err = WorkflowPDSRestore.CreateRestore(restoreName, latestBackupUid, restoreNamespace, *deployment.Create.Meta.Uid)
				log.FailOnError(err, "Restore Failed")
				log.Infof("All restores - [%+v]", WorkflowPDSRestore.Restores)
				log.Infof("Restore Created Name - [%s], UID - [%s]", *WorkflowPDSRestore.Restores[restoreName].Meta.Name, *WorkflowPDSRestore.Restores[restoreName].Meta.Uid)
			})

			Step("Reboot nodes while restore", func() {
				log.InfoD("Reboot nodes while restore")
				// Global Resiliency TC marker
				pdsResLib.MarkResiliencyTC(true)
				// Type of failure that this TC needs to cover
				failuretype := pdsResLib.TypeOfFailure{
					Type: pdsResLib.RestoreDuringNodesAreRebooted,
					Method: func() error {
						return pdsResLib.RebootActiveNodeDuringDeployment(
							restoreNamespace,
							*WorkflowPDSRestore.RestoredDeployments.DataServiceDeployment[restoreResponse.Create.Config.DestinationReferences.DataServiceDeploymentId].Deployment.Status.CustomResourceName,
							1)
					},
				}

				pdsResLib.DefineFailureType(failuretype)
				pdsResLib.AccountID = AccID
				err = pdsResLib.InduceFailureAfterWaitingForCondition(
					&WorkflowPDSRestore.RestoredDeployments.DataServiceDeployment[restoreResponse.Create.Config.DestinationReferences.DataServiceDeploymentId].Deployment,
					restoreNamespace,
					1,
					ds)
				log.FailOnError(err, fmt.Sprintf("Error happened while rebooting nodes %v", *deployment.Create.Status.CustomResourceName))
			})

			Step("Validate Data Service after restore", func() {
				log.InfoD("Validate Data Service after restore")
				err = dslibs.ValidateRestoreDeployment(*restoreResponse.Create.Meta.Uid, restoreNamespace)
				log.FailOnError(err, "Error while Validating dataservice after px-agent reboot")
			})

			stepLog := "Checking data after restore"
			Step(stepLog, func() {
				err := WorkflowPDSRestore.ValidateDataAfterRestore(restoreResponse.Create.Config.DestinationReferences.DataServiceDeploymentId, latestBackupUid, restoreNamespace)
				log.FailOnError(err, "Error while validating data after restore")
			})

		}

	})

	JustAfterEach(func() {
		defer EndPDSTorpedoTest()
		delete(WorkflowDataService.SkipValidatation, pds.ValidatePdsWorkloads)
		delete(WorkflowDataService.SkipValidatation, pds.ValidatePdsDeployment)
	})
})

var _ = Describe("{KillTeleportAndDeploymentOperatorDuringWorkloadRun}", func() {
	var (
		deployment *automationModels.PDSDeploymentResponse
		err        error
	)

	JustBeforeEach(func() {
		StartPDSTorpedoTest("KillTeleportAndDeploymentOperatorDuringWorkloadRun", "Kill Teleport Agent Pod When Workload is running", nil, 0)
		WorkflowPDSRestore.SkipValidation[pds.ValidatePdsRestore] = true
		WorkflowDataService.SkipValidatation[pds.ValidatePdsWorkloads] = true
	})

	It("Kill Teleport Agent Pod When Workload is running", func() {
		for _, ds := range NewPdsParams.DataServiceToTest {
			Step("Deploy DataService and start workloads", func() {
				deployment, err = WorkflowDataService.DeployDataService(ds, ds.Image, ds.Version, PDS_DEFAULT_NAMESPACE)
				log.FailOnError(err, "Error while deploying ds")
				log.Debugf("Source Deployment Id: [%s]", *deployment.Create.Meta.Uid)
			})

			stepLog := "Starting workload before killing teleport agent pod"
			Step(stepLog, func() {
				//Initializing the parameters required for workload generation
				WorkflowDataService.WorkloadGenParams = &dslibs.LoadGenParams{
					LoadGenDepName:  NewPdsParams.LoadGen.LoadGenDepName,
					Namespace:       PDS_DEFAULT_NAMESPACE,
					NumOfRows:       NewPdsParams.LoadGen.NumOfRows,
					Timeout:         NewPdsParams.LoadGen.Timeout,
					Replicas:        NewPdsParams.LoadGen.Replicas,
					FailOnError:     NewPdsParams.LoadGen.FailOnError,
					ReplacePassword: NewPdsParams.LoadGen.ReplacePassword,
					Iterations:      "0",
					Mode:            "crud",
					ClusterMode:     "true",
					TableName:       "testwl",
				}

				if ds.Name == "Redis" && ds.Replicas == 1 {
					WorkflowDataService.WorkloadGenParams.ClusterMode = "false"
				}
				crdName := dslibs.CrdMap[ds.Name]
				_, _, err = dslibs.GenerateWorkload(*deployment.Create.Status.CustomResourceName, ds.Name, crdName, *WorkflowDataService.WorkloadGenParams)
				log.FailOnError(err, "Error while running workloads on ds")
			})

			Step("Delete System Pods while workload is running", func() {
				log.InfoD("Delete Px System Agent Pods")
				err := WorkflowDataService.DeletePDSPods([]string{"px-teleport-agent", "pds-deployments-operator", "px-tc-operator"}, PlatformNamespace)
				log.FailOnError(err, fmt.Sprintf("Error happened while killing px pods during deployment test for data service %v", *deployment.Create.Status.CustomResourceName))
			})
		}
	})

	JustAfterEach(func() {
		defer EndPDSTorpedoTest()
		delete(WorkflowDataService.SkipValidatation, pds.ValidatePdsWorkloads)
		delete(WorkflowDataService.SkipValidatation, pds.ValidatePdsDeployment)
	})
})

var _ = Describe("{RebootReplicaNodeDuringStorageResize}", func() {
	var (
		deployment *automationModels.PDSDeploymentResponse
		err        error
	)

	JustBeforeEach(func() {
		StartPDSTorpedoTest("RebootReplicaNodeDuringStorageResize", "Reboot the Node where replica is present during  application's storage is resized", nil, 0)
	})

	It("Reboot Replica Node during storage resize", func() {
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

			Step("Reboot Node and replica node while storage size increase", func() {
				log.InfoD("Reboot replica node while storage size increase")
				// Global Resiliency TC marker
				pdsResLib.MarkResiliencyTC(true)
				// Type of failure that this TC needs to cover
				failuretype := pdsResLib.TypeOfFailure{
					Type: pdsResLib.RebootReplicaNodeDuringStorageResize,
					Method: func() error {
						return RebootReplicaVolumeNode(PDS_DEFAULT_NAMESPACE, *deployment.Create.Status.CustomResourceName)
					},
				}

				pdsResLib.DefineFailureType(failuretype)
				pdsResLib.AccountID = AccID
				err = pdsResLib.InduceFailureAfterWaitingForCondition(&deployment.Create, PDS_DEFAULT_NAMESPACE, 1, ds)
				log.FailOnError(err, fmt.Sprintf("Error happened while executing Kill Agent Pod test for data service %v", *deployment.Create.Status.CustomResourceName))
			})

			Step("Validate Data Service to after replica Node reboot and storageresrtart", func() {
				log.InfoD("Validate Data Service to after px restart")
				err = WorkflowDataService.ValidatePdsDataServiceDeployments(*deployment.Create.Meta.Uid, ds, ds.ScaleReplicas, pdsResLib.UpdateTemplate, WorkflowDataService.PDSTemplates.StorageTemplateId, PDS_DEFAULT_NAMESPACE, ds.Version, ds.Image)
				log.FailOnError(err, "Error while Validating dataservice after replica node reboot")
			})

			stepLog := "Running Workloads after replica Node reboot and StorageResize"
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

var _ = Describe("{KillDbMasterNodeDuringStorageResize}", func() {
	var (
		deployment *automationModels.PDSDeploymentResponse
		err        error
	)

	JustBeforeEach(func() {
		StartPDSTorpedoTest("KillDbMasterNodeDuringStorageResize", "Kill the DB master node during application's storage is resized", nil, 0)
	})

	It("Reboot Replica Node during storage resize", func() {
		for _, ds := range NewPdsParams.DataServiceToTest {
			Step("Deploy DataService", func() {
				if (ds.Name == postgresql) || (ds.Name == mysql) {
					deployment, err = WorkflowDataService.DeployDataService(ds, ds.Image, ds.Version, PDS_DEFAULT_NAMESPACE)
					log.FailOnError(err, "Deployment failed")
					log.Debugf("Source Deployment Id: [%s]", *deployment.Create.Meta.Uid)
				} else {
					log.InfoD("This testcase is valid only for SQL databases (PG & MYSQL), Skipping this testcase as DB is- [%v]", ds.Name)
				}
			})

			Step("Create and associate update template to the project", func() {
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

			Step("Reboot Node and replica node while storage size increase", func() {
				log.InfoD("Reboot replica node while storage size increase")
				// Global Resiliency TC marker
				pdsResLib.MarkResiliencyTC(true)
				// Type of failure that this TC needs to cover
				failuretype := pdsResLib.TypeOfFailure{
					Type: pdsResLib.KillDbMasterNodeDuringStorageResize,
					Method: func() error {
						return WorkflowDataService.KillDBMasterNodeToValidateHA(*deployment.Create.Meta.Uid)
					},
				}

				pdsResLib.DefineFailureType(failuretype)
				pdsResLib.AccountID = AccID
				err = pdsResLib.InduceFailureAfterWaitingForCondition(&deployment.Create, PDS_DEFAULT_NAMESPACE, 1, ds)
				log.FailOnError(err, fmt.Sprintf("Error happened while executing Kill Agent Pod test for data service %v", *deployment.Create.Status.CustomResourceName))
			})

			Step("Validate Data Service to DB MasterNode Restart", func() {
				log.InfoD("Validate Data Service to after DB Master node reboot")
				err = WorkflowDataService.ValidatePdsDataServiceDeployments(*deployment.Create.Meta.Uid, ds, ds.ScaleReplicas, pdsResLib.UpdateTemplate, WorkflowDataService.PDSTemplates.StorageTemplateId, PDS_DEFAULT_NAMESPACE, ds.Version, ds.Image)
				log.FailOnError(err, "Error while Validating dataservice after px-agent reboot")
			})

			stepLog := "Running Workloads after DB Master Node reboot and Storage Resize"
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
