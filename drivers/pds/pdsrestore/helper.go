package pdsrestore

import (
	"fmt"
	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	tc "github.com/portworx/torpedo/drivers/pds/targetcluster"
	"github.com/portworx/torpedo/pkg/log"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	NameSuffixLength  = 3
	lowerAlphaNumeric = "abcdefghijklmnopqrstuvwxyz0123456789"
)

var (
	random = rand.New(rand.NewSource(time.Now().UnixNano()))
	mu     sync.Mutex
)

func GetRestoreTCContext() ([]*tc.TargetCluster, error) {
	values, err := getKubeConfigPathsFromEnv("PDS_RESTORE_TARGET_CLUSTER")
	if err != nil {
		return nil, err
	}
	var tcContexts []*tc.TargetCluster
	for _, val := range values {
		log.Infof("Kube-config Path: %v", val)
		restoreTarget := tc.NewTargetCluster(val)
		tcContexts = append(tcContexts, restoreTarget)
	}
	return tcContexts, nil
}

func resourceStructToMap(resources *pds.ModelsDeploymentResources) map[string]interface{} {
	resourceMap := make(map[string]interface{})
	if resources.CpuLimit != nil {
		resourceMap["cpu_limit"] = resources.GetCpuLimit()
	}
	if resources.CpuRequest != nil {
		resourceMap["cpu_request"] = resources.GetCpuRequest()
	}
	if resources.MemoryLimit != nil {
		resourceMap["memory_limit"] = resources.GetMemoryLimit()
	}
	if resources.MemoryRequest != nil {
		resourceMap["memory_request"] = resources.GetMemoryRequest()
	}
	if resources.StorageRequest != nil {
		resourceMap["storage_request"] = resources.GetStorageRequest()
	}
	return resourceMap
}

func storageOptionsStructToMap(storageOptions *pds.ModelsDeploymentStorageOptions) map[string]interface{} {
	storageMap := make(map[string]interface{})
	if storageOptions.Fg != nil {
		storageMap["fg"] = storageOptions.GetFg()
	}
	if storageOptions.Fs != nil {
		storageMap["fs"] = storageOptions.GetFg()
	}
	if storageOptions.Repl != nil {
		storageMap["repl"] = storageOptions.GetRepl()
	}
	if storageOptions.Secure != nil {
		storageMap["secure"] = storageOptions.GetSecure()
	}
	return storageMap
}

func AlphaNumericString(length int) string {
	mu.Lock()
	defer mu.Unlock()

	result := make([]uint8, length)
	for i := range result {
		result[i] = lowerAlphaNumeric[random.Intn(len(lowerAlphaNumeric))]
	}
	return string(result)
}

func generateRandomName(prefix string) string {
	nameSuffix := AlphaNumericString(NameSuffixLength)
	return fmt.Sprintf("%s-systest-%s", prefix, nameSuffix)
}

func getKubeConfigPathsFromEnv(env string) ([]string, error) {
	filePaths := os.Getenv(env)
	if filePaths == "" {
		return nil, fmt.Errorf("env var {%v} is not set", env)
	}
	if strings.Contains(filePaths, ":") {
		elements := strings.Split(filePaths, ":")
		for _, element := range elements {
			if err := pathExists(element); err != nil {
				return nil, err
			}
		}
		return elements, nil
	}
	if err := pathExists(filePaths); err != nil {
		return nil, err
	}
	return []string{filePaths}, nil

}

func pathExists(filePath string) error {
	_, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file at path %v does not exist", filePath)
		}
	}
	return nil
}
