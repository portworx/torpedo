package backup

import (
	"fmt"

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

	// DeletrCloudCredential deletes a cloud credential object
	DeleteCloudCredential(orgID string, name string) error

	// CreateCluster creates a cluste object
	CreateCluster(orgID string, name string, cloudCredntial string) error

	// GetCluster enumerates the cluster objects
	GetCluster(OrgID string) error

	// DeleteCluster deletes a cluster object
	DeleteCluster(orgID string, name string) error

	// CreateBackupLocation creates backup location object
	CreateBackupLocation(orgID string, name string, cloudCredential string, path string) error

	// GetBackupLocation enumerates backup location objects
	GetBackupLocation(orgID string, name string) error

	// DeleteBackupLocation deletes backup location objects
	DeleteBackupLocation(orgID string, name string) error

	// CreateBackup creates backup
	CreateBackup(orgID string, name string, backupLocation string, cluster string) error

	// GetBackupList enumerates backup objects
	GetBackupList(orgID string) error

	// InspectBackup inspects a backup object
	InspectBackup(orgID string, name string) error

	// DeleteBackup deletes backup
	DeleteBackup(orgID string, name string) error
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
