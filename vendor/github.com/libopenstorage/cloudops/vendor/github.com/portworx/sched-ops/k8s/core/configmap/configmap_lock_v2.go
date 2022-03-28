package configmap

import (
	"encoding/json"
	"fmt"
	"time"

	v1 "k8s.io/api/core/v1"

	"github.com/libopenstorage/openstorage/pkg/dbg"
	"github.com/portworx/sched-ops/k8s/core"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
)

func (c *configMap) LockWithKey(owner, key string) error {
	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}

	fn := "LockWithKey"
	count := uint(0)
	// try acquiring a lock on the ConfigMap
	newOwner, err := c.tryLock(owner, key, false)
	// if it fails, keep trying for the provided number of retries until it succeeds
	for maxCount := c.lockAttempts; err != nil && count < maxCount; count++ {
		time.Sleep(lockSleepDuration)
		newOwner, err = c.tryLock(owner, key, false)
		if count > 0 && count%15 == 0 && err != nil {
			configMapLog(fn, c.name, newOwner, key, err).Warnf("Locked for"+
				" %v seconds", float64(count)*lockSleepDuration.Seconds())
		}
	}
	if err != nil {
		// We failed to acquire the lock
		return err
	}
	if count >= 30 {
		configMapLog(fn, c.name, newOwner, key, err).Warnf("Spent %v iteration"+
			" locking.", count)
	}
	c.kLocksV2Mutex.Lock()
	c.kLocksV2[key] = &k8sLock{done: make(chan struct{}), id: owner}
	c.kLocksV2Mutex.Unlock()

	go c.refreshLock(owner, key)
	return nil
}

func (c *configMap) UnlockWithKey(key string) error {
	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}

	fn := "UnlockWithKey"

	// Get the lock reference now so we don't have to keep locking and unlocking
	c.kLocksV2Mutex.Lock()
	lock, ok := c.kLocksV2[key]
	c.kLocksV2Mutex.Unlock()
	if !ok {
		return nil
	}

	lock.Lock()
	defer lock.Unlock()

	if lock.unlocked {
		// The lock is already unlocked
		return nil
	}
	lock.unlocked = true
	lock.done <- struct{}{}

	var (
		err error
		cm  *v1.ConfigMap
	)

	// Get the existing ConfigMap
	for retries := 0; retries < maxConflictRetries; retries++ {
		cm, err = core.Instance().GetConfigMap(
			c.name,
			k8sSystemNamespace,
		)
		if err != nil {
			// A ConfigMap should always be created.
			return err
		}

		lockOwners, lockExpirations, err := c.parseLocks(cm)
		if err != nil {
			return fmt.Errorf("failed to get locks from configmap: %v", err)
		}

		currentOwner := lockOwners[key]
		if currentOwner != lock.id {
			return nil
		}

		// We are holding the lock, let's remove it
		delete(lockOwners, key)
		delete(lockExpirations, key)

		err = c.generateConfigMapData(cm, lockOwners, lockExpirations)
		if err != nil {
			return err
		}

		if k8sConflict, err := c.updateConfigMap(cm); err != nil {
			configMapLog(fn, c.name, "", "", err).Errorf("Failed to update" +
				" config map during unlock")
			if k8sConflict {
				// try unlocking again
				continue
			}
			// else unknown error - return immediately
			return err
		}

		// Clean up the lock
		c.kLocksV2Mutex.Lock()
		delete(c.kLocksV2, key)
		c.kLocksV2Mutex.Unlock()
		return nil
	}

	return err
}

func (c *configMap) IsKeyLocked(key string) (bool, string, error) {
	// Get the existing ConfigMap
	cm, err := core.Instance().GetConfigMap(
		c.name,
		k8sSystemNamespace,
	)
	if err != nil {
		return false, "", err
	}

	if cm.Data == nil {
		cm.Data = make(map[string]string)
	}

	lockIDs, lockExpirations, err := c.parseLocks(cm)
	if err != nil {
		return false, "", fmt.Errorf("failed to get locks from configmap: %v", err)
	}

	if owner, ok := lockIDs[key]; ok {
		// Someone owns the lock: let's check the expiration and the owner
		expiration := lockExpirations[key]
		if time.Now().After(expiration) {
			// Lock is expired
			return false, "", nil
		}

		return true, owner, nil
	}

	// Nobody owns the lock
	return false, "", nil
}

func (c *configMap) tryLock(owner string, key string, refresh bool) (string, error) {
	// Get the existing ConfigMap
	cm, err := core.Instance().GetConfigMap(
		c.name,
		k8sSystemNamespace,
	)
	if err != nil {
		// A ConfigMap should always be created.
		return "", err
	}

	if cm.Data == nil {
		cm.Data = make(map[string]string)
	}

	lockIDs, lockExpirations, err := c.parseLocks(cm)
	if err != nil {
		return "", fmt.Errorf("failed to get locks from configmap: %v", err)
	}

	finalOwner, err := c.checkAndTakeLock(owner, key, refresh, lockIDs, lockExpirations)
	if err != nil {
		return finalOwner, err
	}

	err = c.generateConfigMapData(cm, lockIDs, lockExpirations)
	if err != nil {
		return finalOwner, err
	}

	if _, err := c.updateConfigMap(cm); err != nil {
		return "", err
	}
	return owner, nil
}

// parseLocks reads the lock data from the given ConfigMap and then converts it to:
// * a map of keys to lock owners
// * a map of keys to lock expiration times
func (c *configMap) parseLocks(cm *v1.ConfigMap) (map[string]string, map[string]time.Time, error) {
	// Check all the locks: will be an empty string if key is not present indicating no lock
	parsedLocks := []lockData{}
	if lock, ok := cm.Data[pxLockKey]; ok && len(lock) > 0 {
		err := json.Unmarshal([]byte(cm.Data[pxLockKey]), &parsedLocks)
		if err != nil {
			return nil, nil, err
		}
	}

	// Check all the locks first and store them, makes the looping a little easier
	lockOwners := map[string]string{}
	lockExpirations := map[string]time.Time{}

	for _, lock := range parsedLocks {
		lockOwners[lock.Key] = lock.Owner
		lockExpirations[lock.Key] = lock.Expiration
	}

	return lockOwners, lockExpirations, nil
}

// checkAndTakeLock tries to take the given lock (owner, key) given the current state of the lock
// (lockOwners, lockExpirations). Refresh indicates if this is the refreshLock goroutine
// refreshing the lock or an initial Lock call taking the lock.
func (c *configMap) checkAndTakeLock(owner, key string, refresh bool,
	lockOwners map[string]string, lockExpirations map[string]time.Time) (string, error) {
	fn := "checkAndTakeLock"

	_, ownerOK := lockOwners[key]
	_, expOK := lockExpirations[key]

	// Just check to make sure these are consistent and that we either have both or don't
	if ownerOK != expOK {
		return "", fmt.Errorf("inconsistent lock ID and expiration")
	}
	k8sTTL := v2DefaultK8sLockTTL
	if c.lockK8sLockTTL > 0 {
		k8sTTL = c.lockK8sLockTTL
	}

	// Now that we've parsed all the lock lines, let's check the specific key we're taking
	if !ownerOK {
		// There is no lock, we can take it
		lockOwners[key] = owner
		lockExpirations[key] = time.Now().Add(k8sTTL)
		return owner, nil
	}

	// There is a lock: let's check it out and see what the details are

	// First let's see if we're refreshing and who holds it
	if refresh && lockOwners[key] == owner {
		// We hold the lock, just refresh it
		lockExpirations[key] = time.Now().Add(k8sTTL)
		return owner, nil
	}

	// Now let's see if the lock is expired or not: if it's not expired, we can't take it
	if time.Now().Before(lockExpirations[key]) {
		return lockOwners[key], ErrConfigMapLocked
	}

	configMapLog(fn, c.name, owner, key, nil).Infof("Lock from owner '%s' is expired, now claiming for new owner '%s'", lockOwners[key], owner)

	// Lock is expired: let's take it
	lockOwners[key] = owner
	lockExpirations[key] = time.Now().Add(k8sTTL)
	return owner, nil
}

// generateConfigMapData converts the given lock data (lockOwners, lockExpirations) to JSON and
// stores it in the given ConfigMap.
func (c *configMap) generateConfigMapData(cm *v1.ConfigMap, lockOwners map[string]string, lockExpirations map[string]time.Time) error {
	var locks []lockData
	for key, lockOwner := range lockOwners {
		locks = append(locks, lockData{
			Owner:      lockOwner,
			Key:        key,
			Expiration: lockExpirations[key],
		})
	}

	cmData, err := json.Marshal(locks)
	if err != nil {
		return err
	}
	cm.Data[pxLockKey] = string(cmData)
	return nil
}

func (c *configMap) updateConfigMap(cm *v1.ConfigMap) (bool, error) {
	if _, err := core.Instance().UpdateConfigMap(cm); err != nil {
		return k8s_errors.IsConflict(err), err
	}
	return false, nil
}

// refreshLock is the goroutine running in the background after calling LockWithKey.
// It keeps the lock refreshed in k8s until we call Unlock. This is so that if the
// node dies, the lock can have a short timeout and expire quickly but we can still
// take longer-term locks.
func (c *configMap) refreshLock(id, key string) {
	fn := "refreshLock"
	refresh := time.NewTicker(v2DefaultK8sLockRefreshDuration)
	if c.lockRefreshDuration > 0 {
		refresh = time.NewTicker(c.lockRefreshDuration)
	}
	var (
		currentRefresh time.Time
		prevRefresh    time.Time
		startTime      time.Time
	)

	// get a reference to the lock object so we don't have to hold open a
	// map reference - this makes it easier for concurrency purposes (can't
	// lock in a select condition)
	c.kLocksV2Mutex.Lock()
	lock := c.kLocksV2[key]
	c.kLocksV2Mutex.Unlock()

	startTime = time.Now()
	defer refresh.Stop()
	for {
		select {
		case <-refresh.C:
			lock.Lock()

			for !lock.unlocked {
				c.checkLockTimeout(startTime, id)
				currentRefresh = time.Now()
				if _, err := c.tryLock(id, key, true); err != nil {
					configMapLog(fn, c.name, "", key, err).Errorf(
						"Error refreshing lock. [ID %v] [Key %v] [Err: %v]"+
							" [Current Refresh: %v] [Previous Refresh: %v]",
						id, key, err, currentRefresh, prevRefresh,
					)
					if k8s_errors.IsConflict(err) {
						// try refreshing again
						continue
					}
				}
				prevRefresh = currentRefresh
				break
			}
			lock.Unlock()
		case <-lock.done:
			return
		}
	}

}

func (c *configMap) checkLockTimeout(startTime time.Time, id string) {
	if c.lockTimeout > 0 && time.Since(startTime) > c.lockTimeout {
		panicMsg := fmt.Sprintf("Lock timeout triggered for K8s configmap lock key %s", id)
		if fatalCb != nil {
			fatalCb(panicMsg)
		} else {
			dbg.Panicf(panicMsg)
		}
	}
}
