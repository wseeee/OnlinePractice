package storage

import (
	"context"
	"time"
)

// ObjectStorage 对象存储抽象接口
type ObjectStorage interface {
	Put(ctx context.Context, bucket, key string, data []byte, contentType string) error
	Get(ctx context.Context, bucket, key string) ([]byte, error)
	Remove(ctx context.Context, bucket, key string) error
	PresignedGetURL(ctx context.Context, bucket, key string, expiry time.Duration) (string, error)
}
