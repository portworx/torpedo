package pxbackupworkflows

import (
	"fmt"

	"github.com/gosuri/uitable"
	. "github.com/portworx/torpedo/tests/backup/longevity/pxbackupevents"
	. "github.com/portworx/torpedo/tests/backup/longevity/pxbackuplongevitytypes"
)

func OneSuccessOneFail() EventResponse {
	result := GetLongevityEventResponse()
	result.Name = "OneSuccessOneFail"

	inputForBuilder := GetLongevityInputParams()
	inputForBuilder.CustomData.Integers["timeToBlock"] = 3

	RunBuilder(EventBuilder1, &inputForBuilder, &result)

	RunBuilder(EventBuilder1Fail, &inputForBuilder, &result)

	UpdateEventResponse(&result)

	return result
}

func OneSuccessTwoFail() EventResponse {
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

func UpdateEventResponse(eventResponse *EventResponse) {
	for _, builderResponse := range eventResponse.EventBuilders {
		eventResponse.TimeTakenInMinutes += builderResponse.TimeTakenInMinutes
		eventResponse.Errors = append(eventResponse.Errors, builderResponse.Error)
		eventResponse.HighlightEvents = append(eventResponse.HighlightEvents, builderResponse.HighlightEvent)
	}
	LogEventData(eventResponse)
}

func LogEventData(eventResponse *EventResponse) {
	var allErrors []string
	var allHighlightEvents []string

	for _, err := range eventResponse.Errors {
		if err != nil {
			allErrors = append(allErrors, err.Error())
		}
	}
	for _, hEvents := range eventResponse.HighlightEvents {
		allHighlightEvents = append(allHighlightEvents, hEvents)
	}
	fmt.Printf("\n\n")

	table := uitable.New()
	table.MaxColWidth = 50
	table.Wrap = false
	table.AddRow("NAME", "ERROR", "HIGHLIGHT", "TimeTakenInMinutes")
	for eventName, response := range eventResponse.EventBuilders {
		table.AddRow(eventName, response.Error, response.HighlightEvent, response.TimeTakenInMinutes)
	}

	fmt.Println(table)
	fmt.Printf("\n\n\n")

}
