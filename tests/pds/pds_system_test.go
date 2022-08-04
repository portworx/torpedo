package tests

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	"github.com/portworx/torpedo/drivers/node"
	"github.com/portworx/torpedo/drivers/pds/api"
	. "github.com/portworx/torpedo/drivers/pds/api"
	. "github.com/portworx/torpedo/drivers/pds/lib"
	. "github.com/portworx/torpedo/tests"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
)

const (
	duration  = 900
	sleepTime = 10

	defaultNumPods = 3
	// FIX-ME : Use Seeded template
	// Create the template manually for all the data serices with below name (i.e QaDefault)
	storageTemplateName          = "QaDefault"
	resourceTemplateName         = "QaDefault"
	appConfigTemplateName        = "QaDefault"
	deploymentName               = "automation"
	templateName                 = "QaDefault"
	defaultWaitRebootTimeout     = 5 * time.Minute
	defaultWaitRebootRetry       = 10 * time.Second
	defaultCommandRetry          = 5 * time.Second
	defaultCommandTimeout        = 1 * time.Minute
	defaultTestConnectionTimeout = 15 * time.Minute
	defaultRebootTimeRange       = 5 * time.Minute
)

var (
	dnsZone                string
	err                    error
	env                    Environment
	ctx                    context.Context
	components             *Components
	controlPlane           *ControlPlane
	apiClient              *pds.APIClient
	targetCluster          *TargetCluster
	account                pds.ModelsAccount
	tenant                 pds.ModelsTenant
	project                pds.ModelsProject
	ns                     *v1.Namespace
	serviceType            = "LoadBalancer"
	pdsDeploymentNamesapce = "qa-deployment"
	supportedDataServices  = map[string]string{"cas": "Cassandra", "zk": "ZooKeeper", "kf": "Kafka", "rmq": "RabbitMQ", "pg": "PostgreSQL"}
)

func TestPDS(t *testing.T) {
	RegisterFailHandler(Fail)

	var specReporters []Reporter
	junitReporter := reporters.NewJUnitReporter("/testresults/junit_basic.xml")
	specReporters = append(specReporters, junitReporter)
	RunSpecsWithDefaultAndCustomReporters(t, "Torpedo : Basic", specReporters)
}

var _ = BeforeSuite(func() {
	InitInstance()
})

var _ = Describe("{RebootOneNode}", func() {

	JustBeforeEach(func() {
		log.Info("Check for environmental variable.")
		env = MustHaveEnvVariables()

		endpointURL, err := url.Parse(env.PDSControlPlaneURL)
		if err != nil {
			log.Panicf("Unable to access the URL: %s", env.PDSControlPlaneURL)
		}
		apiConf := pds.NewConfiguration()
		apiConf.Host = endpointURL.Host
		apiConf.Scheme = endpointURL.Scheme
		token, err := GetBearerToken()
		ctx := context.WithValue(context.Background(), pds.ContextAPIKeys, map[string]pds.APIKey{"ApiKeyAuth": {Key: token, Prefix: "Bearer"}})
		apiClient := pds.NewAPIClient(apiConf)
		components = api.NewComponents(ctx, apiClient)

		log.Info("Get control plane.")
		controlPlane = NewControlPlane(env.PDSControlPlaneURL, components)

		log.Info("Get target plane.")
		targetCluster = NewTargetCluster(env.PDSTargetKUBECONFIG)

		accounts, _ := components.Account.GetAccountsList()

		for i := 0; i < len(accounts); i++ {
			log.Infof("Account Name: %v", accounts[i].GetName())
			if accounts[i].GetName() == env.PDSTestAccountName {
				account = accounts[i]
			}
		}
		log.Infof("Account Detail- Name: %s", account.GetName())
		tnts := components.Tenant
		tenants, _ := tnts.GetTenantsList(account.GetId())
		tenant = tenants[0]
		log.Infof("Tenant Details- Name: %s", tenant.GetName())
		projcts := components.Project
		projects, _ := projcts.GetprojectsList(tenant.GetId())
		project = projects[0]
		log.Infof("Project Details- Name: %s", project.GetName())

		if strings.EqualFold(env.PDSTargetClusterType, "onprem") || strings.EqualFold(env.PDSTargetClusterType, "ocp") {
			serviceType = "ClusterIP"
		}
		log.Infof("Deployment service type %s", serviceType)

	})

	It("has to deploy data service and reboot node(s) while the data services will be running.", func() {
		Step("Register target cluster to PDS control plane and Verify the CRD and other K8s objects become up and healthy in the pds-system namespace", func() {
			log.Info("Get helm version")
			version, _ := components.APIVersion.GetHelmChartVersion()
			log.Infof("Helm chart Version : %s ", version)

			token, err := controlPlane.GetRegistrationToken(tenant.GetId())
			Expect(err).NotTo(HaveOccurred())

			err = targetCluster.RegisterToControlPlane(env.PDSControlPlaneURL, version, token, tenant.GetId(), env.PDSTargetClusterType)
			Expect(err).NotTo(HaveOccurred())
		})
		Step("Create a namespace for data service deployment and add the label for PDS to identify the namespace.", func() {
			ns, err = targetCluster.CreateNamespace(pdsDeploymentNamesapce)
			Expect(err).NotTo(HaveOccurred())
		})
		Step("Deploy a data service and verify the deployment status using the API as well as with the target cluster by means of PODS health.", func() {

			var (
				deploymentTargetID, storageTemplateID   string
				deploymentTargetComponent               = components.DeploymentTarget
				nsComponent                             = components.Namespace
				storagetemplateComponent                = components.StorageSettingsTemplate
				resourceTemplateComponent               = components.ResourceSettingsTemplate
				dataServiceComponent                    = components.DataService
				versionComponent                        = components.Version
				imageComponent                          = components.Image
				appConfigComponent                      = components.AppConfigTemplate
				dataServiceDefaultResourceTemplateIDMap = make(map[string]string)
				dataServiceNameIDMap                    = make(map[string]string)
				dataServiceNameVersionMap               = make(map[string][]string)
				dataServiceIDImagesMap                  = make(map[string]string)
				dataServiceNameDefaultAppConfigMap      = make(map[string]string)
				deployementIDNameMap                    = make(map[string]string)
				namespaceNameIDMap                      = make(map[string]string)
			)

			clusterID, err := targetCluster.GetClusterID()
			Expect(err).NotTo(HaveOccurred())

			log.Info("Get the Target cluster details")
			targetClusters, err := deploymentTargetComponent.ListDeploymentTargetsBelongsToTenant(tenant.GetId())
			Expect(err).NotTo(HaveOccurred())

			for i := 0; i < len(targetClusters); i++ {
				if targetClusters[i].GetClusterId() == clusterID {
					deploymentTargetID = targetClusters[i].GetId()
					log.Infof("Cluster ID: %v, Name: %v,Status: %v", targetClusters[i].GetClusterId(), targetClusters[i].GetName(), targetClusters[i].GetStatus())
				}
			}

			log.Infof("Get the available namespaces in the Cluster having Id: %v", clusterID)
			namespaces, err := nsComponent.ListNamespaces(deploymentTargetID)
			Expect(err).NotTo(HaveOccurred())

			for i := 0; i < len(namespaces); i++ {
				if namespaces[i].GetStatus() == "available" {
					namespaceNameIDMap[namespaces[i].GetName()] = namespaces[i].GetId()
					log.Infof("Available namespace - Name: %v , Id: %v , Status: %v", namespaces[i].GetName(), namespaces[i].GetId(), namespaces[i].GetStatus())
				}
			}

			log.Info("Fetching the storage template")
			storageTemplates, _ := storagetemplateComponent.ListTemplates(tenant.GetId())
			for i := 0; i < len(storageTemplates); i++ {
				if storageTemplates[i].GetName() == storageTemplateName {
					log.Infof("Storage template details -----> Name %v,Repl %v , Fg %v , Fs %v",
						storageTemplates[i].GetName(),
						storageTemplates[i].GetRepl(),
						storageTemplates[i].GetFg(),
						storageTemplates[i].GetFs())
					storageTemplateID = storageTemplates[i].GetId()
					log.Infof("Storage Id: %v", storageTemplateID)
				}
			}

			log.Info("Get the resource template for each data services")
			resourceTemplates, _ := resourceTemplateComponent.ListTemplates(tenant.GetId())
			for i := 0; i < len(resourceTemplates); i++ {
				if resourceTemplates[i].GetName() == resourceTemplateName {
					dataService, _ := dataServiceComponent.GetDataService(resourceTemplates[i].GetDataServiceId())
					log.Infof("Data service name: %v", dataService.GetName())
					log.Infof("Resource template details ---> Name %v, Id : %v ,DataServiceId %v , StorageReq %v , Memoryrequest %v",
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

			log.Info("Fetching the Versions.")
			for key := range dataServiceNameIDMap {
				versions, _ := versionComponent.ListDataServiceVersions(dataServiceNameIDMap[key])
				for i := 0; i < len(versions); i++ {
					dataServiceNameVersionMap[key] = append(dataServiceNameVersionMap[key], versions[i].GetId())
				}
			}

			for key := range dataServiceNameVersionMap {
				images, _ := imageComponent.ListImages(dataServiceNameVersionMap[key][0])
				for i := 0; i < len(images); i++ {
					dataServiceIDImagesMap[images[i].GetDataServiceId()] = images[i].GetId()
				}
			}

			log.Info("Get the Application configuration template")
			appConfigs, _ := appConfigComponent.ListTemplates(tenant.GetId())
			for i := 0; i < len(appConfigs); i++ {
				if appConfigs[i].GetName() == appConfigTemplateName {
					for key := range dataServiceNameIDMap {
						if dataServiceNameIDMap[key] == appConfigs[i].GetDataServiceId() {
							dataServiceNameDefaultAppConfigMap[key] = appConfigs[i].GetId()
						}
					}
				}

			}

			for key := range dataServiceNameIDMap {
				log.Infof("DS name- %v,id- %v", key, dataServiceNameIDMap[key])
			}

			for key := range dataServiceDefaultResourceTemplateIDMap {
				log.Infof("DS Res template name- %v,id- %v", key, dataServiceDefaultResourceTemplateIDMap[key])
			}
			for key := range dataServiceIDImagesMap {
				log.Infof("DS Image name- %v,id- %v", key, dataServiceIDImagesMap[key])
			}

			for key := range namespaceNameIDMap {
				log.Infof("namespace name- %v,id- %v", key, namespaceNameIDMap[key])
			}

			log.Info("Create dataservices without backup.")
			for i := range supportedDataServices {
				log.Infof("Key: %v, Value %v", supportedDataServices[i], dataServiceNameDefaultAppConfigMap[supportedDataServices[i]])
				namespace := pdsDeploymentNamesapce
				namespaceID := namespaceNameIDMap[namespace]
				log.Infof("Created %v deployment  in the namespace %v", supportedDataServices[i], namespace)
				log.Infof(`Request params: 
					project.GetId()- %v deploymentTargetID - %v, 
					dnsZone - %v,deploymentName-%v,namespaceID - %v
					App config ID - %v, ImageId - %v
					num pods- 3, service-type - %v
					Resource template id - %v, storageTemplateID-%v`,
					project.GetId(), deploymentTargetID, dnsZone, deploymentName, namespaceID, dataServiceNameDefaultAppConfigMap[supportedDataServices[i]],
					dataServiceIDImagesMap[dataServiceNameIDMap[supportedDataServices[i]]], serviceType, dataServiceDefaultResourceTemplateIDMap[supportedDataServices[i]], storageTemplateID)

				deployment, _ :=
					components.DataServiceDeployment.CreateDeployment(project.GetId(),
						deploymentTargetID,
						dnsZone,
						deploymentName,
						namespaceID,
						dataServiceNameDefaultAppConfigMap[supportedDataServices[i]],
						dataServiceIDImagesMap[dataServiceNameIDMap[supportedDataServices[i]]],
						3,
						serviceType,
						dataServiceDefaultResourceTemplateIDMap[supportedDataServices[i]],
						storageTemplateID,
					)

				status, _ := components.DataServiceDeployment.GetDeploymentSatus(deployment.GetId())
				sleeptime := 0
				for status.GetHealth() != "Healthy" && sleeptime < duration {
					if sleeptime > 30 && len(status.GetHealth()) < 2 {
						log.Infof("Deployment details: Health status -  %v, procceeding with next deployment", status.GetHealth())
						break
					}
					time.Sleep(10 * time.Second)
					sleeptime += 10
					status, _ = components.DataServiceDeployment.GetDeploymentSatus(deployment.GetId())
					log.Infof("Health status -  %v", status.GetHealth())
				}
				if status.GetHealth() == "Healthy" {
					deployementIDNameMap[deployment.GetId()] = deployment.GetName()
				}
				log.Infof("Deployment details: Health status -  %v,Replicas - %v, Ready replicas - %v", status.GetHealth(), status.GetReplicas(), status.GetReadyReplicas())

			}
		})
		Step("Reboot nodes", func() {
			nodesToReboot := node.GetWorkerNodes()
			for _, n := range nodesToReboot {
				Step(fmt.Sprintf("reboot node: %s", n.Name), func() {
					err = Inst().N.RebootNode(n, node.RebootNodeOpts{
						Force: true,
						ConnectionOpts: node.ConnectionOpts{
							Timeout:         defaultCommandTimeout,
							TimeBeforeRetry: defaultCommandRetry,
						},
					})
					Expect(err).NotTo(HaveOccurred())
				})

				Step(fmt.Sprintf("wait for node: %s to be back up", n.Name), func() {
					err = Inst().N.TestConnection(n, node.ConnectionOpts{
						Timeout:         defaultTestConnectionTimeout,
						TimeBeforeRetry: defaultWaitRebootRetry,
					})
					Expect(err).NotTo(HaveOccurred())
				})
			}

		})

		Step("Verify if all the existing data service deployments are in healthy state.", func() {
			err = targetCluster.ValidatePDSComponents()
			Expect(err).NotTo(HaveOccurred())
		})

	})

})

func TestMain(m *testing.M) {
	// call flag.Parse() here if TestMain uses flags
	ParseFlags()
	os.Exit(m.Run())
}
