package tests

import (
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	"github.com/portworx/torpedo/drivers/unifiedPlatform"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/stworkflows"
	"github.com/portworx/torpedo/drivers/utilities"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests"
	. "github.com/portworx/torpedo/tests/unifiedPlatform"
)

var _ = Describe("{PlatformBasicTest}", func() {
	var (
		workflowPlatform       stworkflows.WorkflowPlatform
		workflowTargetCluster  stworkflows.WorkflowTargetCluster
		workflowProject        stworkflows.WorkflowProject
		workflowNamespace      stworkflows.WorkflowNamespace
		workflowCloudCreds     stworkflows.WorkflowCloudCredentials
		workflowBackupLocation stworkflows.WorkflowBackupLocation
		namespace              string
		VARIABLE_FROM_JENKINS  string
	)
	JustBeforeEach(func() {

		StartTorpedoTest("PlatformBasicTest", "Basic CRUD operations on platform", nil, 0)
		namespace = fmt.Sprintf("pds-namespace-%s", utilities.RandomString(5))
		workflowPlatform.Accounts = map[string]map[string]string{
			NewPdsParams.Users.AdminUsername: map[string]string{
				automationModels.UserName:        NewPdsParams.Users.AdminUsername,
				automationModels.UserDisplayName: NewPdsParams.Users.AdminUsername,
				automationModels.UserEmail:       NewPdsParams.Users.AdminEmailAddress,
			},
			NewPdsParams.Users.NonAdminUsername: map[string]string{
				automationModels.UserName:        NewPdsParams.Users.NonAdminUsername,
				automationModels.UserDisplayName: NewPdsParams.Users.NonAdminUsername,
				automationModels.UserEmail:       NewPdsParams.Users.NonAdminEmailAddress,
			},
		}
		VARIABLE_FROM_JENKINS = GetEnv(unifiedPlatform.UNIFIED_PLATFORM_INTERFACE, unifiedPlatform.REST_API)
		workflowPlatform.TenantInit()
	})

	It("Basic CRUD operations on platform", func() {
		Step("Create Cloud Credentials", func() {
			// TODO: This needs to be removed once API support is added for cloud creds
			if VARIABLE_FROM_JENKINS == unifiedPlatform.GRPC {
				workflowCloudCreds.Platform = workflowPlatform
				workflowCloudCreds.CloudCredentials = make(map[string]stworkflows.CloudCredentialsType)
				_, err := workflowCloudCreds.CreateCloudCredentials(NewPdsParams.BackUpAndRestore.TargetLocation)
				log.FailOnError(err, "Unable to create cloud credentials")
				for _, value := range workflowCloudCreds.CloudCredentials {
					log.Infof("cloud credentials name: [%s]", value.Name)
					log.Infof("cloud credentials id: [%s]", value.ID)
					log.Infof("cloud provider type: [%s]", value.CloudProviderType)
				}
			}
		})

		Step("Create Backup Location", func() {
			//// TODO: This needs to be removed once API support is added for backup location
			if VARIABLE_FROM_JENKINS == unifiedPlatform.GRPC {
				workflowBackupLocation.WfCloudCredentials = workflowCloudCreds
				_, err := workflowBackupLocation.CreateBackupLocation(PDSBucketName, NewPdsParams.BackUpAndRestore.TargetLocation)
				log.FailOnError(err, "error while creating backup location")
				log.Infof("wfBkpLoc id: [%s]", workflowBackupLocation.BkpLocation.BkpLocationId)
				log.Infof("wfBkpLoc name: [%s]", workflowBackupLocation.BkpLocation.Name)
			}
		})

		Step("Create Project", func() {
			workflowProject.Platform = workflowPlatform
			workflowProject.ProjectName = fmt.Sprintf("project-%s", utilities.RandomString(5))
			workflowProject, err := workflowProject.CreateProject()
			log.FailOnError(err, "Unable to create project")
			log.Infof("Project created with ID - [%s]", workflowProject.ProjectId)
		})

		Step("Register Target Cluster", func() {
			workflowTargetCluster.Project = workflowProject
			log.Infof("Tenant ID [%s]", workflowTargetCluster.Project.Platform.TenantId)
			workflowTargetCluster, err := workflowTargetCluster.RegisterToControlPlane(false)
			log.FailOnError(err, "Unable to register target cluster")
			log.Infof("Target cluster registered with uid - [%s]", workflowTargetCluster.ClusterUID)
		})

		Step("Create a PDS Namespace", func() {
			workflowNamespace.TargetCluster = workflowTargetCluster
			workflowNamespace.Namespaces = make(map[string]string)
			_, err := workflowNamespace.CreateNamespaces(namespace)
			log.FailOnError(err, "Unable to create namespace")
			log.Infof("Namespaces created - [%s]", workflowNamespace.Namespaces)
		})

		Step("Associate namespace and cluster to Project", func() {
			err := workflowProject.Associate(
				[]string{workflowTargetCluster.ClusterUID},
				[]string{workflowNamespace.Namespaces[namespace]},
				[]string{},
				[]string{},
				[]string{},
				[]string{},
			)
			log.FailOnError(err, "Unable to associate Cluster to Project")
			log.Infof("Associated Resources - [%+v]", workflowProject.AssociatedResources)
		})

		Step("Dissociate cluster from Project", func() {
			err := workflowProject.Dissociate(
				[]string{workflowTargetCluster.ClusterUID},
				[]string{workflowNamespace.Namespaces[namespace]},
				[]string{},
				[]string{},
				[]string{},
				[]string{},
			)
			log.FailOnError(err, "Unable to dissociated Cluster from Project")
			log.Infof("Dissociated Clusters - [%s]", workflowTargetCluster.ClusterUID)
			log.Infof("Dissociated namespaces - [%s]", workflowNamespace.Namespaces[namespace])
			log.Infof("Associated Resources - [%+v]", workflowProject.AssociatedResources)
		})

		Step("Delete PDS Namespace", func() {
			err := workflowNamespace.DeleteNamespace(namespace)
			log.FailOnError(err, "Unable to delete namespace")
			log.Infof("Namespaces deleted - [%s]", namespace)
			log.Infof("Namespaces - [%+v]", workflowNamespace.Namespaces)
		})

		Step("Delete Project", func() {
			err := workflowProject.DeleteProject()
			log.FailOnError(err, "Delete project")
			log.Infof("Project deleted successfully")
		})

		//Step("Install PDS Applications", func() {
		//	err := workflowTargetCluster.InstallPDSAppOnTC()
		//	log.FailOnError(err, "Install PDS apps failed on Target cluster")
		//	log.Infof("PDS Apps deployed successfully on the targte cluster")
		//})

	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
	})
})

var _ = Describe("{PlatformRBACTest}", func() {
	var (
		workflowPlatform       stworkflows.WorkflowPlatform
		workflowTargetCluster  stworkflows.WorkflowTargetCluster
		workflowProject        stworkflows.WorkflowProject
		workflowNamespace      stworkflows.WorkflowNamespace
		workflowServiceAccount stworkflows.WorkflowServiceAccount
		namespace              string
		projectAdmin           string
		projectUser            string
	)
	JustBeforeEach(func() {

		StartTorpedoTest("PlatformRBACTest", "Basic RBAC operations on platform", nil, 0)
		namespace = fmt.Sprintf("pds-namespace-%s", utilities.RandomString(5))
		workflowPlatform.Accounts = map[string]map[string]string{
			NewPdsParams.Users.AdminUsername: map[string]string{
				automationModels.UserName:        NewPdsParams.Users.AdminUsername,
				automationModels.UserDisplayName: NewPdsParams.Users.AdminUsername,
				automationModels.UserEmail:       NewPdsParams.Users.AdminEmailAddress,
			},
			NewPdsParams.Users.NonAdminUsername: map[string]string{
				automationModels.UserName:        NewPdsParams.Users.NonAdminUsername,
				automationModels.UserDisplayName: NewPdsParams.Users.NonAdminUsername,
				automationModels.UserEmail:       NewPdsParams.Users.NonAdminEmailAddress,
			},
		}
		workflowPlatform.TenantInit()
		projectAdmin = "project-admin-" + RandomString(5)
		projectUser = "project-user-" + RandomString(5)

	})

	It("Basic RBAC operations on platform", func() {

		Step("Create Project", func() {
			workflowProject.Platform = workflowPlatform
			workflowProject.ProjectName = fmt.Sprintf("project-%s", utilities.RandomString(5))
			workflowProject, err := workflowProject.CreateProject()
			log.FailOnError(err, "Unable to create project")
			log.Infof("Project created with ID - [%s]", workflowProject.ProjectId)
		})

		Step("Register Target Cluster", func() {
			workflowTargetCluster.Project = workflowProject
			log.Infof("Tenant ID [%s]", workflowTargetCluster.Project.Platform.TenantId)
			workflowTargetCluster, err := workflowTargetCluster.RegisterToControlPlane(false)
			log.FailOnError(err, "Unable to register target cluster")
			log.Infof("Target cluster registered with uid - [%s]", workflowTargetCluster.ClusterUID)
		})

		Step("Create a PDS Namespace", func() {
			workflowNamespace.TargetCluster = workflowTargetCluster
			workflowNamespace.Namespaces = make(map[string]string)
			_, err := workflowNamespace.CreateNamespaces(namespace)
			log.FailOnError(err, "Unable to create namespace")
			log.Infof("Namespaces created - [%s]", workflowNamespace.Namespaces)
		})

		Step("Create project admin user", func() {
			rbacParams := NewPdsParams.RbacParams
			workflowServiceAccount.CreateServiceAccount(
				workflowPlatform.AdminAccountId,
				projectAdmin,
				stworkflows.ProjectAdmin,
				rbacParams.ResourceId,
			)
		})

		Step("Create project user", func() {
			rbacParams := NewPdsParams.RbacParams
			workflowServiceAccount.CreateServiceAccount(
				workflowPlatform.AdminAccountId,
				projectUser,
				stworkflows.User,
				rbacParams.ResourceId,
			)
		})

		Step("Create Project with Project Admin - Expected Failure", func() {
			defer func() {
				workflowServiceAccount.SwitchToAdmin()
			}()
			workflowServiceAccount.SwitchToServiceAccount(projectAdmin)
			workflowProject.Platform = workflowPlatform
			workflowProject.ProjectName = fmt.Sprintf("project-%s", utilities.RandomString(5))
			_, err := workflowProject.CreateProject()
			// TODO: Error needs to be changed with actual error at the time of validation
			dash.VerifyFatal(err, err, "Create Project with project-admin failed with project admin privelege")
		})

		Step("Create Project with Project Users - Expected Failure", func() {
			defer func() {
				workflowServiceAccount.SwitchToAdmin()
			}()
			workflowServiceAccount.SwitchToServiceAccount(projectUser)
			workflowProject.Platform = workflowPlatform
			workflowProject.ProjectName = fmt.Sprintf("project-%s", utilities.RandomString(5))
			_, err := workflowProject.CreateProject()
			// TODO: Error needs to be changed with actual error at the time of validation
			dash.VerifyFatal(err, err, "Create Project with project-user failed with project admin privelege")
		})

		Step("Register Target Cluster - Project Admin", func() {
			defer func() {
				workflowServiceAccount.SwitchToAdmin()
			}()
			workflowServiceAccount.SwitchToServiceAccount(projectAdmin)
			workflowTargetCluster.Project = workflowProject
			log.Infof("Tenant ID [%s]", workflowTargetCluster.Project.Platform.TenantId)
			_, err := workflowTargetCluster.RegisterToControlPlane(false)
			// TODO: Error needs to be changed with actual error at the time of validation
			dash.VerifyFatal(err, err, "Register target cluster failed with project admin account")
		})

		Step("Register Target Cluster - Project User", func() {
			defer func() {
				workflowServiceAccount.SwitchToAdmin()
			}()
			workflowServiceAccount.SwitchToServiceAccount(projectUser)
			workflowTargetCluster.Project = workflowProject
			log.Infof("Tenant ID [%s]", workflowTargetCluster.Project.Platform.TenantId)
			_, err := workflowTargetCluster.RegisterToControlPlane(false)
			// TODO: Error needs to be changed with actual error at the time of validation
			dash.VerifyFatal(err, err, "Register target cluster failed with project user account")
		})

		Step("Register Target Cluster - Project User", func() {
			defer func() {
				workflowServiceAccount.SwitchToAdmin()
			}()
			workflowServiceAccount.SwitchToServiceAccount(projectUser)
			workflowTargetCluster.Project = workflowProject
			log.Infof("Tenant ID [%s]", workflowTargetCluster.Project.Platform.TenantId)
			_, err := workflowTargetCluster.RegisterToControlPlane(false)
			// TODO: Error needs to be changed with actual error at the time of validation
			dash.VerifyFatal(err, err, "Register target cluster failed with project user account")
		})

		Step("Associate namespace to Project", func() {
			defer func() {
				workflowServiceAccount.SwitchToAdmin()
			}()
			workflowServiceAccount.SwitchToServiceAccount(projectAdmin)
			err := workflowProject.Associate(
				[]string{},
				[]string{workflowTargetCluster.ClusterUID},
				[]string{},
				[]string{},
				[]string{},
				[]string{},
			)
			log.FailOnError(err, "Unable to associate Cluster to Project")
			log.Infof("Associated Resources - [%+v]", workflowProject.AssociatedResources)
		})

		Step("Dissociate namespace from Project", func() {
			defer func() {
				workflowServiceAccount.SwitchToAdmin()
			}()
			workflowServiceAccount.SwitchToServiceAccount(projectAdmin)
			err := workflowProject.Dissociate(
				[]string{},
				[]string{workflowNamespace.Namespaces[namespace]},
				[]string{},
				[]string{},
				[]string{},
				[]string{},
			)
			log.FailOnError(err, "Unable to dissociated Cluster from Project")
			log.Infof("Dissociated Clusters - [%s]", workflowTargetCluster.ClusterUID)
			log.Infof("Dissociated namespaces - [%s]", workflowNamespace.Namespaces[namespace])
			log.Infof("Associated Resources - [%+v]", workflowProject.AssociatedResources)
		})

		Step("Associate namespace to Project", func() {
			defer func() {
				workflowServiceAccount.SwitchToAdmin()
			}()
			workflowServiceAccount.SwitchToServiceAccount(projectUser)
			err := workflowProject.Associate(
				[]string{},
				[]string{workflowTargetCluster.ClusterUID},
				[]string{},
				[]string{},
				[]string{},
				[]string{},
			)
			// TODO: Error needs to be changed with actual error at the time of validation
			dash.VerifyFatal(err, err, "Associate namespace failed with project user account")
		})

		Step("Dissociate namespace from Project", func() {
			defer func() {
				workflowServiceAccount.SwitchToAdmin()
			}()
			workflowServiceAccount.SwitchToServiceAccount(projectUser)
			err := workflowProject.Dissociate(
				[]string{},
				[]string{workflowNamespace.Namespaces[namespace]},
				[]string{},
				[]string{},
				[]string{},
				[]string{},
			)
			// TODO: Error needs to be changed with actual error at the time of validation
			dash.VerifyFatal(err, err, "Dissociate namespace failed with project user account")
		})

		Step("Delete Project", func() {
			defer func() {
				workflowServiceAccount.SwitchToAdmin()
			}()
			workflowServiceAccount.SwitchToServiceAccount(projectUser)
			err := workflowProject.DeleteProject()
			// TODO: Error needs to be changed with actual error at the time of validation
			dash.VerifyFatal(err, err, "Delete Project failed with project user account")
		})

		Step("Delete Project", func() {
			defer func() {
				workflowServiceAccount.SwitchToAdmin()
			}()
			workflowServiceAccount.SwitchToServiceAccount(projectAdmin)
			err := workflowProject.DeleteProject()
			// TODO: Error needs to be changed with actual error at the time of validation
			dash.VerifyFatal(err, err, "Delete Project failed with project user account")
		})

	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
	})
})
