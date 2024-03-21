package tests

import (
	. "github.com/onsi/ginkgo/v2"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/stworkflows"
	"github.com/portworx/torpedo/drivers/utilities"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests"
	. "github.com/portworx/torpedo/tests/unifiedPlatform"
	"strings"
)

var _ = Describe("{DeployDataServicesOnDemand}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("DeployDataService", "Deploy data services", nil, 0)
	})
	var (
		workflowDataservice stworkflows.WorkflowDataService
	)

	It("Deploy and Validate DataService", func() {

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
			_, err := workflowDataservice.DeployDataService(ds, true)
			log.FailOnError(err, "Error while deploying ds")
		}
	})

	It("Update DataService", func() {
		for _, ds := range NewPdsParams.DataServiceToTest {
			_, err := workflowDataservice.UpdateDataService(ds)
			log.FailOnError(err, "Error while updating ds")
		}
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
	})
})
