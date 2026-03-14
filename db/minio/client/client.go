package client

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Config struct {
	Endpoint        string `env:"MINIO_ENDPOINT"`
	AccessKeyID     string `env:"MINIO_ACCESS_KEY_ID"`
	SecretAccessKey string `env:"MINIO_SECRET_ACCESS_KEY"`
}

func New(ctx context.Context, l *slog.Logger, cfg Config, buckets []string) (*minio.Client, error) {
	minioClient, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: false,
	})
	if err != nil {
		return nil, fmt.Errorf("create minio client: %w", err)
	}

	for _, bucket := range buckets {
		exists, err := minioClient.BucketExists(ctx, bucket)
		if err != nil {
			return nil, fmt.Errorf("check bucket exists: %w", err)
		}

		if exists {
			l.Info("MinIO bucket already exists", slog.String("bucket", bucket))
			continue
		}

		err = minioClient.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("create bucket %s: %w", buckets, err)
		}
	}

	l.Info("Successfully initialized Minio client")

	return minioClient, nil
}
