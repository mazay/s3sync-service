package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func generateS3Key(bucketPath string, root string, path string) string {
	relativePath, _ := filepath.Rel(root, path)
	return filepath.Join(bucketPath, relativePath)
}

func getS3Session(site Site) *session.Session {
	config := aws.Config{
		Region:     aws.String(site.BucketRegion),
		MaxRetries: aws.Int(-1),
	}

	if site.AccessKey != "" && site.SecretAccessKey != "" {
		config.Credentials = credentials.NewStaticCredentials(site.AccessKey, site.SecretAccessKey, "")
	}

	return session.Must(session.NewSession(&config))
}

func getS3Service(site Site) *s3.S3 {
	return s3.New(getS3Session(site))
}

func getAwsS3ItemMap(s3Service *s3.S3, bucketName string) (map[string]string, error) {
	var loi s3.ListObjectsInput
	var items = make(map[string]string)

	loi.SetBucket(bucketName)
	obj, err := s3Service.ListObjects(&loi)

	if err == nil {
		for _, s3obj := range obj.Contents {
			eTag := strings.Trim(*(s3obj.ETag), "\"")
			items[*(s3obj.Key)] = eTag
		}
		return items, nil
	}

	return nil, err
}

func uploadFile(s3Service *s3.S3, file string, timeout time.Duration, site Site) {
	var cancelFn func()

	ctx := context.Background()

	if timeout > 0 {
		ctx, cancelFn = context.WithTimeout(ctx, timeout)
	}

	if cancelFn != nil {
		defer cancelFn()
	}

	// Set default value for StorageClass, available values are here
	// https://docs.aws.amazon.com/AmazonS3/latest/dev/storage-class-intro.html#sc-compare
	if site.StorageClass == "" {
		site.StorageClass = "STANDARD"
	}

	s3Key := generateS3Key(site.BucketPath, site.LocalPath, file)

	f, fileErr := os.Open(file)

	if fileErr != nil {
		logger.Printf("failed to open file %q, %v", file, fileErr)
	} else {
		_, err := s3Service.PutObjectWithContext(ctx, &s3.PutObjectInput{
			Bucket:       aws.String(site.Bucket),
			Key:          aws.String(s3Key),
			Body:         f,
			StorageClass: aws.String(site.StorageClass),
		})

		if err != nil {
			if aerr, ok := err.(awserr.Error); ok && aerr.Code() == request.CanceledErrorCode {
				logger.Printf("upload canceled due to timeout, %v\n", err)
			} else {
				logger.Printf("failed to upload object, %v\n", err)
			}
			os.Exit(1)
		}

		logger.Printf("successfully uploaded file: %s/%s\n", site.Bucket, s3Key)
	}
}

func deleteFile(s3Service *s3.S3, bucketName string, s3Key string) {
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(s3Key),
	}

	_, err := s3Service.DeleteObject(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				logger.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			logger.Println(err.Error())
		}
		return
	}

	logger.Printf("removed s3 object: %s/%s\n", bucketName, s3Key)
}

func syncSite(timeout time.Duration, site Site, uploadCh chan<- UploadCFG) {
	s3Service := s3.New(getS3Session(site))

	awsItems, err := getAwsS3ItemMap(s3Service, site.Bucket)
	uploadFiles, deleteKeys, err := FilePathWalkDir(site.LocalPath, site.Exclusions, awsItems, site.BucketPath)

	if err != nil {
		logger.Fatal(err)
	}

	// Upload files
	for _, file := range uploadFiles {
		uploadCh <- UploadCFG{s3Service, file, timeout, site}
	}

	// Delete retired files
	if site.RetireDeleted {
		for _, key := range deleteKeys {
			go deleteFile(s3Service, site.Bucket, key)
			// Send 1 request per 5 seconds to avoid AWS throtling
			time.Sleep(5)
		}
	}

	// Watch directory for realtime sync
	watch(s3Service, timeout, site, uploadCh)
}
