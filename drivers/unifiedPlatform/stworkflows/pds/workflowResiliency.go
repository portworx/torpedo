package pds

import (
	resiLibs "github.com/portworx/torpedo/drivers/unifiedPlatform/pdsLibs"
	"github.com/portworx/torpedo/pkg/log"
)

type WorkflowPDSResiliency struct {
	ScenarioType   string
	ErrorType      error
	ResiliencyFlag bool
}

// MarkResiliencyTC Function to enable Resiliency Test
func (wkflwResi *WorkflowPDSResiliency) MarkResiliencyTC(resiliency bool) {
	wkflwResi.ResiliencyFlag = resiliency
	log.InfoD("Execution of a Resiliency TestCase Begins ...")
}
func (wkflwResi *WorkflowPDSResiliency) InduceFailureAndExecuteResiliencyScenario(namespace string, failureType string) error {
	err := resiLibs.InduceFailureAfterWaitingForCondition(namespace, failureType, wkflwResi.ResiliencyFlag)
	if err != nil {
		return err
	}

	return nil
}
