package pxbackup

type PxbController struct {
	profile       Profile
	currentOrgId  string
	organizations map[string]*OrganizationObjects
}

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
