package pxbackup

//type BackupLocationInfo struct {
//	*api.BackupLocationObject
//	bucketName string
//}
//
//func (p *PxBackupController) setBackupLocationInfo(backupLocationName string, backupLocationInfo *BackupLocationInfo) {
//	if p.organizations[p.currentOrgId].backupLocations == nil {
//		p.organizations[p.currentOrgId].backupLocations = make(map[string]*BackupLocationInfo, 0)
//	}
//	p.organizations[p.currentOrgId].backupLocations[backupLocationName] = backupLocationInfo
//}
//
//func (p *PxBackupController) GetBackupLocationInfo(backupLocationName string) (*BackupLocationInfo, bool) {
//	backupLocationInfo, ok := p.organizations[p.currentOrgId].backupLocations[backupLocationName]
//	if !ok {
//		return nil, false
//	}
//	return backupLocationInfo, true
//}
//
//func (p *PxBackupController) delBackupLocationInfo(backupLocationName string) {
//	delete(p.organizations[p.currentOrgId].backupLocations, backupLocationName)
//}
//
//type AddBackupLocationConfig struct {
//	cloudAccountName   string
//	backupLocationName string
//	bucketName         string
//	backupLocationUid  string         // default
//	encryptionKey      string         // default
//	controller         *PxBackupController // fixed
//}
//
//func (c *AddBackupLocationConfig) SetEncryptionKey(encryptionKey string) *AddBackupLocationConfig {
//	c.encryptionKey = encryptionKey
//	return c
//}
//
//func (c *AddBackupLocationConfig) SetBackupLocationUid(backupLocationUid string) *AddBackupLocationConfig {
//	c.backupLocationUid = backupLocationUid
//	return c
//}
//
//func (p *PxBackupController) BackupLocation(backupLocationName string, cloudAccountName string, bucketName string) *AddBackupLocationConfig {
//	return &AddBackupLocationConfig{
//		backupLocationName: backupLocationName,
//		cloudAccountName:   cloudAccountName,
//		bucketName:         bucketName,
//		encryptionKey:      "torpedo",
//		backupLocationUid:  uuid.New(),
//		controller:         p,
//	}
//}
//
//func (c *AddBackupLocationConfig) Add() error {
//	cloudAccountInfo, ok := c.controller.GetCloudAccountInfo(c.cloudAccountName)
//	if !ok {
//		return fmt.Errorf("cloud account [%s] not found in cache", c.cloudAccountName)
//	}
//	log.Infof("Adding backup location [%s] for org [%s] and provider [%s]", c.backupLocationName, c.controller.currentOrgId, cloudAccountInfo.provider)
//	var backupLocationCreateReq *api.BackupLocationCreateRequest
//	switch cloudAccountInfo.provider {
//	case drivers.ProviderAws:
//		_, _, endpoint, region, disableSSLBool := s3utils.GetAWSDetailsFromEnv()
//		backupLocationCreateReq = &api.BackupLocationCreateRequest{
//			CreateMetadata: &api.CreateMetadata{
//				Name:  c.backupLocationName,
//				OrgId: c.controller.currentOrgId,
//				Uid:   c.backupLocationUid,
//			},
//			BackupLocation: &api.BackupLocationInfo{
//				Path:          c.bucketName,
//				EncryptionKey: c.encryptionKey,
//				CloudCredentialRef: &api.ObjectRef{
//					Name: c.cloudAccountName,
//					Uid:  cloudAccountInfo.GetUid(),
//				},
//				Type: api.BackupLocationInfo_S3,
//				Config: &api.BackupLocationInfo_S3Config{
//					S3Config: &api.S3Config{
//						Endpoint:   endpoint,
//						Region:     region,
//						DisableSsl: disableSSLBool,
//					},
//				},
//			},
//		}
//	case drivers.ProviderAzure:
//		backupLocationCreateReq = &api.BackupLocationCreateRequest{
//			CreateMetadata: &api.CreateMetadata{
//				Name:  c.backupLocationName,
//				OrgId: c.controller.currentOrgId,
//				Uid:   c.backupLocationUid,
//			},
//			BackupLocation: &api.BackupLocationInfo{
//				Path:          c.bucketName,
//				EncryptionKey: c.encryptionKey,
//				CloudCredentialRef: &api.ObjectRef{
//					Name: c.cloudAccountName,
//					Uid:  c.controller.currentOrgId,
//				},
//				Type: api.BackupLocationInfo_Azure,
//			},
//		}
//	default:
//		return fmt.Errorf("unsupported cloud provider [%s] for backup location; supported providers: %s", cloudAccountInfo.provider, []string{drivers.ProviderAws, drivers.ProviderAzure})
//	}
//	_, err := c.controller.processPxBackupRequest(backupLocationCreateReq)
//	if err != nil {
//		return err
//	}
//	backupLocationInspectInspectReq := &api.BackupLocationInspectRequest{
//		OrgId:          c.controller.currentOrgId,
//		Name:           c.backupLocationName,
//		IncludeSecrets: false,
//		Uid:            c.backupLocationUid,
//	}
//	resp, err := c.controller.processPxBackupRequest(backupLocationInspectInspectReq)
//	if err != nil {
//		return err
//	}
//	backupLocation := resp.(*api.BackupLocationInspectResponse).GetBackupLocation()
//	c.controller.setBackupLocationInfo(c.backupLocationName, &BackupLocationInfo{
//		BackupLocationObject: backupLocation,
//		bucketName:           c.bucketName,
//	})
//	return nil
//}
//
//func (p *PxBackupController) DeleteBackupLocation(backupLocationName string) error {
//	backupLocationInfo, ok := p.GetBackupLocationInfo(backupLocationName)
//	if ok {
//		log.Infof("Deleting backup location [%s] of org [%s]", backupLocationName, p.currentOrgId)
//		backupLocationDeleteReq := &api.BackupLocationDeleteRequest{
//			Name:  backupLocationName,
//			OrgId: p.currentOrgId,
//			Uid:   backupLocationInfo.GetUid(),
//		}
//		if _, err := p.processPxBackupRequest(backupLocationDeleteReq); err != nil {
//			return err
//		}
//		p.delBackupLocationInfo(backupLocationName)
//	}
//	return nil
//}
