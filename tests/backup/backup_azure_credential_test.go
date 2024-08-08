package tests

import (
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	"github.com/pborman/uuid"
	api "github.com/portworx/px-backup-api/pkg/apis/v1"
	"github.com/portworx/torpedo/drivers/backup"
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests"
	"golang.org/x/sync/errgroup"
)

// This testcase creates Azure cloud account with mandatory and non-mandatory fields and take backup & restore
var _ = Describe("{AzureCloudAccountCreationWithMandatoryAndNonMandatoryFields}", Label(TestCaseLabelsMap[AzureCloudAccountCreationWithMandatoryAndNonMandatoryFields]...), func() {

	var (
		credUidWithAllFields                                   string
		sourceClusterUid                                       string
		backupNameWithBkpLocationHavingMandatoryParameters     string
		restoreNameWithBkpLocationHavingMandatoryParameters    string
		backupNameWithBkpLocationHavingAllParameters           string
		restoreNameWithBkpLocationHavingAllParameters          string
		backupNameWithBkpLocationHavingOnlyMandatoryParameters string
		restoreNameWithBkpLocationHavingOnlyMandatoryFields    string
		credUidWithMandatoryFields                             string
		azureCredNameWithAllFields                             string
		azureCredNameWithMandatoryFields                       string
		azureBackupLocationNameWithAllFields                   string
		azureBackupLocationNameWithMandatoryFields             string
		backupLocationAllFieldsUID                             string
		backupLocationMandatoryFieldsUID                       string
		tenantID                                               string
		clientID                                               string
		clientSecret                                           string
		subscriptionID                                         string
		accountName                                            string
		accountKey                                             string
		appNamespaces                                          []string
		controlChannel                                         chan string
		backupLocationMap1                                     map[string]string
		cloudCredentialMap1                                    map[string]string
		backupLocationMap2                                     map[string]string
		cloudCredentialMap2                                    map[string]string
		errorGroup                                             *errgroup.Group
		azureConfigFields                                      *api.AzureConfig
		scheduledAppContexts                                   []*scheduler.Context
		contexts                                               []*scheduler.Context
		appContexts                                            []*scheduler.Context
	)

	JustBeforeEach(func() {
		StartPxBackupTorpedoTest("AzureCloudAccountCreationWithMandatoryAndNonMandatoryFields", "Azure cloud account with mandatory and non mandatory fields", nil, 31661, Sagrawal, Q2FY25)
		backupLocationMap1 = make(map[string]string)
		cloudCredentialMap1 = make(map[string]string)
		backupLocationMap2 = make(map[string]string)
		cloudCredentialMap2 = make(map[string]string)
		log.InfoD("Deploying applications required for the testcase")
		contexts = make([]*scheduler.Context, 0)
		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			taskName := fmt.Sprintf("%s-%d", TaskNamePrefix, i)
			appContexts = ScheduleApplications(taskName)
			contexts = append(contexts, appContexts...)
			for _, ctx := range appContexts {
				ctx.ReadinessTimeout = AppReadinessTimeout
				namespace := GetAppNamespace(ctx, taskName)
				appNamespaces = append(appNamespaces, namespace)
				scheduledAppContexts = append(scheduledAppContexts, ctx)
			}
		}
	})

	It("Azure cloud account with mandatory and non mandatory fields", func() {
		ctx, err := backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")
		tenantID, clientID, clientSecret, subscriptionID, accountName, accountKey = GetAzureCredsFromEnv()
		Step("Validating applications", func() {
			log.InfoD("Validating applications")
			ctx, _ := backup.GetAdminCtxFromSecret()
			controlChannel, errorGroup = ValidateApplicationsStartData(scheduledAppContexts, ctx)
		})

		Step("Creating azure cloud account and backup location with all fields", func() {
			credUidWithAllFields = uuid.New()
			azureConfigFields = &api.AzureConfig{
				TenantId:       tenantID,
				ClientId:       clientID,
				ClientSecret:   clientSecret,
				AccountName:    accountName,
				AccountKey:     accountKey,
				SubscriptionId: subscriptionID,
			}
			log.InfoD("Creating azure cloud account with all fields")
			azureCredNameWithAllFields = fmt.Sprintf("%s-azure-cred-with-all-fields", RandomString(5))
			err := CreateAzureCloudCredential(azureCredNameWithAllFields, credUidWithAllFields, BackupOrgID, azureConfigFields, ctx)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of azure cloud credential named [%s] for org [%s] having all fields", azureCredNameWithAllFields, BackupOrgID))
			cloudCredentialMap1[azureCredNameWithAllFields] = credUidWithAllFields

			log.InfoD("Creating backup location using cloud credential having all fields")
			azureBackupLocationNameWithAllFields = fmt.Sprintf("azure-backup-location-with-all-fields-%v", RandomString(5))
			backupLocationAllFieldsUID = uuid.New()
			backupLocationMap1[backupLocationAllFieldsUID] = azureBackupLocationNameWithAllFields
			err = CreateBackupLocation("azure", azureBackupLocationNameWithAllFields, backupLocationAllFieldsUID, azureCredNameWithAllFields, credUidWithAllFields, getGlobalBucketName("azure"), BackupOrgID, "", true)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Creating backup location %s with all fields in azure credentials", azureBackupLocationNameWithAllFields))
		})

		Step("Creating azure cloud account and backup location with only mandatory fields", func() {
			log.InfoD("Creating azure cloud account with only mandatory fields")
			credUidWithMandatoryFields = uuid.New()
			azureConfigFields = &api.AzureConfig{
				AccountName: accountName,
				AccountKey:  accountKey,
			}
			azureCredNameWithMandatoryFields = fmt.Sprintf("%s-azure-cred-with-mandatory-fields", RandomString(5))
			err = CreateAzureCloudCredential(azureCredNameWithMandatoryFields, credUidWithMandatoryFields, BackupOrgID, azureConfigFields, ctx)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of azure cloud credential named [%s] for org [%s] having only mandatory fields", azureCredNameWithMandatoryFields, BackupOrgID))
			cloudCredentialMap2[azureCredNameWithMandatoryFields] = credUidWithMandatoryFields

			log.InfoD("Creating backup location with mandatory fields in azure credentials")
			azureBackupLocationNameWithMandatoryFields = fmt.Sprintf("azure-backup-location-with-mandatory-fields-%v", RandomString(5))
			backupLocationMandatoryFieldsUID = uuid.New()
			backupLocationMap2[backupLocationMandatoryFieldsUID] = azureBackupLocationNameWithMandatoryFields
			err = CreateBackupLocation("azure", azureBackupLocationNameWithMandatoryFields, backupLocationMandatoryFieldsUID, azureCredNameWithMandatoryFields, credUidWithMandatoryFields, getGlobalBucketName("azure"), BackupOrgID, "", true)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Creating backup location %s with mandatory fields in azure credentials", azureBackupLocationNameWithMandatoryFields))

		})

		Step("Registering azure application clusters using cloud credential having all fields", func() {
			log.InfoD("Registering azure application clusters using cloud credential having all fields")
			err = AddAzureApplicationClusters(BackupOrgID, azureCredNameWithAllFields, credUidWithAllFields, ctx)
			dash.VerifyFatal(err, nil, "Adding source and destination cluster using cloud credential having all fields")
			sourceClusterUid, err = Inst().Backup.GetClusterUID(ctx, BackupOrgID, SourceClusterName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", SourceClusterName))
		})

		Step("Taking backup of applications using backup location with cloud credential having only mandatory parameters", func() {
			log.InfoD("Taking backup of applications using backup location with cloud credential having only mandatory parameters")
			backupNameWithBkpLocationHavingMandatoryParameters = fmt.Sprintf("backup-with-bkp-loc-having-mandatory-params-%v", RandomString(5))
			appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, appNamespaces)
			err = CreateBackupWithValidation(ctx, backupNameWithBkpLocationHavingMandatoryParameters, SourceClusterName, azureBackupLocationNameWithMandatoryFields, backupLocationMandatoryFieldsUID, appContextsToBackup, nil, BackupOrgID, sourceClusterUid, "", "", "", "")
			dash.VerifyFatal(err, nil, fmt.Sprintf("Creation and Validation of backup [%s]", backupNameWithBkpLocationHavingMandatoryParameters))
		})

		Step("Restoring the backed up application from backup location with cloud credential having only mandatory parameters", func() {
			log.InfoD("Restoring the backed up application from backup location with cloud credential having only mandatory parameters")
			restoreNameWithBkpLocationHavingMandatoryParameters = fmt.Sprintf("restore-%s-%v", backupNameWithBkpLocationHavingMandatoryParameters, RandomString(5))
			appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, appNamespaces)
			err = CreateRestoreWithValidation(ctx, restoreNameWithBkpLocationHavingMandatoryParameters, backupNameWithBkpLocationHavingMandatoryParameters, make(map[string]string), make(map[string]string), DestinationClusterName, BackupOrgID, appContextsToBackup)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Validating restore-%s", restoreNameWithBkpLocationHavingMandatoryParameters))
		})

		Step("Removing the added application cluster and registering new azure application clusters using cloud credential having only mandatory fields", func() {
			log.InfoD("Deleting registered clusters for admin context")
			err = DeleteCluster(SourceClusterName, BackupOrgID, ctx, true)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Deleting cluster %s", SourceClusterName))
			err = Inst().Backup.WaitForClusterDeletion(ctx, SourceClusterName, BackupOrgID, ClusterDeleteTimeout, ClusterDeleteRetryTime)
			log.FailOnError(err, fmt.Sprintf("waiting for cluster [%s] deletion", SourceClusterName))

			err = DeleteCluster(DestinationClusterName, BackupOrgID, ctx, true)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Deleting cluster %s", DestinationClusterName))
			err = Inst().Backup.WaitForClusterDeletion(ctx, DestinationClusterName, BackupOrgID, ClusterDeleteTimeout, ClusterDeleteRetryTime)
			log.FailOnError(err, fmt.Sprintf("waiting for cluster [%s] deletion", DestinationClusterName))

			log.InfoD("Registering new azure application clusters using cloud credential having only mandatory fields")
			err = AddAzureApplicationClusters(BackupOrgID, azureCredNameWithMandatoryFields, credUidWithMandatoryFields, ctx)
			dash.VerifyFatal(err, nil, "Adding source and destination cluster using cloud credential having only mandatory fields")
			sourceClusterUid, err = Inst().Backup.GetClusterUID(ctx, BackupOrgID, SourceClusterName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", SourceClusterName))
		})

		Step("Taking backup of applications using backup location with cloud credential having all fields", func() {
			log.InfoD("Taking backup of applications using backup location with cloud credential having all fields")
			backupNameWithBkpLocationHavingAllParameters = fmt.Sprintf("backup-with-bkp-loc-having-all-params-%v", RandomString(5))
			appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, appNamespaces)
			err = CreateBackupWithValidation(ctx, backupNameWithBkpLocationHavingAllParameters, SourceClusterName, azureBackupLocationNameWithAllFields, backupLocationAllFieldsUID, appContextsToBackup, nil, BackupOrgID, sourceClusterUid, "", "", "", "")
			dash.VerifyFatal(err, nil, fmt.Sprintf("Creation and Validation of backup [%s]", backupNameWithBkpLocationHavingAllParameters))
		})

		Step("Restoring the backed up application from backup location with cloud credential having all fields", func() {
			log.InfoD("Restoring the backed up application from backup location with cloud credential having all fields")
			restoreNameWithBkpLocationHavingAllParameters = fmt.Sprintf("restore-%s-%v", backupNameWithBkpLocationHavingAllParameters, RandomString(5))
			appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, appNamespaces)
			err = CreateRestoreWithValidation(ctx, restoreNameWithBkpLocationHavingAllParameters, backupNameWithBkpLocationHavingAllParameters, make(map[string]string), make(map[string]string), DestinationClusterName, BackupOrgID, appContextsToBackup)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Validating restore-%s", restoreNameWithBkpLocationHavingAllParameters))
		})

		Step("Taking backup of applications using backup location with cloud credential having only mandatory fields", func() {
			log.InfoD("Taking backup of applications using backup location with cloud credential having only mandatory fields")
			backupNameWithBkpLocationHavingOnlyMandatoryParameters = fmt.Sprintf("backup-with-bkp-loc-having-only-mandatory-fields-%v", RandomString(5))
			appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, appNamespaces)
			err = CreateBackupWithValidation(ctx, backupNameWithBkpLocationHavingOnlyMandatoryParameters, SourceClusterName, azureBackupLocationNameWithMandatoryFields, backupLocationMandatoryFieldsUID, appContextsToBackup, nil, BackupOrgID, sourceClusterUid, "", "", "", "")
			dash.VerifyFatal(err, nil, fmt.Sprintf("Creation and Validation of backup [%s]", backupNameWithBkpLocationHavingOnlyMandatoryParameters))
		})

		Step("Restoring the backed up application from backup location with cloud credential having only mandatory parameters", func() {
			log.InfoD("Restoring the backed up application from backup location with cloud credential having only mandatory parameters")
			restoreNameWithBkpLocationHavingOnlyMandatoryFields = fmt.Sprintf("restore-%s-%v", backupNameWithBkpLocationHavingOnlyMandatoryParameters, RandomString(5))
			appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, appNamespaces)
			err = CreateRestoreWithValidation(ctx, restoreNameWithBkpLocationHavingOnlyMandatoryFields, backupNameWithBkpLocationHavingOnlyMandatoryParameters, make(map[string]string), make(map[string]string), DestinationClusterName, BackupOrgID, appContextsToBackup)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Validating restore-%s", restoreNameWithBkpLocationHavingOnlyMandatoryFields))
		})
	})

	JustAfterEach(func() {
		defer EndPxBackupTorpedoTest(scheduledAppContexts)
		ctx, err := backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")

		// We need to delete the cluster before deleting the cloud credential
		log.InfoD("Deleting registered clusters for admin context")
		err = DeleteCluster(SourceClusterName, BackupOrgID, ctx, true)
		dash.VerifySafely(err, nil, fmt.Sprintf("Deleting cluster %s", SourceClusterName))
		err = DeleteCluster(DestinationClusterName, BackupOrgID, ctx, true)
		dash.VerifySafely(err, nil, fmt.Sprintf("Deleting cluster %s", DestinationClusterName))

		log.InfoD("Cleaning up cloud settings and application clusters")
		for cloudCredName, cloudCredUID := range cloudCredentialMap1 {
			CleanupCloudSettingsAndClusters(backupLocationMap1, cloudCredName, cloudCredUID, ctx)
		}
		for cloudCredName, cloudCredUID := range cloudCredentialMap2 {
			CleanupCloudSettingsAndClusters(backupLocationMap2, cloudCredName, cloudCredUID, ctx)
		}

		opts := make(map[string]bool)
		opts[SkipClusterScopedObjects] = true
		log.InfoD("Deleting deployed namespaces - %v", appNamespaces)
		err = DestroyAppsWithData(scheduledAppContexts, opts, controlChannel, errorGroup)
		log.FailOnError(err, "Data validations failed")
	})
})
