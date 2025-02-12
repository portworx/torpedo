package tests

import (
	"context"
	"fmt"
	"sync"

	. "github.com/onsi/ginkgo/v2"
	"github.com/pborman/uuid"
	"github.com/pure-px/torpedo/drivers"
	"github.com/pure-px/torpedo/drivers/backup"
	"github.com/pure-px/torpedo/drivers/scheduler"
	"github.com/pure-px/torpedo/pkg/log"
	. "github.com/pure-px/torpedo/tests"
)

// Add 5 backuplocation in parallel while verification is going on for the previous BLs
var _ = Describe("{Add5BackuplocationInParallelWhileVerificationIsGoingOnForThePreviousBLs}", Label(TestCaseLabelsMap[AddBackupLocationWhileNfsValidatorPodIsRunning]...), func() {
	var (
		adminCtx            context.Context
		backupLocationMap   map[string]string
		backupLocationMap2  map[string]string
		err                 error
		numOfBackupLocation int
	)

	JustBeforeEach(func() {
		StartPxBackupTorpedoTest("VerifyAdd5BackuplocationInParallelWhileVerificationIsGoingOnForThePreviousBLs", "Add 5 backuplocation in parallel while verification is going on for the previous BLs", nil, 300464, ABadgujar, Q1FY25)
		numOfBackupLocation = 5
		backupLocationMap = make(map[string]string)
		backupLocationMap2 = make(map[string]string)
		adminCtx, err = backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")

	})

	//Add 5 backuplocation in parallel while verification is going on for the previous BLs
	It("Add 5 backuplocation in parallel while verification is going on for the previous BLs", func() {

		// Step 1: Create 5 NFS Backup Location
		Step("Create 5 NFS Backup Location", func() {
			log.InfoD("Create 5 NFS Backup Location")

			for i := 1; i <= numOfBackupLocation; i++ {
				nfsBackupLocationName := fmt.Sprintf("%s-%v", "nfs", i)
				nfsBackupLocationUID := uuid.New()
				backupLocationMap[nfsBackupLocationUID] = nfsBackupLocationName

				log.InfoD("Creating NFS backup location: %s", nfsBackupLocationName)

				err := CreateNFSBackupLocation(nfsBackupLocationName, nfsBackupLocationUID, BackupOrgID, " ", getGlobalBucketName(drivers.ProviderNfs), true)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of NFS backup location [%s]", nfsBackupLocationName))
			}

		})

		// Step 2: Validate the existing 5 Backup Locations, Create another 5 Backup Locations
		Step("Validate the existing 5 Backup Locations, Create another 5 Backup Locations", func() {
			log.InfoD("Validate the existing 5 Backup Locations, Create another 5 Backup Locations")
			var wg sync.WaitGroup
			// Step 3: Start validating the existing backup locations
			wg.Add(2)

			go func() {
				defer GinkgoRecover()
				defer wg.Done()
				for backupLocationUID, backupLocationName := range backupLocationMap {
					log.InfoD("Validating NFS backup location: %s", backupLocationName)
					err := ValidateBackupLocation(adminCtx, BackupOrgID, backupLocationName, backupLocationUID)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying validation of NFS backup location [%s]", backupLocationName))
				}
			}()

			// Step 4: Start creating 5 new backup locations in parallel while the validation is happening

			go func() {
				defer GinkgoRecover()
				defer wg.Done()
				for i := 1; i <= numOfBackupLocation; i++ {
					nfsBackupLocationName2 := fmt.Sprintf("%s-%v", "nfs", i+5)
					nfsBackupLocationUID2 := uuid.New()
					backupLocationMap2[nfsBackupLocationUID2] = nfsBackupLocationName2

					log.InfoD("Making Another NFS backup location: %s", nfsBackupLocationName2)
					err := CreateNFSBackupLocation(nfsBackupLocationName2, nfsBackupLocationUID2, BackupOrgID, " ", getGlobalBucketName(drivers.ProviderNfs), true)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of another NFS backup location [%s]", nfsBackupLocationName2))
				}
			}()
			wg.Wait()
		})

	})

	JustAfterEach(func() {
		defer EndPxBackupTorpedoTest([]*scheduler.Context{})
		log.InfoD("Cleanup Entered")
		//Merge the 2 Maps
		for bkLocationUid, bkLocationName := range backupLocationMap2 {
			backupLocationMap[bkLocationUid] = bkLocationName
		}
		// Clean up the cluster
		CleanupCloudSettingsAndClusters(backupLocationMap, "", "", adminCtx)
	})
})
