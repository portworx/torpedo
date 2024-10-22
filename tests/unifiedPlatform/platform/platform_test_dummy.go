package tests

import (
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	"github.com/pure-px/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/pure-px/torpedo/drivers/unifiedPlatform/platformLibs"
	"github.com/pure-px/torpedo/drivers/unifiedPlatform/stworkflows/platform"
	"github.com/pure-px/torpedo/drivers/utilities"
	"github.com/pure-px/torpedo/pkg/log"
	. "github.com/pure-px/torpedo/tests"
	. "github.com/pure-px/torpedo/tests/unifiedPlatform"
	"time"
)

var _ = Describe("{CreateAndGeBackupLocation}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("CreateAndGeBackupLocation", "create backup locations", nil, 0)
	})

	It("CreateAndGeBackupLocation", func() {
		Step("create credentials and backup location", func() {
			var (
				workflowCc     platform.WorkflowCloudCredentials
				workflowbkpLoc platform.WorkflowBackupLocation
			)

			tenantId, err := platformLibs.GetDefaultTenantId(AccID)
			log.FailOnError(err, "error occured while fetching tenantID")

			workflowCc.Platform.TenantId = tenantId
			workflowCc.CloudCredentials = make(map[string]platform.CloudCredentialsType)
			cc, err := workflowCc.CreateCloudCredentials(NewPdsParams.BackUpAndRestore.TargetLocation)
			log.FailOnError(err, "error occured while creating cloud credentials")

			for _, value := range cc.CloudCredentials {
				log.Infof("cloud credentials name: [%s]", value.Name)
				log.Infof("cloud credentials id: [%s]", value.ID)
				log.Infof("cloud provider type: [%s]", value.CloudProviderType)
			}

			workflowbkpLoc.WfCloudCredentials = workflowCc

			wfbkpLoc, err := workflowbkpLoc.CreateBackupLocation(PDSBucketName, NewPdsParams.BackUpAndRestore.TargetLocation)
			log.FailOnError(err, "error while creating backup location")
			log.Infof("wfBkpLoc id: [%s]", wfbkpLoc.BkpLocation.BkpLocationId)
			log.Infof("wfBkpLoc name: [%s]", wfbkpLoc.BkpLocation.Name)

			// Listing backuplocation after creation
			bkpLocations, err := workflowbkpLoc.ListBackupLocation()
			log.FailOnError(err, "error while listing backup location")

			for _, bkpLocation := range bkpLocations {
				log.Infof("wfBkpLoc Name: [%s]", bkpLocation.BkpLocation.Name)
				log.Infof("wfBkpLoc Id: [%s]", bkpLocation.BkpLocation.BkpLocationId)
				for _, cred := range bkpLocation.WfCloudCredentials.CloudCredentials {
					log.Infof("credentials Id: [%s]", cred.ID)
				}
			}

			//Deleting the backuplocation ID
			log.Infof("Deleting backuplocation id [%s]", wfbkpLoc.BkpLocation.BkpLocationId)
			err = wfbkpLoc.DeleteBackupLocation(wfbkpLoc.BkpLocation.BkpLocationId)
			log.FailOnError(err, "error while deleting backup location")

			// Listing backup location after deletion
			bkpLocations, err = workflowbkpLoc.ListBackupLocation()
			log.FailOnError(err, "error while listing backup location")

			for _, bkpLocation := range bkpLocations {
				log.Infof("wfBkpLoc Name: [%s]", bkpLocation.BkpLocation.Name)
				log.Infof("wfBkpLoc Id: [%s]", bkpLocation.BkpLocation.BkpLocationId)
				for _, cred := range bkpLocation.WfCloudCredentials.CloudCredentials {
					log.Infof("credentials Id: [%s]", cred.ID)
				}
			}
		})
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
	})
})

var _ = Describe("{CreateAndGetCloudCredentials}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("CreateCloudCredentials", "create cloud credentials", nil, 0)
	})

	It("CreateCloudCredentials", func() {
		Step("create and cloud credentials", func() {
			tenantId, err := platformLibs.GetDefaultTenantId(AccID)
			log.FailOnError(err, "error occured while fetching tenantID")
			credResp, err := platformLibs.CreateCloudCredentials(tenantId, NewPdsParams.BackUpAndRestore.TargetLocation)
			log.FailOnError(err, "error while creating cloud creds")
			log.Infof("creds resp [%+v]", credResp.Create.Config.S3Credentials.AccessKey)
			log.Infof("creds id [%+v]", *credResp.Create.Meta.Uid)

			isconfigRequiredTrue, err := platformLibs.GetCloudCredentials(*credResp.Create.Meta.Uid, NewPdsParams.BackUpAndRestore.TargetLocation, true)
			log.FailOnError(err, "error occured while getting cloud required with false flag")
			log.Debugf("Cred Name [%+v]", *isconfigRequiredTrue.Create.Meta.Name)
			log.Debugf("Cred Id [%+v]", *isconfigRequiredTrue.Create.Meta.Uid)
		})
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
	})
})

var _ = Describe("{ScheduleDummyTest}", func() {
	var (
		workflowPlatform platform.WorkflowPlatform
		schedules        platform.WorkflowSchedule
		scheduleName     string
	)
	JustBeforeEach(func() {

		StartTorpedoTest("ScheduleDummyTest", "Delete and Purge Schedule", nil, 0)
		scheduleName = fmt.Sprintf("pds-schedule-%s", utilities.RandomString(5))
		workflowPlatform.TenantInit()
		schedules.Platform = &workflowPlatform
		schedules.Schedules = make(map[string]*automationModels.V1BackupPolicy)
	})

	It("Delete and Purge Namespaces", func() {

		Step("Create Schedule", func() {
			scheduleCreated, err := schedules.CreateSchedule(scheduleName, 10*time.Minute, "1")
			log.FailOnError(err, "Unable to create schedule")
			log.InfoD("Schedule created with uid - [%s]", *scheduleCreated.Meta.Uid)
		})

		Step("Get Schedule", func() {
			scheduleGet, err := schedules.GetSchedule(*schedules.Schedules[scheduleName].Meta.Uid)
			log.FailOnError(err, "Unable to get schedule")
			log.InfoD("Schedule retrieved with uid - [%s]", *scheduleGet.Meta.Uid)
		})

		Step("Purge", func() {
			err := schedules.Purge()
			log.FailOnError(err, "Unable to purge schedules")
			log.InfoD("Schedules purged successfully")
		})

	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
	})
})
