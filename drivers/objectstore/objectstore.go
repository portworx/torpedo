package objectstore

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	stork_api "github.com/libopenstorage/stork/pkg/apis/stork/v1alpha1"
	"github.com/portworx/sched-ops/k8s/stork"
	"github.com/portworx/torpedo/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

var (
	objectstoredriver = make(map[string]Driver)
	k8sStork          = stork.Instance()
)

const (
	driverName           = "objectstore"
	defaultRetryInterval = 10 * time.Second
	defaultTimeout       = 2 * time.Minute
)

// Object gives details about the object being inspected
type Object struct {
	Name         string
	Size         uint64
	LastModified time.Time
}

// Bucket is the container holding the backups in cloud
type Bucket struct {
	Name string
}

// Driver defines an external volume driver interface that must be implemented
type Driver interface {
	// String returns the string name of this driver.
	String() string

	// ValidateBackupsDeletedFromCloud validates if bucket has been deleted from the cloud objectstore
	ValidateBackupsDeletedFromCloud(backupLocation *stork_api.BackupLocation, backupPath string) error

	// ListBuckets()
	ListBuckets(continuationToken string) ([]Bucket, bool, string, error)

	// ListObjectS()
	ListObjects(
		bucket string,
		delimiter string,
		prefix string,
		continuationToken string,
		version string,
		maxObjects int64,
	) ([]Object, []string, bool, string, string, error)

	// CheckConnection() error
	CheckConnection() error
}

type objstore struct {
	DefaultDriver
}

// Get returns the objecstore drive
func Get() (Driver, error) {
	d, ok := objectstoredriver[driverName]
	if ok {
		return d, nil
	}

	return nil, &errors.ErrNotFound{
		ID:   driverName,
		Type: "ObjectstoreDriver",
	}
}

// Register registers the objectstore driver
func Register(driverName string, d Driver) error {
	if _, ok := objectstoredriver[driverName]; !ok {
		objectstoredriver[driverName] = d
	} else {
		return fmt.Errorf("objecstore driver: %s is already registered", driverName)
	}

	fmt.Printf("Successfully registered objectstore driver %s \n", driverName)
	return nil
}

func (o *objstore) String() string {
	return driverName
}

func (o *objstore) ListBuckets(continuationToken string) ([]Bucket, bool, string, error) {
	return []Bucket{}, false, "", nil
}

func (o *objstore) ListObjects(
	bucket string,
	delimiter string,
	prefix string,
	continuationToken string,
	version string,
	maxObjects int64,
) ([]Object, []string, bool, string, string, error) {
	return []Object{}, []string{}, false, "", "", nil
}

func (o *objstore) CheckConnection() error {
	return nil
}

func checkEnvVar(envVars *map[string]string) error {
	var missingVars []string
	for name, v := range *envVars {
		if v == "" {
			missingVars = append(missingVars, name)
		}
	}
	if len(missingVars) > 0 {
		return fmt.Errorf("Missing environment variables %v", missingVars)
	}
	return nil
}

// OperationStats describes the stats involved in operation
type OperationStats struct {
	GetCount     uint64
	ListCount    uint64
	DownloadSize uint64
	UploadSize   uint64
}

type operationAPIStats struct {
	sync.Mutex
	OperationStats
}

// IncrementGetCount increments the get count by one
func (o *operationAPIStats) IncrementGetCount() {
	o.Lock()
	defer o.Unlock()
	o.GetCount++
}

// IncrementListCount increments the list count by one
func (o *operationAPIStats) IncrementListCount() {
	o.Lock()
	defer o.Unlock()
	o.ListCount++
}

func newDefaultHTTPTransport(
	localIP string,
	proxyURL *url.URL,
	timeoutSeconds int,
	tr *oauth2.Transport,

) *http.Client {
	var proxyFn func(*http.Request) (*url.URL, error)
	if proxyURL != nil {
		proxyFn = http.ProxyURL(proxyURL)
	}
	var localAddr net.Addr
	var err error
	if localIP != "" {
		localAddr, err = net.ResolveTCPAddr("tcp", localIP+":0")
		if err != nil {
			logrus.Infof("Err resolving tcp addr:%v for:%v", err, localIP)
		}
	}

	if tr != nil {
		tr.Base = &http.Transport{
			Proxy: proxyFn,
			Dial: (&net.Dialer{
				Timeout:   time.Duration(timeoutSeconds) * time.Second,
				KeepAlive: 30 * time.Second,
				DualStack: true,
				LocalAddr: localAddr,
			}).Dial,
			MaxIdleConns:          64,
			MaxIdleConnsPerHost:   64,
			MaxConnsPerHost:       64,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		}

		return &http.Client{
			Transport: tr,
		}
	}

	return &http.Client{
		Transport: &http.Transport{
			Proxy: proxyFn,
			Dial: (&net.Dialer{
				Timeout:   time.Duration(timeoutSeconds) * time.Second,
				KeepAlive: 30 * time.Second,
				DualStack: true,
				LocalAddr: localAddr,
			}).Dial,
			MaxIdleConns:          64,
			MaxIdleConnsPerHost:   64,
			MaxConnsPerHost:       64,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
}

func init() {
	Register(driverName, &objstore{})
}
