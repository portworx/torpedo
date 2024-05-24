package tests

import (
	"fmt"
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
	"strings"
	"testing"
)

var dash *aetosutil.Dashboard

var _ = BeforeSuite(func() {
	InitInstance()
	dash = Inst().Dash
	dash.TestSet.Product = "pds"
	dash.TestSetBegin(dash.TestSet)

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
		AccID = "acc:e593e80c-9142-4286-91e5-76dc8bb9b4d6"

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

	Step("Dumping kubeconfigs file", func() {
		kubeconfigs := os.Getenv("KUBECONFIGS")
		if kubeconfigs != "" {
			kubeconfigList := strings.Split(kubeconfigs, ",")
			if len(kubeconfigList) < 2 {
				log.FailOnError(fmt.Errorf("At least minimum two kubeconfigs required but has"),
					"Failed to get k8s config path.At least minimum two kubeconfigs required")
			}
			DumpKubeconfigs(kubeconfigList)
		}
	})

})

var _ = AfterSuite(func() {
	// TODO: Add platform cleanup for these tests
	defer Inst().Dash.TestSetEnd()
	defer EndTorpedoTest()
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
