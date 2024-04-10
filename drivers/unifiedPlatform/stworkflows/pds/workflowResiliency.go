package pds

import (
	resiLibs "github.com/portworx/torpedo/drivers/unifiedPlatform/pdsLibs"
	"github.com/portworx/torpedo/pkg/log"
)

type WorkflowResiliency struct {
	ScenarioType   string
	ErrorType      error
	ResiliencyFlag bool
	WfDataService  *WorkflowDataService
}

// MarkResiliencyTC Function to enable Resiliency Test
func (wkflwResi *WorkflowResiliency) MarkResiliencyTC(resiliency bool) {
	wkflwResi.ResiliencyFlag = resiliency
	log.InfoD("Execution of a Resiliency TestCase Begins ...")
}
func (wkflwResi *WorkflowResiliency) InduceFailureAndExecuteResiliencyScenario(ds resiLibs.PDSDataService, failureType string) error {
	wfDataService := wkflwResi.WfDataService
	namespace := wfDataService.Namespace.Namespaces[wfDataService.NamespaceName]
	projectId := wfDataService.Namespace.TargetCluster.Project.ProjectId
	targetClusterId := wfDataService.Namespace.TargetCluster.ClusterUID
	appConfigId := wfDataService.PDSTemplates.ServiceConfigTemplateId
	resConfigId := wfDataService.PDSTemplates.ResourceTemplateId
	stConfigId := wfDataService.PDSTemplates.StorageTemplateId
	image := ds.OldImage
	version := ds.OldImage
	imageId, err := resiLibs.GetDataServiceImageId(ds.Name, image, version)
	err = resiLibs.InduceFailureAfterWaitingForCondition(ds, namespace, projectId, targetClusterId, imageId, appConfigId, resConfigId, stConfigId, namespace, failureType, wkflwResi.ResiliencyFlag)
	if err != nil {
		return err
	}
	return nil
}
