package s3

import (
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/rumpl/ctrun/pkg/storage/types"
)

type s3Storage struct {
	client *minio.Client
}

func New(opts types.StorageOpts) (types.Storage, error) {
	minioClient, err := minio.New(opts.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(opts.AccessKey, opts.SecretKey, ""),
		Secure: true,
	})
	if err != nil {
		panic(err)
	}

	return &s3Storage{
		client: minioClient,
	}, nil
}

func (s *s3Storage) Put() error {
	return nil
}
