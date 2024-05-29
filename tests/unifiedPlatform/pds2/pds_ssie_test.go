package tests

import (
	. "github.com/onsi/ginkgo/v2"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests"
	. "github.com/portworx/torpedo/tests/unifiedPlatform"
	"strconv"
	"time"
)

var _ = Describe("{PerformStorageResizeBy1GbnTimes}", func() {
	var (
		deployment           *automationModels.PDSDeploymentResponse
		templatesToBeDeleted []string
		err                  error
	)

	JustBeforeEach(func() {
		StartPDSTorpedoTest("PerformStorageResizeBy1GbnTimes", "Perform PVC Resize by 1GB for n times in a loop and validate the updated vol in the storage config.", nil, 0)
	})

	It("Deploy,Validate and ScaleUp DataService", func() {

		for _, ds := range NewPdsParams.DataServiceToTest {
			Step("Deploy DataService", func() {
				WorkflowDataService.SkipValidatation = make(map[string]bool)
				WorkflowDataService.SkipValidatation["VALIDATE_PDS_WORKLOADS"] = true
				deployment, err = WorkflowDataService.DeployDataService(ds, ds.Image, ds.Version, PDS_DEFAULT_NAMESPACE)
				log.FailOnError(err, "Error while deploying ds")
				log.Debugf("Source Deployment Id: [%s]", *deployment.Create.Meta.Uid)
			})

			for i := 1; i <= NewPdsParams.StorageConfigurationsSSIE.Iterations; i++ {
				//Update Ds With New Values of Resource Templates
				NewPdsParams.ResourceConfiguration.New_Storage_Request = strconv.Itoa(i+1) + "G"
				log.Debugf("Updating the Storage to: [%s]", NewPdsParams.ResourceConfiguration.New_Storage_Request)

				resConfigIdUpdated, err := WorkflowPDSTemplate.CreateResourceTemplateWithCustomValue(NewPdsParams)
				log.FailOnError(err, "Unable to create Custom Templates for PDS")
				log.InfoD("Updated Resource Template ID- [updated- %v]", resConfigIdUpdated)
				templatesToBeDeleted = append(templatesToBeDeleted, resConfigIdUpdated)
				log.Infof("Associate newly created template to the project")
				err = WorkflowProject.Associate(
					[]string{},
					[]string{},
					[]string{},
					[]string{},
					[]string{resConfigIdUpdated},
					[]string{},
				)
				log.FailOnError(err, "Unable to associate Templates to Project")
				log.Infof("Associated Resources - [%+v]", WorkflowProject.AssociatedResources)

				beforeResizePodAge, err := WorkflowDataService.GetPodAgeForDeployment(*deployment.Create.Meta.Uid)
				log.FailOnError(err, "unable to get pods AGE before Storage resize")
				log.InfoD("Pods Age before storage resize is- [%v]Min", beforeResizePodAge)

				//sleeping 30sec before the next update
				time.Sleep(30 * time.Second)

				WorkflowDataService.UpdateDeploymentTemplates = true
				WorkflowDataService.PDSTemplates = WorkflowPDSTemplate
				_, err = WorkflowDataService.UpdateDataService(ds, *deployment.Create.Meta.Uid, ds.Image, ds.Version)
				log.FailOnError(err, "Error while updating ds")

				afterResizePodAge, err := WorkflowDataService.GetPodAgeForDeployment(*deployment.Create.Meta.Uid)
				log.FailOnError(err, "unable to get pods AGE before Storage resize")
				log.InfoD("Pods Age after storage resize is- [%v]Min", afterResizePodAge)

				dash.VerifyFatal(afterResizePodAge > beforeResizePodAge, true, "Validating if the pod restarted after storage size increase")
			}

			stepLog := "Running Workloads after StorageSizeIncrease of DataService"
			Step(stepLog, func() {
				_, err := WorkflowDataService.RunDataServiceWorkloads(*deployment.Create.Meta.Uid)
				log.FailOnError(err, "Error while running workloads on ds")
			})
		}
	})

	JustAfterEach(func() {
		defer EndPDSTorpedoTest()
		//TODO: Remove this once https://purestorage.atlassian.net/browse/DS-9648 this is fixed
		//err = WorkflowDataService.PDSTemplates.DeleteCreatedCustomPdsTemplates(templatesToBeDeleted)
		//log.FailOnError(err, "Error while deleting the templates")
	})
})
