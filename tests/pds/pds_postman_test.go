package tests

import (
	. "github.com/onsi/ginkgo"
	postmanLib "github.com/portworx/torpedo/drivers/postmanApiLoadDriver"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests"
	"time"
)

var _ = Describe("{RunPdsPostManApiLoadTests}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("RunPdsPostManApiLoadTests", "Run PDS Specific Api Load Tests using Postman-Newman ", pdsLabels, 0)
	})
	It("Deploy Dataservices", func() {
		Step("Starting to execute Postman-Newman on PDS", func() {
			var resultsFileName = "result_" + fmt.Sprint(time.Now().Unix()) + ".json"
			ctx, err := GetSourceClusterConfigPath()
			log.FailOnError(err, "failed while getting dest cluster path")
			postmanParams := postmanLib.PostmanDriver{
				ResultsFileName: resultsFileName,
				ResultType:      "cli,json",
				Namespace:       params.InfraToTest.Namespace,
				Iteration:       "2",
				Kubeconfig:      ctx}
			err = postmanLib.GetProjectNameToExecutePostman("pds", &postmanParams)
			log.FailOnError(err, "Postman-Newman Execution has failed due to- ")
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
	})
})
