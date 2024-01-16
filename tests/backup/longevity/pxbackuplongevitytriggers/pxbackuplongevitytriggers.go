package pxbackuplongevitytriggers

import (
	"sync"
	"time"

	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests/backup/longevity/pxbackuplongevitytypes"
	workflows "github.com/portworx/torpedo/tests/backup/longevity/pxbackupworkflows"
)

const (
	OneSuccessOneFail                    = "OneSuccessOneFail"
	OneSuccessTwoFail                    = "OneSuccessTwoFail"
	DisruptiveEvent                      = "DisruptiveEvent"
	CreateReport                         = "CreateReport"
	SetupAddCloudBackupLocation          = "SetupAddCloudBackupLocation"
	GlobalScheduleAndValidateApplication = "GlobalScheduleAndValidateApplication"
	SetupAddClusters                     = "SetupAddClusters"
	AppCreateBackup                      = "AppCreateBackup"
	AppCreateBackupandRestore            = "AppCreateBackupandRestore"
)

// All these events will run in longevity
var AppWorkflows = map[string]func(*sync.WaitGroup) EventResponse{
	// OneSuccessOneFail: workflows.OneSuccessOneFail,
	// OneSuccessTwoFail: workflows.OneSuccessTwoFail,
	AppCreateBackup:           workflows.AppCreateBackup,
	AppCreateBackupandRestore: workflows.AppCreateBackupandRestore,
}

var ReportingWorkflows = map[string]func(*sync.WaitGroup) EventResponse{
	CreateReport: workflows.CreateReport,
}

var DisruptiveWorkflows = map[string]func(*sync.WaitGroup) EventResponse{
	DisruptiveEvent: workflows.DisruptiveEvent,
}

var GlobalWorkflows = map[string]func(*sync.WaitGroup) EventResponse{
	GlobalScheduleAndValidateApplication: workflows.GlobalScheduleAndValidateApplication,
}

var SetupWorkflows = map[string]func(*sync.WaitGroup) EventResponse{
	SetupAddCloudBackupLocation: workflows.SetupAddCloudBackupLocation,
	SetupAddClusters:            workflows.SetupAddClusters,
}

var WorkflowFrequecy = map[string]int{
	OneSuccessOneFail:                    60,
	OneSuccessTwoFail:                    180,
	DisruptiveEvent:                      1,
	GlobalScheduleAndValidateApplication: 40,
	SetupAddCloudBackupLocation:          1,
	SetupAddClusters:                     1,
	AppCreateBackup:                      120,
	AppCreateBackupandRestore:            120,
}

var IsDisruptiveEvent = false

var runInterval = 60 // Run Interval in seconds

var baseInterval = 1200 // Base Interval in seconds

var eventsOccurred = make(map[string]int)

var isTriggered = false

func TriggerLongevityWorkflows(startTime time.Time, wg *sync.WaitGroup) time.Time {
	var lastRunTime = time.Now()

	for time.Since(startTime).Seconds() < float64(baseInterval) {

		// Triggering all setup events
		TriggerWorkflow(SetupWorkflows, wg)

		// Waiting for all setup workflows to complete
		wg.Wait()

		if time.Since(lastRunTime).Seconds() > float64(runInterval) || !isTriggered {

			// Triggering all setup events
			TriggerWorkflow(GlobalWorkflows, wg)

			// Waiting for all setup workflows to complete
			wg.Wait()

			if IsDisruptiveEvent {
				// Wait for all previous workflows to finish
				wg.Wait()

				// Starting all disruptive events
				TriggerWorkflow(DisruptiveWorkflows, wg)

				// Waiting for all disruptive events to complete
				wg.Wait()
			}

			// Starting all events
			TriggerWorkflow(AppWorkflows, wg)

			// Starting all reporting events
			TriggerWorkflow(ReportingWorkflows, wg)

			isTriggered = true
			lastRunTime = time.Now()
		} else {
			log.Info("Waiting for next interval to hit")
			time.Sleep(5 * time.Second)
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
			TriggerWorkflow(DisruptiveWorkflows, wg)

			// Waiting for all disruptive events to complete
			wg.Wait()
		}

		// Starting all events
		// Starting all disruptive events
		TriggerWorkflow(AppWorkflows, wg)

		isTriggered = true
		return time.Now()
	} else {
		log.Info("Waiting for next interval to hit")
		time.Sleep(30 * time.Second)
		return startTime
	}

}

func TriggerWorkflow(workflow map[string]func(*sync.WaitGroup) EventResponse, wg *sync.WaitGroup) {
	// Starting all disruptive events
	for workflowName, workflow := range workflow {

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
}
