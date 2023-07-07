package tests

import (
	"encoding/json"
	"fmt"
	"os"
)

type CloudProvider struct {
	Provider        string `json:"provider"`
	AccessKeyID     string `json:"access_key_id,omitempty"`
	SecretAccessKey string `json:"secret_access_key,omitempty"`
	Region          string `json:"region,omitempty"`
	Tag             string `json:"tag,omitempty"`
	SubscriptionID  string `json:"subscription_id,omitempty"`
	ClientID        string `json:"client_id,omitempty"`
	ClientSecret    string `json:"client_secret,omitempty"`
	TenantID        string `json:"tenant_id,omitempty"`
	ProjectID       string `json:"project_id,omitempty"`
	ClusterName     string `json:"cluster_name,omitempty"`
	Location        string `json:"location,omitempty"`
	CredentialsFile string `json:"credentials_file,omitempty"`
	APIKey          string `json:"api_key,omitempty"`
	ResourceGroup   string `json:"resource_group,omitempty"`
}

type BackupTarget struct {
	Name            string `json:"name"`
	Provider        string `json:"provider"`
	AccessKeyID     string `json:"access_key_id,omitempty"`
	SecretAccessKey string `json:"secret_access_key,omitempty"`
	Region          string `json:"region,omitempty"`
	Tag             string `json:"tag,omitempty"`
	IP              string `json:"ip,omitempty"`
	ExportPath      string `json:"export_path,omitempty"`
	SubPath         string `json:"sub_path,omitempty"`
	MountOptions    string `json:"mount_options,omitempty"`
	EncryptionKey   string `json:"encryption_key,omitempty"`
}

type Configuration struct {
	CloudProviders []CloudProvider `json:"cloudProviders"`
	BackupTargets  struct {
		Buckets   []BackupTarget `json:"buckets"`
		NFSServer []BackupTarget `json:"nfs-server"`
	} `json:"backupTargets"`
}

//func main() {
//	provider := getCloudProviderCred("azure", "default")
//	fmt.Println("Access Key IDN:", provider.SubscriptionID)
//	backupTarget := getBackupTargets("s3", "default")
//	fmt.Println("Name:", backupTarget.Name)
//	fmt.Println("Region:", backupTarget.Region)
//}

func getCloudProviderCred(Cloudprovider string, tag string) CloudProvider {
	config := getConfigObj()
	// Access the parsed data
	var cp CloudProvider
	for _, provider := range config.CloudProviders {
		fmt.Println("Provider:", provider.Provider)
		fmt.Println("Tag:", provider.Tag)

		if provider.Provider == Cloudprovider && provider.Tag == tag {
			return provider
		}
	}
	return cp
}

func getBackupTargets(backupTarget string, tag string) BackupTarget {
	config := getConfigObj()
	// Access the parsed data
	var cp BackupTarget
	for _, provider := range config.BackupTargets.Buckets {
		fmt.Println("Provider:", provider.Provider)
		fmt.Println("Tag:", provider.Tag)

		if provider.Provider == backupTarget && provider.Tag == tag {
			return provider
		}
	}
	return cp
}

func getConfigObj() Configuration {

	_, err := os.Getwd()
	if err != nil {
		fmt.Println("Error:", err)
	}
	// Read JSON file into a variable
	jsonData, err := os.ReadFile("../tests/backup/cred.json")
	if err != nil {
		fmt.Println("Error reading file:", err)
	}

	// Parse JSON into Configuration struct
	var config Configuration
	err = json.Unmarshal(jsonData, &config)

	return config
}
