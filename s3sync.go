package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
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

func syncFile(file string, timeout time.Duration, site Site) {
	var cancelFn func()

	svc := s3.New(getS3Session(site))

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
		_, err := svc.PutObjectWithContext(ctx, &s3.PutObjectInput{
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
