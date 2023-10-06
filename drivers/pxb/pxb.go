package pxb

import (
	"context"
	"fmt"
	api "github.com/portworx/px-backup-api/pkg/apis/v1"
	"github.com/portworx/torpedo/drivers/pxb/auth"
	"github.com/portworx/torpedo/drivers/pxb/generics"
	"github.com/portworx/torpedo/drivers/pxb/pxbutils"
	"github.com/portworx/torpedo/pkg/log"
	"time"
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

func (b *PxBackup) AddTestUser() error {
	username := fmt.Sprintf("pxb-user-%v", time.Now().Unix())
	log.Infof("Creating user %s", username)
	//err := auth.AddUser(context.Background(), user, "firstName", "lastName", "fl@cnbu.com", true, "admin", true)
	//log.Errorf("Creating user caused error: %v", err)
	//log.Infof("Created user %s", user)
	//return err

	ctx := context.Background()

	// Test AddUser
	log.Infof("Testing AddUser...")

	addUserReq := &auth.AddUserRequest{
		User: &auth.UserRepresentation{
			Username:      username,
			FirstName:     "Test",
			LastName:      "User",
			EmailVerified: true,
			Enabled:       true,
			Email:         "testuser@example.com",
			Credentials: []auth.CredentialRepresentation{
				{
					Type:      "password",
					Value:     "admin",
					Temporary: false,
				},
			},
		},
	}

	addUserResp, err := auth.AddUser(ctx, addUserReq)
	if err != nil {
		log.Infof("Failed to add user: %v", err)
	} else {
		log.Infof("Successfully added user: %v", addUserResp)
		log.Infof("Successfully added user: %v -- string version %s", addUserResp, pxbutils.ToString(addUserReq))
	}

	//// Test DeleteUser
	//fmt.Println("Testing DeleteUser...")
	//
	//deleteUserReq := &auth.DeleteUserRequest{
	//	Username: username,
	//}
	//
	//deleteUserResp, err := auth.DeleteUser(ctx, deleteUserReq)
	//if err != nil {
	//	fmt.Println("Failed to delete user:", err)
	//} else {
	//	fmt.Println("Successfully deleted user:", deleteUserResp)
	//}
	return err
}
