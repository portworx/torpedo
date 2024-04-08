package platform

import (
	"fmt"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/platformLibs"
)

type WorkflowCloudCredentials struct {
	Platform         WorkflowPlatform
	CloudCredentials map[string]CloudCredentialsType
}

type CloudCredentialsType struct {
	Name              string
	ID                string
	CloudProviderType string
}

func (cloudCredentials *WorkflowCloudCredentials) CreateCloudCredentials(backUpTargetType string) (*WorkflowCloudCredentials, error) {
	cloudCreds, err := platformLibs.CreateCloudCredentials(cloudCredentials.Platform.TenantId, backUpTargetType)
	if err != nil {
		return cloudCredentials, fmt.Errorf("Failed while creating cloud credentials: %v\n", err)
	}

	cloudCredentials.CloudCredentials[backUpTargetType] = CloudCredentialsType{
		ID:                *cloudCreds.Create.Meta.Uid,
		Name:              *cloudCreds.Create.Meta.Name,
		CloudProviderType: backUpTargetType,
	}

	return cloudCredentials, nil
}

func (cloudCredentials *WorkflowCloudCredentials) DeleteCloudCredentials(cloudCredentialId string) error {
	return platformLibs.DeleteCloudCredential(cloudCredentialId)
}
