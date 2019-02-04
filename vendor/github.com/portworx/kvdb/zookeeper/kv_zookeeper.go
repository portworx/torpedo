// Package zookeeper implements the KVDB interface based for zookeeper.
// Code from docker/libkv was leveraged to build parts of this module.

package zookeeper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/portworx/kvdb"
	"github.com/portworx/kvdb/common"
	"github.com/samuel/go-zookeeper/zk"
	"github.com/sirupsen/logrus"
)

const (
	// Name is the name of this kvdb implementation
	Name = "zookeeper-kv"
	// SOH control character
	SOH = "\x01"
	// defaultSessionTimeout is the client session timeout. If the session times out,
	// then all ephemeral zk nodes are deleted by the zk server. As all lock keys are
	// ephemeral nodes, all locks will be released after this timeout when a node
	// dies or closes the zookeeper connection.
	defaultSessionTimeout = 16 * time.Second
	defaultRetryCount     = 10
)

var (
	defaultServers = []string{"127.0.0.1:2181"}
)

type zookeeperKV struct {
	common.BaseKvdb
	client  *zk.Conn
	options map[string]string
	domain  string
	kvdb.Controller
}

// zookeeperLock combines Mutex and channel
type zookeeperLock struct {
	Done     chan struct{}
	Unlocked bool
	sync.Mutex
}

// LockerIDInfo id of locker
type LockerIDInfo struct {
	LockerID string
}

// New constructs a new kvdb.Kvdb with zookeeper implementation
func New(
	domain string,
	servers []string,
	options map[string]string,
	fatalErrorCb kvdb.FatalErrorCB,
) (kvdb.Kvdb, error) {
	return newClient(domain, servers, options, fatalErrorCb)
}

// Version returns the supported version of the zookeeper implementation
func Version(url string, kvdbOptions map[string]string) (string, error) {
	return kvdb.ZookeeperVersion1, nil
}

// Used to create a zookeeper client for testing
func newClient(
	domain string,
	servers []string,
	options map[string]string,
	fatalErrorCb kvdb.FatalErrorCB,
) (*zookeeperKV, error) {
	if len(servers) == 0 {
		servers = defaultServers
	}

	if domain != "" {
		domain = normalize(domain)
	}

	zkClient, _, err := zk.Connect(servers, defaultSessionTimeout)
	if err != nil {
		return nil, err
	}

	return &zookeeperKV{
		BaseKvdb: common.BaseKvdb{
			FatalCb: fatalErrorCb,
		},
		client:     zkClient,
		domain:     domain,
		Controller: kvdb.ControllerNotSupported,
	}, nil
}

func (z *zookeeperKV) closeClient() {
	z.client.Close()
}

func (z *zookeeperKV) String() string {
	return Name
}

func (z *zookeeperKV) Capabilities() int {
	return 0
}

func (z *zookeeperKV) Get(key string) (*kvdb.KVPair, error) {
	var (
		err  error
		resp []byte
		meta *zk.Stat
	)

	key = z.domain + normalize(key)

	for i := 0; i < z.getRetryCount(); i++ {
		resp, meta, err = z.client.Get(key)
		if err == nil && resp != nil {
			// Rare case where Get returns SOH control character
			// instead of the actual value
			if string(resp) == SOH {
				continue
			}
			return z.resultToKvPair(key, resp, "get", meta), nil
		}
		if err == zk.ErrNoNode {
			return nil, kvdb.ErrNotFound
		}
		return nil, err
	}

	return nil, err
}

func (z *zookeeperKV) GetVal(key string, val interface{}) (*kvdb.KVPair, error) {
	kvp, err := z.Get(key)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(kvp.Value, val); err != nil {
		return kvp, kvdb.ErrUnmarshal
	}
	return kvp, nil
}

func (z *zookeeperKV) GetWithCopy(
	key string,
	copySelect kvdb.CopySelect,
) (interface{}, error) {
	return nil, kvdb.ErrNotSupported
}

// Put creates a node if it does not exist and sets the given value
func (z *zookeeperKV) Put(
	key string,
	val interface{},
	ttl uint64,
) (*kvdb.KVPair, error) {
	if ttl != 0 {
		return nil, kvdb.ErrTTLNotSupported
	}

	bval, err := common.ToBytes(val)
	if err != nil {
		return nil, err
	}

	err = z.createFullPath(key, false)
	if err != nil && err != zk.ErrNodeExists {
		return nil, err
	}

	key = z.domain + normalize(key)
	meta, err := z.client.Set(key, bval, -1)
	if err != nil {
		return nil, err
	}
	return z.resultToKvPair(key, bval, "set", meta), nil
}

// Creates a zk node only if it does not exist.
func (z *zookeeperKV) Create(
	key string,
	val interface{},
	ttl uint64,
) (*kvdb.KVPair, error) {
	if ttl != 0 {
		return nil, kvdb.ErrTTLNotSupported
	}

	bval, err := common.ToBytes(val)
	if err != nil {
		return nil, err
	}

	err = z.createFullPath(key, false)
	if err == zk.ErrNodeExists {
		return nil, kvdb.ErrExist
	} else if err != nil {
		return nil, err
	}

	key = z.domain + normalize(key)
	meta, err := z.client.Set(key, bval, -1)
	if err != nil {
		return nil, err
	}
	return z.resultToKvPair(key, bval, "create", meta), nil
}

func (z *zookeeperKV) createEphemeral(
	key string,
	val interface{},
) (*kvdb.KVPair, error) {
	bval, err := common.ToBytes(val)
	if err != nil {
		return nil, err
	}

	err = z.createFullPath(key, true)
	if err == zk.ErrNodeExists {
		return nil, kvdb.ErrExist
	} else if err != nil {
		return nil, err
	}

	key = z.domain + normalize(key)
	meta, err := z.client.Set(key, bval, -1)
	if err != nil {
		return nil, err
	}
	return z.resultToKvPair(key, bval, "create", meta), nil
}

func (z *zookeeperKV) Update(
	key string,
	val interface{},
	ttl uint64,
) (*kvdb.KVPair, error) {
	exists, err := z.exists(key)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, kvdb.ErrNotFound
	}
	return z.Put(key, val, ttl)
}

func (z *zookeeperKV) Enumerate(prefix string) (kvdb.KVPairs, error) {
	prefix = normalize(prefix)
	fullPrefix := z.domain + prefix
	keys, _, err := z.client.Children(fullPrefix)
	if err != nil {
		return nil, err
	}

	kvs := []*kvdb.KVPair{}
	for _, key := range keys {
		kvp, err := z.Get(prefix + normalize(key))
		if err != nil {
			return nil, err
		}
		kvs = append(kvs, kvp)
	}
	return kvs, nil
}

func (z *zookeeperKV) EnumerateWithSelect(
	prefix string,
	enumerateSelect kvdb.EnumerateSelect,
	copySelect kvdb.CopySelect,
) ([]interface{}, error) {
	return nil, kvdb.ErrNotSupported
}

func (z *zookeeperKV) Delete(key string) (*kvdb.KVPair, error) {
	kvp, err := z.Get(key)
	if err != nil {
		return nil, err
	}

	key = z.domain + normalize(key)
	err = z.client.Delete(key, -1)
	if err != nil {
		return nil, err
	}

	kvp.Action = kvdb.KVDelete
	return kvp, nil
}

func (z *zookeeperKV) DeleteTree(prefix string) error {
	fullPrefix := z.domain + normalize(prefix)
	keys, _, err := z.client.Children(fullPrefix)
	if err != nil {
		return err
	}

	var requests []interface{}
	for _, key := range keys {
		requests = append(requests, &zk.DeleteRequest{
			Path:    fullPrefix + normalize(key),
			Version: -1,
		})
	}

	_, err = z.client.Multi(requests...)
	return err
}

func (z *zookeeperKV) Keys(prefix, sep string) ([]string, error) {
	prefix = normalize(prefix)
	fullPrefix := z.domain + prefix
	keys, _, err := z.client.Children(fullPrefix)
	if err != nil {
		return nil, err
	}
	return keys, nil
}

func (z *zookeeperKV) CompareAndSet(
	kvp *kvdb.KVPair,
	flags kvdb.KVFlags,
	prevValue []byte,
) (*kvdb.KVPair, error) {
	action := "create"
	modifiedIndex := int32(-1)

	prevPair, err := z.Get(kvp.Key)
	if err == nil {
		if (flags & kvdb.KVModifiedIndex) != 0 {
			if kvp.ModifiedIndex != prevPair.ModifiedIndex {
				return nil, kvdb.ErrModified
			}
		} else {
			if bytes.Compare(prevValue, prevPair.Value) != 0 {
				return nil, kvdb.ErrValueMismatch
			}
		}
		action = "set"
		modifiedIndex = int32(prevPair.ModifiedIndex)
	} else if err != kvdb.ErrNotFound {
		return nil, err
	}

	err = z.createFullPath(kvp.Key, false)
	if err != nil && err != zk.ErrNodeExists {
		return nil, err
	}

	key := z.domain + normalize(kvp.Key)
	// In case of two nodes trying to call CAS at the same time, setting
	// modified index in Set() will ensure that only one update succeeds.
	// Zookeeper will reject an update if the ModifiedIndex does not
	// match the previous version, unless it is -1.
	meta, err := z.client.Set(key, kvp.Value, modifiedIndex)
	if err == zk.ErrBadVersion {
		return nil, kvdb.ErrModified
	} else if err != nil {
		return nil, err
	}
	return z.resultToKvPair(kvp.Key, kvp.Value, action, meta), nil
}

func (z *zookeeperKV) CompareAndDelete(
	kvp *kvdb.KVPair,
	flags kvdb.KVFlags,
) (*kvdb.KVPair, error) {
	prevPair, err := z.Get(kvp.Key)
	if err != nil {
		return nil, err
	}

	if (flags & kvdb.KVModifiedIndex) != 0 {
		if kvp.ModifiedIndex != prevPair.ModifiedIndex {
			return nil, kvdb.ErrModified
		}
	} else {
		if bytes.Compare(kvp.Value, prevPair.Value) != 0 {
			return nil, kvdb.ErrValueMismatch
		}
	}

	key := z.domain + normalize(kvp.Key)
	// In case of two nodes trying to call CompareAndDelete at the same time,
	// setting modified index in Delete() will ensure that only one call succeeds.
	// Zookeeper will reject an update if the ModifiedIndex does not match the
	// previous version, unless it is -1.
	err = z.client.Delete(key, int32(prevPair.ModifiedIndex))
	if err == zk.ErrNoNode {
		return nil, kvdb.ErrNotFound
	} else if err == zk.ErrBadVersion {
		return nil, kvdb.ErrModified
	} else if err != nil {
		return nil, err
	}
	prevPair.Action = kvdb.KVDelete
	return prevPair, nil
}

func (z *zookeeperKV) WatchKey(
	key string,
	waitIndex uint64,
	opaque interface{},
	watchCB kvdb.WatchCB,
) error {
	return kvdb.ErrNotSupported
}

func (z *zookeeperKV) WatchTree(
	prefix string,
	waitIndex uint64,
	opaque interface{},
	watchCB kvdb.WatchCB,
) error {
	return kvdb.ErrNotSupported
}

func (z *zookeeperKV) Snapshot(prefixes []string) (kvdb.Kvdb, uint64, error) {
	return nil, 0, kvdb.ErrNotSupported
}

func (z *zookeeperKV) SnapPut(kvp *kvdb.KVPair) (*kvdb.KVPair, error) {
	return nil, kvdb.ErrNotSupported
}

func (z *zookeeperKV) Lock(key string) (*kvdb.KVPair, error) {
	return z.LockWithID(key, "locked")
}

func (z *zookeeperKV) LockWithID(key, lockerID string) (*kvdb.KVPair, error) {
	return z.LockWithTimeout(key, lockerID, kvdb.DefaultLockTryDuration, z.GetLockTimeout())
}

func (z *zookeeperKV) LockWithTimeout(
	key string,
	lockerID string,
	lockTryDuration time.Duration,
	lockHoldDuration time.Duration,
) (*kvdb.KVPair, error) {
	key = normalize(key)
	lockTag := LockerIDInfo{LockerID: lockerID}

	kvPair, err := z.createEphemeral(key, lockTag)
	startTime := time.Now()

	for count := 0; err != nil; count++ {
		time.Sleep(time.Second)
		kvPair, err = z.createEphemeral(key, lockTag)
		if count > 0 && count%15 == 0 && err != nil {
			currLockerTag := LockerIDInfo{}
			if _, errGet := z.GetVal(key, &currLockerTag); errGet == nil {
				logrus.Warnf("Lock %v locked for %v seconds, tag: %v, err: %v",
					key, count, currLockerTag, err)
			}
		}
		if err != nil && time.Since(startTime) > lockTryDuration {
			return nil, err
		}
	}
	if err != nil {
		return nil, err
	}

	kvPair.Lock = &zookeeperLock{Done: make(chan struct{})}
	go z.waitForUnlock(kvPair, lockerID, lockHoldDuration)
	return kvPair, err
}

func (z *zookeeperKV) waitForUnlock(
	kvp *kvdb.KVPair,
	lockerID string,
	lockHoldDuration time.Duration,
) {
	l := kvp.Lock.(*zookeeperLock)
	timer := time.NewTimer(lockHoldDuration)
	defer timer.Stop()

	lockMsgString := kvp.Key + ",tag=" + lockerID

	for {
		select {
		case <-timer.C:
			if lockHoldDuration > 0 {
				l.Lock()
				defer l.Unlock()
				if !l.Unlocked {
					if _, err := z.Delete(kvp.Key); err != nil {
						logrus.Errorf("Error deleting lock %s after timeout. %v",
							lockMsgString, err)
					}
					l.Unlocked = true
				}
				z.FatalCb("Lock %s hold timeout triggered", lockMsgString)
				return
			}
		case <-l.Done:
			return
		}
	}
}

func (z *zookeeperKV) Unlock(kvp *kvdb.KVPair) error {
	l, ok := kvp.Lock.(*zookeeperLock)
	if !ok {
		return fmt.Errorf("Invalid lock structure for key: %s", kvp.Key)
	}
	l.Lock()
	if _, err := z.Delete(kvp.Key); err != nil {
		l.Unlock()
		return err
	}
	l.Unlocked = true
	l.Unlock()
	l.Done <- struct{}{}
	return nil
}

func (z *zookeeperKV) TxNew() (kvdb.Tx, error) {
	return nil, kvdb.ErrNotSupported
}

func (z *zookeeperKV) AddUser(username string, password string) error {
	return kvdb.ErrNotSupported
}

func (z *zookeeperKV) RemoveUser(username string) error {
	return kvdb.ErrNotSupported
}

func (z *zookeeperKV) GrantUserAccess(
	username string,
	permType kvdb.PermissionType,
	subtree string,
) error {
	return kvdb.ErrNotSupported
}

func (z *zookeeperKV) RevokeUsersAccess(
	username string,
	permType kvdb.PermissionType,
	subtree string,
) error {
	return kvdb.ErrNotSupported
}

func (z *zookeeperKV) Serialize() ([]byte, error) {
	return nil, kvdb.ErrNotSupported
}

func (z *zookeeperKV) Deserialize(b []byte) (kvdb.KVPairs, error) {
	return nil, kvdb.ErrNotSupported
}

// normalize converts a given path to the form /a/b/c
func normalize(key string) string {
	if key == "" {
		return ""
	}
	path := strings.Split(strings.Trim(key, "/"), "/")
	return "/" + strings.Join(path, "/")
}

// createFullPath creates the entire path for a directory
func (z *zookeeperKV) createFullPath(key string, ephemeral bool) error {
	key = z.domain + normalize(key)
	path := strings.Split(strings.TrimPrefix(key, "/"), "/")

	for i := 1; i <= len(path); i++ {
		newPath := "/" + strings.Join(path[:i], "/")

		if i == len(path) {
			flags := int32(0)
			if ephemeral {
				flags = zk.FlagEphemeral
			}
			_, err := z.client.Create(newPath, []byte{},
				flags, zk.WorldACL(zk.PermAll))
			return err
		}

		_, err := z.client.Create(newPath, []byte{},
			0, zk.WorldACL(zk.PermAll))
		if err != nil && err != zk.ErrNodeExists {
			return err
		}
	}
	return nil
}

// exists checks if the key exists
func (z *zookeeperKV) exists(key string) (bool, error) {
	key = z.domain + normalize(key)
	exists, _, err := z.client.Exists(key)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (z *zookeeperKV) getRetryCount() int {
	retryCount, ok := z.options[kvdb.RetryCountKey]
	if !ok {
		return defaultRetryCount
	}
	retry, err := strconv.ParseInt(retryCount, 10, 0)
	if err != nil {
		return defaultRetryCount
	}
	return int(retry)
}

func (z *zookeeperKV) getAction(action string) kvdb.KVAction {
	switch action {
	case "set":
		return kvdb.KVSet
	case "create":
		return kvdb.KVCreate
	case "get":
		return kvdb.KVGet
	default:
		return kvdb.KVUknown
	}
}

func (z *zookeeperKV) resultToKvPair(
	key string,
	value []byte,
	action string,
	stat *zk.Stat,
) *kvdb.KVPair {
	return &kvdb.KVPair{
		Key:           strings.TrimPrefix(key, z.domain+"/"),
		Value:         value,
		Action:        z.getAction(action),
		ModifiedIndex: uint64(stat.Version),
		CreatedIndex:  uint64(stat.Version),
	}
}

func init() {
	if err := kvdb.Register(Name, New, Version); err != nil {
		panic(err.Error())
	}
}
