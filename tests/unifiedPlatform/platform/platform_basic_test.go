package tests

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	pdslib "github.com/portworx/torpedo/drivers/pds/lib"
	dsUtils "github.com/portworx/torpedo/drivers/unifiedPlatform/pdsLibs"
	platformUtils "github.com/portworx/torpedo/drivers/unifiedPlatform/platformLibs"
	"github.com/portworx/torpedo/pkg/aetosutil"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests"
	. "github.com/portworx/torpedo/tests/unifiedPlatform"
	"os"
	"testing"
)

var dash *aetosutil.Dashboard

var _ = BeforeSuite(func() {
	steplog := "Get prerequisite params to run platform tests"
	log.InfoD(steplog)
	Step(steplog, func() {
		// Read pds params from the configmap
		var err error
		pdsparams := pdslib.GetAndExpectStringEnvVar("PDS_PARAM_CM")
		NewPdsParams, err = ReadNewParams(pdsparams)
		log.FailOnError(err, "Failed to read params from json file")
		infraParams := NewPdsParams.InfraToTest
		PdsLabels["clusterType"] = infraParams.ClusterType

		log.InfoD("Get Account ID")
		AccID = "acc:e9272122-9d7e-4c20-b257-bee7ebb9561a"

		err = platformUtils.InitUnifiedApiComponents(os.Getenv(EnvControlPlaneUrl), "")
		log.FailOnError(err, "error while initialising api components")

		// accList, err := platformUtils.GetAccountListv1()
		// log.FailOnError(err, "error while getting account list")
		// accID = platformUtils.GetPlatformAccountID(accList, defaultTestAccount)
		log.Infof("AccountID - [%s]", AccID)

		err = platformUtils.InitUnifiedApiComponents(infraParams.ControlPlaneURL, AccID)
		log.FailOnError(err, "error while initialising api components")

		//Initialising UnifiedApiComponents in ds utils
		err = dsUtils.InitUnifiedApiComponents(infraParams.ControlPlaneURL, AccID)
		log.FailOnError(err, "error while initialising api components in ds utils")
	})

})

var _ = AfterSuite(func() {
	log.InfoD("Test Finished")
})

func TestDataService(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Torpedo : pds")
}

func TestMain(m *testing.M) {
	// call flag.Parse() here if TestMain uses flags
	ParseFlags()
	os.Exit(m.Run())
}
