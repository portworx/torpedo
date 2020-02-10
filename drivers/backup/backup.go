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

// Driver for backup
type Driver interface {
	// Init initializes the backup driver under a given scheduler
	Init(schedulerDriverName string, nodeDriverName string, volumeDriverName string, token string) error

	// String returns the name of this driver
	String() string

	// CreateOrganization creates Organization
	CreateOrganization(orgName string) error

	// GetOrganization enumerates organizations
	GetOrganization() error

	// CreateCloudCredential creates cloud credential objects
	CreateCloudCredential(orgID string, name string) error

	// InspectCloudCredential describes the cloud credential
	InspectCloudCredential(orgID string, name string) (*api.CloudCredentialInspectResponse, error)

	// EnumerateCloudCredential lists the cloud credentials for given Org
	EnumerateCloudCredential(orgID string) (*api.CloudCredentialEnumerateResponse, error)

	// DeletrCloudCredential deletes a cloud credential object
	DeleteCloudCredential(orgID string, name string) error

	// CreateCluster creates a cluste object
	CreateCluster(name string, orgID string, labels map[string]string, cloudCredential string, pxToken string, config string) (*api.ClusterCreateResponse, error)

	// EnumerateCluster enumerates the cluster objects
	EnumerateCluster(OrgID string) (*api.ClusterEnumerateResponse, error)

	// InsepctCluster describes a cluster
	InspectCluster(OrgID string, name string) (*api.ClusterInspectResponse, error)

	// DeleteCluster deletes a cluster object
	DeleteCluster(orgID string, name string) (*api.ClusterDeleteResponse, error)

	// CreateBackupLocation creates backup location object
	CreateBackupLocation(orgID string, name string, cloudCredential string, path string, encKey string,
		provider string, s3Endpoint string, s3Region string, disableSsl bool, disablePathStyle bool) (*api.BackupLocationCreateResponse, error)

	// EnumerateBackupLocation lists backup locations for an org
	EnumerateBackupLocation(orgID string) (*api.BackupLocationEnumerateResponse, error)

	// InspectBackupLocation enumerates backup location objects
	InspectBackupLocation(orgID string, name string) (*api.BackupLocationInspectResponse, error)

	// DeleteBackupLocation deletes backup location objects
	DeleteBackupLocation(orgID string, name string) (*api.BackupLocationDeleteResponse, error)

	// CreateBackup creates backup
	CreateBackup(orgID string, name string, backupLocation string, clusterName string, owner string, nameSpaces []string, labels map[string]string) (*api.BackupCreateResponse, error)

	// EnumerateBackup enumerates backup objects
	EnumerateBackup(orgID string) (*api.BackupEnumerateResponse, error)

	// InspectBackup inspects a backup object
	InspectBackup(orgID string, name string) (*api.BackupInspectResponse, error)

	// DeleteBackup deletes backup
	DeleteBackup(orgID string, name string) (*api.BackupDeleteResponse, error)
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
