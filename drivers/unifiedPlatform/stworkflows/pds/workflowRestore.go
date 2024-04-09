package pds

import (
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	pdslibs "github.com/portworx/torpedo/drivers/unifiedPlatform/pdsLibs"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/stworkflows/platform"
	"github.com/portworx/torpedo/pkg/log"
)

type WorkflowPDSRestore struct {
	WorkflowProject  platform.WorkflowProject
	Destination      platform.WorkflowNamespace
	SkipValidatation map[string]bool
	Restores         map[string]automationModels.PDSRestore
}

const (
	ValidatePdsRestore = "VALIDATE_PDS_RESTORE"
)

func (restore WorkflowPDSRestore) CreateRestore(name string, backupUid string, namespace string) (*automationModels.PDSRestoreResponse, error) {

	// TODO: Remove harcoded params added for testing the workflow
	createRestore, err := pdslibs.CreateRestore(
		name,
		backupUid, "clu:3868c014-b0a7-4f44-98b1-88fbdd5991c2",
		"nam:8fdb48fe-8ad2-4962-bc20-8980497abd41",
		"prj:2eee1079-83f0-4b88-b61d-d8782c2bdbd4",
		"prj:27ced002-f7a4-4d7d-a381-218dff8defe0",
	)

	if err != nil {
		return nil, err
	}

	if value, ok := restore.SkipValidatation[ValidatePdsRestore]; ok {
		if value == true {
			log.Infof("Skipping Restore Validation")
		}
	} else {
		err = pdslibs.ValidateRestoreDeployment(*createRestore.Create.Meta.Uid, namespace)
		if err != nil {
			return nil, err
		}
	}

	restore.Restores[name] = createRestore.Create
	log.Infof("Updated the restores")

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
