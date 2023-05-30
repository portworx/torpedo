package pxbackup

import (
	"fmt"
	api "github.com/portworx/px-backup-api/pkg/apis/v1"
	"github.com/portworx/torpedo/drivers"
	"github.com/portworx/torpedo/pkg/log"
	"github.com/portworx/torpedo/pkg/s3utils"
)

func (p *PxBackupController) getBackupLocationInfo(backupLocationName string) *BackupLocationInfo {
	backupLocationInfo, ok := p.organizations[p.currentOrgId].backupLocations[backupLocationName]
	if !ok {
		return &BackupLocationInfo{}
	}
	return backupLocationInfo
}

func (p *PxBackupController) saveBackupLocationInfo(backupLocationName string, backupLocationInfo *BackupLocationInfo) {
	if p.organizations[p.currentOrgId].backupLocations == nil {
		p.organizations[p.currentOrgId].backupLocations = make(map[string]*BackupLocationInfo, 0)
	}
	p.organizations[p.currentOrgId].backupLocations[backupLocationName] = backupLocationInfo
}

func (p *PxBackupController) delBackupLocationInfo(backupLocationName string) {
	delete(p.organizations[p.currentOrgId].backupLocations, backupLocationName)
}

func (p *PxBackupController) isBackupLocationRecorded(backupLocationName string) bool {
	_, ok := p.organizations[p.currentOrgId].backupLocations[backupLocationName]
	return ok
}

type BackupLocationConfig struct {
	backupLocationName string
	backupLocationUid  string
	encryptionKey      string
	isRecorded         bool
	controller         *PxBackupController
}

func (c *BackupLocationConfig) AddAsObjectStore(cloudAccountName string, bucketName string) error {
	if !c.controller.isCloudAccountRecorded(cloudAccountName) {
		return fmt.Errorf("cloud-account [%s] is not in records", cloudAccountName)
	}
	cloudAccountInfo := c.controller.getCloudAccountInfo(cloudAccountName)
	var backupLocationCreateReq *api.BackupLocationCreateRequest
	switch cloudAccountInfo.provider {
	case drivers.ProviderAws:
		_, _, endpoint, region, disableSSLBool := s3utils.GetAWSDetailsFromEnv()
		backupLocationCreateReq = &api.BackupLocationCreateRequest{
			CreateMetadata: &api.CreateMetadata{
				Name:  c.backupLocationName,
				OrgId: c.controller.currentOrgId,
				Uid:   c.backupLocationUid,
			},
			BackupLocation: &api.BackupLocationInfo{
				Path:          bucketName,
				EncryptionKey: c.encryptionKey,
				CloudCredentialRef: &api.ObjectRef{
					Name: cloudAccountName,
					Uid:  cloudAccountInfo.GetUid(),
				},
				Type: api.BackupLocationInfo_S3,
				Config: &api.BackupLocationInfo_S3Config{
					S3Config: &api.S3Config{
						Endpoint:   endpoint,
						Region:     region,
						DisableSsl: disableSSLBool,
					},
				},
			},
		}
	case drivers.ProviderAzure:
		backupLocationCreateReq = &api.BackupLocationCreateRequest{
			CreateMetadata: &api.CreateMetadata{
				Name:  c.backupLocationName,
				OrgId: c.controller.currentOrgId,
				Uid:   c.backupLocationUid,
			},
			BackupLocation: &api.BackupLocationInfo{
				Path:          bucketName,
				EncryptionKey: c.encryptionKey,
				CloudCredentialRef: &api.ObjectRef{
					Name: cloudAccountName,
					Uid:  c.controller.currentOrgId,
				},
				Type: api.BackupLocationInfo_Azure,
			},
		}
	default:
		return fmt.Errorf("unsupported cloud provider [%s] for backup location; supported providers: %s", cloudAccountInfo.provider, []string{drivers.ProviderAws, drivers.ProviderAzure})
	}
	log.Infof("Adding backup location [%s] for org [%s] and provider [%s]", c.backupLocationName, c.controller.currentOrgId, cloudAccountInfo.provider)
	_, err := c.controller.processPxBackupRequest(backupLocationCreateReq)
	if err != nil {
		return err
	}
	backupLocationInspectInspectReq := &api.BackupLocationInspectRequest{
		OrgId:          c.controller.currentOrgId,
		Name:           c.backupLocationName,
		IncludeSecrets: false,
		Uid:            c.backupLocationUid,
	}
	resp, err := c.controller.processPxBackupRequest(backupLocationInspectInspectReq)
	if err != nil {
		return err
	}
	backupLocation := resp.(*api.BackupLocationInspectResponse).GetBackupLocation()
	newBackupLocationInfo := &BackupLocationInfo{
		BackupLocationObject: backupLocation,
	}
	c.controller.saveBackupLocationInfo(c.backupLocationName, newBackupLocationInfo)
	return nil
}

func (c *BackupLocationConfig) AddAsNFS(serverAddress string) error {
	backupLocationCreateReq := &api.BackupLocationCreateRequest{
		CreateMetadata: &api.CreateMetadata{
			Name:  c.backupLocationName,
			OrgId: c.controller.currentOrgId,
			Uid:   c.backupLocationUid,
		},
		BackupLocation: &api.BackupLocationInfo{
			Config: &api.BackupLocationInfo_NfsConfig{
				NfsConfig: &api.NFSConfig{
					ServerAddr:  serverAddress,
					SubPath:     subPath,
					MountOption: mountOption,
				},
			},
			Path:          path,
			Type:          api.BackupLocationInfo_NFS,
			EncryptionKey: c.encryptionKey,
		},
	}
	log.Infof("Adding backup location [%s] for org [%s] and provider [%s]", c.backupLocationName, c.controller.currentOrgId, "nfs")
	_, err := c.controller.processPxBackupRequest(backupLocationCreateReq)
	if err != nil {
		return err
	}
	backupLocationInspectInspectReq := &api.BackupLocationInspectRequest{
		OrgId:          c.controller.currentOrgId,
		Name:           c.backupLocationName,
		IncludeSecrets: false,
		Uid:            c.backupLocationUid,
	}
	resp, err := c.controller.processPxBackupRequest(backupLocationInspectInspectReq)
	if err != nil {
		return err
	}
	backupLocation := resp.(*api.BackupLocationInspectResponse).GetBackupLocation()
	newBackupLocationInfo := &BackupLocationInfo{
		BackupLocationObject: backupLocation,
	}
	c.controller.saveBackupLocationInfo(c.backupLocationName, newBackupLocationInfo)
	return nil
}

func (p *PxBackupController) DeleteBackupLocation(backupLocationName string) error {
	//backupLocationInfo, ok := p.getBackupLocationInfo(backupLocationName)
	//if ok {
	//	log.Infof("Deleting backup location [%s] of org [%s]", backupLocationName, p.currentOrgId)
	//	backupLocationDeleteReq := &api.BackupLocationDeleteRequest{
	//		Name:  backupLocationName,
	//		OrgId: p.currentOrgId,
	//		Uid:   backupLocationInfo.GetUid(),
	//	}
	//	if _, err := p.processPxBackupRequest(backupLocationDeleteReq); err != nil {
	//		return err
	//	}
	//	p.delBackupLocationInfo(backupLocationName)
	//}
	return nil
}
