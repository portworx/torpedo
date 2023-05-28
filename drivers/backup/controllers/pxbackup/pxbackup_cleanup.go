package pxbackup

import "github.com/portworx/torpedo/pkg/log"

func (p *PxBackupController) Cleanup() error {
	for organization, organizationObjects := range p.organizations {
		log.Infof("Cleaning up organization [%s]", organization)
		for cloudAccountName := range organizationObjects.cloudAccounts {
			log.Infof("Deleting cloud-account [%s]", cloudAccountName)
			_ = p.CloudAccount(cloudAccountName).Delete()
		}
	}
	return nil
}
