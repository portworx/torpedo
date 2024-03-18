package stworkflows

import (
	"fmt"
	"github.com/portworx/sched-ops/k8s/apps"
	"github.com/portworx/sched-ops/k8s/core"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/platformLibs"
	"github.com/portworx/torpedo/pkg/log"
	"github.com/portworx/torpedo/pkg/osutils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"strings"
	"time"
)

type WorkflowTargetCluster struct {
	KubeConfig string
	Platform   WorkflowPlatform
	Project    WorkflowProject
	ClusterUID string
}

const (
	// DefaultRetryInterval default time to retry
	DefaultRetryInterval = 10 * time.Second
	// DefaultPdsPodsTimeOut default timeout
	DefaultPdsPodsTimeOut = 15 * time.Minute
	// PDSNamespace PDS
	platformNamespace          = "px-system"
	targetClusterHealthOK      = "CONNECTED"
	targetClusterHealthTimeOut = 5 * time.Minute
	pxTargetSecret             = "px-target-cluster-secret"
	ManifestPath               = "/tmp/manifest.yaml"
)

var (
	k8sCore = core.Instance()
)

// RegisterToControlPlane register the target cluster to control plane.
func (targetCluster *WorkflowTargetCluster) RegisterToControlPlane() (*WorkflowTargetCluster, error) {

	var cmd string
	// Get Manifest from API

	clusterName := fmt.Sprintf("Cluster_%v", time.Now().Unix())
	manifest, err := platformLibs.GetManifest(targetCluster.Platform.TenantId, clusterName)
	if err != nil {
		return targetCluster, fmt.Errorf("Failed while getting Manifests: %v\n", err)
	}

	isRegistered := false
	pods, err := k8sCore.GetPods(platformNamespace, nil)
	if err != nil {
		return targetCluster, fmt.Errorf("Failed while getting the pods on %v Namespace: %v\n", platformNamespace, err)
	}

	if len(pods.Items) > 0 {
		log.InfoD("Target cluster is already registered to control plane.")
		isRegistered = true
		// Getting cluster name from the pxTargetSecret in platformNamespace
		secretData, err := core.Instance().GetSecret(pxTargetSecret, platformNamespace)
		if err != nil {
			return targetCluster, fmt.Errorf("Failed while getting px-target-cluster-secret: %v\n", err)
		}
		clusterName = string(secretData.Data["target_cluster_name"])
		log.Infof("ClusterName is [%s]", clusterName)
	}

	if !isRegistered {
		log.InfoD("Installing Manifests..")
		cmd = fmt.Sprintf("echo '%v' > %s && kubectl apply -f %s && rm -f %s", manifest.Manifest, ManifestPath, ManifestPath, ManifestPath)
		log.Infof("Manifest:\n%v\n", cmd)
		output, _, err := osutils.ExecShell(cmd)
		if err != nil {
			return targetCluster, fmt.Errorf("Error occured shile installing manifests: %v\n", err)
		}
		log.Infof("Terminal output: %v", output)
	}

	log.InfoD("Verify the TargetCluster is connected to Control Plane")
	err = wait.Poll(DefaultRetryInterval, targetClusterHealthTimeOut, func() (bool, error) {
		err := targetCluster.GetTargetClusterHealth(clusterName)
		if err != nil {
			return false, nil
		}
		return true, nil
	})

	log.InfoD("Verify the health of all the deployments in %s namespace", platformNamespace)
	err = wait.Poll(DefaultRetryInterval, targetClusterHealthTimeOut, func() (bool, error) {
		err := targetCluster.ValidatePlatformComponents()
		if err != nil {
			return false, nil
		}
		return true, nil
	})

	clusterId, err := targetCluster.GetClusterIdByName(clusterName)
	if err != nil {
		return targetCluster, fmt.Errorf("Failed to get clusterId: %v\n", err)
	}

	targetCluster.ClusterUID = clusterId

	return targetCluster, nil
}

// DeregisterFromControlPlane de-register the target cluster from control plane.
func (targetCluster *WorkflowTargetCluster) DeregisterFromControlPlane() error {
	var cmd string

	// TODO: Currently we are checking for pods in platformNameSpace. Probably  we need to check for deployments
	pods, err := k8sCore.GetPods(platformNamespace, nil)
	if err != nil {
		return fmt.Errorf("Failed while getting the pods on %v Namespace: %v\n", platformNamespace, err)
	}

	if len(pods.Items) > 0 {
		log.InfoD("Uninstalling Manifests ...")
		// Get Manifest from API
		manifest, err := platformLibs.GetManifest(targetCluster.Platform.TenantId, "")
		if err != nil {
			return fmt.Errorf("Failed while getting platform manifests: %v\n", err)
		}
		cmd = fmt.Sprintf("echo '%v' > %s && kubectl delete -f %s && rm -f %s", manifest.Manifest, ManifestPath, ManifestPath, ManifestPath)
		output, _, err := osutils.ExecShell(cmd)
		if err != nil {
			return fmt.Errorf("Failed uninstalling Platform manifests: %v\n", err)
		}
		log.Infof("Terminal output: %v", output)
	}
	return nil

}

// ValidatePlatformComponents used to validate all k8s object in pds-system namespace
func (targetCluster *WorkflowTargetCluster) ValidatePlatformComponents() error {
	var options metav1.ListOptions
	deploymentList, err := apps.Instance().ListDeployments(platformNamespace, options)
	if err != nil {
		return err
	}
	log.Infof("There are %d deployments present in the namespace %s", len(deploymentList.Items), platformNamespace)
	for _, deployment := range deploymentList.Items {
		err = apps.Instance().ValidateDeployment(&deployment, DefaultPdsPodsTimeOut, DefaultRetryInterval)
		if err != nil {
			return err
		}
	}
	return nil
}

func (targetCluster *WorkflowTargetCluster) InstallPDSAppOnTC() error {

	availableApps, err := platformLibs.ListAvailableApplicationsForTenant(targetCluster.ClusterUID, targetCluster.Platform.TenantId)
	if err != nil {
		return fmt.Errorf("Failed to get list of available Apps: %v\n", err)
	}
	var index int
	for index = 0; index < len(availableApps); index++ {
		if strings.Contains("PDS", *availableApps[index].Meta.Name) {
			pdsApp := availableApps[index].Meta
			appName := *pdsApp.Name
			appVersion := *pdsApp.ResourceVersion
			_, err := platformLibs.InstallApplication(appName, appVersion, targetCluster.ClusterUID)
			if err != nil {
				return fmt.Errorf("Failed to install App PDS: %v\n", err)
			}
			return nil
		}
	}
	return fmt.Errorf("PDS App not found in catalog")
}

func (targetCluster *WorkflowTargetCluster) GetClusterIdByName(clusterName string) (string, error) {
	tcList, err := platformLibs.ListTargetClusters(targetCluster.Platform.TenantId)
	if err != nil {
		return "", err
	}

	var index int
	for index = 0; index < len(tcList.Clusters); index++ {
		if *tcList.Clusters[index].Meta.Name == clusterName {
			return *tcList.Clusters[index].Meta.Uid, nil
		}
	}
	return "", fmt.Errorf("Cluster Name not found in list of targetclusters\n")

}

func (targetCluster *WorkflowTargetCluster) GetTargetClusterHealth(clusterName string) error {
	tcList, err := platformLibs.ListTargetClusters(targetCluster.Platform.TenantId)
	if err != nil {
		return err
	}
	var index int
	for index = 0; index < len(tcList.Clusters); index++ {
		if *tcList.Clusters[index].Meta.Name == clusterName {
			phase := tcList.Clusters[index].Status.Phase
			if string(*phase) != targetClusterHealthOK {
				return fmt.Errorf("Target Cluster found in %v Phase\n", phase)
			} else {
				return nil
			}
		}
	}
	return fmt.Errorf("Cluster Name not found in list of targetclusters\n")
}
