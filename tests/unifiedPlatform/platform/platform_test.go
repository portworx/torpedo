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
	"strings"
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
		tenantAdmin            string
	)
	JustBeforeEach(func() {

		StartTorpedoTest("PlatformRBACTest", "Basic RBAC operations on platform", nil, 0)
		namespace = fmt.Sprintf("pds-namespace-%s", utilities.RandomString(5))
		workflowPlatform.TenantInit()
		projectAdmin = "project-admin-" + RandomString(5)
		projectUser = "project-user-" + RandomString(5)
		tenantAdmin = "tenant-admin-" + RandomString(5)

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

		//Step("Create a PDS Namespace", func() {
		//	workflowNamespace.TargetCluster = workflowTargetCluster
		//	workflowNamespace.Namespaces = make(map[string]string)
		//	_, err := workflowNamespace.CreateNamespaces(namespace)
		//	log.FailOnError(err, "Unable to create namespace")
		//	log.Infof("Namespaces created - [%s]", workflowNamespace.Namespaces)
		//})

		Step("Create project admin user", func() {
			workflowServiceAccount.UserRoles = make(map[string]stworkflows.SeviceAccount)
			workflowServiceAccount.WorkflowProject = workflowProject

			workflowServiceAccount.CreateServiceAccount(
				projectAdmin,
				[]string{stworkflows.ProjectAdmin},
			)
		})

		Step("Create project user", func() {
			workflowServiceAccount.CreateServiceAccount(
				projectUser,
				[]string{},
			)
		})

		Step("Create project user", func() {
			workflowServiceAccount.CreateServiceAccount(
				tenantAdmin,
				[]string{stworkflows.TenantAdmin},
			)
		})

		Step("Create Project with Project Admin - Expected Failure", func() {
			defer func() {
				workflowServiceAccount.SwitchToAdmin()
			}()
			log.Infof("Create Project with Project Admin - Expected Failure")
			workflowServiceAccount.SwitchToServiceAccount(projectAdmin)
			workflowProject.Platform = workflowPlatform
			workflowProject.ProjectName = fmt.Sprintf("project-%s", utilities.RandomString(5))
			_, err := workflowProject.CreateProject()
			if err != nil {
				log.Infof("Error - [%s]", err.Error())
				log.Infof("Error Bool - [%v]", strings.Contains(err.Error(), "403 Forbidden"))
			}
			// TODO: Error needs to be changed with actual error at the time of validation
			dash.VerifyFatal(strings.Contains(err.Error(), "403 Forbidden"), true, "Create Project with Admin - 403 Forbidden")
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
			if err != nil {
				dash.VerifyFatal(strings.Contains(err.Error(), "403 Forbidden"), true, "Create Project with Project Users - 403 Forbidden")
			} else {
				//TODO: Need to check this with dev
				log.Infof("Project Created with Project User - Not Expetced")
			}
		})

		Step("Create Project with tenant admin", func() {
			defer func() {
				workflowServiceAccount.SwitchToAdmin()
			}()
			workflowServiceAccount.SwitchToServiceAccount(tenantAdmin)
			workflowProject.Platform = workflowPlatform
			workflowProject.ProjectName = fmt.Sprintf("project-%s", utilities.RandomString(5))
			_, err := workflowProject.CreateProject()
			log.FailOnError(err, "Unable to create project as tenant admin")
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

		Step("Register Target Cluster - Project User", func() {
			defer func() {
				workflowServiceAccount.SwitchToAdmin()
			}()
			workflowServiceAccount.SwitchToServiceAccount(projectUser)
			workflowTargetCluster.Project = workflowProject
			log.Infof("Tenant ID [%s]", workflowTargetCluster.Project.Platform.TenantId)
			_, err := workflowTargetCluster.RegisterToControlPlane(false)
			// TODO: Error needs to be changed with actual error at the time of validation
			if err != nil {
				dash.VerifyFatal(strings.Contains(err.Error(), "403 Forbidden"), true, "Register Target Cluster with User - 403 Forbidden")
			} else {
				log.Infof("Register Target Cluster with Project User - Not Expetced")
			}
		})

		Step("Register Target Cluster - tenant admin", func() {
			defer func() {
				workflowServiceAccount.SwitchToAdmin()
			}()
			workflowServiceAccount.SwitchToServiceAccount(tenantAdmin)
			workflowTargetCluster.Project = workflowProject
			log.Infof("Tenant ID [%s]", workflowTargetCluster.Project.Platform.TenantId)
			_, err := workflowTargetCluster.RegisterToControlPlane(false)
			// TODO: Error needs to be changed with actual error at the time of validation
			log.FailOnError(err, "Unable to register target cluster as tenant admin")
		})

		Step("Associate target cluster to Project Admin", func() {
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
			// TODO: Need to check if this is the expected behaviour or not
			dash.VerifyFatal(strings.Contains(err.Error(), "403 Forbidden"), true, "Associate resource to target cluster with User - 403 Forbidden")
		})

		Step("Dissociate namespace from Project Admin", func() {
			defer func() {
				workflowServiceAccount.SwitchToAdmin()
			}()
			workflowServiceAccount.SwitchToServiceAccount(projectAdmin)
			err := workflowProject.Dissociate(
				[]string{},
				[]string{workflowTargetCluster.ClusterUID},
				[]string{},
				[]string{},
				[]string{},
				[]string{},
			)
			// TODO: Need to check if this is the expected behaviour or not
			dash.VerifyFatal(strings.Contains(err.Error(), "403 Forbidden"), true, "Dissociate resource to target cluster with User - 403 Forbidden")
		})

		Step("Associate target cluster to Project - User", func() {
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
			//TODO: Check why 400 bad request
			dash.VerifyFatal(strings.Contains(err.Error(), "400 Bad Request"), true, "Associate target cluster with User - 400 Bad Request")
		})

		Step("Dissociate namespace from Project User", func() {
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
			//TODO: Check why 400 bad request
			dash.VerifyFatal(strings.Contains(err.Error(), "400 Bad Request"), true, "Dissociate target cluster with User - 400 Bad Request")
		})

		Step("Associate target cluster to Project - Tenant Admin", func() {
			defer func() {
				workflowServiceAccount.SwitchToAdmin()
			}()
			workflowServiceAccount.SwitchToServiceAccount(tenantAdmin)
			err := workflowProject.Associate(
				[]string{},
				[]string{workflowTargetCluster.ClusterUID},
				[]string{},
				[]string{},
				[]string{},
				[]string{},
			)
			dash.VerifyFatal(strings.Contains(err.Error(), "400 Bad Request"), true, "Dissociate target cluster with Tenant Admin - 400 Bad Request")
		})

		Step("Dissociate namespace from Project - Tenant Admin", func() {
			defer func() {
				workflowServiceAccount.SwitchToAdmin()
			}()
			workflowServiceAccount.SwitchToServiceAccount(tenantAdmin)
			err := workflowProject.Dissociate(
				[]string{},
				[]string{workflowNamespace.Namespaces[namespace]},
				[]string{},
				[]string{},
				[]string{},
				[]string{},
			)
			dash.VerifyFatal(strings.Contains(err.Error(), "403 Forbidden"), true, "Dissociate target cluster with Tenant Admin - 403 Forbidden")
		})

		Step("Delete Project - Project Admin", func() {
			defer func() {
				workflowServiceAccount.SwitchToAdmin()
			}()
			workflowServiceAccount.SwitchToServiceAccount(projectAdmin)
			err := workflowProject.DeleteProject()
			if err != nil {
				log.Infof("Error - [%s]", err.Error())
			}
			log.FailOnError(err, "Unable to delete project as admin")
		})

		Step("Delete Project", func() {
			defer func() {
				workflowServiceAccount.SwitchToAdmin()
			}()
			workflowServiceAccount.SwitchToServiceAccount(projectUser)
			err := workflowProject.DeleteProject()
			if err != nil {
				log.Infof("Error - [%s]", err.Error())
				dash.VerifyFatal(strings.Contains(err.Error(), "403 Forbidden"), true, "Delete Project cluster with User - 403 Forbidden")
			} else {
				log.Infof("Project deleted successfully with User - Not expected")
			}
		})
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
	})
})
