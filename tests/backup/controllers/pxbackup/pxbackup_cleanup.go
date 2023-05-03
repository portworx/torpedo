package pxbackup

import "github.com/portworx/torpedo/pkg/log"

func (p *PxbController) CleanUp() error {
	for organization, organizationObjects := range p.organizations {
		log.Infof("The current organization is %s", organization)
		log.Info("The list of cloud accounts added %s", organizationObjects.cloudAccounts)
		//for cloudAccountName, _ := range organizationObjects.cloudAccounts {
		//	_ = p.DeleteCloudAccount(cloudAccountName)
		//}
		log.Info("The list of backup locations added %s", organizationObjects.backupLocations)
		//for backupLocationName, _ := range organizationObjects.backupLocations {
		//	_ = p.DeleteBackupLocation(backupLocationName)
		//}
	}
	return nil
}
