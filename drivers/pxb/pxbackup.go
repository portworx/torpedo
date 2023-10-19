package pxb

import (
	"context"
	"fmt"
	api "github.com/portworx/px-backup-api/pkg/apis/v1"
	"github.com/portworx/torpedo/drivers/pxb/auth"
	"github.com/portworx/torpedo/drivers/pxb/generics"
	"github.com/portworx/torpedo/drivers/pxb/pxbutils"
	"github.com/portworx/torpedo/pkg/log"
	"net/http"
	"time"
)

type User struct {
	Spec                  *auth.UserRepresentation
	PxBackup              *PxBackup
	OrganizationDataStore *generics.DataStore[*Organization]
}

//func (u *User) Delete() error {
//	if u == nil {
//		err := fmt.Errorf("user is nil")
//		return ProcessError(err)
//	}
//	deleteUserReq := &auth.DeleteUserRequest{
//		Username: u.Spec.Username,
//	}
//	_, err := auth.DeleteUser(context.Background(), deleteUserReq)
//	if err != nil {
//		return ProcessError(err)
//	}
//	u.PxBackup.UserDataStore.Remove(u.Spec.Username)
//	return nil
//}

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

func (b *PxBackup) AddTestUser(username string, password string) error {
	addUserReq := &auth.AddUserRequest{
		UserRepresentation: auth.NewTestUserRepresentation(username, password),
	}
	pxbNamespace, err := pxbutils.GetPxBackupNamespace()
	if err != nil {
		log.FailOnError(err, "failed to get px-backup namespace")
	}
	keycloak := &auth.Keycloak{
		Client:    &http.Client{Timeout: 1 * time.Minute},
		Namespace: pxbNamespace,
	}
	_, err = keycloak.AddUser(context.Background(), addUserReq)
	if err != nil {
		return pxbutils.ProcessError(err)
	}
	user := &User{
		Spec:                  addUserReq.UserRepresentation,
		PxBackup:              b,
		OrganizationDataStore: generics.NewDataStore[*Organization](),
	}
	b.UserDataStore.Set(username, user)
	return nil
}

func (b *PxBackup) SelectUser(username string) *User {
	user := b.UserDataStore.Get(username)
	if user == nil {
		err := fmt.Errorf("user with username [%s] not found", username)
		log.Errorf(pxbutils.ProcessError(err).Error())
	}
	return user
}
