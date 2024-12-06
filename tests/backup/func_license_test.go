package tests

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	api "github.com/portworx/px-backup-api/pkg/apis/v1"
	"github.com/pure-px/torpedo/drivers/backup"
	"github.com/pure-px/torpedo/drivers/node"
	"github.com/pure-px/torpedo/drivers/scheduler"
	"github.com/pure-px/torpedo/pkg/log"
	. "github.com/pure-px/torpedo/tests"
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
