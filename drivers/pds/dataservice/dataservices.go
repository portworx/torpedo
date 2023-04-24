package dataservice

import (
	"fmt"
	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	"github.com/portworx/sched-ops/k8s/apiextensions"
	"github.com/portworx/sched-ops/k8s/apps"
	"github.com/portworx/sched-ops/k8s/core"
	pdsapi "github.com/portworx/torpedo/drivers/pds/api"
	pdscontrolplane "github.com/portworx/torpedo/drivers/pds/controlplane"
	"github.com/portworx/torpedo/drivers/pds/parameters"
	"github.com/portworx/torpedo/drivers/pds/targetcluster"
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/drivers/scheduler/spec"
	"github.com/portworx/torpedo/pkg/aetosutil"
	"github.com/portworx/torpedo/pkg/log"
	corev1 "k8s.io/api/core/v1"
	"net/url"
	"strings"
	"time"
)

// PDS vars
var (
	components    *pdsapi.Components
	deployment    *pds.ModelsDeployment
	controlplane  *pdscontrolplane.ControlPlane
	apiClient     *pds.APIClient
	ns            *corev1.Namespace
	pdsAgentpod   corev1.Pod
	ApiComponents *pdsapi.Components

	err                                   error
	isavailable                           bool
	isTemplateavailable                   bool
	isVersionAvailable                    bool
	isBuildAvailable                      bool
	currentReplicas                       int32
	deploymentTargetID, storageTemplateID string
	resourceTemplateID                    string
	appConfigTemplateID                   string
	versionID                             string
	imageID                               string
	serviceAccId                          string
	accountID                             string
	projectID                             string
	tenantID                              string
	istargetclusterAvailable              bool
	isAccountAvailable                    bool
	isStorageTemplateAvailable            bool
	dnsZone                               string
	dataServiceDefaultResourceTemplateID  string
	dataServiceDefaultAppConfigID         string
	dash                                  *aetosutil.Dashboard

	dataServiceDefaultResourceTemplateIDMap = make(map[string]string)
	dataServiceNameIDMap                    = make(map[string]string)
	dataServiceNameVersionMap               = make(map[string][]string)
	dataServiceIDImagesMap                  = make(map[string][]string)
	dataServiceNameDefaultAppConfigMap      = make(map[string]string)
	deploymentsMap                          = make(map[string][]*pds.ModelsDeployment)
	namespaceNameIDMap                      = make(map[string]string)
	dataServiceVersionBuildMap              = make(map[string][]string)
	dataServiceImageMap                     = make(map[string][]string)
)

type PDS_Health_Status string

// PDS const
const (
	PDS_Health_Status_DOWN     PDS_Health_Status = "Down"
	PDS_Health_Status_DEGRADED PDS_Health_Status = "Degraded"
	PDS_Health_Status_HEALTHY  PDS_Health_Status = "Healthy"

	defaultCommandRetry          = 5 * time.Second
	defaultCommandTimeout        = 1 * time.Minute
	storageTemplateName          = "QaDefault"
	resourceTemplateName         = "Small"
	appConfigTemplateName        = "QaDefault"
	defaultRetryInterval         = 10 * time.Minute
	duration                     = 900
	timeOut                      = 30 * time.Minute
	timeInterval                 = 10 * time.Second
	maxtimeInterval              = 30 * time.Second
	resiliencyInterval           = 1 * time.Second
	defaultTestConnectionTimeout = 15 * time.Minute
	defaultWaitRebootRetry       = 10 * time.Second
	envDsVersion                 = "DS_VERSION"
	envDsBuild                   = "DS_BUILD"
	zookeeper                    = "ZooKeeper"
	redis                        = "Redis"
	consul                       = "Consul"
	cassandraStresImage          = "scylladb/scylla:4.1.11"
	postgresqlStressImage        = "portworx/torpedo-pgbench:pdsloadTest"
	consulBenchImage             = "pwxbuild/consul-bench-0.1.1"
	consulAgentImage             = "pwxbuild/consul-agent-0.1.1"
	esRallyImage                 = "elastic/rally"
	cbloadImage                  = "portworx/pds-loadtests:couchbase-0.0.2"
	pdsTpccImage                 = "portworx/torpedo-tpcc-automation:v1"
	redisStressImage             = "redis:latest"
	rmqStressImage               = "pivotalrabbitmq/perf-test:latest"
	postgresql                   = "PostgreSQL"
	cassandra                    = "Cassandra"
	elasticSearch                = "Elasticsearch"
	couchbase                    = "Couchbase"
	rabbitmq                     = "RabbitMQ"
	mysql                        = "MySQL"
	pxLabel                      = "pds.portworx.com/available"
	defaultParams                = "../drivers/pds/parameters/pds_default_parameters.json"
	pdsParamsConfigmap           = "pds-params"
	configmapNamespace           = "default"
	deploymentName               = "qa"
)

// K8s/PDS Instances
var (
	k8sCore       = core.Instance()
	k8sApps       = apps.Instance()
	apiExtentions = apiextensions.Instance()
	serviceType   = "LoadBalancer"
	customparams  *parameters.Customparams
	k8            *targetcluster.K8sType
	cp            *pdscontrolplane.ControlPlane
	//cc            pdscontext.PdsContextCreation
)

type DataserviceType struct{}

type TestParams struct {
	DeploymentTargetId string
	DnsZone            string
	StorageTemplateId  string
	NamespaceId        string
	TenantId           string
	ProjectId          string
}

type PDSDataService struct {
	Name          string "json:\"Name\""
	Version       string "json:\"Version\""
	Image         string "json:\"Image\""
	Replicas      int    "json:\"Replicas\""
	ScaleReplicas int    "json:\"ScaleReplicas\""
	OldVersion    string "json:\"OldVersion\""
	OldImage      string "json:\"OldImage\""
}

// GetResourceTemplate get the resource template id
func GetResourceTemplate(tenantID string, supportedDataService string) (string, error) {
	log.Infof("Get the resource template for each data services")
	resourceTemplates, err := components.ResourceSettingsTemplate.ListTemplates(tenantID)
	if err != nil {
		return "", err
	}
	isavailable = false
	isTemplateavailable = false
	for i := 0; i < len(resourceTemplates); i++ {
		if resourceTemplates[i].GetName() == resourceTemplateName {
			isTemplateavailable = true
			dataService, err := components.DataService.GetDataService(resourceTemplates[i].GetDataServiceId())
			if err != nil {
				return "", err
			}
			if dataService.GetName() == supportedDataService {
				log.Infof("Data service name: %v", dataService.GetName())
				log.Infof("Resource template details ---> Name %v, Id : %v ,DataServiceId %v , StorageReq %v , Memoryrequest %v",
					resourceTemplates[i].GetName(),
					resourceTemplates[i].GetId(),
					resourceTemplates[i].GetDataServiceId(),
					resourceTemplates[i].GetStorageRequest(),
					resourceTemplates[i].GetMemoryRequest())

				isavailable = true
				resourceTemplateID = resourceTemplates[i].GetId()
			}
		}
	}
	if !(isavailable && isTemplateavailable) {
		log.Errorf("Template with Name %v does not exis", resourceTemplateName)
	}
	return resourceTemplateID, nil
}

// GetStorageTemplate return the storage template id
func GetStorageTemplate(tenantID string) (string, error) {
	log.InfoD("Get the storage template")
	storageTemplates, err := components.StorageSettingsTemplate.ListTemplates(tenantID)
	if err != nil {
		return "", err
	}
	isStorageTemplateAvailable = false
	for i := 0; i < len(storageTemplates); i++ {
		if storageTemplates[i].GetName() == storageTemplateName {
			isStorageTemplateAvailable = true
			log.InfoD("Storage template details -----> Name %v,Repl %v , Fg %v , Fs %v",
				storageTemplates[i].GetName(),
				storageTemplates[i].GetRepl(),
				storageTemplates[i].GetFg(),
				storageTemplates[i].GetFs())
			storageTemplateID = storageTemplates[i].GetId()
		}
	}
	if !isStorageTemplateAvailable {
		log.Fatalf("storage template %v is not available ", storageTemplateName)
	}
	return storageTemplateID, nil
}

// GetAppConfTemplate returns the app config template id
func GetAppConfTemplate(tenantID string, supportedDataService string) (string, error) {
	appConfigs, err := components.AppConfigTemplate.ListTemplates(tenantID)
	var d DataserviceType
	if err != nil {
		return "", err
	}
	isavailable = false
	isTemplateavailable = false
	dataServiceId := d.GetDataServiceID(supportedDataService)
	for i := 0; i < len(appConfigs); i++ {
		if appConfigs[i].GetName() == appConfigTemplateName {
			isTemplateavailable = true
			if dataServiceId == appConfigs[i].GetDataServiceId() {
				appConfigTemplateID = appConfigs[i].GetId()
				isavailable = true
			}
		}
	}
	if !(isavailable && isTemplateavailable) {
		log.Errorf("App Config Template with name %v does not exist", appConfigTemplateName)
	}
	return appConfigTemplateID, nil
}

func GetDeploymentTargetID(clusterID, tenantID string) (string, error) {
	log.InfoD("Get the Target cluster details")
	targetClusters, err := components.DeploymentTarget.ListDeploymentTargetsBelongsToTenant(tenantID)
	if err != nil {
		return "", fmt.Errorf("error while listing deployments: %v", err)
	}
	if targetClusters == nil {
		return "", fmt.Errorf("target cluster passed is not available to the account/tenant %v", err)
	}
	for i := 0; i < len(targetClusters); i++ {
		if targetClusters[i].GetClusterId() == clusterID {
			deploymentTargetID = targetClusters[i].GetId()
			log.Infof("deploymentTargetID %v", deploymentTargetID)
			log.InfoD("Cluster ID: %v, Name: %v,Status: %v", targetClusters[i].GetClusterId(), targetClusters[i].GetName(), targetClusters[i].GetStatus())
		}
	}
	return deploymentTargetID, nil
}

// GetVersionsImage returns the required Image of dataservice version
func GetVersionsImage(dsVersion string, dsBuild string, dataServiceID string) (string, string, map[string][]string, error) {
	var versions []pds.ModelsVersion
	var images []pds.ModelsImage

	versions, err = components.Version.ListDataServiceVersions(dataServiceID)
	if err != nil {
		return "", "", nil, err
	}
	isVersionAvailable = false
	isBuildAvailable = false
	for i := 0; i < len(versions); i++ {
		log.Debugf("version name %s and is enabled=%t", *versions[i].Name, *versions[i].Enabled)
		if *versions[i].Name == dsVersion {
			log.Debugf("DS Version %s is enabled in the control plane", dsVersion)
			images, _ = components.Image.ListImages(versions[i].GetId())
			for j := 0; j < len(images); j++ {
				if *images[j].Build == dsBuild {
					versionID = versions[i].GetId()
					imageID = images[j].GetId()
					dataServiceVersionBuildMap[versions[i].GetName()] = append(dataServiceVersionBuildMap[versions[i].GetName()], images[j].GetBuild())
					isBuildAvailable = true
					break
				}
			}
			isVersionAvailable = true
			break
		}
	}
	if !(isVersionAvailable && isBuildAvailable) {
		return "", "", nil, fmt.Errorf("version/build passed is not available")
	}
	return versionID, imageID, dataServiceVersionBuildMap, nil
}

func (d *DataserviceType) GetDataServiceID(ds string) string {
	var dataServiceID string
	dsModel, err := components.DataService.ListDataServices()
	if err != nil {
		log.Errorf("An Error Occured while listing dataservices %v", err)
		return ""
	}
	for _, v := range dsModel {
		if *v.Name == ds {
			dataServiceID = *v.Id
		}
	}
	return dataServiceID
}

// DeployDS deploys dataservices its internally used function
func (d *DataserviceType) DeployDS(ds, projectID, deploymentTargetID, dnsZone, deploymentName, namespaceID, dataServiceDefaultAppConfigID string,
	replicas int32, serviceType, dataServiceDefaultResourceTemplateID, storageTemplateID, dsVersion,
	dsBuild, namespace string) (*pds.ModelsDeployment, map[string][]string, map[string][]string, error) {

	currentReplicas = replicas

	log.Infof("dataService: %v ", ds)
	id := d.GetDataServiceID(ds)
	if id == "" {
		log.Errorf("dataservice ID is empty")
		return nil, nil, nil, err
	}
	log.Infof(`Request params:
				projectID- %v deploymentTargetID - %v,
				dnsZone - %v,deploymentName - %v,namespaceID - %v
				App config ID - %v,
				num pods- %v, service-type - %v
				Resource template id - %v, storageTemplateID - %v`,
		projectID, deploymentTargetID, dnsZone, deploymentName, namespaceID, dataServiceDefaultAppConfigID,
		replicas, serviceType, dataServiceDefaultResourceTemplateID, storageTemplateID)

	if ds == zookeeper && replicas != 3 {
		log.Warnf("Zookeeper replicas cannot be %v, it should be 3", replicas)
		currentReplicas = 3
	}
	if ds == redis {
		log.Infof("Replicas passed %v", replicas)
		log.Warnf("Redis deployment replicas should be any one of the following values 1, 6, 8 and 10")
	}

	//clearing up the previous entries of dataServiceImageMap
	for version := range dataServiceImageMap {
		delete(dataServiceImageMap, version)
	}

	for version := range dataServiceVersionBuildMap {
		delete(dataServiceVersionBuildMap, version)
	}

	log.Infof("Getting versionID  for Data service version %s and buildID for %s ", dsVersion, dsBuild)
	versionID, imageID, dataServiceVersionBuildMap, err = GetVersionsImage(dsVersion, dsBuild, id)
	if err != nil {
		return nil, nil, nil, err
	}

	log.Infof("VersionID %v ImageID %v", versionID, imageID)
	components = pdsapi.NewComponents(apiClient)
	deployment, err = components.DataServiceDeployment.CreateDeployment(projectID,
		deploymentTargetID,
		dnsZone,
		deploymentName,
		namespaceID,
		dataServiceDefaultAppConfigID,
		imageID,
		currentReplicas,
		serviceType,
		dataServiceDefaultResourceTemplateID,
		storageTemplateID)

	if err != nil {
		log.Warnf("An Error Occured while creating deployment %v", err)
		return nil, nil, nil, err
	}

	return deployment, dataServiceImageMap, dataServiceVersionBuildMap, nil
}

func (d *DataserviceType) DeployDataservicesAndCreateContext() ([]*scheduler.Context, error) {
	log.InfoD("*************************Deployment called from schedule applications**************************")
	var deployments = make(map[PDSDataService]*pds.ModelsDeployment)
	var pdsApps []*pds.ModelsDeployment
	contexts := make([]*scheduler.Context, 0)
	var testparams TestParams

	pdsParams := k8.GetAndExpectStringEnvVar("PDS_PARAM_CM")
	params, err := customparams.ReadParams(pdsParams)
	if err != nil {
		return nil, fmt.Errorf("failed to read pds params %v", err)
	}
	infraParams := params.InfraToTest
	namespace := params.InfraToTest.Namespace

	_, isAvailable, err := k8.CreatePDSNamespace(namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to create pds namespace %v", err)
	}
	if !isAvailable {
		return nil, fmt.Errorf("pdsnamespace %v is not available to deploy apps", namespace)
	}
	_, tenantID, dnsZone, projectID, _, clusterID, err := d.SetupPDSTest(infraParams.ClusterType,
		infraParams.AccountName, infraParams.TenantName, infraParams.ProjectName)
	if err != nil {
		return nil, fmt.Errorf("Failed on SetupPDSTest method %v", err)
	}
	testparams.DnsZone = dnsZone

	deploymentTargetID, err = GetDeploymentTargetID(clusterID, tenantID)
	log.FailOnError(err, "Failed to get the deployment TargetID")
	log.InfoD("DeploymentTargetID %s ", deploymentTargetID)
	testparams.DeploymentTargetId = deploymentTargetID

	namespaceId, err := k8.GetnameSpaceID(namespace, deploymentTargetID)
	log.FailOnError(err, "Failed to get the namespace Id")
	log.InfoD("NamespaceId %s ", namespaceId)
	testparams.NamespaceId = namespaceId

	storageTemplateID, err = GetStorageTemplate(tenantID)
	log.FailOnError(err, "Failed while getting storage template ID")
	log.InfoD("storageTemplateID %v", storageTemplateID)
	testparams.StorageTemplateId = storageTemplateID

	for _, ds := range params.DataServiceToTest {
		deployment, _, _, err := d.TriggerDeployDataService(ds, namespace, tenantID, projectID, false, testparams)
		if err != nil {
			return nil, fmt.Errorf("failed to deploy pds apps %v", err)
		}
		deployments[ds] = deployment
		pdsApps = append(pdsApps, deployment)
	}

	log.InfoD("Creating Context for PDS Apps")
	contexts = d.CreateAppContext(pdsApps)
	return contexts, nil
}

func (d *DataserviceType) TriggerDeployDataService(ds PDSDataService, namespace, tenantID,
	projectID string, deployOldVersion bool, testParams TestParams) (*pds.ModelsDeployment, map[string][]string, map[string][]string, error) {
	log.InfoD("Going to start %v app deployment", ds.Name)
	var dsVersion string
	var dsImage string

	if deployOldVersion {
		dsVersion = ds.OldVersion
		dsImage = ds.OldImage
		log.Debugf("Deploying old version %s and image %s", dsVersion, dsImage)
	} else {
		dsVersion = ds.Version
		dsImage = ds.Image
		log.Debugf("Deploying latest version %s and image %s", dsVersion, dsImage)
	}

	log.InfoD("Getting Resource Template ID")
	dataServiceDefaultResourceTemplateID, err = GetResourceTemplate(tenantID, ds.Name)
	log.FailOnError(err, "Error while getting resource template")
	log.InfoD("dataServiceDefaultResourceTemplateID %v ", dataServiceDefaultResourceTemplateID)

	log.InfoD("Getting App Template ID")
	dataServiceDefaultAppConfigID, err = GetAppConfTemplate(tenantID, ds.Name)
	log.FailOnError(err, "Error while getting app configuration template")
	log.InfoD("dataServiceDefaultAppConfigID %v ", dataServiceDefaultAppConfigID)
	//dash.VerifyFatal(dataServiceDefaultAppConfigID != "", true, "Validating dataServiceDefaultAppConfigID")
	log.InfoD(" dataServiceDefaultAppConfigID %v ", dataServiceDefaultAppConfigID)

	log.InfoD("Deploying DataService %v ", ds.Name)
	deployment, dataServiceImageMap, dataServiceVersionBuildMap, err = d.DeployDS(ds.Name, projectID,
		testParams.DeploymentTargetId,
		testParams.DnsZone,
		deploymentName,
		testParams.NamespaceId,
		dataServiceDefaultAppConfigID,
		int32(ds.Replicas),
		serviceType,
		dataServiceDefaultResourceTemplateID,
		testParams.StorageTemplateId,
		dsVersion,
		dsImage,
		namespace,
	)
	log.FailOnError(err, "Error while deploying data services")

	return deployment, dataServiceImageMap, dataServiceVersionBuildMap, err
}

// SetupPDSTest returns few params required to run the test
func (d *DataserviceType) SetupPDSTest(ClusterType, AccountName, TenantName, ProjectName string) (string, string, string, string, string, string, error) {

	acc := components.Account
	accounts, err := acc.GetAccountsList()
	if err != nil {
		return "", "", "", "", "", "", err
	}

	isAccountAvailable = false
	for i := 0; i < len(accounts); i++ {
		log.InfoD("Account Name: %v", accounts[i].GetName())
		if accounts[i].GetName() == AccountName {
			isAccountAvailable = true
			accountID = accounts[i].GetId()
			break
		}
	}
	if !isAccountAvailable {
		return "", "", "", "", "", "", fmt.Errorf("account %v is not available", AccountName)
	}
	log.InfoD("Account Detail- Name: %s, UUID: %s ", AccountName, accountID)
	tnts := components.Tenant
	tenants, _ := tnts.GetTenantsList(accountID)
	for _, tenant := range tenants {
		if tenant.GetName() == TenantName {
			tenantID = tenant.GetId()
			break
		}

	}
	log.InfoD("Tenant Details- Name: %s, UUID: %s ", TenantName, tenantID)

	if strings.EqualFold(ClusterType, "onprem") || strings.EqualFold(ClusterType, "ocp") {
		serviceType = "ClusterIP"
	}
	log.InfoD("Deployment service type %s", serviceType)

	dnsZone, err := controlplane.GetDNSZone(tenantID)
	if err != nil {
		return "", "", "", "", "", "", err
	}
	log.InfoD("DNSZone: %s, tenantName: %s, accountName: %s", dnsZone, TenantName, AccountName)
	projcts := components.Project
	projects, _ := projcts.GetprojectsList(tenantID)
	for _, project := range projects {
		if project.GetName() == ProjectName {
			projectID = project.GetId()
			break
		}
	}
	log.InfoD("Project Details- Name: %s, UUID: %s ", ProjectName, projectID)

	ns, err = k8sCore.GetNamespace("kube-system")
	if err != nil {
		return "", "", "", "", "", "", err
	}
	clusterID := string(ns.GetObjectMeta().GetUID())
	if len(clusterID) > 0 {
		log.InfoD("clusterID %v", clusterID)
	} else {
		return "", "", "", "", "", "", fmt.Errorf("unable to get the clusterID")
	}

	return accountID, tenantID, dnsZone, projectID, serviceType, clusterID, err
}

func (d *DataserviceType) CreateAppContext(pdsApps []*pds.ModelsDeployment) []*scheduler.Context {
	var specObjects []interface{}
	var Contexts []*scheduler.Context
	var ctx *scheduler.Context

	for _, dep := range pdsApps {
		specObjects = append(specObjects, dep)
		ctx = &scheduler.Context{
			UID: dep.GetId(),
			App: &spec.AppSpec{
				Key:      *dep.ClusterResourceName,
				SpecList: specObjects,
			},
		}
		Contexts = append(Contexts, ctx)
	}
	return Contexts
}

//func init() {
//	//dsType := &DataserviceType{}
//	err = pds2.Register("dataservice", &DataserviceType{})
//	log.FailOnError(err, "Error while Registering pds dataservice type driver")
//}

func DataserviceInit(ControlPlaneURL string) (*DataserviceType, error) {
	apiConf := pds.NewConfiguration()
	endpointURL, err := url.Parse(ControlPlaneURL)
	if err != nil {
		return nil, err
	}
	apiConf.Host = endpointURL.Host
	apiConf.Scheme = endpointURL.Scheme

	apiClient = pds.NewAPIClient(apiConf)
	components = pdsapi.NewComponents(apiClient)
	controlplane = pdscontrolplane.NewControlPlane(ControlPlaneURL, components)

	return &DataserviceType{}, nil
}
