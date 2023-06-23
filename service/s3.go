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

package service

import (
	"context"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsCfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// Site is a set of options for backing up data to S3
type Site struct {
	Name            string        `yaml:"name"`
	LocalPath       string        `yaml:"local_path"`
	Bucket          string        `yaml:"bucket"`
	Endpoint        string        `yaml:"endpoint"`
	BucketPath      string        `yaml:"bucket_path"`
	BucketRegion    string        `yaml:"bucket_region"`
	StorageClass    string        `yaml:"storage_class"`
	AccessKey       string        `yaml:"access_key"`
	SecretAccessKey string        `yaml:"secret_access_key"`
	RetireDeleted   bool          `yaml:"retire_deleted"`
	Exclusions      []string      `yaml:",flow"`
	Inclusions      []string      `yaml:",flow"`
	WatchInterval   time.Duration `yaml:"watch_interval"`
	S3OpsRetries    int           `yaml:"s3_ops_retries"`
	client          *s3.Client
}

func generateS3Key(bucketPath string, localPath string, filePath string) string {
	relativePath, _ := filepath.Rel(localPath, filePath)
	return path.Join(bucketPath, relativePath)
}

func (site *Site) getS3Session() {
	var (
		err error
	)

	cfg, err := awsCfg.LoadDefaultConfig(context.TODO(),
		awsCfg.WithRegion(site.BucketRegion),
	)

	if err != nil {
		logger.Fatalf("failed to load SDK configuration, %v", err)
	}

	s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		// override endpoint if needed
		if site.Endpoint != "" {
			u, err := url.Parse(site.Endpoint)
			if err != nil {
				logger.Fatal(err)
			}
			if u.Scheme == "" {
				u.Scheme = "https"
				logger.Debugf("URL scheme for site %s was not provided, using https", site.Name)
			}
			o.EndpointResolver = s3.EndpointResolverFromURL(u.String())
		}
		// set static credentials if needed
		if site.AccessKey != "" && site.SecretAccessKey != "" {
			o.Credentials = aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(site.AccessKey, site.SecretAccessKey, ""))
		}
		o.RetryMaxAttempts = site.S3OpsRetries
	})

	site.client = s3Client
}

func (site *Site) getObjectSize(s3Key string) int64 {
	size := int64(0)

	// Get object size
	obj, err := site.client.HeadObject(context.TODO(), &s3.HeadObjectInput{
		Bucket: aws.String(site.Bucket),
		Key:    aws.String(s3Key),
	})

	if err == nil {
		size = obj.ContentLength
	}

	return size
}

func (site *Site) getAwsS3ItemMap() (map[string]string, error) {
	var (
		err   error
		iter  int
		items = make(map[string]string)
	)

	params := &s3.ListObjectsV2Input{
		Bucket: aws.String(site.Bucket),
		Prefix: aws.String(site.BucketPath),
	}

	p := s3.NewListObjectsV2Paginator(site.client, params)

	for p.HasMorePages() {
		iter++

		page, err := p.NextPage(context.TODO())
		if err != nil {
			errorsMetric.WithLabelValues(site.LocalPath, site.Bucket, site.BucketPath, site.Name, "cloud").Inc()
			logger.Fatalf("%s: failed to get page %v, %v", site.Bucket, iter, err)
		}

		// Log the objects found
		for _, s3obj := range page.Contents {
			if string(s3obj.StorageClass) != site.StorageClass {
				logger.Infof("storage class does not match, marking for re-upload: %s", string(*s3obj.Key))
				items[*aws.String(*s3obj.Key)] = "none"
			} else {
				// Update metrics
				sizeMetric.WithLabelValues(site.LocalPath, site.Bucket, site.BucketPath, site.Name).Add(float64(s3obj.Size))
				objectsMetric.WithLabelValues(site.LocalPath, site.Bucket, site.BucketPath, site.Name).Inc()
				items[*aws.String(*s3obj.Key)] = strings.Trim(*(s3obj.ETag), "\"")
			}
		}
	}

	return items, err
}

func (site *Site) uploadFile(file string) {
	var err error

	s3Key := generateS3Key(site.BucketPath, site.LocalPath, file)
	uploader := manager.NewUploader(site.client, func(u *manager.Uploader) {
		u.PartSize = 5 * 1024 * 1024
		u.Concurrency = 5
	})

	f, err := os.Open(file)

	if err != nil {
		// Update errors metric
		errorsMetric.WithLabelValues(site.LocalPath, site.Bucket, site.BucketPath, site.Name, "local").Inc()
		logger.Errorf("failed to open file %q, %v", file, err)
		return
	}
	defer f.Close()

	// Try to get object size in case we updating already existing
	objSize := site.getObjectSize(s3Key)

	_, err = uploader.Upload(context.TODO(), &s3.PutObjectInput{
		Bucket:       aws.String(site.Bucket),
		Key:          aws.String(s3Key),
		Body:         f,
		StorageClass: types.StorageClass(site.StorageClass),
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

func (site *Site) deleteFile(s3Key string) {
	// Get object size
	objSize := site.getObjectSize(s3Key)

	input := &s3.DeleteObjectInput{
		Bucket: aws.String(site.Bucket),
		Key:    aws.String(s3Key),
	}

	// Delete the object
	_, err := site.client.DeleteObject(context.TODO(), input)
	if err != nil {
		// Update errors metric
		errorsMetric.WithLabelValues(site.LocalPath, site.Bucket, site.BucketPath, site.Name, "cloud").Inc()
		logger.Errorln(err.Error())
		return
	}

	// Update metrics
	if objSize > 0 {
		sizeMetric.WithLabelValues(site.LocalPath, site.Bucket, site.BucketPath, site.Name).Sub(float64(objSize))
		objectsMetric.WithLabelValues(site.LocalPath, site.Bucket, site.BucketPath, site.Name).Dec()
	}

	logger.Infof("removed s3 object: %s/%s", site.Bucket, s3Key)
}
