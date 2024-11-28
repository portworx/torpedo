package tests

import (
	"fmt"
	"github.com/pure-px/torpedo/drivers/backup"
	"sync"

	"github.com/pure-px/torpedo/pkg/log"

	. "github.com/onsi/ginkgo/v2"
	"github.com/pure-px/torpedo/drivers/scheduler"
	. "github.com/pure-px/torpedo/tests"
)

var _ = Describe("{BackupLongevity}", func() {

	IsBackupLongevityRun = true

	contexts := make([]*scheduler.Context, 0)
	var triggerLock sync.Mutex
	var emailTriggerLock sync.Mutex
	var populateDone bool
	CommonPassword = backup.PxCentralAdminPwd + RandomString(4)
	triggerEventsChan := make(chan *EventRecord, 100)
	triggerBackupFunctions = map[string]func(*[]*scheduler.Context, *chan *EventRecord){
		CreatePxBackup:           TriggerCreateBackup,
		CreatePxBackupAndRestore: TriggerCreateBackupAndRestore,
		CreateRandomRestore:      TriggerCreateRandomRestore,
		DeployBackupApps:         TriggerDeployBackupApps,
		//CreatePxLockedBackup:                  TriggerCreateLockedBackup,
		CreateClusterShare:                                    TriggerShareCluster,
		CreateClusterUnshare:                                  TriggerUnShareCluster,
		CreateBackupWithUserFromSharedCluster:                 TriggerCreateBackupWithUserFromSharedCluster,
		CreateBackupRestoreAndDeleteWithUserFromSharedCluster: TriggerCreateBackupRestoreAndDeleteWithUserFromSharedCluster,
		DeletePxBackup:                                        TriggerDeleteSingleBackup,
	}
	//Creating a distinct trigger to make sure email triggers at regular intervals
	emailTriggerFunction = map[string]func(){
		EmailReporter: TriggerBackupEmailReporter,
	}

	BeforeEach(func() {
		if !populateDone {
			tags := map[string]string{
				"longevity": "true",
			}
			StartPxBackupTorpedoTest("BackupLongevityTest",
				"Longevity Run For Backup", tags, 0, ATrivedi, Q1FY24)
			populateBackupIntervals()
			// populateBackupDisruptiveTriggers()
			populateDone = true
		}
	})

	It("has to schedule app and introduce test triggers", func() {
		log.InfoD("schedule apps and start test triggers")
		watchLog := fmt.Sprintf("Start watch on K8S configMap [%s/%s]",
			configMapNS, testTriggersConfigMap)

		Step(watchLog, func() {
			log.InfoD(watchLog)
			watchConfigMap()
		})

		TriggerDeployBackupApps(&contexts, &triggerEventsChan)
		TriggerAddBackupCluster(&contexts, &triggerEventsChan)
		TriggerAddBackupCredAndBucket(&contexts, &triggerEventsChan)
		//TriggerAddLockedBackupCredAndBucket(&contexts, &triggerEventsChan)
		TriggerCreateUsers(&contexts, &triggerEventsChan)

		var wg sync.WaitGroup
		Step("Register test triggers", func() {
			for triggerType, triggerFunc := range triggerBackupFunctions {
				log.InfoD("Registering trigger: [%v]", triggerType)
				go backupEventTrigger(&wg, &contexts, triggerType, triggerFunc, &triggerLock, &triggerEventsChan)
				wg.Add(1)
			}
		})
		log.InfoD("Finished registering test triggers")
		if Inst().MinRunTimeMins != 0 {
			log.InfoD("Longevity Tests  timeout set to %d  minutes", Inst().MinRunTimeMins)
		}

		Step("Register email trigger", func() {
			for triggerType, triggerFunc := range emailTriggerFunction {
				log.InfoD("Registering email trigger: [%v]", triggerType)
				go emailEventTrigger(&wg, triggerType, triggerFunc, &emailTriggerLock)
				wg.Add(1)
			}
		})
		log.InfoD("Finished registering email trigger")

		CollectEventRecords(&triggerEventsChan)
		wg.Wait()
		close(triggerEventsChan)
		Step("teardown all apps", func() {
			for _, ctx := range contexts {
				TearDownContext(ctx, nil)
			}
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts)
	})
})

// Deletion of multiple schedule backups without suspending the schedules
var _ = Describe("{DeletionOfMultipleScheduleBackupsWithoutSuspendingScheduleLongevity}", func() {

	IsBackupLongevityRun = true

	contexts := make([]*scheduler.Context, 0)
	var triggerLock sync.Mutex
	var emailTriggerLock sync.Mutex
	var populateDone bool
	CommonPassword = backup.PxCentralAdminPwd + RandomString(4)
	triggerEventsChan := make(chan *EventRecord, 100)
	triggerBackupFunctions = map[string]func(*[]*scheduler.Context, *chan *EventRecord){
		RestartPxBackupPod: TriggerRestartPxBackupPod,
	}
	// Creating a distinct trigger to make sure email triggers at regular intervals
	emailTriggerFunction = map[string]func(){
		EmailReporter: TriggerBackupEmailReporter,
	}

	BeforeEach(func() {
		if !populateDone {
			tags := map[string]string{
				"longevity": "true",
			}
			StartPxBackupTorpedoTest("DeletionOfMultipleScheduleBackupsWithoutSuspendingScheduleLongevity",
				"Longevity Run For multiple schedule backups without suspending schedule", tags, 0, Sn, Q3FY25)
			populateBackupIntervals()
			populateDone = true
		}
	})

	It("Deletion of multiple schedule backups without suspending the schedules", func() {
		log.InfoD("Deletion of multiple schedule backups without suspending the schedules")
		watchLog := fmt.Sprintf("Start watch on K8S configMap [%s/%s]",
			configMapNS, testTriggersConfigMap)

		Step(watchLog, func() {
			log.InfoD(watchLog)
			watchConfigMap()
		})
		var wg sync.WaitGroup
		TriggerAddBackupCredAndBucket(&contexts, &triggerEventsChan)
		TriggerDeployMultipleApps(&contexts, &triggerEventsChan)
		TriggerAddBackupCluster(&contexts, &triggerEventsChan)
		TriggerAddSchedulePolicy(&contexts, &triggerEventsChan)
		wg.Add(1)
		go func() {
			defer wg.Done() // Ensure the goroutine signals Done after completion
			TriggerMultipleScheduleBackupDeletionAndCollectDeleteStats(&contexts, &triggerEventsChan)
		}()

		Step("Register test triggers", func() {
			for triggerType, triggerFunc := range triggerBackupFunctions {
				log.InfoD("Registering trigger: [%v]", triggerType)
				wg.Add(1) // Add to wait group before launching goroutine
				backupEventTrigger(&wg, &contexts, triggerType, triggerFunc, &triggerLock, &triggerEventsChan)
			}
		})
		log.InfoD("Finished registering test triggers")

		if Inst().MinRunTimeMins != 0 {
			log.InfoD("Longevity Tests timeout set to %d minutes", Inst().MinRunTimeMins)
		}

		Step("Register email trigger", func() {
			for triggerType, triggerFunc := range emailTriggerFunction {
				log.InfoD("Registering email trigger: [%v]", triggerType)
				wg.Add(1) // Add to wait group before launching goroutine
				emailEventTrigger(&wg, triggerType, triggerFunc, &emailTriggerLock)
			}
		})
		log.InfoD("Finished registering email trigger")

		// Step 5: Collect event records and wait for all goroutines to finish
		CollectEventRecords(&triggerEventsChan)

		wg.Wait()                // Wait for all goroutines to finish before proceeding
		close(triggerEventsChan) // Close the channel after all goroutines are done
		Step("teardown all apps", func() {
			for _, ctx := range contexts {
				TearDownContext(ctx, nil)
			}
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts)
	})
})
