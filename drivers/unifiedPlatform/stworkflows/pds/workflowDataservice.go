package pds

import (
	"fmt"
	"github.com/portworx/torpedo/drivers/pds/parameters"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	dslibs "github.com/portworx/torpedo/drivers/unifiedPlatform/pdsLibs"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/stworkflows/platform"
	utils "github.com/portworx/torpedo/drivers/utilities"
	"github.com/portworx/torpedo/pkg/aetosutil"
	"github.com/portworx/torpedo/pkg/log"
)

type WorkflowDataService struct {
	Namespace                     platform.WorkflowNamespace
	PDSTemplates                  WorkflowPDSTemplates
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
		deploymentName, err := dslibs.ValidateDataServiceDeploymentHealth(*deployment.Create.Meta.Uid, namespace)
		if err != nil {
			return nil, err
		}

		// Validate if the dns endpoint is reachable
		err = wfDataService.ValidateDNSEndpoint(*deployment.Create.Meta.Uid)
		if err != nil {
			return nil, err
		}

		// Update the actual deploymentName with deploymentId
		wfDataService.DataServiceDeployment[deploymentName] = *deployment.Create.Meta.Uid

		// Get data service deployment resources
		resourceTemplateOps, storageOps, DeploymentConfigs, err := wfDataService.GetDsDeploymentResources(wfDataService.DataServiceDeployment, ds.Name, resConfigId, stConfigId, namespace)
		if err != nil {
			return nil, err
		}

		// Validate deployment resources
		//TODO: Initialize the dataServiceVersionBuildMap once list ds version api is available
		var dataServiceVersionBuildMap = make(map[string][]string)
		wfDataService.ValidateDeploymentResources(resourceTemplateOps, storageOps, DeploymentConfigs, ds.Replicas, dataServiceVersionBuildMap)
	}

	return deployment, nil
}

func (wfDataService *WorkflowDataService) GetDsDeploymentResources(deployment map[string]string, dataServiceName, resourceTemplateID, storageTemplateID, namespace string) (dslibs.ResourceSettingTemplate, dslibs.StorageOps, dslibs.DeploymentConfig, error) {
	var (
		resourceTemp dslibs.ResourceSettingTemplate
		storageOp    dslibs.StorageOps
		dbConfig     dslibs.DeploymentConfig
		err          error
	)
	deploymentName, deploymentId := GetDeploymentNameAndId(deployment)
	log.Debugf("deployment Name [%s] and Id [%s]", deploymentName, deploymentId)

	dbConfig, err = dslibs.GetDeploymentConfigurations(namespace, dataServiceName, deploymentName)
	if err != nil {
		return resourceTemp, storageOp, dbConfig, err
	}

	resourceTemp, err = dslibs.GetResourceTemplateConfigs(resourceTemplateID)
	if err != nil {
		return resourceTemp, storageOp, dbConfig, err
	}

	storageOp, err = dslibs.GetStorageTemplateConfigs(storageTemplateID)
	if err != nil {
		return resourceTemp, storageOp, dbConfig, err
	}

	return resourceTemp, storageOp, dbConfig, err

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
		deploymentName, err := dslibs.ValidateDataServiceDeploymentHealth(deploymentId, namespace)
		if err != nil {
			return nil, err
		}

		wfDataService.DataServiceDeployment[deploymentName] = *deployment.Update.Config.DeploymentMeta.Uid

		// Validate if the dns endpoint is reachable
		err = wfDataService.ValidateDNSEndpoint(deploymentId)
		if err != nil {
			return nil, err
		}

		// Get deployment resources
		dslibs.GetDeploymentResources(wfDataService.DataServiceDeployment, ds.Name, "resource-template-id", "storage-template-id", namespace)

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

func (wfDataService *WorkflowDataService) ValidateDNSEndpoint(deploymentId string) error {
	deployment, err := dslibs.GetDeployment(deploymentId)
	if err != nil {
		return err
	}
	log.Infof("Deployment Response [+%v]", *deployment)
	log.Infof("ConnectionInfo Response [+%v]", deployment.Get.Status.ConnectionInfo["clusterDetails"])

	clusterDetails := deployment.Get.Status.ConnectionInfo["clusterDetails"]
	dnsEndPoint, err := utils.GetDNSEndPoint(clusterDetails)
	if err != nil {
		return err
	}

	err = dslibs.ValidateDNSEndPoint(dnsEndPoint)
	if err != nil {
		return err
	}

	return nil
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

func (wfDataService *WorkflowDataService) ValidateDeploymentResources(resourceTemp dslibs.ResourceSettingTemplate, storageOp dslibs.StorageOps, config dslibs.DeploymentConfig, replicas int, dataServiceVersionBuildMap map[string][]string) {
	log.Debugf("filesystem used %v ", config.Spec.Topologies[0].StorageOptions.Filesystem)
	log.Debugf("storage replicas used %v ", config.Spec.Topologies[0].StorageOptions.Replicas)
	log.Debugf("cpu requests used %v ", config.Spec.Topologies[0].Resources.Requests.CPU)
	log.Debugf("memory requests used %v ", config.Spec.Topologies[0].Resources.Requests.Memory)
	log.Debugf("storage requests used %v ", config.Spec.Topologies[0].Resources.Requests.Storage)
	log.Debugf("No of nodes requested %v ", config.Spec.Topologies[0].Nodes)
	log.Debugf("volume group %v ", storageOp.VolumeGroup)
	log.Debugf("resource template values cpu req [%s]", resourceTemp.Resources.Requests.CPU)

	dash.VerifyFatal(resourceTemp.Resources.Requests.CPU, config.Spec.Topologies[0].Resources.Requests.CPU, "Validating CPU Request")
	dash.VerifyFatal(resourceTemp.Resources.Requests.Memory, config.Spec.Topologies[0].Resources.Requests.Memory, "Validating Memory Request")
	dash.VerifyFatal(resourceTemp.Resources.Requests.Storage, config.Spec.Topologies[0].Resources.Requests.Storage, "Validating storage")
	dash.VerifyFatal(resourceTemp.Resources.Limits.CPU, config.Spec.Topologies[0].Resources.Limits.CPU, "Validating CPU Limits")
	dash.VerifyFatal(resourceTemp.Resources.Limits.Memory, config.Spec.Topologies[0].Resources.Limits.Memory, "Validating Memory Limits")
	dash.VerifyFatal(storageOp.Replicas, config.Spec.Topologies[0].StorageOptions.Replicas, "Validating storage replicas")
	dash.VerifyFatal(storageOp.Filesystem, config.Spec.Topologies[0].StorageOptions.Filesystem, "Validating filesystems")
	dash.VerifyFatal(replicas, config.Spec.Topologies[0].Nodes, "Validating ds node replicas")

	//for version, build := range dataServiceVersionBuildMap {
	//	dash.VerifyFatal(config.Version, version+"-"+build[0], "validating ds build and version")
	//}
}

func (wfDataService *WorkflowDataService) IncreasePvcSizeBy1gb(namespace string, deployment map[string]string, sizeInGb uint64) error {
	_, err := utils.IncreasePVCby1Gig(namespace, deployment, sizeInGb)
	return err
}

func (wfDataService *WorkflowDataService) KillDBMasterNodeToValidateHA(dsName string, deploymentName string) error {
	dbMaster, isNativelyDistributed := utils.GetDbMasterNode(wfDataService.NamespaceName, dsName, deploymentName, wfDataService.Namespace.TargetCluster.KubeConfig)
	if isNativelyDistributed {
		err := utils.DeleteK8sPods(dbMaster, wfDataService.NamespaceName, wfDataService.Namespace.TargetCluster.KubeConfig)
		if err != nil {
			return err
		}
		//validate DataService Deployment here
		newDbMaster, _ := utils.GetDbMasterNode(wfDataService.NamespaceName, dsName, deploymentName, wfDataService.Namespace.TargetCluster.KubeConfig)
		if dbMaster == newDbMaster {
			log.FailOnError(fmt.Errorf("leader node is not reassigned"), fmt.Sprintf("Leader pod %v", dbMaster))
		}
	} else {
		podName, err := utils.GetAnyPodName(deploymentName, wfDataService.NamespaceName)
		if err != nil {
			return fmt.Errorf("failed while fetching pod for stateful set %v ", deploymentName)
		}
		err = utils.KillPodsInNamespace(wfDataService.NamespaceName, podName)
		if err != nil {
			return fmt.Errorf("failed while deleting pod %v ", deploymentName)
		}
		//validate DataService Deployment here
	}
	return nil
}
