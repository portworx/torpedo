package tests

import (
	. "github.com/onsi/ginkgo/v2"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/stworkflows"
	"github.com/portworx/torpedo/drivers/utilities"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests"
	. "github.com/portworx/torpedo/tests/unifiedPlatform"
	"strconv"
	"strings"
)

var _ = Describe("{DeployDataServicesOnDemand}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("DeployDataService", "Deploy data services", nil, 0)
	})
	var (
		workflowDataservice stworkflows.WorkflowDataService
		workFlowTemplates   stworkflows.CustomTemplates
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
		tenantId := ""
		serviceConfigId, stConfigId, resConfigId, err := workFlowTemplates.CreatePdsCustomTemplatesAndFetchIds(tenantId, NewPdsParams)
		log.FailOnError(err, "Unable to create Custom Templates for PDS")
		//For Dummy test Only Will be removed once PDS build is avail
		resConfigId = strconv.Itoa(12)
		serviceConfigId = strconv.Itoa(12)
		stConfigId = strconv.Itoa(12)

		workflowDataservice.PDSTemplates.ServiceConfigTemplate["serviceConfigTempID"] = serviceConfigId
		workflowDataservice.PDSTemplates.StorageTemplate["storageTempID"] = stConfigId
		workflowDataservice.PDSTemplates.ResourceTemplate["resourceTempID"] = resConfigId
		for _, ds := range NewPdsParams.DataServiceToTest {
			workflowDataservice.Namespace = WorkflowNamespace
			workflowDataservice.NamespaceName = Namespace
			_, err := workflowDataservice.DeployDataService(ds, true)
			log.FailOnError(err, "Error while deploying ds")
		}
	})

	It("Update DataService", func() {
		for _, ds := range NewPdsParams.DataServiceToTest {
			_, err := workflowDataservice.UpdateDataService(ds, true)
			log.FailOnError(err, "Error while updating ds")
		}
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
	})
})

var _ = Describe("{ScaleUpCpuMemLimitsOfDS}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("ScaleUpCpuMemLimitsOfDS", "Deploy a dataservice and scale up its CPU/MEM limits by editing the respective template", nil, 0)
	})
	var (
		workflowDataservice stworkflows.WorkflowDataService
		workFlowTemplates   stworkflows.CustomTemplates
	)

	It("Deploy and Validate DataService", func() {
		tenantId := ""
		Step("Create a PDS Namespace", func() {
			Namespace = strings.ToLower("pds-test-ns-" + utilities.RandString(5))
			WorkflowNamespace.TargetCluster = WorkflowTargetCluster
			WorkflowNamespace.Namespaces = make(map[string]string)
			workflowNamespace, err := WorkflowNamespace.CreateNamespaces(Namespace)
			log.FailOnError(err, "Unable to create namespace")
			log.Infof("Namespaces created - [%s]", workflowNamespace.Namespaces)
			log.Infof("Namespace id - [%s]", workflowNamespace.Namespaces[Namespace])
		})

		serviceConfigId, stConfigId, resConfigId, err := workFlowTemplates.CreatePdsCustomTemplatesAndFetchIds(tenantId, NewPdsParams)
		log.FailOnError(err, "Unable to create Custom Templates for PDS")
		//For Dummy test Only Will be removed once PDS build is avail
		resConfigId = strconv.Itoa(12)
		serviceConfigId = strconv.Itoa(12)
		stConfigId = strconv.Itoa(12)

		workflowDataservice.PDSTemplates.ServiceConfigTemplate["serviceConfigTempID"] = serviceConfigId
		workflowDataservice.PDSTemplates.StorageTemplate["storageTempID"] = stConfigId
		workflowDataservice.PDSTemplates.ResourceTemplate["resourceTempID"] = resConfigId

		for _, ds := range NewPdsParams.DataServiceToTest {
			workflowDataservice.Namespace = WorkflowNamespace
			workflowDataservice.NamespaceName = Namespace
			_, err := workflowDataservice.DeployDataService(ds, true)
			log.FailOnError(err, "Error while deploying ds")
		}
		//create new templates with changed values of CPU Values -
		newCpuLimits, err := strconv.Atoi(NewPdsParams.ResourceConfiguration.CpuLimit)
		NewPdsParams.ResourceConfiguration.CpuLimit = string(newCpuLimits + 1)
		newCpuReq, err := strconv.Atoi(NewPdsParams.ResourceConfiguration.CpuLimit)
		NewPdsParams.ResourceConfiguration.CpuRequest = string(newCpuReq + 1)

		//create new templates with changed values of MEM Values -
		newMemLimits, err := strconv.Atoi(NewPdsParams.ResourceConfiguration.MemoryLimit)
		NewPdsParams.ResourceConfiguration.MemoryLimit = string(newMemLimits + 1)
		newMemReq, err := strconv.Atoi(NewPdsParams.ResourceConfiguration.MemoryLimit)
		NewPdsParams.ResourceConfiguration.MemoryRequest = string(newMemReq + 1)

		_, _, resConfigIdUpdated, err := workFlowTemplates.CreatePdsCustomTemplatesAndFetchIds(tenantId, NewPdsParams)
		log.FailOnError(err, "Unable to create Custom Templates for PDS")
		//For Dummy test Only Will be removed once PDS build is avail
		resConfigIdUpdated = strconv.Itoa(10)

		//Update Ds With New Values of Resource Templates
		workflowDataservice.PDSTemplates.ResourceTemplate["resourceTempID"] = resConfigIdUpdated
		for _, ds := range NewPdsParams.DataServiceToTest {
			_, err := workflowDataservice.UpdateDataService(ds, true)
			log.FailOnError(err, "Error while updating ds")
		}
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
	})
})
