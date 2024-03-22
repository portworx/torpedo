package pdslibs

import (
	"encoding/json"
	"fmt"
	pds "github.com/portworx/torpedo/drivers/pds/dataservice"
	"github.com/portworx/torpedo/pkg/log"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"strconv"
	"strings"
	"time"
)

func GetDeploymentResources(deployment map[string]string, dataService, dataServiceDefaultResourceTemplateID, storageTemplateID, namespace string) (ResourceSettingTemplate, StorageOptions, StorageClassConfig, error) {
	var (
		config       StorageClassConfig
		resourceTemp ResourceSettingTemplate
		storageOp    StorageOptions
		dbConfig     DBConfig
		docImage     string
	)

	deploymentName, deploymentId := GetDeploymentNameAndId(deployment)

	labelSelector := make(map[string]string)
	labelSelector["name"] = deploymentName
	storageClasses, err := k8sStorage.GetStorageClasses(labelSelector)
	if err != nil {
		log.FailOnError(err, "An error occured while getting storage classes")
	}

	objects, err := GetCRObject(namespace, "deployments.pds.io", "v1", "databases")

	// Iterate over the CRD objects and print their names.
	for _, object := range objects.Items {
		log.Debugf("Objects created: %v", object.GetName())
		if object.GetName() == deploymentName {
			crJsonObject, err := object.MarshalJSON()
			if err != nil {
				log.FailOnError(err, "An error occured while marshalling cr")
			}
			err = json.Unmarshal(crJsonObject, &dbConfig)
			if err != nil {
				log.FailOnError(err, "An error occured while unmarshalling cr")
			}
		}
	}

	//Get the ds version from the sts
	if dataService == mssql {
		docImage = dbConfig.Spec.StatefulSet.Template.Spec.Containers[1].Image
	} else {
		docImage = dbConfig.Spec.StatefulSet.Template.Spec.Containers[0].Image
	}
	log.Debugf("docImage [%v]", docImage)
	dsVersionImageTag := strings.Split(docImage, ":")
	log.Debugf("version tag %v", dsVersionImageTag[1])

	scJsonData, err := json.Marshal(storageClasses)
	if err != nil {
		log.FailOnError(err, "An error occured while marshalling statefulset")
	}
	err = json.Unmarshal(scJsonData, &config)
	if err != nil {
		log.FailOnError(err, "An error occured while unmarshalling storage class")
	}

	//Assigning values to the custom struct of storageclass config
	config.Resources.Requests.CPU = dbConfig.Spec.StatefulSet.Template.Spec.Containers[0].Resources.Requests.CPU
	config.Resources.Requests.Memory = dbConfig.Spec.StatefulSet.Template.Spec.Containers[0].Resources.Requests.Memory
	config.Resources.Requests.EphemeralStorage = dbConfig.Spec.Datastorage.PersistentVolumeSpec.Spec.Resources.Requests.Storage
	config.Resources.Limits.CPU = dbConfig.Spec.StatefulSet.Template.Spec.Containers[0].Resources.Limits.CPU
	config.Resources.Limits.Memory = dbConfig.Spec.StatefulSet.Template.Spec.Containers[0].Resources.Limits.Memory
	config.Replicas = dbConfig.Status.Replicas
	config.Version = dsVersionImageTag[1]

	config.Parameters.Fg = dbConfig.Spec.Datastorage.StorageClass.Parameters.Fg
	config.Parameters.Fs = dbConfig.Spec.Datastorage.StorageClass.Parameters.Fs
	config.Parameters.Repl = dbConfig.Spec.Datastorage.StorageClass.Parameters.Repl

	//TODO: Update the template details once the template api's are ready
	log.Infof("deployment Id [%s]", deploymentId)
	//rt, err := components.ResourceSettingsTemplate.GetTemplate(dataServiceDefaultResourceTemplateID)
	//if err != nil {
	//	log.Errorf("Error Occured while getting resource setting template %v", err)
	//}
	//resourceTemp.Resources.Requests.CPU = *rt.CpuRequest
	//resourceTemp.Resources.Requests.Memory = *rt.MemoryRequest
	//resourceTemp.Resources.Requests.Storage = *rt.StorageRequest
	//resourceTemp.Resources.Limits.CPU = *rt.CpuLimit
	//resourceTemp.Resources.Limits.Memory = *rt.MemoryLimit
	//
	//st, err := components.StorageSettingsTemplate.GetTemplate(storageTemplateID)
	//if err != nil {
	//	log.Errorf("Error Occured while getting storage template %v", err)
	//	return resourceTemp, storageOp, config, err
	//}
	//storageOp.Filesystem = st.GetFs()
	//storageOp.Replicas = st.GetRepl()
	//storageOp.VolumeGroup = st.GetFg()

	return resourceTemp, storageOp, config, nil

}

func ValidateDataServiceDeployment(deployment map[string]string, namespace string) error {
	var (
		ss *v1.StatefulSet
	)

	log.Debugf("deployment name [%s] in namespace [%s]", deployment[""], namespace)

	deploymentName, deploymentId := GetDeploymentNameAndId(deployment)

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

func ValidateDeploymentResources(resourceTemp ResourceSettingTemplate, storageOp StorageOptions, config StorageClassConfig, replicas int, dataServiceVersionBuildMap map[string][]string) {
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

// InsertDataAndReturnChecksum Inserts Data into the db and returns the checksum
func InsertDataAndReturnChecksum(deployment map[string]string, wkloadGenParams LoadGenParams) (string, *v1.Deployment, error) {
	wkloadGenParams.Mode = "write"

	deploymentName, _ := GetDeploymentNameAndId(deployment)

	_, dep, err := GenerateWorkload(deploymentName, wkloadGenParams)
	if err == nil {
		err := k8sApps.DeleteDeployment(dep.Name, dep.Namespace)
		if err != nil {
			return "", nil, fmt.Errorf("error while deleting the workload deployment")
		}
	}
	ckSum, wlDep, err := ReadDataAndReturnChecksum(deploymentName, wkloadGenParams)
	return ckSum, wlDep, err
}

// ReadDataAndReturnChecksum Reads Data from the db and returns the checksum
func ReadDataAndReturnChecksum(deploymentName string, wkloadGenParams LoadGenParams) (string, *v1.Deployment, error) {
	wkloadGenParams.Mode = "read"
	ckSum, wlDep, err := GenerateWorkload(deploymentName, wkloadGenParams)
	return ckSum, wlDep, err
}

// GenerateWorkload creates a deployment using the given params(perform read/write) and returns the checksum
func GenerateWorkload(deploymentName string, wkloadGenParams LoadGenParams) (string, *v1.Deployment, error) {
	var checksum string
	dsName := deploymentName
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

	serviceAccount, err := pds.CreatePolicies(namespace)
	if err != nil {
		return "", nil, fmt.Errorf("error while creating policies")
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
								{Name: "PDS_DEPLOYMENT", Value: dsName},
								{Name: "NAMESPACE", Value: namespace},
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
