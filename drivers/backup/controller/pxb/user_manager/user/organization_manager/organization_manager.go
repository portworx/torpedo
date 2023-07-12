package organization_manager

import (
	. "github.com/portworx/torpedo/drivers/backup/controller/generics/entity/entity_manager"
	. "github.com/portworx/torpedo/drivers/backup/controller/pxb/user_manager/user/organization_manager/organization"
)

// OrganizationManager represents a manager for an Organization
type OrganizationManager EntityManager[*Organization]
