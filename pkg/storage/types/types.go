package types

import "io"

type Storage interface {
	Url(name string) string
	Put(name string, r io.Reader, contentType string) error
}

type StorageOpts struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
}
