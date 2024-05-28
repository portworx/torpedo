package pdslibs

//TODO: This needs to be moved to workflow level

import (
	"encoding/json"
	"fmt"
	pds "github.com/portworx/torpedo/drivers/pds/dataservice"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/platformLibs"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/stworkflows"
	"github.com/portworx/torpedo/pkg/aetosutil"
	"github.com/portworx/torpedo/pkg/log"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"net"
	"strings"
	"time"
)

const (
	CRGroup = "deployments.pds.portworx.com"
	Version = "v1"
)

var Dash *aetosutil.Dashboard

type ValidateStorageIncrease struct {
	UpdatedDeployment      *automationModels.PDSDeploymentResponse
	ResConfigIdUpdated     string
	StorageConfigIdUpdated string
	InitialCapacity        uint64
	IncreasedStorageSize   uint64
	BeforeResizePodAge     float64
}

// GetDeploymentConfigurations returns the deployment CRObject response
func GetDeploymentConfigurations(namespace, crdName, deploymentName string) (DeploymentConfig, error) {
	log.Debugf("namespace [%s]", namespace)
	log.Debugf("CRGroup [%s]", CRGroup)
	log.Debugf("Version [%s]", Version)
	log.Debugf("crdName [%s]", crdName)

	var dbConfig DeploymentConfig

	objects, err := GetCRObject(namespace, CRGroup, Version, crdName)
	if err != nil {
		return dbConfig, err
	}

	log.Debugf("objects [%+v]", objects)

	// Iterate over the CRD objects and print their names.
	for _, object := range objects.Items {
		log.Debugf("Objects created: %v", object.GetName())
		if object.GetName() == deploymentName {
			crJsonObject, err := object.MarshalJSON()
			if err != nil {
				return dbConfig, err
			}
			err = json.Unmarshal(crJsonObject, &dbConfig)
			if err != nil {
				return dbConfig, err
			}
		}
	}
	log.Debugf("depVersion [%v]", dbConfig.Spec.Version)
	return dbConfig, nil
}

// GetResourceTemplateConfigs returns the resourceTemplate configs
func GetResourceTemplateConfigs(resourceTemplateID string) (ResourceSettingTemplate, error) {
	var resourceTemp ResourceSettingTemplate
	ResourceTemplateresp, err := platformLibs.GetTemplate(resourceTemplateID)
	if err != nil {
		return resourceTemp, err
	}

	log.Debug("ResourceTemplate Response")
	for key, value := range ResourceTemplateresp.Get.Config.TemplateValues {
		log.Debugf("key [%s]", key)
		log.Debugf("value [%s]", value)
		if key == "cpu_request" {
			resourceTemp.Resources.Requests.CPU = value.(string)
		}
		if key == "memory_request" {
			resourceTemp.Resources.Requests.Memory = value.(string)
		}
		if key == "storage_request" {
			resourceTemp.Resources.Requests.Storage = value.(string)
		}
		if key == "cpu_limit" {
			resourceTemp.Resources.Limits.CPU = value.(string)
		}
		if key == "memory_limit" {
			resourceTemp.Resources.Limits.Memory = value.(string)
		}
	}
	return resourceTemp, nil
}

// GetStorageTemplateConfigs returns the storageTemplate configs
func GetStorageTemplateConfigs(storageTemplateID string) (StorageOps, error) {
	var storageOp StorageOps

	stResponse, err := platformLibs.GetTemplate(storageTemplateID)
	if err != nil {
		return storageOp, err
	}
	log.Debug("StorageTemplate Response")
	for key, value := range stResponse.Get.Config.TemplateValues {
		log.Debugf("key [%s]", key)
		log.Debugf("value [%s]", value)
		if key == "fs" {
			storageOp.Filesystem = value.(string)
		}
		if key == "provisioner" {
			storageOp.Provisioner = value.(string)
		}
		if key == "repl" {
			storageOp.Replicas = value.(string)
		}
		if key == "fg" {
			storageOp.VolumeGroup = value.(string)
		}
		if key == "secure" {
			storageOp.Secure = value.(string)
		}
	}
	return storageOp, nil
}

// ValidateDeploymentConfigUpdate take deploymentConfigUpdateId and validates the status of the update
func ValidateDeploymentConfigUpdate(deploymentConfigUpdateId, expectedPhase string) error {
	log.Infof("DeploymentConfigUpdateId [%s]", deploymentConfigUpdateId)
	waitErr := wait.PollImmediate(maxtimeInterval, validateDeploymentTimeOut, func() (bool, error) {
		deploymentConfig, err := GetDeploymentConfig(deploymentConfigUpdateId)
		if err != nil {
			log.Errorf("Error occured while getting deployment status %v", err)
			return false, err
		}
		log.Debugf("Deployment Config Update phase -  %v", *deploymentConfig.Update.Status.Phase)
		if string(*deploymentConfig.Update.Status.Phase) == expectedPhase {
			log.Infof("Deployment ConfigUpdate status: phase - %v retry-count - %v", *deploymentConfig.Update.Status.Phase, *deploymentConfig.Update.Status.RetryCount)
			if ResiFlag {
				ResiliencyCondition <- true
				log.InfoD("Resiliency Condition Met")
			}
			return true, nil
		}
		log.Infof("Condition still not met. Will retry to see if it has met now.....")
		return false, err
	})

	return waitErr
}

// ValidateStatefulSetHealth validates the health of the statefulset pod
func ValidateStatefulSetHealth(statefulsetName, namespace string) error {
	log.Debugf("deployment name [%s] in namespace [%s]", statefulsetName, namespace)
	var ss *v1.StatefulSet
	err = wait.Poll(validateDeploymentTimeInterval, validateDeploymentTimeOut, func() (bool, error) {
		ss, err = k8sApps.GetStatefulSet(statefulsetName, namespace)
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

	return nil
}

// ValidateDataServiceDeploymentHealth takes the deployment map(name and id), namespace and returns error
func ValidateDataServiceDeploymentHealth(deploymentId string, expectedHealth automationModels.V1StatusHealth) error {
	log.Infof("DeploymentId [%s]", deploymentId)
	conditionError := wait.Poll(maxtimeInterval, validateDeploymentTimeOut, func() (bool, error) {
		res, err := v2Components.PDS.GetDeployment(deploymentId)
		if err != nil {
			log.Errorf("Error occured while getting deployment status %v", err)
			return false, nil
		}
		if *res.Get.Status.Phase == stworkflows.FAILED {
			log.Infof("Deployment details: Health status -  %v, Replicas - %v, Ready replicas - %v", *res.Get.Status.Health, *res.Get.Config.DeploymentTopologies[0].Replicas, *res.Get.Status.DeploymentTopologyStatus[0].ReadyReplicas)
			return true, fmt.Errorf("Deployment [%s] is [%s]", *res.Get.Meta.Name, *res.Get.Status.Phase)
		}
		log.Debugf("Health status - [%v]", *res.Get.Status.Health)
		if *res.Get.Status.Health == expectedHealth {
			log.Infof("Deployment details: Health status -  %v, Replicas - %v, Ready replicas - %v", *res.Get.Status.Health, *res.Get.Config.DeploymentTopologies[0].Replicas, *res.Get.Status.DeploymentTopologyStatus[0].ReadyReplicas)
			if ResiFlag {
				ResiliencyCondition <- true
				log.InfoD("Resiliency Condition Met")
			}
			return true, nil
		}
		log.Infof("Condition still not met. Will retry to see if it has met now.....")
		return false, nil
	})
	if conditionError != nil {
		if ResiliencyFlag {
			ResiliencyCondition <- false
			CapturedErrors <- conditionError
		}
	}
	return conditionError
}

// ValidateDeploymentIsDeleted checks if deployment is deleted
func ValidateDeploymentIsDeleted(deploymentId string) error {
	log.Infof("DeploymentId [%s]", deploymentId)
	err = wait.Poll(maxtimeInterval, validateDeploymentTimeOut, func() (bool, error) {
		res, err := v2Components.PDS.GetDeployment(deploymentId)
		if err != nil && strings.Contains(err.Error(), "resource not found") {
			log.Infof("Error occured while getting deployment status %v", err)
			return true, nil
		}
		log.Debugf("Health status -  %v", *res.Get.Status.Health)
		if *res.Get.Config.DeploymentTopologies[0].Replicas != *res.Get.Status.DeploymentTopologyStatus[0].ReadyReplicas || *res.Get.Status.Health != PDS_DEPLOYMENT_AVAILABLE {
			return false, nil
		}
		log.Infof("Deployment details: Health status -  %v, Replicas - %v, Ready replicas - %v", *res.Get.Status.Health, *res.Get.Config.DeploymentTopologies[0].Replicas, *res.Get.Status.DeploymentTopologyStatus[0].ReadyReplicas)
		return false, nil
	})

	return err
}

// ValidateDataMd5Hash validates the hash of the data service deployments
func ValidateDataMd5Hash(deploymentHash, restoredDepHash map[string]string) bool {
	count := 0

	//Debug block to print hash of the database table
	for depName, hash := range deploymentHash {
		log.Debugf("Dep name %s and hash %s", depName, hash)
	}
	for depName, hash := range restoredDepHash {
		log.Debugf("Restored Dep name %s and hash %s", depName, hash)
	}

	for key, depHash := range deploymentHash {
		depName, _, _ := strings.Cut(key, "-")
		for key1, resDepHash := range restoredDepHash {
			resDepName, _, _ := strings.Cut(key1, "-")
			if depName == resDepName && depHash == resDepHash {
				log.InfoD("data is consistent for restored deployment %s", key1)
				count += 1
			}
		}
	}
	if count < len(restoredDepHash) {
		return false
	}
	return true
}

// InsertDataAndReturnChecksum Inserts Data into the db and returns the checksum
func InsertDataAndReturnChecksum(dataServiceDetails DataServiceDetails, wkloadGenParams LoadGenParams) (string, *v1.Deployment, error) {
	wkloadGenParams.Mode = "write"
	crdName := CrdMap[strings.ToLower(dataServiceDetails.DSParams.Name)]
	dataServiceName := strings.ToLower(dataServiceDetails.DSParams.Name)
	deploymentName := *dataServiceDetails.Deployment.Status.CustomResourceName

	_, dep, err := GenerateWorkload(deploymentName, dataServiceName, crdName, wkloadGenParams)
	if err == nil {
		err := k8sApps.DeleteDeployment(dep.Name, dep.Namespace)
		if err != nil {
			return "", nil, fmt.Errorf("error while deleting the workload deployment")
		}
	} else {
		return "", nil, err
	}
	ckSum, wlDep, err := ReadDataAndReturnChecksum(dataServiceDetails, dataServiceName, crdName, wkloadGenParams)
	return ckSum, wlDep, err
}

// ReadDataAndReturnChecksum Reads Data from the db and returns the checksum
func ReadDataAndReturnChecksum(dataServiceDetails DataServiceDetails, dataServiceName, crdName string, wkloadGenParams LoadGenParams) (string, *v1.Deployment, error) {
	wkloadGenParams.Mode = "read"

	deploymentName := *dataServiceDetails.Deployment.Status.CustomResourceName
	ckSum, wlDep, err := GenerateWorkload(deploymentName, dataServiceName, crdName, wkloadGenParams)
	if err != nil {
		return "", nil, fmt.Errorf("error while reading the workload deployment data")
	}
	return ckSum, wlDep, err
}

// GenerateWorkload creates a deployment using the given params(perform read/write) and returns the checksum
func GenerateWorkload(deploymentName, dataServiceName, crdName string, wkloadGenParams LoadGenParams) (string, *v1.Deployment, error) {
	var checksum string
	workloadDepName := wkloadGenParams.LoadGenDepName
	namespace := wkloadGenParams.Namespace
	failOnError := wkloadGenParams.FailOnError
	mode := wkloadGenParams.Mode
	seed := wkloadGenParams.TableName
	counts := wkloadGenParams.NumOfRows
	iterations := wkloadGenParams.Iterations
	timeout := wkloadGenParams.Timeout
	replicas := wkloadGenParams.Replicas
	replacePassword := wkloadGenParams.ReplacePassword
	clusterMode := wkloadGenParams.ClusterMode

	log.Debugf("DeploymentName [%s]", deploymentName)
	log.Debugf("dataServiceName [%s]", dataServiceName)
	log.Debugf("CrdName [%s]", crdName)

	serviceAccount, err := pds.CreatePolicies(namespace, crdName)
	if err != nil {
		return "", nil, fmt.Errorf("Error while creating policies %v\n", err)
	}

	deploymentSpec := &v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: workloadDepName + "-",
			Namespace:    namespace,
		},
		Spec: v1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": workloadDepName},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": workloadDepName},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            "main",
							Image:           pdsWorkloadImage,
							ImagePullPolicy: "Always",
							Env: []corev1.EnvVar{
								{Name: "PDS_DEPLOYMENT", Value: deploymentName},
								{Name: "NAMESPACE", Value: namespace},
								{Name: "DATASERVICE", Value: dataServiceName},
								{Name: "FAIL_ON_ERROR", Value: failOnError},
								{Name: "MODE", Value: mode},
								{Name: "SEED", Value: seed},
								{Name: "COUNTS", Value: counts},
								{Name: "ITERATIONS", Value: iterations},
								{Name: "TIMEOUT", Value: timeout},
								{Name: "REPLACE_PASSWORD", Value: replacePassword},
								{Name: "CLUSTER_MODE", Value: clusterMode},
							},
							SecurityContext: &corev1.SecurityContext{
								AllowPrivilegeEscalation: boolPtr(false),
								RunAsNonRoot:             boolPtr(true),
								RunAsUser:                int64Ptr(1000),
								Capabilities: &corev1.Capabilities{
									Drop: []corev1.Capability{
										"ALL",
									},
								},
							},
						},
					},
					ServiceAccountName: serviceAccount.Name,
				},
			},
		},
	}
	log.Debugf("Deployment spec %+v", deploymentSpec)
	wlDeployment, err := k8sApps.CreateDeployment(deploymentSpec, metav1.CreateOptions{})
	if err != nil {
		return "", nil, fmt.Errorf("error Occured while creating deployment %v", err)
	}
	err = k8sApps.ValidateDeployment(wlDeployment, timeOut, 10*time.Second)
	if err != nil {
		return "", nil, fmt.Errorf("error Occured while validating the pod %v", err)
	}
	podList, err := k8sCore.GetPods(wlDeployment.Namespace, nil)
	if err != nil {
		return "", nil, fmt.Errorf("error Occured while getting the podlist %v", err)
	}
	for _, pod := range podList.Items {
		if strings.Contains(pod.Name, wlDeployment.Name) {
			log.Debugf("workload pod name %s", pod.Name)
			checksum, err = ReadChecksum(pod.Name, wlDeployment.Namespace, mode)
			if err != nil {
				return "", nil, fmt.Errorf("error Occured while fetching checksum %v", err)
			}
		}
	}
	return checksum, wlDeployment, nil
}

func ReadChecksum(podName, namespace, mode string) (string, error) {
	var checksum string

	log.InfoD("%s operation started...", mode)
	err = wait.Poll(maxtimeInterval, timeOut, func() (bool, error) {
		logs, err := k8sCore.GetPodLog(podName, namespace, &corev1.PodLogOptions{})
		if err != nil {
			return false, fmt.Errorf("error while fetching the pod logs: %v", err)
		}
		log.Infof("%s operation is in progress...", mode)
		if strings.Contains(logs, "Checksum") {
			for _, line := range strings.Split(strings.TrimRight(logs, "\n"), "\n") {
				if strings.Contains(line, "Checksum") {
					words := strings.Split(line, ":")
					checksum = words[1]
					return true, nil
				}
			}
		}
		return false, nil
	})
	log.InfoD("%s operation completed...", mode)
	log.InfoD("Checksum of the table is %s", checksum)
	return checksum, err
}

func DeleteWorkloadDeployments(wlDep *v1.Deployment) error {
	err = k8sApps.DeleteDeployment(wlDep.Name, wlDep.Namespace)
	return err
}

// GetDataServiceImageId returns the pds dsImageId for the given ds version and image build
func GetDataServiceImageId(dsName, dsImageTag, dsVersionBuild string) (string, error) {
	dsId, err := GetDataServiceId(dsName)
	if err != nil {
		return "", err
	}
	log.Debugf("dataserviceId [%s]", dsId)

	versionResps, err := ListDataServiceVersions(dsId)
	if err != nil {
		return "", err
	}

	var dsVersionId string
	for _, versionResp := range versionResps.DataServiceVersionList {
		if *versionResp.Meta.Name == dsVersionBuild {
			dsVersionId = *versionResp.Meta.Uid
			break
		}
	}
	log.Debugf("dsVersionId [%s]", dsVersionId)

	imgResps, err := ListDataServiceImages(dsId, dsVersionId)
	if err != nil {
		return "", err
	}

	dsImageId := ""
	for _, imgResp := range imgResps.DataServiceImageList {
		log.Debugf("imgResp.Info.Build [%v]", *imgResp.Info.Build)
		if *imgResp.Info.Build == dsImageTag {
			dsImageId = *imgResp.Meta.Uid
			break
		}
	}

	log.Debugf("dsImageId [%s]", dsImageId)

	if dsImageId == "" {
		return "", fmt.Errorf("image %s not found for data service %s version %s", dsImageTag, dsName, dsVersionBuild)
	}

	return dsImageId, nil
}

func ValidateDNSEndPoint(dnsEndPoint string) error {
	//log.Debugf("sleeping for 5 min, before validating dns endpoint")
	//time.Sleep(5 * time.Minute)
	conn, err := net.Dial("tcp", dnsEndPoint)
	if err != nil {
		return fmt.Errorf("Failed to connect to the dns endpoint with err: %v", err)
	} else {
		log.Infof("DNS endpoint is reachable and ready to accept connections")
	}

	defer conn.Close()

	return nil
}

func VerifyStorageSizeIncreaseAndNoPodRestarts(initialCapacity uint64, newCapacity uint64, beforeAge float64, afterResizePodAge float64) error {
	if newCapacity > initialCapacity {
		flag := true
		Dash.VerifyFatal(flag, true, "Validating the storage size is updated in the config post resize (PV/PVC-LEVEL)")
		log.InfoD("Initial PVC Capacity is- [%v] and Updated PVC Capacity is- [%v]", initialCapacity, newCapacity)
	} else {
		log.FailOnError(err, "Failed to verify Storage Resize at PV/PVC level")
	}
	log.FailOnError(err, "unable to get pods restart count before PVC resize")
	log.InfoD("Pods Age after storage resize is- [%v]Min", afterResizePodAge)
	if beforeAge < afterResizePodAge {
		flagCount := true
		Dash.VerifyFatal(flagCount, true, "Validating NO pod restarts occurred while storage resize")

	} else {
		log.FailOnError(err, "Pods restarted after storage resize, Please check the logs manually")
	}
	log.InfoD("Successfully validated that NO pod restarted while/after storage resize")
	return nil
}
