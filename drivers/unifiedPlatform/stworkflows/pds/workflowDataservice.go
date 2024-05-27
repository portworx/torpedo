package pds

import (
	"fmt"
	"slices"
	"strings"

	"github.com/portworx/torpedo/drivers/unifiedPlatform/stworkflows"

	"time"

	"github.com/portworx/torpedo/drivers/pds/parameters"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	dslibs "github.com/portworx/torpedo/drivers/unifiedPlatform/pdsLibs"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/stworkflows/platform"
	utils "github.com/portworx/torpedo/drivers/utilities"
	"github.com/portworx/torpedo/pkg/aetosutil"
	"github.com/portworx/torpedo/pkg/log"
	corev1 "k8s.io/api/core/v1"
)

type WorkflowDataService struct {
	Namespace                 *platform.WorkflowNamespace
	PDSTemplates              WorkflowPDSTemplates
	DataServiceDeployment     map[string]*dslibs.DataServiceDetails
	SkipValidatation          map[string]bool
	Dash                      *aetosutil.Dashboard
	PDSParams                 *parameters.NewPDSParams
	ValidateStorageIncrease   dslibs.ValidateStorageIncrease
	UpdateDeploymentTemplates bool
}

const (
	ValidatePdsDeployment      = "VALIDATE_PDS_DEPLOYMENT"
	ValidatePdsWorkloads       = "VALIDATE_PDS_WORKLOADS"
	PlatformNamespace          = "px-system"
	ValidateDeploymentDeletion = "VALIDATE_DELETE_DEPLOYMENT"
	PDS_DEPLOYMENT_AVAILABLE   = "AVAILABLE"
)

func (wfDataService *WorkflowDataService) DeployDataService(ds dslibs.PDSDataService, image, version string, namespace string) (*automationModels.PDSDeploymentResponse, error) {
	namespaceId := wfDataService.Namespace.Namespaces[namespace]
	namespaceName := namespace
	projectId := wfDataService.Namespace.TargetCluster.Project.ProjectId
	targetClusterId := wfDataService.Namespace.TargetCluster.ClusterUID
	appConfigId := wfDataService.PDSTemplates.ServiceConfigTemplateIds[ds.Name]
	resConfigId := wfDataService.PDSTemplates.ResourceTemplateId
	stConfigId := wfDataService.PDSTemplates.StorageTemplateId
	log.Infof("targetClusterId [%s]", targetClusterId)

	log.InfoD("Deploying DataService [%s]", ds.Name)

	imageId, err := dslibs.GetDataServiceImageId(ds.Name, image, version)
	if err != nil {
		return nil, err
	}

	log.Debugf("DS Image id-[%s]", imageId)
	deployment, err := dslibs.DeployDataService(ds, namespaceId, projectId, targetClusterId, imageId, appConfigId, resConfigId, stConfigId)
	if err != nil {
		return nil, err
	}

	wfDataService.DataServiceDeployment[*deployment.Create.Meta.Uid] = &dslibs.DataServiceDetails{
		Deployment:        deployment.Create,
		Namespace:         namespaceName,
		NamespaceId:       namespaceId,
		SourceMd5Checksum: "",
		DSParams:          ds,
	}

	if value, ok := wfDataService.SkipValidatation[ValidatePdsWorkloads]; ok {
		if value == true {
			log.Infof("Skipping DataService Deployment  Validation")
		}
	} else {
		err = wfDataService.ValidatePdsDataServiceDeployments(*deployment.Create.Meta.Uid, ds, ds.Replicas, resConfigId, stConfigId, namespaceName, version, image)
		if err != nil {
			return nil, err
		}
	}

	if value, ok := wfDataService.SkipValidatation[ValidatePdsWorkloads]; ok {
		if value == true {
			log.Infof("Data validation is skipped for this")
		}
	} else {

		// TODO: This needs to be removed once below bugs are fixed:
		// https://purestorage.atlassian.net/issues/DS-9591
		// https://purestorage.atlassian.net/issues/DS-9546
		// https://purestorage.atlassian.net/issues/DS-9305
		log.Infof("Sleeping for 1 minutes to make sure deployment gets healthy")
		time.Sleep(1 * time.Minute)

		_, err := wfDataService.RunDataServiceWorkloads(*deployment.Create.Meta.Uid)
		if err != nil {
			return deployment, fmt.Errorf("unable to run workfload on the data service. Error - [%s]", err.Error())
		}
	}

	return deployment, nil
}

func (wfDataService *WorkflowDataService) UpdateDataService(ds dslibs.PDSDataService, deploymentId, image, version string) (*automationModels.PDSDeploymentResponse, error) {
	var deployment *automationModels.PDSDeploymentResponse

	namespaceId := wfDataService.DataServiceDeployment[deploymentId].NamespaceId
	namespaceName := wfDataService.DataServiceDeployment[deploymentId].Namespace
	projectId := wfDataService.Namespace.TargetCluster.Project.ProjectId
	targetClusterId := wfDataService.Namespace.TargetCluster.ClusterUID
	appConfigId := wfDataService.PDSTemplates.ServiceConfigTemplateIds[ds.Name]
	resConfigId := wfDataService.PDSTemplates.ResourceTemplateId
	stConfigId := wfDataService.PDSTemplates.StorageTemplateId
	newResConfigId := wfDataService.PDSTemplates.UpdateResourceTemplateId
	log.Infof("targetClusterId [%s]", targetClusterId)

	imageId, err := dslibs.GetDataServiceImageId(ds.Name, image, version)
	if err != nil {
		return nil, err
	}

	if wfDataService.UpdateDeploymentTemplates {
		log.Debugf("newResConfigId [%s]", newResConfigId)
		ds.ScaleReplicas = ds.Replicas
		deployment, err = dslibs.UpdateDataService(ds, deploymentId, namespaceId, projectId, imageId, appConfigId, newResConfigId, stConfigId)
	} else {
		deployment, err = dslibs.UpdateDataService(ds, deploymentId, namespaceId, projectId, imageId, appConfigId, resConfigId, stConfigId)
	}
	if err != nil {
		return nil, err
	}
	log.Debugf("Updated Deployment [%+v]", deployment)
	if value, ok := wfDataService.SkipValidatation[ValidatePdsDeployment]; ok {
		if value == true {
			log.Infof("Skipping Validation")
		}
	} else {
		//Validate the deploymentConfig update status
		err := dslibs.ValidateDeploymentConfigUpdate(*deployment.Update.Meta.Uid, "COMPLETED")
		if err != nil {
			return nil, err
		}

		err = wfDataService.ValidatePdsDataServiceDeployments(*deployment.Update.Config.DeploymentMeta.Uid, ds, ds.ScaleReplicas, resConfigId, stConfigId, namespaceName, version, image)
		if err != nil {
			return nil, err
		}
	}

	return deployment, nil
}

// ValidatePdsDataServiceDeployments validates the pds deployments resource, storage, deployment configurations and endpoints
func (wfDataService *WorkflowDataService) ValidatePdsDataServiceDeployments(deploymentId string, ds dslibs.PDSDataService, replicas int, resConfigId, stConfigId, namespace, version, image string) error {

	// Validate the sts object and health of the pds deployment
	err := dslibs.ValidateDataServiceDeploymentHealth(deploymentId, PDS_DEPLOYMENT_AVAILABLE)
	if err != nil {
		return err
	}

	//Validate Statefulset health
	err = dslibs.ValidateStatefulSetHealth(*wfDataService.DataServiceDeployment[deploymentId].Deployment.Status.CustomResourceName, wfDataService.DataServiceDeployment[deploymentId].Namespace)
	if err != nil {
		return err
	}

	// Validate if the dns endpoint is reachable
	err = wfDataService.ValidateDNSEndpoint(deploymentId)
	if err != nil {
		return err
	}

	// Get data service deployment resources
	resourceTemplateOps, storageOps, DeploymentConfigs, err := wfDataService.GetDsDeploymentResources(deploymentId)
	if err != nil {
		return err
	}

	// Validate deployment resources
	dataServiceVersionBuild := version + "-" + image
	wfDataService.ValidateDeploymentResources(resourceTemplateOps, storageOps, DeploymentConfigs, replicas, dataServiceVersionBuild)

	return nil
}

func (wfDataService *WorkflowDataService) GetDsDeploymentResources(deploymentId string) (dslibs.ResourceSettingTemplate, dslibs.StorageOps, dslibs.DeploymentConfig, error) {
	var (
		resourceTemp dslibs.ResourceSettingTemplate
		storageOp    dslibs.StorageOps
		dbConfig     dslibs.DeploymentConfig
		err          error
	)

	deployment, podName, err := dslibs.GetDeploymentAndPodDetails(deploymentId)
	if err != nil {
		return resourceTemp, storageOp, dbConfig, err
	}

	log.Debugf("Crd Name: [%s]", dslibs.CrdMap[strings.ToLower(wfDataService.DataServiceDeployment[deploymentId].DSParams.Name)])
	dbConfig, err = dslibs.GetDeploymentConfigurations(wfDataService.DataServiceDeployment[deploymentId].Namespace, dslibs.CrdMap[strings.ToLower(wfDataService.DataServiceDeployment[deploymentId].DSParams.Name)], podName)
	if err != nil {
		return resourceTemp, storageOp, dbConfig, err
	}

	log.Debugf("Resource Template Id After Update [%s]", *deployment.Get.Config.DeploymentTopologies[0].ResourceSettings.Id)
	resourceTemp, err = dslibs.GetResourceTemplateConfigs(*deployment.Get.Config.DeploymentTopologies[0].ResourceSettings.Id)
	if err != nil {
		return resourceTemp, storageOp, dbConfig, err
	}

	storageOp, err = dslibs.GetStorageTemplateConfigs(*deployment.Get.Config.DeploymentTopologies[0].StorageOptions.Id)
	if err != nil {
		return resourceTemp, storageOp, dbConfig, err
	}

	return resourceTemp, storageOp, dbConfig, err

}

func (wfDataService *WorkflowDataService) DeleteDeployment(deploymentId string) error {
	err := dslibs.DeleteDeployment(deploymentId)
	if err != nil {
		return err
	}
	if value, ok := wfDataService.SkipValidatation[ValidateDeploymentDeletion]; ok {
		if value == true {
			log.Infof("Skipping validation of dataservice deletion")
		}
	} else {
		err = dslibs.ValidateDeploymentIsDeleted(deploymentId)
		if err != nil {
			return err
		}
	}

	// Removing the data service entry from map
	delete(wfDataService.DataServiceDeployment, deploymentId)

	return nil
}

func (wfDataService *WorkflowDataService) ValidateDNSEndpoint(deploymentId string) error {
	deployment, err := dslibs.GetDeployment(deploymentId)
	if err != nil {
		return err
	}
	log.Infof("Deployment Response [+%v]", *deployment)
	log.Infof("ConnectionInfo Response [+%v]", deployment.Get.Status.ConnectionInfo["connectionDetails"])

	connectionDetails := deployment.Get.Status.ConnectionInfo["connectionDetails"]
	dnsEndPoint, err := utils.ParseInterfaceAndGetDetails(connectionDetails, wfDataService.DataServiceDeployment[deploymentId].DSParams.Name)
	if err != nil {
		return err
	}

	err = dslibs.ValidateDNSEndPoint(dnsEndPoint)
	if err != nil {
		return err
	}

	return nil
}

func (wfDataService *WorkflowDataService) RunDataServiceWorkloads(deploymentId string) (string, error) {
	if slices.Contains(stworkflows.SKIPDATASERVICEFROMWORKLOAD, strings.ToLower(wfDataService.DataServiceDeployment[deploymentId].DSParams.Name)) {
		log.Warnf("Workload is not enabled for this - [%s] - data service", wfDataService.DataServiceDeployment[deploymentId].DSParams.Name)
		return "", nil
	}
	//Initializing the parameters required for workload generation
	wkloadParams := dslibs.LoadGenParams{
		LoadGenDepName: wfDataService.PDSParams.LoadGen.LoadGenDepName,
		Namespace:      wfDataService.DataServiceDeployment[deploymentId].Namespace,
		NumOfRows:      wfDataService.PDSParams.LoadGen.NumOfRows,
		Timeout:        wfDataService.PDSParams.LoadGen.Timeout,
		Replicas:       wfDataService.PDSParams.LoadGen.Replicas,
		TableName:      wfDataService.PDSParams.LoadGen.TableName,
		Iterations:     wfDataService.PDSParams.LoadGen.Iterations,
		FailOnError:    wfDataService.PDSParams.LoadGen.FailOnError,
	}

	chkSum, wlDep, err := dslibs.InsertDataAndReturnChecksum(*wfDataService.DataServiceDeployment[deploymentId], wkloadParams)
	if err != nil {
		return "", err
	}

	// Updating the data hash for the deployment
	wfDataService.DataServiceDeployment[deploymentId].SourceMd5Checksum = chkSum

	return chkSum, dslibs.DeleteWorkloadDeployments(wlDep)
}

// Reads and update the md5 hash for the data service
func (wfDataService *WorkflowDataService) ReadAndUpdateDataServiceDataHash(deploymentId string) error {
	if slices.Contains(stworkflows.SKIPDATASERVICEFROMWORKLOAD, strings.ToLower(wfDataService.DataServiceDeployment[deploymentId].DSParams.Name)) {
		log.Warnf("Workload is not enabled for this - [%s] - data service", wfDataService.DataServiceDeployment[deploymentId].DSParams.Name)
		return nil
	}
	wkloadParams := dslibs.LoadGenParams{
		LoadGenDepName: wfDataService.PDSParams.LoadGen.LoadGenDepName,
		Namespace:      wfDataService.DataServiceDeployment[deploymentId].Namespace,
		NumOfRows:      wfDataService.PDSParams.LoadGen.NumOfRows,
		Timeout:        wfDataService.PDSParams.LoadGen.Timeout,
		Replicas:       wfDataService.PDSParams.LoadGen.Replicas,
		TableName:      wfDataService.PDSParams.LoadGen.TableName,
		Iterations:     wfDataService.PDSParams.LoadGen.Iterations,
		FailOnError:    wfDataService.PDSParams.LoadGen.FailOnError,
	}

	chkSum, _, err := dslibs.ReadDataAndReturnChecksum(
		*wfDataService.DataServiceDeployment[deploymentId],
		wfDataService.DataServiceDeployment[deploymentId].DSParams.Name,
		dslibs.CrdMap[strings.ToLower(wfDataService.DataServiceDeployment[deploymentId].DSParams.Name)],
		wkloadParams,
	)

	if err != nil {
		return err
	}

	wfDataService.DataServiceDeployment[deploymentId].SourceMd5Checksum = chkSum

	return nil
}

// TODO: Commenting this methods out, this needs to be refatcored as per current design
//func (wfDataService *WorkflowDataService) ValidateDataServiceWorkloads(params *parameters.NewPDSParams, restoredDeployment *automationModels.PDSRestoreResponse) error {
//	//Initializing the parameters required for workload generation
//	wkloadParams := dslibs.LoadGenParams{
//		LoadGenDepName: params.LoadGen.LoadGenDepName,
//		Namespace:      params.InfraToTest.Namespace,
//		NumOfRows:      params.LoadGen.NumOfRows,
//		Timeout:        params.LoadGen.Timeout,
//		Replicas:       params.LoadGen.Replicas,
//		TableName:      params.LoadGen.TableName,
//		Iterations:     params.LoadGen.Iterations,
//		FailOnError:    params.LoadGen.FailOnError,
//	}
//
//	deployment := make(map[string]string)
//	deployment[*restoredDeployment.Create.Meta.Name] = *restoredDeployment.Create.Meta.Uid
//	// chkSum, wlDep, err := dslibs.ReadDataAndReturnChecksum(deployment, wkloadParams)
//	_, wlDep, err := dslibs.ReadDataAndReturnChecksum(deployment, wkloadParams)
//	if err != nil {
//		return err
//	}
//
//	// deploymentName, _ := GetDeploymentNameAndId(deployment)
//	_, _ = GetDeploymentNameAndId(deployment)
//
//	// TODO: Commenting this out for now, this needs to be refactored as per current design
//	//wfDataService.RestoredDeploymentMd5Hash[deploymentName] = chkSum
//	//
//	//result := dslibs.ValidateDataMd5Hash(wfDataService.SourceDeploymentMd5Hash, wfDataService.RestoredDeploymentMd5Hash)
//	//wfDataService.Dash.VerifyFatal(result, true, "Validate md5 hash after restore")
//
//	return dslibs.DeleteWorkloadDeployments(wlDep)
//}

func (wfDataService *WorkflowDataService) GetDeployment(deploymentId string) (automationModels.V1Deployment, error) {
	dep, err := dslibs.GetDeployment(deploymentId)
	if err != nil {
		return automationModels.V1Deployment{}, err
	}
	return dep.Get, nil
}

func (wfDataService *WorkflowDataService) ValidateDeploymentResources(resourceTemp dslibs.ResourceSettingTemplate, storageOp dslibs.StorageOps, config dslibs.DeploymentConfig, replicas int, dataServiceVersionBuild string) {
	log.Debugf("filesystem used %v ", config.Spec.Topologies[0].StorageOptions.Filesystem)
	log.Debugf("storage replicas used %v ", config.Spec.Topologies[0].StorageOptions.Replicas)
	log.Debugf("cpu requests used %v ", config.Spec.Topologies[0].Resources.Requests.CPU)
	log.Debugf("memory requests used %v ", config.Spec.Topologies[0].Resources.Requests.Memory)
	log.Debugf("storage requests used %v ", config.Spec.Topologies[0].Resources.Requests.Storage)
	log.Debugf("No of nodes requested %v ", config.Spec.Topologies[0].Nodes)
	log.Debugf("volume group %v ", storageOp.VolumeGroup)
	log.Debugf("resource template values cpu req [%s]", resourceTemp.Resources.Requests.CPU)

	wfDataService.Dash.VerifyFatal(config.Spec.Topologies[0].Resources.Requests.CPU, resourceTemp.Resources.Requests.CPU, "Validating CPU Request")
	wfDataService.Dash.VerifyFatal(config.Spec.Topologies[0].Resources.Requests.Memory, resourceTemp.Resources.Requests.Memory, "Validating Memory Request")
	wfDataService.Dash.VerifyFatal(config.Spec.Topologies[0].Resources.Requests.Storage, resourceTemp.Resources.Requests.Storage, "Validating storage Request")
	wfDataService.Dash.VerifyFatal(config.Spec.Topologies[0].Resources.Limits.CPU, resourceTemp.Resources.Limits.CPU, "Validating CPU Limits")
	wfDataService.Dash.VerifyFatal(config.Spec.Topologies[0].Resources.Limits.Memory, resourceTemp.Resources.Limits.Memory, "Validating Memory Limits")
	wfDataService.Dash.VerifyFatal(config.Spec.Topologies[0].StorageOptions.Replicas, storageOp.Replicas, "Validating storage replicas")
	wfDataService.Dash.VerifyFatal(config.Spec.Topologies[0].StorageOptions.Filesystem, storageOp.Filesystem, "Validating filesystems")
	wfDataService.Dash.VerifyFatal(config.Spec.Topologies[0].StorageOptions.Secure, storageOp.Secure, "Validating Secure Storage Option")
	wfDataService.Dash.VerifyFatal(config.Spec.Topologies[0].Nodes, replicas, "Validating ds node replicas")
	wfDataService.Dash.VerifyFatal(config.Spec.Version, dataServiceVersionBuild, "Validating ds version")
}

func (wfDataService *WorkflowDataService) IncreasePvcSizeBy1gb(namespace string, deploymentName string, sizeInGb uint64) error {
	_, err := utils.IncreasePVCby1Gig(namespace, deploymentName, sizeInGb)
	return err
}

func (wfDataService *WorkflowDataService) KillDBMasterNodeToValidateHA(dsName string, deploymentId string) error {
	dbMaster, isNativelyDistributed := utils.GetDbMasterNode(wfDataService.DataServiceDeployment[deploymentId].Namespace, dsName, *wfDataService.DataServiceDeployment[deploymentId].Deployment.Meta.Name, wfDataService.Namespace.TargetCluster.KubeConfig)
	if isNativelyDistributed {
		err := utils.DeleteK8sPods(dbMaster, wfDataService.DataServiceDeployment[deploymentId].Namespace, wfDataService.Namespace.TargetCluster.KubeConfig)
		if err != nil {
			return err
		}
		//validate DataService Deployment here
		newDbMaster, _ := utils.GetDbMasterNode(wfDataService.DataServiceDeployment[deploymentId].Namespace, dsName, *wfDataService.DataServiceDeployment[deploymentId].Deployment.Meta.Name, wfDataService.Namespace.TargetCluster.KubeConfig)
		if dbMaster == newDbMaster {
			log.FailOnError(fmt.Errorf("leader node is not reassigned"), fmt.Sprintf("Leader pod %v", dbMaster))
		}
	} else {
		podName, err := utils.GetAnyPodName(*wfDataService.DataServiceDeployment[deploymentId].Deployment.Meta.Name, wfDataService.DataServiceDeployment[deploymentId].Namespace)
		if err != nil {
			return fmt.Errorf("failed while fetching pod for stateful set %v ", *wfDataService.DataServiceDeployment[deploymentId].Deployment.Meta.Name)
		}
		err = utils.KillPodsInNamespace(wfDataService.DataServiceDeployment[deploymentId].Namespace, podName)
		if err != nil {
			return fmt.Errorf("failed while deleting pod %v ", *wfDataService.DataServiceDeployment[deploymentId].Deployment.Meta.Name)
		}
		//validate DataService Deployment here
	}
	return nil
}

func (wfDataService *WorkflowDataService) DeletePDSPods(podNames []string, namespace string) error {
	pdsPods := make([]corev1.Pod, 0)

	podList, err := utils.GetPods(namespace)
	if err != nil {
		return fmt.Errorf("Error while getting pods: %v", err)
	}
	log.Infof("PDS System Pods")
	for _, pod := range podList.Items {
		for _, name := range podNames {
			if strings.Contains(strings.ToLower(pod.Name), strings.ToLower(name)) {
				log.Infof("%v", pod.Name)
				pdsPods = append(pdsPods, pod)
				break // Once found, break out of the inner loop
			}
		}
	}

	log.InfoD("Deleting PDS System Pods")
	err = utils.DeletePods(pdsPods)
	if err != nil {
		return fmt.Errorf("Error while deleting pods: %\v ", err)
	}
	time.Sleep(90 * time.Second)
	log.InfoD("Validating PDS System Pods")
	err = utils.ValidatePods(namespace, "")
	if err != nil {
		return fmt.Errorf("Error while validating pods: %v", err)
	}
	return nil
}

// GetVolumeCapacityInGBForDeployment Get volume capacity
func (wfDataService *WorkflowDataService) GetVolumeCapacityInGBForDeployment(namespace string, deploymentName string) (uint64, error) {
	capacity, err := utils.GetVolumeCapacityInGB(namespace, deploymentName)
	if err != nil {
		return 0, err
	}
	return capacity, nil
}

func (wfDataService *WorkflowDataService) GetPodAgeForDeployment(deploymentId string) (float64, error) {
	age, err := utils.GetPodAge(
		*wfDataService.DataServiceDeployment[deploymentId].Deployment.Status.CustomResourceName,
		wfDataService.DataServiceDeployment[deploymentId].Namespace,
	)
	if err != nil {
		return 0, err
	}
	return age, nil
}

func (wfDataService *WorkflowDataService) CheckPVCStorageFullCondition(namespace string, deploymentName string, thresholdPercentage float64) error {
	err := utils.CheckStorageFullCondition(namespace, deploymentName, thresholdPercentage)
	if err != nil {
		return err
	}
	return nil
}

func (wfDataService *WorkflowDataService) ValidateDepConfigPostStorageIncrease(deploymentId string, stIncrease *dslibs.ValidateStorageIncrease) error {
	// Get data service deployment resources
	resourceTemp, storageTemp, dbConfig, err := wfDataService.GetDsDeploymentResources(deploymentId)
	if err != nil {
		return err
	}
	wfDataService.Dash.VerifyFatal(resourceTemp.Resources.Requests.CPU, dbConfig.Spec.Topologies[0].Resources.Requests.CPU, "Validating CPU Request")
	wfDataService.Dash.VerifyFatal(resourceTemp.Resources.Requests.Memory, dbConfig.Spec.Topologies[0].Resources.Requests.Memory, "Validating Memory Request")

	log.InfoD("Original resConfigModel.StorageRequest val is- [%v] and Updated resConfigModel.StorageRequest val is- [%v]", resourceTemp.Resources.Requests.Storage, dbConfig.Spec.Topologies[0].Resources.Requests.Storage)
	wfDataService.Dash.VerifyFatal(dbConfig.Spec.Topologies[0].Resources.Requests.Storage, resourceTemp.Resources.Requests.Storage, "Validating the storage size is updated in the config post resize (STS-LEVEL)")

	stringRelFactor := storageTemp.Replicas
	wfDataService.Dash.VerifyFatal(dbConfig.Spec.Topologies[0].StorageOptions.Replicas, stringRelFactor, "Validating the Replication Factor count post storage resize (RepelFactor-LEVEL)")
	podAgeAfterResize, err := utils.GetPodAge(*wfDataService.DataServiceDeployment[deploymentId].Deployment.Meta.Name, wfDataService.DataServiceDeployment[deploymentId].Namespace)
	err = dslibs.VerifyStorageSizeIncreaseAndNoPodRestarts(stIncrease.InitialCapacity, stIncrease.IncreasedStorageSize, stIncrease.BeforeResizePodAge, podAgeAfterResize)
	if err != nil {
		return err
	}
	return err
}

// Purge will delete all dataservice and associated PVCs from the cluster
func (wfDataService *WorkflowDataService) Purge(ignoreError bool) error {

	var errors []string

	errors = make([]string, 0)

	for dsId, dsDetails := range wfDataService.DataServiceDeployment {

		dsName := *wfDataService.DataServiceDeployment[dsId].Deployment.Meta.Name
		log.Infof("Deleting [%s] with id [%s] from [%s]-[%s] namespace ", dsName, dsId, dsDetails.Namespace, dsDetails.NamespaceId)

		deploymentDetails, err := dslibs.GetDeployment(dsId)
		if err != nil {
			log.Warnf("Unable to fetch details for [%s]. Error - [%s]", dsName, err.Error())
			if !ignoreError {
				errors = append(errors, err.Error())
			}
			continue
		}

		err = wfDataService.DeleteDeployment(*deploymentDetails.Get.Meta.Uid)
		if err != nil {
			log.Warnf("Unable to delete [%s]. Error - [%s]", dsName, err.Error())
			if !ignoreError {
				errors = append(errors, err.Error())
			}
			continue
		} else {
			log.Infof("[%s] deleted successfully", dsName)
		}

		err = utils.DeletePvandPVCs(*deploymentDetails.Get.Status.CustomResourceName, false)

		if err != nil {
			log.Warnf("Unable to delete PVs for [%s]. Error - [%s]", dsName, err.Error())
			if !ignoreError {
				errors = append(errors, err.Error())
			}
			continue
		} else {
			log.Infof("All PVs associated with [%s] deleted successfully", dsName)
		}

		err = utils.RemoveFinalizersFromAllResources(dsDetails.Namespace)
		if err != nil {
			log.Warnf("Unable to remove finalizers. Error - [%s]", err.Error())
			if !ignoreError {
				errors = append(errors, err.Error())
			}
		}

	}

	if len(errors) > 0 {
		return fmt.Errorf("Below errors occurred during deployment cleanup.\n\n [%s]", strings.Join(errors, "\n"))
	}

	return nil
}
