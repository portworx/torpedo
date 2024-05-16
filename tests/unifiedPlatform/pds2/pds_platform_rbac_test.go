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
	"strings"
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
		bothAccess             string
		err                    error
	)

	JustBeforeEach(func() {
		StartPDSTorpedoTest("BackupAndRestoreAccrossDifferentProjectsWithDifferentUsers", "Create backup and restore across different project using only project users", nil, 0)
		sourceUser = "source-user-" + RandomString(5)
		destinationUser = "destination-user-" + RandomString(5)
		bothAccess = "both-access-" + RandomString(5)
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
			WorkflowTargetClusterDestination.Project = &destinationProject
		})

		Step("Create project user for source Project", func() {
			workflowServiceAccount.WorkflowProjects = []*platform.WorkflowProject{&WorkflowProject}

			_, err := workflowServiceAccount.CreateServiceAccount(
				sourceUser,
				[]string{platform.ProjectWriter},
			)
			log.FailOnError(err, "Unable to create Project User")
			log.InfoD("Project User Account Created - [%s]", sourceUser)
		})

		Step("Create project user for destination Project", func() {
			workflowServiceAccount.WorkflowProjects = []*platform.WorkflowProject{&destinationProject}

			_, err := workflowServiceAccount.CreateServiceAccount(
				destinationUser,
				[]string{platform.ProjectWriter},
			)
			log.FailOnError(err, "Unable to create Project User")
			log.InfoD("Project User Account Created - [%s]", destinationUser)
		})

		Step("Create project user for destination Project", func() {
			workflowServiceAccount.WorkflowProjects = []*platform.WorkflowProject{&destinationProject, &WorkflowProject}

			_, err := workflowServiceAccount.CreateServiceAccount(
				bothAccess,
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

			Step("Create namespaces for restore", func() {
				workflowServiceAccount.SwitchToAdmin()

				restoreNamespace = "restore-" + RandomString(5)
				restoreName = "restore-" + RandomString(5)

				WorkflowNamespaceDestination.CreateNamespaces(PDS_DEFAULT_NAMESPACE)
				WorkflowNamespaceDestination.CreateNamespaces(restoreNamespace)
			})

			Step("Associate namespaces to destination project", func() {
				err := destinationProject.Associate(
					[]string{},
					[]string{WorkflowNamespaceDestination.Namespaces[restoreNamespace], WorkflowNamespaceDestination.Namespaces[PDS_DEFAULT_NAMESPACE]},
					[]string{},
					[]string{},
					[]string{},
					[]string{},
				)
				log.FailOnError(err, "Unable to associate Templates to Project")
				log.Infof("Associated Resources - [%+v]", WorkflowProject.AssociatedResources)
				WorkflowTargetClusterDestination.Project = &destinationProject
			})

			Step("Switch to destination project user", func() {
				workflowServiceAccount.SwitchToServiceAccount(destinationUser)
			})

			Step("Create Restore from the latest backup Id without having access to source project", func() {
				defer func() {
					err := SetSourceKubeConfig()
					log.FailOnError(err, "failed to switch context to source cluster")
				}()
				CheckforClusterSwitch()
				_, err := WorkflowPDSRestore.CreateRestore(restoreName, latestBackupUid, restoreNamespace, *deployment.Create.Meta.Uid)
				dash.VerifyFatal(strings.Contains(err.Error(), "403 Forbidden"), true, "Create restore without having access to source project - 403 Forbidden")
			})

			Step("Switch to user with access to both project", func() {
				workflowServiceAccount.SwitchToAdmin()
			})

			Step("Create Restore from the latest backup Id with access to source project", func() {
				defer func() {
					err := SetSourceKubeConfig()
					log.FailOnError(err, "failed to switch context to source cluster")
				}()
				CheckforClusterSwitch()
				_, err := WorkflowPDSRestore.CreateRestore(restoreName, latestBackupUid, restoreNamespace, *deployment.Create.Meta.Uid)
				log.FailOnError(err, "Restore Failed")
				log.Infof("Restore created successfully with ID - [%s]", WorkflowPDSRestore.Restores[restoreName].Meta.Uid)
			})
		}

	})

	JustAfterEach(func() {
		defer EndPDSTorpedoTest()
		log.InfoD("Switching back to admin account")
		workflowServiceAccount.SwitchToAdmin()
	})
})

var _ = Describe("{BackupAndRestoreAccrossSameProjectsWithDifferentUsers}", func() {
	var (
		deployment             *automationModels.PDSDeploymentResponse
		workflowServiceAccount platform.WorkflowServiceAccount
		deploymentUser         string
		backupUser             string
		restoreUser            string
		latestBackupUid        string
		pdsBackupConfigName    string
		restoreNamespace       string
		restoreName            string
		err                    error
	)

	JustBeforeEach(func() {
		StartPDSTorpedoTest("BackupAndRestoreAccrossDifferentProjectsWithDifferentUsers", "Create backup and restore across different project using only project users", nil, 0)
		deploymentUser = "deployment-" + RandomString(5)
		backupUser = "backup-" + RandomString(5)
		restoreUser = "restore-" + RandomString(5)
		workflowServiceAccount.UserRoles = make(map[string]platform.SeviceAccount)
		workflowServiceAccount.WorkflowProjects = []*platform.WorkflowProject{&WorkflowProject}
	})

	It("Create backup and restore across different project using only project users", func() {

		Step("Create project user - Deployment User", func() {
			_, err := workflowServiceAccount.CreateServiceAccount(
				deploymentUser,
				[]string{platform.ProjectWriter},
			)
			log.FailOnError(err, "Unable to create Project User")
			log.InfoD("Deployment User Account Created - [%s]", deploymentUser)
		})

		Step("Create project user - Backup User", func() {
			_, err := workflowServiceAccount.CreateServiceAccount(
				backupUser,
				[]string{platform.ProjectWriter},
			)
			log.FailOnError(err, "Unable to create Project User")
			log.InfoD("Backup User Account Created - [%s]", backupUser)
		})

		Step("Create project user - Restore User", func() {
			_, err := workflowServiceAccount.CreateServiceAccount(
				restoreUser,
				[]string{platform.ProjectWriter},
			)
			log.FailOnError(err, "Unable to create Project User")
			log.InfoD("Restore User Account Created - [%s]", restoreUser)
		})

		for _, ds := range NewPdsParams.DataServiceToTest {

			Step("Switch to source project user", func() {
				workflowServiceAccount.SwitchToServiceAccount(deploymentUser)
			})

			Step("Deploy dataservice - Deployment User", func() {
				deployment, err = WorkflowDataService.DeployDataService(ds, ds.Image, ds.Version, PDS_DEFAULT_NAMESPACE)
				log.FailOnError(err, "Error while deploying ds")
				log.Infof("All deployments - [%+v]", WorkflowDataService.DataServiceDeployment)

			})

			Step("Switch to source project user", func() {
				workflowServiceAccount.SwitchToServiceAccount(backupUser)
			})

			Step("Create Adhoc backup config of the existing deployment - Backup User", func() {
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

			Step("Create namespaces for restore", func() {
				workflowServiceAccount.SwitchToAdmin()

				restoreNamespace = "restore-" + RandomString(5)
				restoreName = "restore-" + RandomString(5)

				WorkflowNamespace.CreateNamespaces(PDS_DEFAULT_NAMESPACE)
				WorkflowNamespace.CreateNamespaces(restoreNamespace)
			})

			Step("Switch to destination project user", func() {
				workflowServiceAccount.SwitchToServiceAccount(restoreUser)
			})

			Step("Create Restore from the latest backup - Restore User", func() {
				defer func() {
					err := SetSourceKubeConfig()
					log.FailOnError(err, "failed to switch context to source cluster")
				}()
				CheckforClusterSwitch()
				_, err := WorkflowPDSRestore.CreateRestore(restoreName, latestBackupUid, restoreNamespace, *deployment.Create.Meta.Uid)
				log.FailOnError(err, "Restore Failed")
				log.Infof("Restore created successfully with ID - [%s]", WorkflowPDSRestore.Restores[restoreName].Meta.Uid)
			})
		}

	})

	JustAfterEach(func() {
		defer EndPDSTorpedoTest()
		log.InfoD("Switching back to admin account")
		workflowServiceAccount.SwitchToAdmin()
	})
})
