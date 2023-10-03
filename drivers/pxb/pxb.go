package pxb

//type User struct {
//	Spec                  *auth.UserRepresentation
//	PxBackup              *PxBackup
//	OrganizationDataStore *generics.DataStore[*Organization]
//}
//
//type Organization struct {
//	Spec                     *api.OrganizationObject
//	SchedulePolicyDataStore  *generics.DataStore[*api.SchedulePolicyObject]
//	BackupScheduleDataStore  *generics.DataStore[*api.BackupScheduleObject]
//	ClusterDataStore         *generics.DataStore[*api.ClusterObject]
//	CloudCredentialDataStore *generics.DataStore[*api.CloudCredentialObject]
//	BackupLocationDataStore  *generics.DataStore[*api.BackupLocationObject]
//	BackupDataStore          *generics.DataStore[*api.BackupObject]
//	RestoreDataStore         *generics.DataStore[*api.RestoreObject]
//	RuleDataStore            *generics.DataStore[*api.RuleObject]
//	RoleDataStore            *generics.DataStore[*api.RoleObject]
//}
//
//type PxBackup struct {
//	UserDataStore *generics.DataStore[*User]
//}
//
//func (b *PxBackup) AddTestUser() error {
//	user := fmt.Sprintf("pxb-user-%v", time.Now().Unix())
//	log.Infof("Creating user %s", user)
//	err := auth.AddUser(context.Background(), user, "firstName", "lastName", "fl@cnbu.com", true, "admin", true)
//	log.Errorf("Creating user caused error: %v", err)
//	log.Infof("Created user %s", user)
//	return err
//}
