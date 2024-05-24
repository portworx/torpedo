package pdslibs

//TODO: This needs to be moved to workflow level

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/pkg/log"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	bkpJobCompleted = "APPLIED"
)

const (
	restoreCompleted = "SUCCESSFUL"
	restoreFailed    = "FAILED"
)

// ValidateAdhocBackup triggers the adhoc backup for given ds and store at the given backup target and validate them
func ValidateAdhocBackup(deploymentId string, backupConfigId string) error {
	var bkpJobs *automationModels.PDSBackupConfigResponse

	waitErr := wait.Poll(bkpTimeInterval, bkpMaxtimeInterval, func() (bool, error) {
		bkpJobs, err = GetBackupConfig(backupConfigId)
		if err != nil {
			return false, err
		}
		log.Infof("[Backup job: %s] Name: %s", *bkpJobs.Get.Status.Phase, *bkpJobs.Get.Meta.Name)
		if *bkpJobs.Get.Status.Phase == bkpJobCompleted {
			return true, nil
		} else {
			return false, nil
		}
	})
	if waitErr != nil {
		return fmt.Errorf("error occured while polling the status of backup job object. Err:%v", waitErr)
	}

	log.Infof("Created adhoc backup successfully for %v,"+
		" backup job: %v",
		deploymentId, *bkpJobs.Get.Meta.Name)
	return nil
}

// ValidateRestoreDeployment takes the restoreId and namespace as param and entrypoint to validate the restored deployments
func ValidateRestoreDeployment(restoreId, namespace string) error {
	restore, err := ValidateRestoreStatus(restoreId)
	if err != nil {
		return err
	}

	newDeployment := make(map[string]string)
	newDeployment[*restore.Get.Meta.Name] = restore.Get.Config.DestinationReferences.DeploymentId

	err = ValidateDataServiceDeploymentHealth(restore.Get.Config.DestinationReferences.DeploymentId, PDS_DEPLOYMENT_AVAILABLE)
	if err != nil {
		return fmt.Errorf("error while validating restored deployment readiness")
	}

	sourceDeployment, err := GetDeployment(restore.Get.Config.SourceReferences.DeploymentId)
	if err != nil {
		return fmt.Errorf("error while fetching source deployment object")
	}

	destinationDeployment, err := GetDeployment(restore.Get.Config.DestinationReferences.DeploymentId)
	if err != nil {
		return fmt.Errorf("error while fetching destination deployment object")
	}

	err = ValidateRestore(sourceDeployment.Get.Config.DeploymentTopologies[0], destinationDeployment.Get.Config.DeploymentTopologies[0])
	if err != nil {
		return fmt.Errorf("error while validation data service entities(i.e App config, resource etc). Err: %v", err)
	}

	return nil
}

// ValidateRestoreAfterSourceDeploymentUpgrade Validates the restored deployment config with source deployment config before upgrade
func ValidateRestoreAfterSourceDeploymentUpgrade(restoreId string, sourceDeploymentConfig automationModels.DeploymentTopology) error {
	restore, err := ValidateRestoreStatus(restoreId)
	if err != nil {
		return err
	}

	err = ValidateDataServiceDeploymentHealth(restore.Get.Config.DestinationReferences.DeploymentId, PDS_DEPLOYMENT_AVAILABLE)
	if err != nil {
		return fmt.Errorf("error while validating restored deployment readiness")
	}

	destinationDeployment, err := GetDeployment(restore.Get.Config.DestinationReferences.DeploymentId)
	if err != nil {
		return fmt.Errorf("error while fetching destination deployment object")
	}

	err = ValidateRestore(sourceDeploymentConfig, destinationDeployment.Get.Config.DeploymentTopologies[0])
	if err != nil {
		return fmt.Errorf("error while validation data service entities(i.e App config, resource etc). Err: %v", err)
	}

	return nil
}

// ValidateRestore validates the Resource, App and Storage configurations of source and destination deployments
func ValidateRestore(sourceDep, destDep automationModels.DeploymentTopology) error {

	//TODO : This validation needs to be revisited once we have the working pds templates api

	// Validate the Resource configuration
	log.Debugf("Source Deployment Topology - [%+v]", sourceDep)
	log.Debugf("Destination Deployment Topology - [%+v]", destDep)
	//sourceDep := sourceDeployment.Get.Config.DeploymentTopologies[0]
	//destDep := destinationDeployment.Get.Config.DeploymentTopologies[0]

	sourceResourceSettings := sourceDep.ResourceSettings
	destResourceSettings := destDep.ResourceSettings
	log.Debugf("Source Deployment resource settings - [%+v]", sourceResourceSettings)
	log.Debugf("Destination resource settings - [%+v]", destResourceSettings)

	log.Debugf("source resource settings - [%v]", *sourceResourceSettings.Id)
	if !reflect.DeepEqual(sourceResourceSettings, destResourceSettings) {
		return fmt.Errorf("restored resource configuration are not same as backed up resource config")
	}

	// Validate the StorageOption configuration
	sourceStorageOption := sourceDep.StorageOptions
	destStorageOption := destDep.StorageOptions
	log.Debugf("Source Deployment storage options - [%+v]", sourceStorageOption)
	log.Debugf("Destination storage options - [%+v]", destStorageOption)
	if !reflect.DeepEqual(sourceStorageOption, destStorageOption) {
		return fmt.Errorf("restored storage options configuration are not same as backed up resource storage options config")
	}

	// Validate the Application configuration
	sourceAppConfig := sourceDep.ServiceConfigurations
	destAppConfig := destDep.ServiceConfigurations
	log.Debugf("Source Deployment app config - [%+v]", sourceAppConfig)
	log.Debugf("Destination app config - [%+v]", destAppConfig)
	if !reflect.DeepEqual(sourceAppConfig, destAppConfig) {
		return fmt.Errorf("restored application configuration are not same as backed up application config")
	}

	// Validate the replicas
	sourceReplicas := sourceDep.Replicas
	destReplicas := destDep.Replicas
	log.Debugf("source replicas [%d], dest repicas [%d]", *sourceReplicas, *destReplicas)
	if *sourceReplicas != *destReplicas {
		return fmt.Errorf("restored replicas are not same as backed up resource config")
	}

	return nil
}

// ValidateRestoreStatus validates the health of the restored deployments
func ValidateRestoreStatus(restoreId string) (*automationModels.PDSRestoreResponse, error) {
	//var wfRestore WorkflowRestore
	var restoreResp *automationModels.PDSRestoreResponse

	err := wait.Poll(restoreTimeInterval, timeOut, func() (bool, error) {
		restoreResp, err = GetRestore(restoreId)
		state := restoreResp.Get.Status.Phase
		if err != nil {
			log.Errorf("failed during fetching the restore object, %v", err)
			return false, err
		}
		log.Infof("[%s] phase -  %v", *restoreResp.Get.Meta.Name, state)
		if strings.ToLower(state) == strings.ToLower(restoreFailed) {
			return true, fmt.Errorf("Restore failed - Name [%s], ID [%s], Phase - [%s], ErrorCode - [%s], errorMessage - [%s]", *restoreResp.Get.Meta.Name, *restoreResp.Get.Meta.Uid, restoreResp.Get.Status.Phase, restoreResp.Get.Status.ErrorCode, restoreResp.Get.Status.ErrorMessage)
		}
		if strings.ToLower(state) != strings.ToLower(restoreCompleted) {
			return false, nil
		}

		return true, nil
	})
	if err != nil {
		return nil, fmt.Errorf("Error while restoring the deployment: %v\n", err)
	}
	return restoreResp, nil
}
