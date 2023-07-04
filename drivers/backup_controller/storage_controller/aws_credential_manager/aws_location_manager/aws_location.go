package aws_location_manager

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/portworx/torpedo/drivers/backup_controller/backup_utils"
	"github.com/portworx/torpedo/pkg/s3utils"
)

func (s *AWSLocationSpec) NewSession() (*session.Session, error) {
	awsConfig := &aws.Config{
		Endpoint:         aws.String(s.GetAWSCredential().GetEndpoint()),
		Credentials:      credentials.NewStaticCredentials(s.GetAWSCredential().GetID(), s.GetAWSCredential().GetSecret(), ""),
		Region:           aws.String(s.GetAWSCredential().GetS3Region()),
		DisableSSL:       aws.Bool(s.GetAWSCredential().GetDisableSSL()),
		S3ForcePathStyle: aws.Bool(true),
	}
	sess, err := session.NewSession(awsConfig)
	if err != nil {
		return nil, backup_utils.ProcessError(err, backup_utils.StructToString(awsConfig))
	}
	return sess, nil
}

func (s *AWSLocationSpec) IsObjectLockEnabledForBucket() bool {
	return s.RetainCount > 0 && s.ObjectLock
}

func (c *AWSLocationConfig) CanCreate() error {
	return nil
}

// Create creates a AWSLocation
func (c *AWSLocationConfig) Create(objectLock bool, retainCount int64, objectLockMode string) error {
	err := c.CanCreate()
	if err != nil {
		return backup_utils.ProcessError(err)
	}
	id, secret, endpoint, s3Region, disableSSLBool := s3utils.GetAWSDetailsFromEnv()
	awsLocationSpec := NewAWSLocationSpec(NewAWSCredential(id, secret, endpoint, s3Region, disableSSLBool), objectLock, retainCount, objectLockMode)
	sess, err := awsLocationSpec.NewSession()
	if err != nil {
		return backup_utils.ProcessError(err)
	}
	S3Client := s3.New(sess)
	awsLocationName := c.GetAWSLocationMetaData().GetAWSLocationName()
	isObjectLockEnabledForBucket := awsLocationSpec.IsObjectLockEnabledForBucket()
	awsCreateBucketInput := &s3.CreateBucketInput{
		Bucket:                     aws.String(awsLocationName),
		ObjectLockEnabledForBucket: aws.Bool(isObjectLockEnabledForBucket),
	}
	_, err = S3Client.CreateBucket(awsCreateBucketInput)
	if err != nil {
		debugStruct := struct {
			AWSCreateBucketInput *s3.CreateBucketInput
		}{
			AWSCreateBucketInput: awsCreateBucketInput,
		}
		return backup_utils.ProcessError(err, backup_utils.StructToString(debugStruct))
	}
	awsHeadBucketInput := &s3.HeadBucketInput{
		Bucket: aws.String(awsLocationName),
	}
	err = S3Client.WaitUntilBucketExists(awsHeadBucketInput)
	if err != nil {
		debugStruct := struct {
			AWSHeadBucketInput *s3.HeadBucketInput
		}{
			AWSHeadBucketInput: awsHeadBucketInput,
		}
		return backup_utils.ProcessError(err, backup_utils.StructToString(debugStruct))
	}
	if isObjectLockEnabledForBucket == true {
		// Update ObjectLockConfiguration to bucket
		enabled := "Enabled"
		awsPutObjectLockConfigurationInput := &s3.PutObjectLockConfigurationInput{
			Bucket: aws.String(awsLocationName),
			ObjectLockConfiguration: &s3.ObjectLockConfiguration{
				ObjectLockEnabled: aws.String(enabled),
				Rule: &s3.ObjectLockRule{
					DefaultRetention: &s3.DefaultRetention{
						Days: aws.Int64(retainCount),
						Mode: aws.String(objectLockMode),
					},
				},
			},
		}
		_, err = S3Client.PutObjectLockConfiguration(awsPutObjectLockConfigurationInput)
		if err != nil {
			debugStruct := struct {
				AWSPutObjectLockConfigurationInput *s3.PutObjectLockConfigurationInput
			}{
				AWSPutObjectLockConfigurationInput: awsPutObjectLockConfigurationInput,
			}
			return backup_utils.ProcessError(err, backup_utils.StructToString(debugStruct))
		}
	}
	if err != nil {
		return backup_utils.ProcessError(err)
	}
	return nil
}
