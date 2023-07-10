package tests

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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

func getCloudProviderCred(Cloudprovider string, tag string) (*CloudProvider, error) {
	config, _ := getConfigObj()
	// Access the parsed data
	for _, provider := range config.CloudProviders {
		fmt.Println("Provider:", provider.Provider)
		fmt.Println("Tag:", provider.Tag)

		if provider.Provider == Cloudprovider && provider.Tag == tag {
			return &provider, nil
		}
	}
	return nil, fmt.Errorf("unable to find the cloud provider %s with tag %s", Cloudprovider, tag)
}

func getBackupTargets(backupTarget string, tag string) (*BackupTarget, error) {
	config, _ := getConfigObj()
	// Access the parsed data
	for _, provider := range config.BackupTargets.Buckets {
		fmt.Println("Provider:", provider.Provider)
		fmt.Println("Tag:", provider.Tag)

		if provider.Provider == backupTarget && provider.Tag == tag {
			return &provider, nil
		}
	}
	return nil, fmt.Errorf("unable to find the cloud provider %s with tag %s", backupTarget, tag)
}

func getConfigObj() (*Configuration, error) {

	_, err := os.Getwd()
	// Read JSON file into a variable

	testConfigPath := "../drivers/backup/test_config.json"
	jsonData, err := ioutil.ReadFile(testConfigPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read the test configutation file in the path %s", testConfigPath)
	}
	// Parse JSON into Configuration struct
	var config Configuration
	err = json.Unmarshal(jsonData, &config)

	return &config, nil
}
