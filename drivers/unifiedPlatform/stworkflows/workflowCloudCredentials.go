package stworkflows

import (
	"fmt"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/platformLibs"
)

type WorkflowCloudCredentials struct {
	Platform         WorkflowPlatform
	CloudCredentials map[string]CloudCredentialsType
}

type CloudCredentialsType struct {
	ID                string
	CloudProviderType string
}

func (cloudCredentials *WorkflowCloudCredentials) CreateCloudCredentials(backUpTargetType string) (*WorkflowCloudCredentials, error) {
	cloudCreds, err := platformLibs.CreateCloudCredentials(cloudCredentials.Platform.TenantId, backUpTargetType)
	if err != nil {
		return cloudCredentials, fmt.Errorf("Failed while creating cloud credentials: %v\n", err)
	}

	CC := make(map[string]CloudCredentialsType)

	CC[*cloudCreds.Create.Meta.Name] = CloudCredentialsType{
		ID:                *cloudCreds.Create.Meta.Uid,
		CloudProviderType: backUpTargetType,
	}

	cloudCredentials.CloudCredentials = CC

	return cloudCredentials, nil
}
