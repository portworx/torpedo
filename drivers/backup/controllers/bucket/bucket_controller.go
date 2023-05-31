package bucket

type BucketType int32

const (
	S3 BucketType = iota
)

type BucketInfo struct {
}

type BucketController struct {
	buckets map[string]*BucketInfo
}

func (b *BucketController) getBucketInfo(bucketName string) *BucketInfo {
	bucketInfo, ok := b.buckets[bucketName]
	if !ok {
		return &BucketInfo{}
	}
	return bucketInfo
}

func S3Bucket(bucketName string) *S3BucketConfig {
	return
}
