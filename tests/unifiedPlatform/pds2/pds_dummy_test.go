package tests

import (
	"context"
	. "github.com/onsi/ginkgo/v2"
	"github.com/portworx/torpedo/drivers/applications/hammerdb"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	pdslibs "github.com/portworx/torpedo/drivers/unifiedPlatform/pdsLibs"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/stworkflows"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/stworkflows/pds"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests"
	tests "github.com/portworx/torpedo/tests/unifiedPlatform"
)

var ctx = context.Background()

var _ = Describe("{CleanUpDeployments}", func() {
	It("Delete all deployments", func() {
		err := pdslibs.DeleteAllDeployments(tests.ProjectId)
		log.Errorf("ERROR WHILE DELETING DEPLOYMENT [%v]", err)
		//log.FailOnError(err, "error while deleting deployment")

	})
})

var _ = Describe("{ValidateWorkloads}", func() {

	JustBeforeEach(func() {
		tests.StartPDSTorpedoTest("ValidateWorkloads", "validate  workloads", nil, 0)
	})
	It("Validate Workloads", func() {
		//stepLog := "Running Workloads before upgrading the ds image"
		//Step(stepLog, func() {
		//	err := tests.WorkflowDataService.RunDataServiceWorkloads(tests.NewPdsParams, "postgresql")
		//	log.FailOnError(err, "Error while running workloads on ds")
		//})

	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
	})
})

var _ = Describe("{RunNeo4jQuery}", func() {

	var (
		pgDeployment    *automationModels.PDSDeploymentResponse
		neo4jDeployment *automationModels.PDSDeploymentResponse
		err             error
	)

	JustBeforeEach(func() {
		tests.StartPDSTorpedoTest("ValidateQueryRuns", "Validating query runs", nil, 0)
	})

	It("Run Query", func() {
		Step("Execute Command", func() {

			for _, ds := range tests.NewPdsParams.DataServiceToTest {
				if ds.Name == postgresql {
					Step("Deploy PG dataservice and Export Data", func() {
						pgDeployment, err = tests.WorkflowDataService.DeployDataService(ds, ds.Image, ds.Version, tests.PDS_DEFAULT_NAMESPACE)
						log.FailOnError(err, "Error while deploying ds")
						log.Infof("All deployments - [%+v]", tests.WorkflowDataService.DataServiceDeployment)
					})

					//Export Data from relation db
					err = tests.WorkflowDataService.ExportDataFromRelationalDB(postgresql, *pgDeployment.Create.Meta.Uid, tests.PDS_DEFAULT_NAMESPACE)
					log.FailOnError(err, "error while exporting the data")
				}
			}

			for _, ds := range tests.NewPdsParams.DataServiceToTest {
				if ds.Name == neo4j {
					Step("Deploy Neo4j dataservice and Import Data", func() {
						neo4jDeployment, err = tests.WorkflowDataService.DeployDataService(ds, ds.Image, ds.Version, tests.PDS_DEFAULT_NAMESPACE)
						log.FailOnError(err, "Error while deploying ds")
						log.Infof("All deployments - [%+v]", tests.WorkflowDataService.DataServiceDeployment)

						//Import Data to Neo4j db
						results, err := tests.WorkflowDataService.ImportDataToNeo4jDB(*neo4jDeployment.Create.Meta.Uid,
							tests.PDS_DEFAULT_NAMESPACE, stworkflows.TestDataSetSrcPath, stworkflows.TestDataSetDestPath,
							stworkflows.TestDataQuerySrcPath, stworkflows.TestDataQueryDestPath, "NEXT")
						log.FailOnError(err, "error while importing the data")
						dash.VerifyFatal(results[0], "19", "validating the relationship count after importing the csv file")
					})
				}
			}

			//results, err := tests.WorkflowDataService.ImportDataToNeo4jDB(*deployment.Create.Meta.Uid, tests.PDS_DEFAULT_NAMESPACE)
			//log.FailOnError(err, "error while importing the data")
			//
			//dash.VerifyFatal(results[0], "14446", "validating the relationship count after importing the csv file")

		})
	})

	JustAfterEach(func() {
		EndTorpedoTest()
	})
})

var _ = Describe("{DeployHammerDB}", func() {
	//
	var (
		hammer *hammerdb.HammerDB
	)

	JustBeforeEach(func() {
		tests.StartPDSTorpedoTest("DeployHammerDB", "Deploy HammerDB pod", nil, 0)
	})

	It("Deploy,Validate and ScaleUp DataService", func() {
		Step("Deploy HammerDB Pod", func() {
			hammer = hammerdb.HammerDBInit(
				"hammerdb",
				"sql-restore-uueri-itm13p-restore-uueri-rw.restore-uueri.svc.cluster.local",
				"pds",
				"fV2juAq5rFdhz1EtFpaPIdj9079Cd3YYZq0POe4L",
				"tpccmsql",
				"mssql",
				1)

			err := hammer.DeployHammerDBPod()
			log.FailOnError(err, "Error while deploying hammerdb pod")
		})
		Step("Copy file to hammerdb pod", func() {
			log.Infof("Running hammerdb load")
			out, err := hammer.RunHammerDBLoad()
			log.FailOnError(err, "Error while running hammerdb load")
			log.Infof("Output - [%s]", out)
		})
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
	})

})

var _ = Describe("{ValidateDnsEndPoint}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("ValidateDnsEndPoint", "validate dns endpoint", nil, 0)
	})

	var (
		workflowDataservice pds.WorkflowDataService
		err                 error
	)

	It("ValidateDnsEndPoint", func() {
		Step("validate dns endpoint", func() {
			depId := "dep:3a13954f-ae45-4223-8896-82029c90bca9"
			err = workflowDataservice.ValidateDNSEndpoint(depId)
			log.FailOnError(err, "Error occurred while validating dns endpoint")
		})
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
	})
})
