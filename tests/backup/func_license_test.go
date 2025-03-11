package tests

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	api "github.com/portworx/px-backup-api/pkg/apis/v1"
	"github.com/pure-px/sched-ops/k8s/apps"
	"github.com/pure-px/sched-ops/k8s/core"
	"github.com/pure-px/sched-ops/task"
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

// This test case verifies the license count before and after restarting nodes hosting the px-backup pod.
var _ = Describe("{LicensingCountBeforeAndAfterBackupNodeRestart}", Label(TestCaseLabelsMap[LicensingCountBeforeAndAfterBackupNodeRestart]...), func() {

	var (
		pxbNamespace                  string
		sourceClusterWorkerNodes      []node.Node
		destinationClusterWorkerNodes []node.Node
		totalNumberOfWorkerNodes      []node.Node
		contexts                      []*scheduler.Context
		ctx                           context.Context
		nodeSet                       map[string]node.Node
		labelSelector                 map[string]string
		backupNode                    node.Node
		pxbPods                       *corev1.PodList
		err                           error
	)

	JustBeforeEach(func() {
		StartPxBackupTorpedoTest("LicensingCountBeforeAndAfterBackupNodeRestart",
			"Verifies license count after restarting nodes where px-backup pod is hosted", nil, 300785, Nvettaiyan, Q1FY25)
	})

	It("Verify license count after restarting nodes where px-backup pods are hosted", func() {
		ctx, err = backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")

		Step("Adding source and destination clusters for backup", func() {
			log.InfoD("Adding source and destination clusters for backup")
			err = CreateApplicationClusters(BackupOrgID, "", "", ctx)
			log.FailOnError(err, fmt.Sprintf("Adding source cluster %s and destination cluster %s", SourceClusterName, DestinationClusterName))
		})

		Step("Getting the total number of worker nodes in source and destination cluster", func() {
			log.InfoD("Getting the total number of worker nodes in source and destination cluster")
			sourceClusterWorkerNodes = node.GetWorkerNodes()
			log.InfoD("Total number of worker nodes in source cluster are %v", len(sourceClusterWorkerNodes))
			totalNumberOfWorkerNodes = append(totalNumberOfWorkerNodes, sourceClusterWorkerNodes...)
			log.InfoD("Switching cluster context to destination cluster")
			err = SetDestinationKubeConfig()
			log.FailOnError(err, "Switching context to destination cluster failed")
			destinationClusterWorkerNodes = node.GetWorkerNodes()
			log.InfoD("Total number of worker nodes in destination cluster are %v", len(destinationClusterWorkerNodes))
			totalNumberOfWorkerNodes = append(totalNumberOfWorkerNodes, destinationClusterWorkerNodes...)
			log.InfoD("Total number of worker nodes in source and destination cluster are %v", len(totalNumberOfWorkerNodes))
			log.InfoD("Switching cluster context back to source cluster")
			err = SetSourceKubeConfig()
			log.FailOnError(err, "Switching context to source cluster")
		})

		Step("Verifying the license count before restarting the nodes hosting the px-backup pod", func() {
			log.InfoD("Verifying the license count before restarting the nodes hosting the px-backup pod")
			err = VerifyLicenseConsumedCount(ctx, BackupOrgID, int64(len(totalNumberOfWorkerNodes)))
			dash.VerifyFatal(err, nil, "Verifying license count before restarting the nodes hosting the px-backup pod")
		})

		Step("Restarting node where the px-backup pod is hosted", func() {
			log.InfoD("Restarting node where the px-backup pod is hosted")

			pxbNamespace, err = backup.GetPxBackupNamespace()
			log.FailOnError(err, "Failed to get px-backup namespace")
			log.InfoD("px-backup namespace: %v", pxbNamespace)

			labelSelector = map[string]string{"app": "px-backup"}
			pxbPods, err = core.Instance().GetPods(pxbNamespace, labelSelector)
			log.FailOnError(err, "Failed to get px-backup pods")
			log.InfoD("Fetched px-backup pods")

			nodeSet = make(map[string]node.Node)
			for _, pod := range pxbPods.Items {
				nodeName := pod.Spec.NodeName
				backupNode, err = node.GetNodeByName(nodeName)
				log.FailOnError(err, fmt.Sprintf("Failed to get node %s", nodeName))
				if _, exists := nodeSet[nodeName]; !exists {
					nodeSet[nodeName] = backupNode
					log.InfoD("Found px-backup pod on node %s", backupNode.Name)
				}
			}
			log.InfoD("Total number of nodes hosting px-backup pods: %d", len(nodeSet))

			// Reboot the px-backup nodes concurrently
			var wg sync.WaitGroup
			for _, selectedNode := range nodeSet {
				wg.Add(1)
				go func(selectedNode node.Node) {
					defer wg.Done()
					err = Inst().N.RebootNodeAndWait(selectedNode)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Rebooted node %s, waiting for it to fully recover...", selectedNode.Name))
					err = Inst().S.IsNodeReady(selectedNode)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Node %v is ready after reboot", selectedNode.Name))
				}(selectedNode)
			}
			wg.Wait()

			log.InfoD("Validating if all the pods are up in px-backup namespace")
			err = ValidateAllPodsInPxBackupNamespace()
			dash.VerifyFatal(err, nil, "Successfully validated all pods in px-backup namespace")

		})

		Step("Verifying the license count after restarting the nodes hosting the px-backup pod", func() {
			log.InfoD("Verifying the license count after restarting the nodes hosting the px-backup pod")
			licenseVerification := func() (interface{}, bool, error) {
				err = VerifyLicenseConsumedCount(ctx, BackupOrgID, int64(len(totalNumberOfWorkerNodes)))
				if err != nil {
					log.Warnf("License verification failed, will retry: %v", err)
					return nil, true, err
				}
				return nil, false, nil
			}
			_, err = task.DoRetryWithTimeout(licenseVerification, 30*time.Minute, 5*time.Minute)
			dash.VerifyFatal(err, nil, "Verifying license count after restarting the nodes hosting the px-backup pod")
		})
	})

	JustAfterEach(func() {
		defer EndPxBackupTorpedoTest(contexts)
		err = SetDestinationKubeConfig()
		dash.VerifySafely(err, nil, "Switching context to destination cluster")
		log.InfoD("Switching context to source cluster")
		err = SetSourceKubeConfig()
		dash.VerifySafely(err, nil, "Switching context to source cluster")
		ctx, err = backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")
		CleanupCloudSettingsAndClusters(nil, "", "", ctx)
	})
})

// This test case verifies the license parameters per organization
var _ = Describe("{VerifyGetLicensePerOrganization}", Label(TestCaseLabelsMap[VerifyLicenseParameters]...), func() {
	var (
		licenseConsumedFlag         int64
		licenseStatus               string
		licenseType                 string
		licenseExpiryTime           int64
		licenseStartTime            int64
		licenseExpiryDelta          int64
		backupNodeCount             int64
		expectedLicenseStatus       string
		expectedLicenseConsumedFlag int64
		expectedLicenseExpiryDelta  int64
		expectedbackupNodeCount     int64
		expectedLicenseType         string
	)
	JustBeforeEach(func() {
		StartPxBackupTorpedoTest("VerifyGetLicensePerOrganization",
			"Verify Get License Per Organization", nil, 300747, ABadgujar, Q2FY24)
	})
	It("verify get license per organization", func() {
		//1.Define Expected License Parameters
		Step("Define Expected License Parameters", func() {
			log.InfoD("Define Expected License Parameters")
			expectedLicenseStatus = "Active"
			expectedLicenseConsumedFlag = 0
			expectedLicenseExpiryDelta = 30
			expectedbackupNodeCount = 1000
			expectedLicenseType = "Trial"

		})
		//2.Verify Get Actual License Parameters Per Organization
		Step("Verify Get Actual License Parameters Per Organization", func() {
			log.InfoD("Verify Get Actual License Parameters Per Organization")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			licenseInspectRequestObject := &api.LicenseInspectRequest{
				OrgId: BackupOrgID,
			}
			licenseInspectResponse, err := Inst().Backup.InspectLicense(ctx, licenseInspectRequestObject)
			if err != nil {
				log.FailOnError(err, "Couldn't get Backup License object")
			}
			licenseDetails, err := GetLicenseInfo(licenseInspectResponse)
			if err != nil {
				log.Fatalf("Error extracting license details: %v", err)
			}
			log.InfoD("License Type: %s", licenseDetails.LicenseType)
			log.InfoD("Consumed Flag: %v", licenseDetails.ConsumedFlag)
			log.InfoD("Status: %s", licenseDetails.Status)
			log.InfoD("Expiry Time: %v", time.Unix(licenseDetails.ExpiryTime, 0)) // Convert seconds to time
			log.InfoD("Start Time: %v", time.Unix(licenseDetails.StartTime, 0))   // Convert seconds to time
			log.InfoD("Backup Node Count: %d", licenseDetails.NodeCount)

			licenseConsumedFlag = licenseDetails.ConsumedFlag
			licenseStatus = licenseDetails.Status
			licenseType = licenseDetails.LicenseType
			licenseExpiryTime = licenseDetails.ExpiryTime
			licenseStartTime = licenseDetails.StartTime
			backupNodeCount = licenseDetails.NodeCount
		})
		//3.Verify Different Parameters of License
		Step("Verify Different Parameters of License", func() {
			log.InfoD("Verify Different Parameters of License")
			licenseExpiryDelta = licenseExpiryTime/86400 - licenseStartTime/86400
			if licenseType == expectedLicenseType {
				log.InfoD("Verify if backupNode Count is matching")
				Inst().Dash.VerifyFatal(backupNodeCount, expectedbackupNodeCount, fmt.Sprintf("Actual and Expected Values for backupNode are matching - %v", backupNodeCount))
				log.InfoD("Verify if Expiry and Start time Difference is matching")
				Inst().Dash.VerifyFatal(licenseExpiryDelta, expectedLicenseExpiryDelta, fmt.Sprintf("Actual and Expected Values for Expiry and Start time Difference are matching - %v", licenseExpiryDelta))
				log.InfoD("Verify if License Consumed Flag is matching")
				Inst().Dash.VerifyFatal(licenseConsumedFlag, expectedLicenseConsumedFlag, fmt.Sprintf("Actual and Expected Values for License Consumed Flag are matching - %v", licenseConsumedFlag))
				log.InfoD("Verify if License Status is matching")
				Inst().Dash.VerifyFatal(licenseStatus, expectedLicenseStatus, fmt.Sprintf("Actual and Expected Values for License Status are matching - %v", licenseStatus))
			}
		})
	})
})

// This test case verifies the license count when adding a cluster with an invalid kubeconfig
var _ = Describe("{AddFailedClusterAgainWithInvalidKubeConfig}", Label(TestCaseLabelsMap[AddFailedClusterAgainWithInvalidKubeConfig]...), func() {
	var (
		clusterUid                   string
		clusterStatus                api.ClusterInfo_StatusInfo_Status
		err                          error
		ctx                          context.Context
		expectedLicenseConsumedCount int64 = 0
	)
	JustBeforeEach(func() {
		StartPxBackupTorpedoTest("AddFailedClusterAgainWithInvalidKubeConfig", "Add cluster with invalid kubeconfig and verify license usage count", nil, 300375, Ajisingh, Q1FY25)
		ctx, err = backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")
	})
	It("Add cluster with invalid kubeconfig and verify license usage count", func() {
		// Step 1. Create a Cluster with invlaid kubeconfig
		Step("Add cluster with invalid kubeconfig", func() {
			// Create a Cluster with invlaid kubeconfig
			log.InfoD("Creating a Cluster with Invalid Kubeconfig")
			clusterCreateReq := &api.ClusterCreateRequest{
				CreateMetadata: &api.CreateMetadata{
					Name:  SourceClusterName,
					OrgId: BackupOrgID,
				},
				Kubeconfig: base64.StdEncoding.EncodeToString([]byte(InvalidKubeconfig)),
			}
			_, err := Inst().Backup.CreateCluster(ctx, clusterCreateReq)
			dash.VerifyFatal(strings.Contains(err.Error(), "failed to validate access to the cluster"), true, "Verify the cluster creation")
			clusterUid, err = Inst().Backup.GetClusterUID(ctx, BackupOrgID, SourceClusterName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", SourceClusterName))
			clusterStatus, err = Inst().Backup.GetClusterStatus(BackupOrgID, SourceClusterName, ctx)
			log.FailOnError(err, fmt.Sprintf("Fetching [%s] cluster status", SourceClusterName))
			dash.VerifyFatal(clusterStatus, api.ClusterInfo_StatusInfo_Failed, "Verifying  cluster status")
		})

		// Step 2. Check the license consumption count
		Step("Check the license consumption count", func() {
			log.InfoD("Verifying that license count after adding the invalid kubeconfig")
			err = VerifyLicenseConsumedCount(ctx, BackupOrgID, expectedLicenseConsumedCount)
			dash.VerifyFatal(err, nil, "Verifying that license count after adding the invalid kubeconfig should be 0")
		})

		// Step 3. Add back to the same cluster with invalid kubeconfig and verify license count
		Step("Add back to the same cluster with invalid kubeconfig and verify license count", func() {
			log.InfoD("Update Cluster to point to the invalid Kubeconfig")
			clusterInspectResp, err := Inst().Backup.InspectCluster(ctx, &api.ClusterInspectRequest{
				OrgId: BackupOrgID,
				Name:  SourceClusterName,
				Uid:   clusterUid,
			})
			log.FailOnError(err, "Failed to inspect cluster %s with uid %s", SourceClusterName, clusterUid)
			// Concat the invalidKubeconfig to ensure it appears different from the one above.
			InvalidKubeconfig = InvalidKubeconfig + InvalidKubeconfig
			clusterUpdateRequest := &api.ClusterUpdateRequest{
				CreateMetadata: &api.CreateMetadata{
					Name:  SourceClusterName,
					Uid:   clusterUid,
					OrgId: BackupOrgID,
				},
				Kubeconfig:            base64.StdEncoding.EncodeToString([]byte(InvalidKubeconfig)),
				CloudCredential:       clusterInspectResp.GetCluster().GetCloudCredential(),
				CloudCredentialRef:    clusterInspectResp.GetCluster().GetCloudCredentialRef(),
				PlatformCredentialRef: clusterInspectResp.GetCluster().GetPlatformCredentialRef(),
			}
			log.Infof("Updating kubeconfig for cluster %s from user [%s]", SourceClusterName, ctx)
			_, err = Inst().Backup.UpdateCluster(ctx, clusterUpdateRequest)
			log.Infof("Updating cluster with invalid kubeconfig %v", err)
			// Verify the License count after adding the invalid kubeconfig
			err = VerifyLicenseConsumedCount(ctx, BackupOrgID, expectedLicenseConsumedCount)
			dash.VerifyFatal(err, nil, "Verify update invalid kubeconfig and verify license count.")
		})
	})
	JustAfterEach(func() {
		CleanupCloudSettingsAndClusters(nil, "", "", ctx)
	})
})
