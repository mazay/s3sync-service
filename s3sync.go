package main

import (
	"context"
	"fmt"
	"log"
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

func getS3Session(site Site) *session.Session {
	config := aws.Config{Region: aws.String(site.BucketRegion)}

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

	// Generate S3 key
	relativePath, _ := filepath.Rel(site.LocalPath, file)
	s3Key := filepath.Join(site.BucketPath, relativePath)

	f, fileErr := os.Open(file)

	if fileErr != nil {
		fmt.Fprintf(os.Stderr, "failed to open file %q, %v", file, fileErr)
	} else {
		_, err := s3Service.PutObjectWithContext(ctx, &s3.PutObjectInput{
			Bucket:       aws.String(site.Bucket),
			Key:          aws.String(s3Key),
			Body:         f,
			StorageClass: aws.String(site.StorageClass),
		})

		if err != nil {
			if aerr, ok := err.(awserr.Error); ok && aerr.Code() == request.CanceledErrorCode {
				fmt.Fprintf(os.Stderr, "upload canceled due to timeout, %v\n", err)
			} else {
				fmt.Fprintf(os.Stderr, "failed to upload object, %v\n", err)
			}
			os.Exit(1)
		}

		fmt.Printf("successfully uploaded file to %s/%s\n", site.Bucket, s3Key)
	}
}

func syncFile(timeout time.Duration, site Site) {
	s3Service := s3.New(getS3Session(site))

	awsItems, err := getAwsS3ItemMap(s3Service, site.Bucket)
	files, err := FilePathWalkDir(site.LocalPath, site.Exclusions, awsItems, site.BucketPath)

	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		uploadFile(s3Service, file, timeout, site)
	}
}
