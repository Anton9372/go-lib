package client

import (
	"context"
	"fmt"

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

	// Ping
	_, err = minioClient.ListBuckets(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("ping minio (list buckets): %w", err)
	}

	return minioClient, nil
}
