package dataservice

import (
	"github.com/portworx/torpedo/drivers/unifiedPlatform"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
	"github.com/portworx/torpedo/pkg/log"
	"strconv"
)

var (
	v2Components *unifiedPlatform.UnifiedPlatformComponents
	namespaceId  string
	err          error
)

type PDSDataService struct {
	DeploymentName        string "json:\"DeploymentName\""
	Name                  string "json:\"Name\""
	Version               string "json:\"Version\""
	Image                 string "json:\"Image\""
	Replicas              int    "json:\"Replicas\""
	ScaleReplicas         int    "json:\"ScaleReplicas\""
	OldVersion            string "json:\"OldVersion\""
	OldImage              string "json:\"OldImage\""
	DataServiceEnabledTLS bool   "json:\"DataServiceEnabledTLS\""
	ServiceType           string "json:\"ServiceType\""
}

// InitUnifiedApiComponents
func InitUnifiedApiComponents(controlPlaneURL, accountID string) error {
	v2Components, err = unifiedPlatform.NewUnifiedPlatformComponents(controlPlaneURL, accountID)
	if err != nil {
		return err
	}
	return nil
}

func UpdateDataService(ds PDSDataService) (*apiStructs.WorkFlowResponse, error) {
	log.Info("Update Data service will be performed")

	depInputs := apiStructs.WorkFlowRequest{}

	// TODO call the below methods and fill up the structs
	// Get TargetClusterID
	// Get ImageID
	// Get ProjectID
	// Get App, Resource and storage Template Ids

	depInputs.Deployment.V1Deployment.Config.DeploymentTopologies = []apiStructs.DeploymentTopology{{}}

	depInputs.Deployment.V1Deployment.Meta.Name = &ds.DeploymentName
	depInputs.Deployment.NamespaceID = "nam:6a9bead4-5e2e-473e-b325-ceeda5bbbce6"
	depInputs.Deployment.V1Deployment.Config.DeploymentTopologies[0].ResourceTemplate = &apiStructs.Template{
		Id:              intToPointerString(10),
		ResourceVersion: nil,
		Values:          nil,
	}
	depInputs.Deployment.V1Deployment.Config.DeploymentTopologies[0].ApplicationTemplate = &apiStructs.Template{
		Id:              intToPointerString(11),
		ResourceVersion: nil,
		Values:          nil,
	}
	depInputs.Deployment.V1Deployment.Config.DeploymentTopologies[0].StorageTemplate = &apiStructs.Template{
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

// DeployDataService should be called from workflows
func DeployDataService(ds PDSDataService) (*apiStructs.WorkFlowResponse, error) {
	log.Info("Data service will be deployed as per the config map passed..")

	depInputs := apiStructs.WorkFlowRequest{}

	// TODO call the below methods and fill up the structs
	// Get TargetClusterID
	// Get ImageID
	// Get ProjectID
	// Get App, Resource and storage Template Ids

	depInputs.Deployment.V1Deployment.Config.DeploymentTopologies = []apiStructs.DeploymentTopology{{}}

	depInputs.Deployment.V1Deployment.Meta.Name = &ds.DeploymentName
	depInputs.Deployment.NamespaceID = "nam:6a9bead4-5e2e-473e-b325-ceeda5bbbce6"
	depInputs.Deployment.V1Deployment.Config.DeploymentTopologies[0].ResourceTemplate = &apiStructs.Template{
		Id:              intToPointerString(10),
		ResourceVersion: nil,
		Values:          nil,
	}
	depInputs.Deployment.V1Deployment.Config.DeploymentTopologies[0].ApplicationTemplate = &apiStructs.Template{
		Id:              intToPointerString(11),
		ResourceVersion: nil,
		Values:          nil,
	}
	depInputs.Deployment.V1Deployment.Config.DeploymentTopologies[0].StorageTemplate = &apiStructs.Template{
		Id:              intToPointerString(12),
		ResourceVersion: nil,
		Values:          nil,
	}

	//TODO: Get the namespaceID, write method to get the namespaceID from the give namespace

	log.Infof("deployment name  [%s]", *depInputs.Deployment.V1Deployment.Meta.Name)
	log.Infof("app template ids [%s]", *depInputs.Deployment.V1Deployment.Config.DeploymentTopologies[0].ApplicationTemplate.Id)
	log.Infof("resource template ids [%s]", *depInputs.Deployment.V1Deployment.Config.DeploymentTopologies[0].ResourceTemplate.Id)
	log.Infof("storage template ids [%s]", *depInputs.Deployment.V1Deployment.Config.DeploymentTopologies[0].StorageTemplate.Id)

	log.Infof("depInputs [+%v]", depInputs.Deployment)
	deployment, err := v2Components.PDS.CreateDeployment(&depInputs)
	if err != nil {
		return nil, err
	}
	return deployment, err
}

func ValidateDataServiceDeployment() {
	log.Info("Data service will be validated in this method")
}

func intToPointerString(n int) *string {
	// Convert the integer to a string
	str := strconv.Itoa(n)
	// Create a pointer to the string
	ptr := &str
	// Return the pointer to the string
	return ptr
}
