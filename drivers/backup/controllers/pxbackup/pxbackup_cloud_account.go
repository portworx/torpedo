package pxbackup

import (
	"fmt"
	api "github.com/portworx/px-backup-api/pkg/apis/v1"
	"github.com/portworx/torpedo/drivers"
	"github.com/portworx/torpedo/drivers/backup/utils"
	"github.com/portworx/torpedo/pkg/log"
	"github.com/portworx/torpedo/pkg/s3utils"
	"github.com/portworx/torpedo/tests"
)

type CloudAccountConfig struct {
	cloudAccountName string
	cloudAccountUid  string
	isRecorded       bool
	controller       *PxBackupController
}

func (c *CloudAccountConfig) validate() error {
	if c.cloudAccountName == "" {
		err := fmt.Errorf("cloud-account name cannot be empty")
		return utils.ProcessError(err)
	} else if len(c.cloudAccountName) < GlobalMinCloudAccountNameLength {
		err := fmt.Errorf("cloud-account name [%s] must have a minimum length of [%d] characters", c.cloudAccountName, GlobalMinCloudAccountNameLength)
		return utils.ProcessError(err)
	} else if c.cloudAccountUid == "" {
		err := fmt.Errorf("cloud-account uid cannot be empty")
		return utils.ProcessError(err)
	}
	return nil
}

func (c *CloudAccountConfig) Add(cloudProvider string) error {
	if c.isRecorded {
		err := fmt.Errorf("cloud-account [%s] is already recorded", c.cloudAccountName)
		return utils.ProcessError(err)
	}
	err := c.validate()
	if err != nil {
		return utils.ProcessError(err)
	}
	var cloudCredentialCreateReq *api.CloudCredentialCreateRequest
	switch cloudProvider {
	case drivers.ProviderAws:
		id, secret, _, _, _ := s3utils.GetAWSDetailsFromEnv()
		cloudCredentialCreateReq = &api.CloudCredentialCreateRequest{
			CreateMetadata: &api.CreateMetadata{
				Name:  c.cloudAccountName,
				Uid:   c.cloudAccountUid,
				OrgId: c.controller.currentOrgId,
			},
			CloudCredential: &api.CloudCredentialInfo{
				Type: api.CloudCredentialInfo_AWS,
				Config: &api.CloudCredentialInfo_AwsConfig{
					AwsConfig: &api.AWSConfig{
						AccessKey: id,
						SecretKey: secret,
					},
				},
			},
		}
	case drivers.ProviderAzure:
		tenantID, clientID, clientSecret, subscriptionID, accountName, accountKey := tests.GetAzureCredsFromEnv()
		cloudCredentialCreateReq = &api.CloudCredentialCreateRequest{
			CreateMetadata: &api.CreateMetadata{
				Name:  c.cloudAccountName,
				Uid:   c.cloudAccountUid,
				OrgId: c.controller.currentOrgId,
			},
			CloudCredential: &api.CloudCredentialInfo{
				Type: api.CloudCredentialInfo_Azure,
				Config: &api.CloudCredentialInfo_AzureConfig{
					AzureConfig: &api.AzureConfig{
						TenantId:       tenantID,
						ClientId:       clientID,
						ClientSecret:   clientSecret,
						AccountName:    accountName,
						AccountKey:     accountKey,
						SubscriptionId: subscriptionID,
					},
				},
			},
		}
	default:
		return fmt.Errorf("provider [%s] not supported for adding cloud-account; supported providers: [%s]", cloudProvider, []string{drivers.ProviderAws, drivers.ProviderAzure})
	}
	log.Infof("Adding cloud-account [%s] for org [%s] and provider [%s]", c.cloudAccountName, c.controller.currentOrgId, cloudProvider)
	_, err = c.controller.processPxBackupRequest(cloudCredentialCreateReq)
	if err != nil {
		return err
	}
	cloudCredentialInspectReq := &api.CloudCredentialInspectRequest{
		OrgId:          c.controller.currentOrgId,
		Name:           c.cloudAccountName,
		IncludeSecrets: false,
		Uid:            c.cloudAccountUid,
	}
	resp, err := c.controller.processPxBackupRequest(cloudCredentialInspectReq)
	if err != nil {
		return err
	}
	cloudAccount := resp.(*api.CloudCredentialInspectResponse).GetCloudCredential()
	c.controller.saveCloudAccountInfo(c.cloudAccountName, &CloudAccountInfo{
		CloudCredentialObject: cloudAccount,
		provider:              cloudProvider,
	})
	return nil
}

func (c *CloudAccountConfig) Delete() error {
	if !c.isRecorded {
		err := fmt.Errorf("does not exist")
		return utils.ProcessError(err)
	}
	cloudAccountDeleteReq := &api.CloudCredentialDeleteRequest{
		Name:  c.cloudAccountName,
		OrgId: c.controller.currentOrgId,
		Uid:   c.cloudAccountUid,
	}
	if _, err := c.controller.processPxBackupRequest(cloudAccountDeleteReq); err != nil {
		return err
	}
	c.controller.delCloudAccountInfo(c.cloudAccountName)
	return nil
}
