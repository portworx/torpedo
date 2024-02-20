package platform

import "github.com/portworx/torpedo/drivers/pds/parameters"

const (
	envControlPlaneUrl     = "CONTROL_PLANE_URL"
	defaultTestAccount     = "pds-qa"
	envPlatformAccountName = "PLATFORM_ACCOUNT_NAME"
	envAccountDisplayName  = "PLATFORM_ACCOUNT_DISPLAY_NAME"
	envUserMailId          = "USER_MAIL_ID"
)

var (
	Params       *parameters.NewPDSParams
	customParams *parameters.Customparams
	pdsLabels    = make(map[string]string)
)
