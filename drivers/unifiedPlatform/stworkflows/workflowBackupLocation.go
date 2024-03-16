package stworkflows

import (
	"github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/platformLibs"
	"github.com/portworx/torpedo/pkg/log"
)

type WorkflowBackupLocation struct {
	WfCloudCredentials WorkflowCloudCredentials
	BkpLocation        BkpLocationType
}

type BkpLocationType struct {
	BkpLocationId string
	Name          string
}

func (bkpLoc *WorkflowBackupLocation) CreateBackupLocation(bucketName, backUpTargetType string) (*WorkflowBackupLocation, error) {
	var bkpLocation *apiStructs.BackupLocation
	var err error

	tenantId := bkpLoc.WfCloudCredentials.Platform.TenantId
	cloudCreds := bkpLoc.WfCloudCredentials.CloudCredentials

	log.Infof("tenant id in wfBkpLocation [%s]", tenantId)

	for _, cloudCred := range cloudCreds {
		log.Infof("cloud cred id in wfBkpLocation [%s]", cloudCred.ID)
		bkpLocation, err = platformLibs.CreateBackupLocation(tenantId, cloudCred.ID, bucketName, backUpTargetType)
		if err != nil {
			return nil, err
		}
	}

	bkpLoc.BkpLocation = BkpLocationType{
		BkpLocationId: *bkpLocation.Meta.Uid,
		Name:          *bkpLocation.Meta.Name,
	}

	return bkpLoc, nil
}

func (bkpLoc *WorkflowBackupLocation) ListBackupLocation() ([]*WorkflowBackupLocation, error) {
	bkpLocResponses := []*WorkflowBackupLocation{}
	bkplocs, err := platformLibs.ListBackupLocation(bkpLoc.WfCloudCredentials.Platform.TenantId)
	if err != nil {
		return nil, err
	}

	for _, bkpLocation := range bkplocs {
		newBackupLocation := &WorkflowBackupLocation{
			WfCloudCredentials: WorkflowCloudCredentials{
				Platform:         bkpLoc.WfCloudCredentials.Platform,
				CloudCredentials: []CloudCredentialsType{{ID: bkpLocation.Config.CloudCredentialsId}},
			},
			BkpLocation: BkpLocationType{
				BkpLocationId: *bkpLocation.Meta.Uid,
				Name:          *bkpLocation.Meta.Name,
			},
		}
		bkpLocResponses = append(bkpLocResponses, newBackupLocation)
	}
	return bkpLocResponses, nil
}
