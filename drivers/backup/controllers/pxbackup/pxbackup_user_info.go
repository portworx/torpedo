package pxbackup

import (
	"fmt"
	"github.com/portworx/torpedo/drivers/backup"
	"github.com/portworx/torpedo/drivers/backup/utils"
)

type UserInfo struct {
	username   string
	firstName  string
	lastName   string
	email      string
	isExisting bool
	isAdmin    bool
	password   *string
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

func (u *UserInfo) IsExisting() *UserInfo {
	u.isExisting = true
	return u
}

func (u *UserInfo) IsAdmin() *UserInfo {
	u.isAdmin, u.isExisting = true, true
	return u
}

func (u *UserInfo) DeepCopy() *UserInfo {
	newUserInfo := UserInfo{
		username:   u.username,
		firstName:  u.firstName,
		lastName:   u.lastName,
		email:      u.email,
		isExisting: u.isExisting,
		isAdmin:    u.isAdmin,
	}
	if u.password != nil {
		passwordCopy := *u.password
		newUserInfo.password = &passwordCopy
	}
	return &newUserInfo
}

func (u *UserInfo) register() error {
	if u.password == nil {
		return fmt.Errorf("nil password")
	}
	err := backup.AddUser(u.username, u.firstName, u.lastName, u.email, *u.password)
	if err != nil {
		debugMessage := fmt.Sprintf("username [%s]; first-name [%s]; last-name [%s]; email [%s]", u.username, u.firstName, u.lastName, u.email)
		return utils.ProcessError(err, debugMessage)
	}
	return nil
}

func (u *UserInfo) delete() error {
	err := backup.DeleteUser(u.username)
	if err != nil {
		debugMessage := fmt.Sprintf("username [%s]", u.username)
		return utils.ProcessError(err, debugMessage)
	}
	return nil
}

func (u *UserInfo) GetController() (*PxBackupController, error) {
	if !u.isExisting {
		err := u.register()
		if err != nil {
			return nil, utils.ProcessError(err)
		}
	}
	pxBackupController := &PxBackupController{
		UserInfo:      u.DeepCopy(),
		currentOrgId:  DefaultPxBackupOrganizationId,
		organizations: make(map[string]*OrganizationObjects, 0),
	}
	pxBackupController.organizations[DefaultPxBackupOrganizationId] = &OrganizationObjects{}
	return pxBackupController, nil
}
