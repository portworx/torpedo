package utils

import (
	"fmt"
	"github.com/portworx/sched-ops/k8s/core"
	"github.com/portworx/torpedo/drivers/node/ssh"
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/drivers/scheduler/k8s"
	"github.com/portworx/torpedo/drivers/scheduler/spec"
	"github.com/portworx/torpedo/pkg/log"
	"github.com/portworx/torpedo/tests"
	"os"
	"reflect"
	"runtime"
	"strings"
)

type TestMaintainer string

const (
	Sagrawal TestMaintainer = "sagrawal-px"
)

const (
	// GlobalTorpedoWorkDirectory is where the Torpedo is located
	GlobalTorpedoWorkDirectory = "/go/src/github.com/portworx/"
	// GlobalKubeconfigDirectory is where the kubeconfig files should be stored
	GlobalKubeconfigDirectory = "/tmp"
)

const (
	// DefaultConfigMapName is the default config-map that stores kubeconfigs data
	DefaultConfigMapName = "kubeconfigs"
	// DefaultConfigMapNamespace is the namespace of the DefaultConfigMapName
	DefaultConfigMapNamespace = "default"
)

const (
	// GlobalInClusterName is the cluster where PX-Backup and Torpedo are running
	GlobalInClusterName = "in-cluster"
	// GlobalSourceClusterName is the cluster where applications are installed
	GlobalSourceClusterName = "source-cluster"
	// GlobalDestinationClusterName is the cluster where applications will be restored
	GlobalDestinationClusterName = "destination-cluster"
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

// StructToString returns the string representation of the given struct
func StructToString(s interface{}) string {
	v := reflect.ValueOf(s)
	if stringer, ok := s.(fmt.Stringer); ok {
		return stringer.String()
	}
	if v.Kind() != reflect.Struct {
		return fmt.Sprintf("%v", s)
	}
	t := v.Type()
	var fields []string
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.IsExported() {
			fieldVal := v.Field(i)
			var fieldString string
			if stringer, ok := fieldVal.Interface().(fmt.Stringer); ok {
				fieldString = fmt.Sprintf("%s: %s", field.Name, stringer.String())
			} else {
				switch fieldVal.Kind() {
				case reflect.Ptr:
					if fieldVal.IsNil() {
						fieldString = fmt.Sprintf("%s: nil", field.Name)
					} else if fieldVal.Type().Elem().Kind() == reflect.Struct {
						fieldString = fmt.Sprintf("%s: %s", field.Name, StructToString(fieldVal.Elem().Interface()))
					} else {
						fieldString = fmt.Sprintf("%s: %v", field.Name, fieldVal.Elem())
					}
				case reflect.Slice:
					if fieldVal.IsNil() {
						fieldString = fmt.Sprintf("%s: nil", field.Name)
					} else {
						fieldString = fmt.Sprintf("%s: %v", field.Name, fieldVal.Interface())
					}
				case reflect.Map:
					if fieldVal.IsNil() {
						fieldString = fmt.Sprintf("%s: nil", field.Name)
					} else {
						fieldString = fmt.Sprintf("%s: %v", field.Name, fieldVal.Interface())
					}
				case reflect.Struct:
					fieldString = fmt.Sprintf("%s: %s", field.Name, StructToString(fieldVal.Interface()))
				case reflect.String:
					if fieldVal.Len() == 0 {
						fieldString = fmt.Sprintf("%s: \"\"", field.Name)
					} else {
						fieldString = fmt.Sprintf("%s: %v", field.Name, fieldVal.Interface())
					}
				default:
					fieldString = fmt.Sprintf("%s: %v", field.Name, fieldVal.Interface())
				}
			}
			fields = append(fields, fieldString)
		}
	}
	return fmt.Sprintf("%s: {%s}", t.Name(), strings.Join(fields, ", "))
}

// GetKubeconfigKeysFromEnv returns the kubeconfig keys from the env var KUBECONFIGS
func GetKubeconfigKeysFromEnv() []string {
	envVarKubeconfigs := os.Getenv("KUBECONFIGS")
	kubeconfigKeys := strings.Split(envVarKubeconfigs, ",")
	return kubeconfigKeys
}

// GetClusterConfigPath saves the retrieved kubeconfig from the ConfigMap and returns the file path
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

func GetAppSpec(appKey string) (*spec.AppSpec, error) {
	var specFactory *spec.Factory
	var err error
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
	appSpec, err := specFactory.Get(appKey)
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
