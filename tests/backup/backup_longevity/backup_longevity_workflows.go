package tests

import (
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/pkg/log"
)

// Global variables to be used by all flows
var (
	LongevityBackupLocationName   string
	LongevityBackupLocationUID    string
	LongevityAllNamespaces        []string
	LongevityClusterUID           string
	LongevityScheduledAppContexts []*scheduler.Context
)

func SetupAddCloudBackupLocation() EventResponse {
	result := GetLongevityEventResponse()
	result.Name = "Add global cloud location for backup"
	inputForBuilder := GetLongevityInputParams()

	eventData := RunBuilder(EventAddCredentialandBackupLocation, &inputForBuilder, &result)

	// Setting global variables for backup
	LongevityBackupLocationName = eventData.BackupLocationName
	LongevityBackupLocationUID = eventData.BackupLocationUID

	UpdateEventResponse(&result)

	return result
}

func SetupAddClusters() EventResponse {
	result := GetLongevityEventResponse()
	result.Name = "Add global cloud location for backup"
	inputForBuilder := GetLongevityInputParams()

	eventData := RunBuilder(EventAddSourceAndDestinationCluster, &inputForBuilder, &result)

	// Setting global variables for backup
	LongevityClusterUID = eventData.ClusterUid

	UpdateEventResponse(&result)

	return result
}

func GlobalScheduleAndValidateApplication() EventResponse {
	result := GetLongevityEventResponse()
	result.Name = "Schedule And Validate App"
	inputForBuilder := GetLongevityInputParams()

	eventData := RunBuilder(EventScheduleApps, &inputForBuilder, &result)
	LongevityScheduledAppContexts = append(LongevityScheduledAppContexts, eventData.SchedulerContext...)
	LongevityAllNamespaces = append(LongevityAllNamespaces, eventData.BackupNamespaces...)

	_ = RunBuilder(EventValidateScheduleApplication, &inputForBuilder, &result)

	UpdateEventResponse(&result)

	return result

}

func AppCreateBackup() EventResponse {
	result := GetLongevityEventResponse()
	result.Name = "Create Backup"
	inputForBuilder := GetLongevityInputParams()

	log.Infof("Creating Backup")
	inputForBuilder.BackupData.BackupLocationName = LongevityBackupLocationName
	inputForBuilder.BackupData.BackupLocationUID = LongevityBackupLocationUID
	inputForBuilder.BackupData.ClusterUid = LongevityClusterUID
	inputForBuilder.BackupData.Namespaces = GetRandomNamespacesForBackup()
	inputForBuilder.ApplicationData.SchedulerContext = LongevityScheduledAppContexts

	_ = RunBuilder(EventCreateBackup, &inputForBuilder, &result)

	UpdateEventResponse(&result)

	return result
}

func AppCreateBackupandRestore() EventResponse {
	result := GetLongevityEventResponse()
	result.Name = "Create Backup"
	inputForBuilder := GetLongevityInputParams()

	log.Infof("Creating Backup")
	inputForBuilder.BackupData.BackupLocationName = LongevityBackupLocationName
	inputForBuilder.BackupData.BackupLocationUID = LongevityBackupLocationUID
	inputForBuilder.BackupData.ClusterUid = LongevityClusterUID
	inputForBuilder.BackupData.Namespaces = GetRandomNamespacesForBackup()
	inputForBuilder.ApplicationData.SchedulerContext = LongevityScheduledAppContexts

	eventData := RunBuilder(EventCreateBackup, &inputForBuilder, &result)

	inputForBuilder.BackupData.BackupName = eventData.BackupNames[0]

	_ = RunBuilder(EventRestore, &inputForBuilder, &result)

	UpdateEventResponse(&result)

	return result
}

func OneSuccessOneFail() EventResponse {
	result := GetLongevityEventResponse()
	result.Name = "OneSuccessOneFail"

	inputForBuilder := GetLongevityInputParams()
	inputForBuilder.CustomData.Integers["timeToBlock"] = 3

	RunBuilder(EventBuilder1, &inputForBuilder, &result)

	//RunBuilder(EventBuilder1Fail, &inputForBuilder, &result)

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

func DisruptiveEvent() EventResponse {
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

func CreateReport() EventResponse {
	// report.DumpResult()
	result := GetLongevityEventResponse()
	result.Name = "DumpingData"

	return result
}

// func LogEventData(eventResponse *EventResponse) {
// 	// fmt.Println(table)
// 	// fmt.Printf("\n\n\n")
// 	report.ResultsMutex.Lock()
// 	report.Results[eventResponse.Name+"-"+uuid.NewString()+"-"+time.Now().Format("02 Jan 06 15:04 MST")] = report.ResultForReport{
// 		Data:   *eventResponse,
// 		Status: eventResponse.Status,
// 	}
// 	report.ResultsMutex.Unlock()
// }
