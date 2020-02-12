package backup

import (
	"fmt"

	api "github.com/portworx/px-backup-api/pkg/apis/v1"
	"github.com/portworx/torpedo/pkg/errors"
)

// Image Generic struct
type Image struct {
	Type    string
	Version string
}

type Request struct {
	// Organization ID
	OrgID string
	// CloudCredential object name
	CloudCredentialName string
	// BackupLocation object name
	BackupLocationName string
	// Cluster object name
	ClusterName string
	// Backup object name
	BackupName string
	// Labels
	Labels map[string]string
	// Namespaces
	NameSpaces []string
	// Token
	Token string
	// Kubeconfig
	Kubeconfig string
	// BackupLocationPath
	BackupLocationPath string
	// Encryption Key
	EncKey string
	// Provider (AWS, Azure, Google)
	Provider string
	// S3 endpoint
	S3Endpoint string
	// S3 region
	S3Region string
	// DisableSsl
	DisableSsl bool
	// DisablePathStyle
	DisablePathStyle bool
	// Owner
	Owner string
	// AccessKey
	AccessKey string
	// SecretKey
	SecretKey string
}

// Driver for backup
type Driver interface {
	// Init initializes the backup driver under a given scheduler
	Init(schedulerDriverName string, nodeDriverName string, volumeDriverName string, token string) error

	// String returns the name of this driver
	String() string
}

// Org object interface
type Org interface {
	// CreateOrganization creates Organization
	CreateOrganization(req *Request) (*api.OrganizationCreateResponse, error)

	// GetOrganization enumerates organizations
	EnumerateOrganization() (*api.OrganizationEnumerateResponse, error)
}

// CloudCredential object interface
type CloudCredential interface {
	// CreateCloudCredential creates cloud credential objects
	CreateCloudCredential(req *Request) (*api.CloudCredentialCreateResponse, error)

	// InspectCloudCredential describes the cloud credential
	InspectCloudCredential(req *Request) (*api.CloudCredentialInspectResponse, error)

	// EnumerateCloudCredential lists the cloud credentials for given Org
	EnumerateCloudCredential(req *Backup) (*api.CloudCredentialEnumerateResponse, error)

	// DeletrCloudCredential deletes a cloud credential object
	DeleteCloudCredential(req *Request) (*api.CloudCredentialDeleteResponse, error)
}

// Cluster obj interface
type Cluster interface {
	// CreateCluster creates a cluste object
	CreateCluster(req *Request) (*api.ClusterCreateResponse, error)

	// EnumerateCluster enumerates the cluster objects
	EnumerateCluster(req *Request) (*api.ClusterEnumerateResponse, error)

	// InsepctCluster describes a cluster
	InspectCluster(req *Request) (*api.ClusterInspectResponse, error)

	// DeleteCluster deletes a cluster object
	DeleteCluster(req *Request) (*api.ClusterDeleteResponse, error)
}

// BackupLocation obj interface
type BackupLocation interface {
	// CreateBackupLocation creates backup location object
	CreateBackupLocation(req *Request) (*api.BackupLocationCreateResponse, error)

	// EnumerateBackupLocation lists backup locations for an org
	EnumerateBackupLocation(req *Request) (*api.BackupLocationEnumerateResponse, error)

	// InspectBackupLocation enumerates backup location objects
	InspectBackupLocation(req *Request) (*api.BackupLocationInspectResponse, error)

	// DeleteBackupLocation deletes backup location objects
	DeleteBackupLocation(req *Request) (*api.BackupLocationDeleteResponse, error)
}

// Backup obj interface
type Backup interface {
	// CreateBackup creates backup
	CreateBackup(req *Request) (*api.BackupCreateResponse, error)

	// EnumerateBackup enumerates backup objects
	EnumerateBackup(req *Request) (*api.BackupEnumerateResponse, error)

	// InspectBackup inspects a backup object
	InspectBackup(req *Request) (*api.BackupInspectResponse, error)

	// DeleteBackup deletes backup
	DeleteBackup(req *Request) (*api.BackupDeleteResponse, error)
}

var backupDrivers = make(map[string]Driver)

// Register backup driver
func Register(name string, d Driver) error {
	if _, ok := backupDrivers[name]; !ok {
		backupDrivers[name] = d
	} else {
		return fmt.Errorf("backup driver: %s is already registered", name)
	}

	return nil
}

// Get backup driver name
func Get(name string) (Driver, error) {
	d, ok := backupDrivers[name]
	if ok {
		return d, nil
	}

	return nil, &errors.ErrNotFound{
		ID:   name,
		Type: "BackupDriver",
	}
}
