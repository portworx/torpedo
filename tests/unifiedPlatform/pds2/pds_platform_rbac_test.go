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
	"reflect"
	"strconv"
	"strings"
	"sync"
)

var _ = Describe("{BackupAndRestoreAcrossDifferentProjectsWithDifferentUsers}", func() {
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
				TemplateIds,
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

var _ = Describe("{BackupAndRestoreAcrossSameProjectsWithDifferentUsers}", func() {
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

			Step("Associate both namespaces to the project", func() {
				err := WorkflowProject.Associate(
					[]string{},
					[]string{WorkflowNamespace.Namespaces[PDS_DEFAULT_NAMESPACE], WorkflowNamespace.Namespaces[restoreNamespace]},
					[]string{},
					[]string{},
					[]string{},
					[]string{},
				)
				log.FailOnError(err, "Unable to associate Templates to Project")
				log.Infof("Associated Resources - [%+v]", WorkflowProject.AssociatedResources)
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

var _ = Describe("{DeployDsOnMultipleNSAndProjects}", func() {
	var (
		numberOfNamespacesTobeCreated int
		namespacePrefix               string
		allError                      []string
		namespaces                    []string
		projects                      []string
		project2                      platform.WorkflowProject
		project3                      platform.WorkflowProject
		namespaceNameAndId            map[string]string
	)

	JustBeforeEach(func() {
		StartPDSTorpedoTest("DeployDsOnMultipleNSAndProjects", "Create Multiple Namespaces, Projects and Associates Namespaces to projects then validates cross projects rbac", nil, 0)
		numberOfNamespacesTobeCreated = 3 // Number of namespaces to be created by the testcase
		namespacePrefix = "rbac-ns-"
	})

	It("Enables and Disables pds on a namespace multiple times", func() {
		namespaceNameAndId = make(map[string]string)
		Step(fmt.Sprintf("Creating [%d] namespaces with labels", numberOfNamespacesTobeCreated), func() {
			var wg sync.WaitGroup

			log.InfoD("Creating [%d] namespaces with PDS labels present", numberOfNamespacesTobeCreated)
			for i := 0; i < numberOfNamespacesTobeCreated; i++ {
				wg.Add(1)
				nsName := namespacePrefix + RandomString(5) + "-" + strconv.Itoa(i)
				go func() {
					defer wg.Done()
					defer GinkgoRecover()

					_, err := WorkflowNamespace.CreateNamespaces(nsName)
					if err != nil {
						allError = append(allError, err.Error())
					}
				}()
				namespaces = append(namespaces, nsName)
			}
			wg.Wait()
			if allError != nil {
				log.Errorf(strings.Join(allError, "\n"))
			}
			dash.VerifyFatal(len(allError), 0, "Verifying namespaces creation")
		})

		Step("Validating all current namespaces", func() {
			log.InfoD("Validating all current namespaces")
			for _, namespace := range namespaces {
				ns, err := WorkflowNamespace.GetNamespace(namespace)
				if err != nil {
					allError = append(allError, fmt.Sprintf("Some error occurred while listing namespace. Error - [%s]", err.Error()))
				} else {
					if *ns.Status.Phase != AVAILABLE {
						allError = append(allError, fmt.Sprintf("[%s] is in [%s] state. Expected - [%s]", namespace, *ns.Status.Phase, AVAILABLE))
					}
				}
				log.Infof("[%s] - [%s]", namespace, *ns.Status.Phase)
				namespaceNameAndId[*ns.Meta.Name] = *ns.Meta.Uid
			}

			if allError != nil {
				log.Errorf(strings.Join(allError, "\n"))
			}
			dash.VerifyFatal(len(allError), 0, "Verifying namespaces on control plane")
		})

		steplog := "Create Project1 and Associate Namespace to the projects"
		Step(steplog, func() {
			log.InfoD(steplog)
			//project1.Platform = WorkflowPlatform
			//PROJECT_NAME := "rbac-project-" + RandomString(5) + "-1"
			//project1.ProjectName = PROJECT_NAME
			//_, err := project1.CreateProject()
			//log.FailOnError(err, "unable to create project")
			//ProjectId, err = project1.GetDefaultProject(PROJECT_NAME)
			//log.FailOnError(err, "Unable to get current project")
			//log.Infof("Current project ID - [%s]", ProjectId)
			//projects = append(projects, ProjectId)

			log.Debugf("namespace-[%s], namespaceId-[%s]", namespaces[0], namespaceNameAndId[namespaces[0]])
			log.Debugf("namespace-[%s], namespaceId-[%s]", namespaces[1], namespaceNameAndId[namespaces[1]])
			log.Debugf("namespace-[%s], namespaceId-[%s]", namespaces[2], namespaceNameAndId[namespaces[2]])

			//Associate namespace to the project
			log.InfoD("Associate namespace to the Project1")
			err := WorkflowProject.Associate(
				[]string{},
				[]string{namespaceNameAndId[namespaces[0]], namespaceNameAndId[namespaces[1]]},
				[]string{},
				[]string{},
				[]string{},
				[]string{},
			)
			log.FailOnError(err, "Unable to associate Templates to Project")
			log.Infof("Associated Resources - [%+v]", WorkflowProject.AssociatedResources)
		})

		steplog = "Create Project2 and Associate Namespace to the projects"
		Step(steplog, func() {
			log.InfoD(steplog)
			project2.Platform = WorkflowPlatform
			PROJECT_NAME := "rbac-project-" + RandomString(5) + "-2"
			project2.ProjectName = PROJECT_NAME
			_, err := project2.CreateProject()
			log.FailOnError(err, "unable to create project")
			ProjectId, err = project2.GetDefaultProject(PROJECT_NAME)
			log.FailOnError(err, "Unable to get current project")
			log.Infof("Current project ID - [%s]", ProjectId)
			projects = append(projects, ProjectId)

			//Associate namespace to the project
			log.InfoD("Associate namespace to the Project2")
			err = project2.Associate(
				[]string{},
				[]string{namespaceNameAndId[namespaces[1]], namespaceNameAndId[namespaces[2]]},
				[]string{},
				[]string{},
				[]string{},
				[]string{},
			)
			log.FailOnError(err, "Unable to associate Templates to Project")
			log.Infof("Associated Resources - [%+v]", project2.AssociatedResources)
		})

		steplog = "Create Project3 and Associate Namespace to the projects"
		Step(steplog, func() {
			log.InfoD(steplog)
			project3.Platform = WorkflowPlatform
			PROJECT_NAME := "rbac-project-" + RandomString(5) + "-3"
			project3.ProjectName = PROJECT_NAME
			_, err := project3.CreateProject()
			log.FailOnError(err, "unable to create project")
			ProjectId, err = project3.GetDefaultProject(PROJECT_NAME)
			log.FailOnError(err, "Unable to get current project")
			log.Infof("Current project ID - [%s]", ProjectId)
			projects = append(projects, ProjectId)

			//Associate namespace to the project
			log.InfoD("Associate namespace to the Project3")
			err = project3.Associate(
				[]string{},
				[]string{namespaceNameAndId[namespaces[2]], namespaceNameAndId[namespaces[0]]},
				[]string{},
				[]string{},
				[]string{},
				[]string{},
			)
			log.FailOnError(err, "Unable to associate Templates to Project")
			log.Infof("Associated Resources - [%+v]", project3.AssociatedResources)
		})

		steplog = "Validate the Namespaces are not accessible from the projects to which it is not associated"
		Step(steplog, func() {
			log.InfoD(steplog)

			//project1
			prj1, err := WorkflowProject.GetProject()
			log.FailOnError(err, "Error while getting project")
			namespaceFetchedFromApi := prj1.Config.InfraResources.Namespaces
			associatedNamespaces := WorkflowProject.AssociatedResources.Namespaces
			dash.VerifyFatal(reflect.DeepEqual(namespaceFetchedFromApi, associatedNamespaces), true, "validating the associated namespaces in project1")

			//project2
			prj2, err := project2.GetProject()
			log.FailOnError(err, "Error while getting project")
			namespaceFetchedFromApi2 := prj2.Config.InfraResources.Namespaces
			associatedNamespaces2 := project2.AssociatedResources.Namespaces
			dash.VerifyFatal(reflect.DeepEqual(namespaceFetchedFromApi2, associatedNamespaces2), true, "validating the associated namespaces in project2")

			//project3
			prj3, err := project3.GetProject()
			log.FailOnError(err, "Error while getting project")
			namespaceFetchedFromApi3 := prj3.Config.InfraResources.Namespaces
			associatedNamespaces3 := project3.AssociatedResources.Namespaces
			dash.VerifyFatal(reflect.DeepEqual(namespaceFetchedFromApi3, associatedNamespaces3), true, "validating the associated namespaces in project3")
		})

		//Templates are already associated to project1 in pds_basic_test.go file

		steplog = "Validate the Templates are not accessible from the projects to which it is not associated"
		Step(steplog, func() {
			log.InfoD(steplog)
			//project1
			wfProject, err := WorkflowProject.GetProject()
			log.FailOnError(err, "Error while getting project")
			templatesFetchedFromApi := wfProject.Config.InfraResources.Templates
			associatedTemplates := WorkflowProject.AssociatedResources.Templates
			dash.VerifyFatal(reflect.DeepEqual(templatesFetchedFromApi, associatedTemplates), true, "validating the associated templates in project1")

			//project2
			prj2, err := project2.GetProject()
			log.FailOnError(err, "Error while getting project")
			templatesFetchedFromApi2 := prj2.Config.InfraResources.Templates
			//associatedTemplates2 := project2.AssociatedResources.Templates
			dash.VerifyFatal(reflect.DeepEqual(templatesFetchedFromApi2, associatedTemplates), false, "validating the associated templates in project1")

			//project3
			prj3, err := project3.GetProject()
			log.FailOnError(err, "Error while getting project")
			namespaceFetchedFromApi3 := prj3.Config.InfraResources.Namespaces
			//associatedNamespaces3 := project3.AssociatedResources.Namespaces
			dash.VerifyFatal(reflect.DeepEqual(namespaceFetchedFromApi3, associatedTemplates), false, "validating the associated namespaces in project3")

		})

		steplog = "Deploy DataService on the above created projects"
		Step(steplog, func() {
			log.InfoD(steplog)

			//Deployment expected to fail
			WorkflowDataService.Namespace.TargetCluster.Project = &project3
			for _, ds := range NewPdsParams.DataServiceToTest {
				_, err := WorkflowDataService.DeployDataService(ds, ds.Image, ds.Version, namespaces[0])
				if strings.Contains(err.Error(), "403 Forbidden") {
					log.Errorf(err.Error(), "Error while deploying ds")
				}
				break // running for only one dataService
			}

			//Deployment expected to fail
			WorkflowDataService.Namespace.TargetCluster.Project = &project2
			for _, ds := range NewPdsParams.DataServiceToTest {
				_, err := WorkflowDataService.DeployDataService(ds, ds.Image, ds.Version, namespaces[0])
				if strings.Contains(err.Error(), "403 Forbidden") {
					log.Errorf(err.Error(), "Error while deploying ds")
				}
				break // running for only one dataService
			}

			//working deployment with all associations
			WorkflowDataService.Namespace.TargetCluster.Project = &WorkflowProject
			for _, ds := range NewPdsParams.DataServiceToTest {
				deployment, err := WorkflowDataService.DeployDataService(ds, ds.Image, ds.Version, namespaces[0])
				log.FailOnError(err, "Error while deploying ds")
				log.Debugf("Source Deployment Id: [%s]", *deployment.Create.Meta.Uid)
				break // running for only one dataService
			}
		})
	})

	JustAfterEach(func() {
		defer EndPDSTorpedoTest()
	})
})
