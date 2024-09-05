package tests

import (
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests"
	"os"
	"sync"
	"time"
)

var _ = Describe("{RunSSIE}", func() {
	contexts := make([]*scheduler.Context, 0)
	var emailTriggerLock sync.Mutex
	var populateDone bool
	triggerEventsChan := make(chan *EventRecord, 100)

	BeforeEach(func() {
		if !populateDone {
			tags := map[string]string{
				"ssie": "true",
			}
			StartTorpedoTest("RunSSIE", "Validate SSIE workflows", tags, 0)

			populateTriggerFuncs()
			populateIntervals()
			populateDisruptiveTriggers()
			populateDone = true
		}
	})

	It("has to schedule app and run SSIE", func() {
		log.InfoD("schedule apps and start SSIE triggers")
		watchLog := fmt.Sprintf("Start watch on K8S configMap [%s/%s]",
			configMapNS, testTriggersConfigMap)
		err := os.Setenv(IsSSIEKey, "true")
		log.FailOnError(err, "Failed to set env variable IS_SSIE")

		Step(watchLog, func() {
			log.InfoD(watchLog)
			err := watchConfigMap()
			if err != nil {
				log.Fatalf(fmt.Sprintf("%v", err))
			}
		})

		if pureTopologyEnabled {
			var err error
			labels, err = SetTopologyLabelsOnNodes()
			if err != nil {
				log.Fatalf(fmt.Sprintf("%v", err))
			}
			Inst().TopologyLabels = labels
		}

		Inst().IsHyperConverged = hyperConvergedTypeEnabled
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer GinkgoRecover()
			TriggerDeployNewApps(&contexts, &triggerEventsChan)
		}()
		wg.Wait()

		stepLog := "Register SSIE test triggers"
		Step(stepLog, func() {
			wg.Add(1)
			go func() {
				defer wg.Done()
				triggerSSIECombo(&contexts, &triggerEventsChan)
				time.Sleep(2 * time.Minute)
				close(StopLongevityChan)
				close(triggerEventsChan)
			}()

		})
		log.InfoD("Finished registering test triggers")
		if Inst().MinRunTimeMins != 0 {
			log.InfoD("SSIE Tests  timeout set to %d  minutes", Inst().MinRunTimeMins)
		}

		Step("Register email trigger", func() {
			for triggerType, triggerFunc := range emailTriggerFunction {
				wg.Add(1)
				log.InfoD("Registering email trigger: [%v]", triggerType)
				go emailEventTrigger(&wg, triggerType, triggerFunc, &emailTriggerLock)
			}
		})
		log.InfoD("Finished registering email trigger")

		CollectEventRecords(&triggerEventsChan)
		wg.Wait()
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

func triggerSSIECombo(contexts *[]*scheduler.Context, triggerEventsChan *chan *EventRecord) {
	waitTime := 2 * time.Minute
	for {

		if baseInterval, ok := ChaosMap[BaseInterval]; ok {
			waitTime = time.Duration(baseInterval) * time.Minute
		}

		nextCombinationToRun := getNextCombination()
		if nextCombinationToRun == nil {
			log.Infof("No more SSIE combinations to trigger")
			break
		}

		log.Infof("Waiting for %v before triggering SSIE combination %v", waitTime, nextCombinationToRun)
		time.Sleep(waitTime)
		var wg sync.WaitGroup
		for _, triggerType := range nextCombinationToRun {

			triggerFunc := triggerFunctions[triggerType]

			wg.Add(1)
			go func() {
				defer wg.Done()
				defer GinkgoRecover()
				triggerFunc(contexts, triggerEventsChan)
			}()
		}
		wg.Wait()
		err := UpdateAlreadyRanCombinations(nextCombinationToRun)
		if err != nil {
			log.Errorf("Failed to update already ran combinations: %v", err)
		}

		log.InfoD(fmt.Sprintf("Validating SSIE status after running [%s]", nextCombinationToRun))
		ValidateSSIEStatus(contexts)
	}

}

func getNextCombination() []string {
	GenerateAndStoreEventCombinations()
	if len(eventCombinations) > 0 {
		return eventCombinations[0]
	}
	return nil
}