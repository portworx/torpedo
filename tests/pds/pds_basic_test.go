package tests

import (
	"os"
	"testing"

	"github.com/portworx/torpedo/pkg/log"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
	pdslib "github.com/portworx/torpedo/drivers/pds/lib"

	"github.com/portworx/torpedo/pkg/aetosutil"
	. "github.com/portworx/torpedo/tests"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

const (
	deploymentName          = "qa"
	envDeployAllDataService = "DEPLOY_ALL_DATASERVICE"
	postgresql              = "PostgreSQL"
	cassandra               = "Cassandra"
	redis                   = "Redis"
	rabbitmq                = "RabbitMQ"
	mysql                   = "MySQL"
	kafka                   = "Kafka"
	zookeeper               = "ZooKeeper"
)

var (
	namespace                               string
	pxnamespace                             string
	tenantID                                string
	dnsZone                                 string
	projectID                               string
	serviceType                             string
	deploymentTargetID                      string
	clusterID                               string
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
	dataServiceVersionBuildMap              map[string][]string
	dep                                     *v1.Deployment
	pod                                     *corev1.Pod
	params                                  *pdslib.Parameter
	dash                                    *aetosutil.Dashboard
)

//var ds struct
var ds struct {
	Name          string "json:\"Name\""
	Version       string "json:\"Version\""
	Image         string "json:\"Image\""
	Replicas      int    "json:\"Replicas\""
	ScaleReplicas int    "json:\"ScaleReplicas\""
	OldVersion    string "json:\"OldVersion\""
	OldImage      string "json:\"OldImage\""
}

func TestDataService(t *testing.T) {
	RegisterFailHandler(Fail)

	var specReporters []Reporter
	junitReporter := reporters.NewJUnitReporter("/testresults/junit_basic.xml")
	specReporters = append(specReporters, junitReporter)
	RunSpecsWithDefaultAndCustomReporters(t, "Torpedo : pds", specReporters)

}

var _ = BeforeSuite(func() {
	Step("get prerequisite params to run the pds tests", func() {
		dash = Inst().Dash
		dash.Info("Initializing torpedo instance.")
		//InitInstance()
		dash.TestSetBegin(dash.TestSet)
		pdsparams := pdslib.GetAndExpectStringEnvVar("PDS_PARAM_CM")
		params, err = pdslib.ReadParams(pdsparams)
		Expect(err).NotTo(HaveOccurred())
		infraParams := params.InfraToTest

		dash.Info("Getting prerequisite for pds test")
		tenantID, dnsZone, projectID, serviceType, clusterID, err = pdslib.SetupPDSTest(infraParams.ControlPlaneURL, infraParams.ClusterType, infraParams.AccountName)
		dash.Infof("tenantID %v, projectID %v, serviceType %v, clusterID %v ", tenantID, projectID, serviceType, clusterID)
		Expect(err).NotTo(HaveOccurred())
	})

	Step("Check Target Cluster is registered to control", func() {
		infraParams := params.InfraToTest
		err = pdslib.RegisterToControlPlane(infraParams.ControlPlaneURL, tenantID, infraParams.ClusterType)
		Expect(err).NotTo(HaveOccurred())
	})

	Step("Get Deployment TargetID", func() {
		deploymentTargetID, err = pdslib.GetDeploymentTargetID(clusterID, tenantID)
		Expect(err).NotTo(HaveOccurred())
		dash.Infof("DeploymentTargetID %v ", deploymentTargetID)
	})

	Step("Get StorageTemplateID and Replicas", func() {
		storageTemplateID, err = pdslib.GetStorageTemplate(tenantID)
		Expect(err).NotTo(HaveOccurred())
		log.Infof("storageTemplateID %v", storageTemplateID)
	})

	Step("Create/Get Namespace and NamespaceID", func() {
		namespace = params.InfraToTest.Namespace
		isavailabbe, err := pdslib.CheckNamespace(namespace)
		Expect(err).NotTo(HaveOccurred())
		Expect(isavailabbe).To(BeTrue())
		namespaceID, err = pdslib.GetnameSpaceID(namespace, deploymentTargetID)
		Expect(err).NotTo(HaveOccurred())
		Expect(namespaceID).NotTo(BeEmpty())
	})
})

var _ = AfterSuite(func() {
	defer dash.TestSetEnd()
	defer dash.TestCaseEnd()
})

func TestMain(m *testing.M) {
	// call flag.Parse() here if TestMain uses flags
	ParseFlags()
	os.Exit(m.Run())
}
