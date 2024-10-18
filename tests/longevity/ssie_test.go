package tests

import (
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	"github.com/pure-px/torpedo/drivers/scheduler"
	"github.com/pure-px/torpedo/pkg/log"
	. "github.com/pure-px/torpedo/tests"
	"math/rand"
	"os"
	"slices"
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
		var ssieErr error
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer GinkgoRecover()
			TriggerDeployNewApps(&contexts, &triggerEventsChan)
			ssieErr = ValidateSSIEStatus(&contexts)
		}()
		wg.Wait()
		if ssieErr != nil {
			dash.VerifyFatal(ssieErr, nil, "validate SSIE status after deploying apps")
		}

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

	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts)
	})
})

func triggerSSIECombo(contexts *[]*scheduler.Context, triggerEventsChan *chan *EventRecord) {
	waitTime := 2 * time.Minute

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer GinkgoRecover()
		for {

			select {
			case <-StopSSIEChan:
				log.InfoD("Received stop signal. Exiting non-destructive test triggers")
				return
			default:
				// Continuing the loop as no stop signal is received
			}

			if baseInterval, ok := ChaosMap[BaseInterval]; ok {
				waitTime = time.Duration(baseInterval) * time.Minute
			}

			nextCombinationToRun := getNextCombination()

			if nextCombinationToRun == nil {
				log.Infof("No more SSIE combinations to trigger")
				return
			}

			log.InfoD("Waiting for %v before triggering SSIE combination %v", waitTime, nextCombinationToRun)
			time.Sleep(waitTime)

			var ndwg sync.WaitGroup

			for _, triggerType := range nextCombinationToRun {
				triggerFunc := triggerFunctions[triggerType]
				ndwg.Add(1)
				go func() {
					defer ndwg.Done()
					defer GinkgoRecover()
					triggerFunc(contexts, triggerEventsChan)
				}()
			}
			ndwg.Wait()
		}
	}()

	var ssieErr error
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer GinkgoRecover()

		for {
			select {
			case <-StopSSIEChan:
				log.InfoD("Received stop signal. Exiting destructive test triggers")
				return
			default:
				// Continuing the loop as no stop signal is received
			}
			destructiveEvent := getDestructiveEvent()
			log.InfoD("Running destructive event [%s]", destructiveEvent)

			triggerFunc := triggerFunctions[destructiveEvent]
			triggerFunc(contexts, triggerEventsChan)
			log.InfoD(fmt.Sprintf("Validating SSIE status after running [%s]", destructiveEvent))
			ssieErr = ValidateSSIEStatus(contexts)
			if ssieErr != nil {
				dash.VerifySafely(ssieErr, nil, fmt.Sprintf("verify SSIE status after running [%s]", destructiveEvent))
				close(StopSSIEChan)
			}
		}
	}()

	wg.Wait()

}

func getNextCombination() []string {
	GenerateAndStoreEventCombinations()
	if len(eventCombinations) > 0 {
		rand.Seed(time.Now().UnixNano())

		// Generate a random index within the range of the list
		randomIndex := rand.Intn(len(eventCombinations))
		return eventCombinations[randomIndex]
	}
	return nil
}

func getDestructiveEvent() string {
	var enabledDestructiveEvents []string
	for event := range triggerFunctions {
		_, enableEvent := isTriggerEnabled(event)
		if enableEvent {
			if slices.Contains(disruptiveTriggers, event) {
				enabledDestructiveEvents = append(enabledDestructiveEvents, event)
			}
		}
	}

	rand.Seed(time.Now().UnixNano())
	var selectedValue string

	// Keep picking a random value until it's not in the recentPicks slice
	for {
		selectedValue = enabledDestructiveEvents[rand.Intn(len(enabledDestructiveEvents))]
		if !slices.Contains(recentDestructivePicks, selectedValue) {
			break
		}
	}

	// Add the selected value to recentPicks
	recentDestructivePicks = append(recentDestructivePicks, selectedValue)
	if len(recentDestructivePicks) > len(enabledDestructiveEvents)/2 {
		recentDestructivePicks = recentDestructivePicks[1:] // Remove the oldest entry to maintain only the last half picks
	}
	return selectedValue
}
