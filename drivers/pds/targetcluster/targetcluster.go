package targetcluster

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	apps "github.com/portworx/sched-ops/k8s/apps"
	"github.com/portworx/sched-ops/k8s/core"
	"github.com/portworx/sched-ops/task"
	"github.com/portworx/torpedo/pkg/log"
	"github.com/portworx/torpedo/pkg/osutils"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	k8sNodeReadyTimeout    = 5 * time.Minute
	volDirCleanupTimeout   = 5 * time.Minute
	k8sObjectCreateTimeout = 2 * time.Minute
	k8sDestroyTimeout      = 5 * time.Minute

	// DefaultRetryInterval default time to retry
	DefaultRetryInterval = 10 * time.Second

	// DefaultTimeout default timeout
	DefaultTimeout = 10 * time.Minute

	// PDSNamespace PDS
	PDSNamespace = "pds-system"
	PDSChartRepo = "https://portworx.github.io/pds-charts"
)

var (
	pdsLabel = map[string]string{
		"pds.portworx.com/available": "true",
	}
	k8sCore = core.Instance()
	k8sApps = apps.Instance()
)

// TargetCluster struct
type TargetCluster struct {
	kubeconfig string
}

func (targetCluster *TargetCluster) IsLatestPDSHelm(helmChartversion string) (bool, error) {
	cmd := "helm ls --all -n pds-system | grep pds-target"
	output, _, err := osutils.ExecShell(cmd)
	if err != nil {
		return false, err
	}
	output = strings.ReplaceAll(output, "\t", " ")
	bfr, pdsHelmVersion, found := strings.Cut(output, "pds-target-")
	log.Debugf("Get pds chart: %q, %q, %v\n", bfr, pdsHelmVersion, found)
	helmVersion, after, found := strings.Cut(pdsHelmVersion, " ")
	log.Debugf("Get pds chart version: %q, %q, %v\n", helmVersion, after, found)
	helmVersion = strings.TrimSpace(helmVersion)
	helmChartversion = strings.TrimSpace(helmChartversion)
	log.Debugf("Installed PDS Helm version %s and helm chart version passed %s", helmVersion, helmChartversion)
	return strings.EqualFold(helmVersion, helmChartversion), nil
}

func (targetCluster *TargetCluster) DeRegisterFromControlPlane() error {
	pods, err := k8sCore.GetPods(PDSNamespace, nil)
	if err != nil {
		return err
	}
	if len(pods.Items) > 0 {
		log.InfoD("Uninstalling PDS from the Target cluster")
		cmd := fmt.Sprintf("helm uninstall  pds --namespace %v", PDSNamespace)
		output, _, err := osutils.ExecShell(cmd)
		if err != nil {
			return fmt.Errorf("error occured while removing the pds helm chart: %v", err)
		}
		log.InfoD("helm uninstall output: %v", output)
	} else {
		log.InfoD("No pods are avaialble in the %s namespace", PDSNamespace)
		cmd := "helm list -A"
		output, _, err := osutils.ExecShell(cmd)
		if err != nil {
			return fmt.Errorf("error occured while listing helm installations: %v", err)
		}
		log.InfoD("helm list output: %v", output)
	}

	log.Infof("wait till all the pds-system pods are deleted")
	err = wait.Poll(10*time.Second, 5*time.Minute, func() (bool, error) {
		pods, err := k8sCore.GetPods(PDSNamespace, nil)
		if err != nil {
			return false, nil
		}
		if len(pods.Items) == 0 {
			log.InfoD("pds is uninstalled")
			return true, nil
		}
		log.Infof("There are %d pods present in the namespace %s", len(pods.Items), PDSNamespace)
		return false, nil
	})

	return nil
}

// RegisterToControlPlane register the target cluster to control plane.
func (targetCluster *TargetCluster) RegisterToControlPlane(controlPlaneURL string, helmChartversion string, bearerToken string, tenantId string, clusterType string) error {
	var cmd string
	apiEndpoint := fmt.Sprintf(controlPlaneURL + "api")
	isRegistered := false
	pods, err := k8sCore.GetPods(PDSNamespace, nil)
	if err != nil {
		return err
	}
	if len(pods.Items) > 0 {
		log.InfoD("Target cluster is already registered to control plane.")
		cmd = "helm list -A "
		isLatest, err := targetCluster.IsLatestPDSHelm(helmChartversion)
		if err != nil {
			return err
		}
		if !isLatest {
			log.InfoD("Upgrading PDS helm chart to %v", helmChartversion)
			cmd = fmt.Sprintf("helm upgrade --create-namespace --namespace=%s pds pds-target --repo=%s --version=%s --set tenantId=%s "+
				"--set bearerToken=%s --set apiEndpoint=%s", PDSNamespace, PDSChartRepo, helmChartversion, tenantId, bearerToken, apiEndpoint)
		}
		isRegistered = true
	}
	if !isRegistered {
		log.InfoD("Installing PDS ( helm version -  %v)", helmChartversion)
		cmd = fmt.Sprintf("helm install --create-namespace --namespace=%s pds pds-target --repo=https://portworx.github.io/pds-charts --version=%s --set tenantId=%s "+
			"--set bearerToken=%s --set apiEndpoint=%s", PDSNamespace, helmChartversion, tenantId, bearerToken, apiEndpoint)
		if strings.EqualFold(clusterType, "ocp") {
			cmd = fmt.Sprintf("%s %s ", cmd, "--set platform=ocp")
		}
		log.Infof("helm command: %v ", cmd)
	}
	output, _, err := osutils.ExecShell(cmd)
	if err != nil {
		return fmt.Errorf("kindly remove the PDS chart properly and retry: %v", err)
	}
	log.Infof("Terminal output: %v", output)

	log.InfoD("Verify the health of all the pods in %s namespace", PDSNamespace)
	err = wait.Poll(10*time.Second, 5*time.Minute, func() (bool, error) {
		pods, err := k8sCore.GetPods(PDSNamespace, nil)
		if err != nil {
			return false, nil
		}
		log.Infof("There are %d pods present in the namespace %s", len(pods.Items), PDSNamespace)
		for _, pod := range pods.Items {
			err = k8sCore.ValidatePod(&pod, 10*time.Second, 5*time.Second)
			if err != nil {
				return false, nil
			}
		}
		return true, nil
	})

	return err

}

// isReachbale verify if the control plane is accessable.
func isReachbale(url string) (bool, error) {
	timeout := time.Duration(15 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}
	_, err := client.Get(url)
	if err != nil {
		log.Error(err.Error())
		return false, err
	}
	return true, nil
}

// ValidatePDSComponents used to validate all k8s object in pds-system namespace
func (targetCluster *TargetCluster) ValidatePDSComponents() error {
	pods, err := k8sCore.GetPods(PDSNamespace, nil)
	if err != nil {
		return err
	}

	for _, pod := range pods.Items {
		err = k8sCore.ValidatePod(&pod, DefaultTimeout, DefaultRetryInterval)
		if err != nil {
			return err
		}
	}

	return nil
}

// CreateNamespace create namespace for data service depoloyment
func (targetCluster *TargetCluster) CreateNamespace(namespace string) (*corev1.Namespace, error) {
	t := func() (interface{}, bool, error) {
		nsSpec := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:   namespace,
				Labels: pdsLabel,
			},
		}
		ns, err := k8sCore.CreateNamespace(nsSpec)

		if k8serrors.IsAlreadyExists(err) {
			if ns, err = k8sCore.GetNamespace(namespace); err == nil {
				return ns, false, nil
			}
		}

		return ns, false, nil
	}

	nsObj, err := task.DoRetryWithTimeout(t, k8sObjectCreateTimeout, DefaultRetryInterval)
	if err != nil {
		return nil, err
	}
	return nsObj.(*corev1.Namespace), nil
}

// GetClusterID of target cluster.
func (targetCluster *TargetCluster) GetClusterID() (string, error) {
	log.Infof("Fetch Cluster id ")
	cmd := fmt.Sprintf("kubectl get ns kube-system -o jsonpath={.metadata.uid} --kubeconfig %s", targetCluster.kubeconfig)
	output, _, err := osutils.ExecShell(cmd)
	if err != nil {
		log.Error(err)
		return "Unable to fetch the cluster ID.", err
	}
	return output, nil
}

// SetConfig sets kubeconfig for the target cluster
func (targetCluster *TargetCluster) SetConfig() error {
	var config *rest.Config
	var err error
	config, err = clientcmd.BuildConfigFromFlags("", targetCluster.kubeconfig)
	if err != nil {
		return err
	}
	k8sCore.SetConfig(config)
	k8sApps.SetConfig(config)
	return nil
}

// NewTargetCluster initialize the target cluster
func NewTargetCluster(context string) *TargetCluster {
	return &TargetCluster{
		kubeconfig: context,
	}
}
