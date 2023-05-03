package pxbackup

import (
	"fmt"
	"github.com/pborman/uuid"
	api "github.com/portworx/px-backup-api/pkg/apis/v1"
	"github.com/portworx/torpedo/pkg/log"
	"github.com/portworx/torpedo/tests/backup/utils"
)

type RestoreInfo struct {
	*api.RestoreObject
}

func (p *PxbController) setRestoreInfo(restoreName string, restoreInfo *RestoreInfo) {
	if p.organizations[p.currentOrgId].restores == nil {
		p.organizations[p.currentOrgId].restores = make(map[string]*RestoreInfo, 0)
	}
	p.organizations[p.currentOrgId].restores[restoreName] = restoreInfo
}

func (p *PxbController) GetRestoreInfo(restoreName string) (*RestoreInfo, bool) {
	restoreInfo, ok := p.organizations[p.currentOrgId].restores[restoreName]
	if !ok {
		return nil, false
	}
	return restoreInfo, true
}

func (p *PxbController) delRestoreInfo(restoreName string) {
	delete(p.organizations[p.currentOrgId].restores, restoreName)
}

type CreateRestoreConfig struct {
	restoreName         string
	backupName          string
	clusterName         string
	namespaceMapping    map[string]string   // default
	storageClassMapping map[string]string   // default
	resources           []*api.ResourceInfo // default
	restoreUid          string              // default
	controller          *PxbController      // fixed
}

func (c *CreateRestoreConfig) SetNamespaceMapping(namespaceMapping map[string]string) *CreateRestoreConfig {
	c.namespaceMapping = namespaceMapping
	return c
}

func (c *CreateRestoreConfig) SetStorageClassMapping(storageClassMapping map[string]string) *CreateRestoreConfig {
	c.storageClassMapping = storageClassMapping
	return c
}

func (c *CreateRestoreConfig) SetResources(resources []*api.ResourceInfo) *CreateRestoreConfig {
	c.resources = resources
	return c
}

func (p *PxbController) Restore(restoreName string, backupName string, clusterName string) *CreateRestoreConfig {
	return &CreateRestoreConfig{
		restoreName:         restoreName,
		backupName:          backupName,
		clusterName:         clusterName,
		namespaceMapping:    make(map[string]string, 0),
		storageClassMapping: make(map[string]string, 0),
		resources:           make([]*api.ResourceInfo, 0),
		restoreUid:          uuid.New(),
		controller:          p,
	}
}

func (c *CreateRestoreConfig) Create() error {
	backupInfo, ok := c.controller.GetBackupInfo(c.backupName)
	if !ok {
		return fmt.Errorf("backup not found in cache")
	}
	clusterInfo, ok := c.controller.GetClusterInfo(c.clusterName)
	if !ok {
		return fmt.Errorf("cluster not found in cache")
	}
	createRestoreReq := &api.RestoreCreateRequest{
		CreateMetadata: &api.CreateMetadata{
			Name:  c.restoreName,
			OrgId: c.controller.currentOrgId,
			Uid:   c.restoreUid,
		},
		Backup:              c.backupName,
		Cluster:             clusterInfo.Name,
		NamespaceMapping:    c.namespaceMapping,
		StorageClassMapping: c.storageClassMapping,
		BackupRef: &api.ObjectRef{
			Name: backupInfo.Name,
			Uid:  backupInfo.Uid,
		},
		IncludeResources: c.resources,
	}
	if _, err := c.controller.processPxBackupRequest(createRestoreReq); err != nil {
		return err
	}
	restoreInspectReq := &api.RestoreInspectRequest{
		OrgId: c.controller.currentOrgId,
		Name:  c.restoreName,
	}
	resp, err := c.controller.processPxBackupRequest(restoreInspectReq)
	if err != nil {
		return err
	}
	restore := resp.(*api.RestoreInspectResponse).GetRestore()
	c.controller.setRestoreInfo(c.restoreName, &RestoreInfo{
		RestoreObject: restore,
	})
	return nil
}

func (p *PxbController) DeleteRestore(restoreName string) error {
	restoreInfo, ok := p.GetRestoreInfo(restoreName)
	if !ok {
		return fmt.Errorf("restore [%s] not found in cache", restoreName)
	}
	log.Infof("Deleting restore [%s] of org [%s]", restoreName, restoreInfo.OrgId)
	restoreDeleteReq := &api.RestoreDeleteRequest{
		Name:  restoreInfo.Name,
		OrgId: restoreInfo.OrgId,
	}
	if _, err := p.processPxBackupRequest(restoreDeleteReq); err != nil {
		return err
	}
	return nil
}

type restoreStatusInfo struct {
	Status api.RestoreInfo_StatusInfo_Status
	Reason string
	Error  error
}

func (p *PxbController) getRestoreStatus(restoreName string) (restoreStatusInfo, error) {
	restoreInfo, ok := p.GetRestoreInfo(restoreName)
	if !ok {
		return restoreStatusInfo{}, fmt.Errorf("restore [%s] not found in cache", restoreName)
	}
	restoreInspectRequest := &api.RestoreInspectRequest{
		OrgId: restoreInfo.OrgId,
		Name:  restoreName,
	}
	resp, err := p.processPxBackupRequest(restoreInspectRequest)
	if err != nil {
		return restoreStatusInfo{
			Status: api.RestoreInfo_StatusInfo_Invalid,
			Reason: "",
			Error:  err,
		}, nil
	}
	restore := resp.(*api.RestoreInspectResponse).GetRestore()
	status := restore.GetStatus().Status
	reason := restore.GetStatus().Reason
	return restoreStatusInfo{
		Status: status,
		Reason: reason,
		Error:  nil,
	}, nil
}

func (p *PxbController) WaitForRestoreCompletion(restoreName string) (api.RestoreInfo_StatusInfo_Status, error) {
	getRestoreStatus := func() interface{} {
		status, err := p.getRestoreStatus(restoreName)
		if err != nil {
			return restoreStatusInfo{
				Status: api.RestoreInfo_StatusInfo_Invalid,
				Error:  err,
			}
		}
		return status
	}
	shouldRetry := func(result interface{}) bool {
		res, ok := result.(restoreStatusInfo)
		if !ok || res.Error != nil {
			return false
		}
		finalStates := [...]api.RestoreInfo_StatusInfo_Status{
			api.RestoreInfo_StatusInfo_Invalid,
			api.RestoreInfo_StatusInfo_Aborted,
			api.RestoreInfo_StatusInfo_Failed,
			api.RestoreInfo_StatusInfo_Success,
			api.RestoreInfo_StatusInfo_Retained,
			api.RestoreInfo_StatusInfo_PartialSuccess,
		}
		log.Infof("restore status for [%s] is [%s]", restoreName, res.Status)
		for _, status := range finalStates {
			if res.Status == status {
				return false
			}
		}
		return true
	}
	res, err := utils.DoRetryWithTimeout(getRestoreStatus, utils.DefaultRestoreCreationTimeout, utils.DefaultRestoreCreationRetryTime, shouldRetry)
	if err != nil {
		return res.(restoreStatusInfo).Status, err
	}
	return res.(restoreStatusInfo).Status, nil
}

func (p *PxbController) WaitForRestoreDeletion(restoreName string) error {
	getRestoreStatus := func() interface{} {
		status, err := p.getRestoreStatus(restoreName)
		if err != nil {
			return restoreStatusInfo{
				Status: api.RestoreInfo_StatusInfo_Invalid,
				Error:  err,
			}
		}
		if status.Status != api.RestoreInfo_StatusInfo_Deleting {
			return restoreStatusInfo{
				Status: status.Status,
				Reason: status.Reason,
				Error:  fmt.Errorf("restore [%s] is not in [%s] state rather in [%s] state", restoreName, api.RestoreInfo_StatusInfo_Deleting, status.Status),
			}
		}
		return status
	}
	shouldRetry := func(result interface{}) bool {
		res := result.(restoreStatusInfo)
		if res.Error != nil {
			return false
		}
		if res.Status == api.RestoreInfo_StatusInfo_Invalid {
			return false
		}
		log.Infof("Restore status for [%s]: expected %s but got [%s] because [%s]", restoreName, api.RestoreInfo_StatusInfo_Invalid, res.Status, res.Reason)
		return true
	}
	res, err := utils.DoRetryWithTimeout(getRestoreStatus, utils.DefaultRestoreDeletionTimeout, utils.DefaultRestoreDeletionRetryTime, shouldRetry)
	if err != nil {
		return err
	}
	if res.(restoreStatusInfo).Status != api.RestoreInfo_StatusInfo_Invalid {
		return fmt.Errorf("expected restore status not reached because %v", res.(restoreStatusInfo).Error)
	}
	p.delRestoreInfo(restoreName)
	return nil
}
