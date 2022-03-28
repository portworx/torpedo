package configmap

import (
	"errors"
	"regexp"
	"sync"
	"time"
)

const (
	// DefaultK8sLockAttempts is the number of times to try taking the lock before failing.
	// It defaults to 300, the same number for a kvdb lock.
	DefaultK8sLockAttempts = 300
	// DefaultK8sLockTimeout is time duration within which a lock should be released
	// else it assumes that the node is stuck and panics.
	DefaultK8sLockTimeout = 3 * time.Minute
	// v1DefaultK8sLockTTL is the time duration after which the lock will expire
	v1DefaultK8sLockTTL = 16 * time.Second
	// v1DefaultK8sLockRefreshDuration is the time duration after which a lock is refreshed
	v1DefaultK8sLockRefreshDuration = 8 * time.Second
	// v2DefaultK8sLockTTL is the time duration after which the lock will expire
	v2DefaultK8sLockTTL = 60 * time.Second
	// v2DefaultK8sLockRefreshDuration is the time duration after which a lock is refreshed
	v2DefaultK8sLockRefreshDuration = 20 * time.Second
	// k8sSystemNamespace is the namespace in which we create the ConfigMap
	k8sSystemNamespace = "kube-system"

	// ***********************
	//   ConfigMap Lock Keys
	// ***********************

	// pxOwnerKey is key which indicates the node holding the ConfigMap lock.
	// This is specifically for the deprecated Lock and Unlock methods.
	pxOwnerKey = "px-owner"
	// pxExpirationKey is the key which indicates the time at which the
	// current lock will expire.
	// This is specifically for the deprecated Lock and Unlock methods.
	pxExpirationKey = "px-expiration"

	// pxLockKey is the key which stores the lock data. The data in this key is stored in JSON as an array of lockData
	// objects.
	pxLockKey = "px-lock"

	lockSleepDuration     = 1 * time.Second
	configMapUserLabelKey = "user"
	maxConflictRetries    = 3
)

var (
	// ErrConfigMapLocked is returned when the ConfigMap is locked
	ErrConfigMapLocked = errors.New("ConfigMap is locked")
	fatalCb            FatalCb
	configMapNameRegex = regexp.MustCompile("[^a-zA-Z0-9]+")
)

// FatalCb is a callback function which will be executed if the Lock
// routine encounters a panic situation
type FatalCb func(format string, args ...interface{})

type configMap struct {
	name                string
	kLockV1             k8sLock
	kLocksV2Mutex       sync.Mutex
	kLocksV2            map[string]*k8sLock
	lockTimeout         time.Duration
	lockAttempts        uint
	lockRefreshDuration time.Duration
	lockK8sLockTTL      time.Duration
}

type k8sLock struct {
	done     chan struct{}
	unlocked bool
	id       string
	sync.Mutex
}

// ConfigMap is an interface that provides a set of APIs over a single
// k8s configmap object. The data in the configMap is managed as a map of string
// to string
type ConfigMap interface {
	// Lock locks a configMap where id is the identification of
	// the holder of the lock.
	Lock(id string) error
	// LockWithKey locks a configMap where owner is the identification
	// of the holder of the lock and key is the specific lock to take.
	LockWithKey(owner, key string) error
	// Unlock unlocks the configMap.
	Unlock() error
	// UnlockWithKey unlocks the given key in the configMap.
	UnlockWithKey(key string) error
	// IsKeyLocked returns if the given key is locked, and if so, by which owner.
	IsKeyLocked(key string) (bool, string, error)

	// Patch updates only the keys provided in the input map.
	// It does not replace the complete map
	Patch(data map[string]string) error
	// Update overwrites the data of the configmap
	Update(data map[string]string) error
	// Get returns the contents of the configMap
	Get() (map[string]string, error)
	// Delete deletes the configMap
	Delete() error
}

// lockData structs are serialized into JSON and stored as a list inside a ConfigMap.
// Each lockData struct contains the owner (usually which node took the lock), key
// (which specific lock it's taking), and an expiration time after which the lock is invalid.
type lockData struct {
	Owner      string    `json:"owner"`
	Key        string    `json:"key"`
	Expiration time.Time `json:"expiration"`
}
