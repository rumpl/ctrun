package s3

import (
	"context"
	"fmt"
	"io"

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
		Secure: true,
	})
	if err != nil {
		return nil, err
	}

	return &s3Storage{
		client:   minioClient,
		bucket:   opts.Bucket,
		endpoint: opts.Endpoint,
	}, nil
}

func (s *s3Storage) URL(_ context.Context, _ string, name string) string {
	return fmt.Sprintf("https://%s.%s/blobs/sha256/%s", s.bucket, s.endpoint, name)
}

func (s *s3Storage) Put(ctx context.Context, name string, r io.Reader, contentType string) error {
	_, err := s.client.PutObject(ctx, s.bucket, name, r, -1, minio.PutObjectOptions{
		UserMetadata: map[string]string{
			"x-amz-acl": "public-read",
		},
		ContentType: contentType,
	})

	return err
}
