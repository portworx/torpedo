package pdslibs

import (
	"fmt"
	"github.com/portworx/sched-ops/k8s/apps"
	"github.com/portworx/sched-ops/k8s/core"
	"github.com/portworx/sched-ops/k8s/rbac"
	"github.com/portworx/torpedo/drivers/unifiedPlatform"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/pkg/log"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"strconv"
	"time"
)

var (
	v2Components *unifiedPlatform.UnifiedPlatformComponents
	namespaceId  string
	err          error
)

var (
	k8sCore = core.Instance()
	k8sApps = apps.Instance()
	k8sRbac = rbac.Instance()
)

var (
	validateDeploymentTimeOut      = 50 * time.Minute
	validateDeploymentTimeInterval = 60 * time.Second
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

func UpdateDataService(ds PDSDataService, namespaceId, projectId string) (*automationModels.WorkFlowResponse, error) {
	log.Info("Update Data service will be performed")

	depInputs := automationModels.PDSDeploymentRequest{}

	// TODO call the below methods and fill up the structs
	// Get TargetClusterID
	// Get ImageID
	// Get App, Resource and storage PdsTemplates Ids

	depInputs.Update.V1Deployment.Config.DeploymentTopologies = []automationModels.DeploymentTopology{{}}

	depInputs.Update.V1Deployment.Meta.Name = &ds.DeploymentName
	depInputs.Create.NamespaceID = namespaceId
	depInputs.Create.ProjectID = projectId
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

func ValidateDataServiceDeployment(deployment map[string]string, namespace string) error {
	var (
		ss             *v1.StatefulSet
		deploymentName string
		deploymentId   string
	)

	log.Debugf("deployment name [%s] in namespace [%s]", deployment[""], namespace)

	for key, value := range deployment {
		deploymentName = key
		deploymentId = value
	}

	err = wait.Poll(validateDeploymentTimeInterval, validateDeploymentTimeOut, func() (bool, error) {
		ss, err = k8sApps.GetStatefulSet(deploymentName, namespace)
		if err != nil {
			log.Warnf("An Error Occured while getting statefulsets %v", err)
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		log.Errorf("An Error Occured while getting statefulsets %v", err)
		return err
	}

	//validate the statefulset deployed in the k8s namespace
	err = k8sApps.ValidateStatefulSet(ss, validateDeploymentTimeOut)
	if err != nil {
		log.Errorf("An Error Occured while validating statefulsets %v", err)
		return err
	}

	//TODO : Update the below code once we get the deployment status api
	log.Infof("DeploymentId [%s]", deploymentId)
	//err = wait.Poll(maxtimeInterval, validateDeploymentTimeOut, func() (bool, error) {
	//	status, res, err := components.DataServiceDeployment.GetDeploymentStatus(deployment.GetId())
	//	log.Infof("Health status -  %v", status.GetHealth())
	//	if err != nil {
	//		log.Errorf("Error occured while getting deployment status %v", err)
	//		return false, nil
	//	}
	//	if res.StatusCode != state.StatusOK {
	//		log.Errorf("Error when calling `ApiDeploymentsIdCredentialsGet``: %v\n", err)
	//		log.Errorf("Full HTTP response: %v\n", res)
	//		return false, err
	//	}
	//	if status.GetHealth() != PdsDeploymentAvailable {
	//		return false, nil
	//	}
	//	log.Infof("Deployment details: Health status -  %v,Replicas - %v, Ready replicas - %v", status.GetHealth(), status.GetReplicas(), status.GetReadyReplicas())
	//	return true, nil
	//})
	return err
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

func intToPointerString(n int) *string {
	// Convert the integer to a string
	str := strconv.Itoa(n)
	// Create a pointer to the string
	ptr := &str
	// Return the pointer to the string
	return ptr
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
