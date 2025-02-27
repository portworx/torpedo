package tests

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	"github.com/pborman/uuid"
	api "github.com/portworx/px-backup-api/pkg/apis/v1"
	"github.com/portworx/sched-ops/k8s/talisman"
	"github.com/portworx/talisman/pkg/apis/portworx/v1beta1"
	"github.com/pure-px/sched-ops/k8s/core"
	"github.com/pure-px/sched-ops/k8s/storage"
	"github.com/pure-px/torpedo/drivers/backup"
	"github.com/pure-px/torpedo/drivers/scheduler"
	"github.com/pure-px/torpedo/drivers/scheduler/k8s"
	"github.com/pure-px/torpedo/pkg/log"
	"github.com/pure-px/torpedo/pkg/vpsutil"
	. "github.com/pure-px/torpedo/tests"
	corev1 "k8s.io/api/core/v1"
	storageApi "k8s.io/api/storage/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Try to restore when token manager is not reachable
var _ = Describe("{TestRestoreResilienceToKeycloakScale}", Label(TestCaseLabelsMap[PxBackupLabel]...), func() {
	var (
		scheduledAppContexts   []*scheduler.Context
		bkpNamespaces          []string
		clusterUid             string
		clusterStatus          api.ClusterInfo_StatusInfo_Status
		backupName             string
		backupLocationUID      string
		cloudCredName          string
		cloudCredUID           string
		bkpLocationName        string
		backupNames            []string
		providers              []string
		ctx                    context.Context
		backupLocationMap      map[string]string
		err                    error
		restoreName            string
		numOfBackup            int
		pxbNamespace           string
		scaledDownReplicaCount int32
		originalReplicaCount   int32
	)

	JustBeforeEach(func() {
		StartPxBackupTorpedoTest("VerifyTestRestoreResilienceToKeycloakScale", "Try to restore when token manager is not reachable", nil, 3006567, Pingle, Q3FY25)

		backupLocationMap = make(map[string]string)
		bkpNamespaces = make([]string, 0)
		backupNames = make([]string, 0)
		scheduledAppContexts = make([]*scheduler.Context, 0)
		numOfBackup = 1
		scaledDownReplicaCount = 0
		originalReplicaCount = 1
		ctx, err = backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")
		providers = GetBackupProviders()

		// Schedule an Application
		appContexts := ScheduleApplications(TaskNamePrefix)
		for _, ctx := range appContexts {
			ctx.ReadinessTimeout = AppReadinessTimeout
			namespace := GetAppNamespace(ctx, TaskNamePrefix)
			bkpNamespaces = append(bkpNamespaces, namespace)
			scheduledAppContexts = append(scheduledAppContexts, ctx)
		}
	})

	//Try to restore when token manager is not reachable
	It("Try to restore when token manager is not reachable", func() {
		// 1. validate application
		Step("Validate applications", func() {
			log.InfoD("Validating applications")
			ValidateApplications(scheduledAppContexts)
		})

		// 2. Create cloud credentials and backup location
		Step("Creating cloud credentials and backup location", func() {
			log.InfoD("Creating cloud credentials and backup location")
			for _, provider := range providers {
				cloudCredName = fmt.Sprintf("%s-%s-%v", "cloudcred", provider, time.Now().Unix())
				bkpLocationName = fmt.Sprintf("%s-%s-%v-bl", provider, getGlobalBucketName(provider), time.Now().Unix())
				cloudCredUID = uuid.New()
				backupLocationUID = uuid.New()
				backupLocationMap[backupLocationUID] = bkpLocationName
				err := CreateCloudCredential(provider, cloudCredName, cloudCredUID, BackupOrgID, ctx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of cloud credential named [%s] for org [%s] with [%s] as provider", cloudCredName, BackupOrgID, provider))
				err = CreateBackupLocation(provider, bkpLocationName, backupLocationUID, cloudCredName, cloudCredUID, getGlobalBucketName(provider), BackupOrgID, "", true)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creating backup location %s", bkpLocationName))
			}
		})

		// 3. Create application cluster for backup
		Step("Register cluster for backup", func() {
			err := CreateApplicationClusters(BackupOrgID, "", "", ctx)
			dash.VerifyFatal(err, nil, "Creating source and destination cluster")
			clusterStatus, err = Inst().Backup.GetClusterStatus(BackupOrgID, SourceClusterName, ctx)
			log.FailOnError(err, fmt.Sprintf("Fetching [%s] cluster status", SourceClusterName))
			dash.VerifyFatal(clusterStatus, api.ClusterInfo_StatusInfo_Online, fmt.Sprintf("Verifying if [%s] cluster is online", SourceClusterName))
			clusterUid, err = Inst().Backup.GetClusterUID(ctx, BackupOrgID, SourceClusterName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", SourceClusterName))
			log.InfoD("Uid of [%s] cluster is %s", SourceClusterName, clusterUid)
		})

		// 4. Take backup of applications
		Step("Taking backup of applications", func() {
			for i := 0; i < numOfBackup; i++ {
				backupName = fmt.Sprintf("%s-%s", BackupNamePrefix, RandomString(6))
				appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, bkpNamespaces)
				err = CreateBackupWithValidation(ctx, backupName, SourceClusterName, bkpLocationName, backupLocationUID, appContextsToBackup, nil, BackupOrgID, clusterUid, "", "", "", "")
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creation and Validation of backup [%s]", backupName))
				backupNames = append(backupNames, backupName)
			}
		})

		// 5.create restore
		Step("Restoring the backups application", func() {
			restoreName = fmt.Sprintf("%s-restore", backupName)
			for _, backupName := range backupNames {
				ctx, err := backup.GetAdminCtxFromSecret()
				log.FailOnError(err, "Fetching px-central-admin ctx")
				_, err = CreateRestoreWithoutCheck(restoreName, backupName, nil, SourceClusterName, clusterUid, BackupOrgID, ctx)
				log.FailOnError(err, "Failed while trying to restore [%s] the backup [%s]", fmt.Sprintf("%s-restore", backupName), backupName)
			}
		})

		// 6. Scale down pxcentral-keycloak
		Step("Scale down pxcentral-keycloak", func() {
			log.Infof("Scaling down %s to %d replicas in namespace %s", KeyCloakStateFulSet, scaledDownReplicaCount, pxbNamespace)
			pxbNamespace, err = backup.GetPxBackupNamespace()
			dash.VerifyFatal(err, nil, "failed to get Px-Backup namespace")
			err = ScaleStatefulSetReplicas(KeyCloakStateFulSet, pxbNamespace, scaledDownReplicaCount, 0, PodStatusTimeOut, PodStatusRetryTime)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Scale of: %s with scale count: %d", KeyCloakStateFulSet, scaledDownReplicaCount))
			err = VerifyStatefulSetCount(KeyCloakStateFulSet, pxbNamespace, 0, PodStatusTimeOut, PodStatusRetryTime)
			dash.VerifyFatal(err, nil, "Verify scale count after scale down to 0")
		})

		// 7. Scale up pxcentral-keycloak to actual state
		Step("Scale up pxcentral-keycloak", func() {
			log.Infof("Scaling up %s to %d replicas in namespace %s", KeyCloakStateFulSet, originalReplicaCount, pxbNamespace)
			err = ScaleStatefulSetReplicas(KeyCloakStateFulSet, pxbNamespace, originalReplicaCount, 1, PodStatusTimeOut, PodStatusRetryTime)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Scale of: %s with scale count: %d", KeyCloakStateFulSet, originalReplicaCount))
		})

		// 8. Success check for restore
		Step("Success check for restore", func() {
			log.Infof("Checking success state for the backup %s", backupName)
			err := RestoreSuccessCheck(restoreName, BackupOrgID, MaxWaitPeriodForRestoreCompletionInMinute*time.Minute, 30*time.Second, ctx)
			dash.VerifyFatal(err, nil, "Verify restore success check")
		})
	})

	JustAfterEach(func() {
		defer EndPxBackupTorpedoTest(scheduledAppContexts)
		// Clean up the cluster
		log.InfoD("deleting restores")
		backupUID, err := Inst().Backup.GetBackupUID(ctx, backupName, BackupOrgID)
		log.FailOnError(err, "Failed while trying to get backup UID for - %s", backupName)
		_, err = DeleteBackup(backupName, backupUID, BackupOrgID, ctx)
		dash.VerifySafely(err, nil, fmt.Sprintf("Verifying backup deletion : %v", backupName))

		// Delete restore
		err = DeleteRestore(restoreName, BackupOrgID, ctx)
		dash.VerifySafely(err, nil, fmt.Sprintf("deleting Restore [%s]", restoreName))
		CleanupCloudSettingsAndClusters(backupLocationMap, cloudCredName, cloudCredUID, ctx)

	})
})

// This test verifies if Restore a backup of a app which was deployed with VolumePlacementStrategy Rules is possible
var _ = Describe("{RestoreBackupOfAppDeployedWithVolumePlacementStrategyRules}", Label(TestCaseLabelsMap[BackupAppDeployedUsingVPS]...), func() {
	var (
		scheduledAppContexts []*scheduler.Context
		vpsName              string
		scName               string
		backupNs             string
		clusterUid           string
		clusterStatus        api.ClusterInfo_StatusInfo_Status
		cloudCredName        string
		cloudCredUID         string
		backupLocationUID    string
		providers            []string
		backupLocationMap    map[string]string
		backupLocation       string
		adminCtx             context.Context
		backupName           string
		pvcName              string
		labelSelectors       map[string]string
		bkpNamespaces        []string
		restoreNames         []string
	)

	JustBeforeEach(func() {
		StartPxBackupTorpedoTest("VerifyRestoreBackupOfAppWhichWasDeployedWithVolumePlacementStrategyRules", "Restore a backup of a app which was deployed with VolumePlacementStrategy Rules", nil, 300571, ABadgujar, Q1FY25)
		vpsName = fmt.Sprintf("%v-%v", "postgres-volume-affinity", time.Now().Unix())
		backupNs = "mysql"
		backupLocationMap = make(map[string]string)
		log.InfoD("Deploy applications")
		providers = GetBackupProviders()
		var err error
		adminCtx, err = backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")
		labelSelectors = make(map[string]string)
		originalAppList := Inst().AppList
		bkpNamespaces = make([]string, 0)
		restoreNames = make([]string, 0)

		//Step - 1. Create Namespace

		ns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: backupNs,
			},
		}
		log.InfoD("Creating namespace %v", backupNs)
		_, err = core.Instance().CreateNamespace(ns)

		if err != nil {
			if apierrors.IsAlreadyExists(err) {
				log.Infof("Namespace %s already exists. Skipping creation.", ns.Name)
			} else {
				log.FailOnError(err, fmt.Sprintf("error creating namespace [%s]", backupNs))
			}
		}

		for i := 0; i < len(originalAppList); i++ {
			taskName := fmt.Sprintf("%s-%s", TaskNamePrefix, RandomString(3))
			appContexts := ScheduleApplicationsOnNamespace(backupNs, taskName)
			for _, ctx := range appContexts {
				ctx.ReadinessTimeout = 2 * AppReadinessTimeout
				namespace := GetAppNamespace(ctx, taskName)
				bkpNamespaces = append(bkpNamespaces, namespace)
				scheduledAppContexts = append(scheduledAppContexts, ctx)
			}
		}
		// Step - 2.Validate the application
		ValidateApplications(scheduledAppContexts)

	})

	// Verify if Restore a backup of a app which was deployed with Volume Placement Strategy Rules is possible
	It("Verify if Restore a backup of a app which was deployed with Volume Placement Strategy Rules is possible", func() {

		// Step - 3. Apply volume placement strategy"
		Step("Apply volume placement strategy", func() {
			log.InfoD("Apply volume placement strategy")
			matchExpression := []*v1beta1.LabelSelectorRequirement{
				{
					Key:      "app",
					Operator: v1beta1.LabelSelectorOpIn,
					Values:   []string{"postgres"},
				},
			}
			vpsSpec := vpsutil.VolumeAffinityByMatchExpression(vpsName, matchExpression)
			_, err := talisman.Instance().CreateVolumePlacementStrategy(&vpsSpec)
			dash.VerifyFatal(err, nil, "Check if able to apply volume placement strategy")
		})

		// Step - 4. Create StorageClass"
		Step("Create StorageClass", func() {
			log.InfoD("Create StorageClass")
			scName = fmt.Sprintf("%v-%v", "postgres-storage-class", time.Now().Unix())
			params := make(map[string]string)
			log.InfoD("Apply Storage Class")
			k8sStorage := storage.Instance()
			params["placement_strategy"] = vpsName
			v1obj := metav1.ObjectMeta{
				Name:      scName,
				Namespace: backupNs,
			}
			bindMode := storageApi.VolumeBindingImmediate
			scObj := storageApi.StorageClass{
				ObjectMeta:        v1obj,
				Provisioner:       k8s.CsiProvisioner,
				Parameters:        params,
				VolumeBindingMode: &bindMode,
			}
			_, err := k8sStorage.CreateStorageClass(&scObj)
			dash.VerifyFatal(err, nil, "Verifying creation of new storage class")
		})

		// Step - 5. Create Persistent Volume Claim
		Step("Create Persistent Volume Claim", func() {
			log.InfoD("Create Persistent Volume Claim")
			pvcName = fmt.Sprintf("%v-%v", "postgres-pvc", time.Now().Unix())
			namespace := backupNs
			_, err := core.Instance().CreatePersistentVolumeClaim(&corev1.PersistentVolumeClaim{
				TypeMeta: metav1.TypeMeta{
					Kind: "PersistentVolumeClaim",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      pvcName,
					Namespace: backupNs,
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					AccessModes:      []corev1.PersistentVolumeAccessMode{corev1.ReadWriteMany},
					StorageClassName: &scName,
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceStorage: resource.MustParse("2Gi"),
						},
					},
				},
			})
			dash.VerifyFatal(err, nil, "Verifying creation of PVC")

			_, err = core.Instance().GetPersistentVolumeClaim(pvcName, namespace)
			log.FailOnError(err, "Failed to get pvc")
			err = Inst().S.WaitForSinglePVCToBound(pvcName, namespace, 3)
			log.FailOnError(err, "Failed to wait for pvc to bound")
		})

		// Step - 6. Creating cloud credentials and backup location
		Step("Creating cloud credentials and backup location", func() {
			log.InfoD("Creating cloud credentials and backup location")
			for _, provider := range providers {
				cloudCredName = fmt.Sprintf("%s-%s-%v", "cred", provider, time.Now().Unix())
				cloudCredUID = uuid.New()
				err := CreateCloudCredential(provider, cloudCredName, cloudCredUID, BackupOrgID, adminCtx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of cloud credential named [%s] for org [%s] with [%s] as provider", cloudCredName, BackupOrgID, provider))
				log.InfoD("Created Cloud Credentials with name - %s", cloudCredName)
				backupLocation = fmt.Sprintf("autogenerated-backup-location-%v", time.Now().Unix())
				backupLocationUID = uuid.New()
				backupLocationMap[backupLocationUID] = backupLocation
				err = CreateBackupLocation(provider, backupLocation, backupLocationUID, cloudCredName, cloudCredUID, getGlobalBucketName(provider), BackupOrgID, "", true)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creating backup location %v", backupLocation))
				log.InfoD("Created Backup Location with name - %s", backupLocation)

			}
		})

		//Step - 7. Register cluster for backup
		Step("Register cluster for backup", func() {
			log.InfoD("Register cluster for backup")
			err := CreateApplicationClusters(BackupOrgID, "", "", adminCtx)
			dash.VerifyFatal(err, nil, "Creating source cluster")
			clusterStatus, err = Inst().Backup.GetClusterStatus(BackupOrgID, SourceClusterName, adminCtx)
			log.FailOnError(err, fmt.Sprintf("Fetching [%s] cluster status", SourceClusterName))
			dash.VerifyFatal(clusterStatus, api.ClusterInfo_StatusInfo_Online, fmt.Sprintf("Verifying if [%s] cluster is online", SourceClusterName))
			clusterUid, err = Inst().Backup.GetClusterUID(adminCtx, BackupOrgID, SourceClusterName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", SourceClusterName))
		})

		//Step - 8. Take backup
		Step("Taking backup of applications", func() {
			log.InfoD("Taking backup of applications")
			backupName = fmt.Sprintf("%s-%v", BackupNamePrefix, time.Now().Unix())
			appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, bkpNamespaces)
			err := CreateBackupWithValidation(adminCtx, backupName, SourceClusterName, backupLocation, backupLocationUID, appContextsToBackup, labelSelectors, BackupOrgID, clusterUid, "", "", "", "")
			dash.VerifyFatal(err, nil, fmt.Sprintf("Creation of backup successful [%s]", backupName))
		})

		//Step - 9. Restore backup of Dependent Resources
		Step("Restore backup of Dependent Resources", func() {
			log.InfoD("Restore backup of Dependent Resources")
			DestClusterUid, err := Inst().Backup.GetClusterUID(adminCtx, BackupOrgID, DestinationClusterName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", DestinationClusterName))
			restoreName := fmt.Sprintf("%s-%v", RestoreNamePrefix, time.Now().Unix())
			appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, bkpNamespaces)
			err = CreateRestoreWithValidation(adminCtx, restoreName, backupName, make(map[string]string), make(map[string]string), DestinationClusterName, DestClusterUid, BackupOrgID, appContextsToBackup)
			dash.VerifyNotNilFatal(err, fmt.Sprintf("Validation of Restore [%s] ", restoreName))
			restoreNames = append(restoreNames, restoreName)
		})

	})
	JustAfterEach(func() {
		defer EndPxBackupTorpedoTest(scheduledAppContexts)
		log.InfoD("Entered Cleanup")
		log.InfoD("Deleting the deployed apps after the testcase")

		err := DeleteNamespaces([]string{backupNs})
		dash.VerifyFatal(err, nil, fmt.Sprintf("Deleting Namespace - %v", backupNs))

		log.Infof("Deleting the newly created VPS")
		err = talisman.Instance().DeleteVolumePlacementStrategy(vpsName)
		log.FailOnError(err, "Failed to remove VPS: %v", vpsName)

		log.Infof("Deleting the newly created storage class")
		err = storage.Instance().DeleteStorageClass(scName)
		log.FailOnError(err, "Failed to remove storage class: %v", scName)

		// Cleaning up applications created
		opts := make(map[string]bool)
		opts[SkipClusterScopedObjects] = true
		DestroyApps(scheduledAppContexts, opts)

		//Cleanup Restores
		log.InfoD("Delete Restore")
		for _, restoreNameIteration := range restoreNames {
			err = DeleteRestore(restoreNameIteration, BackupOrgID, adminCtx)
			dash.VerifySafely(err, nil, fmt.Sprintf("Deleting restore [%s]", restoreNameIteration))
		}

		//Cleaning up Backups, Backuplocations, cloud settings
		CleanupCloudSettingsAndClusters(backupLocationMap, cloudCredName, cloudCredUID, adminCtx)
	})
})
