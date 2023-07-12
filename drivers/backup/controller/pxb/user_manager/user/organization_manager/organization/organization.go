package organization

import (
	. "github.com/portworx/torpedo/drivers/backup/controller/pxb/user_manager/user/organization_manager/organization/cloud_credential_manager"
	. "github.com/portworx/torpedo/drivers/backup/controller/pxb/user_manager/user/organization_spec"
)

// Organization represents Organization
type Organization struct {
	Spec                   *OrganizationSpec
	CloudCredentialManager *CloudCredentialManager
}

// GetSpec returns the Spec associated with the Organization
func (o *Organization) GetSpec() *OrganizationSpec {
	return o.Spec
}

// SetSpec sets the Spec for the Organization
func (o *Organization) SetSpec(spec *OrganizationSpec) *Organization {
	o.Spec = spec
	return o
}

// GetCloudCredentialManager returns the CloudCredentialManager associated with the Organization
func (o *Organization) GetCloudCredentialManager() *CloudCredentialManager {
	return o.CloudCredentialManager
}

// SetCloudCredentialManager sets the CloudCredentialManager for the Organization
func (o *Organization) SetCloudCredentialManager(manager *CloudCredentialManager) *Organization {
	o.CloudCredentialManager = manager
	return o
}
