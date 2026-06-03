package storage

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinIOStorage struct {
	Client *minio.Client
}

// NewMinIOStorage 从环境变量创建 MinIO 客户端，确保桶存在
func NewMinIOStorage() (*MinIOStorage, error) {
	endpoint := os.Getenv("MINIO_ENDPOINT")
	accessKey := os.Getenv("MINIO_ACCESS_KEY")
	secretKey := os.Getenv("MINIO_SECRET_KEY")
	useSSL := os.Getenv("MINIO_USE_SSL") == "true"

	if endpoint == "" || accessKey == "" || secretKey == "" {
		return nil, fmt.Errorf("MINIO_ENDPOINT / MINIO_ACCESS_KEY / MINIO_SECRET_KEY are required")
	}

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("minio.New: %w", err)
	}

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err = client.ListBuckets(ctx)
	if err != nil {
		return nil, fmt.Errorf("minio connection failed: %w", err)
	}
	log.Println("MinIO connected successfully")

	s := &MinIOStorage{Client: client}

	// 确保桶存在
	buckets := []string{
		getEnv("MINIO_BUCKET_AVATAR", "oj-avatar"),
		getEnv("MINIO_BUCKET_CODE", "oj-code"),
	}
	for _, b := range buckets {
		if err := s.EnsureBucket(ctx, b); err != nil {
			return nil, fmt.Errorf("ensure bucket %s: %w", b, err)
		}
	}

	return s, nil
}

func (s *MinIOStorage) EnsureBucket(ctx context.Context, bucket string) error {
	exists, err := s.Client.BucketExists(ctx, bucket)
	if err != nil {
		return err
	}
	if !exists {
		if err := s.Client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
			return err
		}
		log.Printf("MinIO bucket created: %s", bucket)
	}
	return nil
}

func (s *MinIOStorage) Put(ctx context.Context, bucket, key string, data []byte, contentType string) error {
	_, err := s.Client.PutObject(ctx, bucket, key, bytes.NewReader(data), int64(len(data)),
		minio.PutObjectOptions{ContentType: contentType})
	return err
}

func (s *MinIOStorage) Get(ctx context.Context, bucket, key string) ([]byte, error) {
	obj, err := s.Client.GetObject(ctx, bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	defer obj.Close()
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(obj); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (s *MinIOStorage) Remove(ctx context.Context, bucket, key string) error {
	return s.Client.RemoveObject(ctx, bucket, key, minio.RemoveObjectOptions{})
}

func (s *MinIOStorage) PresignedGetURL(ctx context.Context, bucket, key string, expiry time.Duration) (string, error) {
	u, err := s.Client.PresignedGetObject(ctx, bucket, key, expiry, nil)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
