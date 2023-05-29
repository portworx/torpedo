package pxbackup

import (
	"github.com/pborman/uuid"
	api "github.com/portworx/px-backup-api/pkg/apis/v1"
	"github.com/portworx/torpedo/drivers/backup/utils"
)

const (
	GlobalAdminUsername = "admin"
)

const (
	GlobalMinCloudAccountNameLength = 3
	DefaultPxBackupOrganizationId   = "default"
)

type CloudAccountInfo struct {
	*api.CloudCredentialObject
	provider string
}

type OrganizationObjects struct {
	cloudAccounts map[string]*CloudAccountInfo
	//backupLocations  map[string]*BackupLocationInfo
	//clusters         map[string]*ClusterInfo
	//rules            map[string]*RuleInfo
	//backups          map[string]*BackupInfo
	//restores         map[string]*RestoreInfo
	//schedulePolicies map[string]*SchedulePolicyInfo
}

type PxBackupController struct {
	*UserInfo
	currentOrgId  string
	organizations map[string]*OrganizationObjects
}

func (p *PxBackupController) getCloudAccountInfo(cloudAccountName string) *CloudAccountInfo {
	cloudAccountInfo, ok := p.organizations[p.currentOrgId].cloudAccounts[cloudAccountName]
	if !ok {
		return &CloudAccountInfo{}
	}
	return cloudAccountInfo
}

func (p *PxBackupController) saveCloudAccountInfo(cloudAccountName string, cloudAccountInfo *CloudAccountInfo) {
	if p.organizations[p.currentOrgId].cloudAccounts == nil {
		p.organizations[p.currentOrgId].cloudAccounts = make(map[string]*CloudAccountInfo, 0)
	}
	p.organizations[p.currentOrgId].cloudAccounts[cloudAccountName] = cloudAccountInfo
}

func (p *PxBackupController) delCloudAccountInfo(cloudAccountName string) {
	delete(p.organizations[p.currentOrgId].cloudAccounts, cloudAccountName)
}

func (p *PxBackupController) isCloudAccountNameRecorded(cloudAccountName string) bool {
	_, ok := p.organizations[p.currentOrgId].cloudAccounts[cloudAccountName]
	return ok
}

func (p *PxBackupController) CloudAccount(cloudAccountName string) *CloudAccountConfig {
	if !p.isCloudAccountNameRecorded(cloudAccountName) {
		return &CloudAccountConfig{
			cloudAccountName: cloudAccountName,
			cloudAccountUid:  uuid.New(),
			controller:       p,
		}
	}
	cloudAccountInfo := p.getCloudAccountInfo(cloudAccountName)
	return &CloudAccountConfig{
		cloudAccountName: cloudAccountName,
		cloudAccountUid:  cloudAccountInfo.GetUid(),
		isRecorded:       true,
		controller:       p,
	}
}

func AddPxBackupControllersToMap(pxBackupControllerMap *map[string]*PxBackupController, usersInfo []*UserInfo) error {
	if *pxBackupControllerMap == nil {
		*pxBackupControllerMap = make(map[string]*PxBackupController, 0)
	}
	adminUserInfo := User(GlobalAdminUsername, nil).IsAdmin()
	usersInfo = append(usersInfo, adminUserInfo)
	for _, userInfo := range usersInfo {
		clusterController, err := userInfo.GetController()
		if err != nil {
			return utils.ProcessError(err)
		}
		(*pxBackupControllerMap)[userInfo.username] = clusterController
	}
	return nil
}
