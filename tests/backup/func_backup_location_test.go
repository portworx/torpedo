package tests

import (
	"context"
	"fmt"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	"github.com/pborman/uuid"
	"github.com/pure-px/torpedo/drivers/backup"
	"github.com/pure-px/torpedo/drivers/scheduler"
	"github.com/pure-px/torpedo/pkg/log"
	. "github.com/pure-px/torpedo/tests"
)

// Delete More than one Backuplocation while periodic validation go routine is being executed.
var _ = Describe("{TestNfsBackupLocationCreationAndDeletion}", Label(TestCaseLabelsMap[PxBackupLabel]...), func() {
	var (
		backupLocationUID     string
		bkpLocationName       string
		providers             []string
		ctx                   context.Context
		backupLocationMap     map[string]string
		err                   error
		numOfBackupLocation   int
		wg                    sync.WaitGroup
		goroutinesToAdd       int
		backupLocationDeleted bool
	)

	JustBeforeEach(func() {
		StartPxBackupTorpedoTest("VerifyTestNfsBackupLocationCreationAndDeletion", "Delete More than one Backuplocation while periodic validation go routine is being executed.", nil, 300465, Pingle, Q3FY25)

		backupLocationMap = make(map[string]string)
		ctx, err = backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")
		providers = GetBackupProviders()
		numOfBackupLocation = 10
		goroutinesToAdd = 2

	})

	//Delete More than one Backuplocation while periodic validation go routine is being executed.
	It("Delete More than one Backuplocation while periodic validation go routine is being executed", func() {
		wg.Add(goroutinesToAdd)

		// 1. Create 10 backup location
		Step("Creating 10 backup locations", func() {

			for _, provider := range providers {
				log.InfoD("Creating backup location")
				for i := 1; i <= numOfBackupLocation; i++ {
					bkpLocationName = fmt.Sprintf("%s-%s-%v-bl", provider, getGlobalBucketName(provider), time.Now().Unix())
					backupLocationUID = uuid.New()
					backupLocationMap[backupLocationUID] = bkpLocationName
					err = CreateNFSBackupLocation(bkpLocationName, backupLocationUID, BackupOrgID, " ", getGlobalBucketName(provider), true)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of NFS backup location [%s]", bkpLocationName))
				}
			}
		})

		// 2. Validate backup locations
		Step("Validating backup locations", func() {
			log.InfoD("Validating backup locations")
			go func() {
				defer GinkgoRecover()
				defer wg.Done()
				for backupLocationUID, bkpLocationName := range backupLocationMap {
					err := ValidateBackupLocation(ctx, BackupOrgID, bkpLocationName, backupLocationUID)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying validation of NFS backup location [%s]", bkpLocationName))
				}
			}()
		})

		// 3. Delete backup locations
		Step("Deleting backup locations", func() {
			log.InfoD("Deleting backup location")
			go func() {
				defer GinkgoRecover()
				defer wg.Done()
				for backupLocationUID, bkpLocationName := range backupLocationMap {
					err := DeleteBackupLocation(bkpLocationName, backupLocationUID, BackupOrgID, true)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying backup location deletion - %s", bkpLocationName))
				}
			}()
			backupLocationDeleted = true
		})

		wg.Wait()
	})

	JustAfterEach(func() {
		defer EndPxBackupTorpedoTest([]*scheduler.Context{})
		// Clean up the cluster
		if backupLocationDeleted == false {
			CleanupCloudSettingsAndClusters(backupLocationMap, "", "", ctx)
		} else {
			CleanupCloudSettingsAndClusters(nil, "", "", ctx)
		}
	})
})
