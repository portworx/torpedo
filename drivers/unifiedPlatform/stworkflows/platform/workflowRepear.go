package platform

import (
	"github.com/pure-px/torpedo/drivers/unifiedPlatform/platformLibs"
	"github.com/pure-px/torpedo/pkg/log"
	"strings"
)

type WorkflowRepear struct {
	TargetCluster   TargetCluster
	CloudCredential CloudCredential
	BackUpLocation  BackUpLocation
}

type TargetCluster struct {
	TargetClusters map[string]string
}

type CloudCredential struct {
	CloudCredentials map[string]string
}

type BackUpLocation struct {
	TargetLocations map[string]string
}

func (tc *TargetCluster) DeleteAllUnhealthyTargetClustersWithPrefix(tenantId, targetClusterPrefix string, force bool) []error {
	var (
		allErros []error
		count    int
	)
	ListTCResponse, err := platformLibs.ListTargetClusters(tenantId)
	if err != nil {
		allErros = append(allErros, err)
		return allErros
	}
	for _, tc := range ListTCResponse.Clusters {
		if strings.Contains(*tc.Meta.Name, targetClusterPrefix) && (tc.Status.Phase == "DISCONNECTED") {
			log.Debugf("Deleting TC [%s] of status [%s]", *tc.Meta.Name, tc.Status.Phase)
			err = platformLibs.DeleteTargetCluster(*tc.Meta.Uid, force)
			if err != nil {
				allErros = append(allErros, err)
			}
			count++
		}
	}
	if count == 0 {
		log.Infof("No Target Cluster found with the given prefix-[%s]", targetClusterPrefix)
	}
	log.Debugf("Total no of target clusters found in DISCONNECTED phase [%d]", count)
	return allErros
}

func (tl *BackUpLocation) DeleteAllBackUpLocationWithPrefix(tenantId, bkpLocationPrefix string) []error {
	var (
		allErros []error
		count    int
	)
	ListBackUpLocationResponse, err := platformLibs.ListBackupLocation(tenantId, "", "CREATED_AT", "DESC")
	if err != nil {
		allErros = append(allErros, err)
		return allErros
	}

	for _, backupLocation := range ListBackUpLocationResponse.List.BackupLocations {
		if strings.Contains(*backupLocation.Meta.Name, bkpLocationPrefix) {
			log.Debugf("Deleting BackUpLocation [%s]", *backupLocation.Meta.Name)
			err = platformLibs.DeleteBackupLocation(*backupLocation.Meta.Uid)
			if err != nil {
				allErros = append(allErros, err)
			}
			count++
		}
	}
	if count == 0 {
		log.Infof("No BackupLocation found with the given prefix-[%s]", bkpLocationPrefix)
	}
	log.Debugf("Total no of BackUpLocation found-[%d] with the given prefix-[%s]", count, bkpLocationPrefix)

	return allErros
}

func (cc *CloudCredential) DeleteAllCloudCredentialWithPrefix(tenantId, cloudCredPrefix string) []error {

	var (
		allErros []error
		count    int
	)
	ListCloudCredResponse, err := platformLibs.ListCloudCredential(tenantId, "", "CREATED_AT", "DESC")
	if err != nil {
		allErros = append(allErros, err)
		return allErros
	}

	for _, cloudCred := range ListCloudCredResponse.List.CloudCredentials {
		if strings.Contains(*cloudCred.Meta.Name, cloudCredPrefix) {
			log.Debugf("Deleting Cloud Credential [%s]", *cloudCred.Meta.Name)
			err = platformLibs.DeleteCloudCredential(*cloudCred.Meta.Uid)
			if err != nil {
				allErros = append(allErros, err)
			}
			count++
		}
	}
	if count == 0 {
		log.Infof("No CloudCredential found with the given prefix-[%s]", cloudCredPrefix)
	}
	log.Debugf("Total no of CloudCredential found-[%d] with the given prefix-[%s]", count, cloudCredPrefix)

	return allErros
}
