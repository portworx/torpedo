package grpc

import (
	"context"
	"fmt"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	"github.com/portworx/torpedo/pkg/log"
	publicBackupConfigapis "github.com/pure-px/apis/public/portworx/pds/restore/apiv1"
)

func (backupConf *PdsGrpc) getBackupConfigClient() (context.Context, publicBackupConfigapis.BackupConfigServiceClient, string, error) {
	log.Infof("Creating client from grpc package")
	var depClient publicBackupConfigapis.BackupConfigServiceClient

	ctx, token, err := GetBearerToken()
	if err != nil {
		return nil, nil, "", fmt.Errorf("Error in getting bearer token: %v\n", err)
	}
	credentials = &Credentials{
		Token: token,
	}
	depClient = publicBackupConfigapis.NewBackupConfigServiceClient(backupConf.ApiClientV2)
	return ctx, depClient, token, nil
}
