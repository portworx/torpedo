package tests

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	pdsbkp "github.com/portworx/torpedo/drivers/pds/pdsbackup"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests"
)

const (
	bkpTargetName = "bkptarget-automation"
)

var (
	bkpClient              *pdsbkp.BackupClient
	awsBkpTargets          []*pds.ModelsBackupTarget
	azureBkpTargets        []*pds.ModelsBackupTarget
	s3CompatibleBkpTargets []*pds.ModelsBackupTarget
)

var _ = Describe("{ValidateBackupTargetsOnSupportedObjectStores}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("ValidateBackupTargetsOnSupportedObjectStores", "Validate backup targets for all supported object stores.", nil, 0)
		bkpClient, err = pdsbkp.InitializePdsBackup()
		log.FailOnError(err, "Failed to initialize backup for pds.")
	})

	It("add all supported Object stores as backup target for data services", func() {
		Step("Create AWS S3 Backup target.", func() {
			bkpTarget, err := bkpClient.CreateAwsS3BackupCredsAndTarget(tenantID, fmt.Sprintf("%v-aws", bkpTargetName))
			log.FailOnError(err, "Failed to create AWS backup target.")
			log.InfoD("AWS S3 target - %v created successfully", bkpTarget.GetName())
			awsBkpTargets = append(awsBkpTargets, bkpTarget)
		})
		Step("Create Azure(blob) Backup target.", func() {
			bkpTarget, err := bkpClient.CreateAzureBackupCredsAndTarget(tenantID, fmt.Sprintf("%v-azure", bkpTargetName))
			log.FailOnError(err, "Failed to create Azure backup target.")
			log.InfoD("Azure backup target - %v created successfully", bkpTarget.GetName())
			azureBkpTargets = append(azureBkpTargets, bkpTarget)
		})
		Step("Create S3 compatible(Minio) Backup target.", func() {
			bkpTarget, err := bkpClient.CreateAwsS3BackupCredsAndTarget(tenantID, fmt.Sprintf("%v-s3compatible", bkpTargetName))
			log.FailOnError(err, "Failed to create AWS backup target.")
			log.InfoD("S3 compatible backup target - %v created successfully", bkpTarget.GetName())
			s3CompatibleBkpTargets = append(s3CompatibleBkpTargets, bkpTarget)
		})
	})
	JustAfterEach(func() {
		for _, bkptarget := range awsBkpTargets {
			log.FailOnError(bkpClient.DeleteAwsS3BackupCredsAndTarget(bkptarget.GetId()), "Failed while deleting the Aws backup target")
		}
		for _, bkptarget := range azureBkpTargets {
			log.FailOnError(bkpClient.DeleteAzureBackupCredsAndTarget(bkptarget.GetId()), "Failed while deleting the Azure backup target")
		}
		for _, bkptarget := range s3CompatibleBkpTargets {
			log.FailOnError(bkpClient.DeleteS3CompatibleBackupCredsAndTarget(bkptarget.GetId()), "Failed while deleting the S3 compatible backup target")
		}
	})
})
