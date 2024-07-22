package applicationbackup

import (
	"fmt"
	"time"

	storkv1 "github.com/libopenstorage/stork/pkg/apis/stork/v1alpha1"
	"github.com/portworx/sched-ops/k8s/core"
	storkops "github.com/portworx/sched-ops/k8s/stork"
	"github.com/portworx/sched-ops/task"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	

	"github.com/portworx/torpedo/pkg/log"
)

const (
	configMapName                          = "secret-configmap"
	backupLocationType                     = storkv1.BackupLocationS3
	backupLocationPath                     = "testpath"
	s3SecretName                           = "s3secret"
	applicationBackupScheduleRetryInterval = 10 * time.Second
	applicationBackupScheduleRetryTimeout  = 5 * time.Minute
	applicationRestoreScheduleRetryInterval = 10 * time.Second
	applicationRestoreScheduleRetryTimeout  = 5 * time.Minute
)

func CreateBackupLocation(
	name string,
	namespace string,
	secretName string,
) (*storkv1.BackupLocation, error) {
	log.Infof("Using backup location type as %v", backupLocationType)
	secretObj, err := core.Instance().GetSecret(secretName, "default")
	if err != nil {
		return nil, fmt.Errorf("secret %v is not present in default namespace", secretName)
	}
	_, err = core.Instance().GetSecret(secretName, namespace)
	if err != nil {
		// copy secret to the app namespace
		newSecretObj := secretObj.DeepCopy()
		newSecretObj.Namespace = namespace
		newSecretObj.ResourceVersion = ""
		_, err = core.Instance().CreateSecret(newSecretObj)
		if err != nil {
			return nil, fmt.Errorf("secret %v is not getting created in %v namespace", secretName, namespace)
		}
	}
	backupLocation := &storkv1.BackupLocation{
		ObjectMeta: meta.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Annotations: map[string]string{"stork.libopenstorage.ord/skipresource": "true"},
		},
		Location: storkv1.BackupLocationItem{
			Type:         backupLocationType,
			Path:         backupLocationPath,
			SecretConfig: secretObj.Name,
		},
	}

	backupLocation, err = storkops.Instance().CreateBackupLocation(backupLocation)
	if err != nil {
		return nil, err
	}

	// Doing a "Get" on the backuplocation created to add any missing info from the secrets,
	// that might be required to later get buckets from the cloud objectstore
	backupLocation, err = storkops.Instance().GetBackupLocation(backupLocation.Name, backupLocation.Namespace)
	if err != nil {
		return nil, err
	}
	return backupLocation, nil
}

func CreateApplicationBackup(
	name string,
	namespace string,
	backupLocation *storkv1.BackupLocation,
) (*storkv1.ApplicationBackup, error) {

	appBackup := &storkv1.ApplicationBackup{
		ObjectMeta: meta.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: storkv1.ApplicationBackupSpec{
			Namespaces:     []string{namespace},
			BackupLocation: backupLocation.Name,
		},
	}

	return storkops.Instance().CreateApplicationBackup(appBackup)
}

func CreateApplicationBackupKs(
	name string,
	namespace string,
	backupLocation *storkv1.BackupLocation,
	namespaces []string,
) (*storkv1.ApplicationBackup, error) {

	appBackup := &storkv1.ApplicationBackup{
		ObjectMeta: meta.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: storkv1.ApplicationBackupSpec{
			Namespaces:     namespaces,
			BackupLocation: backupLocation.Name,
		},
	}

	return storkops.Instance().CreateApplicationBackup(appBackup)
}

func CreateApplicationRestore(
	name string,
	namespace string,
	backupLocation *storkv1.BackupLocation,
	backupName string,
	namespaceMapping map[string]string,
) (*storkv1.ApplicationRestore, error) {

	appRestore := &storkv1.ApplicationRestore{
		ObjectMeta: meta.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: storkv1.ApplicationRestoreSpec{
			BackupName:     backupName,
			BackupLocation: backupLocation.Name,
			NamespaceMapping: namespaceMapping,
		},
	}
	return storkops.Instance().CreateApplicationRestore(appRestore)
}

func WaitForAppBackupCompletion(name, namespace string, timeout time.Duration) error {
	getAppBackup := func() (interface{}, bool, error) {
		appBackup, err := storkops.Instance().GetApplicationBackup(name, namespace)
		if err != nil {
			return "", false, err
		}

		if appBackup.Status.Status != storkv1.ApplicationBackupStatusSuccessful {
			return "", true, fmt.Errorf("app backups %s in %s not complete yet.Retrying", name, namespace)
		}
		return "", false, nil
	}
	_, err := task.DoRetryWithTimeout(getAppBackup, timeout, applicationBackupScheduleRetryInterval)
	return err

}

func WaitForAppBackupToStart(name, namespace string, timeout time.Duration) error {
	getAppBackup := func() (interface{}, bool, error) {
		appBackup, err := storkops.Instance().GetApplicationBackup(name, namespace)
		if err != nil {
			return "", false, err
		}

		if appBackup.Status.Status != storkv1.ApplicationBackupStatusInProgress {
			return "", true, fmt.Errorf("app backups %s in %s has not started yet.Retrying Status: %s", name, namespace, appBackup.Status.Status)
		}
		return "", false, nil
	}
	_, err := task.DoRetryWithTimeout(getAppBackup, timeout, applicationBackupScheduleRetryInterval)
	return err
}

func WaitForAppRestoreCompletion(name, namespace string, timeout time.Duration) error {
	getAppRestore := func() (interface{}, bool, error) {
		appRestore, err := storkops.Instance().GetApplicationRestore(name, namespace)
		if err != nil {
			return "", false, err
		}

		if appRestore.Status.Status != storkv1.ApplicationRestoreStatusSuccessful {
			return "", true, fmt.Errorf("app backups %s in %s not complete yet.Retrying", name, namespace)
		}
		return "", false, nil
	}
	_, err := task.DoRetryWithTimeout(getAppRestore, timeout, applicationBackupScheduleRetryInterval)
	return err

}

func WaitForAppRestoreToStart(name, namespace string, timeout time.Duration) error {
	getAppRestore := func() (interface{}, bool, error) {
		appRestore, err := storkops.Instance().GetApplicationRestore(name, namespace)
		if err != nil {
			return "", false, err
		}

		if appRestore.Status.Status != storkv1.ApplicationRestoreStatusInProgress {
			return "", true, fmt.Errorf("app backups %s in %s has not started yet.Retrying Status: %s", name, namespace, appRestore.Status.Status)
		}
		return "", false, nil
	}
	_, err := task.DoRetryWithTimeout(getAppRestore, timeout, applicationRestoreScheduleRetryInterval)
	return err
}