package pxbackup

//type BackupInfo struct {
//	*api.BackupObject
//}
//
//func (p *PxBackupController) setBackupInfo(backupName string, backupInfo *BackupInfo) {
//	if p.organizations[p.currentOrgId].backups == nil {
//		p.organizations[p.currentOrgId].backups = make(map[string]*BackupInfo, 0)
//	}
//	p.organizations[p.currentOrgId].backups[backupName] = backupInfo
//}
//
//func (p *PxBackupController) GetBackupInfo(backupName string) (*BackupInfo, bool) {
//	backupInfo, ok := p.organizations[p.currentOrgId].backups[backupName]
//	if !ok {
//		return nil, false
//	}
//	return backupInfo, true
//}
//
//func (p *PxBackupController) delBackupInfo(backupName string) {
//	delete(p.organizations[p.currentOrgId].backups, backupName)
//}
//
//type CreateBackupConfig struct {
//	backupName         string
//	backupLocationName string
//	clusterName        string
//	namespaces         []string
//	labelSelectors     map[string]string // default
//	preRuleName        string            // default
//	postRuleName       string            // default
//	resourceTypes      []string          // default
//	nsLabelSelectors   string            // default
//	//backupUid          string            // default
//	controller *PxBackupController // fixed
//}
//
//func (c *CreateBackupConfig) SetLabelSelectors(labelSelectors map[string]string) *CreateBackupConfig {
//	c.labelSelectors = labelSelectors
//	return c
//}
//
//func (c *CreateBackupConfig) SetPreRuleName(preRuleName string) *CreateBackupConfig {
//	c.preRuleName = preRuleName
//	return c
//}
//
//func (c *CreateBackupConfig) SetPostRuleName(postRuleName string) *CreateBackupConfig {
//	c.postRuleName = postRuleName
//	return c
//}
//
//func (c *CreateBackupConfig) SetResourceTypes(resourceTypes []string) *CreateBackupConfig {
//	c.resourceTypes = resourceTypes
//	return c
//}
//
//func (c *CreateBackupConfig) SetNsLabelSelectors(nsLabelSelectors string) *CreateBackupConfig {
//	c.nsLabelSelectors = nsLabelSelectors
//	return c
//}
//
////func (c *CreateBackupConfig) SetBackupUid(backupUid string) *CreateBackupConfig {
////	c.backupUid = backupUid
////	return c
////}
//
//func (p *PxBackupController) getBackupUid(backupName string) (string, error) {
//	var backupUid *string
//	backupEnumerateReq := &api.BackupEnumerateRequest{
//		OrgId: p.currentOrgId,
//	}
//	resp, err := p.processPxBackupRequest(backupEnumerateReq)
//	if err != nil {
//		return "", err
//	}
//	backups := resp.(*api.BackupEnumerateResponse).GetBackups()
//	for _, backup := range backups {
//		if backup.GetName() == backupName {
//			backupUid = &backup.Uid
//		}
//	}
//	if backupUid == nil {
//		return "", fmt.Errorf("backup with name '%s' not found for org '%s'", backupName, p.currentOrgId)
//	}
//	return *backupUid, nil
//}
//
//func (p *PxBackupController) Backup(backupName string, backupLocationName string, clusterName string, namespaces []string) *CreateBackupConfig {
//	return &CreateBackupConfig{
//		backupName:         backupName,
//		backupLocationName: backupLocationName,
//		clusterName:        clusterName,
//		namespaces:         namespaces,
//		labelSelectors:     make(map[string]string, 0),
//		preRuleName:        "",
//		postRuleName:       "",
//		resourceTypes:      make([]string, 0),
//		nsLabelSelectors:   "",
//		//backupUid:          uuid.New(),
//		controller: p,
//	}
//}
//
//func (c *CreateBackupConfig) Create() error {
//	clusterInfo, ok := c.controller.GetClusterInfo(c.clusterName)
//	if !ok {
//		return fmt.Errorf("cluster not found in cache")
//	}
//	backupLocationInfo, ok := c.controller.GetBackupLocationInfo(c.backupLocationName)
//	if !ok {
//		return fmt.Errorf("backup location not found in cache")
//	}
//	var preExecRuleRef = &api.ObjectRef{}
//	if preRuleInfo, ok := c.controller.GetRuleInfo(c.preRuleName); ok {
//		preExecRuleRef = &api.ObjectRef{
//			Name: c.preRuleName,
//			Uid:  preRuleInfo.Uid,
//		}
//	}
//	var postExecRuleRef = &api.ObjectRef{}
//	if postRuleInfo, ok := c.controller.GetRuleInfo(c.postRuleName); ok {
//		postExecRuleRef = &api.ObjectRef{
//			Name: c.postRuleName,
//			Uid:  postRuleInfo.Uid,
//		}
//	}
//	log.Infof("Creating backup [%s]", c.backupName)
//	backupCreateRequest := &api.BackupCreateRequest{
//		CreateMetadata: &api.CreateMetadata{
//			Name:  c.backupName,
//			OrgId: c.controller.currentOrgId,
//			//Uid:   c.backupUid,
//		},
//		BackupLocationRef: &api.ObjectRef{
//			Name: backupLocationInfo.Name,
//			Uid:  backupLocationInfo.Uid,
//		},
//		Cluster:        c.clusterName,
//		Namespaces:     c.namespaces,
//		LabelSelectors: c.labelSelectors,
//		ClusterRef: &api.ObjectRef{
//			Name: c.clusterName,
//			Uid:  clusterInfo.GetUid(),
//		},
//		PreExecRuleRef:  preExecRuleRef,
//		PostExecRuleRef: postExecRuleRef,
//	}
//	if _, err := c.controller.processPxBackupRequest(backupCreateRequest); err != nil {
//		return err
//	}
//	backupUid, err := c.controller.getBackupUid(c.backupName)
//	if err != nil {
//		return err
//	}
//	backupInspectReq := &api.BackupInspectRequest{
//		OrgId: c.controller.currentOrgId,
//		Name:  c.backupName,
//		//Uid:   c.backupUid,
//		Uid: backupUid,
//	}
//	log.Infof("Backup uid: %s", backupUid)
//	resp, err := c.controller.processPxBackupRequest(backupInspectReq)
//	if err != nil {
//		return err
//	}
//	backup := resp.(*api.BackupInspectResponse).GetBackup()
//	c.controller.setBackupInfo(c.backupName, &BackupInfo{
//		BackupObject: backup,
//	})
//	return nil
//}
//
//func (p *PxBackupController) DeleteBackup(backupName string) error {
//	backupInfo, ok := p.GetBackupInfo(backupName)
//	if !ok {
//		return fmt.Errorf("backup [%s] not found in cache", backupName)
//	}
//	log.Infof("Deleting backup [%s] of org [%s]", backupName, backupInfo.OrgId)
//	backupDeleteReq := &api.BackupDeleteRequest{
//		Name:    backupName,
//		OrgId:   backupInfo.OrgId,
//		Cluster: backupInfo.Cluster,
//		Uid:     backupInfo.Uid,
//	}
//	if _, err := p.processPxBackupRequest(backupDeleteReq); err != nil {
//		return err
//	}
//	return nil
//}
//
//type backupStatusInfo struct {
//	Status api.BackupInfo_StatusInfo_Status
//	Reason string
//	Error  error
//}
//
//func (p *PxBackupController) getBackupStatus(backupName string) (backupStatusInfo, error) {
//	backupInfo, ok := p.GetBackupInfo(backupName)
//	if !ok {
//		return backupStatusInfo{}, fmt.Errorf("backup [%s] not found in cache", backupName)
//	}
//	backupInspectRequest := &api.BackupInspectRequest{
//		OrgId: backupInfo.OrgId,
//		Name:  backupInfo.Name,
//		Uid:   backupInfo.Uid,
//	}
//	resp, err := p.processPxBackupRequest(backupInspectRequest)
//	if err != nil {
//		return backupStatusInfo{
//			Status: api.BackupInfo_StatusInfo_Invalid,
//			Reason: "",
//			Error:  err,
//		}, nil
//	}
//	backup := resp.(*api.BackupInspectResponse).GetBackup()
//	status := backup.GetStatus().Status
//	reason := backup.GetStatus().Reason
//	return backupStatusInfo{
//		Status: status,
//		Reason: reason,
//		Error:  nil,
//	}, nil
//}
//
//func (p *PxBackupController) WaitForBackupCompletion(backupName string) (api.BackupInfo_StatusInfo_Status, error) {
//	getBackupStatus := func() interface{} {
//		status, err := p.getBackupStatus(backupName)
//		log.Infof("backup status for [%s] is [%s]", backupName, status)
//		if err != nil {
//			return backupStatusInfo{
//				Status: api.BackupInfo_StatusInfo_Invalid,
//				Error:  err,
//			}
//		}
//		return status
//	}
//	shouldRetry := func(result interface{}) bool {
//		res, ok := result.(backupStatusInfo)
//		if !ok || res.Error != nil {
//			return false
//		}
//		finalStates := [...]api.BackupInfo_StatusInfo_Status{
//			api.BackupInfo_StatusInfo_Invalid,
//			api.BackupInfo_StatusInfo_Aborted,
//			api.BackupInfo_StatusInfo_Failed,
//			api.BackupInfo_StatusInfo_Success,
//			api.BackupInfo_StatusInfo_Captured,
//			api.BackupInfo_StatusInfo_PartialSuccess,
//			api.BackupInfo_StatusInfo_CloudBackupMissing,
//		}
//		log.Infof("backup status for [%s] is [%s] bc [%s]", backupName, res.Status, res.Reason)
//		for _, status := range finalStates {
//			if res.Status == status {
//				return false
//			}
//		}
//		return true
//	}
//	res, err := utils.DoRetryWithTimeout(getBackupStatus, utils.DefaultBackupCreationTimeout, utils.DefaultBackupCreationRetryTime, shouldRetry)
//	if err != nil {
//		return res.(backupStatusInfo).Status, err
//	}
//	return res.(backupStatusInfo).Status, nil
//}
//
//func (p *PxBackupController) WaitForBackupDeletion(backupName string) error {
//	getBackupStatus := func() interface{} {
//		status, err := p.getBackupStatus(backupName)
//		if err != nil {
//			return backupStatusInfo{
//				Status: api.BackupInfo_StatusInfo_Invalid,
//				Error:  err,
//			}
//		}
//		if status.Status != api.BackupInfo_StatusInfo_DeletePending && status.Status != api.BackupInfo_StatusInfo_Deleting {
//			return backupStatusInfo{
//				Status: status.Status,
//				Reason: status.Reason,
//				Error:  fmt.Errorf("backup [%s] is not in [%s] state rather in [%s] state", backupName, [...]api.BackupInfo_StatusInfo_Status{api.BackupInfo_StatusInfo_DeletePending, api.BackupInfo_StatusInfo_Deleting}, status.Status),
//			}
//		}
//		return status
//	}
//	shouldRetry := func(result interface{}) bool {
//		res := result.(backupStatusInfo)
//		if res.Error != nil {
//			return false
//		}
//		if res.Status == api.BackupInfo_StatusInfo_Invalid {
//			return false
//		}
//		log.Infof("Backup status for [%s]: expected %s but got [%s] because [%s]", backupName, api.BackupInfo_StatusInfo_Invalid, res.Status, res.Reason)
//		return true
//	}
//	res, err := utils.DoRetryWithTimeout(getBackupStatus, utils.DefaultBackupDeletionTimeout, utils.DefaultBackupDeletionRetryTime, shouldRetry)
//	if err != nil {
//		return err
//	}
//	if res.(backupStatusInfo).Status != api.BackupInfo_StatusInfo_Invalid {
//		return fmt.Errorf("expected backup status not reached because %v", res.(backupStatusInfo).Error)
//	}
//	p.delBackupInfo(backupName)
//	return nil
//}
