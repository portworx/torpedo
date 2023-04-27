package tests

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	api "github.com/portworx/px-backup-api/pkg/apis/v1"
	"github.com/portworx/sched-ops/k8s/core"
	"github.com/portworx/torpedo/drivers/backup"
	"github.com/portworx/torpedo/drivers/node"
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests"
)

// NodeCountForLicensing applies label portworx.io/nobackup=true on any worker node of application cluster and verifies that this worker node is not counted for licensing
var _ = Describe("{NodeCountForLicensing}", func() {
	var (
		sourceClusterWorkerNodes      []node.Node
		destinationClusterWorkerNodes []node.Node
		totalNumberOfWorkerNodes      []node.Node
		srcClusterStatus              api.ClusterInfo_StatusInfo_Status
		destClusterStatus             api.ClusterInfo_StatusInfo_Status
		contexts                      []*scheduler.Context
	)
	JustBeforeEach(func() {
		StartTorpedoTest("NodeCountForLicensing",
			"Verify worker node on application cluster with label portworx.io/nobackup=true is not counted for licensing", nil, 82777)
	})

	It("Verify worker node on application cluster with label portworx.io/nobackup=true is not counted for licensing", func() {
		Step("Registering source and destination clusters for backup", func() {
			log.InfoD("Registering source and destination clusters for backup")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			err = CreateSourceAndDestClusters(orgID, "", "", ctx)
			log.FailOnError(err, fmt.Sprintf("Creating source cluster %s and destination cluster %s", SourceClusterName, destinationClusterName))
			srcClusterStatus, err = Inst().Backup.GetClusterStatus(orgID, SourceClusterName, ctx)
			log.FailOnError(err, fmt.Sprintf("Fetching [%s] cluster status", SourceClusterName))
			dash.VerifyFatal(srcClusterStatus, api.ClusterInfo_StatusInfo_Online, fmt.Sprintf("Verifying if [%s] cluster is online", SourceClusterName))
			destClusterStatus, err = Inst().Backup.GetClusterStatus(orgID, destinationClusterName, ctx)
			log.FailOnError(err, fmt.Sprintf("Fetching [%s] cluster status", destinationClusterName))
			dash.VerifyFatal(destClusterStatus, api.ClusterInfo_StatusInfo_Online, fmt.Sprintf("Verifying if [%s] cluster is online", destinationClusterName))
		})
		Step("Getting the total number of worker nodes in source and destination cluster", func() {
			log.InfoD("Getting the total number of worker nodes in source and destination cluster")
			sourceClusterWorkerNodes = node.GetWorkerNodes()
			log.InfoD("Total number of worker nodes in source cluster are %v", len(sourceClusterWorkerNodes))
			totalNumberOfWorkerNodes = append(totalNumberOfWorkerNodes, sourceClusterWorkerNodes...)
			log.InfoD("Switching cluster context to destination cluster")
			SetDestinationKubeConfig()
			destinationClusterWorkerNodes = node.GetWorkerNodes()
			log.InfoD("Total number of worker nodes in destination cluster are %v", len(destinationClusterWorkerNodes))
			totalNumberOfWorkerNodes = append(totalNumberOfWorkerNodes, destinationClusterWorkerNodes...)
			log.InfoD("Total number of worker nodes in source and destination cluster are %v", len(totalNumberOfWorkerNodes))
			log.InfoD("Switching cluster context back to source cluster")
			err := SetSourceKubeConfig()
			log.FailOnError(err, "Switching context to source cluster")
		})
		Step("Verifying the license count after adding source and destination clusters", func() {
			log.InfoD("Verifying the license count after adding source and destination clusters")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			err = VerifyLicenseConsumedCount(ctx, orgID, int64(len(totalNumberOfWorkerNodes)))
			dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying the license count when source cluster with %d worker nodes and destination cluster with %d worker nodes are added to backup", len(sourceClusterWorkerNodes), len(destinationClusterWorkerNodes)))
		})
		Step("Verify worker node on application cluster with label portworx.io/nobackup=true is not counted for licensing", func() {
			log.InfoD("Applying label portworx.io/nobackup=true to one of the worker node on source cluster and verifying the license count")
			err := Inst().S.AddLabelOnNode(sourceClusterWorkerNodes[0], "portworx.io/nobackup", "true")
			log.FailOnError(err, fmt.Sprintf("Failed to apply label portworx.io/nobackup=true to worker node %v", sourceClusterWorkerNodes[0].Name))
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			err = VerifyLicenseConsumedCount(ctx, orgID, int64(len(totalNumberOfWorkerNodes)-1))
			dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying license count after applying label portworx.io/nobackup=true to node %v", sourceClusterWorkerNodes[0].Name))
			log.InfoD("Switching cluster context to destination cluster")
			SetDestinationKubeConfig()
			log.InfoD("Applying label portworx.io/nobackup=true to one of the worker node on destination cluster and verifying the license count")
			err = Inst().S.AddLabelOnNode(destinationClusterWorkerNodes[0], "portworx.io/nobackup", "true")
			log.FailOnError(err, fmt.Sprintf("Failed to apply label portworx.io/nobackup=true to worker node %v", destinationClusterWorkerNodes[0].Name))
			log.InfoD("Switching cluster context back to source cluster")
			err = SetSourceKubeConfig()
			log.FailOnError(err, "Switching context to source cluster")
			err = VerifyLicenseConsumedCount(ctx, orgID, int64(len(totalNumberOfWorkerNodes)-2))
			dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying license count after applying label portworx.io/nobackup=true to node %v", destinationClusterWorkerNodes[0].Name))
		})
		Step("Removing label portworx.io/nobackup=true from worker nodes and verifying the license count", func() {
			log.InfoD("Removing label from worker node on source cluster on which label was applied earlier and verifying the license count")
			err := Inst().S.RemoveLabelOnNode(sourceClusterWorkerNodes[0], "portworx.io/nobackup")
			log.FailOnError(err, fmt.Sprintf("Failed to remove label portworx.io/nobackup=true from worker node %v", sourceClusterWorkerNodes[0].Name))
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			err = VerifyLicenseConsumedCount(ctx, orgID, int64(len(totalNumberOfWorkerNodes)-1))
			dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying license count after removing label portworx.io/nobackup=true from node %v", sourceClusterWorkerNodes[0].Name))
			log.InfoD("Switching cluster context to destination cluster")
			SetDestinationKubeConfig()
			log.InfoD("Removing label from worker node on destination cluster on which label was applied earlier and verifying the license count")
			err = Inst().S.RemoveLabelOnNode(destinationClusterWorkerNodes[0], "portworx.io/nobackup")
			log.FailOnError(err, fmt.Sprintf("Failed to remove label portworx.io/nobackup=true from worker node %v", destinationClusterWorkerNodes[0].Name))
			log.InfoD("Switching cluster context back to source cluster")
			err = SetSourceKubeConfig()
			log.FailOnError(err, "Switching context to source cluster")
			err = VerifyLicenseConsumedCount(ctx, orgID, int64(len(totalNumberOfWorkerNodes)))
			dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying license count after removing label portworx.io/nobackup=true from node %v", destinationClusterWorkerNodes[0].Name))
		})
	})
	JustAfterEach(func() {
		defer EndPxBackupTorpedoTest(contexts)
		ctx, err := backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")
		err = SetDestinationKubeConfig()
		log.FailOnError(err, "Switching context to destination cluster failed")
		nodeLabels, err := core.Instance().GetLabelsOnNode(destinationClusterWorkerNodes[0].Name)
		if err != nil {
			dash.VerifySafely(err, nil, fmt.Sprintf("Getting label from worker node %v", destinationClusterWorkerNodes[0].Name))
		}
		for key := range nodeLabels {
			if key == "portworx.io/nobackup" {
				log.InfoD("Removing the applied label portworx.io/nobackup=true from worker nodes on destination cluster at the end of the testcase")
				err = Inst().S.RemoveLabelOnNode(destinationClusterWorkerNodes[0], "portworx.io/nobackup")
				dash.VerifySafely(err, nil, fmt.Sprintf("Removing label portworx.io/nobackup=true from worker node %v", destinationClusterWorkerNodes[0].Name))
				break
			}
		}
		log.InfoD("Switching cluster context back to source cluster")
		err = SetSourceKubeConfig()
		log.FailOnError(err, "Switching context to source cluster")
		nodeLabels, err = core.Instance().GetLabelsOnNode(sourceClusterWorkerNodes[0].Name)
		if err != nil {
			dash.VerifySafely(err, nil, fmt.Sprintf("Getting label from worker node %v", sourceClusterWorkerNodes[0].Name))
		}
		for key := range nodeLabels {
			if key == "portworx.io/nobackup" {
				log.InfoD("Removing the applied label portworx.io/nobackup=true from worker nodes on source cluster at the end of the testcase")
				err = Inst().S.RemoveLabelOnNode(sourceClusterWorkerNodes[0], "portworx.io/nobackup")
				dash.VerifySafely(err, nil, fmt.Sprintf("Removing label portworx.io/nobackup=true from worker node %v", sourceClusterWorkerNodes[0].Name))
				break
			}
		}
		CleanupCloudSettingsAndClusters(nil, "", "", ctx)
	})
})

// LicensingCountBeforeAndAfterBackupPodRestart verifies the license count before and after the backup pod restarts
var _ = Describe("{LicensingCountBeforeAndAfterBackupPodRestart}", func() {
	var (
		bkpNamespaces                 []string
		sourceClusterWorkerNodes      []node.Node
		destinationClusterWorkerNodes []node.Node
		totalNumberOfWorkerNodes      []node.Node
		contexts                      []*scheduler.Context
		appContexts                   []*scheduler.Context
	)
	JustBeforeEach(func() {
		StartTorpedoTest("LicensingCountBeforeAndAfterBackupPodRestart",
			"Verifies the license count before and after the backup pod restarts", nil, 82956)
		log.InfoD("Deploy applications needed for taking backup")
		contexts = make([]*scheduler.Context, 0)
		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			taskName := fmt.Sprintf("%s-%d", taskNamePrefix, i)
			appContexts = ScheduleApplications(taskName)
			contexts = append(contexts, appContexts...)
			for _, ctx := range appContexts {
				ctx.ReadinessTimeout = appReadinessTimeout
				namespace := GetAppNamespace(ctx, taskName)
				bkpNamespaces = append(bkpNamespaces, namespace)
			}
		}
	})

	It("Verify the license count before and after the backup pod restarts", func() {
		ctx, err := backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")
		Step("Adding source and destination clusters for backup", func() {
			log.InfoD("Adding source and destination clusters for backup")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			err = CreateSourceAndDestClusters(orgID, "", "", ctx)
			log.FailOnError(err, fmt.Sprintf("Registering source cluster %s and destination cluster %s", SourceClusterName, destinationClusterName))
		})
		Step("Getting the total number of worker nodes in source and destination cluster", func() {
			log.InfoD("Getting the total number of worker nodes in source and destination cluster")
			sourceClusterWorkerNodes = node.GetWorkerNodes()
			log.InfoD("Total number of worker nodes in source cluster are %v", len(sourceClusterWorkerNodes))
			totalNumberOfWorkerNodes = append(totalNumberOfWorkerNodes, sourceClusterWorkerNodes...)
			log.InfoD("Switching cluster context to destination cluster")
			err := SetDestinationKubeConfig()
			log.FailOnError(err, "Switching context to destination cluster failed")
			destinationClusterWorkerNodes = node.GetWorkerNodes()
			log.InfoD("Total number of worker nodes in destination cluster are %v", len(destinationClusterWorkerNodes))
			totalNumberOfWorkerNodes = append(totalNumberOfWorkerNodes, destinationClusterWorkerNodes...)
			log.InfoD("Total number of worker nodes in source and destination cluster are %v", len(totalNumberOfWorkerNodes))
			log.InfoD("Switching cluster context back to source cluster")
			err = SetSourceKubeConfig()
			log.FailOnError(err, "Switching context to source cluster")
		})
		Step("Verifying the license count after adding source and destination clusters", func() {
			log.InfoD("Verifying the license count after adding source and destination clusters")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			err = VerifyLicenseConsumedCount(ctx, orgID, int64(len(totalNumberOfWorkerNodes)))
			dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying the license count when source cluster with %d worker nodes and destination cluster with %d worker nodes are added to backup", len(sourceClusterWorkerNodes), len(destinationClusterWorkerNodes)))
		})
		Step("Verify worker nodes on application cluster with label portworx.io/nobackup=true is not counted for licensing before pod restart", func() {
			log.InfoD("Applying label portworx.io/nobackup=true to one of the worker node on source cluster and destination cluster and verifying the license count")
			err := Inst().S.AddLabelOnNode(sourceClusterWorkerNodes[0], "portworx.io/nobackup", "true")
			log.FailOnError(err, fmt.Sprintf("Failed to apply label portworx.io/nobackup=true to worker node %v", sourceClusterWorkerNodes[0].Name))
			err = Inst().S.AddLabelOnNode(destinationClusterWorkerNodes[0], "portworx.io/nobackup", "true")
			log.FailOnError(err, fmt.Sprintf("Failed to apply label portworx.io/nobackup=true to worker node %v", destinationClusterWorkerNodes[0].Name))
			err = VerifyLicenseConsumedCount(ctx, orgID, int64(len(totalNumberOfWorkerNodes)-2))
			dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying license count after applying label portworx.io/nobackup=true to node %v", sourceClusterWorkerNodes[0].Name))
		})
		Step("Restart all the backup pod and wait for it to come up", func() {
			pxbNamespace, err := backup.GetPxBackupNamespace()
			dash.VerifyFatal(err, nil, "Getting px-backup namespace")
			err = DeletePodWithLabelInNamespace(pxbNamespace, nil)
			dash.VerifyFatal(err, nil, "Restart all the backup pods")
			log.InfoD("Validate if all the backup pods are up")
			pods, err := core.Instance().GetPods(pxbNamespace, nil)
			dash.VerifyFatal(err, nil, "Getting all the backup pods")
			for _, pod := range pods.Items {
				err = core.Instance().ValidatePod(&pod, podReadyTimeout, podReadyRetryTime)
				log.FailOnError(err, fmt.Sprintf("Failed to validate pod [%s]", pod.GetName()))
			}
		})
		Step("Verify the license count after pod restart", func() {
			err = VerifyLicenseConsumedCount(ctx, orgID, int64(len(totalNumberOfWorkerNodes)-2))
			dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying license count after applying label portworx.io/nobackup=true to node %v", sourceClusterWorkerNodes[0].Name))
		})
		Step("Label all the remaining worker nodes after pod restart", func() {
			log.InfoD("Applying label portworx.io/nobackup=true to all the remaining worker nodes on source cluster after pod restart")
			for _, workerNode := range sourceClusterWorkerNodes[1:] {
				err := Inst().S.AddLabelOnNode(workerNode, "portworx.io/nobackup", "true")
				log.FailOnError(err, fmt.Sprintf("Failed to apply label portworx.io/nobackup=true to source cluster worker node %v", workerNode.Name))
			}
			log.InfoD("Applying label portworx.io/nobackup=true to all the remaining worker nodes on destination cluster after pod restart")
			for _, workerNode := range sourceClusterWorkerNodes[1:] {
				err := Inst().S.AddLabelOnNode(workerNode, "portworx.io/nobackup", "true")
				log.FailOnError(err, fmt.Sprintf("Failed to apply label portworx.io/nobackup=true to source cluster worker node %v", workerNode.Name))
			}
		})
		Step("Restart all the backup pod again and wait for it to come up", func() {
			pxbNamespace, err := backup.GetPxBackupNamespace()
			dash.VerifyFatal(err, nil, "Getting px-backup namespace")
			err = DeletePodWithLabelInNamespace(pxbNamespace, nil)
			dash.VerifyFatal(err, nil, "Restart all the backup pods")
			log.InfoD("Validate if all the backup pods are up")
			pods, err := core.Instance().GetPods(pxbNamespace, nil)
			dash.VerifyFatal(err, nil, "Getting all the backup pods")
			for _, pod := range pods.Items {
				err = core.Instance().ValidatePod(&pod, podReadyTimeout, podReadyRetryTime)
				log.FailOnError(err, fmt.Sprintf("Failed to validate pod [%s]", pod.GetName()))
			}
		})
		Step("Verify the license count again after pod restart with all the worker nodes labelled portworx.io/nobackup=true", func() {
			err = VerifyLicenseConsumedCount(ctx, orgID, 0)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying license count after applying label portworx.io/nobackup=true to node %v", sourceClusterWorkerNodes[0].Name))
		})
		Step("Removing label portworx.io/nobackup=true from all the worker nodes from both source and destination cluster", func() {
			log.InfoD("Removing label from all worker nodes on source cluster")
			for _, workerNode := range sourceClusterWorkerNodes {
				err := Inst().S.RemoveLabelOnNode(workerNode, "portworx.io/nobackup")
				log.FailOnError(err, fmt.Sprintf("Failed to remove label portworx.io/nobackup=true from worker node %v", workerNode.Name))
			}
			log.InfoD("Removing label from all worker nodes on destination cluster")
			for _, workerNode := range destinationClusterWorkerNodes {
				err := Inst().S.RemoveLabelOnNode(workerNode, "portworx.io/nobackup")
				log.FailOnError(err, fmt.Sprintf("Failed to remove label portworx.io/nobackup=true from worker node %v", workerNode.Name))
			}
		})
		Step("Verify the license count when no worker nodes are labelled portworx.io/nobackup=true", func() {
			err = VerifyLicenseConsumedCount(ctx, orgID, int64(len(totalNumberOfWorkerNodes)))
			dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying license count when no worker nodes are labelled portworx.io/nobackup=true"))
		})

	})
	JustAfterEach(func() {
		defer EndPxBackupTorpedoTest(contexts)
		ctx, err := backup.GetAdminCtxFromSecret()
		dash.VerifySafely(err, nil, "Fetching px-central-admin ctx")
		err = SetDestinationKubeConfig()
		dash.VerifySafely(err, nil, "Switching context to destination cluster")
		log.InfoD("Removing label portworx.io/nobackup=true from all worker nodes on destination cluster if present")
		for _, workerNode := range destinationClusterWorkerNodes {
			err = RemoveLabelFromNodesIfPresent(workerNode, "portworx.io/nobackup")
			dash.VerifySafely(err, nil, fmt.Sprintf("Removing label portworx.io/nobackup=true from worker node %s", workerNode))
		}
		log.InfoD("Switching cluster context back to source cluster")
		err = SetSourceKubeConfig()
		dash.VerifySafely(err, nil, "Switching context to source cluster")
		log.InfoD("Removing label portworx.io/nobackup=true from all worker nodes on source cluster if present")
		for _, workerNode := range sourceClusterWorkerNodes {
			err = RemoveLabelFromNodesIfPresent(workerNode, "portworx.io/nobackup")
			dash.VerifySafely(err, nil, fmt.Sprintf("Removing label portworx.io/nobackup=true from worker node %s", workerNode))
		}
		CleanupCloudSettingsAndClusters(nil, "", "", ctx)
	})
})
