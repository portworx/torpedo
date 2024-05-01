package tests

import (
	"fmt"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/stworkflows/platform"
	"github.com/portworx/torpedo/drivers/utilities"
	"os"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	pdslib "github.com/portworx/torpedo/drivers/pds/lib"
	dsUtils "github.com/portworx/torpedo/drivers/unifiedPlatform/pdsLibs"
	platformUtils "github.com/portworx/torpedo/drivers/unifiedPlatform/platformLibs"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests"
	. "github.com/portworx/torpedo/tests/unifiedPlatform"
)

var _ = BeforeSuite(func() {
	steplog := "Get prerequisite params to run platform tests"

	log.InfoD(steplog)
	Step(steplog, func() {
		InitInstance()
		dash = Inst().Dash
		dash.TestSet.Product = "pds"
		dash.TestSetBegin(dash.TestSet)
		// Read pds params from the configmap
		var err error
		pdsparams := pdslib.GetAndExpectStringEnvVar("PDS_PARAM_CM")
		NewPdsParams, err = ReadNewParams(pdsparams)
		log.FailOnError(err, "Failed to read params from json file")
		infraParams := NewPdsParams.InfraToTest
		PdsLabels["clusterType"] = infraParams.ClusterType

		log.InfoD("Get Account ID")
		//TODO: Get the accountID
		AccID = "acc:e0886df7-aa22-4090-a6ea-9f08a0097f99"

		err = platformUtils.InitUnifiedApiComponents(os.Getenv(EnvControlPlaneUrl), AccID)
		log.FailOnError(err, "error while initialising api components")

		// accList, err := platformUtils.GetAccountListv1()
		// log.FailOnError(err, "error while getting account list")
		// accID = platformUtils.GetPlatformAccountID(accList, defaultTestAccount)
		log.Infof("AccountID - [%s]", AccID)

		err = platformUtils.InitUnifiedApiComponents(infraParams.ControlPlaneURL, AccID)
		log.FailOnError(err, "error while initialising api components")

		//Initialising UnifiedApiComponents in ds utils
		err = dsUtils.InitUnifiedApiComponents(infraParams.ControlPlaneURL, AccID)
		log.FailOnError(err, "error while initialising api components in ds utils")
	})

	Step("Dumping kubeconfigs file", func() {
		kubeconfigs := os.Getenv("KUBECONFIGS")
		if kubeconfigs != "" {
			kubeconfigList := strings.Split(kubeconfigs, ",")
			if len(kubeconfigList) < 2 {
				log.FailOnError(fmt.Errorf("At least minimum two kubeconfigs required but has"),
					"Failed to get k8s config path.At least minimum two kubeconfigs required")
			}
			DumpKubeconfigs(kubeconfigList)
		}
	})

	Step("Get Default Tenant", func() {
		log.Infof("Initialising values for tenant")
		WorkflowPlatform.AdminAccountId = AccID
		_, err := WorkflowPlatform.TenantInit()
		log.FailOnError(err, "error while getting Default TenantId")
	})

	Step("Get Default Project", func() {
		var err error
		DEFAULT_PROJECT_NAME := "pds-project-" + RandomString(5)
		WorkflowProject.Platform = WorkflowPlatform
		WorkflowProject.ProjectName = DEFAULT_PROJECT_NAME
		_, err = WorkflowProject.CreateProject()
		log.FailOnError(err, "unable to create project")
		ProjectId, err = WorkflowProject.GetDefaultProject(DEFAULT_PROJECT_NAME)
		log.FailOnError(err, "Unable to get current project")
		log.Infof("Current project ID - [%s]", ProjectId)
	})

	Step("Register Target Cluster and Install PDS app", func() {
		WorkflowTargetCluster.Project = WorkflowProject
		log.Infof("Tenant ID [%s]", WorkflowTargetCluster.Project.Platform.TenantId)
		WorkflowTargetCluster, err := WorkflowTargetCluster.RegisterToControlPlane(false)
		log.FailOnError(err, "Unable to register target cluster")
		log.Infof("Target cluster registered with uid - [%s]", WorkflowTargetCluster.ClusterUID)

		err = WorkflowTargetCluster.InstallPDSAppOnTC(WorkflowTargetCluster.ClusterUID)
		log.FailOnError(err, "Unable to Install pds on target cluster")
	})

	Step("Register Destination target Cluster", func() {

		defer func() {
			err := SetSourceKubeConfig()
			log.FailOnError(err, "failed to switch context to source cluster")
		}()

		err := SetDestinationKubeConfig()
		log.FailOnError(err, "Failed to switched to destination cluster")

		WorkflowTargetClusterDestination.Project = WorkflowProject
		log.Infof("Tenant ID [%s]", WorkflowTargetClusterDestination.Project.Platform.TenantId)
		WorkflowTargetClusterDestination, err := WorkflowTargetClusterDestination.RegisterToControlPlane(false)
		log.FailOnError(err, "Unable to register target cluster")
		log.Infof("Destination Target cluster registered with uid - [%s]", WorkflowTargetCluster.ClusterUID)
		err = WorkflowTargetClusterDestination.InstallPDSAppOnTC(WorkflowTargetCluster.ClusterUID)
		log.FailOnError(err, "Unable to Install pds on destination target cluster")
	})

	Step("Create Buckets", func() {
		if NewPdsParams.BackUpAndRestore.RunBkpAndRestrTest {
			PDSBucketName = strings.ToLower("pds-test-buck-" + utilities.RandString(5))
			switch NewPdsParams.BackUpAndRestore.TargetLocation {
			case "s3-comp":
				err := platformUtils.CreateS3CompBucket(PDSBucketName)
				log.FailOnError(err, "error while creating s3-comp bucket")
			case "s3":
				err := platformUtils.CreateS3Bucket(PDSBucketName)
				log.FailOnError(err, "error while creating s3 bucket")
			case "azure":
				err := platformUtils.CreateAzureBucket(PDSBucketName)
				log.FailOnError(err, "error while creating azure bucket")
			default:
				err := platformUtils.CreateS3CompBucket(PDSBucketName)
				log.FailOnError(err, "error while creating s3-comp bucket")
			}
		}
	})

	Step("Create Cloud Credential and BackUpLocation", func() {
		log.Debugf("TenantId [%s]", WorkflowTargetCluster.Project.Platform.TenantId)
		WorkflowCc.Platform = WorkflowPlatform
		WorkflowCc.CloudCredentials = make(map[string]platform.CloudCredentialsType)
		cc, err := WorkflowCc.CreateCloudCredentials(NewPdsParams.BackUpAndRestore.TargetLocation)
		log.FailOnError(err, "error occured while creating cloud credentials")
		for _, value := range cc.CloudCredentials {
			log.Infof("cloud credentials name: [%s]", value.Name)
			log.Infof("cloud credentials id: [%s]", value.ID)
			log.Infof("cloud provider type: [%s]", value.CloudProviderType)
		}

		WorkflowbkpLoc.WfCloudCredentials = WorkflowCc
		wfbkpLoc, err := WorkflowbkpLoc.CreateBackupLocation(PDSBucketName, NewPdsParams.BackUpAndRestore.TargetLocation)
		log.FailOnError(err, "error while creating backup location")
		log.Infof("wfBkpLoc id: [%s]", wfbkpLoc.BkpLocation.BkpLocationId)
		log.Infof("wfBkpLoc name: [%s]", wfbkpLoc.BkpLocation.Name)
	})

	Step("Associate platform resources to Project", func() {
		err := WorkflowProject.Associate(
			[]string{WorkflowTargetCluster.ClusterUID, WorkflowTargetClusterDestination.ClusterUID},
			[]string{},
			[]string{WorkflowCc.CloudCredentials[NewPdsParams.BackUpAndRestore.TargetLocation].ID},
			[]string{WorkflowbkpLoc.BkpLocation.BkpLocationId},
			[]string{},
			[]string{},
		)
		log.FailOnError(err, "Unable to associate platform resources to Project")
		log.Infof("Associated Resources - [%+v]", WorkflowProject.AssociatedResources)
	})

})

var _ = AfterSuite(func() {

	Step("Purging all platform related objects", func() {

		// TODO: This needs to be added back once cleanup issues are fixed
		log.Warnf("Skipping Platform resource cleanup")
		//log.InfoD("Deleting projects")
		//err := WorkflowProject.DeleteProject()
		//log.FailOnError(err, "unable to delete projects")
	})

	EndTorpedoTest()
	//TODO: Steps to delete Backup location, Target and Bucket
	// TODO: Add namespace cleanup once deployment cleanup cleans up the services too
	//err := WorkflowNamespace.Purge()
	//log.FailOnError(err, "Unable to cleanup all namespaces")
	//log.InfoD("All namespaces cleaned up successfully")
	log.InfoD("Test Finished")
})

func TestDataService(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Torpedo : pds")
}

func TestMain(m *testing.M) {
	// call flag.Parse() here if TestMain uses flags
	ParseFlags()
	os.Exit(m.Run())
}
