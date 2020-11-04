package storage

import (
	"github.com/rumpl/ctrun/pkg/storage/minio"
	"github.com/rumpl/ctrun/pkg/storage/types"
)

func New(opts types.StorageOpts) (types.Storage, error) {
	return minio.New(opts)
}
