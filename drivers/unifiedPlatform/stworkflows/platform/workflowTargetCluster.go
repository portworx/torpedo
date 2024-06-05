package platform

import (
	"fmt"
	"github.com/portworx/sched-ops/k8s/apps"
	"github.com/portworx/sched-ops/k8s/core"
	k8sutils "github.com/portworx/torpedo/drivers/pds/lib"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/platformLibs"
	"github.com/portworx/torpedo/drivers/utilities"
	"github.com/portworx/torpedo/pkg/log"
	"github.com/portworx/torpedo/pkg/osutils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"time"
)

type WorkflowTargetCluster struct {
	KubeConfig string
	Project    *WorkflowProject
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
	//PDS-Deployments
	PDS_BACKUP_OPERATOR     = "pds-backups-operator"
	PDS_DEPLOYMENT_OPERATOR = "pds-deployments-operator"
	PDS_EXTERNAL_DNS        = "pds-external-dns"
	PDS_MUTATOR             = "pds-mutator"
)

var (
	k8sCore = core.Instance()
)

// RegisterToControlPlane register the target cluster to control plane.
func (targetCluster *WorkflowTargetCluster) RegisterToControlPlane() (*WorkflowTargetCluster, error) {

	var cmd string
	// Get Manifest from API

	clusterName := fmt.Sprintf("Cluster_%v", time.Now().Unix())
	manifest, err := platformLibs.GetManifest(targetCluster.Project.Platform.TenantId, clusterName)
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
		cmd = fmt.Sprintf("echo '%s' > %s && kubectl apply -f %s && rm -f %s", *manifest.Manifest, ManifestPath, ManifestPath, ManifestPath)
		log.Infof("Manifest:\n%v\n", cmd)
		output, _, err := osutils.ExecShell(cmd)
		log.Infof("Output: %v", output)
		if err != nil {
			return targetCluster, fmt.Errorf("Error occured shile installing manifests: %v\n", err)
		}
		log.Infof("Terminal output: %v", output)
	}

	clusterId, err := targetCluster.GetClusterIdByName(clusterName)
	if err != nil {
		return targetCluster, fmt.Errorf("Failed to get clusterId: %v\n", err)
	}

	targetCluster.ClusterUID = clusterId

	log.InfoD("Verify the TargetCluster is connected to Control Plane")
	err = wait.Poll(DefaultRetryInterval, targetClusterHealthTimeOut, func() (bool, error) {
		err := targetCluster.CheckTargetClusterHealth()
		if err != nil {
			return false, fmt.Errorf("Error - [%s]", err.Error())
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
		manifest, err := platformLibs.GetManifest(targetCluster.Project.Platform.TenantId, "")
		if err != nil {
			return fmt.Errorf("Failed while getting platform manifests: %v\n", err)
		}
		cmd = fmt.Sprintf("echo '%s' > %s && kubectl delete -f %s && rm -f %s", *manifest.Manifest, ManifestPath, ManifestPath, ManifestPath)
		output, _, err := osutils.ExecShell(cmd)
		if err != nil {
			return fmt.Errorf("Failed uninstalling Platform manifests: %v\n", err)
		}
		log.Infof("Terminal output: %v", output)
	}
	return nil

}

// ValidatePlatformComponents used to validate all k8s object in pds-system namespace
func (targetCluster *WorkflowTargetCluster) ValidatePdsComponents() error {
	var options metav1.ListOptions
	pdsComponents := []string{PDS_MUTATOR, PDS_EXTERNAL_DNS, PDS_BACKUP_OPERATOR, PDS_DEPLOYMENT_OPERATOR}
	waitErr := wait.Poll(DefaultRetryInterval, targetClusterHealthTimeOut, func() (bool, error) {
		var count = 0
		//gets the available deployments in the platformNamespace
		deploymentList, err := apps.Instance().ListDeployments(platformNamespace, options)
		if err != nil {
			return false, err
		}

		//checks if the list deployments captures the pds deployments aswell
		for _, deployment := range deploymentList.Items {
			if utilities.Contains(pdsComponents, deployment.Name) {
				log.InfoD("Deployment %s found in namespace %s\n", deployment.Name, platformNamespace)
				count++
			}
		}

		//once all the pds deployments are captured by the list deployments return true and break the polling
		if count == len(pdsComponents) {
			return true, nil
		}
		log.Infof("PDS Deployments not available in the %s namespace, Retrying...", platformNamespace)
		return false, nil
	})
	if waitErr != nil {
		return waitErr
	}

	// ValidatePlatformComponents validates all the deployments are up and running
	err := targetCluster.ValidatePlatformComponents()
	if err != nil {
		return err
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

func (targetCluster *WorkflowTargetCluster) InstallPDSAppOnTC(clusterId string) error {
	appName := "pds"
	// Check if PDS tcApp already exists
	objects, err := k8sutils.GetCRObject(platformNamespace, "core.portworx.com", "v1", "targetclusterapplications")
	if err != nil {
		return err
	}
	// Iterate over the CRD objects and print their names.
	for _, object := range objects.Items {
		log.Debugf("Objects created: %v", object.GetName())
		if object.GetName() == appName {
			log.Infof("PDS TCApp already exists in the cluster %v", clusterId)
			return nil
		}
	}
	_, err = platformLibs.InstallApplication(appName, clusterId)
	if err != nil {
		return fmt.Errorf("Failed to install App PDS: %v\n", err)
	}

	log.InfoD("Verify the health of all the deployments in %s namespace", platformNamespace)
	err = targetCluster.ValidatePdsComponents()
	if err != nil {
		return err
	}

	return nil
}

func (targetCluster *WorkflowTargetCluster) GetClusterIdByName(clusterName string) (string, error) {
	var ClusterId string

	waitErr := wait.Poll(DefaultRetryInterval, targetClusterHealthTimeOut, func() (bool, error) {
		tcList, err := platformLibs.ListTargetClusters(targetCluster.Project.Platform.TenantId)
		if err != nil {
			return false, err
		}

		var index int
		log.Infof("All clusters - [%+v]", tcList)
		for index = 0; index < len(tcList.Clusters); index++ {
			log.Infof("Cluster Details - [%+v]", *tcList.Clusters[index].Meta)
			log.Infof("Cluster Name - [%s]", *tcList.Clusters[index].Meta.Name)
			if *tcList.Clusters[index].Meta.Name == clusterName {
				ClusterId = *tcList.Clusters[index].Meta.Uid
				return true, nil
			}
		}
		return false, nil
	})

	if ClusterId == "" {
		return "", fmt.Errorf("Cluster Name not found in list of targetclusters\n")
	}

	return ClusterId, waitErr

}

func (targetCluster *WorkflowTargetCluster) GetTargetCluster() (*automationModels.V1TargetCluster, error) {
	tc, err := platformLibs.GetTargetCluster(targetCluster.ClusterUID)
	if err != nil {
		return nil, err
	}
	return tc, nil
}

func (targetCluster *WorkflowTargetCluster) CheckTargetClusterHealth() error {
	tc, err := targetCluster.GetTargetCluster()
	if err != nil {
		return err
	}
	if string(tc.Status.Phase) != targetClusterHealthOK {
		return fmt.Errorf("Target Cluster found in %v Phase\n", tc.Status.Phase)
	}

	return nil
}
