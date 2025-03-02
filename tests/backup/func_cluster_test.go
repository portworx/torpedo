package tests

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	api "github.com/portworx/px-backup-api/pkg/apis/v1"
	"github.com/pure-px/torpedo/drivers/backup"
	"github.com/pure-px/torpedo/drivers/scheduler"
	"github.com/pure-px/torpedo/pkg/log"
	. "github.com/pure-px/torpedo/tests"
)

// This test case verifies that both admin and infra-admin users can add a cluster with the same name and have the correct ownership
var _ = Describe("{ClusterAdditionwithDifferentUserRoles}", Label(TestCaseLabelsMap[ClusterTestLabel]...), func() {

	var (
		users                []string
		infraAdminUserName   string
		adminCtx             context.Context
		infraAdminCtx        context.Context
		infraAdminUserID     string
		clusterName          string
		cluster1Uid          string
		cluster2Uid          string
		sourceKubeConfigPath string
		err                  error
	)

	JustBeforeEach(func() {
		StartPxBackupTorpedoTest("ClusterAdditionwithDifferentUserRoles",
			"Verifies that both admin and infra-admin users can add a cluster with the same name", nil, 300480, Nvettaiyan, Q1FY25)
		adminCtx, err = backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")
		sourceKubeConfigPath, err = GetSourceClusterConfigPath()
		log.FailOnError(err, "Fetching source cluster kubeconfig path")
	})

	It("Verifies that both admin and infra-admin users can add a cluster with the same name", func() {
		clusterName = fmt.Sprintf("cluster-%s", RandomString(5))
		Step("Create User with Infra Admin Roles", func() {
			log.InfoD("Creating User with Infra Admin Roles")
			users = CreateUsers(1)
			infraAdminUserName = users[0]
			// Assign Infra Admin role
			err = backup.AddRoleToUser(infraAdminUserName, backup.InfrastructureOwner, fmt.Sprintf("Adding Infra Admin role to %s", infraAdminUserName))
			log.FailOnError(err, "Failed to add Infra Admin role for user - %s", infraAdminUserName)
			infraAdminUserID, err = backup.FetchIDOfUser(infraAdminUserName)
			log.FailOnError(err, "Failed to fetch user id for infra-admin - %s", infraAdminUserName)
			log.InfoD("Created Infra-Admin user: %s with ID: %s", infraAdminUserName, infraAdminUserID)
		})

		Step("Create cluster using infra-admin", func() {
			log.InfoD("Creating cluster using infra-admin")
			infraAdminCtx, err = backup.GetNonAdminCtx(infraAdminUserName, CommonPassword)
			log.FailOnError(err, "Fetching infra-admin Context")
			err = CreateCluster(clusterName, sourceKubeConfigPath, BackupOrgID, "", "", infraAdminCtx)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Adding %s cluster using infra-admin", clusterName))
			cluster1Uid, err = Inst().Backup.GetClusterUID(infraAdminCtx, BackupOrgID, clusterName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid for infra-admin", cluster1Uid))
		})

		Step("Create cluster using admin", func() {
			log.InfoD("Creating cluster using admin")
			err = CreateCluster(clusterName, sourceKubeConfigPath, BackupOrgID, "", "", adminCtx)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Adding %s cluster", clusterName))
			cluster2Uid, err = Inst().Backup.GetClusterUID(adminCtx, BackupOrgID, clusterName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", cluster2Uid))
		})

		Step("Verify both clusters are listed with respective owners", func() {
			log.InfoD("Verifying that both clusters are listed with the correct owners")

			// Fetch the list of clusters using the admin context
			log.InfoD("Fetching the list of clusters using the admin context")
			clusters, err := Inst().Backup.EnumerateAllCluster(adminCtx, &api.ClusterEnumerateRequest{
				OrgId: BackupOrgID,
			})
			log.FailOnError(err, "Failed to fetch admin cluster list")

			foundInfraAdminCluster := false
			foundAdminCluster := false

			// Loop through clusters and verify that both the infra-admin and admin clusters are listed
			log.InfoD("length of clusters %v", len(clusters.Clusters))
			for _, cluster := range clusters.Clusters {
				if cluster.Uid == cluster1Uid && cluster.Metadata.Ownership.Owner == infraAdminUserID {
					foundInfraAdminCluster = true
				}
				if cluster.Uid == cluster2Uid {
					foundAdminCluster = true
				}
			}
			// Verify that both clusters are found and have the correct owners
			dash.VerifyFatal(foundInfraAdminCluster, true, "Infra-admin cluster was successfully found with correct ownership")
			dash.VerifyFatal(foundAdminCluster, true, "Admin cluster was successfully found with correct ownership")
		})
	})

	JustAfterEach(func() {
		log.Infof("Cleaning up user")
		err = backup.DeleteUser(infraAdminUserName)
		log.FailOnError(err, "Error deleting user %v", infraAdminUserName)
		CleanupCloudSettingsAndClusters(nil, "", "", adminCtx)
	})
})

// Check correct failure status is captured in cluster object when it's failed to be added.
var _ = Describe("{AddClusterWithInvalidKubeConfigAndVerifyFailure}", Label(TestCaseLabelsMap[AddClusterWithInvalidKubeConfigAndVerifyFailure]...), func() {

	var (
		scheduledAppContexts []*scheduler.Context
		ctx                  context.Context
		clusterName          string
		clusterUid           string
	)

	JustBeforeEach(func() {
		StartPxBackupTorpedoTest("AddClusterWithInvalidKubeConfigAndVerifyFailure", "Check correct failure status is captured in cluster object when it's failed to be added.", nil, 300378, Pingle, Q3FY25)
		scheduledAppContexts = make([]*scheduler.Context, 0)
		var err error
		ctx, err = backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")
	})

	// Check correct failure status is captured in cluster object when it's failed to be added.
	It("Check correct failure status is captured in cluster object when it's failed to be added.", func() {
		// 1. Create cluster with invalid kube-config and verify failure
		Step("Register cluster with invalid kubeconfig", func() {
			backupDriver := Inst().Backup
			clusterName = fmt.Sprintf("%s-%v", "cluster-", RandomString(6))
			clusterCreateReq := &api.ClusterCreateRequest{
				CreateMetadata: &api.CreateMetadata{
					Name:  clusterName,
					OrgId: BackupOrgID,
				},
				Kubeconfig: base64.StdEncoding.EncodeToString([]byte(InvalidKubeconfig)),
			}
			_, err := backupDriver.CreateCluster(ctx, clusterCreateReq)
			dash.VerifyFatal(strings.Contains(err.Error(), "failed to validate access to the cluster"), true, "Verify the cluster creation")
			clusterUid, err = Inst().Backup.GetClusterUID(ctx, BackupOrgID, clusterName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", clusterName))
			clusterReq := &api.ClusterInspectRequest{OrgId: BackupOrgID, Name: clusterName, IncludeSecrets: true, Uid: clusterUid}
			clusterResp, err := backupDriver.InspectCluster(ctx, clusterReq)
			log.FailOnError(err, fmt.Sprintf("Fetching [%s] cluster status", clusterName))
			log.Infof("cluster failed with status %v and with reason %s", clusterResp.Cluster.GetStatus().Status, clusterResp.Cluster.GetStatus().Reason)
			dash.VerifyFatal(clusterResp.Cluster.GetStatus().Status, api.ClusterInfo_StatusInfo_Failed, "Verifying  cluster status")
			dash.VerifyFatal(strings.Contains(clusterResp.Cluster.GetStatus().Reason, "failed to validate access to the cluster"), true, "Verify the cluster reason")
		})
	})

	JustAfterEach(func() {
		defer EndPxBackupTorpedoTest(scheduledAppContexts)
		err := DeleteCluster(clusterName, BackupOrgID, ctx, false)
		dash.VerifySafely(err, nil, "Delete created cluster")
		CleanupCloudSettingsAndClusters(nil, "", "", ctx)
	})

})
