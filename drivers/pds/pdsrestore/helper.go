package pdsrestore

import (
	"fmt"
	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	"math/rand"
	"sync"
	"time"
)

const NameSuffixLength = 6

var (
	random = rand.New(rand.NewSource(time.Now().UnixNano()))
	mu     sync.Mutex
)

const lowerAlphaNumeric = "abcdefghijklmnopqrstuvwxyz0123456789"

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
	return fmt.Sprintf("%s-systemtest-%s", prefix, nameSuffix)
}

func compareMaps(map1, map2 map[string]interface{}) bool {
	if len(map1) != len(map2) {
		return false
	}
	for key, value1 := range map1 {
		value2, ok := map2[key]
		if !ok {
			return false
		}
		// Check if the values are equal
		if value1 != value2 {
			return false
		}
	}
	return true
}

func resourceStructToMap(resources *pds.ModelsDeploymentResources) map[string]interface{} {
	resourceMap := make(map[string]interface{})
	if resources.CpuLimit != nil {
		resourceMap["cpu_limit"] = &resources.CpuLimit
	}
	if resources.CpuRequest != nil {
		resourceMap["cpu_request"] = &resources.CpuRequest
	}
	if resources.MemoryLimit != nil {
		resourceMap["memory_limit"] = &resources.MemoryLimit
	}
	if resources.MemoryRequest != nil {
		resourceMap["memory_request"] = &resources.MemoryRequest
	}
	if resources.StorageRequest != nil {
		resourceMap["storage_request"] = &resources.StorageRequest
	}
	return resourceMap
}

func storageOptionsStructToMap(storageOptions *pds.ModelsDeploymentStorageOptions) map[string]interface{} {
	storageMap := make(map[string]interface{})
	if storageOptions.Fg != nil {
		storageMap["fg"] = &storageOptions.Fg
	}
	if storageOptions.Fs != nil {
		storageMap["fs"] = &storageOptions.Fs
	}
	if storageOptions.Provisioner != nil {
		storageMap["provisioner"] = &storageOptions.Provisioner
	}
	if storageOptions.Repl != nil {
		storageMap["repl"] = &storageOptions.Repl
	}
	if storageOptions.Secure != nil {
		storageMap["secure"] = &storageOptions.Secure
	}
	return storageMap
}
