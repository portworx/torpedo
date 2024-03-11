package tests

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	pdslib "github.com/portworx/torpedo/drivers/pds/lib"
	dsUtils "github.com/portworx/torpedo/drivers/unifiedPlatform/pdsLibs"
	platformUtils "github.com/portworx/torpedo/drivers/unifiedPlatform/platformLibs"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests"
	"os"
	"testing"
)

var _ = BeforeSuite(func() {
	steplog := "Get prerequisite params to run platform tests"
	log.InfoD(steplog)
	Step(steplog, func() {

		// Read pds params from the configmap
		var err error
		log.InfoD("Get Account ID")
		accID := "acc:2199f82a-9c39-4070-a431-4a8c8b1c2ca7"
		pdsparams := pdslib.GetAndExpectStringEnvVar("PDS_PARAM_CM")
		NewPdsParams, err := ReadNewParams(pdsparams)
		infraParams := NewPdsParams.InfraToTest
		pdsLabels["clusterType"] = infraParams.ClusterType
		log.FailOnError(err, "Failed to read params from json file")
		err = platformUtils.InitUnifiedApiComponents(os.Getenv(envControlPlaneUrl), "")
		log.FailOnError(err, "error while initialising api components")

		// accList, err := platformUtils.GetAccountListv1()
		// log.FailOnError(err, "error while getting account list")
		// accID = platformUtils.GetPlatformAccountID(accList, defaultTestAccount)
		log.Infof("AccountID - [%s]", accID)

		err = platformUtils.InitUnifiedApiComponents(infraParams.ControlPlaneURL, accID)
		log.FailOnError(err, "error while initialising api components")

		//Initialising UnifiedApiComponents in ds utils
		err = dsUtils.InitUnifiedApiComponents(infraParams.ControlPlaneURL, accID)
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
