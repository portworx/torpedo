package tests

import (
	. "github.com/onsi/ginkgo/v2"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/pdsLibs"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/stworkflows/pds"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests"
	tests "github.com/portworx/torpedo/tests/unifiedPlatform"
)

var _ = Describe("{CleanUpDeployments}", func() {
	It("Delete all deployments", func() {
		err := pdslibs.DeleteAllDeployments(tests.ProjectId)
		log.FailOnError(err, "error while deleting deployment")
	})
})

var _ = Describe("{GetCRObject}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("GetCRObject", "get the cr object", nil, 0)
	})

	var wfDataService pds.WorkflowDataService

	It("Get the CR object", func() {
		for _, ds := range tests.NewPdsParams.DataServiceToTest {
			wfDataService.DataServiceDeployment = make(map[string]string)
			wfDataService.DataServiceDeployment["pg-qa-bxgoxo"] = "dep:8965032c-c8e3-447f-af05-9b7a4badf5a9"
			resourceSettings, storageOps, deploymentConfig, err := pdslibs.GetDeploymentResources(wfDataService.DataServiceDeployment, ds.Name, "tmpl:04dab835-1fe2-4526-824f-d7a45694676c", "tmpl:a584ede7-811e-48bd-b000-ae799e3e084e", "pds-namespace-fdrey")
			log.FailOnError(err, "Error occured while getting deployment resources")
			var dataServiceVersionBuildMap = make(map[string][]string)
			wfDataService.ValidateDeploymentResources(resourceSettings, storageOps, deploymentConfig, ds.Replicas, dataServiceVersionBuildMap)
		}
	})
})

var _ = Describe("{ValidateDnsEndPoint}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("ValidateDnsEndPoint", "validate dns endpoint", nil, 0)
	})

	var (
		workflowDataservice pds.WorkflowDataService
		err                 error
	)

	It("ValidateDnsEndPoint", func() {
		Step("validate dns endpoint", func() {
			depId := "dep:8965032c-c8e3-447f-af05-9b7a4badf5a9"
			err = workflowDataservice.ValidateDNSEndpoint(depId)
			log.FailOnError(err, "Error occurred while validating dns endpoint")
		})
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
	})
})

var _ = Describe("{DummyBackupTest}", func() {

	var (
		workflowDataservice pds.WorkflowDataService
		workflowBackup      pds.WorkflowPDSBackup
		workflowRestore     pds.WorkflowPDSRestore
		deploymentName      string
		latestBackupUid     string
	)

	JustBeforeEach(func() {
		StartTorpedoTest("DummyBackupTest", "DummyBackupTest", nil, 0)
		workflowDataservice.DataServiceDeployment = make(map[string]string)
		deploymentName = "samore-pg-test-1"
		workflowDataservice.DataServiceDeployment[deploymentName] = "dep:fa70e52d-0563-4258-b96b-7d6ca6ed4799"

		workflowRestore.Destination = tests.WorkflowNamespace
		workflowRestore.WorkflowProject = tests.WorkflowProject
	})

	It("Dummy to verify backup and restore creation", func() {

		Step("Get latest backup from a backup config", func() {
			backupResponse, err := workflowBackup.GetLatestBackup(deploymentName)
			log.FailOnError(err, "Error occured while creating backup")
			latestBackupUid = *backupResponse.Meta.Uid
			log.Infof("Latest backup ID [%s], Name [%s]", *backupResponse.Meta.Uid, *backupResponse.Meta.Name)
		})

		Step("Get all backup from a backup config", func() {
			backupResponse, err := workflowBackup.ListAllBackups(deploymentName)
			log.FailOnError(err, "Error occured while creating backup")
			log.Infof("Number of backups - [%d]", len(backupResponse))

			log.Infof("Listing all backups \n\n")
			for _, backup := range backupResponse {
				log.Infof("Latest backup ID [%s], Name [%s]", *backup.Meta.Uid, *backup.Meta.Name)
			}

		})

		Step("Create Restore", func() {
			restoreName := "testing_restore_" + RandomString(5)
			_, err := workflowRestore.CreateRestore(restoreName, latestBackupUid, tests.PDS_DEFAULT_NAMESPACE)
			log.FailOnError(err, "Restore Failed")

			log.Infof("Restore created successfully with ID - [%s]", workflowRestore.Restores[restoreName].Meta.Uid)
		})

	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
	})
})

func PointerTo[T ~string](s T) *T {
	return &s
}
