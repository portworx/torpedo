package pdsbackup

import (
	"context"
	"fmt"
	"os"
	"time"

	"cloud.google.com/go/storage"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/portworx/torpedo/pkg/log"
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
}

func (awsObj *awsStorageClient) createBucket(bucketName string) error {
	sess, err := session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Region:      aws.String(awsObj.region),
			Credentials: credentials.NewStaticCredentials(awsObj.accessKey, awsObj.secretKey, ""),
		},
	})

	if err != nil {
		return fmt.Errorf("failed to initialize new session: %v", err)
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
				return fmt.Errorf("couldn't create bucket: %v", err)
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
		return fmt.Errorf("failed to initialize new session: %v", err)
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
				return fmt.Errorf("couldn't delete bucket: %v", err)
			}

		}

	}
	log.Infof("[AWS] Successfully deleted the bucket: %v", bucketName)
	return nil
}

func (azObj *azureStorageClient) createBucket(containerName string) error {
	cred, err := azblob.NewSharedKeyCredential(azObj.accountName, azObj.accountKey)
	if err != nil {
		return err
	}
	client, err := azblob.NewClientWithSharedKeyCredential(fmt.Sprintf("https://%s.blob.core.windows.net/", azObj.accountName), cred, nil)
	if err != nil {
		return fmt.Errorf("error -> %v", err.Error())
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
		return fmt.Errorf("error -> %v", err.Error())
	}
	client, err := azblob.NewClientWithSharedKeyCredential(fmt.Sprintf("https://%s.blob.core.windows.net/", azObj.accountName), cred, nil)
	if err != nil {
		return fmt.Errorf("error -> %v", err.Error())
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
	createGcpJsonFile("/tmp/json")
	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "/tmp/json")
	client, err := storage.NewClient(context.Background())
	if err != nil {
		return fmt.Errorf("failed to create gcp client: %v", err)
	}

	defer client.Close()

	ctx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()
	bucketClient := client.Bucket(bucketName)
	exist, err := bucketClient.Attrs(ctx)
	if err != nil {
		return err
	}
	if exist != nil {
		if err != nil {
			return fmt.Errorf("unexpected error occured: %v", err)
		} else {
			log.Infof("[GCP] Bucket: %v already exists.!!", bucketName)
		}
	} else {
		err := bucketClient.Create(ctx, gcpObj.projectId, nil)
		if err != nil {
			return fmt.Errorf("Bucket(%v).Create: %v", bucketName, err)
		}
		log.Infof("[GCP] Successfully create the Bucket: %v", bucketName)
	}
	return nil

}

func (gcpObj *gcpStorageClient) deleteBucket(bucketName string) error {
	ctx := context.Background()
	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "/tmp/json")
	client, err := storage.NewClient(context.Background())
	if err != nil {
		return fmt.Errorf("failed to create client: %v", err)
	}

	defer client.Close()

	ctx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	bucketClient := client.Bucket(bucketName)
	exist, err := bucketClient.Attrs(ctx)

	if exist != nil {
		if err != nil {
			return fmt.Errorf("unexpected error occured: %v", err)
		} else {
			err := bucketClient.Delete(ctx)
			if err != nil {
				return fmt.Errorf("Bucket(%v).Delete: %v", bucketName, err)
			}
			log.Infof("[GCP]Successfully deleted the Bucket: %v", bucketName)
		}
	} else {
		log.Infof("[GCP]Bucket: %v doesn't exist.", bucketName)
	}
	return nil
}

func createGcpJsonFile(path string) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("error while creating the file -> %v", err)
	}
	defer f.Close()
	_, err = f.WriteString(os.Getenv("PDS_QA_GCP_JSON_PATH"))
	if err != nil {
		return fmt.Errorf("error while writing the data to file -> %v", err)
	}
	return nil
}
