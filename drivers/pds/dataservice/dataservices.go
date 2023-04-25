package dataservice

import (
	"fmt"
	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	pdsapi "github.com/portworx/torpedo/drivers/pds/api"
	pdscontrolplane "github.com/portworx/torpedo/drivers/pds/controlplane"
	pdslib "github.com/portworx/torpedo/drivers/pds/lib"
	"github.com/portworx/torpedo/drivers/pds/parameters"
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/drivers/scheduler/spec"
	"github.com/portworx/torpedo/pkg/log"
	"net/url"
)

// PDS vars
var (
	components   *pdsapi.Components
	deployment   *pds.ModelsDeployment
	controlplane *pdscontrolplane.ControlPlane
	apiClient    *pds.APIClient

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
	isStorageTemplateAvailable            bool
	dataServiceDefaultResourceTemplateID  string
	dataServiceDefaultAppConfigID         string
	dataServiceVersionBuildMap            = make(map[string][]string)
	dataServiceImageMap                   = make(map[string][]string)
)

// PDS const
const (
	storageTemplateName   = "QaDefault"
	resourceTemplateName  = "Small"
	appConfigTemplateName = "QaDefault"
	zookeeper             = "ZooKeeper"
	redis                 = "Redis"
	deploymentName        = "qa"
)

// K8s/PDS Instances
var (
	serviceType  = "LoadBalancer"
	customparams *parameters.Customparams
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

	namespaceID, err := pdslib.GetnameSpaceID(namespace, testParams.DeploymentTargetId)
	log.FailOnError(err, "Error while getting namespace id")

	log.InfoD("Deploying DataService %v ", ds.Name)
	deployment, dataServiceImageMap, dataServiceVersionBuildMap, err = d.DeployDS(ds.Name, projectID,
		testParams.DeploymentTargetId,
		testParams.DnsZone,
		deploymentName,
		namespaceID,
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

func (d *DataserviceType) DeployDataservicesAndCreateContext() ([]*scheduler.Context, error) {
	log.InfoD("Deployment of pds apps called from schedule applications")
	var deployments = make(map[PDSDataService]*pds.ModelsDeployment)
	var pdsApps []*pds.ModelsDeployment
	contexts := make([]*scheduler.Context, 0)
	var testparams TestParams

	pdsParams := pdslib.GetAndExpectStringEnvVar("PDS_PARAM_CM")
	params, err := customparams.ReadParams(pdsParams)
	if err != nil {
		return nil, fmt.Errorf("failed to read pds params %v", err)
	}
	infraParams := params.InfraToTest
	namespace := params.InfraToTest.Namespace

	_, isAvailable, err := pdslib.CreatePDSNamespace(namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to create pds namespace %v", err)
	}
	if !isAvailable {
		return nil, fmt.Errorf("pdsnamespace %v is not available to deploy apps", namespace)
	}
	_, tenantID, dnsZone, projectID, _, clusterID, err := pdslib.SetupPDSTest(infraParams.ControlPlaneURL, infraParams.ClusterType,
		infraParams.AccountName, infraParams.TenantName, infraParams.ProjectName)
	if err != nil {
		return nil, fmt.Errorf("Failed on SetupPDSTest method %v", err)
	}
	testparams.DnsZone = dnsZone

	deploymentTargetID, err = pdslib.GetDeploymentTargetID(clusterID, tenantID)
	log.FailOnError(err, "Failed to get the deployment TargetID")
	log.InfoD("DeploymentTargetID %s ", deploymentTargetID)
	testparams.DeploymentTargetId = deploymentTargetID

	namespaceId, err := pdslib.GetnameSpaceID(namespace, deploymentTargetID)
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
