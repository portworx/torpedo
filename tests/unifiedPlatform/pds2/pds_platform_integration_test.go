package tests

import (
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/stworkflows/platform"
	"github.com/portworx/torpedo/drivers/utilities"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests"
	. "github.com/portworx/torpedo/tests/unifiedPlatform"
)

var _ = Describe("{BackupAndRestoreAccrossDifferentProjectsWithDifferentUsers}", func() {
	var (
		deployment             *automationModels.PDSDeploymentResponse
		destinationProject     platform.WorkflowProject
		workflowServiceAccount platform.WorkflowServiceAccount
		sourceUser             string
		destinationUser        string
		latestBackupUid        string
		pdsBackupConfigName    string
		restoreNamespace       string
		restoreName            string
		err                    error
	)

	JustBeforeEach(func() {
		StartPDSTorpedoTest("BackupAndRestoreAccrossDifferentProjectsWithDifferentUsers", "Create backup and restore across different project using only project users", nil, 0)
		restoreNamespace = "restore-" + RandomString(5)
		restoreName = "restore-" + RandomString(5)
		sourceUser = "source-user-" + RandomString(5)
		destinationUser = "destination-user-" + RandomString(5)
		workflowServiceAccount.UserRoles = make(map[string]platform.SeviceAccount)
		WorkflowPDSRestore.Destination = &WorkflowNamespaceDestination
	})

	It("Create backup and restore across different project using only project users", func() {

		Step("Create Destination Project", func() {
			destinationProject.Platform = WorkflowPlatform
			destinationProject.ProjectName = fmt.Sprintf("destination-project-%s", utilities.RandomString(5))
			workflowProject, err := destinationProject.CreateProject()
			log.FailOnError(err, "Unable to create project")
			log.InfoD("Destination Project created with ID - [%s]", workflowProject.ProjectId)
		})

		Step("Associate resources to destination project", func() {
			err := destinationProject.Associate(
				[]string{WorkflowTargetCluster.ClusterUID, WorkflowTargetClusterDestination.ClusterUID},
				[]string{},
				[]string{WorkflowCc.CloudCredentials[NewPdsParams.BackUpAndRestore.TargetLocation].ID},
				[]string{WorkflowbkpLoc.BkpLocation.BkpLocationId},
				[]string{},
				[]string{},
			)
			log.FailOnError(err, "Unable to associate Templates to Project")
			log.Infof("Associated Resources - [%+v]", WorkflowProject.AssociatedResources)
		})

		Step("Create project user for source Project", func() {
			workflowServiceAccount.WorkflowProject = WorkflowProject

			_, err := workflowServiceAccount.CreateServiceAccount(
				sourceUser,
				[]string{platform.ProjectWriter},
			)
			log.FailOnError(err, "Unable to create Project User")
			log.InfoD("Project User Account Created - [%s]", sourceUser)
		})

		Step("Create project user for destination Project", func() {
			workflowServiceAccount.WorkflowProject = destinationProject

			_, err := workflowServiceAccount.CreateServiceAccount(
				destinationUser,
				[]string{platform.ProjectWriter},
			)
			log.FailOnError(err, "Unable to create Project User")
			log.InfoD("Project User Account Created - [%s]", destinationUser)
		})

		for _, ds := range NewPdsParams.DataServiceToTest {

			Step("Switch to source project user", func() {
				workflowServiceAccount.SwitchToServiceAccount(sourceUser)
			})

			Step("Deploy dataservice", func() {
				deployment, err = WorkflowDataService.DeployDataService(ds, ds.Image, ds.Version, PDS_DEFAULT_NAMESPACE)
				log.FailOnError(err, "Error while deploying ds")
				log.Infof("All deployments - [%+v]", WorkflowDataService.DataServiceDeployment)

			})

			Step("Create Adhoc backup config of the existing deployment - Project User", func() {
				pdsBackupConfigName = "pds-adhoc-backup-" + RandomString(5)
				bkpConfigResponse, err := WorkflowPDSBackupConfig.CreateBackupConfig(pdsBackupConfigName, *deployment.Create.Meta.Uid)
				log.FailOnError(err, "Error occured while creating backupConfig")
				log.Infof("BackupConfigName: [%s], BackupConfigId: [%s]", *bkpConfigResponse.Create.Meta.Name, *bkpConfigResponse.Create.Meta.Uid)
				log.Infof("All deployments - [%+v]", WorkflowDataService.DataServiceDeployment)
			})

			Step("Get the latest backup detail for the deployment", func() {
				backupResponse, err := WorkflowPDSBackup.GetLatestBackup(*deployment.Create.Meta.Uid)
				log.FailOnError(err, "Error occured while creating backup")
				latestBackupUid = *backupResponse.Meta.Uid
				log.Infof("Latest backup ID [%s], Name [%s]", *backupResponse.Meta.Uid, *backupResponse.Meta.Name)
				err = WorkflowPDSBackup.WaitForBackupToComplete(*backupResponse.Meta.Uid)
				log.FailOnError(err, "Error occured while waiting for backup to complete")
			})

			Step("Switch to destination project user", func() {
				workflowServiceAccount.SwitchToServiceAccount(destinationUser)
			})

			Step("Create Restore from the latest backup Id", func() {
				defer func() {
					err := SetSourceKubeConfig()
					log.FailOnError(err, "failed to switch context to source cluster")
				}()
				CheckforClusterSwitch()
				_, err := WorkflowPDSRestore.CreateRestore(restoreName, latestBackupUid, restoreNamespace, *deployment.Create.Meta.Uid)
				log.FailOnError(err, "Restore Failed")
				log.Infof("All restores - [%+v]", WorkflowPDSRestore.Restores)
				log.Infof("Restore Created Name - [%s], UID - [%s]", *WorkflowPDSRestore.Restores[restoreName].Meta.Name, *WorkflowPDSRestore.Restores[restoreName].Meta.Uid)
			})
		}

	})

	JustAfterEach(func() {
		defer EndPDSTorpedoTest()
		log.InfoD("Switching back to admin account")
		workflowServiceAccount.SwitchToAdmin()
	})
})
