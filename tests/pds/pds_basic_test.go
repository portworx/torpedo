package tests

import (
	"os"
	"testing"

	"github.com/portworx/torpedo/pkg/log"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
	pdslib "github.com/portworx/torpedo/drivers/pds/lib"
	. "github.com/portworx/torpedo/tests"
)

func TestDataService(t *testing.T) {
	RegisterFailHandler(Fail)

	var specReporters []Reporter
	junitReporter := reporters.NewJUnitReporter("/testresults/junit_basic.xml")
	specReporters = append(specReporters, junitReporter)
	RunSpecsWithDefaultAndCustomReporters(t, "Torpedo : pds", specReporters)

}

var _ = BeforeSuite(func() {
	steplog := "Get prerequisite params to run the pds tests"
	Step(steplog, func() {
		//InitInstance()
		log.InfoD(steplog)
		dash = Inst().Dash
		dash.TestSet.Product = "pds"
		dash.TestSetBegin(dash.TestSet)
		pdsLabels["pds"] = "true"
		pdsparams := pdslib.GetAndExpectStringEnvVar("PDS_PARAM_CM")
		params, err = pdslib.ReadParams(pdsparams)
		log.FailOnError(err, "Failed to read params from json file")
		infraParams := params.InfraToTest
		pdsLabels["clusterType"] = infraParams.ClusterType

		tenantID, dnsZone, projectID, serviceType, clusterID, err = SetupPDSTest(infraParams.ControlPlaneURL, infraParams.ClusterType, infraParams.AccountName, infraParams.TenantName, infraParams.ProjectName)
		log.InfoD("DeploymentTargetID %v ", deploymentTargetID)
		log.FailOnError(err, "Failed on SetupPDSTest method")
	})

	steplog = "Check and Register Target Cluster to ControlPlane"
	Step(steplog, func() {
		log.InfoD(steplog)
		infraParams := params.InfraToTest
		err = pdslib.RegisterClusterToControlPlane(infraParams.ControlPlaneURL, tenantID, infraParams.ClusterType)
		log.FailOnError(err, "Target Cluster Registeration failed")
	})

	steplog = "Get Deployment TargetID"
	Step(steplog, func() {
		log.InfoD(steplog)
		deploymentTargetID, err = pdslib.GetDeploymentTargetID(clusterID, tenantID)
		log.FailOnError(err, "Failed to get the deployment TargetID")
		log.InfoD("DeploymentTargetID %v ", deploymentTargetID)
	})

	steplog = "Get StorageTemplateID and Replicas"
	Step(steplog, func() {
		log.InfoD(steplog)
		storageTemplateID, err = pdslib.GetStorageTemplate(tenantID)
		log.FailOnError(err, "Failed while getting storage template ID")
		log.InfoD("storageTemplateID %v", storageTemplateID)
	})

	Step(steplog, func() {
		log.InfoD(steplog)
		namespace = params.InfraToTest.Namespace
		_, isavailable, err := pdslib.CreatePDSNamespace(namespace)
		log.FailOnError(err, "Error while Create/Get Namespaces")
		dash.VerifyFatal(bool(true), isavailable, "Namespace is not available for pds to deploy data services")
		namespaceID, err = pdslib.GetnameSpaceID(namespace, deploymentTargetID)
		log.FailOnError(err, "Error while getting namespace id")
		dash.VerifyFatal(namespaceID != "", true, "validating namespace ID")
	})
})

var _ = AfterSuite(func() {
	defer dash.TestSetEnd()
})

func TestMain(m *testing.M) {
	// call flag.Parse() here if TestMain uses flags
	ParseFlags()
	os.Exit(m.Run())
}
