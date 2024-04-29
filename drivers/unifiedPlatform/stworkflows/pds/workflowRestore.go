package pds

import (
	"fmt"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	pdslibs "github.com/portworx/torpedo/drivers/unifiedPlatform/pdsLibs"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/stworkflows/platform"
	"github.com/portworx/torpedo/pkg/log"
)

type WorkflowPDSRestore struct {
	SourceNamespace     string
	Source              *platform.WorkflowNamespace
	Destination         *platform.WorkflowNamespace
	SkipValidatation    map[string]bool
	Restores            map[string]automationModels.PDSRestore
	RestoredDeployments WorkflowDataService
}

const (
	ValidatePdsRestore = "VALIDATE_PDS_RESTORE"
)

func (restore WorkflowPDSRestore) CreateRestore(name string, backupUid string, namespace string) (*automationModels.PDSRestoreResponse, error) {

	log.Infof("Name of restore - [%s]", name)
	log.Infof("Backup UUID - [%s]", backupUid)
	log.Infof("Destination Cluster Id - [%s]", restore.Destination.TargetCluster.ClusterUID)
	log.Infof("Source project Id - [%s]", restore.Source.TargetCluster.Project.ProjectId)
	log.Infof("Destination project Id - [%s]", restore.Destination.TargetCluster.Project.ProjectId)
	err := restore.CreateAndAssociateRestoreNamespace(namespace)
	if err != nil {
		return nil, err
	}

	log.Infof("Destination Namespace Id - [%s]", restore.Destination.Namespaces[namespace])

	createRestore, err := pdslibs.CreateRestore(
		name,
		backupUid, restore.Destination.TargetCluster.ClusterUID,
		restore.Destination.Namespaces[namespace],
		restore.Source.TargetCluster.Project.ProjectId,
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

	restore.RestoredDeployments.Namespace = restore.Destination
	restore.RestoredDeployments.DataServiceDeployment[name] = createRestore.Create.Config.DestinationReferences.DeploymentId

	log.Infof("Restore completed successfully - [%s]", *createRestore.Create.Meta.Name)

	return createRestore, nil
}

// Get Restore fetches the first given restore id
func (restore WorkflowPDSRestore) GetRestore(id string) (*automationModels.PDSRestoreResponse, error) {
	getRestore, err := pdslibs.GetRestore(id)

	if err != nil {
		return nil, err
	}

	return getRestore, nil
}

// Purge Deletes all created restores
func (restore WorkflowPDSRestore) Purge() error {

	err := restore.RestoredDeployments.Purge()
	if err != nil {
		return err
	}

	return nil
}

func (restore WorkflowPDSRestore) CreateAndAssociateRestoreNamespace(namespace string) error {

	// TODO: Remove this once https://purestorage.atlassian.net/browse/DS-9443 is resolved
	log.InfoD("Creating restore namespace on source")
	_, err := restore.Destination.CreateNamespaces(restore.SourceNamespace)
	if err != nil {
		return fmt.Errorf("unable to create source namespace - [%s]", err.Error())
	}

	log.InfoD("Creating restore namespace")
	_, err = restore.Destination.CreateNamespaces(namespace)
	if err != nil {
		return fmt.Errorf("unable to create restore namespace - [%s]", err.Error())
	}

	log.InfoD("Associating restore namespace to destination project")

	err = restore.Destination.TargetCluster.Project.Associate(
		[]string{},
		[]string{restore.Destination.Namespaces[namespace], restore.Destination.Namespaces[restore.SourceNamespace]},
		[]string{},
		[]string{},
		[]string{},
		[]string{},
	)
	if err != nil {
		return fmt.Errorf("unable to associate restore namespace - [%s]", err.Error())
	}

	return nil
}
