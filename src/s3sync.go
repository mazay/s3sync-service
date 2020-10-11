//
// s3sync-service - Realtime S3 synchronisation tool
// Copyright (c) 2020  Yevgeniy Valeyev
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.
//

package main

import (
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

func getObjectSize(s3Service *s3.S3, site Site, s3Key string) int64 {
	objSize := int64(0)

	// Generate S3 parameters
	params := &s3.ListObjectsInput{
		Bucket: aws.String(site.Bucket),
		Prefix: aws.String(s3Key),
	}

	// Get object size
	obj, objErr := s3Service.ListObjects(params)
	if objErr == nil {
		for _, s3obj := range obj.Contents {
			objSize = *s3obj.Size
		}
	}

	return objSize
}

func generateS3Key(bucketPath string, localPath string, filePath string) string {
	relativePath, _ := filepath.Rel(localPath, filePath)
	return path.Join(bucketPath, relativePath)
}

func getS3Session(site Site) *session.Session {
	config := aws.Config{
		Region:     aws.String(site.BucketRegion),
		MaxRetries: aws.Int(site.S3OpsRetries),
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

	params := &s3.ListObjectsInput{
		Bucket: aws.String(site.Bucket),
		Prefix: aws.String(site.BucketPath),
	}

	err := s3Service.ListObjectsPages(params,
		func(page *s3.ListObjectsOutput, last bool) bool {
			// Process the objects for each page
			for _, s3obj := range page.Contents {
				if aws.StringValue(s3obj.StorageClass) != site.StorageClass {
					logger.Infof("storage class does not match, marking for re-upload: %s", aws.StringValue(s3obj.Key))
					items[aws.StringValue(s3obj.Key)] = "none"
				} else {
					// Update metrics
					sizeMetric.WithLabelValues(site.LocalPath, site.Bucket, site.BucketPath, site.Name).Add(float64(*s3obj.Size))
					objectsMetric.WithLabelValues(site.LocalPath, site.Bucket, site.BucketPath, site.Name).Inc()
					items[aws.StringValue(s3obj.Key)] = strings.Trim(*(s3obj.ETag), "\"")
				}
			}
			return true
		},
	)
	if err != nil {
		// Update errors metric
		errorsMetric.WithLabelValues(site.LocalPath, site.Bucket, site.BucketPath, site.Name, "cloud").Inc()
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

	// Try to get object size in case we updating already existing
	objSize := getObjectSize(s3Service, site, s3Key)

	if fileErr != nil {
		// Update errors metric
		errorsMetric.WithLabelValues(site.LocalPath, site.Bucket, site.BucketPath, site.Name, "local").Inc()
		logger.Errorf("failed to open file %q, %v", file, fileErr)
	} else {
		_, err := uploader.Upload(&s3manager.UploadInput{
			Bucket:       aws.String(site.Bucket),
			Key:          aws.String(s3Key),
			Body:         f,
			StorageClass: aws.String(site.StorageClass),
		})

		if err != nil {
			// Update errors metric
			errorsMetric.WithLabelValues(site.LocalPath, site.Bucket, site.BucketPath, site.Name, "cloud").Inc()
			logger.Errorf("failed to upload object, %v", err)
		} else {
			// Get file size
			fs, _ := f.Stat()
			fileSize := fs.Size()
			// Update metrics
			sizeMetric.WithLabelValues(site.LocalPath, site.Bucket, site.BucketPath, site.Name).Add(float64(fileSize))
			if objSize > 0 {
				// Substitute old file size
				sizeMetric.WithLabelValues(site.LocalPath, site.Bucket, site.BucketPath, site.Name).Sub(float64(objSize))
			} else {
				// Only upodate object counter when it's a new object
				objectsMetric.WithLabelValues(site.LocalPath, site.Bucket, site.BucketPath, site.Name).Inc()
			}
			logger.Infof("successfully uploaded file: %s/%s", site.Bucket, s3Key)
		}
	}
	defer f.Close()
}

func deleteFile(s3Service *s3.S3, s3Key string, site Site) {
	// Get object size
	objSize := getObjectSize(s3Service, site, s3Key)

	input := &s3.DeleteObjectInput{
		Bucket: aws.String(site.Bucket),
		Key:    aws.String(s3Key),
	}

	// Delete the object
	_, err := s3Service.DeleteObject(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				// Update errors metric
				errorsMetric.WithLabelValues(site.LocalPath, site.Bucket, site.BucketPath, site.Name, "cloud").Inc()
				logger.Errorln(aerr.Error())
			}
		} else {
			// Update errors metric
			errorsMetric.WithLabelValues(site.LocalPath, site.Bucket, site.BucketPath, site.Name, "cloud").Inc()
			logger.Errorln(err.Error())
		}
		return
	}

	// Update metrics
	if objSize > 0 {
		sizeMetric.WithLabelValues(site.LocalPath, site.Bucket, site.BucketPath, site.Name).Sub(float64(objSize))
		objectsMetric.WithLabelValues(site.LocalPath, site.Bucket, site.BucketPath, site.Name).Dec()
	}

	logger.Infof("removed s3 object: %s/%s", site.Bucket, s3Key)
}

func syncSite(site Site, uploadCh chan<- UploadCFG, checksumCh chan<- ChecksumCFG) {
	// Initi S3 session
	s3Service := s3.New(getS3Session(site))
	// Watch directory for realtime sync
	go watch(s3Service, site, uploadCh)
	// Fetch S3 objects
	awsItems, err := getAwsS3ItemMap(s3Service, site)
	if err != nil {
		logger.Errorln(err)
		osExit(4)
	} else {
		// Compare S3 objects with local
		FilePathWalkDir(site, awsItems, s3Service, uploadCh, checksumCh)
		logger.Infof("finished initial sync for site %s", site.Name)
	}
}
