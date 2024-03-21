package stworkflows

import (
	"github.com/portworx/torpedo/drivers/pds/parameters"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	dslibs "github.com/portworx/torpedo/drivers/unifiedPlatform/pdsLibs"
	"github.com/portworx/torpedo/pkg/log"
)

type WorkflowDataService struct {
	Namespace             WorkflowNamespace
	NamespaceName         string
	DataServiceDeployment map[string]string
	SkipValidatation      map[string]bool
}

const (
	ValidatePdsDeployment = "VALIDATE_PDS_DEPLOYMENT"
)

func (wfDataService *WorkflowDataService) DeployDataService(ds dslibs.PDSDataService) (*automationModels.WorkFlowResponse, error) {
	namespace := wfDataService.Namespace.Namespaces[wfDataService.NamespaceName]
	projectId := wfDataService.Namespace.TargetCluster.Project.ProjectId
	targetClusterId := wfDataService.Namespace.TargetCluster.ClusterUID
	log.Infof("targetClusterId [%s]", targetClusterId)
	deployment, err := dslibs.DeployDataService(ds, namespace, projectId, targetClusterId)
	if err != nil {
		return nil, err
	}
	wfDataService.DataServiceDeployment[*deployment.PDSDeployment.V1Deployment.Meta.Name] = *deployment.PDSDeployment.V1Deployment.Meta.Uid

	if value, ok := wfDataService.SkipValidatation[ValidatePdsDeployment]; ok {
		if value == true {
			log.Infof("Skipping Validation")
		}
	} else {
		// Validate the sts object and health of the pds deployment
		err = dslibs.ValidateDataServiceDeployment(wfDataService.DataServiceDeployment, namespace)
		if err != nil {
			return nil, err
		}

		// Get deployment resources
		resourceTemp, storageOp, config, err := dslibs.GetDeploymentResources(wfDataService.DataServiceDeployment, ds.Name, "resource-template-id", "storage-template-id", namespace)
		if err != nil {
			return nil, err
		}

		// Validate deployment resources
		//TODO: Initialize the dataServiceVersionBuildMap once list ds version api is available
		var dataServiceVersionBuildMap = make(map[string][]string)
		dslibs.ValidateDeploymentResources(resourceTemp, storageOp, config, ds.Replicas, dataServiceVersionBuildMap)
	}

	return deployment, nil
}

func (wfDataService *WorkflowDataService) UpdateDataService(ds dslibs.PDSDataService) (*automationModels.WorkFlowResponse, error) {
	namespace := wfDataService.Namespace.Namespaces[wfDataService.NamespaceName]
	projectId := wfDataService.Namespace.TargetCluster.Project.ProjectId
	targetClusterId := wfDataService.Namespace.TargetCluster.ClusterUID
	log.Infof("targetClusterId [%s]", targetClusterId)

	deployment, err := dslibs.UpdateDataService(ds, namespace, projectId)
	if err != nil {
		return nil, err
	}
	wfDataService.DataServiceDeployment[*deployment.PDSDeployment.V1Deployment.Meta.Name] = *deployment.PDSDeployment.V1Deployment.Meta.Uid
	if value, ok := wfDataService.SkipValidatation[ValidatePdsDeployment]; ok {
		if value == true {
			log.Infof("Skipping Validation")
		}
	} else {
		// Validate the sts object and health of the pds deployment
		err = dslibs.ValidateDataServiceDeployment(wfDataService.DataServiceDeployment, namespace)
		if err != nil {
			return nil, err
		}

		// Get deployment resources
		resourceTemp, storageOp, config, err := dslibs.GetDeploymentResources(wfDataService.DataServiceDeployment, ds.Name, "resource-template-id", "storage-template-id", namespace)
		if err != nil {
			return nil, err
		}

		// Validate deployment resources
		//TODO: Initialize the dataServiceVersionBuildMap once list ds version api is available
		var dataServiceVersionBuildMap = make(map[string][]string)
		dslibs.ValidateDeploymentResources(resourceTemp, storageOp, config, ds.Replicas, dataServiceVersionBuildMap)
	}
	return deployment, nil
}

func (wfDataService *WorkflowDataService) DeleteDeployment() error {
	return dslibs.DeleteDeployment(wfDataService.DataServiceDeployment)
}

func (wfDataService *WorkflowDataService) RunDataServiceWorkloads(params *parameters.NewPDSParams) error {

	//Initializing the parameters required for workload generation
	wkloadParams := dslibs.LoadGenParams{
		LoadGenDepName: params.LoadGen.LoadGenDepName,
		Namespace:      params.InfraToTest.Namespace,
		NumOfRows:      params.LoadGen.NumOfRows,
		Timeout:        params.LoadGen.Timeout,
		Replicas:       params.LoadGen.Replicas,
		TableName:      params.LoadGen.TableName,
		Iterations:     params.LoadGen.Iterations,
		FailOnError:    params.LoadGen.FailOnError,
	}

	_, wlDep, err := dslibs.InsertDataAndReturnChecksum(wfDataService.DataServiceDeployment, wkloadParams)
	if err != nil {
		return err
	}

	return dslibs.DeleteWorkloadDeployments(wlDep)
}
