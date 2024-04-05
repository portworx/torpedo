package stworkflows

import (
	"github.com/portworx/torpedo/drivers/pds/parameters"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	dslibs "github.com/portworx/torpedo/drivers/unifiedPlatform/pdsLibs"
	"github.com/portworx/torpedo/pkg/aetosutil"
	"github.com/portworx/torpedo/pkg/log"
	"strconv"
)

type WorkflowDataService struct {
	Namespace                     WorkflowNamespace
	PDSTemplates                  CustomTemplates
	NamespaceName                 string
	DataServiceDeployment         map[string]string
	RestoredDataServiceDeployment map[string]string
	SkipValidatation              map[string]bool
	SourceDeploymentMd5Hash       map[string]string
	RestoredDeploymentMd5Hash     map[string]string
}

const (
	ValidatePdsDeployment = "VALIDATE_PDS_DEPLOYMENT"
	ValidatePdsWorkloads  = "VALIDATE_PDS_WORKLOADS"
)

var (
	dash *aetosutil.Dashboard
)

func (wfDataService *WorkflowDataService) DeployDataService(ds dslibs.PDSDataService, image, version string) (*automationModels.PDSDeploymentResponse, error) {
	namespace := wfDataService.Namespace.Namespaces[wfDataService.NamespaceName]
	projectId := wfDataService.Namespace.TargetCluster.Project.ProjectId
	targetClusterId := wfDataService.Namespace.TargetCluster.ClusterUID
	appConfigId := wfDataService.PDSTemplates.ServiceConfigTemplateId
	resConfigId := wfDataService.PDSTemplates.ResourceTemplateId
	stConfigId := wfDataService.PDSTemplates.StorageTemplateId
	log.Infof("targetClusterId [%s]", targetClusterId)

	imageId, err := dslibs.GetDataServiceImageId(ds.Name, image, version)
	if err != nil {
		return nil, err
	}

	log.Debugf("DS Image id-[%s]", imageId)
	deployment, err := dslibs.DeployDataService(ds, namespace, projectId, targetClusterId, imageId, appConfigId, resConfigId, stConfigId)
	if err != nil {
		return nil, err
	}
	wfDataService.DataServiceDeployment = make(map[string]string)
	wfDataService.DataServiceDeployment[*deployment.Create.Meta.Name] = *deployment.Create.Meta.Uid
	if value, ok := wfDataService.SkipValidatation[ValidatePdsDeployment]; ok {
		if value == true {
			log.Infof("Skipping DataService Deployment  Validation")
		}
	} else {
		// Validate the sts object and health of the pds deployment
		err = dslibs.ValidateDataServiceDeployment(*deployment.Create.Meta.Uid, namespace)
		if err != nil {
			return nil, err
		}

		// Get deployment resources
		//resourceTemp, storageOp, config, err := dslibs.GetDeploymentResources(wfDataService.DataServiceDeployment, ds.Name, resConfigId, stConfigId, namespace)
		//if err != nil {
		//	return nil, err
		//}

		// Validate deployment resources
		//TODO: Initialize the dataServiceVersionBuildMap once list ds version api is available
		//var dataServiceVersionBuildMap = make(map[string][]string)
		//ValidateDeploymentResources(resourceTemp, storageOp, config, ds.Replicas, dataServiceVersionBuildMap)
	}

	return deployment, nil
}

func (wfDataService *WorkflowDataService) UpdateDataService(ds dslibs.PDSDataService, deploymentId, image, version string) (*automationModels.PDSDeploymentResponse, error) {
	namespace := wfDataService.Namespace.Namespaces[wfDataService.NamespaceName]
	projectId := wfDataService.Namespace.TargetCluster.Project.ProjectId
	targetClusterId := wfDataService.Namespace.TargetCluster.ClusterUID
	appConfigId := wfDataService.PDSTemplates.ServiceConfigTemplateId
	resConfigId := wfDataService.PDSTemplates.ResourceTemplateId
	stConfigId := wfDataService.PDSTemplates.StorageTemplateId
	log.Infof("targetClusterId [%s]", targetClusterId)

	imageId, err := dslibs.GetDataServiceImageId(ds.Name, image, version)
	if err != nil {
		return nil, err
	}

	deployment, err := dslibs.UpdateDataService(ds, deploymentId, namespace, projectId, imageId, appConfigId, resConfigId, stConfigId)
	if err != nil {
		return nil, err
	}
	log.Debugf("Updated Deployment [%v]", deployment)
	wfDataService.DataServiceDeployment = make(map[string]string)
	wfDataService.DataServiceDeployment[*deployment.Update.Config.DeploymentMeta.Name] = *deployment.Update.Config.DeploymentMeta.Uid
	if value, ok := wfDataService.SkipValidatation[ValidatePdsDeployment]; ok {
		if value == true {
			log.Infof("Skipping Validation")
		}
	} else {
		// Validate the sts object and health of the pds deployment
		err = dslibs.ValidateDataServiceDeployment(deploymentId, namespace)
		if err != nil {
			return nil, err
		}

		// Get deployment resources
		//resourceTemp, storageOp, config, err := dslibs.GetDeploymentResources(wfDataService.DataServiceDeployment, ds.Name, "resource-template-id", "storage-template-id", namespace)
		//if err != nil {
		//	return nil, err
		//}

		// Validate deployment resources
		//TODO: Initialize the dataServiceVersionBuildMap once list ds version api is available
		//var dataServiceVersionBuildMap = make(map[string][]string)
		//ValidateDeploymentResources(resourceTemp, storageOp, config, ds.Replicas, dataServiceVersionBuildMap)
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

	chkSum, wlDep, err := dslibs.InsertDataAndReturnChecksum(wfDataService.DataServiceDeployment, wkloadParams)
	if err != nil {
		return err
	}

	deploymentName, _ := GetDeploymentNameAndId(wfDataService.DataServiceDeployment)

	wfDataService.SourceDeploymentMd5Hash[deploymentName] = chkSum

	return dslibs.DeleteWorkloadDeployments(wlDep)
}

func (wfDataService *WorkflowDataService) ValidateDataServiceWorkloads(params *parameters.NewPDSParams, restoredDeployment *automationModels.PDSRestoreResponse) error {
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

	deployment := make(map[string]string)
	deployment[*restoredDeployment.Create.Meta.Name] = *restoredDeployment.Create.Meta.Uid
	chkSum, wlDep, err := dslibs.ReadDataAndReturnChecksum(deployment, wkloadParams)
	if err != nil {
		return err
	}

	deploymentName, _ := GetDeploymentNameAndId(deployment)

	wfDataService.RestoredDeploymentMd5Hash[deploymentName] = chkSum

	result := dslibs.ValidateDataMd5Hash(wfDataService.SourceDeploymentMd5Hash, wfDataService.RestoredDeploymentMd5Hash)

	dash.VerifyFatal(result, true, "Validate md5 hash after restore")

	return dslibs.DeleteWorkloadDeployments(wlDep)
}

func GetDeploymentNameAndId(deployment map[string]string) (string, string) {
	var (
		deploymentName string
		deploymentId   string
	)

	for key, value := range deployment {
		deploymentName = key
		deploymentId = value
	}

	return deploymentName, deploymentId

}

func ValidateDeploymentResources(resourceTemp dslibs.ResourceSettingTemplate, storageOp dslibs.StorageOptions, config dslibs.StorageClassConfig, replicas int, dataServiceVersionBuildMap map[string][]string) {
	log.InfoD("filesystem used %v ", config.Parameters.Fs)
	log.InfoD("storage replicas used %v ", config.Parameters.Fg)
	log.InfoD("cpu requests used %v ", config.Resources.Requests.CPU)
	log.InfoD("memory requests used %v ", config.Resources.Requests.Memory)
	log.InfoD("storage requests used %v ", config.Resources.Requests.EphemeralStorage)
	log.InfoD("No of nodes requested %v ", config.Replicas)
	log.InfoD("volume group %v ", storageOp.VolumeGroup)

	dash.VerifyFatal(resourceTemp.Resources.Requests.CPU, config.Resources.Requests.CPU, "Validating CPU Request")
	dash.VerifyFatal(resourceTemp.Resources.Requests.Memory, config.Resources.Requests.Memory, "Validating Memory Request")
	dash.VerifyFatal(resourceTemp.Resources.Requests.Storage, config.Resources.Requests.EphemeralStorage, "Validating storage")
	dash.VerifyFatal(resourceTemp.Resources.Limits.CPU, config.Resources.Limits.CPU, "Validating CPU Limits")
	dash.VerifyFatal(resourceTemp.Resources.Limits.Memory, config.Resources.Limits.Memory, "Validating Memory Limits")
	repl, err := strconv.Atoi(config.Parameters.Repl)
	log.FailOnError(err, "failed on atoi method")
	dash.VerifyFatal(storageOp.Replicas, int32(repl), "Validating storage replicas")
	dash.VerifyFatal(storageOp.Filesystem, config.Parameters.Fs, "Validating filesystems")
	dash.VerifyFatal(config.Replicas, replicas, "Validating ds node replicas")

	for version, build := range dataServiceVersionBuildMap {
		dash.VerifyFatal(config.Version, version+"-"+build[0], "validating ds build and version")
	}
}
