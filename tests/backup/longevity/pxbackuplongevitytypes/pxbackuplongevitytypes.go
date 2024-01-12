package pxbackuplongevitytypes

import (
	"context"

	"github.com/portworx/torpedo/drivers/scheduler"
)

type PxBackupLongevity struct {
	CustomData      *CustomData
	ApplicationData *ApplicationData
	BackupData      *BackupData
}

type CustomData struct {
	Integers map[string]int
	Strings  map[string]string
}

type BackupData struct {
	Namespaces         []string
	BackupLocationName string
	BackupLocationUID  string
	ClusterUid         string
}

type ApplicationData struct {
	SchedulerContext []*scheduler.Context
}

type EventData struct {
	SchedulerContext   []*scheduler.Context
	AppContext         context.Context
	BackupNamespaces   []string
	BackupLocationName string
	BackupLocationUID  string
	ClusterUid         string
	BackupNames        []string
}

type EventBuilderResponse struct {
	Error              error
	TimeTakenInMinutes float32
	HighlightEvent     string
	EventData          EventData
}

type EventResponse struct {
	Name               string
	EventBuilders      map[string]*EventBuilderResponse
	Errors             []error
	TimeTakenInMinutes float32
	HighlightEvents    []string
	DisruptiveEventRan []string
	Status             bool
}

func GetLongevityInputParams() PxBackupLongevity {
	var customData = CustomData{
		Integers: make(map[string]int),
		Strings:  make(map[string]string),
	}

	var backupData = BackupData{
		Namespaces:         make([]string, 0),
		BackupLocationName: "",
		BackupLocationUID:  "",
		ClusterUid:         "",
	}

	var applicationData = ApplicationData{
		SchedulerContext: make([]*scheduler.Context, 0),
	}

	var longevityStruct = PxBackupLongevity{
		CustomData:      &customData,
		ApplicationData: &applicationData,
		BackupData:      &backupData,
	}

	return longevityStruct
}

func GetLongevityEventResponse() EventResponse {
	var someOtherVar = make(map[string]*EventBuilderResponse)

	var eventResponse = EventResponse{
		EventBuilders: someOtherVar,
	}

	return eventResponse
}
