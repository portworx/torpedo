package pxbackuplongevitytriggers

import (
	"sync"
	"time"

	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests/backup/longevity/pxbackuplongevitytypes"
	workflows "github.com/portworx/torpedo/tests/backup/longevity/pxbackupworkflows"
)

const (
	OneSuccessOneFail = "OneSuccessOneFail"
	OneSuccessTwoFail = "OneSuccessTwoFail"
	DisruptiveEvent   = "DisruptiveEvent"
	CreateReport      = "CreateReport"
)

// All these events will run in longevity
var AllWorkflows = map[string]func(*sync.WaitGroup) EventResponse{
	OneSuccessOneFail: workflows.OneSuccessOneFail,
	OneSuccessTwoFail: workflows.OneSuccessTwoFail,
	CreateReport:      workflows.CreateReport,
}

var AllDisruptiveWorkflows = map[string]func(*sync.WaitGroup) EventResponse{
	DisruptiveEvent: workflows.DisruptiveEvent,
}

var WorkflowFrequecy = map[string]int{
	OneSuccessOneFail: 60,
	OneSuccessTwoFail: 180,
	DisruptiveEvent:   1,
	CreateReport:      2,
}

var IsDisruptiveEvent = true

var runInterval = 10 // Run Interval in seconds

var baseInterval = 30 // Base Interval in seconds

var eventsOccurred = make(map[string]int)

var isTriggered = false

func TriggerLongevityWorkflows(startTime time.Time, wg *sync.WaitGroup) time.Time {
	var lastRunTime = time.Now()
	for time.Since(startTime).Seconds() < float64(baseInterval) {

		if time.Since(lastRunTime).Seconds() > float64(runInterval) || !isTriggered {

			if IsDisruptiveEvent {
				// Wait for all previous workflows to finish
				wg.Wait()

				// Starting all disruptive events
				log.Info("Starting all disruptive events")
				for workflowName, workflow := range AllDisruptiveWorkflows {

					numberOfEventsInThisInterval := WorkflowFrequecy[workflowName] * runInterval / baseInterval
					if numberOfEventsInThisInterval == 0 {
						if alreadyOccurred, ok := eventsOccurred[workflowName]; ok {
							if alreadyOccurred >= WorkflowFrequecy[workflowName] {
								continue
							} else {
								numberOfEventsInThisInterval = 1
							}
						} else {
							numberOfEventsInThisInterval = 1
						}
					}
					log.Infof("Starting %s - %d times", workflowName, numberOfEventsInThisInterval)
					for i := 0; i < numberOfEventsInThisInterval; i++ {
						go workflow(wg)
						wg.Add(1)
						if alreadyOccurred, ok := eventsOccurred[workflowName]; ok {
							eventsOccurred[workflowName] = alreadyOccurred + 1
						} else {
							eventsOccurred[workflowName] = 1
						}
					}
				}

				// Waiting for all disruptive events to complete
				wg.Wait()
			}

			// Starting all events
			log.Info("Starting all normal events")
			for workflowName, workflow := range AllWorkflows {
				numberOfEventsInThisInterval := WorkflowFrequecy[workflowName] * runInterval / baseInterval
				if numberOfEventsInThisInterval == 0 {
					if alreadyOccurred, ok := eventsOccurred[workflowName]; ok {
						if alreadyOccurred >= WorkflowFrequecy[workflowName] {
							continue
						} else {
							numberOfEventsInThisInterval = 1
						}
					} else {
						numberOfEventsInThisInterval = 1
					}
				}
				log.Infof("Starting %s - %d times", workflowName, numberOfEventsInThisInterval)
				for i := 0; i < numberOfEventsInThisInterval; i++ {
					go workflow(wg)
					wg.Add(1)
					if alreadyOccurred, ok := eventsOccurred[workflowName]; ok {
						eventsOccurred[workflowName] = alreadyOccurred + 1
					} else {
						eventsOccurred[workflowName] = 1
					}
				}
			}

			isTriggered = true
			lastRunTime = time.Now()
		} else {
			log.Info("Waiting for next interval to hit")
			time.Sleep(2 * time.Second)
		}
	}
	log.Infof("Total events occurred - [%v]", eventsOccurred)
	eventsOccurred = make(map[string]int)

	return time.Now()

}

func TriggerStressWorkflows(startTime time.Time, wg *sync.WaitGroup) time.Time {

	if time.Since(startTime).Minutes() > float64(runInterval) || !isTriggered {

		if IsDisruptiveEvent {
			// Wait for all previous workflows to finish
			wg.Wait()

			// Starting all disruptive events
			log.Info("Starting all disruptive events")
			for workflowName, workflow := range AllDisruptiveWorkflows {
				log.Infof("Starting %s - %d times", workflowName, WorkflowFrequecy[workflowName])
				for i := 0; i < WorkflowFrequecy[workflowName]; i++ {
					go workflow(wg)
					wg.Add(1)
				}
			}

			// Waiting for all disruptive events to complete
			wg.Wait()
		}

		// Starting all events
		log.Info("Starting all normal events")
		for workflowName, workflow := range AllWorkflows {
			log.Infof("Starting %s - %d times", workflowName, WorkflowFrequecy[workflowName])
			for i := 0; i < WorkflowFrequecy[workflowName]; i++ {
				go workflow(wg)
				wg.Add(1)
			}
		}

		isTriggered = true
		return time.Now()
	} else {
		log.Info("Waiting for next interval to hit")
		return startTime
	}

}
