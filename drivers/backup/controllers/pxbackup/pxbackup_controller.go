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
}

type PxBackupController struct {
	*UserInfo
	currentOrgId  string
	organizations map[string]*OrganizationObjects
}

func User(username string, password *string) *UserInfo {
	return &UserInfo{
		username:  username,
		password:  password,
		firstName: "first-" + username,
		lastName:  username + "-last",
		email:     username + "@cnbu.com",
	}
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
