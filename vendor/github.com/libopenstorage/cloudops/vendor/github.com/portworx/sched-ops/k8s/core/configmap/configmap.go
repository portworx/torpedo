package configmap

import (
	"fmt"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"

	"github.com/portworx/sched-ops/k8s/core"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// New returns the ConfigMap interface. It also creates a new
// configmap in k8s for the given name if not present and puts the data in it.
func New(
	name string,
	data map[string]string,
	lockTimeout time.Duration,
	lockAttempts uint,
	lockRefreshDuration time.Duration,
	lockK8sLockTTL time.Duration,
) (ConfigMap, error) {
	if data == nil {
		data = make(map[string]string)
	}

	labels := map[string]string{
		configMapUserLabelKey: TruncateLabel(name),
	}
	data[pxOwnerKey] = ""

	cm := &v1.ConfigMap{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      name,
			Namespace: k8sSystemNamespace,
			Labels:    labels,
		},
		Data: data,
	}

	if _, err := core.Instance().CreateConfigMap(cm); err != nil &&
		!k8s_errors.IsAlreadyExists(err) {
		return nil, fmt.Errorf("Failed to create configmap %v: %v",
			name, err)
	}
	return &configMap{
		name:                name,
		lockTimeout:         lockTimeout,
		kLocksV2:            map[string]*k8sLock{},
		lockAttempts:        lockAttempts,
		lockRefreshDuration: lockRefreshDuration,
		lockK8sLockTTL:      lockK8sLockTTL,
	}, nil
}

func (c *configMap) Get() (map[string]string, error) {
	cm, err := core.Instance().GetConfigMap(
		c.name,
		k8sSystemNamespace,
	)
	if err != nil {
		return nil, err
	}

	return cm.Data, nil
}

func (c *configMap) Delete() error {
	return core.Instance().DeleteConfigMap(
		c.name,
		k8sSystemNamespace,
	)
}

func (c *configMap) Patch(data map[string]string) error {
	var (
		err error
		cm  *corev1.ConfigMap
	)
	for retries := 0; retries < maxConflictRetries; retries++ {
		cm, err = core.Instance().GetConfigMap(
			c.name,
			k8sSystemNamespace,
		)
		if err != nil {
			return err
		}

		if cm.Data == nil {
			cm.Data = make(map[string]string, 0)
		}

		for k, v := range data {
			cm.Data[k] = v
		}
		_, err = core.Instance().UpdateConfigMap(cm)
		if k8s_errors.IsConflict(err) {
			// try again
			continue
		}
		return err
	}
	return err
}

func (c *configMap) Update(data map[string]string) error {
	var (
		err error
		cm  *corev1.ConfigMap
	)
	for retries := 0; retries < maxConflictRetries; retries++ {
		cm, err = core.Instance().GetConfigMap(
			c.name,
			k8sSystemNamespace,
		)
		if err != nil {
			return err
		}
		cm.Data = data
		_, err = core.Instance().UpdateConfigMap(cm)
		if k8s_errors.IsConflict(err) {
			// try again
			continue
		}
		return err
	}
	return err
}

// SetFatalCb sets the fatal callback for the package which will get invoked in panic situations
func SetFatalCb(fb FatalCb) {
	fatalCb = fb
}

func configMapLog(fn, name, owner, key string, err error) *logrus.Entry {
	if len(owner) > 0 && len(key) > 0 {
		return logrus.WithFields(logrus.Fields{
			"Module":   "ConfigMap",
			"Name":     name,
			"Owner":    owner,
			"Key":      key,
			"Function": fn,
			"Error":    err,
		})
	}
	if len(owner) > 0 {
		return logrus.WithFields(logrus.Fields{
			"Module":   "ConfigMap",
			"Name":     name,
			"Owner":    owner,
			"Function": fn,
			"Error":    err,
		})
	}
	return logrus.WithFields(logrus.Fields{
		"Module":   "ConfigMap",
		"Name":     name,
		"Function": fn,
		"Error":    err,
	})
}

// GetName is a helper function that returns a valid k8s
// configmap name given a prefix identifying the component using
// the configmap and a clusterID
func GetName(prefix, clusterID string) string {
	return prefix + strings.ToLower(configMapNameRegex.ReplaceAllString(clusterID, ""))
}

// TruncateLabel is a helper function that returns a valid k8s
// label stripped down to 63 characters. It removes the trailing characters
func TruncateLabel(label string) string {
	if len(label) > 63 {
		return label[:63]
	}
	return label
}
