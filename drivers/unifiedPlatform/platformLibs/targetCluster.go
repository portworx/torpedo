package platformLibs

import (
	"context"
	"fmt"
	"github.com/jinzhu/copier"
	"github.com/portworx/sched-ops/k8s/apps"
	"github.com/portworx/sched-ops/k8s/core"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
	"github.com/portworx/torpedo/pkg/log"
	"github.com/portworx/torpedo/pkg/osutils"
	platformv1 "github.com/pure-px/platform-api-go-client/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"strings"
	"time"
)

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
)

var (
	k8sCore = core.Instance()
)

// TargetCluster structure
type TargetCluster struct {
	kubeConfig string
}

// type proxyConfig utils.ProxyConfig
// type customRegistryConfig utils.CustomRegistryConfig

// RegisterToControlPlane register the target cluster to control plane.
func (targetCluster *TargetCluster) RegisterToControlPlane(platformVersion string, tenantId string) (string, error) {
	var cmd string
	// Get Manifest from API

	clusterName := fmt.Sprintf("Cluster_%v", time.Now().Unix())
	manifest, err := GetManifest(tenantId, clusterName)
	if err != nil {
		return "", fmt.Errorf("Failed while getting Manifests: %v\n", err)
	}

	isRegistered := false
	pods, err := k8sCore.GetPods(platformNamespace, nil)
	if err != nil {
		return "", fmt.Errorf("Failed while getting the pods on %v Namespace: %v\n", platformNamespace, err)
	}

	if len(pods.Items) > 0 {
		log.InfoD("Target cluster is already registered to control plane.")
		isRegistered = true
		// Getting cluster name from the pxTargetSecret in platformNamespace
		secretData, err := core.Instance().GetSecret(pxTargetSecret, platformNamespace)
		if err != nil {
			return "", fmt.Errorf("Failed while getting px-target-cluster-secret: %v\n", err)
		}
		clusterName = string(secretData.Data["target_cluster_name"])
		log.Infof("ClusterName is [%s]", clusterName)
	}

	if !isRegistered {
		log.InfoD("Installing Manifests %v", platformVersion)
		cmd = fmt.Sprintf("echo '%v' > /tmp/manifest.yaml && kubectl apply -f /tmp/manifest.yaml && rm -f /tmp/manifest.yaml", manifest)
		log.Infof("Manifest:\n%v\n", cmd)
		output, _, err := osutils.ExecShell(cmd)
		if err != nil {
			return "", fmt.Errorf("Error occured shile installing manifests: %v\n", err)
		}
		log.Infof("Terminal output: %v", output)
	}

	log.InfoD("Verify the TargetCluster is connected to Control Plane")
	err = wait.Poll(DefaultRetryInterval, targetClusterHealthTimeOut, func() (bool, error) {
		err := getTargetClusterHealth(clusterName, tenantId)
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

	clusterId, err := GetClusterIdByName(clusterName, tenantId)
	if err != nil {
		return "", fmt.Errorf("Failed to get clusterId: %v\n", err)
	}
	return clusterId, nil
}

// ValidatePlatformComponents used to validate all k8s object in pds-system namespace
func (targetCluster *TargetCluster) ValidatePlatformComponents() error {
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

// DeregisterFromControlPlane de-register the target cluster from control plane.
func (targetCluster *TargetCluster) DeregisterFromControlPlane(platformVersion string, tenantId string) error {
	var cmd string

	// TODO: Currently we are checking for pods in platformNameSpace. Probably  we need to check for deployments
	pods, err := k8sCore.GetPods(platformNamespace, nil)
	if err != nil {
		return fmt.Errorf("Failed while getting the pods on %v Namespace: %v\n", platformNamespace, err)
	}

	if len(pods.Items) > 0 {
		log.InfoD("Uninstalling Manifests %v", platformVersion)
		// Get Manifest from API
		manifest, err := GetManifest(tenantId, "")
		if err != nil {
			return fmt.Errorf("Failed while getting platform manifests: %v\n", err)
		}
		cmd = fmt.Sprintf("echo '%v' > /tmp/manifest.yaml && kubectl delete -f /tmp/manifest.yaml && rm -f /tmp/manifest.yaml", manifest)
		output, _, err := osutils.ExecShell(cmd)
		if err != nil {
			return fmt.Errorf("Failed uninstalling Platform manifests: %v\n", err)
		}
		log.Infof("Terminal output: %v", output)
	}
	return nil

}

func (targetCluster *TargetCluster) InstallPDSAppOnTC(clusterId string, tenantId string) error {
	var pdsAppRequest apiStructs.WorkFlowRequest
	depInputs := apiStructs.WorkFlowRequest{}
	pdsAppRequest.ClusterId = clusterId
	pdsAppRequest.TenantId = tenantId
	if err != nil {
		return fmt.Errorf("Failed to get Context: %v\n", err)
	}
	availableApps, err := v2Components.Platform.ListAvailableApplicationsForTenant(&pdsAppRequest)
	if err != nil {
		return fmt.Errorf("Failed to get list of available Apps: %v\n", err)
	}
	var index int
	for index = 0; index < len(availableApps); index++ {
		if strings.Contains("PDS", *availableApps[index].Meta.Name) {
			pdsApp := availableApps[index].Meta
			appName := pdsApp.Name
			appVersion := pdsApp.ResourceVersion
			var createRequest platformv1.ApiApplicationServiceInstallApplicationRequest
			createRequest = createRequest.ApiService.ApplicationServiceInstallApplication(context.Background(), clusterId)
			createRequest = createRequest.V1Application1(platformv1.V1Application1{
				Meta: &platformv1.V1Meta{
					Name: appName,
				},
				Config: &platformv1.V1Config{
					Version: appVersion,
				},
			})
			err = copier.Copy(&depInputs, createRequest)
			if err != nil {
				return fmt.Errorf("Failed while copying createRequest: %v\n", err)
			}
			_, err := v2Components.Platform.InstallApplication(&depInputs)
			if err != nil {
				return fmt.Errorf("Failed to install App PDS: %v\n", err)
			}
			return nil
		}
	}
	return fmt.Errorf("PDS App not found in catalog")
}

// GetManifest Get the manifest for the account and tenant-id that can be used to install the platform agent
func GetManifest(tenantId string, clusterName string) (string, error) {

	manifestInputs := apiStructs.WorkFlowRequest{}

	// TODO: Proxy and Registry configs need to be added to this call

	if clusterName == "" {
		clusterName = fmt.Sprintf("Cluster_%v", time.Now().Unix())
	}

	manifestInputs.TargetClusterManifest.ClusterName = clusterName
	manifestInputs.TargetClusterManifest.TenantId = tenantId
	log.Infof("cluster name [%s]", manifestInputs.TargetClusterManifest.ClusterName)

	// Get Manifest from API
	manifest, err := v2Components.Platform.GetTargetClusterRegistrationManifest(&manifestInputs)
	if err != nil {
		return "", err
	}
	return manifest, nil
}

func GetClusterIdByName(clusterName string, tenantId string) (string, error) {
	wfRequest := apiStructs.WorkFlowRequest{}
	wfRequest.TenantId = tenantId
	tcList, err := v2Components.Platform.ListTargetClusters(&wfRequest)
	if err != nil {
		return "", err
	}
	var index int
	for index = 0; index < len(tcList); index++ {
		if *tcList[index].Meta.Name == clusterName {
			return *tcList[index].Meta.Uid, nil
		}
	}
	return "", fmt.Errorf("Cluster Name not found in list of targetclusters\n")
}

func getTargetClusterHealth(clusterName string, tenantId string) error {
	wfRequest := apiStructs.WorkFlowRequest{}
	wfRequest.TenantId = tenantId
	tcList, err := v2Components.Platform.ListTargetClusters(&wfRequest)
	if err != nil {
		return err
	}
	var index int
	for index = 0; index < len(tcList); index++ {
		if *tcList[index].Meta.Name == clusterName {
			phase := tcList[index].Status.Phase
			if phase != targetClusterHealthOK {
				return fmt.Errorf("Target Cluster found in %v Phase\n", phase)
			} else {
				return nil
			}
		}
	}
	return fmt.Errorf("Cluster Name not found in list of targetclusters\n")
}
