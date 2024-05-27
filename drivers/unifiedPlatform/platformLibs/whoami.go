package platformLibs

import "github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"

func Whoami() (*automationModels.WhoamiResponse, error) {
	res, err := v2Components.Platform.WhoAmI()
	if err != nil {
		return nil, err
	}
	return res, nil
}
