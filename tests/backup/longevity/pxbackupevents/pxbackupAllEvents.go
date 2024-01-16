package pxbackupevents

var AllBuilders = map[string]PxBackupEventBuilder{
	EventBuilder1:                       eventBuilder1,
	EventBuilder1Fail:                   eventBuilder1Fail,
	EventScheduleApps:                   eventScheduleApps,
	EventValidateScheduleApplication:    eventValidateScheduleApplication,
	EventAddCredentialandBackupLocation: eventAddCredentialandBackupLocation,
	EventAddSourceAndDestinationCluster: eventAddSourceAndDestinationCluster,
	EventCreateBackup:                   eventCreateBackup,
	EventRestore:                        eventRestore,
}
