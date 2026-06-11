package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"platform/internal/config"
)

type MinIOStorage struct {
	client         *minio.Client
	publicEndpoint string
}

func NewMinIOStorage(cfg config.MinIOConfig) (*MinIOStorage, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.SSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}
	return &MinIOStorage{client: client, publicEndpoint: cfg.PublicEndpoint}, nil
}

func (s *MinIOStorage) Upload(ctx context.Context, bucket, objectName string, reader io.Reader, size int64, contentType string) error {
	exists, err := s.client.BucketExists(ctx, bucket)
	if err != nil {
		return fmt.Errorf("bucket exists check failed: %w", err)
	}
	if !exists {
		if err := s.client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
			return fmt.Errorf("create bucket failed: %w", err)
		}
	}

	_, err = s.client.PutObject(ctx, bucket, objectName, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	return err
}

func (s *MinIOStorage) GetObject(ctx context.Context, bucket, objectName string) (io.ReadCloser, error) {
	obj, err := s.client.GetObject(ctx, bucket, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("get object failed: %w", err)
	}
	return obj, nil
}

func (s *MinIOStorage) GetDownloadURL(ctx context.Context, bucket, objectName string, expiry time.Duration) (string, error) {
	reqParams := make(url.Values)
	reqParams.Set("response-content-disposition", fmt.Sprintf(`attachment; filename="%s"`, filepath.Base(objectName)))
	u, err := s.client.PresignedGetObject(ctx, bucket, objectName, expiry, reqParams)
	if err != nil {
		return "", err
	}
	signed := u.String()
	if s.publicEndpoint != "" {
		return rewriteURLHost(signed, s.publicEndpoint)
	}
	return signed, nil
}

func rewriteURLHost(signed, public string) (string, error) {
	parsed, err := url.Parse(signed)
	if err != nil {
		return "", err
	}
	if !strings.Contains(public, "://") {
		public = "http://" + public
	}
	pub, err := url.Parse(public)
	if err != nil {
		return "", err
	}
	if pub.Scheme != "" {
		parsed.Scheme = pub.Scheme
	}
	parsed.Host = pub.Host
	if pub.Path != "" && pub.Path != "/" {
		parsed.Path = strings.TrimRight(pub.Path, "/") + parsed.Path
	}
	if pub.RawQuery != "" && !strings.Contains(signed, pub.RawQuery) {
		if parsed.RawQuery == "" {
			parsed.RawQuery = pub.RawQuery
		} else {
			parsed.RawQuery = parsed.RawQuery + "&" + pub.RawQuery
		}
	}
	return parsed.String(), nil
}
