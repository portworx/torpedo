package lib

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	state "net/http"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	"github.com/portworx/sched-ops/k8s/apps"
	"github.com/portworx/sched-ops/k8s/core"
	pdsapi "github.com/portworx/torpedo/drivers/pds/api"
	"github.com/sirupsen/logrus"
)

//PDS const
const (
	storageTemplateName   = "Volume replication (best-effort spread)"
	resourceTemplateName  = "Small"
	appConfigTemplateName = "QaDefault"
	defaultRetryInterval  = 10 * time.Minute
	duration              = 900
	timeOut               = 5 * time.Minute
	timeInterval          = 10 * time.Second
	envDsVersion          = "DS_VERSION"
	envDsBuild            = "DS_BUILD"
	envDeployAllVersions  = "DEPLOY_ALL_VERSIONS"
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
	isVersionAvailable                    bool
	isBuildAvailable                      bool
	currentReplicas                       int32

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

// GetClusterID retruns the cluster id
func GetClusterID(pathToKubeconfig string) (string, error) {
	logrus.Infof("Fetch Cluster id ")
	cmd := fmt.Sprintf("kubectl get ns kube-system -o jsonpath={.metadata.uid} --kubeconfig %s", pathToKubeconfig)
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

//GetStorageTemplate return the storage template id
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
	isavailable = false
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
					isavailable = true
				}
			}
			isavailable = true
		}
	}
	if !isavailable {
		logrus.Errorf("Template with Name %v does not exis", resourceTemplateName)
	}
	return dataServiceDefaultResourceTemplateIDMap, dataServiceNameIDMap
}

// GetAppConfTemplate returns the app config templates
func GetAppConfTemplate(tenantID string, dataServiceNameIDMap map[string]string) map[string]string {
	appConfigs, _ := components.AppConfigTemplate.ListTemplates(tenantID)
	isavailable = false
	for i := 0; i < len(appConfigs); i++ {
		if appConfigs[i].GetName() == appConfigTemplateName {
			for key := range dataServiceNameIDMap {
				if dataServiceNameIDMap[key] == appConfigs[i].GetDataServiceId() {
					dataServiceNameDefaultAppConfigMap[key] = appConfigs[i].GetId()
					isavailable = true
				}
			}
			isavailable = true
		}
	}
	if !isavailable {
		logrus.Errorf("App Config Template with name %v does not exist", appConfigTemplateName)
	}
	return dataServiceNameDefaultAppConfigMap
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

// GetVersionsImage returns the required Image of dataservice version
func GetVersionsImage(dsVersion string, dsBuild string, dataServiceID string) (map[string][]string, map[string][]string) {
	var versions []pds.ModelsVersion
	var images []pds.ModelsImage

	versions, _ = components.Version.ListDataServiceVersions(dataServiceID)
	isVersionAvailable = false
	isBuildAvailable = false
	for i := 0; i < len(versions); i++ {
		if (*versions[i].Enabled) && (*versions[i].Name == dsVersion) {
			dataServiceNameVersionMap[dataServiceID] = append(dataServiceNameVersionMap[dataServiceID], versions[i].GetId())
			images, _ = components.Image.ListImages(versions[i].GetId())
			for j := 0; j < len(images); j++ {
				if *images[j].Build == dsBuild {
					dataServiceIDImagesMap[versions[i].GetId()] = append(dataServiceIDImagesMap[versions[i].GetId()], images[j].GetId())
					isBuildAvailable = true
					break //remove this break to deploy all images for selected version
				}
			}
			isVersionAvailable = true
			break
		}
	}
	if !(isVersionAvailable && isBuildAvailable) {
		logrus.Fatal("Version/Build passed is not available")
	}

	for key := range dataServiceNameVersionMap {
		logrus.Infof("DS name- %v,version ids- %v", key, dataServiceNameVersionMap[key])
	}

	for key := range dataServiceIDImagesMap {
		logrus.Infof("DS Verion id - %v, DS Image id - %v", key, dataServiceIDImagesMap[key])
	}
	return dataServiceNameVersionMap, dataServiceIDImagesMap
}

// GetAllVersionsImages returns all the versions and Images of dataservice
func GetAllVersionsImages(dataServiceID string) (map[string][]string, map[string][]string) {
	var versions []pds.ModelsVersion
	var images []pds.ModelsImage

	versions, _ = components.Version.ListDataServiceVersions(dataServiceID)
	for i := 0; i < len(versions); i++ {
		if *versions[i].Enabled {
			dataServiceNameVersionMap[dataServiceID] = append(dataServiceNameVersionMap[dataServiceID], versions[i].GetId())
			images, _ = components.Image.ListImages(versions[i].GetId())
			for j := 0; j < len(images); j++ {
				dataServiceIDImagesMap[versions[i].GetId()] = append(dataServiceIDImagesMap[versions[i].GetId()], images[j].GetId())
			}
		}
	}

	for key := range dataServiceNameVersionMap {
		logrus.Infof("DS name- %v, version ids- %v", key, dataServiceNameVersionMap[key])
	}
	for key := range dataServiceIDImagesMap {
		logrus.Infof("DS Verion id - %v,DS Image id - %v", key, dataServiceIDImagesMap[key])
	}
	return dataServiceNameVersionMap, dataServiceIDImagesMap
}

//ValidateDataServiceDeployment checks if deployment is healthy and running
func ValidateDataServiceDeployment(deployment *pds.ModelsDeployment) {
	//To get the list of statefulsets in particular namespace
	time.Sleep(30 * time.Second)

	ss, err := k8sApps.GetStatefulSet(deployment.GetClusterResourceName(), GetAndExpectStringEnvVar("NAMESPACE"))
	if err != nil {
		logrus.Warnf("An Error Occured while getting statefulsets %v", err)
	}

	//validate the statefulset deployed in the namespace
	err = k8sApps.ValidateStatefulSet(ss, defaultRetryInterval)
	if err != nil {
		logrus.Fatalf("An Error Occured while validating statefulsets %v", err)
	}

	status, res, err := components.DataServiceDeployment.GetDeploymentSatus(deployment.GetId())
	if err != nil {
		logrus.Fatalf("An Error Occured while get the deployment status %v", err)
	}
	if res.StatusCode != state.StatusOK {
		logrus.Errorf("Error when calling `ApiDeploymentsIdStatusGet``: %v\n", err)
		logrus.Errorf("Full HTTP response: %v\n", res)
	}
	sleeptime := 0
	for status.GetHealth() != "Healthy" && sleeptime < duration {
		if sleeptime > 30 && len(status.GetHealth()) < 2 {
			logrus.Infof("Deployment details: Health status -  %v, procceeding with next deployment", status.GetHealth())
			break
		}
		time.Sleep(10 * time.Second)
		sleeptime += 10
		status, res, err = components.DataServiceDeployment.GetDeploymentSatus(deployment.GetId())
		logrus.Infof("Health status -  %v", status.GetHealth())
		if err != nil {
			logrus.Fatalf("Error occured while getting deployment status %v", err)
		}
		if res.StatusCode != state.StatusOK {
			logrus.Errorf("Error when calling `ApiDeploymentsIdCredentialsGet``: %v\n", err)
			logrus.Errorf("Full HTTP response: %v\n", res)
		}
	}
	if status.GetHealth() == "Healthy" {
		deployementIDNameMap[deployment.GetId()] = deployment.GetName()
	}
	logrus.Infof("Deployment details: Health status -  %v,Replicas - %v, Ready replicas - %v", status.GetHealth(), status.GetReplicas(), status.GetReadyReplicas())

}

// DeleteDeployment deletes the given deployment
func DeleteDeployment(deploymentID string) (*state.Response, error) {
	resp, err := components.DataServiceDeployment.DeleteDeployment(deploymentID)
	if err != nil {
		logrus.Errorf("An Error Occured while deleting deployment %v", err)
	}
	if resp.StatusCode != state.StatusAccepted {
		logrus.Errorf("HTTP response failed: %v\n", resp)
	}
	return resp, err
}

// DeployDataServices deploys all dataservices, versions and images that are supported
func DeployDataServices(supportedDataServicesMap map[string]string, projectID string, deploymentTargetID string, dnsZone string, deploymentName string,
	namespaceID string, dataServiceNameDefaultAppConfigMap map[string]string, replicas int32,
	serviceType string, dataServiceDefaultResourceTemplateIDMap map[string]string, storageTemplateID string) map[string]string {

	currentReplicas = replicas
	var dataServiceImageMap map[string][]string

	for ds, id := range supportedDataServicesMap {
		logrus.Infof("dataService: %v ", ds)
		logrus.Infof(`Request params:
				projectID- %v deploymentTargetID - %v,
				dnsZone - %v,deploymentName - %v,namespaceID - %v
				App config ID - %v,
				num pods- %v, service-type - %v
				Resource template id - %v, storageTemplateID - %v`,
			projectID, deploymentTargetID, dnsZone, deploymentName, namespaceID, dataServiceNameDefaultAppConfigMap[ds],
			replicas, serviceType, dataServiceDefaultResourceTemplateIDMap[ds], storageTemplateID)

		if ds == "ZooKeeper" && replicas != 3 {
			logrus.Warnf("Zookeeper replicas cannot be %v, it should be 3", replicas)
			currentReplicas = 3
		}
		if ds == "Redis" {
			logrus.Infof("Replicas passed %v", replicas)
			logrus.Warn("Redis deployment replicas should be any one of the following values 1, 6, 8 and 10")
		}

		//clearing up the previous entries of dataServiceImageMap
		for ds := range dataServiceImageMap {
			delete(dataServiceImageMap, ds)
		}

		if !GetAndExpectBoolEnvVar(envDeployAllVersions) {
			dsVersion := GetAndExpectStringEnvVar(envDsVersion)
			dsBuild := GetAndExpectStringEnvVar(envDsBuild)
			logrus.Infof("Getting versionID  for Data service version %s and buildID for %s ", dsVersion, dsBuild)
			_, dataServiceImageMap = GetVersionsImage(dsVersion, dsBuild, id)
		} else {
			_, dataServiceImageMap = GetAllVersionsImages(id)
		}

		for version := range dataServiceImageMap {
			for index := range dataServiceImageMap[version] {
				imageID := dataServiceImageMap[version][index]
				logrus.Infof("VersionID %v ImageID %v", version, imageID)
				deployment, err = components.DataServiceDeployment.CreateDeployment(projectID,
					deploymentTargetID,
					dnsZone,
					deploymentName,
					namespaceID,
					dataServiceNameDefaultAppConfigMap[ds],
					imageID,
					currentReplicas,
					serviceType,
					dataServiceDefaultResourceTemplateIDMap[ds],
					storageTemplateID)

				if err != nil {
					logrus.Warnf("An Error Occured while creating deployment %v", err)
				}
				ValidateDataServiceDeployment(deployment)
				deployementIDNameMap[deployment.GetId()] = deployment.GetName()

			}
		}
	}
	return deployementIDNameMap
}

//GetAllSupportedDataServices get the supported datasservices and returns the map
func GetAllSupportedDataServices() map[string]string {
	dataService, _ := components.DataService.ListDataServices()
	for _, ds := range dataService {
		if !*ds.ComingSoon {
			dataServiceNameIDMap[ds.GetName()] = ds.GetId()
		}
	}
	for key, value := range dataServiceNameIDMap {
		logrus.Infof("dsKey %v dsValue %v", key, value)
	}
	return dataServiceNameIDMap
}
