package pxbackup

import (
	"fmt"
	"github.com/pborman/uuid"
	api "github.com/portworx/px-backup-api/pkg/apis/v1"
	"github.com/portworx/torpedo/drivers"
	"github.com/portworx/torpedo/pkg/log"
	"github.com/portworx/torpedo/pkg/s3utils"
	"github.com/portworx/torpedo/tests"
)

type CloudAccountInfo struct {
	*api.CloudCredentialObject
	provider string
}

func (p *PxbController) setCloudAccountInfo(cloudAccountName string, cloudAccountInfo *CloudAccountInfo) {
	if p.organizations[p.currentOrgId].cloudAccounts == nil {
		p.organizations[p.currentOrgId].cloudAccounts = make(map[string]*CloudAccountInfo, 0)
	}
	p.organizations[p.currentOrgId].cloudAccounts[cloudAccountName] = cloudAccountInfo
}

func (p *PxbController) GetCloudAccountInfo(cloudAccountName string) (*CloudAccountInfo, bool) {
	cloudAccountInfo, ok := p.organizations[p.currentOrgId].cloudAccounts[cloudAccountName]
	if !ok {
		return nil, false
	}
	return cloudAccountInfo, true
}

func (p *PxbController) delCloudAccountInfo(cloudAccountName string) {
	delete(p.organizations[p.currentOrgId].cloudAccounts, cloudAccountName)
}

type AddCloudAccountConfig struct {
	cloudProvider    string
	cloudAccountName string
	cloudAccountUid  string         // default
	controller       *PxbController // fixed
}

func (c *AddCloudAccountConfig) SetCloudAccountUid(cloudAccountUid string) *AddCloudAccountConfig {
	c.cloudAccountUid = cloudAccountUid
	return c
}

func (p *PxbController) CloudAccount(cloudAccountName string, cloudProvider string) *AddCloudAccountConfig {
	return &AddCloudAccountConfig{
		cloudAccountName: cloudAccountName,
		cloudProvider:    cloudProvider,
		cloudAccountUid:  uuid.New(),
		controller:       p,
	}
}

func (c *AddCloudAccountConfig) Add() error {
	log.Infof("Adding cloud account [%s] for org [%s] and provider [%s]", c.cloudAccountName, c.controller.currentOrgId, c.cloudProvider)
	var cloudCredentialCreateReq *api.CloudCredentialCreateRequest
	switch c.cloudProvider {
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
		return fmt.Errorf("provider [%s] not supported for adding cloud account; supported providers: %s", c.cloudProvider, []string{drivers.ProviderAws, drivers.ProviderAzure})
	}
	_, err := c.controller.processPxBackupRequest(cloudCredentialCreateReq)
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
	c.controller.setCloudAccountInfo(c.cloudAccountName, &CloudAccountInfo{
		CloudCredentialObject: cloudAccount,
		provider:              c.cloudProvider,
	})
	return nil
}

func (p *PxbController) DeleteCloudAccount(cloudAccountName string) error {
	cloudAccountInfo, ok := p.GetCloudAccountInfo(cloudAccountName)
	if ok {
		log.Infof("Deleting cloud account [%s] of org [%s]", cloudAccountName, p.currentOrgId)
		cloudAccountDeleteReq := &api.CloudCredentialDeleteRequest{
			Name:  cloudAccountName,
			OrgId: p.currentOrgId,
			Uid:   cloudAccountInfo.GetUid(),
		}
		if _, err := p.processPxBackupRequest(cloudAccountDeleteReq); err != nil {
			return err
		}
		p.delCloudAccountInfo(cloudAccountName)
	}
	return nil
}
