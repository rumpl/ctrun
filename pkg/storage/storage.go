package storage

import (
	"github.com/rumpl/ctrun/pkg/storage/s3"
	"github.com/rumpl/ctrun/pkg/storage/types"
)

func New(opts types.StorageOpts) (types.Storage, error) {
	return s3.New(opts)
}
