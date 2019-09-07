package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
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

func getAwsS3ItemMap(s3Service *s3.S3, site Site) (map[string]string, error) {
	var items = make(map[string]string)

	params := &s3.ListObjectsInput{Bucket: aws.String(site.Bucket)}

	err := s3Service.ListObjectsPages(params,
		func(page *s3.ListObjectsOutput, last bool) bool {
			// Process the objects for each page
			for _, s3obj := range page.Contents {
				if aws.StringValue(s3obj.StorageClass) != site.StorageClass {
					logger.Info("storage class does not match, marking for re-upload: %s", aws.StringValue(s3obj.Key))
					items[aws.StringValue(s3obj.Key)] = "none"
				} else {
					items[aws.StringValue(s3obj.Key)] = strings.Trim(*(s3obj.ETag), "\"")
				}
			}
			return true
		},
	)
	if err != nil {
		logger.Errorf("Error listing %s objects: %s", *params.Bucket, err)
		return nil, err
	}
	return items, nil
}

func uploadFile(s3Service *s3.S3, file string, site Site) {
	s3Key := generateS3Key(site.BucketPath, site.LocalPath, file)
	uploader := s3manager.NewUploader(getS3Session(site), func(u *s3manager.Uploader) {
		u.PartSize = 5 * 1024 * 1024
		u.Concurrency = 5
	})

	f, fileErr := os.Open(file)

	if fileErr != nil {
		logger.Errorf("failed to open file %q, %v", file, fileErr)
	} else {
		_, err := uploader.Upload(&s3manager.UploadInput{
			Bucket:       aws.String(site.Bucket),
			Key:          aws.String(s3Key),
			Body:         f,
			StorageClass: aws.String(site.StorageClass),
		})

		if err != nil {
			logger.Errorf("failed to upload object, %v", err)
		}

		logger.Infof("successfully uploaded file: %s/%s", site.Bucket, s3Key)
	}
	defer f.Close()
}

func deleteFile(s3Service *s3.S3, s3Key string, site Site) {
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(site.Bucket),
		Key:    aws.String(s3Key),
	}

	_, err := s3Service.DeleteObject(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				logger.Errorln(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			logger.Errorln(err.Error())
		}
		return
	}

	logger.Infof("removed s3 object: %s/%s", site.Bucket, s3Key)
}

func syncSite(site Site, uploadCh chan<- UploadCFG, checksumCh chan<- ChecksumCFG) {
	s3Service := s3.New(getS3Session(site))

	awsItems, err := getAwsS3ItemMap(s3Service, site)

	if err != nil {
		logger.Errorln(err)
		os.Exit(3)
	} else {
		FilePathWalkDir(site, awsItems, s3Service, uploadCh, checksumCh)
	}

	// Watch directory for realtime sync
	watch(s3Service, site, uploadCh)
}
