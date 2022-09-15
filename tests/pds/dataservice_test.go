package tests

import (
	"os"
	"strconv"
	"testing"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
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
	dataServiceNameDefaultAppConfigMap      map[string]string
	namespaceID                             string
	storageTemplateID                       string
	dataServiceDefaultResourceTemplateIDMap map[string]string
	dataServiceNameIDMap                    map[string]string
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
			deployements, err := pdslib.DeployDataServices(dataServiceNameIDMap, projectID,
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
			Expect(err).NotTo(HaveOccurred())
			Step("Validate Storage Configurations", func() {
				for ds, deployment := range deployements {
					for index := range deployment {
						logrus.Infof("data service deployed %v ", ds)
						resourceTemp, storageOp, config := pdslib.ValidateDataServiceVolumes(deployment[index], ds, dataServiceDefaultResourceTemplateIDMap, storageTemplateID)
						logrus.Infof("filesystem used %v ", config.Spec.StorageOptions.Filesystem)
						logrus.Infof("storage replicas used %v ", config.Spec.StorageOptions.Replicas)
						logrus.Infof("cpu requests used %v ", config.Spec.Resources.Requests.CPU)
						logrus.Infof("memory requests used %v ", config.Spec.Resources.Requests.Memory)
						logrus.Infof("storage requests used %v ", config.Spec.Resources.Requests.Storage)
						logrus.Infof("No of nodes requested %v ", config.Spec.Nodes)
						logrus.Infof("volume group %v ", storageOp.VolumeGroup)

						Expect(resourceTemp.Resources.Requests.CPU).Should(Equal(config.Spec.Resources.Requests.CPU))
						Expect(resourceTemp.Resources.Requests.Memory).Should(Equal(config.Spec.Resources.Requests.Memory))
						Expect(resourceTemp.Resources.Requests.Storage).Should(Equal(config.Spec.Resources.Requests.Storage))
						Expect(resourceTemp.Resources.Limits.CPU).Should(Equal(config.Spec.Resources.Limits.CPU))
						Expect(resourceTemp.Resources.Limits.Memory).Should(Equal(config.Spec.Resources.Limits.Memory))
						repl, err := strconv.Atoi(config.Spec.StorageOptions.Replicas)
						Expect(err).NotTo(HaveOccurred())
						Expect(storageOp.Replicas).Should(Equal(int32(repl)))
						Expect(storageOp.Filesystem).Should(Equal(config.Spec.StorageOptions.Filesystem))
						Expect(config.Spec.Nodes).Should(Equal(replicas))

					}
				}
			})
			defer func() {
				Step("Delete created deployments")
				for _, dep := range deployements {
					for index := range dep {
						_, err := pdslib.DeleteDeployment(dep[index].GetId())
						Expect(err).NotTo(HaveOccurred())
					}
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
			deployements, err := pdslib.DeployDataServices(supportedDataServicesNameIDMap, projectID,
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
			Expect(err).NotTo(HaveOccurred())
			Step("Validate Storage Configurations", func() {
				for ds, deployment := range deployements {
					for index := range deployment {
						logrus.Infof("data service deployed %v ", ds)
						resourceTemp, storageOp, config := pdslib.ValidateDataServiceVolumes(deployment[index], ds, dataServiceDefaultResourceTemplateIDMap, storageTemplateID)
						logrus.Infof("filesystem used %v ", config.Spec.StorageOptions.Filesystem)
						logrus.Infof("storage replicas used %v ", config.Spec.StorageOptions.Replicas)
						logrus.Infof("cpu requests used %v ", config.Spec.Resources.Requests.CPU)
						logrus.Infof("memory requests used %v ", config.Spec.Resources.Requests.Memory)
						logrus.Infof("storage requests used %v ", config.Spec.Resources.Requests.Storage)
						logrus.Infof("No of nodes requested %v ", config.Spec.Nodes)
						logrus.Infof("volume group %v ", storageOp.VolumeGroup)

						Expect(resourceTemp.Resources.Requests.CPU).Should(Equal(config.Spec.Resources.Requests.CPU))
						Expect(resourceTemp.Resources.Requests.Memory).Should(Equal(config.Spec.Resources.Requests.Memory))
						Expect(resourceTemp.Resources.Requests.Storage).Should(Equal(config.Spec.Resources.Requests.Storage))
						Expect(resourceTemp.Resources.Limits.CPU).Should(Equal(config.Spec.Resources.Limits.CPU))
						Expect(resourceTemp.Resources.Limits.Memory).Should(Equal(config.Spec.Resources.Limits.Memory))
						repl, err := strconv.Atoi(config.Spec.StorageOptions.Replicas)
						Expect(err).NotTo(HaveOccurred())
						Expect(storageOp.Replicas).Should(Equal(int32(repl)))
						Expect(storageOp.Filesystem).Should(Equal(config.Spec.StorageOptions.Filesystem))
						Expect(config.Spec.Nodes).Should(Equal(replicas))
					}
				}
			})
			defer func() {
				Step("Delete created deployments")
				for _, dep := range deployements {
					for index := range dep {
						_, err := pdslib.DeleteDeployment(dep[index].GetId())
						Expect(err).NotTo(HaveOccurred())
					}
				}
			}()
		})
	})
})

func TestMain(m *testing.M) {
	// call flag.Parse() here if TestMain uses flags
	ParseFlags()
	os.Exit(m.Run())
}
