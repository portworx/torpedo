package bucket

type S3BucketInfo struct {
	objectLock     bool
	retainCount    int64
	objectLockMode string
}

func (i *S3BucketInfo) setObjectLock(objectLock bool) *S3BucketInfo {

}

func (i *S3BucketInfo) setRetainCount(retainCount int64) *S3BucketInfo {

}

func (i *S3BucketInfo) setObjectLockMode(objectLockMode string) *S3BucketInfo {

}

type BucketController struct {
	s3Buckets map[string]*S3BucketInfo
}

func (b *BucketController) S3Bucket(bucketName string) *S3BucketConfig {
	return &S3BucketConfig{
		bucketName:     bucketName,
		objectLock:     false,
		retainCount:    0,
		objectLockMode: "",
	}
}
