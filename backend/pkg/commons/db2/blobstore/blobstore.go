package blobstore

import (
	"context"
	"io"
)

type BlobStore interface {
	Put(ctx context.Context, bucket string, key string, blob Blob) error
	Exist(ctx context.Context, bucket string, key string) (bool, error)
	Get(ctx context.Context, bucket string, key string) (Blob, error)
}

type Blob struct {
	Body        io.Reader
	ContentType *string
	Metadata    map[string]string
}
