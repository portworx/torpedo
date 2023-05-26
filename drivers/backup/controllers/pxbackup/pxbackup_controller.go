package pxbackup

import (
	"fmt"
	"github.com/portworx/torpedo/drivers/backup"
	"github.com/portworx/torpedo/drivers/backup/utils"
)

type Profile struct {
	isAdmin         bool
	isFirstTimeUser bool
	username        string
	password        string
}

type OrganizationObjects struct {
	cloudAccounts map[string]*CloudAccountInfo
	//backupLocations  map[string]*BackupLocationInfo
	//clusters         map[string]*ClusterInfo
	//rules            map[string]*RuleInfo
	//backups          map[string]*BackupInfo
	//restores         map[string]*RestoreInfo
	//schedulePolicies map[string]*SchedulePolicyInfo
}

type PxBackupController struct {
	profile       Profile
	currentOrgId  string
	organizations map[string]*OrganizationObjects
}

func (p *PxBackupController) initializeDefaults() error {
	p.currentOrgId = "default"
	p.organizations = make(map[string]*OrganizationObjects, 0)
	p.organizations[p.currentOrgId] = &OrganizationObjects{}
	return nil
}

func (p *PxBackupController) signInAsAdmin() error {
	p.profile.isAdmin = true
	p.profile.isFirstTimeUser = false
	p.profile.username = "admin"
	if err := p.initializeDefaults(); err != nil {
		return err
	}
	return nil
}

func (p *PxBackupController) signInAsExistingUser(username string, password string) error {
	p.profile.isAdmin = false
	p.profile.isFirstTimeUser = false
	p.profile.username = username
	p.profile.password = password
	if err := p.initializeDefaults(); err != nil {
		return err
	}
	return nil
}

func (p *PxBackupController) signInAsFirstTimeUser(username string, password string) error {
	p.profile.isAdmin = false
	p.profile.isFirstTimeUser = true
	p.profile.username = username
	p.profile.password = password
	if err := p.initializeDefaults(); err != nil {
		return err
	}
	return nil
}

func AddPxBackupControllersToMap(pxBackupControllerMap *map[string]*PxBackupController, userCredentials map[string]string) error {
	if userCredentials != nil {
		for username, password := range userCredentials {
			present, err := backup.IsUserPresent(username)
			if err != nil {
				debugMessage := fmt.Sprintf("username [%s]", username)
				return utils.ProcessError(err, debugMessage)
			}
			userController := &PxBackupController{}
			if present {
				err = userController.signInAsExistingUser(username, password)
				if err != nil {
					return err
				}
			} else {
				err = NewUser(username, password).Register()
				if err != nil {
					debugMessage := fmt.Sprintf("username [%s]", username)
					return utils.ProcessError(err, debugMessage)
				}
				err = userController.signInAsFirstTimeUser(username, password)
				if err != nil {
					debugMessage := fmt.Sprintf("username [%s]", username)
					return utils.ProcessError(err, debugMessage)
				}
			}
			(*pxBackupControllerMap)[username] = userController
		}
	}
	adminController := &PxBackupController{}
	err := adminController.signInAsAdmin()
	if err != nil {
		return utils.ProcessError(err)
	}
	(*pxBackupControllerMap)["admin"] = adminController
	return nil
}
