package pdslibs

import (
	"fmt"
	"github.com/portworx/torpedo/drivers/unifiedPlatform"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/pkg/log"
)

// InitUnifiedApiComponents
func InitUnifiedApiComponents(controlPlaneURL, accountID string) error {
	v2Components, err = unifiedPlatform.NewUnifiedPlatformComponents(controlPlaneURL, accountID)
	if err != nil {
		return err
	}
	return nil
}

func UpdateDataService(ds PDSDataService, namespaceId, projectId string) (*automationModels.WorkFlowResponse, error) {
	log.Info("Update Data service will be performed")

	depInputs := automationModels.PDSDeploymentRequest{}

	// TODO call the below methods and fill up the structs
	// Get TargetClusterID
	// Get ImageID
	// Get App, Resource and storage PdsTemplates Ids

	depInputs.Update.V1Deployment.Config.DeploymentTopologies = []automationModels.DeploymentTopology{{}}

	depInputs.Update.V1Deployment.Meta.Name = &ds.DeploymentName
	depInputs.Update.NamespaceID = namespaceId
	depInputs.Update.ProjectID = projectId
	depInputs.Update.V1Deployment.Config.DeploymentTopologies[0].Replicas = intToPointerString(ds.ScaleReplicas)
	depInputs.Update.V1Deployment.Config.DeploymentTopologies[0].ResourceSettings = &automationModels.PdsTemplates{
		Id:              intToPointerString(10),
		ResourceVersion: nil,
		Values:          nil,
	}
	depInputs.Update.V1Deployment.Config.DeploymentTopologies[0].ServiceConfigurations = &automationModels.PdsTemplates{
		Id:              intToPointerString(11),
		ResourceVersion: nil,
		Values:          nil,
	}
	depInputs.Update.V1Deployment.Config.DeploymentTopologies[0].StorageOptions = &automationModels.PdsTemplates{
		Id:              intToPointerString(12),
		ResourceVersion: nil,
		Values:          nil,
	}

	//TODO: Get the namespaceID, write method to get the namespaceID from the give namespace
	deployment, err := v2Components.PDS.UpdateDeployment(&depInputs)
	if err != nil {
		return nil, err
	}
	return deployment, err
}

func DeleteDeployment(deployment map[string]string) error {
	_, deploymentId := GetDeploymentNameAndId(deployment)
	return v2Components.PDS.DeleteDeployment(deploymentId)
}

// DeployDataService should be called from workflows
func DeployDataService(ds PDSDataService, namespaceId, projectId, targetClusterId string) (*automationModels.WorkFlowResponse, error) {
	log.Info("Data service will be deployed as per the config map passed..")

	depInputs := automationModels.PDSDeploymentRequest{}

	// TODO call the below methods and fill up the structs
	// Get ImageID
	// Get App, Resource and storage PdsTemplates Ids

	depInputs.Create.V1Deployment.Config.DeploymentTopologies = []automationModels.DeploymentTopology{{}}

	depInputs.Create.V1Deployment.Meta.Name = &ds.DeploymentName
	depInputs.Create.NamespaceID = namespaceId
	depInputs.Create.ProjectID = projectId
	depInputs.Create.V1Deployment.Config.References.TargetClusterId = targetClusterId
	depInputs.Create.V1Deployment.Config.References.ProjectId = &projectId
	depInputs.Create.V1Deployment.Config.DeploymentTopologies[0].ResourceSettings = &automationModels.PdsTemplates{
		Id:              intToPointerString(10),
		ResourceVersion: nil,
		Values:          nil,
	}
	depInputs.Create.V1Deployment.Config.DeploymentTopologies[0].ServiceConfigurations = &automationModels.PdsTemplates{
		Id:              intToPointerString(11),
		ResourceVersion: nil,
		Values:          nil,
	}
	depInputs.Create.V1Deployment.Config.DeploymentTopologies[0].StorageOptions = &automationModels.PdsTemplates{
		Id:              intToPointerString(12),
		ResourceVersion: nil,
		Values:          nil,
	}

	//TODO: Get the namespaceID, write method to get the namespaceID from the give namespace

	log.Infof("deployment name  [%s]", *depInputs.Create.V1Deployment.Meta.Name)
	log.Infof("app template ids [%s]", *depInputs.Create.V1Deployment.Config.DeploymentTopologies[0].ServiceConfigurations.Id)
	log.Infof("resource template ids [%s]", *depInputs.Create.V1Deployment.Config.DeploymentTopologies[0].ResourceSettings.Id)
	log.Infof("storage template ids [%s]", *depInputs.Create.V1Deployment.Config.DeploymentTopologies[0].StorageOptions.Id)

	log.Infof("depInputs [+%v]", depInputs.Create)
	deployment, err := v2Components.PDS.CreateDeployment(&depInputs)
	if err != nil {
		return nil, err
	}
	return deployment, err
}

func GetDataServiceId(dsName string) (string, error) {
	ds, err := v2Components.PDS.ListDataServices()
	if err != nil {
		return "", fmt.Errorf("Failed to list DataServices: %v", err)
	}
	for _, dataService := range ds {
		if dataService.Meta.Name == &dsName {
			return dataService.Id, nil
		}
	}
	return "", fmt.Errorf("Failed to find DataService with name %s", dsName)
}

func ListDataServiceVersions(dsId string) ([]automationModels.WorkFlowResponse, error) {
	input := automationModels.WorkFlowRequest{
		DataServiceId: dsId,
	}
	ds, err := v2Components.PDS.ListDataServiceVersions(&input)
	return ds, err
}

func ListDataServiceImages(dsId string) ([]automationModels.WorkFlowResponse, error) {
	input := automationModels.WorkFlowRequest{
		DataServiceId: dsId,
	}
	ds, err := v2Components.PDS.ListDataServiceImages(&input)
	return ds, err
}
