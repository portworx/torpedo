package pdsbackup

import (
	"context"
	"fmt"
	"github.com/minio/minio-go/v7"
	"os"
	"time"

	"cloud.google.com/go/storage"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	minioCred "github.com/minio/minio-go/v7/pkg/credentials"
	log "github.com/sirupsen/logrus"
)

type awsStorageClient struct {
	accessKey string
	secretKey string
	region    string
}

type azureStorageClient struct {
	accountName string
	accountKey  string
}

type gcpStorageClient struct {
	projectId string
	jsongPath string
}

type s3CompatibleStorageClient struct {
	accessKey string
	secretKey string
	region    string
	endpoint  string
}

func (awsObj *awsStorageClient) createBucket(bucketName string) error {
	sess, err := session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Region:      aws.String(awsObj.region),
			Credentials: credentials.NewStaticCredentials(awsObj.accessKey, awsObj.secretKey, ""),
		},
	})

	if err != nil {
		log.Errorf("Failed to initialize new session: %v", err)
		return err
	}

	client := s3.New(sess)
	bucketObj, err := client.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	})

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			if (aerr.Code() == s3.ErrCodeBucketAlreadyOwnedByYou) || (aerr.Code() == s3.ErrCodeBucketAlreadyExists) {
				log.Infof("Bucket: %v ,already exist.", bucketName)
				return nil
			} else {
				log.Errorf("Couldn't create bucket: %v", err)
				return err
			}

		}

	}

	log.Infof("[AWS]Successfully created the bucket. Info: %v", bucketObj)
	return nil
}

func (awsObj *awsStorageClient) deleteBucket(bucketName string) error {
	sess, err := session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Region:      aws.String(awsObj.region),
			Credentials: credentials.NewStaticCredentials(awsObj.accessKey, awsObj.secretKey, ""),
		},
	})

	if err != nil {
		log.Infof("Failed to initialize new session: %v", err)
		return err
	}

	client := s3.New(sess)
	_, err = client.DeleteBucket(&s3.DeleteBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			if aerr.Code() == s3.ErrCodeNoSuchBucket {
				log.Infof("[AWS] Bucket: %v doesn't exists.!!", bucketName)
				return nil
			} else {
				log.Errorf("Couldn't delete bucket: %v", err)
				return err
			}

		}

	}
	log.Infof("[AWS] Successfully deleted the bucket: %v", bucketName)
	return nil
}

func (azObj *azureStorageClient) createBucket(containerName string) error {
	cred, err := azblob.NewSharedKeyCredential(azObj.accountName, azObj.accountKey)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	client, err := azblob.NewClientWithSharedKeyCredential(fmt.Sprintf("https://%s.blob.core.windows.net/", azObj.accountName), cred, nil)
	if err != nil {
		log.Error(err.Error())
		return err
	}

	_, err = client.CreateContainer(context.TODO(), containerName, nil)
	if err != nil {
		log.Infof("Container: %s, already exists.", containerName)
	} else {
		log.Infof("[Azure]Successfully created the container: %s", containerName)
	}
	return nil
}

func (azObj *azureStorageClient) deleteBucket(containerName string) error {
	cred, err := azblob.NewSharedKeyCredential(azObj.accountName, azObj.accountKey)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	client, err := azblob.NewClientWithSharedKeyCredential(fmt.Sprintf("https://%s.blob.core.windows.net/", azObj.accountName), cred, nil)
	if err != nil {
		log.Error(err.Error())
		return err
	}

	_, err = client.DeleteContainer(context.TODO(), containerName, nil)
	if err != nil {
		log.Infof("[Azure]Container: %s not found!!", containerName)
	} else {
		log.Infof("[Azure]Container: %s deleted successfully!!", containerName)
	}
	return nil
}

func (gcpObj *gcpStorageClient) createBucket(bucketName string) error {
	ctx := context.Background()
	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", gcpObj.jsongPath)
	client, err := storage.NewClient(context.Background())
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	defer client.Close()

	ctx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()
	bucketClient := client.Bucket(bucketName)
	exist, err := bucketClient.Attrs(ctx)

	if exist != nil {
		if err != nil {
			log.Errorf("Unexpected error occured: %v", err)
			return err
		} else {
			log.Infof("[GCP] Bucket: %v already exists.!!", bucketName)
		}
	} else {
		err := bucketClient.Create(ctx, gcpObj.projectId, nil)
		if err != nil {
			log.Errorf("Bucket(%v).Create: %v", bucketName, err)
			return err
		}
		log.Infof("[GCP] Successfully create the Bucket: %v", bucketName)
	}
	return nil

}

func (gcpObj *gcpStorageClient) deleteBucket(bucketName string) error {
	ctx := context.Background()
	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", gcpObj.jsongPath)
	client, err := storage.NewClient(context.Background())
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	defer client.Close()

	ctx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	bucketClient := client.Bucket(bucketName)
	exist, err := bucketClient.Attrs(ctx)

	if exist != nil {
		if err != nil {
			log.Errorf("Unexpected error occured: %v", err)
			return err
		} else {
			err := bucketClient.Delete(ctx)
			if err != nil {
				log.Errorf("Bucket(%v).Delete: %v", bucketName, err)
				return err
			}
			log.Infof("[GCP]Successfully deleted the Bucket: %v", bucketName)
		}
	} else {
		log.Infof("[GCP]Bucket: %v doesn't exist.", bucketName)
	}
	return nil

}

func (minioObj *s3CompatibleStorageClient) createBucket(bucketName string) error {
	minioClient, err := minio.New(minioObj.endpoint, &minio.Options{
		Creds:  minioCred.NewStaticV4(minioObj.accessKey, minioObj.secretKey, ""),
		Secure: true,
	})
	if err != nil {
		log.Error(err)
	}
	found, err := minioClient.BucketExists(context.Background(), bucketName)
	if err != nil {
		log.Error(err)
		return err
	}
	if !found {
		err = minioClient.MakeBucket(context.Background(), bucketName, minio.MakeBucketOptions{Region: minioObj.region})
		if err != nil {
			fmt.Println(err)
			return err
		}
		log.Infof("[MINIO] Successfully created bucket:%v.", bucketName)
	} else {
		log.Infof("[MINIO] Bucket:%v already exists.", bucketName)
	}
	return nil

}

func (minioObj *s3CompatibleStorageClient) deleteBucket(bucketName string) error {
	minioClient, err := minio.New(minioObj.endpoint, &minio.Options{
		Creds:  minioCred.NewStaticV4(minioObj.accessKey, minioObj.secretKey, ""),
		Secure: true,
	})
	if err != nil {
		log.Error(err)
	}
	found, err := minioClient.BucketExists(context.Background(), bucketName)
	if err != nil {
		log.Error(err)
		return err
	}
	if found {
		err = minioClient.RemoveBucket(context.Background(), bucketName)
		if err != nil {
			fmt.Println(err)
			return err
		}
		log.Infof("[MINIO] Successfully deleted the bucket: %v", bucketName)
	} else {
		log.Infof("[MINIO] Bucket:%v doesn't exist.", bucketName)
	}
	return nil

}
