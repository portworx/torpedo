package pxbackupworkflows

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests/backup/longevity/pxbackupevents"
	report "github.com/portworx/torpedo/tests/backup/longevity/pxbackuplongevityreport"
	. "github.com/portworx/torpedo/tests/backup/longevity/pxbackuplongevitytypes"
)

func OneSuccessOneFail(wg *sync.WaitGroup) EventResponse {
	defer wg.Done()
	result := GetLongevityEventResponse()
	result.Name = "OneSuccessOneFail"

	inputForBuilder := GetLongevityInputParams()
	inputForBuilder.CustomData.Integers["timeToBlock"] = 3

	RunBuilder(EventBuilder1, &inputForBuilder, &result)

	//RunBuilder(EventBuilder1Fail, &inputForBuilder, &result)

	UpdateEventResponse(&result)

	return result
}

func OneSuccessTwoFail(wg *sync.WaitGroup) EventResponse {
	defer wg.Done()
	print("I am started")
	result := GetLongevityEventResponse()
	result.Name = "OneSuccessTwoFail"

	inputForBuilder := GetLongevityInputParams()
	inputForBuilder.CustomData.Integers["timeToBlock"] = 2

	RunBuilder(EventBuilder1, &inputForBuilder, &result)

	RunBuilder(EventBuilder1Fail, &inputForBuilder, &result)

	RunBuilder(EventBuilder1Fail, &inputForBuilder, &result)

	UpdateEventResponse(&result)

	return result
}

func DisruptiveEvent(wg *sync.WaitGroup) EventResponse {
	defer wg.Done()
	result := GetLongevityEventResponse()
	result.Name = "DisruptiveEvent"

	inputForBuilder := GetLongevityInputParams()
	inputForBuilder.CustomData.Integers["timeToBlock"] = 2

	RunBuilder(EventBuilder1Fail, &inputForBuilder, &result)

	RunBuilder(EventBuilder1Fail, &inputForBuilder, &result)

	RunBuilder(EventBuilder1Fail, &inputForBuilder, &result)

	UpdateEventResponse(&result)

	return result
}

func CreateReport(wg *sync.WaitGroup) EventResponse {
	defer wg.Done()
	report.DumpResult()
	result := GetLongevityEventResponse()
	result.Name = "DumpingData"

	return result
}

func UpdateEventResponse(eventResponse *EventResponse) {
	for _, builderResponse := range eventResponse.EventBuilders {
		eventResponse.TimeTakenInMinutes += builderResponse.TimeTakenInMinutes
		if builderResponse.Error != nil {
			eventResponse.Errors = append(eventResponse.Errors, builderResponse.Error)
		}
		eventResponse.HighlightEvents = append(eventResponse.HighlightEvents, builderResponse.HighlightEvent)
	}
	log.Infof("Error in event - %+v", eventResponse.Errors)
	if eventResponse.Errors != nil {
		eventResponse.Status = false
	} else {
		eventResponse.Status = true
	}
	LogEventData(eventResponse)
}

func LogEventData(eventResponse *EventResponse) {
	// fmt.Println(table)
	// fmt.Printf("\n\n\n")
	report.ResultsMutex.Lock()
	report.Results[eventResponse.Name+"-"+uuid.NewString()+"-"+time.Now().Format("02 Jan 06 15:04 MST")] = report.ResultForReport{
		Data:   *eventResponse,
		Status: eventResponse.Status,
	}
	report.ResultsMutex.Unlock()
}
