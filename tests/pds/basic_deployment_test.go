package tests

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	pdslib "github.com/portworx/torpedo/drivers/pds/lib"
	. "github.com/portworx/torpedo/tests"
	"github.com/sirupsen/logrus"
)

const (
	deploymentName       = "qa"
	envDsVersion         = "DS_VERSION"
	envDsBuild           = "DS_BUILD"
	envReplicas          = "NO_OF_NODES"
	envNamespace         = "NAMESPACE"
	envDataService       = "DATA_SERVICE"
	envDeployAllVersions = "DEPLOY_ALL_VERSIONS"
)

var (
	namespace                               string
	tenantID                                string
	dnsZone                                 string
	projectID                               string
	serviceType                             string
	deploymentTargetID                      string
	dsVersion                               string
	dsBuild                                 string
	replicas                                int32
	supportedDataServices                   []string
	dataServiceIDImagesMap                  map[string][]string
	dataServiceNameDefaultAppConfigMap      map[string]string
	namespaceID                             string
	storageTemplateID                       string
	dataServiceDefaultResourceTemplateIDMap map[string]string
	dataServiceNameIDMap                    map[string]string
	deploymentIDs                           []string
	deployment                              *pds.ModelsDeployment
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

	supportedDataServices = append(supportedDataServices, pdslib.GetAndExpectStringEnvVar(envDataService))

	Step("Get StorageTemplateID, App ConfigID, ResourceTemplateID, Replicas and Supported Dataservice", func() {
		storageTemplateID = pdslib.GetStorageTemplate(tenantID)
		logrus.Infof("storageTemplateID %v", storageTemplateID)

		replicas = int32(pdslib.GetAndExpectIntEnvVar(envReplicas))

		dataServiceDefaultResourceTemplateIDMap, dataServiceNameIDMap = pdslib.GetResourceTemplate(tenantID, supportedDataServices)
		dataServiceNameDefaultAppConfigMap = pdslib.GetAppConfTemplate(tenantID, dataServiceNameIDMap)

	})

	Step("Get Versions and Builds of Dataservice and form supported dataServiceIDImagesMap", func() {
		deployallDataServiceVersion := pdslib.GetAndExpectStringEnvVar(envDeployAllVersions)
		if deployallDataServiceVersion != "true" {
			dsVersion = pdslib.GetAndExpectStringEnvVar(envDsVersion)
			dsBuild = pdslib.GetAndExpectStringEnvVar(envDsBuild)
			logrus.Infof("Getting versionID  for Data service version %s and buildID for %s ", dsVersion, dsBuild)
			_, dataServiceIDImagesMap = pdslib.GetVersions(dsVersion, dsBuild, dataServiceNameIDMap)
		} else {
			_, dataServiceIDImagesMap = pdslib.GetAllVersions(dataServiceNameIDMap)
		}

	})

	//InitInstance()
})

//This test deploys dataservices and validates the health and cleans up the deployed dataservice

var _ = Describe("{DeployDataService}", func() {
	JustBeforeEach(func() {
		Step("Get Namespace and NamespaceID", func() {
			namespace = pdslib.GetAndExpectStringEnvVar(envNamespace)
			namespaceID = pdslib.GetnameSpaceID(namespace)
		})
	})

	It("deploy dataservcies", func() {
		logrus.Info("Create dataservices without backup.")
		for i := range supportedDataServices {
			logrus.Infof("Key: %v, Value %v", supportedDataServices[i], dataServiceNameDefaultAppConfigMap[supportedDataServices[i]])
			logrus.Infof(`Request params: 
			projectID- %v deploymentTargetID - %v, 
			dnsZone - %v,deploymentName - %v,namespaceID - %v
			App config ID - %v,
			num pods- %v, service-type - %v
			Resource template id - %v, storageTemplateID - %v`,
				projectID, deploymentTargetID, dnsZone, deploymentName, namespaceID, dataServiceNameDefaultAppConfigMap[supportedDataServices[i]],
				replicas, serviceType, dataServiceDefaultResourceTemplateIDMap[supportedDataServices[i]], storageTemplateID)
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

	It("trigger Workload on dataServices", func() {
		for i := range supportedDataServices {
			Step("Deploy Dataservices", func() {
				logrus.Infof("Key: %v, Value %v", supportedDataServices[i], dataServiceNameDefaultAppConfigMap[supportedDataServices[i]])
				logrus.Infof(`Request params: 
			projectID- %v deploymentTargetID - %v, 
			dnsZone - %v,deploymentName - %v,namespaceID - %v
			App config ID - %v,
			num pods- %v, service-type - %v
			Resource template id - %v, storageTemplateID - %v`,
					projectID, deploymentTargetID, dnsZone, deploymentName, namespaceID, dataServiceNameDefaultAppConfigMap[supportedDataServices[i]],
					replicas, serviceType, dataServiceDefaultResourceTemplateIDMap[supportedDataServices[i]], storageTemplateID)
				deployment = pdslib.DeployDataServices(projectID,
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
				deploymentIDs = append(deploymentIDs, deployment.GetId())
			})
		}
		Step("Trigger workload generation for PostgreSQL")

	})

	It("scaleUp dataservices", func() {
		logrus.Info("Scale Test for dataservices")
		for i := range supportedDataServices {
			Step("Deploy Dataservices", func() {
				logrus.Infof("Key: %v, Value %v", supportedDataServices[i], dataServiceNameDefaultAppConfigMap[supportedDataServices[i]])
				logrus.Infof(`Request params: 
			projectID- %v deploymentTargetID - %v, 
			dnsZone - %v,deploymentName - %v,namespaceID - %v
			App config ID - %v,
			num pods- %v, service-type - %v
			Resource template id - %v, storageTemplateID - %v`,
					projectID, deploymentTargetID, dnsZone, deploymentName, namespaceID, dataServiceNameDefaultAppConfigMap[supportedDataServices[i]],
					replicas, serviceType, dataServiceDefaultResourceTemplateIDMap[supportedDataServices[i]], storageTemplateID)
				deployment = pdslib.DeployDataServices(projectID,
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
				deploymentIDs = append(deploymentIDs, deployment.GetId())
			})
		}

		Step("Scaling up the dataservice replicas", func() {
			replicas = 2
			for index := range deploymentIDs {
				deployment = pdslib.UpdateDataServices(deploymentIDs[index],
					dataServiceNameDefaultAppConfigMap[supportedDataServices[index]],
					dataServiceIDImagesMap,
					replicas,
					dataServiceDefaultResourceTemplateIDMap[supportedDataServices[index]],
				)
			}
		})
	})
})

func TestMain(m *testing.M) {
	// call flag.Parse() here if TestMain uses flags
	ParseFlags()
	os.Exit(m.Run())
}
