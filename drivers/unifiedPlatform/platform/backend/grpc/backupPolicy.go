package grpc

import (
	"context"
	"fmt"
	"github.com/jinzhu/copier"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	"github.com/portworx/torpedo/pkg/log"
	publicbackuppolicyapis "github.com/pure-px/apis/public/portworx/platform/backuppolicy/apiv1"
	"google.golang.org/grpc"
)

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

func (BackupPolicyV1 *PlatformGrpc) CreateBackupPolicy(createRequest *WorkFlowRequest) (WorkFlowResponse, error) {
	backupPolicyResponse := WorkFlowResponse{}
	ctx, client, _, err := BackupPolicyV1.getBackupPolicyClient()
	if err != nil {
		return backupPolicyResponse, fmt.Errorf("Error while getting updated client with auth header: %v\n", err)
	}

	createBackupPolicyRequest := publicbackuppolicyapis.CreateBackupPolicyRequest{}
	err = copier.Copy(&createBackupPolicyRequest, createRequest)
	apiResponse, err := client.CreateBackupPolicy(ctx, &createBackupPolicyRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return backupPolicyResponse, fmt.Errorf("Error calling createBackupPolicy: %v\n", err)
	}

	err = copier.Copy(&backupPolicyResponse, apiResponse)
	if err != nil {
		return backupPolicyResponse, fmt.Errorf("Error while copying: %v\n", err)
	}
	return backupPolicyResponse, nil
}

func (BackupPolicyV1 *PlatformGrpc) ListBackupPolicies(tenantId string) ([]WorkFlowResponse, error) {
	backupPolicyResponse := []WorkFlowResponse{}
	ctx, client, _, err := BackupPolicyV1.getBackupPolicyClient()
	if err != nil {
		return backupPolicyResponse, fmt.Errorf("Error while getting updated client with auth header: %v\n", err)
	}

	listBackupPolicyRequest := publicbackuppolicyapis.ListBackupPoliciesRequest{
		TenantId: tenantId,
	}

	apiResponse, err := client.ListBackupPolicies(ctx, &listBackupPolicyRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return backupPolicyResponse, fmt.Errorf("Error calling ListBackupPolicies: %v\n", err)
	}

	err = copier.Copy(&backupPolicyResponse, apiResponse)
	if err != nil {
		return backupPolicyResponse, fmt.Errorf("Error while copying: %v\n", err)
	}
	return backupPolicyResponse, nil
}

func (BackupPolicyV1 *PlatformGrpc) GetBackupPolicy(backupPolicyId string) (WorkFlowResponse, error) {
	backupPolicyResponse := WorkFlowResponse{}
	ctx, client, _, err := BackupPolicyV1.getBackupPolicyClient()

	if err != nil {
		return backupPolicyResponse, fmt.Errorf("Error while getting updated client with auth header: %v\n", err)
	}

	getBackupPolicyRequest := publicbackuppolicyapis.GetBackupPolicyRequest{
		Id: backupPolicyId,
	}

	apiResponse, err := client.GetBackupPolicy(ctx, &getBackupPolicyRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return backupPolicyResponse, fmt.Errorf("Error calling GetBackupPolicy: %v\n", err)
	}

	err = copier.Copy(&backupPolicyResponse, apiResponse)
	if err != nil {
		return backupPolicyResponse, fmt.Errorf("Error while copying: %v\n", err)
	}
	return backupPolicyResponse, nil
}

func (BackupPolicyV1 *PlatformGrpc) DeleteBackupPolicy(backupPolicyId string) error {
	ctx, client, _, err := BackupPolicyV1.getBackupPolicyClient()

	if err != nil {
		return fmt.Errorf("Error while getting updated client with auth header: %v\n", err)
	}

	deleteBackupPolicyRequest := publicbackuppolicyapis.DeleteBackupPolicyRequest{
		Id: backupPolicyId,
	}
	_, err = client.DeleteBackupPolicy(ctx, &deleteBackupPolicyRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return fmt.Errorf("Error calling DeleteBackupPolicy: %v\n", err)
	}
	return nil
}
