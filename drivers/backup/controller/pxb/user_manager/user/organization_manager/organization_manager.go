package organization_manager

import (
	. "github.com/portworx/torpedo/drivers/backup/controller/pxb/user_manager/user/organization_manager/organization"
	. "github.com/portworx/torpedo/drivers/backup/controller/torpedo/torpedo_utils/entity_generics"
)

// OrganizationManager represents a manager for an Organization
type OrganizationManager EntityManager[*Organization]
