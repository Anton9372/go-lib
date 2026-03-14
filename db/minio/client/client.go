package client

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Config struct {
	Endpoint        string `env:"MINIO_ENDPOINT"`
	AccessKeyID     string `env:"MINIO_ACCESS_KEY_ID"`
	SecretAccessKey string `env:"MINIO_SECRET_ACCESS_KEY"`
}

func New(cfg Config) (*minio.Client, error) {
	minioClient, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: false,
	})
	if err != nil {
		return nil, fmt.Errorf("create minio client: %w", err)
	}

	return minioClient, nil
}

func InitBuckets(ctx context.Context, l *slog.Logger, client *minio.Client, buckets []string) error {
	for _, bucket := range buckets {
		exists, err := client.BucketExists(ctx, bucket)
		if err != nil {
			return fmt.Errorf("check bucket %s exists: %w", bucket, err)
		}

		if exists {
			l.Info("MinIO bucket already exists", slog.String("bucket", bucket))
			continue
		}

		err = client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("create bucket %s: %w", bucket, err)
		}
	}

	l.Info("Successfully initialized MinIO buckets", slog.String("buckets", strings.Join(buckets, ",")))
	return nil
}
