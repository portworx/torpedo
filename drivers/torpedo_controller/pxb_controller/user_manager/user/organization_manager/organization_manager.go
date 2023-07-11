package organization_manager

import (
	. "github.com/portworx/torpedo/drivers/torpedo_controller/pxb_controller/user_manager/user/organization_manager/organization"
	. "github.com/portworx/torpedo/drivers/torpedo_controller/torpedo_utils/entity_generics"
)

// OrganizationManager represents a manager for an Organization
type OrganizationManager EntityManager[*Organization]
