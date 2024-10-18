package platform

import (
	"github.com/pure-px/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/pure-px/torpedo/drivers/unifiedPlatform/platformLibs"
	"github.com/pure-px/torpedo/pkg/log"
	"strconv"
	"time"
)

type WorkflowSchedule struct {
	Platform  *WorkflowPlatform
	Schedules map[string]*automationModels.V1BackupPolicy
}

var (
	typeOfPolicy = automationModels.IntervalPolicy
)

// CreateSchedule will create a schedule with given name, minutes and retainCount
func (workflowSchedule *WorkflowSchedule) CreateSchedule(name string, minutes time.Duration, retainCount string) (*automationModels.V1BackupPolicy, error) {

	minutesToStr := strconv.Itoa(int(minutes / time.Minute))
	log.Infof("Creating schedule - [%s] with interval - [%s] minutes and retain count - [%s]", name, minutesToStr, retainCount)

	schedule, err := platformLibs.CreateBackupPolicy(name, typeOfPolicy, minutesToStr, retainCount, workflowSchedule.Platform.TenantId)
	if err != nil {
		return &automationModels.V1BackupPolicy{}, err
	}

	workflowSchedule.Schedules[name] = &schedule
	return &schedule, nil
}

// GetSchedule will get the schedule details of given schedule name
func (workflowSchedule *WorkflowSchedule) GetSchedule(id string) (*automationModels.V1BackupPolicy, error) {
	schedule, err := platformLibs.GetBackupPolicy(id)
	if err != nil {
		return &automationModels.V1BackupPolicy{}, err
	}

	return &schedule, nil
}

// DeleteSchedule will delete the schedule with given schedule name
func (workflowSchedule *WorkflowSchedule) DeleteSchedule(id string) error {
	err := platformLibs.DeleteBackupPolicy(id)
	if err != nil {
		return err
	}

	delete(workflowSchedule.Schedules, id)

	return nil
}

// Purge will delete all schedules
func (workflowSchedule *WorkflowSchedule) Purge() error {
	for _, schedule := range workflowSchedule.Schedules {
		log.Infof("Deleting schedule - [%s]", *schedule.Meta.Uid)
		err := workflowSchedule.DeleteSchedule(*schedule.Meta.Uid)
		if err != nil {
			return err
		}
	}

	return nil
}
