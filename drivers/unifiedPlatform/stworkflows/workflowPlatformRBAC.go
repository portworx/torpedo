package stworkflows

import (
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/pkg/log"
)

type WorkflowRBAC struct {
	AccountName string
	AccountRole string
	Platform    WorkflowPlatform
}

// Public method to verify any user access on platform
func (user *WorkflowRBAC) VerifyUserPlatformAccess() error {
	switch user.AccountName {
	case automationModels.User:
		err := user.verifyUserAccess()
		return err
	case automationModels.ProjectAdmin:
		err := user.verifyProjectAdminAccess()
		return err

	case automationModels.TenantAdmin:
		err := user.verifyTenantAdminAccess()
		return err
	default:
		err := user.verifyUserAccess()
		return err
	}
}

// verifyProjectAdminAccess verifies all project admin level access
func (user *WorkflowRBAC) verifyProjectAdminAccess() error {
	defer func() {
		// Switch back to admin user in this
		log.Infof("Switching back to admin")
	}()

	log.Infof("Verifying project admin level access")
	// All project admin level validations need to be made here
	return nil
}

// verifyUserAccess verifies all user level access
func (user *WorkflowRBAC) verifyUserAccess() error {
	defer func() {
		// Switch back to admin user in this
		log.Infof("Switching back to admin")
	}()

	log.Infof("Verifying user level access")
	// All user level validations need to be made here
	return nil
}

// verifyTenantAdminAccess verifies all tenant admin level access
func (user *WorkflowRBAC) verifyTenantAdminAccess() error {
	defer func() {
		// Switch back to admin user in this
		log.Infof("Switching back to admin")
	}()

	log.Infof("Verifying tenant admin level access")
	// All tenant admin level validations need to be made here
	return nil
}
