package pds

import (
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	pdslibs "github.com/portworx/torpedo/drivers/unifiedPlatform/pdsLibs"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/stworkflows/platform"
	"github.com/portworx/torpedo/pkg/log"
)

type WorkflowPDSRestore struct {
	WorkflowProject     platform.WorkflowProject
	Destination         platform.WorkflowNamespace
	SkipValidatation    map[string]bool
	Restores            map[string]automationModels.PDSRestore
	RestoredDeployments map[string]WorkflowDataService
}

const (
	ValidatePdsRestore = "VALIDATE_PDS_RESTORE"
)

func (restore WorkflowPDSRestore) CreateRestore(name string, backupUid string, namespace string) (*automationModels.PDSRestoreResponse, error) {

	log.Infof("Name of restore - [%s]", name)
	log.Infof("Backup UUID - [%s]", backupUid)
	log.Infof("Destination Cluster Id - [%s]", restore.Destination.TargetCluster.ClusterUID)
	log.Infof("Destination Namespace Id - [%s]", restore.Destination.Namespaces[namespace])
	log.Infof("Source project Id - [%s]", restore.WorkflowProject.ProjectId)
	log.Infof("Destination project Id - [%s]", restore.Destination.TargetCluster.Project.ProjectId)

	createRestore, err := pdslibs.CreateRestore(
		name,
		backupUid, restore.Destination.TargetCluster.ClusterUID,
		restore.Destination.Namespaces[namespace],
		restore.WorkflowProject.ProjectId,
		restore.Destination.TargetCluster.Project.ProjectId,
	)

	if err != nil {
		return nil, err
	}

	if value, ok := restore.SkipValidatation[ValidatePdsRestore]; ok {
		if value == true {
			log.Infof("Skipping Restore Validation")
		}
	} else {
		log.Infof("Restore UID - [%s]", *createRestore.Create.Meta.Uid)
		err = pdslibs.ValidateRestoreDeployment(*createRestore.Create.Meta.Uid, namespace)
		if err != nil {
			return nil, err
		}
	}

	restore.Restores[name] = createRestore.Create
	log.Infof("Restore completed successfully - [%s]", *createRestore.Create.Meta.Name)

	return createRestore, nil
}

func (restore WorkflowPDSRestore) GetRestore(id string) (*automationModels.PDSRestoreResponse, error) {
	getRestore, err := pdslibs.GetRestore(id)

	if err != nil {
		return nil, err
	}

	return getRestore, nil
}

//func (restore WorkflowPDSRestore) DeleteRestore(id string) error {
//	err := pdslibs.DeleteRestore(id)
//
//	if err != nil {
//		return err
//	}
//
//	return nil
//}
