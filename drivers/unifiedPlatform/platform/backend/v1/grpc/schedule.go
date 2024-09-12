package grpc

import (
	"context"
	"fmt"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	"github.com/portworx/torpedo/pkg/log"
	publicbackuppolicyapis "github.com/pure-px/apis/public/portworx/platform/backuppolicy/apiv1"
)

//TODO: This whole file needs to be revisited while implementing the actual methods for GRPC calls

// GetClient updates the header with bearer token and returns the new client
func (BackupPolicyV1 *PlatformGrpc) getBackupPolicyClient() (context.Context, publicbackuppolicyapis.BackupPolicyServiceClient, string, error) {
	log.Infof("Creating client from grpc package")
	var backupPolicyClient publicbackuppolicyapis.BackupPolicyServiceClient
	ctx, token, err := GetBearerToken()
	if err != nil {
		return nil, nil, "", fmt.Errorf("Error in getting bearer token: %v\n", err)
	}

	credentials = &Credentials{
		Token: token,
	}

	backupPolicyClient = publicbackuppolicyapis.NewBackupPolicyServiceClient(BackupPolicyV1.ApiClientV1)

	return ctx, backupPolicyClient, token, nil
}

// GetBackupPolicy gets the backup policy model by its ID.
func (backupPolicy *PlatformGrpc) GetBackupPolicy(getRequest *BackupPolicyRequest) (*BackupPolicyResponse, error) {
	return &BackupPolicyResponse{}, nil
}

// ListBackupPolicy lists all backup policies.
func (backupPolicy *PlatformGrpc) ListBackupPolicy(listRequest *BackupPolicyRequest) (*BackupPolicyResponse, error) {
	return &BackupPolicyResponse{}, nil
}

// CreateBackupPolicy creates a backup policy.
func (backupPolicy *PlatformGrpc) CreateBackupPolicy(createRequest *BackupPolicyRequest) (*BackupPolicyResponse, error) {
	return &BackupPolicyResponse{}, nil
}

// DeleteBackupPolicy deletes a backup policy.
func (backupPolicy *PlatformGrpc) DeleteBackupPolicy(deleteRequest *BackupPolicyRequest) error {

	return nil
}
