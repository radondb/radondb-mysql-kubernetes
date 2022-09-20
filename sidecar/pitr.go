/*
Copyright 2021 RadonDB.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package sidecar

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/pkg/errors"
	"github.com/radondb/radondb-mysql-kubernetes/utils"
)

type S3struct struct {
	minioClient *minio.Client   // minio client for work with storage
	ctx         context.Context // context for client operations
	bucketName  string          // S3 bucket name where binlogs will be stored
	//prefix      string          // prefix for S3 requests
}

// NewS3 return new Manager, useSSL using ssl for connection with storage
func NewS3(endpoint, accessKeyID, secretAccessKey, bucketName string, sec bool) (*S3struct, error) {
	log.Info("NewS3", "endpoint", endpoint, "accesskeyId", accessKeyID, "secret", secretAccessKey, "bucketName", bucketName)
	minioClient, err := minio.New(strings.TrimRight(endpoint, "/"), &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: sec,
		// Region: "pek3b",
	})
	if err != nil {
		return nil, errors.Wrap(err, "new minio client")
	}

	return &S3struct{
		minioClient: minioClient,
		ctx:         context.TODO(),
		bucketName:  bucketName,
		//prefix:      prefix,
	}, nil
}

// download to InitFileVolumeMountPath
func (s3 *S3struct) S3Download(cfg *Config, prefix string) error {
	log.Info("now do the s3 downlad", "prefix", prefix)
	objectCh := s3.minioClient.ListObjects(s3.ctx, cfg.XCloudS3Bucket, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	})
	for object := range objectCh {
		if object.Err != nil {
			fmt.Println(object.Err)
			return object.Err
		}
		log.Info("S3Download:", "object", object.Key)
		oldObj, err := s3.minioClient.GetObject(s3.ctx, cfg.XCloudS3Bucket, object.Key, minio.GetObjectOptions{})
		if err != nil {
			log.Info("S3Download:", "GetObject Error", err.Error())
			return errors.Wrap(err, "get object")
		}
		newObj, err := os.Create(path.Join(utils.InitFileVolumeMountPath, object.Key))
		if err != nil {
			oldObj.Close()
			log.Info("S3Download:", "CreateFile Error", err.Error())
			return errors.Wrap(err, "create file")
		}
		io.Copy(newObj, oldObj)
		newObj.Close()
		oldObj.Close()
	}
	return nil
}

// Upload to S3
func (s3 *S3struct) S3Upload(cfg *BackupClientConfig, prefix string, file string) error {
	fh, err := os.Open(file)
	if err != nil {
		return err
	}
	upinfo, err := s3.minioClient.PutObject(s3.ctx, cfg.XCloudS3Bucket,
		prefix+filepath.Base(file), fh, -1, minio.PutObjectOptions{})
	if err != nil {
		return err
	}
	log.Info("S3 upload", "upinfo", upinfo)
	return nil
}
