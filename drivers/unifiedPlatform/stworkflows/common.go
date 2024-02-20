package stworkflows

import "github.com/portworx/torpedo/pkg/log"

const (
	envControlPlaneUrl     = "CONTROL_PLANE_URL"
	defaultTestAccount     = "demo-milestone-one"
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
