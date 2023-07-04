package cluster_utils

import (
	"fmt"
	"github.com/portworx/sched-ops/k8s/core"
	. "github.com/portworx/torpedo/drivers/backup_controller/backup_utils"
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/drivers/scheduler/k8s"
	"github.com/portworx/torpedo/drivers/scheduler/spec"
	"github.com/portworx/torpedo/tests"
	"os"
	"reflect"
	"strings"
	"time"
)

const (
	// GlobalKubeconfigDirectory is where the kubeconfig files should be stored
	GlobalKubeconfigDirectory = "/tmp"
)

const (
	// GlobalInClusterConfigPath is the config-path of the cluster_manager.Cluster in which Torpedo and Px-Backup are running
	GlobalInClusterConfigPath = "" // as described in the doc string of the `SetConfig` function in the k8s.go file of the k8s package
)

const (
	// DefaultConfigMapName is the default config-map that stores kubeconfigs data
	DefaultConfigMapName = "kubeconfigs"
	// DefaultConfigMapNamespace is the namespace of the DefaultConfigMapName
	DefaultConfigMapNamespace = "default"
)

const (
	// DefaultWaitForRunningTimeout indicates the duration to wait for an app to reach the running state
	DefaultWaitForRunningTimeout = 10 * time.Minute
	// DefaultWaitForRunningRetryInterval indicates the interval between retries when waiting for an app to reach the running state
	DefaultWaitForRunningRetryInterval = 10 * time.Second
	// DefaultValidateVolumeTimeout indicates the duration to wait for volume validation of an app
	DefaultValidateVolumeTimeout = 10 * time.Minute
	// DefaultValidateVolumeRetryInterval indicates the interval between retries when performing volume validation of an app
	DefaultValidateVolumeRetryInterval = 10 * time.Second
)

const (
	// DefaultWaitForDestroy indicates whether to wait for resources to be destroyed during the teardown process
	DefaultWaitForDestroy = true // indicates the value of the `scheduler.OptionsWaitForDestroy` key
	// DefaultWaitForResourceLeakCleanup indicates whether to wait for resource leak cleanup during the teardown process
	DefaultWaitForResourceLeakCleanup = true // indicates the value of the `scheduler.OptionsWaitForResourceLeakCleanup` key
	// DefaultSkipClusterScopedObjects indicates whether to skip cluster-scoped objects during the teardown process
	DefaultSkipClusterScopedObjects = false // indicates the value of the `SkipClusterScopedObject` field in the `scheduler.Context`
)

var (
	// DefaultLogsLocation is default location for the logs
	DefaultLogsLocation = tests.Inst().LogLoc
)

// GetKubeconfigKeysFromEnv returns the kubeconfig keys from the env var KUBECONFIGS
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
		debugStruct := struct {
			ConfigMapName      string
			ConfigMapNamespace string
		}{
			ConfigMapName:      configMapName,
			ConfigMapNamespace: configMapNamespace,
		}
		return "", ProcessError(err, StructToString(debugStruct))
	}
	kubeconfig, ok := cm.Data[kubeconfigKey]
	if !ok {
		err = fmt.Errorf("kubeconfig key [%s] not found in the config map [%s]", kubeconfigKey, configMapName)
		return "", ProcessError(err)
	}
	err = os.WriteFile(filePath, []byte(kubeconfig), 0644)
	if err != nil {
		debugStruct := struct {
			FilePath   string
			Kubeconfig string
		}{
			FilePath:   filePath,
			Kubeconfig: kubeconfig,
		}
		return "", ProcessError(err, StructToString(debugStruct))
	}
	return filePath, nil
}

// GetSourceClusterConfigPath returns the file path of the source cluster kubeconfig
func GetSourceClusterConfigPath() (string, error) {
	kubeconfigKeys := GetKubeconfigKeysFromEnv()
	if len(kubeconfigKeys) < 2 {
		err := fmt.Errorf("the environment variable KUBECONFIGS should have two kubeconfig keys: one for the source cluster and one for the destination cluster")
		debugStruct := struct {
			KubeconfigKeys []string
		}{
			KubeconfigKeys: kubeconfigKeys,
		}
		return "", ProcessError(err, StructToString(debugStruct))
	}
	sourceClusterKubeconfigKey := kubeconfigKeys[0]
	sourceClusterConfigPath, err := GetClusterConfigPath(sourceClusterKubeconfigKey, DefaultConfigMapName, DefaultConfigMapNamespace)
	if err != nil {
		debugStruct := struct {
			KubeconfigKey string
		}{
			KubeconfigKey: sourceClusterKubeconfigKey,
		}
		return "", ProcessError(err, StructToString(debugStruct))
	}
	return sourceClusterConfigPath, nil
}

// GetDestinationClusterConfigPath returns the file path of the destination cluster kubeconfig
func GetDestinationClusterConfigPath() (string, error) {
	kubeconfigKeys := GetKubeconfigKeysFromEnv()
	if len(kubeconfigKeys) < 2 {
		err := fmt.Errorf("the environment variable KUBECONFIGS should have two kubeconfig keys: one for the source cluster and one for the destination cluster")
		debugStruct := struct {
			KubeconfigKeys []string
		}{
			KubeconfigKeys: kubeconfigKeys,
		}
		return "", ProcessError(err, StructToString(debugStruct))
	}
	destinationClusterKubeconfigKey := kubeconfigKeys[1]
	destinationClusterConfigPath, err := GetClusterConfigPath(destinationClusterKubeconfigKey, DefaultConfigMapName, DefaultConfigMapNamespace)
	if err != nil {
		debugStruct := struct {
			KubeconfigKey string
		}{
			KubeconfigKey: destinationClusterKubeconfigKey,
		}
		return "", ProcessError(err, StructToString(debugStruct))
	}
	return destinationClusterConfigPath, nil
}

// GetAppSpec returns the *spec.AppSpec associated with the appKey
func GetAppSpec(appKey string) (appSpec *spec.AppSpec, err error) {
	var specFactory *spec.Factory
	switch driver := tests.Inst().S.(type) {
	case *k8s.K8s:
		specFactory = driver.SpecFactory
	default:
		specDir := tests.Inst().SpecDir
		storageProvisioner := tests.Inst().V.String()
		parser := tests.Inst().S
		specFactory, err = spec.NewFactory(specDir, storageProvisioner, parser)
		if err != nil {
			debugStruct := struct {
				SpecDir            string
				StorageProvisioner string
				Parser             scheduler.Driver
			}{
				SpecDir:            specDir,
				StorageProvisioner: storageProvisioner,
				Parser:             parser,
			}
			return nil, ProcessError(err, StructToString(debugStruct))
		}
	}
	appSpec, err = specFactory.Get(appKey)
	if err != nil {
		debugStruct := struct {
			AppKey string
		}{
			AppKey: appKey,
		}
		return nil, ProcessError(err, StructToString(debugStruct))
	}
	return appSpec, nil
}

// DeepCopyAppSpec returns a deep copy of the AppSpec
//
// DeepCopyAppSpec replicates the behavior of the (in *spec.AppSpec) DeepCopy function in the spec.go file of the spec package.
// It performs a shallow copy of the SpecList within the AppSpec due to it being a list of pointers, requiring individual
// handling to ensure a proper deep copy
func DeepCopyAppSpec(in *spec.AppSpec) *spec.AppSpec {
	if in == nil {
		return nil
	}
	out := new(spec.AppSpec)
	*out = *in
	if in.SpecList != nil {
		out.SpecList = make([]interface{}, len(in.SpecList))
		for i, val := range in.SpecList {
			rv := reflect.ValueOf(val)
			if rv.Kind() == reflect.Ptr || rv.Kind() == reflect.Interface {
				method := rv.MethodByName("DeepCopy")
				if !method.IsValid() {
					out.SpecList[i] = val // SoftCopy
				} else {
					result := method.Call(nil)
					if len(result) > 0 {
						out.SpecList[i] = result[0].Interface()
					}
				}
			} else {
				out.SpecList[i] = val // SoftCopy
			}
		}
	}
	return out
}
