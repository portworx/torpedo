package objectstore

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	stork_api "github.com/libopenstorage/stork/pkg/apis/stork/v1alpha1"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const (
	// maxRetries on failures
	s3MaxRetries = 2
	maxS3Retries = 1
	// time to wait between retries
	retryDelay = 10 * time.Second
	// standrd amzn end point
	amznAwsEndPoint = "s3.amazonaws.com"
)

var endpointMutex sync.Mutex

// S3CfgInput describes the s3 client options to create with
type S3CfgInput struct {
	Bucket           string
	S3Iam            bool
	EndPoints        []string
	AccessKeyID      string
	SecretAccessKey  string
	Region           string
	Secure           bool
	DisablePathStyle bool
	UseProxy         bool
	Proxy            string
	StorageClass     string
	LocalIP          string
}

// S3Client holsd details about s3 client with options
type S3Client struct {
	*operationAPIStats
	mu               *sync.Mutex
	renewalInrogress bool
	client           *s3.S3
	bucket           string
	s3Iam            bool
	endPoints        []string
	accessKeyID      string
	secretAccessKey  string
	region           string
	secure           bool
	disablePathStyle bool
	useProxy         bool
	proxy            string
	storageClass     string
	localIP          string
	maxRetries       int
}

func isSwitchRequiredForThisError(err error) bool {
	if strings.Contains(err.Error(), "i/o timeout") ||
		strings.Contains(err.Error(), "connect: connection refused") ||
		strings.Contains(err.Error(), "connect: no route to host") {
		return true
	}

	return false
}

// do will execute the function and renew connection to s3 client on error.
// It will retry a few times, trying to establish a new connection every time.
func (s *S3Client) do(f func() error) error {
	log := logrus.WithField("pkg", "objectstore/s3")
	err := f()

	if err != nil && isSwitchRequiredForThisError(err) {
		// try renew of client under a lock
		s.mu.Lock()
		defer s.mu.Unlock()

		// try one more time since another thread might have renewed client successfully
		log.Debug("Trying func one more time before retrying")
		if err = f(); err == nil {
			log.Debug("S3 Operation successful")
			return nil
		}

		// Retry a few times with a delay to give time for the objectstore to
		// fail over if required
		retries := s3MaxRetries
		if len(s.endPoints) > 1 {
			retries = s3MaxRetries + 3
		}
		for j := 0; j < retries; j++ {
			// try over all available end points
			for i := 0; i < len(s.endPoints); i++ {
				log.Errorf("Trying endpoint:%s on err:%v", s.endPoints[i], err)
				if renewalErr := s.renewClient(i); renewalErr != nil {
					return fmt.Errorf("%v:%v", renewalErr, err)
				}

				if err = f(); err == nil {
					log.Debugf("Successful conn with endpoint:%s", s.endPoints[i])
					return nil
				}
				if !isSwitchRequiredForThisError(err) {
					return err
				}
			}
			time.Sleep(retryDelay)
		}
	}

	return err
}

// GetS3Credentials returns credentials for s3 endpoint
func GetS3Credentials() (map[string]string, error) {
	s3Creds := map[string]string{
		"S3_ENDPOINT":          os.Getenv("S3_ENDPOINT"),
		"S3_ACCESS_KEY_ID":     os.Getenv("S3_ACCESS_KEY_ID"),
		"S3_SECRET_ACCESS_KEY": os.Getenv("S3_SECRET_ACCESS_KEY"),
		"S3_REGION":            os.Getenv("S3_REGION"),
	}
	if err := checkEnvVar(&s3Creds); err != nil {
		return nil, fmt.Errorf("S3, missing account creds: %v", err)
	}
	return s3Creds, nil
}

func loadCustomCABundle() (*os.File, error) {
	customCABundleFile := os.Getenv("AWS_CA_BUNDLE")
	if customCABundleFile == "" {
		return nil, nil
	}
	return os.Open(customCABundleFile)
}

// renewClient renews a client by trying all available end points.
// This function is not thread-safe, so it should always be called
// using a lock
func (s *S3Client) renewClient(index int) error {

	// no Proxy or unsigned certs for IAM
	if s.s3Iam {
		// defaults for endpoints, accessKey and secretkey
		config := &aws.Config{
			Region:     &(s.region),
			MaxRetries: &(s.maxRetries),
		}
		sess, err := session.NewSession(config)
		if err != nil {
			return err
		}
		client := s3.New(sess)
		s.client = client
		return nil
	}

	if index >= len(s.endPoints) {
		return fmt.Errorf("out of bounds error for endpoints list")
	}

	endPoint := s.endPoints[index]
	// push current end point into cache
	endpointMutex.Lock()
	viper.Set(s.accessKeyID, endPoint)
	endpointMutex.Unlock()
	if endPoint == amznAwsEndPoint {
		endPoint = ""
	}
	config := &aws.Config{
		Credentials: credentials.NewStaticCredentials(s.accessKeyID,
			s.secretAccessKey, ""),
		Endpoint:         &endPoint,
		DisableSSL:       aws.Bool(!s.secure),
		Region:           &(s.region),
		S3ForcePathStyle: aws.Bool(!s.disablePathStyle),
		MaxRetries:       &(s.maxRetries),
	}
	var customCABundle *os.File
	var err error
	timeout := 15
	if s.useProxy {
		s3HttpProxy := s.proxy
		if s3HttpProxy != "" {
			proxyURL, err := url.Parse(s3HttpProxy)
			if err != nil {
				logrus.Errorf("Failed to parse proxy url:%v err: %v", s3HttpProxy, err)
				return err
			}
			// default client with default transport, init the proxy field
			// TODO get rid of this and just do Clone of default transport when
			// we move to go 1.13
			proxyClient := newDefaultHTTPTransport(s.localIP, proxyURL, timeout, nil)
			config = config.WithHTTPClient(proxyClient)
		} else {
			return fmt.Errorf("no proxy env provided")
		}
		if s.secure {
			customCABundle, err = loadCustomCABundle()
			if err != nil {
				logrus.Errorf("Failed load CAbundle :%v", err)
				return err
			}
		}
	} else {
		// only set http timeout if there are multiple endpoints.
		// local objectstore, 1 min required for failover
		if len(s.endPoints) > 1 {
			timeout = 60
		}
		httpClient := newDefaultHTTPTransport(s.localIP, nil, timeout, nil)
		config = config.WithHTTPClient(httpClient)
	}

	sess := &session.Session{}
	if customCABundle == nil {
		sess, err = session.NewSession(config)
	} else {
		defer func() {
			if err := customCABundle.Close(); err != nil {
				logrus.Warnf("failed to close CA bundle: %v", err)
			}
		}()
		sess, err = session.NewSessionWithOptions(session.Options{
			Config:         *config,
			CustomCABundle: customCABundle,
		})
	}

	if err != nil {
		return err
	}

	client := s3.New(sess)
	s.client = client

	return nil
}

// NewS3Client obtains a new S3 client by trying out all available end points
// and returning on first successful connection.
func NewS3Client(
	usrConfig S3CfgInput,
) (*S3Client, error) {

	if !usrConfig.S3Iam {
		endpointMutex.Lock()
		// access cache to see if there was an
		endPoint := viper.GetString(usrConfig.AccessKeyID)

		// reorder end point list if a match is found
		for i := range usrConfig.EndPoints {
			if usrConfig.EndPoints[i] == endPoint {
				usrConfig.EndPoints[0], usrConfig.EndPoints[i] = usrConfig.EndPoints[i], usrConfig.EndPoints[0]
				break
			}
		}

		// push first end point back to cache
		endPoint = usrConfig.EndPoints[0]
		viper.Set(usrConfig.AccessKeyID, endPoint)
		endpointMutex.Unlock()
	}
	s3client := new(S3Client)
	s3client.endPoints = usrConfig.EndPoints
	s3client.accessKeyID = usrConfig.AccessKeyID
	s3client.secretAccessKey = usrConfig.SecretAccessKey
	s3client.region = usrConfig.Region
	s3client.secure = usrConfig.Secure
	s3client.bucket = usrConfig.Bucket
	s3client.disablePathStyle = usrConfig.DisablePathStyle
	s3client.useProxy = usrConfig.UseProxy
	s3client.proxy = usrConfig.Proxy
	s3client.s3Iam = usrConfig.S3Iam
	s3client.mu = new(sync.Mutex)
	s3client.storageClass = usrConfig.StorageClass
	s3client.localIP = usrConfig.LocalIP
	s3client.maxRetries = maxS3Retries
	s3client.operationAPIStats = &operationAPIStats{}
	if err := s3client.renewClient(0); err != nil {
		return nil, err
	}
	return s3client, nil
}

// ListBuckets lists the buckets for given end point
func (s *S3Client) ListBuckets(continuationToken string) ([]Bucket, bool, string, error) {
	var buckets []Bucket

	f := func() error {
		s.IncrementListCount()
		input := &s3.ListBucketsInput{}
		output, err := s.client.ListBuckets(input)
		if err != nil {
			// May not have listbuckets permission. return preconfigured
			// bucket if configured
			if s.bucket != "" {
				buckets = append(buckets, Bucket{Name: s.bucket})
				return nil
			}
			return err
		}

		for i := range output.Buckets {
			bucket := Bucket{
				Name: aws.StringValue(output.Buckets[i].Name),
			}
			buckets = append(buckets, bucket)
		}

		return nil
	}

	if err := s.do(f); err != nil {
		return buckets, false, "", err
	}

	return buckets, false, "", nil
}

// ListObjectsV1 lists objects for given prefix using V1 api
func (s *S3Client) ListObjectsV1(
	bucket string,
	delimiter string,
	prefix string,
	continuationToken string,
	version string,
	maxObjects int64,
) ([]Object, []string, bool, string, string, error) {
	moreObjects := false
	nextToken := ""
	ver := "v1"
	input := &s3.ListObjectsInput{
		Bucket:    &bucket,
		Prefix:    &prefix,
		Delimiter: &delimiter,
	}

	if maxObjects != 0 {
		input.MaxKeys = &maxObjects
	}
	if continuationToken != "" {
		input.Marker = &continuationToken
	}

	var output *s3.ListObjectsOutput
	var err error

	f := func() error {
		s.IncrementListCount()
		output, err = s.client.ListObjects(input)
		return err
	}

	err = s.do(f)
	if err != nil {
		return nil, nil, false, "", ver, err
	}
	var objects []Object
	for i := range output.Contents {
		object := Object{
			Name:         aws.StringValue(output.Contents[i].Key),
			Size:         uint64(aws.Int64Value(output.Contents[i].Size)),
			LastModified: aws.TimeValue(output.Contents[i].LastModified),
		}
		objects = append(objects, object)
	}
	if aws.BoolValue(output.IsTruncated) {
		moreObjects = true
		nextToken = aws.StringValue(output.NextMarker)
	}

	var prefixes []string
	for i := range output.CommonPrefixes {
		prefix := aws.StringValue(output.CommonPrefixes[i].Prefix)
		prefixes = append(prefixes, prefix)
	}
	return objects, prefixes, moreObjects, nextToken, ver, nil
}

// ListObjects lists objects for given prefix using V1/V2 api
func (s *S3Client) ListObjects(
	bucket string,
	delimiter string,
	prefix string,
	continuationToken string,
	version string,
	maxObjects int64,
) ([]Object, []string, bool, string, string, error) {

	moreObjects := false
	nextToken := ""
	ver := "v2"
	if version == "v1" {
		return s.ListObjectsV1(bucket, delimiter, prefix, continuationToken, version, maxObjects)
	}
	input := &s3.ListObjectsV2Input{
		Bucket:    &bucket,
		Prefix:    &prefix,
		Delimiter: &delimiter,
	}
	if maxObjects != 0 {
		input.MaxKeys = &maxObjects
	}

	if continuationToken != "" {
		input.ContinuationToken = &continuationToken
	}

	var output *s3.ListObjectsV2Output
	var err error

	f := func() error {
		s.IncrementListCount()
		output, err = s.client.ListObjectsV2(input)
		return err
	}

	err = s.do(f)
	if err != nil {
		return nil, nil, false, "", ver, err
	}

	var objects []Object
	for i := range output.Contents {
		object := Object{
			Name:         aws.StringValue(output.Contents[i].Key),
			Size:         uint64(aws.Int64Value(output.Contents[i].Size)),
			LastModified: aws.TimeValue(output.Contents[i].LastModified),
		}
		objects = append(objects, object)
	}

	if aws.BoolValue(output.IsTruncated) {
		moreObjects = true
		nextToken = aws.StringValue(output.NextContinuationToken)
		if nextToken == "" {
			return s.ListObjectsV1(bucket, delimiter, prefix, continuationToken, version, maxObjects)
		}
	}

	var prefixes []string
	for i := range output.CommonPrefixes {
		prefix := aws.StringValue(output.CommonPrefixes[i].Prefix)
		prefixes = append(prefixes, prefix)
	}

	return objects, prefixes, moreObjects, nextToken, ver, nil
}

// ValidateBackupsDeletedFromCloud checks it given backups are deleted from the cloud
func (s *S3Client) ValidateBackupsDeletedFromCloud(backupLocation *stork_api.BackupLocation, backupPath string) error {
	return fmt.Errorf("Not Supported")
}
