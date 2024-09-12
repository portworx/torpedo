package platform

import (
	"fmt"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/platformLibs"
	"github.com/portworx/torpedo/pkg/log"
	"strings"
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

// Purge method to delete all cloud creds at the end of testcase
func (cloudCredentials *WorkflowCloudCredentials) Purge() error {
	var allError []string

	for name, details := range cloudCredentials.CloudCredentials {
		log.Infof("Deleting [%s] cred", name)
		err := cloudCredentials.DeleteCloudCredentials(details.ID)
		if err != nil {
			allError = append(allError, err.Error())
		} else {
			delete(cloudCredentials.CloudCredentials, name)
		}
	}

	if len(allError) > 0 {
		return fmt.Errorf("%s", strings.Join(allError, "\n"))
	}
	return nil
}
