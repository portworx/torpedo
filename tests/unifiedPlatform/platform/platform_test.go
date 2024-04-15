package tests

import (
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	"github.com/portworx/torpedo/drivers/unifiedPlatform"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/stworkflows/platform"
	"github.com/portworx/torpedo/drivers/utilities"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests"
	. "github.com/portworx/torpedo/tests/unifiedPlatform"
	"strings"
)

var _ = Describe("{PlatformBasicTest}", func() {
	var (
		workflowPlatform       platform.WorkflowPlatform
		workflowTargetCluster  platform.WorkflowTargetCluster
		workflowProject        platform.WorkflowProject
		workflowNamespace      platform.WorkflowNamespace
		workflowCloudCreds     platform.WorkflowCloudCredentials
		workflowBackupLocation platform.WorkflowBackupLocation
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
				workflowCloudCreds.CloudCredentials = make(map[string]platform.CloudCredentialsType)
				_, err := workflowCloudCreds.CreateCloudCredentials(NewPdsParams.BackUpAndRestore.TargetLocation)
				log.FailOnError(err, "Unable to create cloud credentials")
				log.InfoD("Cloud credentials created")
				for _, value := range workflowCloudCreds.CloudCredentials {
					log.InfoD("cloud credentials name: [%s]", value.Name)
					log.InfoD("cloud credentials id: [%s]", value.ID)
					log.InfoD("cloud provider type: [%s]", value.CloudProviderType)
				}
			}
		})

		Step("Create Backup Location", func() {
			//// TODO: This needs to be removed once API support is added for backup location
			if VARIABLE_FROM_JENKINS == unifiedPlatform.GRPC {
				workflowBackupLocation.WfCloudCredentials = workflowCloudCreds
				_, err := workflowBackupLocation.CreateBackupLocation(PDSBucketName, NewPdsParams.BackUpAndRestore.TargetLocation)
				log.FailOnError(err, "error while creating backup location")
				log.InfoD("wfBkpLoc id: [%s]", workflowBackupLocation.BkpLocation.BkpLocationId)
				log.InfoD("wfBkpLoc name: [%s]", workflowBackupLocation.BkpLocation.Name)
			}
		})

		Step("Create Project", func() {
			workflowProject.Platform = workflowPlatform
			workflowProject.ProjectName = fmt.Sprintf("project-%s", utilities.RandomString(5))
			workflowProject, err := workflowProject.CreateProject()
			log.FailOnError(err, "Unable to create project")
			log.InfoD("Project created with ID - [%s]", workflowProject.ProjectId)
		})

		Step("Register Target Cluster", func() {
			workflowTargetCluster.Project = workflowProject
			log.Infof("Tenant ID [%s]", workflowTargetCluster.Project.Platform.TenantId)
			workflowTargetCluster, err := workflowTargetCluster.RegisterToControlPlane(false)
			log.FailOnError(err, "Unable to register target cluster")
			log.InfoD("Target cluster registered with uid - [%s]", workflowTargetCluster.ClusterUID)
		})

		Step("Create a PDS Namespace", func() {
			workflowNamespace.TargetCluster = workflowTargetCluster
			workflowNamespace.Namespaces = make(map[string]string)
			_, err := workflowNamespace.CreateNamespaces(namespace)
			log.FailOnError(err, "Unable to create namespace")
			log.InfoD("Namespaces created - [%s]", workflowNamespace.Namespaces)
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
			log.InfoD("Associated Resources - [%+v]", workflowProject.AssociatedResources)
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
			log.InfoD("Dissociated Clusters - [%s]", workflowTargetCluster.ClusterUID)
			log.InfoD("Dissociated namespaces - [%s]", workflowNamespace.Namespaces[namespace])
			log.InfoD("Current Associated Resources - [%+v]", workflowProject.AssociatedResources)
		})

		Step("Delete PDS Namespace", func() {
			err := workflowNamespace.DeleteNamespace(namespace)
			log.FailOnError(err, "Unable to delete namespace")
			log.InfoD("Namespaces deleted - [%s]", namespace)
			log.Infof("Namespaces - [%+v]", workflowNamespace.Namespaces)
		})

		Step("Delete Project", func() {
			err := workflowProject.DeleteProject()
			log.FailOnError(err, "Delete project")
			log.InfoD("Project deleted successfully")
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
		workflowPlatform       platform.WorkflowPlatform
		workflowTargetCluster  platform.WorkflowTargetCluster
		workflowProject        platform.WorkflowProject
		workflowProject1       platform.WorkflowProject
		workflowProject2       platform.WorkflowProject
		workflowServiceAccount platform.WorkflowServiceAccount
		projectAdmin           string
		tenantAdmin            string
		user                   string
	)
	JustBeforeEach(func() {

		StartTorpedoTest("PlatformRBACTest", "Basic RBAC operations on platform", nil, 0)
		workflowPlatform.TenantInit()
		projectAdmin = "project-Admin-" + RandomString(5)
		tenantAdmin = "tenant-Admin-" + RandomString(5)
		user = "project-User-" + RandomString(5)

	})

	It("Basic RBAC operations on platform", func() {

		Step("Create Project", func() {
			workflowProject.Platform = workflowPlatform
			workflowProject.ProjectName = fmt.Sprintf("project-%s", utilities.RandomString(5))
			workflowProject, err := workflowProject.CreateProject()
			log.FailOnError(err, "Unable to create project")
			log.InfoD("Project created with ID - [%s]", workflowProject.ProjectId)
		})

		Step("Register Target Cluster", func() {
			workflowTargetCluster.Project = workflowProject
			log.Infof("Tenant ID [%s]", workflowTargetCluster.Project.Platform.TenantId)
			workflowTargetCluster, err := workflowTargetCluster.RegisterToControlPlane(false)
			log.FailOnError(err, "Unable to register target cluster")
			log.InfoD("Target cluster registered with uid - [%s]", workflowTargetCluster.ClusterUID)
		})

		Step("Create project user", func() {
			workflowServiceAccount.UserRoles = make(map[string]platform.SeviceAccount)
			workflowServiceAccount.WorkflowProject = workflowProject

			_, err := workflowServiceAccount.CreateServiceAccount(
				user,
				[]string{},
			)
			log.FailOnError(err, "Unable to create project user")
			log.InfoD("Created project user - [%s]", user)
		})

		Step("Create project admin user", func() {
			_, err := workflowServiceAccount.CreateServiceAccount(
				projectAdmin,
				[]string{platform.ProjectAdmin},
			)
			log.FailOnError(err, "Unable to create project Admin")
			log.InfoD("Created project admin - [%s]", projectAdmin)
		})

		Step("Create tenant admin", func() {
			_, err := workflowServiceAccount.CreateServiceAccount(
				tenantAdmin,
				[]string{platform.TenantAdmin},
			)
			log.FailOnError(err, "Unable to create tenant Admin")
			log.InfoD("Created tenant admin - [%s]", tenantAdmin)
		})

		Step("Create Project with User - Expected Failure", func() {
			defer func() {
				workflowServiceAccount.SwitchToAdmin()
			}()
			workflowServiceAccount.SwitchToServiceAccount(user)
			workflowProject1.Platform = workflowPlatform
			workflowProject1.ProjectName = fmt.Sprintf("project-%s", utilities.RandomString(5))
			_, err := workflowProject1.CreateProject()
			dash.VerifyFatal(strings.Contains(err.Error(), "403 Forbidden"), true, "Create Project with User - 403 Forbidden")
		})

		Step("Create Project with Project Admin - Expected Failure", func() {
			defer func() {
				workflowServiceAccount.SwitchToAdmin()
			}()
			workflowServiceAccount.SwitchToServiceAccount(projectAdmin)
			workflowProject1.Platform = workflowPlatform
			workflowProject1.ProjectName = fmt.Sprintf("project-%s", utilities.RandomString(5))
			_, err := workflowProject1.CreateProject()
			dash.VerifyFatal(strings.Contains(err.Error(), "403 Forbidden"), true, "Create Project with Admin - 403 Forbidden")
		})

		Step("Create Project with tenant admin", func() {
			defer func() {
				workflowServiceAccount.SwitchToAdmin()
			}()
			workflowServiceAccount.SwitchToServiceAccount(tenantAdmin)
			workflowProject2.Platform = workflowPlatform
			workflowProject2.ProjectName = fmt.Sprintf("project-%s", utilities.RandomString(5))
			_, err := workflowProject2.CreateProject()
			log.FailOnError(err, "Unable to create project as tenant admin")
			log.InfoD("Project created with ID - [%s] - Tenant Admin", workflowProject2.ProjectId)
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
			dash.VerifyFatal(strings.Contains(err.Error(), "403 Forbidden"), true, "Register Target Cluster with Admin - 403 Forbidden")
		})

		Step("Register Target Cluster - tenant admin", func() {
			defer func() {
				workflowServiceAccount.SwitchToAdmin()
			}()
			workflowServiceAccount.SwitchToServiceAccount(tenantAdmin)
			workflowTargetCluster.Project = workflowProject
			log.Infof("Tenant ID [%s]", workflowTargetCluster.Project.Platform.TenantId)
			_, err := workflowTargetCluster.RegisterToControlPlane(false)
			log.FailOnError(err, "Unable to register target cluster as tenant admin")
			log.InfoD("Target cluster registered - Tenant Admin")
		})

		Step("Associate target cluster to Project Admin", func() {
			defer func() {
				workflowServiceAccount.SwitchToAdmin()
			}()
			workflowServiceAccount.SwitchToServiceAccount(projectAdmin)
			err := workflowProject.Associate(
				[]string{workflowTargetCluster.ClusterUID},
				[]string{},
				[]string{},
				[]string{},
				[]string{},
				[]string{},
			)
			log.FailOnError(err, "Unable to associate cluster as project admin")
			log.InfoD("Associated target cluster - Project Admin")
		})

		Step("Dissociate target cluster from Project Admin", func() {
			defer func() {
				workflowServiceAccount.SwitchToAdmin()
			}()
			workflowServiceAccount.SwitchToServiceAccount(projectAdmin)
			err := workflowProject.Dissociate(
				[]string{workflowTargetCluster.ClusterUID},
				[]string{},
				[]string{},
				[]string{},
				[]string{},
				[]string{},
			)
			log.FailOnError(err, "Unable to dissociate cluster as project admin")
			log.InfoD("Dissociated target cluster - Project Admin")
		})

		Step("Associate target cluster to Project - Tenant Admin", func() {
			defer func() {
				workflowServiceAccount.SwitchToAdmin()
			}()
			workflowServiceAccount.SwitchToServiceAccount(tenantAdmin)
			err := workflowProject.Associate(
				[]string{workflowTargetCluster.ClusterUID},
				[]string{},
				[]string{},
				[]string{},
				[]string{},
				[]string{},
			)
			log.FailOnError(err, "Tenant admin is not able to associate resource to project")
			log.InfoD("Associated target cluster - Tenant Admin")
		})

		Step("Dissociate target cluster from Project - Tenant Admin", func() {
			defer func() {
				workflowServiceAccount.SwitchToAdmin()
			}()
			workflowServiceAccount.SwitchToServiceAccount(tenantAdmin)
			err := workflowProject.Dissociate(
				[]string{workflowTargetCluster.ClusterUID},
				[]string{},
				[]string{},
				[]string{},
				[]string{},
				[]string{},
			)
			log.FailOnError(err, "Tenant admin is not able to associate resource from project")
			log.InfoD("Dissociated target cluster - Tenant Admin")
		})

		Step("Associate target cluster to Project - User", func() {
			defer func() {
				workflowServiceAccount.SwitchToAdmin()
			}()
			workflowServiceAccount.SwitchToServiceAccount(user)
			err := workflowProject.Associate(
				[]string{workflowTargetCluster.ClusterUID},
				[]string{},
				[]string{},
				[]string{},
				[]string{},
				[]string{},
			)
			dash.VerifyFatal(strings.Contains(err.Error(), "403 Forbidden"), true, "Associate project with user - 403 Forbidden")
		})

		Step("Dissociate namespace from Project - User", func() {
			defer func() {
				workflowServiceAccount.SwitchToAdmin()
			}()
			workflowServiceAccount.SwitchToServiceAccount(user)
			err := workflowProject.Dissociate(
				[]string{workflowTargetCluster.ClusterUID},
				[]string{},
				[]string{},
				[]string{},
				[]string{},
				[]string{},
			)
			dash.VerifyFatal(strings.Contains(err.Error(), "403 Forbidden"), true, "Dissociate project with user - 403 Forbidden")
		})

		Step("Delete Project - Project Admin", func() {
			defer func() {
				workflowServiceAccount.SwitchToAdmin()
			}()
			workflowServiceAccount.SwitchToServiceAccount(projectAdmin)
			err := workflowProject.DeleteProject()
			log.FailOnError(err, "Unable to delete project as project Admin with access")
			log.InfoD("Project - [%s] - deleted with project admin", workflowProject.ProjectName)

			err = workflowProject2.DeleteProject() // Delete project without access
			dash.VerifyFatal(strings.Contains(err.Error(), "403 Forbidden"), true, "Delete project on another project - 403 Forbidden")
		})

		Step("Delete Project - Tenant Admin", func() {
			defer func() {
				workflowServiceAccount.SwitchToAdmin()
			}()
			workflowServiceAccount.SwitchToServiceAccount(tenantAdmin)
			err := workflowProject2.DeleteProject() // Delete project without access
			log.FailOnError(err, "Unable to delete project as tenant Admin")
			log.InfoD("Project - [%s] - deleted with tenant admin", workflowProject2.ProjectName)
		})

	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
	})
})
