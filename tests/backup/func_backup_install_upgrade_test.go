package tests

import (
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	"github.com/pborman/uuid"
	api "github.com/portworx/px-backup-api/pkg/apis/v1"
	"github.com/portworx/sched-ops/k8s/core"
	"github.com/pure-px/torpedo/drivers/backup"
	"github.com/pure-px/torpedo/drivers/scheduler"
	"github.com/pure-px/torpedo/pkg/log"
	. "github.com/pure-px/torpedo/tests"
	"helm.sh/helm/v3/pkg/release"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubevirtv1 "kubevirt.io/api/core/v1"
	"strings"
	"time"
)

// Test case to check resource quota for namespace
var _ = Describe("{InstallPxBackupResourceQuotaCheck}", Label(TestCaseLabelsMap[InstallPxBackupResourceQuotaCheck]...), func() {
	var (
		namespace string
	)
	JustBeforeEach(func() {
		StartPxBackupTorpedoTest("InstallPxBackupResourceQuotaCheck", "Installing Px Backup and checking if resource quota satisfies the min requirement", nil, 304868, Mkoppal, Q4FY25)
		log.InfoD("Uninstalling Px-Backup...")
		err := UninstallPxBackup()
		log.FailOnError(err, "Failed to delete px-backup")
		namespace = "px-backup"
	})
	It("Should check resource quota for namespace", func() {

		Step("Create namespace with insufficient resource quota", func() {
			log.InfoD("Create namespace with insufficient resource quota")
			err := CreateNamespaceAndResourceQuota(namespace, "300Gi")
			log.FailOnError(err, "Failed to create namespace and resource quota")
		})

		Step("Install Px-Backup and validate the failure message", func() {
			log.InfoD("Install Px-Backup and validate the failure message")
			valuesMap, err := ParseValuesFromFile("values1")
			log.FailOnError(err, "Failed to parse values file")
			_, err = InstallPxBackup(LatestPxBackupVersion, DefaultPxBackupHelmBranch, namespace, valuesMap)
			if err != nil {
				log.InfoD("Install Px-Backup error - %s", err.Error())
				dash.VerifyFatal(strings.Contains(err.Error(), "px-backup chart installation failed"), true, "Expecting installation to fail")
			} else {
				log.FailOnError(fmt.Errorf("installation was successful when it should not have been"),
					"Installation was successful when it should not have been")
			}
		})

		Step("Validate the pre-install job pod logs", func() {
			log.InfoD("Validate the pre-install job pod logs")
			searchString := "Namespace px-backup does not meet the storage quota requirements for Px-Backup installation"
			present, err := SearchPodLogs(map[string]string{"job-name": "pre-install-check"}, namespace, searchString)
			log.FailOnError(err, "Failed to search pod logs")
			dash.VerifyFatal(present, true, "Found the expected string in the pod logs")
		})

		Step("Delete namespace and resource quota", func() {
			log.InfoD("Delete namespace and resource quota")
			err := DeleteAppNamespace(namespace)
			log.FailOnError(err, "Failed to delete namespace and resource quota")
		})

		Step("Create namespace with sufficient resource quota", func() {
			log.InfoD("Create namespace with sufficient resource quota")
			err := CreateNamespaceAndResourceQuota(namespace, "600Gi")
			log.FailOnError(err, "Failed to create namespace and resource quota")
		})

		Step("Install Px-Backup and validate the successful installation", func() {
			log.InfoD("Install Px-Backup and validate the successful installation")
			valuesMap, err := ParseValuesFromFile("values1")
			log.FailOnError(err, "Failed to parse values file")
			_, err = InstallPxBackup(LatestPxBackupVersion, DefaultPxBackupHelmBranch, namespace, valuesMap)
			log.FailOnError(err, "Failed to install px-backup")
		})
	})

	JustAfterEach(func() {
		defer EndPxBackupTorpedoTest(make([]*scheduler.Context, 0))
		log.Infof("No cleanup required for this testcase")
	})
})

// Test case to check if there is a string in the helm notes
var _ = Describe("{InstallPxBackupHelmNotesCheck}", Label(TestCaseLabelsMap[InstallPxBackupHelmNotesCheck]...), func() {
	var (
		namespace string
	)
	rel := &release.Release{}
	JustBeforeEach(func() {
		StartPxBackupTorpedoTest("InstallPxBackupHelmNotesCheck",
			"Installing Px Backup and checking if helm notes contains the desired string", nil, 304870, Mkoppal, Q4FY25)
		log.InfoD("Uninstalling Px-Backup...")
		err := UninstallPxBackup()
		log.FailOnError(err, "Failed to delete px-backup")
		namespace = "px-backup"
	})
	It("Should check if there is a string in the helm notes", func() {

		Step("Install Px-Backup and validate the helm notes", func() {
			log.InfoD("Install Px-Backup and validate the helm notes")
			valuesMap, err := ParseValuesFromFile("values1")
			log.FailOnError(err, "Failed to parse values file")
			rel, err = InstallPxBackup(LatestPxBackupVersion, DefaultPxBackupHelmBranch, namespace, valuesMap)
			log.FailOnError(err, "Failed to install px-backup")
			log.InfoD("Helm notes - \n%s", rel.Info.Notes)
		})

		Step("Validate the helm notes", func() {
			log.InfoD("Validate the helm notes")
			searchString := "For more information on network pre-requisites: " +
				"https://docs.portworx.com/portworx-backup-on-prem/install/install-prereq/nw-prereqs.html"
			dash.VerifyFatal(strings.Contains(rel.Info.Notes, searchString), true,
				"Found the expected string in the helm notes")
		})
	})
	JustAfterEach(func() {
		defer EndPxBackupTorpedoTest(make([]*scheduler.Context, 0))
		log.Infof("No cleanup required for this testcase")
	})
})

// Test case to verify that the installation fails with a PVC in pre-existing in the namespace where the installation is being attempted
var _ = Describe("{InstallPxBackupPvcCheck}", Label(TestCaseLabelsMap[InstallPxBackupPvcCheck]...), func() {

	var (
		namespace string
	)
	JustBeforeEach(func() {
		StartPxBackupTorpedoTest("InstallPxBackupPvcCheck",
			"Installing Px Backup and checking if installation fails when a PVC is pre-existing in the namespace", nil, 304869, Mkoppal, Q4FY25)
		log.InfoD("Uninstalling Px-Backup...")
		err := UninstallPxBackup()
		log.FailOnError(err, "Failed to delete px-backup")
		namespace = "px-backup"
	})
	It("Should check if installation fails when a PVC is pre-existing in the namespace", func() {

		Step("Create namespace and PVC", func() {
			log.InfoD("Create namespace and PVC")
			// Create the Namespace object
			ns := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: namespace,
				},
			}
			// Create Namespace
			_, err := core.Instance().CreateNamespace(ns)
			log.FailOnError(err, "Failed to create namespace")
			sc, err := GetDefaultStorageClass()
			log.Infof("Using storage class - %s", sc.Name)

			// Get a random px-backup associated PVC
			pxbPvc, err := GetRandomSubset(PxBackupPVCs, 1)
			log.FailOnError(err, "Failed to get PVC")
			if len(pxbPvc) == 0 {
				log.FailOnError(fmt.Errorf("no PVCs found"), "No PVCs found")
			}

			// Create PVC
			log.Infof("Creating PVC - %s", pxbPvc[0])
			pvc := &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      pxbPvc[0],
					Namespace: namespace,
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					StorageClassName: &sc.Name,
					AccessModes:      []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceStorage: resource.MustParse("1Gi"),
						},
					},
				},
			}
			_, err = core.Instance().CreatePersistentVolumeClaim(pvc)
			log.FailOnError(err, "Failed to create PVC")
		})

		Step("Install Px-Backup and validate the failure message", func() {
			log.InfoD("Install Px-Backup and validate the failure message")
			valuesMap, err := ParseValuesFromFile("values1")
			log.FailOnError(err, "Failed to parse values file")
			_, err = InstallPxBackup(LatestPxBackupVersion, DefaultPxBackupHelmBranch, namespace, valuesMap)
			if err != nil {
				log.InfoD("Install Px-Backup error - %s", err.Error())
				dash.VerifyFatal(strings.Contains(err.Error(), "px-backup chart installation failed"),
					true, "Expecting installation to fail")
			} else {
				log.FailOnError(fmt.Errorf("installation was successful when it should not have been"),
					"Installation was successful when it should not have been")
			}
		})

		Step("Validate the pre-install job pod logs", func() {
			log.InfoD("Validate the pre-install job pod logs")
			searchString := fmt.Sprintf("Namespace contains %d stale PX-Backup PVC(s).", 1)
			present, err := SearchPodLogs(map[string]string{"job-name": "pre-install-check"}, namespace, searchString)
			log.FailOnError(err, "Failed to search pod logs")
			dash.VerifyFatal(present, true, "Found the expected string in the pod logs")
		})

	})
	JustAfterEach(func() {
		defer EndPxBackupTorpedoTest(make([]*scheduler.Context, 0))
		log.Infof("No cleanup required for this testcase")
	})
})

// Test case to check px-backup install failure due to incorrect storage class and images
var _ = Describe("{InstallPxBackupStorageClassAndImagesCheck}", Label(TestCaseLabelsMap[InstallPxBackupStorageClassAndImagesCheck]...), func() {
	var (
		backupNames          []string
		restoreNames         []string
		scheduledAppContexts []*scheduler.Context
		preRuleNameList      []string
		postRuleNameList     []string
		sourceClusterUid     string
		destClusterUid       string
		cloudCredName        string
		cloudCredUID         string
		backupLocationUID    string
		backupLocationName   string
		appList              []string
		backupLocationMap    map[string]string
		labelSelectors       map[string]string
		providers            []string
		intervalName         string
		dailyName            string
		weeklyName           string
		monthlyName          string
		allVirtualMachines   []kubevirtv1.VirtualMachine
		namespace            string
	)
	JustBeforeEach(func() {
		StartPxBackupTorpedoTest("InstallPxBackupStorageClassAndImagesCheck", "Installing Px Backup and checking if storage class and images are validated", nil, 304879, Mkoppal, Q4FY25)
		appList = Inst().AppList
		backupLocationMap = make(map[string]string)
		labelSelectors = make(map[string]string)
		providers = GetBackupProviders()
		intervalName = fmt.Sprintf("%s-%v", "interval", time.Now().Unix())
		dailyName = fmt.Sprintf("%s-%v", "daily", time.Now().Unix())
		weeklyName = fmt.Sprintf("%s-%v", "weekly", time.Now().Unix())
		monthlyName = fmt.Sprintf("%s-%v", "monthly", time.Now().Unix())
		log.InfoD("Uninstalling Px-Backup...")
		err := UninstallPxBackup()
		log.FailOnError(err, "Failed to delete px-backup")
		namespace = "px-backup"
	})
	It("Should check px-backup install failure due to incorrect storage class and images", func() {

		Step("Install Px-Backup with incorrect images", func() {
			log.InfoD("Install Px-Backup with incorrect images")
			valuesMap, err := ParseValuesFromFile("values-incorrect-image")
			log.FailOnError(err, "Failed to parse values file")
			_, err = InstallPxBackup(LatestPxBackupVersion, DefaultPxBackupHelmBranch, namespace, valuesMap)
			if err != nil {
				log.InfoD("Install Px-Backup error - %s", err.Error())
				dash.VerifyFatal(strings.Contains(err.Error(), "px-backup chart installation failed"), true, "Expecting installation to fail")
			} else {
				log.FailOnError(fmt.Errorf("installation was successful when it should not have been"),
					"Installation was successful when it should not have been")
			}
		})

		Step("Validate the preflight job pod logs", func() {
			log.InfoD("Validate the preflight job pod logs")
			searchString := "Preflight Check failed: failed to validate images: ValidateImages: image validation failed"
			present, err := SearchPodLogs(map[string]string{"job-name": "preflight-check"}, namespace, searchString)
			log.FailOnError(err, "Failed to search pod logs")
			dash.VerifyFatal(present, true, "Found the expected string in the pod logs")
		})

		Step("Delete px-backup namespace", func() {
			log.InfoD("Delete px-backup namespace")
			err := DeleteAppNamespace(namespace)
			log.FailOnError(err, "Failed to delete namespace")
		})

		Step("Install Px-Backup with incorrect storage class", func() {
			log.InfoD("Install Px-Backup with incorrect storage class")
			valuesMap, err := ParseValuesFromFile("values-incorrect-sc")
			log.FailOnError(err, "Failed to parse values file")
			_, err = InstallPxBackup(LatestPxBackupVersion, DefaultPxBackupHelmBranch, namespace, valuesMap)
			if err != nil {
				log.InfoD("Install Px-Backup error - %s", err.Error())
				dash.VerifyFatal(strings.Contains(err.Error(), "px-backup chart installation failed"), true, "Expecting installation to fail")
			} else {
				log.FailOnError(fmt.Errorf("installation was successful when it should not have been"),
					"Installation was successful when it should not have been")
			}
		})

		Step("Validate the preflight job pod logs", func() {
			log.InfoD("Validate the preflight job pod logs")
			searchString := "Preflight Check failed: storage class check failed: failed to get storage class"
			present, err := SearchPodLogs(map[string]string{"job-name": "preflight-check"}, namespace, searchString)
			log.FailOnError(err, "Failed to search pod logs")
			dash.VerifyFatal(present, true, "Found the expected string in the pod logs")
		})

		Step("Delete px-backup namespace", func() {
			log.InfoD("Delete px-backup namespace")
			err := DeleteAppNamespace(namespace)
			log.FailOnError(err, "Failed to delete namespace")
		})

		Step("Install Px-Backup with correct storage class and images", func() {
			log.InfoD("Install Px-Backup with correct storage class and images")
			valuesMap, err := ParseValuesFromFile("values1")
			log.FailOnError(err, "Failed to parse values file")
			_, err = InstallPxBackup(LatestPxBackupVersion, DefaultPxBackupHelmBranch, namespace, valuesMap)
			log.FailOnError(err, "Failed to install px-backup")
		})

		Step("Scheduling applications", func() {
			log.InfoD("scheduling applications")
			scheduledAppContexts = make([]*scheduler.Context, 0)
			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				taskName := fmt.Sprintf("%s-%d", TaskNamePrefix, i)
				appContexts := ScheduleApplications(taskName)
				for _, appCtx := range appContexts {
					appCtx.ReadinessTimeout = AppReadinessTimeout
					scheduledAppContexts = append(scheduledAppContexts, appCtx)
				}
			}
		})

		Step("Validating applications", func() {
			log.InfoD("Validating applications")
			ValidateApplications(scheduledAppContexts)
		})

		Step("Creating rules for backup", func() {
			log.InfoD("Creating rules for backup")
			log.InfoD("Creating pre rule for deployed apps")
			for i := 0; i < len(appList); i++ {
				preRuleStatus, ruleName, err := Inst().Backup.CreateRuleForBackup(appList[i], BackupOrgID, "pre")
				log.FailOnError(err, "Creating pre rule for deployed app [%s] failed", appList[i])
				dash.VerifyFatal(preRuleStatus, true, "Verifying pre rule for backup")
				if ruleName != "" {
					preRuleNameList = append(preRuleNameList, ruleName)
				}
			}
			log.InfoD("Creating post rule for deployed apps")
			for i := 0; i < len(appList); i++ {
				postRuleStatus, ruleName, err := Inst().Backup.CreateRuleForBackup(appList[i], BackupOrgID, "post")
				log.FailOnError(err, "Creating post rule for deployed app [%s] failed", appList[i])
				dash.VerifyFatal(postRuleStatus, true, "Verifying Post rule for backup")
				if ruleName != "" {
					postRuleNameList = append(postRuleNameList, ruleName)
				}
			}
		})

		Step("Creating backup location and cloud setting", func() {
			log.InfoD("Creating backup location and cloud setting")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			for _, provider := range providers {
				cloudCredName = fmt.Sprintf("%s-%s-%v", "cred", provider, time.Now().Unix())
				backupLocationName = fmt.Sprintf("%s-%s-bl-%v", provider, getGlobalBucketName(provider), time.Now().Unix())
				cloudCredUID = uuid.New()
				backupLocationUID = uuid.New()
				backupLocationMap[backupLocationUID] = backupLocationName
				err := CreateCloudCredential(provider, cloudCredName, cloudCredUID, BackupOrgID, ctx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of cloud credential named [%s] for org [%s] with [%s] as provider", cloudCredName, BackupOrgID, provider))
				err = CreateBackupLocation(provider, backupLocationName, backupLocationUID, cloudCredName, cloudCredUID, getGlobalBucketName(provider), BackupOrgID, "", true)
				dash.VerifyFatal(err, nil, "Creating backup location")
			}
		})

		Step("Creating backup schedule policies", func() {
			log.InfoD("Creating backup schedule policies")
			log.InfoD("Creating backup interval schedule policy")
			intervalSchedulePolicyInfo := Inst().Backup.CreateIntervalSchedulePolicy(5, 15, 2)
			intervalPolicyStatus := Inst().Backup.BackupSchedulePolicy(intervalName, uuid.New(), BackupOrgID, intervalSchedulePolicyInfo)
			dash.VerifyFatal(intervalPolicyStatus, nil, "Creating interval schedule policy")

			log.InfoD("Creating backup daily schedule policy")
			dailySchedulePolicyInfo := Inst().Backup.CreateDailySchedulePolicy(1, "9:00AM", 2)
			dailyPolicyStatus := Inst().Backup.BackupSchedulePolicy(dailyName, uuid.New(), BackupOrgID, dailySchedulePolicyInfo)
			dash.VerifyFatal(dailyPolicyStatus, nil, "Creating daily schedule policy")

			log.InfoD("Creating backup weekly schedule policy")
			weeklySchedulePolicyInfo := Inst().Backup.CreateWeeklySchedulePolicy(1, backup.Friday, "9:10AM", 2)
			weeklyPolicyStatus := Inst().Backup.BackupSchedulePolicy(weeklyName, uuid.New(), BackupOrgID, weeklySchedulePolicyInfo)
			dash.VerifyFatal(weeklyPolicyStatus, nil, "Creating weekly schedule policy")

			log.InfoD("Creating backup monthly schedule policy")
			monthlySchedulePolicyInfo := Inst().Backup.CreateMonthlySchedulePolicy(1, 29, "9:20AM", 2)
			monthlyPolicyStatus := Inst().Backup.BackupSchedulePolicy(monthlyName, uuid.New(), BackupOrgID, monthlySchedulePolicyInfo)
			dash.VerifyFatal(monthlyPolicyStatus, nil, "Creating monthly schedule policy")
		})

		Step("Registering cluster for backup", func() {
			log.InfoD("Registering cluster for backup")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")

			err = CreateApplicationClusters(BackupOrgID, "", "", ctx)
			dash.VerifyFatal(err, nil, "Creating source and destination cluster")

			clusterStatus, err := Inst().Backup.GetClusterStatus(BackupOrgID, SourceClusterName, ctx)
			log.FailOnError(err, fmt.Sprintf("Fetching [%s] cluster status", SourceClusterName))
			dash.VerifyFatal(clusterStatus, api.ClusterInfo_StatusInfo_Online, fmt.Sprintf("Verifying if [%s] cluster is online", SourceClusterName))

			sourceClusterUid, err = Inst().Backup.GetClusterUID(ctx, BackupOrgID, SourceClusterName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", SourceClusterName))

			clusterStatus, err = Inst().Backup.GetClusterStatus(BackupOrgID, DestinationClusterName, ctx)
			log.FailOnError(err, fmt.Sprintf("Fetching [%s] cluster status", DestinationClusterName))
			dash.VerifyFatal(clusterStatus, api.ClusterInfo_StatusInfo_Online, fmt.Sprintf("Verifying if [%s] cluster is online", DestinationClusterName))

			destClusterUid, err = Inst().Backup.GetClusterUID(ctx, BackupOrgID, DestinationClusterName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", DestinationClusterName))
		})

		Step("Taking backup of application from source cluster", func() {
			log.InfoD("taking backup of applications")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")

			backupNames = make([]string, 0)
			for i, appCtx := range scheduledAppContexts {
				scheduledNamespace := appCtx.ScheduleOptions.Namespace
				backupName := fmt.Sprintf("%s-%s-%v", "autogenerated-backup", scheduledNamespace, time.Now().Unix())
				log.InfoD("creating backup [%s] in source cluster [%s] (%s), organization [%s], of namespace [%s], in backup location [%s]", backupName, SourceClusterName, sourceClusterUid, BackupOrgID, scheduledNamespace, backupLocationName)
				err := CreateBackupWithValidation(ctx, backupName, SourceClusterName, backupLocationName, backupLocationUID, scheduledAppContexts[i:i+1], labelSelectors, BackupOrgID, sourceClusterUid, "", "", "", "")
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creation and Validation of backup [%s]", backupName))
				backupNames = append(backupNames, backupName)
				if len(allVirtualMachines) > 0 {
					backupName := fmt.Sprintf("%s-%s-%v", "autogenerated-backup", scheduledNamespace, time.Now().Unix())
					err = CreateVMBackupWithValidation(ctx, backupName, allVirtualMachines, SourceClusterName, backupLocationName, backupLocationUID, scheduledAppContexts,
						labelSelectors, BackupOrgID, sourceClusterUid, "", "", "", "", true)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Creation of VM backup and Validation of backup [%s]", backupName))
					backupNames = append(backupNames, backupName)
				}
			}
		})

		Step("Restoring the backed up namespaces", func() {
			log.InfoD("Restoring the backed up namespaces")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			for i, appCtx := range scheduledAppContexts {
				scheduledNamespace := appCtx.ScheduleOptions.Namespace
				restoreName := fmt.Sprintf("%s-%s-%s", "test-restore", scheduledNamespace, RandomString(4))
				for strings.Contains(strings.Join(restoreNames, ","), restoreName) {
					restoreName = fmt.Sprintf("%s-%s-%s", "test-restore", scheduledNamespace, RandomString(4))
				}
				log.InfoD("Restoring [%s] namespace from the [%s] backup", scheduledNamespace, backupNames[i])
				err = CreateRestoreWithValidation(ctx, restoreName, backupNames[i], make(map[string]string), make(map[string]string), DestinationClusterName, destClusterUid, BackupOrgID, scheduledAppContexts[i:i+1])
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creation and Validation of restore [%s]", restoreName))
				restoreNames = append(restoreNames, restoreName)
			}
		})
	})
	JustAfterEach(func() {
		defer EndPxBackupTorpedoTest(make([]*scheduler.Context, 0))
		defer func() {
			log.InfoD("switching to default context")
			err := SetClusterContext("")
			log.FailOnError(err, "failed to SetClusterContext to default cluster")
		}()

		policyList := []string{intervalName, dailyName, weeklyName, monthlyName}
		ctx, err := backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")
		if len(preRuleNameList) > 0 {
			for _, ruleName := range preRuleNameList {
				err := Inst().Backup.DeleteRuleForBackup(BackupOrgID, ruleName)
				dash.VerifySafely(err, nil, fmt.Sprintf("Deleting backup pre rules [%s]", ruleName))
			}
		}
		if len(postRuleNameList) > 0 {
			for _, ruleName := range postRuleNameList {
				err := Inst().Backup.DeleteRuleForBackup(BackupOrgID, ruleName)
				dash.VerifySafely(err, nil, fmt.Sprintf("Deleting backup post rules [%s]", ruleName))
			}
		}
		err = Inst().Backup.DeleteBackupSchedulePolicy(BackupOrgID, policyList)
		dash.VerifySafely(err, nil, "Deleting backup schedule policies")
		opts := make(map[string]bool)
		opts[SkipClusterScopedObjects] = true

		log.Info("Destroying scheduled apps on source cluster")
		DestroyApps(scheduledAppContexts, opts)
		log.FailOnError(err, "Data validations failed")

		log.InfoD("switching to destination context")
		err = SetDestinationKubeConfig()
		log.FailOnError(err, "failed to switch to context to destination cluster")

		log.InfoD("Destroying restored apps on destination clusters")
		restoredAppContexts := make([]*scheduler.Context, 0)
		for _, scheduledAppContext := range scheduledAppContexts {
			restoredAppContext, err := CloneAppContextAndTransformWithMappings(scheduledAppContext, make(map[string]string), make(map[string]string), true)
			if err != nil {
				log.Errorf("TransformAppContextWithMappings: %v", err)
				continue
			}
			restoredAppContexts = append(restoredAppContexts, restoredAppContext)
		}
		DestroyApps(restoredAppContexts, opts)

		log.InfoD("switching to default context")
		err = SetClusterContext("")
		log.FailOnError(err, "failed to SetClusterContext to default cluster")

		backupDriver := Inst().Backup
		log.Info("Deleting backed up namespaces")
		for _, backupName := range backupNames {
			backupUID, err := backupDriver.GetBackupUID(ctx, backupName, BackupOrgID)
			log.FailOnError(err, "Failed while trying to get backup UID for - %s", backupName)
			backupDeleteResponse, err := DeleteBackup(backupName, backupUID, BackupOrgID, ctx)
			log.FailOnError(err, "Backup [%s] could not be deleted", backupName)
			dash.VerifyFatal(backupDeleteResponse.String(), "", fmt.Sprintf("Verifying [%s] backup deletion is successful", backupName))
		}
		log.Info("Deleting restored namespaces")
		for _, restoreName := range restoreNames {
			err = DeleteRestore(restoreName, BackupOrgID, ctx)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Deleting Restore [%s]", restoreName))
		}
		CleanupCloudSettingsAndClusters(backupLocationMap, cloudCredName, cloudCredUID, ctx)
	})
})

// Test case to check px-backup upgrade failure due to incorrect storage class and images
var _ = Describe("{UpgradePxBackupStorageClassAndImagesCheck}", Label(TestCaseLabelsMap[UpgradePxBackupStorageClassAndImagesCheck]...), func() {
	var (
		backupNames          []string
		restoreNames         []string
		scheduledAppContexts []*scheduler.Context
		preRuleNameList      []string
		postRuleNameList     []string
		sourceClusterUid     string
		destClusterUid       string
		cloudCredName        string
		cloudCredUID         string
		backupLocationUID    string
		backupLocationName   string
		appList              []string
		backupLocationMap    map[string]string
		labelSelectors       map[string]string
		providers            []string
		intervalName         string
		dailyName            string
		weeklyName           string
		monthlyName          string
		allVirtualMachines   []kubevirtv1.VirtualMachine
		namespace            string
	)
	JustBeforeEach(func() {
		StartPxBackupTorpedoTest("UpgradePxBackupStorageClassAndImagesCheck", "Installing Px Backup and checking if storage class and images are validated", nil, 304880, Mkoppal, Q4FY25)
		appList = Inst().AppList
		namespace = "px-backup"
		backupLocationMap = make(map[string]string)
		labelSelectors = make(map[string]string)
		providers = GetBackupProviders()
		intervalName = fmt.Sprintf("%s-%v", "interval", time.Now().Unix())
		dailyName = fmt.Sprintf("%s-%v", "daily", time.Now().Unix())
		weeklyName = fmt.Sprintf("%s-%v", "weekly", time.Now().Unix())
		monthlyName = fmt.Sprintf("%s-%v", "monthly", time.Now().Unix())
		log.InfoD("Uninstalling Px-Backup...")
		err := UninstallPxBackup()
		log.FailOnError(err, "Failed to delete px-backup")
		log.InfoD("Installing Px-Backup...")
		valuesMap, err := ParseValuesFromFile("values1")
		log.FailOnError(err, "Failed to parse values file")
		_, err = InstallPxBackup("2.6.0", "master", namespace, valuesMap)
		log.FailOnError(err, "Failed to install px-backup")
	})
	It("Should check px-backup upgrade failure due to incorrect storage class and images", func() {

		Step("Upgrade Px-Backup with incorrect images", func() {
			log.InfoD("Upgrade Px-Backup and validate the helm notes")
			valuesMap, err := ParseValuesFromFile("values-incorrect-image")
			log.FailOnError(err, "Failed to parse values file")
			_, err = HelmUpgradePxBackup(LatestPxBackupVersion, DefaultPxBackupHelmBranch, namespace, valuesMap)
			if err != nil {
				log.InfoD("Upgrade Px-Backup error - %s", err.Error())
				dash.VerifyFatal(strings.Contains(err.Error(), "pre-upgrade hooks failed"), true, "Expecting upgrade to fail")
			} else {
				log.FailOnError(fmt.Errorf("upgrade was successful when it should not have been"),
					"Upgrade was successful when it should not have been")
			}
		})

		Step("Validate the preflight job pod logs", func() {
			log.InfoD("Validate the preflight job pod logs")
			searchString := "Preflight Check failed: failed to validate images: ValidateImages: image validation failed"
			present, err := SearchPodLogs(map[string]string{"job-name": "preflight-check"}, namespace, searchString)
			log.FailOnError(err, "Failed to search pod logs")
			dash.VerifyFatal(present, true, "Found the expected string in the pod logs")
		})

		Step("Upgrade Px-Backup with incorrect storage class", func() {
			log.InfoD("Upgrade Px-Backup with incorrect storage class")
			valuesMap, err := ParseValuesFromFile("values-incorrect-sc")
			log.FailOnError(err, "Failed to parse values file")
			_, err = HelmUpgradePxBackup(LatestPxBackupVersion, DefaultPxBackupHelmBranch, namespace, valuesMap)
			if err != nil {
				log.InfoD("Upgrade Px-Backup error - %s", err.Error())
				dash.VerifyFatal(strings.Contains(err.Error(), "pre-upgrade hooks failed"), true, "Expecting upgrade to fail")
			} else {
				log.FailOnError(fmt.Errorf("upgrade was successful when it should not have been"),
					"Upgrade was successful when it should not have been")
			}
		})

		Step("Validate the preflight job pod logs", func() {
			log.InfoD("Validate the preflight job pod logs")
			searchString := "Preflight Check failed: storage class check failed: failed to get storage class"
			present, err := SearchPodLogs(map[string]string{"job-name": "preflight-check"}, namespace, searchString)
			log.FailOnError(err, "Failed to search pod logs")
			dash.VerifyFatal(present, true, "Found the expected string in the pod logs")
		})

		Step("Upgrade Px-Backup with correct storage class and images", func() {
			log.InfoD("Upgrade Px-Backup with correct storage class and images")
			valuesMap, err := ParseValuesFromFile("values-correct-images-sc")
			log.FailOnError(err, "Failed to parse values file")
			_, err = HelmUpgradePxBackup(LatestPxBackupVersion, DefaultPxBackupHelmBranch, namespace, valuesMap)
			log.FailOnError(err, "Failed to upgrade px-backup")
		})

		Step("Scheduling applications", func() {
			log.InfoD("scheduling applications")
			scheduledAppContexts = make([]*scheduler.Context, 0)
			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				taskName := fmt.Sprintf("%s-%d", TaskNamePrefix, i)
				appContexts := ScheduleApplications(taskName)
				for _, appCtx := range appContexts {
					appCtx.ReadinessTimeout = AppReadinessTimeout
					scheduledAppContexts = append(scheduledAppContexts, appCtx)
				}
			}
		})

		Step("Validating applications", func() {
			log.InfoD("Validating applications")
			ValidateApplications(scheduledAppContexts)
		})

		Step("Creating rules for backup", func() {
			log.InfoD("Creating rules for backup")
			log.InfoD("Creating pre rule for deployed apps")
			for i := 0; i < len(appList); i++ {
				preRuleStatus, ruleName, err := Inst().Backup.CreateRuleForBackup(appList[i], BackupOrgID, "pre")
				log.FailOnError(err, "Creating pre rule for deployed app [%s] failed", appList[i])
				dash.VerifyFatal(preRuleStatus, true, "Verifying pre rule for backup")
				if ruleName != "" {
					preRuleNameList = append(preRuleNameList, ruleName)
				}
			}
			log.InfoD("Creating post rule for deployed apps")
			for i := 0; i < len(appList); i++ {
				postRuleStatus, ruleName, err := Inst().Backup.CreateRuleForBackup(appList[i], BackupOrgID, "post")
				log.FailOnError(err, "Creating post rule for deployed app [%s] failed", appList[i])
				dash.VerifyFatal(postRuleStatus, true, "Verifying Post rule for backup")
				if ruleName != "" {
					postRuleNameList = append(postRuleNameList, ruleName)
				}
			}
		})

		Step("Creating backup location and cloud setting", func() {
			log.InfoD("Creating backup location and cloud setting")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			for _, provider := range providers {
				cloudCredName = fmt.Sprintf("%s-%s-%v", "cred", provider, time.Now().Unix())
				backupLocationName = fmt.Sprintf("%s-%s-bl-%v", provider, getGlobalBucketName(provider), time.Now().Unix())
				cloudCredUID = uuid.New()
				backupLocationUID = uuid.New()
				backupLocationMap[backupLocationUID] = backupLocationName
				err := CreateCloudCredential(provider, cloudCredName, cloudCredUID, BackupOrgID, ctx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of cloud credential named [%s] for org [%s] with [%s] as provider", cloudCredName, BackupOrgID, provider))
				err = CreateBackupLocation(provider, backupLocationName, backupLocationUID, cloudCredName, cloudCredUID, getGlobalBucketName(provider), BackupOrgID, "", true)
				dash.VerifyFatal(err, nil, "Creating backup location")
			}
		})

		Step("Creating backup schedule policies", func() {
			log.InfoD("Creating backup schedule policies")
			log.InfoD("Creating backup interval schedule policy")
			intervalSchedulePolicyInfo := Inst().Backup.CreateIntervalSchedulePolicy(5, 15, 2)
			intervalPolicyStatus := Inst().Backup.BackupSchedulePolicy(intervalName, uuid.New(), BackupOrgID, intervalSchedulePolicyInfo)
			dash.VerifyFatal(intervalPolicyStatus, nil, "Creating interval schedule policy")

			log.InfoD("Creating backup daily schedule policy")
			dailySchedulePolicyInfo := Inst().Backup.CreateDailySchedulePolicy(1, "9:00AM", 2)
			dailyPolicyStatus := Inst().Backup.BackupSchedulePolicy(dailyName, uuid.New(), BackupOrgID, dailySchedulePolicyInfo)
			dash.VerifyFatal(dailyPolicyStatus, nil, "Creating daily schedule policy")

			log.InfoD("Creating backup weekly schedule policy")
			weeklySchedulePolicyInfo := Inst().Backup.CreateWeeklySchedulePolicy(1, backup.Friday, "9:10AM", 2)
			weeklyPolicyStatus := Inst().Backup.BackupSchedulePolicy(weeklyName, uuid.New(), BackupOrgID, weeklySchedulePolicyInfo)
			dash.VerifyFatal(weeklyPolicyStatus, nil, "Creating weekly schedule policy")

			log.InfoD("Creating backup monthly schedule policy")
			monthlySchedulePolicyInfo := Inst().Backup.CreateMonthlySchedulePolicy(1, 29, "9:20AM", 2)
			monthlyPolicyStatus := Inst().Backup.BackupSchedulePolicy(monthlyName, uuid.New(), BackupOrgID, monthlySchedulePolicyInfo)
			dash.VerifyFatal(monthlyPolicyStatus, nil, "Creating monthly schedule policy")
		})

		Step("Registering cluster for backup", func() {
			log.InfoD("Registering cluster for backup")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")

			err = CreateApplicationClusters(BackupOrgID, "", "", ctx)
			dash.VerifyFatal(err, nil, "Creating source and destination cluster")

			clusterStatus, err := Inst().Backup.GetClusterStatus(BackupOrgID, SourceClusterName, ctx)
			log.FailOnError(err, fmt.Sprintf("Fetching [%s] cluster status", SourceClusterName))
			dash.VerifyFatal(clusterStatus, api.ClusterInfo_StatusInfo_Online, fmt.Sprintf("Verifying if [%s] cluster is online", SourceClusterName))

			sourceClusterUid, err = Inst().Backup.GetClusterUID(ctx, BackupOrgID, SourceClusterName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", SourceClusterName))

			clusterStatus, err = Inst().Backup.GetClusterStatus(BackupOrgID, DestinationClusterName, ctx)
			log.FailOnError(err, fmt.Sprintf("Fetching [%s] cluster status", DestinationClusterName))
			dash.VerifyFatal(clusterStatus, api.ClusterInfo_StatusInfo_Online, fmt.Sprintf("Verifying if [%s] cluster is online", DestinationClusterName))

			destClusterUid, err = Inst().Backup.GetClusterUID(ctx, BackupOrgID, DestinationClusterName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", DestinationClusterName))
		})

		Step("Taking backup of application from source cluster", func() {
			log.InfoD("taking backup of applications")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")

			backupNames = make([]string, 0)
			for i, appCtx := range scheduledAppContexts {
				scheduledNamespace := appCtx.ScheduleOptions.Namespace
				backupName := fmt.Sprintf("%s-%s-%v", "autogenerated-backup", scheduledNamespace, time.Now().Unix())
				log.InfoD("creating backup [%s] in source cluster [%s] (%s), organization [%s], of namespace [%s], in backup location [%s]", backupName, SourceClusterName, sourceClusterUid, BackupOrgID, scheduledNamespace, backupLocationName)
				err := CreateBackupWithValidation(ctx, backupName, SourceClusterName, backupLocationName, backupLocationUID, scheduledAppContexts[i:i+1], labelSelectors, BackupOrgID, sourceClusterUid, "", "", "", "")
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creation and Validation of backup [%s]", backupName))
				backupNames = append(backupNames, backupName)
				if len(allVirtualMachines) > 0 {
					backupName := fmt.Sprintf("%s-%s-%v", "autogenerated-backup", scheduledNamespace, time.Now().Unix())
					err = CreateVMBackupWithValidation(ctx, backupName, allVirtualMachines, SourceClusterName, backupLocationName, backupLocationUID, scheduledAppContexts,
						labelSelectors, BackupOrgID, sourceClusterUid, "", "", "", "", true)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Creation of VM backup and Validation of backup [%s]", backupName))
					backupNames = append(backupNames, backupName)
				}
			}
		})

		Step("Restoring the backed up namespaces", func() {
			log.InfoD("Restoring the backed up namespaces")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			for i, appCtx := range scheduledAppContexts {
				scheduledNamespace := appCtx.ScheduleOptions.Namespace
				restoreName := fmt.Sprintf("%s-%s-%s", "test-restore", scheduledNamespace, RandomString(4))
				for strings.Contains(strings.Join(restoreNames, ","), restoreName) {
					restoreName = fmt.Sprintf("%s-%s-%s", "test-restore", scheduledNamespace, RandomString(4))
				}
				log.InfoD("Restoring [%s] namespace from the [%s] backup", scheduledNamespace, backupNames[i])
				err = CreateRestoreWithValidation(ctx, restoreName, backupNames[i], make(map[string]string), make(map[string]string), DestinationClusterName, destClusterUid, BackupOrgID, scheduledAppContexts[i:i+1])
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creation and Validation of restore [%s]", restoreName))
				restoreNames = append(restoreNames, restoreName)
			}
		})
	})
	JustAfterEach(func() {
		defer EndPxBackupTorpedoTest(make([]*scheduler.Context, 0))
		defer func() {
			log.InfoD("switching to default context")
			err := SetClusterContext("")
			log.FailOnError(err, "failed to SetClusterContext to default cluster")
		}()

		policyList := []string{intervalName, dailyName, weeklyName, monthlyName}
		ctx, err := backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")
		if len(preRuleNameList) > 0 {
			for _, ruleName := range preRuleNameList {
				err := Inst().Backup.DeleteRuleForBackup(BackupOrgID, ruleName)
				dash.VerifySafely(err, nil, fmt.Sprintf("Deleting backup pre rules [%s]", ruleName))
			}
		}
		if len(postRuleNameList) > 0 {
			for _, ruleName := range postRuleNameList {
				err := Inst().Backup.DeleteRuleForBackup(BackupOrgID, ruleName)
				dash.VerifySafely(err, nil, fmt.Sprintf("Deleting backup post rules [%s]", ruleName))
			}
		}
		err = Inst().Backup.DeleteBackupSchedulePolicy(BackupOrgID, policyList)
		dash.VerifySafely(err, nil, "Deleting backup schedule policies")
		opts := make(map[string]bool)
		opts[SkipClusterScopedObjects] = true

		log.Info("Destroying scheduled apps on source cluster")
		DestroyApps(scheduledAppContexts, opts)
		log.FailOnError(err, "Data validations failed")

		log.InfoD("switching to destination context")
		err = SetDestinationKubeConfig()
		log.FailOnError(err, "failed to switch to context to destination cluster")

		log.InfoD("Destroying restored apps on destination clusters")
		restoredAppContexts := make([]*scheduler.Context, 0)
		for _, scheduledAppContext := range scheduledAppContexts {
			restoredAppContext, err := CloneAppContextAndTransformWithMappings(scheduledAppContext, make(map[string]string), make(map[string]string), true)
			if err != nil {
				log.Errorf("TransformAppContextWithMappings: %v", err)
				continue
			}
			restoredAppContexts = append(restoredAppContexts, restoredAppContext)
		}
		DestroyApps(restoredAppContexts, opts)

		log.InfoD("switching to default context")
		err = SetClusterContext("")
		log.FailOnError(err, "failed to SetClusterContext to default cluster")

		backupDriver := Inst().Backup
		log.Info("Deleting backed up namespaces")
		for _, backupName := range backupNames {
			backupUID, err := backupDriver.GetBackupUID(ctx, backupName, BackupOrgID)
			log.FailOnError(err, "Failed while trying to get backup UID for - %s", backupName)
			backupDeleteResponse, err := DeleteBackup(backupName, backupUID, BackupOrgID, ctx)
			log.FailOnError(err, "Backup [%s] could not be deleted", backupName)
			dash.VerifyFatal(backupDeleteResponse.String(), "", fmt.Sprintf("Verifying [%s] backup deletion is successful", backupName))
		}
		log.Info("Deleting restored namespaces")
		for _, restoreName := range restoreNames {
			err = DeleteRestore(restoreName, BackupOrgID, ctx)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Deleting Restore [%s]", restoreName))
		}
		CleanupCloudSettingsAndClusters(backupLocationMap, cloudCredName, cloudCredUID, ctx)
	})
})

// Test case to verify that helm rollback is not allowed and check the error log
var _ = Describe("{InstallPxBackupHelmRollbackCheck}", Label(TestCaseLabelsMap[InstallPxBackupHelmRollbackCheck]...), func() {
	var namespace string
	JustBeforeEach(func() {
		StartPxBackupTorpedoTest("InstallPxBackupHelmRollbackCheck",
			"Installing Px Backup and checking if helm rollback is not allowed", nil, 304881, Mkoppal, Q4FY25)
		log.InfoD("Uninstalling Px-Backup...")
		err := UninstallPxBackup()
		log.FailOnError(err, "Failed to delete px-backup")
		namespace = "px-backup"
	})
	It("Should check if helm rollback is not allowed", func() {

		Step("Install Px-Backup", func() {
			log.InfoD("Install Px-Backup")
			valuesMap, err := ParseValuesFromFile("values1")
			log.FailOnError(err, "Failed to parse values file")
			_, err = InstallPxBackup(LatestPxBackupVersion, DefaultPxBackupHelmBranch, namespace, valuesMap)
			log.FailOnError(err, "Failed to install px-backup")
		})

		Step("Rollback the helm release", func() {
			log.InfoD("Rollback the helm release")
			err := RollbackPxBackup()
			if err != nil {
				log.InfoD("Rollback Px-Backup error - %s", err.Error())
				dash.VerifyFatal(strings.Contains(err.Error(), "failed to roll back release"), true, "Expecting rollback to fail")
			} else {
				log.FailOnError(fmt.Errorf("rollback was successful when it should not have been"),
					"Rollback was successful when it should not have been")
			}
		})

		Step("Validate the pre-rollback job pod logs", func() {
			log.InfoD("Validate the pre-rollback job pod logs")
			searchString := "Rollback is not supported due to potential data inconsistency risks. Please contact support."
			present, err := SearchPodLogs(map[string]string{"job-name": "pre-rollback-check"}, namespace, searchString)
			log.FailOnError(err, "Failed to search pod logs")
			dash.VerifyFatal(present, true, "Found the expected string in the pod logs")
		})
	})
	JustAfterEach(func() {
		defer EndPxBackupTorpedoTest(make([]*scheduler.Context, 0))
		log.Infof("No cleanup required for this testcase")
	})
})

// Test case to check port info in the release notes for upgrade
var _ = Describe("{UpgradePxBackupPortInfoCheck}", Label(TestCaseLabelsMap[UpgradePxBackupPortInfoCheck]...), func() {
	var (
		namespace string
	)
	rel := &release.Release{}
	JustBeforeEach(func() {
		StartPxBackupTorpedoTest("UpgradePxBackupPortInfoCheck", "Upgrading Px-Backup and checking if desired port information is displayed in the helm release notes", nil, 304871, SS, Q4FY25)
		log.InfoD("Uninstalling Px-Backup...")
		err := UninstallPxBackup()
		log.FailOnError(err, "Failed to delete px-backup")
		namespace = "px-backup"

		log.InfoD("Installing Px-Backup...")
		valuesMap, err := ParseValuesFromFile("values1")
		log.FailOnError(err, "Failed to parse values file")
		rel, err = InstallPxBackup("2.8.0", "master", namespace, valuesMap)
		log.FailOnError(err, "Failed to install px-backup")
	})
	It("Should check port info in the release notes for upgrade", func() {

		Step("Upgrade Px-Backup and validate the helm notes", func() {
			log.InfoD("Upgrade Px-Backup and validate the helm notes")
			valuesMap, err := ParseValuesFromFile("values1")
			log.FailOnError(err, "Failed to parse values file")
			rel, err = HelmUpgradePxBackup(LatestPxBackupVersion, DefaultPxBackupHelmBranch, namespace, valuesMap)
			log.FailOnError(err, "Failed to upgrade px-backup")
			log.InfoD("Helm notes - \n%s", rel.Info.Notes)
		})

		Step("Validate the helm notes", func() {
			searchString := "For more information on network pre-requisites: " +
				"https://docs.portworx.com/portworx-backup-on-prem/install/install-prereq/nw-prereqs.html"
			dash.VerifyFatal(strings.Contains(rel.Info.Notes, searchString), true,
				"Found the expected string in the helm notes")
		})
	})

	JustAfterEach(func() {
		defer EndPxBackupTorpedoTest(make([]*scheduler.Context, 0))
		log.Infof("No cleanup required for this testcase")
	})
})
