package lib

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	"github.com/portworx/sched-ops/k8s/apps"
	"github.com/portworx/sched-ops/k8s/core"
	pdsapi "github.com/portworx/torpedo/drivers/pds/api"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

//PDS const
const (
	storageTemplateName   = "Volume replication (best-effort spread)"
	resourceTemplateName  = "Small"
	appConfigTemplateName = "Default"
	deploymentName        = "automation"
	defaultRetryInterval  = 10 * time.Minute
	duration              = 900
	timeOut               = 5 * time.Minute
	timeInterval          = 10 * time.Second
)

//PDS vars
var (
	k8sCore = core.Instance()
	k8sApps = apps.Instance()

	components                            *pdsapi.Components
	deploymentTargetID, storageTemplateID string
	accountID                             string
	tenantID                              string
	projectID                             string
	serviceType                           = "LoadBalancer"
	accountName                           = "Portworx"
	deployment                            *pds.ModelsDeployment
	err                                   error
	isavailable                           bool

	dataServiceDefaultResourceTemplateIDMap = make(map[string]string)
	dataServiceNameIDMap                    = make(map[string]string)
	dataServiceNameVersionMap               = make(map[string][]string)
	dataServiceIDImagesMap                  = make(map[string][]string)
	dataServiceNameDefaultAppConfigMap      = make(map[string]string)
	deployementIDNameMap                    = make(map[string]string)
	namespaceNameIDMap                      = make(map[string]string)
)

//ExecShell to execute local command
func ExecShell(command string) (string, string, error) {
	return ExecShellWithEnv(command)
}

// ExecShellWithEnv to execute local command
func ExecShellWithEnv(command string, envVars ...string) (string, string, error) {
	var stout, sterr []byte
	cmd := exec.Command("bash", "-c", command)
	logrus.Debugf("Command %s ", command)
	cmd.Env = append(cmd.Env, envVars...)
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	if err := cmd.Start(); err != nil {
		logrus.Debugf("Command %s failed to start. Cause: %v", command, err)
		return "", "", err
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		stout, _ = copyAndCapture(os.Stdout, stdout)
		wg.Done()
	}()

	sterr, _ = copyAndCapture(os.Stderr, stderr)

	wg.Wait()

	err := cmd.Wait()
	return string(stout), string(sterr), err
}

// copyAndCapture
func copyAndCapture(w io.Writer, r io.Reader) ([]byte, error) {
	var out []byte
	buf := make([]byte, 1024)
	for {
		n, err := r.Read(buf[:])
		if n > 0 {
			d := buf[:n]
			out = append(out, d...)
			_, err := w.Write(d)
			if err != nil {
				return out, err
			}
		}
		if err != nil {
			// Read returns io.EOF at the end of file, which is not an error for us
			if err == io.EOF {
				err = nil
			}
			return out, err
		}
	}
}

// GetDataServices lists data services
func GetDataServices(url string) {
	endPointURL := url + "/api/data-services"
	logrus.Infof("endPointURL %v", endPointURL)
	accessToken := "eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsImlzcyI6Imh0dHBzOi8vYXBpY2VudHJhbC5wb3J0d29yeC5jb20vYXBpIn0.eyJhdWQiOiI2IiwianRpIjoiN2MwMDVkMGQwZmYwMDNiYzk1NGQ2YjVlODgyZmY2YzZlZDdiZjcwMjdjN2Q2ZWQxNTg1MTliNWY3MmRjM2EyNDg3MTJkZjc4ZDNlZjUwMjQiLCJpYXQiOjE2NTk2NzQ1ODguODY4Njc1LCJuYmYiOjE2NTk2NzQyODguODY4Njc1LCJleHAiOjE2NTk2NzYzODguODI2MTg5LCJzdWIiOiIyMTk5NiIsImlzcyI6Imh0dHBzOi8vYXBpY2VudHJhbC5wb3J0d29yeC5jb20vYXBpIiwibmFtZSI6Ik1hZGFuYWdvcGFsIEFydW5hY2hhbGFtIiwiZW1haWwiOiJtYXJ1bmFjaGFsYW1AcHVyZXN0b3JhZ2UuY29tIiwic2NvcGVzIjpbXX0.oKrTnVr871h4WmKnmrKXBn4jEp-ndaxiz16v4manzpxuFoZiUJC8b5GZFL4jSgke6Jis0baAhLi19sE86waQmA5jl67zWDEL4LfPXyxWplIwmb-voWqhmroxOYqIegjaiWq_Okrf1rNUyavIdnERfOrMEiWh-FfkHEJ6UBwZ_3ZCBNL42dD0wyf5634lD_a6oR3UxaPHuoG_ejW79sDPSpxn5sz9Q5QaSG2MfnACeKmfWUsFfifIixGzj8DQSxb2igRtQkgKD34kuyBIkRXmu0I-7105wx3rstTswTBrOjACUHA3kpgYtoFnrVx6hIuYqzGHDtfOFDWCV7nH0kiJlfzk05Y-uyQodDNYBTTbIjcDhlboS5dhsQw5g0gilrDjaHH6grmZC3O4q8dwIdq5F9KsU5ZVJRqf0f3bAeARNZsOHxQixGxmPMLOoIHkla_fHch7ftOaxZ5YLvkj4JIis8_JQ2rkjlJqRsFC4Ma-j9ZVUlF5VRJ5UAKCLe_EJF5X9SjtgO9cQTIJNQ-CHS7xYTfpBnSLqbOWajI3MsUMeEhfCcvaZVZ4e1c7mtTuJnhI5h20S1qZx_t-d3lZ7YXrj0iR4-SO1r5CcJ5qkuXVCZryWNpp6RPf67b9GxVpBzh6iAsiBO8EBbv2TgYgpnmxlA0Lt7SiPz8XTb4G2BeSsZM"
	req, err := http.NewRequest("GET", endPointURL, nil)
	if err != nil {
		logrus.Fatalf("An Error Occured %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logrus.Fatalf("An Error Occured %v", err)
	}
	defer resp.Body.Close()
	logrus.Infof("status %v", resp.Status)
}

// GetClusterID retruns the cluster id
func GetClusterID(pathToKubeconfig string) (string, error) {
	logrus.Infof("Fetch Cluster id ")
	cmd := fmt.Sprintf("kubectl get ns kube-system -o jsonpath={.metadata.uid} --kubeconfig %s", pathToKubeconfig)
	// cmdArgs := []string{"get ns kube-system -o jsonpath={.metadata.uid"}
	// err := osutils.Kubectl(cmdArgs)
	output, _, err := ExecShell(cmd)
	if err != nil {
		logrus.Fatalf("An Error Occured %v", err)
	}
	return output, nil
}

// SetupPDSTest returns few params required to run the test
func SetupPDSTest() (string, string, string, string, string) {
	var err error
	apiConf := pds.NewConfiguration()
	endpointURL, err := url.Parse(GetAndExpectStringEnvVar(envControlPlaneURL))
	if err != nil {
		logrus.Fatalf("An Error Occured %v", err)
	}
	apiConf.Host = endpointURL.Host
	apiConf.Scheme = endpointURL.Scheme

	//ctx := context.WithValue(context.Background(), pds.ContextAPIKeys, map[string]pds.APIKey{"ApiKeyAuth": {Key: GetAndExpectStringEnvVar("BEARER_TOKEN"), Prefix: "Bearer"}})
	ctx := context.WithValue(context.Background(), pds.ContextAPIKeys, map[string]pds.APIKey{"ApiKeyAuth": {Key: GetBearerToken(), Prefix: "Bearer"}})
	apiClient := pds.NewAPIClient(apiConf)
	components = pdsapi.NewComponents(ctx, apiClient)
	controlplane := NewControlPlane(GetAndExpectStringEnvVar(envControlPlaneURL), components)

	//Test methods - Remove me
	//GetDataServices(GetAndExpectStringEnvVar(envControlPlaneURL))

	clusterID, err := GetClusterID(GetAndExpectStringEnvVar(envTargetKubeconfig))
	logrus.Infof("clusterID %v", clusterID)
	if err != nil {
		logrus.Fatalf("An Error Occured %v", err)
	}

	if strings.EqualFold(GetAndExpectStringEnvVar(envClusterType), "onprem") || strings.EqualFold(GetAndExpectStringEnvVar(envClusterType), "ocp") {
		serviceType = "ClusterIP"
	}
	logrus.Infof("Deployment service type %s", serviceType)

	acc := components.Account
	accounts, err := acc.GetAccountsList()
	if err != nil {
		logrus.Fatalf("An Error Occured %v", err)
	}
	logrus.Infof("length of account %v", len(accounts))

	for i := 0; i < len(accounts); i++ {
		logrus.Infof("Account Name: %v", accounts[i].GetName())
		if accounts[i].GetName() == accountName {
			accountID = accounts[i].GetId()
		}
	}
	logrus.Infof("Account Detail- Name: %s, UUID: %s ", accountName, accountID)
	tnts := components.Tenant
	tenants, _ := tnts.GetTenantsList(accountID)
	tenantID = tenants[0].GetId()
	tenantName := tenants[0].GetName()
	logrus.Infof("Tenant Details- Name: %s, UUID: %s ", tenantName, tenantID)
	dnsZone := controlplane.GetDNSZone(tenantID)
	logrus.Infof("DNSZone info - Name: %s, tenant: %s , account: %s", dnsZone, tenantName, accountName)
	projcts := components.Project
	projects, _ := projcts.GetprojectsList(tenantID)
	projectID = projects[0].GetId()
	projectName := projects[0].GetName()
	logrus.Infof("Project Details- Name: %s, UUID: %s ", projectName, projectID)

	logrus.Info("Get the Target cluster details")
	targetClusters, _ := components.DeploymentTarget.ListDeploymentTargetsBelongsToTenant(tenantID)
	for i := 0; i < len(targetClusters); i++ {
		if targetClusters[i].GetClusterId() == clusterID {
			deploymentTargetID = targetClusters[i].GetId()
			logrus.Infof("Cluster ID: %v, Name: %v,Status: %v", targetClusters[i].GetClusterId(), targetClusters[i].GetName(), targetClusters[i].GetStatus())
		}
	}

	return tenantID, dnsZone, projectID, serviceType, deploymentTargetID
}

//GetStorageTemplate gets the template id
func GetStorageTemplate(tenantID string) string {
	logrus.Infof("Get the storage template")
	storageTemplates, _ := components.StorageSettingsTemplate.ListTemplates(tenantID)
	for i := 0; i < len(storageTemplates); i++ {
		if storageTemplates[i].GetName() == storageTemplateName {
			logrus.Infof("Storage template details -----> Name %v,Repl %v , Fg %v , Fs %v",
				storageTemplates[i].GetName(),
				storageTemplates[i].GetRepl(),
				storageTemplates[i].GetFg(),
				storageTemplates[i].GetFs())
			storageTemplateID = storageTemplates[i].GetId()
			logrus.Infof("Storage Id: %v", storageTemplateID)
		}
	}
	return storageTemplateID
}

// GetResourceTemplate get the resource template id and forms supported dataserviceNameIdMap
func GetResourceTemplate(tenantID string, supportedDataServices []string) (map[string]string, map[string]string) {
	logrus.Infof("Get the resource template for each data services")
	resourceTemplates, _ := components.ResourceSettingsTemplate.ListTemplates(tenantID)
	for i := 0; i < len(resourceTemplates); i++ {
		if resourceTemplates[i].GetName() == resourceTemplateName {
			dataService, _ := components.DataService.GetDataService(resourceTemplates[i].GetDataServiceId())
			for dataKey := range supportedDataServices {
				if dataService.GetName() == supportedDataServices[dataKey] {
					logrus.Infof("Data service name: %v", dataService.GetName())
					logrus.Infof("Resource template details ---> Name %v, Id : %v ,DataServiceId %v , StorageReq %v , Memoryrequest %v",
						resourceTemplates[i].GetName(),
						resourceTemplates[i].GetId(),
						resourceTemplates[i].GetDataServiceId(),
						resourceTemplates[i].GetStorageRequest(),
						resourceTemplates[i].GetMemoryRequest())

					dataServiceDefaultResourceTemplateIDMap[dataService.GetName()] =
						resourceTemplates[i].GetId()
					dataServiceNameIDMap[dataService.GetName()] = dataService.GetId()
				}
			}
		}
	}
	return dataServiceDefaultResourceTemplateIDMap, dataServiceNameIDMap
}

// GetVersions returns the required versions and build ID of dataservices
func GetVersions(dsVersion string, dsBuild string, dataServiceNameIDMap map[string]string) (map[string][]string, map[string][]string) {
	var versions []pds.ModelsVersion
	var images []pds.ModelsImage
	for key := range dataServiceNameIDMap {
		versions, _ = components.Version.ListDataServiceVersions(dataServiceNameIDMap[key])
		for i := 0; i < len(versions); i++ {
			if (*versions[i].Enabled) && (*versions[i].Name == dsVersion) {
				dataServiceNameVersionMap[key] = append(dataServiceNameVersionMap[key], versions[i].GetId())
				images, _ = components.Image.ListImages(versions[i].GetId())
				for j := 0; j < len(images); j++ {
					if *images[j].Build == dsBuild {
						dataServiceIDImagesMap[versions[i].GetId()] = append(dataServiceIDImagesMap[versions[i].GetId()], images[j].GetId())
						break //remove this break to deploy all images for selected version
					} else {
						dataServiceIDImagesMap[versions[i].GetId()] = append(dataServiceIDImagesMap[versions[i].GetId()], images[j].GetId())
					}
				}
				break
			}
		}
	}
	for key := range dataServiceNameVersionMap {
		logrus.Infof("DS Version name- %v,id- %v", key, dataServiceNameVersionMap[key])
	}

	for key := range dataServiceIDImagesMap {
		logrus.Infof("DS Image name- %v,id- %v", key, dataServiceIDImagesMap[key])
	}
	return dataServiceNameVersionMap, dataServiceIDImagesMap
}

// GetAllVersions returns all the versions and build ID of dataservices
func GetAllVersions(dataServiceNameIDMap map[string]string) (map[string][]string, map[string][]string) {
	var versions []pds.ModelsVersion
	var images []pds.ModelsImage
	for key := range dataServiceNameIDMap {
		versions, _ = components.Version.ListDataServiceVersions(dataServiceNameIDMap[key])
		for i := 0; i < len(versions); i++ {
			if *versions[i].Enabled {
				dataServiceNameVersionMap[key] = append(dataServiceNameVersionMap[key], versions[i].GetId())
				images, _ = components.Image.ListImages(versions[i].GetId())
				for j := 0; j < len(images); j++ {
					dataServiceIDImagesMap[versions[i].GetId()] = append(dataServiceIDImagesMap[versions[i].GetId()], images[j].GetId())
				}
			}
		}
	}

	for key := range dataServiceNameVersionMap {
		logrus.Infof("DS Version name- %v,id- %v", key, dataServiceNameVersionMap[key])
	}

	for key := range dataServiceIDImagesMap {
		logrus.Infof("DS Image name- %v,id- %v", key, dataServiceIDImagesMap[key])
	}

	return dataServiceNameVersionMap, dataServiceIDImagesMap
}

// GetAppConfTemplate returns the app config templates
func GetAppConfTemplate(tenantID string, dataServiceNameIDMap map[string]string) map[string]string {
	appConfigs, _ := components.AppConfigTemplate.ListTemplates(tenantID)
	for i := 0; i < len(appConfigs); i++ {
		if appConfigs[i].GetName() == appConfigTemplateName {
			for key := range dataServiceNameIDMap {
				if dataServiceNameIDMap[key] == appConfigs[i].GetDataServiceId() {
					dataServiceNameDefaultAppConfigMap[key] = appConfigs[i].GetId()
				}
			}
		}
	}
	return dataServiceNameDefaultAppConfigMap
}

// DeleteDeploymentPods deletes the given pods
func DeleteDeploymentPods(podList *v1.PodList) error {
	var pods, newPods []v1.Pod
	k8sOps := k8sCore

	pods = append(pods, podList.Items...)
	err := k8sOps.DeletePods(pods, true)
	if err != nil {
		return err
	}

	//get the newly created pods
	time.Sleep(10 * time.Second)
	newPodList, err := GetPods("pds-system")
	if err != nil {
		return err
	}
	//reinitializing the pods
	newPods = append(newPods, newPodList.Items...)

	//validate deployment pods are up and running after deletion
	for _, pod := range newPods {
		logrus.Infof("pds system pod name %v", pod.Name)
		err = k8sOps.ValidatePod(&pod, timeOut, timeInterval)
		if err != nil {
			return err
		}
	}
	return nil
}

//ValidateNamespaces validates the namespace is available for pds
func ValidateNamespaces(deploymentTargetID string, ns string, status string) error {
	isavailable = false
	waitErr := wait.Poll(timeOut, timeInterval, func() (bool, error) {
		pdsNamespaces, err := components.Namespace.ListNamespaces(deploymentTargetID)
		if err != nil {
			return false, err
		}
		for _, pdsNamespace := range pdsNamespaces {
			logrus.Infof("namespace name %v and status %v", *pdsNamespace.Name, *pdsNamespace.Status)
			if *pdsNamespace.Name == ns && *pdsNamespace.Status == status {
				isavailable = true
			}
		}
		if isavailable {
			return true, nil
		}

		return false, nil
	})
	return waitErr
}

//DeletePDSNamespace deletes the given namespace
func DeletePDSNamespace(namespace string) error {
	k8sOps := k8sCore
	err := k8sOps.DeleteNamespace(namespace)
	return err
}

//UpdatePDSNamespce updates the namespace
func UpdatePDSNamespce(name string, nsLables map[string]string) (*v1.Namespace, error) {
	k8sOps := k8sCore
	nsSpec := &v1.Namespace{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Namespace",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: nsLables,
		},
	}
	ns, err := k8sOps.UpdateNamespace(nsSpec)
	time.Sleep(10 * time.Second)
	if err != nil {
		return nil, err
	}
	return ns, nil
}

//CreatePDSNamespace creates pds namespace
func CreatePDSNamespace(name string) (*v1.Namespace, error) {
	k8sOps := k8sCore
	nsSpec := &v1.Namespace{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Namespace",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"name": name,
			},
		},
	}
	ns, err := k8sOps.CreateNamespace(nsSpec)
	if err != nil {
		return nil, err
	}
	return ns, nil
}

// GetPods returns the list of pods in namespace
func GetPods(namespace string) (*v1.PodList, error) {
	k8sOps := k8sCore
	podList, err := k8sOps.GetPods(namespace, nil)
	if err != nil {
		return nil, err
	}
	return podList, err
}

// GetnameSpaceID returns the namespace ID
func GetnameSpaceID(namespace string) string {
	var namespaceID string
	namespaces, err := components.Namespace.ListNamespaces(deploymentTargetID)
	for i := 0; i < len(namespaces); i++ {
		if namespaces[i].GetStatus() == "available" {
			if namespaces[i].GetName() == namespace {
				namespaceID = namespaces[i].GetId()
			}
			namespaceNameIDMap[namespaces[i].GetName()] = namespaces[i].GetId()
			logrus.Infof("Available namespace - Name: %v , Id: %v , Status: %v", namespaces[i].GetName(), namespaces[i].GetId(), namespaces[i].GetStatus())
		}
	}
	if err != nil {
		logrus.Fatalf("An Error Occured %v", err)
	}
	return namespaceID
}

//ValidateDataServiceDeployment checks if deployment is healthy and running
func ValidateDataServiceDeployment(deployment *pds.ModelsDeployment) {
	//To get the list of statefulsets in particular namespace
	time.Sleep(30 * time.Second)

	ss, err := k8sApps.GetStatefulSet(deployment.GetClusterResourceName(), GetAndExpectStringEnvVar("NAMESPACE"))
	if err != nil {
		logrus.Warnf("An Error Occured %v", err)
	}

	//validate the statefulset deployed in the namespace
	err = k8sApps.ValidateStatefulSet(ss, defaultRetryInterval)
	if err != nil {
		logrus.Fatalf("An Error Occured %v", err)
	}

	status, _ := components.DataServiceDeployment.GetDeploymentSatus(deployment.GetId())
	sleeptime := 0
	for status.GetHealth() != "Healthy" && sleeptime < duration {
		if sleeptime > 30 && len(status.GetHealth()) < 2 {
			logrus.Infof("Deployment details: Health status -  %v, procceeding with next deployment", status.GetHealth())
			break
		}
		time.Sleep(10 * time.Second)
		sleeptime += 10
		status, _ = components.DataServiceDeployment.GetDeploymentSatus(deployment.GetId())
		logrus.Infof("Health status -  %v", status.GetHealth())
	}
	if status.GetHealth() == "Healthy" {
		deployementIDNameMap[deployment.GetId()] = deployment.GetName()
	}
	logrus.Infof("Deployment details: Health status -  %v,Replicas - %v, Ready replicas - %v", status.GetHealth(), status.GetReplicas(), status.GetReadyReplicas())
}

//UpdateDataServices modifies the existing deployment
func UpdateDataServices(deploymentID string, appConfigID string, dataServiceImageMap map[string][]string, nodeCount int32, resourceTemplateID string) *pds.ModelsDeployment {

	for version := range dataServiceImageMap {
		for i := range dataServiceImageMap[version] {
			imageID := dataServiceImageMap[version][i]
			logrus.Infof("Version %v ImageID %v", version, imageID)

			deployment, err = components.DataServiceDeployment.UpdateDeployment(deploymentID, appConfigID, imageID, nodeCount, resourceTemplateID)
			if err != nil {
				logrus.Fatalf("An Error Occured %v", err)
			}
			ValidateDataServiceDeployment(deployment)
		}
	}
	return deployment
}

// DeployDataServices deploys the dataservice
func DeployDataServices(projectID string, deploymentTargetID string, dnsZone string, deploymentName string,
	namespaceID string, dataServiceNameDefaultAppConfig string, dataServiceImageMap map[string][]string, replicas int32,
	serviceType string, dataServiceDefaultResourceTemplateID string, storageTemplateID string) *pds.ModelsDeployment {

	for version := range dataServiceImageMap {
		for index := range dataServiceImageMap[version] {
			imageID := dataServiceImageMap[version][index]
			logrus.Infof("Version %v ImageID %v", version, imageID)
			deployment, err = components.DataServiceDeployment.CreateDeployment(projectID,
				deploymentTargetID,
				dnsZone,
				deploymentName,
				namespaceID,
				dataServiceNameDefaultAppConfig,
				imageID,
				replicas,
				serviceType,
				dataServiceDefaultResourceTemplateID,
				storageTemplateID)

			if err != nil {
				logrus.Warnf("An Error Occured while creating deployment %v", err)
			}
			ValidateDataServiceDeployment(deployment)
		}
	}
	return deployment
}

// func CreateKubernetesClientFromConfig(kubeConfigPath string) (clientset.Interface, error) {
// 	cfg, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
// 	if err != nil {
// 		return nil, err
// 	}

// 	client, err := clientset.NewForConfig(cfg)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return client, nil
// }

// func Getk8sClient() clientset.Interface {
// 	var err error
// 	var k8sClient clientset.Interface
// 	if k8sClient == nil {
// 		if k8senv := os.Getenv("TARGET_KUBECONFIG"); k8senv != "" {
// 			k8sClient, err = CreateKubernetesClientFromConfig(k8senv)
// 			if err != nil {
// 				log.Panicf("Unable to get the client")
// 			}
// 		}
// 	}
// 	return k8sClient
// }

// func WaitForDeployment(ctx context.Context, client clientset.Interface, deployment *appsv1.Deployment, namespace string, replicas int32) error {
// 	var err error
// 	waitErr := wait.PollImmediate(timeInterval, timeOut, func() (bool, error) {
// 		log.Info("waiting for the deployment to complete")
// 		deployment, err = client.AppsV1().Deployments(namespace).Get(ctx, deployment.Name, metav1.GetOptions{})
// 		if err != nil {
// 			return false, nil
// 		}
// 		if deployment.Status.ReadyReplicas == replicas {
// 			return true, nil
// 		}
// 		return true, nil
// 	})
// 	return waitErr
// }

//GetDeploymentConnectionInfo returns the dns endpoint
func GetDeploymentConnectionInfo(deploymentID string) string {
	var isfound bool
	var dnsEndpoint string

	dataServiceDeployment := components.DataServiceDeployment
	deploymentConnectionDetails, clusterDetails, err := dataServiceDeployment.GetConnectionDetails(deploymentID)
	deploymentConnectionDetails.MarshalJSON()
	if err != nil {
		log.Fatalf("An Error Occured %v", err)
	}
	deploymentNodes := deploymentConnectionDetails.GetNodes()
	log.Infof("Deployment nodes %v", deploymentNodes)
	isfound = false
	for key, value := range clusterDetails {
		logrus.Infof("host details key %v value %v", key, value)
		if key == "host" {
			dnsEndpoint = fmt.Sprint(value)
			isfound = true
		}
	}
	if !isfound {
		log.Fatal("No connection string found")
	}

	return dnsEndpoint
}

//GetDeploymentCredentials returns the password to connect to the dataservice
func GetDeploymentCredentials(deploymentID string) string {
	dataServiceDeployment := components.DataServiceDeployment
	dataServicePassword, err := dataServiceDeployment.GetDeploymentCredentials(deploymentID)
	if err != nil {
		logrus.Fatalf("An Error Occured %v", err)
	}
	pdsPassword := dataServicePassword.GetPassword()
	return pdsPassword
}

//CreateDataServiceWorkloads func
func CreateDataServiceWorkloads(dataServiceName string, deploymentID string, scalefactor string, iterations string, deploymentName string, namespace string) {
	dnsEndpoint := GetDeploymentConnectionInfo(deploymentID)
	log.Infof("Dataservice DNS endpoint %s", dnsEndpoint)

	pdsPassword := GetDeploymentCredentials(deploymentID)

	switch dataServiceName {
	case "PostgreSQL":
		CreatepostgresqlWorkload(dnsEndpoint, pdsPassword, scalefactor, iterations, deploymentName, namespace)

	case "RabbitMQ":
		env := []string{"AMQP_HOST", "PDS_USER", "PDS_PASS"}
		CreateRmqWorkload(dnsEndpoint, pdsPassword, namespace, env)
	}

}

//CreatepostgresqlWorkload generate workloads on the pg db
func CreatepostgresqlWorkload(dnsEndpoint string, pdsPassword string, scalefactor string, iterations string, deploymentName string, namespace string) (*appsv1.Deployment, error) {
	var replicas int32 = 1
	deploymentSpec := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
			Namespace: namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": deploymentName},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": deploymentName},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:    "pgbench",
							Image:   "madan19/pgbench:pgloadTest1",
							Command: []string{"/pgloadgen.sh"},
							Args:    []string{dnsEndpoint, pdsPassword, scalefactor, iterations},
						},
					},
					RestartPolicy: v1.RestartPolicyAlways,
				},
			},
		},
	}
	deployment, err := k8sApps.CreateDeployment(deploymentSpec, metav1.CreateOptions{})
	if err != nil {
		logrus.Fatalf("An Error Occured while creating deployment %v", err)
	}
	err = k8sApps.ValidateDeployment(deployment, timeOut, timeInterval)
	if err != nil {
		logrus.Fatalf("An Error Occured while validating the pod %v", err)
	}
	//run workload for few min and terminate the deployment
	time.Sleep(1 * time.Minute)
	err = k8sApps.DeleteDeployment(deployment.Name, namespace)
	if err != nil {
		logrus.Fatalf("An Error Occured while deleting the deployment %v", err)
	}
	return deployment, err
}

//CreateRmqWorkload generate workloads for rmq
func CreateRmqWorkload(dnsEndpoint string, pdsPassword string, namespace string, env []string) (*v1.Pod, error) {
	var value []string
	podSpec := &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "rmq-perf-",
			Namespace:    namespace,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:    "rmqperf",
					Image:   "pivotalrabbitmq/perf-test:latest",
					Command: []string{"/bin/sh", "-c", "while true; do bin/runjava com.rabbitmq.perf.PerfTest --uri amqp://${PDS_USER}:${PDS_PASS}@${AMQP_HOST} -jb -s 10240 -z 30; done"},
					Env:     make([]v1.EnvVar, 3),
				},
			},
			RestartPolicy: v1.RestartPolicyOnFailure,
		},
	}

	value = append(value, dnsEndpoint)
	value = append(value, "pds")
	value = append(value, pdsPassword)

	for index := range env {
		podSpec.Spec.Containers[0].Env[index].Name = env[index]
		podSpec.Spec.Containers[0].Env[index].Value = value[index]
	}

	pod, err := k8sCore.CreatePod(podSpec)
	if err != nil {
		logrus.Fatalf("An Error Occured while creating %v", err)
	}

	err = k8sCore.ValidatePod(pod, timeOut, timeInterval)
	if err != nil {
		logrus.Fatalf("An Error Occured while validating the pod %v", err)
	}

	time.Sleep(1 * time.Minute)
	err = k8sCore.DeletePod(pod.Name, pod.Namespace, true)
	if err != nil {
		logrus.Fatalf("An Error Occured while deleting namespace %v", err)
	}

	return pod, nil
}
