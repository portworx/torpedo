package zookeeper

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/portworx/kvdb"
	"github.com/portworx/kvdb/test"
	"github.com/stretchr/testify/assert"
)

const (
	domain  = "pwx/test"
	dataDir = "/data/zookeeper"
)

func TestAll(t *testing.T) {
	test.RunBasic(New, t, Start, Stop)
}

func TestVersion(t *testing.T) {
	fmt.Println("verify version")
	version, err := Version("", nil)
	assert.NoError(t, err, "Unexpected error on Version")
	assert.Equal(t, kvdb.ZookeeperVersion1, version)
}

func TestZookeeperOps(t *testing.T) {
	err := Start(true)
	assert.NoError(t, err, "Unable to start kvdb")
	// Wait for kvdb to start
	time.Sleep(5 * time.Second)

	kv, err := New(domain, nil, nil, nil)
	if err != nil {
		t.Fatalf(err.Error())
	}

	testName(t, kv)
	testCreateWithTTL(t, kv)
	testPutWithTTL(t, kv)
	testUpdateWithTTL(t, kv)
	testCreateEphemeral(t)
	testLockBetweenClientRestarts(t)
	testLockWithIDBetweenClientRestarts(t)
	testLockWithTimeoutBetweenClientRestarts(t)

	err = Stop()
	assert.NoError(t, err, "Unable to stop kvdb")
}

func testName(t *testing.T, kv kvdb.Kvdb) {
	fmt.Println("verify name")
	assert.Equal(t, Name, kv.String())
}

func testCreateWithTTL(t *testing.T, kv kvdb.Kvdb) {
	fmt.Println("create with ttl")
	key := "create/foottl"
	kv.Delete(key)

	_, err := kv.Create(key, []byte("barttl"), 10)
	assert.Equal(t, err, kvdb.ErrTTLNotSupported)
}

func testPutWithTTL(t *testing.T, kv kvdb.Kvdb) {
	fmt.Println("put with ttl")
	key := "put/foottl"
	kv.Delete(key)

	_, err := kv.Put(key, []byte("barttl"), 10)
	assert.Equal(t, err, kvdb.ErrTTLNotSupported)
}

func testUpdateWithTTL(t *testing.T, kv kvdb.Kvdb) {
	fmt.Println("update with ttl")
	key := "update/foottl"
	kv.Delete(key)

	_, err := kv.Create(key, []byte("bar"), 0)
	assert.NoError(t, err, "Unexpected error on create")

	_, err = kv.Update(key, []byte("barttl"), 10)
	assert.Equal(t, err, kvdb.ErrTTLNotSupported)
}

func testCreateEphemeral(t *testing.T) {
	fmt.Println("create ephemeral node")
	zk, err := newClient(domain, nil, nil, nil)
	assert.NoError(t, err)
	assert.NotNil(t, zk)

	key := "create/ephemeral"
	value := []byte("val")
	zk.Delete(key)

	kvp, err := zk.createEphemeral(key, value)
	assert.NoError(t, err, "Unexpected error on ephemeral create")

	defer func() {
		zk.Delete(key)
	}()
	assert.Equal(t, kvp.Action, kvdb.KVCreate,
		"Expected action KVCreate, action %v", kvp.Action)
	assert.Equal(t, value, kvp.Value)
	assert.Equal(t, uint64(1), kvp.ModifiedIndex)

	_, err = zk.createEphemeral(key, []byte("val"))
	assert.Equal(t, err, kvdb.ErrExist)
}

func testLockBetweenClientRestarts(t *testing.T) {
	fmt.Println("Lock between client restarts")
	zk, err := newClient(domain, nil, nil, nil)
	assert.NoError(t, err, "Unable to create a client")
	assert.NotNil(t, zk)

	zk.SetLockTimeout(time.Minute)

	// Lock before restarting client
	kvPair, err := zk.Lock("lock_key")
	assert.NoError(t, err, "Unable to take a lock")

	// Stopping client connection
	zk.closeClient()

	// We don't need to wait for session timeout, as we are doing
	// a proper session close. In this case, the server will delete
	// all ephemeral nodes created in that session immediately.

	// Reconnecting the client
	zk, err = newClient(domain, nil, nil, nil)
	assert.NoError(t, err, "Unable to reconnect client")

	// Locking again should succeed as the previous client died
	lockChan := make(chan struct{})
	go func() {
		kvPair, err = zk.Lock("lock_key")
		lockChan <- struct{}{}
	}()
	select {
	case <-time.After(5 * time.Second):
		assert.Fail(t, "Unable to take a lock even when previous session was expired")
	case <-lockChan:
	}
	err = zk.Unlock(kvPair)
	assert.NoError(t, err, "Unable to unlock")
}

func testLockWithIDBetweenClientRestarts(t *testing.T) {
	fmt.Println("LockWithID between client restarts")
	zk, err := newClient(domain, nil, nil, nil)
	assert.NoError(t, err)
	assert.NotNil(t, zk)

	zk.SetLockTimeout(time.Minute)

	// Lock before restarting client
	kvPair, err := zk.LockWithID("lock_key", "lock_with_id")
	assert.NoError(t, err, "Unable to take a lock")

	// Stopping client connection
	zk.closeClient()

	// We don't need to wait for session timeout, as we are doing
	// a proper session close. In this case, the server will delete
	// all ephemeral nodes created in that session immediately.

	// Reconnecting the client
	zk, err = newClient(domain, nil, nil, nil)
	assert.NoError(t, err, "Unable to reconnect client")

	// Locking again should succeed as the previous client died
	lockChan := make(chan struct{})
	go func() {
		kvPair, err = zk.LockWithID("lock_key", "lock_with_id")
		lockChan <- struct{}{}
	}()
	select {
	case <-time.After(5 * time.Second):
		assert.Fail(t, "Unable to take a lock even when previous session was expired")
	case <-lockChan:
	}
	err = zk.Unlock(kvPair)
	assert.NoError(t, err, "Unable to unlock")
}

func testLockWithTimeoutBetweenClientRestarts(t *testing.T) {
	fmt.Println("LockWithTimeout between client restarts")
	zk, err := newClient(domain, nil, nil, nil)
	assert.NoError(t, err)
	assert.NotNil(t, zk)

	zk.SetLockTimeout(time.Minute)

	// Lock before restarting client
	kvPair, err := zk.LockWithTimeout("lock_key", "lock_with_id",
		10*time.Second, time.Minute)
	assert.NoError(t, err, "Unable to take a lock")

	// Stopping client connection
	zk.closeClient()

	// We don't need to wait for session timeout, as we are doing
	// a proper session close. In this case, the server will delete
	// all ephemeral nodes created in that session immediately.

	// Reconnecting the client
	zk, err = newClient(domain, nil, nil, nil)
	assert.NoError(t, err, "Unable to reconnect client")

	// Locking again should succeed as the previous client died
	lockChan := make(chan struct{})
	go func() {
		kvPair, err = zk.LockWithTimeout("lock_key", "lock_with_id",
			10*time.Second, time.Minute)
		lockChan <- struct{}{}
	}()
	select {
	case <-time.After(5 * time.Second):
		assert.Fail(t, "Unable to take a lock even when previous session was expired")
	case <-lockChan:
	}
	err = zk.Unlock(kvPair)
	assert.NoError(t, err, "Unable to unlock")
}

func Start(removeData bool) error {
	if removeData {
		err := os.RemoveAll(dataDir)
		if err != nil {
			return err
		}
	}
	err := os.MkdirAll(dataDir, 0644)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(dataDir+"/myid", []byte("1"), 0644)
	if err != nil {
		return err
	}
	cmd := exec.Command("/tmp/test-zookeeper/bin/zkServer.sh", "start")
	err = cmd.Start()
	time.Sleep(5 * time.Second)
	return err
}

func Stop() error {
	cmd := exec.Command("/tmp/test-zookeeper/bin/zkServer.sh", "stop")
	err := cmd.Start()
	time.Sleep(5 * time.Second)
	return err
}
