package pxb

import (
	"context"
	api "github.com/portworx/px-backup-api/pkg/apis/v1"
	"github.com/portworx/torpedo/drivers/pxb/auth"
	"github.com/portworx/torpedo/drivers/pxb/generics"
	. "github.com/portworx/torpedo/drivers/pxb/pxbutils"
)

type User struct {
	Spec                  *auth.UserRepresentation
	PxBackup              *PxBackup
	OrganizationDataStore *generics.DataStore[*Organization]
}

type Organization struct {
	Spec                     *api.OrganizationObject
	BackupDataStore          *generics.DataStore[*api.BackupObject]
	BackupLocationDataStore  *generics.DataStore[*api.BackupLocationObject]
	BackupScheduleDataStore  *generics.DataStore[*api.BackupScheduleObject]
	ClusterDataStore         *generics.DataStore[*api.ClusterObject]
	CloudCredentialDataStore *generics.DataStore[*api.CloudCredentialObject]
	RoleDataStore            *generics.DataStore[*api.RoleObject]
	RestoreDataStore         *generics.DataStore[*api.RestoreObject]
	RuleDataStore            *generics.DataStore[*api.RuleObject]
	SchedulePolicyDataStore  *generics.DataStore[*api.SchedulePolicyObject]
}

type PxBackup struct {
	UserDataStore *generics.DataStore[*User]
}

func (b *PxBackup) SelectUser(username string) *User {
	return b.UserDataStore.Get(username)
}

func (b *PxBackup) AddTestUser(ctx context.Context, username string, password string) error {
	addUserReq := &auth.AddUserRequest{
		UserRepresentation: auth.NewTestUserRepresentation(username, password),
	}
	_, err := auth.AddUser(ctx, addUserReq)
	if err != nil {
		return ProcessError(err)
	}
	return nil
}

func (b *PxBackup) DeleteUser(ctx context.Context, username string) error {
	deleteUserReq := &auth.DeleteUserRequest{
		Username: username,
	}
	_, err := auth.DeleteUser(ctx, deleteUserReq)
	if err != nil {
		return ProcessError(err)
	}
	return nil
}
