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
	BucketName      string `env:"MINIO_BUCKET_NAME"`
}

func New(ctx context.Context, l *slog.Logger, cfg Config) (*minio.Client, error) {
	minioClient, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: false,
	})
	if err != nil {
		return nil, fmt.Errorf("create minio client: %w", err)
	}

	exists, err := minioClient.BucketExists(ctx, cfg.BucketName)
	if err != nil {
		return nil, fmt.Errorf("check bucket exists: %w", err)
	}

	if exists {
		l.Info("MinIO bucket already exists", slog.String("bucket", cfg.BucketName))
		return minioClient, nil
	}

	err = minioClient.MakeBucket(ctx, cfg.BucketName, minio.MakeBucketOptions{})
	if err != nil {
		return nil, fmt.Errorf("create bucket %s: %w", cfg.BucketName, err)
	}

	l.Info("Successfully created missing MinIO bucket", slog.String("bucket", cfg.BucketName))

	return minioClient, nil
}
