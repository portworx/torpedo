package tests

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type CloudProviders struct {
	AWS   map[string]AWSCredential   `json:"aws"`
	Azure map[string]AzureCredential `json:"azure"`
	GKE   map[string]GKECredential   `json:"gke"`
	IBM   map[string]IBMCredential   `json:"ibm"`
}

type AWSCredential struct {
	AccessKeyID     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key"`
	Region          string `json:"region"`
}

type AzureCredential struct {
	SubscriptionID string `json:"subscription_id"`
	ClientID       string `json:"client_id"`
	ClientSecret   string `json:"client_secret"`
	TenantID       string `json:"tenant_id"`
}

type GKECredential struct {
	ProjectID       string `json:"project_id"`
	ClusterName     string `json:"cluster_name"`
	Location        string `json:"location"`
	CredentialsFile string `json:"credentials_file"`
}

type IBMCredential struct {
	APIKey        string `json:"api_key"`
	Region        string `json:"region"`
	ResourceGroup string `json:"resource_group"`
}

type BackupTargets struct {
	Buckets    Buckets    `json:"buckets"`
	NFSServers NFSServers `json:"nfsServer"`
}

type Buckets struct {
	AWS BucketsAWS `json:"aws"`
}

type BucketsAWS struct {
	Default Bucket `json:"default"`
}

type Bucket struct {
	Name            string `json:"name"`
	Provider        string `json:"provider"`
	AccessKeyID     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key"`
	Region          string `json:"region"`
	Tag             string `json:"tag"`
}

type NFSServers struct {
	Default NFSServer `json:"default"`
}

type NFSServer struct {
	Name          string `json:"name"`
	IP            string `json:"ip"`
	ExportPath    string `json:"export_path"`
	SubPath       string `json:"sub_path"`
	MountOptions  string `json:"mount_options"`
	EncryptionKey string `json:"encryption_key"`
	Tag           string `json:"tag"`
}

type Config struct {
	CloudProviders CloudProviders `json:"cloudProviders"`
	BackupTargets  BackupTargets  `json:"backupTargets"`
}

func getConfigObj() (*Config, error) {

	_, err := os.Getwd()
	// Read JSON file into a variable

	testConfigPath := "../drivers/backup/cloud_config.json"
	jsonData, err := ioutil.ReadFile(testConfigPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read the test configutation file in the path %s", testConfigPath)
	}
	// Parse JSON into Configuration struct
	var config Config
	err = json.Unmarshal(jsonData, &config)

	return &config, nil
}

func (p *CloudProviders) GetAWSCredential(tag string) AWSCredential {
	return p.AWS[tag]
}

func (p *CloudProviders) AzureCredential(tag string) AzureCredential {
	return p.Azure[tag]
}
