package tests

import (
	. "github.com/onsi/ginkgo/v2"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/stworkflows"
	"github.com/portworx/torpedo/drivers/utilities"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests"
	. "github.com/portworx/torpedo/tests/unifiedPlatform"
	"strings"
)

var _ = Describe("{PerformRestoreToDifferentClusterSameProject}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("PerformRestoreToDifferentClusterSameProject", "Deploy data services and perform backup and restore on a different cluster on the same project", nil, 0)
	})
	var (
		workflowDataservice  stworkflows.WorkflowDataService
		workflowBackUpConfig stworkflows.WorkflowPDSBackupConfig
		deployment           *automationModels.WorkFlowResponse
		pdsBackupConfigName  string
		err                  error
	)

	It("Deploy, Validate and RunWorkloads on DataService", func() {
		Step("Create a PDS Namespace", func() {
			Namespace = strings.ToLower("pds-test-ns-" + utilities.RandString(5))
			WorkflowNamespace.TargetCluster = WorkflowTargetCluster
			WorkflowNamespace.Namespaces = make(map[string]string)
			workflowNamespace, err := WorkflowNamespace.CreateNamespaces(Namespace)
			log.FailOnError(err, "Unable to create namespace")
			log.Infof("Namespaces created - [%s]", workflowNamespace.Namespaces)
			log.Infof("Namespace id - [%s]", workflowNamespace.Namespaces[Namespace])
		})

		for _, ds := range NewPdsParams.DataServiceToTest {
			workflowDataservice.Namespace = WorkflowNamespace
			workflowDataservice.NamespaceName = Namespace
			deployment, err = workflowDataservice.DeployDataService(ds, ds.Image, ds.Version)
			log.FailOnError(err, "Error while deploying ds")
		}

		defer func() {
			Step("Delete DataServiceDeployment", func() {
				err := workflowDataservice.DeleteDeployment()
				log.FailOnError(err, "Error while deleting dataservice")
			})
		}()

		stepLog := "Running Workloads before taking backups"
		Step(stepLog, func() {
			err := workflowDataservice.RunDataServiceWorkloads(NewPdsParams)
			log.FailOnError(err, "Error while running workloads on ds")
		})
	})

	It("Perform adhoc backup, restore and validate them", func() {
		workflowBackUpConfig.WorkflowDataService = workflowDataservice
		pdsBackupConfigName = strings.ToLower("pds-qa-bkpConfig-" + utilities.RandString(5))
		Step("Take Backup and validate", func() {
			bkpConfigResponse, err := workflowBackUpConfig.CreateBackupConfig(pdsBackupConfigName, *deployment.PDSDeployment.V1Deployment.Meta.Uid)
			log.FailOnError(err, "Error occured while creating backupConfig")
			log.Infof("BackupConfigName: [%s], BackupConfigId: [%s]", bkpConfigResponse.Create.Meta.Name, bkpConfigResponse.Create.Meta.Uid)
		})

		defer func() {
			Step("Delete Backups", func() {
				err = workflowBackUpConfig.DeleteBackupConfig(pdsBackupConfigName)
				log.FailOnError(err, "Error while deleting BackupConfig [%s]", pdsBackupConfigName)
			})
		}()

		Step("Perform Restore and validate", func() {
			//TODO: Restore steps will be added once WorkFlow is available
		})

		defer func() {
			Step("Delete RestoredDeployment", func() {

			})
		}()

		Step("Validate md5hash for the restored deployments", func() {
			err := workflowDataservice.ValidateDataServiceWorkloads(NewPdsParams)
			log.FailOnError(err, "Error occured in ValidateDataServiceWorkloads method")
		})

	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
	})
})
