package pxbackupworkflows

import (
	"sync"
	"time"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo"
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests/backup/longevity/pxbackupevents"
	report "github.com/portworx/torpedo/tests/backup/longevity/pxbackuplongevityreport"
	. "github.com/portworx/torpedo/tests/backup/longevity/pxbackuplongevitytypes"
)

// Global variables to be used by all flows
var (
	BackupLocationName   string
	BackupLocationUID    string
	AllNamespaces        []string
	ClusterUID           string
	ScheduledAppContexts []*scheduler.Context
)

func SetupAddCloudBackupLocation(wg *sync.WaitGroup) EventResponse {
	defer GinkgoRecover()
	defer wg.Done()
	result := GetLongevityEventResponse()
	result.Name = "Add global cloud location for backup"
	inputForBuilder := GetLongevityInputParams()

	eventData := RunBuilder(EventAddCredentialandBackupLocation, &inputForBuilder, &result)

	// Setting global variables for backup
	BackupLocationName = eventData.BackupLocationName
	BackupLocationUID = eventData.BackupLocationUID

	UpdateEventResponse(&result)

	return result
}

func SetupAddClusters(wg *sync.WaitGroup) EventResponse {
	defer GinkgoRecover()
	defer wg.Done()
	result := GetLongevityEventResponse()
	result.Name = "Add global cloud location for backup"
	inputForBuilder := GetLongevityInputParams()

	eventData := RunBuilder(EventAddSourceAndDestinationCluster, &inputForBuilder, &result)

	// Setting global variables for backup
	ClusterUID = eventData.ClusterUid

	UpdateEventResponse(&result)

	return result
}

func GlobalScheduleAndValidateApplication(wg *sync.WaitGroup) EventResponse {
	defer GinkgoRecover()
	defer wg.Done()
	result := GetLongevityEventResponse()
	result.Name = "Schedule And Validate App"
	inputForBuilder := GetLongevityInputParams()

	eventData := RunBuilder(EventScheduleApps, &inputForBuilder, &result)
	ScheduledAppContexts = append(ScheduledAppContexts, eventData.SchedulerContext...)
	AllNamespaces = append(AllNamespaces, eventData.BackupNamespaces...)

	_ = RunBuilder(EventValidateScheduleApplication, &inputForBuilder, &result)

	UpdateEventResponse(&result)

	return result

}

func AppCreateBackup(wg *sync.WaitGroup) EventResponse {
	defer GinkgoRecover()
	defer wg.Done()
	result := GetLongevityEventResponse()
	result.Name = "Create Backup"
	inputForBuilder := GetLongevityInputParams()

	log.Infof("Creating Backup")
	inputForBuilder.BackupData.BackupLocationName = BackupLocationName
	inputForBuilder.BackupData.BackupLocationUID = BackupLocationUID
	inputForBuilder.BackupData.ClusterUid = ClusterUID
	inputForBuilder.BackupData.Namespaces = GetRandomNamespacesForBackup()
	inputForBuilder.ApplicationData.SchedulerContext = ScheduledAppContexts

	_ = RunBuilder(EventCreateBackup, &inputForBuilder, &result)

	UpdateEventResponse(&result)

	return result
}

func OneSuccessOneFail(wg *sync.WaitGroup) EventResponse {
	defer GinkgoRecover()
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
	defer GinkgoRecover()
	defer wg.Done()
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
	defer GinkgoRecover()
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
	defer GinkgoRecover()
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
