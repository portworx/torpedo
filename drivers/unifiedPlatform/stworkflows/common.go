package stworkflows

import (
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/pkg/log"
)

const (
	COMPLETED = "COMPLETED"
	FAILED    = "FAILED"
)

var SKIPDATASERVICEFROMWORKLOAD = []string{
	"sqlserver",
	"mongodb",
}

const (
	envPlatformAccountName = "PLATFORM_ACCOUNT_NAME"
	envAccountDisplayName  = "PLATFORM_ACCOUNT_DISPLAY_NAME"
	envUserMailId          = "USER_MAIL_ID"
)

func startStep(name string) {
	log.Infof("---------------------------------------")
	log.Infof("---------------------------------------")
	log.Infof("StepName - %s", name)
	log.Infof("Output Key - %s", name)
	log.Infof("---------------------------------------")
	log.Infof("---------------------------------------")
}

func addResultToResponse(result []automationModels.WorkFlowResponse, stepName string, resultMap map[string][]automationModels.WorkFlowResponse) {
	if _, ok := resultMap[stepName]; ok {
		log.Infof("Already found result for %s in result map, appending result to same key", stepName)
		resultMap[stepName] = append(resultMap[stepName], result...)
	} else {
		log.Infof("Storing result to %s key", stepName)
		resultMap[stepName] = append(resultMap[stepName], result...)
	}
}
