package pds

import (
	"fmt"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	pdslibs "github.com/portworx/torpedo/drivers/unifiedPlatform/pdsLibs"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/stworkflows"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/stworkflows/platform"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	"github.com/portworx/torpedo/pkg/log"
	"slices"
	"time"
)

type WorkflowPDSRestore struct {
	Source                              *WorkflowDataService
	Destination                         *platform.WorkflowNamespace
	Validatation                        map[string]bool
	Restores                            map[string]automationModels.PDSRestore
	RestoredDeployments                 *WorkflowDataService
	SourceDeploymentConfigBeforeUpgrade *automationModels.DeploymentTopology
	SkipValidation                      map[string]bool
	WorkflowBackup                      *WorkflowPDSBackup
}

const (
	ValidatePdsRestore                          = "VALIDATE_PDS_RESTORE"
	ValidateRestoreAfterSourceDeploymentUpgrade = "VALIDATE_RESTORE_AFTER_SRC_DEPLOYMENT_UPGRADE"
	CheckDataAfterRestore                       = "CHECKDATAAFTERRESTORE"
)

func (restore WorkflowPDSRestore) CreateRestore(name string, backupUid string, namespace string, sourceDeploymentId string) (*automationModels.PDSRestoreResponse, error) {

	log.Infof("Restore [%s] started at [%s]", name, time.Now().Format("2006-01-02 15:04:05"))

	log.Infof("Name of restore - [%s]", name)
	log.Infof("Backup UUID - [%s]", backupUid)
	log.Infof("Destination Cluster Id - [%s]", restore.Destination.TargetCluster.ClusterUID)
	log.Infof("Source project Id - [%s]", restore.Source.Namespace.TargetCluster.Project.ProjectId)
	log.Infof("Destination project Id - [%s]", restore.Destination.TargetCluster.Project.ProjectId)
	err := restore.CreateAndAssociateRestoreNamespace(namespace, restore.Source.DataServiceDeployment[sourceDeploymentId].Namespace)
	if err != nil {
		return nil, err
	}

	log.Infof("Destination Namespace Id - [%s]", restore.Destination.Namespaces[namespace])

	log.Infof("Creating restore - [%s]", name)
	createRestore, err := pdslibs.CreateRestore(
		name,
		backupUid,
		restore.Destination.TargetCluster.Project.ProjectId,
	)

	if err != nil {
		return nil, err
	}

	log.InfoD("Restore triggered. Name - [%s], UID - [%s]", *createRestore.Create.Meta.Name, *createRestore.Create.Meta.Uid)

	if value, ok := restore.SkipValidation[ValidatePdsRestore]; ok {
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
	restore.RestoredDeployments.DataServiceDeployment[createRestore.Create.Config.DestinationReferences.DeploymentId] = &pdslibs.DataServiceDetails{
		Deployment:        deployment.Get,
		Namespace:         namespace,
		NamespaceId:       restore.Destination.Namespaces[namespace],
		SourceMd5Checksum: "",
		DSParams:          restore.Source.DataServiceDeployment[sourceDeploymentId].DSParams,
	}

	log.Infof("Validating data after restore")
	if value, ok := restore.SkipValidation[CheckDataAfterRestore]; ok ||
		slices.Contains(stworkflows.SKIPDATASERVICEFROMWORKLOAD, restore.RestoredDeployments.DataServiceDeployment[createRestore.Create.Config.DestinationReferences.DeploymentId].DSParams.Name) {
		if value == true {
			log.Infof("Skipping data validation for the restore [%s]", name)
		}
	} else {
		err := restore.ValidateDataAfterRestore(createRestore.Create.Config.DestinationReferences.DeploymentId, backupUid)
		if err != nil {
			return nil, fmt.Errorf("data validation failed. Error - [%s]", err.Error())
		}
	}

	return createRestore, nil
}

func (restore WorkflowPDSRestore) ValidateDataAfterRestore(destinationDeploymentId string, backupId string) error {

	err := restore.RestoredDeployments.ReadAndUpdateDataServiceDataHash(destinationDeploymentId)
	if err != nil {
		return fmt.Errorf("unable to read data from restored database. Error - [%s]", err.Error())
	}

	sourceCheckSum := restore.WorkflowBackup.Backups[backupId].Md5Hash
	destinationCheckSum := restore.RestoredDeployments.DataServiceDeployment[destinationDeploymentId].SourceMd5Checksum

	log.Infof("Source Md5 Hash - [%s]", sourceCheckSum)
	log.Infof("Restore Md5 Hash - [%s]", destinationCheckSum)

	if sourceCheckSum != destinationCheckSum {
		return fmt.Errorf("Data validation failed for restore. Expected - [%s], Found - [%s]", sourceCheckSum, destinationCheckSum)
	}

	return nil
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

	if utils.RunWithRBAC.RbacFlag != true {

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
		log.Infof("Associated Resources - [%+v]", restore.Destination.TargetCluster.Project.AssociatedResources)
	} else {
		log.Infof("Please create namespace and associate with Account-Admin account.")
	}

	return nil
}
