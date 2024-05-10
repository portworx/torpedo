package pds

import (
	"fmt"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	pdslibs "github.com/portworx/torpedo/drivers/unifiedPlatform/pdsLibs"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/stworkflows/platform"
	"github.com/portworx/torpedo/pkg/log"
	"time"
)

type WorkflowPDSRestore struct {
	Source              *platform.WorkflowNamespace
	Destination         *platform.WorkflowNamespace
	Validatation    map[string]bool
	Restores            map[string]automationModels.PDSRestore
	RestoredDeployments WorkflowDataService
	SourceDeploymentConfigBeforeUpgrade *automationModels.DeploymentTopology
}

const (
	ValidatePdsRestore                          = "VALIDATE_PDS_RESTORE"
	ValidateRestoreAfterSourceDeploymentUpgrade = "VALIDATE_RESTORE_AFTER_SRC_DEPLOYMENT_UPGRADE"
)

func (restore WorkflowPDSRestore) CreateRestore(name string, backupUid string, namespace string, sourceNamespace string) (*automationModels.PDSRestoreResponse, error) {

	log.Infof("Restore [%s] started at [%s]", name, time.Now().Format("2006-01-02 15:04:05"))

	log.Infof("Name of restore - [%s]", name)
	log.Infof("Backup UUID - [%s]", backupUid)
	log.Infof("Destination Cluster Id - [%s]", restore.Destination.TargetCluster.ClusterUID)
	log.Infof("Source project Id - [%s]", restore.Source.TargetCluster.Project.ProjectId)
	log.Infof("Destination project Id - [%s]", restore.Destination.TargetCluster.Project.ProjectId)
	err := restore.CreateAndAssociateRestoreNamespace(namespace, sourceNamespace)
	if err != nil {
		return nil, err
	}

	log.Infof("Destination Namespace Id - [%s]", restore.Destination.Namespaces[namespace])

	log.Infof("Creating restore - [%s]", name)
	createRestore, err := pdslibs.CreateRestore(
		name,
		backupUid, restore.Destination.TargetCluster.ClusterUID,
		restore.Destination.Namespaces[namespace],
		restore.Source.TargetCluster.Project.ProjectId,
		restore.Destination.TargetCluster.Project.ProjectId,
	)

	log.InfoD("Restore triggered. Name - [%s], UID - [%s]", *createRestore.Create.Meta.Name, *createRestore.Create.Meta.Uid)

	if err != nil {
		return nil, err
	}

	if value, ok := restore.Validatation[ValidatePdsRestore]; ok {
		if value == true {
			log.Infof("Skipping Restore Validation")
		}
	} else {
		log.Infof("Restore UID - [%s]", *createRestore.Create.Meta.Uid)
		if value, ok = restore.Validatation[ValidateRestoreAfterSourceDeploymentUpgrade]; ok {
			if value == true {
				log.Debugf("validating restore after source deployment upgrade")
				err = pdslibs.ValidateRestoreAfterSourceDeploymentUpgrade(*createRestore.Create.Meta.Uid, *restore.SourceDeploymentConfigBeforeUpgrade)
				if err != nil {
					return nil, err
				}
			}
		} else {
			log.Debugf("Starting Restore Validation")
			err = pdslibs.ValidateRestoreDeployment(*createRestore.Create.Meta.Uid, namespace)
			if err != nil {
				return nil, err
			}
		}
	}

	restore.Restores[name] = createRestore.Create
	deployment, _, err := pdslibs.GetDeployment(createRestore.Create.Config.DestinationReferences.DeploymentId)
	if err != nil {
		return nil, err
	}

	// TODO: The Get MD5Hash needs to be run here to get the Md5CheckSum
	restore.RestoredDeployments.DataServiceDeployment[createRestore.Create.Config.DestinationReferences.DeploymentId] = pdslibs.DataServiceDetails{
		Deployment:        deployment.Get,
		Namespace:         namespace,
		NamespaceId:       restore.Destination.Namespaces[namespace],
		SourceMd5Checksum: "",
	}

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

func (restore WorkflowPDSRestore) CreateAndAssociateRestoreNamespace(namespace string, sourceNamespace string) error {

	// TODO: Remove this once https://purestorage.atlassian.net/browse/DS-9443 is resolved
	log.InfoD("Creating restore namespace on source - [%s]", sourceNamespace)
	_, err := restore.Destination.CreateNamespaces(sourceNamespace)
	if err != nil {
		return fmt.Errorf("unable to create source namespace - [%s]", err.Error())
	}

	log.InfoD("Creating restore namespace - [%s]", namespace)
	_, err = restore.Destination.CreateNamespaces(namespace)
	if err != nil {
		return fmt.Errorf("unable to create restore namespace - [%s]", err.Error())
	}

	log.InfoD("Associating restore namespace to destination project")

	err = restore.Destination.TargetCluster.Project.Associate(
		[]string{},
		[]string{restore.Destination.Namespaces[namespace], restore.Destination.Namespaces[sourceNamespace]},
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
