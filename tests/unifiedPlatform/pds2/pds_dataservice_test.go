package tests

import (
	pdslib "github.com/portworx/torpedo/drivers/pds/lib"
	dslibs "github.com/portworx/torpedo/drivers/unifiedPlatform/pdsLibs"
	corev1 "k8s.io/api/core/v1"

	"fmt"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	"github.com/portworx/sched-ops/k8s/core"
	"github.com/portworx/torpedo/drivers/node"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/stworkflows/pds"
	"github.com/portworx/torpedo/drivers/utilities"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests"
	. "github.com/portworx/torpedo/tests/unifiedPlatform"
)

const (
	PlatformNamespace = "px-system"
)

var _ = Describe("{DeployDataServicesOnDemandAndScaleUp}", func() {
	var (
		deployment *automationModels.PDSDeploymentResponse
		err        error
	)

	JustBeforeEach(func() {
		StartPDSTorpedoTest("DeployDataServicesOnDemandAndScaleUp", "Deploy data services and perform scale up", nil, 0)
	})

	It("Deploy,Validate and ScaleUp DataService", func() {
		for _, ds := range NewPdsParams.DataServiceToTest {
			Step("Deploy DataService", func() {
				deployment, err = WorkflowDataService.DeployDataService(ds, ds.Image, ds.Version, PDS_DEFAULT_NAMESPACE)
				log.FailOnError(err, "Error while deploying ds")
				log.Debugf("Source Deployment Id: [%s]", *deployment.Create.Meta.Uid)
			})

			Step("ScaleUp DataService", func() {
				log.InfoD("Scaling Up dataServices...")
				updateDeployment, err := WorkflowDataService.UpdateDataService(ds, *deployment.Create.Meta.Uid, ds.Image, ds.Version)
				log.FailOnError(err, "Error while updating ds")
				log.Debugf("Updated Deployment Id: [%s]", *updateDeployment.Update.Meta.Uid)
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

var _ = Describe("{UpgradeDataServiceImage}", func() {
	var (
		deployment *automationModels.PDSDeploymentResponse
		err        error
	)

	JustBeforeEach(func() {
		StartPDSTorpedoTest("UpgradeDataServiceImage", "Upgrade Data Service Image", nil, 0)
	})

	It("Deploy, Validate and Upgrade Data service Image", func() {
		for _, ds := range NewPdsParams.DataServiceToTest {
			Step("Deploy DataService", func() {
				deployment, err = WorkflowDataService.DeployDataService(ds, ds.OldImage, ds.Version, PDS_DEFAULT_NAMESPACE)
				log.FailOnError(err, "Error while deploying ds")
				log.Debugf("Source Deployment Id: [%s]", *deployment.Create.Meta.Uid)
			})

			Step("Upgrade DataService Image", func() {
				_, err := WorkflowDataService.UpdateDataService(ds, *deployment.Create.Meta.Uid, ds.Image, ds.Version)
				log.FailOnError(err, "Error while updating ds")
			})

			stepLog := "Running Workloads after upgrading the ds image"
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

var _ = Describe("{ScaleUpCpuMemLimitsandStorageOfDS}", func() {
	var (
		deployment *automationModels.PDSDeploymentResponse
		err        error
	)

	JustBeforeEach(func() {
		StartPDSTorpedoTest("ScaleUpCpuMemLimitsandStorageOfDS", "Deploy a dataservice and scale up its CPU/MEM limits and storage size by editing the respective template", nil, 0)
	})

	It("Deploy,Validate and ScaleUp DataService", func() {
		for _, ds := range NewPdsParams.DataServiceToTest {
			Step("Deploy DataService", func() {
				deployment, err = WorkflowDataService.DeployDataService(ds, ds.Image, ds.Version, PDS_DEFAULT_NAMESPACE)
				log.FailOnError(err, "Error while deploying ds")
				log.Debugf("Source Deployment Id: [%s]", *deployment.Create.Meta.Uid)
			})

			//Update Ds With New Values of Resource Templates
			resConfigIdUpdated, err := WorkflowPDSTemplate.CreateResourceTemplateWithCustomValue(NewPdsParams)
			log.FailOnError(err, "Unable to create Custom Templates for PDS")
			log.InfoD("Updated Resource Template ID- [updated- %v]", resConfigIdUpdated)

			WorkflowDataService.UpdateDeploymentTemplates = true
			_, err = WorkflowDataService.UpdateDataService(ds, *deployment.Create.Meta.Uid, ds.Image, ds.Version)
			log.FailOnError(err, "Error while updating ds")

			stepLog := "Running Workloads after upgrading the ds image"
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

var _ = Describe("{GetPVCFullCondition}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("GetPVCFullCondition", "Deploy a dataservice and fill-up the PVC, Once full, resize the PVC", nil, 0)
	})
	var (
		workflowDataservice pds.WorkflowDataService
		workFlowTemplates   pds.WorkflowPDSTemplates
		deployment          *automationModels.PDSDeploymentResponse
		deployments         = make(map[dslibs.PDSDataService]*automationModels.PDSDeploymentResponse)
		templates           []string
		err                 error
	)
	It("Deploy and Validate DataService", func() {
		Step("Create a PDS Namespace", func() {
			Namespace = strings.ToLower("pds-test-ns-" + utilities.RandString(5))
			WorkflowNamespace.TargetCluster = &WorkflowTargetCluster
			workFlowTemplates.Platform = WorkflowPlatform
			WorkflowNamespace.Namespaces = make(map[string]string)
			workflowNamespace, err := WorkflowNamespace.CreateNamespaces(Namespace)
			log.FailOnError(err, "Unable to create namespace")
			log.Infof("Namespaces created - [%s]", workflowNamespace.Namespaces)
			log.Infof("Namespace id - [%s]", workflowNamespace.Namespaces[Namespace])
		})

		for _, ds := range NewPdsParams.DataServiceToTest {
			workflowDataservice.Namespace = &WorkflowNamespace
			deployment, err = workflowDataservice.DeployDataService(ds, ds.OldImage, ds.OldVersion, PDS_DEFAULT_NAMESPACE)
			log.FailOnError(err, "Error while deploying ds")
			deployments[ds] = deployment

			defer func() {
				Step("Delete PDS CustomTemplates", func() {
					log.InfoD("Cleaning Up templates...")
					err := workFlowTemplates.DeleteCreatedCustomPdsTemplates(templates)
					log.FailOnError(err, "Error while deleting dataservice")
				})
			}()

			defer func() {
				for _, deployment := range deployments {
					Step("Delete DataServiceDeployment", func() {
						log.InfoD("Cleaning Up dataservice...")
						err := workflowDataservice.DeleteDeployment(*deployment.Create.Meta.Uid)
						log.FailOnError(err, "Error while deleting dataservice")
					})
				}
			}()

			log.InfoD("Running Workloads to fill up the PVC")
			_, err = workflowDataservice.RunDataServiceWorkloads(*deployment.Create.Meta.Uid)
			log.FailOnError(err, "Error while running workloads on ds")

			log.InfoD("Compute the PVC usage")
			err = workflowDataservice.CheckPVCStorageFullCondition(workflowDataservice.DataServiceDeployment[*deployment.Create.Meta.Uid].Namespace, *deployment.Create.Status.CustomResourceName, 85)
			log.FailOnError(err, "Error while checking for pvc full condition")

			log.InfoD("Once pvc has reached threshold, increase the ovc by 1gb")
			err = workflowDataservice.IncreasePvcSizeBy1gb(workflowDataservice.DataServiceDeployment[*deployment.Create.Meta.Uid].Namespace, *deployment.Create.Status.CustomResourceName, 1)
			log.FailOnError(err, "Failing while Increasing the PVC name...")

			//log.InfoD("Validate deployment after PVC increase")
			//err = workflowDataservice.ValidatePdsDataServiceDeployments(*deployment.Create.Meta.Uid, ds, ds.Replicas, resConfigId, stConfigId, workflowDataservice.DataServiceDeployment[*deployment.Create.Meta.Uid].Namespace, ds.Version, ds.Image)
		}
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
	})
})

var _ = Describe("{DeletePDSPods}", func() {
	var (
		deployment *automationModels.PDSDeploymentResponse
		err        error
	)

	JustBeforeEach(func() {
		StartPDSTorpedoTest("DeletePDSPods", "delete pds pods and validate if its coming back online and dataServices are not affected", nil, 0)
	})

	It("Delete pds pods and validate if its coming back online and dataservices are not affected", func() {
		for _, ds := range NewPdsParams.DataServiceToTest {
			Step("Deploy DataService", func() {
				deployment, err = WorkflowDataService.DeployDataService(ds, ds.Image, ds.Version, PDS_DEFAULT_NAMESPACE)
				log.FailOnError(err, "Error while deploying ds")
				log.Debugf("Source Deployment Id: [%s]", *deployment.Create.Meta.Uid)
			})

			stepLog := "Running Workloads before deleting pods in Px-System namespace"
			Step(stepLog, func() {
				_, err := WorkflowDataService.RunDataServiceWorkloads(*deployment.Create.Meta.Uid)
				log.FailOnError(err, "Error while running workloads on ds")
			})

			Step("Delete PDSPods", func() {
				err := WorkflowDataService.DeletePDSPods([]string{"pds-backups", "pds-target"}, PlatformNamespace)
				log.FailOnError(err, "Error while deleting pds pods")
				err = WorkflowDataService.ValidatePdsDataServiceDeployments(
					*deployment.Create.Meta.Uid,
					ds,
					ds.Replicas,
					WorkflowDataService.PDSTemplates.ResourceTemplateId,
					WorkflowDataService.PDSTemplates.StorageTemplateId,
					PDS_DEFAULT_NAMESPACE,
					ds.Version,
					ds.Image)
				log.FailOnError(err, "Error while Validating dataservice")
			})

			stepLog = "Running Workloads after deleting pods in Px-System namespace"
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

var _ = Describe("{ValidatePdsHealthIncaseofFailures}", func() {
	var (
		deployment *automationModels.PDSDeploymentResponse
		err        error
	)

	JustBeforeEach(func() {
		StartPDSTorpedoTest("ValidatePdsHealthIncaseofFailures", "Deploy data services and validate PDS health in case of PDS pod deletion", nil, 0)
	})

	It("Deploy data services, Delete Pds Agent pods and perform backup and restore on the same cluster", func() {
		for _, ds := range NewPdsParams.DataServiceToTest {

			steplog := "Deploy dataservice"
			Step(steplog, func() {
				log.InfoD(steplog)
				deployment, err = WorkflowDataService.DeployDataService(ds, ds.Image, ds.Version, PDS_DEFAULT_NAMESPACE)
				log.FailOnError(err, "Error while deploying ds")
				log.Infof("All deployments - [%+v]", WorkflowDataService.DataServiceDeployment)
				WorkflowPDSRestore.SourceDeploymentConfigBeforeUpgrade = &deployment.Create.Config.DeploymentTopologies[0]
			})

			steplog = "Restart PDS Agent Pods and Validate if it comes up"
			Step(steplog, func() {
				log.InfoD(steplog)
				err := WorkflowDataService.DeletePDSPods([]string{*deployment.Create.Status.CustomResourceName}, PDS_DEFAULT_NAMESPACE)
				log.FailOnError(err, "Error while deleting pds pods")
				err = WorkflowDataService.ValidatePdsDataServiceDeployments(
					*deployment.Create.Meta.Uid,
					ds,
					ds.Replicas,
					WorkflowDataService.PDSTemplates.ResourceTemplateId,
					WorkflowDataService.PDSTemplates.StorageTemplateId,
					PDS_DEFAULT_NAMESPACE,
					ds.Version,
					ds.Image)
				log.FailOnError(err, "Error while Validating dataservice")
			})

			steplog = "ScaleUp DataService"
			Step(steplog, func() {
				log.InfoD(steplog)
				updateDeployment, err := WorkflowDataService.UpdateDataService(ds, *deployment.Create.Meta.Uid, ds.Image, ds.Version)
				log.FailOnError(err, "Error while updating ds")
				log.Debugf("Updated Deployment Id: [%s]", *updateDeployment.Update.Meta.Uid)
			})
		}
	})

	JustAfterEach(func() {
		defer EndPDSTorpedoTest()
	})
})

var _ = Describe("{DrainAndDecommissionNode}", func() {
	var (
		deployment          *automationModels.PDSDeploymentResponse
		err                 error
		nodeName            string
		k8sCore             core.Ops
		timeOut             time.Duration
		maxtimeInterval     time.Duration
		deploymentNamespace string
	)

	JustBeforeEach(func() {
		StartPDSTorpedoTest("DrainAndDecommissionNode", "Deploys a data service, drains one selected node, decommissions that node", nil, 0)
		k8sCore = core.Instance()
		timeOut = 30 * time.Minute
		maxtimeInterval = 30 * time.Second
	})

	It("Deploys a data service, drains one selected node, decommissions that node", func() {
		for _, ds := range NewPdsParams.DataServiceToTest {
			Step("Deploy DataService", func() {
				deployment, err = WorkflowDataService.DeployDataService(ds, ds.Image, ds.Version, PDS_DEFAULT_NAMESPACE)
				log.FailOnError(err, "Error while deploying ds")
				log.Debugf("Source Deployment Id: [%s]", *deployment.Create.Meta.Uid)
				nodes, err := pdslib.GetNodesOfSS(*deployment.Create.Status.CustomResourceName, PDS_DEFAULT_NAMESPACE)
				log.FailOnError(err, "Cannot fetch nodes of the running Data Service")
				nodeName = nodes[0].Name // Selecting the 1st node in the list to cordon
			})

			steplog := "Drain Pods from a node"
			Step(steplog, func() {
				log.InfoD(steplog)
				podsList, err := pdslib.GetPodsOfSsByNode(*deployment.Create.Status.CustomResourceName, nodeName, PDS_DEFAULT_NAMESPACE)
				log.FailOnError(err, fmt.Sprintf("Pod not found on this Node : %s", nodeName))
				log.InfoD("Pods found on %v node. Trying to Drain pods from this node now.", nodeName)
				err = k8sCore.DrainPodsFromNode(nodeName, podsList, timeOut, maxtimeInterval)
				log.FailOnError(err, fmt.Sprintf("Draining pod from the node %s failed", nodeName))
				log.InfoD("Pods successfully drained from the node %s", nodeName)
			})

			steplog = "Validate Data Service to see if Pods have rescheduled on another node"
			Step(steplog, func() {
				log.InfoD(steplog)
				err = WorkflowDataService.ValidatePdsDataServiceDeployments(*deployment.Create.Meta.Uid, ds, ds.Replicas, WorkflowDataService.PDSTemplates.ResourceTemplateId, WorkflowDataService.PDSTemplates.StorageTemplateId, PDS_DEFAULT_NAMESPACE, ds.Version, ds.Image)
				log.FailOnError(err, "Error while Validating dataservice after cordoned node")
			})

			steplog = "Validate no pods are on the cordoned node anymore"
			Step(steplog, func() {
				log.InfoD(steplog)
				nodes, err := pdslib.GetNodesOfSS(*deployment.Create.Status.CustomResourceName, PDS_DEFAULT_NAMESPACE)
				log.FailOnError(err, fmt.Sprintf("Cannot fetch nodes of the running Data Service %v", *deployment.Create.Status.CustomResourceName))
				for _, nodeObj := range nodes {
					if nodeObj.Name == nodeName {
						log.FailOnError(fmt.Errorf("New Pod came up on the node that was cordoned."), "Unexpected error")
					}
				}
				log.InfoD("The pods of the Stateful Set %v are not on the cordoned node. Moving ahead now.", *deployment.Create.Status.CustomResourceName)
			})

			steplog = "Create a namespace for PDS"
			Step(steplog, func() {
				log.InfoD(steplog)
				deploymentNamespace = fmt.Sprintf("%s-%s", strings.ToLower(ds.Name), RandomString(5))
				_, err := WorkflowNamespace.CreateNamespaces(deploymentNamespace)
				log.FailOnError(err, "Error while creating namespace for New Deployment")
				log.Infof("Namespaces created - [%s]", WorkflowNamespace.Namespaces)
			})

			steplog = "Associate namespace to the project"
			Step(steplog, func() {
				log.InfoD(steplog)
				err := WorkflowProject.Associate(
					[]string{},
					[]string{WorkflowNamespace.Namespaces[deploymentNamespace]},
					[]string{},
					[]string{},
					[]string{},
					[]string{},
				)
				log.FailOnError(err, "Error while associating namespace to the project")
			})

			Step("Deploy DataService", func() {
				deployment, err = WorkflowDataService.DeployDataService(ds, ds.Image, ds.Version, deploymentNamespace)
				log.FailOnError(err, "Error while deploying ds")
				log.Debugf("Source Deployment Id: [%s]", *deployment.Create.Meta.Uid)
				nodes, err := pdslib.GetNodesOfSS(*deployment.Create.Status.CustomResourceName, deploymentNamespace)
				log.FailOnError(err, "Cannot fetch nodes of the running Data Service")
				for _, nodeObj := range nodes {
					if nodeObj.Name == nodeName {
						log.FailOnError(fmt.Errorf("New Pod came up on the node that was cordoned."), "Unexpected error")
					}
				}
			})

			steplog = "UnCordon Selected Node"
			Step(steplog, func() {
				log.InfoD(steplog)
				err = k8sCore.UnCordonNode(nodeName, timeOut, maxtimeInterval)
				log.FailOnError(err, fmt.Sprintf("UnCordoning the node %s Failed", nodeName))
				log.InfoD("Node %s successfully UnCordoned", nodeName)
			})

		}
	})

	JustAfterEach(func() {
		defer EndPDSTorpedoTest()
	})
})

var _ = Describe("{RollingRebootNodes}", func() {
	var (
		deployment *automationModels.PDSDeploymentResponse
		err        error
	)

	JustBeforeEach(func() {
		StartPDSTorpedoTest("RollingRebootNodes", "Reboot node(s) while the data services will be running", nil, 0)
	})

	It("Reboot node(s) while the data services will be running", func() {
		for _, ds := range NewPdsParams.DataServiceToTest {
			Step("Deploy DataService", func() {
				deployment, err = WorkflowDataService.DeployDataService(ds, ds.Image, ds.Version, PDS_DEFAULT_NAMESPACE)
				log.FailOnError(err, "Error while deploying ds")
				log.Debugf("Source Deployment Id: [%s]", *deployment.Create.Meta.Uid)
			})

			steplog := "Reboot nodes"
			Step(steplog, func() {
				log.InfoD("Reboot nodes")
				nodesToReboot := node.GetWorkerNodes()
				err = RebootNodes(nodesToReboot)
				log.FailOnError(err, "Error while rebooting nodes")
			})

			steplog = "Validate Data Service to see if Pods have rescheduled on another node"
			Step(steplog, func() {
				log.InfoD(steplog)
				err = WorkflowDataService.ValidatePdsDataServiceDeployments(*deployment.Create.Meta.Uid, ds, ds.Replicas, WorkflowDataService.PDSTemplates.ResourceTemplateId, WorkflowDataService.PDSTemplates.StorageTemplateId, PDS_DEFAULT_NAMESPACE, ds.Version, ds.Image)
				log.FailOnError(err, "Error while Validating dataservice after node reboot node")
			})

			steplog = "Running Workloads after node reoot"
			Step(steplog, func() {
				log.InfoD(steplog)
				_, err := WorkflowDataService.RunDataServiceWorkloads(*deployment.Create.Meta.Uid)
				log.FailOnError(err, "Error while running workloads on ds")
			})

		}
	})

	JustAfterEach(func() {
		defer EndPDSTorpedoTest()
	})
})

var _ = Describe("{ScaleDownScaleupPXCluster}", func() {
	var (
		deployment *automationModels.PDSDeploymentResponse
		err        error
		nodeList   []*corev1.Node
	)

	JustBeforeEach(func() {
		StartPDSTorpedoTest("ScaleDownScaleupPXCluster", "Scales Down the PX cluster, Verify Data Services and Scales up the Px cluster", nil, 0)
	})

	It("Scales Down the PX cluster, Verify Data Services and Scales up the Px cluster", func() {

		for _, ds := range NewPdsParams.DataServiceToTest {
			Step("Deploy DataService", func() {
				deployment, err = WorkflowDataService.DeployDataService(ds, ds.Image, ds.Version, PDS_DEFAULT_NAMESPACE)
				log.FailOnError(err, "Error while deploying ds")
				log.Debugf("Source Deployment Id: [%s]", *deployment.Create.Meta.Uid)
				nodes, err := pdslib.GetNodesOfSS(*deployment.Create.Status.CustomResourceName, PDS_DEFAULT_NAMESPACE)
				log.FailOnError(err, "Error while getting Data Serice Nodes")
				nodeList = append(nodeList, nodes[0])
			})

			Step("Scale down PX Nodes on which Data Services are running", func() {
				// disable PX Pod on the first node of each deployed Data Service
				err := StopPxServiceOnNodes(nodeList)
				log.FailOnError(err, "unable to stop px service on given nodes")
				log.InfoD("Successfully Scaled Down PX Nodes...")
			})

			log.InfoD("Sleeping 300 seconds after scale down of Px Nodes before we check the health of Data Services")
			time.Sleep(300 * time.Second)

			Step("Verify the Data Services status after scaling down the Px Nodes", func() {
				log.InfoD("Verify the Data Services status after scaling down the Px Nodes")
				err = WorkflowDataService.ValidatePdsDataServiceDeployments(*deployment.Create.Meta.Uid, ds, ds.Replicas, WorkflowDataService.PDSTemplates.ResourceTemplateId, WorkflowDataService.PDSTemplates.StorageTemplateId, PDS_DEFAULT_NAMESPACE, ds.Version, ds.Image)
				log.FailOnError(err, "Error while Validating dataservice after px nodes scale down")
			})

			stepLog := "Running Workloads after scale down of PX Nodes"
			Step(stepLog, func() {
				_, err := WorkflowDataService.RunDataServiceWorkloads(*deployment.Create.Meta.Uid)
				log.FailOnError(err, "Error while running workloads on ds")
			})

			Step("Scale Up PX  Nodes", func() {
				err := StartPxServiceOnNodes(nodeList)
				log.FailOnError(err, "unable to start px service on given nodes")
				log.InfoD("Successfully Scaled up Px Nodes")
			})
			log.InfoD("Sleeping 300 seconds again after scale up of Px Nodes before we check the health of Data Services")
			time.Sleep(300 * time.Second)

			Step("Verify the Data Services status after scaling up the Px Nodes", func() {
				log.InfoD("Verify the Data Services status after scaling up the Px Nodes")
				err = WorkflowDataService.ValidatePdsDataServiceDeployments(*deployment.Create.Meta.Uid, ds, ds.Replicas, WorkflowDataService.PDSTemplates.ResourceTemplateId, WorkflowDataService.PDSTemplates.StorageTemplateId, PDS_DEFAULT_NAMESPACE, ds.Version, ds.Image)
				log.FailOnError(err, "Error while Validating dataservice after px nodes scale up")
			})

			stepLog = "Running Workloads after scale up of PX Nodes"
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
