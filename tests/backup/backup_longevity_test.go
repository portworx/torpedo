package tests

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/portworx/torpedo/pkg/log"

	. "github.com/onsi/ginkgo"
	"github.com/portworx/sched-ops/k8s/core"
	"github.com/portworx/torpedo/drivers/scheduler"
	. "github.com/portworx/torpedo/tests"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	testTriggersConfigMap = "pxb-longevity-triggers"
	configMapNS           = "default"
	controlLoopSleepTime  = time.Second * 15
)

var (
	// Stores mapping between chaos level and its freq. Values are hardcoded
	triggerInterval map[string]map[int]time.Duration
	// Stores which are disruptive triggers. When disruptive triggers are happening in test,
	// other triggers are allowed to happen only after existing triggers are complete.
	disruptiveTriggers map[string]bool

	triggerFunctions     map[string]func(*[]*scheduler.Context, *chan *EventRecord)
	emailTriggerFunction map[string]func()

	// Pure Topology is disabled by default
	pureTopologyEnabled = false

	//Default is allow deploying apps both in storage and storageless nodes
	hyperConvergedTypeEnabled = true

	// Pure Topology Label array
	labels []map[string]string
)

// TriggerFunction represents function signature of a testTrigger
type TriggerFunction func(*[]*scheduler.Context, *chan *EventRecord)

var _ = Describe("{BackupLongevity}", func() {
	contexts := make([]*scheduler.Context, 0)
	var triggerLock sync.Mutex
	var emailTriggerLock sync.Mutex
	var populateDone bool
	triggerEventsChan := make(chan *EventRecord, 100)
	triggerFunctions = map[string]func(*[]*scheduler.Context, *chan *EventRecord){
		CreatePxBackup:           TriggerCreateBackup,
		CreatePxBackupAndRestore: TriggerCreateBackupAndRestore,
		CreateRandomRestore:      TriggerCreateRandomRestore,
		DeployBackupApps:         TriggerDeployBackupApps,
	}
	//Creating a distinct trigger to make sure email triggers at regular intervals
	emailTriggerFunction = map[string]func(){
		EmailReporter: TriggerEmailReporter,
	}

	BeforeEach(func() {
		if !populateDone {
			StartPxBackupTorpedoTest("BackupLongevityTest",
				"Longevity Run For Backup", nil, 0, ATrivedi, Q1FY24)
			populateIntervals()
			//  populateDisruptiveTriggers()
			populateDone = true
		}
	})

	It("has to schedule app and introduce test triggers", func() {
		log.InfoD("schedule apps and start test triggers")
		watchLog := fmt.Sprintf("Start watch on K8S configMap [%s/%s]",
			configMapNS, testTriggersConfigMap)

		Step(watchLog, func() {
			log.InfoD(watchLog)
			err := watchConfigMap()
			if err != nil {
				log.Fatalf(fmt.Sprintf("%v", err))
			}
		})

		TriggerDeployBackupApps(&contexts, &triggerEventsChan)
		TriggerAddBackupCluster(&contexts, &triggerEventsChan)
		TriggerAddBackupCredAndBucket(&contexts, &triggerEventsChan)

		var wg sync.WaitGroup
		Step("Register test triggers", func() {
			for triggerType, triggerFunc := range triggerFunctions {
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

func backupEventTrigger(wg *sync.WaitGroup,
	contexts *[]*scheduler.Context,
	triggerType string,
	triggerFunc func(*[]*scheduler.Context, *chan *EventRecord),
	triggerLoc *sync.Mutex,
	triggerEventsChan *chan *EventRecord) {
	defer wg.Done()

	minRunTime := Inst().MinRunTimeMins
	timeout := (minRunTime) * 60

	start := time.Now().Local()
	lastInvocationTime := start

	for {
		// if timeout is 0, run indefinitely
		if timeout != 0 && int(time.Since(start).Seconds()) > timeout {
			log.InfoD("Longevity Tests timed out with timeout %d  minutes", minRunTime)
			break
		}

		waitTime, isTriggerEnabled := isTriggerEnabled(triggerType)

		if isTriggerEnabled && time.Since(lastInvocationTime) > time.Duration(waitTime) {

			triggerFunc(contexts, triggerEventsChan)
			log.Infof("Trigger Function completed for [%s]\n", triggerType)

			lastInvocationTime = time.Now().Local()

		}
		time.Sleep(controlLoopSleepTime)
	}
	os.Exit(0)
}

func emailEventTrigger(wg *sync.WaitGroup,
	triggerType string,
	triggerFunc func(),
	emailTriggerLock *sync.Mutex) {
	defer wg.Done()

	start := time.Now().Local()
	lastInvocationTime := start

	for {

		// Get next interval of when trigger should happen
		// This interval can dynamically change by editing configMap
		waitTime, isTriggerEnabled := isTriggerEnabled(triggerType)

		if isTriggerEnabled && time.Since(lastInvocationTime) > time.Duration(waitTime) {
			// If trigger is not disabled and its right time to trigger,

			log.InfoD("Waiting for lock for trigger [%s]\n", triggerType)
			emailTriggerLock.Lock()
			log.InfoD("Successfully taken lock for trigger [%s]\n", triggerType)

			triggerFunc()
			log.InfoD("Trigger Function completed for [%s]\n", triggerType)

			emailTriggerLock.Unlock()
			log.InfoD("Successfully released lock for trigger [%s]\n", triggerType)

			lastInvocationTime = time.Now().Local()

		}
		time.Sleep(controlLoopSleepTime)
	}
}

func watchConfigMap() error {
	ChaosMap = map[string]int{}
	cm, err := core.Instance().GetConfigMap(testTriggersConfigMap, configMapNS)
	if err != nil {
		return fmt.Errorf("Error reading config map: %v", err)
	}
	err = populateDataFromConfigMap(&cm.Data)
	if err != nil {
		return err
	}

	// Apply watch if configMap exists
	fn := func(object runtime.Object) error {
		cm, ok := object.(*v1.ConfigMap)
		if !ok {
			err := fmt.Errorf("invalid object type on configmap watch: %v", object)
			return err
		}
		if len(cm.Data) > 0 {
			err = populateDataFromConfigMap(&cm.Data)
			if err != nil {
				return err
			}
		}
		return nil
	}

	err = core.Instance().WatchConfigMap(cm, fn)
	if err != nil {
		return fmt.Errorf("Failed to watch on config map: %s due to: %v", testTriggersConfigMap, err)
	}
	return nil
}

func populateDisruptiveTriggers() {
	disruptiveTriggers = map[string]bool{}
}

func isDisruptiveTrigger(triggerType string) bool {
	return disruptiveTriggers[triggerType]
}

func populateDataFromConfigMap(configData *map[string]string) error {
	log.Infof("ChaosMap provided: %v", configData)
	setEmailRecipients(configData)
	setEmailHost(configData)
	setEmailSubject(configData)

	err := populateTriggers(configData)
	if err != nil {
		return err
	}
	return nil
}

func setEmailRecipients(configData *map[string]string) {
	// Get email recipients from configMap
	if emailRecipients, ok := (*configData)[EmailRecipientsConfigMapField]; !ok {
		log.Warnf("No [%s] field found in [%s] config-map in [%s] namespace."+
			"Defaulting email recipients to [%s].\n",
			EmailRecipientsConfigMapField, testTriggersConfigMap, configMapNS, DefaultEmailRecipient)
		EmailRecipients = []string{DefaultEmailRecipient}
	} else {
		EmailRecipients = strings.Split(emailRecipients, ";")
		delete(*configData, EmailRecipientsConfigMapField)
	}
}

func setEmailHost(configData *map[string]string) error {
	if emailhost, ok := (*configData)[EmailHostServerField]; ok {
		EmailServer = emailhost
		delete(*configData, EmailHostServerField)
		return nil
	}
	return fmt.Errorf("Failed to find [%s] field in config-map [%s] in namespace [%s]",
		EmailHostServerField, testTriggersConfigMap, configMapNS)
}

func setEmailSubject(configData *map[string]string) {
	if emailsubject, ok := (*configData)[EmailSubjectField]; ok {
		EmailSubject = emailsubject
		delete(*configData, EmailSubjectField)
	} else {
		EmailSubject = "Torpedo Longevity Report"
	}
}

func setSendGridEmailAPIKey(configData *map[string]string) error {
	if apiKey, ok := (*configData)[SendGridEmailAPIKeyField]; ok {
		SendGridEmailAPIKey = apiKey
		delete(*configData, SendGridEmailAPIKeyField)
		return nil
	}
	return fmt.Errorf("Failed to find [%s] field in config-map [%s] in namespace [%s]",
		SendGridEmailAPIKeyField, testTriggersConfigMap, configMapNS)
}

func populateTriggers(triggers *map[string]string) error {
	for triggerType, chaosLevel := range *triggers {
		chaosLevelInt, err := strconv.Atoi(chaosLevel)
		if err != nil {
			return fmt.Errorf("Failed to get chaos levels from configMap [%s] in [%s] namespace. Error:[%v]",
				testTriggersConfigMap, configMapNS, err)
		}
		ChaosMap[triggerType] = chaosLevelInt
		if triggerType == BackupScheduleAll || triggerType == BackupScheduleScale {
			SetScheduledBackupInterval(triggerInterval[triggerType][chaosLevelInt], triggerType)
		}
	}

	RunningTriggers = map[string]time.Duration{}
	for triggerType := range triggerFunctions {
		chaosLevel, ok := ChaosMap[triggerType]
		if !ok {
			chaosLevel = Inst().ChaosLevel
		}
		if chaosLevel != 0 {
			RunningTriggers[triggerType] = triggerInterval[triggerType][chaosLevel]
		}

	}
	return nil
}

func populateIntervals() {
	triggerInterval = map[string]map[int]time.Duration{}
	triggerInterval[CreatePxBackup] = map[int]time.Duration{}
	triggerInterval[EmailReporter] = map[int]time.Duration{}
	triggerInterval[CreatePxBackupAndRestore] = map[int]time.Duration{}
	triggerInterval[CreateRandomRestore] = map[int]time.Duration{}
	triggerInterval[DeployBackupApps] = map[int]time.Duration{}

	baseInterval := 1 * time.Second

	triggerInterval[CreatePxBackup][10] = 6 * baseInterval
	triggerInterval[CreatePxBackup][9] = 12 * baseInterval
	triggerInterval[CreatePxBackup][8] = 18 * baseInterval
	triggerInterval[CreatePxBackup][7] = 24 * baseInterval
	triggerInterval[CreatePxBackup][6] = 30 * baseInterval
	triggerInterval[CreatePxBackup][5] = 36 * baseInterval
	triggerInterval[CreatePxBackup][4] = 42 * baseInterval
	triggerInterval[CreatePxBackup][3] = 48 * baseInterval
	triggerInterval[CreatePxBackup][2] = 54 * baseInterval
	triggerInterval[CreatePxBackup][1] = 60 * baseInterval

	triggerInterval[CreatePxBackupAndRestore][10] = 6 * baseInterval
	triggerInterval[CreatePxBackupAndRestore][9] = 12 * baseInterval
	triggerInterval[CreatePxBackupAndRestore][8] = 18 * baseInterval
	triggerInterval[CreatePxBackupAndRestore][7] = 24 * baseInterval
	triggerInterval[CreatePxBackupAndRestore][6] = 30 * baseInterval
	triggerInterval[CreatePxBackupAndRestore][5] = 36 * baseInterval
	triggerInterval[CreatePxBackupAndRestore][4] = 42 * baseInterval
	triggerInterval[CreatePxBackupAndRestore][3] = 48 * baseInterval
	triggerInterval[CreatePxBackupAndRestore][2] = 54 * baseInterval
	triggerInterval[CreatePxBackupAndRestore][1] = 60 * baseInterval

	triggerInterval[CreateRandomRestore][10] = 1 * baseInterval
	triggerInterval[CreateRandomRestore][9] = 3 * baseInterval
	triggerInterval[CreateRandomRestore][8] = 6 * baseInterval
	triggerInterval[CreateRandomRestore][7] = 9 * baseInterval
	triggerInterval[CreateRandomRestore][6] = 12 * baseInterval
	triggerInterval[CreateRandomRestore][5] = 15 * baseInterval
	triggerInterval[CreateRandomRestore][4] = 18 * baseInterval
	triggerInterval[CreateRandomRestore][3] = 21 * baseInterval
	triggerInterval[CreateRandomRestore][2] = 24 * baseInterval
	triggerInterval[CreateRandomRestore][1] = 27 * baseInterval

	triggerInterval[DeployBackupApps][10] = 5 * baseInterval
	triggerInterval[DeployBackupApps][9] = 10 * baseInterval
	triggerInterval[DeployBackupApps][8] = 15 * baseInterval
	triggerInterval[DeployBackupApps][7] = 30 * baseInterval
	triggerInterval[DeployBackupApps][6] = 3 * 15 * baseInterval
	triggerInterval[DeployBackupApps][5] = 3 * 30 * baseInterval
	triggerInterval[DeployBackupApps][4] = 3 * 60 * baseInterval
	triggerInterval[DeployBackupApps][3] = 6 * 60 * baseInterval
	triggerInterval[DeployBackupApps][2] = 12 * 60 * baseInterval
	triggerInterval[DeployBackupApps][1] = 24 * 60 * baseInterval

	baseInterval = 1 * time.Minute

	triggerInterval[EmailReporter][10] = 1 * baseInterval
	triggerInterval[EmailReporter][9] = 2 * baseInterval
	triggerInterval[EmailReporter][8] = 3 * baseInterval
	triggerInterval[EmailReporter][7] = 4 * baseInterval
	triggerInterval[EmailReporter][6] = 5 * baseInterval
	triggerInterval[EmailReporter][5] = 6 * baseInterval
	triggerInterval[EmailReporter][4] = 7 * baseInterval
	triggerInterval[EmailReporter][3] = 8 * baseInterval
	triggerInterval[EmailReporter][2] = 12 * baseInterval
	triggerInterval[EmailReporter][1] = 24 * baseInterval

}

func isTriggerEnabled(triggerType string) (time.Duration, bool) {
	var chaosLevel int
	var ok bool
	chaosLevel, ok = ChaosMap[triggerType]
	if !ok {
		chaosLevel = Inst().ChaosLevel
		log.Warnf("Chaos level for trigger [%s] not found in chaos map. Using global chaos level [%d]",
			triggerType, Inst().ChaosLevel)
	}
	if triggerInterval[triggerType][chaosLevel] != 0 {
		return triggerInterval[triggerType][chaosLevel], true
	}
	return triggerInterval[triggerType][chaosLevel], false
}
