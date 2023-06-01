package bucket

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/portworx/torpedo/drivers/backup/utils"
	"github.com/portworx/torpedo/pkg/log"
	"github.com/portworx/torpedo/pkg/s3utils"
)

type S3BucketConfig struct {
	bucketName     string
	objectLock     bool
	retainCount    int64
	objectLockMode string
}

func (b *BucketController) getS3BucketInfo(bucketName string) *S3BucketInfo {
	return nil
}

func (c *S3BucketConfig) Create() error {
	id, secret, endpoint, s3Region, disableSSLBool := s3utils.GetAWSDetailsFromEnv()
	sess, err := session.NewSession(
		&aws.Config{
			Endpoint:         aws.String(endpoint),
			Credentials:      credentials.NewStaticCredentials(id, secret, ""),
			Region:           aws.String(s3Region),
			DisableSSL:       aws.Bool(disableSSLBool),
			S3ForcePathStyle: aws.Bool(true),
		},
	)
	if err != nil {
		return utils.ProcessError(err)
	}
	S3Client := s3.New(sess)
	if c.retainCount > 0 && c.objectLock == true {
		log.Infof("Creating object locked bucket [%s]", c.bucketName)
		_, err = S3Client.CreateBucket(&s3.CreateBucketInput{
			Bucket:                     aws.String(c.bucketName),
			ObjectLockEnabledForBucket: aws.Bool(true),
		})
	} else {
		log.Infof("Creating standard [%s]", c.bucketName)
		_, err = S3Client.CreateBucket(&s3.CreateBucketInput{
			Bucket: aws.String(c.bucketName),
		})
	}
	if err != nil {
		return utils.ProcessError(err)
	}
	err = S3Client.WaitUntilBucketExists(&s3.HeadBucketInput{
		Bucket: aws.String(c.bucketName),
	})
	if err != nil {
		return utils.ProcessError(err)
	}
	if c.retainCount > 0 && c.objectLock == true {
		// Update ObjectLockConfiguration to bucket
		enabled := "Enabled"
		_, err = S3Client.PutObjectLockConfiguration(
			&s3.PutObjectLockConfigurationInput{
				Bucket: aws.String(c.bucketName),
				ObjectLockConfiguration: &s3.ObjectLockConfiguration{
					ObjectLockEnabled: aws.String(enabled),
					Rule: &s3.ObjectLockRule{
						DefaultRetention: &s3.DefaultRetention{
							Days: aws.Int64(c.retainCount),
							Mode: aws.String(c.objectLockMode),
						},
					},
				},
			},
		)
		if err != nil {
			//err = fmt.Errorf("Failed to update Objectlock config with Retain Count [%v] and Mode [%v]. Error: [%v]", c.retainCount, c.objectLockMode, err)
			return utils.ProcessError(err)
		}
	}
	return err
}

// DeleteS3Bucket deletes bucket in S3
func (c *S3BucketConfig) Delete() error {
	id, secret, endpoint, s3Region, disableSSLBool := s3utils.GetAWSDetailsFromEnv()
	sess, err := session.NewSession(
		&aws.Config{
			Endpoint:         aws.String(endpoint),
			Credentials:      credentials.NewStaticCredentials(id, secret, ""),
			Region:           aws.String(s3Region),
			DisableSSL:       aws.Bool(disableSSLBool),
			S3ForcePathStyle: aws.Bool(true),
		},
	)

	if err != nil {
		//err = fmt.Errorf("Failed to update Objectlock config with Retain Count [%v] and Mode [%v]. Error: [%v]", c.retainCount, c.objectLockMode, err)
		return utils.ProcessError(err)
	}

	//expect(err).NotTo(haveOccurred(),
	//	fmt.Sprintf("Failed to get S3 session to create bucket. Error: [%v]", err))

	S3Client := s3.New(sess)

	iter := s3manager.NewDeleteListIterator(S3Client, &s3.ListObjectsInput{
		Bucket: aws.String(c.bucketName),
	})

	err = s3manager.NewBatchDeleteWithClient(S3Client).Delete(aws.BackgroundContext(), iter)
	if err != nil {
		return utils.ProcessError(err)
	}
	_, err = S3Client.DeleteBucket(&s3.DeleteBucketInput{
		Bucket: aws.String(c.bucketName),
	})
	if err != nil {
		return utils.ProcessError(err)
	}
	return nil
}

// IsEmpty returns true if bucket empty else false
func (c *S3BucketConfig) IsEmpty() (bool, error) {
	id, secret, endpoint, s3Region, disableSSLBool := s3utils.GetAWSDetailsFromEnv()
	sess, err := session.NewSession(&aws.Config{
		Endpoint:         aws.String(endpoint),
		Credentials:      credentials.NewStaticCredentials(id, secret, ""),
		Region:           aws.String(s3Region),
		DisableSSL:       aws.Bool(disableSSLBool),
		S3ForcePathStyle: aws.Bool(true),
	})
	if err != nil {
		return false, fmt.Errorf("failed to get S3 session to create bucket with %s", err)
	}

	S3Client := s3.New(sess)
	maxKeys := int64(1)
	input := &s3.ListObjectsInput{
		Bucket:  aws.String(c.bucketName),
		MaxKeys: &maxKeys,
	}

	result, err := S3Client.ListObjects(input)
	if err != nil {
		return false, fmt.Errorf("unable to fetch cotents from s3 failing with %s", err)
	}

	log.Info(fmt.Sprintf("Result content %d", len(result.Contents)))
	if len(result.Contents) > 0 {
		return false, nil
	}
	return true, nil
}
