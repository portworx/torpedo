package pxbackup

import (
	"github.com/portworx/torpedo/drivers/backup"
)

// const (
// 	cloudAccountDeleteTimeout                 = 30 * time.Minute
// 	cloudAccountDeleteRetryTime               = 30 * time.Second
// 	storkDeploymentName                       = "stork"
// 	defaultStorkDeploymentNamespace           = "kube-system"
// 	upgradeStorkImage                         = "UPGRADE_STORK_IMAGE"
// 	latestStorkImage                          = "openstorage/stork:23.2.0"
// 	restoreNamePrefix                         = "tp-restore"
// 	destinationClusterName                    = "destination-cluster"
// 	appReadinessTimeout                       = 10 * time.Minute
// 	taskNamePrefix                            = "pxbackuptask"
// 	orgID                                     = "default"
// 	usersToBeCreated                          = "USERS_TO_CREATE"
// 	groupsToBeCreated                         = "GROUPS_TO_CREATE"
// 	maxUsersInGroup                           = "MAX_USERS_IN_GROUP"
// 	maxBackupsToBeCreated                     = "MAX_BACKUPS"
// 	maxWaitPeriodForBackupCompletionInMinutes = 40
// 	maxWaitPeriodForRestoreCompletionInMinute = 40
// 	maxWaitPeriodForBackupJobCancellation     = 20
// 	backupJobCancellationRetryTime            = 30
// 	K8sNodeReadyTimeout                       = 10
// 	K8sNodeRetryInterval                      = 30
// 	GlobalAWSBucketPrefix                     = "global-aws"
// 	GlobalAzureBucketPrefix                   = "global-azure"
// 	GlobalGCPBucketPrefix                     = "global-gcp"
// 	GlobalAWSLockedBucketPrefix               = "global-aws-locked"
// 	GlobalAzureLockedBucketPrefix             = "global-azure-locked"
// 	GlobalGCPLockedBucketPrefix               = "global-gcp-locked"
// 	mongodbStatefulset                        = "pxc-backup-mongodb"
// 	pxBackupDeployment                        = "px-backup"
// 	backupDeleteTimeout                       = 20 * time.Minute
// 	backupDeleteRetryTime                     = 30 * time.Second
// 	backupLocationDeleteTimeout               = 30 * time.Minute
// 	backupLocationDeleteRetryTime             = 30 * time.Second
// 	rebootNodeTimeout                         = 1 * time.Minute
// 	rebootNodeTimeBeforeRetry                 = 5 * time.Second
// 	latestPxBackupVersion                     = "2.4.0"
// 	latestPxBackupHelmBranch                  = "master"
// 	pxCentralPostInstallHookJobName           = "pxcentral-post-install-hook"
// 	quickMaintenancePod                       = "quick-maintenance-repo"
// 	fullMaintenancePod                        = "full-maintenance-repo"
// 	jobDeleteTimeout                          = 5 * time.Minute
// 	jobDeleteRetryTime                        = 10 * time.Second
// 	podStatusTimeOut                          = 20 * time.Minute
// 	podStatusRetryTime                        = 30 * time.Second
// )

// // My

// var (
// 	// GlobalAWSBucketName         string
// 	// GlobalAzureBucketName       string
// 	// GlobalGCPBucketName         string
// 	// GlobalAWSLockedBucketName   string
// 	// GlobalAzureLockedBucketName string
// 	// GlobalGCPLockedBucketName   string
// 	DefaultCloudProviders = []string{"aws"}
// )

var (
	DefaultPassword string
)

// // My end

// // var (
// // 	// User should keep updating preRuleApp, postRuleApp
// // 	preRuleApp                  = []string{"cassandra", "postgres"}
// // 	postRuleApp                 = []string{"cassandra"}
// // 	GlobalAWSBucketName         string
// // 	GlobalAzureBucketName       string
// // 	GlobalGCPBucketName         string
// // 	GlobalAWSLockedBucketName   string
// // 	GlobalAzureLockedBucketName string
// // 	GlobalGCPLockedBucketName   string
// // 	CloudProviders              = []string{"aws"}
// // 	DefaultPassword              string
// // )

// type userRoleAccess struct {
// 	user     string
// 	roles    backup.PxBackupRole
// 	accesses BackupAccess
// 	context  context.Context
// }

// type userAccessContext struct {
// 	user     string
// 	accesses BackupAccess
// 	context  context.Context
// }

// var backupAccessKeyValue = map[BackupAccess]string{
// 	1: "ViewOnlyAccess",
// 	2: "RestoreAccess",
// 	3: "FullAccess",
// }

// var storkLabel = map[string]string{
// 	"name": "stork",
// }

// type BackupAccess int32

// const (
// 	ViewOnlyAccess BackupAccess = 1
// 	RestoreAccess               = 2
// 	FullAccess                  = 3
// )

// // // Set default provider as aws
// // func getProviders() []string {
// // 	providersStr := os.Getenv("PROVIDERS")
// // 	if providersStr != "" {
// // 		return strings.Split(providersStr, ",")
// // 	}
// // 	return DefaultCloudProviders
// // }

// // getPXNamespace fetches px namespace from env else sends backup kube-system
// func getPXNamespace() string {
// 	namespace := os.Getenv("PX_NAMESPACE")
// 	if namespace != "" {
// 		return namespace
// 	}
// 	return defaultStorkDeploymentNamespace
// }

// // func GetAllBackupsForUser(username, password string) ([]string, error) {
// // 	backupNames := make([]string, 0)
// // 	backupDriver := Inst().Backup
// // 	ctx, err := backup.GetNonAdminCtx(username, password)
// // 	if err != nil {
// // 		return nil, err
// // 	}

// // 	backupEnumerateReq := &api.BackupEnumerateRequest{
// // 		OrgId: orgID,
// // 	}
// // 	currentBackups, err := backupDriver.EnumerateBackup(ctx, backupEnumerateReq)
// // 	if err != nil {
// // 		return nil, err
// // 	}
// // 	for _, backup := range currentBackups.GetBackups() {
// // 		backupNames = append(backupNames, backup.GetName())
// // 	}
// // 	return backupNames, nil
// // }

// func getSizeOfMountPoint(podName string, namespace string, kubeConfigFile string) (int, error) {
// 	var number int
// 	ret, err := kubectlExec([]string{podName, "-n", namespace, "--kubeconfig=", kubeConfigFile, " -- /bin/df"})
// 	if err != nil {
// 		return 0, err
// 	}
// 	for _, line := range strings.SplitAfter(ret, "\n") {
// 		if strings.Contains(line, "pxd") {
// 			ret = strings.Fields(line)[3]
// 		}
// 	}
// 	number, err = strconv.Atoi(ret)
// 	if err != nil {
// 		return 0, err
// 	}
// 	return number, nil
// }

// func kubectlExec(arguments []string) (string, error) {
// 	if len(arguments) == 0 {
// 		return "", fmt.Errorf("no arguments supplied for kubectl command")
// 	}
// 	cmd := exec.Command("kubectl exec -it", arguments...)
// 	output, err := cmd.Output()
// 	log.Debugf("command output for '%s': %s", cmd.String(), string(output))
// 	if err != nil {
// 		return "", fmt.Errorf("error on executing kubectl command, Err: %+v", err)
// 	}
// 	return string(output), err
// }

// // func createUsers(numberOfUsers int) []string {
// // 	users := make([]string, 0)
// // 	log.InfoD("Creating %d users", numberOfUsers)
// // 	var wg sync.WaitGroup
// // 	for i := 1; i <= numberOfUsers; i++ {
// // 		userName := fmt.Sprintf("testuser%v-%v", i, time.Now().Unix())
// // 		firstName := fmt.Sprintf("FirstName%v", i)
// // 		lastName := fmt.Sprintf("LastName%v", i)
// // 		email := fmt.Sprintf("%v@cnbu.com", userName)
// // 		wg.Add(1)
// // 		go func(userName, firstName, lastName, email string) {
// // 			defer GinkgoRecover()
// // 			defer wg.Done()
// // 			err := backup.AddUser(userName, firstName, lastName, email, DefaultPassword)
// // 			Inst().Dash.VerifyFatal(err, nil, fmt.Sprintf("Creating user - %s", userName))
// // 			users = append(users, userName)
// // 		}(userName, firstName, lastName, email)
// // 	}
// // 	wg.Wait()
// // 	return users
// // }

// // // CleanupCloudSettingsAndClusters removes the backup location(s), cloud accounts and source/destination clusters for the given context
// // func CleanupCloudSettingsAndClusters(backupLocationMap map[string]string, credName string, cloudCredUID string, ctx context.Context) {
// // 	log.InfoD("Cleaning backup locations in map [%v], cloud credential [%s], source [%s] and destination [%s] cluster", backupLocationMap, credName, SourceClusterName, destinationClusterName)
// // 	if len(backupLocationMap) != 0 {
// // 		for backupLocationUID, bkpLocationName := range backupLocationMap {
// // 			err := DeleteBackupLocation(bkpLocationName, backupLocationUID, orgID, true)
// // 			Inst().Dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying deletion of backup location [%s]", bkpLocationName))
// // 			backupLocationDeleteStatusCheck := func() (interface{}, bool, error) {
// // 				status, err := IsBackupLocationPresent(bkpLocationName, ctx, orgID)
// // 				if err != nil {
// // 					return "", true, fmt.Errorf("backup location %s still present with error %v", bkpLocationName, err)
// // 				}
// // 				if status {
// // 					return "", true, fmt.Errorf("backup location %s is not deleted yet", bkpLocationName)
// // 				}
// // 				return "", false, nil
// // 			}
// // 			_, err = task.DoRetryWithTimeout(backupLocationDeleteStatusCheck, cloudAccountDeleteTimeout, cloudAccountDeleteRetryTime)
// // 			Inst().Dash.VerifySafely(err, nil, fmt.Sprintf("Verifying backup location deletion status %s", bkpLocationName))
// // 		}
// // 		err := DeleteCloudCredential(credName, orgID, cloudCredUID)
// // 		Inst().Dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying deletion of cloud cred [%s]", credName))
// // 		cloudCredDeleteStatus := func() (interface{}, bool, error) {
// // 			status, err := IsCloudCredPresent(credName, ctx, orgID)
// // 			if err != nil {
// // 				return "", true, fmt.Errorf("cloud cred %s still present with error %v", credName, err)
// // 			}
// // 			if status {
// // 				return "", true, fmt.Errorf("cloud cred %s is not deleted yet", credName)
// // 			}
// // 			return "", false, nil
// // 		}
// // 		_, err = task.DoRetryWithTimeout(cloudCredDeleteStatus, cloudAccountDeleteTimeout, cloudAccountDeleteRetryTime)
// // 		Inst().Dash.VerifySafely(err, nil, fmt.Sprintf("Deleting cloud cred %s", credName))
// // 	}
// // 	err := DeleteCluster(SourceClusterName, orgID, ctx)
// // 	Inst().Dash.VerifySafely(err, nil, fmt.Sprintf("Deleting cluster %s", SourceClusterName))
// // 	err = DeleteCluster(destinationClusterName, orgID, ctx)
// // 	Inst().Dash.VerifySafely(err, nil, fmt.Sprintf("Deleting cluster %s", destinationClusterName))
// // }

// // func getEnv(environmentVariable string, defaultValue string) string {
// // 	value, present := os.LookupEnv(environmentVariable)
// // 	if !present {
// // 		value = defaultValue
// // 	}
// // 	return value
// // }

// // GetAllBackupsAdmin returns all the backups that px-central-admin has access to
// func GetAllBackupsAdmin() ([]string, error) {
// 	var bkp *api.BackupObject
// 	backupNames := make([]string, 0)
// 	backupDriver := Inst().Backup
// 	ctx, err := backup.GetAdminCtxFromSecret()
// 	if err != nil {
// 		return nil, err
// 	}
// 	bkpEnumerateReq := &api.BackupEnumerateRequest{
// 		OrgId: orgID}
// 	curBackups, err := backupDriver.EnumerateBackup(ctx, bkpEnumerateReq)
// 	if err != nil {
// 		return nil, err
// 	}
// 	for _, bkp = range curBackups.GetBackups() {
// 		backupNames = append(backupNames, bkp.GetName())
// 	}
// 	return backupNames, nil
// }

// // GetAllRestoresAdmin returns all the backups that px-central-admin has access to
// func GetAllRestoresAdmin() ([]string, error) {
// 	restoreNames := make([]string, 0)
// 	backupDriver := Inst().Backup
// 	ctx, err := backup.GetAdminCtxFromSecret()
// 	if err != nil {
// 		return restoreNames, err
// 	}

// 	restoreEnumerateRequest := &api.RestoreEnumerateRequest{
// 		OrgId: orgID,
// 	}
// 	restoreResponse, err := backupDriver.EnumerateRestore(ctx, restoreEnumerateRequest)
// 	if err != nil {
// 		return restoreNames, err
// 	}
// 	for _, restore := range restoreResponse.GetRestores() {
// 		restoreNames = append(restoreNames, restore.Name)
// 	}
// 	return restoreNames, nil
// }

// func generateEncryptionKey() string {
// 	var lower = []byte("abcdefghijklmnopqrstuvwxyz")
// 	var upper = []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
// 	var number = []byte("0123456789")
// 	var special = []byte("~=+%^*/()[]{}/!@#$?|")
// 	allChar := append(lower, upper...)
// 	allChar = append(allChar, number...)
// 	allChar = append(allChar, special...)

// 	b := make([]byte, 12)
// 	// select 1 upper, 1 lower, 1 number and 1 special
// 	b[0] = lower[rand.Intn(len(lower))]
// 	b[1] = upper[rand.Intn(len(upper))]
// 	b[2] = number[rand.Intn(len(number))]
// 	b[3] = special[rand.Intn(len(special))]
// 	for i := 4; i < 12; i++ {
// 		// randomly select 1 character from given charset
// 		b[i] = allChar[rand.Intn(len(allChar))]
// 	}

// 	//shuffle character
// 	rand.Shuffle(len(b), func(i, j int) {
// 		b[i], b[j] = b[j], b[i]
// 	})

// 	return string(b)
// }

// func GetScheduleUID(scheduleName string, orgID string, ctx context.Context) (string, error) {
// 	backupDriver := Inst().Backup
// 	backupScheduleInspectRequest := &api.BackupScheduleInspectRequest{
// 		Name:  scheduleName,
// 		Uid:   "",
// 		OrgId: orgID,
// 	}
// 	resp, err := backupDriver.InspectBackupSchedule(ctx, backupScheduleInspectRequest)
// 	if err != nil {
// 		return "", err
// 	}
// 	scheduleUid := resp.GetBackupSchedule().GetUid()
// 	return scheduleUid, err
// }

// func removeStringItemFromSlice(itemList []string, item []string) []string {
// 	sort.Sort(sort.StringSlice(itemList))
// 	for _, element := range item {
// 		index := sort.StringSlice(itemList).Search(element)
// 		itemList = append(itemList[:index], itemList[index+1:]...)
// 	}
// 	return itemList
// }

// func removeIntItemFromSlice(itemList []int, item []int) []int {
// 	sort.Sort(sort.IntSlice(itemList))
// 	for _, element := range item {
// 		index := sort.IntSlice(itemList).Search(element)
// 		itemList = append(itemList[:index], itemList[index+1:]...)
// 	}
// 	return itemList
// }

// func getAllBackupLocations(ctx context.Context) (map[string]string, error) {
// 	backupLocationMap := make(map[string]string, 0)
// 	backupDriver := Inst().Backup
// 	backupLocationEnumerateRequest := &api.BackupLocationEnumerateRequest{
// 		OrgId: orgID,
// 	}
// 	response, err := backupDriver.EnumerateBackupLocation(ctx, backupLocationEnumerateRequest)
// 	if err != nil {
// 		return backupLocationMap, err
// 	}
// 	if len(response.BackupLocations) > 0 {
// 		for _, backupLocation := range response.BackupLocations {
// 			backupLocationMap[backupLocation.Uid] = backupLocation.Name
// 		}
// 		log.Infof("The backup locations and their UID are %v", backupLocationMap)
// 	} else {
// 		log.Info("No backup locations found")
// 	}
// 	return backupLocationMap, nil
// }

// func getAllCloudCredentials(ctx context.Context) (map[string]string, error) {
// 	cloudCredentialMap := make(map[string]string, 0)
// 	backupDriver := Inst().Backup
// 	cloudCredentialEnumerateRequest := &api.CloudCredentialEnumerateRequest{
// 		OrgId: orgID,
// 	}
// 	response, err := backupDriver.EnumerateCloudCredential(ctx, cloudCredentialEnumerateRequest)
// 	if err != nil {
// 		return cloudCredentialMap, err
// 	}
// 	if len(response.CloudCredentials) > 0 {
// 		for _, cloudCredential := range response.CloudCredentials {
// 			cloudCredentialMap[cloudCredential.Uid] = cloudCredential.Name
// 		}
// 		log.Infof("The cloud credentials and their UID are %v", cloudCredentialMap)
// 	} else {
// 		log.Info("No cloud credentials found")
// 	}
// 	return cloudCredentialMap, nil
// }

// func GetAllRestoresNonAdminCtx(ctx context.Context) ([]string, error) {
// 	restoreNames := make([]string, 0)
// 	backupDriver := Inst().Backup
// 	restoreEnumerateRequest := &api.RestoreEnumerateRequest{
// 		OrgId: orgID,
// 	}
// 	restoreResponse, err := backupDriver.EnumerateRestore(ctx, restoreEnumerateRequest)
// 	if err != nil {
// 		return restoreNames, err
// 	}
// 	for _, restore := range restoreResponse.GetRestores() {
// 		restoreNames = append(restoreNames, restore.Name)
// 	}
// 	return restoreNames, nil
// }

// // DeletePodWithLabelInNamespace kills pod with the given label in the given namespace
// func DeletePodWithLabelInNamespace(namespace string, label map[string]string) error {
// 	pods, err := core.Instance().GetPods(namespace, label)
// 	if err != nil {
// 		return err
// 	}
// 	for _, pod := range pods.Items {
// 		err := core.Instance().DeletePod(pod.GetName(), namespace, false)
// 		if err != nil {
// 			return err
// 		}
// 		err = core.Instance().WaitForPodDeletion(pod.GetUID(), namespace, 5*time.Minute)
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }

// // IsBackupLocationPresent checks whether the backup location is present or not
// func IsBackupLocationPresent(bkpLocation string, ctx context.Context, orgID string) (bool, error) {
// 	backupLocationNames := make([]string, 0)
// 	backupLocationEnumerateRequest := &api.BackupLocationEnumerateRequest{
// 		OrgId: orgID,
// 	}
// 	response, err := Inst().Backup.EnumerateBackupLocation(ctx, backupLocationEnumerateRequest)
// 	if err != nil {
// 		return false, err
// 	}

// 	for _, backupLocationObj := range response.GetBackupLocations() {
// 		backupLocationNames = append(backupLocationNames, backupLocationObj.GetName())
// 		if backupLocationObj.GetName() == bkpLocation {
// 			log.Infof("Backup location [%s] is present", bkpLocation)
// 			return true, nil
// 		}
// 	}
// 	log.Infof("Backup locations fetched - %s", backupLocationNames)
// 	return false, nil
// }

// // IsCloudCredPresent checks whether the Cloud Cred is present or not
// func IsCloudCredPresent(cloudCredName string, ctx context.Context, orgID string) (bool, error) {
// 	cloudCredEnumerateRequest := &api.CloudCredentialEnumerateRequest{
// 		OrgId:          orgID,
// 		IncludeSecrets: false,
// 	}
// 	cloudCredObjs, err := Inst().Backup.EnumerateCloudCredential(ctx, cloudCredEnumerateRequest)
// 	if err != nil {
// 		return false, err
// 	}
// 	for _, cloudCredObj := range cloudCredObjs.GetCloudCredentials() {
// 		if cloudCredObj.GetName() == cloudCredName {
// 			log.Infof("Cloud Credential [%s] is present", cloudCredName)
// 			return true, nil
// 		}
// 	}
// 	return false, nil
// }

// // IsPresent verifies if the given data is present in slice of data
// func IsPresent(dataSlice interface{}, data interface{}) bool {
// 	s := reflect.ValueOf(dataSlice)
// 	for i := 0; i < s.Len(); i++ {
// 		if s.Index(i).Interface() == data {
// 			return true
// 		}
// 	}
// 	return false
// }

// // GetPxBackupVersion return the version of Px Backup as a VersionInfo struct
// func GetPxBackupVersion() (*api.VersionInfo, error) {
// 	ctx, err := backup.GetAdminCtxFromSecret()
// 	if err != nil {
// 		return nil, err
// 	}
// 	versionResponse, err := Inst().Backup.GetPxBackupVersion(ctx, &api.VersionGetRequest{})
// 	if err != nil {
// 		return nil, err
// 	}
// 	backupVersion := versionResponse.GetVersion()
// 	return backupVersion, nil
// }

// // GetPxBackupVersionString returns the version of Px Backup like 2.4.0-e85b680
// func GetPxBackupVersionString() (string, error) {
// 	backupVersion, err := GetPxBackupVersion()
// 	if err != nil {
// 		return "", err
// 	}
// 	return fmt.Sprintf("%s.%s.%s-%s", backupVersion.GetMajor(), backupVersion.GetMinor(), backupVersion.GetPatch(), backupVersion.GetGitCommit()), nil
// }

// // GetPxBackupVersionSemVer returns the version of Px Backup in semver format like 2.4.0
// func GetPxBackupVersionSemVer() (string, error) {
// 	backupVersion, err := GetPxBackupVersion()
// 	if err != nil {
// 		return "", err
// 	}
// 	return fmt.Sprintf("%s.%s.%s", backupVersion.GetMajor(), backupVersion.GetMinor(), backupVersion.GetPatch()), nil
// }

// // GetPxBackupBuildDate returns the Px Backup build date
// func GetPxBackupBuildDate() (string, error) {
// 	ctx, err := backup.GetAdminCtxFromSecret()
// 	if err != nil {
// 		return "", err
// 	}
// 	versionResponse, err := Inst().Backup.GetPxBackupVersion(ctx, &api.VersionGetRequest{})
// 	if err != nil {
// 		return "", err
// 	}
// 	backupVersion := versionResponse.GetVersion()
// 	return backupVersion.GetBuildDate(), nil
// }

// // UpgradePxBackup will perform the upgrade tasks for Px Backup to the version passed as string
// // Eg: versionToUpgrade := "2.4.0"
// func UpgradePxBackup(versionToUpgrade string) error {
// 	var cmd string

// 	// Compare and validate the upgrade path
// 	currentBackupVersionString, err := GetPxBackupVersionSemVer()
// 	if err != nil {
// 		return err
// 	}
// 	currentBackupVersion, err := version.NewSemver(currentBackupVersionString)
// 	if err != nil {
// 		return err
// 	}
// 	versionToUpgradeSemVer, err := version.NewSemver(versionToUpgrade)
// 	if err != nil {
// 		return err
// 	}

// 	if currentBackupVersion.GreaterThanOrEqual(versionToUpgradeSemVer) {
// 		return fmt.Errorf("px backup cannot be upgraded from version [%s] to version [%s]", currentBackupVersion.String(), versionToUpgradeSemVer.String())
// 	} else {
// 		log.InfoD("Upgrade path chosen (%s) ---> (%s)", currentBackupVersionString, versionToUpgrade)
// 	}

// 	// Getting Px Backup Namespace
// 	pxBackupNamespace, err := backup.GetPxBackupNamespace()
// 	if err != nil {
// 		return err
// 	}

// 	// Delete the pxcentral-post-install-hook job is it exists
// 	allJobs, err := batch.Instance().ListAllJobs(pxBackupNamespace, metav1.ListOptions{})
// 	if err != nil {
// 		return err
// 	}
// 	if len(allJobs.Items) > 0 {
// 		log.Infof("List of all the jobs in Px Backup Namespace [%s] - ", pxBackupNamespace)
// 		for _, job := range allJobs.Items {
// 			log.Infof(job.Name)
// 		}

// 		for _, job := range allJobs.Items {
// 			if strings.Contains(job.Name, pxCentralPostInstallHookJobName) {
// 				err = deleteJobAndWait(job)
// 				if err != nil {
// 					return err
// 				}
// 			}
// 		}
// 	} else {
// 		log.Infof("%s job not found", pxCentralPostInstallHookJobName)
// 	}

// 	// Get storage class using for px-backup deployment
// 	statefulSet, err := apps.Instance().GetStatefulSet(mongodbStatefulset, pxBackupNamespace)
// 	if err != nil {
// 		return err
// 	}
// 	pvcs, err := apps.Instance().GetPVCsForStatefulSet(statefulSet)
// 	if err != nil {
// 		return err
// 	}
// 	storageClassName := pvcs.Items[0].Spec.StorageClassName

// 	// Get the tarball required for helm upgrade
// 	cmd = fmt.Sprintf("curl -O  https://raw.githubusercontent.com/portworx/helm/%s/stable/px-central-%s.tgz", latestPxBackupHelmBranch, versionToUpgrade)
// 	log.Infof("curl command to get tarball: %v ", cmd)
// 	output, _, err := osutils.ExecShell(cmd)
// 	if err != nil {
// 		return fmt.Errorf("error downloading of tarball: %v", err)
// 	}
// 	log.Infof("Terminal output: %s", output)

// 	// Checking if all pods are healthy before upgrade
// 	err = ValidateAllPodsInPxBackupNamespace()
// 	if err != nil {
// 		return err
// 	}

// 	// Execute helm upgrade using cmd
// 	log.Infof("Upgrading Px-Backup version from %s to %s", currentBackupVersionString, versionToUpgrade)
// 	cmd = fmt.Sprintf("helm upgrade px-central px-central-%s.tgz --namespace %s --version %s --set persistentStorage.enabled=true,persistentStorage.storageClassName=\"%s\",pxbackup.enabled=true",
// 		versionToUpgrade, pxBackupNamespace, versionToUpgrade, *storageClassName)
// 	log.Infof("helm command: %v ", cmd)
// 	output, _, err = osutils.ExecShell(cmd)
// 	if err != nil {
// 		return fmt.Errorf("upgrade failed with error: %v", err)
// 	}
// 	log.Infof("Terminal output: %s", output)

// 	// Wait for post install hook job to be completed
// 	postInstallHookJobCompletedCheck := func() (interface{}, bool, error) {
// 		job, err := batch.Instance().GetJob(pxCentralPostInstallHookJobName, pxBackupNamespace)
// 		if err != nil {
// 			return "", true, err
// 		}
// 		if job.Status.Succeeded > 0 {
// 			log.Infof("Status of job %s after completion - "+
// 				"\nactive count - %d"+
// 				"\nsucceeded count - %d"+
// 				"\nfailed count - %d\n", job.Name, job.Status.Active, job.Status.Succeeded, job.Status.Failed)
// 			return "", false, nil
// 		}
// 		return "", true, fmt.Errorf("status of job %s not yet in desired state - "+
// 			"\nactive count - %d"+
// 			"\nsucceeded count - %d"+
// 			"\nfailed count - %d\n", job.Name, job.Status.Active, job.Status.Succeeded, job.Status.Failed)
// 	}
// 	_, err = task.DoRetryWithTimeout(postInstallHookJobCompletedCheck, 10*time.Minute, 30*time.Second)
// 	if err != nil {
// 		return err
// 	}

// 	// Checking if all pods are running
// 	err = ValidateAllPodsInPxBackupNamespace()
// 	if err != nil {
// 		return err
// 	}

// 	postUpgradeVersion, err := GetPxBackupVersionSemVer()
// 	if err != nil {
// 		return err
// 	}
// 	if !strings.EqualFold(postUpgradeVersion, versionToUpgrade) {
// 		return fmt.Errorf("expected version after upgrade was %s but got %s", versionToUpgrade, postUpgradeVersion)
// 	}
// 	log.InfoD("Px-Backup upgrade from %s to %s is complete", currentBackupVersionString, postUpgradeVersion)
// 	return nil
// }

// // deleteJobAndWait waits for the provided job to be deleted
// func deleteJobAndWait(job batchv1.Job) error {
// 	t := func() (interface{}, bool, error) {
// 		err := batch.Instance().DeleteJob(job.Name, job.Namespace)

// 		if err != nil {
// 			if strings.Contains(err.Error(), "not found") {
// 				return "", false, nil
// 			}
// 			return "", false, err
// 		}
// 		return "", true, fmt.Errorf("job %s not deleted", job.Name)
// 	}

// 	_, err := task.DoRetryWithTimeout(t, jobDeleteTimeout, jobDeleteRetryTime)
// 	if err != nil {
// 		return err
// 	}
// 	log.Infof("job %s deleted", job.Name)
// 	return nil
// }

// func ValidateAllPodsInPxBackupNamespace() error {
// 	pxBackupNamespace, err := backup.GetPxBackupNamespace()
// 	allPods, err := core.Instance().GetPods(pxBackupNamespace, nil)
// 	for _, pod := range allPods.Items {
// 		if strings.Contains(pod.Name, pxCentralPostInstallHookJobName) ||
// 			strings.Contains(pod.Name, quickMaintenancePod) ||
// 			strings.Contains(pod.Name, fullMaintenancePod) {
// 			continue
// 		}
// 		log.Infof("Checking status for pod - %s", pod.GetName())
// 		err = core.Instance().ValidatePod(&pod, 5*time.Minute, 30*time.Second)
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }

// // getStorkImageVersion returns current stork image version.
// func getStorkImageVersion() (string, error) {
// 	storkDeploymentNamespace, err := k8sutils.GetStorkPodNamespace()
// 	if err != nil {
// 		return "", err
// 	}
// 	storkDeployment, err := apps.Instance().GetDeployment(storkDeploymentName, storkDeploymentNamespace)
// 	if err != nil {
// 		return "", err
// 	}
// 	storkImage := storkDeployment.Spec.Template.Spec.Containers[0].Image
// 	storkImageVersion := strings.Split(storkImage, ":")[len(strings.Split(storkImage, ":"))-1]
// 	return storkImageVersion, nil
// }

// // upgradeStorkVersion upgrades the stork to the provided version.
// func upgradeStorkVersion(storkImageToUpgrade string) error {
// 	storkDeploymentNamespace, err := k8sutils.GetStorkPodNamespace()
// 	if err != nil {
// 		return err
// 	}
// 	currentStorkImageStr, err := getStorkImageVersion()
// 	if err != nil {
// 		return err
// 	}
// 	currentStorkVersion, err := version.NewSemver(currentStorkImageStr)
// 	if err != nil {
// 		return err
// 	}

// 	storkImageVersionToUpgradeStr := strings.Split(storkImageToUpgrade, ":")[len(strings.Split(storkImageToUpgrade, ":"))-1]
// 	storkImageVersionToUpgrade, err := version.NewSemver(storkImageVersionToUpgradeStr)
// 	if err != nil {
// 		return err
// 	}

// 	log.Infof("Current stork version : %s", currentStorkVersion)
// 	log.Infof("Upgrading stork version to : %s", storkImageVersionToUpgrade)

// 	if currentStorkVersion.GreaterThanOrEqual(storkImageVersionToUpgrade) {
// 		return fmt.Errorf("Cannot upgrade stork version from %s to %s as the current version is higher than the provided version", currentStorkVersion, storkImageVersionToUpgrade)
// 	}

// 	isOpBased, _ := Inst().V.IsOperatorBasedInstall()
// 	if isOpBased {
// 		log.Infof("Operator based Portworx deployment, Upgrading stork via StorageCluster")
// 		storageSpec, err := Inst().V.GetDriver()
// 		if err != nil {
// 			return err
// 		}
// 		storageSpec.Spec.Stork.Image = storkImageToUpgrade
// 		_, err = operator.Instance().UpdateStorageCluster(storageSpec)
// 		if err != nil {
// 			return err
// 		}
// 	} else {
// 		log.Infof("Non-Operator based Portworx deployment, Upgrading stork via Deployment")
// 		storkDeployment, err := apps.Instance().GetDeployment(storkDeploymentName, storkDeploymentNamespace)
// 		if err != nil {
// 			return err
// 		}

// 		storkDeployment.Spec.Template.Spec.Containers[0].Image = storkImageToUpgrade
// 		_, err = apps.Instance().UpdateDeployment(storkDeployment)
// 		if err != nil {
// 			return err
// 		}
// 	}

// 	// validate stork pods after upgrade
// 	updatedStorkDeployment, err := apps.Instance().GetDeployment(storkDeploymentName, storkDeploymentNamespace)
// 	if err != nil {
// 		return err
// 	}
// 	err = apps.Instance().ValidateDeployment(updatedStorkDeployment, k8s.DefaultTimeout, k8s.DefaultRetryInterval)
// 	if err != nil {
// 		return err
// 	}

// 	postUpgradeStorkImageVersionStr, err := getStorkImageVersion()
// 	if err != nil {
// 		return err
// 	}

// 	if !strings.EqualFold(postUpgradeStorkImageVersionStr, storkImageVersionToUpgradeStr) {
// 		return fmt.Errorf("expected version after upgrade was %s but got %s", storkImageVersionToUpgradeStr, postUpgradeStorkImageVersionStr)
// 	}

// 	log.Infof("Succesfully upgraded stork version from %v to %v", currentStorkImageStr, postUpgradeStorkImageVersionStr)
// 	return nil
// }

//
//
//
//

func isUserPresent(username string) (bool, error) {
	allUsers, err := backup.GetAllUsers()
	if err != nil {
		return false, err
	}
	for _, user := range allUsers {
		if user.Name == username {
			return true, nil
		}
	}
	return false, nil
}
