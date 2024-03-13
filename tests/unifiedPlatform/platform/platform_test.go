package tests

import (
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/stworkflows"
	"github.com/portworx/torpedo/drivers/utilities"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests"
)

var _ = Describe("{PlatformOnboardingTest}", func() {
	var (
		workflowPlatform      stworkflows.WorkflowPlatform
		workflowTargetCluster stworkflows.WorkflowTargetCluster
		workflowProject       stworkflows.WorkflowProject
		workflowNamespace     stworkflows.WorkflowNamespace
		namespace             string
	)
	JustBeforeEach(func() {

		StartTorpedoTest("TestingWorkflowToOnboardAccounts", "Onboard Accounts", nil, 0)
		namespace = fmt.Sprintf("pds-namespace-%s", utilities.RandomString(5))
		workflowPlatform.Accounts = map[string]map[string]string{
			"testAccount1": map[string]string{
				automationModels.UserName:        "testAccount1",
				automationModels.UserDisplayName: "testAccount1",
				automationModels.UserEmail:       "atrivedi+1@purestorage.com",
			},
		}
	})

	It("Onboard accounts", func() {
		Step("Onboarding Accounts", func() {
			// TODO: This needs to be enabled once https://portworx.atlassian.net/browse/DS-8552 is fixed
			//workflowResponse, err := workflowPlatform.OnboardAccounts()
			//log.FailOnError(err, "Unable to create accounts")
			//log.Infof("All Account onboarded Successfully, Response - [%+v]", workflowResponse)
			log.Infof("Initilaising values for tenant")
			workflowPlatform.TenantInit()
		})

		Step("Create Project", func() {
			workflowProject.Platform = workflowPlatform
			workflowProject.ProjectName = fmt.Sprintf("project-%s", utilities.RandomString(5))
			workflowProject, err := workflowProject.CreateProject()
			log.FailOnError(err, "Unable to create project")
			log.Infof("Project created with ID - [%s]", workflowProject.ProjectId)
		})

		Step("Register Target Cluster", func() {
			workflowTargetCluster.Platform = workflowPlatform
			workflowTargetCluster.Project = workflowProject
			log.Infof("Tenant ID [%s]", workflowTargetCluster.Platform.TenantId)
			workflowTargetCluster, err := workflowTargetCluster.RegisterToControlPlane()
			log.FailOnError(err, "Unable to register target cluster")
			log.Infof("Target cluster registered with uid - [%s]", workflowTargetCluster.ClusterUID)
		})

		Step("Create a PDS Namespace", func() {
			workflowNamespace.TargetCluster = workflowTargetCluster
			workflowNamespace.Namespaces = make(map[string]map[string]string)
			_, err := workflowNamespace.CreateNamespaces(namespace)
			log.FailOnError(err, "Unable to create namespace")
			log.Infof("Namespaces created - [%s]", workflowNamespace.Namespaces)
		})

		Step("Associate cluster to Project", func() {
			err := workflowProject.Associate(
				[]string{workflowTargetCluster.ClusterUID},
				[]string{},
				[]string{},
				[]string{},
				[]string{},
				[]string{},
			)
			log.FailOnError(err, "Associated Cluster to Project")
			log.Infof("Associated Clusters - [%s]", workflowProject.AssociatedResources.Clusters)
		})

		Step("Dissociate cluster from Project", func() {
			err := workflowProject.Dissociate(
				[]string{workflowTargetCluster.ClusterUID},
				[]string{},
				[]string{},
				[]string{},
				[]string{},
				[]string{},
			)
			log.FailOnError(err, "Dissociated Cluster from Project")
			log.Infof("Dissociated Clusters - [%s]", workflowTargetCluster.ClusterUID)
			log.Infof("Current associated Clusters - [%s]", workflowProject.AssociatedResources.Clusters)
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
