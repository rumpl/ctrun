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

func New(ctx context.Context, opts types.StorageOpts) (types.Storage, error) {
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

func (s *s3Storage) Url(ctx context.Context, repo string, name string) string {
	u, _ := s.client.PresignedGetObject(ctx, "ctrun", repo+"/blobs/sha256/"+name, 5*time.Minute, nil)
	return u.String()
}

func (s *s3Storage) Put(ctx context.Context, name string, r io.Reader, contentType string) error {
	_, err := s.client.PutObject(ctx, s.bucket, name, r, -1, minio.PutObjectOptions{
		ContentType: contentType,
	})

	return err
}
