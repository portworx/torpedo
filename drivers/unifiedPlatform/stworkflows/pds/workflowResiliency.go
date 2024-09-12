package pds

import (
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	resiLibs "github.com/portworx/torpedo/drivers/unifiedPlatform/pdsLibs"
	"github.com/portworx/torpedo/pkg/log"
)

type WorkflowPDSResiliency struct {
	ScenarioType   string
	ErrorType      error
	ResiliencyFlag bool
	WfDataService  *WorkflowDataService
}

// MarkResiliencyTC Function to enable Resiliency Test
func (wkflwResi *WorkflowPDSResiliency) MarkResiliencyTC(resiliency bool) {
	wkflwResi.ResiliencyFlag = resiliency
	log.InfoD("Execution of a Resiliency TestCase Begins ...")
}

func (wkflwResi *WorkflowPDSResiliency) InduceFailureAndExecuteResiliencyScenario(ds resiLibs.PDSDataService, deployment *automationModels.PDSDeploymentResponse, failureType string) error {
	wfDataService := wkflwResi.WfDataService
	deploymentId := *deployment.Create.Meta.Uid
	namespaceName := wfDataService.DataServiceDeployment[deploymentId].Namespace
	namespaceId := wfDataService.DataServiceDeployment[deploymentId].NamespaceId
	projectId := wfDataService.Namespace.TargetCluster.Project.ProjectId
	appConfigId := wfDataService.PDSTemplates.ServiceConfigTemplateIds[ds.Name]
	resConfigId := wfDataService.PDSTemplates.ResourceTemplateId
	stConfigId := wfDataService.PDSTemplates.StorageTemplateId
	image := ds.Image
	version := ds.Version
	imageId, err := resiLibs.GetDataServiceImageId(ds.Name, image, version)

	_, dsPodName, err := resiLibs.GetDeploymentAndPodDetails(deploymentId)
	if err != nil {
		return err
	}

	err = resiLibs.InduceFailureAfterWaitingForCondition(ds, dsPodName+"-0", deploymentId, namespaceId, projectId, imageId, appConfigId, resConfigId, stConfigId, namespaceName, failureType, wkflwResi.ResiliencyFlag)
	if err != nil {
		return err
	}
	return nil
}
