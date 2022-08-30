package tests

import (
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
	deploymentName          = "qa"
	envDsVersion            = "DS_VERSION"
	envDsBuild              = "DS_BUILD"
	envReplicas             = "NO_OF_NODES"
	envNamespace            = "NAMESPACE"
	envDataService          = "DATA_SERVICE"
	envDeployAllVersions    = "DEPLOY_ALL_VERSIONS"
	envDeployAllDataService = "DEPLOY_ALL_DATASERVICE"
)

var (
	namespace                               string
	tenantID                                string
	dnsZone                                 string
	projectID                               string
	serviceType                             string
	deploymentTargetID                      string
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
	supportedDataServicesNameIDMap          map[string]string
)

func TestDataService(t *testing.T) {
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

	Step("Get StorageTemplateID and Replicas", func() {
		storageTemplateID = pdslib.GetStorageTemplate(tenantID)
		logrus.Infof("storageTemplateID %v", storageTemplateID)
		replicas = int32(pdslib.GetAndExpectIntEnvVar(envReplicas))

	})

	Step("Get Namespace and NamespaceID", func() {
		namespace = pdslib.GetAndExpectStringEnvVar(envNamespace)
		namespaceID = pdslib.GetnameSpaceID(namespace)
	})
})

var _ = Describe("{DeployDataServicesOnDemand}", func() {

	JustBeforeEach(func() {
		if !pdslib.GetAndExpectBoolEnvVar(envDeployAllDataService) {
			supportedDataServices = append(supportedDataServices, pdslib.GetAndExpectStringEnvVar(envDataService))
			for _, ds := range supportedDataServices {
				logrus.Infof("supported dataservices %v", ds)
			}
			Step("Get the resource and app config template for supported dataservice", func() {
				dataServiceDefaultResourceTemplateIDMap, dataServiceNameIDMap = pdslib.GetResourceTemplate(tenantID, supportedDataServices)
				dataServiceNameDefaultAppConfigMap = pdslib.GetAppConfTemplate(tenantID, dataServiceNameIDMap)
			})
		} else {
			Expect(pdslib.GetAndExpectBoolEnvVar(envDeployAllDataService)).To(Equal(true))
		}
	})

	It("deploy Dataservices", func() {
		logrus.Info("Create dataservices without backup.")
		Step("Deploy Data Services", func() {
			deployementIDNameMap := pdslib.DeployDataServices(dataServiceNameIDMap, projectID,
				deploymentTargetID,
				dnsZone,
				deploymentName,
				namespaceID,
				dataServiceNameDefaultAppConfigMap,
				replicas,
				serviceType,
				dataServiceDefaultResourceTemplateIDMap,
				storageTemplateID,
			)
			defer func() {
				Step("Delete created deployments")
				for depID := range deployementIDNameMap {
					logrus.Infof("deplymentID %v ", depID)
					_, err := pdslib.DeleteDeployment(depID)
					Expect(err).NotTo(HaveOccurred())
				}
			}()
		})
	})
})

var _ = Describe("{DeployAllDataServices}", func() {

	JustBeforeEach(func() {
		Step("Check the required env param is available to run this test", func() {
			if !pdslib.GetAndExpectBoolEnvVar(envDeployAllDataService) && pdslib.GetAndExpectBoolEnvVar(envDeployAllVersions) {
				logrus.Fatal("Env Var are not set as expected")
			}
		})
		Step("Get All Supported Dataservices and Versions", func() {
			supportedDataServicesNameIDMap = pdslib.GetAllSupportedDataServices()
			for dsName := range supportedDataServicesNameIDMap {
				supportedDataServices = append(supportedDataServices, dsName)
			}
			for index := range supportedDataServices {
				logrus.Infof("supported data service %v ", supportedDataServices[index])
			}
			Step("Get the resource and app config template for supported dataservice")
			dataServiceDefaultResourceTemplateIDMap, dataServiceNameIDMap = pdslib.GetResourceTemplate(tenantID, supportedDataServices)
			dataServiceNameDefaultAppConfigMap = pdslib.GetAppConfTemplate(tenantID, dataServiceNameIDMap)
		})

	})

	It("Deploy All SupportedDataServices", func() {
		Step("Deploy All Supported Data Services", func() {
			deployementIDNameMap := pdslib.DeployDataServices(supportedDataServicesNameIDMap, projectID,
				deploymentTargetID,
				dnsZone,
				deploymentName,
				namespaceID,
				dataServiceNameDefaultAppConfigMap,
				replicas,
				serviceType,
				dataServiceDefaultResourceTemplateIDMap,
				storageTemplateID,
			)
			defer func() {
				Step("Delete created deployments")
				for depID := range deployementIDNameMap {
					logrus.Infof("deplymentID %v ", depID)
					_, err := pdslib.DeleteDeployment(depID)
					Expect(err).NotTo(HaveOccurred())
				}
			}()
		})
	})
})
