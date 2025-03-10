package tests

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	"github.com/pborman/uuid"
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

// For a failed cluster try adding the cluster back with correct parameters
var _ = Describe("{ClusterAdditionFailWithInvalidThenSucceedWithValidConfig}", Label(TestCaseLabelsMap[AddClusterWithInvalidAndValidKubeConfigAndVerify]...), func() {

	var (
		clusterStatus     api.ClusterInfo_StatusInfo_Status
		backupLocationUID string
		cloudCredName     string
		cloudCredUID      string
		bkpLocationName   string
		providers         []string
		ctx               context.Context
		backupLocationMap map[string]string
		clusterUid        string
	)

	JustBeforeEach(func() {
		StartPxBackupTorpedoTest("ClusterAdditionFailWithInvalidThenSucceedWithValidConfig", "For a failed cluster try adding the cluster back with correct parameters.", nil, 300376, Pingle, Q1FY25)
		backupLocationMap = make(map[string]string)
		var err error
		ctx, err = backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")
		providers = GetBackupProviders()
	})

	// For a failed cluster try adding the cluster back with correct parameters
	It("For a failed cluster try adding the cluster back with correct parameters.", func() {

		// 1.Create cloud credentials and backup location
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

		// 2.Register cluster with invalid kubeconfig
		Step("Register cluster with invalid kubeconfig", func() {
			backupDriver := Inst().Backup
			clusterCreateReq := &api.ClusterCreateRequest{
				CreateMetadata: &api.CreateMetadata{
					Name:  SourceClusterName,
					OrgId: BackupOrgID,
				},
				Kubeconfig: base64.StdEncoding.EncodeToString([]byte(InvalidKubeconfig)),
			}
			_, err := backupDriver.CreateCluster(ctx, clusterCreateReq)
			dash.VerifyFatal(strings.Contains(err.Error(), "failed to validate access to the cluster"), true, "Verify the cluster creation")
			clusterUid, err = Inst().Backup.GetClusterUID(ctx, BackupOrgID, SourceClusterName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", SourceClusterName))
			clusterStatus, err = Inst().Backup.GetClusterStatus(BackupOrgID, SourceClusterName, ctx)
			log.FailOnError(err, fmt.Sprintf("Fetching [%s] cluster status", SourceClusterName))
			dash.VerifyFatal(clusterStatus, api.ClusterInfo_StatusInfo_Failed, "Verifying  cluster status")
		})

		// 3.Update Cluster to point to the valid Kubeconfig
		Step("Update Cluster to point to the valid Kubeconfig", func() {
			log.InfoD("Update Cluster to point to the valid Kubeconfig")
			backupDriver := Inst().Backup
			clusterInspectResp, err := Inst().Backup.InspectCluster(ctx, &api.ClusterInspectRequest{
				OrgId: BackupOrgID,
				Name:  SourceClusterName,
				Uid:   clusterUid,
			})
			log.FailOnError(err, "failed to inspect cluster %s with uid %s", SourceClusterName, clusterUid)

			kubeconfigPath, err := GetSourceClusterConfigPath()
			log.FailOnError(err, "Failed to get source cluster config")
			kubeconfigRaw, err := ioutil.ReadFile(kubeconfigPath)
			log.FailOnError(err, "Failed to read kubeconfig file")

			clusterUpdateRequest := &api.ClusterUpdateRequest{
				CreateMetadata: &api.CreateMetadata{Name: SourceClusterName,
					Uid:   clusterUid,
					OrgId: BackupOrgID,
				},
				Kubeconfig:            base64.StdEncoding.EncodeToString(kubeconfigRaw),
				CloudCredential:       clusterInspectResp.GetCluster().GetCloudCredential(),
				CloudCredentialRef:    clusterInspectResp.GetCluster().GetCloudCredentialRef(),
				PlatformCredentialRef: clusterInspectResp.GetCluster().GetPlatformCredentialRef(),
			}

			_, err = backupDriver.UpdateCluster(ctx, clusterUpdateRequest)
			dash.VerifyFatal(err, nil, "failed to update the cluster with valid kube-config")

			clusterStatus, err = Inst().Backup.GetClusterStatus(BackupOrgID, SourceClusterName, ctx)
			log.FailOnError(err, fmt.Sprintf("Fetching [%s] cluster status", SourceClusterName))
			dash.VerifyFatal(clusterStatus, api.ClusterInfo_StatusInfo_Online, fmt.Sprintf("Verifying if [%s] cluster is online", SourceClusterName))
		})
	})

	JustAfterEach(func() {
		defer EndPxBackupTorpedoTest(nil)
		CleanupCloudSettingsAndClusters(backupLocationMap, cloudCredName, cloudCredUID, ctx)
	})

})

// Check failed cluster object should not toggle between Online/Offline status.
var _ = Describe("{TestClusterStatusPersistentFailureOnInvalidKubeconfig}", Label(TestCaseLabelsMap[ClusterTestLabel]...), func() {

	var (
		ctx context.Context
	)

	JustBeforeEach(func() {
		StartPxBackupTorpedoTest("TestClusterStatusPersistentFailureOnInvalidKubeconfig", "Check failed cluster object should not toggle between Online/Offline status.", nil, 300377, Pingle, Q3FY25)
		var err error
		ctx, err = backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")
	})

	// Check failed cluster object should not toggle between Online/Offline status.
	It("Check failed cluster object should not toggle between Online/Offline status.", func() {
		// 1.Create cluster with invalid kube-config and verify failure
		Step("Register cluster with invalid kubeconfig", func() {
			backupDriver := Inst().Backup
			invalidKubeconfig := "\"\""
			clusterCreateReq := &api.ClusterCreateRequest{
				CreateMetadata: &api.CreateMetadata{
					Name:  SourceClusterName,
					OrgId: BackupOrgID,
				},
				Kubeconfig: base64.StdEncoding.EncodeToString([]byte(invalidKubeconfig)),
			}
			_, err := backupDriver.CreateCluster(ctx, clusterCreateReq)
			dash.VerifyFatal(strings.Contains(err.Error(), "failed to validate access to the cluster"), true, "Verify the cluster creation")
		})

		// 2.Verifying  cluster status remain same failed after some time
		Step("Verifying  cluster status remain same failed after some time", func() {
			waitTimeForClusterStatus := 5 * time.Second
			time.Sleep(waitTimeForClusterStatus)
			clusterStatus, err := Inst().Backup.GetClusterStatus(BackupOrgID, SourceClusterName, ctx)
			log.FailOnError(err, fmt.Sprintf("Fetching [%s] cluster status", SourceClusterName))
			dash.VerifyFatal(clusterStatus, api.ClusterInfo_StatusInfo_Failed, "Verifying  cluster status")
		})
	})

	JustAfterEach(func() {
		defer EndPxBackupTorpedoTest(nil)
		CleanupCloudSettingsAndClusters(nil, "", "", ctx)
	})
})

// List the available clusters and add one cluster as part of discovery
var _ = Describe("{AddClusterFromDiscoveredList}", Label(TestCaseLabelsMap[AddClusterFromDiscoveredList]...), func() {

	var (
		cloudCredName string
		cloudCredUID  string
		providers     []string
		ctx           context.Context
		err           error
		region        string
		clusterName   string
	)

	JustBeforeEach(func() {
		StartPxBackupTorpedoTest("VerifyAddClusterFromDiscoveredList", "Add one cluster as part of discovery", nil, 300362, Pingle, Q1FY25)
		providers = GetBackupProviders()
		ctx, err = backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")
		region = os.Getenv("REGION_FOR_CLUSTER_DISCOVERY")
		clusterName = os.Getenv("CLUSTER_NAME_FOR_CLOUD_DISCOVERY")
	})
	// List the available clusters and add one cluster as part of discovery
	It("Add one cluster as part of discovery", func() {

		// 1.Creating cloud credential
		Step("Creating cloud credentials", func() {
			log.InfoD("Creating cloud credentials")
			for _, provider := range providers {
				cloudCredName = fmt.Sprintf("%s-%s-%v", "cred", provider, time.Now().Unix())
				cloudCredUID = uuid.New()
				err := CreateDiscoveryCloudCredential(provider, cloudCredName, cloudCredUID, BackupOrgID, ctx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of cloud credential named [%s] for org [%s] with [%s] as provider", cloudCredName, BackupOrgID, provider))
			}
		})

		// 3.List the discovery cluster
		Step("List the discovery clusters and verify name passed from Env to add", func() {
			log.InfoD("List discovery Cluster")
			enumerateRequest := PopulateMangeClusterEnumerateRequest(cloudCredName, cloudCredUID, region)
			enumerateResponse, err := Inst().Backup.EnumerateManagedCluster(ctx, enumerateRequest)
			dash.VerifyFatal(err, nil, "enumerate discovery cluster")
			log.Infof("length of clusters %d", len(enumerateResponse.Cluster))
			isFound := false
			for _, clusterObj := range enumerateResponse.Cluster {
				if clusterObj.Name == clusterName {
					log.Infof("cluster name is%s", clusterObj.Name)
					isFound = true
				}
			}
			dash.VerifyFatal(isFound, true, "Verify cluster name same as passed from ENV")
			log.Infof("cluster name is%s", clusterName)
		})

		// 3.Add one cluster as part of discovery
		Step("Create discovery Cluster", func() {
			bulkAddRequest := PopulateMangeClusterBuldAddRequest(cloudCredName, cloudCredUID, region, []string{clusterName})
			bulkAddResponse, err := Inst().Backup.BulkAddManagedCluster(ctx, bulkAddRequest)
			dash.VerifyFatal(err, nil, "Verify creating discovery cluster")
			log.InfoD("Addded discovery cluster%v", bulkAddResponse)
			clusterStatus, err := Inst().Backup.GetClusterStatus(BackupOrgID, clusterName, ctx)
			log.FailOnError(err, fmt.Sprintf("Fetching [%s] cluster status", clusterName))
			log.Infof("cluster status is %s", clusterStatus.String())
			dash.VerifyFatal(clusterStatus, api.ClusterInfo_StatusInfo_Online, fmt.Sprintf("Verifying if [%s] cluster is online", clusterName))
		})
	})

	JustAfterEach(func() {
		defer EndPxBackupTorpedoTest(nil)
		err := DeleteCluster(clusterName, BackupOrgID, ctx, false)
		dash.VerifyFatal(err, nil, "Delete created cluster")
		err = Inst().Backup.WaitForClusterDeletion(ctx, clusterName, BackupOrgID, ClusterDeleteTimeout, ClusterDeleteRetryTime)
		dash.VerifyFatal(err, nil, fmt.Sprintf("waiting for cluster [%s] deletion", clusterName))
		CleanupCloudSettingsAndClusters(nil, cloudCredName, cloudCredUID, ctx)
	})
})

// Try listing clusters with invalid cloud cred and region
var _ = Describe("{ClusterDiscoveryFailureWithInvalidCloudCredsAndRegion}", Label(TestCaseLabelsMap[ClusterDiscoveryFailureWithInvalidCloudCredsAndRegion]...), func() {
	var (
		cloudCredName string
		cloudCredUID  string
		region        string
		ctx           context.Context
		err           error
	)

	JustBeforeEach(func() {
		StartPxBackupTorpedoTest("ClusterDiscoveryFailureWithInvalidCloudCredsAndRegion", "Try listing clusters with invalid cloud cred and region", nil, 300371, Pingle, Q1FY25)
		ctx, err = backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")
		region = os.Getenv("REGION_FOR_CLUSTER_DISCOVERY")
	})

	// List clusters with invalid cloud cred and region and verify app error
	It("Try listing clusters with invalid cloud cred and region", func() {
		Step("list Managed Clusters with invalid cloud cred and region", func() {
			cloudCredName = "wrong-credName"
			cloudCredUID = "wrong-credUid"
			enumerateRequest := PopulateMangeClusterEnumerateRequest(cloudCredName, cloudCredUID, region)
			enumerateResponse, err := Inst().Backup.EnumerateManagedCluster(ctx, enumerateRequest)
			dash.VerifyFatal(strings.Contains(err.Error(), "failed to retrieve cloud credential"), true, "Enumerate cluster with invalid cloud creds")
			log.Infof("Response is:%s", enumerateResponse.String())
		})
	})

	JustAfterEach(func() {
		defer EndPxBackupTorpedoTest(nil)
	})
})

// This Test Case Verifies if we can add cluster object with valid kubeconfig and change it to invalid kubeconfig later with suitable error response.
var _ = Describe("{ModifyTheClusterObjectToPointToInvalidKubeConfigAndVerifyFailure}", Label(TestCaseLabelsMap[AddClusterAndPointToInvalidKubeConfigAndVerifyFailure]...), func() {

	var (
		scheduledAppContexts []*scheduler.Context
		clusterStatus        api.ClusterInfo_StatusInfo_Status
		ctx                  context.Context
		clusterUid           string
	)

	JustBeforeEach(func() {
		StartPxBackupTorpedoTest("ModifyTheClusterObjectToPointToInvalidKubeConfigAndVerifyFailure", "Making the cluster object addition to fail and try to add the same", nil, 300372, ABadgujar, Q3FY25)
		scheduledAppContexts = make([]*scheduler.Context, 0)
		var err error
		ctx, err = backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")
	})

	// Add cluster object with valid kubeconfig and change it to invalid kubeconfig later with suitable error response
	It("Add cluster object with valid kubeconfig and change it to invalid kubeconfig later with suitable error response", func() {

		// 2. Create Valid Source Cluster with Valid KubeConfig
		Step("Create Valid Source Cluster with Valid KubeConfig", func() {
			log.InfoD("Create Valid Source Cluster with Valid KubeConfig")
			err := RegisterCluster(SourceClusterName, "", BackupOrgID, ctx)
			dash.VerifyFatal(err, nil, "Verifying if source-cluster cluster is registered")
			clusterStatus, err = Inst().Backup.GetClusterStatus(BackupOrgID, SourceClusterName, ctx)
			log.FailOnError(err, fmt.Sprintf("Fetching [%s] cluster status", SourceClusterName))
			dash.VerifyFatal(clusterStatus, api.ClusterInfo_StatusInfo_Online, fmt.Sprintf("Verifying if [%s] cluster is online", SourceClusterName))
			clusterUid, err = Inst().Backup.GetClusterUID(ctx, BackupOrgID, SourceClusterName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", SourceClusterName))
		})

		// 3. Update Cluster to point to the invalid Kubeconfig
		Step("Update Cluster to point to the invalid Kubeconfig", func() {
			log.InfoD("Update Cluster to point to the invalid Kubeconfig")
			backupDriver := Inst().Backup
			invalidKubeConfig := "\"\""
			clusterInspectResp, err := Inst().Backup.InspectCluster(ctx, &api.ClusterInspectRequest{
				OrgId: BackupOrgID,
				Name:  SourceClusterName,
				Uid:   clusterUid,
			})
			log.FailOnError(err, "failed to inspect cluster %s with uid %s", SourceClusterName, clusterUid)

			clusterUpdateRequest := &api.ClusterUpdateRequest{
				CreateMetadata: &api.CreateMetadata{Name: SourceClusterName,
					Uid:   clusterUid,
					OrgId: BackupOrgID,
				},
				Kubeconfig:            invalidKubeConfig,
				CloudCredential:       clusterInspectResp.GetCluster().GetCloudCredential(),
				CloudCredentialRef:    clusterInspectResp.GetCluster().GetCloudCredentialRef(),
				PlatformCredentialRef: clusterInspectResp.GetCluster().GetPlatformCredentialRef(),
			}
			log.Infof("Updating kubeconfig for cluster %s from user [admin]", SourceClusterName)
			_, err = backupDriver.UpdateCluster(ctx, clusterUpdateRequest)
			dash.VerifyFatal(strings.Contains(err.Error(), "failed to validate access to the cluster"), true, fmt.Sprintf("Updated [%s] Cluster kubeconfig to invalid kubeconfig", SourceClusterName))

			clusterStatus, err = Inst().Backup.GetClusterStatus(BackupOrgID, SourceClusterName, ctx)
			log.FailOnError(err, fmt.Sprintf("Fetching [%s] cluster status", SourceClusterName))
			dash.VerifyFatal(clusterStatus, api.ClusterInfo_StatusInfo_Failed, fmt.Sprintf("Verifying if [%s] cluster is invalid", SourceClusterName))
		})

	})
	JustAfterEach(func() {
		defer EndPxBackupTorpedoTest(scheduledAppContexts)
		CleanupCloudSettingsAndClusters(nil, "", "", ctx)
	})
})

// This test case verifies that a cluster addition fails with an invalid kubeconfig and is resolved by re-adding the cluster with a valid kubeconfig, transitioning it to Online status.
var _ = Describe("{ClusterAdditionWithInvalidAndValidKubeconfigHandling}", Label(TestCaseLabelsMap[ClusterAdditionWithInvalidAndValidKubeconfigHandling]...), func() {
	var (
		adminCtx             context.Context
		sourceKubeConfigPath string
		clusterStatus        api.ClusterInfo_StatusInfo_Status
		backupDriver         backup.Driver
		clusterName          string
		clusterUid           string
		err                  error
	)

	JustBeforeEach(func() {
		StartPxBackupTorpedoTest("ClusterAdditionWithInvalidAndValidKubeconfigHandling",
			"Verifies that a cluster addition failure due to an invalid kubeconfig is resolved by deleting and re-adding the cluster with a valid kubeconfig, transitioning it to Online status", nil, 300379, Nvettaiyan, Q1FY25)
		backupDriver = Inst().Backup
		adminCtx, err = backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")
		sourceKubeConfigPath, err = GetSourceClusterConfigPath()
		log.FailOnError(err, "Fetching source cluster kubeconfig path")
	})

	It("Ensures that a cluster addition fails with an invalid kubeconfig, but succeeds and transitions to Online status once the cluster is deleted and re-added with a valid kubeconfig", func() {

		// 1. Add a cluster object with an invalid kubeconfig
		Step("Attempt to create cluster using invalid kube-config", func() {
			log.InfoD("Attempt to create cluster using invalid kube-config")
			clusterName = fmt.Sprintf("%s-%v", "cluster-", RandomString(6))
			clusterCreateReq := &api.ClusterCreateRequest{
				CreateMetadata: &api.CreateMetadata{
					Name:  clusterName,
					OrgId: BackupOrgID,
				},
				Kubeconfig: base64.StdEncoding.EncodeToString([]byte(InvalidKubeconfig)),
			}
			_, err = backupDriver.CreateCluster(adminCtx, clusterCreateReq)
			dash.VerifyFatal(strings.Contains(err.Error(), "failed to validate access to the cluster"), true, "Verify the cluster creation")
			clusterUid, err = Inst().Backup.GetClusterUID(adminCtx, BackupOrgID, clusterName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", clusterName))
			clusterReq := &api.ClusterInspectRequest{OrgId: BackupOrgID, Name: clusterName, IncludeSecrets: true, Uid: clusterUid}
			clusterResp, err := backupDriver.InspectCluster(adminCtx, clusterReq)
			log.FailOnError(err, fmt.Sprintf("Fetching [%s] cluster status", clusterName))
			log.Infof("cluster failed with status %v and with reason %s", clusterResp.Cluster.GetStatus().Status, clusterResp.Cluster.GetStatus().Reason)
			dash.VerifyFatal(clusterResp.Cluster.GetStatus().Status, api.ClusterInfo_StatusInfo_Failed, "Verifying  cluster status")
			dash.VerifyFatal(strings.Contains(clusterResp.Cluster.GetStatus().Reason, "failed to validate access to the cluster"), true, "Verify the cluster reason")
		})

		// 2. Delete the failed cluster object
		Step("Delete the failed cluster", func() {
			log.InfoD("Delete the failed cluster")
			err = DeleteCluster(clusterName, BackupOrgID, adminCtx, true)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Successfully initiated the cleanup for the failed cluster: %s", clusterName))
			err = backupDriver.WaitForClusterDeletion(adminCtx, clusterName, BackupOrgID, ClusterDeleteTimeout, ClusterDeleteRetryTime)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Successfully verified that the cluster [%s] is deleted", clusterName))
		})

		// 3. Add the cluster back with a valid kubeconfig
		Step("Create cluster using valid kubeconfig", func() {
			log.InfoD("Creating cluster using valid kubeconfig")
			err = CreateCluster(clusterName, sourceKubeConfigPath, BackupOrgID, "", "", adminCtx)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Adding %s cluster with valid kubeconfig", clusterName))
			clusterStatus, err = Inst().Backup.GetClusterStatus(BackupOrgID, clusterName, adminCtx)
			log.FailOnError(err, fmt.Sprintf("Fetching status of [%s] cluster", clusterName))
			dash.VerifyFatal(clusterStatus, api.ClusterInfo_StatusInfo_Online, fmt.Sprintf("Verifying if [%s] cluster is online", clusterName))
			clusterUid, err = Inst().Backup.GetClusterUID(adminCtx, BackupOrgID, clusterName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching UID for [%s] cluster", clusterUid))
		})
	})

	JustAfterEach(func() {
		defer EndPxBackupTorpedoTest(nil)
		CleanupCloudSettingsAndClusters(nil, "", "", adminCtx)
	})
})

// This test case validates the addition of clusters from the discovery list and verifies that their status is 'Added' after bulk addition
var _ = Describe("{ClusterDiscoveryAndStatusValidation}", Label(TestCaseLabelsMap[ClusterDiscoveryAndStatusValidation]...), func() {
	var (
		providers     []string
		clusterNames  []string
		cloudCredName string
		cloudCredUID  string
		region        string
		ctx           context.Context
		err           error
	)

	JustBeforeEach(func() {
		StartPxBackupTorpedoTest("ClusterDiscoveryAndStatusValidation", "Validating the addition of clusters from the discovered list and confirming their status as Added", nil, 300358, Nvettaiyan, Q1FY25)
		providers = GetBackupProviders()
		ctx, err = backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")
		region = os.Getenv("REGION_FOR_CLUSTER_DISCOVERY")
	})

	It("Ensuring that clusters added from the discovery list are successfully updated with an Added status", func() {
		// 1.Creating cloud credentials
		Step("Creating cloud credentials", func() {
			log.InfoD("Creating cloud credentials")
			for _, provider := range providers {
				cloudCredName = fmt.Sprintf("%s-%s-%v", "cred", provider, RandomString(10))
				cloudCredUID = uuid.New()
				err = CreateDiscoveryCloudCredential(provider, cloudCredName, cloudCredUID, BackupOrgID, ctx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of cloud credential named [%s] for org [%s] with [%s] as provider", cloudCredName, BackupOrgID, provider))
			}
		})

		// 2. Discover the clusters and add them
		Step("Enumerating and Bulk Adding Managed Clusters", func() {
			log.InfoD("Enumerating and Bulk Adding Managed Clusters")
			enumerateRequest := PopulateMangeClusterEnumerateRequest(cloudCredName, cloudCredUID, region)
			enumerateResponse, err := Inst().Backup.EnumerateManagedCluster(ctx, enumerateRequest)
			dash.VerifyFatal(err, nil, "Enumerating discovery clusters")
			log.Infof("Number of discovered clusters: %d", len(enumerateResponse.Cluster))
			for _, cluster := range enumerateResponse.Cluster {
				clusterNames = append(clusterNames, cluster.Name)
			}
			bulkAddRequest := PopulateMangeClusterBuldAddRequest(cloudCredName, cloudCredUID, region, clusterNames)
			_, err = Inst().Backup.BulkAddManagedCluster(ctx, bulkAddRequest)
			dash.VerifyFatal(err, nil, "Successfully added managed clusters")
		})

		// 3. After adding, re-enumerate the clusters and check their status to ensure they are updated to 'Added'
		Step("Re-enumerating the discovery clusters and verifying their status after addition", func() {
			log.InfoD("Re-enumerating the discovery clusters and verifying their status after addition")
			enumerateResponseAfterAddition, err := Inst().Backup.EnumerateManagedCluster(ctx, PopulateMangeClusterEnumerateRequest(cloudCredName, cloudCredUID, region))
			dash.VerifyFatal(err, nil, "Re-enumerating discovery clusters after addition")
			for _, cluster := range enumerateResponseAfterAddition.Cluster {
				dash.VerifyFatal(cluster.Status, api.ManagedClusterObject_Added, fmt.Sprintf("Verifying cluster [%s] status is 'Added'", cluster.Name))
			}
			log.InfoD("All clusters have been successfully added and verified")
		})
	})

	JustAfterEach(func() {
		defer EndPxBackupTorpedoTest(nil)
		for _, clusterName := range clusterNames {
			err = DeleteCluster(clusterName, BackupOrgID, ctx, false)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Delete created cluster [%s]", clusterName))
			err = Inst().Backup.WaitForClusterDeletion(ctx, clusterName, BackupOrgID, ClusterDeleteTimeout, ClusterDeleteRetryTime)
			dash.VerifyFatal(err, nil, fmt.Sprintf("waiting for cluster [%s] deletion", clusterName))
		}
		CleanupCloudSettingsAndClusters(nil, cloudCredName, cloudCredUID, ctx)
	})
})
