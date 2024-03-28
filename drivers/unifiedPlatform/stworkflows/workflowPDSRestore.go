package stworkflows

import (
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	pdslibs "github.com/portworx/torpedo/drivers/unifiedPlatform/pdsLibs"
	"github.com/portworx/torpedo/pkg/log"
)

type WorkflowPDSRestore struct {
	WorkflowDataService    WorkflowDataService
	Destination            WorkflowTargetCluster
	WorkflowBackupLocation WorkflowBackupLocation
	SkipValidatation       map[string]bool
}

const (
	ValidatePdsRestore = "VALIDATE_PDS_RESTORE"
)

func (restore WorkflowPDSRestore) CreateRestore(backupUid string, deploymentName string, cloudSnapId string) (*automationModels.PDSRestoreResponse, error) {
	createRestore, err := pdslibs.CreateRestore(
		backupUid, restore.Destination.ClusterUID,
		restore.WorkflowDataService.DataServiceDeployment[deploymentName],
		restore.Destination.Project.ProjectId,
		cloudSnapId, // This needs to be replaced with clodSnapID
		restore.WorkflowBackupLocation.BkpLocation.BkpLocationId)

	if err != nil {
		return nil, err
	}

	if value, ok := restore.SkipValidatation[ValidatePdsRestore]; ok {
		if value == true {
			log.Infof("Skipping Restore Validation")
		}
	} else {
		err = pdslibs.ValidateRestoreDeployment(*createRestore.Create.Meta.Uid, restore.WorkflowDataService.NamespaceName)
		if err != nil {
			return nil, err
		}
	}

	return createRestore, nil
}

func (restore WorkflowPDSRestore) GetRestore(id string) (*automationModels.PDSRestoreResponse, error) {
	getRestore, err := pdslibs.GetRestore(id)

	if err != nil {
		return nil, err
	}

	return getRestore, nil
}

func (restore WorkflowPDSRestore) DeleteRestore(id string) error {
	err := pdslibs.DeleteRestore(id)

	if err != nil {
		return err
	}

	return nil
}
