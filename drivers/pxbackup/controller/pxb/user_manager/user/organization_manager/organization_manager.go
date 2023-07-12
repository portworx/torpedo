package organization_manager

import (
	. "github.com/portworx/torpedo/drivers/pxbackup/controller/pxb/user_manager/user/organization_manager/organization"
	. "github.com/portworx/torpedo/drivers/pxbackup/controller_utils/entity/entity_config/entity_manager"
)

// OrganizationManager represents a manager for an Organization
type OrganizationManager EntityManager[*Organization]
