package tests

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	api "github.com/portworx/px-backup-api/pkg/apis/v1"
	"github.com/pure-px/sched-ops/k8s/apps"
	"github.com/pure-px/torpedo/drivers/backup"
	"github.com/pure-px/torpedo/drivers/node"
	"github.com/pure-px/torpedo/drivers/scheduler"
	"github.com/pure-px/torpedo/pkg/log"
	. "github.com/pure-px/torpedo/tests"
	appsV1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// Add clusters and delete them from backend and check if px-backup license page is breaking or not
var _ = Describe("{AddAndUpdateClusterCheckLicensePageIntegrity}", Label(TestCaseLabelsMap[CheckLicensePageIntegrity]...), func() {
	var (
		scheduledAppContexts []*scheduler.Context
		clusterUid           string
		clusterStatus        api.ClusterInfo_StatusInfo_Status
		cloudCredName        string
		cloudCredUID         string
		backupLocationMap    map[string]string
		err                  error
		ctx                  context.Context
		bkpNamespaces        []string
		licenceCount         int
	)
	JustBeforeEach(func() {
		StartPxBackupTorpedoTest("VerifyAddAndUpdateClusterCheckLicensePageIntegrity", "Add clusters and delete them from backend and check if px-backup license page is breaking or not", nil, 91966, Pingle, Q2FY25)
		ctx, err = backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")
		scheduledAppContexts = make([]*scheduler.Context, 0)
		backupLocationMap = make(map[string]string)
		bkpNamespaces = make([]string, 0)
		// Schedule an Application
		appContexts := ScheduleApplications(TaskNamePrefix)
		for _, ctx := range appContexts {
			ctx.ReadinessTimeout = AppReadinessTimeout
			namespace, err := backup.GetPxBackupNamespace()
			dash.VerifyFatal(err, nil, fmt.Sprintf("pxbNamespace :: %s", namespace))
			bkpNamespaces = append(bkpNamespaces, namespace)
			scheduledAppContexts = append(scheduledAppContexts, ctx)
		}
	})

	//Add clusters and delete them from backend and verify license page is breaking or not
	It("Add clusters and delete them from backend and verify license page is breaking or not", func() {
		invalidKubeconfig := "\"\""
		// 1. Validate application
		Step("Validate applications", func() {
			log.Infof("Validate applications")
			ValidateApplications(scheduledAppContexts)

		})

		// 2. Check the license count by inspecting
		Step("Check the license count by inspect shouldn't fail", func() {
			licenseInspectRequestObject := &api.LicenseInspectRequest{
				OrgId: BackupOrgID,
			}
			expected := len(node.GetWorkerNodes())
			licenseInspectResponse, err := Inst().Backup.InspectLicense(ctx, licenseInspectRequestObject)
			dash.VerifyFatal(err, nil, "verify cluster inspect and check licence page")

			log.Infof("license res is: %v", licenseInspectResponse.LicenseRespInfo.FeatureInfo)
			var actualLicenseCount int
			for _, info := range licenseInspectResponse.LicenseRespInfo.FeatureInfo {
				actualLicenseCount = actualLicenseCount + int(info.GetConsumed())
			}

			log.Infof("actual license count is %d and expected count is %d:", actualLicenseCount, expected)
			dash.VerifyFatal(actualLicenseCount, licenceCount, "verify cluster inspect and check licence page initial")
		})

		// 3. Create application cluster.
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

		// 4. Verify licence count after cluster added.
		Step("Verify licence count after cluster added.", func() {
			licenseInspectRequestObject := &api.LicenseInspectRequest{
				OrgId: BackupOrgID,
			}

			licenceCount = len(node.GetWorkerNodes()) * 2
			licenseInspectResponse, err := Inst().Backup.InspectLicense(ctx, licenseInspectRequestObject)
			dash.VerifyFatal(err, nil, "verify cluster inspect and check licence page")

			log.Infof("license res is: %v", licenseInspectResponse.LicenseRespInfo.FeatureInfo)
			var actualLicenseCount int
			for _, info := range licenseInspectResponse.LicenseRespInfo.FeatureInfo {
				actualLicenseCount = actualLicenseCount + int(info.GetConsumed())
			}

			log.Infof("actual license count is %d and expected after added cluster: %d", actualLicenseCount)
			dash.VerifyFatal(actualLicenseCount, licenceCount, "Verifing licence count after cluster added")
		})

		// 5. Update kube-config and verify license count
		Step("Update kube-config and verify license count", func() {
			clusterInspectResp, err := Inst().Backup.InspectCluster(ctx, &api.ClusterInspectRequest{
				OrgId: BackupOrgID,
				Name:  SourceClusterName,
				Uid:   clusterUid,
			})

			log.FailOnError(err, "failed to inspect cluster %s with uid %s", SourceClusterName, clusterUid)

			clusterUpdateRequest := &api.ClusterUpdateRequest{
				CreateMetadata: &api.CreateMetadata{
					Name:  SourceClusterName,
					Uid:   clusterUid,
					OrgId: BackupOrgID,
				},
				Kubeconfig:            invalidKubeconfig,
				CloudCredential:       clusterInspectResp.GetCluster().GetCloudCredential(),
				CloudCredentialRef:    clusterInspectResp.GetCluster().GetCloudCredentialRef(),
				PlatformCredentialRef: clusterInspectResp.GetCluster().GetPlatformCredentialRef(),
			}

			log.Infof("Updating kubeconfig for cluster %s from user [%s]", SourceClusterName, ctx)
			_, err = Inst().Backup.UpdateCluster(ctx, clusterUpdateRequest)
			log.Infof("updating cluster with invalid kubeconfig %v", err)

			licenceCount = licenceCount - len(node.GetWorkerNodes())

			err = VerifyLicenseCount(ctx, licenceCount)
			dash.VerifyFatal(err, nil, "Verify Update kube-config and verify license count.")
		})

	})

	JustAfterEach(func() {
		// Cleaning up px-backup cluster
		ctx, err := backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")

		defer EndPxBackupTorpedoTest(scheduledAppContexts)
		log.InfoD("Deleting the deployed apps after the testcase")
		opts := make(map[string]bool)
		opts[SkipClusterScopedObjects] = true
		DestroyApps(scheduledAppContexts, opts)

		CleanupCloudSettingsAndClusters(backupLocationMap, cloudCredName, cloudCredUID, ctx)
	})
})

// This test case verifies if Px-Backup Namespace deployment spec has SOFT_LICENSING_PERIOD set to correct Value
var _ = Describe("{CheckIfDeploymentSpecHasSOFT_LICENSING_PERIODSet}", Label(TestCaseLabelsMap[VerifySoftLicensingPeriodInDeploymentSpec]...), func() {
	var (
		backupDeployment            *appsV1.Deployment
		expectedSoftLicensingPeriod string
		actualSoftLicensingPeriod   string
	)

	JustBeforeEach(func() {
		StartPxBackupTorpedoTest("VerifyIfDeploymentSpecHasSOFT_LICENSING_PERIODSet",
			"Deployment spec has SOFT_LICENSING_PERIOD Set to correct value or not", nil, 300718, ABadgujar, Q2FY24)
		//1.Define what is expected licensing period
		log.InfoD("Define what is expected licensing period")
		expectedSoftLicensingPeriod = "36160"
		log.Infof("Expected licensing period is [%s]", expectedSoftLicensingPeriod)
	})
	It("Deployment spec has SOFT_LICENSING_PERIOD Set to correct value or not", func() {
		//2.Get Px-Backup Namespace and Deployment Spec and soft licensing period from there
		Step("Get Px-Backup Namespace and Deployment Spec and soft licensing period from there", func() {
			log.InfoD("Get Px-Backup Namespace and Deployment Spec and soft licensing period from there")
			pxBackupNS, err := backup.GetPxBackupNamespace()
			log.FailOnError(err, "Getting Px-Backup namespace")
			log.Infof("Get Px-Backup Namespace and Deployment Spec from px-backup namespace [%s]", pxBackupNS)
			backupDeployment, err = apps.Instance().GetDeployment(PxBackupDeployment, pxBackupNS)
			log.FailOnError(err, fmt.Sprintf("Getting px-backup deployment in backup namespace %s", pxBackupNS))
			actualSoftLicensingPeriod = backupDeployment.Spec.Template.Spec.Containers[0].Env[6].Value
			log.InfoD("Actual Soft Licensing Period in Px-Backup Namespace is %v", actualSoftLicensingPeriod)
		})

		//3.Verify if Actual and Expected SOFT_LICENSING_PERIOD values are matching
		Step("Verify if Actual and Expected SOFT_LICENSING_PERIOD values are matching", func() {
			log.InfoD("Verify if Actual and Expected SOFT_LICENSING_PERIOD values are matching")
			Inst().Dash.VerifyFatal(actualSoftLicensingPeriod, expectedSoftLicensingPeriod, fmt.Sprintf("Actual and Expected Values for SOFT_LICENSING_PERIOD are matching - %s", actualSoftLicensingPeriod))
		})
	})
})

// This test case verifies if addition of cluster with nodes entitled for goes through by disabling soft licensing
var _ = Describe("{CheckIfAdditionOfClusterWithNodesEntitledForGoesThroughByDisablingSoftLicensing}", Label(TestCaseLabelsMap[VerifySoftLicensingPeriodInDeploymentSpec]...), func() {
	var (
		backupDeployment        *appsV1.Deployment
		updatedBackupDeployment *appsV1.Deployment
		scheduledAppContexts    []*scheduler.Context
	)

	JustBeforeEach(func() {
		StartPxBackupTorpedoTest("VerifyIfAdditionOfClusterWithNodesEntitledForGoesThroughByDisablingSoftLicensing",
			"Addition of cluster with nodes entitled for goes through by disabling soft licensing", nil, 300719, ABadgujar, Q2FY24)

	})
	It("Disable Soft Licensing Period and Check Deployment Spec", func() {
		//1.Disable Soft Licensing Period and Check Deployment Spec
		Step("Disable Soft Licensing Period and Check Deployment Spec", func() {
			log.InfoD("Get Px-Backup Namespace and Deployment Spec and soft licensing period from there")
			pxBackupNS, err := backup.GetPxBackupNamespace()
			log.FailOnError(err, "Getting Px-Backup namespace")
			log.Infof("Get Px-Backup Namespace and Deployment Spec backup namespace [%s]", pxBackupNS)
			backupDeployment, err = apps.Instance().GetDeployment(PxBackupDeployment, pxBackupNS)
			log.FailOnError(err, fmt.Sprintf("Getting px-backup deployment in backup namespace %s", pxBackupNS))
			log.InfoD("The Deployment Spec looks like before removal of SoftLicensing Period - \n %v", backupDeployment.Spec.Template.Spec.Containers[0].Env)
			for envIndex := len(backupDeployment.Spec.Template.Spec.Containers[0].Env) - 1; envIndex >= 0; envIndex-- {
				log.InfoD("Found SOFT_LICENSING_PERIOD in Deployment Spec")
				if backupDeployment.Spec.Template.Spec.Containers[0].Env[envIndex].Name == "SOFT_LICENSING_PERIOD" {
					backupDeployment.Spec.Template.Spec.Containers[0].Env = append(
						backupDeployment.Spec.Template.Spec.Containers[0].Env[:envIndex],
						backupDeployment.Spec.Template.Spec.Containers[0].Env[envIndex+1:]...,
					)
					break
				}
			}
			updatedBackupDeployment, err = apps.Instance().UpdateDeployment(backupDeployment)
			log.FailOnError(err, "Error Updating Deployment Spec after removing SoftLicensing Period")
			log.InfoD("The Deployment Spec looks like after removal of SoftLicensing Period - \n %v", updatedBackupDeployment.Spec.Template.Spec.Containers[0].Env)
		})

		//2.Backup Pod will be restarted, wait for it to come back up and validate it's status
		Step("Backup Pod will be restarted, wait for it to come back up and validate it's status", func() {
			pxBackupNS, err := backup.GetPxBackupNamespace()
			log.FailOnError(err, "Getting Px-Backup namespace")
			log.InfoD("Wait for Pod to come up")
			time.Sleep(time.Minute * 2)
			backupPodLabel := map[string]string{
				"app": "px-backup",
			}
			log.InfoD("Validate Pod's Status")
			err = ValidatePodByLabel(backupPodLabel, pxBackupNS, 5*time.Minute, 30*time.Second)
			log.FailOnError(err, "Checking if px-backup pod is in running state")
		})

		//3.Registering cluster for backup and add cluster
		Step("Registering cluster for backup and add cluster", func() {
			log.InfoD("Registering cluster for backup and add cluster")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			err = CreateApplicationClusters(BackupOrgID, "", "", ctx)
			dash.VerifyFatal(err, nil, "Creating source and destination cluster")
			clusterStatus, err := Inst().Backup.GetClusterStatus(BackupOrgID, SourceClusterName, ctx)
			log.FailOnError(err, fmt.Sprintf("Fetching [%s] cluster status", SourceClusterName))
			dash.VerifyFatal(clusterStatus, api.ClusterInfo_StatusInfo_Online, fmt.Sprintf("Verifying if [%s] cluster is online", SourceClusterName))
		})
	})
	JustAfterEach(func() {
		defer EndPxBackupTorpedoTest(scheduledAppContexts)
		pxBackupNS, err := backup.GetPxBackupNamespace()
		log.FailOnError(err, "Getting Px-Backup namespace")
		log.Infof("Get Px-Backup Namespace and Deployment Spec backup namespace [%s]", pxBackupNS)
		finalbackupDeployment, err := apps.Instance().GetDeployment(PxBackupDeployment, pxBackupNS)
		log.FailOnError(err, fmt.Sprintf("Getting px-backup deployment in backup namespace %s", pxBackupNS))
		log.InfoD("Adding SOFT_LICENSING_PERIOD back to the Deployment Spec")
		foundSoftLicensing := false
		for envIndex := range finalbackupDeployment.Spec.Template.Spec.Containers[0].Env {
			// Check if the environment variable-SOFT_LICENSING_PERIOD already exists
			if finalbackupDeployment.Spec.Template.Spec.Containers[0].Env[envIndex].Name == "SOFT_LICENSING_PERIOD" {
				foundSoftLicensing = true
				break
			}
		}
		// If the environment variable - SOFT_LICENSING_PERIOD doesn't exist, add it
		if !foundSoftLicensing {
			finalbackupDeployment.Spec.Template.Spec.Containers[0].Env = append(
				finalbackupDeployment.Spec.Template.Spec.Containers[0].Env,
				corev1.EnvVar{
					Name:  "SOFT_LICENSING_PERIOD",
					Value: "36160",
				},
			)

			log.InfoD("Added SOFT_LICENSING_PERIOD with value 36160 to Deployment Spec")
		}
		// Now update the deployment with the modified spec
		finalUpdatedBackupDeployment, err := apps.Instance().UpdateDeployment(finalbackupDeployment)
		log.FailOnError(err, "Error Updating Deployment Spec after removing SoftLicensing Period")
		log.InfoD("The Deployment Spec looks like after Addition of SoftLicensing Period after test case - \n %v", finalUpdatedBackupDeployment.Spec.Template.Spec.Containers[0].Env)

		ctx, err := backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")
		log.InfoD("Deleting the deployed apps after the testcase")
		CleanupCloudSettingsAndClusters(nil, "", "", ctx)

	})
})
