package bucket

type BucketType int32

const (
	S3 BucketType = iota
)

type S3BucketInfo struct {
}

type BucketController struct {
	s3Buckets map[string]*S3BucketInfo
}

func (b *BucketController) getS3BucketInfo(bucketName string) {

}

func (b *BucketController) S3Bucket(bucketName string) *S3BucketConfig {
	return
}
