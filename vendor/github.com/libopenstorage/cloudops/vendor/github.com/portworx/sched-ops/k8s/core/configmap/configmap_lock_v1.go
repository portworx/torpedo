package configmap

import (
	"github.com/portworx/sched-ops/k8s/core"
	corev1 "k8s.io/api/core/v1"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	"time"
)

func (c *configMap) Lock(id string) error {
	fn := "Lock"
	count := uint(0)
	// try acquiring a lock on the ConfigMap
	owner, err := c.tryLockV1(id, false)
	// This is the same no. of times (300) we try while acquiring a kvdb lock
	for maxCount := c.lockAttempts; err != nil && count < maxCount; count++ {
		time.Sleep(lockSleepDuration)
		owner, err = c.tryLockV1(id, false)
		if count > 0 && count%15 == 0 && err != nil {
			configMapLog(fn, c.name, owner, "", err).Warnf("Locked for"+
				" %v seconds", float64(count)*lockSleepDuration.Seconds())
		}
	}
	if err != nil {
		// We failed to acquire the lock
		return err
	}
	if count >= 30 {
		configMapLog(fn, c.name, owner, "", err).Warnf("Spent %v iteration"+
			" locking.", count)
	}
	c.kLockV1 = k8sLock{done: make(chan struct{}), id: id}
	go c.refreshLockV1(id)
	return nil
}

func (c *configMap) Unlock() error {
	fn := "Unlock"
	// Get the existing ConfigMap
	c.kLockV1.Lock()
	defer c.kLockV1.Unlock()
	if c.kLockV1.unlocked {
		// The lock is already unlocked
		return nil
	}
	c.kLockV1.unlocked = true
	c.kLockV1.done <- struct{}{}

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
			// A ConfigMap should always be created.
			return err
		}

		currentOwner := cm.Data[pxOwnerKey]
		if currentOwner != c.kLockV1.id {
			// We are currently not holding the lock
			return nil
		}
		delete(cm.Data, pxOwnerKey)
		delete(cm.Data, pxExpirationKey)

		if _, err = core.Instance().UpdateConfigMap(cm); err != nil {
			configMapLog(fn, c.name, "", "", err).Errorf("Failed to update" +
				" config map during unlock")
			if k8s_errors.IsConflict(err) {
				// try unlocking again
				continue
			} // else unknown error - return immediately
			return err
		}
		c.kLockV1.id = ""
		return nil
	}

	return err
}

func (c *configMap) tryLockV1(id string, refresh bool) (string, error) {
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

	currentOwner := cm.Data[pxOwnerKey]
	if currentOwner != "" {
		if currentOwner == id && refresh {
			// We already hold the lock just refresh
			// our expiry
			goto increase_expiry
		} // refresh not requested
		// Someone might have a lock on the cm
		// Check expiration
		expiration := cm.Data[pxExpirationKey]
		if expiration != "" {
			expiresAt, err := time.Parse(time.UnixDate, expiration)
			if err != nil {
				return currentOwner, err
			}
			if time.Now().Before(expiresAt) {
				// Lock is currently held by the owner
				// Retry after sometime
				return currentOwner, ErrConfigMapLocked
			} // else lock is expired. Try to lock it.
		}
	}

	// Take the lock or increase our expiration if we are already holding the lock
	cm.Data[pxOwnerKey] = id
increase_expiry:
	cm.Data[pxExpirationKey] = time.Now().Add(v1DefaultK8sLockTTL).Format(time.UnixDate)

	if _, err = core.Instance().UpdateConfigMap(cm); err != nil {
		return "", err
	}
	return id, nil
}

func (c *configMap) refreshLockV1(id string) {
	fn := "refreshLock"
	refresh := time.NewTicker(v1DefaultK8sLockRefreshDuration)
	var (
		currentRefresh time.Time
		prevRefresh    time.Time
		startTime      time.Time
	)
	startTime = time.Now()
	defer refresh.Stop()
	for {
		select {
		case <-refresh.C:
			c.kLockV1.Lock()
			for !c.kLockV1.unlocked {
				c.checkLockTimeout(startTime, id)
				currentRefresh = time.Now()
				if _, err := c.tryLockV1(id, true); err != nil {
					configMapLog(fn, c.name, "", "", err).Errorf(
						"Error refreshing lock. [Owner %v] [Err: %v]"+
							" [Current Refresh: %v] [Previous Refresh: %v]",
						id, err, currentRefresh, prevRefresh,
					)
					if k8s_errors.IsConflict(err) {
						// try refreshing again
						continue
					}
				}
				prevRefresh = currentRefresh
				break
			}
			c.kLockV1.Unlock()
		case <-c.kLockV1.done:
			return
		}
	}
}
