package platformLibs

import "github.com/portworx/torpedo/drivers/unifiedPlatform"

var (
	v2Components *unifiedPlatform.UnifiedPlatformComponents
	err          error
)

// InitUnifiedApiComponents
func InitUnifiedApiComponents(controlPlaneURL, accountID string) error {
	v2Components, err = unifiedPlatform.NewUnifiedPlatformComponents(controlPlaneURL, accountID)
	if err != nil {
		return err
	}
	return nil
}
