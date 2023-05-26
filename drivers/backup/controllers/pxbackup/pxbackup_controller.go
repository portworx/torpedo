package pxbackup

import "github.com/portworx/torpedo/drivers/backup"

type Profile struct {
	isAdmin         bool
	isFirstTimeUser bool
	username        string
	password        string
}

type OrganizationObjects struct {
	cloudAccounts    map[string]*CloudAccountInfo
	backupLocations  map[string]*BackupLocationInfo
	clusters         map[string]*ClusterInfo
	rules            map[string]*RuleInfo
	backups          map[string]*BackupInfo
	restores         map[string]*RestoreInfo
	schedulePolicies map[string]*SchedulePolicyInfo
}

type PxbController struct {
	profile       Profile
	currentOrgId  string
	organizations map[string]*OrganizationObjects
}

func (p *PxbController) initializeDefaults() error {
	p.currentOrgId = "default"
	p.organizations = make(map[string]*OrganizationObjects, 0)
	p.organizations[p.currentOrgId] = &OrganizationObjects{}
	return nil
}

func (p *PxbController) signInAsAdmin() error {
	p.profile.isAdmin = true
	p.profile.isFirstTimeUser = false
	p.profile.username = "admin"
	if err := p.initializeDefaults(); err != nil {
		return err
	}
	return nil
}

func (p *PxbController) signInAsExistingUser(username string, password string) error {
	p.profile.isAdmin = false
	p.profile.isFirstTimeUser = false
	p.profile.username = username
	p.profile.password = password
	if err := p.initializeDefaults(); err != nil {
		return err
	}
	return nil
}

func (p *PxbController) signInAsFirstTimeUser(username string, password string) error {
	p.profile.isAdmin = false
	p.profile.isFirstTimeUser = true
	p.profile.username = username
	p.profile.password = password
	if err := p.initializeDefaults(); err != nil {
		return err
	}
	return nil
}

type RegisterNewUserConfig struct {
	username  string
	firstName string
	lastName  string
	email     string
	password  string
}

func NewUser(username string, password string) *RegisterNewUserConfig {
	return &RegisterNewUserConfig{
		username:  username,
		password:  password,
		firstName: "first-" + username,
		lastName:  "last-" + username,
		email:     username + "@cnbu.com",
	}
}

func (c *RegisterNewUserConfig) SetFirstName(firstName string) *RegisterNewUserConfig {
	c.firstName = firstName
	return c
}

func (c *RegisterNewUserConfig) SetLastName(lastName string) *RegisterNewUserConfig {
	c.lastName = lastName
	return c
}

func (c *RegisterNewUserConfig) SetEmail(email string) *RegisterNewUserConfig {
	c.email = email
	return c
}

func (c *RegisterNewUserConfig) GetController() (*PxbController, error) {
	if err := backup.AddUser(c.username, c.firstName, c.lastName, c.email, c.password); err != nil {
		return nil, err
	}
	userController := &PxbController{}
	if err := userController.signInAsFirstTimeUser(c.username, c.password); err != nil {
		return nil, err
	}
	return userController, nil
}

func SetControllers(controllers *map[string]*PxbController, userCredentials map[string]string) error {
	if userCredentials != nil {
		for username, password := range userCredentials {
			exists, err := isUserPresent(username)
			if err != nil {
				return err
			}
			if exists {
				userController := &PxbController{}
				if err := userController.signInAsExistingUser(username, password); err != nil {
					return err
				}
				(*controllers)[username] = userController
			} else {
				userController, err := NewUser(username, password).GetController()
				if err != nil {
					return err
				}
				(*controllers)[username] = userController
			}
		}
	}
	adminController := &PxbController{}
	if err := adminController.signInAsAdmin(); err != nil {
		return err
	}
	(*controllers)["admin"] = adminController
	return nil
}
