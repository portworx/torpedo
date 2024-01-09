package pxbackupevents

import (
	"fmt"
	"time"

	. "github.com/portworx/torpedo/tests/backup/longevity/pxbackuplongevitytypes"
)

const (
	EventBuilder1     = "EventBuilder1"
	EventBuilder1Fail = "EventBuilder1Fail"
)

type PxBackupEventBuilder func(*PxBackupLongevity) (error, string)

var AllBuilders = map[string]PxBackupEventBuilder{
	EventBuilder1:     eventBuilder1,
	EventBuilder1Fail: eventBuilder1Fail,
}

func eventBuilder1(inputsForEventBuilder *PxBackupLongevity) (error, string) {
	time.Sleep(time.Second * time.Duration(inputsForEventBuilder.CustomData.Integers["timeToBlock"]))
	return nil, ""
}

func eventBuilder1Fail(inputsForEventBuilder *PxBackupLongevity) (error, string) {
	time.Sleep(time.Second * time.Duration(inputsForEventBuilder.CustomData.Integers["timeToBlock"]))
	return fmt.Errorf("This is the returned error"), "This is the highlight event from - EventBuilder1Fail"
}

func RunBuilder(eventBuilderName string, inputsForEventBuilder *PxBackupLongevity, eventResponse *EventResponse) {
	eventBuilder := AllBuilders[eventBuilderName]
	eventBuilderIdentifier := eventBuilderName + "-" + time.Now().Format("15:04:05.000")
	eventResponse.EventBuilders[eventBuilderIdentifier] = &EventBuilderResponse{}

	startTime := time.Now()

	err, highlightEvent := eventBuilder(inputsForEventBuilder)
	if err != nil {
		eventResponse.EventBuilders[eventBuilderIdentifier].Error = err
	}
	if highlightEvent != "" {
		eventResponse.EventBuilders[eventBuilderIdentifier].HighlightEvent = highlightEvent
	}
	eventResponse.EventBuilders[eventBuilderIdentifier].TimeTakenInMinutes = float32(time.Since(startTime).Minutes())

}
