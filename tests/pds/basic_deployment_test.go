package tests

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
	pdslib "github.com/portworx/torpedo/drivers/pds/lib"
	. "github.com/portworx/torpedo/tests"
	"github.com/sirupsen/logrus"
)

const (
	deploymentName = "qa"
	replicas       = int32(3)
)

var (
	namespace                          string
	tenantID                           string
	dnsZone                            string
	projectID                          string
	serviceType                        string
	deploymentTargetID                 string
	dsVersion                          string
	dsBuild                            string
	supportedDataServices              []string
	dataServiceIDImagesMap             map[string][]string
	dataServiceNameDefaultAppConfigMap map[string]string
)

func TestBasicDeployment(t *testing.T) {
	RegisterFailHandler(Fail)

	var specReporters []Reporter
	junitReporter := reporters.NewJUnitReporter("/testresults/junit_basic.xml")
	specReporters = append(specReporters, junitReporter)
	RunSpecsWithDefaultAndCustomReporters(t, "Torpedo : pds", specReporters)

}

var _ = BeforeSuite(func() {
	logrus.Info("Starting the test in before suite...")
	Step("get prerequisite params to run the pds tests")
	tenantID, dnsZone, projectID, serviceType, deploymentTargetID = pdslib.SetupPDSTest()
	//InitInstance()
})

//This test deploys dataservices and validates the health and cleans up the deployed dataservice

var _ = Describe("{DeployDataService}", func() {
	It("deploy dataservcies", func() {
		storageTemplateID := pdslib.GetStorageTemplate(tenantID)
		logrus.Infof("storageTemplateID %v", storageTemplateID)

		namespace = pdslib.GetAndExpectStringEnvVar("NAMESPACE")
		namespaceID := pdslib.GetnameSpaceID(namespace)

		//supportedDataServices := map[string]string{"pg": "PostgreSQL"}
		supportedDataServices = append(supportedDataServices, pdslib.GetAndExpectStringEnvVar("DATA_SERVICE"))

		dataServiceDefaultResourceTemplateIDMap, dataServiceNameIDMap := pdslib.GetResourceTemplate(tenantID, supportedDataServices)
		deployallDataServiceVersion := pdslib.GetAndExpectStringEnvVar("DEPLOY_ALL_VERSIONS")
		if deployallDataServiceVersion != "true" {
			dsVersion = pdslib.GetAndExpectStringEnvVar("DS_VERSION")
			dsBuild = pdslib.GetAndExpectStringEnvVar("DS_BUILD")
			_, dataServiceIDImagesMap = pdslib.GetVersions(dsVersion, dsBuild, dataServiceNameIDMap)
		} else {
			_, dataServiceIDImagesMap = pdslib.GetAllVersions(dataServiceNameIDMap)
		}
		dataServiceNameDefaultAppConfigMap = pdslib.GetAppConfTemplate(tenantID)

		logrus.Info("Create dataservices without backup.")
		for i := range supportedDataServices {
			logrus.Infof("Key: %v, Value %v", supportedDataServices[i], dataServiceNameDefaultAppConfigMap[supportedDataServices[i]])
			logrus.Infof(`Request params: 
			projectID- %v deploymentTargetID - %v, 
			dnsZone - %v,deploymentName - %v,namespaceID - %v
			App config ID - %v,
			num pods- 3, service-type - %v
			Resource template id - %v, storageTemplateID - %v`,
				projectID, deploymentTargetID, dnsZone, deploymentName, namespaceID, dataServiceNameDefaultAppConfigMap[supportedDataServices[i]],
				serviceType, dataServiceDefaultResourceTemplateIDMap[supportedDataServices[i]], storageTemplateID)
			deployment := pdslib.DeployDataServices(projectID,
				deploymentTargetID,
				dnsZone,
				deploymentName,
				namespaceID,
				dataServiceNameDefaultAppConfigMap[supportedDataServices[i]],
				dataServiceIDImagesMap,
				replicas,
				serviceType,
				dataServiceDefaultResourceTemplateIDMap[supportedDataServices[i]],
				storageTemplateID,
			)
			logrus.Infof("data service id %v", deployment.GetDataServiceId())
		}
	})
})

func TestMain(m *testing.M) {
	// call flag.Parse() here if TestMain uses flags
	ParseFlags()
	os.Exit(m.Run())
}
