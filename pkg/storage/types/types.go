package types

import (
	"context"
	"io"
)

type Storage interface {
	URL(ctx context.Context, repo string, name string) string
	Put(ctx context.Context, name string, r io.Reader, contentType string) error
}

type StorageOpts struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
}
