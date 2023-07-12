package cloud_credential_manager

import (
	. "github.com/portworx/torpedo/drivers/backup/controller/pxb/user_manager/user/organization_manager/organization/cloud_credential_manager/cloud_credential"
	. "github.com/portworx/torpedo/drivers/pxbackup/controller/pxb/user_manager/user/organization_manager/organization/cloud_credential_manager/cloud_credential"
	. "github.com/portworx/torpedo/drivers/pxbackup/controller_utils/entity/entity_config/entity_manager"
)

type (
	AWSCredentialManager = EntityManager[*AWSCredential]
)

// CloudCredentialManager represents a manager for a cloud_credential.CloudCredential
type CloudCredentialManager struct {
	AWSCredentialManager *AWSCredentialManager
}

// GetAWSCredentialManager returns the AWSCredentialManager associated with the CloudCredentialManager
func (m *CloudCredentialManager) GetAWSCredentialManager() *AWSCredentialManager {
	return m.AWSCredentialManager
}

// SetAWSCredentialManager sets the AWSCredentialManager for the CloudCredentialManager
func (m *CloudCredentialManager) SetAWSCredentialManager(manager *AWSCredentialManager) *CloudCredentialManager {
	m.AWSCredentialManager = manager
	return m
}

// NewCloudCredentialManager creates a new instance of the CloudCredentialManager
func NewCloudCredentialManager(awsCredentialManager *AWSCredentialManager) *CloudCredentialManager {
	cloudCredentialManager := &CloudCredentialManager{}
	cloudCredentialManager.SetAWSCredentialManager(awsCredentialManager)
	return cloudCredentialManager
}

// NewDefaultCloudCredentialManager creates a new instance of the CloudCredentialManager with default values
func NewDefaultCloudCredentialManager() *CloudCredentialManager {
	return NewCloudCredentialManager(NewDefaultEntityManager[*AWSCredential]())
}
