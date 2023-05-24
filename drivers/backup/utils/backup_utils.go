package utils

import (
	"fmt"
	snapv1 "github.com/kubernetes-incubator/external-storage/snapshot/pkg/apis/crd/v1"
	apapi "github.com/libopenstorage/autopilot-api/pkg/apis/autopilot/v1alpha1"
	storkapi "github.com/libopenstorage/stork/pkg/apis/stork/v1alpha1"
	"github.com/portworx/sched-ops/k8s/core"
	"github.com/portworx/torpedo/drivers/node/ssh"
	"github.com/portworx/torpedo/drivers/scheduler/spec"
	"github.com/portworx/torpedo/pkg/log"
	"github.com/portworx/torpedo/tests"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	appsapi "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
	storageapi "k8s.io/api/storage/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"os"
	"runtime"
	"strings"
)

const (
	// GlobalTorpedoWorkDirectory is where the Torpedo is located
	GlobalTorpedoWorkDirectory = "/go/src/github.com/portworx/"
	// GlobalKubeconfigDirectory is where the kubeconfig files should be stored
	GlobalKubeconfigDirectory = "/tmp"
)

const (
	// DefaultConfigMapName is the default ConfigMap that stores kubeconfigs data
	DefaultConfigMapName = "kubeconfigs"
	// DefaultConfigMapNamespace is the namespace of the DefaultConfigMapName
	DefaultConfigMapNamespace = "default"
)

const (
	// DefaultSourceClusterName is the default Cluster where the PX-Backup is installed
	DefaultSourceClusterName = "source-cluster"
	// DefaultDestinationClusterName is the default Cluster where applications are deployed
	DefaultDestinationClusterName = "destination-cluster"
)

// ProcessError formats the error message with caller information and an optional debug message
func ProcessError(err error, debugMessage ...string) error {
	if err == nil {
		return nil
	}
	_, file, line, _ := runtime.Caller(1)
	file = strings.TrimPrefix(file, GlobalTorpedoWorkDirectory)
	callerInfo := fmt.Sprintf("%s:%d", file, line)
	debugInfo := "no debug message"
	if len(debugMessage) > 0 {
		debugInfo = "debug message: " + debugMessage[0]
	}
	processedError := fmt.Errorf("%s\n  at %s <-> %s", err.Error(), callerInfo, debugInfo)
	return processedError
}

// GetKubeconfigKeysFromEnv returns kubeconfigKeys from the environment variable KUBECONFIGS
func GetKubeconfigKeysFromEnv() []string {
	envVarKubeconfigs := os.Getenv("KUBECONFIGS")
	kubeconfigKeys := strings.Split(envVarKubeconfigs, ",")
	return kubeconfigKeys
}

// GetClusterConfigPath saves the retrieved kubeconfig from the config-map and returns the file path
func GetClusterConfigPath(kubeconfigKey string, configMapName string, configMapNamespace string) (string, error) {
	filePath := fmt.Sprintf("%s/%s", GlobalKubeconfigDirectory, kubeconfigKey)
	if _, err := os.Stat(filePath); err == nil {
		return filePath, nil
	}
	cm, err := core.Instance().GetConfigMap(configMapName, configMapNamespace)
	if err != nil {
		debugMessage := fmt.Sprintf("config map: name [%s], namespace [%s]", configMapName, configMapNamespace)
		return "", ProcessError(err, debugMessage)
	}
	kubeconfig, ok := cm.Data[kubeconfigKey]
	if !ok {
		err = fmt.Errorf("kubeconfig key [%s] not found in the config map [%s]", kubeconfigKey, configMapName)
		return "", ProcessError(err)
	}
	err = os.WriteFile(filePath, []byte(kubeconfig), 0644)
	if err != nil {
		debugMessage := fmt.Sprintf("kubeconfig file: path [%s]; kubeconfig [%s]", filePath, kubeconfig)
		return "", ProcessError(err, debugMessage)
	}
	return filePath, nil
}

// GetSourceClusterConfigPath returns the file path of the source cluster kubeconfig
func GetSourceClusterConfigPath() (string, error) {
	kubeconfigKeys := GetKubeconfigKeysFromEnv()
	if len(kubeconfigKeys) < 2 {
		err := fmt.Errorf("the environment variable KUBECONFIGS should have two kubeconfig keys: one for the source cluster and one for the destination cluster")
		debugMessage := fmt.Sprintf("kubeconfig-keys: [%s]", kubeconfigKeys)
		return "", ProcessError(err, debugMessage)
	}
	sourceClusterKubeconfigKey := kubeconfigKeys[0]
	sourceClusterConfigPath, err := GetClusterConfigPath(sourceClusterKubeconfigKey, DefaultConfigMapName, DefaultConfigMapNamespace)
	if err != nil {
		debugMessage := fmt.Sprintf("source-cluster: kubeconfig-key [%s]", sourceClusterKubeconfigKey)
		return "", ProcessError(err, debugMessage)
	}
	return sourceClusterConfigPath, nil
}

// GetDestinationClusterConfigPath returns the file path of the destination cluster kubeconfig
func GetDestinationClusterConfigPath() (string, error) {
	kubeconfigKeys := GetKubeconfigKeysFromEnv()
	if len(kubeconfigKeys) < 2 {
		err := fmt.Errorf("the environment variable KUBECONFIGS should have two kubeconfig keys: one for the source cluster and one for the destination cluster")
		debugMessage := fmt.Sprintf("kubeconfig-keys: [%s]", kubeconfigKeys)
		return "", ProcessError(err, debugMessage)
	}
	destinationClusterKubeconfigKey := kubeconfigKeys[1]
	destinationClusterConfigPath, err := GetClusterConfigPath(destinationClusterKubeconfigKey, DefaultConfigMapName, DefaultConfigMapNamespace)
	if err != nil {
		debugMessage := fmt.Sprintf("destination-cluster: kubeconfig-key [%s]", destinationClusterKubeconfigKey)
		return "", ProcessError(err, debugMessage)
	}
	return destinationClusterConfigPath, nil
}

// SwitchClusterContext switches the cluster context to the cluster specified by the clusterConfigPath
//
// SwitchClusterContext replicates the behaviour of the `tests.SetClusterContext` function in the common.go file of the tests package,
// ensuring that errors encountered during the context switching process, including the case when retrieving the SSH node driver fails,
// are appropriately processed using the ProcessError function and returned
func SwitchClusterContext(clusterConfigPath string) error {
	if clusterConfigPath != tests.CurrentClusterConfigPath {
		log.Infof("Switching the cluster context specified by [%s] to [%s]", tests.CurrentClusterConfigPath, clusterConfigPath)
		err := tests.Inst().S.SetConfig(clusterConfigPath)
		if err != nil {
			debugMessage := fmt.Sprintf("cluster: config-path [%s]", clusterConfigPath)
			return ProcessError(err, debugMessage)
		}
		err = tests.Inst().S.RefreshNodeRegistry()
		if err != nil {
			return ProcessError(err)
		}
		err = tests.Inst().V.RefreshDriverEndpoints()
		if err != nil {
			return ProcessError(err)
		}
		if sshNodeDriver, ok := tests.Inst().N.(*ssh.SSH); ok {
			err = ssh.RefreshDriver(sshNodeDriver)
			if err != nil {
				return ProcessError(err)
			}
		} else {
			err = fmt.Errorf("failed to get SSH node driver [%s]", sshNodeDriver.String())
			return ProcessError(err)
		}
	}
	log.Infof("Switched the cluster context specified by [%s] to [%s]", tests.CurrentClusterConfigPath, clusterConfigPath)
	tests.CurrentClusterConfigPath = clusterConfigPath
	return nil
}

// DeepCopyAppSpec returns a deep copy of the AppSpec
//
// DeepCopyAppSpec replicates the behavior of the `(in *spec.AppSpec) DeepCopy` function in the spec.go file of the spec package,
// which had a special case issue. It performed a shallow copy of the SpecList within the AppSpec due to it being a list of pointers,
// requiring individual handling to ensure a proper deep copy
func DeepCopyAppSpec(in *spec.AppSpec) *spec.AppSpec {
	if in == nil {
		return nil
	}
	out := new(spec.AppSpec)
	out.Key = in.Key
	out.Enabled = in.Enabled
	out.SpecList = make([]interface{}, len(in.SpecList))
	for i, spec := range in.SpecList {
		switch v := spec.(type) {
		case *appsapi.Deployment:
			out.SpecList[i] = v.DeepCopy()
		case *appsapi.StatefulSet:
			out.SpecList[i] = v.DeepCopy()
		case *appsapi.DaemonSet:
			out.SpecList[i] = v.DeepCopy()
		case *corev1.Service:
			out.SpecList[i] = v.DeepCopy()
		case *corev1.PersistentVolumeClaim:
			out.SpecList[i] = v.DeepCopy()
		case *storageapi.StorageClass:
			out.SpecList[i] = v.DeepCopy()
		case *snapv1.VolumeSnapshot:
			out.SpecList[i] = v.DeepCopy()
		case *storkapi.GroupVolumeSnapshot:
			out.SpecList[i] = v.DeepCopy()
		case *corev1.Secret:
			out.SpecList[i] = v.DeepCopy()
		case *corev1.ConfigMap:
			out.SpecList[i] = v.DeepCopy()
		case *storkapi.Rule:
			out.SpecList[i] = v.DeepCopy()
		case *corev1.Pod:
			out.SpecList[i] = v.DeepCopy()
		case *storkapi.ClusterPair:
			out.SpecList[i] = v.DeepCopy()
		case *storkapi.Migration:
			out.SpecList[i] = v.DeepCopy()
		case *storkapi.MigrationSchedule:
			out.SpecList[i] = v.DeepCopy()
		case *storkapi.BackupLocation:
			out.SpecList[i] = v.DeepCopy()
		case *storkapi.ApplicationBackup:
			out.SpecList[i] = v.DeepCopy()
		case *storkapi.SchedulePolicy:
			out.SpecList[i] = v.DeepCopy()
		case *storkapi.ApplicationRestore:
			out.SpecList[i] = v.DeepCopy()
		case *storkapi.ApplicationClone:
			out.SpecList[i] = v.DeepCopy()
		case *storkapi.VolumeSnapshotRestore:
			out.SpecList[i] = v.DeepCopy()
		case *apapi.AutopilotRule:
			out.SpecList[i] = v.DeepCopy()
		case *corev1.ServiceAccount:
			out.SpecList[i] = v.DeepCopy()
		case *rbacv1.ClusterRole:
			out.SpecList[i] = v.DeepCopy()
		case *rbacv1.ClusterRoleBinding:
			out.SpecList[i] = v.DeepCopy()
		case *rbacv1.Role:
			out.SpecList[i] = v.DeepCopy()
		case *rbacv1.RoleBinding:
			out.SpecList[i] = v.DeepCopy()
		case *batchv1beta1.CronJob:
			out.SpecList[i] = v.DeepCopy()
		case *batchv1.Job:
			out.SpecList[i] = v.DeepCopy()
		case *corev1.LimitRange:
			out.SpecList[i] = v.DeepCopy()
		case *networkingv1beta1.Ingress:
			out.SpecList[i] = v.DeepCopy()
		case *monitoringv1.Prometheus:
			out.SpecList[i] = v.DeepCopy()
		case *monitoringv1.PrometheusRule:
			out.SpecList[i] = v.DeepCopy()
		case *monitoringv1.ServiceMonitor:
			out.SpecList[i] = v.DeepCopy()
		case *corev1.Namespace:
			out.SpecList[i] = v.DeepCopy()
		case *apiextensionsv1beta1.CustomResourceDefinition:
			out.SpecList[i] = v.DeepCopy()
		case *apiextensionsv1.CustomResourceDefinition:
			out.SpecList[i] = v.DeepCopy()
		case *policyv1beta1.PodDisruptionBudget:
			out.SpecList[i] = v.DeepCopy()
		case *netv1.NetworkPolicy:
			out.SpecList[i] = v.DeepCopy()
		case *corev1.Endpoints:
			out.SpecList[i] = v.DeepCopy()
		case *storkapi.ResourceTransformation:
			out.SpecList[i] = v.DeepCopy()
		case *admissionregistrationv1.ValidatingWebhookConfiguration:
			out.SpecList[i] = v.DeepCopy()
		case *admissionregistrationv1.ValidatingWebhookConfigurationList:
			out.SpecList[i] = v.DeepCopy()
		default:
			out.SpecList[i] = v
		}
	}
	return out
}
