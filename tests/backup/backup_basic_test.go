package tests

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
	"github.com/portworx/torpedo/drivers"
	"github.com/portworx/torpedo/drivers/backup"
	"github.com/portworx/torpedo/drivers/backup/pxbackup"
	"github.com/portworx/torpedo/drivers/node"
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/drivers/scheduler/anthos"
	"github.com/portworx/torpedo/drivers/scheduler/dcos"
	"github.com/portworx/torpedo/drivers/scheduler/k8s"
	"github.com/portworx/torpedo/drivers/scheduler/openshift"
	"github.com/portworx/torpedo/drivers/scheduler/rke"
	"github.com/portworx/torpedo/drivers/volume"
	"github.com/portworx/torpedo/pkg/aetosutil"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests"
)

func getBucketNameSuffix() string {
	bucketNameSuffix, present := os.LookupEnv("BUCKET_NAME")
	if present {
		return bucketNameSuffix
	} else {
		return "default-suffix"
	}
}

func getGlobalBucketName(provider string) string {
	switch provider {
	case drivers.ProviderAws:
		return globalAWSBucketName
	case drivers.ProviderAzure:
		return globalAzureBucketName
	case drivers.ProviderGke:
		return globalGCPBucketName
	default:
		return globalAWSBucketName
	}
}

func getGlobalLockedBucketName(provider string) string {
	switch provider {
	case drivers.ProviderAws:
		return globalAWSLockedBucketName
	default:
		log.Errorf("environment variable [%s] not provided with valid values", "PROVIDERS")
		return ""
	}
}

func TestBasic(t *testing.T) {
	RegisterFailHandler(Fail)

	var specReporters []Reporter
	junitReporter := reporters.NewJUnitReporter("/testresults/junit_basic.xml")
	specReporters = append(specReporters, junitReporter)
	RunSpecsWithDefaultAndCustomReporters(t, "Torpedo : Backup", specReporters)
}

// BackupInitInstance initialises instances required for backup
func BackupInitInstance() {
	var err error

	// Initialization of Scheduler Driver
	schedulerOptions := scheduler.InitOptions{
		SpecDir:                    Inst().SpecDir,
		NodeDriverType:             Inst().N.String(),
		VolumeDriverName:           Inst().V.String(),
		UseGlobalSchedopsInstances: true,
	}
	err = Inst().S.Init(schedulerOptions)
	log.FailOnError(err, "Error occured while Scheduler Driver Initialization")

	// Initialization of Node Driver
	nodeOptions := node.InitOptions{
		SpecDir:          Inst().SpecDir,
		VolumeDriverName: Inst().V.String(),
	}
	if k8sScheduler, ok := Inst().S.(*k8s.K8s); ok {
		nodeOptions.NodeRegistry = k8sScheduler.NodeRegistry
		nodeOptions.K8sCore = k8sScheduler.K8sCore
		nodeOptions.K8sApps = k8sScheduler.K8sApps
	} else if rkeScheduler, ok := Inst().S.(*rke.Rke); ok {
		nodeOptions.NodeRegistry = rkeScheduler.NodeRegistry
		nodeOptions.K8sCore = rkeScheduler.K8sCore
		nodeOptions.K8sApps = rkeScheduler.K8sApps
	} else if dcosScheduler, ok := Inst().S.(*dcos.Dcos); ok {
		nodeOptions.NodeRegistry = dcosScheduler.NodeRegistry
	} else if anthosScheduler, ok := Inst().S.(*anthos.Anthos); ok {
		nodeOptions.NodeRegistry = anthosScheduler.NodeRegistry
		nodeOptions.K8sCore = anthosScheduler.K8sCore
		nodeOptions.K8sApps = anthosScheduler.K8sApps
	} else if openshiftScheduler, ok := Inst().S.(*openshift.Openshift); ok {
		nodeOptions.NodeRegistry = openshiftScheduler.NodeRegistry
		nodeOptions.K8sCore = openshiftScheduler.K8sCore
		nodeOptions.K8sApps = openshiftScheduler.K8sApps
	}
	err = Inst().N.Init(nodeOptions)
	log.FailOnError(err, "Error occured while Node Driver Initialization")

	// Initialization of Volume Driver
	volOptions := volume.InitOptions{
		NodeDriver:                Inst().N,
		SchedulerDriverName:       Inst().S.String(),
		StorageProvisionerType:    volume.StorageProvisionerType(Inst().ProvisionerType),
		CsiGenericDriverConfigMap: Inst().CsiGenericDriverConfigMap,
	}
	if k8sScheduler, ok := Inst().S.(*k8s.K8s); ok {
		volOptions.NodeRegistry = k8sScheduler.NodeRegistry
		volOptions.K8sApps = k8sScheduler.K8sApps
		volOptions.K8sAutopilot = k8sScheduler.K8sAutopilot
		volOptions.K8sBatch = k8sScheduler.K8sBatch
		volOptions.K8sRbac = k8sScheduler.K8sRbac
		volOptions.K8sApiExtensions = k8sScheduler.K8sApiExtensions
		volOptions.PxOperator = k8sScheduler.PxOperator
		volOptions.K8sCore = k8sScheduler.K8sCore
	} else if rkeScheduler, ok := Inst().S.(*rke.Rke); ok {
		volOptions.NodeRegistry = rkeScheduler.NodeRegistry
		volOptions.K8sApps = rkeScheduler.K8sApps
		volOptions.K8sAutopilot = rkeScheduler.K8sAutopilot
		volOptions.K8sBatch = rkeScheduler.K8sBatch
		volOptions.K8sRbac = rkeScheduler.K8sRbac
		volOptions.K8sApiExtensions = rkeScheduler.K8sApiExtensions
		volOptions.PxOperator = rkeScheduler.PxOperator
		volOptions.K8sCore = rkeScheduler.K8sCore
	} else if dcosScheduler, ok := Inst().S.(*dcos.Dcos); ok {
		volOptions.NodeRegistry = dcosScheduler.NodeRegistry
	} else if anthosScheduler, ok := Inst().S.(*anthos.Anthos); ok {
		volOptions.NodeRegistry = anthosScheduler.NodeRegistry
		volOptions.K8sApps = anthosScheduler.K8sApps
		volOptions.K8sAutopilot = anthosScheduler.K8sAutopilot
		volOptions.K8sBatch = anthosScheduler.K8sBatch
		volOptions.K8sRbac = anthosScheduler.K8sRbac
		volOptions.K8sApiExtensions = anthosScheduler.K8sApiExtensions
		volOptions.PxOperator = anthosScheduler.PxOperator
		volOptions.K8sCore = anthosScheduler.K8sCore
	} else if openshiftScheduler, ok := Inst().S.(*openshift.Openshift); ok {
		volOptions.NodeRegistry = openshiftScheduler.NodeRegistry
		volOptions.K8sApps = openshiftScheduler.K8sApps
		volOptions.K8sAutopilot = openshiftScheduler.K8sAutopilot
		volOptions.K8sBatch = openshiftScheduler.K8sBatch
		volOptions.K8sRbac = openshiftScheduler.K8sRbac
		volOptions.K8sApiExtensions = openshiftScheduler.K8sApiExtensions
		volOptions.PxOperator = openshiftScheduler.PxOperator
		volOptions.K8sCore = openshiftScheduler.K8sCore
	}
	err = Inst().V.Init(volOptions)
	log.FailOnError(err, "Error occured while Volume Driver Initialization")

	// finish setting up scheduler
	if k8sScheduler, ok := Inst().S.(*k8s.K8s); ok {
		k8sScheduler.NodeDriver = Inst().N
		k8sScheduler.VolumeDriver = Inst().V
	} else if rkeScheduler, ok := Inst().S.(*rke.Rke); ok {
		rkeScheduler.NodeDriver = Inst().N
		rkeScheduler.VolumeDriver = Inst().V
	} else if dcosScheduler, ok := Inst().S.(*dcos.Dcos); ok {
		dcosScheduler.VolumeDriver = Inst().V
	} else if anthosScheduler, ok := Inst().S.(*anthos.Anthos); ok {
		anthosScheduler.NodeDriver = Inst().N
		anthosScheduler.VolumeDriver = Inst().V
	} else if openshiftScheduler, ok := Inst().S.(*openshift.Openshift); ok {
		openshiftScheduler.NodeDriver = Inst().N
		openshiftScheduler.VolumeDriver = Inst().V
	}

	backupOptions := backup.InitOptions{}
	if k8sScheduler, ok := Inst().S.(*k8s.K8s); ok {
		backupOptions.K8sCore = k8sScheduler.K8sCore
	} else if rkeScheduler, ok := Inst().S.(*rke.Rke); ok {
		backupOptions.K8sCore = rkeScheduler.K8sCore
	} else if anthosScheduler, ok := Inst().S.(*anthos.Anthos); ok {
		backupOptions.K8sCore = anthosScheduler.K8sCore
	} else if openshiftScheduler, ok := Inst().S.(*openshift.Openshift); ok {
		backupOptions.K8sCore = openshiftScheduler.K8sCore
	}
	err = Inst().Backup.Init(backupOptions)
	log.FailOnError(err, "Error occured while Backup Driver Initialization")

	SetupTestRail()

	// Getting Px version info
	pxVersion, err := Inst().V.GetDriverVersion()
	log.FailOnError(err, "Error occurred while getting PX version")
	commitID := strings.Split(pxVersion, "-")[1]
	t := Inst().Dash.TestSet
	t.CommitID = commitID
	if pxVersion != "" {
		t.Tags["px-version"] = pxVersion
	}

	// Getting Px-Backup server version info and setting Aetos Dashboard tags
	PxBackupVersion, err = GetPxBackupVersionString()
	log.FailOnError(err, "Error getting Px Backup version")
	PxBackupBuildDate, err := GetPxBackupBuildDate()
	log.FailOnError(err, "Error getting Px Backup build date")
	t.Tags["px-backup-version"] = PxBackupVersion
	t.Tags["px-backup-build-date"] = PxBackupBuildDate

	Inst().Dash.TestSetUpdate(t)

	// Setting the common password
	pxCentralAdminPwd, err := Inst().Backup.(*pxbackup.PXBackup).GetPxCentralAdminPwd()
	log.FailOnError(err, "Error in pxbackup.GetPxCentralAdminPwd()")
	commonPassword = pxCentralAdminPwd + RandomString(4)

	// Dumping source and destination kubeconfig to file system path
	log.Infof("Dumping source and destination kubeconfig to file system path")
	kubeconfigs := os.Getenv("KUBECONFIGS")
	if kubeconfigs == "" {
		log.FailOnError(fmt.Errorf("Getting KUBECONFIGS Environment variable"), "")
	}
	kubeconfigList := strings.Split(kubeconfigs, ",")
	if len(kubeconfigList) != 2 {
		log.FailOnError(fmt.Errorf("2 kubeconfigs are required for source and destination cluster"), "")
	}
	// The way Backup tests use schedulers is:
	// 1. Inst().S - default scheduler is for PX-Backup
	// 2. Inst().SchedulerDrivers[kubeconfigsPaths[0]] - scheduler for source
	// 3. Inst().SchedulerDrivers[kubeconfigsPaths[1]] - scheduler for destination
	kubeconfigsPaths, err := InitTorpedoDriversForKubeconfigs(kubeconfigList)
	dash.VerifyFatal(err, nil, fmt.Sprintf("Initialization of drivers using kubeconfigs [%v]", kubeconfigList))
	KubeconfigsPaths[0] = kubeconfigsPaths[0]
	KubeconfigsPaths[1] = kubeconfigsPaths[1]
	// redundant variables (easy to use)
	SourceClusterConfigPath = kubeconfigsPaths[0]
	DestinationClusterConfigPath = kubeconfigsPaths[1]
}

var dash *aetosutil.Dashboard
var _ = BeforeSuite(func() {
	dash = Inst().Dash
	dash.TestSetBegin(dash.TestSet)
	log.Infof("Backup Init instance")
	BackupInitInstance()
	StartTorpedoTest("Setup buckets", "Creating one generic bucket to be used in all cases", nil, 0)
	defer EndTorpedoTest()
	// Create the first bucket from the list to be used as generic bucket
	providers := getProviders()
	bucketNameSuffix := getBucketNameSuffix()
	for _, provider := range providers {
		switch provider {
		case drivers.ProviderAws:
			globalAWSBucketName = fmt.Sprintf("%s-%s", globalAWSBucketPrefix, bucketNameSuffix)
			CreateBucket(provider, globalAWSBucketName)
			log.Infof("Bucket created with name - %s", globalAWSBucketName)
		case drivers.ProviderAzure:
			globalAzureBucketName = fmt.Sprintf("%s-%s", globalAzureBucketPrefix, bucketNameSuffix)
			CreateBucket(provider, globalAzureBucketName)
			log.Infof("Bucket created with name - %s", globalAzureBucketName)
		case drivers.ProviderGke:
			globalGCPBucketName = fmt.Sprintf("%s-%s", globalGCPBucketPrefix, bucketNameSuffix)
			CreateBucket(provider, globalGCPBucketName)
			log.Infof("Bucket created with name - %s", globalGCPBucketName)
		}
	}
	lockedBucketNameSuffix, present := os.LookupEnv("LOCKED_BUCKET_NAME")
	if present {
		for _, provider := range providers {
			switch provider {
			case drivers.ProviderAws:
				globalAWSLockedBucketName = fmt.Sprintf("%s-%s", globalAWSLockedBucketPrefix, lockedBucketNameSuffix)
			case drivers.ProviderAzure:
				globalAzureLockedBucketName = fmt.Sprintf("%s-%s", globalAzureLockedBucketPrefix, lockedBucketNameSuffix)
			case drivers.ProviderGke:
				globalGCPLockedBucketName = fmt.Sprintf("%s-%s", globalGCPLockedBucketPrefix, lockedBucketNameSuffix)
			}
		}
	} else {
		log.Infof("Locked bucket name not provided")
	}
})

var _ = AfterSuite(func() {

	StartTorpedoTest("Environment cleanup", "Removing Px-Backup entities created during the test execution", nil, 0)
	defer dash.TestSetEnd()
	defer EndTorpedoTest()

	// Cleanup all non admin users
	ctx, err := Inst().Backup.(*pxbackup.PXBackup).GetPxCentralAdminCtx()
	log.FailOnError(err, "Fetching px-central-admin ctx")
	allUsers, err := Inst().Backup.(*pxbackup.PXBackup).GetAllUsers()
	dash.VerifySafely(err, nil, "Verifying cleaning up of all users from keycloak")
	for _, user := range allUsers {
		if !strings.Contains(user.Name, "admin") {
			err = Inst().Backup.(*pxbackup.PXBackup).DeleteUser(user.Name)
			dash.VerifySafely(err, nil, fmt.Sprintf("Verifying user [%s] deletion", user.Name))
		} else {
			log.Infof("User %s was not deleted", user.Name)
		}
	}
	// Cleanup all non admin groups
	allGroups, err := Inst().Backup.(*pxbackup.PXBackup).GetAllUsers()
	dash.VerifySafely(err, nil, "Verifying cleaning up of all groups from keycloak")
	for _, group := range allGroups {
		if !strings.Contains(group.Name, "admin") && !strings.Contains(group.Name, "app") {
			err = Inst().Backup.(*pxbackup.PXBackup).DeleteGroup(group.Name)
			dash.VerifySafely(err, nil, fmt.Sprintf("Verifying group [%s] deletion", group.Name))
		} else {
			log.Infof("Group %s was not deleted", group.Name)
		}
	}

	// Cleanup all backups
	allBackups, err := GetAllBackupsAdmin()
	for _, backupName := range allBackups {
		backupUID, err := Inst().Backup.GetBackupUID(ctx, backupName, orgID)
		dash.VerifySafely(err, nil, fmt.Sprintf("Getting backuip UID for backup %s", backupName))
		_, err = DeleteBackup(backupName, backupUID, orgID, ctx)
		dash.VerifySafely(err, nil, fmt.Sprintf("Verifying backup deletion - %s", backupName))
	}

	// Cleanup all restores
	allRestores, err := GetAllRestoresAdmin()
	for _, restoreName := range allRestores {
		err = DeleteRestore(restoreName, orgID, ctx)
		dash.VerifySafely(err, nil, fmt.Sprintf("Verifying restore deletion - %s", restoreName))
	}

	// Cleanup all backup locations
	allBackupLocations, err := getAllBackupLocations(ctx)
	dash.VerifySafely(err, nil, "Verifying fetching of all backup locations")
	for backupLocationUid, backupLocationName := range allBackupLocations {
		err = DeleteBackupLocation(backupLocationName, backupLocationUid, orgID, true)
		dash.VerifySafely(err, nil, fmt.Sprintf("Verifying backup location deletion - %s", backupLocationName))
	}

	backupLocationDeletionSuccess := func() (interface{}, bool, error) {
		allBackupLocations, err := getAllBackupLocations(ctx)
		dash.VerifySafely(err, nil, "Verifying fetching of all backup locations")
		if len(allBackupLocations) > 0 {
			return "", true, fmt.Errorf("found %d backup locations", len(allBackupLocations))
		} else {
			return "", false, nil
		}
	}
	_, err = DoRetryWithTimeoutWithGinkgoRecover(backupLocationDeletionSuccess, 5*time.Minute, 30*time.Second)
	dash.VerifySafely(err, nil, "Verifying backup location deletion success")

	// Cleanup all cloud credentials
	allCloudCredentials, err := getAllCloudCredentials(ctx)
	dash.VerifySafely(err, nil, "Verifying fetching of all cloud credentials")
	for cloudCredentialUid, cloudCredentialName := range allCloudCredentials {
		err = DeleteCloudCredential(cloudCredentialName, orgID, cloudCredentialUid)
		dash.VerifySafely(err, nil, fmt.Sprintf("Deleting cloud cred %s", cloudCredentialName))
	}

	cloudCredentialDeletionSuccess := func() (interface{}, bool, error) {
		allCloudCredentials, err := getAllCloudCredentials(ctx)
		dash.VerifySafely(err, nil, "Verifying fetching of all cloud credentials")
		if len(allCloudCredentials) > 0 {
			return "", true, fmt.Errorf("found %d cloud credentials", len(allBackupLocations))
		} else {
			return "", false, nil
		}
	}
	_, err = DoRetryWithTimeoutWithGinkgoRecover(cloudCredentialDeletionSuccess, 5*time.Minute, 30*time.Second)
	dash.VerifySafely(err, nil, "Verifying backup location deletion success")

	// Cleanup all buckets after suite
	providers := getProviders()
	for _, provider := range providers {
		switch provider {
		case drivers.ProviderAws:
			DeleteBucket(provider, globalAWSBucketName)
			log.Infof("Bucket deleted - %s", globalAWSBucketName)
		case drivers.ProviderAzure:
			DeleteBucket(provider, globalAzureBucketName)
			log.Infof("Bucket deleted - %s", globalAzureBucketName)
		case drivers.ProviderGke:
			DeleteBucket(provider, globalGCPBucketName)
			log.Infof("Bucket deleted - %s", globalGCPBucketName)
		}
	}

})

func TestMain(m *testing.M) {
	// call flag.Parse() here if TestMain uses flags
	ParseFlags()
	os.Exit(m.Run())
}
