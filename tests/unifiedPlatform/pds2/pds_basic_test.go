package tests

import (
	"fmt"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/stworkflows/platform"
	"github.com/portworx/torpedo/drivers/utilities"
	"os"
	"strings"
	"testing"
	"time"

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
		AccID = "acc:f50fbd20-626a-43c1-8cf1-cb2518770c4e"

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

	steplog = "Dumping kubeconfigs file"
	Step(steplog, func() {
		log.InfoD(steplog)
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

	steplog = "Get Default Tenant"
	Step(steplog, func() {
		log.InfoD(steplog)
		log.Infof("Initialising values for tenant")
		WorkflowPlatform.AdminAccountId = AccID
		_, err := WorkflowPlatform.TenantInit()
		log.FailOnError(err, "error while getting Default TenantId")
	})

	steplog = "Create default project"
	Step(steplog, func() {
		log.InfoD(steplog)
		WorkflowProject.Platform = WorkflowPlatform
		WorkflowProject.ProjectName = fmt.Sprintf("project-%s", utilities.RandomString(5))
		workflowProject, err := WorkflowProject.CreateProject()
		log.FailOnError(err, "Unable to create project")
		log.InfoD("Project created with ID - [%s]", workflowProject.ProjectId)

	})

	steplog = "Register Target Cluster and Install PDS app"
	Step(steplog, func() {
		log.InfoD(steplog)
		WorkflowTargetCluster.Project = &WorkflowProject
		log.Infof("Tenant ID [%s]", WorkflowTargetCluster.Project.Platform.TenantId)
		WorkflowTargetCluster, err := WorkflowTargetCluster.RegisterToControlPlane()
		log.FailOnError(err, "Unable to register target cluster")
		log.Infof("Target cluster registered with uid - [%s]", WorkflowTargetCluster.ClusterUID)

		time.Sleep(1 * time.Minute)

		err = WorkflowTargetCluster.InstallPDSAppOnTC(WorkflowTargetCluster.ClusterUID)
		log.FailOnError(err, "Unable to Install pds on target cluster")
	})

	steplog = "Register Destination target Cluster"
	Step(steplog, func() {
		log.InfoD(steplog)
		defer func() {
			err := SetSourceKubeConfig()
			log.FailOnError(err, "failed to switch context to source cluster")
		}()

		err := SetDestinationKubeConfig()
		log.FailOnError(err, "Failed to switched to destination cluster")

		WorkflowTargetClusterDestination.Project = &WorkflowProject
		log.Infof("Tenant ID [%s]", WorkflowTargetClusterDestination.Project.Platform.TenantId)
		WorkflowTargetClusterDestination, err := WorkflowTargetClusterDestination.RegisterToControlPlane()
		log.FailOnError(err, "Unable to register target cluster")
		log.Infof("Destination Target cluster registered with uid - [%s]", WorkflowTargetCluster.ClusterUID)
		err = WorkflowTargetClusterDestination.InstallPDSAppOnTC(WorkflowTargetCluster.ClusterUID)
		log.FailOnError(err, "Unable to Install pds on destination target cluster")
	})

	steplog = "Create Service Configuration, Resource and Storage Templates"
	Step(steplog, func() {
		log.InfoD(steplog)
		var err error
		WorkflowPDSTemplate.Platform = WorkflowPlatform
		DsNameAndAppTempId, StTemplateId, ResourceTemplateId, err = WorkflowPDSTemplate.CreatePdsCustomTemplatesAndFetchIds(NewPdsParams)
		log.FailOnError(err, "Unable to create Custom Templates for PDS")

		for _, AppTemplateId := range DsNameAndAppTempId {
			TemplateIds = append(TemplateIds, AppTemplateId)
		}
		TemplateIds = append(TemplateIds, ResourceTemplateId, StTemplateId)

		for _, tempId := range TemplateIds {
			log.Debugf("TemplateID: [%s]", tempId)
		}
	})

	steplog = "Associate templates to the Project"
	Step(steplog, func() {
		log.InfoD(steplog)
		err := WorkflowProject.Associate(
			[]string{},
			[]string{},
			[]string{},
			[]string{},
			TemplateIds,
			[]string{},
		)
		log.FailOnError(err, "Unable to associate Templates to Project")
		log.Infof("Associated Resources - [%+v]", WorkflowProject.AssociatedResources)
	})

	if NewPdsParams.BackUpAndRestore.RunBkpAndRestrTest {
		steplog = "Create Buckets"
		Step(steplog, func() {
			log.InfoD(steplog)
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
		})

		steplog = "Create Cloud Credential and BackUpLocation"
		Step(steplog, func() {
			log.InfoD(steplog)
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

		steplog = "Associate platform resources to Project"
		Step(steplog, func() {
			log.InfoD(steplog)
			err := WorkflowProject.Associate(
				[]string{WorkflowTargetCluster.ClusterUID, WorkflowTargetClusterDestination.ClusterUID},
				[]string{},
				[]string{WorkflowCc.CloudCredentials[NewPdsParams.BackUpAndRestore.TargetLocation].ID},
				[]string{WorkflowbkpLoc.BkpLocation.BkpLocationId},
				[]string{},
				[]string{},
			)
			log.FailOnError(err, "Unable to associate Cluster to Project")
			log.Infof("Associated Resources - [%+v]", WorkflowProject.AssociatedResources)
		})

	}

})

var _ = AfterSuite(func() {

	defer Inst().Dash.TestSetEnd()
	defer EndTorpedoTest()

	var allErrors []error

	// TODO: Need to add platform cleanup here
	log.InfoD("Purging all templates")
	err := WorkflowPDSTemplate.Purge(true)
	if err != nil {
		log.Errorf("some error occurred while purging data service templates - [%s]", err.Error())
		allErrors = append(allErrors, err)
	}

	log.InfoD("Purging all backup locations")
	err = WorkflowbkpLoc.Purge()
	if err != nil {
		log.Errorf("some error occurred while purging backup locations - [%s]", err.Error())
		allErrors = append(allErrors, err)
	}

	log.InfoD("Purging all cloud credentials")
	err = WorkflowCc.Purge()
	if err != nil {
		log.Errorf("some error occurred while purging cloud credentials - [%s]", err.Error())
		allErrors = append(allErrors, err)
	}

	if len(allErrors) > 0 {
		var allErrorStrings []string
		for _, err := range allErrors {
			allErrorStrings = append(allErrorStrings, err.Error())
		}
		log.FailOnError(fmt.Errorf("[%s]", strings.Join(allErrorStrings, "\n\n")), "errors occurred while cleanup")
	}

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
