package tests

import (
	"math/rand"
	"time"

	"github.com/portworx/torpedo/pkg/log"
)

// ...

func GetRandomNamespacesForBackup() []string {
	var allNamespacesForBackupMap = make(map[string]bool)
	var allNamepsacesForBackup []string
	rand.Seed(time.Now().Unix()) // initialize global pseudo random generator

	numberOfNamespaces := rand.Intn(len(LongevityAllNamespaces))

	for i := 0; i <= numberOfNamespaces; i++ {
		allNamespacesForBackupMap[LongevityAllNamespaces[rand.Intn(len(LongevityAllNamespaces))]] = true
	}

	for namespaceName, _ := range allNamespacesForBackupMap {
		allNamepsacesForBackup = append(allNamepsacesForBackup, namespaceName)
	}

	log.Infof("Returning This - [%v]", allNamepsacesForBackup)
	return allNamepsacesForBackup
}

func UpdateEventResponse(eventResponse *EventResponse) {
	for _, builderResponse := range eventResponse.EventBuilders {
		eventResponse.TimeTakenInMinutes += builderResponse.TimeTakenInMinutes
		if builderResponse.Error != nil {
			eventResponse.Errors = append(eventResponse.Errors, builderResponse.Error)
		}
		eventResponse.HighlightEvents = append(eventResponse.HighlightEvents, builderResponse.HighlightEvent)
	}
	if eventResponse.Errors != nil {
		eventResponse.Status = false
	} else {
		eventResponse.Status = true
	}
	// LogEventData(eventResponse)
}
