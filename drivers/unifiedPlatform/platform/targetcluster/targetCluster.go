package targetcluster

import (
	"fmt"
	"github.com/portworx/sched-ops/k8s/apps"
	"github.com/portworx/sched-ops/k8s/core"
	pdsv2api "github.com/portworx/torpedo/drivers/unifiedPlatform"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/platform/platformUtils"
	"github.com/portworx/torpedo/pkg/log"
	"github.com/portworx/torpedo/pkg/osutils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"strings"
	"time"
)

const (
	k8sObjectCreateTimeout = 2 * time.Minute
	// DefaultRetryInterval default time to retry
	DefaultRetryInterval = 10 * time.Second

	// DefaultTimeout default timeout
	DefaultTimeout        = 10 * time.Minute
	MaxTimeout            = 30 * time.Minute
	timeOut               = 30 * time.Minute
	DefaultPdsPodsTimeOut = 15 * time.Minute
	timeInterval          = 10 * time.Second

	// PDSNamespace PDS
	platformNamespace  = "px-system"
	PDSChartRepo       = "https://pds.pure-px.io/charts/target"
	pxLabel            = "pds.portworx.com/available"
	CertManager        = "jetstack/cert-manager"
	CertManagerVersion = "v1.11.0"
	TLSFeatureGates    = "AdditionalCertificateOutputFormats=true"
	PDSAppUID          = "ABCD123456"
)

var (
	v2Components *pdsv2api.UnifiedPlatformComponents
	k8sCore      = core.Instance()
)

// TargetCluster structure
type TargetCluster struct {
	kubeConfig string
}

// RegisterToControlPlane register the target cluster to control plane.
func (targetCluster *TargetCluster) RegisterToControlPlane(platformVersion string, tenantId string, clusterType string) (string, error) {
	var cmd string
	// Get Manifest from API
	clusterName := fmt.Sprintf("Cluster_%v", time.Now())
	manifest, err := platformUtils.GetManifest(tenantId, clusterName)
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

		isLatest, err := IsLatestManifest(platformVersion)
		if err != nil {
			return "", err
		}
		if !isLatest {
			log.InfoD("Upgrading manifest version to %v", platformVersion)
			//TODO: Logic to upgrade Manifests
		}
		isRegistered = true
	}
	if !isRegistered {
		log.InfoD("Installing Manifests %v", platformVersion)
		cmd = fmt.Sprintf("echo '%v' > /tmp/manifest.yaml && kubectl apply -f /tmp/manifest.yaml && rm -f /tmp/manifest.yaml", manifest)
		log.Infof("Manifest:\n%v\n", cmd)
	}
	output, _, err := osutils.ExecShell(cmd)
	if err != nil {
		return "", fmt.Errorf("Error occured shile installing manifests: %v\n", err)
	}

	log.Infof("Terminal output: %v", output)

	log.InfoD("Verify the health of all the deployments in %s namespace", platformNamespace)
	err = wait.Poll(10*time.Second, 5*time.Minute, func() (bool, error) {
		err := targetCluster.ValidatePlatformComponents()
		if err != nil {
			return false, nil
		}
		return true, nil
	})

	clusterId, err := platformUtils.GetClusterIdByName(clusterName)
	if err != nil {
		return "", fmt.Errorf("Failed to get clusterId: %v", err)
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
		manifest, err := platformUtils.GetManifest(tenantId, "")
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
	availableApps, err := v2Components.Platform.ApplicationV2.ListAvailableApplicationsForTenant(tenantId)
	if err != nil {
		return fmt.Errorf("Failed to get list of available Apps: %v", err)
	}
	var index int
	for index = 0; index < len(availableApps); index++ {
		if strings.Contains("PDS", *availableApps[index].Name) {
			pdsApp := availableApps[index]
			appName := pdsApp.GetName()
			appVersion := pdsApp.GetVersion()
			_, err := v2Components.Platform.ApplicationV2.InstallApplication(appName, appVersion, clusterId)
			if err != nil {
				return fmt.Errorf("Failed to install App PDS: %v", err)
			}
			return nil
		}
	}
	return fmt.Errorf("PDS App not found in catalog")
}

func IsLatestManifest(platformVersion string) (bool, error) {
	//TODO: For now this will always return true. We have to come up with a way to verify if version is latest
	return true, nil
}
