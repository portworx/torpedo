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
	SchedulePolicyDataStore  *generics.DataStore[*api.SchedulePolicyObject]
	BackupScheduleDataStore  *generics.DataStore[*api.BackupScheduleObject]
	ClusterDataStore         *generics.DataStore[*api.ClusterObject]
	CloudCredentialDataStore *generics.DataStore[*api.CloudCredentialObject]
	BackupLocationDataStore  *generics.DataStore[*api.BackupLocationObject]
	BackupDataStore          *generics.DataStore[*api.BackupObject]
	RestoreDataStore         *generics.DataStore[*api.RestoreObject]
	RuleDataStore            *generics.DataStore[*api.RuleObject]
	RoleDataStore            *generics.DataStore[*api.RoleObject]
}

type PxBackup struct {
	UserDataStore *generics.DataStore[*User]
}

func (b *PxBackup) SelectUser(username string) *User {
	return b.UserDataStore.Get(username)
}

func (b *PxBackup) AddUser(ctx context.Context, userRepresentation *auth.UserRepresentation) error {
	addUserReq := &auth.AddUserRequest{
		UserRepresentation: userRepresentation,
	}
	_, err := auth.AddUser(ctx, addUserReq)
	if err != nil {
		return ProcessError(err)
	}
	b.UserDataStore.Set(
		userRepresentation.Username,
		&User{
			Spec:                  userRepresentation,
			PxBackup:              b,
			OrganizationDataStore: generics.NewDataStore[*Organization](),
		},
	)
	return nil
}

func (b *PxBackup) AddTestUser(ctx context.Context, username string, password string) error {
	user := &auth.UserRepresentation{
		ID:            "",
		Username:      username,
		FirstName:     "first-" + username,
		LastName:      username + "last",
		Email:         username + "@cnbu.com",
		EmailVerified: true,
		Enabled:       true,
		Credentials: []auth.CredentialRepresentation{
			{
				Type:      auth.Password.String(),
				Value:     password,
				Temporary: false,
			},
		},
	}
	err := b.AddUser(ctx, user)
	if err != nil {
		return ProcessError(err, ToString(user))
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
