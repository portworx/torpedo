package tests

import (
	"fmt"
	"net/url"
	"strings"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	"github.com/portworx/sched-ops/k8s/core"
	pdsapi "github.com/portworx/torpedo/drivers/pds/api"
	pdscontrolplane "github.com/portworx/torpedo/drivers/pds/controlplane"
	pdslib "github.com/portworx/torpedo/drivers/pds/lib"
	"github.com/portworx/torpedo/pkg/aetosutil"
	"github.com/portworx/torpedo/pkg/log"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

type PDSDataService struct {
	Name          string "json:\"Name\""
	Version       string "json:\"Version\""
	Image         string "json:\"Image\""
	Replicas      int    "json:\"Replicas\""
	ScaleReplicas int    "json:\"ScaleReplicas\""
	OldVersion    string "json:\"OldVersion\""
	OldImage      string "json:\"OldImage\""
}

const (
	pdsNamespace            = "pds-system"
	deploymentName          = "qa"
	envDeployAllDataService = "DEPLOY_ALL_DATASERVICE"
	postgresql              = "PostgreSQL"
	cassandra               = "Cassandra"
	elasticSearch           = "Elasticsearch"
	couchbase               = "Couchbase"
	redis                   = "Redis"
	rabbitmq                = "RabbitMQ"
	mongodb                 = "MongoDB"
	mysql                   = "MySQL"
	kafka                   = "Kafka"
	zookeeper               = "ZooKeeper"
	pdsNamespaceLabel       = "pds.portworx.com/available"
)

var (
	namespace                               string
	pxnamespace                             string
	tenantID                                string
	dnsZone                                 string
	clusterID                               string
	projectID                               string
	serviceType                             string
	deploymentTargetID                      string
	replicas                                int32
	err                                     error
	supportedDataServices                   []string
	dataServiceNameDefaultAppConfigMap      map[string]string
	namespaceID                             string
	storageTemplateID                       string
	dataServiceDefaultResourceTemplateIDMap map[string]string
	dataServiceNameIDMap                    map[string]string
	supportedDataServicesNameIDMap          map[string]string
	DeployAllVersions                       bool
	DataService                             string
	DeployAllImages                         bool
	dataServiceDefaultResourceTemplateID    string
	dataServiceDefaultAppConfigID           string
	accountID                               string
	dataServiceVersionBuildMap              map[string][]string
	dataServiceImageMap                     map[string][]string
	dep                                     *v1.Deployment
	pod                                     *corev1.Pod
	params                                  *pdslib.Parameter
	podList                                 *corev1.PodList
	ns                                      *corev1.Namespace
	isDeploymentsDeleted                    bool
	isNamespacesDeleted                     bool
	isAccountAvailable                      bool
	dash                                    *aetosutil.Dashboard
	deployment                              *pds.ModelsDeployment
	k8sCore                                 = core.Instance()
	pdsLabels                               = make(map[string]string)
	apiClient                               *pds.APIClient
	components                              *pdsapi.Components
)

var dataServiceDeploymentWorkloads = []string{cassandra, elasticSearch, postgresql}
var dataServicePodWorkloads = []string{redis, rabbitmq, couchbase}

func RunWorkloads(params pdslib.WorkloadGenerationParams, ds PDSDataService, deployment *pds.ModelsDeployment, namespace string) (*corev1.Pod, *v1.Deployment, error) {
	params.DataServiceName = ds.Name
	params.DeploymentID = deployment.GetId()
	params.Namespace = namespace
	log.Infof("Dataservice Name : %s", ds.Name)

	if ds.Name == postgresql {
		params.DeploymentName = "pgload"
		params.ScaleFactor = "100"
		params.Iterations = "1"

		log.Infof("Running Workloads on DataService %v ", ds.Name)
		pod, dep, err = pdslib.CreateDataServiceWorkloads(params)

	}
	if ds.Name == rabbitmq {
		params.DeploymentName = "rmq"
		log.Infof("Running Workloads on DataService %v ", ds.Name)
		pod, dep, err = pdslib.CreateDataServiceWorkloads(params)

	}
	if ds.Name == redis {
		params.DeploymentName = "redisbench"
		log.Infof("Running Workloads on DataService %v ", ds.Name)
		pod, dep, err = pdslib.CreateDataServiceWorkloads(params)

	}
	if ds.Name == cassandra {
		params.DeploymentName = "cassandra-stress"
		log.Infof("Running Workloads on DataService %v ", ds.Name)
		pod, dep, err = pdslib.CreateDataServiceWorkloads(params)

	}
	if ds.Name == elasticSearch {
		params.DeploymentName = "es-rally"
		params.User = "elastic"
		params.UseSSL = "false"
		params.VerifyCerts = "false"
		params.TimeOut = "60"
		log.Infof("Running Workloads on DataService %v ", ds.Name)
		pod, dep, err = pdslib.CreateDataServiceWorkloads(params)

	}
	if ds.Name == couchbase {
		params.DeploymentName = "cb-load"
		log.Infof("Running Workloads on DataService %v ", ds.Name)
		pod, dep, err = pdslib.CreateDataServiceWorkloads(params)

	}

	return pod, dep, err

}

// SetupPDSTest returns few params required to run the test
func SetupPDSTest(ControlPlaneURL, ClusterType, AccountName, TenantName, ProjectName string) (string, string, string, string, string, error) {
	var err error
	apiConf := pds.NewConfiguration()
	endpointURL, err := url.Parse(ControlPlaneURL)
	if err != nil {
		return "", "", "", "", "", err
	}
	apiConf.Host = endpointURL.Host
	apiConf.Scheme = endpointURL.Scheme

	apiClient = pds.NewAPIClient(apiConf)
	components = pdsapi.NewComponents(apiClient)
	controlplane := pdscontrolplane.NewControlPlane(ControlPlaneURL, components)

	if strings.EqualFold(ClusterType, "onprem") || strings.EqualFold(ClusterType, "ocp") {
		serviceType = "ClusterIP"
	}
	log.InfoD("Deployment service type %s", serviceType)

	acc := components.Account
	accounts, err := acc.GetAccountsList()
	if err != nil {
		return "", "", "", "", "", err
	}

	isAccountAvailable = false
	for i := 0; i < len(accounts); i++ {
		log.InfoD("Account Name: %v", accounts[i].GetName())
		if accounts[i].GetName() == AccountName {
			isAccountAvailable = true
			accountID = accounts[i].GetId()
		}
	}
	if !isAccountAvailable {
		log.Fatalf("Account %v is not available", AccountName)
	}
	log.InfoD("Account Detail- Name: %s, UUID: %s ", AccountName, accountID)
	tnts := components.Tenant
	tenants, _ := tnts.GetTenantsList(accountID)
	for _, tenant := range tenants {
		if tenant.GetName() == TenantName {
			tenantID = tenant.GetId()
		}

	}
	log.InfoD("Tenant Details- Name: %s, UUID: %s ", TenantName, tenantID)
	dnsZone, err := controlplane.GetDNSZone(tenantID)
	if err != nil {
		return "", "", "", "", "", err
	}
	log.InfoD("DNSZone: %s, tenantName: %s, accountName: %s", dnsZone, TenantName, AccountName)
	projcts := components.Project
	projects, _ := projcts.GetprojectsList(tenantID)
	for _, project := range projects {
		if project.GetName() == ProjectName {
			projectID = project.GetId()
		}
	}
	log.InfoD("Project Details- Name: %s, UUID: %s ", ProjectName, projectID)

	ns, err = k8sCore.GetNamespace("kube-system")
	if err != nil {
		return "", "", "", "", "", err
	}
	clusterID := string(ns.GetObjectMeta().GetUID())
	if len(clusterID) > 0 {
		log.InfoD("clusterID %v", clusterID)
	} else {
		return "", "", "", "", "", fmt.Errorf("unable to get the clusterID")
	}

	return tenantID, dnsZone, projectID, serviceType, clusterID, err
}
