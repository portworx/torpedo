package tests

import (
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
	PDS_DEFAULT_NAMESPACE = "pds-namespace-" + RandomString(5)
	steplog := "Get prerequisite params to run platform tests"
	log.InfoD(steplog)
	Step(steplog, func() {
		InitInstance()
		dash = Inst().Dash
		// Read pds params from the configmap
		var err error
		pdsparams := pdslib.GetAndExpectStringEnvVar("PDS_PARAM_CM")
		NewPdsParams, err = ReadNewParams(pdsparams)
		log.FailOnError(err, "Failed to read params from json file")
		infraParams := NewPdsParams.InfraToTest
		PdsLabels["clusterType"] = infraParams.ClusterType

		log.InfoD("Get Account ID")
		//TODO: Get the accountID
		AccID = "acc:8b6e5023-2ec9-474f-acda-7ab662987409"

		err = platformUtils.InitUnifiedApiComponents(os.Getenv(EnvControlPlaneUrl), "")
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

	Step("Get Default Tenant", func() {
		log.Infof("Initialising values for tenant")
		WorkflowPlatform.AdminAccountId = AccID
		_, err := WorkflowPlatform.TenantInit()
		log.FailOnError(err, "error while getting Default TenantId")
	})

	Step("Get Default Project", func() {
		var err error
		WorkflowProject.Platform = WorkflowPlatform
		ProjectId, err = WorkflowProject.GetDefaultProject(DefaultProject)
		log.FailOnError(err, "Unable to get default project")
		log.Infof("Default project ID - [%s]", ProjectId)
		WorkflowProject.ProjectId = ProjectId
		WorkflowProject.ProjectName = DefaultProject
	})

	Step("Register Target Cluster", func() {
		WorkflowTargetCluster.Project = WorkflowProject
		log.Infof("Tenant ID [%s]", WorkflowTargetCluster.Project.Platform.TenantId)
		WorkflowTargetCluster, err := WorkflowTargetCluster.RegisterToControlPlane(false)
		log.FailOnError(err, "Unable to register target cluster")
		log.Infof("Target cluster registered with uid - [%s]", WorkflowTargetCluster.ClusterUID)
	})

	Step("Create a namespace for PDS", func() {
		WorkflowNamespace.TargetCluster = WorkflowTargetCluster
		WorkflowNamespace.Namespaces = make(map[string]string)
		_, err := WorkflowNamespace.CreateNamespaces(PDS_DEFAULT_NAMESPACE)
		log.FailOnError(err, "Unable to create namespace")
		log.Infof("Namespaces created - [%s]", WorkflowNamespace.Namespaces)
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

	Step("Associate namespace and cluster to Project", func() {
		err := WorkflowProject.Associate(
			[]string{WorkflowTargetCluster.ClusterUID},
			[]string{WorkflowNamespace.Namespaces[PDS_DEFAULT_NAMESPACE]},
			[]string{WorkflowCc.CloudCredentials[NewPdsParams.BackUpAndRestore.TargetLocation].ID},
			[]string{WorkflowbkpLoc.BkpLocation.BkpLocationId},
			[]string{},
			[]string{},
		)
		log.FailOnError(err, "Unable to associate Cluster to Project")
		log.Infof("Associated Resources - [%+v]", WorkflowProject.AssociatedResources)
	})

})

var _ = AfterSuite(func() {
	//TODO: Steps to delete Backup location, Target and Bucket
	err := WorkflowNamespace.Purge()
	log.FailOnError(err, "Unable to cleanup all namespaces")
	log.InfoD("All namespaces cleaned up successfully")

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
