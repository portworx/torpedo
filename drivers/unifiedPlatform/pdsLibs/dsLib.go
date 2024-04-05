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

func UpdateDataService(ds PDSDataService, deploymentId, namespaceId, projectId, imageId, appConfigId, resConfigId, stConfigId string) (*automationModels.PDSDeploymentResponse, error) {
	log.Info("Update Data service will be performed")
	//depInputs := automationModels.PDSDeploymentRequest{}

	depInputs := &automationModels.PDSDeploymentRequest{
		Update: automationModels.PDSDeploymentUpdate{
			NamespaceID: namespaceId,
			ProjectID:   projectId,
			V1Deployment: automationModels.V1DeploymentUpdate{
				Meta: automationModels.Meta{
					Uid:             &deploymentId,
					Name:            nil,
					Description:     nil,
					ResourceVersion: nil,
					CreateTime:      nil,
					UpdateTime:      nil,
					Labels:          nil,
					Annotations:     nil,
				},
				Config: automationModels.DeploymentUpdateConfig{
					DeploymentMeta: automationModels.Meta{
						Uid:             nil,
						Name:            nil,
						Description:     nil,
						ResourceVersion: nil,
						CreateTime:      nil,
						UpdateTime:      nil,
						Labels:          nil,
						Annotations:     nil,
					},
					DeploymentConfig: automationModels.V1Config1{
						TlsEnabled: nil,
						DeploymentTopologies: []automationModels.DeploymentTopology{
							{
								Name:     StringPtr("pds-qa-test-topology"),
								Replicas: intToPointerString(ds.ScaleReplicas),
							},
						},
					},
				},
			},
		},
	}
	deployment, err := v2Components.PDS.UpdateDeployment(depInputs)
	if err != nil {
		return nil, err
	}
	return deployment, err
}

// DeleteDeployment Deletes the given deployment
func DeleteDeployment(deployment map[string]string) error {
	_, deploymentId := GetDeploymentNameAndId(deployment)
	return v2Components.PDS.DeleteDeployment(deploymentId)
}

// DeployDataService Deploys the dataservices based on the given params
func DeployDataService(ds PDSDataService, namespaceId, projectId, targetClusterId, imageId, appConfigId, resConfigId, stConfigId string) (*automationModels.PDSDeploymentResponse, error) {
	log.Info("Data service will be deployed as per the config map passed..")

	depInputs := &automationModels.PDSDeploymentRequest{
		Create: automationModels.PDSDeployment{
			NamespaceID: namespaceId,
			ProjectID:   projectId,
			V1Deployment: automationModels.V1Deployment{
				Meta: automationModels.Meta{
					Name: &ds.DeploymentName,
				},
				Config: automationModels.V1Config1{
					References: automationModels.Reference{
						TargetClusterId: targetClusterId,
						ProjectId:       &projectId,
						ImageId:         &imageId,
					},
					TlsEnabled: nil,
					DeploymentTopologies: []automationModels.DeploymentTopology{
						{
							Name:     StringPtr("pds-qa-test-topology"),
							Replicas: intToPointerString(ds.Replicas),
							ResourceSettings: &automationModels.PdsTemplates{
								Id: &resConfigId,
							},
							ServiceConfigurations: &automationModels.PdsTemplates{
								Id: &appConfigId,
							},
							StorageOptions: &automationModels.PdsTemplates{
								Id: &stConfigId,
							},
						},
					},
				},
			},
		},
	}

	log.Infof("deployment name  [%s]", *depInputs.Create.V1Deployment.Meta.Name)
	log.Infof("app template ids [%s]", *depInputs.Create.V1Deployment.Config.DeploymentTopologies[0].ServiceConfigurations.Id)
	log.Infof("resource template ids [%s]", *depInputs.Create.V1Deployment.Config.DeploymentTopologies[0].ResourceSettings.Id)
	log.Infof("storage template ids [%s]", *depInputs.Create.V1Deployment.Config.DeploymentTopologies[0].StorageOptions.Id)

	log.Infof("depInputs [+%v]", depInputs.Create)
	deployment, err := v2Components.PDS.CreateDeployment(depInputs)
	if err != nil {
		return nil, err
	}
	return deployment, err
}

// GetDataServiceId gets the DataService's ID
func GetDataServiceId(dsName string) (string, error) {
	ds, err := v2Components.PDS.ListDataServices()
	if err != nil {
		return "", fmt.Errorf("Failed to list DataServices: %v", err)
	}
	for _, dataService := range ds.DataServiceList {
		log.Debugf("Dataservice name: [%s]", *dataService.Meta.Name)
		if *dataService.Meta.Name == dsName {
			return *dataService.Meta.Uid, nil
		}
	}
	return "", fmt.Errorf("Failed to find DataService with name %s", dsName)
}

func ListDataServiceVersions(dsId string) (*automationModels.CatalogResponse, error) {
	input := automationModels.WorkFlowRequest{
		DataServiceId: dsId,
	}
	ds, err := v2Components.PDS.ListDataServiceVersions(&input)
	return ds, err
}

func ListDataServiceImages(dsId, dsVersionId string) (*automationModels.CatalogResponse, error) {
	input := automationModels.WorkFlowRequest{
		DataServiceId:        dsId,
		DataServiceVersionId: dsVersionId,
	}
	ds, err := v2Components.PDS.ListDataServiceImages(&input)
	return ds, err
}
