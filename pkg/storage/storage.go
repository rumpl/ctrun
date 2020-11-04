package storage

import (
	"context"

	"github.com/rumpl/ctrun/pkg/storage/minio"
	"github.com/rumpl/ctrun/pkg/storage/types"
)

func New(ctx context.Context, opts types.StorageOpts) (types.Storage, error) {
	return minio.New(ctx, opts)
}
