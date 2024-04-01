package stworkflows

import (
	resiLibs "github.com/portworx/torpedo/drivers/unifiedPlatform/pdsLibs"
	"github.com/portworx/torpedo/pkg/log"
)

type WorkflowResiliency struct {
	ScenarioType   string
	ErrorType      error
	ResiliencyFlag bool
}

// MarkResiliencyTC Function to enable Resiliency Test
func (wkflwResi *WorkflowResiliency) MarkResiliencyTC(resiliency bool) {
	wkflwResi.ResiliencyFlag = resiliency
	log.InfoD("Execution of a Resiliency TestCase Begins ...")
}
func (wkflwResi *WorkflowResiliency) InduceFailureAndExecuteResiliencyScenario(namespace string, failureType string, method error) error {
	err := resiLibs.InduceFailureAfterWaitingForCondition(namespace, failureType, method, wkflwResi.ResiliencyFlag)
	if err != nil {
		return err
	}

	return nil
}
