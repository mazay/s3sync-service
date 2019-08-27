package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func syncFile(file string, timeout time.Duration, site Site) {
	var cancelFn func()

	sess := session.Must(session.NewSession())
	svc := s3.New(sess)

	ctx := context.Background()

	if timeout > 0 {
		ctx, cancelFn = context.WithTimeout(ctx, timeout)
	}

	if cancelFn != nil {
		defer cancelFn()
	}

	// Generate S3 key
	relatiPath, _ := filepath.Rel(site.LocalPath, file)
	s3Key := filepath.Join(site.BucketPath, relatiPath)
	fmt.Printf("%+v", s3Key)

	_, err := svc.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket:       aws.String(site.Bucket),
		Key:          aws.String(s3Key),
		Body:         os.Stdin,
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
