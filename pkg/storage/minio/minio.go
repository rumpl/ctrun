package minio

import (
	"context"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/rumpl/ctrun/pkg/storage/types"
)

type s3Storage struct {
	client   *minio.Client
	bucket   string
	endpoint string
}

func New(opts types.StorageOpts) (types.Storage, error) {
	minioClient, err := minio.New(opts.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(opts.AccessKey, opts.SecretKey, ""),
		Secure: false,
	})
	if err != nil {
		panic(err)
	}

	return &s3Storage{
		client:   minioClient,
		bucket:   opts.Bucket,
		endpoint: opts.Endpoint,
	}, nil
}

func (s *s3Storage) Url(name string) string {
	u, _ := s.client.PresignedGetObject(context.Background(), "ctrun", "blobs/sha256/"+name, 5*time.Minute, nil)
	u.Host = "host.docker.internal:9000"
	return u.String()
}

func (s *s3Storage) Put(name string, r io.Reader, contentType string) error {
	_, err := s.client.PutObject(context.Background(), s.bucket, name, r, -1, minio.PutObjectOptions{
		UserMetadata: map[string]string{
			"x-amz-acl": "public-read",
		},
		ContentType: contentType,
	})

	return err
}
